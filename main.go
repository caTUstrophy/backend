package main

import (
	"fmt"
	"log"
	"time"

	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/itsjamie/gin-cors"
	"github.com/jinzhu/gorm"
)

// Structs

type App struct {
	IP                string
	Port              string
	Router            *gin.Engine
	DB                *gorm.DB
	HashCost          int
	SessionValidFor   time.Duration
	Validator         *validator.Validate
	OffReqSleepOffset time.Duration
	NotifExpOffset    time.Duration
	NotifSleepOffset  time.Duration
	TagsWeightAlpha   float64
	DescWeightBeta    float64
}

// Functions

func InitApp() *App {

	// Parse command line flags and build application config.
	app := InitAndConfig()

	// Enable compliance to CORS.
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
	app.Router.GET("/users", app.ListUsers)
	app.Router.GET("/users/:userID", app.GetUser)
	app.Router.PUT("/users/:userID", app.UpdateUser)
	// This endpoint might change.
	app.Router.POST("/users/admins", app.PromoteToSystemAdmin)

	app.Router.GET("/groups", app.GetGroups)
	app.Router.GET("/tags", app.GetTags)

	app.Router.POST("/offers", app.CreateOffer)
	app.Router.GET("/offers/:offerID", app.GetOffer)
	app.Router.PUT("/offers/:offerID", app.UpdateOffer)

	app.Router.POST("/requests", app.CreateRequest)
	app.Router.GET("/requests/:requestID", app.GetRequest)
	app.Router.PUT("/requests/:requestID", app.UpdateRequest)

	app.Router.POST("/matchings", app.CreateMatching)
	app.Router.GET("/matchings/:matchingID", app.GetMatching)
	app.Router.PUT("/matchings/:matchingID", app.UpdateMatching)

	app.Router.POST("/regions", app.CreateRegion)
	app.Router.GET("/regions", app.ListRegions)
	app.Router.GET("/regions/:regionID", app.GetRegion)
	app.Router.PUT("/regions/:regionID", app.UpdateRegion)
	app.Router.GET("/regions/:regionID/requests", app.ListRequestsForRegion)
	app.Router.GET("/regions/:regionID/offers", app.ListOffersForRegion)
	app.Router.GET("/regions/:regionID/matchings", app.ListMatchingsForRegion)
	app.Router.GET("/regions/:regionID/admins", app.ListAdminsForRegion)
	app.Router.POST("/regions/:regionID/admins", app.PromoteToRegionAdmin)
	app.Router.GET("/regions/:regionID/recommendations", app.ListRecommendationsForRegion)
	app.Router.GET("/regions/:regionID/requests/:requestID/recommendations", app.ListOffersForRequest)
	app.Router.GET("/regions/:regionID/offers/:offerID/recommendations", app.ListRequestsForOffer)

	// This endpoint might change.
	app.Router.GET("/system/admins", app.ListSystemAdmins)
	app.Router.POST("/system/admins", app.PromoteToSystemAdmin)

	app.Router.GET("/me", app.GetMe)
	app.Router.PUT("/me", app.UpdateMe)
	app.Router.GET("/me/offers", app.ListUserOffers)
	app.Router.GET("/me/requests", app.ListUserRequests)
	app.Router.GET("/me/matchings", app.ListUserMatchings)

	app.Router.GET("/notifications", app.ListNotifications)
	app.Router.PUT("/notifications/:notificationID", app.UpdateNotification)

	// This endpoint is for internal usage only.
	app.Router.GET("/help", app.GetJsonResponseInfo)

	return app
}

func main() {

	app := InitApp()

	// Start goroutine that sets expired field of offers.
	go db.OfferRequestReaper(app.DB, "Offers", app.OffReqSleepOffset)
	log.Printf("\n[main] Dispatched offers reaper with %s sleep time.", app.OffReqSleepOffset.String())

	// Start goroutine that sets expired field of requests.
	go db.OfferRequestReaper(app.DB, "Requests", app.OffReqSleepOffset)
	log.Printf("\n[main] Dispatched requests reaper with %s sleep time.", app.OffReqSleepOffset.String())

	// Start goroutine to delete old notifications.
	go db.NotificationReaper(app.DB, app.NotifExpOffset, app.NotifSleepOffset)
	log.Printf("\n[main] Dispatched notification reaper with %s expiry time and %s sleep time.\n\n", app.NotifExpOffset.String(), app.NotifSleepOffset.String())

	// Run our application.
	app.Router.Run(fmt.Sprintf("%s:%s", app.IP, app.Port))
}
