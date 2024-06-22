package main

import (
	"foodbuddy/database"
	"foodbuddy/helper"

	"github.com/gin-gonic/gin"
)

func init() {
	database.ConnectToDB()
	database.AutoMigrate()
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("../templates/*")

	router.Use(helper.RateLimitMiddleware())
	router.Use(helper.CorsMiddleware())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})

	PublicRoutes(router)
	AuthenticationRoutes(router)
	AdminRoutes(router)
	UserRoutes(router)
	RestaurantRoutes(router)
	AdditionalRoutes(router)

	router.Run(":8080")
}

