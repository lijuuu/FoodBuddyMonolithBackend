package utils

import (
	"foodbuddy/internal/model"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVariables() model.EnvVariables {
	cwd, _ := os.Getwd()

	envFilePath := cwd + "/.env"
	err := godotenv.Load(envFilePath)
	if err != nil {
	}

	//retrieving all the variables and storing it on the struct and returning it
	EnvVariables := model.EnvVariables{
		ServerIP:            os.Getenv("SERVERIP"),
		Port:                os.Getenv("PORT"),
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
