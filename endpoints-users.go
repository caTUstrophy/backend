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

type UpdateUserPayload struct {
	Name          string   `conform:"trim" validate:"excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	PreferredName string   `conform:"trim" validate:"excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Mail          string   `conform:"trim"`
	PhoneNumbers  []string `conform:"trim"`
	Password      string
	Groups        []GroupPayload
}

type GroupPayload struct {
	ID string `conform:"trim" validate:"required,uuid4"`
}

type PasswordPayload struct {
	Password string `validate:"min=16,containsany=0123456789,containsany=!@#$%^&*()_+-=:;?/0x2C0x7C"`
}
type EmailPayload struct {
	Mail string `conform:"trim,email" validate:"email"`
}

type CreateUserPayload struct {
	Name          string   `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	PreferredName string   `conform:"trim" validate:"excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Mail          string   `conform:"trim,email" validate:"required,email"`
	PhoneNumbers  []string `conform:"trim" validate:"required"`
	Password      string   `validate:"required,min=16,containsany=0123456789,containsany=!@#$%^&*()_+-=:;?/0x2C0x7C"`
}

// Functions

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

	jsonPhoneNumbers := new(db.PhoneNumbers)
	err = jsonPhoneNumbers.Scan(Payload.PhoneNumbers)
	if err != nil {

		// Signal client that the supplied phone numbers contained some kind of error.
		c.JSON(http.StatusBadRequest, gin.H{
			"PhoneNumbers": "Are not valid",
		})

		return
	}

	User.PhoneNumbers = *jsonPhoneNumbers

	// Password hash generation.
	hash, err := bcrypt.GenerateFromPassword([]byte(Payload.Password), app.HashCost)
	User.PasswordHash = string(hash)
	if err != nil {
		// If there was an error during hash creation - terminate immediately.
		log.Fatal("[CreateUser] Error while generating hash in user creation. Terminating.")
	}

	var DefaultGroup db.Group
	app.DB.First(&DefaultGroup, "default_group = ?", true)

	// Add user to default user group and enable the user.
	User.Groups = []db.Group{DefaultGroup}
	User.Enabled = true

	// Create user object in database.
	app.DB.Create(&User)

	model := CopyNestedModel(User, fieldsUser)

	c.JSON(http.StatusOK, model)
}

// This function is not thought be used as handler, it updates a given user object with no permission checking
// Used by UpdateMe and UpdateUser
func (app *App) UpdateUserObject(User *db.User, c *gin.Context, updateGroups bool) {

	var Payload UpdateUserPayload

	// Expect user struct fields in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Supplied values in JSON body could not be parsed",
		})

		return
	}

	// Object containing fields to be updated
	var updatedUser db.User

	// Validate sent user registration data.
	conform.Strings(&Payload)
	errs, isErr := app.Validator.Struct(&Payload).(validator.ValidationErrors)
	if isErr {
		CheckErrors(errs, c)
		return
	}
	if Payload.Mail != "" {
		emailPayload := EmailPayload{Payload.Mail}
		errs, isErr = app.Validator.Struct(&emailPayload).(validator.ValidationErrors)
		if isErr {
			CheckErrors(errs, c)
			return
		}
	}
	if Payload.Password != "" {
		passwordPayload := PasswordPayload{Payload.Password}
		errs, isErr = app.Validator.Struct(&passwordPayload).(validator.ValidationErrors)
		if isErr {
			CheckErrors(errs, c)
			return
		}
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(Payload.Password), app.HashCost)
		updatedUser.PasswordHash = string(hash)
		if hashErr != nil {
			// If there was an error during hash creation - terminate immediately.
			log.Fatal("[UpdateUserObject] Error while generating hash in user creation. Terminating.")
		}
	}

	updatedUser.Name = Payload.Name
	updatedUser.PreferredName = Payload.PreferredName
	updatedUser.Mail = Payload.Mail
	updatedUser.MailVerified = false

	if len(Payload.PhoneNumbers) > 0 {
		jsonPhoneNumbers := new(db.PhoneNumbers)
		err = jsonPhoneNumbers.Scan(Payload.PhoneNumbers)
		if err != nil {

			// Signal client that the supplied phone numbers contained some kind of error.
			c.JSON(http.StatusBadRequest, gin.H{
				"PhoneNumbers": "Are not valid",
			})

			return
		}

		updatedUser.PhoneNumbers = *jsonPhoneNumbers
	}

	// Update user
	app.DB.Model(&User).Updates(updatedUser)

	if len(Payload.Groups) > 0 && updateGroups {
		// Load full user to save groups
		app.DB.First(&updatedUser, "id = ?", User.ID)
		// Groups
		updatedUser.Groups = make([]db.Group, len(Payload.Groups))
		for i, gid := range Payload.Groups {
			group := app.GetGroupObject(gid.ID)
			if group.ID == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"Groups": (gid.ID + " does not exist"),
				})
				return
			}
			updatedUser.Groups[i] = group
		}
		// Delete all prior groups
		app.DB.Exec("DELETE FROM user_groups WHERE user_id = ?", updatedUser.ID)
		//app.DB.Delete(&User.Groups)
		// Save all current groups
		//app.DB.Save(&updatedUser.Groups)
		app.DB.Model(&User).Updates(updatedUser)
	}

	// Return updated user
	var checkUser db.User
	app.DB.First(&checkUser, "id = ?", User.ID)
	app.DB.Preload("Groups").First(&checkUser, "id = ?", User.ID)
	for i, _ := range checkUser.Groups {
		app.DB.Model(&checkUser.Groups[i]).Related(&checkUser.Groups[i].Region)
	}
	model := CopyNestedModel(checkUser, fieldsUser)
	c.JSON(http.StatusOK, model)
}

func (app *App) ListUsers(c *gin.Context) {
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

	// The real request

	var Users []db.User
	app.DB.Preload("Groups").Find(&Users)
	for userLoop, _ := range Users {
		for groupLoop, _ := range Users[userLoop].Groups { //groupLoop rhymes
			app.DB.Model(&Users[userLoop].Groups[groupLoop]).Related(&Users[userLoop].Groups[groupLoop].Region)
		}
	}

	model := CopyNestedModel(Users, fieldsUser)

	c.JSON(http.StatusOK, model)
}
func (app *App) GetUser(c *gin.Context) {
	// TODO: Implement this function.
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

	// The real request
	userID := app.getUUID(c, "userID")
	if userID == "" { // Bad request header set by getUUID
		return
	}

	var requestUser db.User
	app.DB.Preload("Groups").First(&requestUser, "id = ?", userID)

	for groupLoop, _ := range requestUser.Groups { //groupLoop rhymes
		app.DB.Model(&requestUser.Groups[groupLoop]).Related(&requestUser.Groups[groupLoop].Region)
	}

	model := CopyNestedModel(requestUser, fieldsUser)

	c.JSON(http.StatusOK, model)
}

func (app *App) UpdateUser(c *gin.Context) {
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

	// The real request

	userID := app.getUUID(c, "userID")
	if userID == "" {
		return
	}

	var updateUser db.User
	app.DB.Preload("Groups").First(&updateUser, "id = ?", userID)

	if updateUser.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"User": "The user you tried to update does not exist.",
		})
		return
	}

	for groupLoop, _ := range updateUser.Groups { //groupLoop rhymes
		app.DB.Model(&updateUser.Groups[groupLoop]).Related(&updateUser.Groups[groupLoop].Region)
	}

	if app.CheckScope(&updateUser, db.Region{}, "superadmin") && updateUser.ID != User.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"Error": "You tried to update a system admin. But as you are equal bosses, you have to respect that your power is limited where the power of the other boss starts.",
		})
		return
	}
	app.UpdateUserObject(&updateUser, c, true)
}
