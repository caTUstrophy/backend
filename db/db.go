package db

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Models

// Users and permissions

type Permission struct {
	gorm.Model

	AccessRight string `gorm:"index;not null;unique"`
	Description string
}

type Group struct {
	gorm.Model

	Location    string `gorm:"index"`
	Permissions []Permission
}

type User struct {
	gorm.Model

	Name          string
	PreferredName string

	Mail         string `gorm:"index;not null;unique"`
	MailVerified bool

	PasswordHash string `gorm:"unique"`

	Groups []Group

	Enabled bool
}

// Offer, Request and Matching

type Tag struct {
	gorm.Model
	Name string `gorm:"index;not null;unique"`
}

type Matching struct {
	gorm.Model
	Offer   Offer
	Request Request
}

type Offer struct {
	gorm.Model

	Tags []Tag
	Name string

	User User

	ValidityPeriod time.Time
	Expired        bool
}

type Request struct {
	gorm.Model

	Tags []Tag
	Name string

	User User

	ValidityPeriod time.Time
	Expired        bool
}

// Set Expired-flag for all requests and offers
func CheckForExpired(db *gorm.DB) {
	// TODO
}

// Create connection to our database.
func InitDB(databaseType string, databaseName string) *gorm.DB {

	// Tries to connect to specified database.
	db, err := gorm.Open(databaseType, databaseName)
	if err != nil {
		log.Fatal(err)
	}

	// Check connection to database in order to be sure.
	err = db.DB().Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Check if our tables are present, otherwise create them.
	db.CreateTable(&Permission{})
	db.CreateTable(&Group{})
	db.CreateTable(&User{})
	db.CreateTable(&Tag{})
	db.CreateTable(&Matching{})
	db.CreateTable(&Offer{})
	db.CreateTable(&Request{})

	return db
}

// Insert default data into Permissions table
func InsertDefaultPermissions(db *gorm.DB) {
	// TODO
}

// Insert default data into Groups table
func InsertDefaultGroups(db *gorm.DB) {
	// TODO
}
