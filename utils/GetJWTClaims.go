package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetJWTEmailClaim attempts to extract the email claim from a JWT token.
func GetJWTEmailClaim(c *gin.Context) (string){
	JWTToken, err := c.Cookie("Authorization")
	if JWTToken == "" || err != nil {
		fmt.Println("no Authorization token available")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no authorization token available", "ok": false})
		c.Abort()
		return ""
	}

	hmacSecretString := GetEnvVariables().JWTSecret
	hmacSecret := []byte(hmacSecretString)

	// Parse the token
	token, err := jwt.Parse(JWTToken, func(token *jwt.Token) (interface{}, error) {
		if err !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse the token, access denied", "ok": false})
		}
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return hmacSecret, nil
	})

	if err != nil {
		log.Printf("Error parsing token: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse jwt token, access denied", "ok": false})
		c.Abort()
		return ""
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check for expiration
		expirationTime, ok := claims["exp"].(float64)
		if !ok {
			log.Printf("Token does not contain 'exp' claim")
			c.JSON(http.StatusBadRequest, gin.H{"error": "token does not contain 'exp' claim, access denied", "ok": false})
			c.Abort()
			return ""
		}

		// Convert the expiration time to a Time value
		expiration := time.Unix(int64(expirationTime), 0)

		// Check if the token is expired
		if time.Now().After(expiration) {
			log.Printf("Token is expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is expired, access denied", "ok": false})
			c.Abort()
			return ""
		}

		email := claims["email"].(string)
		return email
	} else {
		log.Printf("Token is not valid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token, access denied", "ok": false})
		c.Abort()
		return ""
	}
}
