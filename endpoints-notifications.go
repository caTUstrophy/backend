package main

import (
	"fmt"
	"log"
	"time"

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

// This function usually runs in the background of the backend
// service and deletes all read notifications that were created
// longer than a supplied offset ago.
func (app *App) NotificationReaper(expiryOffset time.Duration, sleepOffset time.Duration) {

	log.Println("Notification reaper started.")

	// Buffer up to 10 items before bulk delete.
	deleteBufferSize := 10

	for {

		// Retrieve all read notifications in ascending order by date.
		var Notifications []db.Notification
		app.DB.Where("\"read\" = ?", "true").Order("\"created_at\" ASC").Find(&Notifications)

		// Initialize slice to buffer to-be-deleted notifications.
		deleteBuffer := make([]string, 0, deleteBufferSize)
		i := 0

		for _, notification := range Notifications {

			// If notification was created more than expiryOffset duration
			// ago, add ID of notification to deletion buffer.
			if notification.CreatedAt.Add(expiryOffset).Before(time.Now()) {

				deleteBuffer = append(deleteBuffer, notification.ID)
				i++
			}

			// If delete buffer is full, issue a bulk deletion.
			if i == deleteBufferSize {

				app.DB.Where("\"id\" IN (?)", deleteBuffer).Delete(&db.Notification{})
				deleteBuffer = make([]string, 0, deleteBufferSize)
				i = 0
			}
		}

		if len(deleteBuffer) > 0 {
			app.DB.Where("\"id\" IN (?)", deleteBuffer).Delete(&db.Notification{})
		}

		log.Println("Notification reaper done. Sleeping for", sleepOffset)

		// Let function execution sleep until next round.
		time.Sleep(sleepOffset)
	}
}
