package main

import (
	"fmt"
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
	PreferredName string `conform:"trim" validate:"excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Mail          string `conform:"trim,email" validate:"required,email"`
	Password      string `validate:"required,min=16,containsany=0123456789,containsany=!@#$%^&*()_+-=:;?/0x2C0x7C"`
}

type LoginPayload struct {
	Mail     string `conform:"trim,email" validate:"required,email"`
	Password string `validate:"required"`
}

type CreateRequestPayload struct {
	Name           string   `conform:"trim" validate:"required"`
	Location       string   `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Tags           []string `conform:"trim" validate:"dive,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	ValidityPeriod int64    `conform:"trim" validate:"required"`
}

type CreateOfferPayload struct {
	Name           string   `conform:"trim" validate:"required"`
	Location       string   `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Tags           []string `conform:"trim" validate:"dive,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	ValidityPeriod int64    `conform:"trim" validate:"required"`
}

// Functions
func (app *App) Authorize(req *http.Request) (bool, *db.User, string) {
	jwtSigningSecret := []byte(os.Getenv("JWT_SIGNING_SECRET"))

	// Extract JWT from request headers.
	requestJWT, err := jwt.ParseFromRequest(req, func(token *jwt.Token) (interface{}, error) {
		// Verfiy that JWT was signed with correct algorithm.

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("[Authorize] Unexpected signing method: %v.", token.Header["alg"])
		}

		// Return our JWT signing secret as the key to verify integrity of JWT.
		return jwtSigningSecret, nil
	})


	// Check if JWT is valid.
	if err != nil {

		// Check every useful error variant.
		if validationError, ok := err.(*jwt.ValidationError); ok {

			if (validationError.Errors & (jwt.ValidationErrorExpired | jwt.ValidationErrorNotValidYet)) != 0 {
				// JWT is not yet valid or expired.
				return false, nil, "JWT not yet valid or expired"
			} else {
				// Invalid JWT was delivered.
				return false, nil, "JWT was invalid"
			}
		} else {
			// Something went wrong.
			return false, nil, "JWT was invalid"
		}
	}

	// Check if an entry with mail from JWT exists in our session store.
	email := requestJWT.Claims["iss"].(string)
	sessionJWTInterface, found := app.Sessions.Get(email)
	if !found {
		return false, nil, "JWT was invalid"
	}

	// Check if JWT from request matches JWT from session store.
	if sessionJWTInterface.(string) != requestJWT.Raw {
		return false, nil, "JWT was invalid"
	}

	// Retrieve user from database.
	var User db.User
	app.DB.First(&User, "mail = ?", email)

	return true, &User, ""
}

func (app *App) CheckScope(user *db.User, location string, permission string) bool {

	// Check if User.Groups contains a group with location.
	// * No -> false
	// * Yes -> Has this group the necessary permission?

	// Fast, because the typical user is member of few groups.
	for _, group := range user.Groups {

		if group.Location == location {

			// Fast, because there are not so many different permissions.
			for _, groupPermission := range group.Permissions {

				if groupPermission.AccessRight == permission {
					return true
				}
			}
		}
	}

	// No group found that gives permission to user.
	return false
}

func (app *App) makeToken(c *gin.Context, user *db.User) (string, int64) {

	// Retrieve the session signing key from environment.
	jwtSigningSecret := os.Getenv("JWT_SIGNING_SECRET")

	// Save current timestamp.
	nowTime := time.Now()

	// At this point, the user exists and provided a correct password.
	// Create a JWT with claims to identify user.
	sessionJWT := jwt.New(jwt.SigningMethodHS512)

	// Add these claims.
	// TODO: Add important claims for security!
	//       Hash(PasswordHash)? Needs database call, which is what we want to avoid.
	sessionJWT.Claims["iss"] = user.Mail
	sessionJWT.Claims["iat"] = nowTime.Unix()
	sessionJWT.Claims["nbf"] = nowTime.Add((-1 * time.Minute)).Unix()
	sessionJWT.Claims["exp"] = nowTime.Add(app.SessionValidFor).Unix()

	sessionJWTString, err := sessionJWT.SignedString([]byte(jwtSigningSecret))
	if err != nil {
		log.Fatalf("[makeToken] Creating JWT went wrong: %s.\nTerminating.", err)
	}

	// Add JWT to session in-memory cache.
	app.Sessions.Set(user.Mail, sessionJWTString, cache.DefaultExpiration)

	return sessionJWTString, nowTime.Add((app.SessionValidFor - (30 * time.Second))).Unix()
}

// Endpoint handlers

func (app *App) CreateUser(c *gin.Context) {

	var Payload CreateUserPayload

	// Expect user struct fields in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(400, gin.H{
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
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(400, gin.H{
			"Error": "Supplied values in JSON body could not be parsed",
		})

		return
	}

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
	err = bcrypt.CompareHashAndPassword([]byte(User.PasswordHash), []byte(Payload.Password))
	if err != nil {

		// Signal client that an error occured.
		c.JSON(400, gin.H{
			"Error": "Mail and/or password is wrong",
		})

		return
	}

	// Create session JWT and expiration time of JWT.
	sessionJWTString, sessionExpTime := app.makeToken(c, &User)

	// Deliver JWT to client that made the request.
	c.JSON(200, gin.H{
		"AccessToken": sessionJWTString,
		"ExpiresIn":   sessionExpTime,
	})
}

func (app *App) RenewToken(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(401)

		return
	}

	// Create session JWT and expiration time of JWT.
	sessionJWTString, sessionExpTime := app.makeToken(c, User)

	// Deliver JWT to client that made the request.
	c.JSON(200, gin.H{
		"AccessToken": sessionJWTString,
		"ExpiresIn":   sessionExpTime,
	})
}

func (app *App) Logout(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(401)
		return
	}

	// Remove user's JWT from session store.
	app.Sessions.Delete(User.Mail)

	c.Status(200)

	return

}

func (app *App) ListOffers(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(401)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	// if ok := CheckScope(User, "worldwide", "admin"); !ok {
	//     c.Status(401)
	// }

	var Offers []db.Offer

	// Retrieve all offers from database.
	app.DB.Find(&Offers)

	// Send back results to client.
	c.JSON(200, Offers)
}

func (app *App) ListRequests(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(401)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	// if ok := CheckScope(User, "worldwide", "admin"); !ok {
	//     c.Status(401)
	// }

	var Requests []db.Request

	// Retrieve all requests from database.
	app.DB.Find(&Requests)

	// Send back results to client.
	c.JSON(200, Requests)
}

func (app *App) CreateOffer(c *gin.Context) {
	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {
		fmt.Printf(message + "\n")
		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(401)
		c.JSON(401, gin.H{ "message": "jwt invalid", })
		return
	}

	var Payload CreateOfferPayload

	// Expect offer struct fields for creation in JSON request body.
	c.BindJSON(&Payload)


	// Validate sent request creation data.
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
			} else if err.Tag == "dive" {
				errResp[err.Field] = "Needs to be an array"
			}
		}

		// Send prepared error message to client.
		c.JSON(400, errResp)

		return
	}


	var Offer db.Offer

	// Set insert struct to values from payload.
	Offer.Name = Payload.Name
	Offer.User = *User
	Offer.Location = Payload.Location
	Offer.Tags = make([]db.Tag, 0)

	fmt.Printf(Offer.Name)
	fmt.Printf(Offer.Location)

	// If tags were supplied, check if they exist in our system.
	log.Printf("Do tags exist in payload?", Payload.Tags)
	if len(Payload.Tags) > 0 {

		var TagExists int
		allTagsExist := true

		for t := range Payload.Tags {

			var Tag db.Tag

			// Count number of results for query of name of tags.
			app.DB.Where("name = ?", t).First(&Tag).Count(&TagExists)

			// Set flag to false, if one tag was not found.
			if TagExists <= 0 {
				allTagsExist = false
			} else {
				Offer.Tags = append(Offer.Tags, Tag)
			}
		}

		// If at least one of the tags does not exist - return error.
		if !allTagsExist {
			c.JSON(400, gin.H{
				"Tags": "One or multiple tags do not exist",
			})

			return
		}
	}

	// Check if validity period is yet to come.
	if Payload.ValidityPeriod <= time.Now().Unix() {
		c.JSON(400, gin.H{
			"ValidityPeriod": "Request has to be valid until a date in the future",
		})

		return
	} else {
		Offer.ValidityPeriod = Payload.ValidityPeriod
	}

	Offer.Expired = false

	// Save request to database.
	app.DB.Create(&Offer)
	fmt.Printf(" -> Created new offer")

	// Signal success to client.
	c.Status(200)
}

func (app *App) CreateRequest(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(401)

		return
	}

	var Payload CreateRequestPayload

	// Expect request struct fields for creation in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		// Check if error was caused by failed unmarshalling string -> []string.
		if err.Error() == "json: cannot unmarshal string into Go value of type []string" {

			c.JSON(400, gin.H{
				"Tags": "Provide an array, not a string",
			})

			return
		}
	}

	// Validate sent request creation data.
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
		c.JSON(400, errResp)

		return
	}

	var Request db.Request

	// Set insert struct to values from payload.
	Request.Name = Payload.Name
	Request.User = *User
	Request.Location = Payload.Location
	Request.Tags = make([]db.Tag, 0)

	// If tags were supplied, check if they exist in our system.
	if len(Payload.Tags) > 0 {

		allTagsExist := true

		for _, tag := range Payload.Tags {

			var Tag db.Tag

			// Count number of results for query of name of tags.
			app.DB.First(&Tag, "name = ?", tag)

			// Set flag to false, if one tag was not found.
			if Tag.Name == "" {
				allTagsExist = false
			} else {
				Request.Tags = append(Request.Tags, Tag)
			}
		}

		// If at least one of the tags does not exist - return error.
		if !allTagsExist {

			c.JSON(400, gin.H{
				"Tags": "One or multiple tags do not exist",
			})

			return
		}
	} else {
		Request.Tags = nil
	}

	// Check if validity period is yet to come.
	if Payload.ValidityPeriod <= time.Now().Unix() {

		c.JSON(400, gin.H{
			"ValidityPeriod": "Request has to be valid until a date in the future",
		})

		return
	} else {
		Request.ValidityPeriod = Payload.ValidityPeriod
		Request.Expired = false
	}

	// Save request to database.
	app.DB.Create(&Request)

	// Signal success to client.
	c.Status(201)
}

func (app *App) CreateMatching(c *gin.Context) {

	// Check authorization for this function.
	// Authorize()
}

func (app *App) GetMatching(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(401)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	// if ok := CheckScope(User, "worldwide", "admin"); !ok {
	//     c.Status(401)
	// }

	var Matching db.Matching

	matchingID := c.Params.ByName("matchingID")

	// Retrieve all requests from database.
	app.DB.First(&Matching, "id = ?", matchingID)

	// Send back results to client.
	c.JSON(200, Matching)
}
