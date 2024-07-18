package api

import (
	"foodbuddy/internal/controllers"
	"foodbuddy/view"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServerHealth(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})
}
func AuthenticationRoutes(router *gin.Engine) {
	//admin
	router.GET("/api/v1/auth/admin/login", controllers.AdminLogin) //

	//user
	router.POST("/api/v1/auth/user/email/login", controllers.EmailLogin)   //
	router.POST("/api/v1/auth/user/email/signup", controllers.EmailSignup) //
	router.GET("/api/v1/auth/google/login", controllers.GoogleHandleLogin) //
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback) //

	//additional endpoints for email verification and password reset
	router.GET("/api/v1/auth/verifyemail/:role/:email/:otp", controllers.VerifyEmail) //

	router.POST("/api/v1/auth/passwordreset/step1", controllers.Step1PasswordReset) //
	router.GET("/api/v1/auth/passwordreset", controllers.LoadPasswordReset)         //
	router.POST("/api/v1/auth/passwordreset/step2", controllers.Step2PasswordReset) //

	//restaurant
	router.POST("/api/v1/auth/restaurant/signup", controllers.RestaurantSignup) //
	router.POST("/api/v1/auth/restaurant/login", controllers.RestaurantLogin)   //
}

func UserRoutes(router *gin.Engine) {
	userRoutes := router.Group("/api/v1/user")
	{
		// User Profile Management
		userRoutes.GET("/profile", controllers.GetUserProfile)       //
		userRoutes.POST("/edit", controllers.UpdateUserInformation)  //
		userRoutes.GET("/wallet/all", controllers.GetUserWalletData) //

		// Favorite Products
		userRoutes.GET("/favorites/all", controllers.GetUsersFavouriteProduct)     //
		userRoutes.POST("/favorites/add", controllers.AddFavouriteProduct)         //
		userRoutes.DELETE("/favorites/delete", controllers.RemoveFavouriteProduct) //

		// User Address Management
		userRoutes.GET("/address/all", controllers.GetUserAddress)          //
		userRoutes.POST("/address/add", controllers.AddUserAddress)         //
		userRoutes.PATCH("/address/edit", controllers.EditUserAddress)      //
		userRoutes.DELETE("/address/delete", controllers.DeleteUserAddress) //

		// Cart Management
		userRoutes.POST("/cart/add", controllers.AddToCart) //
		userRoutes.POST("/cart/cookingrequest", controllers.AddCookingRequest)
		userRoutes.GET("/cart/all", controllers.GetCartTotal)             //
		userRoutes.DELETE("/cart/delete/", controllers.ClearCart)         //specify restaurant_id for clearing individual carts
		userRoutes.DELETE("/cart/remove", controllers.RemoveItemFromCart) //
		userRoutes.PUT("/cart/update/", controllers.UpdateQuantity)       //
		userRoutes.GET("/coupon/cart/", controllers.ApplyCouponOnCart)    //

		// Order Management
		userRoutes.POST("/order/step1/placeorder", controllers.PlaceOrder)
		userRoutes.GET("/order/deliverycode", controllers.SendOrderDeliveryVerificationCode)
		userRoutes.POST("/order/step2/initiatepayment", controllers.InitiatePayment)
		userRoutes.PUT("/order/update/paymentmode", controllers.ChangeOrderPaymentMode) //orderid in the query param //CHANGE COD , ONLINE MODE
		userRoutes.POST("/order/step3/razorpaycallback/:orderid", controllers.RazorPayGatewayCallback)
		userRoutes.GET("/order/step3/razorpaycallback/failed/:orderid", controllers.RazorPayFailed)
		userRoutes.GET("/order/step3/stripecallback", controllers.StripeCallback)
		userRoutes.POST("/order/cancel/online", controllers.CancelOrderedProductOnline)
		userRoutes.POST("/order/cancel/cod", controllers.CancelOrderedProductCOD)
		userRoutes.GET("/order/items", controllers.UserOrderItems)
		userRoutes.GET("/order/info", controllers.GetOrderInfoByOrderIDasJSON)
		userRoutes.GET("/order/invoice/", controllers.GetOrderInfoByOrderIDAndGeneratePDF)
		userRoutes.GET("/order/paymenthistory", controllers.PaymentDetailsByOrderID)
		userRoutes.GET("/order/verifypayment", controllers.VerifyOnlinePayment)
		userRoutes.POST("/order/review", controllers.UserReviewonOrderItem)
		userRoutes.POST("/order/rating", controllers.UserRatingOrderItem)

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
		restaurantRoutes.POST("/edit", controllers.EditRestaurant)       //update restaurant profile
		restaurantRoutes.POST("/products/add", controllers.AddProduct)   //
		restaurantRoutes.POST("/products/edit", controllers.EditProduct) //
		restaurantRoutes.DELETE("/products", controllers.DeleteProduct)  //

		// Order History and Status Updates
		restaurantRoutes.GET("/order/history", controllers.OrderHistoryRestaurants)            //without order_status and with
		restaurantRoutes.POST("/order/confirmcod", controllers.ConfirmCODPayment)              //authentication for rest add rest id in the order
		restaurantRoutes.POST("/order/confirmdelivery", controllers.DeliveryComplete)          //query param order_id,authentication
		restaurantRoutes.POST("/order/nextstatus", controllers.UpdateOrderStatusForRestaurant) //authentication rest

		// Product Offers
		restaurantRoutes.POST("/product/offer/add", controllers.AddProductOffer)      //
		restaurantRoutes.PUT("/product/offer/remove", controllers.RemoveProductOffer) //

		//orderitem information in excel
		restaurantRoutes.GET("/orderitems/excel/all", controllers.OrderItemsCSVFileForRestaurant) //
		restaurantRoutes.GET("/orderitems/json/all", controllers.ListOrderItemsForRestaurants)    //

		//report
		restaurantRoutes.GET("/report/all", controllers.RestaurantOverallSalesReport)
		//new customers this week

		//restaurant wallet balance and history
		restaurantRoutes.GET("/wallet/all", controllers.GetRestaurantWalletData) //
	}
}

func AdminRoutes(router *gin.Engine) {
	adminRoutes := router.Group("/api/v1/admin")
	{
		// User Management
		//get profile info , update online stats
		adminRoutes.GET("/users", controllers.GetUserList)                //
		adminRoutes.GET("/users/blocked", controllers.GetBlockedUserList) //
		adminRoutes.PUT("/users/block/", controllers.BlockUser)           //
		adminRoutes.PUT("/users/unblock", controllers.UnblockUser)        //

		// Category Management
		adminRoutes.POST("/categories/add", controllers.AddCategory)         //
		adminRoutes.PATCH("/categories/edit", controllers.EditCategory)      //
		adminRoutes.DELETE("/categories/delete", controllers.DeleteCategory) //

		// Restaurant Management
		adminRoutes.GET("/restaurants", controllers.GetRestaurants)
		adminRoutes.PUT("/restaurants/block", controllers.BlockRestaurant)     //
		adminRoutes.PUT("/restaurants/unblock", controllers.UnblockRestaurant) //
		adminRoutes.PUT("/restaurants/verify/success", controllers.VerifyRestaurant)
		adminRoutes.PUT("/restaurants/verify/failed", controllers.RemoveVerifyStatusRestaurant)

		// Coupon Management
		adminRoutes.POST("/coupon/create", controllers.CreateCoupon)  //
		adminRoutes.PATCH("/coupon/update", controllers.UpdateCoupon) //
	}
}

func PublicRoutes(router *gin.Engine) {
	// Public API Endpoints
	publicRoute := router.Group("/api/v1/public")
	{
		//get restaurant profile info
		publicRoute.GET("/restaurant/profile", controllers.GetRestaurantProfile)
		publicRoute.GET("/coupon/all", controllers.GetAllCoupons)                   //
		publicRoute.GET("/categories", controllers.GetCategoryList)                 //
		publicRoute.GET("/categories/products", controllers.GetCategoryProductList) //
		publicRoute.GET("/products", controllers.GetProductList)                    //
		publicRoute.GET("/product/reviewandrating", controllers.ListAllReviewsandRating)
		publicRoute.GET("/restaurants", controllers.GetRestaurants)                          //
		publicRoute.GET("/restaurants/products/", controllers.GetProductsByRestaurantID)     //
		publicRoute.GET("/products/onlyveg", controllers.OnlyVegProducts)                    //
		publicRoute.GET("/products/newarrivals", controllers.NewArrivals)                    //
		publicRoute.GET("/products/lowtohigh", controllers.PriceLowToHigh)                   //
		publicRoute.GET("/products/hightolow", controllers.PriceHighToLow)                   //
		publicRoute.GET("/products/offerproducts", controllers.GetProductOffers)             //
		publicRoute.GET("/report/products", controllers.ProductReport)                       //
		publicRoute.GET("/report/products/best", controllers.BestSellingProducts)            //
		publicRoute.GET("/report/overallreport/all", controllers.PlatformOverallSalesReport) //

	}
}

func AdditionalRoutes(router *gin.Engine) {
	// Additional Endpoints
	router.GET("/api/v1/documentation", APIDocumentation)
	router.GET("/api/v1/user/profileimage", view.LoadUpload)                                 //
	router.POST("/api/v1/user/profileimage", controllers.UserProfileImageUpload)             //
	router.GET("/api/v1/restaurant/profileimage", view.LoadUpload)                           //
	router.POST("/api/v1/restaurant/profileimage", controllers.RestaurantProfileImageUpload) //
	router.GET("/api/v1/logout", controllers.Logout)                                         //
}

func APIDocumentation(c *gin.Context) {
	url := "https://documenter.getpostman.com/view/32055383/2sA3e488Sh"
	c.Redirect(http.StatusMovedPermanently, url)
}
