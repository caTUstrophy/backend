package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
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

// Functions

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

	// Add custom middleware to call stack.
	app.Router.Use(CORSMiddleware())

	// Define endpoint to handler mapping
	app.Router.POST("/users", app.CreateUser)
	app.Router.POST("/auth", app.Login)
	app.Router.GET("/auth", app.RenewToken)
	app.Router.DELETE("/auth", app.Logout)
	app.Router.POST("/offers", app.CreateOffer)
	app.Router.GET("/offers", app.ListOffers)
	app.Router.GET("/requests", app.ListRequests)
	app.Router.POST("/requests", app.CreateRequest)
	app.Router.POST("/matchings", app.CreateMatching)
	app.Router.GET("/matchings/:matchingID", app.GetMatching)

	// Run our application.
	app.Router.Run(":3001")
}
