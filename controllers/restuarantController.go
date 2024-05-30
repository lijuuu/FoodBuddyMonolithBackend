package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetRestaurants(c *gin.Context) {
	var restaurants []model.Restaurant
	// Search db and get all
	if err := database.DB.Select("*").Find(&restaurants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    false,
			"message":   "failed to retrieve data from the database",
			"error_code": http.StatusInternalServerError,
			"data":      gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    true,
		"message":   "restaurants retrieved successfully",
		"data":      gin.H{"restaurantslist": restaurants},
	})
}

func AddRestaurant(c *gin.Context) {
	// Bind JSON
	var restaurant model.Restaurant
	if err := c.BindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"message":   "invalid request data",
			"error_code": http.StatusBadRequest,
			"data":      gin.H{},
		})
		return
	}

	// Validate the restaurant data
	if err := validate(restaurant); err!= nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"message":   err,
			"error_code": http.StatusBadRequest,
			"data":      gin.H{},
		})
		return
	}

	var existingRestaurant model.Restaurant
	if err := database.DB.Where("name =?", restaurant.Name).Find(&existingRestaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    false,
			"message":   "error while checking restaurant name",
			"error_code": http.StatusInternalServerError,
			"data":      gin.H{},
		})
		return
	}

	if restaurant.Name == existingRestaurant.Name {
		c.JSON(http.StatusConflict, gin.H{
			"status":    false,
			"message":   "restaurant already exists",
			"error_code": http.StatusConflict,
			"data":      gin.H{},
		})
		return
	}

	// Set blocked as false
	restaurant.Blocked = false

	// Add to db
	if err := database.DB.Create(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    false,
			"message":   "failed to add restaurant to the database",
			"error_code": http.StatusInternalServerError,
			"data":      gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    true,
		"message":   "successfully added new restaurant",
		"data":      gin.H{},
	})
}

func EditRestaurant(c *gin.Context) {
	// Bind JSON
	var restaurant model.Restaurant
	if err := c.BindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"message":   "failed to bind request",
			"error_code": http.StatusBadRequest,
			"data":      gin.H{},
		})
		return
	}

	// Check if present and update it with the new data
	var existingRestaurant model.Restaurant
	if err := database.DB.First(&existingRestaurant, restaurant.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":    false,
			"message":   "restaurant doesn't exist",
			"error_code": http.StatusNotFound,
			"data":      gin.H{},
		})
		return
	}

	// Edit the restaurant
	if err := database.DB.Model(&existingRestaurant).Updates(restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    false,
			"message":   "failed to edit the restaurant",
			"error_code": http.StatusInternalServerError,
			"data":      gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    true,
		"message":   "successfully edited the restaurant",
		"data":      gin.H{},
	})
}

func DeleteRestaurant(c *gin.Context) {
	// Get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"message":   "invalid restaurant ID",
			"error_code": http.StatusBadRequest,
			"data":      gin.H{},
		})
		return
	}

	// Check if it's already present
	var existingRestaurant model.Restaurant
	if err := database.DB.First(&existingRestaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":    false,
			"message":   "restaurant doesn't exist",
			"error_code": http.StatusNotFound,
			"data":      gin.H{},
		})
		return
	}

	// Delete it
	if err := database.DB.Delete(&existingRestaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    false,
			"message":   "failed to delete the restaurant",
			"error_code": http.StatusInternalServerError,
			"data":      gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    true,
		"message":   "successfully deleted the restaurant",
		"data":      gin.H{},
	})
}

func BlockRestaurant(c *gin.Context) {
	// Get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"message":   "invalid restaurant ID",
			"error_code": http.StatusBadRequest,
			"data":      gin.H{},
		})
		return
	}

	// Check restaurant by id
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":    false,
			"message":   "restaurant not found",
			"error_code": http.StatusNotFound,
			"data":      gin.H{},
		})
		return
	}

	if restaurant.Blocked {
		c.JSON(http.StatusConflict, gin.H{
			"status":    false,
			"message":   "restaurant is already blocked",
			"error_code": http.StatusConflict,
			"data":      gin.H{"restaurant": restaurant},
		})
		return
	}

	// Set blocked as true
	restaurant.Blocked = true

	if err := database.DB.Save(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    false,
			"message":   "failed to change the block status",
			"error_code": http.StatusInternalServerError,
			"data":      gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    true,
		"message":   "restaurant is blocked",
		"data":      gin.H{"restaurant": restaurant},
	})
}

func UnblockRestaurant(c *gin.Context) {
	// Get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"message":   "invalid restaurant ID",
			"error_code": http.StatusBadRequest,
			"data":      gin.H{},
		})
		return
	}

	// Check restaurant by id
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":    false,
			"message":   "restaurant not found",
			"error_code": http.StatusNotFound,
			"data":      gin.H{},
		})
		return
	}

	if !restaurant.Blocked {
		c.JSON(http.StatusConflict, gin.H{
			"status":    false,
			"message":   "restaurant is already unblocked",
			"error_code": http.StatusConflict,
			"data":      gin.H{"restaurant": restaurant},
		})
		return
	}

	// Set blocked as false
	restaurant.Blocked = false

	if err := database.DB.Save(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    false,
			"message":   "failed to change the block status",
			"error_code": http.StatusInternalServerError,
			"data":      gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    true,
		"message":   "restaurant is unblocked",
		"data":      gin.H{"restaurant": restaurant},
	})
}
