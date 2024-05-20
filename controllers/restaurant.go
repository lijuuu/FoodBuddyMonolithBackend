package controllers

import (
	"github.com/gin-gonic/gin"
)

func GetRestaurants(c *gin.Context) {
	// var restaurants model.Restaurant
	//search db and get all
}

func AddRestaurant(c *gin.Context) {
   //bind json
   //check if it already exists
   //add to db
}

func EditRestaurant(c *gin.Context) {
   //check existing restuarant 
   //if present update it with the new data
}

func DeleteRestaurant(c *gin.Context)  {
	//check if its already present
   //check if its deleted if yes sent already deletd
   //if no delete it
}

func BlockRestaurant(c *gin.Context) {
   //check resturant by id 
   //set Blocked as true
}

func UnblockRestaurant(c *gin.Context) {
  //check restruant id 
  //check whether if is active, if no unblock it
  //use column in db, false is a zero value
}
