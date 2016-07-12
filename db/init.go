package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

	// Add some requests and offers
	req1 := Request{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "Toothbrushes",
		UserID:         UserAdmin.ID,
		User:           UserAdmin,
		Radius:         10.0,
		Tags:           []Tag{TagMedical},
		Location:       gormGIS.GeoPoint{13.326863, 52.513142},
		Description:    "I need toothbrushes for me and my family, we are four peaple but if necessary we can share! Toothpaste would also be really nice!",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	req2 := Request{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "Mini USB charger",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []Tag{TagTool, TagOther},
		Location:       gormGIS.GeoPoint{13.326860, 52.513142},
		Description:    "Hey everyone, I lost my charger and would love to get exchange for it as I really need my phone, as fast as possible. We have electricity here, you can use it; my phone has a mini usb plot",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	req3 := Request{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "A 2 meters sized chocolate letter",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []Tag{TagFood, TagChildren},
		Location:       gormGIS.GeoPoint{13.326859, 52.513143},
		Description:    "Sorry guys I lost my 3 meter sized chocolate 'A'. Its dark chocolate wih I guess 60percent cacao. I am very sad and if this cant be found another big sized chocolate letter would help, but i really want to eat chocolate",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off1 := Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "hygiene stuff",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []Tag{TagMedical},
		Location:       gormGIS.GeoPoint{13.326861, 52.513145},
		Description:    "hey, i have some toothbrushes, toothpasta, cacao shampoo and a electric shaver to offer",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off2 := Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "phone charger",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []Tag{TagOther},
		Location:       gormGIS.GeoPoint{13.326861, 52.513142},
		Description:    "i have a charger for mobile phones, but no public electricity around",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off3 := Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "Children stuff",
		UserID:         UserAdmin.ID,
		Radius:         10,
		Tags:           []Tag{TagChildren},
		Location:       gormGIS.GeoPoint{13.326862, 52.513143},
		Description:    "Hey, I have some stuff kids like to eat, choco sweets and chips",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}
	off4 := Offer{
		ID:             fmt.Sprintf("%s", uuid.NewV4()),
		Name:           "This is a very bad offer",
		UserID:         UserAdmin.ID,
		Radius:         0.0001,
		Tags:           []Tag{TagOther},
		Location:       gormGIS.GeoPoint{13.326861, 52.513143},
		Description:    "Extraordnariy unusefullness that does not fit anything really good.",
		ValidityPeriod: time.Now().Add(time.Hour * 1000),
		Matched:        false,
		Expired:        false,
	}

	log.Println(req1)

	db.Create(req1)
	db.Create(req2)
	db.Create(req3)
	db.Create(off1)
	db.Create(off2)
	db.Create(off3)
	db.Create(off4)

	RegionTU.Offers = append(RegionTU.Offers, off1)
	RegionTU.Offers = append(RegionTU.Offers, off2)
	RegionTU.Offers = append(RegionTU.Offers, off3)
	RegionTU.Offers = append(RegionTU.Offers, off4)
	RegionTU.Requests = append(RegionTU.Requests, req1)
	RegionTU.Requests = append(RegionTU.Requests, req2)
	RegionTU.Requests = append(RegionTU.Requests, req3)
	db.Save(&RegionTU)

}
