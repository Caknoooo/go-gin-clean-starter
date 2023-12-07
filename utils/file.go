package utils

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strings"
)

const PATH = "storage"

func UploadFile(file *multipart.FileHeader, path string) error {
	parts := strings.Split(path, "/")

	fileId := parts[1]
	dirPath := fmt.Sprintf("%s/%s", PATH, parts[0])

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			return err
		}
	}

	filePath := fmt.Sprintf("%s/%s", dirPath, fileId)

	uploadedFile, err := file.Open()
	if err != nil {
		return err
	}
	defer uploadedFile.Close()

	fileData, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, fileData, 0666)
	if err != nil {
		return err
	}

	return nil
}

func GetExtensions(filename string) string {
	return strings.Split(filename, ".")[1]
}