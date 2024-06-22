package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/helper"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddToCart(c *gin.Context) {
	//check user api authentication
	email, role, err := helper.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)
	//bind the json
	var Request model.AddToCartReq
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "Failed to fetch incoming request. Please provide valid JSON data.",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := validate(&Request); err != nil {
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
	var User model.User
	if err := database.DB.Where("id = ?", UserID).First(&User).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to fetch user information, make sure the user exists",
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

func GetCartTotal(c *gin.Context) {
	//check user api authentication
	email, role, err := helper.GetJWTClaim(c)
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
	//total price of the cart
	sum := 0
	var ProductOffer float64
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

		ProductOffer += Product.OfferAmount * float64(item.Quantity)
		sum += int(Product.Price) * int(item.Quantity)
	}

	FinalAmount := sum - int(ProductOffer)
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"cartitems":    CartItems,
			"productoffer": ProductOffer,
			"totalamount":  sum,
			"finalamount":  FinalAmount,
		},
		"message": "Cart items retrieved successfully",
	})
}

func ClearCart(c *gin.Context) {

	//check user api authentication
	email, role, err := helper.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)

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
}

func RemoveItemFromCart(c *gin.Context) {
	//check user api authentication
	email, role, err := helper.GetJWTClaim(c)
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
	if err := validate(&CartItems); err != nil {
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
	email, role, err := helper.GetJWTClaim(c)
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
	if err := validate(&CartItems); err != nil {
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

func CalculateCartTotal(userID uint) (TotalAmount float64, ProductOffer float64, err error) {
	var cartItems []model.CartItems

	if err := database.DB.Where("user_id = ?", userID).Find(&cartItems).Error; err != nil {
		return 0, 0, errors.New("failed to fetch cart items")
	}

	if len(cartItems) == 0 {
		return 0, 0, errors.New("your cart is empty")
	}

	for _, item := range cartItems {
		var product model.Product
		if err := database.DB.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
			return 0,0, errors.New("failed to fetch product information")
		}
		TotalAmount += (product.Price) * float64(item.Quantity)
		ProductOffer += product.OfferAmount * float64(item.Quantity)
	}

	return TotalAmount, ProductOffer, nil
}
