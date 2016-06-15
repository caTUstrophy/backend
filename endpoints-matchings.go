package main

import (
	"fmt"
	"time"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"github.com/satori/go.uuid"
)

// Structs.

type CreateMatchingPayload struct {
	Request string `conform:"trim" validate:"required,uuid4"`
	Offer   string `conform:"trim" validate:"required,uuid4"`
}

// Matching related functions.

func (app *App) CreateMatching(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, "worldwide", "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
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

	// Check that offer and request do exist.
	var CountOffer int
	app.DB.Model(&db.Offer{}).Where("id = ?", Payload.Offer).Count(&CountOffer)
	var CountRequest int
	app.DB.Model(&db.Request{}).Where("id = ?", Payload.Request).Count(&CountRequest)

	if (CountOffer == 0) || (CountRequest == 0) {

		// Signal request failure to client.
		c.JSON(http.StatusBadRequest, gin.H{
			"Matching": "Offer or request does not exist",
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

	// Get request and offer to resolve foreign key dependencies.
	var Offer db.Offer
	app.DB.First(&Offer, "id = ?", Payload.Offer)
	var Request db.Request
	app.DB.First(&Request, "id = ?", Payload.Request)

	// Save matching.
	var Matching db.Matching
	Matching.ID = fmt.Sprintf("%s", uuid.NewV4())
	Matching.OfferId = Payload.Offer
	Matching.Offer = Offer
	Matching.RequestId = Payload.Request
	Matching.Request = Request

	// Save matching to database.
	app.DB.Create(&Matching)

	c.JSON(http.StatusCreated, Matching)
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

	matchingID := c.Params.ByName("matchingID")

	// Validate sent matchingID.
	errs := app.Validator.Field(matchingID, "required,uuid4")
	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp["matchingID"] = "Is required"
			} else if err.Tag == "uuid4" {
				errResp["matchingID"] = "Needs to be an UUID version 4"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	var Matching db.Matching

	// Retrieve all requests from database.
	app.DB.First(&Matching, "id = ?", matchingID)

	// Send back results to client.
	c.JSON(http.StatusOK, Matching)
}

func (app *App) ListMatchings(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, "worldwide", "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Change this stub to real function.
	// 1) Check for expired fields in offers and requests - via extra argument?
	c.JSON(http.StatusOK, gin.H{
		"Offer": struct {
			ID             string
			Name           string
			User           interface{}
			Location       interface{}
			Tags           interface{}
			ValidityPeriod string
		}{
			"a55b22de-5955-4d2a-9078-1479bda097e7",
			"Offering bread",
			struct {
				ID string
			}{
				"08f5588f-40c0-4ad1-9fd3-ce20e37903d3",
			},
			struct {
				Longitude float64
				Latitude  float64
			}{
				13.9,
				50.1,
			},
			struct {
				Name string
			}{
				"Food",
			},
			time.Now().Format(time.RFC3339),
		},
		"Request": struct {
			ID             string
			Name           string
			User           interface{}
			Location       interface{}
			Tags           interface{}
			ValidityPeriod string
		}{
			"adbe5c76-e3ac-4dee-90a4-d85054c64bdd",
			"Looking for bread",
			struct {
				ID string
			}{
				"aaf84b79-ec0f-45df-9282-58850064fcbe",
			},
			struct {
				Longitude float64
				Latitude  float64
			}{
				13.9,
				50.1,
			},
			struct {
				Name string
			}{
				"Food",
			},
			time.Now().Format(time.RFC3339),
		},
	})
}
