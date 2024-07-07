package helper

import (
	"fmt"
	"foodbuddy/model"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func GetEnvVariables() model.EnvVariables {
	
    path, ok := os.LookupEnv("PROJECTROOT")
    if !ok {
        log.Fatal("PROJECTROOT environment variable not found,Please set your PROJECTROOT variable by running this on your terminal\n [ export PROJECTROOT={absolute_path_till_project_root_dir} ]")
    }

    envFilePath := path + "/.env"
    fmt.Println(envFilePath)
    err := godotenv.Load(envFilePath)
    if err != nil {
        log.Fatal("Error loading .env file:", err)
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
func AbsoluetPath() string {
	AbsPath, err := filepath.Abs(".")
	if err != nil {
		fmt.Printf("Error converting path to absolute: %v\n", err)
		return AbsPath
	}
	return AbsPath
}
