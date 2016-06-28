package main

import (
	"fmt"
	"log"
	"time"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"github.com/satori/go.uuid"
)

// Structs

type CreateMatchingPayload struct {
	Region  string `conform:"trim" validate:"required,uuid4"`
	Request string `conform:"trim" validate:"required,uuid4"`
	Offer   string `conform:"trim" validate:"required,uuid4"`
}

// Functions

func (app *App) CreateMatching(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Payload CreateMatchingPayload

	// Expect user struct fields in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Supplied values in JSON body could not be parsed",
		})

		return
	}

	// Validate sent user login data.
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
			} else if err.Tag == "uuid4" {
				errResp[err.Field] = "Needs to be an UUID version 4"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	// Fetch involved request and offer.
	var Offer db.Offer
	app.DB.First(&Offer, "id = ?", Payload.Offer)
	var Request db.Request
	app.DB.First(&Request, "id = ?", Payload.Request)

	// Check that offer and request do exist.
	if (Offer.UserID == "") || (Request.UserID == "") {

		// Signal request failure to client.
		c.JSON(http.StatusBadRequest, gin.H{
			"Matching": "Offer or request does not exist",
		})

		return
	}

	// Check that offer or request are not already expired.
	if (Offer.Expired) || (Request.Expired) {

		// Signal request failure to client.
		c.JSON(http.StatusBadRequest, gin.H{
			"Matching": "Offer or request already expired",
		})

		return
	}

	// Check that offer or request are not already matched.
	if (Offer.Matched) || (Request.Matched) {

		// Signal request failure to client.
		c.JSON(http.StatusBadRequest, gin.H{
			"Matching": "Offer or request already matched",
		})

		return
	}

	// Check for duplicate of matching.
	var CountDup int
	app.DB.Model(&db.Matching{}).Where("offer_id = ? AND request_id = ?", Payload.Offer, Payload.Request).Count(&CountDup)

	if CountDup > 0 {

		// Signal request failure to client.
		c.JSON(http.StatusBadRequest, gin.H{
			"Matching": "Already exists",
		})

		return
	}

	// Check if user has actually admin rights for specified region.
	var ContainingRegion db.Region
	app.DB.First(&ContainingRegion, "id = ?", Payload.Region)

	if ok := app.CheckScope(User, ContainingRegion, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Save matching.
	var Matching db.Matching
	Matching.ID = fmt.Sprintf("%s", uuid.NewV4())
	Matching.RegionId = Payload.Region
	Matching.OfferId = Payload.Offer
	Matching.Offer = Offer
	Matching.RequestId = Payload.Request
	Matching.Request = Request

	// Save matching to database.
	app.DB.Create(&Matching)

	// Set 'Matched' field of involved request and offer to true.
	Offer.Matched = true
	Request.Matched = true

	app.DB.Save(&Offer)
	app.DB.Save(&Request)

	// Trigger a notification for involved users.
	NotifyOfferUser := db.Notification{
		ID:        fmt.Sprintf("%s", uuid.NewV4()),
		Type:      db.NotificationMatching,
		UserID:    Offer.UserID,
		ItemID:    Matching.ID,
		Read:      false,
		CreatedAt: time.Now(),
	}

	NotifyRequestUser := db.Notification{
		ID:        fmt.Sprintf("%s", uuid.NewV4()),
		Type:      db.NotificationMatching,
		UserID:    Request.UserID,
		ItemID:    Matching.ID,
		Read:      false,
		CreatedAt: time.Now(),
	}

	log.Println("time.Now():", time.Now())
	log.Println("time.Now().Format(RFC):", time.Now().Format(time.RFC3339))

	app.DB.Create(&NotifyOfferUser)
	app.DB.Create(&NotifyRequestUser)

	app.DB.Model(&Matching).Related(&Matching.Offer)
	app.DB.Model(&Matching.Offer).Related(&Matching.Offer.User)
	app.DB.Model(&Matching).Related(&Matching.Request)
	app.DB.Model(&Matching.Request).Related(&Matching.Request.User)
	// Only expose fields that are necessary.
	model := CopyNestedModel(Matching, fieldsMatching)

	c.JSON(http.StatusCreated, model)
}

func (app *App) GetMatching(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	matchingID := app.getUUID(c, "matchingID")
	if matchingID == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"Error": "macthingID is not a valid UUID",
		})
		return
	}

	var Matching db.Matching

	// Retrieve the specified matching
	app.DB.First(&Matching, "id = ?", matchingID)
	app.DB.Model(&Matching).Related(&Matching.Offer)
	app.DB.Model(&Matching.Offer).Related(&Matching.Offer.User)
	app.DB.Model(&Matching).Related(&Matching.Request)
	app.DB.Model(&Matching.Request).Related(&Matching.Request.User)

	// Only expose fields that are necessary.
	model := CopyNestedModel(Matching, fieldsMatching)

	// Send back results to client.
	c.JSON(http.StatusOK, model)
}
