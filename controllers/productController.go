package controllers

import (
	"fmt"
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
			"status":     false,
			"message":    "failed to retrieve data from the database, or the product doesn't exist",
			"error_code": http.StatusNotFound,
			
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "successfully retrieved products",
		"data":       gin.H{
			"products": Products,
		},
	})
}

func GetProductsByRestaurantID(c *gin.Context) {
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "invalid restaurant ID",
			"error_code": http.StatusNotFound,
			
		})
		return
	}

	var products []model.Product
	if err := database.DB.Where("restaurant_id =?", restaurantID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to retrieve products",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved products",
		"data": gin.H{
			"products": products,
		},
	})
}

func AddProduct(c *gin.Context) {
	   
	// if status := c.IsAborted(); !status{
	// 	return
	// }

	// Bind JSON
	var product model.Product

	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process request",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	if err := validate(product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	product.ID = 0

	// Check if the restaurant ID is correct and present in the database
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, product.RestaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "restaurant not found",
			"error_code": http.StatusNotFound,
			
		})
		return
	}

	// if _,err := CheckRestaurant(c,restaurant.Email);err!=nil{
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"status":     false,
	// 		"message":    err.Error(),
	// 		"error_code": http.StatusUnauthorized,
	// 	})
	// 	return
	//    }

	// Check if the category is present
	var category model.Category
	if err := database.DB.First(&category, product.CategoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "category doesn't exist",
			"error_code": http.StatusNotFound,
			
		})
		return
	}

	// Check if the product name already exists within the same restaurant
	var existingProduct model.Product
	if err := database.DB.Where("name =? AND restaurant_id =? AND deleted_at IS NULL", product.Name, product.RestaurantID).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "product with the same name already exists in this restaurant",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	// Proceed with adding the product if all checks pass
	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create product",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "successfully added new product",
		"error_code": http.StatusOK,
		"data":       gin.H{
			"product": product,
		},
	})
}

func EditProduct(c *gin.Context) {
	if status := c.IsAborted(); !status{
		return
	}
	// Bind JSON
	var product model.Product
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process request",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	if err := validate(product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
			
		})
		return
	}
	// Check if the product exists by id
	var existingProduct model.Product
	if err := database.DB.Where("id = ?", product.ID).First(&existingProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to fetch product from the database",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	//check if the restaurant id exists
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, product.RestaurantID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "restaurant doesn't exist",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}
	// Check if the category is present
	var category model.Category
	if err := database.DB.First(&category, product.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "category doesn't exist",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}
	// Update product details
	if err := database.DB.Updates(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update product",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "successfully updated product information",
		"data":       gin.H{
			"product": product,
		},
	})
}

func DeleteProduct(c *gin.Context) {
	if status := c.IsAborted(); !status{
		return
	}
	// Get product id from parameters
	productIDStr := c.Param("productid")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "invalid product ID",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}
	// Check if the product exists by id
	var product model.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "product is not present in the database",
			"error_code": http.StatusNotFound,
			
		})
		return
	}

	//delete the product
	if err := database.DB.Delete(&product, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "unable to delete the product from the database",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "successfully deleted the product",
		
	})
}

func GetFavouriteProductByUserID(c *gin.Context) {
	userIDStr := c.Param("userid")
	UserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "invalid user ID format",
			"error_code": http.StatusNotFound,
			
		})
		return
	}

	//checking the user sending is performing in his/her account..
	email, ok := EmailFromUserID(uint(UserID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to get user email from the database",
			"error_code": http.StatusNotFound,
			
		})
		return
	}
	fmt.Println(email)
	if err := VerifyJWT(c, model.RestaurantRole,email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
			
		})
		return
	}

	var FavouriteProducts []model.FavouriteProduct

	if err := database.DB.Where("user_id =?", UserID).Find(&FavouriteProducts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "the user ID doesn't exist in the database",
			"error_code": http.StatusNotFound,
			
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        true,
		"message":       "successfully retrieved favourite products",
		"favouritelist": FavouriteProducts,
		"data":          gin.H{},
	})
}

func AddFavouriteProduct(c *gin.Context) {
	var request struct {
		UserID    uint `validate:"required,number" json:"user_id"`
		ProductID uint `validate:"required,number" json:"product_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind the JSON",
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}

	//check if the user is not impersonating other users through jwt email and users email match..
	email, ok := EmailFromUserID(request.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user email from the database",
			"error_code":http.StatusNotFound,
			"data":    gin.H{},
		})
		return
	}
	if err := VerifyJWT(c,model.RestaurantRole, email); err!=nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized user",
			"error_code":http.StatusUnauthorized,
			"data":    gin.H{},
		})
		return
	}

	// Extracted validation logic for clarity
	if err := validate(request); err != nil  {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err,
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}

	var existingFavouriteProduct model.FavouriteProduct
	var userinfo model.User
	var productinfo model.Product

	// Check if the user exists
	if err := database.DB.Where("id =?", request.UserID).First(&userinfo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "user not found",
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}

	// Check if the product exists
	if err := database.DB.Where("id =?", request.ProductID).First(&productinfo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "product not found",
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}

	// Check if the favorite product combination already exists
	if err := database.DB.Where("user_id =? AND product_id =?", request.UserID, request.ProductID).First(&existingFavouriteProduct).Error; err == nil {
		// If there's no error, it means the favorite product already exists
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "favorite product already exists",
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}

	// If everything checks out, add the favorite product
	if err := database.DB.Create(&model.FavouriteProduct{UserID: request.UserID, ProductID: request.ProductID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to add favorite product",
			"error_code":http.StatusInternalServerError,
			"data":    gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "favorite product added successfully",
		"data":    gin.H{},
	})
}

func RemoveFavouriteProduct(c *gin.Context) {
	var request struct {
		UserID    uint `validate:"required,number" json:"user_id"`
		ProductID uint `validate:"required,number" json:"product_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind the JSON",
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}

	//check if the user is not impersonating other users through jwt email and users email match..
	email, ok := EmailFromUserID(request.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user email from the database",
			"error_code":http.StatusNotFound,
			"data":    gin.H{},
		})
		return
	}
	if err := VerifyJWT(c, model.RestaurantRole,email); err!= nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized user",
			"error_code":http.StatusUnauthorized,
			"data":    gin.H{},
		})
		return
	}

	// Extracted validation logic for clarity
	if err := validate(request); err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "unauthorized user",
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}

	var existingFavouriteProduct model.FavouriteProduct

	if err := database.DB.Where(&model.FavouriteProduct{UserID: request.UserID, ProductID: request.ProductID}).First(&existingFavouriteProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "favorite product doesn't exist",
			"error_code":http.StatusBadRequest,
			"data":    gin.H{},
		})
		return
	}
	if err := database.DB.Where("user_id =? AND product_id =?", request.UserID, request.ProductID).Delete(&model.FavouriteProduct{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to delete favorite product",
			"error_code":http.StatusInternalServerError,
			"data":    gin.H{},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "favorite product deleted successfully",
		"data":    gin.H{},
	})
}