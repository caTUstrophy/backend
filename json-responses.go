package main

var fieldsUser = map[string]interface{}{
	"ID":            "ID",
	"Name":          "Name",
	"PreferredName": "PreferredName",
	"Mail":          "Mail",
	"MailVerified":  "MailVerified",
	"PhoneNumbers":  "PhoneNumbers",
	"Groups": map[string]interface{}{
		"ID": "ID",
		"Region": map[string]interface{}{
			"ID":          "ID",
			"Description": "Description",
			"Name":        "Name",
		},
		"AccessRight": "AccessRight",
		"Description": "Description",
	},
}

var fieldsUserNoGroups = map[string]interface{}{
	"ID":           "ID",
	"Name":         "Name",
	"Mail":         "Mail",
	"MailVerified": "MailVerified",
	"PhoneNumbers": "PhoneNumbers",
}

var fieldsSystemAdmin = map[string]interface{}{
	"ID":            "ID",
	"Name":          "Name",
	"PreferredName": "PreferredName",
	"Mail":          "Mail",
	"MailVerified":  "MailVerified",
	"PhoneNumbers":  "PhoneNumbers",
}

var fieldsGroup = map[string]interface{}{
	"ID":          "ID",
	"RegionId":    "RegionId",
	"Region":      "Region",
	"Users":       fieldsUserNoGroups,
	"AccessRight": "AccessRight",
	"Description": "Description",
}

var fieldsRequestWithUser = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Location": map[string]interface{}{
		"Lng": "lng",
		"Lat": "lat",
	},
	"Tags": map[string]interface{}{
		"Name": "Name",
	},
	"ValidityPeriod": "ValidityPeriod",
	"Matched":        "Matched",
	"Expired":        "Expired",
	"User": map[string]interface{}{
		"ID":           "ID",
		"Name":         "Name",
		"Mail":         "Mail",
		"PhoneNumbers": "PhoneNumbers",
	},
}

var fieldsRequest = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Location": map[string]interface{}{
		"Lng": "lng",
		"Lat": "lat",
	},
	"Tags": map[string]interface{}{
		"Name": "Name",
	},
	"ValidityPeriod": "ValidityPeriod",
	"Matched":        "Matched",
	"Expired":        "Expired",
}

var fieldsOfferWithUser = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Location": map[string]interface{}{
		"Lng": "lng",
		"Lat": "lat",
	},
	"Tags": map[string]interface{}{
		"Name": "Name",
	},
	"ValidityPeriod": "ValidityPeriod",
	"Matched":        "Matched",
	"Expired":        "Expired",
	"User": map[string]interface{}{
		"ID":           "ID",
		"Name":         "Name",
		"Mail":         "Mail",
		"PhoneNumbers": "PhoneNumbers",
	},
}

var fieldsOffer = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Location": map[string]interface{}{
		"Lng": "lng",
		"Lat": "lat",
	},
	"Tags": map[string]interface{}{
		"Name": "Name",
	},
	"ValidityPeriod": "ValidityPeriod",
	"Matched":        "Matched",
	"Expired":        "Expired",
}

var fieldsRegion = map[string]interface{}{
	"ID":   "ID",
	"Name": "Name",
	"Boundaries": map[string]interface{}{
		"Points": map[string]interface{}{
			"Lng": "lng",
			"Lat": "lat",
		},
	},
	"Description": "Description",
}

var fieldsMatching = map[string]interface{}{
	"ID":       "ID",
	"RegionId": "RegionId",
	"Request":  fieldsRequestWithUser,
	"Offer":    fieldsOfferWithUser,
	"Invalid":  "Invalid",
}

var fieldsNotificationWithRead = map[string]interface{}{
	"ID":     "ID",
	"Type":   "Type",
	"ItemID": "ItemID",
	"Read":   "Read",
}

var fieldsNotification = map[string]interface{}{
	"ID":     "ID",
	"Type":   "Type",
	"ItemID": "ItemID",
}

var fieldsNotificationForMatching = map[string]interface{}{
	"ID":       "ID",
	"Type":     "Type",
	"ItemID":   "ItemID",
	"Matching": fieldsMatching,
}

// Other global response schemes, not thought to be used for CopyNestedModel

var notFound = map[string]interface{}{
	"Error": "Requested item does not exist in database",
}
