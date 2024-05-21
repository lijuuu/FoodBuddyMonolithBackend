package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func EmailLogin(c *gin.Context) {
	var form model.LoginForm

	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if form.Email == "" || form.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid credentials",
			"ok":    false,
		})
		return
	}

	//validate the struct body
	// validate := validator.New()
	// err := validate.Struct(form)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"error": "failed to validate the struct body",
	// 		"ok":    false,
	// 	})
	// 	return
	// }

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

// removing cookie "authorization"
func Logout(c *gin.Context) {
	utils.RemoveCookies(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully logged out",
		"ok":      true,
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

	//validate the struct body
	validate := validator.New()
	err := validate.Struct(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to validate the struct body",
			"ok":    false,
		})
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
	}

	tx := database.DB.Where("email =? AND deleted_at IS NULL", body.Email).First(&User)

	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": tx.Error,
			"ok":    false,
		})
		return

	} else if tx.Error == gorm.ErrRecordNotFound {
		// User does not exist, proceed to create
		tx = database.DB.Create(&User)
		if tx.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
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
