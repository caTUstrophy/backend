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

type Location struct {
	Lng float64 `json:"lng" conform:"trim"`
	Lat float64 `json:"lat" conform:"trim"`
}

type Boundaries struct {
	Points []Location `validate:"dive,required"`
}

type CreateRegionPayload struct {
	Name        string     `conform:"trim" validate:"required"`
	Description string     `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Boundaries  Boundaries `conform:"trim" validate:"required"`
}

type UpdateRegionPayload struct {
	Name        string     `conform:"trim" validate:"required"`
	Description string     `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Boundaries  Boundaries `conform:"trim" validate:"required"`
}

type PromoteUserPayload struct {
	Mail string `conform:"trim,email" validate:"required,email"`
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

	if len(Payload.Boundaries.Points) == 0 {

		c.JSON(http.StatusBadRequest, gin.H{
			"Boundaries": "Has to contain {\"Points:\" [{\"lng\": float64, \"lat\": float64}, ...]}",
		})

		return
	}

	Points := make([]gormGIS.GeoPoint, len(Payload.Boundaries.Points))

	for i, point := range Payload.Boundaries.Points {
		Points[i] = gormGIS.GeoPoint{Lng: point.Lng, Lat: point.Lat}
	}

	Region.Boundaries = db.GeoPolygon{
		Points: Points,
	}

	app.DB.Create(&Region)

	// Create admin group for this region
	var admins db.Group
	admins.RegionId = Region.ID
	admins.AccessRight = "admin"
	admins.Description = ("Group for admins of the region " + Region.Name)
	admins.DefaultGroup = false
	admins.ID = fmt.Sprintf("%s", uuid.NewV4())
	app.DB.Create(&admins)

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
		models[i] = CopyNestedModel(region, fieldsRegion).(map[string]interface{})
	}

	c.JSON(http.StatusOK, models)
}

func (app *App) GetRegion(c *gin.Context) {

	// Retrieve region ID from request URL.
	regionID := app.getUUID(c, "regionID")
	if regionID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "No valid region UUID",
		})

		return
	}

	var Region db.Region

	// Select region based on supplied ID from database.
	app.DB.First(&Region, "id = ?", regionID)
	if Region.ID == "" {

		c.JSON(http.StatusNotFound, notFound)

		return
	}

	// Only expose necessary fields in JSON response.
	model := CopyNestedModel(Region, fieldsRegion)

	c.JSON(http.StatusOK, model)
}

func (app *App) UpdateRegion(c *gin.Context) {

	// Authorize user via JWT.
	User := app.AuthorizeShort(c)
	if User == nil {
		return
	}

	// Bind and validate payload.
	var Payload UpdateRegionPayload
	if ok := app.ValidatePayloadShort(c, &Payload); !ok {
		return
	}

	// Get valid regionID.
	regionID := app.getUUID(c, "regionID")
	if regionID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "regionID is no valid UUID",
		})

		return
	}

	// Find and update region parameters.
	var Region db.Region
	app.DB.First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	Region.Name = Payload.Name
	Region.Description = Payload.Description

	// Update boundaries.
	Points := make([]gormGIS.GeoPoint, len(Payload.Boundaries.Points))

	for i, point := range Payload.Boundaries.Points {
		Points[i] = gormGIS.GeoPoint{Lng: point.Lng, Lat: point.Lat}
	}

	Region.Boundaries = db.GeoPolygon{
		Points: Points,
	}

	// Write update to database and return updated object.
	app.DB.Model(&Region).Updates(Region)

	// Only marshal needed fields.
	model := CopyNestedModel(Region, fieldsRegion)

	c.JSON(http.StatusOK, model)
}

func (app *App) ListOffersForRegion(c *gin.Context) {

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

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "regionID is no valid UUID",
		})

		return
	}

	// Load all offers for specified region that were not yet matched.
	var Region db.Region
	app.DB.Preload("Offers", "\"matched\" = ?", false).First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	model := make([]map[string]interface{}, len(Region.Offers))

	for i, offer := range Region.Offers {
		model[i] = CopyNestedModel(offer, fieldsOffer).(map[string]interface{})
	}

	// Send back results to client.
	c.JSON(http.StatusOK, model)
}

func (app *App) ListRequestsForRegion(c *gin.Context) {

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

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "regionID is no valid UUID",
		})

		return
	}

	// Load all requests for specified region that were not yet matched.
	var Region db.Region
	app.DB.Preload("Requests", "\"matched\" = ?", false).First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	model := make([]map[string]interface{}, len(Region.Requests))

	for i, offer := range Region.Requests {
		model[i] = CopyNestedModel(offer, fieldsRequest).(map[string]interface{})
	}

	// Send back results to client.
	c.JSON(http.StatusOK, model)
}

func (app *App) ListMatchingsForRegion(c *gin.Context) {

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

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "regionID is no valid UUID",
		})

		return
	}

	var Region db.Region
	app.DB.First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Find all matchings contained in this region.
	var Matchings []db.Matching
	app.DB.Model(&Region).Related(&Matchings)

	model := make([]map[string]interface{}, len(Matchings))

	for i, matching := range Matchings {

		// Load in matching involved offer and request.
		app.DB.Model(&matching).Related(&matching.Offer).Related(&matching.Request)
		app.DB.Model(&matching.Offer).Related(&matching.Offer.User)
		app.DB.Model(&matching.Request).Related(&matching.Request.User)

		// Marshal it to outside representation and add it to response list.
		model[i] = CopyNestedModel(matching, fieldsMatching).(map[string]interface{})
	}

	c.JSON(http.StatusOK, model)
}

func (app *App) ListAdminsForRegion(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Retrieve region ID from request URL.
	regionID := app.getUUID(c, "regionID")
	if regionID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "regionID is no valid UUID",
		})

		return
	}

	var Region db.Region

	// Select region based on supplied ID from database.
	app.DB.First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Find users that are admins for this region.
	var group db.Group
	app.DB.Preload("Users").First(&group, "region_id = ? AND access_right = ?", regionID, "admin")
	app.DB.Model(&group).Related(&group.Region)

	model := CopyNestedModel(group.Users, fieldsUserNoGroups)

	c.JSON(http.StatusOK, model)
}

func (app *App) PromoteToRegionAdmin(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Retrieve region ID from request URL.
	regionID := app.getUUID(c, "regionID")
	if regionID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "regionID is no valid UUID",
		})

		return
	}

	var Region db.Region

	// Select region based on supplied ID from database.
	app.DB.First(&Region, "id = ?", regionID)

	if Region.ID == "" {

		c.JSON(http.StatusNotFound, gin.H{
			"Error": "The region you requested does not exist.",
		})

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Parse the JSON and check for errors.
	var Payload PromoteUserPayload

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

	// Everything seems fine, promote that user.
	var group db.Group
	app.DB.First(&group, "region_id = ? AND access_right = ?", regionID, "admin")
	if group.ID == "" {

		c.JSON(http.StatusNotFound, gin.H{
			"Error": "For the requested region, no admin group exists. Please tell the developers, this should not occur.",
		})

		return
	}

	// Find the user who is to be promoted and add the group to his or her groups.
	var promotedUser db.User
	app.DB.Preload("Groups").First(&promotedUser, "mail = ?", Payload.Mail)

	if promotedUser.Mail != Payload.Mail {

		c.JSON(http.StatusNotFound, gin.H{
			"Error": "Email unkown to system.",
		})

		return
	}

	promotedUser.Groups = append(promotedUser.Groups, group)
	// app.DB.Model(&promotedUser).Updates(db.User{Groups: promotedUser.Groups})
	app.DB.Save(promotedUser)

	model := CopyNestedModel(promotedUser, fieldsUserNoGroups)

	c.JSON(http.StatusOK, model)
}
