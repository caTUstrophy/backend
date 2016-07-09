package main

import (
	"math"

	"github.com/caTUstrophy/backend/db"
)

// Calculates a matching score between 0 and 100 for a pair of offer and request
func GetScore(request db.Request, offer db.Offer) float64 {
	return GetDistanceFactor(request, offer) * GetContentFactor(request, offer)
}

// Calculates a distance score between 0 and 10 for a pair of offer and request
func GetDistanceFactor(request db.Request, offer db.Offer) float64 {
	distance := Distance(GeoLocation{request.Location.Lng, request.Location.Lat}, GeoLocation{offer.Location.Lng, offer.Location.Lat})
	if distance > request.Radius+offer.Radius {
		return 0.0
	}

	return scale((request.Radius+offer.Radius)/distance, 1, 0, 10)
}

// Calculates a score between 0 and 10 for a pair of offer and request.
// This score shall be imagened as a probability that the offer fits the request
func GetContentFactor(request db.Request, offer db.Offer) float64 {
	return 0.0
}

func scale(value, currMin, supposedFrom, supposedTo float64) float64 {
	return (math.Tanh(value-2*currMin)+1)*(supposedTo-supposedFrom)/2 + supposedFrom
}
