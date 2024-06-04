package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddToCart(c *gin.Context)  {
	//bind the json
	var Request model.AddToCartReq
	if err:= c.BindJSON(&Request);err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":false,
			"message":"failed to fetch incoming request",
			"error_code":http.StatusBadRequest,
		})
		return
	}
	// 	- **Validation**:
	// 	- Validate the product ID to ensure it exists.
	var Product model.Product
    if err:= database.DB.Where("id = ?",Request.ProductID).First(&Product).Error;err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":false,
			"message":"failed to fetch product information, make sure the product exists",
			"error_code":http.StatusBadRequest,
		})
		return
	}
	// 	- Validate the user ID to ensure the user is authenticated.
	var User model.User
    if err:= database.DB.Where("id = ?",Request.UserID).First(&User).Error;err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"status":false,
			"message":"failed to fetch user information, make sure the user exists",
			"error_code":http.StatusBadRequest,
		})
		return
	}

	// - **Stock Check**:
	// 	- Fetch the current stock level of the product.
	
	// 	- Ensure the requested quantity does not exceed available stock.
	// 	- Ensure the requested quantity does not exceed any per-user purchase limits.


	// - **Update Cart**:
	// 	- If the product is already in the cart, update the quantity.
	// 	- If the product is not in the cart, add it with the specified quantity.


	// - **Update Cart Total**:
	// 	- Recalculate the cart total after adding the product.


	// - **Response**:
	// 	- Provide feedback to the user about the action (success or failure).
}