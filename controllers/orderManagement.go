package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

func CreateOrderID(UserID uint) (string, error) {
	//get user info - name
	var UserInfo model.User
	if err := database.DB.Where("id = ?", UserID).First(&UserInfo).Error; err != nil {
		return "", errors.New("failed to fetch user information")
	}
	//create and check for orderID...if order id is present create another one
	random := utils.GenerateRandomString(10)
	OrderID := fmt.Sprintf("%v_%v", UserInfo.Name, random)
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

		OrderItem := model.OrderItem{
			OrderID:        OrderID,
			ProductID:      v.ProductID,
			Quantity:       v.Quantity,
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

func PlaceOrder(c *gin.Context) {
	//bind json (userid,addressid,paymentmethod)
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
	//check user
	ok := CheckUser(PlaceOrder.UserID)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "user doesnt exist, please verify user id",
			"error_code": http.StatusConflict,
		})
		return
	}
	//check address
	ok = ValidAddress(PlaceOrder.UserID, PlaceOrder.AddressID)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "invalid address, please retry with user's address",
			"error_code": http.StatusConflict,
		})
		return
	}
	//check for stock avaiability using check stock
	available := CheckStock(PlaceOrder.UserID)
	if !available {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "items on the cart is out of stock, please update the cart to make sure the cart is containing items in stock",
			"error_code": http.StatusConflict,
		})
		return
	}
	//create orderid
	OrderID, err := CreateOrderID(PlaceOrder.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create orderid",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//get cart total
	TotalAmount, err := CalculateCartTotal(PlaceOrder.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to calculate cart total or the cart is empty",
			"error_code": http.StatusInternalServerError,
		})
		return
	}
	//create order row with userid,addressid,totalprice,paymentmethod,etc...,payment status as "pending"
	var Order model.Order
	Order.OrderID = OrderID
	Order.UserID = PlaceOrder.UserID
	Order.AddressID = PlaceOrder.AddressID
	Order.TotalAmount = float64(TotalAmount)
	Order.PaymentMethod = PlaceOrder.PaymentMethod
	Order.OrderedAt = time.Now()

	if PlaceOrder.PaymentMethod == model.CashOnDelivery {
		Order.PaymentStatus = model.PaymentComplete
	} else {
		Order.PaymentStatus = model.PaymentPending
	}

	fmt.Println(Order)
	if err := database.DB.Where("order_id = ?", OrderID).FirstOrCreate(&Order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create order",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//get cart details
	//insert everything to orderItems
	ok = CartToOrderItems(PlaceOrder.UserID, OrderID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to transfer cart items to order",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	if PlaceOrder.PaymentMethod == model.CashOnDelivery {
		//decrement stock
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Order is Successfully created",
		"data": gin.H{
			"OrderDetails": Order,
		},
	})
}

// get response from place order render the pay button with initiate payment logic
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
		if v.PaymentStatus == string(model.PaymentComplete) {
			c.JSON(http.StatusAlreadyReported, gin.H{
				"status":  true,
				"message": "Payment already done",
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

	// Create Razorpay order
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))
	data := map[string]interface{}{
		"amount":          order.TotalAmount * 100, // Amount in paisa
		"currency":        "INR",
		"receipt":         order.OrderID,
		"payment_capture": 1,
	}
	rzpOrder, err := client.Order.Create(data, nil)
	if err != nil {
		PaymentFailedOrderTable(initiatePayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to create Razorpay order",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//add to payment tables
	RazorpayOrderID := rzpOrder["id"]
	PaymentDetails := model.Payment{
		OrderID:         initiatePayment.OrderID,
		RazorpayOrderID: RazorpayOrderID.(string),
	}
	if err := database.DB.Create(&PaymentDetails).Error; err != nil {
		PaymentFailedOrderTable(initiatePayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to add Payment order details",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	callbackurl := fmt.Sprintf("http://localhost:8080/api/v1/user/order/step3/paymentcallback/%v", initiatePayment.OrderID)

	responseData := map[string]interface{}{
		"razorpay_order_id": rzpOrder["id"],
		"amount":            rzpOrder["amount"],
		"key":               os.Getenv("RAZORPAY_KEY_ID"),
		"callbackurl":       callbackurl,
	}

	// Render the payment page
	c.HTML(http.StatusOK, "payment.html", responseData)
}

func PaymentGatewayCallback(c *gin.Context) {

	OrderID := c.Param("orderid")
	fmt.Println(OrderID)
	if OrderID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to get orderid",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var RazorpayPayment model.RazorpayPayment
	if err := c.ShouldBind(&RazorpayPayment); err != nil {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind Razorpay payment details",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	fmt.Println(RazorpayPayment)

	// Now you can proceed with your verification logic
	if !verifyRazorpaySignature(RazorpayPayment.OrderID, RazorpayPayment.PaymentID, RazorpayPayment.Signature, os.Getenv("RAZORPAY_KEY_SECRET")) {
		PaymentFailedOrderTable(OrderID)
		PaymentFailedPaymentTable(RazorpayPayment.OrderID)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to verify",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	PaymentDetails := model.Payment{
		RazorpayOrderID:   RazorpayPayment.OrderID,
		RazorpayPaymentID: RazorpayPayment.PaymentID,
		RazorpaySignature: RazorpayPayment.Signature,
		PaymentStatus:     model.PaymentComplete,
	}
	if err := database.DB.Where("order_id = ? AND razorpay_order_id = ?", OrderID, PaymentDetails.RazorpayOrderID).Updates(&PaymentDetails).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		PaymentFailedPaymentTable(RazorpayPayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update payment informations",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	var Order model.Order
	if err := database.DB.Model(&Order).Where("order_id = ?", OrderID).Update("payment_status", model.PaymentComplete).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		PaymentFailedPaymentTable(RazorpayPayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update payment informations",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//decrement stock based on orderid

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"paymentdata": RazorpayPayment,
		},
	})
}

func verifyRazorpaySignature(orderID, paymentID, signature, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(orderID + "|" + paymentID))
	computedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(computedSignature), []byte(signature))
}

func PaymentFailedOrderTable(OrderID string) bool {
	var Order model.Order
	Order.PaymentStatus = model.PaymentFailed
	if err := database.DB.Model(&model.Order{}).Where("order_id = ?", OrderID).Update("payment_status", model.PaymentFailed).Error; err != nil {
		return false
	}
	return true
}

func PaymentFailedPaymentTable(RazorpayOrderID string) bool {
	var PaymentDetails model.Payment
	PaymentDetails.PaymentStatus = model.PaymentFailed
	if err := database.DB.Model(&model.Payment{}).Where("razorpay_order_id = ?", RazorpayOrderID).Update("payment_status", model.PaymentFailed).Error; err != nil {
		return false
	}
	return true
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
		OrderTransition := []string{model.OrderStatusInPreparation, model.OrderStatusPrepared, model.OrderStatusDelivered}

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

func UserReviewonOrderItem(c *gin.Context) {
	//orderid, productid,review text
	//check delivered
	//if no, return
	//check review text
	//if yes, update the text to row
}

func UserRatingOrderItem(c *gin.Context) {

}

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
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&OrderItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch the order item",
		})
		return
	}

	//check order status is processing
	if OrderItem.OrderStatus != model.OrderStatusProcessing && OrderItem.OrderStatus != model.OrderStatusInPreparation {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "order can only be cancelled during processing or in preparation status",
		})
		return
	}

	//if mentioned with productid only cancel that
	if Request.ProductId != 0 {
		//change status to cancelled
		if err := database.DB.Model(&model.OrderItem{}).Where("order_id = ? AND product_id = ?", Request.OrderID, Request.ProductId).Update("order_status", model.OrderStatusCancelled).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "failed to cancel the order item",
			})
			return
		}
	} else {
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

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully cancelled the order",
		"data": gin.H{
			"order_id": Request.OrderID,
		},
	})
}

//after payment_confirmed change the stockleft based on orderitems quantity
