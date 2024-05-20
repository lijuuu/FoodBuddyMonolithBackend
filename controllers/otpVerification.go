package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"math/rand"
	"net/http"
	"net/smtp"
	"time"

	"github.com/gin-gonic/gin"
)

func SendOTP(c *gin.Context, userID uint, to string, otpexpiry int64) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := r.Intn(900000) + 100000

	// Check if the provided otpexpiry has already passed
	now := time.Now().Unix()
	if otpexpiry > 0 && now < otpexpiry {
		// OTP is still valid, respond with a message and do not send a new OTP
		c.JSON(http.StatusOK, gin.H{
			"error": "OTP is still valid. wait before sending another request.",
			"ok":    false,
		})
		return
	}

	// Proceed to send a new OTP if the previous one has expired or if no otpexpiry was provided
	// Set expiryTime as 10 minutes from now
	expiryTime := now + 10*60 // 10 minutes in seconds

	fmt.Printf("Sending mail because OTP has expired: %v\n", expiryTime)

	from := "foodbuddycode@gmail.com"
	appPassword := "emdnwucohpvcoyin"
	auth := smtp.PlainAuth("", from, appPassword, "smtp.gmail.com")

	mail := fmt.Sprintf("FoodBuddy OTP Verification\nThis is your Verification code: %d", otp)

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(mail))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "sending otp failed",
			"ok":    false,
		})
		return
	}

	user := model.User{
		ID:        userID,
		OTP:       otp,
		OTPexpiry: expiryTime, // Store the Unix timestamp directly
	}

	tx := database.DB.Updates(&user)
	if tx.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "failed to save otp on database",
			"ok":    false,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "verification code sent to mail,verify to continue",
			"ok":      true,
		})
	}
}

func VerifyOTP(c *gin.Context) {

	var userRequest model.User
	var user model.User

	if err := c.BindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	tx := database.DB.Where("email =?", userRequest.Email).First(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user not found",
			"ok":    false,
		})
		return
	}

	if user.VerificationStatus == model.VerificationStatusVerified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user already verified",
			"ok":    false,
		})
		return
	}

	if user.OTP == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "login before verifying otp",
			"ok":    false,
		})
		return
	}

	if user.OTPexpiry < time.Now().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "otp expired",
			"ok":    false,
		})
		return
	}

	if user.OTP != userRequest.OTP {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid otp",
			"ok":    false,
		})
		return
	}

	// Correctly placed inside the if block to ensure it only executes if the OTP is correct
	user.VerificationStatus = model.VerificationStatusVerified

	tx = database.DB.Updates(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "otp verification failed",
			"ok":    false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "otp verified successfully",
		"user":    user,
		"ok":      true,
	})
}
