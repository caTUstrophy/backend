package main

import (
	"flag"
	"log"
	"time"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

func InitAndConfig() *App {

	// Define an initialization flag.
	initFlag := flag.Bool("init", false, "Set this flag to true to initialize a fresh database with default data.")
	flag.Parse()

	// Load .env configuration files.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[InitAndConfig] Failed to load .env files. Terminating.")
	}

	// Make space for a application struct containing our global context.
	app := new(App)

	// Set cost factor of bcrypt password hashing to 10.
	// TODO: On production, set this to 16.
	app.HashCost = 10

	// Set JWT session token validity to 30 minutes.
	app.SessionValidFor = 30 * time.Minute

	// Create new gin instance with default functionalities and add it to app struct.
	app.Router = gin.Default()

	// Open connection to database and insert middleware.
	app.DB = db.InitDB("sqlite3", "caTUstrophy.sqlite")

	// If init flag was set to true, add default data to database.
	if *initFlag {
		db.AddDefaultData(app.DB)
	}

	// Instantiate a new go-cache instance to hold the JWTs of user sessions.
	app.Sessions = cache.New(app.SessionValidFor, 10*time.Second)

	// Initialize the validator instance to validate fields with tag 'validate'
	validatorConfig := &validator.Config{TagName: "validate"}
	app.Validator = validator.New(validatorConfig)

	return app
}
