package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"net/http"
	"strconv"
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

	if err := utils.Validate(&Request); err != nil {
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

	if Request.Percentage > model.CouponDiscountPercentageLimit {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "coupon discount percentage should not exceed more than " + strconv.Itoa(model.CouponDiscountPercentageLimit),
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
		CouponCode:    Request.CouponCode,
		Expiry:        Request.Expiry,
		Percentage:    Request.Percentage,
		MaximumUsage:  Request.MaximumUsage,
		MinimumAmount: float64(Request.MinimumAmount),
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

	if err := utils.Validate(&request); err != nil {
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
	CouponCode := c.Query("couponcode")

	var CartItems []model.CartItems

	if err := database.DB.Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to fetch cart items. Please try again later.",
		})
		return
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "Your cart is empty.",
		})
		return
	}

	// Total price of the cart
	var sum, ProductOfferAmount float64
	for _, item := range CartItems {
		var Product model.Product
		if err := database.DB.Where("id = ?", item.ProductID).First(&Product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "Failed to fetch product information. Please try again later.",
			})
			return
		}

		ProductOfferAmount += float64(Product.OfferAmount) * float64((item.Quantity))
		sum += ((Product.Price) * float64(item.Quantity))
	}

	// Apply coupon if provided
	var CouponDiscount float64
	var FinalAmount float64
	if CouponCode != "" {
		var coupon model.CouponInventory
		if err := database.DB.Where("coupon_code = ?", CouponCode).First(&coupon).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "Invalid coupon code. Please check and try again.",
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
			"cart_items":           CartItems,
			"total_amount":         sum,
			"coupon_discount":      CouponDiscount,
			"product_offer_amount": ProductOfferAmount,
			"final_amount":         FinalAmount,
		},
		"message": "Cart items retrieved successfully",
	})
}

func ApplyCouponToOrder(order model.Order, UserID uint, CouponCode string) (bool, string, model.Order) {

	if order.CouponCode != "" {
		errMsg := fmt.Sprintf("%v coupon already exists, remove this coupon to add a new coupon", order.CouponCode)
		return false, errMsg, order
	}

	var coupon model.CouponInventory
	if err := database.DB.Where("coupon_code = ?", CouponCode).First(&coupon).Error; err != nil {
		return false, "coupon not found", order
	}

	if time.Now().Unix() > int64(coupon.Expiry) {
		return false, "coupon has expired", order
	}

	var couponUsage model.CouponUsage
	err := database.DB.Where("coupon_code = ? AND user_id = ?", CouponCode, UserID).First(&couponUsage).Error

	if err == nil {
		if couponUsage.UsageCount >= coupon.MaximumUsage {
			return false, "coupon usage limit reached", order
		}
	} else if err != gorm.ErrRecordNotFound {
		return false, "database error", order
	}

	//check minimum amount
	if order.TotalAmount < coupon.MinimumAmount {
		errmsg := fmt.Sprintf("minimum of %v is needed for using this coupon", coupon.MinimumAmount)
		return false, errmsg, order
	}

	discountAmount := order.TotalAmount * float64(coupon.Percentage) / 100
	finalAmount := order.TotalAmount - (discountAmount + order.ProductOfferAmount)

	order.CouponCode = CouponCode
	order.CouponDiscountAmount = discountAmount
	order.FinalAmount = finalAmount

	if err := database.DB.Where("order_id = ?", order.OrderID).Updates(&order).Error; err != nil {
		return false, "failed to apply coupon to order", order
	}

	if err == gorm.ErrRecordNotFound {
		couponUsage = model.CouponUsage{
			UserID:     UserID,
			CouponCode: CouponCode,
			UsageCount: 1,
		}
		if err := database.DB.Create(&couponUsage).Error; err != nil {
			return false, "failed to create coupon usage record", order
		}
	} else {
		couponUsage.UsageCount++
		if err := database.DB.Where("order_id = ?", order.OrderID).Save(&couponUsage).Error; err != nil {
			return false, "failed to update coupon usage record", order
		}
	}

	return true, "coupon applied successfully", order
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

func GetRefferalCode(c *gin.Context) {

	email, role, _ := utils.GetJWTClaim(c)
	if role != model.UserRole {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var User model.User
	if err := database.DB.Where("email = ?", email).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user information",
		})
		return
	}

	if User.ReferralCode != "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "referral code : " + User.ReferralCode,
		})
		return
	}

	refCode := utils.GenerateRandomString(5)

	if err := database.DB.Model(&User).Where("email = ?", email).Update("referral_code", refCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to generate referral code",
		})
		return
	}

	if !CreateReferralEntry(User.ID) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to save referral history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "referral code : " + refCode,
	})

}

func ActivateReferral(c *gin.Context) {
	RefCode := c.Query("referralcode")
	email, role, _ := utils.GetJWTClaim(c)
	if role != model.UserRole {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var User model.User
	if err := database.DB.Where("email =?", email).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user information",
		})
		return
	}

	fmt.Println(User)

	var UserReferralHistory model.UserReferralHistory
	if err := database.DB.Where("referral_code =?", User.ReferralCode).First(&UserReferralHistory).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user referral information",
		})
		return
	}

	fmt.Println(UserReferralHistory)

	if UserReferralHistory.ReferredBy != "" {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "user already been referred",
		})
		return
	}

	if UserReferralHistory.ReferredBy == RefCode {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "user's cannot refer each other",
		})
		return
	}

	if RefCode == User.ReferralCode {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"status":  false,
			"message": "usage of same referral code restricted",
		})
		return
	}

	if err := database.DB.Where("referral_code =?", User.ReferralCode).First(&UserReferralHistory).Update("referred_by", RefCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update user referral information",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully finished refer process",
	})
}
func GenerateReferralCodeForUser(email string) (string, error) {
	var User model.User
	if err := database.DB.Where("email =?", email).First(&User).Error; err != nil {
		return "", err
	}

	if User.ReferralCode != "" {
		return User.ReferralCode, nil
	}

	refCode := utils.GenerateRandomString(5)

	if err := database.DB.Model(&User).Where("email =?", email).Update("referral_code", refCode).Error; err != nil {
		return "", err
	}

	if !CreateReferralEntry(User.ID) {
		return "", errors.New("failed to save referral history")
	}

	return refCode, nil
}

func CreateReferralEntry(UserID uint) bool {
	var User model.User
	if err := database.DB.Where("id =?", UserID).First(&User).Error; err != nil {
		return false
	}
	var existingEntry model.UserReferralHistory
	if err := database.DB.Where("user_id =?", UserID).First(&existingEntry).Error; err == nil {
		return true
	} else if err != gorm.ErrRecordNotFound {
		return false
	}

	RefHistory := model.UserReferralHistory{
		UserID:       UserID,
		ReferralCode: User.ReferralCode,
	}
	if err := database.DB.Create(&RefHistory).Error; err != nil {
		return false
	}
	return true
}

func ClaimReferralRewards(c *gin.Context) {
	email, role, _ := utils.GetJWTClaim(c)
	if role != model.UserRole {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var User model.User
	if err := database.DB.Where("email =?", email).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user information",
		})
		return
	}

	var Referrals []model.UserReferralHistory
	if err := database.DB.Where("referred_by =? AND refer_claimed =?", User.ReferralCode, false).Find(&Referrals).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user refer history",
		})
		return
	}

	var eligibleClaims int64
	for _, v := range Referrals {
		if err := database.DB.Model(&model.OrderItem{}).Where("user_id =? AND order_status =?", v.UserID, model.OrderStatusDelivered).Count(&eligibleClaims).Error; err != nil {
			continue
		}
	}

	if eligibleClaims < model.ReferralClaimLimit {
		errMsg := fmt.Sprintf("need a minimum of %v referrals with at least one order delivered to claim reward", model.ReferralClaimLimit)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"status":  false,
			"message": errMsg,
		})
		return
	}

	fmt.Println(User)

	if err := database.DB.Model(&model.UserReferralHistory{}).Where("referred_by =? AND refer_claimed =?", User.ReferralCode, false).Update("refer_claimed", 1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update user refer history",
		})
		return
	}

	PossibleClaimAmount := eligibleClaims * model.ReferralClaimAmount

	UserWalletHistory := model.UserWalletHistory{
		TransactionTime: time.Now(),
		UserID:          User.ID,
		Type:            model.WalletIncoming,
		Amount:          float64(PossibleClaimAmount),
		CurrentBalance:  float64(PossibleClaimAmount) + User.WalletAmount,
		Reason:          model.WalletTxTypeReferralReward,
	}

	User.WalletAmount += float64(PossibleClaimAmount)
	if err := database.DB.Updates(&User).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update user",
		})
		return
	}

	if err := database.DB.Create(&UserWalletHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to  create user refer history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"message":           "successfully claimed referral reward",
			"eligible_referrals": eligibleClaims,
			"claim_refund":     PossibleClaimAmount,
		},
	})
}

func GetReferralStats(c *gin.Context) {
	email, role, _ := utils.GetJWTClaim(c)
	if role != model.UserRole {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	var User model.User
	if err := database.DB.Where("email =?", email).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user information",
		})
		return
	}

	fmt.Println(User)

	var CompleteReferrals []model.UserReferralHistory
	if err := database.DB.Where("referred_by = ?", User.ReferralCode).Find(&CompleteReferrals).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user refer history",
		})
		return
	}

	var EligibleReferrals int64
	for _, v := range CompleteReferrals {
		if v.ReferClaimed {
			continue
		}

		if err := database.DB.Model(&model.OrderItem{}).Where("user_id =? AND order_status =?", v.UserID, model.OrderStatusDelivered).Count(&EligibleReferrals).Error; err != nil {
			continue
		}
	}

	var claimsDone int64
	if err := database.DB.Model(&model.UserReferralHistory{}).Where("referred_by = ? AND refer_claimed = ?", User.ReferralCode, true).Count(&claimsDone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve total claims",
		})
		return
	}

	var totalClaimedAmount float64
	var userWalletHistories []model.UserWalletHistory
	if err := database.DB.Where("reason = ? AND user_id =?", model.WalletTxTypeReferralReward, User.ID).Find(&userWalletHistories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve total claimed amount",
		})
		return
	}

	for _, history := range userWalletHistories {
		totalClaimedAmount += history.Amount
	}

	fmt.Println("length is  :", len(CompleteReferrals))
	var IneligibleReferrals int
	if len(CompleteReferrals) == 0 {
		IneligibleReferrals = 0
	} else {
		IneligibleReferrals = len(CompleteReferrals) - int(claimsDone) - int(EligibleReferrals)
		if IneligibleReferrals < 0 {
			IneligibleReferrals = 0
		}
	}

	referralStats := gin.H{
		"total_referrals":           len(CompleteReferrals),
		"ineligible_referrals":      IneligibleReferrals,
		"eligible_referrals":        EligibleReferrals,
		"claims_done":               claimsDone,
		"total_claim_amount_received": totalClaimedAmount,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          true,
		"data":            referralStats,
		"referral_history": CompleteReferrals,
	})
}
