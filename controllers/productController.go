package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// public
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
		"status":  true,
		"message": "successfully retrieved products",
		"data": gin.H{
			"products": Products,
		},
	})
}

// public
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

// restuarant id
func AddProduct(c *gin.Context) {

	//check restaurant api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	RestaurantID, ok := RestIDfromEmail(email)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to get restaurant information",
			"error_code": http.StatusBadRequest,
		})
		return
	}
	// Bind JSON
	var Request model.AddProductRequest

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := validate(Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Check if the restaurant ID is correct and present in the database
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, Request.RestaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "restaurant not found",
			"error_code": http.StatusNotFound,
		})
		return
	}

	// Check if the category is present
	var category model.Category
	if err := database.DB.First(&category, Request.CategoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "category doesn't exist",
			"error_code": http.StatusNotFound,
		})
		return
	}

	// Check if the product name already exists within the same restaurant
	var existingProduct model.Product
	if err := database.DB.Where("name =? AND restaurant_id =? AND deleted_at IS NULL", Request.Name, Request.RestaurantID).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "product with the same name already exists in this restaurant",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Proceed with adding the product if all checks pass
	Request.RestaurantID = RestaurantID
	if err := database.DB.Create(&Request).Error; err != nil {
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
		"data": gin.H{
			"product": Request,
		},
	})
}

// restaurant id
func EditProduct(c *gin.Context) {
	//check restaurant api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.RestaurantRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	JWTRestaurantID, ok := RestIDfromEmail(email)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to get restaurant information",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Bind JSON
	var Request model.EditProductRequest
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := validate(Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//check jwt rest id and product rest id
	Request.RestaurantID = RestaurantIDByProductID(Request.ProductID)
	if JWTRestaurantID != Request.RestaurantID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	// Check if the product exists by id
	var existingProduct model.Product
	if err := database.DB.Where("id = ?", Request.ProductID).First(&existingProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to fetch product from the database",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//check if the restaurant id exists
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, Request.RestaurantID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "restaurant doesn't exist",
			"error_code": http.StatusBadRequest,
		})
		return
	}
	// Check if the category is present
	var category model.Category
	if err := database.DB.First(&category, Request.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "category doesn't exist",
			"error_code": http.StatusBadRequest,
		})
		return
	}
	// Update product details
	if err := database.DB.Updates(&Request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update product",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully updated product information",
		"data": gin.H{
			"product": Request,
		},
	})
}

// restaurant id
func DeleteProduct(c *gin.Context) {
	//check restaurant api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.RestaurantRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	JWTRestaurantID, ok := RestIDfromEmail(email)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to get restaurant information",
			"error_code": http.StatusBadRequest,
		})
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
	ProductRestaurantID := RestaurantIDByProductID(uint(productID))

	if JWTRestaurantID != ProductRestaurantID{
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
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
		"status":  true,
		"message": "successfully deleted the product",
	})
}

// user id
func GetUsersFavouriteProduct(c *gin.Context) {
	//check user api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)

	//checking the user sending is performing in his/her account..
	email, ok := EmailFromUserID(UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to get user email from the database",
			"error_code": http.StatusNotFound,
		})
		return
	}
	fmt.Println(email)
	if err := VerifyJWT(c, model.RestaurantRole, email); err != nil {
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

// user id
func AddFavouriteProduct(c *gin.Context) {

	//check user api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)

	var request struct {
		ProductID uint `validate:"required,number" json:"product_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the JSON",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Extracted validation logic for clarity
	if err := validate(request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err,
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	var existingFavouriteProduct model.FavouriteProduct
	var userinfo model.User
	var productinfo model.Product

	// Check if the user exists
	if err := database.DB.Where("id =?", UserID).First(&userinfo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "user not found",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Check if the product exists
	if err := database.DB.Where("id =?", request.ProductID).First(&productinfo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "product not found",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Check if the favorite product combination already exists
	if err := database.DB.Where("user_id =? AND product_id =?", UserID, request.ProductID).First(&existingFavouriteProduct).Error; err == nil {
		// If there's no error, it means the favorite product already exists
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "favorite product already exists",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// If everything checks out, add the favorite product
	if err := database.DB.Create(&model.FavouriteProduct{UserID: UserID, ProductID: request.ProductID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to add favorite product",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "favorite product added successfully",
		"data":    gin.H{},
	})
}

// user id
func RemoveFavouriteProduct(c *gin.Context) {
	//check user api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)

	var request struct {
		ProductID uint `validate:"required,number" json:"product_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the JSON",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Extracted validation logic for clarity
	if err := validate(request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	var existingFavouriteProduct model.FavouriteProduct

	if err := database.DB.Where(&model.FavouriteProduct{UserID: UserID, ProductID: request.ProductID}).First(&existingFavouriteProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "favorite product doesn't exist",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}
	if err := database.DB.Where("user_id =? AND product_id =?", UserID, request.ProductID).Delete(&model.FavouriteProduct{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to delete favorite product",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "favorite product deleted successfully",
		"data":    gin.H{},
	})
}

func RestaurantIDByProductID(ProductID uint) uint {

	var Product model.Product
	if err := database.DB.Where("id = ?", ProductID).First(&Product).Error; err != nil {
		return 0
	}
	return Product.RestaurantID
}
