package main

import (
	"foodbuddy/controllers"
	"foodbuddy/database"
	"foodbuddy/utils"
	"foodbuddy/view"

	"github.com/gin-gonic/gin"
)

func init() {
	database.ConnectToDB()
	database.AutoMigrate()
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	// authentication routes
	// router.POST("/api/v1/adminlogin", controllers.AdminLogin)
	router.POST("/api/v1/user/emaillogin", controllers.EmailLogin)
	router.POST("/api/v1/user/emailsignup", controllers.EmailSignup)
	router.POST("/api/v1/user/verifyotp", controllers.VerifyOTP)

	// social login routes
	router.GET("/api/v1/googlesso", controllers.GoogleHandleLogin)
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)

	// public routes for viewing categories, products, and restaurants
	router.GET("/api/v1/public/categories/all", controllers.GetCategoryList)
	router.GET("/api/v1/public/categories/products/all", controllers.GetCategoryProductList)
	router.GET("/api/v1/public/products/all", controllers.GetProductList)
	router.GET("/api/v1/public/products/restaurants/:restaurantid", controllers.GetProductsByRestaurantID)
	router.GET("/api/v1/public/restaurants/all", controllers.GetRestaurants)

	// admin routes with check admin middleware
	adminRoutes := router.Group("/api/v1/admin",controllers.CheckAdmin)
	{
		// user management
		adminRoutes.GET("/users/all", controllers.GetUserList)
		adminRoutes.GET("/users/blocked", controllers.GetBlockedUserList)
		adminRoutes.GET("/users/block/:userid", controllers.BlockUser)
		adminRoutes.GET("/users/unblock/:userid", controllers.UnblockUser)

		// category management
		adminRoutes.GET("/categories/all", controllers.GetCategoryList)
		adminRoutes.GET("/categories/products/all", controllers.GetCategoryProductList)
		adminRoutes.POST("/categories/add", controllers.AddCategory)
		adminRoutes.POST("/categories/edit", controllers.EditCategory)
		adminRoutes.GET("/categories/delete/:categoryid", controllers.DeleteCategory)

		// restaurant management
		adminRoutes.GET("/restaurants/all", controllers.GetRestaurants)
		adminRoutes.POST("/restaurants/add", controllers.AddRestaurant)
		adminRoutes.POST("/restaurants/edit", controllers.EditRestaurant)
		adminRoutes.GET("/restaurants/delete/:restaurantid", controllers.DeleteRestaurant)
		adminRoutes.GET("/restaurants/block/:restaurantid", controllers.BlockRestaurant)
		adminRoutes.GET("/restaurants/unblock/:restaurantid", controllers.UnblockRestaurant)

		// product management
		adminRoutes.GET("/products/all", controllers.GetProductList)
		adminRoutes.GET("/products/restaurants/:restaurantid", controllers.GetProductsByRestaurantID)
		adminRoutes.POST("/products/add", controllers.AddProduct)
		adminRoutes.POST("/products/edit", controllers.EditProduct)
		adminRoutes.GET("/products/delete/:productid", controllers.DeleteProduct)
	}

	// user favorite products routes
	router.GET("/api/v1/user/products/favourite/:userid", controllers.GetFavouriteProductByUserID)
	router.POST("/api/v1/user/products/favourite/add", controllers.AddFavouriteProduct)
	router.POST("/api/v1/user/products/favourite/delete", controllers.RemoveFavouriteProduct)

	// user address routes
	router.POST("/api/v1/user/address/add", controllers.AddUserAddress)
	router.GET("/api/v1/user/address/:userid", controllers.GetUserAddress)
	router.POST("/api/v1/user/address/edit", controllers.EditUserAddress)
	router.POST("/api/v1/user/address/delete", controllers.DeleteUserAddress)

	// image upload route
	router.GET("/api/v1/uploadimage", view.LoadUpload)
	router.POST("/api/v1/uploadimage", utils.ImageUpload)

	// logout route
	router.GET("/api/v1/logout",controllers.Logout)


	router.Run(":8080")

}



// /controllers
//     admin_controller.go          // Handles admin-specific operations.
//     auth_controller.go           // Handles authentication and social login routes.
//     category_controller.go        // Manages CRUD operations for categories.
//     product_controller.go         // Manages CRUD operations for products.
//     restaurant_controller.go      // Manages CRUD operations for restaurants.
//     user_controller.go           // Handles user-related operations like address management, favorite products, etc.
//     upload_controller.go         // Manages file uploads.


//request and reponse for all the endpoints in the postman
//add validation 
//error in google  //solved --changed googleclient api and secret

//change the way sendotp fn works....make it dynamic to process on user and restaurants as well
//restaurant routes
