package model

import (
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
	ID                 uint   `validate:"required"`
	Name               string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email              string `gorm:"column:email;type:varchar(255);unique_index" validate:"email" json:"email"`
	PhoneNumber        string `gorm:"column:email;type:varchar(255);unique_index" validate:"number" json:"phone_number"`
	Picture            string `gorm:"column:picture;type:text" json:"picture"`
	HashedPassword     string `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
	Salt               string `gorm:"column:salt;type:varchar(255)" validate:"required" json:"salt"`
	LoginMethod        string `gorm:"column:login_method;type:varchar(255)" validate:"required" json:"login_method"`
	VerificationStatus string `gorm:"column:verification_status;type:varchar(255)" validate:"required" json:"verification_status"`
	Blocked            bool   `gorm:"column:blocked;type:bool" json:"blocked"`
	OTP                int    `gorm:"column:otp;type:int" json:"otp"`
	OTPexpiry          int64  `gorm:"column:otp_expiry" json:"otp_expiry"`
}

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

type OTPTable struct {
	email              string `validate:"required,email"`
	OTP                int
	OTPexpiry          int64
	VerificationStatus string `gorm:"column:verification_status;type:varchar(255)" validate:"required" json:"verification_status"`
}

type LoginForm struct {
	Email    string `form:"email" validate:"required,email" json:"email"`
	Password string `form:"password" validate:"required" json:"password"`
}

type Category struct {
	gorm.Model
	ID          uint
	Name        string    `validate:"required" json:"name"`
	Description string    `gorm:"column:description" validate:"required" json:"description"`
	ImageURL    string    `gorm:"column:image_url" validate:"required" json:"image_url"`
	Products    []Product `gorm:"foreignKey:CategoryID"`
}

type ImageSlice []string

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
	ID          uint
	Name        string `validate:"required" json:"name"`
	Description string `gorm:"column:description" validate:"required" json:"description"`
	ImageURL    string `gorm:"column:image_url" validate:"required" json:"image_url"`
	Blocked     bool
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
