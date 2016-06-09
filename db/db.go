package db

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"

	// TEMPORARY: Comment in whichever you need.
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Models

type Permission struct {
	gorm.Model

	AccessRight string `gorm:"index;not null;unique"`
	Description string
}

type Group struct {
	gorm.Model

	DefaultGroup bool
	Location     string       `gorm:"index"`
	Permissions  []Permission `gorm:"many2many:group_permissions"`
}

type User struct {
	gorm.Model

	Name          string
	PreferredName string

	Mail         string `gorm:"index;not null;unique"`
	MailVerified bool

	PasswordHash string `gorm:"unique"`

	Groups []Group `gorm:"many2many:user_groups"`

	Enabled bool
}

type Tag struct {
	gorm.Model
	Name string `gorm:"index;not null;unique"`
}

type Matching struct {
	gorm.Model
	Offer     Offer `gorm:"ForeignKey:OfferId;AssociationForeignKey:Refer"`
	OfferId   int
	Request   Request `gorm:"ForeignKey:RequestId;AssociationForeignKey:Refer"`
	RequestId int
}

type Offer struct {
	gorm.Model

	Name     string `gorm:"index;not null"`
	User     User
	Location string `gorm:"index;not null"`

	Tags           []Tag `gorm:"many2many:offer_tags"`
	ValidityPeriod int64
	Expired        bool
}

type Request struct {
	gorm.Model

	Name     string `gorm:"index;not null"`
	User     User
	Location string `gorm:"index;not null"`

	Tags           []Tag `gorm:"many2many:request_tags"`
	ValidityPeriod int64
	Expired        bool
}

// Set Expired flag for all requests and offers that are not valid anymore.
func CheckForExpired(db *gorm.DB) {
	// TODO
}

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
	db.CreateTable(&Matching{})
	db.CreateTable(&Offer{})
	db.CreateTable(&Request{})

	// Two default permission entities.

	PermUser := Permission{
		AccessRight: "user",
		Description: "This permission represents a standard, registered but not privileged user in our system.",
	}

	PermAdmin := Permission{
		AccessRight: "admin",
		Description: "This permission represents a registered and fully authorized user in our system. Users with this permission have full API access to our system.",
	}

	// Two default group entities.

	GroupUser := Group{
		DefaultGroup: true,
		Location:     "worldwide",
		Permissions:  []Permission{PermUser},
	}

	GroupAdmin := Group{
		DefaultGroup: false,
		Location:     "worldwide",
		Permissions:  []Permission{PermAdmin},
	}

	// Some default tag entities.
	TagFood := Tag{Name: "Food"}
	TagWater := Tag{Name: "Water"}
	TagVehicle := Tag{Name: "Vehicle"}
	TagTool := Tag{Name: "Tool"}
	TagInformation := Tag{Name: "Information"}

	Tags := []Tag{TagFood, TagWater, TagVehicle, TagTool, TagInformation}

	// Create default admin user ('admin@example.org', 'CaTUstrophyAdmin123$').
	// TODO: Replace this with an interactive dialog, when starting
	//       the backend for the first time.
	UserAdmin := User{
		Name:          "admin",
		PreferredName: "The Boss Around Here",
		Mail:          "admin@example.org",
		PasswordHash:  "$2a$10$SkmaOImXqNS/PSWp65p1ougtA1N.o8r5qyu8M4RPTfGSMEf2k.Q1C",
		Groups:        []Group{GroupAdmin},
		Enabled:       true,
	}

	// Create the database elements for these default values.

	db.Create(&GroupUser)
	db.Create(&GroupAdmin)
	db.Create(&UserAdmin)

	for _, Tag := range Tags {
		log.Println(Tag)
		db.Create(&Tag)
	}
}
