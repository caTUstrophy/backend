package db

import (
	"time"

	"github.com/nferruzzi/gormGIS"
)

// Constants

const (
	NotificationMatching string = "matching"
	// Place for more, future notification types.
	// Add them like e.g.:
	// NotificationPromotion string = "promotion"
)

// Models

type Group struct {
	ID           string `gorm:"primary_key"`
	DefaultGroup bool   `gorm:"not null"`
	Region       Region `gorm:"ForeignKey:RegionId;AssociationForeignKey:Refer"`
	RegionId     string `gorm:"index;not null"`
	Users        []User `gorm:"many2many:user_groups"`
	AccessRight  string
	Description  string
}

type User struct {
	ID            string       `gorm:"primary_key"`
	Name          string       `gorm:"not null"`
	PreferredName string       `gorm:"not null"`
	Mail          string       `gorm:"index;not null;unique"`
	MailVerified  bool         `gorm:"not null"`
	PhoneNumbers  PhoneNumbers `gorm:"not null" sql:"type:jsonb"`
	PasswordHash  string       `gorm:"not null;unique"`
	Groups        []Group      `gorm:"many2many:user_groups"`
	Enabled       bool         `gorm:"not null"`
}

type Tag struct {
	Name string `gorm:"primary_key"`
}

type Offer struct {
	ID             string           `gorm:"primary_key"`
	Name           string           `gorm:"index;not null"`
	User           User             `gorm:"ForeignKey:UserID;AssociationForeignKey:Refer"`
	UserID         string           `gorm:"index;not null"`
	Location       gormGIS.GeoPoint `gorm:"not null" sql:"type:geometry(Geometry,4326)"`
	Tags           []Tag            `gorm:"many2many:offer_tags"`
	Regions        []Region         `gorm:"many2many:region_offers"`
	ValidityPeriod time.Time        `gorm:"not null"`
	Matched        bool             `gorm:"not null"`
	Expired        bool             `gorm:"not null"`
}

type Request struct {
	ID             string           `gorm:"primary_key"`
	Name           string           `gorm:"index;not null"`
	User           User             `gorm:"ForeignKey:UserID;AssociationForeignKey:Refer"`
	UserID         string           `gorm:"index;not null"`
	Location       gormGIS.GeoPoint `gorm:"not null" sql:"type:geometry(Geometry,4326)"`
	Tags           []Tag            `gorm:"many2many:request_tags"`
	Regions        []Region         `gorm:"many2many:region_requests"`
	ValidityPeriod time.Time        `gorm:"not null"`
	Matched        bool             `gorm:"not null"`
	Expired        bool             `gorm:"not null"`
}

type Matching struct {
	ID        string  `gorm:"primary_key"`
	Region    Region  `gorm:"ForeignKey:RegionId;AssociationForeignKey:Refer"`
	RegionId  string  `gorm:"index;not null"`
	Offer     Offer   `gorm:"ForeignKey:OfferId;AssociationForeignKey:Refer"`
	OfferId   string  `gorm:"index;not null"`
	Request   Request `gorm:"ForeignKey:RequestId;AssociationForeignKey:Refer"`
	RequestId string  `gorm:"index;not null"`
	Invalid   bool    `gorm:"not null"`
}

type Region struct {
	ID          string     `gorm:"primary_key"`
	Name        string     `gorm:"not null"`
	Boundaries  GeoPolygon `gorm:"not null" sql:"type:geometry(Geometry,4326)"`
	Description string     `gorm:"not null"`
	Matchings   []Matching
	Offers      []Offer   `gorm:"many2many:region_offers"`
	Requests    []Request `gorm:"many2many:region_requests"`
}

type Notification struct {
	ID        string    `gorm:"primary_key"`
	Type      string    `gorm:"index;not null"`
	UserID    string    `gorm:"not null"`
	ItemID    string    `gorm:"not null"`
	Read      bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

// Helpers

// Make tags model sortable.
type TagsByName []Tag

func (t TagsByName) Len() int           { return len(t) }
func (t TagsByName) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TagsByName) Less(i, j int) bool { return t[i].Name < t[j].Name }
