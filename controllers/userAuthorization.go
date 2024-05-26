package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
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
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code parameter", "ok": false})
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token", "ok": false})
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("google signup done")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info", "ok": false})
		return
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read user info", "ok": false})
		return
	}

	var User model.GoogleResponse
	err = json.Unmarshal(content, &User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info", "ok": false})
		return
	}

	newUser := model.User{
		Name:               User.Name,
		Email:              User.Email,
		LoginMethod:        model.GoogleSSOMethod,
		Picture:            User.Picture,
		VerificationStatus: model.VerificationStatusVerified,
		Blocked:            false,
	}

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
	c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully", "user": existingUser, "jwttoken": tokenstring, "ok": true})

}


func EmailLogin(c *gin.Context) {
	var form model.LoginForm

	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok := validate(form,c)
	if !ok{
		return
	}

	var user model.User
	tx := database.DB.Where("email =? AND deleted_at IS NULL", form.Email).First(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid credentials or user doesn't exist on the database",
			"ok":    false,
		})
		return
	}

	if user.LoginMethod != model.EmailLoginMethod {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "please login through google",
			"ok":    false,
		})
		return
	}

	//check is the user is blocked by the admin
	if user.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user is restricted from accessing, blocked by the administrator", "ok": false})
		return
	}

	// password with salt = user.salt + form.password 

	//get the hash and compare it with password from body
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(form.Password))
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
	//check jwt via cookie thru /api/v1/verifyjwt
	// and using json through /api/v1/checkjwt
	tokenstring := GenerateJWT(c, user.Email)

	if tokenstring == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "jwt token is empty please try again",
			"ok":    false,
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "logged in successfully", "user": user, "jwttoken": tokenstring, "ok": true})

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
	//salt+password
	//hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
			"ok":    false,
		})
		return
	}

	User := model.User{
		Name:               body.Name,
		Email:              body.Email,
		HashedPassword:     string(hash),
		LoginMethod:        model.EmailLoginMethod,
		VerificationStatus: model.VerificationStatusPending,
		Blocked:            false,
		//add salt 
	}

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
