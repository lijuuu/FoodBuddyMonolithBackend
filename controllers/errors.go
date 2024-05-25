package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func validate(value interface{}, c *gin.Context) bool {
	var translator = map[string]string{
		"Name_required":            "Please enter your Name",
		"Password_required":        "Please enter your Password",
		"ConfirmPassword_required": "Please enter your ConfirmPassword",
		"Email_email":              "Please enter a valid email address", 
		"UserID_number":"Please enter a valid user id",
		"ProductID_number":"Please enter a valid product id",
	}
	// validate the struct body
	validate := validator.New()
	err := validate.Struct(value)
	if err != nil {
		var errs []string
		for _, e := range err.(validator.ValidationErrors) {
			translationKey := e.Field() + "_" + e.Tag()
			errMsg := translator[translationKey]
			if errMsg == "" {
				errMsg = e.Error()
			}
			errs = append(errs, errMsg)
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": errs,
			"ok":    false,
		})
		return false
	}
	return true
}
