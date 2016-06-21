package main

var fieldsRequestU = map[string]interface{}{
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
		"ID":   "ID",
		"Name": "Name",
		"Mail": "Mail",
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

var fieldsOfferU = map[string]interface{}{
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
		"ID":   "ID",
		"Name": "Name",
		"Mail": "Mail",
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
	"ID":        "ID",
	"RegionId":  "RegionId",
	"OfferId":   "OfferId",
	"RequestId": "RequestId",
}

var fieldsUser = map[string]interface{}{
	"ID":            "ID",
	"Name":          "Name",
	"PreferredName": "PreferredName",
	"Mail":          "Mail",
	"MailVerified":  "MailVerified",
	"Groups": map[string]interface{}{
		"ID": "ID",
		"Permissions": map[string]interface{}{
			"AccessRight": "AccessRight",
		},
	},
}
