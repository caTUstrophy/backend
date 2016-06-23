// Based on: http://www.booneputney.com/development/gorm-golang-jsonb-value-copy/
// Therefore credit is due to Boone Putney. Thank you!

package db

import (
	"fmt"
	"strings"

	"database/sql/driver"
	"encoding/json"
)

type PhoneNumbers []string

func (num *PhoneNumbers) Scan(value interface{}) error {

	// Start of JSON string to be unmarshalled later.
	jsonByteString := "[ "

	// Make sure we operate on a string slice.
	numbers := value.([]string)

	// Range over supplied numbers and append each to JSON string.
	for _, n := range numbers {
		jsonByteString = fmt.Sprintf("%s\"%s\", ", jsonByteString, n)
	}

	// Remove trailing comma and space and replace by closing bracket of JSON string.
	jsonByteString = strings.TrimRight(jsonByteString, ", ")
	jsonByteString = fmt.Sprintf("%s ]", jsonByteString)

	// Attempt to unmarshal into type.
	if err := json.Unmarshal([]byte(jsonByteString), &num); err != nil {
		return err
	}

	return nil
}

func (num PhoneNumbers) Value() (driver.Value, error) {

	valueString, err := json.Marshal(num)
	if err != nil {
		return nil, err
	}

	return string(valueString), nil
}
