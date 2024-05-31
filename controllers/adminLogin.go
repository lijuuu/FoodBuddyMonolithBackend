package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckAdmin(c *gin.Context) {
	email := utils.GetJWTEmailClaim(c)

	if err := VerifyJWT(c, model.AdminRole, email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "access denied, request is unauthorized",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "authorized request, proceed to login",
	})
	c.Next()
}

func AdminLogin(c *gin.Context) {
	//get the json from the request
	var form struct {
		Email string
	}
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process the incoming request",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	//validate the content of the json
	err := validate(form)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	//update the otp and expiry
	var VerificationTable model.VerificationTable

	if err := database.DB.Where("email = ? AND role = ?", form.Email, model.AdminRole).First(&VerificationTable).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
		})
		return
	}

	//sendotp
	err = SendOTP(c, form.Email, VerificationTable.OTPExpiry, VerificationTable.Role)
	if err != nil {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusAlreadyReported,
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  false,
		"message": "OTP is send successfully",
	})
}
