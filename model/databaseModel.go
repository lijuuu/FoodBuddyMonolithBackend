package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	EmailLoginMethod           = "email"
	GoogleSSOMethod            = "googlesso"
	VerificationStatusVerified = "verified"
	VerificationStatusPending  = "pending"
	UserRole                   = "user"
	AdminRole                  = "admin"
	RestaurantRole             = "restaurant"
	MaxUserQuantity            = 50
	YES                        = "YES"
	NO                         = "NO"

	CashOnDelivery = "COD"
	OnlinePayment  = "ONLINE"

	Razorpay = "RAZORPAY"
	Stripe   = "STRIPE"
	Wallet   = "WALLET"

	WalletIncoming = "INCOMING"
	WalletOutgoing = "OUTGOING"

	OnlinePaymentPending   = "ONLINE_PENDING"
	OnlinePaymentConfirmed = "ONLINE_CONFIRMED"
	OnlinePaymentFailed    = "ONLINE_FAILED"

	CODStatusPending   = "COD_PENDING"
	CODStatusConfirmed = "COD_CONFIRMED"
	CODStatusFailed    = "COD_FAILED"

	OrderStatusProcessing    = "PROCESSING"
	OrderStatusInPreparation = "PREPARATION"
	OrderStatusPrepared      = "PREPARED"
	OrderStatusOntheway      = "ONTHEWAY"
	OrderStatusDelivered     = "DELIVERED"
	OrderStatusCancelled     = "CANCELLED"


	CouponDiscountPercentageLimit = 50

	ReferralClaimAmount = 30
	ReferralClaimLimit  = 1
)

type EnvVariables struct {
	ClientID            string
	ClientSecret        string
	DBUser              string
	DBPassword          string
	DBName              string
	JWTSecret           string
	CloudinaryCloudName string
	CloudinaryAccessKey string
	CloudinarySecretKey string
}

type Admin struct {
	gorm.Model
	Email string `validate:"required,email"`
}

type User struct {
	gorm.Model
	ID             uint    `validate:"required"`
	Name           string  `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string  `gorm:"column:email;type:varchar(255);unique_index" validate:"email" json:"email"`
	PhoneNumber    string  `gorm:"column:phone_number;type:varchar(255);unique_index" validate:"number" json:"phone_number"`
	Picture        string  `gorm:"column:picture;type:text" json:"picture"`
	ReferralCode   string  `gorm:"column:referral_code" json:"referral_code"`
	WalletAmount   float64 `gorm:"column:wallet_amount;type:double" json:"wallet_amount"`
	LoginMethod    string  `gorm:"column:login_method;type:varchar(255)" validate:"required" json:"login_method"`
	Blocked        bool    `gorm:"column:blocked;type:bool" json:"blocked"`
	Salt           string  `gorm:"column:salt;type:varchar(255)" validate:"required" json:"salt"`
	HashedPassword string  `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
}

type UserReferralHistory struct {
	UserID       uint   `gorm:"column:user_id" json:"user_id"`
	ReferralCode string `gorm:"column:referral_code" json:"referral_code"`
	ReferredBy   string `gorm:"column:referred_by" json:"referred_by"`
	ReferClaimed bool   `gorm:"column:refer_claimed" json:"refer_claimed"`
}

type VerificationTable struct {
	Email              string `validate:"required,email" gorm:"type:varchar(255);unique_index"`
	Role               string
	OTP                uint64
	OTPExpiry          uint64
	VerificationStatus string `gorm:"type:varchar(255)"`
}

type Category struct {
	gorm.Model
	ID              uint
	Name            string    `validate:"required" json:"name"`
	Description     string    `gorm:"column:description" validate:"required" json:"description"`
	ImageURL        string    `gorm:"column:image_url" validate:"required" json:"image_url"`
	OfferPercentage uint      `gorm:"column:offer_percentage" json:"offer_percentage"`
	Products        []Product `gorm:"foreignKey:CategoryID"`
}

type Product struct {
	gorm.Model
	ID            uint
	RestaurantID  uint    `gorm:"foreignKey:RestaurantID" validate:"required" json:"restaurant_id"`
	CategoryID    uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Name          string  `validate:"required" json:"name"`
	Description   string  `gorm:"column:description" validate:"required" json:"description"`
	ImageURL      string  `gorm:"column:image_url" validate:"required" json:"image_url"`
	Price         float64 `validate:"required,number" json:"price"`
	MaxStock      uint    `validate:"required,number" json:"max_stock"`
	OfferAmount   float64 `gorm:"column:offer_amount" json:"offer_amount"`
	StockLeft     uint    `validate:"required,number" json:"stock_left"`
	AverageRating float64 `json:"average_rating"`
	Veg           string  `validate:"required" json:"veg" gorm:"column:veg"`
}

type Restaurant struct {
	gorm.Model
	ID                 uint
	Name               string `validate:"required" json:"name"`
	Description        string `gorm:"column:description" validate:"required" json:"description"`
	Address            string
	Email              string
	PhoneNumber        string
	WalletAmount       float64 `gorm:"column:wallet_amount;type:double" json:"wallet_amount"`
	ImageURL           string  `gorm:"column:image_url" validate:"required" json:"image_url"`
	CertificateURL     string  `gorm:"column:certificate_url" validate:"required" json:"certificate_url"`
	VerificationStatus string
	Blocked            bool
	Salt               string
	HashedPassword     string
}

type FavouriteProduct struct {
	UserID    uint `validate:"required"`
	ProductID uint `validate:"required"`
}

type Address struct {
	UserID       uint   `validate:"required,number" json:"user_id" gorm:"column:user_id"`
	AddressID    uint   `gorm:"primaryKey;autoIncrement;column:address_id" json:"address_id"`
	AddressType  string `validate:"required" json:"address_type" gorm:"column:address_type"`
	StreetName   string `validate:"required" json:"street_name" gorm:"column:street_name"`
	StreetNumber string `validate:"required" json:"street_number" gorm:"column:street_number"`
	City         string `validate:"required" json:"city" gorm:"column:city"`
	State        string `validate:"required" json:"state" gorm:"column:state"`
	PostalCode   string `validate:"required" json:"postal_code" gorm:"column:postal_code"`
}

type CartItems struct {
	UserID         uint   `gorm:"column:user_id" validate:"required,number" json:"user_id"`
	ProductID      uint   ` validate:"required,number" json:"product_id"`
	Quantity       uint   ` validate:"required,number" json:"quantity"`
	CookingRequest string // similar to zomato,, requesting restaurant to add or remove specific ingredients etc
}
type Order struct {
	OrderID              string    `validate:"required" json:"order_id"`
	UserID               uint      `validate:"required,number" json:"user_id"`
	AddressID            uint      `validate:"required,number" json:"address_id"`
	ItemCount            uint      `json:"item_count"`
	CouponCode           string    `json:"coupon_code"`
	CouponDiscountAmount float64   `validate:"required,number" json:"coupon_discount_amount"`
	ProductOfferAmount   float64   `json:"product_offer_amount"`
	TotalAmount          float64   `validate:"required,number" json:"total_amount"`
	FinalAmount          float64   `validate:"required,number" json:"final_amount"`
	PaymentMethod        string    `validate:"required" json:"payment_method" gorm:"column:payment_method"`
	PaymentStatus        string    `validate:"required" json:"payment_status" gorm:"column:payment_status"`
	OrderedAt            time.Time `gorm:"autoCreateTime" json:"ordered_at"`
}
type OrderItem struct {
	OrderID            string  `validate:"required"`
	UserID             uint    `validate:"required,number" json:"user_id"`
	RestaurantID       uint    `validate:"required,number" json:"restaurant_id"`
	ProductID          uint    ` validate:"required,number" json:"product_id"`
	Quantity           uint    ` validate:"required,number" json:"quantity"`
	Amount             float64 ` validate:"required,number" json:"amount"`
	ProductOfferAmount float64 `json:"product_offer_amount"`
	AfterDeduction     float64 `gorm:"column:after_deduction" json:"after_deduction"`
	CookingRequest     string
	OrderStatus        string `json:"order_status" gorm:"column:order_status"`
	OrderReview        string
	OrderRating        float64
}
type Payment struct {
	OrderID           string `validate:"required"`
	StripeSessionID   string `json:"stripe_session_id" column:"stripe_session_id"`
	StripePaymentID   string `json:"stripe_payment_id" column:"stripe_payment_id"`
	RazorpayOrderID   string `validate:"required" json:"razorpay_order_id" gorm:"column:razorpay_order_id"`
	RazorpayPaymentID string `validate:"required" json:"razorpay_payment_id" gorm:"column:razorpay_payment_id"`
	RazorpaySignature string `validate:"required" json:"razorpay_signature" gorm:"column:razorpay_signature"`
	PaymentGateway    string `json:"payment_gateway" gorm:"payment_gateway"`
	PaymentStatus     string `validate:"required" json:"payment_status" gorm:"column:payment_status"`
}

type PasswordReset struct {
	gorm.Model
	Email      string `validate:"email"`
	Role       string `validate:"required"`
	ResetToken string `gorm:"column:reset_token" json:"reset_token"`
	Active       string   `json:"active"`
	ExpiryTime uint   `gorm:"expiry_time" json:"expiry_time"`
}

type CouponInventory struct {
	CouponCode    string  `validate:"required" json:"coupon_code" gorm:"primary_key"`
	Expiry        uint    `validate:"required" json:"expiry"`
	Percentage    uint    `validate:"required" json:"percentage"`
	MaximumUsage  uint    `validate:"required" json:"maximum_usage"`
	MinimumAmount float64 `validate:"required" json:"minimum_amount"`
}

type CouponUsage struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	CouponCode string `json:"coupon_code"`
	UsageCount uint   `json:"usage_count"`
}

type UserWalletHistory struct {
	TransactionTime time.Time `gorm:"autoCreateTime" json:"transaction_time"`
	UserID          uint      `gorm:"column:user_id" json:"user_id"`
	Type            string    `gorm:"column:type" json:"type"` //incoming //outgoing
	OrderID         string    `gorm:"column:order_id" json:"order_id"`
	Amount          float64   `gorm:"column:amount" json:"amount"`
	CurrentBalance  float64   `gorm:"column:current_balance" json:"current_balance"`
	Reason          string    `gorm:"column:reason" json:"reason"`
}

type RestaurantWalletHistory struct {
	TransactionTime time.Time `gorm:"autoCreateTime"`
	Type            string    `gorm:"column:type" json:"type"` //incoming //outgoing
	OrderID         string    `gorm:"column:order_id" json:"order_id"`
	RestaurantID    uint      `gorm:"column:restaurant_id" json:"restaurant_id"`
	Amount          float64   `gorm:"column:amount" json:"amount"`
	CurrentBalance  float64   `gorm:"column:current_balance" json:"current_balance"`
	Reason          string    `gorm:"column:reason" json:"reason"`
}
