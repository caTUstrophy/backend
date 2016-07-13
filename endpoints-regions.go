package main

import (
	"fmt"
	"sort"

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
	Description string     `conform:"trim" validate:"required"`
	Boundaries  Boundaries `conform:"trim" validate:"required"`
}

type UpdateRegionPayload struct {
	Name        string     `conform:"trim" validate:"required"`
	Description string     `conform:"trim" validate:"required"`
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

	// Save region.
	var Region db.Region
	Region.ID = fmt.Sprintf("%s", uuid.NewV4())
	Region.Name = Payload.Name
	Region.Description = Payload.Description

	// There are currently no offers or requests in this region, so there can't be a recommendation.
	Region.RecommendationUpdated = true

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

	// Create admin group for this region.
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

	// Load all offers for specified region that are
	// - not yet expired
	// - and not yet matched.
	var Region db.Region
	app.DB.Preload("Offers.Tags").Preload("Offers", "\"expired\" = ? AND \"matched\" = ?", false, false).First(&Region, "id = ?", regionID)

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

	// Load all requests for specified region that are
	// - not yet expired
	// - and not yet matched.
	var Region db.Region
	app.DB.Preload("Requests.Tags").Preload("Requests", "\"expired\" = ? AND \"matched\" = ?", false, false).First(&Region, "id = ?", regionID)

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

func (app *App) ListRecommendationsForRegion(c *gin.Context) {

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

	if !Region.RecommendationUpdated {
		app.RecommendMatching(Region)
	}

	var recommendations []db.MatchingScore
	app.DB.Find(&recommendations, "recommended = ?", true)
	for i, rec := range recommendations {
		app.DB.Model(&rec).Preload("Tags").Related(&recommendations[i].Offer)
		app.DB.Model(&rec).Preload("Tags").Related(&recommendations[i].Request)
	}

	model := CopyNestedModel(recommendations, fieldsRecommendations)

	c.JSON(http.StatusOK, model)
}

func (app *App) ListOffersForRequest(c *gin.Context) {

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

	// Load all offers for specified region that are
	// - not yet expired
	// - and not yet matched.
	app.DB.Preload("Offers", "\"expired\" = ? AND \"matched\" = ?", false, false).First(&Region, "\"id\" = ?", regionID)

	// Sort offers by UUID.
	sort.Sort(db.OffersByUUID(Region.Offers))

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

	// Retrieve request ID from request URL.
	requestID := app.getUUID(c, "requestID")
	if requestID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "requestID is no valid UUID",
		})

		return
	}

	var Request db.Request

	// Select region based on supplied ID from database.
	app.DB.First(&Request, "\"id\" = ?", requestID)

	if Request.ID == "" {

		c.JSON(http.StatusNotFound, gin.H{
			"Error": "The request you requested does not exist.",
		})

		return
	}

	// Retrieve matching scores from database table for (Region, *, Request).
	var MatchingScores []db.MatchingScore
	app.DB.Order("\"matching_score\" DESC").Find(&MatchingScores, "\"region_id\" = ? AND \"request_id\" = ?", Region.ID, Request.ID)

	model := make([]map[string]interface{}, len(Region.Offers))

	// Iterate over all found elements in matching scores list.
	for _, matchingScore := range MatchingScores {

		// Find offer that matches MatchingScores.OfferID in sorted offers list.
		i := sort.Search(len(Region.Offers), func(i int) bool {
			return Region.Offers[i].ID >= matchingScore.OfferID
		})

		if i < len(Region.Offers) && Region.Offers[i].ID == matchingScore.OfferID {

			// We found the correct offer, add it to result list
			model[i] = CopyNestedModel(Region.Offers[i], fieldsOffer).(map[string]interface{})

			// TODO: Add matching score field and recommended field.
		}
	}

	// Send back results to client.
	c.JSON(http.StatusOK, model)
}

func (app *App) ListRequestsForOffer(c *gin.Context) {

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

	// Retrieve offer ID from request URL.
	offerID := app.getUUID(c, "offerID")
	if offerID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "offerID is no valid UUID",
		})

		return
	}

	var Offer db.Offer

	// Select region based on supplied ID from database.
	app.DB.First(&Offer, "id = ?", offerID)

	if Offer.ID == "" {

		c.JSON(http.StatusNotFound, gin.H{
			"Error": "The offer you requested does not exist.",
		})

		return
	}

	// TODO: finish. See above function for reference.
}
