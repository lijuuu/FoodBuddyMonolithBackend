package utils

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetJWTEmailClaim attempts to extract the email claim from a JWT token.
func GetJWTClaim(c *gin.Context) (email string,role string,err error) {
	JWTToken, err := c.Cookie("Authorization")
	if JWTToken == "" || err != nil {
		return "","", errors.New("no authorization token available")
	}

	hmacSecretString := GetEnvVariables().JWTSecret
	hmacSecret := []byte(hmacSecretString)

	// Parse the token
	token, err := jwt.Parse(JWTToken, func(token *jwt.Token) (interface{}, error) {
		if err != nil {
			return "", errors.New("request unauthorized")
		}
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return hmacSecret, nil
	})

	if err != nil {
		return "","", errors.New("request unauthorized")
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check for expiration
		expirationTime, ok := claims["exp"].(float64)
		if !ok {
			return "","", errors.New("request unauthorized")
		}

		// Convert the expiration time to a Time value
		expiration := time.Unix(int64(expirationTime), 0)

		// Check if the token is expired
		if time.Now().After(expiration) {
			return "","", errors.New("request unauthorized")
		}

		email := claims["email"].(string)
		role := claims["role"].(string)
		return email,role, nil
	} else {
		return "","", errors.New("request unauthorized")
	}
}
