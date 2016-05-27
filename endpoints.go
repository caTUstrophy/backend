package main

import (
	"log"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"golang.org/x/crypto/bcrypt"
)

// Endpoints

func (app *App) CreateUser(c *gin.Context) {

	var Payload CreateUserPayload
	var CountDup int
	var User db.User

	// Expect user struct fields in JSON request body.
	c.BindJSON(&Payload)

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
		c.JSON(400, errResp)

		return
	}

	// Check for user duplicate attempt: entry with mail exists?
	app.DB.Model(&db.User{}).Where("mail = ?", Payload.Mail).Count(&CountDup)

	if CountDup > 0 {

		// Signal client that this mail is already in use.
		c.JSON(400, gin.H{
			"Mail": "Already exists",
		})

		return
	}

	User.Name = Payload.Name
	User.PreferredName = Payload.PreferredName
	User.Mail = Payload.Mail
	User.MailVerified = false

	// Password hash generation.
	// TODO: Cost currently set to 10, in production increase to 16.
	hash, err := bcrypt.GenerateFromPassword([]byte(Payload.Password), bcrypt.DefaultCost)
	User.PasswordHash = string(hash)

	if err != nil {
		// If there was an error during hash creation - terminate immediately.
		log.Fatal("[CreateUser] Error while generating hash in user creation. Terminating.")
	}

	// User.Groups =
	User.Enabled = true

	// Create user object in database.
	app.DB.Create(&User)

	// On success: return ID of newly created user.
	c.JSON(201, gin.H{
		"ID": User.ID,
	})
}

func (app *App) Login(c *gin.Context) {

}

func (app *App) RenewToken(c *gin.Context) {

}

func (app *App) Logout(c *gin.Context) {

}

func (app *App) ListOffers(c *gin.Context) {

}

func (app *App) ListRequests(c *gin.Context) {

}

func (app *App) CreateRequest(c *gin.Context) {

}

func (app *App) CreateMatching(c *gin.Context) {

}

func (app *App) GetMatching(c *gin.Context) {
	// matchingID := c.Params.ByName("matchingID")
}
