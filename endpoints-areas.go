package main

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

/*
func (app *App) CreateArea(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Change stub to real function.
	c.JSON(http.StatusCreated, gin.H{
		"ID":          fmt.Sprintf("%s", uuid.NewV4()),
		"Name":        "Algeria",
		"Description": "Mountain region hit by an earth quake of strength 4.0",
		"Boundaries": struct {
			Boundaries []db.Point
		}{
			[]db.Point{
				db.Point{3.389017, 36.416215},
				db.Point{3.358667, 36.391414},
				db.Point{3.391039, 36.362402},
				db.Point{3.418206, 36.392172},
				db.Point{3.389017, 36.416215},
			},
		},
	})
}
*/

func (app *App) ListAreas(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	type TmpPoint struct {
		Longitude float64
		Latitude  float64
	}

	type TmpArea struct {
		ID          string
		Name        string
		Description string
		Boundaries  []TmpPoint
	}

	TmpResponse := []TmpArea{
		TmpArea{
			fmt.Sprintf("%s", uuid.NewV4()),
			"Algeria",
			"Mountain region hit by an earth quake of strength 4.0",
			[]TmpPoint{
				TmpPoint{3.389017, 36.416215},
				TmpPoint{3.358667, 36.391414},
				TmpPoint{3.391039, 36.362402},
				TmpPoint{3.418206, 36.392172},
				TmpPoint{3.389017, 36.416215},
			},
		},
	}

	// TODO: Change stub to real function.
	c.JSON(http.StatusOK, TmpResponse)
}

/*
func (app *App) GetArea(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

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

	c.JSON(http.StatusOK, Area)
}
*/
