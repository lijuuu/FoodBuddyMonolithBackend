package controllers

// AdminLogin godoc
// @Summary Admin login
// @Description Login an admin using email
// @Tags authentication
// @Accept json
// @Produce json
// @Param AdminLogin body model.AdminLoginRequest true "Admin Login"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/admin/login [post]

// EmailLogin godoc
// @Summary Email login
// @Description Login a user using email
// @Tags authentication
// @Accept json
// @Produce json
// @Param EmailLogin body model.EmailLoginRequest true "Email Login"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/user/email/login [post]

// EmailSignup godoc
// @Summary Email signup
// @Description Signup a new user using email
// @Tags authentication
// @Accept json
// @Produce json
// @Param EmailSignup body model.EmailSignupRequest true "Email Signup"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/user/email/signup [post]

// VerifyOTP godoc
// @Summary Verify OTP
// @Description Verify OTP for email verification
// @Tags authentication
// @Accept json
// @Produce json
// @Param role path string true "User role"
// @Param email path string true "User email"
// @Param otp path string true "OTP"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/verifyotp/{role}/{email}/{otp} [get]

// GoogleHandleLogin godoc
// @Summary Google login
// @Description Login using Google authentication
// @Tags authentication
// @Produce json
// @Success 200 {object} model.SuccessResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/google/login [get]

// GoogleHandleCallback godoc
// @Summary Google callback
// @Description Callback handler for Google authentication
// @Tags authentication
// @Produce json
// @Success 200 {object} model.SuccessResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/google/callback [get]

// RestaurantSignup godoc
// @Summary Restaurant signup
// @Description Signup a new restaurant
// @Tags authentication
// @Accept json
// @Produce json
// @Param RestaurantSignup body model.RestaurantSignupRequest true "Restaurant Signup"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/restaurant/signup [post]

// RestaurantLogin godoc
// @Summary Restaurant login
// @Description Login a restaurant
// @Tags authentication
// @Accept json
// @Produce json
// @Param RestaurantLogin body model.RestaurantLoginRequest true "Restaurant Login"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/restaurant/login [post]

// UpdateUserInformation godoc
// @Summary Update user information
// @Description Update user information
// @Tags user
// @Accept json
// @Produce json
// @Param UpdateUserInformation body model.UpdateUserInformationRequest true "Update User Information"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/update [post]
// @Router /api/v1/user/UpdateUserInformation [post]

// GetCategoryList godoc
// @Summary Get all categories
// @Description Get a list of all categories
// @Tags public
// @Produce json
// @Success 200 {object} model.CategoryListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/public/categories [get]

// GetCategoryProductList godoc
// @Summary Get products by category
// @Description Get products by category ID
// @Tags public
// @Produce json
// @Param categoryid path string true "Category ID"
// @Success 200 {object} model.ProductListResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/public/categories/{categoryid}/products [get]

// GetProductList godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags public
// @Produce json
// @Success 200 {object} model.ProductListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/public/products [get]

// GetRestaurants godoc
// @Summary Get all restaurants
// @Description Get a list of all restaurants
// @Tags public
// @Produce json
// @Success 200 {object} model.RestaurantListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/public/restaurants [get]

// GetProductsByRestaurantID godoc
// @Summary Get products by restaurant ID
// @Description Get products by restaurant ID
// @Tags public
// @Produce json
// @Param restaurantid path string true "Restaurant ID"
// @Success 200 {object} model.ProductListResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/public/restaurants/{restaurantid}/products [get]

// GetUserList godoc
// @Summary Get all users
// @Description Get a list of all users
// @Tags admin
// @Produce json
// @Success 200 {object} model.UserListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/users [get]

// GetBlockedUserList godoc
// @Summary Get blocked users
// @Description Get a list of blocked users
// @Tags admin
// @Produce json
// @Success 200 {object} model.BlockedUserListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/users/blocked [get]

// BlockUser godoc
// @Summary Block a user
// @Description Block a user by user ID
// @Tags admin
// @Produce json
// @Param userid path string true "User ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/users/block/{userid} [post]

// UnblockUser godoc
// @Summary Unblock a user
// @Description Unblock a user by user ID
// @Tags admin
// @Produce json
// @Param userid path string true "User ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/users/unblock/{userid} [post]

// GetCategoryList godoc
// @Summary Get all categories
// @Description Get a list of all categories
// @Tags admin
// @Produce json
// @Success 200 {object} model.CategoryListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/categories [get]

// AddCategory godoc
// @Summary Add a new category
// @Description Add a new category
// @Tags admin
// @Accept json
// @Produce json
// @Param AddCategory body model.AddCategoryRequest true "Add Category"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/categories [post]

// EditCategory godoc
// @Summary Edit a category
// @Description Edit a category by category ID
// @Tags admin
// @Accept json
// @Produce json
// @Param categoryid path string true "Category ID"
// @Param EditCategory body model.EditCategoryRequest true "Edit Category"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/categories/{categoryid} [put]

// DeleteCategory godoc
// @Summary Delete a category
// @Description Deletea category by category ID
// @Tags admin
// @Produce json
// @Param categoryid path string true "Category ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/categories/{categoryid} [delete]

// GetRestaurants godoc
// @Summary Get all restaurants
// @Description Get a list of all restaurants
// @Tags admin
// @Produce json
// @Success 200 {object} model.RestaurantListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/restaurants [get]

// EditRestaurant godoc
// @Summary Edit a restaurant
// @Description Edit a restaurant by restaurant ID
// @Tags admin
// @Accept json
// @Produce json
// @Param restaurantid path string true "Restaurant ID"
// @Param EditRestaurant body model.EditRestaurantRequest true "Edit Restaurant"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/restaurants/{restaurantid} [put]

// DeleteRestaurant godoc
// @Summary Delete a restaurant
// @Description Delete a restaurant by restaurant ID
// @Tags admin
// @Produce json
// @Param restaurantid path string true "Restaurant ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/restaurants/{restaurantid} [delete]

// BlockRestaurant godoc
// @Summary Block a restaurant
// @Description Block a restaurant by restaurant ID
// @Tags admin
// @Produce json
// @Param restaurantid path string true "Restaurant ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/restaurants/block/{restaurantid} [post]

// UnblockRestaurant godoc
// @Summary Unblock a restaurant
// @Description Unblock a restaurant by restaurant ID
// @Tags admin
// @Produce json
// @Param restaurantid path string true "Restaurant ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/restaurants/unblock/{restaurantid} [post]

// GetProductList godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags admin
// @Produce json
// @Success 200 {object} model.ProductListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/products [get]

// AddProduct godoc
// @Summary Add a new product
// @Description Add a new product
// @Tags admin
// @Accept json
// @Produce json
// @Param AddProduct body model.AddProductRequest true "Add Product"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/products [post]

// EditProduct godoc
// @Summary Edit a product
// @Description Edit a product by product ID
// @Tags admin
// @Accept json
// @Produce json
// @Param productid path string true "Product ID"
// @Param EditProduct body model.EditProductRequest true "Edit Product"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/products/{productid} [put]

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by product ID
// @Tags admin
// @Produce json
// @Param productid path string true "Product ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/admin/products/{productid} [delete]

// GetProductList godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags restaurant
// @Produce json
// @Success 200 {object} model.ProductListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/restaurant/products [get]

// AddProduct godoc
// @Summary Add a new product
// @Description Add a new product
// @Tags restaurant
// @Accept json
// @Produce json
// @Param AddProduct body model.AddProductRequest true "Add Product"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/restaurant/products [post]

// EditProduct godoc
// @Summary Edit a product
// @Description Edit a product by product ID
// @Tags restaurant
// @Accept json
// @Produce json
// @Param productid path string true "Product ID"
// @Param EditProduct body model.EditProductRequest true "Edit Product"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/restaurant/products/{productid} [put]

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by product ID
// @Tags restaurant
// @Produce json
// @Param productid path string true "Product ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/restaurant/products/{productid} [delete]

// GetFavouriteProductByUserID godoc
// @Summary Get favorite products by user ID
// @Description Get favorite products by user ID
// @Tags user
// @Produce json
// @Param userid path string true "User ID"
// @Success 200 {object} model.FavoriteProductListResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/{userid}/favorites [get]

// AddFavouriteProduct godoc
// @Summary Add a favorite product
// @Description Add a product to favorites by user ID
// @Tags user
// @Accept json
// @Produce json
// @Param userid path string true "User ID"
// @Param AddFavouriteProduct body model.AddFavouriteProductRequest true "Add Favorite Product"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/{userid}/favorites [post]

// RemoveFavouriteProduct godoc
// @Summary Remove a favorite product
// @Description Remove a product from favorites by user ID
// @Tags user
// @Produce json
// @Param userid path string true "User ID"
// @Param productid query string true "Product ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/{userid}/favorites [delete]

// AddUserAddress godoc
// @Summary Add a user address
// @Description Add a new address for a user
// @Tags user
// @Accept json
// @Produce json
// @Param userid path string true "User ID"
// @Param AddUserAddress body model.AddUserAddressRequest true "Add User Address"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/{userid}/address [post]

// GetUserAddress godoc
// @Summary Get user address
// @Description Get address of a user by user ID
// @Tags user
// @Produce json
// @Param userid path string true "User ID"
// @Success 200 {object} model.UserAddressResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/{userid}/address [get]

// EditUserAddress godoc
// @Summary Edit a user address
// @Description Edit an address of a user by user ID and address ID
// @Tags user
// @Accept json
// @Produce json
// @Param userid path string true "User ID"
// @Param addressid path string true "Address ID"
// @Param EditUserAddress body model.EditUserAddressRequest true "Edit User Address"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/{userid}/address/{addressid} [put]

// DeleteUserAddress godoc
// @Summary Delete a user address
// @Description Delete an address of a user by user ID and address ID
// @Tags user
// @Produce json
// @Param userid path string true "User ID"
// @Param addressid path string true "Address ID"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/user/{userid}/address/{addressid} [delete]

// LoadUpload godoc
// @Summary Load image upload page
// @Description Load the page for uploading images
// @Tags image
// @Produce html
// @Router /api/v1/uploadimage [get]

// ImageUpload godoc
// @Summary Upload an image
// @Description Upload an image
// @Tags image
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file to upload"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/uploadimage [post]

// Logout godoc
// @Summary Logout
// @Description Logout the user or admin
// @Tags authentication
// @Produce json
// @Success 200 {object} model.SuccessResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/logout [get]







