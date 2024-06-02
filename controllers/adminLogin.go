package controllers

import (
	"errors"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	c.Next()
}

// AdminLogin godoc
// @Summary Admin login
// @Description Login an admin using email
// @Tags authentication
// @Accept json
// @Produce json
// @Param AdminLogin body model.AdminLoginRequest true "Admin Login"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/admin/login [post]
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
			"data":       gin.H{},
		})
		return
	}

	// Validate the content of the JSON
	if err := validate(form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
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
				"data":       gin.H{},
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Database error",
				"error_code": http.StatusInternalServerError,
				"data":       gin.H{},
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
					"data":       gin.H{},
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Database error",
				"error_code": http.StatusInternalServerError,
				"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Verification link sent successfully. Please verify via that. Link expires soon .",
	})
}
