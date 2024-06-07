package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"io"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/api/v1/googlecallback",
	ClientID:     utils.GetEnvVariables().ClientID,
	ClientSecret: utils.GetEnvVariables().ClientSecret,
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

// GoogleHandleLogin godoc
// @Summary Google login
// @Description Login using Google authentication
// @Tags authentication
// @Produce json
// @Success 200 {object} model.SuccessResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/google/login [get]
func GoogleHandleLogin(c *gin.Context) {
	utils.NoCache(c)
	url := googleOauthConfig.AuthCodeURL("hjdfyuhadVFYU6781235")
	c.Redirect(http.StatusTemporaryRedirect, url)
	c.Next()
}

func GoogleHandleCallback(c *gin.Context) {
	utils.NoCache(c)
	fmt.Println("Starting to handle callback")
	code := c.Query("code")

	//check for code defined on googlehandlelogin still exists
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "missing code parameter",
			"data":    gin.H{},
		})
		return
	}

	//exchange code for token, code is exchanged to make sure the state is same
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to exchange token",
			"data":    gin.H{},
		})
		return
	}

	//use access token and get reponse of the user
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("google signup done")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to get user information",
			"data":    gin.H{},
		})
		return
	}
	defer response.Body.Close()

	//read the content of the reponse.body
	content, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to read user information",
			"data":    gin.H{},
		})
		return
	}

	//store the content from the json to the user struct of model.GoogleResponse
	var User model.GoogleResponse
	err = json.Unmarshal(content, &User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to parse user information",
			"data":    gin.H{},
		})
		return
	}

	//pass the values needed from the google response to the newuser struct
	newUser := model.User{
		Name:        User.Name,
		Email:       User.Email,
		LoginMethod: model.GoogleSSOMethod,
		Picture:     User.Picture,
		Blocked:     false,
	}

	//if no name is present on the response use the email as the name
	if newUser.Name == "" {
		newUser.Name = User.Email
	}

	// Check if the user already exists
	var existingUser model.User
	if err := database.DB.Where("email =? AND deleted_at IS NULL", newUser.Email).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a new user
			if err := database.DB.Create(&newUser).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  false,
					"message": "failed to create user through google sso",
					"data":    gin.H{},
				})
				return
			}
		} else {
			// Handle case where user already exists but not found due to other errors
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "failed to fetch user information",
				"data":    gin.H{},
			})
			return
		}
	}

	// User already exists, check login method
	if existingUser.LoginMethod == model.EmailLoginMethod {
		c.JSON(http.StatusSeeOther, gin.H{
			"status":  false,
			"message": "please login through email method",
			"data":    gin.H{},
		})
		return
	}

	//check is the user is blocked by the admin
	if existingUser.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "user is unauthorized to access",
			"data":    gin.H{},
		})
		return
	}

	// Generate JWT and set cookie within GenerateJWT
	tokenstring, err := GenerateJWT(c, newUser.Email, model.UserRole)
	if tokenstring == "" || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "failed to create authorization token",
			"data":    gin.H{},
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "login is successful",
		"data": gin.H{
			"user":  User,
			"token": tokenstring,
		},
	})

}

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
func EmailSignup(c *gin.Context) {

	utils.NoCache(c)

	//get the body
	var EmailSignupRequest model.EmailSignupRequest

	if err := c.BindJSON(&EmailSignupRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process the incoming request",
			"data":    gin.H{},
		})
		return
	}

	err := validate(EmailSignupRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
			"data":    gin.H{},
		})
		return
	}

	//check if the password and the confirm password is correct
	if EmailSignupRequest.Password != EmailSignupRequest.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "passwords doesn't match",
			"data":    gin.H{},
		})
		return
	}

	//create salt and add it to the password
	Salt := utils.GenerateRandomString(7)
	//salt+password
	saltedPassword := Salt + EmailSignupRequest.Password

	//hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to hash the password",
			"data":    gin.H{},
		})
		return
	}

	//add the data to user struct
	User := model.User{
		Name:           EmailSignupRequest.Name,
		Email:          EmailSignupRequest.Email,
		HashedPassword: string(hash),
		LoginMethod:    model.EmailLoginMethod,
		Blocked:        false,
		Salt:           Salt,
	}

	//check if the user exists on the database
	tx := database.DB.Where("email =? AND deleted_at IS NULL", EmailSignupRequest.Email).First(&User)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retreive information from the database",
			"data":    gin.H{},
		})
		return

	} else if tx.Error == gorm.ErrRecordNotFound {
		// User does not exist, proceed to create
		tx = database.DB.Create(&User)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "failed to create a new user",
				"data":    gin.H{},
			})
			return
		}
	} else {
		// User already exists
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "user already exists",
			"data":    gin.H{},
		})
		return
	}

	//update otp on the otp table along with user email, role, verification status
	VerificationTable := model.VerificationTable{
		Email:              User.Email,
		Role:               model.UserRole,
		VerificationStatus: model.VerificationStatusPending,
	}

	if err := database.DB.Create(&VerificationTable).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to process otp verification process",
			"data":    gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Email login successful, please login to complete your email verification",
		"data": gin.H{
			"user": gin.H{
				"name":         User.Name,
				"email":        User.Email,
				"phone_number": User.PhoneNumber,
				"picture":      User.Picture,
				"login_method": User.LoginMethod,
				"block_status": User.Blocked,
			},
		},
	})
	c.Next()
}

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
func EmailLogin(c *gin.Context) {
	var EmailLoginRequest model.EmailLoginRequest
	//get the json from the request
	if err := c.BindJSON(&EmailLoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process the incoming request",
			"data":    gin.H{},
		})
		return
	}

	//validate the content of the json
	err := validate(EmailLoginRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
			"data":    gin.H{},
		})
		return
	}

	//chekc whether the email exist on the database, if not return an error
	var user model.User
	tx := database.DB.Where("email =? AND deleted_at IS NULL", EmailLoginRequest.Email).First(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid email or password",
			"data":    gin.H{},
		})
		return
	}

	//check if the login methods are the same as email, if google prompt to use google login
	if user.LoginMethod != model.EmailLoginMethod {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "email uses another method for logging in, use google sso",
			"data":    gin.H{},
		})
		return
	}

	//check is the user is blocked by the admin
	if user.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "user is not authorized to access",
			"data":    gin.H{},
		})
		return
	}

	// password with salt = user.salt + EmailLoginRequest.password
	saltedPassword := user.Salt + EmailLoginRequest.Password

	//get the hash and compare it with password from body
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(saltedPassword))
	if err != nil {
		//passwords do not match
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid email or password",
			"data":    gin.H{},
		})
		return
	}

	//checking verification status of the user ,
	//if pending it will sent a response to login and verify the otp, use  /api/v1/verifyotp to verify the otp
	var VerificationTable model.VerificationTable

	if err := database.DB.Where("email = ? AND role = ?", user.Email, model.UserRole).First(&VerificationTable).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to process otp verification",
			"data":    gin.H{},
		})
		return
	}

	if VerificationTable.VerificationStatus != model.VerificationStatusVerified {
		err := SendOTP(c, user.Email, VerificationTable.OTPExpiry, model.UserRole)
		if err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  false,
				"message": err.Error(),
				"data":    gin.H{},
			})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status":  false,
			"message": "please complete your email verification",
			"data":    gin.H{},
		})
		return
	}

	//generate the jwt token and set it in cookie using generatejwt fn,
	tokenstring, err := GenerateJWT(c, user.Email, model.UserRole)

	if tokenstring == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to create JWT token due to an internal server error.Try again",
			"data":    gin.H{},
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Email login successful.",
		"data": gin.H{
			"user": gin.H{
				"name":         user.Name,
				"email":        user.Email,
				"phone_number": user.PhoneNumber,
				"picture":      user.Picture,
				"login_method": user.LoginMethod,
				"block_status": user.Blocked,
			},
		},
	})

	c.Next()

}
func SendOTP(c *gin.Context, to string, otpexpiry uint64, role string) error {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := r.Intn(900000) + 100000
	// Check if the provided otpexpiry has already passed
	now := time.Now().Unix()
	if otpexpiry > 0 && uint64(now) < otpexpiry {
		// OTP is still valid, respond with a message and do not send a new OTP
		//send back tim left before trying another one
		timeLeft := otpexpiry - uint64(now) 
		str := fmt.Sprintf("OTP is still valid. wait before sending another request, %v seconds left", int(timeLeft))

		return errors.New(str)
	}

	var expiryTime int64
	switch role {
	case model.AdminRole:
		expiryTime = now + 2*60
	case model.RestaurantRole:
		expiryTime = now + 5*60
	case model.UserRole:
		expiryTime = now + 10*60
	}

	// fmt.Printf("Sending mail because OTP has expired: %v\n", expiryTime)

	from := "foodbuddycode@gmail.com"
	appPassword := "emdnwucohpvcoyin"
	auth := smtp.PlainAuth("", from, appPassword, "smtp.gmail.com")
	url := fmt.Sprintf("http://localhost:8080/api/v1/auth/verifyotp/%v/%v/%v", role, to, otp)
	mail := fmt.Sprintf("FoodBuddy Email Verification \n Click here to verify your email %v", url)

	//send the otp to the specified email
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(mail))
	if err != nil {
		return errors.New("failed to send otp")
	}

	//update the otp and expiry
	VerificationTable := model.VerificationTable{
		Email:              to,
		Role:               role,
		OTP:                uint64(otp),
		OTPExpiry:         uint64(expiryTime),
		VerificationStatus: model.VerificationStatusPending, //already metioned during signup //mentioning it sprtly for all routes as well
	}

	if err := database.DB.Where("email = ? AND role = ?", VerificationTable.Email, role).Updates(&VerificationTable).Error; err != nil {
		return errors.New("failed to get information using email")
	}

	return nil
}

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
func VerifyOTP(c *gin.Context) {
	///welcome?firstname=Jane&lastname=Doe
	entityRole := c.Param("role")
	entityEmail := c.Param("email")
	entityOTP, _ := strconv.Atoi(c.Param("otp"))

	if entityRole == "" || entityEmail == "" || entityOTP == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process incoming request",
			"data":    gin.H{},
		})
		return
	}

	var VerificationTable model.VerificationTable

	tx := database.DB.Where("email = ? AND role = ?", entityEmail, entityRole).First(&VerificationTable)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve  information",
			"data":    gin.H{},
		})
		return
	}

	if VerificationTable.VerificationStatus == model.VerificationStatusVerified {
		c.JSON(http.StatusIMUsed, gin.H{
			"status":  false,
			"message": "already verified",
			"data":    gin.H{},
		})
		return
	}

	if VerificationTable.OTP == 0 {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":  false,
			"message": "please login once again to verify your otp",
			"data":    gin.H{},
		})
		return
	}

	if VerificationTable.OTPExpiry < uint64(time.Now().Unix()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "otp has expired ,please login once again to verify your otp",
			"data":    gin.H{},
		})
		return
	}

	if VerificationTable.OTP != uint64(entityOTP) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "otp is invalid ,please login once again to verify your otp",
			"data":    gin.H{},
		})
		return
	}

	VerificationTable.VerificationStatus = model.VerificationStatusVerified

	tx = database.DB.Where("email = ? AND role = ?", entityEmail, entityRole).Updates(&VerificationTable)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to verify otp, please try again",
			"data":    gin.H{},
		})
		return
	}
	var token string
	var err error

	token, err = GenerateJWT(c, entityEmail, entityRole)
	if token == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to generate token, please try again",
			"data":    gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Email verification is Successful",
		"data": gin.H{
			"token": token,
		},
	})
}

func GenerateJWT(c *gin.Context, email string, role string) (string, error) {
	//generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	//sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(utils.GetEnvVariables().JWTSecret))
	if err != nil {
		return "", err
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	return tokenString, nil
}

func VerifyJWT(c *gin.Context, role string, useremail string) error {
	utils.NoCache(c)

	// Attempt to retrieve the JWT token from the cookie
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		return errors.New("no authorization token found in the cookie")
	}

	// Decode and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(utils.GetEnvVariables().JWTSecret), nil
	})

	if err != nil {
		return errors.New("internal server error occurred while parsing the JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("internal server error occurred while parsing the JWT token")
	}

	// Check if the token is expired
	if claimsExpiration, ok := claims["exp"].(float64); ok && claimsExpiration < float64(time.Now().Unix()) {
		return errors.New("authorization token is expired please log in again")
	}

	// Retrieve the user associated with the token
	email, ok := claims["email"].(string)
	if !ok {
		return errors.New("invalid email claim in token")
	}

	tokenRole, ok := claims["role"].(string)
	if !ok {
		return errors.New("invalid role claim in token")
	}

	// Ensure the token role matches the passed role parameter
	if tokenRole != role {
		return errors.New("role mismatch")
	}

	var VerificationTable model.VerificationTable
	tx := database.DB.Where("email = ? AND role = ?", email, role).First(&VerificationTable)
	if tx.Error != nil {
		return errors.New("failed to process user information")
	}

	// If we reach this point, the JWT is valid and the user is authenticated
	return nil
}

func EmailFromUserID(UserID uint) (string, bool) {
	var userinfo model.User
	if err := database.DB.Where("id = ?", UserID).First(&userinfo).Error; err != nil {
		return "", false
	}

	return userinfo.Email, true
}

// removing cookie "authorization"
func Logout(c *gin.Context) {
	utils.RemoveCookies(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully logged out",
		"ok":      true,
	})
	c.Next()
}

