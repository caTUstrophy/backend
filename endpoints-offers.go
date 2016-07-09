package main

import (
	"fmt"
	"sort"
	"time"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
)

// Structs

type CreateOfferPayload struct {
	Name     string `conform:"trim" validate:"required"`
	Location struct {
		Longitude float64 `json:"lng" conform:"trim"`
		Latitude  float64 `json:"lat" conform:"trim"`
	} `validate:"dive,required"`
	Radius         float64  `validate:"required"`
	Tags           []string `conform:"trim" validate:"dive,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Description    string   `conform:"trim" validate:"excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	ValidityPeriod string   `conform:"trim" validate:"required"`
}

type UpdateOfferPayload struct {
	Name     string `conform:"trim" validate:"required"`
	Location struct {
		Longitude float64 `json:"lng" conform:"trim"`
		Latitude  float64 `json:"lat" conform:"trim"`
	} `validate:"dive,required"`
	Radius         float64  `validate:"required"`
	Tags           []string `conform:"trim" validate:"dive,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Description    string   `conform:"trim" validate:"excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	ValidityPeriod string   `conform:"trim" validate:"required"`
	Matched        bool     `conform:"trim" validate:"exists"`
}

// Functions

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

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Supplied values in JSON body could not be parsed",
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

	var Offer db.Offer

	// Set insert struct to values from payload.
	Offer.ID = fmt.Sprintf("%s", uuid.NewV4())
	Offer.Name = Payload.Name
	Offer.User = *User
	Offer.UserID = User.ID
	Offer.Location = gormGIS.GeoPoint{Lng: Payload.Location.Longitude, Lat: Payload.Location.Latitude}
	Offer.Radius = Payload.Radius
	Offer.Description = Payload.Description
	Offer.Tags = make([]db.Tag, 0)

	// If tags were supplied, check if they exist in our system.
	if len(Payload.Tags) > 0 {

		// Retrieve all currently available tags from database.
		var Tags []db.Tag
		app.DB.Find(&Tags)

		// Make tags list searchable in fast time.
		sort.Sort(db.TagsByName(Tags))

		allTagsExist := true

		for _, tag := range Payload.Tags {

			// Find tag in sorted tags list.
			i := sort.Search(len(Tags), func(i int) bool {
				return Tags[i].Name >= tag
			})

			if i < len(Tags) && Tags[i].Name == tag {
				// We found the supplied tag, add it to tags list.
				Offer.Tags = append(Offer.Tags, Tags[i])
			} else {
				// Set flag to false, if one tag was not found.
				allTagsExist = false
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

	// Load all regions to which we just mapped the offer's location.
	app.DB.Preload("Regions").First(&Offer)

	// Calculate the matching score of this offer with all possible requests.
	go app.CalcMatchScoreForOffer(Offer)

	model := CopyNestedModel(Offer, fieldsOfferWithUser)

	c.JSON(http.StatusCreated, model)
}

func (app *App) GetOffer(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Load offerID from request.
	offerID := app.getUUID(c, "offerID")
	if offerID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "offerID is no valid UUID",
		})

		return
	}

	// Retrieve corresponding entry from database.
	var offer db.Offer
	app.DB.Preload("Regions").Preload("Tags").First(&offer, "id = ?", offerID)
	app.DB.Model(&offer).Related(&offer.User)

	// Validity check:
	// User accessing this offer has to be either an admin in any region
	// of this offer or has to be the owning user of this offer.
	if ok := ((offer.UserID == User.ID) || app.CheckScopes(User, offer.Regions, "admin")); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// He or she can have it, if he or she wants it so badly!
	model := CopyNestedModel(offer, fieldsOfferWithUser)

	c.JSON(http.StatusOK, model)
}

func (app *App) UpdateOffer(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Load offerID from request.
	offerID := app.getUUID(c, "offerID")
	if offerID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "offerID is no valid UUID",
		})

		return
	}

	// Retrieve corresponding entry from database.
	var Offer db.Offer
	app.DB.Preload("Regions").Preload("Tags").First(&Offer, "id = ?", offerID)
	app.DB.Model(&Offer).Related(&Offer.User)

	// Validity check:
	// User accessing this offer has to be either an admin in any region
	// of this offer or has to be the owning user of this offer.
	if ok := ((Offer.UserID == User.ID) || app.CheckScopes(User, Offer.Regions, "admin")); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Bind payload.
	var Payload UpdateRequestPayload
	if ok := app.ValidatePayloadShort(c, &Payload); !ok {
		return
	}

	Offer.Name = Payload.Name
	Offer.Location = gormGIS.GeoPoint{Lng: Payload.Location.Longitude, Lat: Payload.Location.Latitude}
	Offer.Radius = Payload.Radius
	Offer.Description = Payload.Description

	// Delete all tags associated with request.
	for _, Tag := range Offer.Tags {
		app.DB.Exec("DELETE FROM offer_tags WHERE \"request_id\" = ? AND \"tag_name\" = ?", Offer.ID, Tag.Name)
	}

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
		Offer.ValidityPeriod = PayloadTime
		Offer.Expired = false
	}

	// delete all regions associated with request
	app.DB.Exec("DELETE FROM region_requests WHERE request_id = ?", Offer.ID)
	// Try to map the provided location to all containing regions.

	// map into new regions
	Offer.Regions = []db.Region{}
	app.mapLocationToRegions(Offer)

	app.DB.Model(&Offer).Updates(Offer)
	model := CopyNestedModel(Offer, fieldsRequestWithUser)
	c.JSON(http.StatusOK, model)

}
