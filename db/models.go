package db

import (
	"fmt"
	"reflect"
	"time"

	"github.com/nferruzzi/gormGIS"
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
	Location     Region `gorm:"ForeignKey:LocationId;AssociationForeignKey:Refer"`
	LocationId   string
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
	Location       gormGIS.GeoPoint `gorm:"not null" sql:"type:geometry(Geometry,4326)"`
	Tags           []Tag            `gorm:"many2many:offer_tags"`
	ValidityPeriod time.Time
	Matched        bool
	Expired        bool
}

type Request struct {
	ID             string `gorm:"primary_key"`
	Name           string `gorm:"index;not null"`
	User           User   `gorm:"ForeignKey:UserID;AssociationForeignKey:Refer"`
	UserID         string
	Location       gormGIS.GeoPoint `gorm:"not null" sql:"type:geometry(Geometry,4326)"`
	Tags           []Tag            `gorm:"many2many:request_tags"`
	ValidityPeriod time.Time
	Matched        bool
	Expired        bool
}

type Matching struct {
	ID        string `gorm:"primary_key"`
	Offer     Offer  `gorm:"ForeignKey:OfferId;AssociationForeignKey:Refer"`
	OfferId   string
	Request   Request `gorm:"ForeignKey:RequestId;AssociationForeignKey:Refer"`
	RequestId string
}

type Region struct {
	ID          string `gorm:"primary_key"`
	Name        string
	Boundaries  GeoPolygon `sql:"type:geometry(Geometry,4326)"`
	Description string
	Offers      []Offer   `gorm:"many2many:region_offers"`
	Requests    []Request `gorm:"many2many:region_requests"`
}

func CopyNestedModel(i interface{}, fields map[string]interface{}) map[string]interface{} {
	var m map[string]interface{}
	m = make(map[string]interface{})

	// get value + type of source interface
	valInterface := reflect.ValueOf(i)
	typeOfT := reflect.ValueOf(i).Type()

	// iterate over all fields that will be copied
	for key := range fields {
		var exists = false
		newKey, _ := fields[key].(string)

		// search for field in source type
		for i := 0; i < valInterface.NumField(); i++ {
			if typeOfT.Field(i).Name == key {

				// check for nesting through type assertion
				nestedMap, nested := fields[key].(map[string]interface{})

				if !nested {
					// NOT nested -> copy value directly
					m[newKey] = valInterface.Field(i).Interface()
				} else {

					// NESTED copied via recursion
					var slice = reflect.ValueOf(valInterface.Field(i).Interface())

					// if nested ARRAY
					if valInterface.Field(i).Kind() == reflect.Slice {
						sliceMapped := make([]interface{}, slice.Len())

						for i := 0; i < slice.Len(); i++ {
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
