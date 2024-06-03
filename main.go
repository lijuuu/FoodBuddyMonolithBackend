package main

import (
	"foodbuddy/controllers"
	"foodbuddy/database"
	"foodbuddy/utils"
	"foodbuddy/view"

	_ "foodbuddy/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	database.ConnectToDB()
	database.AutoMigrate()
}

// @title FoodBuddy API
// @version 1.0
// @description Documentation

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @query.collection.format multi
func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Use(controllers.RateLimitMiddleware())

	// Health check route
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello, its working",
		})
	})

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) //https://github.com/swaggo/swag/issues/197#issuecomment-1100847754

	// Authentication routes
	router.POST("/api/v1/auth/admin/login", controllers.AdminLogin)
	router.POST("/api/v1/auth/user/email/login", controllers.EmailLogin)
	router.POST("/api/v1/auth/user/email/signup", controllers.EmailSignup)
	router.GET("/api/v1/auth/verifyotp/:role/:email/:otp", controllers.VerifyOTP)

	// Social login routes
	router.GET("/api/v1/auth/google/login", controllers.GoogleHandleLogin)
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)

	// Restaurant authentication routes
	router.POST("/api/v1/auth/restaurant/signup", controllers.RestaurantSignup)
	router.POST("/api/v1/auth/restaurant/login", controllers.RestaurantLogin)

	// Public routes for viewing categories, products, and restaurants
	router.GET("/api/v1/public/categories", controllers.GetCategoryList)
	router.GET("/api/v1/public/categories/products", controllers.GetCategoryProductList)
	router.GET("/api/v1/public/products", controllers.GetProductList)
	router.GET("/api/v1/public/restaurants", controllers.GetRestaurants)
	router.GET("/api/v1/public/restaurants/products/:restaurantid", controllers.GetProductsByRestaurantID)

	// Admin routes with admin middleware
	adminRoutes := router.Group("/api/v1/admin")
	adminRoutes.Use(controllers.CheckAdmin)
	{
		// User management
		adminRoutes.GET("/users", controllers.GetUserList)
		adminRoutes.GET("/users/blocked", controllers.GetBlockedUserList)
		adminRoutes.POST("/users/block/:userid", controllers.BlockUser)
		adminRoutes.POST("/users/unblock/:userid", controllers.UnblockUser)

		// Category management
		adminRoutes.GET("/categories", controllers.GetCategoryList)
		adminRoutes.POST("/categories/add", controllers.AddCategory)
		adminRoutes.PUT("/categories/:categoryid", controllers.EditCategory)
		adminRoutes.DELETE("/categories/:categoryid", controllers.DeleteCategory)

		// Restaurant management
		adminRoutes.GET("/restaurants", controllers.GetRestaurants)
		adminRoutes.POST("/restaurants/edit", controllers.EditRestaurant)
		adminRoutes.DELETE("/restaurants/:restaurantid", controllers.DeleteRestaurant)
		adminRoutes.POST("/restaurants/block/:restaurantid", controllers.BlockRestaurant)
		adminRoutes.POST("/restaurants/unblock/:restaurantid", controllers.UnblockRestaurant)

		// Product management
		adminRoutes.GET("/products", controllers.GetProductList)
		adminRoutes.POST("/products/add", controllers.AddProduct)
		adminRoutes.POST("/products/:productid", controllers.EditProduct)
		adminRoutes.DELETE("/products/:productid", controllers.DeleteProduct)
	}

	// Restaurant routes with restaurant middleware
	restaurantRoutes := router.Group("/api/v1/restaurants")
	restaurantRoutes.Use(controllers.CheckRestaurant)
	{
		restaurantRoutes.POST("/edit", controllers.EditRestaurant)
		restaurantRoutes.POST("/products/add", controllers.AddProduct)
		restaurantRoutes.POST("/products/edit", controllers.EditProduct)
		restaurantRoutes.DELETE("/products/:productid", controllers.DeleteProduct)
	}

	userRoutes := router.Group("/api/v1/user")
	{
		userRoutes.GET("/favorites/:userid", controllers.GetFavouriteProductByUserID)
		userRoutes.POST("/favorites/:userid", controllers.AddFavouriteProduct)
		userRoutes.DELETE("/favorites/:userid", controllers.RemoveFavouriteProduct)
		userRoutes.GET("/:userid", controllers.GetUserProfile)
		userRoutes.POST("/update", controllers.UpdateUserInformation)
		userRoutes.POST("/address/add", controllers.AddUserAddress)
		userRoutes.GET("/address/:userid", controllers.GetUserAddress)
		userRoutes.PUT("/address/edit", controllers.EditUserAddress)
		userRoutes.DELETE("/address", controllers.DeleteUserAddress)
	}

	// Image upload route
	router.GET("/api/v1/uploadimage", view.LoadUpload)
	router.POST("/api/v1/uploadimage", utils.ImageUpload)

	// Logout route
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

//https://changelly.com/transaction/r01ya2dpntiu7507
//https://localmonero.co/blocks/tx/485cae6969fa60f98a4ad5a74ebf7b6b89412cd13e1b2caba29cbd2b33814889
