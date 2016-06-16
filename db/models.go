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
	Regions        []Region         `gorm:"many2many:region_offers"`
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
	Regions        []Region         `gorm:"many2many:region_requests"`
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
