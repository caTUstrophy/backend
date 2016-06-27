package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
)

var ReplacementsJSON = map[string]interface{}{
	"time.Time":           "RFC3339 date",
	"db.PhoneNumbers":     "[string, ...]",
	"db.NotificationType": "string",
}

var ReplacementsJSONbyKey = map[string]interface{}{
	"ID": "UUID v4",
}

// This function generates documentation like we present in README
// for a hard coded list of data we send in responses. Sends back
// the generated stuff via HTTP and also prints it to the terminal
// with additional info. Last is used to generate text we can copy
// directly into README that automates the documentation of our
// frequently updated API.
func (app *App) GetJsonResponseInfo(c *gin.Context) {

	// Check authorization for this function.

	// TODO: decomment when authorization is pulled and works
	ok, authUser, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}
	// Check if user permissions are sufficient (user is admin).
	if ok = app.CheckScope(authUser, db.Region{}, "superadmin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	var User db.User
	User.Enabled = false
	app.DB.First(&User.Groups, "region_id <> null")

	for _, group := range User.Groups {
		app.DB.Model(&group).Related(&group.Region)
	}

	User.ID = "Wer als erstes diese Zeile liest und mich drauf anspricht kriegt eine Mate ausgegeben"
	User.Mail = "m@i.le"
	User.MailVerified = false
	User.PasswordHash = "Gar kein echtes Hash"
	User.PreferredName = "Gott"
	User.Name = "Mensch"

	// Generate response maps
	var currResponseMap map[string]interface{}
	allResponses := make(map[string]interface{})

	// USER
	currResponseMap = getJSONResponseInfo(User, fieldsUser)
	allResponses["User"] = currResponseMap

	// USERS LIST
	var users [1]map[string]interface{}
	users[0] = allResponses["User"].(map[string]interface{})
	allResponses["Users"] = users

	// USER WITHOUT GROUP DETAILS
	currResponseMap = getJSONResponseInfo(User, fieldsUserNoGroups)
	allResponses["User without groups"] = currResponseMap

	// LIST OF USERs WITHOUT GROUP DETAILS
	var usersNoGroup [1]map[string]interface{}
	usersNoGroup[0] = allResponses["User without groups"].(map[string]interface{})
	allResponses["List of users without group"] = usersNoGroup

	// REGION
	var region db.Region
	app.DB.First(&region)
	currResponseMap = getJSONResponseInfo(region, fieldsRegion)
	allResponses["Region"] = currResponseMap

	// REGIONS LIST
	var regions [1]map[string]interface{}
	regions[0] = allResponses["Region"].(map[string]interface{})
	allResponses["Regions"] = regions

	// OFFER
	var offer db.Offer
	app.DB.Preload("Regions").Preload("Tags").First(&offer)
	app.DB.Model(&offer).Related(&offer.User)
	currResponseMap = getJSONResponseInfo(offer, fieldsOffer)
	allResponses["Offer"] = currResponseMap

	// OFFERS LIST
	var offers [1]map[string]interface{}
	offers[0] = allResponses["Offer"].(map[string]interface{})
	allResponses["Offers"] = offers

	// REQUEST
	var request db.Request
	app.DB.Preload("Regions").Preload("Tags").First(&request)
	app.DB.Model(&request).Related(&request.User)
	currResponseMap = getJSONResponseInfo(request, fieldsRequest)
	allResponses["Request"] = currResponseMap

	// REQUESTS LIST
	var requests [1]map[string]interface{}
	requests[0] = allResponses["Request"].(map[string]interface{})
	allResponses["Requests"] = requests

	// MATCHING
	var matching db.Matching
	app.DB.First(&matching)
	currResponseMap = getJSONResponseInfo(matching, fieldsMatching)
	allResponses["Matching"] = currResponseMap

	// MATCHING LIST
	var matchings [1]map[string]interface{}
	matchings[0] = allResponses["Matching"].(map[string]interface{})
	allResponses["Matchings"] = matchings

	// NOTIFICATION
	var notification db.Notification
	app.DB.First(&notification)
	currResponseMap = getJSONResponseInfo(notification, fieldsNotification)
	allResponses["Notification"] = currResponseMap

	// NOTIFICATIONS LIST
	var notifications [1]map[string]interface{}
	notifications[0] = allResponses["Notification"].(map[string]interface{})
	allResponses["Notifications"] = notifications

	// generate text from that map that can be copied to README
	// Open file for writing footer
	f, err := os.Create("README_footer.md")
	if err != nil {
		log.Println("Unable to open file to write. Please check if backend has the permission to write on your server.")
		// Send as http response
		c.JSON(http.StatusOK, allResponses)
	}

	// Write footer
	f.WriteString("\n### Responses\n")

	writeFooterSection(f, "\n#### Single user complete\n", allResponses["User"])

	writeFooterSection(f, "\n#### List users complete\n", allResponses["Users"])

	writeFooterSection(f, "\n#### User without group\n", allResponses["User without groups"])

	writeFooterSection(f, "\n#### List of users without group\n", allResponses["List of users without group"])

	writeFooterSection(f, "\n#### Offer object\n", allResponses["Offer"])

	writeFooterSection(f, "\n#### Offer list\n", allResponses["Offers"])

	writeFooterSection(f, "\n#### Request object\n", allResponses["Request"])

	writeFooterSection(f, "\n#### Request list\n", allResponses["Requests"])

	writeFooterSection(f, "\n#### Matching object\n", allResponses["Matching"])

	writeFooterSection(f, "\n#### Matching list\n", allResponses["Matchings"])

	writeFooterSection(f, "\n#### Region object\n", allResponses["Region"])

	writeFooterSection(f, "\n#### Region list\n", allResponses["Regions"])

	writeFooterSection(f, "\n#### Notification object\n", allResponses["Notification"])

	writeFooterSection(f, "\n#### Notification list\n", allResponses["Notifications"])

	// Send as http response
	c.JSON(http.StatusOK, allResponses)
}

func writeFooterSection(f *os.File, title string, content interface{}) {
	f.WriteString(title)

	var out bytes.Buffer
	jsonText, _ := json.Marshal(content)
	json.Indent(&out, jsonText, "", "\t")
	out.WriteTo(f)
}
