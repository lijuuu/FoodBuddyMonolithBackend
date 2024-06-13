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

	CashOnDelivery = "COD"
	OnlinePayment  = "ONLINE"

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
	ID             uint   `validate:"required"`
	Name           string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string `gorm:"column:email;type:varchar(255);unique_index" validate:"email" json:"email"`
	PhoneNumber    string `gorm:"column:phone_number;type:varchar(255);unique_index" validate:"number" json:"phone_number"`
	Picture        string `gorm:"column:picture;type:text" json:"picture"`
	WalletAmount   uint   `gorm:"column:wallet_amount;" json:"wallet_amount"`
	LoginMethod    string `gorm:"column:login_method;type:varchar(255)" validate:"required" json:"login_method"`
	Blocked        bool   `gorm:"column:blocked;type:bool" json:"blocked"`
	Salt           string `gorm:"column:salt;type:varchar(255)" validate:"required" json:"salt"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
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
	ID          uint
	Name        string    `validate:"required" json:"name"`
	Description string    `gorm:"column:description" validate:"required" json:"description"`
	ImageURL    string    `gorm:"column:image_url" validate:"required" json:"image_url"`
	Products    []Product `gorm:"foreignKey:CategoryID"`
}

type Product struct {
	gorm.Model
	ID            uint
	RestaurantID  uint    `gorm:"foreignKey:RestaurantID" validate:"required" json:"restaurant_id"`
	CategoryID    uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Name          string  `validate:"required" json:"name"`
	Description   string  `gorm:"column:description" validate:"required" json:"description"`
	ImageURL      string  `gorm:"column:image_url" validate:"required" json:"image_url"`
	Price         uint    `validate:"required,number" json:"price"`
	MaxStock      uint    `validate:"required,number" json:"max_stock"`
	StockLeft     uint    `validate:"required,number" json:"stock_left"`
	AverageRating float64 `json:"average_rating"`
	//totalorders till now
	//avg rating
	//veg or non veg, validate this
}

type Restaurant struct {
	gorm.Model
	ID                 uint
	Name               string `validate:"required" json:"name"`
	Description        string `gorm:"column:description" validate:"required" json:"description"`
	Address            string
	Email              string
	PhoneNumber        string
	WalletAmount       int    `gorm:"column:wallet_amount;" json:"wallet_amount"`
	ImageURL           string `gorm:"column:image_url" validate:"required" json:"image_url"`
	CertificateURL     string `gorm:"column:certificate_url" validate:"required" json:"certificate_url"`
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
	OrderID        string    `validate:"required" json:"order_id"`
	UserID         uint      `validate:"required,number" json:"user_id"`
	AddressID      uint      `validate:"required,number" json:"address_id"`
	DiscountAmount float64   `validate:"required,number" json:"discount_amount"`
	CouponCode     string    `json:"coupon_code"`
	TotalAmount    float64   `validate:"required,number" json:"total_amount"`
	FinalAmount    float64   `validate:"required,number" json:"final_amount"`
	PaymentMethod  string    `validate:"required" json:"payment_method" gorm:"column:payment_method"`
	PaymentStatus  string    `validate:"required" json:"payment_status" gorm:"column:payment_status"`
	OrderedAt      time.Time `gorm:"autoCreateTime" json:"ordered_at"`
}

type Payment struct {
	OrderID           string `validate:"required"`
	RazorpayOrderID   string `validate:"required" json:"razorpay_order_id" gorm:"column:razorpay_order_id"`
	RazorpayPaymentID string `validate:"required" json:"razorpay_payment_id" gorm:"column:razorpay_payment_id"`
	RazorpaySignature string `validate:"required" json:"razorpay_signature" gorm:"column:razorpay_signature"`
	PaymentStatus     string `validate:"required" json:"payment_status" gorm:"column:payment_status"`
}

type OrderItem struct {
	OrderID        string `validate:"required"`
	RestaurantID   uint   `validate:"required,number" json:"restaurant_id"`
	ProductID      uint   ` validate:"required,number" json:"product_id"`
	Quantity       uint   ` validate:"required,number" json:"quantity"`
	Amount         uint   ` validate:"required,number" json:"amount"`
	CookingRequest string
	OrderStatus    string `json:"order_status" gorm:"column:order_status"`
	OrderReview    string
	OrderRating    float64
}

type UserPasswordReset struct {
	gorm.Model
	Email      string `validate:"email"`
	ResetToken string `gorm:"column:reset_token" json:"reset_token"`
	ExpiryTime uint   `gorm:"expiry_time" json:"expiry_time"`
}

type CouponInventory struct {
	CouponCode   string `validate:"required" json:"coupon_code" gorm:"primary_key"`
	Expiry       uint   `validate:"required" json:"expiry"`
	Percentage   uint   `validate:"required" json:"percentage"`
	MaximumUsage uint   `validate:"required" json:"maximum_usage"`
}

type CouponUsage struct {
	gorm.Model
	OrderID    uint   `json:"order_id"`
	UserID     uint   `json:"user_id"`
	CouponCode string `json:"coupon_code"`
	UsageCount uint   `json:"usage_count"`
}
