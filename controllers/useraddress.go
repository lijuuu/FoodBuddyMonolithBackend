package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddUserAddress(c *gin.Context) {
	var UserAddress model.Address

	//bind the json to the struct
	if err := c.BindJSON(&UserAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind the incoming request",
			"ok":    false,
		})
		return
	}

	//checking the user sending is performing in his/her account..
	email, ok := EmailFromUserID(UserAddress.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get user email from the database",
			"ok":    false,
		})
		return
	}
	if ok := VerifyJWT(c, email); !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "unauthorized user",
			"ok":    false,
		})
		return
	}

	//to make sure the addressid is autoincremented by the gorm
	UserAddress.AddressID = 0

	//validating the useraddress for required,number tag etc....
	if ok := validate(UserAddress, c); !ok {
		return
	}

	//check if the user exists...
	var userinfo model.User
	if err := database.DB.First(&userinfo, UserAddress.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"ok":    false,
		})
		return
	}

	//check if there is 3 addresses, if >= 3 return address limit reached
	var UserAddresses []model.Address
	if err := database.DB.Where("user_id = ?", UserAddress.UserID).Find(&UserAddresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve the existing user addresses from the database",
			"ok":    false,
		})
		return
	}

	//check if the user already has 3 address...3 is the limit
	if len(UserAddresses) >= 3 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "user already have three addresses, please delete or edit the existing addresses",
			"ok":    false,
		})
		return
	}

	//create the address, provide address_id similar to serial numbers...no 1,2,3 for addresses based on user_id
	if err := database.DB.Create(&UserAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create the address on the database",
			"ok":    false,
		})
		return
	}

	//return the addresses of the particular user
	c.JSON(http.StatusOK, gin.H{
		"useraddresses": UserAddress,
		"ok":            true,
	})
}

func GetUserAddress(c *gin.Context) {

	//get the userid from the param
	userIDStr := c.Param("userid")
	UserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userid format",
			"ok":    true,
		})
		return
	}

	//get the addresses where user_id == UserID
	var UserAddresses []model.Address
	if err := database.DB.Where("user_id = ?", UserID).Find(&UserAddresses).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get informations from the database",
			"ok":    false,
		})
		return
	}

	//checking the user sending is performing in his/her account..
	email, ok := EmailFromUserID(uint(UserID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get user email from the database",
			"ok":    false,
		})
		return
	}
	if ok := VerifyJWT(c, email); !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "unauthorized user",
			"ok":    false,
		})
		return
	}

	//check if there's any address related to the user
	if len(UserAddresses) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no addresses related to the userid",
		})
		return
	}

	//return the addresses
	c.JSON(http.StatusOK, gin.H{
		"useraddress": UserAddresses,
		"ok":          true,
	})
}

func EditUserAddress(c *gin.Context) {
	var updateUserAddress model.Address

	// Bind the incoming JSON to the updateUserAddress struct
	if err := c.BindJSON(&updateUserAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind the incoming request",
			"ok":    false,
		})
		return
	}

	// Retrieve the existing UserAddress record
	var existingUserAddress model.Address
	if err := database.DB.Where("user_id =? AND address_id =?", updateUserAddress.UserID, updateUserAddress.AddressID).First(&existingUserAddress).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "address not found",
			"ok":    false,
		})
		return
	}

	//check if the user is not impersonating other users through jwt email and users email match..
	email, ok := EmailFromUserID(existingUserAddress.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get user email from the database",
			"ok":    false,
		})
		return
	}
	if ok := VerifyJWT(c, email); !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "unauthorized user",
			"ok":    false,
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
			"error": "failed to update the address",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "address updated successfully",
		"ok":      true,
	})
}

func DeleteUserAddress(c *gin.Context) {
	var AddressInfo model.Address

	// Bind the incoming JSON to the updateUserAddress struct
	if err := c.BindJSON(&AddressInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind the incoming request",
			"ok":    false,
		})
		return
	}

	//check the userid and addressid
	var existingUserAddress model.Address
	if err := database.DB.Where("user_id =? AND address_id =?", AddressInfo.UserID, AddressInfo.AddressID).First(&existingUserAddress).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "address not found",
			"ok":    false,
		})
		return
	}

	//check if the user is not impersonating other users through jwt email and users email match..
	email, ok := EmailFromUserID(existingUserAddress.UserID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get user email from the database",
			"ok":    false,
		})
		return
	}
	if ok := VerifyJWT(c, email); !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "unauthorized user",
			"ok":    false,
		})
		return
	}

	//delete the user address
	if err := database.DB.Delete(&existingUserAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete the address",
			"ok":    false,
		})
		return
	}

	//return the response
	c.JSON(http.StatusOK, gin.H{
		"message": "address deleted successfully",
		"ok":      true,
	})

}
