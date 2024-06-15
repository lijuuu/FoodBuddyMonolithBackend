package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

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

	callbackurl := fmt.Sprintf("http://localhost:8080/api/v1/user/order/step3/razorpaycallback/%v", initiatePayment.OrderID)

	responseData := map[string]interface{}{
		"razorpay_order_id": rzpOrder["id"],
		"amount":            rzpOrder["amount"],
		"key":               os.Getenv("RAZORPAY_KEY_ID"),
		"callbackurl":       callbackurl,
	}

	// Render the payment page
	c.HTML(http.StatusOK, "payment.html", responseData)
}

func HandleStripe(c *gin.Context, initiatePayment model.InitiatePayment, order model.Order) {
	// Set your Stripe secret key
	stripe.Key = os.Getenv("STRIPE_KEY")

	// Calculate the total amount in the smallest currency unit (e.g., cents for USD, paise for INR)
	totalAmount := order.FinalAmount * 100 // Assuming FinalAmount is in the base currency unit (e.g., dollars for USD, rupees for INR)

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("inr"), // Assuming INR currency
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(order.OrderID),
						Description: stripe.String(order.PaymentMethod),
					},
					UnitAmount: stripe.Int64(int64(totalAmount)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String("http://localhost:8080/api/v1/user/order/step3/stripecallback?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("http://localhost:8080/api/v1/user/order/step3/stripecallback?session_id={CHECKOUT_SESSION_ID}"),
	}

	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Replace the placeholder with the actual session ID
	successURL := fmt.Sprintf("http://localhost:8080/api/v1/user/order/step3/stripecallback?session_id=%s", s.ID)
	cancelURL := fmt.Sprintf("http://localhost:8080/api/v1/user/order/step3/stripecallback?session_id=%s", s.ID)

	// Update the session with the correct URLs
	params.SuccessURL = stripe.String(successURL)
	params.CancelURL = stripe.String(cancelURL)

	// Return the URL to the client
	c.JSON(http.StatusSeeOther, gin.H{"url": s.URL})
}
func HandleWebhookStripe(c *gin.Context) {
	sessionID := c.Query("session_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id"})
		return
	}

	// Set your Stripe secret key
	stripe.Key = os.Getenv("STRIPE_KEY")

	// Fetch session details from Stripe
	stripeSession, err := session.Get(sessionID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve session details from Stripe",
			"error":   err.Error(),
		})
		return
	}

	// Extract the necessary details
	response := gin.H{
		"message": "Payment complete",
		"status":  "complete",
		"stripe": gin.H{
			"amount_subtotal":        stripeSession.AmountSubtotal,
			"amount_total":           stripeSession.AmountTotal,
			"currency":               stripeSession.Currency,
			"customer_email":         stripeSession.CustomerDetails.Email,
			"customer_name":          stripeSession.CustomerDetails.Name,
			"payment_method_types":   stripeSession.PaymentMethodTypes,
			"payment_status":         stripeSession.PaymentStatus,
			"success_url":            stripeSession.SuccessURL,
			"cancel_url":             stripeSession.CancelURL,
			"created":                stripeSession.Created,
			"expires_at":             stripeSession.ExpiresAt,
			"id":                     stripeSession.ID,
			"client_reference_id":    stripeSession.ClientReferenceID,
			"billing_address":        stripeSession.CustomerDetails.Address,
			"line_items":             stripeSession.LineItems,
			"customer_creation":      stripeSession.CustomerCreation,
			"livemode":               stripeSession.Livemode,
			"locale":                 stripeSession.Locale,
			"mode":                   stripeSession.Mode,
			"metadata":               stripeSession.Metadata,
			"object":                 stripeSession.Object,
			"ui_mode":                stripeSession.UIMode,
		},
	}

	c.JSON(http.StatusOK, response)
}