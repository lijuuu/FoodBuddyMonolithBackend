package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserList(c *gin.Context) {
	var users []model.User

	tx := database.DB.Select("*").Find(&users)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}

	if len(blockedUsers) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve blocked user data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}

	if err := database.DB.First(&user, userId).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to fetch user information",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
		})
		return
	}

	if user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":     false,
			"message":    "user is already blocked",
			"error_code": http.StatusAlreadyReported,
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}

	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "User not found",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
		})
		return
	}

	if !user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":     false,
			"message":    "user is already unblocked",
			"error_code": http.StatusAlreadyReported,
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}
	if err := VerifyJWT(c, email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}

	//check if the user already has 3 address...3 is the limit
	if len(UserAddresses) >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "user already have three addresses, please delete or edit the existing addresses",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	//create the address, provide address_id similar to serial numbers...no 1,2,3 for addresses based on user_id
	if err := database.DB.Create(&UserAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to create the address on the database",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	//return the addresses of the particular user
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  true,
		"message": "successfully added new address",
		"data": gin.H{
			"address": UserAddresses,
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
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}
	if err := VerifyJWT(c, email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
		})
		return
	}

	//check if there's any address related to the user
	if len(UserAddresses) == 0 {

		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "no addresses related to the userid",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}
	if err := VerifyJWT(c, email); err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
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
			"data":       gin.H{},
		})
		return
	}
	if errs := VerifyJWT(c, email); errs != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized user",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
		})

		return
	}

	//delete the user address
	if err := database.DB.Delete(&existingUserAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to delete the address",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	//return the response
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "address deleted successfully",
	})

}
