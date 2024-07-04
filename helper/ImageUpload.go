package helper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func ImageUpload(fileHeader *multipart.FileHeader) (string, error) {
	if !isValidFormat(fileHeader.Filename) {
		log.Printf("Invalid file format: %s", fileHeader.Filename)
		return "", fmt.Errorf("invalid file format")
	}

	url := fmt.Sprintf("cloudinary://%v:%v@%v", GetEnvVariables().CloudinaryAccessKey, GetEnvVariables().CloudinarySecretKey, GetEnvVariables().CloudinaryCloudName)
	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		log.Printf("Failed to initialize Cloudinary: %v", err)
		return "", fmt.Errorf("failed to initialize Cloudinary")
	}

	ctx := context.Background()

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return "", fmt.Errorf("failed to open file")
	}
	defer file.Close()

	// read the file into a buffer
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Printf("Failed to read file: %v", err)
		return "", fmt.Errorf("failed to read file")
	}

	// define the upload parameters with the desired transformations
	uploadParams := uploader.UploadParams{
		Transformation: "f_auto/q_auto/c_crop,w_300,h_300,c_fill",
		Folder:         "foodbuddy",
	}

	// upload the file with transformation to Cloudinary
	uploadResult, err := cld.Upload.Upload(ctx, buf, uploadParams)
	if err != nil {
		log.Printf("Failed to upload file: %v", err)
		return "", fmt.Errorf("failed to upload file")
	}

	return uploadResult.SecureURL, nil
}

// to check the file type
func isValidFormat(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}
