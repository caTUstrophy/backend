package db

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nferruzzi/gormGIS"
	"github.com/satori/go.uuid"
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
	} else {
		log.Fatal("[InitDB] Unsupported database type in environment file, please use PostgreSQL with PostGIS instance. Did you forget to specify a database in your .env file?")
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

	// Drop all existing tables that will be created afterwards.
	db.DropTableIfExists(&Permission{})
	db.DropTableIfExists(&Group{})
	db.DropTableIfExists(&User{})
	db.DropTableIfExists(&Tag{})
	db.DropTableIfExists(&Offer{})
	db.DropTableIfExists(&Request{})
	db.DropTableIfExists(&Matching{})
	db.DropTableIfExists(&Region{})
	db.DropTableIfExists("region_offers")
	db.DropTableIfExists("region_requests")
	db.DropTableIfExists("group_permissions")
	db.DropTableIfExists("offer_tags")
	db.DropTableIfExists("request_tags")
	db.DropTableIfExists("user_groups")

	// Check if our tables are present, otherwise create them.
	db.CreateTable(&Permission{})
	db.CreateTable(&Group{})
	db.CreateTable(&User{})
	db.CreateTable(&Tag{})
	db.CreateTable(&Offer{})
	db.CreateTable(&Request{})
	db.CreateTable(&Matching{})
	db.CreateTable(&Region{})

	// Three default permission entities.

	PermUser := Permission{
		ID:          fmt.Sprintf("%s", uuid.NewV4()),
		AccessRight: "user",
		Description: "This permission represents a standard, registered but not privileged user in our system.",
	}

	PermAdmin := Permission{
		ID:          fmt.Sprintf("%s", uuid.NewV4()),
		AccessRight: "admin",
		Description: "This permission represents a registered and fully authorized user for a specified region in our system. Users with this permission can view all offers, requests and matches in the region they have this permission for. Also, they can create matches.",
	}

	PermSuperAdmin := Permission{
		ID:          fmt.Sprintf("%s", uuid.NewV4()),
		AccessRight: "superadmin",
		Description: "This permission represents a registered and fully authorized user in our system. Users with this permission have full API access to our system.",
	}

	RegionTU := Region{
		ID:          fmt.Sprintf("%s", uuid.NewV4()),
		Name:        "TU Berlin",
		Description: "The campus of the Technische Universität Berlin in Charlottenburg, Berlin.",
		Boundaries:  GeoPolygon{[]gormGIS.GeoPoint{gormGIS.GeoPoint{13.324401, 52.516872}, gormGIS.GeoPoint{13.322599, 52.514740}, gormGIS.GeoPoint{13.322679, 52.512611}, gormGIS.GeoPoint{13.322674, 52.511743}, gormGIS.GeoPoint{13.328280, 52.508302}, gormGIS.GeoPoint{13.331077, 52.512191}, gormGIS.GeoPoint{13.329763, 52.513787}, gormGIS.GeoPoint{13.324401, 52.516872}}},
	}

	// Three default group entities.

	GroupUser := Group{
		ID:           fmt.Sprintf("%s", uuid.NewV4()),
		DefaultGroup: true,
		Location:     Region{},
		LocationId:   "",
		Permissions:  []Permission{PermUser},
	}

	GroupAdmin := Group{
		ID:           fmt.Sprintf("%s", uuid.NewV4()),
		DefaultGroup: false,
		Location:     RegionTU,
		LocationId:   RegionTU.ID,
		Permissions:  []Permission{PermAdmin},
	}

	GroupSuperAdmin := Group{
		ID:           fmt.Sprintf("%s", uuid.NewV4()),
		DefaultGroup: false,
		Location:     Region{},
		LocationId:   "",
		Permissions:  []Permission{PermSuperAdmin},
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
		Groups:        []Group{GroupSuperAdmin, GroupAdmin},
		Enabled:       true,
	}

	// Create the database elements for these default values.
	db.Create(&PermUser)
	db.Create(&PermAdmin)
	db.Create(&PermSuperAdmin)
	db.Create(&GroupUser)
	db.Create(&GroupAdmin)
	db.Create(&GroupSuperAdmin)
	db.Create(&UserAdmin)
	db.Create(&RegionTU)

	for _, Tag := range Tags {
		db.Create(&Tag)
	}
}

// Set Expired flag for all requests and offers that are not valid anymore.
func CheckForExpired(db *gorm.DB) {
	// TODO: Write this expired element reaper function.
	//       Cycle through request and offer tables and set Expired fields to true
	//       where a ValidityPeriod is smaller than current time.
}
