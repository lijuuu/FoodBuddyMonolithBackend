package helper

import (
	"foodbuddy/model"
	"os"
)

func GetEnvVariables() model.EnvVariables {

	// err := godotenv.Load("../.env")
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	//retrieving all the variables and storing it on the struct and returning it
	EnvVariables := model.EnvVariables{
		ServerIP:            os.Getenv("SERVERIP"),
		ServerPort:          os.Getenv("SERVERPORT"),
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
