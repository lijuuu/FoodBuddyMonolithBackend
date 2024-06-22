package helper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RemoveCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode) //for security against csrf and usability
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	c.Next()
}
