package controllers

import (
	"log"
	"net/http"

	"foodbuddy/internal/database"
	"foodbuddy/internal/utils"
	"foodbuddy/internal/model"

	"github.com/gin-gonic/gin"
)

func UserProfileImageUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Failed to get multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get multipart form"})
		return
	}

	// Assuming there's only one file uploaded
	fileHeaders := form.File["file"] // Replace "yourFieldName" with the actual field name used in the form
	if len(fileHeaders) > 0 {
		fileHeader := fileHeaders[0]
		imageURL, err := utils.ImageUpload(fileHeader)
		if err != nil {
			log.Printf("Error uploading image: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		// fmt.Println(imageURL)
		// c.Redirect(http.StatusMovedPermanently, imageURL)
		email, role, err := utils.GetJWTClaim(c)
		if role != model.UserRole || err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "unauthorized request",
			})
			return
		}
		UserID, _ := UserIDfromEmail(email)
		if err:=database.DB.Model(&model.User{}).Where("id = ?",UserID).Update("picture",imageURL).Error;err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"status":false,"message":"failed to add image, please try again"+err.Error()})
		}

		c.JSON(http.StatusOK,gin.H{"status":"success","message":"profile image uploaded"})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
	}
}

func RestaurantProfileImageUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Failed to get multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get multipart form"})
		return
	}

	// Assuming there's only one file uploaded
	fileHeaders := form.File["file"] // Replace "yourFieldName" with the actual field name used in the form
	if len(fileHeaders) > 0 {
		fileHeader := fileHeaders[0]
		imageURL, err := utils.ImageUpload(fileHeader)
		if err != nil {
			log.Printf("Error uploading image: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		// fmt.Println(imageURL)
		// c.Redirect(http.StatusMovedPermanently, imageURL)
		email, role, err := utils.GetJWTClaim(c)
		if role != model.RestaurantRole || err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "unauthorized request",
			})
			return
		}
		RestID, _ := RestIDfromEmail(email)
		if err:=database.DB.Model(&model.Restaurant{}).Where("id = ?",RestID).Update("image_url",imageURL).Error;err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"status":false,"message":"failed to add image, please try again"+err.Error()})
		}

		c.JSON(http.StatusOK,gin.H{"status":"success","message":"profile image uploaded"})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
	}
}
