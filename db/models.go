package db

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
	User           User
	Location       string `gorm:"index;not null"`
	Tags           []Tag  `gorm:"many2many:offer_tags"`
	ValidityPeriod int64
	Expired        bool
}

type Request struct {
	ID             string `gorm:"primary_key"`
	Name           string `gorm:"index;not null"`
	User           User
	Location       string `gorm:"index;not null"`
	Tags           []Tag  `gorm:"many2many:request_tags"`
	ValidityPeriod int64
	Expired        bool
}

type Matching struct {
	ID        string `gorm:"primary_key"`
	Offer     Offer  `gorm:"ForeignKey:OfferId;AssociationForeignKey:Refer"`
	OfferId   string
	Request   Request `gorm:"ForeignKey:RequestId;AssociationForeignKey:Refer"`
	RequestId string
}
