package main

import (
	"github.com/caTUstrophy/backend/db"
)

// Calculates a matching score between 0 and 100 for a pair of offer and request
func GetScore(request db.Request, offer db.Offer) float64 {
	return GetDistanceFactor(request, offer) * GetContentFactor(request, offer)
}

// Calculates a distance score between 0 and 10 for a pair of offer and request
func GetDistanceFactor(request db.Request, offer db.Offer) float64 {
	return 0.0
}

// Calculates a score between 0 and 10 for a pair of offer and request.
// This score shall be imagened as a probability that the offer fits the request
func GetContentFactor(request db.Request, offer db.Offer) float64 {
	return 0.0
}
