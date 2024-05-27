package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"io"
	"math/rand"
	"net/http"
	"net/smtp"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code parameter", "ok": false})
		return
	}

	//exchange code for token, code is exchanged to make sure the state is same
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token", "ok": false})
		return
	}

	//use access token and get reponse of the user
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("google signup done")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info", "ok": false})
		return
	}
	defer response.Body.Close()

	//read the content of the reponse.body
	content, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read user info", "ok": false})
		return
	}

	//store the content from the json to the user struct of model.GoogleResponse
	var User model.GoogleResponse
	err = json.Unmarshal(content, &User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info", "ok": false})
		return
	}

	//pass the values needed from the google response to the newuser struct
	newUser := model.User{
		Name:               User.Name,
		Email:              User.Email,
		LoginMethod:        model.GoogleSSOMethod,
		Picture:            User.Picture,
		VerificationStatus: model.VerificationStatusVerified,
		Blocked:            false,
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
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create user using google signup method", "ok": false})
				return
			}
		} else {
			// Handle case where user already exists but not found due to other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error while fetching user", "ok": false})
			return
		}
	}

	// User already exists, check login method
	if existingUser.LoginMethod == model.EmailLoginMethod {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists, please use email for login", "ok": false})
		return
	}

	//check is the user is blocked by the admin
	if existingUser.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user is restricted from accessing, blocked by the administrator", "ok": false})
		return
	}

	// Generate JWT and set cookie within GenerateJWT
	tokenstring := GenerateJWT(c, newUser.Email)
	if tokenstring == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "jwt token is empty please try again",
			"ok":    false,
		})
		return
	}

	// Return success response
	fmt.Println("google signup done")
	c.JSON(http.StatusOK, gin.H{
		"message":  "Logged in successfully",
		"user":     existingUser,
		"jwttoken": tokenstring,
		"ok":       true,
	})

}

func EmailLogin(c *gin.Context) {
	var form model.LoginForm

	//get the json from the request
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//validate the content of the json
	ok := validate(form, c)
	if !ok {
		return
	}

	//chekc whether the email exist on the database, if not return an error
	var user model.User
	tx := database.DB.Where("email =? AND deleted_at IS NULL", form.Email).First(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid credentials or user doesn't exist on the database",
			"ok":    false,
		})
		return
	}

	//check if the login methods are the same as email, if google prompt to use google login
	if user.LoginMethod != model.EmailLoginMethod {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "please login through google",
			"ok":    false,
		})
		return
	}

	//check is the user is blocked by the admin
	if user.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user is restricted from accessing, blocked by the administrator",
			"ok":    false,
		})
		return
	}

	// password with salt = user.salt + form.password
	saltedPassword := user.Salt + form.Password

	//get the hash and compare it with password from body
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(saltedPassword))
	if err != nil {
		//passwords do not match
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "credentials are wrong",
			"ok":    false,
		})
		return
	}

	//checking verification status of the user ,
	//if pending it will sent a response to login and verify the otp, use  /api/v1/verifyotp to verify the otp
	if user.VerificationStatus == model.VerificationStatusPending {
		SendOTP(c, user.ID, user.Email, user.OTPexpiry)
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "email verification status is pending, please verify via email verification code",
			"ok":    false,
		})
		return
	}

	//generate the jwt token and set it in cookie using generatejwt fn,
	tokenstring := GenerateJWT(c, user.Email)

	if tokenstring == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "jwt token is empty please try again",
			"ok":    false,
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":  "logged in successfully",
		"user":     user,
		"jwttoken": tokenstring,
		"ok":       true,
	})

	c.Next()

}

// using signup via email
func EmailSignup(c *gin.Context) {

	utils.NoCache(c)

	//get the body
	var body struct {
		Name            string `validate:"required" json:"name"`
		Email           string `validate:"required,email" json:"email"`
		Password        string `validate:"required" json:"password"`
		ConfirmPassword string `validate:"required" json:"confirmpassword"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "ok": false})
		return
	}

	ok := validate(body, c)
	if !ok {
		return
	}

	//check if the password and the confirm password is correct
	if body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "password is not a match",
			"ok":    false,
		})
		return
	}

	//create salt and add it to the password
	Salt := utils.GenerateRandomString(7)
	//salt+password
	saltedPassword := Salt + body.Password

	//hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
			"ok":    false,
		})
		return
	}

	//add the data to user struct
	User := model.User{
		Name:               body.Name,
		Email:              body.Email,
		HashedPassword:     string(hash),
		LoginMethod:        model.EmailLoginMethod,
		VerificationStatus: model.VerificationStatusPending,
		Blocked:            false,
		Salt:               Salt,
	}

	//check if the user exists on the database
	tx := database.DB.Where("email =? AND deleted_at IS NULL", body.Email).First(&User)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": tx.Error,
			"ok":    false,
		})
		return

	} else if tx.Error == gorm.ErrRecordNotFound {
		// User does not exist, proceed to create
		tx = database.DB.Create(&User)
		if tx.Error != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": tx.Error,
				"ok":    false,
			})
			return
		}
	} else {
		// User already exists
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user already exists",
			"ok":    false,
		})
		return
	}

	// //Generating JWT to access the home page //using verification otp ,reason for commenting the code
	// tokenstring  := GenerateJWT(c, body.Email)

	// if tokenstring == ""{
	// 	c.JSON(http.StatusInternalServerError,gin.H{
	// 		"Error":"not able to generate jwt token",
	// 	})
	// }

	c.JSON(http.StatusOK, gin.H{
		// "jwttoken":tokenstring,
		"message": "signup is successfull, login and complete your otp verification",
		"user":    User,
		"ok":      true,
	})
	c.Next()
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

func SendOTP(c *gin.Context, userID uint, to string, otpexpiry int64) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := r.Intn(900000) + 100000

	// Check if the provided otpexpiry has already passed
	now := time.Now().Unix()
	if otpexpiry > 0 && now < otpexpiry {
		// OTP is still valid, respond with a message and do not send a new OTP
		c.JSON(http.StatusOK, gin.H{
			"error": "OTP is still valid. wait before sending another request.",
			"ok":    false,
		})
		return
	}

	// Proceed to send a new OTP if the previous one has expired or if no otpexpiry was provided
	// Set expiryTime as 10 minutes from now
	expiryTime := now + 10*60 // 10 minutes in seconds

	fmt.Printf("Sending mail because OTP has expired: %v\n", expiryTime)

	from := "foodbuddycode@gmail.com"
	appPassword := "emdnwucohpvcoyin"
	auth := smtp.PlainAuth("", from, appPassword, "smtp.gmail.com")

	mail := fmt.Sprintf("FoodBuddy OTP Verification\nThis is your Verification code: %d", otp)

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(mail))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "sending otp failed",
			"ok":    false,
		})
		return
	}

	user := model.User{
		ID:        userID,
		OTP:       otp,
		OTPexpiry: expiryTime, // Store the Unix timestamp directly
	}

	tx := database.DB.Updates(&user)
	if tx.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "failed to save otp on database",
			"ok":    false,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "verification code sent to mail,verify to continue",
			"ok":      true,
		})
	}
}

func VerifyOTP(c *gin.Context) {

	var userRequest model.User
	var user model.User

	if err := c.BindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	tx := database.DB.Where("email =?", userRequest.Email).First(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user not found",
			"ok":    false,
		})
		return
	}

	if user.VerificationStatus == model.VerificationStatusVerified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user already verified",
			"ok":    false,
		})
		return
	}

	if user.OTP == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "login before verifying otp",
			"ok":    false,
		})
		return
	}

	if user.OTPexpiry < time.Now().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "otp expired",
			"ok":    false,
		})
		return
	}

	if user.OTP != userRequest.OTP {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid otp",
			"ok":    false,
		})
		return
	}

	// Correctly placed inside the if block to ensure it only executes if the OTP is correct
	user.VerificationStatus = model.VerificationStatusVerified

	tx = database.DB.Updates(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "otp verification failed",
			"ok":    false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "otp verified successfully",
		"user":    user,
		"ok":      true,
	})
}

func GenerateJWT(c *gin.Context, email string) string {
	//generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	//sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(utils.GetEnvVariables().JWTSecret))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "error while generating jwt ",
		})
		return ""
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	return tokenString
}

func VerifyJWT(c *gin.Context, useremail string) bool {
	utils.NoCache(c)

	// Attempt to retrieve the JWT token from the cookie
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no JWT token found in the cookie",
			"ok":    false,
		})
		return false
	}

	// Decode and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(utils.GetEnvVariables().JWTSecret), nil
	})

	if err != nil {
		errstr := fmt.Sprintf("internal server error occurred while parsing the JWT token : /n %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errstr,
			"ok":    false,
		})
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check if the token is expired
		if claimsExpiration, ok := claims["exp"].(float64); ok && claimsExpiration < float64(time.Now().Unix()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "JWT token expired, please log in again",
				"ok":    false,
			})
			return false
		}

		// Retrieve the user associated with the token
		var user model.User
		tx := database.DB.FirstOrInit(&user, "email =?", claims["sub"])
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve user information from the database",
				"ok":    false,
			})
			return false
		}
		ok := IsAdmin(user.Email)

		if useremail != user.Email || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized user",
				"ok":    false,
			})
			return false
		}

		// If we reach this point, the JWT is valid and the user is authenticated
		c.JSON(http.StatusAccepted, gin.H{
			"message": "jwt is a valid one, proceed to login",
			"ok":      true,
		})
		return true

	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "internal server error occurred while parsing the JWT token",
			"ok":    false,
		})
		return false
	}
	return true
}

func EmailFromUserID(UserID uint) (string, bool) {
	var userinfo model.User
	if err := database.DB.Where("id = ?", UserID).First(&userinfo).Error; err != nil {
		return "", false
	}

	return userinfo.Email, true
}

func IsAdmin(email string) bool {
	var Admin model.Admin
	if err := database.DB.Where("email = ?", email).First(&Admin).Error; err != nil {
		return false
	}
	if Admin.Email == "" {
		return false
	}

	return true
}
