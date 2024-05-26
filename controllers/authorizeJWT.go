package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)


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


func VerifyJWT(c *gin.Context,useremail string)bool {
	utils.NoCache(c)

	// Attempt to retrieve the JWT token from the cookie
	tokenString, err := c.Cookie("Authorization")
	if err!= nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no JWT token found in the cookie",
			"ok": false,
		})
		return false
	}

	// Decode and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC);!ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(utils.GetEnvVariables().JWTSecret), nil
	})

	if err!= nil {
		errstr := fmt.Sprintf("internal server error occurred while parsing the JWT token : /n %v",err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errstr,
			"ok": false,
		})
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check if the token is expired
		if claimsExpiration, ok := claims["exp"].(float64); ok && claimsExpiration < float64(time.Now().Unix()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "JWT token expired, please log in again",
				"ok": false,
			})
			return false
		}

		// Retrieve the user associated with the token
		var user model.User
		tx := database.DB.FirstOrInit(&user, "email =?", claims["sub"])
		if tx.Error!= nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve user information from the database",
				"ok": false,
			})
			return false
		}

		if useremail != user.Email{
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized user",
				"ok": false,
			})
			return false
		}

		// If we reach this point, the JWT is valid and the user is authenticated
		c.JSON(http.StatusAccepted,gin.H{
			"message":"jwt is a valid one, proceed to login",
			"ok": true,
		})
		return true
		
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "internal server error occurred while parsing the JWT token",
			"ok": false,
		})
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
