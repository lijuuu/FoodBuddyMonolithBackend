package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	if err := database.DB.Model(&model.User{}).Where("referral_code = ?", RefCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "the referral code doesnt exist",
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
			"message":            "successfully claimed referral reward",
			"eligible_referrals": eligibleClaims,
			"claim_refund":       PossibleClaimAmount,
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
		"total_referrals":             len(CompleteReferrals),
		"ineligible_referrals":        IneligibleReferrals,
		"eligible_referrals":          EligibleReferrals,
		"claims_done":                 claimsDone,
		"total_claim_amount_received": totalClaimedAmount,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           true,
		"data":             referralStats,
		"referral_history": CompleteReferrals,
	})
}
