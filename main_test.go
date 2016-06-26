package main

import (
	"fmt"
	"log"
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
	userJeryPW       string
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
	userJeryPW = "stupidtestthingthatsuckshard!12"
	userJeryMail = "jery1@jery.jery"
	plCreateUserJery = CreateUserPayload{
		Name:          "German Jery",
		PreferredName: "Jery",
		Mail:          userJeryMail,
		Password:      userJeryPW,
		PhoneNumbers:  make([]string, 1),
	}
	resp := app.Request("POST", "/users", plCreateUserJery)

	// check if everything went well
	if resp.Code != 200 && resp.Code != 400 {
		t.Error("User creation failed", resp.Code)
		return
	}

	loginB := LoginPayload{
		plCreateUserJery.Mail,
		plCreateUserJery.Password,
	}
	resp = app.Request("POST", "/auth", loginB)

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
	resp := app.Request("POST", "/auth", loginB)

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
	log.Println("Admin Token: ", tokenAdmin)
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

	PromoteNonExisting := PromoteAdminPayload{"This is no valid email"}
	resp := app.RequestWithJWT("POST", url, PromoteNonExisting, tokenAdmin)
	if resp.Code == 500 {
		t.Error("Promoting non-existant user gave 500: ", resp.Body.String())
		return
	}
	if resp.Code == 200 || resp.Code == 201 {
		t.Error("Promoting non-existant user gave ", resp.Code, resp.Body.String())
		return
	}

	PromoteNonExisting = PromoteAdminPayload{"emailthatsnotinthesystem@example.org"}
	resp = app.RequestWithJWT("POST", url, PromoteNonExisting, tokenAdmin)
	if resp.Code == 500 {
		t.Error("Promoting non-existant user gave 500: ", resp.Body.String())
		return
	}
	if resp.Code == 200 || resp.Code == 201 {
		t.Error("Promoting non-existant user gave ", resp.Code)
		return
	}

	resp = app.RequestWithJWT("POST", url, PromoteJery, tokenAdmin)
	if resp.Code != 201 && resp.Code != 200 {
		t.Error("Promoting User Jery as Admin did not work, but should: ", resp.Body.String())
		return
	}
}


func TestCreateMatching(t *testing.T) {
	// create offer
	plCreateOffer := CreateOfferPayload {
		"new name",
		GeoLocation{123.2, 113.1},
		[]string{},
		"2017-11-01T22:08:41+00:00",
	}

	resp := app.RequestWithJWT("POST", "/offers", plCreateOffer, tokenUserJery)
	dataCreateOffer := parseResponse(resp)

	if resp.Code != 201 {
		t.Error("Could not create offer")
		return
	}

	// create request
	plCreateRequest := CreateRequestPayload {
		"new name",
		GeoLocation{123.2, 113.1},
		[]string{},
		"2017-11-01T22:08:41+00:00",
	}

	resp = app.RequestWithJWT("POST", "/requests", plCreateRequest, tokenUserJery)
	dataCreateRequest := parseResponse(resp)

	if resp.Code != 201 {
		t.Error("Could not create request")
		return
	}


	// create matching
	plCreateMatching := CreateMatchingPayload{
		"830c13c6-30b4-407d-966c-50d46d9f8d31",
		dataCreateRequest["ID"].(string),
		dataCreateOffer["ID"].(string),
	}
	resp = app.RequestWithJWT("POST", "/matchings", plCreateMatching, tokenAdmin)

	if resp.Code != 201 {
		t.Error("Could not create matching")
		return
	}
}


func TestMeMatchings(t *testing.T) {
	resp := app.RequestWithJWT("GET", "/me/matchings", nil, tokenUserJery)
	if resp.Code != 200 {
		t.Error("could not load matchings")
		return
	}
}