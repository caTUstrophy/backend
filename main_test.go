	package main


import (
	"fmt"
	"bytes"
	"net/http"
	"encoding/json"

	"testing"
    "net/http/httptest"
)

var (
	app *App
	plCreateUserJery CreateUserPayload
	tokenUserJery string
)

// init() will always be called before TestMain or Tests
func init() {
	fmt.Println("initiating...")

	// initialize and configure server
    app = InitApp()

}


// if TestMain exists no Test functions will be called
func TestMain(m *testing.M) {
	m.Run()
}

func TestUser(t *testing.T) {

    // create user jery
	plCreateUserJery = CreateUserPayload {
		"German Jery",
		"Jery",
		"jery@jery.com",
		"stupidtestthingthatsuckshard!12",
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

func TestWhatevs(t *testing.T) {
	fmt.Println("new test with token:", tokenUserJery)

}

func NewRequestPOST(url string, body interface{}) *http.Request{
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
    req.Header.Set("Content-Type", "application/json")
    return req
}

func ResponsePOST(url string, body interface{}) *httptest.ResponseRecorder{
    resp := httptest.NewRecorder()
    req := NewRequestPOST(url, body)
    app.Router.ServeHTTP(resp, req)
    return resp
}

func parseResponse(resp *httptest.ResponseRecorder) map[string]interface{}{
	var dat map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &dat); err != nil {
        panic(err)
    }
    return dat
}



