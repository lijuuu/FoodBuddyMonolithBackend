package database

import (
	"fmt"
	"foodbuddy/model"
	"foodbuddy/helper"

	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	databaseCredentials := helper.GetEnvVariables()

	dsn := fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v?charset=utf8mb4&parseTime=True&loc=Local", databaseCredentials.DBUser, databaseCredentials.DBPassword, databaseCredentials.DBName)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("unable to connect to database, ", databaseCredentials.DBName)
	} else {
		fmt.Println("connection to database :OK")
	}

}

func AutoMigrate() {
	DB.AutoMigrate(
		&model.User{},
		&model.Restaurant{},
		&model.Category{},
		&model.Product{},
		&model.FavouriteProduct{},
		&model.Address{},
		&model.Admin{},
		&model.VerificationTable{},
		&model.CartItems{},
		&model.Order{},
		&model.OrderItem{},
		&model.Payment{},
		&model.PasswordReset{},
		&model.CouponInventory{},
		&model.CouponUsage{},
		&model.UserWalletHistory{},
		&model.RestaurantWalletHistory{},
		&model.UserReferralHistory{}
	)
}
