package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//add to cart with restaurantid, find the rest id from the product
//add endpoint listcartwithrestaurants show restaurnat id and name,

func AddToCart(c *gin.Context) {
	// Check user API authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)
	// Bind the JSON
	var Request model.AddToCartReq
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "Failed to fetch incoming request. Please provide valid JSON data.",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := utils.Validate(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var Product model.Product
	if err := database.DB.Where("id = ?", Request.ProductID).First(&Product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "Failed to fetch product information. Please ensure the specified product exists.",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if Request.Quantity > Product.StockLeft {
		message := fmt.Sprintf("Requested quantity exceeds available stock. Available stock: %v", Product.StockLeft)
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    message,
			"error_code": http.StatusConflict,
		})
		return
	}
	if Request.Quantity > model.MaxUserQuantity {
		message := fmt.Sprintf("Requested quantity exceeds allowed limit. Maximum quantity per cart:  %v", model.MaxUserQuantity)
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    message,
			"error_code": http.StatusConflict,
		})
		return
	}

	var CartItem model.CartItems
	if err := database.DB.Where("user_id = ? AND product_id = ?", UserID, Request.ProductID).First(&CartItem).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Failed to fetch items of the user. Please provide a valid user ID.",
				"error_code": http.StatusInternalServerError,
			})
			return
		}

		var AddCartItems model.CartItems

		AddCartItems.UserID = UserID
		AddCartItems.ProductID = Request.ProductID
		AddCartItems.RestaurantID = Product.RestaurantID // Set the RestaurantID
		AddCartItems.Quantity = Request.Quantity
		AddCartItems.CookingRequest = Request.CookingRequest

		if err := database.DB.Create(&AddCartItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Failed to update cart items. Please try again later.",
				"error_code": http.StatusInternalServerError,
			})
			return
		}
	} else {
		if CartItem.Quantity+Request.Quantity > model.MaxUserQuantity {
			c.JSON(http.StatusConflict, gin.H{
				"status":     false,
				"message":    "Total of Requested and Current need of quantity exceeds the max user quantity",
				"error_code": http.StatusConflict,
			})
			return
		}

		CartItem.Quantity += Request.Quantity

		if Request.CookingRequest != "" {
			CartItem.CookingRequest = Request.CookingRequest
		}

		if err := database.DB.Where("user_id = ? AND product_id = ?", UserID, Request.ProductID).Updates(&CartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Failed to update cart items. Please try again later.",
				"error_code": http.StatusInternalServerError,
			})
			return
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Product successfully added to cart",
	})
}

// get cart total by restaurant
// add restaurant_id in the query
func GetCartTotal(c *gin.Context) {
	// Check user API authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)

	var CartItems []model.CartItems

	if err := database.DB.Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to fetch cart items. Please try again later.",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "Your cart is empty.",
			"error_code": http.StatusNotFound,
		})
		return
	}

	// Create a map to store cart totals per restaurant
	cartTotals := make(map[uint]struct {
		Items       []model.CartItems
		ProductOffer float64
		TotalAmount  float64
	})

	// Iterate through cart items and calculate totals per restaurant
	for _, item := range CartItems {
		var Product model.Product
		if err := database.DB.Where("id = ?", item.ProductID).First(&Product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":     false,
				"message":    "Failed to fetch product information. Please try again later.",
				"error_code": http.StatusNotFound,
			})
			return
		}

		restaurantID := Product.RestaurantID
		if _, exists := cartTotals[restaurantID]; !exists {
			cartTotals[restaurantID] = struct {
				Items       []model.CartItems
				ProductOffer float64
				TotalAmount  float64
			}{}
		}

		if val, ok := cartTotals[restaurantID]; ok {
			val.Items = append(val.Items, item)
			val.ProductOffer += Product.OfferAmount * float64(item.Quantity)
			val.TotalAmount += float64(Product.Price) * float64(item.Quantity)
			cartTotals[restaurantID] = val
		} else {
			cartTotals[restaurantID] = struct {
				Items       []model.CartItems
				ProductOffer float64
				TotalAmount  float64
			}{
				Items:       []model.CartItems{item},
				ProductOffer: Product.OfferAmount * float64(item.Quantity),
				TotalAmount:  float64(Product.Price) * float64(item.Quantity),
			}
		}

	}

	// Prepare response data
	responseData := make(map[uint]interface{})
	for restaurantID, totals := range cartTotals {
		FinalAmount := totals.TotalAmount - totals.ProductOffer
		responseData[restaurantID] = gin.H{
			"cartitems":    totals.Items,
			"productoffer": totals.ProductOffer,
			"totalamount":  totals.TotalAmount,
			"finalamount":  FinalAmount,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   responseData,
		"message": "Cart items retrieved successfully",
	})
}


// clear whole cart
// also clear whole cart by restaurant_id
func ClearCart(c *gin.Context) {
	// Check user API authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)

	restaurantID := c.Query("restaurant_id")
	if restaurantID == "" || restaurantID == "0"{
		// Clear entire cart
		var CartItems model.CartItems
		if err := database.DB.Where("user_id = ?", UserID).Delete(&CartItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Failed to delete cart items. Please try again later.",
				"error_code": http.StatusInternalServerError,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "Deleted entire cart of the User",
		})
	} else {
		// Clear cart for specific restaurant
		var CartItems model.CartItems
		if err := database.DB.Where("user_id = ? AND restaurant_id = ?", UserID, restaurantID).Delete(&CartItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "Failed to delete cart items for the specified restaurant. Please try again later.",
				"error_code": http.StatusInternalServerError,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "Deleted cart items for the specified restaurant",
		})
	}
}

func RemoveItemFromCart(c *gin.Context) {
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
	//bindthe json
	var CartItems model.RemoveItem
	if err := c.BindJSON(&CartItems); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the json",
			"error_code": http.StatusBadRequest,
		})
		return
	}
	//validate
	if err := utils.Validate(&CartItems); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var CartItem model.CartItems
	if err := database.DB.Where("user_id = ? AND product_id = ?", UserID, CartItems.ProductID).First(&CartItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusNotFound,
		})
		return
	}
	//if yes, remove the item
	if err := database.DB.Where("user_id = ? AND product_id = ?", UserID, CartItems.ProductID).Delete(&CartItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Removed the Item Successfully",
	})
}

func UpdateQuantity(c *gin.Context) {
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
	//bindthe json
	var CartItems model.CartItems
	if err := c.BindJSON(&CartItems); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the json",
			"error_code": http.StatusBadRequest,
		})
		return
	}
	CartItems.UserID = UserID
	//validate
	if err := utils.Validate(&CartItems); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var CartItem model.CartItems
	if err := database.DB.Where("user_id = ? AND product_id = ?", CartItems.UserID, CartItems.ProductID).First(&CartItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusNotFound,
		})
		return
	}

	if CartItems.Quantity > model.MaxUserQuantity {
		message := fmt.Sprintf("Requested quantity exceeds allowed limit. Maximum quantity per cart:  %v", model.MaxUserQuantity)
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    message,
			"error_code": http.StatusConflict,
		})
		return
	}

	//update quantity
	if err := database.DB.Where("user_id = ? AND product_id = ?", CartItems.UserID, CartItems.ProductID).Updates(&CartItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Updated the Quantity Successfully",
	})
}

// calculate cart total by restaurant_id
func CalculateCartTotal(userID uint,RestaurantID uint) (TotalAmount float64, ProductOffer float64, err error) {
	var cartItems []model.CartItems

	if err := database.DB.Where("user_id = ? AND restaurant_id = ?", userID,RestaurantID).Find(&cartItems).Error; err != nil {
		return 0, 0, errors.New("failed to fetch cart items")
	}

	if len(cartItems) == 0 {
		return 0, 0, errors.New("your cart is empty")
	}

	for _, item := range cartItems {
		var product model.Product
		if err := database.DB.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
			return 0, 0, errors.New("failed to fetch product information")
		}
		TotalAmount += (product.Price) * float64(item.Quantity)
		ProductOffer += product.OfferAmount * float64(item.Quantity)
	}

	return TotalAmount, ProductOffer, nil
}

func AddCookingRequest(c *gin.Context) {
	var Request model.AddCookingRequest
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "provide product_id, cooking_request in the payload"})
		return
	}

	//check whether the cooking_request contains atleast the minimum words
	words := strings.Fields(Request.CookingRequest)
	wordCount := len(words)
	if wordCount < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "cooking_request must contain atleast 2 words"})
		return
	}

	//update the cart with cooking_request
	if err := database.DB.Where("product_id = ?", Request.ProductID).Updates(model.CartItems{CookingRequest: Request.CookingRequest}).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "cart is empty or the specified product is not in this cart,please make sure the product exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "successfully updated cooking request"})

}
