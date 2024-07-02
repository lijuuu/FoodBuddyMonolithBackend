package main

import (
	"foodbuddy/controllers"
	"foodbuddy/helper"
	"foodbuddy/view"

	"github.com/gin-gonic/gin"
)

func ServerHealth(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})
}
func AuthenticationRoutes(router *gin.Engine) {
	// Authentication Endpoints
	//admin
	router.POST("/api/v1/auth/admin/login", controllers.AdminLogin) //mark

	//user
	router.POST("/api/v1/auth/user/email/login", controllers.EmailLogin)   //mark
	router.POST("/api/v1/auth/user/email/signup", controllers.EmailSignup) //mark
	router.GET("/api/v1/auth/google/login", controllers.GoogleHandleLogin) //mark
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback) //mark

	//additional endpoints for email verification and password reset
	router.GET("/api/v1/auth/verifyotp/:role/:email/:otp", controllers.VerifyOTP) //mark

	router.POST("/api/v1/auth/passwordreset/step1", controllers.Step1PasswordReset) //mark
	router.GET("/api/v1/auth/passwordreset", controllers.LoadPasswordReset)         //mark
	router.POST("/api/v1/auth/passwordreset/step2", controllers.Step2PasswordReset) //mark

	//restaurant
	router.POST("/api/v1/auth/restaurant/signup", controllers.RestaurantSignup) //mark
	router.POST("/api/v1/auth/restaurant/login", controllers.RestaurantLogin)   //mark
}

func UserRoutes(router *gin.Engine) {
	userRoutes := router.Group("/api/v1/user")
	{
		// User Profile Management
		userRoutes.GET("/profile", controllers.GetUserProfile)       //mark
		userRoutes.POST("/edit", controllers.UpdateUserInformation)  //mark
		userRoutes.GET("/wallet/all", controllers.GetUserWalletData) //mark

		// Favorite Products
		userRoutes.GET("/favorites/all", controllers.GetUsersFavouriteProduct)     //mark
		userRoutes.POST("/favorites/add", controllers.AddFavouriteProduct)         //mark
		userRoutes.DELETE("/favorites/delete", controllers.RemoveFavouriteProduct) //mark

		// User Address Management
		userRoutes.GET("/address/all", controllers.GetUserAddress)          //mark
		userRoutes.POST("/address/add", controllers.AddUserAddress)         //mark
		userRoutes.PATCH("/address/edit", controllers.EditUserAddress)      //mark
		userRoutes.DELETE("/address/delete", controllers.DeleteUserAddress) //mark

		// Cart Management
		userRoutes.POST("/cart/add", controllers.AddToCart)
		userRoutes.GET("/cart/all", controllers.GetCartTotal)
		userRoutes.DELETE("/cart/delete/", controllers.ClearCart)
		userRoutes.DELETE("/cart/remove", controllers.RemoveItemFromCart)
		userRoutes.PUT("/cart/update/", controllers.UpdateQuantity)

		// Order Management
		userRoutes.POST("/order/step1/placeorder", controllers.PlaceOrder)
		userRoutes.POST("/order/step2/initiatepayment", controllers.InitiatePayment)
		userRoutes.POST("/order/step3/razorpaycallback/:orderid", controllers.RazorPayGatewayCallback)
		userRoutes.GET("/order/step3/stripecallback", controllers.StripeCallback)
		userRoutes.POST("/order/cancel", controllers.CancelOrderedProduct)
		userRoutes.POST("/order/history", controllers.UserOrderHistory)
		userRoutes.GET("/order/invoice/:orderid", controllers.GetOrderInfoByOrderIDAndGeneratePDF)
		userRoutes.POST("/order/paymenthistory", controllers.PaymentDetailsByOrderID)
		userRoutes.POST("/order/review", controllers.UserReviewonOrderItem)
		userRoutes.POST("/order/rating", controllers.UserRatingOrderItem)
		userRoutes.GET("/coupon/all", controllers.GetAllCoupons)
		userRoutes.GET("/coupon/cart/:couponcode", controllers.ApplyCouponOnCart)

		// Referral System
		userRoutes.GET("/referral/code", controllers.GetRefferalCode)
		userRoutes.PATCH("/referral/activate", controllers.ActivateReferral)
		userRoutes.GET("/referral/claim", controllers.ClaimReferralRewards)
		userRoutes.GET("/referral/stats", controllers.GetReferralStats)
	}
}

func RestaurantRoutes(router *gin.Engine) {
	restaurantRoutes := router.Group("/api/v1/restaurants")
	{
		// Restaurant Management
		restaurantRoutes.POST("/edit", controllers.EditRestaurant)
		restaurantRoutes.POST("/products/add", controllers.AddProduct)
		restaurantRoutes.POST("/products/edit", controllers.EditProduct)
		restaurantRoutes.DELETE("/products/:productid", controllers.DeleteProduct)

		// Order History and Status Updates
		restaurantRoutes.POST("/order/history/", controllers.OrderHistoryRestaurants)
		restaurantRoutes.POST("/order/nextstatus", controllers.UpdateOrderStatusForRestaurant)

		// Product Offers
		restaurantRoutes.POST("/product/offer/add", controllers.AddProductOffer)
		restaurantRoutes.PATCH("/product/offer/remove/:productid", controllers.RemoveProductOffer)

		//restaurant wallet balance and history
		restaurantRoutes.GET("/wallet/all", controllers.GetRestaurantWalletData)
		// restaurantRoutes.POST("/profile/update",controllers.RestaurantProfileUpdate)
	}
}

func AdminRoutes(router *gin.Engine) {
	adminRoutes := router.Group("/api/v1/admin")
	{
		// User Management
		adminRoutes.GET("/users", controllers.GetUserList) 
		adminRoutes.GET("/users/blocked", controllers.GetBlockedUserList) //mark
		adminRoutes.PUT("/users/block/:userid", controllers.BlockUser) //mark
		adminRoutes.PUT("/users/unblock/:userid", controllers.UnblockUser) //mark

		// Category Management
		adminRoutes.POST("/categories/add", controllers.AddCategory) //mark
		adminRoutes.PATCH("/categories/edit", controllers.EditCategory)  //mark
		adminRoutes.DELETE("/categories/delete/:categoryid", controllers.DeleteCategory) //mark

		// Restaurant Management
		adminRoutes.GET("/restaurants", controllers.GetRestaurants) 
		// adminRoutes.DELETE("/restaurants/:restaurantid", controllers.DeleteRestaurant)
		adminRoutes.PUT("/restaurants/block/:restaurantid", controllers.BlockRestaurant) //mark
		adminRoutes.PUT("/restaurants/unblock/:restaurantid", controllers.UnblockRestaurant) //mark

		// Coupon Management
		adminRoutes.POST("/coupon/create", controllers.CreateCoupon)
		adminRoutes.POST("/coupon/update", controllers.UpdateCoupon)
	}
}

func PublicRoutes(router *gin.Engine) {
	// Public API Endpoints
	publicRoute := router.Group("/api/v1/public")
	{
		publicRoute.GET("/categories", controllers.GetCategoryList) //mark
		publicRoute.GET("/categories/products", controllers.GetCategoryProductList) //mark
		publicRoute.GET("/products", controllers.GetProductList) //mark 
		publicRoute.GET("/restaurants", controllers.GetRestaurants) //mark
		publicRoute.GET("/restaurants/products/:restaurantid", controllers.GetProductsByRestaurantID) //mark
		publicRoute.GET("/products/onlyveg", controllers.OnlyVegProducts)  //mark
		publicRoute.GET("/products/newarrivals", controllers.NewArrivals) //mark
		publicRoute.GET("/products/lowtohigh", controllers.PriceLowToHigh) //mark
		publicRoute.GET("/products/hightolow", controllers.PriceHighToLow) //mark
		publicRoute.GET("/products/offerproducts", controllers.GetProductOffers) //mark
		publicRoute.GET("/report/products/:productid", controllers.ProductReport) //mark
		publicRoute.GET("/report/products/best", controllers.BestSellingProducts) //mark
		publicRoute.GET("/report/totalorders/all", controllers.PlatformOverallSalesReport) //mark

	}
}

func AdditionalRoutes(router *gin.Engine) {
	// Additional Endpoints
	router.GET("/api/v1/uploadimage", view.LoadUpload) //mark
	router.POST("/api/v1/uploadimage", helper.ImageUpload) //mark
	router.GET("/api/v1/logout", controllers.Logout) //mark
}
