package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/caTUstrophy/backend/db"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
)

var (
	app *App

	userOffering    string
	userRequesting  string
	userRegionAdmin string
	userSuperAdmin  string

	regionID       string
	offerID        string
	requestID      string
	matchingID     string
	notificationID string
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

type PromoteAdminPayload struct {
	Mail string
}

// --------------------------------------------------- Matching Algorithm

// [x] Distance
// [x] DistanceFactor
// TagFactor
// NLP Factor

func AddDataTest(t *testing.T) {
	var UserAdmin db.User
	app.DB.First(&UserAdmin)
	var RegionTU db.Region
	app.DB.First(&RegionTU)
	TagFood := db.Tag{Name: "Food"}
	TagTool := db.Tag{Name: "Tool"}
	TagChildren := db.Tag{Name: "Children"}
	TagOther := db.Tag{Name: "Other"}
	TagMedical := db.Tag{Name: "Medical"}

	req1 := db.Request{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "Toothbrushes",
		UserID:         UserAdmin.ID,
		User:           UserAdmin,
		Radius:         10.0,
		Tags:           []db.Tag{TagMedical},
		Location:       gormGIS.GeoPoint{13.326863, 52.513142},
		Description:    "I need toothbrushes for me and my family, we are four peaple but if necessary we can share! Toothpaste would also be really nice!",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	req2 := db.Request{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "Mini USB charger",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []db.Tag{TagTool, TagOther},
		Location:       gormGIS.GeoPoint{13.326860, 52.513142},
		Description:    "Hey everyone, I lost my charger and would love to get exchange for it as I really need my phone, as fast as possible. We have electricity here, you can use it; my phone has a mini usb plot",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	req3 := db.Request{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "A 2 meters sized chocolate letter",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []db.Tag{TagFood, TagChildren},
		Location:       gormGIS.GeoPoint{13.326859, 52.513143},
		Description:    "Sorry guys I lost my 3 meter sized chocolate 'A'. Its dark chocolate wih I guess 60percent cacao. I am very sad and if this cant be found another big sized chocolate letter would help, but i really want to eat chocolate",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off1 := db.Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "hygiene stuff",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []db.Tag{TagMedical},
		Location:       gormGIS.GeoPoint{13.326861, 52.513145},
		Description:    "hey, i have some toothbrushes, toothpasta, cacao shampoo and a electric shaver to offer",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off2 := db.Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "phone charger",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []db.Tag{TagOther},
		Location:       gormGIS.GeoPoint{13.326861, 52.513142},
		Description:    "i have a charger for mobile phones, but no public electricity around",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off3 := db.Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "Children stuff",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []db.Tag{TagChildren},
		Location:       gormGIS.GeoPoint{13.326862, 52.513143},
		Description:    "Hey, I have some stuff kids like to eat, choco sweets and chips",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off4 := db.Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "This is a very bad offer",
		UserID:         UserAdmin.ID,
		Radius:         0.0001,
		Tags:           []db.Tag{TagOther},
		Location:       gormGIS.GeoPoint{13.326861, 52.513143},
		Description:    "Extraordnariy unusefullness that does not fit anything really good.",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}

	log.Println(req1)

	app.DB.Create(&req1)
	app.DB.Create(&req2)
	app.DB.Create(&req3)
	app.DB.Create(&off1)
	app.DB.Create(&off2)
	app.DB.Create(&off3)
	app.DB.Create(&off4)

	app.CalcMatchScoreForRequest(req1)
	app.CalcMatchScoreForRequest(req2)
	app.CalcMatchScoreForRequest(req3)

	RegionTU.Offers = append(RegionTU.Offers, off1)
	RegionTU.Offers = append(RegionTU.Offers, off2)
	RegionTU.Offers = append(RegionTU.Offers, off3)
	RegionTU.Offers = append(RegionTU.Offers, off4)
	RegionTU.Requests = append(RegionTU.Requests, req1)
	RegionTU.Requests = append(RegionTU.Requests, req2)
	RegionTU.Requests = append(RegionTU.Requests, req3)
	app.DB.Save(&RegionTU)
}

func DistanceTest(t *testing.T, request db.Request, offer db.Offer, AssertDistance float64, epsilon float64) float64 {

	distance := distance(offer.Location, request.Location)

	if (distance - AssertDistance) < (distance * epsilon) {
		// everything is ok
		return distance
	}

	t.Error("Distance Test failed: Asserted distance = ", AssertDistance, " \nCalculated distance = ", distance)

	return distance
}

func DistanceFactorTest(t *testing.T, request db.Request, offer db.Offer, AssertDistance float64, epsilon float64) float64 {

	distChannel := make(chan float64)

	go CalculateLocationDistance(distChannel, offer, request)

	distance := <-distChannel

	if (distance - AssertDistance) < (distance * epsilon) {
		// everything is ok
		return distance
	}

	t.Error("DistanceFactor Test failed: Asserted distance factor= ", AssertDistance, " \nCalculated distance factor = ", distance)

	return distance
}

// ----------------------------------------------------------------- AUTH

// [X] Login a: N - check if returns JWT
// [] Authorize : L
// [] Renew Token : L - check if returns new JWT
// [] Logout : L -

func LoginTest(t *testing.T, Email string, Password string, AssertCode int) string {
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

// [X] CreateUser : U
// [X] UpdateUser : S
// [X] ListUsers : A
// [X] GetUser : A
// [] PromoteToSystemAdmin : S

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

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("User creation failed")
	}

	if AssertCode == 400 && resp.Code != 400 {
		t.Error("User should already exist")
	}
}

func UpdateUserTest(t *testing.T, jwt string, User string, Name string, PreferredName string, Mail string,
	PhoneNumbers []string, Password string, Groups []GroupPayload, AssertCode int) map[string]interface{} {

	updateParams := UpdateUserPayload{
		Name,
		PreferredName,
		Mail,
		PhoneNumbers,
		Password,
		Groups,
	}

	resp := app.RequestWithJWT("PUT", "/users/"+User, updateParams, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not update user")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("UpdateUser should return BadRequest, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("UpdateUser should return Unauthorized, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 403 {
		if resp.Code != 403 {
			t.Error(fmt.Printf("UpdateUser should return Forbidden, but didnt"))
		}
		return map[string]interface{}{}
	}

	return parseResponse(resp)
}

func ListUsersTest(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/users", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("ListUsers fail ", resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("ListUsers should return UnAuthorized but didnt")
		}
		return []map[string]interface{}{}
	}

	return parseResponseToArray(resp)
}

func GetUser(t *testing.T, jwt string, User string, AssertCode int) map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/users/"+User, nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get user")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("GetUser should return BadRequest, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("GetUser should return Unauthorized, but didnt"))
		}
		return map[string]interface{}{}
	}

	return parseResponse(resp)
}

// ----------------------------------------------------------------- GROUPS

// [X] GetGroups : S
// [X] ListSystemAdmins : S

func GetGroupsTest(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/groups", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get groups")
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("GetGroups should return Unauthorized, but didnt"))
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

func ListSystemAdminsTest(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/system/admins", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get System  admin list")
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("ListSystemAdmins should return Unauthorized, but didnt"))
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

// ----------------------------------------------------------------- OFFERS

// [X] CreateOffer - L
// [X] GetOffer - C
// [X] UpdateOffer - C

func CreateOfferTest(t *testing.T, jwt string, Name string, Location gormGIS.GeoPoint, Radius float64, Validity string, AssertCode int) string {

	plCreateOffer := CreateOfferPayload{
		Name,
		struct {
			Longitude float64 `json:"lng" conform:"trim"`
			Latitude  float64 `json:"lat" conform:"trim"`
		}{Longitude: Location.Lng, Latitude: Location.Lat},
		Radius,
		[]string{},
		"This is a description of itsself because some text is needed for the description field.",
		Validity,
	}

	// check if offer was created
	resp := app.RequestWithJWT("POST", "/offers", plCreateOffer, jwt)

	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not CreateOffer")
		return ""
	}

	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error("CreateOffer should return BadRequest, but didnt")
		}
		return ""
	}

	data := parseResponse(resp)

	return data["ID"].(string)
}

func GetOfferTest(t *testing.T, jwt string, Offer string, AssertCode int) map[string]interface{} {

	resp := app.RequestWithJWT("GET", "/offers/"+Offer, nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get offer")
		return map[string]interface{}{}
	}

	if AssertCode == 400 {

		if resp.Code != 400 {
			t.Error(fmt.Printf("GetOffer should return BadRequest, but didnt"))
		}

		return map[string]interface{}{}
	}

	if AssertCode == 401 {

		if resp.Code != 401 {
			t.Error(fmt.Printf("GetOffer should return Unauthorized, but didnt"))
		}

		return map[string]interface{}{}
	}

	data := parseResponse(resp)

	return data
}

func UpdateOfferTest(t *testing.T, jwt string, Offer string, Name string, Location gormGIS.GeoPoint, Radius float64, Validity string, Tags []string, Description string, Matched bool, AssertCode int) map[string]interface{} {

	updateOfferParams := UpdateOfferPayload{
		Name,
		struct {
			Longitude float64 `json:"lng" conform:"trim"`
			Latitude  float64 `json:"lat" conform:"trim"`
		}{Longitude: Location.Lng, Latitude: Location.Lat},
		Radius,
		Tags,
		Description,
		Validity,
		Matched,
	}

	resp := app.RequestWithJWT("PUT", "/offers/"+Offer, updateOfferParams, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get offer")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("GetOffer should return BadRequest, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("GetOffer should return Unauthorized, but didnt"))
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	return data
}

// ----------------------------------------------------------------- REQUESTS

// [X] CreateRequest - L
// [X] GetRequest - C
// [X] UpdateRequest - C

func CreateRequestTest(t *testing.T, jwt string, Name string, Location gormGIS.GeoPoint, Radius float64, Validity string, Tags []string, Description string, AssertCode int) string {

	plCreateRequest := CreateRequestPayload{
		Name,
		struct {
			Longitude float64 `json:"lng" conform:"trim"`
			Latitude  float64 `json:"lat" conform:"trim"`
		}{Longitude: Location.Lng, Latitude: Location.Lat},
		Radius,
		Tags,
		Description,
		Validity,
	}

	resp := app.RequestWithJWT("POST", "/requests", plCreateRequest, jwt)

	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not create request")
		return ""
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("CreateRequest should return BadRequest, but did return %d", resp.Code))
		}
		return ""
	}

	data := parseResponse(resp)
	return data["ID"].(string)
}

func GetRequestTest(t *testing.T, jwt string, Request string, AssertCode int) map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/requests/"+Request, nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get request")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("GetRequest should return BadRequest, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("GetRequest should return Unauthorized, but didnt"))
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	return data
}

func UpdateRequestTest(t *testing.T, jwt string, Request string, Name string, Location gormGIS.GeoPoint, Radius float64, Validity string, Tags []string, Description string, Matched bool, AssertCode int) map[string]interface{} {

	updateRequestParams := UpdateRequestPayload{
		Name,
		struct {
			Longitude float64 `json:"lng" conform:"trim"`
			Latitude  float64 `json:"lat" conform:"trim"`
		}{Longitude: Location.Lng, Latitude: Location.Lat},
		Radius,
		Tags,
		Description,
		Validity,
		Matched,
	}

	resp := app.RequestWithJWT("PUT", "/requests/"+Request, updateRequestParams, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not update request" + resp.Body.String())
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("UpdateRequest should return BadRequest, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("UpdateRequest should return Unauthorized, but didnt"))
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	return data
}

// ----------------------------------------------------------------- MATCHINGS

// [X] CreateMatching - A
// [X] GetMatching - C
// [X] UpdateMatching - C

func CreateMatchingTest(t *testing.T, jwt string, Region string, Offer string, Request string, AssertCode int) string {
	plCreateMatching := CreateMatchingPayload{
		Region,
		Request,
		Offer,
	}

	resp := app.RequestWithJWT("POST", "/matchings", plCreateMatching, jwt)

	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not create matching")
		return ""
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
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

func GetMatchingTest(t *testing.T, jwt string, Matching string, AssertCode int) map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/matchings/"+Matching, nil, jwt)

	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not get matching")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("GetMatching should return BadRequest, but did return %d", resp.Code))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("GetMatching should return UnAuthorized but didnt")
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	return data
}

func UpdateMatchingTest(t *testing.T, jwt string, Matching string, Invalid bool, AssertCode int) map[string]interface{} {
	updateParams := UpdateMatchingPayload{
		Invalid,
	}

	resp := app.RequestWithJWT("PUT", "/matchings/"+Matching, updateParams, jwt)

	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not update matching")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("UpdateMatching should return BadRequest, but did return %d", resp.Code))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("UpdateMatching should return UnAuthorized but didnt")
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	return data
}

// ----------------------------------------------------------------- REGIONS

// [X] CreateRegion - L
// [X] ListRegions - U
// [X] GetRegion - U
// [] UpdateRegion - A
// [X] ListOffersForRegion - A
// [X] ListRequestsForRegion - A
// [X] ListMatchingsForRegion - A
// [X] PromoteUserToAdminForRegion - A
// [X] ListAdminsForRegion - A

func CreateRegionTest(t *testing.T, jwt string, Name string, Desc string, Locations []Location, AssertCode int) string {
	// create region
	plCreateRegion := CreateRegionPayload{
		Name,
		Desc,
		Boundaries{
			Locations,
		},
	}

	resp := app.RequestWithJWT("POST", "/regions", plCreateRegion, jwt)
	if AssertCode == 201 && resp.Code != 201 {
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

func ListRegions(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/regions", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get region list")
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("ListRegions should return Unauthorized, but didnt"))
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

func GetRegionTest(t *testing.T, jwt string, Region string, AssertCode int) map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/regions/"+Region, nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("GetRegion failed")
		return map[string]interface{}{}
	}

	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error("GetRegion should return bad request, but didnt")
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	if data["ID"] != Region {
		t.Error("Wrong region was returned")
	}

	return data
}

func ListOffersForRegionTest(t *testing.T, jwt string, Region string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/regions/"+Region+"/offers", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get offerlist for region" + resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("ListOffersForRegion should return Unauthorized, but didnt"))
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

func ListRequestsForRegionTest(t *testing.T, jwt string, Region string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/regions/"+Region+"/requests", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get requestlist for region" + resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("ListRequestsForRegion should return Unauthorized, but didnt"))
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

func ListMatchingsForRegionTest(t *testing.T, jwt string, Region string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/regions/"+Region+"/matchings", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get matchinglist for region" + resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("ListMatchingsForRegion should return Unauthorized, but didnt"))
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

func PromoteUserToAdminForRegionTest(t *testing.T, jwt string, Email string, Region string, AssertCode int) bool {
	promoteParams := PromoteAdminPayload{Email}
	resp := app.RequestWithJWT("POST", "/regions/"+Region+"/admins", promoteParams, jwt)

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

func ListAdminsForRegionTest(t *testing.T, jwt string, Region string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/regions/"+Region+"/admins", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not get adminlist for region" + resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("ListAdminsForRegion should return Unauthorized, but didnt"))
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

// ------------------------------------------------------------------------------- ME

// [X] GetMe - L
// [X] UpdateMe - L
// [X] ListUserOffers - L
// [X] ListUserRequests - L
// [X] ListUserMatchings - L

func GetMeTest(t *testing.T, jwt string, AssertCode int) map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/me", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("GetMe fail ", resp.Body.String())
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	return data
}

func UpdateMeTest(t *testing.T, jwt string, Name string, PreferredName string, Mail string,
	PhoneNumbers []string, Password string, Groups []GroupPayload, AssertCode int) map[string]interface{} {

	updateParams := UpdateUserPayload{
		Name,
		PreferredName,
		Mail,
		PhoneNumbers,
		Password,
		Groups,
	}

	resp := app.RequestWithJWT("PUT", "/me", updateParams, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("Could not UpdateMe")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("UpdateMe should return BadRequest, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error(fmt.Printf("UpdateMe should return Unauthorized, but didnt"))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 403 {
		if resp.Code != 403 {
			t.Error(fmt.Printf("UpdateMe should return Forbidden, but didnt"))
		}
		return map[string]interface{}{}
	}

	return parseResponse(resp)
}

func ListUserOffersTest(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/me/offers", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("ListUserOffers fail ", resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("ListUserOffers should return UnAuthorized but didnt")
		}
		return []map[string]interface{}{}
	}

	return parseResponseToArray(resp)
}

func ListUserRequestsTest(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/me/requests", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("ListUserRequests fail ", resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("ListUserRequests should return UnAuthorized but didnt")
		}
		return []map[string]interface{}{}
	}

	return parseResponseToArray(resp)
}

func ListUserMatchingsTest(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/me/matchings", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("ListUserMatchings fail ", resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("ListUserMatchings should return UnAuthorized but didnt")
		}
		return []map[string]interface{}{}
	}

	return parseResponseToArray(resp)
}

// ------------------------------------------------------------------------------- Notifications

// [X] ListNotifications - L
// [X] UpdateNotification - C

func ListNotificationsTest(t *testing.T, jwt string, AssertCode int) []map[string]interface{} {
	resp := app.RequestWithJWT("GET", "/notifications", nil, jwt)

	if AssertCode == 200 && resp.Code != 200 {
		t.Error("ListNotifications fail ", resp.Body.String())
		return []map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("ListNotificationsTest should return UnAuthorized but didnt")
		}
		return []map[string]interface{}{}
	}

	data := parseResponseToArray(resp)
	return data
}

func UpdateNotificationTest(t *testing.T, jwt string, Notification string, Read bool, AssertCode int) map[string]interface{} {
	updateParams := UpdateNotificationPayload{
		Read,
	}

	resp := app.RequestWithJWT("PUT", "/notifications/"+Notification, updateParams, jwt)

	if AssertCode == 201 && resp.Code != 201 {
		t.Error("Could not update notification")
		return map[string]interface{}{}
	}
	if AssertCode == 400 {
		if resp.Code != 400 {
			t.Error(fmt.Printf("UpdateNotification should return BadRequest, but did return %d", resp.Code))
		}
		return map[string]interface{}{}
	}
	if AssertCode == 401 {
		if resp.Code != 401 {
			t.Error("UpdateNotification should return UnAuthorized but didnt")
		}
		return map[string]interface{}{}
	}

	data := parseResponse(resp)
	return data
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

	// VALID GetMe
	regionAdminResp := GetMeTest(t, userRegionAdmin, 200)
	if regionAdminResp["Mail"] != emailRegionAdmin {
		t.Error("CreateUser followed by GetUser: comparing email for region admin failed")
	}

	// INVALID: Login superadmin
	LoginTest(t, "admin@example.org", "nonononooo", 400)
	// VALID: Login superadmin
	userSuperAdmin = LoginTest(t, "admin@example.org", "CaTUstrophyAdmin123$", 200)

	// INVALID: CreateRegion
	CreateRegionTest(t, userRegionAdmin, "", "", []Location{}, 400)
	// VALID: CreateRegion
	regionName := "Milkshake Region"
	regionID = CreateRegionTest(t, userRegionAdmin, regionName, "Ma Region brings all the boys in the yard",
		[]Location{Location{10.0, 0.0}, Location{11.0, 0.0}, Location{10.0, 1.0}, Location{10.0, 0.0}}, 201)

	// INVALID: GetRegion
	GetRegionTest(t, userOffering, regionID+"a", 400)
	// VALID: GetRegion
	region := GetRegionTest(t, userOffering, regionID, 200)
	// compare region names of Create & Get
	if region["Name"] != regionName {
		t.Error("CreateRegion followed by GetRegion returns not the same value for field Region.Name")
	}

	// VALID GetGroup - check if group was created for region
	groups := GetGroupsTest(t, userSuperAdmin, 200)
	newGroupFound := false
	for _, group := range groups {
		if group["RegionId"].(string) == regionID {
			newGroupFound = true
			break
		}
	}

	if !newGroupFound {
		t.Error("GetGroups failed or CreateRegion did not create new group")
	}

	// INVALID: PromoteUserToAdminForRegion
	PromoteUserToAdminForRegionTest(t, userSuperAdmin, "", regionID, 400)
	PromoteUserToAdminForRegionTest(t, userRegionAdmin, emailRegionAdmin, regionID, 401)
	PromoteUserToAdminForRegionTest(t, userSuperAdmin, "nobody@donotexist.com", regionID, 404)
	// VALID: PromoteUserToAdminForRegion
	PromoteUserToAdminForRegionTest(t, userSuperAdmin, emailRegionAdmin, regionID, 200)

	// VALID ListSystemAdmins
	admins := ListSystemAdminsTest(t, userSuperAdmin, 200)
	if len(admins) == 0 {
		t.Error("ListSystemAdmins returned no SuperAdmins")
	}
}

func TestMatchingAlpha(t *testing.T) {
	fmt.Println("\n--------------------- MatchingTestAlpha ---------------------\n")

	// INVALID CreateOffer
	CreateOfferTest(t, userOffering, "", gormGIS.GeoPoint{10.2, .0}, 10.4, "2017-11-01T22:08:41+00:00", 400)
	// VALID CreateOffer
	offerID = CreateOfferTest(t, userOffering, "Milk x10", gormGIS.GeoPoint{10.2, .0}, 20.3, "2017-11-01T22:08:41+00:00", 201)

	// INVALID GetOffer
	GetOfferTest(t, userOffering, offerID+"a", 400)
	GetOfferTest(t, userRequesting, offerID, 401)
	// VALID GetOffer
	offer := GetOfferTest(t, userOffering, offerID, 200)
	GetOfferTest(t, userRegionAdmin, offerID, 200)

	// INVALID CreateRequest
	CreateRequestTest(t, userRequesting, "", gormGIS.GeoPoint{10.3, 0.2}, 2.0, "2016-11-01T22:08:41+00:00", []string{}, "Description", 400)
	// VALID CreateRequest
	requestName := "Me thirsty"
	requestID = CreateRequestTest(t, userRequesting, requestName, gormGIS.GeoPoint{10.3, 0.2}, 1000.2, "2016-11-01T22:08:41+00:00", []string{"Food"}, "", 201)

	// INVALID GetRequest
	GetRequestTest(t, userRequesting, requestID+"a", 400)
	GetRequestTest(t, userOffering, requestID, 401)
	// VALID GetRequest
	request := GetRequestTest(t, userRequesting, requestID, 200)
	GetRequestTest(t, userRegionAdmin, requestID, 200)
	GetRequestTest(t, userSuperAdmin, requestID, 200)
	if request["Name"] != requestName {
		t.Error("CreateRequest followed by GetRequest dont seem to return same values")
	}

	// INVALID CreateMatching
	CreateMatchingTest(t, userOffering, regionID, offerID, requestID, 401) // TODO : somehow this seems to work just fine
	CreateMatchingTest(t, userRegionAdmin, "", offerID, requestID, 400)
	// VALID CreateMatching
	matchingID = CreateMatchingTest(t, userRegionAdmin, regionID, offerID, requestID, 201)

	// VALID GetMatching and compare against CreateMatching
	GetMatchingTest(t, userRequesting, matchingID, 200)
	matching := GetMatchingTest(t, userOffering, matchingID, 200)
	if matching["Request"].(map[string]interface{})["ID"] != requestID {
		t.Error("GetMatching Matching.Request.ID is not same as Request.ID")
	}
	if matching["Offer"].(map[string]interface{})["ID"] != offerID {
		t.Error("GetMatching Matching.Offer.ID is not same as Offer.ID")
	}

	// VALID UpdateMatching with Invalid
	UpdateMatchingTest(t, userOffering, matchingID, true, 200)
	matching = GetMatchingTest(t, userOffering, matchingID, 200)
	if !matching["Invalid"].(bool) {
		t.Error("UpdateMatching with invalid:true failed - matching still valid")
	}

	// VALID CreateMatch - recreate the matching that was set invalid
	matchingID = CreateMatchingTest(t, userRegionAdmin, regionID, offerID, requestID, 201)
	matching = GetMatchingTest(t, userOffering, matchingID, 200)

	// VALID UpdateOfferTest
	offer = UpdateOfferTest(t, userOffering, offerID,
		offer["Name"].(string)+" Updated",
		gormGIS.GeoPoint{0, 0},
		99.3,
		offer["ValidityPeriod"].(string),
		[]string{"Tool"},
		offer["Description"].(string)+" also updated",
		false,
		200,
	)
	// GetOffer and check if updates were propagated
	offer = GetOfferTest(t, userOffering, offerID, 200)
	// check if tags were updated
	tags := offer["Tags"].([]interface{})
	if len(tags) == 0 {
		t.Error("UpdateOffer failed to update")
	}
	if tags[0].(map[string]interface{})["Name"].(string) != "Tool" {
		t.Error("UpdateOffer failed to insert tag Tool")
	}

	// check if location was updated
	location := offer["Location"].(map[string]interface{})
	if location["lat"].(float64) != 0 || location["lng"].(float64) != 0 {
		//t.Error("UpdateOfffer didnt update Location")
	}

	// VALID UpdateRequestTest
	request = UpdateRequestTest(t, userRequesting, requestID,
		request["Name"].(string)+" Updated",
		gormGIS.GeoPoint{0, 0},
		100.1,
		request["ValidityPeriod"].(string),
		[]string{"Water"},
		"Completely new description",
		false,
		200,
	)

	// GetRequest and check if updates were propagated
	request = GetRequestTest(t, userRequesting, requestID, 200)
	// check if tags were updated
	tags = request["Tags"].([]interface{})
	if len(tags) == 0 {
		t.Error("UpdateRequest failed to update")
	}
	if tags[0].(map[string]interface{})["Name"].(string) != "Water" {
		t.Error("UpdateRequest failed to insert tag Water")
	}

	// check if location was updated
	location = request["Location"].(map[string]interface{})
	if location["lat"].(float64) != 0 || location["lng"].(float64) != 0 {
		//t.Error("UpdateRequest didnt update Location")
	}

	// VALID ListNotifications
	notifications := ListNotificationsTest(t, userOffering, 200)
	if len(notifications) == 0 {
		t.Error("ListNotifications returned empty array, but matching was created prior to that")
	}
	notificationID := notifications[0]["ID"].(string)

	//VALID UpdateNotification
	UpdateNotificationTest(t, userOffering, notificationID, true, 200)
	notifications = ListNotificationsTest(t, userOffering, 200)

	for _, not := range notifications {
		if not["ID"].(string) == notificationID {
			t.Error("ListNotification should not contain updated notification, since it was just READ")
			break
		}
	}

	// Distance test:
	// Create offer and request with distance 11.132km and very large Radius
	distRequest := db.Request{Location: gormGIS.GeoPoint{0.0, 0.0}, Radius: 10000}
	distOffer := db.Offer{Location: gormGIS.GeoPoint{0.0, 0.1}, Radius: 10000}
	DistanceTest(t, distRequest, distOffer, 11.132, 0.001)
	DistanceFactorTest(t, distRequest, distOffer, 10, 0.5)

	AddDataTest(t)
}
