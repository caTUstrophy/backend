package main

import (
	"fmt"
	"math"

	"github.com/caTUstrophy/backend/db"
	"github.com/caTUstrophy/munkres"
	"github.com/numbleroot/go-tfidf"
)

// Functions

/*
// This is helpful for developing the matching algorithm but mustn't be available in deployment
func (app *App) getScoreMatrix(c *gin.Context) {

	var recommendations []db.MatchingScore
	app.DB.Find(&recommendations)

	for i, rec := range recommendations {
		app.DB.Model(&rec).Related(&recommendations[i].Offer)
		app.DB.Model(&rec).Related(&recommendations[i].Request)
	}

	model := CopyNestedModel(recommendations, fieldsRecommendations)

	c.JSON(200, model)
}
*/

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
	numIntersect := len(tagsIntersection)
	numUnion := min(len(tagsUnion), 1)
	tagSimilarity := float64(numIntersect) / math.Pow(float64(numUnion), exp)
	tagSimilarity = scale(tagSimilarity, 2, 0.5, 0, 1)

	// Pass result into tag channel.
	if math.IsNaN(tagSimilarity) {
		tagChannel <- 0.5
	}
	tagChannel <- tagSimilarity
}

// Returns the similarity in terms of free text description fields
// of offer and request. Result is normalized to be within [0, 1].
func CalculateDescriptionSimilarity(descChannel chan float64, offerDesc, requestDesc string) {

	// Own implementation:
	tokOfferDesc := tfidf.TokenizeDocument(offerDesc)
	tokRequestDesc := tfidf.TokenizeDocument(requestDesc)
	docs := [][]string{tokOfferDesc, tokRequestDesc}

	tfOffer := tfidf.TermFrequencies(tokOfferDesc, docs)
	tfRequest := tfidf.TermFrequencies(tokRequestDesc, docs)

	// Compute cosine similarity between both tf vectors.
	descSimilarity := 4 * math.Pow(cosineSimilarity(tfOffer, tfRequest), 2)
	descSimilarity = math.Min(1, descSimilarity)

	// Pass result into description channel.
	if math.IsNaN(descSimilarity) {
		descChannel <- 0.5
	}
	descChannel <- descSimilarity
}

// Returns the geometric distance between the offer's and the request's
// location. Result is normalized to be within [0, 10].
func CalculateLocationDistance(distChannel chan float64, offer db.Offer, request db.Request) {

	// Calculate distance between offer's and request's location.
	distance := distance(offer.Location, request.Location)
	if distance == 0.0 {
		distChannel <- 10.0
	}

	// Depending on result, pass normalized distance into channel.
	if distance > (request.Radius + offer.Radius) {
		distChannel <- 0.0
	} else {
		loc := scale(((request.Radius + offer.Radius) / distance), 1, 1, 0, 10)
		if math.IsNaN(loc) {
			distChannel <- 5.0
		}
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
	if math.IsNaN(finalScore) {
		finalScore = 20
	}
	fmt.Printf("Request: %s\n  Offer: %s \n", request.Name, offer.Name)
	fmt.Printf("((alpha: %f * tagSimilarity: %f) + (beta: %f * descSimilarity: %f)) * locDistance: %f = %f\n", app.TagsWeightAlpha, tagSimilarity, app.DescWeightBeta, descSimilarity, locDistance, finalScore)

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

	if len(offer.Regions) == 0 {
		app.DB.Preload("Regions").First(&offer, "id = ?", offer.ID)
	}

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

		// Tell database that region has an updated recommendation calculation.
		app.DB.Model(&Region).Update("RecommendationUpdated", false)
	}
}

// Wrapper function to simplify usage of main calculation method.
func (app *App) CalcMatchScoreForRequest(request db.Request) {
	if len(request.Regions) == 0 {
		app.DB.Preload("Regions").First(&request, "id = ?", request.ID)
	}

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

		// Tell database that region has an updated recommendation calculation.
		app.DB.Model(&Region).Update("RecommendationUpdated", false)
	}
}

// Caclulate assignment problem for offers und requests of this region
// and set recommended flag to matching scores.
func (app *App) RecommendMatching(region db.Region) {

	// load all scores for this region
	var scores []db.MatchingScore
	app.DB.Order("request_id, offer_id").Find(&scores, "region_id = ?", region.ID)
	// get number of offers and request in order to get the matrix size
	numOffers := 0
	numRequests := 0
	app.DB.Raw("SELECT COUNT (*) FROM region_requests WHERE region_id = '" + region.ID + "'").Row().Scan(&numRequests)
	app.DB.Raw("SELECT COUNT (*) FROM region_offers WHERE region_id = '" + region.ID + "'").Row().Scan(&numOffers)

	// Check db for inkonsistence and try to recover it if necessary. Debug states can stay as this implies something went wrong before
	if numRequests*numOffers != len(scores) {
		fmt.Println("Inkonsistent data in DB! In region ", region.Name, " is the number of matching scores not as expected. Calculate all new :(")
		app.DB.Delete(&db.MatchingScore{}, "region_id = ?", region.ID)
		app.DB.Preload("Offers.Tags").Preload("Offers").Preload("Requests.Tags").Preload("Requests").First(&region, "id = ?", region.ID)
		for _, offer := range region.Offers {
			app.MapLocationToRegions(offer)
		}
		for _, request := range region.Requests {
			app.MapLocationToRegions(request)
			for _, offer := range region.Offers {
				fmt.Println("Calculate for ", offer.Name, "/", offer.ID, "and ", request.Name, "/", request.ID)
				app.CalculateMatchingScore(region, offer, request)
			}
		}
		app.DB.Order("request_id, offer_id").Find(&scores, "region_id = ?", region.ID)
		app.DB.Raw("SELECT COUNT (*) FROM region_requests WHERE region_id = '" + region.ID + "'").Row().Scan(&numRequests)
		app.DB.Raw("SELECT COUNT (*) FROM region_offers WHERE region_id = '" + region.ID + "'").Row().Scan(&numOffers)
	}
	if numRequests*numOffers != len(scores) {
		fmt.Println("Could not fix inkonsistent data. Please contact more skilled developers\n\nNumRequests: ", numRequests, "\nnumOffers: ", numOffers, "num scores: ", len(scores))
		panic("Inkonsistent data could not be fixed")
	}

	size := Max(numOffers, numRequests)

	scoreValues := make([]int64, len(scores))
	for i, score := range scores {

		scoreValues[i] = 100 - int64(score.MatchingScore)

	}
	// create dummy rows and cols; rows: request; cols: offers

	scoreMatrixArray := make([]int64, size*size)

	for row := 0; row < size; row++ {

		for col := 0; col < size; col++ {

			if row < numRequests && col < numOffers {
				scoreMatrixArray[row*size+col] = scoreValues[row*numOffers+col]
			} else {
				scoreMatrixArray[row*size+col] = 100
			}
		}
	}

	// create Matrix and solve assignment problem
	m := munkres.NewMatrix(size)
	m.A = scoreMatrixArray
	solution := munkres.ComputeMunkresMin(m)

	// save recommendations to db
	for _, recommendation := range solution {

		if recommendation.Row < numRequests && recommendation.Col < numOffers {
			index := recommendation.Row*numOffers + recommendation.Col
			scores[index].Recommended = true
			app.DB.Model(&scores[index]).Update("recommended", true)
		}
	}

	// Save that this region has up to date recommendations
	app.DB.Model(&region).Update("RecommendationUpdated", true)
}
