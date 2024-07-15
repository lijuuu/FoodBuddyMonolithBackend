package model

import "time"

type GoogleResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type SuccessResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  bool   `json:"status" example:"false"`
	Message string `json:"message" example:"Error message"`
}

type ProductSales struct {
	TotalAmount uint    `json:"total_amount"`
	TotalOrders uint    `json:"total_orders"`
	AvgRating   float64 `json:"avg_rating"` // Pointer to allow for NULL values
	Quantity    uint    `json:"quantity"`
}

type BestProduct struct {
	ProductID    uint    `json:"product_id"`
	Name         string  `json:"name"`
	CategoryName string  `json:"category_name"`
	Description  string  `json:"description"`
	ImageURL     string  `json:"image_url"`
	Price        float64 `json:"price"`
	Rating       float64 `json:"rating"`
	TotalSales   uint    `json:"TotalSales"`
}

type ProductResponse struct {
	ID             uint    `json:"product_id"`
	RestaurantName string  `json:"restaurant_name"`
	CategoryName   string  `json:"category_name"`
	Name           string  `json:"product_name"`
	Description    string  `json:"description"`
	ImageURL       string  `json:"image_url"`
	Price          float64 `json:"price"`
	StockLeft      uint    `json:"stock_left"`
	AverageRating  float64 `json:"average_rating"`
	Veg            string  `json:"veg"`
}

type OrderCount struct {
	TotalOrder         uint `json:"total_order"`
	TotalProcessing    uint `json:"total_processing"`
	TotalInitiated     uint `json:"total_initiated"`
	TotalInPreparation uint `json:"total_in_preparation"`
	TotalPrepared      uint `json:"total_prepared"`
	TotalOnTheWay      uint `json:"total_onthway"`
	TotalDelivered     uint `json:"total_delivered"`
	TotalCancelled     uint `json:"total_cancelled"`
}
type AmountInformation struct {
	TotalCouponDeduction       float64 `json:"total_coupon_deduction"`
	TotalProductOfferDeduction float64 `json:"total_product_offer_deduction"`
	TotalAmountBeforeDeduction float64 `json:"total_amount_before_deduction"`
	TotalAmountAfterDeduction  float64 `json:"total_amount_after_deduction"`
}

type OrderSales struct {
	TotalRevenue            float64 `json:"total_revenue"`
	CouponDiscounts         float64 `json:"coupon_discounts"`
	ProductOffers           float64 `json:"product_offers"`
	TotalCancelOrderRefunds float64 `json:"total_cancelorder_refunds"`
}

type OverallOrderReport struct {
	From  time.Time  `json:"from"`
	Till  time.Time  `json:"till"`
	Count OrderCount `json:"count"`
}

type PlatformSalesReportInput struct {
	StartDate     string `json:"start_date,omitempty" time_format:"2006-01-02"`
	EndDate       string `json:"end_date,omitempty" time_format:"2006-01-02"`
	Limit         string `json:"limit,omitempty"`
	PaymentStatus string `json:"payment_status"`
}

type RestaurantOverallSalesReport struct {
	StartDate     string `json:"start_date,omitempty" time_format:"2006-01-02"`
	EndDate       string `json:"end_date,omitempty" time_format:"2006-01-02"`
	Limit         string `json:"limit,omitempty"`
	PaymentStatus string `json:"payment_status"`
}


type BlockedUserResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	PhoneNumber  uint    `json:"phone_number"`
	Picture      string  `json:"picture"`
	ReferralCode string  `json:"referral_code"`
	WalletAmount float64 `json:"wallet_amount"`
	LoginMethod  string  `json:"login_method"`
	Blocked      bool    `json:"blocked"`
}
