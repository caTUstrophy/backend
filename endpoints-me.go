package main

import (
	"fmt"
	"time"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
)

// Functions

func (app *App) GetMe(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Marshal only necessary fields.
	response := CopyNestedModel(*User, fieldsUser)

	c.JSON(http.StatusOK, response)
}

func (app *App) UpdateMe(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	app.UpdateUserObject(User, c, false)

	return
}

func (app *App) ListUserOffers(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Offers []db.Offer
	app.DB.Preload("Tags").Find(&Offers, "user_id = ?", User.ID)

	response := make([]interface{}, len(Offers))

	for i, o := range Offers {

		// Check if offer is not already expired.
		if o.ValidityPeriod.After(time.Now()) {

			// Marshal only needed fields.
			model := CopyNestedModel(o, fieldsOffer)
			response[i] = model
		}
	}

	c.JSON(http.StatusOK, response)
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

	var Requests []db.Request
	app.DB.Preload("Tags").Find(&Requests, "user_id = ?", User.ID)

	response := make([]interface{}, len(Requests))

	for i, r := range Requests {

		// Check if offer is not already expired.
		if r.ValidityPeriod.After(time.Now()) {

			// Marshal only needed fields.
			model := CopyNestedModel(r, fieldsRequest)
			response[i] = model
		}
	}

	c.JSON(http.StatusOK, response)
}

func (app *App) ListUserMatchings(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Get all matchings in order to filter for this user's ones.
	// This might not be a viable option for big matching sets.
	var Matchings []db.Matching
	app.DB.Find(&Matchings, "\"invalid\" = ?", false)

	response := make([]interface{}, 0)

	for _, Matching := range Matchings {

		// Load offer and request for this matching.
		app.DB.Model(&Matching).Related(&Matching.Offer).Related(&Matching.Request)

		// Check if user is owner of either the offer or the request.
		if (Matching.Offer.UserID == User.ID) || (Matching.Request.UserID == User.ID) {

			// If so - load some more related data.
			app.DB.Model(&Matching).Related(&Matching.Region)
			app.DB.Model(&Matching.Offer).Related(&Matching.Offer.User)
			app.DB.Model(&Matching.Request).Related(&Matching.Request.User)

			// Add marshalled version of matching to response list.
			response = append(response, CopyNestedModel(Matching, fieldsMatching))
		}
	}

	c.JSON(http.StatusOK, response)
}
