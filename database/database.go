package database

import (
	"fmt"
	"foodbuddy/model"
	"foodbuddy/utils"

	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	databaseCredentials := utils.GetEnvVariables()

	dsn := fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v?charset=utf8mb4&parseTime=True&loc=Local", databaseCredentials.DBUser, databaseCredentials.DBPassword, databaseCredentials.DBName)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("unable to connect to database, ", databaseCredentials.DBName)
	} else {
		fmt.Println("connection to database :OK")
	}

}

func AutoMigrate() {
	DB.AutoMigrate(&model.User{})
	DB.AutoMigrate(&model.Restaurant{})
	DB.AutoMigrate(&model.Category{})
	DB.AutoMigrate(&model.Product{})
	DB.AutoMigrate(&model.FavouriteProduct{})
	DB.AutoMigrate(&model.Address{})
	DB.AutoMigrate(&model.Admin{})
	DB.AutoMigrate(&model.VerificationTable{})
}
