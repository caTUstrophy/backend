package main

import (
	"fmt"
	"time"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

func (app *App) GetUser(c *gin.Context) {

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
		"Name":          "Bernd",
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
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Change this stub to real function.
	// 1) Only retrieve offers from user.
	// 2) Check expired field - extra argument for that?
	// 3) Only return what's needed.

	type TmpLocation struct {
		Longitude float64
		Latitude  float64
	}

	type TmpTag struct {
		Name string
	}

	type TmpUserOffer struct {
		ID             string
		Name           string
		Location       TmpLocation
		Tags           []TmpTag
		ValidityPeriod string
	}

	TmpResponse := []TmpUserOffer{
		TmpUserOffer{
			fmt.Sprintf("%s", uuid.NewV4()),
			"Offering bread",
			TmpLocation{
				12.7,
				51.0,
			},
			[]TmpTag{
				TmpTag{
					"Food",
				},
			},
			time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, TmpResponse)
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

	// TODO: Change this stub to real function.
	// replace tmp location - once postgis branch merged

	type TmpLocation struct {
		Longitude float64
		Latitude  float64
	}

	type TmpUserRequest struct {
		ID             string
		Name           string
		Location       TmpLocation
		Tags           []db.Tag // IMO return []string instead, no need to know db.Tag.ID
		ValidityPeriod string
	}

	// 1) Only retrieve requests from user.
	var Requests []db.Request
	app.DB.Find(&Requests, "user_id = ?", User.ID)

	var ReturnRequests []TmpUserRequest
	ReturnRequests = make([]TmpUserRequest, 0)
	for _, r := range Requests {

		// 2) Check expired field - extra argument for that?
		if r.ValidityPeriod.After(time.Now()) {
			// find associated tags
			app.DB.Model(&r).Association("Tags").Find(&r.Tags)

			// 3) Only return what's needed
			// TODO : replace through proper payload struct
			req := TmpUserRequest{
				r.ID,
				r.Name,
				TmpLocation{
					13.9,
					50.1,
				},
				r.Tags,
				r.ValidityPeriod.Format(time.RFC3339),
			}

			ReturnRequests = append(ReturnRequests, req)
		}
	}

	c.JSON(http.StatusOK, ReturnRequests)
}
