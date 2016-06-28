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

		// 2) Check expired field - extra argument for that?
		if o.ValidityPeriod.After(time.Now()) {

			// 3) Only return what's needed
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

		// 2) Check expired field - extra argument for that?
		if r.ValidityPeriod.After(time.Now()) {

			// 3) Only return what's needed
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

	var Notifications []db.Notification
	app.DB.Find(&Notifications, "user_id = ?", User.ID)

	response := make([]interface{}, len(Notifications))

	for i, notification := range Notifications {

		// Find matching element and connected elements.
		var Matching db.Matching
		app.DB.First(&Matching, "id = ?", notification.ItemID)
		app.DB.Model(&Matching).Related(&Matching.Offer)
		app.DB.Model(&Matching.Offer).Related(&Matching.Offer.User)
		app.DB.Model(&Matching).Related(&Matching.Request)
		app.DB.Model(&Matching.Request).Related(&Matching.Request.User)

		response[i] = CopyNestedModel(Matching, fieldsMatching)
	}

	c.JSON(http.StatusOK, response)
}
