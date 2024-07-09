package utils

import (
	"foodbuddy/internal/model"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVariables() model.EnvVariables {
	path, exist := os.LookupEnv(model.ProjectRoot)
	if !exist {
		log.Fatal("PROJECTROOT environment variable not found,Please set your PROJECTROOT variable by running this on your terminal\n [ export PROJECTROOT={absolute_path_till_project_root_dir} ]")
	}
	envFilePath := path + "/.env"
	err := godotenv.Load(envFilePath)
	if err != nil {
	}

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
