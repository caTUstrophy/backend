package main

import (
	"log"

	//"github.com/caTUstrophy/backend/cache"
	"github.com/caTUstrophy/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Work around until we get a proper context file.
func DatabaseMiddleware(db *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Next()
	}
}

/**
 * new endpoints
 */
func CreateUser(c *gin.Context) {

	var User db.User

	// Check for existence of mail attribute in user supplied data
    if User.Mail

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		log.Fatal("[CreateUser] Could not retrieve database connection from gin context.")
	}

	// Expect user struct fields in JSON request body.
	c.BindJSON(&User)

	// IMPORTANT CHECKS HAVE TO HAPPEN HERE:
	// * Validity checks (certain fields not null etc.)
	// * Security checks (filter out malicious input)

	// Create user object in database.
	db.Create(&User)

	// On success: return newly created user.
	c.JSON(201, User)
}

func Login(c *gin.Context) {

}

func RenewToken(c *gin.Context) {

}

func Logout(c *gin.Context) {

}

func ListOffers(c *gin.Context) {

}

func ListRequests(c *gin.Context) {

}

func CreateRequest(c *gin.Context) {

}

func CreateMatching(c *gin.Context) {

}

func GetMatching(c *gin.Context) {
	//matchingID := c.Params.ByName("matchingID")
}

func main() {

	// Create new gin instance with default functionalities.
	router := gin.Default()

	// Open connection to database and insert middleware.
	db := db.InitDB("sqlite3", "caTUstrophy.sqlite")

	// Add custom middleware to call stack.
	router.Use(DatabaseMiddleware(db))
	router.Use(CORSMiddleware())

	// Define endpoint to handler mapping
	router.POST("/users", CreateUser)
	router.POST("/auth", Login)
	router.GET("/auth", RenewToken)
	router.DELETE("/auth", Logout)
	router.GET("/offers", ListOffers)
	router.GET("/requests", ListRequests)
	router.POST("/requests", CreateRequest)
	router.POST("/matchings", CreateMatching)
	router.GET("/matchings/:matchingID", GetMatching)

	// Run our application.
	router.Run(":3001")
}
