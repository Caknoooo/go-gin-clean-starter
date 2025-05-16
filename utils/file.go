package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

var PATH = "assets"

func UploadFile(file *multipart.FileHeader, path string) error {
	// Split the path and get the file ID (last part)
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		return fmt.Errorf("invalid path: %s", path)
	}
	fileID := parts[len(parts)-1]
	// Directory is all parts except the last one
	dirParts := parts[:len(parts)-1]
	dirPath := filepath.Join(PATH, filepath.Join(dirParts...))

	// Create the directory structure if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	// Construct the full file path
	filePath := filepath.Join(dirPath, fileID)

	// Open the uploaded file
	uploadedFile, err := file.Open()
	if err != nil {
		return err
	}
	defer func(uploadedFile multipart.File) {
		err := uploadedFile.Close()
		if err != nil {
			panic(err)
		}
	}(uploadedFile)

	// Create the target file
	targetFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func(targetFile *os.File) {
		err := targetFile.Close()
		if err != nil {
			panic(err)
		}
	}(targetFile)

	// Copy file contents
	_, err = io.Copy(targetFile, uploadedFile)
	if err != nil {
		return err
	}

	return nil
}

func GetExtensions(filename string) string {
	return strings.Split(filename, ".")[len(strings.Split(filename, "."))-1]
}
