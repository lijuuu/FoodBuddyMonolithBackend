package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

func RoundDecimalValue(value float64) float64 {
	multiplier := math.Pow(10, 2)
	return math.Round(value*multiplier) / multiplier
}

func HandleRazorpay(c *gin.Context, initiatePayment model.InitiatePayment, order model.Order) {
	// Create Razorpay order
	fmt.Println(initiatePayment, order)
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))
	data := map[string]interface{}{
		"amount":          order.FinalAmount * 100, // Amount in paisa
		"currency":        "INR",
		"receipt":         order.OrderID,
		"payment_capture": 1,
	}

	fmt.Println(data)
	rzpOrder, err := client.Order.Create(data, nil)
	if err != nil {
		PaymentFailedOrderTable(initiatePayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//add to payment tables
	RazorpayOrderID := rzpOrder["id"]
	PaymentDetails := model.Payment{
		OrderID:         initiatePayment.OrderID,
		RazorpayOrderID: RazorpayOrderID.(string),
		PaymentGateway:  model.Razorpay,
		PaymentStatus:   model.OnlinePaymentPending,
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

	callbackurl := fmt.Sprintf("https://%v/api/v1/user/order/step3/razorpaycallback/%v", utils.GetEnvVariables().ServerIP, initiatePayment.OrderID)
	cancelurl := fmt.Sprintf("https://%v/api/v1/user/order/step3/razorpaycallback/failed/%v", utils.GetEnvVariables().ServerIP, initiatePayment.OrderID)

	responseData := map[string]interface{}{
		"razorpay_order_id": rzpOrder["id"],
		"amount":            rzpOrder["amount"],
		"key":               os.Getenv("RAZORPAY_KEY_ID"),
		"callbackurl":       callbackurl,
		"cancelurl":         cancelurl,
	}

	// Render the payment page
	c.HTML(http.StatusOK, "payment.html", responseData)
}
func RazorPayGatewayCallback(c *gin.Context) {

	OrderID := c.Param("orderid")
	fmt.Println(OrderID)
	if OrderID == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to get orderid",
		})
		return
	}

	var RazorpayPayment model.RazorpayPayment
	if err := c.ShouldBind(&RazorpayPayment); err != nil {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind Razorpay payment details" + err.Error(),
		})
		return
	}

	fmt.Println(RazorpayPayment)

	// Now you can proceed with your verification logic
	if !verifyRazorpaySignature(RazorpayPayment.OrderID, RazorpayPayment.PaymentID, RazorpayPayment.Signature, os.Getenv("RAZORPAY_KEY_SECRET")) {
		PaymentFailedOrderTable(OrderID)
		PaymentFailedPaymentTable(RazorpayPayment.OrderID)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to verify",
		})
		return
	}

	PaymentDetails := model.Payment{
		RazorpayOrderID:   RazorpayPayment.OrderID,
		RazorpayPaymentID: RazorpayPayment.PaymentID,
		RazorpaySignature: RazorpayPayment.Signature,
		PaymentStatus:     model.OnlinePaymentConfirmed,
	}
	if err := database.DB.Where("order_id = ? AND razorpay_order_id = ?", OrderID, PaymentDetails.RazorpayOrderID).Updates(&PaymentDetails).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		PaymentFailedPaymentTable(RazorpayPayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update payment informations",
		})
		return
	}

	var Order model.Order
	if err := database.DB.Model(&Order).Where("order_id = ?", OrderID).Update("payment_status", model.OnlinePaymentConfirmed).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		PaymentFailedPaymentTable(RazorpayPayment.OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update payment informations",
		})
		return
	}

	//update all the orderitems as intiated
	if err := database.DB.Model(&model.OrderItem{}).Where("order_id = ?", OrderID).Update("order_status", model.OrderStatusInitiated).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update order status",
		})
		return
	}

	//decrement stock based on orderid
	done := DecrementStock(OrderID)
	if !done {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to decrement order stock",
		})
		return
	}

	//update payment for each restaurant by splitting payment for each restaurant
	done = SplitMoneyToRestaurants(OrderID)
	if !done {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to split payment for restaurant",
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

func RazorPayFailed(c *gin.Context) {
	OrderID := c.Param("orderid")
	fmt.Println(OrderID)
	if OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "failed to get orderid"})
		return
	}

	PaymentFailedOrderTable(OrderID)

}

func HandleStripe(c *gin.Context, initiatePayment model.InitiatePayment, order model.Order) {
	stripe.Key = os.Getenv("STRIPE_KEY")

	totalAmount := order.FinalAmount * 100 // Convert to paise, same as Razorpay

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("inr"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(order.OrderID),
						Description: stripe.String(order.PaymentMethod),
					},
					UnitAmount: stripe.Int64(int64(totalAmount)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata:   map[string]string{"order_id": order.OrderID},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(fmt.Sprintf("http://%v/api/v1/user/order/step3/stripecallback?session_id={CHECKOUT_SESSION_ID}", utils.GetEnvVariables().ServerIP)),
		CancelURL:  stripe.String(fmt.Sprintf("http://%v/api/v1/user/order/step3/stripecallback?session_id={CHECKOUT_SESSION_ID}", utils.GetEnvVariables().ServerIP)),
	}

	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	StripePaymentDetail := model.Payment{
		OrderID:         order.OrderID,
		PaymentGateway:  model.Stripe,
		StripeSessionID: s.ID,
	}

	if err := database.DB.Create(&StripePaymentDetail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to store stripe payment information",
			"error":   err.Error(),
		})
		return
	}

	// Return the URL to the client
	c.JSON(http.StatusSeeOther, gin.H{"url": s.URL})
}

func StripeCallback(c *gin.Context) {
	sessionID := c.Query("session_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id"})
		return
	}

	// Set your Stripe secret key
	stripe.Key = os.Getenv("STRIPE_KEY")

	//using session id get the stripe session info, payment information and its id
	stripeSession, err := session.Get(sessionID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve session details from Stripe",
			"error":   err.Error(),
		})
		return
	}

	OrderID := stripeSession.Metadata["order_id"]
	StripePayment := model.Payment{
		OrderID:         OrderID,
		StripePaymentID: stripeSession.PaymentIntent.ID,
	}

	if stripeSession.PaymentStatus == "paid" {
		StripePayment.PaymentStatus = model.OnlinePaymentConfirmed
		if err := database.DB.Where("stripe_session_id = ?", stripeSession.ID).Updates(&StripePayment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to update status of payment",
				"error":   err.Error(),
			})
			return
		}

		//update payment status on order as well
		if err := database.DB.Model(&model.Order{}).Where("order_id = ?", OrderID).Update("payment_status", model.OnlinePaymentConfirmed).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to update status of payment",
				"error":   err.Error(),
			})
			return
		}
		//update all the orderitems as intiated
		if err := database.DB.Model(&model.OrderItem{}).Where("order_id = ?", OrderID).Update("order_status", model.OrderStatusInitiated).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "failed to update order status",
			})
			return
		}
		//decrement stock based on orderid
		done := DecrementStock(OrderID)
		if !done {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "failed to decrement order stock",
			})
			return
		}

		//update payment for each restaurant by splitting payment for each restaurant
		done = SplitMoneyToRestaurants(OrderID)
		if !done {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "failed to split payment for restaurant",
			})
			return
		}
	} else {
		StripePayment.PaymentStatus = model.OnlinePaymentFailed
		if err := database.DB.Where("stripe_payment_id = ?", stripeSession.PaymentIntent.ID).Updates(&StripePayment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to update status of payment",
				"error":   err.Error(),
			})
			return
		}
		if err := database.DB.Model(&model.Payment{}).Where("order_id = ?", OrderID).Update("payment_status", model.OnlinePaymentFailed).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to update status of payment",
				"error":   err.Error(),
			})
			return
		}

	}

	response := gin.H{
		"message": "Payment complete",
		"status":  "complete",
		"stripe": gin.H{
			"payment_id":      stripeSession.PaymentIntent.ID,
			"amount_subtotal": stripeSession.AmountSubtotal / 100,
			"amount_total":    stripeSession.AmountTotal / 100,
			"payment_mode":    stripeSession.PaymentMethodTypes,
			"currency":        stripeSession.Currency,
			"payment_status":  stripeSession.PaymentStatus,
			"created":         stripeSession.Created,
			"expires_at":      stripeSession.ExpiresAt,
			"id":              stripeSession.ID,
			"mode":            stripeSession.Mode,
			"order_id":        stripeSession.Metadata["order_id"],
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"paymentdata": response,
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
	if err := database.DB.Model(&model.Order{}).Where("order_id = ?", OrderID).Update("payment_status", model.OnlinePaymentFailed).Error; err != nil {
		return false
	}
	return true
}

func PaymentFailedPaymentTable(RazorpayOrderID string) bool {
	if err := database.DB.Model(&model.Payment{}).Where("razorpay_order_id = ?", RazorpayOrderID).Update("payment_status", model.OnlinePaymentFailed).Error; err != nil {
		return false
	}
	return true
}

func GetUserWalletData(c *gin.Context) {
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

	var User model.User
	if err := database.DB.Where("id = ?", UserID).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": false, "message": "failed to get wallet balance",
		})
		return
	}

	var Result []model.UserWalletHistory
	if err := database.DB.Where("user_id = ?", UserID).Find(&Result).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": false, "message": "failed to get wallet history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true, "data": gin.H{
			"walletbalance": User.WalletAmount,
			"history":       Result,
		},
	})

}

func HandleWalletPayment(OrderID string, UserID uint, c *gin.Context) {
	// Verify if user has sufficient wallet balance
	var user model.User
	if err := database.DB.Where("id = ?", UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch user details",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	var order model.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&order).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to fetch order information",
		})
		return
	}

	if float64(user.WalletAmount) < order.FinalAmount {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusPaymentRequired, gin.H{
			"status":  false,
			"message": "insufficient wallet balance",
		})
		return
	}

	// Deduct amount from user wallet
	newBalance := float64(user.WalletAmount) - order.FinalAmount
	if err := database.DB.Model(&user).Update("wallet_amount", newBalance).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to deduct wallet balance",
		})
		return
	}

	// Record the wallet transaction
	walletHistory := model.UserWalletHistory{
		TransactionTime: time.Now(),
		WalletPaymentID: uuid.New().String(),
		UserID:          UserID,
		Type:            model.WalletOutgoing,
		Amount:          order.FinalAmount,
		CurrentBalance:  newBalance,
		OrderID:         OrderID,
		Reason:          model.WalletTxTypeOrderPayment,
	}

	if err := database.DB.Create(&walletHistory).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to record wallet transaction",
		})
		return
	}

	// Update payment status
	PaymentDetails := model.Payment{
		OrderID:        OrderID,
		PaymentStatus:  model.OnlinePaymentConfirmed,
		PaymentGateway: model.Wallet,
	}
	if err := database.DB.Model(&model.Payment{}).Where("order_id = ?", OrderID).Updates(&PaymentDetails).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update payment information",
		})
		return
	}

	if err := database.DB.Model(&order).Where("order_id = ?", OrderID).Update("payment_status", model.OnlinePaymentConfirmed).Error; err != nil {
		PaymentFailedOrderTable(OrderID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update payment status",
		})
		return
	}

	// Decrement stock based on order ID
	if !DecrementStock(OrderID) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to decrement order stock",
		})
		return
	}

	// Split payment for each restaurant
	if !SplitMoneyToRestaurants(OrderID) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to split payment for restaurants",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"payment": OrderID + " Status : Payment Confirmed, Payment Method :" + model.Wallet,
		},
	})
}

func CreateRestaurantWalletHistory(r model.RestaurantWalletHistory) bool {
	if err := database.DB.Create(&r).Error; err != nil {
		return false
	}
	return true
}

func VerifyOnlinePayment(c *gin.Context) {
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)
	OrderID := c.Query("order_id")
	if OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "please provide the orderid in the query"})
		return
	}

	oUserID, _ := UserIDfromOrderID(OrderID)
	if UserID != oUserID {
		c.JSON(http.StatusUnauthorized, gin.H{"status": false, "message": "please provide orderid from the orders you have initiated"})
		return
	}

	var PaymentInfo model.Payment
	if err := database.DB.Where("order_id = ? AND payment_status = ?", OrderID, model.OnlinePaymentConfirmed).First(&PaymentInfo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "online payment is not done"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "online payment done",
		"data":    PaymentInfo,
	})
}

func ChangeOrderPaymentMode(c *gin.Context) { //check if payment confirmed, change the order items payment status to cod pending
	//check user api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": false, "message": "unauthorized request"})
		return
	}
	JWTUserID, _ := UserIDfromEmail(email)

	var Request model.ChangeOrderPaymentMode
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "make sure the order_id and the payment_mode exist on the json payload"})
		return
	}
	//check if the order is of the user who is sending the request
	oUserID, _ := UserIDfromOrderID(Request.OrderID)
	if JWTUserID != oUserID {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "unauthorized request"})
		return
	}

	//check if the request is having COD or ONLINE
	if Request.PaymentMethod != model.CashOnDelivery && Request.PaymentMethod != model.OnlinePayment {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "payment method should be either COD or ONLINE"})
		return
	}

	//retreive the order
	var Order model.Order
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&Order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "order_id doesn't exist or failed to retrieve the order_id, try again"})
		return
	}

	//check
	if Request.PaymentMethod == model.CashOnDelivery && Order.TotalAmount >= model.CODMaximumAmount {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "switch to ONLINE payment for total amounts greater than or equal to 1000"})
		return
	}

	//check if its already the same requested payment method
	if Request.PaymentMethod == Order.PaymentMethod {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "payment method is already : " + Request.PaymentMethod})
		return
	}

	if Order.PaymentStatus == model.OnlinePaymentConfirmed || Order.PaymentStatus == model.CODStatusConfirmed {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "payment is already done, cannot update the payment status"})
		return
	}

	//update the payment method
	Order.PaymentMethod = Request.PaymentMethod
	if Request.PaymentMethod == model.CashOnDelivery { //COD
		Order.PaymentStatus = model.CODStatusPending
		if err := database.DB.Where("order_id = ?", Request.OrderID).Updates(&Order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to change payment method"})
			return
		}
		c.JSON(http.StatusOK,
			gin.H{"status": false, "message": "payment method changed to COD and the order is updated to COD_PENDING", "order_details": Order})
		return
	}

	//ONLINE
	Order.PaymentStatus = model.OnlinePaymentPending
	if err := database.DB.Where("order_id = ?", Request.OrderID).Updates(&Order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to change payment method"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": false, "message": "payment method changed to " + Request.PaymentMethod})
}
