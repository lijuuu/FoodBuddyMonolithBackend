package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateOrderID(UserID uint) (string, error) {
	var UserInfo model.User
	if err := database.DB.Where("id = ?", UserID).First(&UserInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", fmt.Errorf("failed to fetch user information: %w", err)
	}

	// Replace spaces and special characters with an underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	cleanedName := re.ReplaceAllString(UserInfo.Name, "_")

	var OrderID string
	for {
		random := utils.GenerateRandomString(10)
		OrderID = fmt.Sprintf("%v_%v", cleanedName, random)

		var existingOrder model.Order
		if err := database.DB.Where("order_id = ?", OrderID).First(&existingOrder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return "", fmt.Errorf("error checking existing order ID: %w", err)
		}
	}

	return OrderID, nil
}

func ValidAddress(UserID uint, AddressID uint) bool {
	var Address model.Address
	if err := database.DB.Where("user_id = ? AND address_id = ?", UserID, AddressID).First(&Address).Error; err != nil {
		return false
	}
	return true
}

func CheckStock(UserID uint) bool {

	var CartItems []model.CartItems
	//get cartItems,find the product check with stockleft
	if err := database.DB.Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {
		return false
	}
	for _, v := range CartItems {
		var Product model.Product
		if err := database.DB.Where("id = ?", v.ProductID).First(&Product).Error; err != nil {
			return false
		}
		if v.Quantity > Product.StockLeft {
			return false
		}
	}
	return true
}

func CartToOrderItems(UserID uint, OrderID string) bool {
	var CartItems []model.CartItems
	if err := database.DB.Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {

		return false
	}

	for _, v := range CartItems {

		var Product model.Product
		if err := database.DB.Where("id = ?", v.ProductID).First(&Product).Error; err != nil {
			return false
		}

		OrderItem := model.OrderItem{
			OrderID:        OrderID,
			ProductID:      v.ProductID,
			Quantity:       v.Quantity,
			Amount:         v.Quantity * Product.Price,
			CookingRequest: v.CookingRequest,
			OrderStatus:    model.OrderStatusProcessing,
			RestaurantID:   RestaurantIDByProductID(v.ProductID),
		}

		if err := database.DB.Create(&OrderItem).Error; err != nil {
			return false
		}
	}

	//then remove the cartdetail for that user
	var CartItem model.CartItems
	if err := database.DB.Where("user_id = ?", UserID).Delete(&CartItem).Error; err != nil {
		return false
	}

	return true
}

//user - check userid
func PlaceOrder(c *gin.Context) {
	var PlaceOrder model.PlaceOrder
	if err := c.BindJSON(&PlaceOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the json",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := validate(&PlaceOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if !CheckUser(PlaceOrder.UserID) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "user doesn't exist, please verify user id",
			"error_code": http.StatusNotFound,
		})
		return
	}

	if !ValidAddress(PlaceOrder.UserID, PlaceOrder.AddressID) {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "invalid address, please retry with user's address",
			"error_code": http.StatusConflict,
		})
		return
	}

	if !CheckStock(PlaceOrder.UserID) {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "items in the cart are out of stock, please update the cart to ensure all items are in stock",
			"error_code": http.StatusConflict,
		})
		return
	}

	OrderID, err := CreateOrderID(PlaceOrder.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create order ID",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	TotalAmount, err := CalculateCartTotal(PlaceOrder.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to calculate cart total or the cart is empty",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	if PlaceOrder.PaymentMethod == model.CashOnDelivery && TotalAmount >= 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "please switch to ONLINE payment for total amounts greater than or equal to 1000",
		})
		return
	}

	order := model.Order{
		OrderID:       OrderID,
		UserID:        PlaceOrder.UserID,
		AddressID:     PlaceOrder.AddressID,
		TotalAmount:   float64(TotalAmount),
		FinalAmount:   float64(TotalAmount),
		PaymentMethod: PlaceOrder.PaymentMethod,
		OrderedAt:     time.Now(),
	}

	if PlaceOrder.PaymentMethod == model.CashOnDelivery {
		order.PaymentStatus = model.CODStatusPending
	} else {
		order.PaymentStatus = model.OnlinePaymentPending
	}

	// Attempt to create order record
	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create order",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	// Apply coupon if provided
	if PlaceOrder.CouponCode != "" {
		success, msg := ApplyCouponToOrder(order, PlaceOrder.UserID, PlaceOrder.CouponCode)
		if !success {
			// Rollback order creation if coupon application fails
			database.DB.Where("order_id = ?", OrderID).Delete(&order)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": msg,
			})
			return
		}
	}

	// Transfer cart items to order
	if !CartToOrderItems(PlaceOrder.UserID, OrderID) {
		// Rollback order creation and coupon application if cart items transfer fails
		database.DB.Where("order_id = ?", OrderID).Delete(&order)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to transfer cart items to order",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	// Decrement stock for COD orders
	if PlaceOrder.PaymentMethod == model.CashOnDelivery {
		if !DecrementStock(OrderID) {
			// Rollback order creation, coupon application, and cart items transfer if stock decrement fails
			database.DB.Where("order_id = ?", OrderID).Delete(&order)
			c.JSON(http.StatusConflict, gin.H{
				"status":  false,
				"message": "failed to decrement order stock",
			})
			return
		}
	}

	// Fetch final order details
	if err := database.DB.Where("order_id = ?", OrderID).First(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed fetch order details",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Order is successfully created",
		"data": gin.H{
			"OrderDetails": order,
		},
	})
}

// get response from place order render the pay button with initiate payment logic
//user - check userid by order.userid
func InitiatePayment(c *gin.Context) {
	// Get order id from request body
	var initiatePayment model.InitiatePayment
	if err := c.BindJSON(&initiatePayment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "Failed to bind the JSON",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Check if payment status is confirmed
	var Order []model.Order
	if err := database.DB.Where("order_id = ?", initiatePayment.OrderID).Find(&Order).Error; err != nil {
		PaymentFailedOrderTable(initiatePayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to get payment information",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	for _, v := range Order {
		if v.PaymentStatus == string(model.OnlinePaymentConfirmed) {
			c.JSON(http.StatusAlreadyReported, gin.H{
				"status":  true,
				"message": "Payment already done",
			})
			return
		}
		if v.PaymentStatus == string(model.CODStatusPending) || v.PaymentStatus == string(model.CODStatusConfirmed) {
			c.JSON(http.StatusAlreadyReported, gin.H{
				"status":  true,
				"message": "Customer chose payment via COD",
			})
			return
		}

	}

	// Fetch order details
	var order model.Order
	if err := database.DB.Where("order_id = ?", initiatePayment.OrderID).First(&order).Error; err != nil {
		PaymentFailedOrderTable(initiatePayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to fetch order information",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	switch initiatePayment.PaymentGateway {
	case model.Razorpay:
		HandleRazorpay(c, initiatePayment, order)
	case model.Stripe:
		HandleStripe(c, initiatePayment, order)
	default:
		HandleRazorpay(c, initiatePayment, order)
	}
}

func CheckUser(UserID uint) bool {

	var User model.User
	if err := database.DB.Where("id = ?", UserID).First(&User).Error; err != nil {
		return false
	}
	return true
}
func RestaurantIDByProductID(ProductID uint) uint {

	var Product model.Product
	if err := database.DB.Where("id = ?", ProductID).First(&Product).Error; err != nil {
		return 0
	}
	return Product.RestaurantID
}

// active orders of restaurants
//restaurant
func OrderHistoryRestaurants(c *gin.Context) {
	//Restaurant id, if order status is provided use it or get the whole history
	var Request model.OrderHistoryRestaurants
	if err := c.Bind(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the json",
			"error_code": http.StatusBadRequest,
		})
		return
	}
	//if provided with order_status show the specific order's of that status

	var OrderItems []model.OrderItem
	if Request.OrderStatus != "" {
		//condition one order status not empty
		//return all the orders with order_id, restaurant_id,order_status is met with the condition
		if err := database.DB.Where("restaurant_id = ? AND order_status = ?", Request.RestaurantID, Request.OrderStatus).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":     false,
				"message":    "failed to fetch orders assigned to this restaurant",
				"error_code": http.StatusNotFound,
			})
			return
		}
	} else {
		//condition two order status empty
		//return all the orders with order_id, restaurant_id is met with the condition
		if err := database.DB.Where("restaurant_id = ?", Request.RestaurantID).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":     false,
				"message":    "failed to fetch orders assigned to this restaurant",
				"error_code": http.StatusNotFound,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"message": gin.H{
			"orderhistory": OrderItems,
		},
	})
}

//user
func UserOrderHistory(c *gin.Context) {
	//same like restaurant
	var Request model.UserOrderHistory
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  true,
			"message": "failed to bind the request",
		})
		return
	}
	var Orders []model.Order

	if Request.PaymentStatus != "" {
		if err := database.DB.Where("user_id = ? AND payment_status = ?", Request.UserID, Request.PaymentStatus).Find(&Orders).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  true,
				"message": "failed to fetch order history of the user by order status",
			})
			return
		}
	} else {
		if err := database.DB.Where("user_id = ?", Request.UserID).Find(&Orders).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  true,
				"message": "failed to fetch order history of the user",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully fetched order history",
		"data": gin.H{
			"orderhistory": Orders,
		},
	})
}

func GetOrderInfoByOrderID(c *gin.Context) {
	//get order id
	var Request model.GetOrderInfoByOrderID
	if err := c.Bind(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", Request.OrderID).Find(&OrderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch order information",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successfully retrieved OrderInformation",
		"data": gin.H{
			"orderitems": OrderItems,
		},
	})
}

func PaymentDetailsByOrderID(c *gin.Context) {

	var Request model.PaymentDetailsByOrderID
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var PaymentDetails []model.Payment
	if Request.PaymentStatus != "" {
		if err := database.DB.Where("order_id = ? AND payment_status = ?", Request.OrderID, Request.PaymentStatus).Find(&PaymentDetails).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  true,
				"message": "failed to fetch payment information",
			})
			return
		}

	} else {
		if err := database.DB.Where("order_id = ?", Request.OrderID).Find(&PaymentDetails).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  true,
				"message": "failed to fetch payment information",
			})
			return
		}
	}

	if len(PaymentDetails) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  true,
			"message": "failed to fetch payment information with the specified order_status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successfully retrieved payment information",
		"data": gin.H{
			"paymentdetails": PaymentDetails,
		},
	})
}

//restaurant - check restid with product.rest id
func UpdateOrderStatusForRestaurant(c *gin.Context) {
	var Request model.UpdateOrderStatusForRestaurant

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var OrderItemDetail model.OrderItem
	if err := database.DB.Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).First(&OrderItemDetail).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to fetch order information for the specific product",
			"error_code": http.StatusNotFound,
		})
		return
	}

	var NextOrderStatus string

	if OrderItemDetail.OrderStatus == model.OrderStatusProcessing {
		NextOrderStatus = model.OrderStatusInPreparation
	} else {
		OrderTransition := []string{model.OrderStatusInPreparation, model.OrderStatusPrepared, model.OrderStatusOntheway, model.OrderStatusDelivered}

		//get current index of the status transition
		fmt.Println(OrderItemDetail.OrderStatus)
		var orderIndex int
		for i, v := range OrderTransition {
			if OrderItemDetail.OrderStatus == v {
				orderIndex = i
				break
			}
		}

		//check if transition ended
		if orderIndex == len(OrderTransition)-1 {
			c.JSON(http.StatusNotFound, gin.H{
				"status":     false,
				"message":    "Reached maximum level of order transition",
				"error_code": http.StatusNotFound,
			})
			return
		}

		NextOrderStatus = OrderTransition[orderIndex+1]
		fmt.Println(NextOrderStatus)
	}

	//update the new status to the orderitem table
	if err := database.DB.Model(&model.OrderItem{}).Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).Update("order_status", NextOrderStatus).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to update order status for the specific product",
			"error_code": http.StatusNotFound,
		})
		return
	}

	OrderItemDetail.OrderStatus = NextOrderStatus
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successfully changed to next order status",
		"data": gin.H{
			"orderdetails": OrderItemDetail,
		},
	})
}

//user - check userid by order.userid
func CancelOrderedProduct(c *gin.Context) {
	//get orderid product
	var Request model.CancelOrderedProduct
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var OrderItem model.OrderItem
	if Request.ProductId != 0 {
		if err := database.DB.Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductId).First(&OrderItem).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to fetch the order item",
			})
			return
		}
	} else {
		if err := database.DB.Where("order_id = ?", Request.OrderID).First(&OrderItem).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to fetch the order item",
			})
			return
		}
	}

	//check order status is processing
	if OrderItem.OrderStatus != model.OrderStatusProcessing && OrderItem.OrderStatus != model.OrderStatusInPreparation {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "Order can only be cancelled during processing or in preparation status",
		})
		return
	}

	//if mentioned with productid only cancel that
	var OrderItems []model.OrderItem
	if Request.ProductId != 0 {
		if err := database.DB.Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductId).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to fetch the order item",
			})
			return
		}
		//change status to cancelled
		if err := database.DB.Model(&model.OrderItem{}).Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductId).Update("order_status", model.OrderStatusCancelled).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to cancel the order item",
			})
			return
		}
	} else {
		if err := database.DB.Where("order_id = ?", Request.OrderID).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to fetch the order item",
			})
			return
		}

		if err := database.DB.Model(&model.OrderItem{}).Where("order_id = ?", Request.OrderID).Update("order_status", model.OrderStatusCancelled).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to cancel the order item",
			})
			return
		}
		//if no productid is mentioned cancel all the orders equal to and under preparing
		//increment stockleft
	}

	done := IncrementStock(OrderItems)
	if !done {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "failed to increment order stock",
		})
		return
	}

	var order model.Order
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch order information",
		})
		return
	}

	if order.PaymentStatus == model.OnlinePaymentConfirmed {
		done = ProvideWalletRefundToUser(order.UserID, OrderItems)
		if !done {
			c.JSON(http.StatusConflict, gin.H{
				"status":  false,
				"message": "failed to refund to the wallet",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully cancelled the order",
		"data": gin.H{
			"order_id": Request.OrderID,
		},
	})
}

func ProvideWalletRefundToUser(UserID uint, OrderItems []model.OrderItem) bool {
	var sum uint
	for _, item := range OrderItems {
		sum += item.Amount

		var Restaurant model.Restaurant
		if err := database.DB.Where("id = ?", item.RestaurantID).First(&Restaurant).Error; err != nil {
			return false
		}

		Restaurant.WalletAmount -= int(item.Amount)

		if err := database.DB.Updates(&Restaurant).Error; err != nil {
			return false
		}
	}

	var User model.User
	if err := database.DB.Where("id = ?", UserID).First(&User).Error; err != nil {
		return false
	}

	User.WalletAmount += sum
	if err := database.DB.Updates(&User).Error; err != nil {
		return false
	}

	return true
}

func SplitMoneyToRestaurants(OrderID string) bool {
	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", OrderID).Find(&OrderItems).Error; err != nil {
		return false
	}
	for _, item := range OrderItems {
		var Restaurant model.Restaurant
		if err := database.DB.Where("id = ?", item.RestaurantID).First(&Restaurant).Error; err != nil {
			return false
		}

		Restaurant.WalletAmount += int(item.Amount)
		if err := database.DB.Updates(&Restaurant).Error; err != nil {
			return false
		}

	}
	return true
}

func IncrementStock(OrderItems []model.OrderItem) bool {

	//loop and get the cancelled orders
	for _, v := range OrderItems {
		if v.OrderStatus == model.OrderStatusCancelled {
			//get product id
			var Product model.Product
			if err := database.DB.Where("id = ?", v.ProductID).First(&Product).Error; err != nil {
				return false
			}
			//update stock left by incrementing it by producty.stockleft - v.quantity
			Product.StockLeft += v.Quantity
			if err := database.DB.Updates(&Product).Error; err != nil {
				return false
			}
		}
	}
	return true
}

func DecrementStock(OrderID string) bool {
	//get orderitems
	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", OrderID).Find(&OrderItems).Error; err != nil {
		return false
	}

	//loop and get the cancelled orders
	for _, v := range OrderItems {
		if v.OrderStatus == model.OrderStatusProcessing || v.OrderStatus == model.OrderStatusInPreparation {
			//get product id
			var Product model.Product
			if err := database.DB.Where("id = ?", v.ProductID).First(&Product).Error; err != nil {
				return false
			}
			//update stock left by decrementing it by producty.stockleft - v.quantity
			Product.StockLeft -= v.Quantity
			if Product.StockLeft <= 0 {
				return false
			}
			if err := database.DB.Updates(&Product).Error; err != nil {
				return false
			}
		}
	}
	return true
}

//user - check userid by order.userid
func UserReviewonOrderItem(c *gin.Context) {
	//orderid, productid,review text
	var Request model.UserReviewonOrderItem
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind request",
		})
		return
	}
	fmt.Println(Request)
	//get orderitem
	var OrderItem model.OrderItem
	if err := database.DB.Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).First(&OrderItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retreive the order item",
		})
		return
	}
	//check delivered
	if OrderItem.OrderStatus != model.OrderStatusDelivered {
		c.JSON(http.StatusConflict, gin.H{"status": false, "message": "reviews can only be added after order is delivered"})
		return
	}
	//check review text
	OrderItem.OrderReview = Request.ReviewText
	if err := database.DB.Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).Updates(&OrderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to add order review, please try again"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "successfully added the review"})
}

//user - check userid by order.userid
func UserRatingOrderItem(c *gin.Context) {
	//get the orderid,productid,rating
	var Request model.UserRatingOrderItem
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "failed to bind the json"})
		return
	}

	if Request.UserRating <= 0 && Request.UserRating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "enter a valid rating between 1 to 5"})
		return
	}

	//check if user already rated
	var OrderItem model.OrderItem
	if err := database.DB.Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).First(&OrderItem).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "failed to retrieve order information"})
		return
	}

	if OrderItem.OrderRating != 0 {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "user already rated the order"})
		return
	}

	var OrderItems []model.OrderItem
	if err := database.DB.Model(&model.OrderItem{}).Where("product_id = ? AND order_rating BETWEEN ? AND ?", Request.ProductID, 1, 5).Update("order_rating", Request.UserRating).Find(&OrderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update product rating"})
		return
	}

	var RatingSum float64
	for _, v := range OrderItems {
		RatingSum += v.OrderRating
	}

	newRating := (RatingSum + Request.UserRating) / float64(len(OrderItems)+1)
	fmt.Println(newRating)

	//update
	if err := database.DB.Model(&model.OrderItem{}).Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).Update("order_rating", Request.UserRating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update product rating"})
		return
	}

	if err := database.DB.Model(&model.Product{}).Where("id = ?", Request.ProductID).Update("average_rating", newRating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update product rating"})
		return
	}
	//success
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "successfully updated rating"})
}


func UpdatePaymentGatewayMethod(OrderID string, PaymentGateway string) bool {
	if err := database.DB.Model(&model.Payment{}).Where("order_id = ?", OrderID).Update("payment_gateway", PaymentGateway).Error; err != nil {
		return false
	}
	return true
}
