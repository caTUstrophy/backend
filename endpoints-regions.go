package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
)

type CreateRegionPayload struct {
	Name        string             `conform:"trim" validate:"required"`
	Description string             `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Boundaries  []gormGIS.GeoPoint `conform:"trim"`
}

var fieldsListRegions = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Boundaries": map[string]interface{}{
		"Points": map[string]interface{}{
			"Lng": "lng",
			"Lat": "lat",
		},
	},
	"Description": "Description",
}

var fieldsGetRegion = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Boundaries": map[string]interface{}{
		"Points": map[string]interface{}{
			"Lng": "lng",
			"Lat": "lat",
		},
	},
	"Description": "Description",
}

func (app *App) CreateRegion(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// check scope if we want it for admins only

	// TODO: Change stub to real function.
	c.JSON(http.StatusCreated, gin.H{
		"ID":          fmt.Sprintf("%s", uuid.NewV4()),
		"Name":        "Algeria",
		"Description": "Mountain region hit by an earth quake of strength 4.0",
		"Boundaries": struct {
			Boundaries []gormGIS.GeoPoint
		}{
			[]gormGIS.GeoPoint{
				gormGIS.GeoPoint{3.389017, 36.416215},
				gormGIS.GeoPoint{3.358667, 36.391414},
				gormGIS.GeoPoint{3.391039, 36.362402},
				gormGIS.GeoPoint{3.418206, 36.392172},
				gormGIS.GeoPoint{3.389017, 36.416215},
			},
		},
	})
}

func (app *App) ListRegions(c *gin.Context) {

	var Regions []db.Region

	// Retrieve all offers from database.
	app.DB.Find(&Regions)

	models := make([]map[string]interface{}, len(Regions))

	// Iterate over all regions in database return and marshal it.
	for i, region := range Regions {
		models[i] = CopyNestedModel(region, fieldsListRegions)
	}

	c.JSON(http.StatusOK, models)
}

func (app *App) GetRegion(c *gin.Context) {

	// Retrieve region ID from request URL.
	regionID := app.getUUID(c, "regionID")
	if regionID == "" {
		return
	}

	var Region db.Region

	// Select region based on supplied ID from database.
	app.DB.First(&Region, "id = ?", regionID)

	// Only expose necessary fields in JSON response.
	model := CopyNestedModel(Region, fieldsGetRegion)

	c.JSON(http.StatusOK, model)
}

func (app *App) UpdateRegion(c *gin.Context) {

}

func (app *App) GetOffersForRegion(c *gin.Context) {

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
		return
	}

	var Region db.Region
	app.DB.Preload("Users").Preload("Offers").First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Send back results to client.
	c.JSON(http.StatusOK, Region.Offers)
}

func (app *App) GetRequestsForRegion(c *gin.Context) {

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
		return
	}

	var Region db.Region
	app.DB.Preload("Users").Preload("Requests").First(&Region, "id = ?", regionID)

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, Region, "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Send back results to client.
	c.JSON(http.StatusOK, Region.Requests)
}

func (app *App) GetMatchingsForRegion(c *gin.Context) {

}
