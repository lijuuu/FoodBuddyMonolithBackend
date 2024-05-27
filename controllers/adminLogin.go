package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	var AdminLogin model.Admin

	if err := c.BindJSON(&AdminLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind json",
			"ok":    false,
		})
	}

	if err := database.DB.Where("email = ?", AdminLogin.Email).First(&AdminLogin).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get admin info from the database",
			"ok":    false,
		})
		return
	}

	tokenstring := GenerateJWT(c, AdminLogin.Email)
	if tokenstring == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to create jwt",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admin": AdminLogin,
		"ok":    true,
	})

}

func CheckAdmin(c *gin.Context) {
	email := utils.GetJWTEmailClaim(c)

	var AdminLogin model.Admin
	if err := database.DB.Where("email = ?", email).First(&AdminLogin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "the email doesnt exist in database, unauthorized user",
			"ok":    false,
		})
		c.Abort()
		return
	}

	

	c.JSON(http.StatusOK, gin.H{
		"message": "user is an admin",
		"ok":    true,
	})
	c.Next()
}
