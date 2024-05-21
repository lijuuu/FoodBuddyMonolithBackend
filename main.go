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

	router.POST("/api/v1/emaillogin", controllers.EmailLogin) //pass
	router.POST("/api/v1/emailsignup", controllers.EmailSignup) //pass

	//otpverification
	// router.GET("/api/v1/emailotp") // Uncomment and use this if you decide to implement it
	router.POST("/api/v1/verifyotp", controllers.VerifyOTP) //pass

	//check whether the jwt is a valid one, takes the jwttoken from cookie "Authorization"
	router.GET("/api/v1/verifyjwt-cookie", controllers.VerifyJWT)//pass

	//pass jwt token as a json
	router.POST("/api/v1/verifyjwt-json", utils.GetJWTEmailClaim)//pass

	//load google sso page and get result as json
	router.GET("/api/v1/googlesso", controllers.GoogleHandleLogin)//pass
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)//pass

	//admin user management
	router.GET("/api/v1/admin/users/all", controllers.GetUserList)//pass
	// router.GET("/api/v1/admin/users/edit", controllers.EditUser)
	router.GET("/api/v1/admin/users/blocked", controllers.GetBlockedUserList)//pass
	router.GET("/api/v1/admin/users/block/:userid", controllers.BlockUser)//pass
	router.GET("/api/v1/admin/users/unblock/:userid", controllers.UnblockUser)//pass

	//admin category management
	router.GET("/api/v1/admin/categories/all", controllers.GetCategoryList)//pass
	router.GET("/api/v1/admin/categories/products/all", controllers.GetCategoryProductList)//pass
	router.POST("/api/v1/admin/categories/add", controllers.AddCategory)//pass
	router.POST("/api/v1/admin/categories/edit", controllers.EditCategory)//pass
	router.GET("/api/v1/admin/categories/delete/:categoryid", controllers.DeleteCategory)//pass

	//admin product management
	router.GET("/api/v1/admin/products/all", controllers.GetProductList)//pass
	router.POST("/api/v1/admin/products/add", controllers.AddProduct)                 //productid = 0 ;add product only if doesnt exist on the the tables,and only with the valid category
	router.POST("/api/v1/admin/products/edit", controllers.EditProduct)               //productid = 0,only allow values from categrory
	router.GET("/api/v1/admin/products/delete/:productid", controllers.DeleteProduct) //check if the product is deleted

	//admin category management
	router.GET("/api/v1/admin/restaurants/all", controllers.GetRestaurants)
	router.GET("/api/v1/admin/restaurants/products/:restaurantid", controllers.GetRestaurantProductsByID)
	router.POST("/api/v1/admin/restaurants/add", controllers.AddRestaurant)
	router.POST("/api/v1/admin/restaurants/edit", controllers.EditRestaurant)
	router.GET("/api/v1/admin/restaurants/delete/:restaurantid", controllers.DeleteRestaurant)
	router.GET("/api/v1/admin/restaurants/block/:restaurantid", controllers.BlockRestaurant)
	router.GET("/api/v1/admin/restaurants/unblock/:restaurantid", controllers.UnblockRestaurant)
	
	//logout - removes the cookie "Authorization"
	router.GET("/api/v1/logout", controllers.Logout)

	router.Run(":8080")
}
