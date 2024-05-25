package model

import (
	"gorm.io/gorm"
)

const (
	EmailLoginMethod           = "email"
	GoogleSSOMethod            = "googlesso"
	VerificationStatusVerified = "verified"
	VerificationStatusPending  = "pending"
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

type User struct {
	gorm.Model
	ID                 uint   `validate:"required"`
	Name               string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email              string `gorm:"column:email;type:varchar(255);unique_index" validate:"email" json:"email"`
	Picture            string `gorm:"column:picture;type:text" json:"picture"`
	HashedPassword     string `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
	Salt               string `gorm:"column:salt;type:varchar(255)" validate:"required" json:"salt"`
	LoginMethod        string `gorm:"column:login_method;type:varchar(255)" validate:"required" json:"login_method"`
	VerificationStatus string `gorm:"column:verification_status;type:varchar(255)" validate:"required" json:"verification_status"`
	Blocked            bool   `gorm:"column:blocked;type:bool" json:"blocked"`
	OTP                int    `gorm:"column:otp;type:int" json:"otp"`
	OTPexpiry          int64  `gorm:"column:otp_expiry" json:"otp_expiry"`
}

type Admin struct {
	gorm.Model
	ID    uint   `gorm:"column:id;type:int;auto_increment;primary_key" json:"id"`
	Name  string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email string `gorm:"column:email;type:varchar(255)" validate:"email" json:"email"`
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

type LoginForm struct {
	Email    string `form:"email" validate:"required, email" json:"email"`
	Password string `form:"password" json:"password"`
}

type Category struct {
	gorm.Model
	ID          uint      `validate:"required"`
	Name        string    `validate:"required" json:"name"`
	Description string    `gorm:"column:description" validate:"required" json:"description"`
	ImageURL    string    `gorm:"column:image_url" validate:"required" json:"image_url"`
	Products    []Product `gorm:"foreignKey:CategoryID"`
}

type ImageSlice []string

type Product struct {
	gorm.Model
	ID           uint     `validate:"required"`
	RestaurantID uint     `gorm:"foreignKey:RestaurantID" validate:"required" json:"restaurant_id"`
	CategoryID   uint     `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Name         string   `validate:"required" json:"name"`
	Description  string   `gorm:"column:description" validate:"required" json:"description"`
	ImageURL     string `gorm:"column:image_url" validate:"required" json:"image_url"`
	Price        uint     `validate:"required" json:"price"`
	Stock        uint     `validate:"required" json:"stock"`
	//totalorders till now
	//avg rating
	//veg or non veg, validate this
}

type Restaurant struct {
	gorm.Model
	ID          uint   `validate:"required"`
	Name        string `validate:"required" json:"name"`
	Description string `gorm:"column:description" validate:"required" json:"description"`
	ImageURL    string `gorm:"column:image_url" validate:"required" json:"image_url"`
	Blocked     bool
}
