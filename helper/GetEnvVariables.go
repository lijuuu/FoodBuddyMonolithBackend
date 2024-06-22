package helper

import (
	"foodbuddy/model"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVariables() model.EnvVariables {

	//checking for error, to find if there's any env file available
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//retrieving all the variables and storing it on the struct and returning it
	EnvVariables := model.EnvVariables{
		ClientID:            os.Getenv("CLIENTID"),
		ClientSecret:        os.Getenv("CLIENTSECRET"),
		DBUser:              os.Getenv("DBUSER"),
		DBPassword:          os.Getenv("DBPASSWORD"),
		DBName:              os.Getenv("DBNAME"),
		JWTSecret:           os.Getenv("JWTSECRET"),
		CloudinaryCloudName: os.Getenv("CLOUDNAME"),
		CloudinaryAccessKey: os.Getenv("CLOUDINARYACCESSKEY"),
		CloudinarySecretKey: os.Getenv("CLOUDINARYSECRETKEY"),
	}
	return EnvVariables
}
