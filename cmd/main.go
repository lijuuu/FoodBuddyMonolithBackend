package main

import (
	"log"
	"os"

	"foodbuddy/internal/database"
	"foodbuddy/internal/utils"
	"foodbuddy/internal/api"

	"github.com/gin-gonic/gin"
)

func init() {
	_, ok := os.LookupEnv("PROJECTROOT")
	if !ok {
		log.Fatal("PROJECTROOT environment variable not found,Please set your PROJECTROOT variable by running this on your terminal\n [ export PROJECTROOT={absolute_path_till_project_root_dir} ]")
	}
	//connect to db
	database.ConnectToDB() //export PROJECTROOT=/home/xstill/Desktop/Week8/onlyapi
	database.AutoMigrate()
}

func main() {
	//start server with default logger and recovery
	router := gin.Default()
	//load html from templates folder
	path, _ := os.LookupEnv("PROJECTROOT")
	router.LoadHTMLGlob(path + "/templates/*")

	//middleware for cors and api rate limiting`
	router.Use(utils.RateLimitMiddleware())
	router.Use(utils.CorsMiddleware())

	//access all the routes
	api.ServerHealth(router)
	api.PublicRoutes(router)
	api.AuthenticationRoutes(router)
	api.AdminRoutes(router)
	api.UserRoutes(router)
	api.RestaurantRoutes(router)
	api.AdditionalRoutes(router)

	//run the server at port :8080
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
