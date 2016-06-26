package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
)

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
	// TODO: Implement this function.
}
