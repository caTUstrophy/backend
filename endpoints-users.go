package main

import (
	"fmt"
	"log"

	"net/http"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Structs

type CreateUserPayload struct {
	Name          string `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	PreferredName string `conform:"trim" validate:"excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Mail          string `conform:"trim,email" validate:"required,email"`
	Password      string `validate:"required,min=16,containsany=0123456789,containsany=!@#$%^&*()_+-=:;?/0x2C0x7C"`
}

// Endpoint handler functions

func (app *App) CreateUser(c *gin.Context) {

	var Payload CreateUserPayload

	// Expect user struct fields in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Supplied values in JSON body could not be parsed",
		})

		return
	}

	// Validate sent user registration data.
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
			} else if err.Tag == "min" {
				errResp[err.Field] = "Is too short"
			} else if err.Tag == "containsany" {
				errResp[err.Field] = "Does not contain numbers and special characters"
			} else if err.Tag == "email" {
				errResp[err.Field] = "Is not a valid mail address"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	// Check for user duplicate attempt: entry with mail exists?
	var CountDup int
	app.DB.Model(&db.User{}).Where("mail = ?", Payload.Mail).Count(&CountDup)

	if CountDup > 0 {

		// Signal client that this mail is already in use.
		c.JSON(http.StatusBadRequest, gin.H{
			"Mail": "Already exists",
		})

		return
	}

	var User db.User

	User.ID = fmt.Sprintf("%s", uuid.NewV4())
	User.Name = Payload.Name
	User.PreferredName = Payload.PreferredName
	User.Mail = Payload.Mail
	User.MailVerified = false

	// Password hash generation.
	hash, err := bcrypt.GenerateFromPassword([]byte(Payload.Password), app.HashCost)
	User.PasswordHash = string(hash)
	if err != nil {
		// If there was an error during hash creation - terminate immediately.
		log.Fatal("[CreateUser] Error while generating hash in user creation. Terminating.")
	}

	var DefaultGroup db.Group
	app.DB.Preload("Permissions").First(&DefaultGroup, "default_group = ?", true)

	// Add user to default user group and enable the user.
	User.Groups = []db.Group{DefaultGroup}
	User.Enabled = true

	// Create user object in database.
	app.DB.Create(&User)

	// On success: return ID of newly created user.
	c.JSON(http.StatusCreated, gin.H{
		"ID":            User.ID,
		"Name":          User.Name,
		"PreferredName": User.PreferredName,
		"Mail":          User.Mail,
		"Groups":        User.Groups,
	})
}

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
