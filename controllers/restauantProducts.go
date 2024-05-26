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
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid restaurant ID",
			"ok":    false,
		})
		return
	}

	var products []model.Product
	if err := database.DB.Where("restaurant_id =?", restaurantID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve products",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"ok":       true,
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

func GetFavouriteProductByUserID(c *gin.Context) {

	userIDStr := c.Param("userid")
	UserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userid format",
			"ok":    true,
		})
		return
	}

	//checking the user sending is performing in his/her account..
	email, ok := EmailFromUserID(uint(UserID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get user email from the database",
			"ok":    false,
		})
		return
	}
	if ok := VerifyJWT(c, email); !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "unauthorized user",
			"ok":    false,
		})
		return
	}

	var FavouriteProducts []model.FavouriteProduct

	if err := database.DB.Where("user_id =?", UserID).Find(&FavouriteProducts).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "the userid doesn't exist on the database",
			"ok":    true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favouritelist": FavouriteProducts,
		"ok":            true,
	})

}

func AddFavouriteProduct(c *gin.Context) {
	var request struct {
		UserID    uint `validate:"required,number" json:"user_id"`
		ProductID uint `validate:"required,number" json:"product_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind the json",
			"ok":    false,
		})
		return
	}


	//check if the user is not impersonating other users through jwt email and users email match..
	email, ok := EmailFromUserID(request.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get user email from the database",
			"ok":    false,
		})
		return
	}
	if ok := VerifyJWT(c, email);!ok{
		c.JSON(http.StatusNotFound, gin.H{
			"error": "unauthorized user",
			"ok":    false,
		})
		return
	}

	// Extracted validation logic for clarity
	if ok := validate(request, c); !ok {
		return
	}

	var existingFavouriteProduct model.FavouriteProduct
	var userinfo model.User
	var productinfo model.Product

	// Check if the user exists
	if err := database.DB.Where("id =?", request.UserID).First(&userinfo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User not found.",
			"ok":    false,
		})
		return
	}

	// Check if the product exists
	if err := database.DB.Where("id =?", request.ProductID).First(&productinfo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Product not found.",
			"ok":    false,
		})
		return
	}

	// Check if the favorite product combination already exists
	if err := database.DB.Where("user_id =? AND product_id =?", request.UserID, request.ProductID).First(&existingFavouriteProduct).Error; err == nil {
		// If there's no error, it means the favorite product already exists
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Favorite product already exists.",
			"ok":    false,
		})
		return
	}

	// If everything checks out, add the favorite product
	if err := database.DB.Create(&model.FavouriteProduct{UserID: request.UserID, ProductID: request.ProductID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add favorite product.",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Favorite product added successfully.",
		"ok":      true,
	})
}

func RemoveFavouriteProduct(c *gin.Context) {
	var request struct {
		UserID    uint `validate:"required,number" json:"user_id"`
		ProductID uint `validate:"required,number" json:"product_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind the json",
			"ok":    false,
		})
		return
	}

	//check if the user is not impersonating other users through jwt email and users email match..
	email, ok := EmailFromUserID(request.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get user email from the database",
			"ok":    false,
		})
		return
	}
	if ok := VerifyJWT(c, email);!ok{
		c.JSON(http.StatusNotFound, gin.H{
			"error": "unauthorized user",
			"ok":    false,
		})
		return
	}

	// Extracted validation logic for clarity
	if ok := validate(request, c); !ok {
		return
	}

	var existingFavouriteProduct model.FavouriteProduct

	if err := database.DB.Where(&model.FavouriteProduct{UserID: request.UserID, ProductID: request.ProductID}).First(&existingFavouriteProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Favorite product dont exists.",
			"ok":    false,
		})
		return
	}
	if err := database.DB.Where("user_id =? AND product_id =?", request.UserID, request.ProductID).Delete(&model.FavouriteProduct{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete favorite product.",
			"ok":    false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Favorite product deleted successfully.",
		"ok":      true,
	})
}
