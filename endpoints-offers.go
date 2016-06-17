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

var fieldsGetOffer = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Location": map[string]interface{}{
		"Lng": "lng",
		"Lat": "lat",
	},
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

type CreateOfferPayload struct {
	Name     string `conform:"trim" validate:"required"`
	Location struct {
		Longitude float64 `json:"lng" conform:"trim"`
		Latitude  float64 `json:"lat" conform:"trim"`
	} `validate:"dive,required"`
	Tags           []string `conform:"trim" validate:"dive,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	ValidityPeriod string   `conform:"trim" validate:"required"`
}

// Offers related functions.

func (app *App) CreateOffer(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Payload CreateOfferPayload

	// Expect offer struct fields for creation in JSON request body.
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

	var Offer db.Offer

	// Set insert struct to values from payload.
	Offer.ID = fmt.Sprintf("%s", uuid.NewV4())
	Offer.Name = Payload.Name
	Offer.User = *User
	Offer.UserID = User.ID
	Offer.Location = gormGIS.GeoPoint{Lng: Payload.Location.Longitude, Lat: Payload.Location.Latitude}
	Offer.Tags = make([]db.Tag, 0)

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
				Offer.Tags = append(Offer.Tags, Tag)
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
		Offer.Tags = nil
	}

	// Check if supplied date is a RFC3339 compliant date.
	PayloadTime, err := time.Parse(time.RFC3339, Payload.ValidityPeriod)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"ValidityPeriod": "Offer has to be a RFC3339 compliant date",
		})

		return
	}

	// Check if validity period is yet to come.
	if PayloadTime.Unix() <= time.Now().Unix() {

		c.JSON(http.StatusBadRequest, gin.H{
			"ValidityPeriod": "Offer has to be valid until a date in the future",
		})

		return
	} else {
		Offer.ValidityPeriod = PayloadTime
		Offer.Expired = false
	}

	// Try to map the provided location to all containing regions.
	app.mapLocationToRegions(Offer)

	// Save offer to database.
	app.DB.Create(&Offer)

	// On success: return ID of newly created offer.
	c.JSON(http.StatusCreated, gin.H{
		"ID":             Offer.ID,
		"Name":           Offer.Name,
		"Location":       Offer.Location,
		"Tags":           Offer.Tags,
		"ValidityPeriod": Offer.ValidityPeriod,
	})
}

func (app *App) GetOffer(c *gin.Context) {

	// Check authorization for this function.
	ok, user, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Load offerID from request.
	offerID := app.getUUID(c, "offerID")

	// Retrieve corresponding entry from database.
	var offer db.Offer
	app.DB.Preload("Regions").Preload("Tags").First(&offer, "id = ?", offerID)
	app.DB.Model(&offer).Related(&offer.User)

	// Check if user is admin in any region of this offer.
	if ok := app.CheckScopes(user, offer.Regions, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// He or she can have it, if he or she wants it so badly!
	model := CopyNestedModel(offer, fieldsGetOffer)

	c.JSON(http.StatusOK, model)
}

func (app *App) UpdateOffer(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Add edit rights for concerned user vs. admin.

	offerID := app.getUUID(c, "offerID")
	if offerID == "" {
		return
	}

	// TODO: Change this stub to real function.
	c.JSON(http.StatusOK, gin.H{
		"Name": "Offering COMPLETELY NEW bread",
		"Location": struct {
			Longitude float64
			Latitude  float64
		}{
			15.5,
			45.3,
		},
		"Tags": struct {
			Name string
		}{
			"Food",
		},
		"ValidityPeriod": time.Now().Format(time.RFC3339),
	})
}
