package controllers

import (
	"errors"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
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
	adminEmail, exist := c.GetQuery("email")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "missing field email",
		})
		return
	}
	// Check if email exists in the admin table
	var admin model.Admin
	if tx := database.DB.Where("email = ?", adminEmail).First(&admin); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Email not present in the admin table",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "Database error",
			})
			return
		}
	}

	// Check if email exists in the verification table with admin role
	var verification model.VerificationTable
	if tx := database.DB.Where("email = ? AND role = ?", adminEmail, model.AdminRole).First(&verification); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			// Create a new entry in the verification table
			verification = model.VerificationTable{
				Email:              adminEmail,
				Role:               model.AdminRole,
				VerificationStatus: model.VerificationStatusPending,
			}
			if err := database.DB.Create(&verification).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  false,
					"message": "Failed to create admin verification entry",
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "Database error",
			})
			return
		}
	}

	// Send OTP
	err := SendOTP(c, adminEmail, verification.OTPExpiry, verification.Role)
	if err != nil {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Verification link sent successfully. Please verify via that. Link expires soon .",
	})
}
