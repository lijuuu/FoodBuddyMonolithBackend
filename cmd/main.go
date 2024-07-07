package main

import (
	"foodbuddy/database"
	"foodbuddy/helper"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func init() {
	database.ConnectToDB() //export PROJECTROOT=/home/xstill/Desktop/Week8/onlyapi
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("failed to automigrate models")
	}
}

func main() {
	//start server with default logger and recovery
	router := gin.Default()
	//load html from templates folder
	router.LoadHTMLGlob(os.Getenv("PROJECTROOT") + "/templates/*")

	//middleware for cors and api rate limiting`
	router.Use(helper.RateLimitMiddleware())
	router.Use(helper.CorsMiddleware())

	//access all the routes
	ServerHealth(router)
	PublicRoutes(router)
	AuthenticationRoutes(router)
	AdminRoutes(router)
	UserRoutes(router)
	RestaurantRoutes(router)
	AdditionalRoutes(router)

	//run the server at port :8080
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
