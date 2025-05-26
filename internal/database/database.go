package database

import (
	"fmt"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	databaseCredentials := utils.GetEnvVariables()

	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/?parseTime=true", databaseCredentials.DBUser, databaseCredentials.DBPassword)
	log.Println("Connecting to MySQL server with DSN:", dsn)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to connect to MySQL server: %v", err)
	}

	if !databaseExists(databaseCredentials.DBName) {
		if err := createDatabase(databaseCredentials.DBName); err != nil {
			log.Printf("Warning: %v", err)
		}
	}

	dsn = fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s?parseTime=true", databaseCredentials.DBUser, databaseCredentials.DBPassword, databaseCredentials.DBName)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	} else {
		log.Println("Connection to database: OK")
	}

	AutoMigrate()
}

func databaseExists(dbName string) bool {
	var exists int
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/?parseTime=true", os.Getenv("DBUSER"), os.Getenv("DBPASSWORD"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to connect to MySQL server: %v", err)
	}

	if err := db.Raw("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?", dbName).Scan(&exists).Error; err != nil {
		log.Printf("Failed to check database existence: %v", err)
		return false
	}
	return exists > 0
}

func createDatabase(dbName string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/?parseTime=true", os.Getenv("DBUSER"), os.Getenv("DBPASSWORD"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("unable to connect to MySQL server: %w", err)
	}

	if err := db.Exec("CREATE DATABASE " + dbName).Error; err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	log.Printf("Database %s created successfully!", dbName)
	return nil
}

func AutoMigrate() {
	err := DB.AutoMigrate(
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
		&model.UserReferralHistory{},
		&model.DeliveryVerification{},
	)

	if err != nil {
		log.Fatal("failed to automigrate models:", err)
	}
}
