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
	PhoneNumber    string `gorm:"column:email;type:varchar(255);unique_index" validate:"number" json:"phone_number"`
	Picture        string `gorm:"column:picture;type:text" json:"picture"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
	Salt           string `gorm:"column:salt;type:varchar(255)" validate:"required" json:"salt"`
	LoginMethod    string `gorm:"column:login_method;type:varchar(255)" validate:"required" json:"login_method"`
	Blocked        bool   `gorm:"column:blocked;type:bool" json:"blocked"`
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
	ID           uint   `json:"product_id"`
	RestaurantID uint   `gorm:"foreignKey:RestaurantID" validate:"required" json:"restaurant_id"`
	CategoryID   uint   `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Name         string `validate:"required" json:"name"`
	Description  string `gorm:"column:description" validate:"required" json:"description"`
	ImageURL     string `gorm:"column:image_url" validate:"required" json:"image_url"`
	Price        uint   `validate:"required" json:"price"`
	Stock        uint   `validate:"required" json:"stock"`
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

type CartDetails struct {
	OrderID    uint `validate:"required,number" json:"order_id" gorm:"column:order_id"`
	ProductID  uint `validate:"required,number" json:"product_id" gorm:"column:product_id"`
	Quantity   uint `validate:"required,number" json:"quantity" gorm:"column:quantity"`
	CartAmount uint `validate:"required,number" json:"cart_amount" gorm:"column:cart_amount"`
}

type PaymentDetails struct {
	PaymentID     uint   `validate:"required,number" json:"payment_id" gorm:"column:payment_id"`
	OrderID       uint   `validate:"required,number" json:"order_id" gorm:"column:order_id"`
	PaymentAmount uint   `validate:"required,number" json:"payment_amount" gorm:"column:payment_amount"`
	TransactionID uint   `validate:"required,number" json:"transaction_id" gorm:"column:transaction_id"`
	PaymentStatus string `validate:"required" json:"payment_status" gorm:"column:payment_status"`
}

type Cart struct {
	CartID    uint       `gorm:"primaryKey;autoIncrement"`
	UserID    uint       `gorm:"not null" validate:"required,number" json:"user_id"`
	CartItems []CartItem `json:"cart_items"`
}

type CartItem struct {
	CartID    uint    `gorm:"not null"`
	UserID    uint    `validate:"required,number" json:"user_id"`
	ProductID uint    `gorm:"not null" validate:"required,number" json:"product_id"`
	Quantity  uint    `gorm:"not null" validate:"required,number" json:"quantity"`
	Price     float64 `gorm:"not null" validate:"required,number" json:"price"`
}

type Order struct {
	OrderID    uint        `gorm:"primaryKey;autoIncrement"`
	UserID     uint        `gorm:"not null" validate:"required,number" json:"user_id"`
	AddressID  uint        `gorm:"not null" validate:"required,number" json:"address_id"`
	TotalPrice float64     `gorm:"not null" validate:"required,number" json:"total_price"`
	Status     string      `gorm:"not null" validate:"required" json:"status"`
	CreatedAt  time.Time   `gorm:"autoCreateTime" json:"created_at"`
	OrderItems []OrderItem `json:"order_items"`
}

type OrderItem struct {
	OrderID   uint    `gorm:"not null"`
	ProductID uint    `gorm:"not null" validate:"required,number" json:"product_id"`
	Quantity  uint    `gorm:"not null" validate:"required,number" json:"quantity"`
	Price     float64 `gorm:"not null" validate:"required,number" json:"price"`
}

type OrderDetails struct {
	OrderID        uint   `validate:"required,number" json:"order_id" gorm:"column:order_id"`
	UserID         uint   `validate:"required,number" json:"user_id" gorm:"column:user_id"`
	RestaurantID   uint   `validate:"required,number" json:"restaurant_id" gorm:"column:restaurant_id"`
	OrderStatus    string `validate:"required" json:"order_status" gorm:"column:order_status"`
	TotalPrice     uint   `validate:"required,number" json:"total_price" gorm:"column:total_price"`
	PaymentMethod  string `validate:"required" json:"payment_method" gorm:"column:payment_method"`
	AddressID      uint   `validate:"required,number" json:"address_id" gorm:"column:address_id"`
	DeliveryStatus string `validate:"required" json:"delivery_status" gorm:"column:delivery_status"`
	CustomerReview string `validate:"required" json:"customer_review" gorm:"column:customer_review"`
	Rating         uint   `validate:"required,number" json:"rating" gorm:"column:rating"`
}

type OnlinePayment struct {
	PaymentID     uint    `gorm:"primaryKey;autoIncrement"`
	OrderID       uint    `gorm:"not null" validate:"required,number" json:"order_id"`
	PaymentAmount float64 `gorm:"not null" validate:"required,number" json:"payment_amount"`
	PaymentStatus string  `gorm:"not null" validate:"required" json:"payment_status"`
	TransactionID string  `gorm:"not null" validate:"required" json:"transaction_id"`
}
