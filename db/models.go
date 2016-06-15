package db

import (
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
	Location     Area `gorm:"ForeignKey:LocationId;AssociationForeignKey:Refer"`
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
	User           User
	Location       gormGIS.GeoPoint `gorm:"not null"`
	Tags           []Tag            `gorm:"many2many:offer_tags"`
	ValidityPeriod time.Time
	Matched        bool
	Expired        bool
}

type Request struct {
	ID             string `gorm:"primary_key"`
	Name           string `gorm:"index;not null"`
	User           User
	Location       gormGIS.GeoPoint `gorm:"not null"`
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

type Area struct {
	ID          string `gorm:"primary_key"`
	Name        string
	Boundaries  GeoPolygon `sql:"type:geometry(Geometry,4326)"`
	Description string
	Offers      []Offer   `gorm:"many2many:area_offers"`
	Requests    []Request `gorm:"many2many:area_requests"`
}
