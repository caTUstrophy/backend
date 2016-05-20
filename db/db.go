package db

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Models

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
	ID           uint `gorm:"primary_key"`
	FirstName    string
	LastName     string
	Mail         string `gorm:"index;not null;unique"`
	PasswordHash string
	PasswordSalt string `gorm:"unique"`
	Groups       []Group
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

	return db
}
