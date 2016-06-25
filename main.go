package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/itsjamie/gin-cors"
	"github.com/jinzhu/gorm"
)

// Structs

type App struct {
	IP              string
	Port            string
	Router          *gin.Engine
	DB              *gorm.DB
	HashCost        int
	SessionValidFor time.Duration
	Validator       *validator.Validate
}

// Main

func InitApp() *App {

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

	// Define our endpoints.

	app.Router.POST("/auth", app.Login)
	app.Router.GET("/auth", app.RenewToken)
	app.Router.DELETE("/auth", app.Logout)

	app.Router.POST("/users", app.CreateUser)

	app.Router.POST("/offers", app.CreateOffer)
	app.Router.GET("/offers/:offerID", app.GetOffer)
	app.Router.PUT("/offers/:offerID", app.UpdateOffer)

	app.Router.POST("/requests", app.CreateRequest)
	app.Router.GET("/requests/:requestID", app.GetRequest)
	app.Router.PUT("/requests/:requestID", app.UpdateRequest)

	app.Router.POST("/matchings", app.CreateMatching)
	app.Router.GET("/matchings/:matchingID", app.GetMatching)

	app.Router.POST("/regions", app.CreateRegion)
	app.Router.GET("/regions", app.ListRegions)
	app.Router.GET("/regions/:regionID", app.GetRegion)
	app.Router.PUT("/regions/:regionID", app.UpdateRegion)
	app.Router.GET("/regions/:regionID/requests", app.GetRequestsForRegion)
	app.Router.GET("/regions/:regionID/offers", app.GetOffersForRegion)
	app.Router.GET("/regions/:regionID/matchings", app.GetMatchingsForRegion)
	app.Router.GET("/regions/:regionID/admins", app.GetAdminsForRegion)
	app.Router.POST("/regions/:regionID/admins", app.PromoteToRegionAdmin)

	app.Router.GET("/me", app.GetUser)
	app.Router.PUT("/me", app.UpdateUser)
	app.Router.GET("/me/offers", app.ListUserOffers)
	app.Router.GET("/me/requests", app.ListUserRequests)

	app.Router.GET("/help", app.GetJsonResponseInfo)

	return app
}

func main() {
	app := InitApp()

	// Run our application.
	app.Router.Run(fmt.Sprintf("%s:%s", app.IP, app.Port))
}
