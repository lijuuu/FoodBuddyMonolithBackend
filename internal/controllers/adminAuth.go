package controllers

import (
	"errors"
	"foodbuddy/internal/database"
	"foodbuddy/internal/utils"
	"foodbuddy/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CheckAdmin(c *gin.Context) (string, error) {
	email, _, err := utils.GetJWTClaim(c)
	if err != nil {
		return email, errors.New("request unauthorized")
	}

	if email == "" {
		return email, errors.New("request unauthorized")
	}

	if err := VerifyJWT(c, model.AdminRole, email); err != nil {
		return email, errors.New("request unauthorized")
	}
	return email, nil
}

func AdminLogin(c *gin.Context) {
	// Get the email from the JSON request
	var form struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "Failed to process the incoming request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Validate the content of the JSON
	if err := utils.Validate(form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Check if email exists in the admin table
	var admin model.Admin
	if tx := database.DB.Where("email = ?", form.Email).First(&admin); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":     false,
				"message":    "Email not present in the admin table",
				"error_code": http.StatusUnauthorized,
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Database error",
				"error_code": http.StatusInternalServerError,
			})
			return
		}
	}

	// Check if email exists in the verification table with admin role
	var verification model.VerificationTable
	if tx := database.DB.Where("email = ? AND role = ?", form.Email, model.AdminRole).First(&verification); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			// Create a new entry in the verification table
			verification = model.VerificationTable{
				Email:              form.Email,
				Role:               model.AdminRole,
				VerificationStatus: model.VerificationStatusPending,
			}
			if err := database.DB.Create(&verification).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":     false,
					"message":    "Failed to create admin verification entry",
					"error_code": http.StatusInternalServerError,
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Database error",
				"error_code": http.StatusInternalServerError,
			})
			return
		}
	}

	// Send OTP
	err := SendOTP(c, form.Email, verification.OTPExpiry, verification.Role)
	if err != nil {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusAlreadyReported,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Verification link sent successfully. Please verify via that. Link expires soon .",
	})
}
