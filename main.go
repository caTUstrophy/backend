package main

import (
	"flag"

	"github.com/caTUstrophy/backend/cache"
	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/jinzhu/gorm"
)

// Structs

type App struct {
	Router    *gin.Engine
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

func InitAndConfig() *App {

	// Define an initialization flag.
	initFlag := flag.Bool("init", false, "Set this flag to true to initialize a fresh database with default data.")
	flag.Parse()

	// Make space for a application struct containing our global context.
	app := new(App)

	// Open connection to database and insert middleware.
	app.DB = db.InitDB("sqlite3", "caTUstrophy.sqlite")

	// If init flag was set to true, add default data to database.
	if *initFlag {
		db.AddDefaultData(app.DB)
	}

	// Initialize the validator instance to validate fields with tag 'validate'
	validatorConfig := &validator.Config{TagName: "validate"}
	app.Validator = validator.New(validatorConfig)

	// Create new gin instance with default functionalities and add it to app struct.
	router := gin.Default()
	app.Router = router

	return app
}

func CORSMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Next()
	}
}

// Main

func main() {

	// Parse command line flags and build application config.
	app := InitAndConfig()

	// Initialize cache for jwts.
	jwts := cache.MapCache{make(map[string]bool)}
	jwts.Set("test")

	// Add custom middleware to call stack.
	app.Router.Use(CORSMiddleware())

	// Define endpoint to handler mapping
	app.Router.POST("/users", app.CreateUser)
	app.Router.POST("/auth", app.Login)
	app.Router.GET("/auth", app.RenewToken)
	app.Router.DELETE("/auth", app.Logout)
	app.Router.GET("/offers", app.ListOffers)
	app.Router.GET("/requests", app.ListRequests)
	app.Router.POST("/requests", app.CreateRequest)
	app.Router.POST("/matchings", app.CreateMatching)
	app.Router.GET("/matchings/:matchingID", app.GetMatching)

	// Run our application.
	app.Router.Run(":3001")
}
