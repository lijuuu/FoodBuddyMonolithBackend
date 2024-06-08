package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUserProfile(c *gin.Context) {

	//get id
	userIDStr := c.Param("userid")

	//check user info and save it on struct
	var UserProfile model.User
	if err := database.DB.Where("id = ?", userIDStr).First(&UserProfile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
		})
		return
	}

	//return
	c.JSON(http.StatusNotFound, gin.H{
		"status":  true,
		"message": "successfully fetched user profile",
		"data": gin.H{
			"id":           UserProfile.ID,
			"name":         UserProfile.Name,
			"email":        UserProfile.Email,
			"phone_number": UserProfile.PhoneNumber,
			"picture":      UserProfile.Picture,
			"login_method": UserProfile.LoginMethod,
			"blocked":      UserProfile.Blocked,
		},
	})
}

func GetUserList(c *gin.Context) {
	var users []model.User

	tx := database.DB.Select("*").Find(&users)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved user informations",
		"data": gin.H{
			"users": users,
		},
	})
}

func GetBlockedUserList(c *gin.Context) {

	var blockedUsers []model.User

	tx := database.DB.Where("deleted_at IS NULL AND blocked =?", true).Find(&blockedUsers)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve blocked user data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
		})
		return
	}

	if len(blockedUsers) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve blocked user data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved blocked user's data",
		"data": gin.H{
			"blockedusers": blockedUsers,
		},
	})
}

func BlockUser(c *gin.Context) {

	var user model.User

	userIdStr := c.Param("userid")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "invalid user ID",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := database.DB.First(&user, userId).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to fetch user information",
			"error_code": http.StatusNotFound,
		})
		return
	}

	if user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":     false,
			"message":    "user is already blocked",
			"error_code": http.StatusAlreadyReported,
		})
		return
	}

	user.Blocked = true
	fmt.Println(user)
	tx := database.DB.Updates(&user)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to change the block status ",
			"error_code": http.StatusInternalServerError,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  false,
		"message": "successfully blocked the user",
	})
}

func UnblockUser(c *gin.Context) {

	var user model.User

	userIdStr := c.Param("userid")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "invalid user ID",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "User not found",
			"error_code": http.StatusNotFound,
		})
		return
	}

	if !user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":     false,
			"message":    "user is already unblocked",
			"error_code": http.StatusAlreadyReported,
		})
		return
	}

	user.Blocked = false
	fmt.Println(user)
	tx := database.DB.Model(&user).UpdateColumn("blocked", false)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to change the unblock status",
			"error_code": http.StatusInternalServerError,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "user is succefully blocked",
	})
}

func AddUserAddress(c *gin.Context) {
	var UserAddress model.Address

	//bind the json to the struct
	if err := c.BindJSON(&UserAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the incoming request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := validate(UserAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err,
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//checking the user sending is performing in his/her account..
	email, ok := EmailFromUserID(UserAddress.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to get user email from the database",
			"error_code": http.StatusNotFound,
		})
		return
	}
	if err := VerifyJWT(c, model.UserRole, email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
		})
		return
	}

	//to make sure the addressid is autoincremented by the gorm
	UserAddress.AddressID = 0

	//validating the useraddress for required,number tag etc....
	if errs := validate(UserAddress); errs != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    errs,
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//check if the user exists...
	var userinfo model.User
	if err := database.DB.First(&userinfo, UserAddress.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "User not found",
			"error_code": http.StatusNotFound,
		})
		return
	}

	//check if there is 3 addresses, if >= 3 return address limit reached
	var UserAddresses []model.Address
	if err := database.DB.Where("user_id = ?", UserAddress.UserID).Find(&UserAddresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to retrieve the existing user addresses from the database",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//check if the user already has 3 address...3 is the limit
	if len(UserAddresses) >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "user already have three addresses, please delete or edit the existing addresses",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//create the address, provide address_id similar to serial numbers...no 1,2,3 for addresses based on user_id
	if err := database.DB.Create(&UserAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create the address on the database",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//return the addresses of the particular user
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  true,
		"message": "successfully added new address",
		"data": gin.H{
			"address": UserAddress,
		},
	})
}

func GetUserAddress(c *gin.Context) {

	//get the userid from the param
	userIDStr := c.Param("userid")
	UserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process the request data",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//get the addresses where user_id == UserID
	var UserAddresses []model.Address
	if err := database.DB.Where("user_id = ?", UserID).Find(&UserAddresses).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to get informations from the database",
			"error_code": http.StatusNotFound,
		})
		return
	}

	//checking the user sending is performing in his/her account..
	email, ok := EmailFromUserID(uint(UserID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to get user email from the database",
			"error_code": http.StatusNotFound,
		})
		return
	}
	if err := VerifyJWT(c, model.UserRole, email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
		})
		return
	}

	//check if there's any address related to the user
	if len(UserAddresses) == 0 {

		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "no addresses related to the userid",
			"error_code": http.StatusNotFound,
		})
		return
	}

	//return the addresses
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved user address",
		"data": gin.H{
			"addresses": UserAddresses,
		},
	})
}

func EditUserAddress(c *gin.Context) {
	var updateUserAddress model.Address

	// Bind the incoming JSON to the updateUserAddress struct
	if err := c.BindJSON(&updateUserAddress); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the incoming request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Retrieve the existing UserAddress record
	var existingUserAddress model.Address
	if err := database.DB.Where("user_id =? AND address_id =?", updateUserAddress.UserID, updateUserAddress.AddressID).First(&existingUserAddress).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "address not found",
			"error_code": http.StatusNotFound,
		})
		return
	}

	//check if the user is not impersonating other users through jwt email and users email match..
	email, ok := EmailFromUserID(existingUserAddress.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to get user email from the database",
			"error_code": http.StatusNotFound,
		})
		return
	}
	if err := VerifyJWT(c, model.UserRole, email); err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
		})
		return
	}

	// Update the existing record with the new values
	existingUserAddress.AddressType = updateUserAddress.AddressType
	existingUserAddress.StreetName = updateUserAddress.StreetName
	existingUserAddress.StreetNumber = updateUserAddress.StreetNumber
	existingUserAddress.City = updateUserAddress.City
	existingUserAddress.State = updateUserAddress.State
	existingUserAddress.PostalCode = updateUserAddress.PostalCode

	// Save the updated record back to the database
	if err := database.DB.Updates(&existingUserAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update the address",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "address updated successfully",
	})
}

func DeleteUserAddress(c *gin.Context) {
	var AddressInfo model.Address

	// Bind the incoming JSON to the updateUserAddress struct
	if err := c.BindJSON(&AddressInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind the incoming request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//check the userid and addressid
	var existingUserAddress model.Address
	if err := database.DB.Where("user_id =? AND address_id =?", AddressInfo.UserID, AddressInfo.AddressID).First(&existingUserAddress).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "address not found",
			"error_code": http.StatusNotFound,
		})
		return
	}

	//check if the user is not impersonating other users ,
	//using jwt email and users email fo a match
	email, ok := EmailFromUserID(existingUserAddress.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to get user email from the database",
			"error_code": http.StatusNotFound,
		})
		return
	}
	if errs := VerifyJWT(c, model.UserRole, email); errs != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
		})

		return
	}

	//delete the user address
	if err := database.DB.Delete(&existingUserAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to delete the address",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	//return the response
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "address deleted successfully",
	})
}

func UpdateUserInformation(c *gin.Context) {
	//bind json
	var Request model.UpdateUserInformation

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind the json",
		})
		return
	}

	//validate
	if err := validate(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err,
		})
		return
	}

	if ok := CheckUser(Request.ID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "user doesnt exist",
		})
		return
	}

	//update the user information
	if err := database.DB.Model(&model.User{}).Where("id = ?", Request.ID).Updates(&Request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update user profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully updated user profile",
		"data": gin.H{
			"user": Request,
		},
	})
}

//sent reset email

func Step1PasswordReset(c *gin.Context) {
	//receive email
	var Request model.Step1PasswordReset
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "failed to bind the request"})
		return
	}
	//check if the user exists
	var User model.User
	if err:= database.DB.Where("email = ?",Request.Email).First(&User).Error;err!=nil{
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "user doesnt exists"})
		return
	}

	//generate token,expiry etc
	ResetToken := utils.GenerateRandomString(10)
	ExpiryTime := time.Now().Unix() + 1*60
	//sent email  use smtp with token as query

	from := "foodbuddycode@gmail.com"
	appPassword := os.Getenv("SMTPAPP")
	auth := smtp.PlainAuth("", from, appPassword, "smtp.gmail.com")
	url := fmt.Sprintf("http://localhost:8080/api/v1/auth/passwordreset?email=%v&token=%v", Request.Email, ResetToken)
	mail := fmt.Sprintf("FoodBuddy Password Reset \n Click here to reset your password %v", url)

	//send the otp to the specified email
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{Request.Email}, []byte(mail))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to sent the password reset mail"})
		return
	}

	//save it on the server
	var UserPasswordReset model.UserPasswordReset
	UserPasswordReset.Email = Request.Email
	UserPasswordReset.ResetToken = ResetToken
	UserPasswordReset.ExpiryTime = uint(ExpiryTime)

	var CheckUser model.UserPasswordReset
	if err := database.DB.Where("email = ?", Request.Email).First(&CheckUser).Error; err != nil {
		//email row doesnt exist create new entry
		if err := database.DB.Create(&UserPasswordReset).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to save password reset information, try again"})
			return
		}
	} else {
		//update the mail if it exists
		if err := database.DB.Where("email = ?", Request.Email).Updates(&UserPasswordReset).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to save password reset information, try again"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Successfully sent the password reset email"})
}

func LoadPasswordReset(c *gin.Context) {
	email := c.Query("email")
	token := c.Query("token")

	fmt.Println(email, ": ", token)

	c.HTML(http.StatusOK, "passwordreset.html", gin.H{
		"email": email,
		"token": token,
	})
}

func Step2PasswordReset(c *gin.Context) {
	//user clicks the url with token,
	var Request model.Step2PasswordReset
	if err := c.ShouldBind(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "failed to bind the request"})
		return
	}

	if err := validate(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}
	// this function recieves email, token, passwords(check for password match)
	//check email
	var UserPasswordReset model.UserPasswordReset
	if err := database.DB.Where("email = ?", Request.Email).First(&UserPasswordReset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to fetch password reset information"})
		return
	}
	//call step3
	ok, err := Step3PasswordReset(Request)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Successfully Reseted password"})
}

func Step3PasswordReset(Request model.Step2PasswordReset) (bool, error) {
	// check if passwords match
	if Request.Password != Request.ConfirmPassword {
		return false, errors.New("please ensure both the passwords are same")
	}

	// generate salt
	salt := utils.GenerateRandomString(7)

	// combine salt and password and hash it
	saltedPassword := salt + Request.Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, errors.New("failed to hash the password")
	}

	// create a new user instance with updated password and salt
	user := model.User{
		Email:          Request.Email,
		Salt:           salt,
		HashedPassword: string(hashedPassword),
	}

	// save the updated user record
	if err := database.DB.Model(&user).Where("email = ?", user.Email).Updates(user).Error; err != nil {
		return false, errors.New("failed to update the password")
	}

	//change verification status to pending in the verification table as well
	if err:= database.DB.Model(&model.VerificationTable{}).Where("email = ?",Request.Email).Update("verification_status",model.VerificationStatusPending).Error;err!=nil{
		return false, errors.New("failed to update the verification status")
	}

	return true, nil
}
