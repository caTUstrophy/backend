package db

import (
	"time"
	"fmt"
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


func CopyNestedModel(i interface{}, fields map[string]interface{}) (map[string]interface{}) {
	var m map[string]interface{}
	m = make(map[string]interface{})

	// get value + type of source interface
	valInterface := reflect.ValueOf(i)
	typeOfT := reflect.ValueOf(i).Type()


	// iterate over all fields that will be copied
	for key := range fields {
		var exists = false

		// search for field in source type
		for i := 0; i < valInterface.NumField(); i++ {
	    	if typeOfT.Field(i).Name == key{

	    		// check for nesting through type assertion 
	    		nestedMap, nested := fields[key].(map[string]interface{})

	    		if !nested {
	    			// NOT nested -> copy value directly
					m[key] = valInterface.Field(i).Interface()
				} else { 

					// NESTED copied via recursion
					var slice = reflect.ValueOf(valInterface.Field(i).Interface())

					// if nested ARRAY
					if valInterface.Field(i).Kind() == reflect.Slice {
						sliceMapped := make([]interface{}, slice.Len())
						for i:=0; i<slice.Len(); i++ {
					        sliceMapped[i] = CopyNestedModel(slice.Index(i).Interface(), nestedMap)
					    }
					    m[key] = sliceMapped
					} else {
						// if nested OBJECT
						m[key] = CopyNestedModel(valInterface.Field(i).Interface(), nestedMap)
					}

				}

				exists = true
	    		break
		    }
		}

		if !exists {
			panic(fmt.Sprintf("ERROR: Struct<%s> has no field: %s", typeOfT.Name(), key))
		}
	}

	return m
}
		