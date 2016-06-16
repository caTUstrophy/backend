package main

import (
	"log"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/nferruzzi/gormGIS"
)

// Checks if a generic URL substring is present in the
// current request and, if so, attempts to validate it
// as an UUID version 4.
func (app *App) getUUID(c *gin.Context, par string) string {

	parID := c.Params.ByName(par)
	errs := app.Validator.Field(parID, "required,uuid4")

	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp[err.Field] = "Is required"
			} else if err.Tag == "uuid4" {
				errResp[err.Field] = "Needs to be an UUID version 4"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return ""
	}

	return parID
}

// Intersects the location of the provided item (offer or request)
// with all regions available to determine to which region this
// item should be mapped.
func (app *App) mapLocationToRegions(item interface{}) {

	var itemType string
	var location gormGIS.GeoPoint
	var ContRegionID string
	var ContRegionIDs []string

	// Determine whether we received an offer or a request.
	switch item.(type) {
	case db.Offer:

		itemType = "Offer"

		asssertedItem, ok := item.(db.Offer)
		if !ok {
			log.Fatal("[mapLocationToRegions] Type assertion to db.Offer was unsuccessful. Returning from function.")
			return
		}

		location = asssertedItem.Location
	case db.Request:

		itemType = "Request"

		asssertedItem, ok := item.(db.Request)
		if !ok {
			log.Fatal("[mapLocationToRegions] Type assertion to db.Request was unsuccessful. Returning from function.")
			return
		}

		location = asssertedItem.Location
	default:
		itemType = "UNKNOWN"
		log.Println("[mapLocationToRegions] itemType was UNKNOWN")
		return
	}

	// Find all IDs of regions with which the supplied point intersects.
	regionRows, err := app.DB.Raw("SELECT \"id\" FROM \"regions\" WHERE ST_INTERSECTS(ST_GeographyFromText(?), \"regions\".\"boundaries\")", location.String()).Rows()
	if err != nil {
		log.Fatal(err)
	}

	// Close row connection on function exit.
	defer regionRows.Close()

	// Iterate over all found regions and save regionID to slice.
	for regionRows.Next() {
		regionRows.Scan(&ContRegionID)
		ContRegionIDs = append(ContRegionIDs, ContRegionID)
	}

	if len(ContRegionIDs) > 0 {

		var ContRegions []db.Region

		// Retrieve all regions into above list of containing regions.
		// Only regions with IDs from intersecting region list will be chosen.
		app.DB.Where("id in (?)", ContRegionIDs).Preload("Offers").Preload("Requests").Find(&ContRegions)

		for _, ContRegion := range ContRegions {

			// Depending on type of item, save an offer or a request into list.
			if itemType == "Offer" {
				ContRegion.Offers = append(ContRegion.Offers, item.(db.Offer))
			} else if itemType == "Request" {
				ContRegion.Requests = append(ContRegion.Requests, item.(db.Request))
			}

			// Save changed offers or requests of a region to database.
			app.DB.Save(&ContRegion)
		}
	} else {
		log.Println("[mapLocationToRegions] No intersecting regions found.")
	}
}
