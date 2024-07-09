package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckCookie(c *gin.Context) {

	//get the jwt token string
	cookie, err := c.Cookie("Authorization")
	if cookie == "" || err != nil {
		fmt.Println("CHECKCOOKIE - no Authorization token available, redirecting to login/signup page")
		c.Next()
		return
	} else {
		c.Redirect(http.StatusSeeOther, "/home")
		fmt.Println("CHECKCOOKIE - user cookie is still present , redirecting to homepage")
	}
}
