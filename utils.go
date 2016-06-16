package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

func (app *App) getUUID(c *gin.Context, par string) string {
	parID := c.Params.ByName(par)
	errs := app.Validator.Field(parID, "required,uuid4")
	if errs != nil {

		errResp := make(map[string]string)

		// Iterate over all validation errors.
		for _, err := range errs.(validator.ValidationErrors) {

			if err.Tag == "required" {
				errResp[err.Field] = "Is required"
			} else if err.Tag == "uuid4" {
				errResp[err.Field] = "Needs to be an UUID version 4"
			}
		}

		// Send prepared error message to client.
		c.JSON(http.StatusBadRequest, errResp)
		return ""
	}
	return parID
}

func CopyNestedModel(i interface{}, fields map[string]interface{}) map[string]interface{} {
	var m map[string]interface{}
	m = make(map[string]interface{})

	// get value + type of source interface
	valInterface := reflect.ValueOf(i)
	typeOfT := reflect.ValueOf(i).Type()

	// iterate over all fields that will be copied
	for key := range fields {
		var exists = false
		newKey, _ := fields[key].(string)

		// search for field in source type
		for i := 0; i < valInterface.NumField(); i++ {
			if typeOfT.Field(i).Name == key {

				// check for nesting through type assertion
				nestedMap, nested := fields[key].(map[string]interface{})

				if !nested {
					// NOT nested -> copy value directly
					m[newKey] = valInterface.Field(i).Interface()
				} else {

					// NESTED copied via recursion
					var slice = reflect.ValueOf(valInterface.Field(i).Interface())

					// if nested ARRAY
					if valInterface.Field(i).Kind() == reflect.Slice {
						sliceMapped := make([]interface{}, slice.Len())

						for i := 0; i < slice.Len(); i++ {
							sliceMapped[i] = CopyNestedModel(slice.Index(i).Interface(), nestedMap)
						}
						m[key] = sliceMapped
					} else {
						// if nested OBJECT
						m[key] = CopyNestedModel(valInterface.Field(i).Interface(), nestedMap)
					}
				}

				exists = true
				break
			}
		}

		if !exists {
			panic(fmt.Sprintf("ERROR: Struct<%s> has no field: %s", typeOfT.Name(), key))
		}
	}

	return m
}
