package main

import (
	"fmt"
	"time"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
)

var fieldsGetRequest = map[string]interface{}{
	"ID":       "ID",
	"Name":     "Name",
	"Location": "Location",
	"Tags": map[string]interface{}{
		"Name": "Name",
	},
	"User": map[string]interface{}{
		"ID":   "ID",
		"Name": "Name",
		"Mail": "Mail",
	},
}

// Structs.

type CreateRequestPayload struct {
	Name     string `conform:"trim" validate:"required"`
	Location struct {
		Longitude float64 `json:"lng" conform:"trim"`
		Latitude  float64 `json:"lat" conform:"trim"`
	} `validate:"dive,required"`
	Tags           []string `conform:"trim" validate:"dive,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	ValidityPeriod string   `conform:"trim" validate:"required"`
}

// Requests related functions.

// Looks up all regions that match the location of this offer
func (app *App) assignRegionsToRequest(request db.Request) {

}

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
	Request.Location = gormGIS.GeoPoint{Lng: Payload.Location.Longitude, Lat: Payload.Location.Latitude}
	Request.Tags = make([]db.Tag, 0)
	// The regions that match the location will be assigned outside
	app.assignRegionsToRequest(Request)
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

func (app *App) GetRequest(c *gin.Context) {
	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}
	// Load request
	requestID := app.getUUID(c, "requestID")
	var request db.Request

	app.DB.Preload("Regions").Preload("Tags").First(&request, "id = ?", requestID)
	app.DB.Model(&request).Related(&request.User)

	// Check if user is admin in any region of this request
	hasPerm := app.CheckScope(user, db.Region{}, "superadmin")
	if hasPerm := app.CheckScopes(User, request.Regions, "admin"); !hasPerm {
		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}
	// He or she can have it, if he or she wants it so bad
	model := CopyNestedModel(request, fieldsGetRequest)
	c.JSON(http.StatusOK, model)
	return
}

func (app *App) UpdateRequest(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Add edit rights for concerned user vs. admin.

	requestID := app.getUUID(c, "requestID")
	if requestID == "" {
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
