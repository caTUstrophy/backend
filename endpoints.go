package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/caTUstrophy/backend/db"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/leebenson/conform"
	"github.com/patrickmn/go-cache"
	"golang.org/x/crypto/bcrypt"
)

// Structs

type CreateUserPayload struct {
	Name          string `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	PreferredName string `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Mail          string `conform:"trim,email" validate:"required,email"`
	Password      string `validate:"required,min=16,containsany=0123456789,containsany=!@#$%^&*()_+-=:;?/0x2C0x7C"`
}

type LoginPayload struct {
	Mail     string `conform:"trim,email" validate:"required,email"`
	Password string `validate:"required"`
}

// Functions

// Authorization checking

func Authorize(req *http.Request, scope string) (bool, string) {

	// Extract JWT from request headers.

	// Check if JWT is valid.

	// Check if an entry with mail from JWT exists in our session store.

	// Check if JWT from request matches JWT from session store.

	// Return success.
	return true, ""
}

// Endpoint handlers

func (app *App) CreateUser(c *gin.Context) {

	var Payload CreateUserPayload

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
	var CountDup int
	app.DB.Model(&db.User{}).Where("mail = ?", Payload.Mail).Count(&CountDup)

	if CountDup > 0 {

		// Signal client that this mail is already in use.
		c.JSON(400, gin.H{
			"Mail": "Already exists",
		})

		return
	}

	var User db.User

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
	app.DB.First(&DefaultGroup, "default_group = ?", true)

	// Add user to default user group and enable the user.
	User.Groups = []db.Group{DefaultGroup}
	User.Enabled = true

	// Create user object in database.
	app.DB.Create(&User)

	// On success: return ID of newly created user.
	c.JSON(201, gin.H{
		"ID": User.ID,
	})
}

func (app *App) Login(c *gin.Context) {

	var Payload LoginPayload

	// Expect login struct fields in JSON request body.
	c.BindJSON(&Payload)

	// Validate sent user login data.
	conform.Strings(&Payload)
	errs := app.Validator.Struct(&Payload)

	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp[err.Field] = "Is required"
			} else if err.Tag == "email" {
				errResp[err.Field] = "Is not a valid mail address"
			}
		}

		// Send prepared error message to client.
		c.JSON(400, errResp)

		return
	}

	// Find user in database.
	var User db.User
	app.DB.First(&User, "mail = ?", Payload.Mail)

	// Check if user is not known to our system.
	if User.Mail == "" {
		User.PasswordHash = ""
	}

	// Compare password hash from database with possible plaintext
	// password from request. Compares in constant time.
	err := bcrypt.CompareHashAndPassword([]byte(User.PasswordHash), []byte(Payload.Password))
	if err != nil {

		// Signal client that an error occured.
		c.JSON(400, gin.H{
			"Error": "Mail or password is wrong",
		})

		return
	}

	// Retrieve the session signing key from environment.
	jwtSigningSecret := os.Getenv("JWT_SIGNING_SECRET")

	nowTime := time.Now()
	expTime := nowTime.Add((app.SessionValidFor - (30 * time.Second))).Unix()

	// At this point, the user exists and provided a correct password.
	// Create a JWT with claims to identify user.
	sessionJWT := jwt.New(jwt.SigningMethodHS512)

	// Add these claims.
	// TODO: Add important claims for security!
	//       Hash(PasswordHash)? Needs database call, which is what we want to avoid.
	sessionJWT.Claims["iss"] = User.Mail
	sessionJWT.Claims["iat"] = nowTime.Unix()
	sessionJWT.Claims["nbf"] = nowTime.Add((-1 * time.Minute)).Unix()
	sessionJWT.Claims["exp"] = expTime

	sessionJWTString, err := sessionJWT.SignedString([]byte(jwtSigningSecret))
	if err != nil {
		log.Fatalf("[Login] Signing of user session JWT went wrong: %s.\nTerminating.", err)
	}

	// Add JWT to session in-memory cache.
	app.Sessions.Set(User.Mail, sessionJWTString, cache.DefaultExpiration)

	// Deliver JWT to client that made the request.
	c.JSON(200, gin.H{
		"AccessToken": sessionJWTString,
		"ExpiresIn":   expTime,
	})
}

func (app *App) RenewToken(c *gin.Context) {

	// Example usage of authorization function.
	if ok, message := Authorize(c.Request, "user"); !ok {

		// Signal client an error and expect authorization.
		c.JSON(401, message)

		return
	}
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
