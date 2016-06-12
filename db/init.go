package db

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"

	// TEMPORARY: Use whichever connector you need.
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Functions.

// Create connection to our database from environment file.
func InitDB() *gorm.DB {

	var db *gorm.DB

	// Fetch from environment which database type to connect to.
	dbType := os.Getenv("DB_TYPE")

	// Tries to connect to specified database.
	if dbType == "postgres" {

		port, err := strconv.Atoi(os.Getenv("DB_PORT"))
		if err != nil {
			log.Fatal("[InitDB] Unrecognized port type in .env file. Integer expected.")
		}

		db, err = gorm.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			os.Getenv("DB_USER"), os.Getenv("DB_PW"), os.Getenv("DB_HOST"),
			port, os.Getenv("DB_DBNAME"), os.Getenv("DB_SSLMODE")))
		if err != nil {
			log.Fatal(err)
		}

	} else if dbType == "mysql" {

		port, err := strconv.Atoi(os.Getenv("DB_PORT"))
		if err != nil {
			log.Fatal("[InitDB] Unrecognized port type in .env file. Integer expected.")
		}

		db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@%s:%d/%s?charset=utf8&parseTime=True&loc=Local",
			os.Getenv("DB_USER"), os.Getenv("DB_PW"), os.Getenv("DB_HOST"),
			port, os.Getenv("DB_DBNAME")))
		if err != nil {
			log.Fatal(err)
		}

	} else if dbType == "sqlite" {

		var err error

		db, err = gorm.Open("sqlite3", os.Getenv("SQLITE_DB_PATH"))
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Fatal("[InitDB] Unknown database type in environment file. Did you forget to specify a database in your .env file?")
	}

	// Check connection to database in order to be sure.
	err := db.DB().Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Development: Log SQL by gorm.
	db.LogMode(true)

	return db
}

// Insert default data into Permissions, Groups and Tags table
func AddDefaultData(db *gorm.DB) {

	// Check if our tables are present, otherwise create them.
	db.CreateTable(&Permission{})
	db.CreateTable(&Group{})
	db.CreateTable(&User{})
	db.CreateTable(&Tag{})
	db.CreateTable(&Offer{})
	db.CreateTable(&Request{})
	db.CreateTable(&Matching{})
	// db.CreateTable(&Area{})

	// Two default permission entities.

	PermUser := Permission{
		ID:          fmt.Sprintf("%s", uuid.NewV4()),
		AccessRight: "user",
		Description: "This permission represents a standard, registered but not privileged user in our system.",
	}

	PermAdmin := Permission{
		ID:          fmt.Sprintf("%s", uuid.NewV4()),
		AccessRight: "admin",
		Description: "This permission represents a registered and fully authorized user in our system. Users with this permission have full API access to our system.",
	}

	// Two default group entities.

	GroupUser := Group{
		ID:           fmt.Sprintf("%s", uuid.NewV4()),
		DefaultGroup: true,
		Location:     "worldwide",
		Permissions:  []Permission{PermUser},
	}

	GroupAdmin := Group{
		ID:           fmt.Sprintf("%s", uuid.NewV4()),
		DefaultGroup: false,
		Location:     "worldwide",
		Permissions:  []Permission{PermAdmin},
	}

	// Some default tag entities.

	TagFood := Tag{ID: fmt.Sprintf("%s", uuid.NewV4()), Name: "Food"}
	TagWater := Tag{ID: fmt.Sprintf("%s", uuid.NewV4()), Name: "Water"}
	TagVehicle := Tag{ID: fmt.Sprintf("%s", uuid.NewV4()), Name: "Vehicle"}
	TagTool := Tag{ID: fmt.Sprintf("%s", uuid.NewV4()), Name: "Tool"}
	TagInformation := Tag{ID: fmt.Sprintf("%s", uuid.NewV4()), Name: "Information"}

	Tags := []Tag{TagFood, TagWater, TagVehicle, TagTool, TagInformation}

	// Create default admin user ('admin@example.org', 'CaTUstrophyAdmin123$').
	// TODO: Replace this with an interactive dialog, when starting
	//       the backend for the first time.
	UserAdmin := User{
		ID:            fmt.Sprintf("%s", uuid.NewV4()),
		Name:          "admin",
		PreferredName: "The Boss Around Here",
		Mail:          "admin@example.org",
		PasswordHash:  "$2a$10$SkmaOImXqNS/PSWp65p1ougtA1N.o8r5qyu8M4RPTfGSMEf2k.Q1C",
		Groups:        []Group{GroupAdmin},
		Enabled:       true,
	}

	// A default areas.
	/*
		AreaTU := Area{
			ID:          fmt.Sprintf("%s", uuid.NewV4()),
			Name:        "TU Berlin",
			Description: "The campus of the Technische Universität Berlin in Charlottenburg, Berlin.",
			Boundaries:  []Point{Point{13.324401, 52.516872}, Point{13.322599, 52.514740}, Point{13.322679, 52.512611}, Point{13.322674, 52.511743}, Point{13.328280, 52.508302}, Point{13.331077, 52.512191}, Point{13.329763, 52.513787}, Point{13.324401, 52.516872}},
		}
	*/

	// Create the database elements for these default values.
	db.Create(&PermUser)
	db.Create(&PermAdmin)
	db.Create(&GroupUser)
	db.Create(&GroupAdmin)
	db.Create(&UserAdmin)
	// db.Create(&AreaTU)

	for _, Tag := range Tags {
		log.Println(Tag)
		db.Create(&Tag)
	}
}

// Set Expired flag for all requests and offers that are not valid anymore.
func CheckForExpired(db *gorm.DB) {
	// TODO
}
