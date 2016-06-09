package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"net/http"

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

type CreateMatchingPayload struct {
	Request int `conform:"trim" validate:"required"`
	Offer   int `conform:"trim" validate:"required"`
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
	app.DB.Preload("Groups").Preload("Groups.Permissions").First(&User, "mail = ?", email)

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

func (app *App) makeToken(c *gin.Context, user *db.User) string {

	// Retrieve the session signing key from environment.
	jwtSigningSecret := os.Getenv("JWT_SIGNING_SECRET")

	// Save current timestamp.
	nowTime := time.Now()
	expTime := nowTime.Add(app.SessionValidFor).Unix()

	// At this point, the user exists and provided a correct password.
	// Create a JWT with claims to identify user.
	sessionJWT := jwt.New(jwt.SigningMethodHS512)

	// Add these claims.
	// TODO: Add important claims for security!
	//       Hash(PasswordHash)? Needs database call, which is what we want to avoid.
	sessionJWT.Claims["iss"] = user.Mail
	sessionJWT.Claims["iat"] = nowTime.Unix()
	sessionJWT.Claims["nbf"] = nowTime.Add((-1 * time.Minute)).Unix()
	sessionJWT.Claims["exp"] = expTime

	sessionJWTString, err := sessionJWT.SignedString([]byte(jwtSigningSecret))
	if err != nil {
		log.Fatalf("[makeToken] Creating JWT went wrong: %s.\nTerminating.", err)
	}

	// Add JWT to session in-memory cache.
	app.Sessions.Set(user.Mail, sessionJWTString, cache.DefaultExpiration)

	return sessionJWTString
}

// Endpoint handlers

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
	c.JSON(http.StatusCreated, gin.H{
		"ID":            User.ID,
		"Name":          User.Name,
		"PreferredName": User.PreferredName,
		"Mail":          User.Mail,
		"Groups":        User.Groups,
	})
}

func (app *App) Login(c *gin.Context) {

	var Payload LoginPayload

	// Expect login struct fields in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
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
		c.JSON(http.StatusBadRequest, errResp)

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
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Mail and/or password is wrong",
		})

		return
	}

	// Create session JWT and expiration time of JWT.
	sessionJWTString := app.makeToken(c, &User)

	// Deliver JWT to client that made the request.
	c.JSON(http.StatusOK, gin.H{
		"AccessToken": sessionJWTString,
	})
}

func (app *App) RenewToken(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Create session JWT and expiration time of JWT.
	sessionJWTString := app.makeToken(c, User)

	// Deliver JWT to client that made the request.
	c.JSON(http.StatusOK, gin.H{
		"AccessToken": sessionJWTString,
	})
}

func (app *App) Logout(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Remove user's JWT from session store.
	app.Sessions.Delete(User.Mail)

	// Signal client success and return ID of logged out user.
	c.JSON(http.StatusOK, gin.H{
		"ID": User.ID,
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
		"Name":          "Bernd",
		"PreferredName": "Da Börnd",
		"Mail":          "esistdermomentgekommen@mail.com",
		"Groups": struct {
			Location    interface{}
			Permissions interface{}
		}{
			struct {
				lon float32
				lat float32
			}{
				float32(13.5),
				float32(50.2),
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
		"Name":          "Updated Bernd",
		"PreferredName": "Da Börnd",
		"Mail":          "esistdermomentgekommen@mail.com",
		"Groups": struct {
			Location    interface{}
			Permissions interface{}
		}{
			struct {
				lon float32
				lat float32
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

func (app *App) ListOffers(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, "worldwide", "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	region := c.Params.ByName("region")

	// TODO: Validate region!

	var Offers []db.Offer

	// Retrieve all offers from database.
	app.DB.Find(&Offers, "Location = ?", region)

	// TODO remove loop and exchange for preload
	for i := 0; i < len(Offers); i++ {
		app.DB.Select("name, id").First(&Offers[i].User, "mail = ?", User.Mail)
	}

	// Send back results to client.
	c.JSON(http.StatusOK, Offers)
}

func (app *App) ListUserOffers(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Change this stub to real function.
	// 1) Only retrieve offers from user.
	// 2) Check expired field - extra argument for that?
	// 3) Only return what's needed.
	c.JSON(http.StatusOK, gin.H{
		"Name": "Offering bread",
		"Location": struct {
			lon float32
			lat float32
		}{
			12.7,
			51.0,
		},
		"Tags": struct {
			Name string
		}{
			"Food",
		},
		"ValidityPeriod": time.Now().Format(time.RFC3339),
	})
}

func (app *App) ListRequests(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, "worldwide", "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	region := c.Params.ByName("region")

	// TODO: Validate region!

	var Requests []db.Request

	// Retrieve all requests from database.
	//app.DB.Preload("User").Find(&Requests, "location = ?", region)
	app.DB.Find(&Requests, "location = ?", region)

	// TODO remove loop and exchange for preload
	for i := 0; i < len(Requests); i++ {
		app.DB.Select("name, id").First(&Requests[i].User, "mail = ?", User.Mail)
	}

	// Send back results to client.
	c.JSON(http.StatusOK, Requests)
}

func (app *App) ListUserRequests(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Change this stub to real function.
	// 1) Only retrieve requests from user.
	// 2) Check expired field - extra argument for that?
	// 3) Only return what's needed.
	c.JSON(http.StatusOK, gin.H{
		"Name": "Looking for bread",
		"Location": struct {
			lon float32
			lat float32
		}{
			13.9,
			50.1,
		},
		"Tags": struct {
			Name string
		}{
			"Food",
		},
		"ValidityPeriod": time.Now().Format(time.RFC3339),
	})
}

func (app *App) CreateOffer(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Payload CreateOfferPayload

	// Expect offer struct fields for creation in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		// Check if error was caused by failed unmarshalling string -> []string.
		if err.Error() == "json: cannot unmarshal string into Go value of type []string" {

			c.JSON(http.StatusBadRequest, gin.H{
				"Tags": "Provide an array, not a string",
			})

			return
		}
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

	var Offer db.Offer

	// Set insert struct to values from payload.
	Offer.Name = Payload.Name
	Offer.User = *User
	Offer.Location = Payload.Location
	Offer.Tags = make([]db.Tag, 0)

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
				Offer.Tags = append(Offer.Tags, Tag)
			}
		}

		// If at least one of the tags does not exist - return error.
		if !allTagsExist {

			c.JSON(http.StatusBadRequest, gin.H{
				"Tags": "One or multiple tags do not exist",
			})

			return
		}
	} else {
		Offer.Tags = nil
	}

	// Check if validity period is yet to come.
	if Payload.ValidityPeriod <= time.Now().Unix() {

		c.JSON(http.StatusBadRequest, gin.H{
			"ValidityPeriod": "Offer has to be valid until a date in the future",
		})

		return
	} else {
		Offer.ValidityPeriod = Payload.ValidityPeriod
		Offer.Expired = false
	}

	// Save offer to database.
	app.DB.Create(&Offer)

	// On success: return ID of newly created offer.
	c.JSON(http.StatusCreated, gin.H{
		"ID":             Offer.ID,
		"Name":           Offer.Name,
		"Location":       Offer.Location,
		"Tags":           Offer.Tags,
		"ValidityPeriod": Offer.ValidityPeriod,
	})
}

func (app *App) CreateRequest(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Payload CreateRequestPayload

	// Expect request struct fields for creation in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		// Check if error was caused by failed unmarshalling string -> []string.
		if err.Error() == "json: cannot unmarshal string into Go value of type []string" {

			c.JSON(http.StatusBadRequest, gin.H{
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
		c.JSON(http.StatusBadRequest, errResp)

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

			c.JSON(http.StatusBadRequest, gin.H{
				"Tags": "One or multiple tags do not exist",
			})

			return
		}
	} else {
		Request.Tags = nil
	}

	// Check if validity period is yet to come.
	if Payload.ValidityPeriod <= time.Now().Unix() {

		c.JSON(http.StatusBadRequest, gin.H{
			"ValidityPeriod": "Request has to be valid until a date in the future",
		})

		return
	} else {
		Request.ValidityPeriod = Payload.ValidityPeriod
		Request.Expired = false
	}

	// Save request to database.
	app.DB.Create(&Request)

	// On success: return ID of newly created request.
	c.JSON(http.StatusCreated, gin.H{
		"ID":             Request.ID,
		"Name":           Request.Name,
		"Location":       Request.Location,
		"Tags":           Request.Tags,
		"ValidityPeriod": Request.ValidityPeriod,
	})
}

func (app *App) UpdateUserOffer(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	offerID := c.Params.ByName("offerID")
	log.Println("[UpdateUserOffer] Offer ID was:", offerID)

	// TODO: Validate offerID!

	// TODO: Change this stub to real function.
	c.JSON(http.StatusOK, gin.H{
		"Name": "Offering COMPLETELY NEW bread",
		"Location": struct {
			lon float32
			lat float32
		}{
			15.5,
			45.3,
		},
		"Tags": struct {
			Name string
		}{
			"Food",
		},
		"ValidityPeriod": time.Now().Format(time.RFC3339),
	})
}

func (app *App) UpdateUserRequest(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	requestID := c.Params.ByName("requestID")
	log.Println("[UpdateUserRequest] Request ID was:", requestID)

	// TODO: Validate requestID!

	// TODO: Change this stub to real function.
	c.JSON(http.StatusOK, gin.H{
		"Name": "Looking for COMPLETELY NEW bread",
		"Location": struct {
			lon float32
			lat float32
		}{
			14.0,
			49.9,
		},
		"Tags": struct {
			Name string
		}{
			"Food",
		},
		"ValidityPeriod": time.Now().Format(time.RFC3339),
	})
}

func (app *App) CreateMatching(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, "worldwide", "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	var Payload CreateMatchingPayload

	// Expect user struct fields in JSON request body.
	err := c.BindJSON(&Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
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
			} else if err.Tag == "excludesall" {
				errResp[err.Field] = "Contains unallowed characters"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)

		return
	}

	// check that offer and request do exist
	var CountOffer int
	app.DB.Model(&db.Offer{}).Where("id = ?", Payload.Offer).Count(&CountOffer)
	var CountRequest int
	app.DB.Model(&db.Request{}).Where("id = ?", Payload.Request).Count(&CountRequest)

	if CountOffer == 0 || CountRequest == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"Matching": "Offer / Request doesnt exist",
		})

		return
	}

	// check for matching duplicate
	var CountDup int
	app.DB.Model(&db.Matching{}).Where("offer_id = ? AND request_id = ?", Payload.Offer, Payload.Request).Count(&CountDup)

	if CountDup > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"Matching": "Already exists",
		})

		return
	}

	// get request and offer to resolve foreign key dependencies
	var Offer db.Offer
	app.DB.First(&Offer, "id = ?", Payload.Offer)
	var Request db.Request
	app.DB.First(&Request, "id = ?", Payload.Request)

	// save matching
	var Matching db.Matching
	Matching.OfferId = Payload.Offer
	Matching.Offer = Offer
	Matching.RequestId = Payload.Request
	Matching.Request = Request

	app.DB.Create(&Matching)

	c.JSON(http.StatusCreated, Matching)
}

func (app *App) ListMatchings(c *gin.Context) {

	// Check authorization for this function.
	ok, User, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	// Check if user permissions are sufficient (user is admin).
	if ok := app.CheckScope(User, "worldwide", "admin"); !ok {

		// Signal client that the provided authorization was not sufficient.
		c.Header("WWW-Authenticate", "Bearer realm=\"CaTUstrophy\", error=\"authentication_failed\", error_description=\"Could not authenticate the request\"")
		c.Status(http.StatusUnauthorized)

		return
	}

	// TODO: Change this stub to real function.
	// 1) Check for expired fields in offers and requests - via extra argument?
	c.JSON(http.StatusOK, gin.H{
		"Offer": struct {
			ID             string
			Name           string
			User           interface{}
			Location       interface{}
			Tags           interface{}
			ValidityPeriod string
		}{
			"a-b-c-d",
			"Offering bread",
			struct {
				ID string
			}{
				"1-2-3-4",
			},
			struct {
				lon float32
				lat float32
			}{
				13.9,
				50.1,
			},
			struct {
				Name string
			}{
				"Food",
			},
			time.Now().Format(time.RFC3339),
		},
		"Request": struct {
			ID             string
			Name           string
			User           interface{}
			Location       interface{}
			Tags           interface{}
			ValidityPeriod string
		}{
			"9-d-2-c",
			"Looking for bread",
			struct {
				ID string
			}{
				"u-x-y-z",
			},
			struct {
				lon float32
				lat float32
			}{
				13.9,
				50.1,
			},
			struct {
				Name string
			}{
				"Food",
			},
			time.Now().Format(time.RFC3339),
		},
	})
}

func (app *App) GetMatching(c *gin.Context) {

	// Check authorization for this function.
	ok, _, message := app.Authorize(c.Request)
	if !ok {

		// Signal client an error and expect authorization.
		c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"CaTUstrophy\", error=\"invalid_token\", error_description=\"%s\"", message))
		c.Status(http.StatusUnauthorized)

		return
	}

	var Matching db.Matching

	matchingID := c.Params.ByName("matchingID")

	// TODO: Validate matchingID!

	// Retrieve all requests from database.
	app.DB.First(&Matching, "id = ?", matchingID)

	// Send back results to client.
	c.JSON(http.StatusOK, Matching)
}
