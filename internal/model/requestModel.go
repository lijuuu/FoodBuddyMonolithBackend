package model

type EmailSignupRequest struct {
	Name            string `validate:"required" json:"name"`
	Email           string `validate:"required,email" json:"email"`
	PhoneNumber     string `validate:"required" json:"phone_number"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}

type EmailLoginRequest struct {
	Email    string `form:"email" validate:"required,email" json:"email"`
	Password string `form:"password" validate:"required" json:"password"`
}

type UpdateUserInformation struct {
	Name        string `json:"name"`
	PhoneNumber string `gorm:"column:phone_number" validate:"required" json:"phone_number"`
	Picture     string `json:"picture"`
}

type RestaurantSignupRequest struct {
	Name            string `validate:"required" json:"name"`
	Description     string `gorm:"column:description" validate:"required" json:"description"`
	Address         string `gorm:"column:address" validate:"required" json:"address"`
	Email           string `gorm:"column:email" validate:"required,email" json:"email"`
	ConfirmPassword string `gorm:"column:confirmpassword" validate:"required" json:"confirmpassword"`
	Password        string `gorm:"column:password" validate:"required" json:"password"`
	PhoneNumber     string `gorm:"column:phone_number" validate:"required" json:"phone_number"`
	ImageURL        string `gorm:"column:image_url" validate:"required,url" json:"image_url"`
	CertificateURL  string `gorm:"column:certificate_url" validate:"required,url" json:"certificate_url"`
}

type UpdateQuantityCart struct {
	ProductID uint `validate:"required" json:"product_id"`
	Quantity  uint `validate:"required" json:"quantity"`
}

type RestaurantLoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type AdminLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type AddToCartReq struct {
	ProductID      uint   `gorm:"column:product_id" validate:"required,number" json:"product_id"`
	Quantity       uint   `validate:"required,number" json:"quantity"`
	CookingRequest string `json:"cooking_request"` // similar to zomato,, requesting restaurant to add or remove specific ingredients etc
}

type RemoveItem struct {
	ProductID uint `validate:"required,number" json:"product_id"`
}

type PlaceOrder struct {
	UserID        uint   `json:"user_id"`
	AddressID     uint   `validate:"required,number" json:"address_id"`
	RestaurantID  uint   `validate:"required,number" gorm:"column:restaurant_id" json:"restaurant_id"`
	PaymentMethod string `validate:"required" json:"payment_method"`
	CouponCode    string `json:"coupon_code"`
}

type InitiatePayment struct {
	OrderID        string `json:"order_id"`
	PaymentGateway string `json:"payment_gateway"`
}

type RazorpayPayment struct {
	PaymentID string `form:"razorpay_payment_id" binding:"required" json:"razorpay_payment_id"`
	OrderID   string `form:"razorpay_order_id" binding:"required" json:"razorpay_order_id"`
	Signature string `form:"razorpay_signature" binding:"required" json:"razorpay_signature"`
}

type OrderHistoryRestaurants struct {
	OrderStatus string `json:"order_status"`
}

type UserOrderHistory struct {
	UserID  uint   `json:"user_id"`
	OrderID string `json:"order_id"`
}

type PaymentDetailsByOrderID struct {
	OrderID       string `json:"order_id"`
	PaymentStatus string `json:"payment_status"`
}

type UpdateOrderStatusForRestaurant struct {
	OrderID   string `json:"order_id"`
	ProductID uint   `json:"product_id"`
}

type CancelOrderedProduct struct {
	OrderID   string `json:"order_id"`
	ProductId uint   `json:"product_id"`
}

type IncrementStock struct {
	OrderID string `json:"order_id"`
}

type Step1PasswordReset struct {
	Email string `validate:"required,email"`
	Role  string `validate:"required"`
}

type Step2PasswordReset struct {
	Email           string `form:"email" binding:"required,email" json:"email"`
	Role            string `form:"role" binding:"required" `
	Token           string `form:"token" binding:"required" json:"token"`
	Password        string `form:"password1" binding:"required" json:"password1"`
	ConfirmPassword string `form:"password2" binding:"required" json:"password2"`
}

type UserReviewonOrderItem struct {
	OrderID    string `validate:"required" json:"order_id"`
	ProductID  uint   `validate:"required" json:"product_id"`
	ReviewText string `validate:"required" json:"user_review"`
}

type UserRatingOrderItem struct {
	OrderID    string  `validate:"required" json:"order_id"`
	ProductID  uint    `validate:"required" json:"product_id"`
	UserRating float64 `validate:"required" json:"user_rating"`
}

type CouponInventoryRequest struct {
	CouponCode    string `validate:"required" json:"coupon_code"`
	Expiry        uint   `validate:"required" json:"expiry"`
	Percentage    uint   `validate:"required" json:"percentage"`
	MaximumUsage  uint   `validate:"required" json:"maximum_usage"`
	MinimumAmount uint   `validate:"required" json:"minimum_amount"`
}

type ApplyCouponOnOrderRequest struct {
	UserID     uint   `validate:"required" json:"user_id"`
	CouponCode string `validate:"required" json:"coupon_code"`
	OrderID    string `validate:"required" json:"order_id"`
}

// changes
type AddCategoryRequest struct {
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`
	ImageURL    string `validate:"required,url" json:"image_url"`
}

type EditCategoryRequest struct {
	ID          uint   `validate:"required,number" json:"id"`
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`
	ImageURL    string `validate:"required,url" json:"image_url"`
}

type EditRestaurantRequest struct {
	Name           string `validate:"omitempty" json:"name"`
	Description    string `validate:"omitempty" json:"description"`
	Address        string `validate:"omitempty" json:"address"`
	PhoneNumber    string `validate:"omitempty,number" json:"phone_number"`
	ImageURL       string `validate:"omitempty,url" json:"image_url"`
	CertificateURL string `validate:"omitempty,url" json:"certificate_url"`
}

type AddProductRequest struct {
	CategoryID  uint    `validate:"required,number" json:"category_id"`
	Name        string  `validate:"required" json:"name"`
	Description string  `validate:"required" json:"description"`
	ImageURL    string  `validate:"required" json:"image_url"`
	Price       float64 `validate:"required,number" json:"price"`
	OfferAmount float64 `json:"offer_amount"`
	MaxStock    uint    `validate:"required,number" json:"max_stock"`
	StockLeft   uint    `validate:"required,number" json:"stock_left"`
	Veg         string  `validate:"required" json:"veg"`
}

type EditProductRequest struct {
	ProductID   uint    `validate:"required,number" json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	OfferAmount float64 `json:"offer_amount"`
	MaxStock    uint    `json:"max_stock"`
	StockLeft   uint    `json:"stock_left"`
	Veg         string  `json:"veg"`
}

type AddOfferRequest struct {
	ProductID   uint    `json:"product_id" binding:"required"`
	OfferAmount float64 `json:"offer_amount" binding:"required"`
}

type RestaurantProfileUpdate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	ImageURL    string `json:"image_url"`
}

type EditUserAddress struct {
	UserID       uint   `json:"user_id" gorm:"column:user_id"`
	AddressID    uint   `gorm:"column:address_id" json:"address_id"`
	PhoneNumber  string `validate:"required" json:"phone_number"`
	AddressType  string `validate:"required" json:"address_type" gorm:"column:address_type"`
	StreetName   string `validate:"required" json:"street_name" gorm:"column:street_name"`
	StreetNumber string `validate:"required" json:"street_number" gorm:"column:street_number"`
	City         string `validate:"required" json:"city" gorm:"column:city"`
	State        string `validate:"required" json:"state" gorm:"column:state"`
	PostalCode   string `validate:"required" json:"postal_code" gorm:"column:postal_code"`
}

type AddCookingRequest struct {
	ProductID      uint   `json:"product_id"`
	CookingRequest string `json:"cooking_request"`
}

type ChangeOrderPaymentMode struct {
	OrderID       string `json:"order_id"`
	PaymentMethod string `json:"payment_method"`
}

type ConfirmCODPayment struct {
	OrderID string `json:"order_id"`
}

type ConfirmDelivery struct {
	OrderID     string `json:"order_id"`
	DeliveryOTP uint   `json:"delivery_otp"`
}
