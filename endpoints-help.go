package main

import (
	"fmt"
	"log"
	"reflect"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
)

var ReplacementsJSON = map[string]interface{}{
	"time.Time":       "RFC3339 date",
	"db.PhoneNumbers": "[string, ...]",
}

var ReplacementsJSONbyKey = map[string]interface{}{
	"ID": "UUID v4",
}

// This function generates for a hard coded list of data we send in responses a documentation like we present in README
// Sends back the generated stuff via http and also prints it to the terminal with additional Info
// Last is used to generate text we can copy directly into README that automizes the documentation of our regulary updated API
func (app *App) GetJsonResponseInfo(c *gin.Context) {
	// Check authorization for this function.

	// TODO: decomment when authorization is pulled and works
	/*ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}*/
	// Check if user permissions are sufficient (user is admin).
	// if ok = app.CheckScope(User, db.Region{}, "superadmin"); !ok {
	if false {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: remove hacky User load

	var User db.User
	User.Enabled = false
	app.DB.Preload("Permissions").First(&User.Groups)
	for _, group := range User.Groups {
		app.DB.Model(&group).Related(&group.Region)
	}
	User.ID = "Wer als erstes diese Zeile liest und mich drauf anspricht kriegt eine Mate ausgegeben"
	User.Mail = "m@i.le"
	User.MailVerified = false
	User.PasswordHash = "Gar kein echtes Hash"
	User.PreferredName = "Gott"
	User.Name = "Mensch"

	// Generate Response Maps
	var currResponseMap map[string]interface{}
	allResponses := make(map[string]interface{})

	// USER
	currResponseMap = getJSONResponseInfo(User, fieldsUser)
	allResponses["User"] = currResponseMap

	// USER LIST
	var users [1]map[string]interface{}
	users[0] = allResponses["User"].(map[string]interface{})
	allResponses["User List"] = users

	// REGION
	var region db.Region
	app.DB.First(&region)
	currResponseMap = getJSONResponseInfo(region, fieldsRegion)
	allResponses["Region"] = currResponseMap

	// REGION LIST
	var regions [1]map[string]interface{}
	regions[0] = allResponses["Region"].(map[string]interface{})
	allResponses["Region List"] = regions

	// OFFER
	var offer db.Offer
	app.DB.Preload("Regions").Preload("Tags").First(&offer)
	app.DB.Model(&offer).Related(&offer.User)
	currResponseMap = getJSONResponseInfo(offer, fieldsOffer)
	allResponses["Offer"] = currResponseMap

	// OFFER LIST
	var offers [1]map[string]interface{}
	offers[0] = allResponses["Offer"].(map[string]interface{})
	allResponses["Offer List"] = offers

	// REQUEST
	var request db.Request
	app.DB.Preload("Regions").Preload("Tags").First(&request)
	app.DB.Model(&request).Related(&request.User)
	currResponseMap = getJSONResponseInfo(request, fieldsRequest)
	allResponses["Request"] = currResponseMap

	// REQUEST LIST
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

	// generate text from that map that can be copied to README
	// TODO because no internet in this bus to look up stuff

	// Send as http response
	c.JSON(http.StatusOK, allResponses)

}

// This function works just like CopyNestedModel, but returns no content of i but the type of the content.
// Will be used to automize writing our Documentation
func getJSONResponseInfo(i interface{}, fields map[string]interface{}) map[string]interface{} {

	var m map[string]interface{}
	m = make(map[string]interface{})

	// get value + type of source interface
	valInterface := reflect.ValueOf(i)
	typeOfT := reflect.ValueOf(i).Type()

	// iterate over all fields that occur in response
	for key := range fields {

		var exists = false
		newKey, _ := fields[key].(string)

		// search for field in source type
		for i := 0; i < valInterface.NumField(); i++ {

			if typeOfT.Field(i).Name == key { // original field has been found

				// check for nesting through type assertion
				nestedMap, nested := fields[key].(map[string]interface{})

				if !nested {
					// NOT nested -> save type of this field for the newKey
					if ReplacementsJSONbyKey[newKey] != nil {
						m[newKey] = ReplacementsJSONbyKey[newKey]
					} else {
						newType := fmt.Sprint(valInterface.Field(i).Type())
						if ReplacementsJSON[newType] != nil {
							m[newKey] = ReplacementsJSON[newType]
						} else {
							m[newKey] = newType
						}
					}

				} else {

					// NESTED copied via recursion
					var slice = reflect.ValueOf(valInterface.Field(i).Interface())

					// if nested ARRAY
					if valInterface.Field(i).Kind() == reflect.Slice {

						sliceMapped := make([]interface{}, 1)

						for i := 0; i < slice.Len() && i < 1; i++ {
							sliceMapped[i] = getJSONResponseInfo(slice.Index(i).Interface(), nestedMap)
						}

						m[key] = sliceMapped
					} else {
						// if nested OBJECT
						m[key] = getJSONResponseInfo(valInterface.Field(i).Interface(), nestedMap)
					}
				}

				exists = true
				break
			}
		}

		if !exists {
			log.Fatalf("ERROR: Struct<%s> has no field: %s", typeOfT.Name(), key)
		}
	}

	return m
}
