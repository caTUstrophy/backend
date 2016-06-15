package main

import (
	"fmt"
	"time"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

var fieldsGetUser = map[string]interface{}{
	"Name":"Name",
	"PreferredName":"PreferredName",
	"Mail":"Mail",
	"MailVerified":"MailVerified",
	"Groups":map[string]interface{}{
		"Permissions":map[string]interface{}{
			"AccessRight":"AccessRight",
		},
	},
}

var fieldsListUserOffers = map[string]interface{}{
	"Name":"Name",
	"Location":"Location",
	"Tags":map[string]interface{}{
		"Name":"Name",
	},
}

var fieldsListUserRequests = map[string]interface{}{
	"Name":"Name",
	"Location":"Location",
	"Tags":map[string]interface{}{
		"Name":"Name",
	},
}




func (app *App) GetUser(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var obj db.User
	app.DB.Preload("Groups").Preload("Groups.Permissions").First(&obj, "id = ?", User.ID)
	response := db.CopyNestedModel(obj, fieldsGetUser)
	c.JSON(http.StatusOK, response)
}

func (app *App) UpdateUser(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Change stub to real function.
	c.JSON(http.StatusOK, gin.H{
		"ID":            fmt.Sprintf("%s", uuid.NewV4()),
		"Name":          "Updated Bernd",
		"PreferredName": "Da Börnd",
		"Mail":          "esistdermomentgekommen@mail.com",
		"Groups": struct {
			Location    interface{}
			Permissions interface{}
		}{
			struct {
				Longitude float64
				Latitude  float64
			}{
				13.5,
				50.2,
			},
			struct {
				AccessRight string
				Description string
			}{
				"user",
				"This permission represents a standard, registered but not privileged user in our system.",
			},
		},
	})
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


	response := make([]interface{}, 0)
	for _, o := range Offers {
		// 2) Check expired field - extra argument for that?
		if o.ValidityPeriod.After(time.Now()) {

			// 3) Only return what's needed
			model := db.CopyNestedModel(o, fieldsListUserOffers)
			response = append(response, model)
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

	response := make([]interface{}, 0)
	for _, r := range Requests {
		// 2) Check expired field - extra argument for that?
		if r.ValidityPeriod.After(time.Now()) {
			// 3) Only return what's needed
			model := db.CopyNestedModel(r, fieldsListUserRequests)
			response = append(response, model)
		}
	}


	c.JSON(http.StatusOK, response)
}
