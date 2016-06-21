package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
)

// Structs

type CreateLocation struct {
	Lng float64 `json:"lng" conform:"trim"`
	Lat float64 `json:"lat" conform:"trim"`
}

type CreateRegionPayload struct {
	Name        string           `conform:"trim" validate:"required"`
	Description string           `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Boundaries  []CreateLocation `validate:"dive,required"`
}

// Functions

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

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Couldn't marshal JSON",
		})

		return
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

	// Save Region
	var Region db.Region
	Region.ID = fmt.Sprintf("%s", uuid.NewV4())
	Region.Name = Payload.Name
	Region.Description = Payload.Description

	Points := make([]gormGIS.GeoPoint, len(Payload.Boundaries))

	for i, point := range Payload.Boundaries {
		Points[i] = gormGIS.GeoPoint{Lng: point.Lng, Lat: point.Lat}
	}

	Region.Boundaries = db.GeoPolygon{
		Points: Points,
	}

	app.DB.Create(&Region)

	model := CopyNestedModel(Region, fieldsRegion)

	c.JSON(http.StatusCreated, model)
}

func (app *App) ListRegions(c *gin.Context) {

	Regions := []db.Region{}

	// Retrieve all offers from database.
	app.DB.Find(&Regions)

	models := make([]map[string]interface{}, len(Regions))

	// Iterate over all regions in database return and marshal it.
	for i, region := range Regions {
		models[i] = CopyNestedModel(region, fieldsRegion)
	}

	c.JSON(http.StatusOK, models)
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

	// Only expose necessary fields in JSON response.
	model := CopyNestedModel(Region, fieldsRegion)

	c.JSON(http.StatusOK, model)
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

	for i, offer := range Region.Offers {
		model[i] = CopyNestedModel(offer, fieldsOffer)
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

	for i, offer := range Region.Requests {
		model[i] = CopyNestedModel(offer, fieldsRequest)
	}

	// Send back results to client.
	c.JSON(http.StatusOK, model)
}

func (app *App) GetMatchingsForRegion(c *gin.Context) {

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

	var region db.Region
	app.DB.First(&region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	var matchings []db.Matching
	app.DB.Model(&region).Related(&matchings)

	model := make([]map[string]interface{}, len(matchings))

	for i, matching := range matchings {
		model[i] = CopyNestedModel(matching, fieldsMatching)
	}

	c.JSON(http.StatusOK, model)
}
