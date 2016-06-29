package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
)

func (app *App) PromoteToSystemAdmin(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, db.Region{}, "superadmin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// Parse the JSON and check for errors
	var Payload PromoteUserPayload

	// Expect offer struct fields for creation in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Couldn't marshal JSON",
		})

		return
	}

	// Validate sent offer creation data.
	conform.Strings(&Payload)
	errs := app.Validator.Struct(&Payload)

	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp[err.Field] = "Is required"
			} else if err.Tag == "excludesall" {
				errResp[err.Field] = "Contains unallowed characters"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	// Everything seems fine, promote that user.
	var group db.Group
	app.DB.First(&group, "access_right = ?", "superadmin")

	// Find the user who is to be promoted and add the group to his or her groups.
	var promotedUser db.User
	app.DB.Preload("Groups").First(&promotedUser, "mail = ?", Payload.Mail)

	if promotedUser.Mail != Payload.Mail {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Email unkown to system",
		})

		return
	}

	promotedUser.Groups = append(promotedUser.Groups, group)
	// app.DB.Model(&promotedUser).Updates(db.User{Groups: promotedUser.Groups})
	app.DB.Save(promotedUser)

	model := CopyNestedModel(promotedUser, fieldsUser)

	c.JSON(http.StatusOK, model)
}
