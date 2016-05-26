package main

import (
	"github.com/caTUstrophy/backend/cache"
	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/jinzhu/gorm"
)

// Structs

type App struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

type CreateUserPayload struct {
	Name          string `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	PreferredName string `conform:"trim" validate:"required,excludesall=!@#$%^&*()_+-=:;?/0x2C0x7C"`
	Mail          string `conform:"trim,email" validate:"required,email"`
	Password      string `validate:"required,min=16,containsany=0123456789,containsany=!@#$%^&*()_+-=:;?/0x2C0x7C"`
}

// Functions

func CORSMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Next()
	}
}

// Main

func main() {

	// Create new gin instance with default functionalities.
	router := gin.Default()

	// Make space for a application struct containing our global context.
	app := new(App)

	// Open connection to database and insert middleware.
	app.DB = db.InitDB("sqlite3", "caTUstrophy.sqlite")

	// Initialize the validator instance to validate fields with tag 'validate'
	validatorConfig := &validator.Config{TagName: "validate"}
	app.Validator = validator.New(validatorConfig)

	// Initialize cache for jwts
	jwts := cache.MapCache{make(map[string]bool)}
	jwts.Set("test")

	// Add custom middleware to call stack.
	router.Use(CORSMiddleware())

	// Define endpoint to handler mapping
	router.POST("/users", app.CreateUser)
	router.POST("/auth", app.Login)
	router.GET("/auth", app.RenewToken)
	router.DELETE("/auth", app.Logout)
	router.GET("/offers", app.ListOffers)
	router.GET("/requests", app.ListRequests)
	router.POST("/requests", app.CreateRequest)
	router.POST("/matchings", app.CreateMatching)
	router.GET("/matchings/:matchingID", app.GetMatching)

	// Run our application.
	router.Run(":3001")
}
