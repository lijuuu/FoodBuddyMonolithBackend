package utils

import "github.com/gin-gonic/gin"

func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, no-transform, must-revalidate, private, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Next()
}
