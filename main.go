package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/itsjamie/gin-cors"
	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
)

// Structs

type App struct {
	Router          *gin.Engine
	DB              *gorm.DB
	Sessions        *cache.Cache
	Validator       *validator.Validate
	HashCost        int
	SessionValidFor time.Duration
}

// Main

func main() {

	// Parse command line flags and build application config.
	app := InitAndConfig()

	// Enable compliance to CORS.
	// TODO: Keep the values in check when this backend gets deployed (Origins!).
	app.Router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          2 * time.Hour,
		Credentials:     true,
		ValidateHeaders: false,
	}))

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
