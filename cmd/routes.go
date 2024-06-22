package main

import (
	"foodbuddy/controllers"
	"foodbuddy/view"
	"foodbuddy/helper"

	"github.com/gin-gonic/gin"
)

func ServerHealth(router *gin.Engine)  {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})
}

func UserRoutes(router *gin.Engine) {
    userRoutes := router.Group("/api/v1/user")
    {
        // User Profile Management
        userRoutes.GET("/profile", controllers.GetUserProfile)
        userRoutes.POST("/edit", controllers.UpdateUserInformation)
        userRoutes.GET("/wallet/balance", controllers.UserWalletBalance)

        // Favorite Products
        userRoutes.GET("/favorites/all", controllers.GetUsersFavouriteProduct)
        userRoutes.POST("/favorites/", controllers.AddFavouriteProduct)
        userRoutes.DELETE("/favorites/", controllers.RemoveFavouriteProduct)

        // User Address Management
        userRoutes.GET("/address/all", controllers.GetUserAddress)
        userRoutes.POST("/address/add", controllers.AddUserAddress)
        userRoutes.PUT("/address/edit", controllers.EditUserAddress)
        userRoutes.DELETE("/address/delete", controllers.DeleteUserAddress)

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
        userRoutes.POST("/order/information", controllers.GetOrderInfoByOrderIDAndGeneratePDF)
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
		restaurantRoutes.GET("/wallet/all",controllers.GetRestaurantWalletData)
    }
}

func AdminRoutes(router *gin.Engine) {
    adminRoutes := router.Group("/api/v1/admin")
    {
        // User Management
        adminRoutes.GET("/users", controllers.GetUserList)
        adminRoutes.GET("/users/blocked", controllers.GetBlockedUserList)
        adminRoutes.GET("/users/block/:userid", controllers.BlockUser)
        adminRoutes.GET("/users/unblock/:userid", controllers.UnblockUser)

        // Category Management
        adminRoutes.POST("/categories/add", controllers.AddCategory)
        adminRoutes.PUT("/categories/:categoryid", controllers.EditCategory)
        adminRoutes.DELETE("/categories/:categoryid", controllers.DeleteCategory)

        // Restaurant Management
        adminRoutes.GET("/restaurants", controllers.GetRestaurants)
        adminRoutes.DELETE("/restaurants/:restaurantid", controllers.DeleteRestaurant)
        adminRoutes.PUT("/restaurants/block/:restaurantid", controllers.BlockRestaurant)
        adminRoutes.PUT("/restaurants/unblock/:restaurantid", controllers.UnblockRestaurant)

        // Coupon Management
        adminRoutes.POST("/coupon/create", controllers.CreateCoupon)
        adminRoutes.POST("/coupon/update", controllers.UpdateCoupon)
    }
}

func PublicRoutes(router *gin.Engine) {
    // Public API Endpoints
    router.GET("/api/v1/public/categories", controllers.GetCategoryList)
    router.GET("/api/v1/public/categories/products", controllers.GetCategoryProductList)
    router.GET("/api/v1/public/products", controllers.GetProductList)
    router.GET("/api/v1/public/restaurants", controllers.GetRestaurants)
    router.GET("/api/v1/public/restaurants/products/:restaurantid", controllers.GetProductsByRestaurantID)
    router.GET("/api/v1/public/products/onlyveg", controllers.OnlyVegProducts)
    router.GET("/api/v1/public/products/newarrivals", controllers.NewArrivals)
    router.GET("/api/v1/public/product/lowtohigh", controllers.PriceLowToHigh)
    router.GET("/api/v1/public/product/hightolow", controllers.PriceHighToLow)
    router.GET("/api/v1/public/product/offerproducts", controllers.GetProductOffers)
    router.GET("/api/v1/public/report/product/:productid", controllers.ProductReport)
    router.GET("/api/v1/public/report/product/best/", controllers.BestSellingProducts)
}

func AuthenticationRoutes(router *gin.Engine) {
    // Authentication Endpoints
	//admin
    router.POST("/api/v1/auth/admin/login", controllers.AdminLogin)

	//user
    router.POST("/api/v1/auth/user/email/login", controllers.EmailLogin)
    router.POST("/api/v1/auth/user/email/signup", controllers.EmailSignup)
	router.GET("/api/v1/auth/google/login", controllers.GoogleHandleLogin)
    router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)

	//additional endpoints for email verification and password reset
    router.GET("/api/v1/auth/verifyotp/:role/:email/:otp", controllers.VerifyOTP)
	
    router.POST("/api/v1/auth/passwordreset/step1", controllers.Step1PasswordReset)
    router.GET("/api/v1/auth/passwordreset", controllers.LoadPasswordReset)
    router.POST("/api/v1/auth/passwordreset/step2", controllers.Step2PasswordReset)

	//restaurant
    router.POST("/api/v1/auth/restaurant/signup", controllers.RestaurantSignup)
    router.POST("/api/v1/auth/restaurant/login", controllers.RestaurantLogin)
}

func AdditionalRoutes(router *gin.Engine) {
    // Additional Endpoints
    router.GET("/api/v1/uploadimage", view.LoadUpload)
    router.POST("/api/v1/uploadimage", helper.ImageUpload)
    router.GET("/api/v1/logout", controllers.Logout)
}
