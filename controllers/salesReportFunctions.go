package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"github.com/gin-gonic/gin"
)

//total orders  //oi
//total delivered  //oi
//total cancelled //oi
//revenue generated //oi
//coupon deductions //oi
//product offer deductions //oi
//referral rewards given //oi

func TotalOrders(c *gin.Context) {
   var Order model.Order
   if err:=database.DB.Where("")
}
