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
	router.Use(utils.CorsMiddleware())
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

	//password resets
	router.POST("/api/v1/auth/passwordreset/step1", controllers.Step1PasswordReset)
	router.GET("/api/v1/auth/passwordreset", controllers.LoadPasswordReset)
	router.POST("/api/v1/auth/passwordreset/step2", controllers.Step2PasswordReset)

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

	// Image upload route
	router.GET("/api/v1/uploadimage", view.LoadUpload)
	router.POST("/api/v1/uploadimage", utils.ImageUpload)

	// Logout route
	router.GET("/api/v1/logout", controllers.Logout)

	// Admin routes with admin middleware
	adminRoutes := router.Group("/api/v1/admin")
	{
		// User management
		adminRoutes.GET("/users", controllers.GetUserList)
		adminRoutes.GET("/users/blocked", controllers.GetBlockedUserList)
		adminRoutes.GET("/users/block/:userid", controllers.BlockUser)
		adminRoutes.GET("/users/unblock/:userid", controllers.UnblockUser)

		// Category management
		adminRoutes.POST("/categories/add", controllers.AddCategory)
		adminRoutes.PUT("/categories/:categoryid", controllers.EditCategory)
		adminRoutes.DELETE("/categories/:categoryid", controllers.DeleteCategory)

		// Restaurant management
		adminRoutes.GET("/restaurants", controllers.GetRestaurants)
		adminRoutes.DELETE("/restaurants/:restaurantid", controllers.DeleteRestaurant)
		adminRoutes.PUT("/restaurants/block/:restaurantid", controllers.BlockRestaurant)
		adminRoutes.PUT("/restaurants/unblock/:restaurantid", controllers.UnblockRestaurant)
	}

	// Restaurant routes with restaurant middleware
	restaurantRoutes := router.Group("/api/v1/restaurants")
	{
		restaurantRoutes.POST("/edit", controllers.EditRestaurant)
		restaurantRoutes.POST("/products/add", controllers.AddProduct)
		restaurantRoutes.POST("/products/edit", controllers.EditProduct)
		restaurantRoutes.DELETE("/products/:productid", controllers.DeleteProduct)

		//order history
		restaurantRoutes.POST("/order/history/", controllers.OrderHistoryRestaurants)
		restaurantRoutes.POST("/order/nextstatus", controllers.UpdateOrderStatusForRestaurant)
	}

	userRoutes := router.Group("/api/v1/user")
	{

		userRoutes.GET("/:userid", controllers.GetUserProfile)
		userRoutes.POST("/edit", controllers.UpdateUserInformation)

		//favourite product by usedid
		userRoutes.GET("/favorites/:userid", controllers.GetFavouriteProductByUserID)
		userRoutes.POST("/favorites/", controllers.AddFavouriteProduct)
		userRoutes.DELETE("/favorites/", controllers.RemoveFavouriteProduct)

		//user address
		userRoutes.GET("/address/:userid", controllers.GetUserAddress)
		userRoutes.POST("/address/add", controllers.AddUserAddress)
		userRoutes.PUT("/address/edit", controllers.EditUserAddress)
		userRoutes.DELETE("/address/delete", controllers.DeleteUserAddress)

		//cart management
		userRoutes.POST("/cart/add", controllers.AddToCart)                      //add items to cart
		userRoutes.GET("/cart/:userid", controllers.GetCartTotal)                //get everything that cart holds
		userRoutes.DELETE("/cart/delete/:userid", controllers.ClearCartByUserID) //remove the entire cart
		userRoutes.DELETE("/cart/remove", controllers.RemoveItemFromCart)        //remove specific products form the cart
		userRoutes.PUT("/cart/update/", controllers.UpdateQuantity)

		userRoutes.POST("/order/step1/placeorder", controllers.PlaceOrder)
		userRoutes.POST("/order/step2/initiatepayment", controllers.InitiatePayment)
		userRoutes.POST("/order/step3/razorpaycallback/:orderid", controllers.RazorPayGatewayCallback)

		userRoutes.POST("/order/history", controllers.UserOrderHistory)
		userRoutes.POST("/order/information", controllers.GetOrderInfoByOrderID)
		userRoutes.POST("/order/paymenthistory", controllers.PaymentDetailsByOrderID)
		userRoutes.POST("/order/cancel", controllers.CancelOrderedProduct)

		userRoutes.POST("/order/review", controllers.UserReviewonOrderItem)
		userRoutes.POST("/order/rating", controllers.UserRatingOrderItem)

		userRoutes.GET("/order/step3/stripecallback", controllers.HandleWebhookStripe)

	}

	//coupons
	adminRoutes.POST("/coupon/create", controllers.CreateCoupon)
	adminRoutes.POST("/coupon/update", controllers.UpdateCoupon)
	userRoutes.GET("/coupon/all", controllers.GetAllCoupons)
	userRoutes.GET("/coupon/cart/:userid/:couponcode", controllers.ApplyCouponOnCart)

	//order management
	//checkout
	//createorder with address selection etc..
	//intiatepayment
	//retry payment
	//cancel order

	router.Run(":8080")
}

//forget password
//add wallet to user and restaurant
//after payment, split the cash to the respective restaurant
