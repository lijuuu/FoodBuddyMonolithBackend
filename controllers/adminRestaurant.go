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
	//search db and get all

	if err := database.DB.Select("*").Find(&restaurants).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to retrieve data from the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"restaurantslist": restaurants,
		"ok":              true,
	})
}

func AddRestaurant(c *gin.Context) {
	//bind json
	var restaurant model.Restaurant
	if err := c.BindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
			"ok":    false,
		})
	}

	var existingRestaurant model.Restaurant
	if err := database.DB.Where("name =?", restaurant.Name).Find(&existingRestaurant).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error while checking restaurant name",
			"ok":    false,
		})
		return
	}

	if restaurant.Name == existingRestaurant.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "restaurant already exists",
			"ok":    false,
		})
		return
	}

   //set block as false
  restaurant.Blocked = false

	//add to db
	if err := database.DB.Create(&restaurant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to add restaurant to the database",
			"ok":    false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully added new restaurants",
		"ok":      false,
	})
}

func EditRestaurant(c *gin.Context) {
	//check existing restuarant
	var restaurant model.Restaurant
	if err := c.BindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind request",
			"ok":    false,
		})
		return
	}
	//if present update it with the new data
	var existingRestaurant model.Restaurant
	if err := database.DB.First(&existingRestaurant, restaurant.ID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "restaurant doesnt exist",
			"ok":    false,
		})
		return
	}
	//edit the restaurant
	if err := database.DB.Updates(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to edit the restaurant",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully edited the restaurant",
		"ok":      true,
	})

}

func DeleteRestaurant(c *gin.Context) {
	///get the restaurant id

	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid restaurant ID",
			"ok":    false,
		})
		return
	}

	//check if its already present
	var existingRestaurant model.Restaurant
	if err := database.DB.First(&existingRestaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "restaurant doesnt exist",
			"ok":    false,
		})
		return
	}
	//delete it
	if err := database.DB.Delete(&existingRestaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete the restaurant",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully deleted the restaurant",
		"ok":      true,
	})
}

func BlockRestaurant(c *gin.Context) {
	///get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid restaurant ID",
			"ok":    false,
		})
		return
	}

	//check resturant by id
	var restaurant model.Restaurant
   if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "restaurant not found",
			"ok":    false,
		})
		return
	}

	if restaurant.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"message": "restaurant is already blocked",
			"user":    restaurant,
			"ok":      true,
		})
		return
	}
   //set Blocked as true
	restaurant.Blocked = true

	tx := database.DB.Updates(&restaurant)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to change the block status ",
			"ok":    false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "restaurant is blocked", "restaurant": restaurant, "ok": true,
	})
	
}

func UnblockRestaurant(c *gin.Context) {
	///get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid restaurant ID",
			"ok":    false,
		})
		return
	}

	//check resturant by id
	var restaurant model.Restaurant
   if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "restaurant not found",
			"ok":    false,
		})
		return
	}

	if !restaurant.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"message": "restaurant is already unblocked",
			"user":    restaurant,
			"ok":      true,
		})
		return
	}
   //set Blocked as true
	restaurant.Blocked = false

	if err := database.DB.Model(&restaurant).UpdateColumn("blocked", false);err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to change the block status ",
			"ok":    false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "restaurant is blocked", 
      "restaurant": restaurant, 
      "ok": true,
	})
}
