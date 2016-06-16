package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
)

var fieldsGetOffersForRegion = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Location": map[string]interface{}{
		"Lng": "lng",
		"Lat": "lat",
	},
	"Tags": map[string]interface{}{
		"Name": "Name",
	},
	"ValidityPeriod": "ValidityPeriod",
	"Matched":        "Matched",
	"Expired":        "Expired",
}

var fieldsGetRequestsForRegion = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Location": map[string]interface{}{
		"Lng": "lng",
		"Lat": "lat",
	},
	"Tags": map[string]interface{}{
		"Name": "Name",
	},
	"ValidityPeriod": "ValidityPeriod",
	"Matched":        "Matched",
	"Expired":        "Expired",
}

type CreateRegionPayload struct {
	Name        string             `conform:"trim" validate:"required"`
	Description string             `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Boundaries  []gormGIS.GeoPoint `conform:"trim"`
}

func (app *App) CreateRegion(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Payload CreateRegionPayload

	// Expect offer struct fields for creation in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		// Check if error was caused by failed unmarshalling string -> []string.
		//if err.Error() == "json: cannot unmarshal string into Go value of type []string" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Tags": "Provide an array, not a string",
		})

		return
		//}
	}

	// Validate sent offer creation data.
	conform.Strings(&Payload)
	errs := app.Validator.Struct(&Payload)

	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp[err.Field] = "Is required"
			} else if err.Tag == "excludesall" {
				errResp[err.Field] = "Contains unallowed characters"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}
}

func (app *App) ListRegions(c *gin.Context) {

	Regions := []db.Region{}

	// Retrieve all offers from database.
	app.DB.Find(&Regions)

	c.JSON(http.StatusOK, Regions)
}

func (app *App) GetRegion(c *gin.Context) {

	// Retrieve region ID from request URL.
	regionID := app.getUUID(c, "regionID")
	if regionID == "" {
		return
	}

	var Region db.Region

	// Select region based on supplied ID from database.
	app.DB.First(&Region, "id = ?", regionID)
	app.DB.Model(&Region).Association("Boundaries").Find(&Region.Boundaries)

	c.JSON(http.StatusOK, Region)
}

func (app *App) UpdateRegion(c *gin.Context) {

}

func (app *App) GetOffersForRegion(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	regionID := app.getUUID(c, "regionID")
	if regionID == "" {
		return
	}

	var Region db.Region
	app.DB.Preload("Offers").First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}
	model := make([]map[string]interface{}, len(Region.Offers))
	for _, offer := range Region.Offers {
		model = append(model, CopyNestedModel(offer, fieldsGetOffersForRegion))
	}
	// Send back results to client.
	c.JSON(http.StatusOK, model)
}

func (app *App) GetRequestsForRegion(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	regionID := app.getUUID(c, "regionID")
	if regionID == "" {
		return
	}

	var Region db.Region
	app.DB.Preload("Requests").First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}
	model := make([]map[string]interface{}, len(Region.Requests))
	for _, offer := range Region.Requests {
		model = append(model, CopyNestedModel(offer, fieldsGetRequestsForRegion))
	}
	// Send back results to client.
	c.JSON(http.StatusOK, model)
}

func (app *App) GetMatchingsForRegion(c *gin.Context) {

}
