package main

import (
	"fmt"
	"log"

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

	// Instantiate final response slice.
	response := make([]interface{}, len(Notifications))

	// Range over all notifications and enrich with
	// additional objects if necessary.
	for i, notification := range Notifications {

		jsonNotificationTmp := CopyNestedModel(notification, fieldsNotification)
		jsonNotification, ok := jsonNotificationTmp.(map[string]interface{})
		if !ok {
			log.Println("[ListNotifications] Type assertion of jsonNotificationTmp went wrong.")
			return
		}

		if notification.Type == db.NotificationMatching {

			// Find matching element and connected elements.
			var Matching db.Matching
			app.DB.First(&Matching, "\"id\" = ?", notification.ItemID)
			app.DB.Model(&Matching).Related(&Matching.Offer)
			app.DB.Model(&Matching.Offer).Related(&Matching.Offer.User)
			app.DB.Model(&Matching).Related(&Matching.Request)
			app.DB.Model(&Matching.Request).Related(&Matching.Request.User)

			// Marshal compiled matching model.
			jsonMatchingTmp := CopyNestedModel(Matching, fieldsMatching)
			jsonMatching, ok := jsonMatchingTmp.(map[string]interface{})
			if !ok {
				log.Println("[ListNotifications] Type assertion of jsonMatchingTmp went wrong.")
				return
			}

			// Append marshalled matching to response JSON.
			jsonNotification["Matching"] = jsonMatching
		}

		response[i] = jsonNotification
	}

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

	// Retrieve notificationID from request URL.
	notificationID := app.getUUID(c, "notificationID")
	if notificationID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "notificationID is no valid UUID",
		})

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

	// Validate sent data for updating a notification.
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

	response := CopyNestedModel(Notification, fieldsNotificationWithRead)

	c.JSON(http.StatusOK, response)
}
