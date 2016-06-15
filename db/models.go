package db

import (
	"time"
    "reflect"
)

// Models

type Permission struct {
	ID          string `gorm:"primary_key"`
	AccessRight string `gorm:"index;not null;unique"`
	Description string
}

type Group struct {
	ID           string `gorm:"primary_key"`
	DefaultGroup bool
	Location     string       `gorm:"index"`
	Permissions  []Permission `gorm:"many2many:group_permissions"`
}

type User struct {
	ID            string `gorm:"primary_key"`
	Name          string
	PreferredName string
	Mail          string `gorm:"index;not null;unique"`
	MailVerified  bool
	PasswordHash  string  `gorm:"unique"`
	Groups        []Group `gorm:"many2many:user_groups"`
	Enabled       bool
}

type Tag struct {
	ID   string `gorm:"primary_key"`
	Name string `gorm:"index;not null;unique"`
}

type Offer struct {
	ID             string `gorm:"primary_key"`
	Name           string `gorm:"index;not null"`
	User           User   `gorm:"ForeignKey:UserID;AssociationForeignKey:Refer"`
	UserID         string
	Location       string `gorm:"index;not null"`
	Tags           []Tag  `gorm:"many2many:offer_tags"`
	ValidityPeriod time.Time
	Expired        bool
}

type Request struct {
	ID             string `gorm:"primary_key"`
	Name           string `gorm:"index;not null"`
	User           User   `gorm:"ForeignKey:UserID;AssociationForeignKey:Refer"`
	UserID         string
	Location       string `gorm:"index;not null"`
	Tags           []Tag  `gorm:"many2many:request_tags"`
	ValidityPeriod time.Time
	Expired        bool
}

type Matching struct {
	ID        string `gorm:"primary_key"`
	Offer     Offer  `gorm:"ForeignKey:OfferId;AssociationForeignKey:Refer"`
	OfferId   string
	Request   Request `gorm:"ForeignKey:RequestId;AssociationForeignKey:Refer"`
	RequestId string
}

// TODO: Add an area representation similar to this.
//       Make use of PostGIS and Postgres native geometric types.
//       Points will NOT be represented like this.

/*
type Point struct {
	Longitude float32
	Latitude  float32
}

type Area struct {
	ID          string `gorm:"primary_key"`
	Name        string
	Description string
	Boundaries  []Point
}
*/


func CopyModel(i interface{}, fields []string) (map[string]interface{}) {
	var m map[string]interface{}
	m = make(map[string]interface{})

	// get type of struct
	s := reflect.ValueOf(i)
	typeOfT := reflect.ValueOf(i).Type()

	// iterate over fields of struct
	for i := 0; i < s.NumField(); i++ {
	    for _, field := range fields{
			// set only fields specified in array
	    	if typeOfT.Field(i).Name == field{
	    		m[field] = s.Field(i).Interface()
	    		break
	    	}
	    }
	}

	return m
}
	