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

type CreateRequestPayload struct {
	Name           string   `conform:"trim" validate:"required"`
	Location       string   `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Tags           []string `conform:"trim" validate:"dive,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	ValidityPeriod string   `conform:"trim" validate:"required"`
}

// Requests related functions.

func (app *App) CreateRequest(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Payload CreateRequestPayload

	// Expect request struct fields for creation in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		// Check if error was caused by failed unmarshalling string -> []string.
		if err.Error() == "json: cannot unmarshal string into Go value of type []string" {

			c.JSON(http.StatusBadRequest, gin.H{
				"Tags": "Provide an array, not a string",
			})

			return
		}
	}

	// Validate sent request creation data.
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

	var Request db.Request

	// Set insert struct to values from payload.
	Request.ID = fmt.Sprintf("%s", uuid.NewV4())
	Request.Name = Payload.Name
	Request.User = *User
	Request.UserID = User.ID
	Request.Location = Payload.Location
	Request.Tags = make([]db.Tag, 0)

	// If tags were supplied, check if they exist in our system.
	if len(Payload.Tags) > 0 {

		allTagsExist := true

		for _, tag := range Payload.Tags {

			var Tag db.Tag

			// Count number of results for query of name of tags.
			app.DB.First(&Tag, "name = ?", tag)

			// Set flag to false, if one tag was not found.
			if Tag.Name == "" {
				allTagsExist = false
			} else {
				Request.Tags = append(Request.Tags, Tag)
			}
		}

		// If at least one of the tags does not exist - return error.
		if !allTagsExist {

			c.JSON(http.StatusBadRequest, gin.H{
				"Tags": "One or multiple tags do not exist",
			})

			return
		}
	} else {
		Request.Tags = nil
	}

	// Check if supplied date is a RFC3339 compliant date.
	PayloadTime, err := time.Parse(time.RFC3339, Payload.ValidityPeriod)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"ValidityPeriod": "Request has to be a RFC3339 compliant date",
		})

		return
	}

	// Check if validity period is yet to come.
	if PayloadTime.Unix() <= time.Now().Unix() {

		c.JSON(http.StatusBadRequest, gin.H{
			"ValidityPeriod": "Request has to be valid until a date in the future",
		})

		return
	} else {
		Request.ValidityPeriod = PayloadTime
		Request.Expired = false
	}

	// Save request to database.
	app.DB.Create(&Request)

	// On success: return ID of newly created request.
	c.JSON(http.StatusCreated, gin.H{
		"ID":             Request.ID,
		"Name":           Request.Name,
		"Location":       Request.Location,
		"Tags":           Request.Tags,
		"ValidityPeriod": Request.ValidityPeriod,
	})
}

func (app *App) ListRequests(c *gin.Context) {

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

	region := c.Params.ByName("region")

	// Validate sent region.
	errs := app.Validator.Field(region, "required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C")
	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp["region"] = "Is required"
			} else if err.Tag == "excludesall" {
				errResp["region"] = "Contains unallowed characters"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	var Requests []db.Request

	// Retrieve all requests from database.
	// app.DB.Preload("User").Find(&Requests, "location = ?", region)
	app.DB.Find(&Requests, "location = ?", region)

	// TODO: remove loop and exchange for preload
	for i := 0; i < len(Requests); i++ {
		app.DB.Select("name, id").First(&Requests[i].User, "mail = ?", User.Mail)
	}

	// Send back results to client.
	c.JSON(http.StatusOK, Requests)
}

func (app *App) ListUserRequests(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}




	// TODO: Change this stub to real function.
	// 1) Only retrieve requests from user.
	// 2) Check expired field - extra argument for that?
	// 3) Only return what's needed.

	type TmpLocation struct {
		Longitude float64
		Latitude  float64
	}

	type TmpTag struct {
		Name string
	}

	type TmpUserRequest struct {
		ID             string
		Name           string
		Location       TmpLocation
		Tags           []TmpTag
		ValidityPeriod string
	}

	// 1) Only retrieve requests from user.
	var Requests []db.Request
	app.DB.Find(&Requests, "user_id = ?", User.ID)

	var ReturnRequests []TmpUserRequest
	ReturnRequests = make([]TmpUserRequest, 0)
	for _, r := range Requests {
		// 2) Check expired field - extra argument for that?
		if r.ValidityPeriod.After(time.Now()) {
			// 3) Only return what's needed.
			ReturnRequests = append(ReturnRequests, TmpUserRequest{
				r.ID,
				r.Name,
				TmpLocation{
					13.9,
					50.1,
				},
				[]TmpTag{
					TmpTag{
						"Food",
					},
				},
				r.ValidityPeriod.Format(time.RFC3339),
			})
		}
	}


	c.JSON(http.StatusOK, ReturnRequests)
}

func (app *App) UpdateUserRequest(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	requestID := c.Params.ByName("requestID")

	// Validate sent requestID.
	errs := app.Validator.Field(requestID, "required,uuid4")
	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp["requestID"] = "Is required"
			} else if err.Tag == "uuid4" {
				errResp["requestID"] = "Needs to be an UUID version 4"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	// TODO: Change this stub to real function.
	c.JSON(http.StatusOK, gin.H{
		"Name": "Looking for COMPLETELY NEW bread",
		"Location": struct {
			Longitude float64
			Latitude  float64
		}{
			14.0,
			49.9,
		},
		"Tags": struct {
			Name string
		}{
			"Food",
		},
		"ValidityPeriod": time.Now().Format(time.RFC3339),
	})
}
