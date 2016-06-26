package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
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

	// Read IP and port to run on from environment.
	app.IP = os.Getenv("BACKEND_IP")
	app.Port = os.Getenv("BACKEND_PORT")

	// Create new gin instance with default functionalities and add it to app struct.
	app.Router = gin.Default()

	// Open connection to database and insert middleware.
	app.DB = db.InitDB()

	// If init flag was set to true, add default data to database.
	if *initFlag {
		db.AddDefaultData(app.DB)
	}

	// Set cost factor of bcrypt password hashing to the one loaded from environment.
	app.HashCost, err = strconv.Atoi(os.Getenv("PASSWORD_HASHING_COST"))
	if err != nil {
		log.Fatal("[InitAndConfig] Could not load PASSWORD_HASHING_COST from .env file. Missing or not an integer?")
	}

	// Set JWT session token validity to the duration in minutes loaded from environment.
	validFor, err := strconv.Atoi(os.Getenv("JWT_VALID_FOR"))
	if err != nil {
		log.Fatal("[InitAndConfig] Could not load JWT_VALID_FOR from .env file. Missing or not an integer?")
	}
	app.SessionValidFor = time.Duration(validFor) * time.Minute

	// Initialize the validator instance to validate fields with tag 'validate'
	validatorConfig := &validator.Config{TagName: "validate"}
	app.Validator = validator.New(validatorConfig)

	// Set offsets for notification reaper to delete old notifications.
	notifExpOffset, err := strconv.Atoi(os.Getenv("NOTIFICATION_EXPIRY_OFFSET"))
	if err != nil {
		log.Fatal("[InitAndConfig] Could not load NOTIFICATION_EXPIRY_OFFSET from .env file. Missing or not an integer?")
	}
	app.NotifExpOffset = time.Duration(notifExpOffset) * (time.Hour * 24)

	notifSleepOffset, err := strconv.Atoi(os.Getenv("NOTIFICATION_SLEEP_OFFSET"))
	if err != nil {
		log.Fatal("[InitAndConfig] Could not load NOTIFICATION_SLEEP_OFFSET from .env file. Missing or not an integer?")
	}
	app.NotifSleepOffset = time.Duration(notifSleepOffset) * time.Minute

	return app
}
