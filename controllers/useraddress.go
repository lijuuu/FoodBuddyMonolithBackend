package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddUserAddress(c *gin.Context)  {
	var UserAddress model.UserAddress

	//bind the json to the struct 
	if err:= c.BindJSON(&UserAddress);err != nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"failed to bind the incoming request",
			"ok":false,
		})
		return
	}

	if ok := validate(UserAddress,c);!ok{
		return
	}

	//check if there is 3 addresses, if >= 3 return address limit reached
	var UserAddresses []model.UserAddress
	if err:= database.DB.Where("user_id = ?",UserAddress.UserID).Find(&UserAddresses).Error;err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error":"failed to retrieve the existing user addresses from the database",
			"ok":false,
		})
		return
	}

	if len(UserAddresses) >= 3{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error":"user already have three addresses, please delete or edit the existing addresses",
			"ok":false,
		})
		return
	}

	//create the address, provide address_id similar to serial numbers...no 1,2,3 for addresses based on user_id
    if err:= database.DB.Create(&UserAddress).Error;err!= nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error":"failed to create the address on the database",
			"ok":false,
		})
		return
	} 

    //return the addresses of the particular user
	c.JSON(http.StatusOK,gin.H{
		"useraddresses":UserAddresses,
		"ok":true,
	})

}

func GetUserAddress(c *gin.Context)  {

	//get the userid from the param
	UserID:= c.Param("userid")
	if UserID == ""{
		c.JSON(http.StatusNotFound,gin.H{
			"error":"failed to get the params",
			"ok":false,
		})
		return
	}

	//get the addresses where user_id == UserID
  	var UserAddresses []model.UserAddress
	if err:= database.DB.Where("user_id = ?",UserID).Find(&UserAddresses).Error;err!= nil{
		c.JSON(http.StatusNotFound,gin.H{
			"error":"failed to get informations from the database",
			"ok":false,
		})
		return
	}

	if len(UserAddresses) == 0{
       c.JSON(http.StatusNotFound,gin.H{
		"error":"no addresses related to the userid",
	   })
	   return
	}

	//return the addresses
	c.JSON(http.StatusOK,gin.H{
		"useraddress":UserAddresses,
		"ok":true,
	})
}

func EditUserAddress(c *gin.Context)  {
  	
}

func DeleteUserAddress(c *gin.Context)  {
  	
}