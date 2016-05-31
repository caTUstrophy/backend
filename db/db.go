package db

import (
	"log"

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

// Offer, Request and Matching

type Tag struct {
	gorm.Model
	Name string `gorm:"index;not null;unique"`
}

type Matching struct {
	gorm.Model
	Offer   Offer `gorm:"ForeignKey:OfferId;AssociationForeignKey:Refer"`
	OfferId int
	Request Request `gorm:"ForeignKey:RequestId;AssociationForeignKey:Refer"`
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

	db.Create(&GroupUser)
	db.Create(&GroupAdmin)

	for _, Tag := range Tags {
		log.Println(Tag)
		db.Create(&Tag)
	}
}
