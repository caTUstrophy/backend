package main

import (
	"fmt"
	"log"

	"bytes"
	"encoding/json"
	"net/http"
	"reflect"

	"net/http/httptest"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/nferruzzi/gormGIS"
)

// structs

type GeoLocation struct {
	Longitude float64 `json:"lng" conform:"trim"`
	Latitude  float64 `json:"lat" conform:"trim"`
}

// Checks if a generic URL substring is present in the
// current request and, if so, attempts to validate it
// as an UUID version 4.
func (app *App) getUUID(c *gin.Context, par string) string {

	parID := c.Params.ByName(par)
	errs := app.Validator.Field(parID, "required,uuid4")

	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp[par] = "Is required"
			} else if err.Tag == "uuid4" {
				errResp[par] = "Needs to be an UUID version 4"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return ""
	}

	return parID
}

// Intersects the location of the provided item (offer or request)
// with all regions available to determine to which region this
// item should be mapped.
func (app *App) mapLocationToRegions(item interface{}) {

	var itemType string
	var location gormGIS.GeoPoint
	var ContRegionID string
	var ContRegionIDs []string

	// Determine whether we received an offer or a request.
	switch item.(type) {
	case db.Offer:

		itemType = "Offer"

		asssertedItem, ok := item.(db.Offer)
		if !ok {
			log.Fatal("[mapLocationToRegions] Type assertion to db.Offer was unsuccessful. Returning from function.")
			return
		}

		location = asssertedItem.Location
	case db.Request:

		itemType = "Request"

		asssertedItem, ok := item.(db.Request)
		if !ok {
			log.Fatal("[mapLocationToRegions] Type assertion to db.Request was unsuccessful. Returning from function.")
			return
		}

		location = asssertedItem.Location
	default:
		itemType = "UNKNOWN"
		log.Println("[mapLocationToRegions] itemType was UNKNOWN")
		return
	}

	// Find all IDs of regions with which the supplied point intersects.
	regionRows, err := app.DB.Raw("SELECT \"id\" FROM \"regions\" WHERE ST_INTERSECTS(ST_GeographyFromText(?), \"regions\".\"boundaries\")", location.String()).Rows()
	if err != nil {
		log.Fatal(err)
	}

	// Close row connection on function exit.
	defer regionRows.Close()

	// Iterate over all found regions and save regionID to slice.
	for regionRows.Next() {
		regionRows.Scan(&ContRegionID)
		ContRegionIDs = append(ContRegionIDs, ContRegionID)
	}

	if len(ContRegionIDs) > 0 {

		var ContRegions []db.Region

		// Retrieve all regions into above list of containing regions.
		// Only regions with IDs from intersecting region list will be chosen.
		app.DB.Where("id in (?)", ContRegionIDs).Preload("Offers").Preload("Requests").Find(&ContRegions)

		for _, ContRegion := range ContRegions {

			// Depending on type of item, save an offer or a request into list.
			if itemType == "Offer" {
				ContRegion.Offers = append(ContRegion.Offers, item.(db.Offer))
			} else if itemType == "Request" {
				ContRegion.Requests = append(ContRegion.Requests, item.(db.Request))
			}

			// Save changed offers or requests of a region to database.
			app.DB.Save(&ContRegion)
		}
	} else {
		log.Println("[mapLocationToRegions] No intersecting regions found.")
	}
}

// Takes in any data type as i and a fields map containing
// the fields in i that are supposed to be passed to output
// data. Allows for selection and renaming of e.g. fields in
// structs before returning them as JSON.
func CopyNestedModel(i interface{}, fields map[string]interface{}) interface{} {

	var m map[string]interface{}
	m = make(map[string]interface{})

	// get value + type of source interface
	valInterface := reflect.ValueOf(i)
	typeOfT := reflect.ValueOf(i).Type()

	var slice = reflect.ValueOf(valInterface.Interface())
	if valInterface.Kind() == reflect.Slice {

		sliceMapped := make([]interface{}, slice.Len())

		for i := 0; i < slice.Len(); i++ {
			sliceMapped[i] = CopyNestedModel(slice.Index(i).Interface(), fields)
		}

		return sliceMapped
	}

	// iterate over all fields that will be copied
	for key := range fields {

		var exists = false
		newKey, _ := fields[key].(string)

		if valInterface.NumField() == 0 {
			// This nested data is expected by fields but is empty.
			// We return the data with every field of the not existing nested map set to nil.
			m[newKey] = nil
			exists = true // We dont throw an error in this case
		} else {

			// search for field in source type
			for i := 0; i < valInterface.NumField(); i++ {

				if typeOfT.Field(i).Name == key {

					// check for nesting through type assertion
					nestedMap, nested := fields[key].(map[string]interface{})

					if !nested {
						// NOT nested -> copy value directly
						m[newKey] = valInterface.Field(i).Interface()
					} else {

						// NESTED copied via recursion
						slice = reflect.ValueOf(valInterface.Field(i).Interface())

						// if nested ARRAY
						if valInterface.Field(i).Kind() == reflect.Slice {

							sliceMapped := make([]interface{}, slice.Len())

							for i := 0; i < slice.Len(); i++ {
								sliceMapped[i] = CopyNestedModel(slice.Index(i).Interface(), nestedMap)
							}

							m[key] = sliceMapped
						} else {
							// if nested OBJECT
							m[key] = CopyNestedModel(valInterface.Field(i).Interface(), nestedMap)
						}
					}

					exists = true
					break
				}
			}
		}

		if !exists {
			log.Fatalf("ERROR: Struct<%s> has no field: %s", typeOfT.Name(), key)
		}
	}

	return m
}

// This function works just like CopyNestedModel, but returns
// no content of i but the type of the content. Will be used
// to automate writing our documentation.
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

// USED FOR TESTING ONLY!
// Creates http.Request with authentication, requests url and returns a response
func (app *App) RequestWithJWT(method string, url string, body interface{}, jwt string) *httptest.ResponseRecorder {

	resp := httptest.NewRecorder()
	req := NewRequestWithJWT(method, url, body, jwt)
	app.Router.ServeHTTP(resp, req)

	return resp
}

// USED FOR TESTING ONLY!
// Creates http.Request, requests url and returns a response
func (app *App) Request(method string, url string, body interface{}) *httptest.ResponseRecorder {
	return app.RequestWithJWT(method, url, body, "")
}

// creates and configures http.Request
func NewRequest(method string, url string, body interface{}) *http.Request {
	return NewRequestWithJWT(method, url, body, "")
}

// creates and configures authorized http.Request
func NewRequestWithJWT(method string, url string, body interface{}, jwt string) *http.Request {

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))

	if jwt != "" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", ("Bearer " + jwt))
	}

	return req
}

// USED FOR TESTING ONLY!
// Parse response into interface{}, instead of JSON.bind() onto struct.
func parseResponse(resp *httptest.ResponseRecorder) map[string]interface{} {

	var dat map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &dat); err != nil {
		panic(err)
	}

	return dat
}

// USED FOR TESTING ONLY!
// Parse response into array of interface{}, instead of JSON.bind() onto struct.
func parseResponseToArray(resp *httptest.ResponseRecorder) []map[string]interface{} {

	var dat []map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &dat); err != nil {
		panic(err)
	}

	return dat
}
