package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
)

// Functions

func (app *App) GetTags(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Retrieve all currently available tags from database.
	var Tags []db.Tag
	app.DB.Find(&Tags)

	// Only return defined fields in JSON.
	model := CopyNestedModel(Tags, fieldsTag)

	c.JSON(http.StatusOK, model)
}
