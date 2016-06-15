package main

import (
	"log"

	"net/http"

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

// Intersects the provided location (GeoPoint) in requests
// and offers with all regions available to determine to which
// region this location should be mapped.
func (app *App) mapLocationToRegions(location gormGIS.GeoPoint, structName string) {

	var ContainingRegionIDs []string

	app.DB.Exec("SELECT r.id FROM regions r WHERE ST_INTERSECTS(ST_GeographyFromText(?), r.boundaries);", location.String()).Scan(&ContainingRegionIDs)

	log.Printf("[mapLocationToRegions] IDs of containing regions: %v\n", ContainingRegionIDs)

	/*
		if structName == "Offer" {

		} else if structName == "Request" {

		}
	*/
}
