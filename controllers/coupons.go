package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CouponInventory struct {
	CouponCode   string `json:"coupon_code" gorm:"primary_key"`
	Expiry       uint   `json:"expiry"`
	Percentage   uint   `json:"percentage"`
	TotalSpend   uint   `json:"TotalSpend"` //requirement while claiming
	MaximumUsage uint   `validate:"required" json:"total_usage"`
	GlobalUsage  uint   `validate:"required" json:"gloabl_usage"`
}

type CouponUsage struct {
	gorm.Model
	OrderID    uint   `json:"order_id"`
	UserID     uint   `json:"user_id"`
	CouponCode string `json:"coupon_code"`
	UsageCount uint   `json:"usage_count"`
}

type CouponInventoryRequest struct {
	CouponCode   string `validate:"required" json:"coupon_code" gorm:"primary_key"`
	Expiry       uint   `validate:"required" json:"expiry"`
	Percentage   uint   `validate:"required" json:"percentage"`
	TotalSpend   uint   `validate:"required" json:"total_spend"` //requirement while claiming
	MaximumUsage uint   `validate:"required" json:"total_usage"`
	GlobalUsage  uint   `validate:"required" json:"gloabl_usage"`
}

// create coupons -admin side
func CreateCoupon(c *gin.Context) {
	var Request model.CouponInventoryRequest
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind the json",
		})
		return
	}

	if err := validate(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err,
		})
		return
	}

	if CheckCouponExists(Request.CouponCode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "coupon code already exists",
		})
		return
	}

	if time.Now().Unix()*12*3600 > int64(Request.Expiry) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "please change the expiry time that is more than a day",
		})
		return
	}

	if err:=database.DB.Create(&Request).Error;err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to create coupon",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully created coupon",
	})
}

//update coupon
//add coupon on cart
//remove coupon on cart
//on place order change the amount according to the coupon

func CheckCouponExists(code string) bool {
	var Coupons []model.CouponInventory
	for _, c := range Coupons {
		if c.CouponCode == code {
			return true
		}
	}
	return false
}

//update coupons -admin side
//add coupon -user side on the cart side //check conditions
//calculate and add the finalamount discount amount on the place order
//remove coupon - user side on the cart side //check conditions
