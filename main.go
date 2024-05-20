package main

import (
	"foodbuddy/controllers"
	"foodbuddy/database"
	"foodbuddy/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	database.ConnectToDB()
	database.AutoMigrate()
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.POST("/api/v1/emaillogin", controllers.EmailLogin)
	router.POST("/api/v1/emailsignup", controllers.EmailSignup)

	//otpverification
	// router.GET("/api/v1/emailotp") // Uncomment and use this if you decide to implement it
	router.POST("/api/v1/verifyotp", controllers.VerifyOTP)

	//check whether the jwt is a valid one, takes the jwttoken from cookie "Authorization"
	router.GET("/api/v1/verifyjwt-cookie", controllers.VerifyJWT)

	//pass jwt token as a json
	router.POST("/api/v1/verifyjwt-json", utils.GetJWTEmailClaim)

	//load google sso page and get result as json
	router.GET("/api/v1/googlesso", controllers.GoogleHandleLogin)
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)

	//admin user management
	router.GET("/api/v1/admin/users/all", controllers.GetUserList)
	router.GET("/api/v1/admin/users/blocked", controllers.GetBlockedUserList)
	router.GET("/api/v1/admin/users/block/:userid", controllers.BlockUser)
	router.GET("/api/v1/admin/users/unblock/:userid", controllers.UnblockUser)

	//admin category management
	router.GET("/api/v1/admin/category/all", controllers.GetCategoryList)
	router.POST("/api/v1/admin/category/add", controllers.AddCategory)
	router.POST("/api/v1/admin/category/edit", controllers.EditCategory)
	router.GET("/api/v1/admin/category/delete/:categoryid", controllers.DeleteCategory)

	//admin product management
	router.GET("/api/v1/admin/product/all", controllers.GetProductList)
	router.POST("/api/v1/admin/product/add", controllers.AddProduct)                 //productid = 0 ;add product only if doesnt exist on the the tables,and only with the valid category
	router.POST("/api/v1/admin/product/edit", controllers.EditProduct)               //productid = 0,only allow values from categrory
	router.GET("/api/v1/admin/product/delete/:productid", controllers.DeleteProduct) //check if the product is deleted

	//admin category management
	router.GET("/api/v1/admin/category/all", controllers.GetCategoryList)
	router.POST("/api/v1/admin/category/add", controllers.AddCategory)
	router.POST("/api/v1/admin/category/edit", controllers.EditCategory)
	router.GET("/api/v1/admin/category/delete/:categoryid", controllers.DeleteCategory)
	
	//logout - removes the cookie "Authorization"
	router.GET("/api/v1/logout", controllers.Logout)

	router.Run(":8080")
}
