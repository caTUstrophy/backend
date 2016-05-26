package db

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/leebenson/conform"
)

// Models

// Users and permissions
type Permission struct {
	ID          uint   `gorm:"primary_key"`
	AccessRight string `gorm:"index;not null;unique"`
	Description string
}

type Group struct {
	ID          uint   `gorm:"primary_key"`
	Location    string `gorm:"index"`
	Permissions []Permission
}

type User struct {
	ID            uint `gorm:"primary_key"`
	Name          string
	PreferredName string
	Mail          string `gorm:"index;not null;unique"`
	MailVerified  bool
	PasswordHash  string
	PasswordSalt  string `gorm:"unique"`
	Groups        []Group
	Enabled       bool
}

// Offer, Requests and Matchings

type Tag struct {
	gorm.Model
	Name string
}

type Matching struct {
	ID      uint `gorm:"primary_key"`
	Offer   Offer
	Request Request
}

type Offer struct {
	ID uint `gorm:"primary_key"`

	Tags []Tag `gorm:"index"`
	Name string

	User User `gorm:"index"`

	ValidityPeriod time.Time
	Expired        bool
}

type Request struct {
	ID uint `gorm:"primary_key"`

	Tags []Tag `gorm:"index"`
	Name string

	User User `gorm:"index"`

	ValidityPeriod time.Time
	Expired        bool
}

// Set Expired-flag for all requests and offers
func CheckForExpired(db *gorm.DB) {
	//TODO
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
