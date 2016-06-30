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


	userOffering		string
	userRequesting		string
	userRegionAdmin 	string
	userSuperAdmin		string

	regionID         	string
	offerID				string
	requestID			string
	matchingID			string
	notificationID		string
)

// init() will always be called before TestMain or Tests
func init() {
	fmt.Println("initiating...")

	// initialize and configure server
	app = InitApp()
}

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



func TestMeMatchings(t *testing.T) {
	resp := app.RequestWithJWT("GET", "/me/matchings", nil, tokenUserJery)
	if resp.Code != 200 {
		t.Error("could not load matchings")
		return
	}
}



func TestRegionsDef(t *testing.T) {
	// create region
	plCreateRegion := CreateRegionPayload {
		"Milkshake Region",
		"Ma Region brings all the boys in the yard",
		Boundaries{
			[]Location{
				Location{
				0.0,
				0.0,
				},
				Location{
					5.0,
					0.0,
				},
				Location{
					5.0,
					5.0,
				},
				Location{
					0.0,
					0.0,
				},
			},
		},
	}

	resp := app.RequestWithJWT("POST", "/regions", plCreateRegion, tokenUserJery)
	if resp.Code != http.StatusCreated {
		t.Error("could not create region")
		return
	}

	// change name
	data := parseResponse(resp)
	NameUpdated := "NAME UPDATED"
	plCreateRegion.Name = NameUpdated
	resp = app.RequestWithJWT("PUT", "/regions/" + data["ID"].(string), plCreateRegion, tokenAdmin)
	if resp.Code != http.StatusOK {
		t.Error("could not update region")
		return
	}

	// check if name was actually changed
	resp = app.RequestWithJWT("GET", "/regions/" + data["ID"].(string), nil, tokenUserJery)
	data = parseResponse(resp)
	if data["Name"].(string) != NameUpdated {
		t.Error("failed to update region name")
		return
	}

}


// ----------------------------------------------------------------- AUTH

func TestAuth(t *testing.T) {
	// [X] Login : N - check if returns JWT
	// Authorize : L - test if returns user
	// Renew Token : L - check if returns new JWT
	// Logout : L - NOT COVERED
}


func LoginTest(t *testing.T, Email string, Password string, AssertCode int) string{
	loginParams := LoginPayload{
		Email,
		Password,
	}
	resp := app.Request("POST", "/auth", loginParams)


	// expected login to work 
	if AssertCode == 200 && resp.Code != 200 {
		t.Error("User login failed", resp.Body.String())
		return ""
	}

	// expecting login to fail
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("User login unexpected response %d", resp.Code))
		}
		return ""
	}

	// check if access token exists
	dat := parseResponse(resp)
	if dat["AccessToken"] == nil {
		t.Error("User Access Token is empty")
		return ""
	}

	// return token
	return dat["AccessToken"].(string)
}

// ----------------------------------------------------------------- USERS

func TestUsers(t *testing.T) {
	// [X] CreateUser : U
	// UpdateUser : S
	// ListUsers : A
	// GetUser : A
	// PromoteToSystemAdmin : S
}


func CreateUserTest(t *testing.T, Email string, Password string, Name string, AssertCode int) {
	// Create User
	createParams := CreateUserPayload{
		Name:          Name,
		PreferredName: Name + " Pref",
		Mail:          Email,
		Password:      Password,
		PhoneNumbers:  make([]string, 1),
	}
	resp := app.Request("POST", "/users", createParams)

	if AssertCode == 200 && resp.Code != 200{
		t.Error("User creation failed")
	}

	if AssertCode == 400 && resp.Code != 400 {
		t.Error("User should already exist")
	}
}


// ----------------------------------------------------------------- GROUPS

func TestGroups(t *testing.T) {
	// GetGroups : S
	// ListSystemAdmins : S
}


// ----------------------------------------------------------------- OFFERS

func TestOffers(t *testing.T) {
	// [X] CreateOffer - L
	// GetOffer - C
	// UpdateOffer - C
}

func CreateOfferTest(t *testing.T, jwt string, Name string, Location GeoLocation, Validity string, AssertCode int) string{
	plCreateOffer := CreateOfferPayload {
		Name,
		Location,
		[]string{},
		Validity,
	}

	// check if offer was created
	resp := app.RequestWithJWT("POST", "/offers", plCreateOffer, jwt)
	
	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not CreateOffer")
		return ""
	}
	if AssertCode == 400 {
		if resp.Code != 400{
			t.Error("CreateOffer should return BadRequest, but didnt")
		}
		return ""
	}

	data := parseResponse(resp)
	return data["ID"].(string)
}

// ----------------------------------------------------------------- REQUESTS

func TestRequests(t *testing.T) {
	// [X] CreateRequest - L
	// GetRequest - C
	// UpdateRequest - C
}

func CreateRequestTest(t *testing.T, jwt string, Name string, Location GeoLocation, Validity string, AssertCode int) string {
	plCreateRequest := CreateRequestPayload {
		Name,
		Location,
		[]string{},
		Validity,
	}

	resp := app.RequestWithJWT("POST", "/requests", plCreateRequest, jwt)

	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not create request")
		return ""
	}
	if AssertCode == 400 {
		if resp.Code != 400{
			t.Error(fmt.Printf("CreateRequest should return BadRequest, but did return %d", resp.Code))
		}
		return ""
	}

	data := parseResponse(resp)
	return data["ID"].(string)
}

// ----------------------------------------------------------------- MATCHINGS

func TestMatchings(t *testing.T) {
	// [X] CreateMatching - A
	// GetMatching - C
	// UpdateMatching - C
}

func CreateMatchingTest(t *testing.T, jwt string, Region string, Offer string, Request string, AssertCode int) string {
	plCreateMatching := CreateMatchingPayload{
		Region,
		Request,
		Offer,
	}

	resp := app.RequestWithJWT("POST", "/matchings", plCreateMatching, tokenAdmin)
	
	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not create matching")
		return ""
	}
	if AssertCode == 400 {
		if resp.Code != 400{
			t.Error(fmt.Printf("CreateMatching should return BadRequest, but did return %d", resp.Code))
		}
		return ""
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("CreateMatching should return UnAuthorized but didnt")
		}
		return "" 
	}

	data := parseResponse(resp)
	return data["ID"].(string)
}

// ----------------------------------------------------------------- REGIONS

func TestRegions(t *testing.T) {
	// [X] CreateRegion - L
	// ListRegions - U
	// [X] GetRegion - U
	// UpdateRegion - A
	// ListOffersForRegion - A
	// ListRequestsForRegion - A
	// ListMatchingsForRegion - A
	// [X] PromoteUserToAdminForRegion - A
	// ListAdminsForRegion - A
}

func CreateRegionTest(t *testing.T, jwt string, Name string, Desc string, Locations []Location, AssertCode int) string {
	// create region
	plCreateRegion := CreateRegionPayload {
		Name,
		Desc,
		Boundaries{
			Locations,
		},
	}

	resp := app.RequestWithJWT("POST", "/regions", plCreateRegion, jwt)
	if AssertCode == 201 && resp.Code != 201{
		t.Error("could not create region")
		return ""
	}

	if AssertCode == 400 { 
		if resp.Code != 400 {
			t.Error("CreateRegion should return BadRequest but didnt")
		}
		return ""
	}

	// check if ID exists
	dat := parseResponse(resp)
	if dat["ID"] == nil {
		t.Error("CreateRegion did not return ID parameter")
		return ""
	}

	return dat["ID"].(string)
}

func GetRegionTest(t *testing.T, jwt string, Region string, AssertCode int) map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/regions/" + Region, nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("GetRegion failed")
		return map[string]interface{}{}
	}

	if AssertCode == 400 {
		if resp.Code != 400  {
			t.Error("GetRegion should return bad request, but didnt")
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	if(data["ID"] != Region) {
		t.Error("Wrong region was returned")
	}

	return data
}

func PromoteUserToAdminForRegionTest(t *testing.T, jwt string, Email string, Region string, AssertCode int) bool{
	promoteParams := PromoteAdminPayload{Email}
	resp := app.RequestWithJWT("POST", "/regions/" + Region + "/admins", promoteParams, jwt)
	

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Promoting User to Admin did not work, but should: ", resp.Body.String())
		return false
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error("PromoteUserToAdmin should return BadRequest but didnt")
		}
		return false 
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("PromoteUserToAdmin should return UnAuthorized but didnt")
		}
		return false 
	}
	if AssertCode == 404 {
		if resp.Code != 404 {
			t.Error("PromoteUserToAdmin should return NotFound but didnt")
		}
		return false 
	}

	return true
}


// ------------------------------------------------------------------------------- ME

func TestMe(t *testing.T) {
	// GetMe - L
	// UpdateMe - L
	// ListUserOffers - L
	// ListUserRequests - L
	// ListUserMatchings - L
}

func TestNotifications(t *testing.T) {
	// ListNotifications - L
	// UpdateNotification - C
}



// ----------------------------------------------------- SCENARIO ALPHA

func TestSetupAlpha(t *testing.T) {
	fmt.Println("\n--------------------- SetupTestAlpha ---------------------\n")

	// HACKY : CreateUserTest AssertCode set to zero because no prior knowledge exists about database
	// response 200 and 400 could both be valid, but our testframework does not support multi-asserts
	// VALID: CreateUser + Login
	emailRegionAdmin := "regionadmin@test.org"
	CreateUserTest(t, "offering@test.org", "ICanOfferAllThemHelp666!", "OfferBoy", 0)
	userOffering = LoginTest(t, "offering@test.org", "ICanOfferAllThemHelp666!", 200)
	CreateUserTest(t, "requesting@test.org", "INeedAllThemHelp666!", "RequestDude", 0)
	userRequesting = LoginTest(t, "requesting@test.org", "INeedAllThemHelp666!", 200)
	CreateUserTest(t, emailRegionAdmin, "LetMeAdminAllYourHelp666!", "AdminMate", 0)
	userRegionAdmin = LoginTest(t, emailRegionAdmin, "LetMeAdminAllYourHelp666!", 200)

	// INVALID: Login superadmin
	LoginTest(t, "admin@example.org", "nonononooo", 400)
	// VALID: Login superadmin
	userSuperAdmin = LoginTest(t, "admin@example.org", "CaTUstrophyAdmin123$", 200)

	// INVALID: CreateRegion
	CreateRegionTest(t, userRegionAdmin, "", "", []Location{}, 400)
	// VALID: CreateRegion
	regionName := "Milkshake Region"
	regionID = CreateRegionTest(t, userRegionAdmin, regionName, "Ma Region brings all the boys in the yard", 
		[]Location{ Location{ 10.0, 0.0, }, Location{ 11.0, 0.0, }, Location{ 10.0, 1.0, }, Location{ 10.0, 0.0, }, }, 201)

	// INVALID: GetRegion 
	GetRegionTest(t, userOffering, regionID + "a", 400)
	// VALID: GetRegion 
	region := GetRegionTest(t, userOffering, regionID, 200)
	// compare region names of Create & Get
	if region["Name"] != regionName {
		t.Error("CreateRegion followed by GetRegion returns not the same value for field Region.Name")
	}

	// INVALID: PromoteUserToAdminForRegion
	PromoteUserToAdminForRegionTest(t, userSuperAdmin, "", regionID, 400)
	// INVALID: PromoteUserToAdminForRegion
	PromoteUserToAdminForRegionTest(t, userRegionAdmin, emailRegionAdmin, regionID, 401)
	// INVALID: PromoteUserToAdminForRegion
	PromoteUserToAdminForRegionTest(t, userSuperAdmin, "nobody@donotexist.com", regionID, 404)
	// VALID: PromoteUserToAdminForRegion
	PromoteUserToAdminForRegionTest(t, userSuperAdmin, emailRegionAdmin, regionID, 200)
}

func TestMatchingAlpha(t *testing.T) {
	fmt.Println("\n--------------------- MatchingTestAlpha ---------------------\n")

	// INVALID CreateOffer
	CreateOfferTest(t, userOffering, "", GeoLocation{10.2, .0}, "2017-11-01T22:08:41+00:00", 400)
	// VALID CreateOffer
	offerID = CreateOfferTest(t, userOffering, "Milk x10", GeoLocation{10.2, .0}, "2017-11-01T22:08:41+00:00", 201)

	// INVALID CreateRequest
	CreateRequestTest(t, userRequesting, "", GeoLocation{10.3, 0.2}, "2016-11-01T22:08:41+00:00", 400)
	// VALID CreateRequest
	requestID = CreateRequestTest(t, userRequesting, "Me thirsty", GeoLocation{10.3, 0.2}, "2016-11-01T22:08:41+00:00", 201)

	// INVALID CreateMatching
	CreateMatchingTest(t, userOffering, regionID, offerID, requestID, 401)
	CreateMatchingTest(t, userRegionAdmin, "", offerID, requestID, 400)
	// VALID CreateMatching
	matchingID = CreateMatchingTest(t, userRegionAdmin, regionID, offerID, requestID, 201)



}
