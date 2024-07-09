package controllers

import (
	"foodbuddy/internal/database"
	"foodbuddy/internal/utils"
	"foodbuddy/internal/model"
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
		})
		return
	}

	var products []model.Product
	if err := database.DB.Where("restaurant_id =?", restaurantID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to retrieve products",
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
		})
		return
	}
	// Bind JSON
	var Request model.AddProductRequest

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process request",
		})
		return
	}

	if err := utils.Validate(Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
		})
		return
	}

	if Request.Veg != model.YES && Request.Veg != model.NO {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "please specify if the product is vegetarian by 'YES' or 'NO' ",
		})
		return
	}

	if Request.Price < Request.OfferAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "offer amount should not be more than the product price",
		})
		return
	}

	// Check if the restaurant ID is correct and present in the database
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, JWTRestaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "restaurant not found",
		})
		return
	}

	// Check if the category is present
	var category model.Category
	if err := database.DB.First(&category, Request.CategoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "category doesn't exist",
		})
		return
	}

	// Check if the product name already exists within the same restaurant
	var existingProduct model.Product
	if err := database.DB.Where("name =? AND restaurant_id =? AND deleted_at IS NULL", Request.Name, JWTRestaurantID).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "product with the same name already exists in this restaurant",
		})
		return
	}
	existingProduct = model.Product{
		RestaurantID: JWTRestaurantID,
		CategoryID:   Request.CategoryID,
		Name:         Request.Name,
		Description:  Request.Description,
		ImageURL:     Request.ImageURL,
		Price:        Request.Price,
		MaxStock:     Request.MaxStock,
		StockLeft:    Request.StockLeft,
	}

	if err := database.DB.Create(&existingProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create product",
		})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "successfully added new product",
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
		})
		return
	}

	// Bind JSON
	var Request model.EditProductRequest
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process request",
		})
		return
	}

	if err := utils.Validate(Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
		})
		return
	}

	//check jwt rest id and product rest id
	pRestID := RestaurantIDByProductID(Request.ProductID)
	if JWTRestaurantID != pRestID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request, product is not yours",
		})
		return
	}

	// Check if the product exists by id
	var existingProduct model.Product
	if err := database.DB.Where("id = ?", Request.ProductID).First(&existingProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to fetch product from the database",
		})
		return
	}

	if Request.Veg != model.YES && Request.Veg != model.NO {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "please specify if the product is vegetarian by 'YES' or 'NO' ",
		})
		return
	}

	if Request.Price < Request.OfferAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "offer amount should not be more than the product price",
		})
		return
	}

	existingProduct.Name = Request.Name
	existingProduct.Description = Request.Description
	existingProduct.ImageURL = Request.ImageURL
	existingProduct.Price = Request.Price
	existingProduct.MaxStock = Request.MaxStock
	existingProduct.StockLeft = Request.StockLeft
	existingProduct.Veg = Request.Veg

	// Update product details
	if err := database.DB.Where("id = ?", Request.ProductID).Updates(&existingProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update product",
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
		})
		return
	}
	ProductRestaurantID := RestaurantIDByProductID(uint(productID))

	if JWTRestaurantID != ProductRestaurantID {
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
		})
		return
	}

	//delete the product
	if err := database.DB.Delete(&product, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "unable to delete the product from the database",
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

	// //checking the user sending is performing in his/her account..
	// _, ok := EmailFromUserID(UserID)
	// if !ok {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"status":     false,
	// 		"message":    "failed to get user email from the database",
	// 		"error_code": http.StatusNotFound,
	// 	})
	// 	return
	// }

	var FavouriteProducts []model.FavouriteProduct

	if err := database.DB.Where("user_id =?", UserID).Find(&FavouriteProducts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "the user ID doesn't exist in the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        true,
		"message":       "successfully retrieved favourite products",
		"favourite_list": FavouriteProducts,
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
			"data":       gin.H{},
		})
		return
	}

	// Extracted validation logic for clarity
	if err := utils.Validate(request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err,
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
			"data":       gin.H{},
		})
		return
	}

	// Check if the product exists
	if err := database.DB.Where("id =?", request.ProductID).First(&productinfo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "product not found",
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
			"data":       gin.H{},
		})
		return
	}

	// If everything checks out, add the favorite product
	if err := database.DB.Create(&model.FavouriteProduct{UserID: UserID, ProductID: request.ProductID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to add favorite product",
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
			"data":       gin.H{},
		})
		return
	}

	// Extracted validation logic for clarity
	if err := utils.Validate(request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"data":       gin.H{},
		})
		return
	}

	var existingFavouriteProduct model.FavouriteProduct

	if err := database.DB.Where(&model.FavouriteProduct{UserID: UserID, ProductID: request.ProductID}).First(&existingFavouriteProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "favorite product doesn't exist",
			"data":       gin.H{},
		})
		return
	}
	if err := database.DB.Where("user_id =? AND product_id =?", UserID, request.ProductID).Delete(&model.FavouriteProduct{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to delete favorite product",
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

func OnlyVegProducts(c *gin.Context) {
	var products []model.Product

	tx := database.DB.Where("veg = ?", model.YES).
		Order("price ASC").
		Find(&products)

	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get product information",
		})
		return
	}

	var response []model.ProductResponse
	for _, product := range products {
		var dbCategory model.Category
		var dbRestaurant model.Restaurant

		if err := database.DB.Where("id = ?", product.CategoryID).First(&dbCategory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get category information"})
			return
		}

		if err := database.DB.Where("id = ?", product.RestaurantID).First(&dbRestaurant).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get restaurant information"})
			return
		}

		response = append(response, model.ProductResponse{
			ID:             product.ID,
			RestaurantName: dbRestaurant.Name,
			CategoryName:   dbCategory.Name,
			Name:           product.Name,
			Description:    product.Description,
			ImageURL:       product.ImageURL,
			Price:          product.Price,
			StockLeft:      product.StockLeft,
			AverageRating:  product.AverageRating,
			Veg:            product.Veg,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "data": response})
}

func AddProductOffer(c *gin.Context) {
	var request model.AddOfferRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid request body: ",
		})
		return
	}

	email, role, err := utils.GetJWTClaim(c)
	if role != model.RestaurantRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	pRestID := RestaurantIDByProductID(request.ProductID)
	RestID, _ := RestIDfromEmail(email)
	if pRestID != RestID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "StatusUnauthorized",
		})
		return
	}

	exist, Product := CheckProduct(int(request.ProductID))
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to find the product",
		})
		return
	}

	Product.OfferAmount = request.OfferAmount
	if err := database.DB.Model(&Product).Where("id = ?", request.ProductID).Update("offer_amount", request.OfferAmount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to add the offer amount",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully added the offer amount",
		"data":    Product,
	})
}

func RemoveProductOffer(c *gin.Context) {

	ProductID, err := strconv.Atoi(c.Param("productid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ProductID",
		})
		return
	}

	//check restaurant api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.RestaurantRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	pRestID := RestaurantIDByProductID(uint(ProductID))
	RestID, _ := RestIDfromEmail(email)
	if pRestID != RestID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "StatusUnauthorized",
		})
		return
	}

	exist, Product := CheckProduct(ProductID)
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to find the product",
		})
		return
	}

	Product.OfferAmount = 0
	if err := database.DB.Model(&Product).Where("id = ?", ProductID).Update("offer_amount", 0).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to remove the offer amount",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully removed the offer amount",
		"data":    Product,
	})

}

func CheckProduct(ProductID int) (bool, model.Product) {
	var Product model.Product

	if err := database.DB.Where("id = ?", ProductID).First(&Product).Error; err != nil {
		return false, model.Product{}
	}

	return true, Product

}

func GetProductOffers(c *gin.Context) {
	//get products with more than 0 in offer_amount
	var Products []model.Product

	if err := database.DB.Where("offer_amount > ?", 0).Find(&Products).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": false, "message": "failed to get product details",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true, "data": Products,
	})
}
