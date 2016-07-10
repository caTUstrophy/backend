package tfidf

import (
	"log"
	"strings"
)

// Functions

func ComputeTF(words []string) map[string]int {

	// TODO
	tfMap := make(map[string]int)

	return tfMap
}

func ComputeIDF() {
	// TODO
}

func ComputeTFIDF(text string) []float64 {

	// Should we preprocess the input text here?
	// E.g.:
	// - Remove punctuation
	// - Run the Snowball stemmer algorithm on text

	// Split text at spaces.
	words := strings.Split(text, " ")

	// Compute text's tf vector as a map representation.
	tfMap := make(map[string]int)
	tfMap = ComputeTF(words)

	log.Println(tfMap)

	// Compute the idf vector.

	// Compute the final tf-idf vector.
	tfidf := []float64{1.0}

	return tfidf
}
