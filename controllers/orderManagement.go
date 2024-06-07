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
			OrderStatus: model.OrderStatusProcessing,
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
	ok:= CheckUser(PlaceOrder.UserID)
	if !ok{
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
	if err := database.DB.Where("order_id = ? AND razorpay_order_id = ?", OrderID,PaymentDetails.RazorpayOrderID).Updates(&PaymentDetails).Error; err != nil {
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

func PaymentFailedPaymentTable(RazorpayOrderID string)bool {
	var PaymentDetails model.Payment
	PaymentDetails.PaymentStatus = model.PaymentFailed
	if err := database.DB.Model(&model.Payment{}).Where("razorpay_order_id = ?", RazorpayOrderID).Update("payment_status", model.PaymentFailed).Error; err != nil {
		return false
	}
	return true
}

func CheckUser(UserID uint)bool  {

	var User model.User
	if err:= database.DB.Where("id = ?",UserID).First(&User).Error;err!=nil{
		return false
	}
	return true
}

//active orders odf 