package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProductList(c *gin.Context) {
	var Products []model.Product

	tx := database.DB.Select("*").Find(&Products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to retrieve data from the database, or the product doesn't exists",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"productlist": Products,
		"ok":          true,
	})
}

func AddProduct(c *gin.Context) {
    // Bind JSON
    var product model.Product
    
    if err := c.BindJSON(&product); err!= nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "invalid request body",
			"ok":false,
        })
        return
    }

    // Check if the restaurant ID is correct and present in the database
    var restaurant model.Restaurant
    if err := database.DB.First(&restaurant, product.RestaurantID).Error; err!= nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "restaurant not found",
			"ok":false,
        })
        return
    }

    // Check if the product name already exists within the same restaurant
    var existingProduct model.Product
    if err := database.DB.Where("name =? AND restaurant_id =?", product.Name, product.RestaurantID).First(&existingProduct).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "a product with the same name already exists in this restaurant",
			"ok":false,
        })
        return
    }

    // Proceed with adding the product if all checks pass
    if err := database.DB.Create(&product).Error; err!= nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "failed to create product",
			"ok":false,
        })
        return
    }

    // Return a success response
    c.JSON(http.StatusCreated, gin.H{
        "message": "successfully added new product",
        "data":   product,
    })
}
 
 func EditProduct(c *gin.Context) {
	// Bind JSON
	// Check if the product exists by id
	// If product exists, validate new data
	// Check if the restaurant id is present
	// Check if the category is present
	// Update product details
 }
 
 func DeleteProduct(c *gin.Context) {
	// Get product id from parameters
	// Check if the product exists by id
	// If product exists
	// Delete the product
 }
 