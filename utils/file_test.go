package utils

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadFile(t *testing.T) {
	// Setup test directory
	testDir := "./test_assets"
	t.Cleanup(
		func() {
			err := os.RemoveAll(testDir)
			if err != nil {
				panic(err)
			}
		},
	)

	// Create a test file to upload
	createTestFile := func(t *testing.T, content string) *multipart.FileHeader {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.txt")
		require.NoError(t, err)
		_, err = part.Write([]byte(content))
		require.NoError(t, err)
		err = writer.Close()
		if err != nil {
			return nil
		}

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		file, header, err := req.FormFile("file")
		require.NoError(t, err)
		err = file.Close()
		if err != nil {
			return nil
		}

		return header
	}

	tests := []struct {
		name        string
		path        string
		fileContent string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Successfully upload file",
			path:        "images/test123",
			fileContent: "test file content",
			wantErr:     false,
		},
		{
			name:        "Create nested directory structure",
			path:        "images/nested/test123",
			fileContent: "test file content",
			wantErr:     false,
		},
		{
			name:        "Empty file",
			path:        "images/test123",
			fileContent: "",
			wantErr:     false,
		},
		{
			name:        "Invalid file header",
			path:        "images/test123",
			wantErr:     true,
			errContains: "no such file",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Replace the constant for testing
				originalPath := PATH
				PATH = testDir
				t.Cleanup(
					func() {
						PATH = originalPath
					},
				)

				var fileHeader *multipart.FileHeader
				if tt.name != "Invalid file header" {
					fileHeader = createTestFile(t, tt.fileContent)
				} else {
					fileHeader = &multipart.FileHeader{}
				}

				err := UploadFile(fileHeader, tt.path)

				if tt.wantErr {
					require.Error(t, err)
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				} else {
					require.NoError(t, err)

					// Verify file was created
					parts := strings.Split(tt.path, "/")
					fileID := parts[len(parts)-1]
					dirPath := filepath.Join(testDir, strings.Join(parts[:len(parts)-1], "/"))
					filePath := filepath.Join(dirPath, fileID)

					_, err = os.Stat(filePath)
					assert.NoError(t, err)

					// Verify file content
					if tt.fileContent != "" {
						content, err := os.ReadFile(filePath)
						require.NoError(t, err)
						assert.Equal(t, tt.fileContent, string(content))
					}
				}
			},
		)
	}

	t.Run(
		"Permission denied for directory creation", func(t *testing.T) {
			if os.Geteuid() == 0 {
				t.Skip("Skipping test when running as root")
			}

			// Replace the constant for testing with protected directory
			originalPath := PATH
			PATH = "/root/protected_directory"
			t.Cleanup(
				func() {
					PATH = originalPath
				},
			)

			fileHeader := createTestFile(t, "test content")
			err := UploadFile(fileHeader, "images/test123")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "permission denied")
		},
	)
}

func TestGetExtensions(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "Simple extension",
			filename: "file.txt",
			want:     "txt",
		},
		{
			name:     "Multiple dots",
			filename: "archive.tar.gz",
			want:     "gz",
		},
		{
			name:     "No extension",
			filename: "file",
			want:     "file",
		},
		{
			name:     "Empty string",
			filename: "",
			want:     "",
		},
		{
			name:     "Dot at end",
			filename: "file.",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := GetExtensions(tt.filename)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}
