package main

import (
	"foodbuddy/controllers"
	"foodbuddy/database"
	"foodbuddy/utils"
	"foodbuddy/view"

	"github.com/gin-gonic/gin"
	// swaggerfiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
	// "github.com/swaggo/swag/example/basic/docs"
)

func init() {
	database.ConnectToDB()
	database.AutoMigrate()
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Use(controllers.RateLimitMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello, its working",
		})
	})
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	// authentication routes
	// router.POST("/api/v1/adminlogin", controllers.AdminLogin)
	router.POST("/api/v1/user/emaillogin", controllers.EmailLogin)
	router.POST("/api/v1/user/emailsignup", controllers.EmailSignup)
	router.GET("/api/v1/verifyotp/:role/:email/:otp", controllers.VerifyOTP)

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
	router.POST("/api/v1/admin/login", controllers.AdminLogin)
	adminRoutes := router.Group("/api/v1/admin", controllers.CheckAdmin)
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
	}

	restaurantRoutes := router.Group("/api/v1/restaurants")
	{
		// Restaurant Management
		restaurantRoutes.GET("/all", controllers.GetRestaurants)
		restaurantRoutes.POST("/add", controllers.RestaurantSignup)
		restaurantRoutes.POST("/edit", controllers.EditRestaurant)
		restaurantRoutes.GET("/delete/:restaurantid", controllers.DeleteRestaurant)
		restaurantRoutes.GET("/block/:restaurantid", controllers.BlockRestaurant)
		restaurantRoutes.GET("/unblock/:restaurantid", controllers.UnblockRestaurant)

		// Product Management
		restaurantRoutes.GET("/products/all", controllers.GetProductList)
		restaurantRoutes.GET("/products/:restaurantid", controllers.GetProductsByRestaurantID)
		restaurantRoutes.POST("/products/add", controllers.AddProduct)
		restaurantRoutes.POST("/products/edit", controllers.EditProduct)
		restaurantRoutes.GET("/products/delete/:productid", controllers.DeleteProduct)
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
	router.GET("/api/v1/logout", controllers.Logout)

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
