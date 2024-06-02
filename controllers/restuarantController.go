package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
func RestaurantSignup(c *gin.Context) {
	// bind json to struct
	var restaurantSignup model.RestaurantSignupRequest
	if err := c.BindJSON(&restaurantSignup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process the request",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// validate input
	if err := validate(&restaurantSignup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// check if email exists
	var verification model.VerificationTable
	tx := database.DB.Where("email = ? AND role = ?", restaurantSignup.Email, model.RestaurantRole).First(&verification)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "database error",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	} else if tx.Error == gorm.ErrRecordNotFound {
		// create new entry in verification table
		verification = model.VerificationTable{
			Email:              restaurantSignup.Email,
			Role:               model.RestaurantRole,
			VerificationStatus: model.VerificationStatusPending,
		}
		tx = database.DB.Create(&verification)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":     false,
				"message":    "failed to create restaurant verification entry",
				"error_code": http.StatusInternalServerError,
				"data":       gin.H{},
			})
			return
		}
	} else {
		// email already exists
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "restaurant email already exists",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// generate salt and hash password
	salt := utils.GenerateRandomString(7)
	saltedPassword := salt + restaurantSignup.Password
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to process the request",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	// create restaurant record
	restaurant := model.Restaurant{
		Name:               restaurantSignup.Name,
		Description:        restaurantSignup.Description,
		Address:            restaurantSignup.Address,
		Email:              restaurantSignup.Email,
		PhoneNumber:        restaurantSignup.PhoneNumber,
		ImageURL:           restaurantSignup.ImageURL,
		CertificateURL:     restaurantSignup.CertificateURL,
		VerificationStatus: model.VerificationStatusPending,
		Blocked:            false,
		Salt:               salt,
		HashedPassword:     string(hash),
	}

	// save to database
	if err := database.DB.Create(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to save restaurant data",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	// respond with success
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "restaurant signup successful",
		"data": gin.H{
			"name":                restaurant.Name,
			"description":         restaurant.Description,
			"address":             restaurant.Address,
			"email":               restaurant.Email,
			"phone_number":        restaurant.PhoneNumber,
			"image_url":           restaurant.ImageURL,
			"certificate_url":     restaurant.CertificateURL,
			"verification_status": restaurant.VerificationStatus,
			"blocked":             restaurant.Blocked,
		},
	})
}

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
func RestaurantLogin(c *gin.Context) {
	// Get struct
	var restaurantLogin model.RestaurantLoginRequest
	var existingRestaurant model.Restaurant

	if err := c.BindJSON(&restaurantLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "Failed to process the incoming request",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Validate
	if err := validate(&restaurantLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Check email on restaurant DB
	if err := database.DB.Where("email = ?", restaurantLogin.Email).First(&existingRestaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Error fetching restaurant details",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	// Check block and admin verification status
	if existingRestaurant.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "Restaurant not authorized to access the route",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
		})
		return
	}

	// Check password by salt and password
	password := []byte(existingRestaurant.Salt + restaurantLogin.Password)
	if err := bcrypt.CompareHashAndPassword([]byte(existingRestaurant.HashedPassword), password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "Invalid credentials",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
		})
		return
	}

	// Check email verification status using verification table
	var verificationTable model.VerificationTable
	if err := database.DB.Where("email = ?", restaurantLogin.Email).First(&verificationTable).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to fetch email verification status",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	if verificationTable.VerificationStatus != model.VerificationStatusVerified {
		if err := SendOTP(c, restaurantLogin.Email, verificationTable.OTPExpiry, model.RestaurantRole); err != nil {
			c.JSON(http.StatusAlreadyReported, gin.H{
				"status":     false,
				"message":    err.Error(),
				"error_code": http.StatusAlreadyReported,
				"data":       gin.H{},
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "Please verify your email to continue",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
		})
		return
	}

	token, err := GenerateJWT(c, existingRestaurant.Email, model.RestaurantRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "Failed to generate token",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Login is successful",
		"data": gin.H{
			"Name":               existingRestaurant.Name,
			"Email":              existingRestaurant.Email,
			"Description":        existingRestaurant.Description,
			"Address":            existingRestaurant.Address,
			"PhoneNumber":        existingRestaurant.PhoneNumber,
			"ImageURL":           existingRestaurant.ImageURL,
			"CertificateURL":     existingRestaurant.CertificateURL,
			"VerificationStatus": existingRestaurant.VerificationStatus,
		},
		"token": token,
	})
}

func CheckRestaurant(c *gin.Context) {
	email := utils.GetJWTEmailClaim(c)
	if email == "" {
		c.Abort()
		return
	}

	if err := VerifyJWT(c, model.RestaurantRole, email); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "access denied, request is unauthorized",
			"error_code": http.StatusUnauthorized,
			"data":       gin.H{},
		})
		c.Abort()
		return
	}
	c.Next()
}

func GetRestaurants(c *gin.Context) {
	var restaurants []model.Restaurant
	// Search db and get all
	if err := database.DB.Select("*").Find(&restaurants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to retrieve data from the database",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "restaurants retrieved successfully",
		"data":    gin.H{"restaurantslist": restaurants},
	})
}

func EditRestaurant(c *gin.Context) {
	// Bind JSON
	var restaurant model.Restaurant
	if err := c.BindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	if err := validate(restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Check if present and update it with the new data
	var existingRestaurant model.Restaurant
	if err := database.DB.First(&existingRestaurant, restaurant.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "restaurant doesn't exist",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
		})
		return
	}

	// Edit the restaurant
	if err := database.DB.Model(&existingRestaurant).Updates(restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to edit the restaurant",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully edited the restaurant",
		"data":    gin.H{},
	})
}

func DeleteRestaurant(c *gin.Context) {
	// Get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "invalid restaurant ID",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Check if it's already present
	var existingRestaurant model.Restaurant
	if err := database.DB.First(&existingRestaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "restaurant doesn't exist",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
		})
		return
	}

	// Delete it
	if err := database.DB.Delete(&existingRestaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to delete the restaurant",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully deleted the restaurant",
		"data":    gin.H{},
	})
}

func BlockRestaurant(c *gin.Context) {
	// Get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "invalid restaurant ID",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Check restaurant by id
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "restaurant not found",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
		})
		return
	}

	if restaurant.Blocked {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "restaurant is already blocked",
			"error_code": http.StatusConflict,
			"data":       gin.H{"restaurant": restaurant},
		})
		return
	}

	// Set blocked as true
	restaurant.Blocked = true

	if err := database.DB.Save(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to change the block status",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "restaurant is blocked",
		"data":    gin.H{"restaurant": restaurant},
	})
}

func UnblockRestaurant(c *gin.Context) {
	// Get the restaurant id
	restaurantIDStr := c.Param("restaurantid")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "invalid restaurant ID",
			"error_code": http.StatusBadRequest,
			"data":       gin.H{},
		})
		return
	}

	// Check restaurant by id
	var restaurant model.Restaurant
	if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "restaurant not found",
			"error_code": http.StatusNotFound,
			"data":       gin.H{},
		})
		return
	}

	if !restaurant.Blocked {
		c.JSON(http.StatusConflict, gin.H{
			"status":     false,
			"message":    "restaurant is already unblocked",
			"error_code": http.StatusConflict,
			"data":       gin.H{"restaurant": restaurant},
		})
		return
	}

	// Set blocked as false
	restaurant.Blocked = false

	if err := database.DB.Save(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to change the block status",
			"error_code": http.StatusInternalServerError,
			"data":       gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "restaurant is unblocked",
		"data":    gin.H{"restaurant": restaurant},
	})
}
