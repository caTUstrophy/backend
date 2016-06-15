package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
)

type CreateAreaPayload struct {
	Name        string             `conform:"trim" validate:"required"`
	Description string             `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Boundaries  []gormGIS.GeoPoint `conform:"trim"`
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

	var Areas []db.Area

	// Retrieve all offers from database.
	app.DB.Find(&Areas)

	c.JSON(http.StatusOK, Areas)
}

func (app *App) GetRegion(c *gin.Context) {

	// Retrieve area ID from request URL.
	areaID := c.Params.ByName("areaID")

	errs := app.Validator.Field(areaID, "required,uuid4")
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

		return
	}

	var Area db.Area

	// Select area based on supplied ID from database.
	app.DB.First(&Area, "id = ?", areaID)
	app.DB.Model(&Area).Association("Boundaries").Find(&Area.Boundaries)

	c.JSON(http.StatusOK, Area)
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

	var Offers []db.Offer

	// Retrieve all offers from database.
	app.DB.Find(&Offers, "Location = ?", region)

	// TODO: remove loop and exchange for preload
	for i := 0; i < len(Offers); i++ {
		app.DB.Select("name, id").First(&Offers[i].User, "mail = ?", User.Mail)
	}

	// Send back results to client.
	c.JSON(http.StatusOK, Offers)
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

func (app *App) GetMatchingsForRegion(c *gin.Context) {

}
