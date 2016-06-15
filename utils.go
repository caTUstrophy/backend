package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

func getUUID(c *gin.Context, par string) {
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

		return
	}
	return parID
}
