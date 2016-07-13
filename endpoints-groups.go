package main

import (
	"fmt"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
)

// Functions

func (app *App) GetGroupObject(groupID string) db.Group {

	var Group db.Group

	app.DB.Preload("Users").First(&Group, "id = ?", groupID)
	app.DB.Model(&Group).Related(&Group.Region)

	return Group
}

func (app *App) GetGroups(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok = app.CheckScope(User, db.Region{}, "superadmin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	var Groups []db.Group
	app.DB.Preload("Users").Find(&Groups)

	models := make([]map[string]interface{}, len(Groups))

	// Iterate over all groups in database return and marshal it.
	for i, group := range Groups {

		// TODO: Resolve region in smarter way.
		if group.RegionId != "" {

			var Region db.Region
			app.DB.First(&Region, "id = ?", group.RegionId)
			group.Region = Region
		}

		models[i] = CopyNestedModel(group, fieldsGroup).(map[string]interface{})
	}

	c.JSON(http.StatusOK, models)
}

func (app *App) ListSystemAdmins(c *gin.Context) {

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

	var group db.Group
	app.DB.Preload("Users").First(&group, "access_right = ?", "superadmin")

	model := CopyNestedModel(group.Users, fieldsSystemAdmin)

	c.JSON(http.StatusOK, model)
}
