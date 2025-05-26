package main

import (
	"fmt"

	"foodbuddy/internal/api"
	"foodbuddy/internal/database"
	"foodbuddy/internal/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	database.ConnectToDB()
	database.AutoMigrate()
	fmt.Println("Database intitialization done")
}

func main() {
	//start server with default logger and recovery
	router := gin.Default()
	//load html from templates folder
	// router.LoadHTMLGlob("./templates/*")

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

	err := router.Run(":"+utils.GetEnvVariables().Port)
	if err != nil {
		panic(err)
	}
}
