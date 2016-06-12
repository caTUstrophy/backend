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
	//       Put this config into config.go and make it environment loadable.
	app.Router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          2 * time.Hour,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	// Define endpoint to handler mapping.
	app.Router.POST("/users", app.CreateUser)
	app.Router.POST("/auth", app.Login)
	app.Router.GET("/auth", app.RenewToken)
	app.Router.DELETE("/auth", app.Logout)

	app.Router.GET("/me", app.GetUser)
	app.Router.POST("/me", app.UpdateUser)

	app.Router.GET("/offers/:region", app.ListOffers)
	app.Router.GET("/me/offers", app.ListUserOffers)
	app.Router.GET("/requests/:region", app.ListRequests)
	app.Router.GET("/me/requests", app.ListUserRequests)

	app.Router.POST("/offers", app.CreateOffer)
	app.Router.POST("/requests", app.CreateRequest)

	app.Router.PUT("/me/offers/:offerID", app.UpdateUserOffer)
	app.Router.PUT("/me/requests/:requestID", app.UpdateUserRequest)

	app.Router.POST("/matchings", app.CreateMatching)
	app.Router.GET("/matchings", app.ListMatchings)
	app.Router.GET("/matchings/:matchingID", app.GetMatching)

	// Run our application.
	app.Router.Run(":3001")
}
