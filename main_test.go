package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"net/http/httptest"
	"testing"
)

var (
	app              *App
	plCreateUserJery CreateUserPayload
	tokenUserJery    string
	userJeryMail     string
	tokenAdmin       string
	regionID         string
)

// init() will always be called before TestMain or Tests
func init() {
	fmt.Println("initiating...")

	// initialize and configure server
	app = InitApp()

}

// if TestMain exists no Test functions will be called
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestUser(t *testing.T) {

	// create user jery
	userJeryMail = "stupidtestthingthatsuckshard!12"
	plCreateUserJery = CreateUserPayload{
		"German Jery",
		"Jery",
		"jery@jery.com",
		userJeryMail,
	}
	resp := ResponsePOST("/users", plCreateUserJery)

	// check if everything went well
	if resp.Code != 200 && resp.Code != 400 {
		t.Error("User creation failed", resp.Code)
		return
	}

	loginB := LoginPayload{
		plCreateUserJery.Mail,
		plCreateUserJery.Password,
	}
	resp = ResponsePOST("/auth", loginB)

	if resp.Code != 200 {
		t.Error("User login failed", resp.Body.String())
		return
	}

	dat := parseResponse(resp)
	if dat["AccessToken"] == nil {
		t.Error("User Access Token is empty")
		return
	}
	tokenUserJery = dat["AccessToken"].(string)
}

func TestAdminLogin(t *testing.T) {
	loginB := LoginPayload{
		"admin@example.org",
		"CaTUstrophyAdmin123$",
	}
	resp := ResponsePOST("/auth", loginB)

	if resp.Code != 200 {
		t.Error("User login failed", resp.Body.String())
		return
	}

	dat := parseResponse(resp)
	if dat["AccessToken"] == nil {
		t.Error("Admin Access Token is empty")
		return
	}
	tokenAdmin = dat["AccessToken"].(string)
}

func TestWhatevs(t *testing.T) {
	fmt.Println("new test with token:", tokenUserJery)

}

type PromoteAdminPayload struct {
	Mail string
}

func TestGetRegions(t *testing.T) {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/regions", nil)
	if err != nil {
		t.Error("Error while trying to get regions", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(resp, req)
	regions := parseResponseToArray(resp)
	regionID = regions[0]["ID"].(string)
}

func TestPostRegionAdmin(t *testing.T) {
	PromoteJery := PromoteAdminPayload{userJeryMail}

	url := "/regions/" + regionID + "/admins"
	resp := ResponsePOSTwithJWT(url, PromoteJery, tokenUserJery)
	if resp.Code != 400 {
		t.Error("Promoting User Jery as Jery gave no 400, but should: ", resp.Body.String())
		return
	}
	PromoteNonExisting := PromoteAdminPayload{"This is no valid email"}
	resp = ResponsePOSTwithJWT(url, PromoteNonExisting, tokenAdmin)
	if resp.Code == 500 {
		t.Error("Promoting non existant user gave 500: ", resp.Body.String())
		return
	}
	if resp.Code == 200 || resp.Code == 201 {
		t.Error("Promoting non existant user gave ", resp.Code, resp.Body.String())
		return
	}
	PromoteNonExisting = PromoteAdminPayload{"emailthatsnotinthesystem@example.org"}
	resp = ResponsePOSTwithJWT(url, PromoteNonExisting, tokenAdmin)
	if resp.Code == 500 {
		t.Error("Promoting non existant user gave 500: ", resp.Body.String())
		return
	}
	if resp.Code == 200 || resp.Code == 201 {
		t.Error("Promoting non existant user gave ", resp.Code)
		return
	}
	resp = ResponsePOSTwithJWT(url, PromoteJery, tokenAdmin)
	if resp.Code != 201 && resp.Code != 200 {
		t.Error("Promoting User Jery as Admin did not work, but should: ", resp.Body.String())
		return
	}

}

func NewRequestPOST(url string, body interface{}) *http.Request {
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewRequestPOSTwithJWT(url string, body interface{}, jwt string) *http.Request {
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", ("Bearer " + jwt))
	return req
}

func ResponsePOST(url string, body interface{}) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	req := NewRequestPOST(url, body)
	app.Router.ServeHTTP(resp, req)
	return resp
}

func ResponsePOSTwithJWT(url string, body interface{}, jwt string) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	req := NewRequestPOSTwithJWT(url, body, jwt)
	app.Router.ServeHTTP(resp, req)
	return resp
}

func parseResponse(resp *httptest.ResponseRecorder) map[string]interface{} {
	var dat map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &dat); err != nil {
		panic(err)
	}
	return dat
}

func parseResponseToArray(resp *httptest.ResponseRecorder) []map[string]interface{} {
	var dat []map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &dat); err != nil {
		panic(err)
	}
	return dat
}
