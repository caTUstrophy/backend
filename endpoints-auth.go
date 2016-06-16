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

type LoginPayload struct {
	Mail     string `conform:"trim,email" validate:"required,email"`
	Password string `validate:"required"`
}

// Authorization related functions.

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

func (app *App) CheckScope(user *db.User, location db.Region, permission string) bool {

	// Check if User.Groups contains a group with location.
	// * No  -> false
	// * Yes -> Has this group the necessary permission?

	// Fast, because the typical user is member of few groups.
	for _, group := range user.Groups {

		for _, groupPermission := range group.Permissions {

			if groupPermission.AccessRight == "superadmin" {
				return true
			}
		}
		fmt.Print("location id", location.ID)
		if location.ID == "" { // if someone wants to check only for superadmin without location, he can give an empty location
			return false
		}

		if group.Location.ID == location.ID {

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

// copy pasted the function to
func (app *App) CheckScopes(user *db.User, locations []db.Region, permission string) bool {
	// check for superadmin privilege
	if su := app.CheckScope(user, db.Region{}, "superadmin"); su {
		return true;
	}

	// iterate over regions until region with permission was found
	for _, Region := range locations {
		if ok := app.CheckScope(user, Region, "admin"); ok {
			return true
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
	expTime := nowTime.Add(app.SessionValidFor).Format(time.RFC3339)

	// At this point, the user exists and provided a correct password.
	// Create a JWT with claims to identify user.
	sessionJWT := jwt.New(jwt.SigningMethodHS512)

	// Add these claims.
	// TODO: Add important claims for security!
	//       Hash(PasswordHash)? Needs database call, which is what we want to avoid.
	sessionJWT.Claims["iss"] = user.Mail
	sessionJWT.Claims["iat"] = nowTime.Format(time.RFC3339)
	sessionJWT.Claims["nbf"] = nowTime.Add((-1 * time.Minute)).Format(time.RFC3339)
	sessionJWT.Claims["exp"] = expTime

	sessionJWTString, err := sessionJWT.SignedString([]byte(jwtSigningSecret))
	if err != nil {
		log.Fatalf("[makeToken] Creating JWT went wrong: %s.\nTerminating.", err)
	}

	// Add JWT to session in-memory cache.
	app.Sessions.Set(user.Mail, sessionJWTString, cache.DefaultExpiration)

	return sessionJWTString
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
