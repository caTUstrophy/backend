package main

import (
	"math"

	"github.com/caTUstrophy/backend/db"
	"github.com/caTUstrophy/backend/tfidf"
)

// Functions

// Returns the similarity between the tags' and the requests'
// tag sets. Result is normalized to be within [0, 1].
func CalculateTagSimilarity(tagChannel chan float64, offerTags, requestTags []db.Tag) {

	var exp float64
	exp = 2.0 / 3.0

	// Initialize an empty list to add intersect and union elements to.
	tagsIntersection := make([]db.Tag, 0)
	tagsUnion := make([]db.Tag, 0)

	// Maintain a lookup union map: fast existence check.
	tagsUnionMap := make(map[string]bool)

	// Maintain a lookup request's tags map: fast existence check.
	requestTagsMap := make(map[string]bool)
	for _, requestTag := range requestTags {
		requestTagsMap[requestTag.Name] = true
	}

	// Iterate over offer's tags list and check for intersection.
	for _, tag := range offerTags {

		// If existing in both, add to intersection.
		if requestTagsMap[tag.Name] {
			tagsIntersection = append(tagsIntersection, tag)
		}

		// Add tag to union and update lookup map.
		tagsUnion = append(tagsUnion, tag)
		tagsUnionMap[tag.Name] = true
	}

	for _, tag := range requestTags {

		if !tagsUnionMap[tag.Name] {

			// Add tag to union and update lookup map.
			tagsUnion = append(tagsUnion, tag)
			tagsUnionMap[tag.Name] = true
		}
	}

	// Calculate similarity and normalize it to be within [0, 1].
	tagSimilarity := float64(len(tagsIntersection)) / math.Pow(float64(len(tagsUnion)), exp)
	tagSimilarity = scale(tagSimilarity, 2, 0.5, 0, 1)

	// Pass result into tag channel.
	tagChannel <- tagSimilarity
}

// Returns the similarity in terms of free text description fields
// of offer and request. Result is normalized to be within [0, 1].
func CalculateDescriptionSimilarity(descChannel chan float64, offerDesc, requestDesc string) {

	// Compute tf-idf vector for offer's description field.
	_ = tfidf.ComputeTFIDF(offerDesc)

	// Compute tf-idf vector for request's description field.
	_ = tfidf.ComputeTFIDF(requestDesc)

	// Compute cosine similarity between both tf-idf vectors.
	descSimilarity := 0.75

	// Normalize similarity value to be within [0, 1].
	descSimilarity = scale(descSimilarity, 1, 1, 0, 1)

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
		distChannel <- scale(((request.Radius + offer.Radius) / distance), 1, 1, 0, 10)
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

		// Preload needed tags.
		app.DB.Preload("Tags").Find(&Region.Requests)

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

		// Preload needed tags.
		app.DB.Preload("Tags").Find(&Region.Offers)

		for _, offer := range Region.Offers {

			// Calculate pair wise matching score between
			// request and all offers in region.
			go app.CalculateMatchingScore(Region, offer, request)
		}
	}
}
