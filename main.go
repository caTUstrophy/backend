package main

import (
	"log"

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

func GetUsers(c *gin.Context) {

	var Users []db.User

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		log.Fatal("[GetUsers] Could not retrieve database connection from gin context.")
	}

	// Make a database call for all users.
	db.Find(&Users)

	// Return JSON formatted list of all users.
	c.JSON(200, Users)
}

func GetUser(c *gin.Context) {

	var User db.User

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		log.Fatal("[GetUser] Could not retrieve database connection from gin context.")
	}

	// Retrieve user ID from URL context.
	userID := c.Params.ByName("userID")

	// Make a database call for a specific user.
	db.Find(&User, "id = ?", userID)

	// Return JSON formatted user object.
	c.JSON(200, User)
}

func CreateUser(c *gin.Context) {

	var User db.User

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

func main() {

	// Create new gin instance with default functionalities.
	router := gin.Default()

	// Open connection to database and insert middleware.
	db := db.InitDB("sqlite3", "caTUstrophy.sqlite")

	// Add custom middleware to call stack.
	router.Use(DatabaseMiddleware(db))
	router.Use(CORSMiddleware())

	// Define routes.
	router.GET("/users", GetUsers)
	router.GET("/users/:userID", GetUser)
	router.POST("/users", CreateUser)

	// Run our application.
	router.Run(":3001")
}
