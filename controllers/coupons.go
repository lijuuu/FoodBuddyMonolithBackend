package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// create coupons -admin side
func CreateCoupon(c *gin.Context) { //admin
	// check admin api authentication
	_, role, err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
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
			"message": err.Error(),
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

	if time.Now().Unix()+12*3600 > int64(Request.Expiry) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "please change the expiry time that is more than a day",
		})
		return
	}

	Coupon := model.CouponInventory{
		CouponCode:   Request.CouponCode,
		Expiry:       Request.Expiry,
		Percentage:   Request.Percentage,
		MaximumUsage: Request.MaximumUsage,
	}

	if err := database.DB.Create(&Coupon).Error; err != nil {
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

func GetAllCoupons(c *gin.Context) { //public
	var Coupons []model.CouponInventory

	if err := database.DB.Find(&Coupons).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to fetch coupon details",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   Coupons,
	})
}

// update coupon
func UpdateCoupon(c *gin.Context) { //admin
	// check admin api authentication
	_, role, err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	var request model.CouponInventoryRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind the json",
		})
		return
	}

	if err := validate(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	var existingCoupon model.CouponInventory
	err = database.DB.Where("coupon_code = ?", request.CouponCode).First(&existingCoupon).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "coupon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to find coupon",
		})
		return
	}

	existingCoupon.Expiry = request.Expiry
	existingCoupon.Percentage = request.Percentage
	existingCoupon.MaximumUsage = request.MaximumUsage

	if err := database.DB.Save(&existingCoupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update coupon",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully updated coupon",
	})
}

func ApplyCouponOnCart(c *gin.Context) { //user
	// check restaurant api authentication
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}
	UserID, _ := UserIDfromEmail(email)
	CouponCode := c.Param("couponcode")

	var CartItems []model.CartItems

	if err := database.DB.Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to fetch cart items. Please try again later.",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "Your cart is empty.",
			"error_code": http.StatusNotFound,
		})
		return
	}

	// Total price of the cart
	var sum, ProductOfferAmount float64
	for _, item := range CartItems {
		var Product model.Product
		if err := database.DB.Where("id = ?", item.ProductID).First(&Product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":     false,
				"message":    "Failed to fetch product information. Please try again later.",
				"error_code": http.StatusNotFound,
			})
			return
		}

		ProductOfferAmount += float64(Product.OfferAmount)  * float64((item.Quantity))
		sum += ((Product.Price) * float64(item.Quantity))
	}

	// Apply coupon if provided
	var CouponDiscount float64
	var FinalAmount float64
	if CouponCode != "" {
		var coupon model.CouponInventory
		if err := database.DB.Where("coupon_code = ?", CouponCode).First(&coupon).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":     false,
				"message":    "Invalid coupon code. Please check and try again.",
				"error_code": http.StatusNotFound,
			})
			return
		}

		// Check coupon expiration
		if time.Now().Unix() > int64(coupon.Expiry) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": "The coupon has expired.",
			})
			return
		}

		//check minimum amount
		if sum < coupon.MinimumAmount {
			errmsg := fmt.Sprintf("minimum of %v is needed for using this coupon", coupon.MinimumAmount)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": errmsg,
			})
			return
		}

		// Check coupon usage
		var usage model.CouponUsage
		if err := database.DB.Where("user_id = ? AND coupon_code = ?", UserID, CouponCode).First(&usage).Error; err == nil {
			if usage.UsageCount >= coupon.MaximumUsage {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  false,
					"message": "The coupon usage limit has been reached.",
				})
				return
			}
		}

		// Calculate discount
		CouponDiscount = float64(sum) * (float64(coupon.Percentage) / 100.0)
		FinalAmount = sum - (CouponDiscount + ProductOfferAmount)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"CartItems":            CartItems,
			"TotalAmount":          sum,
			"CouponDiscount":       CouponDiscount,
			"ProductOfferDiscount": ProductOfferAmount,
			"FinalAmount":          FinalAmount,
		},
		"message": "Cart items retrieved successfully",
	})
}

func ApplyCouponToOrder(order model.Order, UserID uint, CouponCode string) (bool, string) {

	if order.CouponCode != "" {
		errMsg := fmt.Sprintf("%v coupon already exists, remove this coupon to add a new coupon", order.CouponCode)
		return false, errMsg
	}

	var coupon model.CouponInventory
	if err := database.DB.Where("coupon_code = ?", CouponCode).First(&coupon).Error; err != nil {
		return false, "coupon not found"
	}

	if time.Now().Unix() > int64(coupon.Expiry) {
		return false, "coupon has expired"
	}

	var couponUsage model.CouponUsage
	err := database.DB.Where("coupon_code = ? AND user_id = ?", CouponCode, UserID).First(&couponUsage).Error

	if err == nil {
		if couponUsage.UsageCount >= coupon.MaximumUsage {
			return false, "coupon usage limit reached"
		}
	} else if err != gorm.ErrRecordNotFound {
		return false, "database error"
	}

	//check minimum amount
	if order.TotalAmount < coupon.MinimumAmount {
		errmsg := fmt.Sprintf("minimum of %v is needed for using this coupon", coupon.MinimumAmount)
		return false, errmsg
	}

	discountAmount := order.TotalAmount * float64(coupon.Percentage) / 100
	finalAmount := order.TotalAmount - (discountAmount + order.ProductOfferAmount)

	order.CouponCode = CouponCode
	order.CouponDiscountAmount = discountAmount
	order.FinalAmount = finalAmount

	if err := database.DB.Where("order_id = ?", order.OrderID).Updates(&order).Error; err != nil {
		return false, "failed to apply coupon to order"
	}

	if err == gorm.ErrRecordNotFound {
		couponUsage = model.CouponUsage{
			UserID:     UserID,
			CouponCode: CouponCode,
			UsageCount: 1,
		}
		if err := database.DB.Create(&couponUsage).Error; err != nil {
			return false, "failed to create coupon usage record"
		}
	} else {
		couponUsage.UsageCount++
		if err := database.DB.Where("order_id = ?", order.OrderID).Save(&couponUsage).Error; err != nil {
			return false, "failed to update coupon usage record"
		}
	}

	return true, "coupon applied successfully"
}

func CheckCouponExists(code string) bool {
	var Coupons []model.CouponInventory
	if err := database.DB.Find(&Coupons).Error; err != nil {
		return false
	}
	fmt.Println(&Coupons)
	for _, c := range Coupons {
		if c.CouponCode == code {
			return true
		}
	}
	return false
}
