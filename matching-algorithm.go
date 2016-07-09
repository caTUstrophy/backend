package main

import (
	"github.com/caTUstrophy/backend/db"
)

// Functions

// Returns the similarity between the tags' and the requests'
// tag sets. Result is normalized to be within [0, 1].
func CalculateTagSimilarity(tagChannel chan float64, offerTags, requestTags []db.Tag) {

	// Normalize similarity value to be within [0, 1].
	tagSimilarity := 0.75

	// Pass result into tag channel.
	tagChannel <- tagSimilarity
}

// Returns the similarity in terms of free text description fields
// of offer and request. Result is normalized to be within [0, 1].
func CalculateDescriptionSimilarity(descChannel chan float64, offerDesc, requestDesc string) {

	// Normalize similarity value to be within [0, 1].
	descSimilarity := 0.75

	// Pass result into description channel.
	descChannel <- descSimilarity
}

// Returns the geometric distance between the offer's and the request's
// location. Result is normalized to be within [0, 10].
func CalculateLocationDistance(distChannel chan float64, offer db.Offer, request db.Request) {

	// Calculate distance between offer's and request's location.
	distance := distance(offer.Location, request.Location)

	// Depending on result, pass normalized distance into channel.
	if distance > (request.Radius + offer.Radius) {
		distChannel <- 0.0
	} else {
		distChannel <- scale((request.Radius+offer.Radius)/distance, 1, 0, 10)
	}
}

// This function calculates the possible matching score
// between an offer and a request in a specified region.
func (app *App) CalculateMatchingScore(region db.Region, offer db.Offer, request db.Request) {

	// Create channels for synchronization of goroutines.
	tagChannel := make(chan float64)
	descChannel := make(chan float64)
	distChannel := make(chan float64)

	// Reserve space for result variables.
	var tagSimilarity float64
	var descSimilarity float64
	var locDistance float64

	// In a goroutine: Calculate the tag similarity between offer and request.
	go CalculateTagSimilarity(tagChannel, offer.Tags, request.Tags)

	// In a goroutine: Calculate the text distance between the offer's and
	// the request's free text description fields.
	go CalculateDescriptionSimilarity(descChannel, offer.Description, request.Description)

	// In a goroutine: Calculate the geometric distance between the offer's
	// and the request's location fields.
	go CalculateLocationDistance(distChannel, offer, request)

	// Wait until all goroutines have finished.
	for i := 0; i < 3; i++ {

		select {
		case firstValue := <-tagChannel:
			tagSimilarity = firstValue
		case secondValue := <-descChannel:
			descSimilarity = secondValue
		case thirdValue := <-distChannel:
			locDistance = thirdValue
		}
	}

	// Compute the content related similarity of offer and request
	// with weighted tag and description similarity.
	contentSimilarity := (app.TagsWeightAlpha * tagSimilarity) + (app.DescWeightBeta * descSimilarity)

	// Final score is the product of content similarity and distance.
	finalScore := contentSimilarity * locDistance

	MatchingScore := db.MatchingScore{
		RegionID:      region.ID,
		Region:        region,
		OfferID:       offer.ID,
		Offer:         offer,
		RequestID:     request.ID,
		Request:       request,
		MatchingScore: finalScore,
	}

	// Check if matching score element already exists in table.
	var TmpMatchingScore db.MatchingScore
	TmpMatchingScore.MatchingScore = -1.0
	app.DB.First(&TmpMatchingScore, "\"matching_scores\".\"region_id\" = ? AND \"matching_scores\".\"offer_id\" = ? AND \"matching_scores\".\"request_id\" = ?", region.ID, offer.ID, request.ID)

	if TmpMatchingScore.MatchingScore != -1.0 {
		// If it exists, update it.
		app.DB.Save(&MatchingScore)
	} else {
		// If it does not exist, create it.
		app.DB.Create(&MatchingScore)
	}
}

// Wrapper function to simplify usage of main calculation method.
func (app *App) CalcMatchScoreForOffer(offer db.Offer) {

	for _, Region := range offer.Regions {

		// Load all requests in this region.
		app.DB.Preload("Requests").First(&Region)

		for _, request := range Region.Requests {

			// Calculate pair wise matching score between
			// offer and all requests in region.
			go app.CalculateMatchingScore(Region, offer, request)
		}
	}
}

// Wrapper function to simplify usage of main calculation method.
func (app *App) CalcMatchScoreForRequest(request db.Request) {

	for _, Region := range request.Regions {

		// Load all offers in this region.
		app.DB.Preload("Offers").First(&Region)

		for _, offer := range Region.Offers {

			// Calculate pair wise matching score between
			// request and all offers in region.
			go app.CalculateMatchingScore(Region, offer, request)
		}
	}
}
