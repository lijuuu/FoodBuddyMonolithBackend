package main

import (
	"foodbuddy/database"
	"foodbuddy/helper"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func init() {
	_, ok := os.LookupEnv("PROJECTROOT")
	if !ok {
		log.Fatal("PROJECTROOT environment variable not found,Please set your PROJECTROOT variable by running this on your terminal\n [ export PROJECTROOT={absolute_path_till_project_root_dir} ]")
	}
	database.ConnectToDB() //export PROJECTROOT=/home/xstill/Desktop/Week8/onlyapi
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("failed to automigrate models")
	}
}

func main() {
	//start server with default logger and recovery
	router := gin.Default()
	//load html from templates folder
	path, _ := os.LookupEnv("PROJECTROOT")
	router.LoadHTMLGlob(path + "/templates/*")

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
