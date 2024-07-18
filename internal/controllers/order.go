package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func ValidRestaurant(RestauarantID uint) (model.Restaurant, bool) {
	var response model.Restaurant
	if err := database.DB.Where("id = ?", RestauarantID).First(&response).Error; err != nil {
		return response, false
	}
	return model.Restaurant{}, true
}

func ValidAddress(UserID uint, AddressID uint) bool {
	var Address model.Address
	if err := database.DB.Where("user_id = ? AND address_id = ?", UserID, AddressID).First(&Address).Error; err != nil {
		return false
	}
	return true
}

func CheckStock(UserID uint) (ItemCount uint, ok bool) {

	var CartItems []model.CartItems
	//get cartItems,find the product check with stockleft
	if err := database.DB.Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {
		return ItemCount, false
	}
	for _, v := range CartItems {
		var Product model.Product
		if err := database.DB.Where("id = ?", v.ProductID).First(&Product).Error; err != nil {
			return ItemCount, false
		}
		if v.Quantity > Product.StockLeft {
			return ItemCount, false
		}

		ItemCount += v.Quantity
	}
	return ItemCount, true
}

func CartToOrderItems(UserID uint, RestaurantID uint, Order model.Order) bool {
	var CartItems []model.CartItems
	if err := database.DB.Where("user_id = ? AND restaurant_id = ?", UserID, RestaurantID).Find(&CartItems).Error; err != nil {

		return false
	}

	fmt.Printf("OrderID: %s\nUserID: %d\nAddressID: %d\nItemCount: %d\nCouponCode: %s\nCouponDiscountAmount: %.2f\nProductOfferAmount: %.2f\nTotalAmount: %.2f\nFinalAmount: %.2f\nPaymentMethod: %s\nPaymentStatus: %s\nOrderedAt: %v\n",
		Order.OrderID, Order.UserID, Order.AddressID, Order.ItemCount, Order.CouponCode, Order.CouponDiscountAmount, Order.ProductOfferAmount, Order.TotalAmount, Order.FinalAmount, Order.PaymentMethod, Order.PaymentStatus, Order.OrderedAt)

	for _, v := range CartItems {

		var Product model.Product
		if err := database.DB.Where("id = ?", v.ProductID).First(&Product).Error; err != nil {
			return false
		}

		OrderItem := model.OrderItem{
			OrderID:            Order.OrderID,
			UserID:             UserID,
			ProductID:          v.ProductID,
			Quantity:           v.Quantity,
			Amount:             (float64(v.Quantity) * Product.Price),
			ProductOfferAmount: float64(v.Quantity) * float64(Product.OfferAmount),
			CookingRequest:     v.CookingRequest,
			OrderStatus:        model.OrderStatusInitiated,
			RestaurantID:       RestaurantIDByProductID(v.ProductID),
		}

		//after offer and coupon deduction amount
		//get ratio for coupon reduction
		fmt.Println(Order)
		couponDeduct := Order.CouponDiscountAmount * (float64(OrderItem.Quantity) / float64(Order.ItemCount))
		afterDeduct := OrderItem.Amount - (OrderItem.ProductOfferAmount + couponDeduct)

		fmt.Println("coupon deduct is : ", couponDeduct)
		fmt.Println(OrderItem)
		OrderItem.AfterDeduction = afterDeduct

		if err := database.DB.Create(&OrderItem).Error; err != nil {
			return false
		}
	}

	//then remove the cartdetail for that user
	var CartItem model.CartItems
	if err := database.DB.Where("user_id = ? AND restaurant_id = ?", UserID, RestaurantID).Delete(&CartItem).Error; err != nil {
		return false
	}

	return true
}

// user - check userid
func PlaceOrder(c *gin.Context) {
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

	var PlaceOrder model.PlaceOrder
	if err := c.BindJSON(&PlaceOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the json",
			"error_code": http.StatusBadRequest,
		})
		return
	}
	PlaceOrder.UserID = UserID

	if err := utils.Validate(&PlaceOrder); err != nil {
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

	var Restaurant model.Restaurant
	Restaurant, restExist := ValidRestaurant(PlaceOrder.RestaurantID)
	if !restExist {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "invalid restaurant_id, please ensure restauarant exists",
			"error_code": http.StatusConflict,
		})
		return
	}

	if Restaurant.Blocked {
		c.JSON(http.StatusConflict, gin.H{"status": false, "message": "cannot place orders of blocked restaurants", "error_code": http.StatusConflict})
		return
	}

	ItemCount, ok := CheckStock(PlaceOrder.UserID)

	if !ok {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "items in the cart are out of stock, please update the cart to ensure all items are in stock",
			"error_code": http.StatusConflict,
		})
		return
	}

	// OrderID, err := CreateOrderID(PlaceOrder.UserID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"status":     false,
	// 		"message":    "failed to create order ID",
	// 		"error_code": http.StatusInternalServerError,
	// 	})
	// 	return
	// }
	OrderID := uuid.New().String()

	TotalAmount, ProductOffer, err := CalculateCartTotal(PlaceOrder.UserID, PlaceOrder.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to calculate cart total or the cart is empty",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	if PlaceOrder.PaymentMethod == model.CashOnDelivery && TotalAmount >= model.CODMaximumAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "please switch to ONLINE payment for total amounts greater than or equal to 1000",
		})
		return
	}

	order := model.Order{
		OrderID:            OrderID,
		UserID:             PlaceOrder.UserID,
		AddressID:          PlaceOrder.AddressID,
		ItemCount:          ItemCount,
		ProductOfferAmount: float64(ProductOffer),
		TotalAmount:        float64(TotalAmount),
		FinalAmount:        float64(TotalAmount) - float64(ProductOffer),
		PaymentMethod:      PlaceOrder.PaymentMethod,
		OrderedAt:          time.Now(),
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
		var success bool
		var msg string
		success, msg, order = ApplyCouponToOrder(order, PlaceOrder.UserID, PlaceOrder.CouponCode)
		if !success {
			database.DB.Where("order_id = ?", OrderID).Delete(&order)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": msg,
			})
			return
		}
	}

	// Transfer cart items to order
	if !CartToOrderItems(PlaceOrder.UserID, PlaceOrder.RestaurantID, order) {
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
			"order_details": order,
		},
	})
}

// get response from place order render the pay button with initiate payment logic
// user - check userid by order.userid
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
	case model.Wallet:
		HandleWalletPayment(initiatePayment.OrderID, order.UserID, c)
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

// active orders of restaurants
// restaurant
func OrderHistoryRestaurants(c *gin.Context) {
	//check restaurant api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.RestaurantRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	RestaurantID, ok := RestIDfromEmail(email)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve restaurant information",
		})
		return
	}
	//Restaurant id, if order status is provided use it or get the whole history
	Request := c.Query("order_status")

	var OrderItems []model.OrderItem
	if Request != "" {
		//condition one order status not empty
		//return all the orders with order_id, restaurant_id,order_status is met with the condition
		if err := database.DB.Where("restaurant_id = ? AND order_status = ?", RestaurantID, Request).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to fetch orders assigned to this restaurant",
			})
			return
		}
	} else {
		//condition two order status empty
		//return all the orders with order_id, restaurant_id is met with the condition
		if err := database.DB.Where("restaurant_id = ?", RestaurantID).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to fetch orders assigned to this restaurant",
			})
			return
		}
	}

	for i := 0; i < len(OrderItems); i++ {
		OrderItems[i].AfterDeduction = RoundDecimalValue(OrderItems[i].AfterDeduction)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"message": gin.H{
			"orderhistory": OrderItems,
		},
	})
}

// user
func UserOrderItems(c *gin.Context) {
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

	//same like restaurant
	var Request model.UserOrderHistory
	Request.OrderID = c.Query("order_id")
	Request.UserID = UserID
	var OrderItems []model.OrderItem

	if Request.OrderID != "" {
		if err := database.DB.Where("user_id = ? AND order_id = ?", Request.UserID, Request.OrderID).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  true,
				"message": "failed to fetch specified orderitems",
			})
			return
		}
	}

	for i := 0; i < len(OrderItems); i++ {
		OrderItems[i].AfterDeduction = RoundDecimalValue(OrderItems[i].AfterDeduction)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully fetched order items based on order status",
		"data": gin.H{
			"orderhistory": OrderItems,
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
			"message": "failed to fetch payment information of this order_id",
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

// restaurant - check restid with product.rest id
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

	//check restaurant api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.RestaurantRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	RestaurantID, ok := RestIDfromEmail(email)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var OrderItemDetail model.OrderItem
	if err := database.DB.Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).First(&OrderItemDetail).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch order information for the specific product",
		})
		return
	}

	if RestaurantID != OrderItemDetail.RestaurantID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var NextOrderStatus string

	if OrderItemDetail.OrderStatus == model.OrderStatusProcessing {
		NextOrderStatus = model.OrderStatusInPreparation
	} else {
		OrderTransition := []string{model.OrderStatusInPreparation, model.OrderStatusPrepared, model.OrderStatusOntheway}

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

func UserIDfromOrderID(OrderID string) (uint, bool) {
	var Order model.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&Order).Error; err != nil {
		return Order.UserID, false
	}
	return Order.UserID, true
}

// user - check userid by order.userid
func CancelOrderedProductOnline(c *gin.Context) {
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	JWTUserID, _ := UserIDfromEmail(email)

	var Request model.CancelOrderedProduct
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	oUserID, _ := UserIDfromOrderID(Request.OrderID)

	if JWTUserID != oUserID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var order model.Order
	if err := database.DB.Where("order_id =?", Request.OrderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch order information",
		})
		return
	}
	if order.PaymentMethod != model.OnlinePayment {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"status":  false,
			"message": "order payment method is not ONLINE",
		})
		return
	}

	if order.PaymentStatus != model.OnlinePaymentConfirmed {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "order has not received the payment, hence cannot initiate the cancellation",
		})
		return
	}

	var OrderItems []model.OrderItem
	if Request.ProductId != 0 {
		// Fetch individual product
		if err := database.DB.Where("order_id =? AND product_id =? AND order_status = ?", Request.OrderID, Request.ProductId, model.OrderStatusInitiated).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to fetch the order item"})
			return
		}
	} else {
		// Fetch all orders
		if err := database.DB.Where("order_id =? AND order_status= ?", Request.OrderID, model.OrderStatusInitiated).Find(&OrderItems).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to fetch the order item"})
			return
		}
	}

	if len(OrderItems) == 0 {
		c.JSON(http.StatusConflict, gin.H{"status": false, "message": "No eligible items found for cancellation"})
		return
	}

	// Update order status to cancelled
	for _, item := range OrderItems {
		item.OrderStatus = model.OrderStatusCancelled
		if err := database.DB.Where("order_id = ? AND product_id = ?", item.OrderID, item.ProductID).Updates(&item).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "failed to do cancellation"})
		}
	}

	done := IncrementStock(OrderItems)
	if !done {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "failed to increment order stock",
		})
		return
	}

	done = ProvideWalletRefundToUser(order.UserID, OrderItems)
	if !done {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "failed to refund to the wallet",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully cancelled the order",
		"data": gin.H{
			"order_id": Request.OrderID,
		},
	})
}

func CancelOrderedProductCOD(c *gin.Context) {
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	JWTUserID, _ := UserIDfromEmail(email)

	var Request model.CancelOrderedProduct
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	oUserID, _ := UserIDfromOrderID(Request.OrderID)

	if JWTUserID != oUserID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var order model.Order
	if err := database.DB.Where("order_id =?", Request.OrderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch order information",
		})
		return
	}
	if order.PaymentMethod != model.CashOnDelivery {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"status":  false,
			"message": "order payment method is not COD",
		})
		return
	}

	if order.PaymentMethod != model.CashOnDelivery {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "order is not COD, hence cannot initiate the cancellation",
		})
		return
	}

	var OrderItems []model.OrderItem
	// Fetch all orders
	if err := database.DB.Where("order_id =? AND order_status IN (?,?)", Request.OrderID, model.OrderStatusProcessing, model.OrderStatusInPreparation).Find(&OrderItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to fetch the order item"})
		return
	}

	if len(OrderItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "No eligible items found for cancellation,order item should be either in Processing or Preparation stage"})
		return
	}

	// Update order status to cancelled
	for _, item := range OrderItems {
		item.OrderStatus = model.OrderStatusCancelled
		if err := database.DB.Where("order_id = ? AND product_id = ?", item.OrderID, item.ProductID).Updates(&item).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "failed to do cancellation"})
		}
	}

	done := IncrementStock(OrderItems)
	if !done {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "failed to increment order stock",
		})
		return
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
	// Fetch the order using the first item's order ID to get the coupon discount amount
	var Order model.Order
	if err := database.DB.Where("order_id =?", OrderItems[0].OrderID).First(&Order).Error; err != nil {
		return false
	}
	var sum float64

	// Iterate over each item to calculate individual refunds
	for _, item := range OrderItems {

		// Fetch and update the restaurant's wallet amount
		var Restaurant model.Restaurant
		if err := database.DB.Where("id =?", item.RestaurantID).First(&Restaurant).Error; err != nil {
			return false
		}

		Restaurant.WalletAmount -= (item.AfterDeduction)

		rWalletHistory := model.RestaurantWalletHistory{
			TransactionTime: time.Now(),
			Type:            model.WalletOutgoing,
			OrderID:         item.OrderID,
			RestaurantID:    item.RestaurantID,
			Amount:          item.AfterDeduction,
			CurrentBalance:  Restaurant.WalletAmount,
			Reason:          "Order Refund",
		}

		if !CreateRestaurantWalletHistory(rWalletHistory) {
			return false
		}

		if err := database.DB.Save(&Restaurant).Error; err != nil {
			return false
		}

		sum += item.AfterDeduction
	}

	// Update the user's wallet amount
	var User model.User
	if err := database.DB.Where("id =?", UserID).First(&User).Error; err != nil {
		return false
	}
	User.WalletAmount += (sum)

	var WalletHistory model.UserWalletHistory

	WalletHistory.TransactionTime = time.Now()
	WalletHistory.Amount = float64(sum)
	WalletHistory.UserID = UserID
	WalletHistory.OrderID = Order.OrderID
	WalletHistory.Reason = model.WalletTxTypeOrderRefund
	WalletHistory.CurrentBalance = User.WalletAmount
	WalletHistory.Type = model.WalletIncoming

	if err := database.DB.Create(&WalletHistory).Error; err != nil {
		return false
	}

	if err := database.DB.Save(&User).Error; err != nil {
		return false
	}

	return true
}

func SplitMoneyToRestaurants(OrderID string) bool {
	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", OrderID).Find(&OrderItems).Error; err != nil {
		return false
	}

	var totalOrderAmount float64
	for _, item := range OrderItems {
		totalOrderAmount += float64(item.Amount)
	}

	var Order model.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&Order).Error; err != nil {
		return false
	}

	for _, item := range OrderItems {
		var Restaurant model.Restaurant
		if err := database.DB.Where("id = ?", item.RestaurantID).First(&Restaurant).Error; err != nil {
			return false
		}
		Restaurant.WalletAmount += (item.AfterDeduction)

		rWalletHistory := model.RestaurantWalletHistory{
			TransactionTime: time.Now(),
			Type:            model.WalletIncoming,
			OrderID:         item.OrderID,
			RestaurantID:    Restaurant.ID,
			Amount:          item.AfterDeduction,
			CurrentBalance:  Restaurant.WalletAmount,
			Reason:          "Order Payment",
		}

		if !CreateRestaurantWalletHistory(rWalletHistory) {
			return false
		}

		if err := database.DB.Save(&Restaurant).Error; err != nil {
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

// user - check userid by order.userid
func UserReviewonOrderItem(c *gin.Context) {
	//check user api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	JWTUserID, _ := UserIDfromEmail(email)

	//orderid, productid,review text
	var Request model.UserReviewonOrderItem
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind request",
		})
		return
	}
	oUserID, _ := UserIDfromOrderID(Request.OrderID)
	if JWTUserID != oUserID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

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

// user - check userid by order.userid
func UserRatingOrderItem(c *gin.Context) {
	//check user api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	JWTUserID, _ := UserIDfromEmail(email)

	//get the orderid,productid,rating
	var Request model.UserRatingOrderItem
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "failed to bind the json"})
		return
	}

	oUserID, _ := UserIDfromOrderID(Request.OrderID)
	if JWTUserID != oUserID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
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

	if OrderItem.OrderStatus != model.OrderStatusDelivered {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "user can only rate an order after delivery"})
		return
	}

	if OrderItem.OrderRating != 0 {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "user already rated the order"})
		return
	}

	//get the count of the product rating and current avg rating
	//use rating_sum + add_new_rating /(rating_count+1)\
	//update the product rating,count etc
	//update the rating on order items
	tx := database.DB.Begin()

	if err := tx.Model(&model.OrderItem{}).Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductID).Update("order_rating", Request.UserRating).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update order rating in the order item table for the specified product"})
		return
	}
	var product model.Product
	if err := tx.Where("id = ?", Request.ProductID).First(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "product not found"})
		return
	}

	//update the products rating sum and count
	newRatingSum := product.RatingSum + Request.UserRating
	newRatingCount := product.RatingCount + 1
	newAverageRating := newRatingSum / float64(newRatingCount)

	if err := tx.Model(&product).Updates(model.Product{RatingSum: newRatingSum, RatingCount: newRatingCount, AverageRating: newAverageRating}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update product rating"})
		return
	}

	//commit the tx
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update product rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "successfully updated rating", "new_average_rating": newAverageRating})
}

func UpdatePaymentGatewayMethod(OrderID string, PaymentGateway string) bool {
	if err := database.DB.Model(&model.Payment{}).Where("order_id = ?", OrderID).Update("payment_gateway", PaymentGateway).Error; err != nil {
		return false
	}
	return true
}

func GetOrderInfoByOrderIDasJSON(c *gin.Context) {
	//get order id

	OrderID := c.Query("order_id")
	if OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "order_id is empty,mention order_id as query params"})
		return
	}

	var Order model.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&Order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "order_id is not present",
		})
		return
	}

	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", OrderID).Find(&OrderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "no order items for order_id in the table,make sure the order contains order items",
		})
		return
	}

	for i := 0; i < len(OrderItems); i++ {
		OrderItems[i].AfterDeduction = RoundDecimalValue(OrderItems[i].AfterDeduction)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successfully retrieved OrderInformation",
		"data": gin.H{
			"order_information": Order,
			"items_ordered":     OrderItems,
		},
	})
}

func CancelCODOrder(c *gin.Context) {
	OrderID := c.Query("order_id")
	if OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "order_id is empty,mention order_id as query params"})
		return
	}

}
func SendOrderDeliveryVerificationCode(c *gin.Context) {
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": false, "message": "unauthorized request"})
		return
	}
	OrderID := c.Query("order_id")
	if OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "order_id is empty,mention order_id as query params"})
		return
	}

	oUserID, _ := UserIDfromOrderID(OrderID)
	JWTUserID, _ := UserIDfromEmail(email)

	if JWTUserID != oUserID {
		c.JSON(http.StatusUnauthorized, gin.H{"status": false, "message": "unauthorized request"})
		return
	}

	otp, err := SendOrderDeliveryCode(OrderID)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": false, "message": "otp is sent in mail", "otp": otp})
}

// function that sends the code
func SendOrderDeliveryCode(OrderID string) (int64, error) {
	var DeliveryVerification model.DeliveryVerification
	//check order_id
	database.DB.Where("order_id = ?", OrderID).First(&DeliveryVerification)

	var Order model.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&Order).Error; err != nil {
		return 0, errors.New("order_id doesnt exist")
	}

	//check if its delivered
	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", OrderID).Find(&OrderItems).Error; err != nil {
		return 0, errors.New("failed to find orderitems")
	}

	for _, v := range OrderItems {
		if v.OrderStatus == model.OrderStatusDelivered {
			return 0, errors.New("order is already delivered, cannot sent delivery code for this order")
		}
	}

	now := time.Now().Unix()
	if now < int64(DeliveryVerification.LastSentAT+model.DeliveryVerificationOTPCooldownTime) {
		return 0, errors.New("wait for a min before trying again")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	DeliveryVerification.OTP = uint(r.Intn(900000) + 100000)
	from := "foodbuddycode@gmail.com"
	appPassword := os.Getenv("SMTPAPP")
	auth := smtp.PlainAuth("", from, appPassword, "smtp.gmail.com")
	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>FoodBuddy Email Verification</title>
		<style>
			.button {
				background-color: #4CAF50;
				border: none;
				color: white;
				padding: 15px 32px;
				text-align: center;
				text-decoration: none;
				display: inline-block;
				font-size: 16px;
				margin: 4px 2px;
				cursor: pointer;
			}
		</style>
	</head>
	<body>
		<h1>FoodBuddy Delivery Verification</h1>
		<p>Your OTP is : %v </p>
	</body>
	</html>
	`, DeliveryVerification.OTP)

	// Set up the email message
	msg := []byte("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n" +
		"Subject: FoodBuddy Email Verification\r\n\r\n" +
		htmlContent)

	UserID, _ := UserIDfromOrderID(OrderID)
	var User model.User
	database.DB.Where("id = ?", UserID).First(&User)
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{User.Email}, []byte(msg))
	if err != nil {
		return 0, errors.New("failed to send email")
	}

	DeliveryVerification.OrderID = OrderID
	DeliveryVerification.LastSentAT = uint(now)
	DeliveryVerification.UserID = UserID
	database.DB.Where("order_id = ?", OrderID).Save(&DeliveryVerification)

	return (int64(DeliveryVerification.OTP)), nil
}

func ConfirmCODPayment(c *gin.Context) {
	var Request model.ConfirmCODPayment
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "provide order_id in the json payload" + err.Error()})
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

	RestID, _ := RestIDfromEmail(email)

	//get order information
	var Order model.Order
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&Order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "order_id is not present"})
		return
	}

	if RestID != Order.RestaurantID {
		c.JSON(http.StatusUnauthorized, gin.H{"status": false, "message": "unauthorized request,order doesnt belong to this restaurant"})
		return
	}

	//check if the order is COD
	if Order.PaymentMethod != model.CashOnDelivery {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "order_id given is not COD"})
		return
	}

	if Order.PaymentStatus == model.CODStatusConfirmed {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "order_id payment already done"})
		return
	}

	Order.PaymentStatus = model.CODStatusConfirmed
	if err := database.DB.Where("order_id = ?", Request.OrderID).Updates(&Order).Error; err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "failed to update COD payment status to paid, please try again"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "order COD payment confirmed"})

}

func DeliveryComplete(c *gin.Context) {
	var Request model.ConfirmDelivery
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "order_id and delivery_otp should be present on the json payload"})
		return
	}

	// Get order information
	var Order model.Order
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&Order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "order_id is not present"})
		return
	}

	if Order.PaymentStatus != model.OnlinePaymentConfirmed && Order.PaymentStatus != model.CODStatusConfirmed {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "payment should be confirmed before using this endpoint"})
		return
	}

	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", Request.OrderID).Find(&OrderItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to fetch order items"})
		return
	}

	var DeliveryVerification model.DeliveryVerification
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&DeliveryVerification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to fetch delivery verification information"})
		return
	}

	if Request.DeliveryOTP != DeliveryVerification.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"status": false, "message": "delivery_otp provided doesnt match with database"})
		return
	}

	for _, item := range OrderItems {
		if item.OrderStatus != model.OrderStatusCancelled {
			item.OrderStatus = model.OrderStatusDelivered
			if err := database.DB.Where("order_id = ? AND product_id = ?", item.OrderID, item.ProductID).Save(&item).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update order item status"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "order status updated to delivered"})

}
