package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"strconv"

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

func GetProductsByRestaurantID(c *gin.Context) {
    restaurantIDStr := c.Param("restaurantid")
    restaurantID, err := strconv.Atoi(restaurantIDStr)
    if err!= nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid restaurant ID",
            "ok":    false,
        })
        return
    }

    var products []model.Product
    if err := database.DB.Where("restaurant_id =?", restaurantID).Find(&products).Error; err!= nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to retrieve products",
            "ok":    false,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "products": products,
        "ok":      true,
    })
}

func AddProduct(c *gin.Context) {
	// Bind JSON
	var product model.Product

	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"ok":    false,
		})
		return
	}
	product.ID = 0

	// Check if the restaurant ID is correct and present in the database
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, product.RestaurantID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "restaurant not found",
			"ok":    false,
		})
		return
	}

	// Check if the category is present
	var category model.Category
	if err := database.DB.First(&category, product.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "category doesnt exists ",
			"ok":    false,
		})
		return
	}

	// Check if the product name already exists within the same restaurant
	var existingProduct model.Product
	if err := database.DB.Where("name =? AND restaurant_id =? AND deleted_at IS NULL", product.Name, product.RestaurantID).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "a product with the same name already exists in this restaurant",
			"ok":    false,
		})
		return
	}

	

	// Proceed with adding the product if all checks pass
	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create product",
			"ok":    false,
		})
		return
	}

	// Return a success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "successfully added new product",
		"data":    product,
	})
}

func EditProduct(c *gin.Context) {
	// Bind JSON
	var product model.Product
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
			"ok":    false,
		})
	}
	// Check if the product exists by id
	var existingProduct model.Restaurant
	if err := database.DB.First(&existingProduct, product.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch product details from the database",
			"ok":    false,
		})
		return
	}

	//check if the restaurant id exists
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, product.RestaurantID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "restaurant doesnt exists ",
			"ok":    false,
		})
		return
	}
	// Check if the category is present
	var category model.Category
	if err := database.DB.First(&category, product.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "category doesnt exists ",
			"ok":    false,
		})
	}
	// Update product details
	if err := database.DB.Updates(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to update products ",
			"ok":    false,
		})
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"message": "successfully changed product info",
		"ok":      true,
	})
}

func DeleteProduct(c *gin.Context) {
	// Get product id from parameters
	productIDStr := c.Param("productid")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
			"ok":    false,
		})
		return
	}
	// Check if the product exists by id
    var product model.Product
    if err := database.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "product is not present on the database",
			"ok":    false,
		})
		return
	}

    //delete the product
	if err := database.DB.Delete(&product, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to delete the product from the database",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "successfully  deleted the product",
		"ok":      true,
	})

}
