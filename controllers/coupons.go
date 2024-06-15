package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CouponInventoryRequest struct {
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
		CouponCode: Request.CouponCode,
		Expiry:     Request.Expiry,
		Percentage: Request.Percentage,
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

func GetAllCoupons(c *gin.Context){
   var Coupons []model.CouponInventory

   if err:= database.DB.Find(&Coupons).Error;err!=nil{
	c.JSON(http.StatusBadRequest,gin.H{
		"status":false,
		"message":"failed to fetch coupon details",
	})
	return
   }

   c.JSON(http.StatusOK,gin.H{
	"status":true,
	"data":Coupons,
	})
}

// update coupon
func UpdateCoupon(c *gin.Context) {
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
	err := database.DB.Where("coupon_code = ?", request.CouponCode).First(&existingCoupon).Error

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

	if time.Now().Unix()+12*3600 > int64(request.Expiry) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "please change the expiry time to more than a day",
		})
		return
	}

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

func ApplyCouponOnCart(c *gin.Context) {
    UserID := c.Param("userid")
    CouponCode := c.Param("couponcode")
    if UserID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  false,
            "message": "ensure userid is present on the URL param",
        })
        return
    }

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
    var sum float64
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

        sum += float64((Product.Price) * (item.Quantity))
    }

    // Apply coupon if provided
    var discount float64
	var Percentage,finalamount float64
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
		Percentage = float64(coupon.Percentage)
        discount = float64(sum) * (float64(coupon.Percentage) / 100.0)
        finalamount = sum - (discount)
    }

    c.JSON(http.StatusOK, gin.H{
        "status": true,
        "data": gin.H{
            "cartitems":   CartItems,
            "totalamount": sum,
			"discountpercentage":Percentage,
            "discount":    discount,
            "finalamount": finalamount,
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

	discountAmount := order.TotalAmount * float64(coupon.Percentage) / 100
	finalAmount := order.TotalAmount - discountAmount

	order.CouponCode = CouponCode
	order.DiscountAmount = discountAmount
	order.FinalAmount = finalAmount

	if err := database.DB.Where("order_id = ?",order.OrderID).Updates(&order).Error; err != nil {
		return false, "failed to apply coupon to order"
	}

	if err == gorm.ErrRecordNotFound {
		couponUsage = model.CouponUsage{
			OrderID:    order.OrderID,
			UserID:     UserID,
			CouponCode: CouponCode,
			UsageCount: 1,
		}
		if err := database.DB.Create(&couponUsage).Error; err != nil {
			return false, "failed to create coupon usage record"
		}
	} else {
		couponUsage.UsageCount++
		if err := database.DB.Where("order_id = ?",order.OrderID).Save(&couponUsage).Error; err != nil {
			return false, "failed to update coupon usage record"
		}
	}

	return true, "coupon applied successfully"
}

//add coupon on cart
//remove coupon on cart
//on place order change the amount according to the coupon

func CheckCouponExists(code string) bool {
	var Coupons []model.CouponInventory
	if err:= database.DB.Find(&Coupons).Error;err!=nil{
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

//update coupons -admin side
//add coupon -user side on the cart side //check conditions
//calculate and add the finalamount discount amount on the place order
//remove coupon - user side on the cart side //check conditions

//add one more payment method if possible
//june 13 - authentication on each endpoints and add request model if possible
//june 14 - sales report, order report etc... pdf,excel file generation
//june 15 - fix all issues...make swagger if possible
