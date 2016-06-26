package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
)

// Structs

type UpdateNotificationPayload struct {
	Read bool `conform:"trim" validate:"exists"`
}

// Functions

func (app *App) ListNotifications(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Notifications []db.Notification
	app.DB.Find(&Notifications, "\"user_id\" = ? AND \"read\" = ?", User.ID, "false")

	// Marshal only necessary fields.
	response := CopyNestedModel(Notifications, fieldsNotification)

	c.JSON(http.StatusOK, response)
}

func (app *App) UpdateNotification(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Retrieve notification ID from request URL.
	notificationID := app.getUUID(c, "notificationID")
	if notificationID == "" {
		return
	}

	var Payload UpdateNotificationPayload

	// Expect the update fields in request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Supplied values in JSON body could not be parsed",
		})

		return
	}

	// Validate sent request creation data.
	conform.Strings(&Payload)
	errs := app.Validator.Struct(&Payload)

	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "exists" {
				errResp[err.Field] = "Has to be present"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	var Notification db.Notification
	app.DB.First(&Notification, "\"id\" = ?", notificationID)

	// Check if user is actually owner of notification.
	// This conforms to the scope 'C' level of this handler.
	if Notification.UserID != User.ID {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"JWT was invalid\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// If Read flag from request was set to 'false' return an error.
	if Payload.Read != true {

		c.JSON(http.StatusBadRequest, gin.H{
			"Read": "Can not be anything different than 'true'",
		})

		return
	}

	// All checks passed - set notification to read.
	Notification.Read = true
	app.DB.Save(&Notification)

	response := CopyNestedModel(Notification, fieldsNotificationU)

	c.JSON(http.StatusOK, response)
}
