package view

import (
	"foodbuddy/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)


func LoadLoginPage(c *gin.Context) {
	helper.NoCache(c)
	helper.CheckCookie(c)
	c.HTML(http.StatusOK, "login.html", nil)
	c.Next()
}

func LoadSignupPage(c *gin.Context) {
	helper.NoCache(c)
	helper.CheckCookie(c)
	c.HTML(http.StatusOK, "signup.html", nil)
}

func LoadUpload(c *gin.Context) {
	c.HTML(http.StatusOK, "image.html", nil)
}