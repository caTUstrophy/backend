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
	db.DropTableIfExists(&Group{})
	db.DropTableIfExists(&User{})
	db.DropTableIfExists(&Tag{})
	db.DropTableIfExists(&Offer{})
	db.DropTableIfExists(&Request{})
	db.DropTableIfExists(&Matching{})
	db.DropTableIfExists(&Region{})
	db.DropTableIfExists(&Notification{})
	db.DropTableIfExists(&MatchingScore{})
	db.DropTableIfExists("region_offers")
	db.DropTableIfExists("region_requests")
	db.DropTableIfExists("offer_tags")
	db.DropTableIfExists("request_tags")
	db.DropTableIfExists("user_groups")

	// Check if our tables are present, otherwise create them.
	db.CreateTable(&Group{})
	db.CreateTable(&User{})
	db.CreateTable(&Tag{})
	db.CreateTable(&Offer{})
	db.CreateTable(&Request{})
	db.CreateTable(&Matching{})
	db.CreateTable(&Region{})
	db.CreateTable(&Notification{})
	db.CreateTable(&MatchingScore{})

	// Three default permission entities.

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
		Region:       Region{},
		RegionId:     "",
		AccessRight:  "user",
		Description:  "This permission represents a standard, registered but not privileged user in our system.",
	}

	GroupAdmin := Group{
		ID:           fmt.Sprintf("%s", uuid.NewV4()),
		DefaultGroup: false,
		AccessRight:  "superadmin",
		Description:  "This permission represents a registered and fully authorized user in our system. Users with this permission have full API access to our system.",
	}

	GroupRegionAdmin := Group{
		ID:           fmt.Sprintf("%s", uuid.NewV4()),
		DefaultGroup: false,
		Region:       RegionTU,
		RegionId:     RegionTU.ID,
		AccessRight:  "admin",
		Description:  "This permission represents a registered and fully authorized user in our system. Users with this permission have full API access to our system.",
	}

	// Some default tag entities.

	TagFood := Tag{Name: "Food"}
	TagWater := Tag{Name: "Water"}
	TagVehicle := Tag{Name: "Vehicle"}
	TagTool := Tag{Name: "Tool"}
	TagInformation := Tag{Name: "Information"}
	TagChildren := Tag{Name: "Children"}
	TagOther := Tag{Name: "Other"}
	TagMedical := Tag{Name: "Medical"}

	Tags := []Tag{TagFood, TagWater, TagVehicle, TagTool, TagInformation, TagChildren, TagOther, TagMedical}

	// Two default phone numbers.
	PhoneNumbers := new(PhoneNumbers)
	err := PhoneNumbers.Scan([]string{"01611234567", "0419123456"})
	if err != nil {
		log.Fatalf("[AddDefaultData] JSON marshaling of default phone numbers went wrong: %s\n", err)
	}

	// Create default admin user ('admin@example.org', 'CaTUstrophyAdmin123$').
	// TODO: Replace this with an interactive dialog, when starting
	//       the backend for the first time.
	UserAdmin := User{
		ID:            fmt.Sprintf("%s", uuid.NewV4()),
		Name:          "admin",
		PreferredName: "The Boss Around Here",
		Mail:          "admin@example.org",
		PhoneNumbers:  *PhoneNumbers,
		PasswordHash:  "$2a$10$SkmaOImXqNS/PSWp65p1ougtA1N.o8r5qyu8M4RPTfGSMEf2k.Q1C",
		Groups:        []Group{GroupAdmin},
		Enabled:       true,
	}

	// Create the database elements for these default values.
	db.Create(&GroupUser)
	db.Create(&GroupRegionAdmin)
	db.Create(&GroupAdmin)
	db.Create(&UserAdmin)
	db.Create(&RegionTU)

	for _, Tag := range Tags {
		db.Create(&Tag)
	}
}
