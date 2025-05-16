package utils_test

import (
	"bytes"
	"github.com/Caknoooo/go-gin-clean-starter/utils"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FileUploadIntegrationTestSuite struct {
	suite.Suite
	testDir string
}

func (suite *FileUploadIntegrationTestSuite) SetupSuite() {
	// Setup test directory
	suite.testDir = "./test_assets_integration"
	utils.PATH = suite.testDir
	err := os.MkdirAll(suite.testDir, 0755)
	if err != nil {
		panic(err)
	}
}

func (suite *FileUploadIntegrationTestSuite) TearDownSuite() {
	// Clean up test directory
	err := os.RemoveAll(suite.testDir)
	if err != nil {
		panic(err)
	}
	utils.PATH = "assets" // Reset to original value
}

func (suite *FileUploadIntegrationTestSuite) createTestFile(content string) *multipart.FileHeader {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(suite.T(), err)
	_, err = part.Write([]byte(content))
	require.NoError(suite.T(), err)
	err = writer.Close()
	if err != nil {
		return nil
	}

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	file, header, err := req.FormFile("file")
	require.NoError(suite.T(), err)
	err = file.Close()
	if err != nil {
		return nil
	}

	return header
}

func (suite *FileUploadIntegrationTestSuite) TestUploadFile_Integration() {
	tests := []struct {
		name        string
		path        string
		fileContent string
		wantErr     bool
	}{
		{
			name:        "Successfully upload file",
			path:        "images/test123",
			fileContent: "integration test content",
			wantErr:     false,
		},
		{
			name:        "Create nested directory structure",
			path:        "images/nested/test123",
			fileContent: "nested content",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		suite.Run(
			tt.name, func() {
				fileHeader := suite.createTestFile(tt.fileContent)
				err := utils.UploadFile(fileHeader, tt.path)

				if tt.wantErr {
					assert.Error(suite.T(), err)
				} else {
					assert.NoError(suite.T(), err)

					// Verify file was created
					parts := strings.Split(tt.path, "/")
					fileID := parts[len(parts)-1]
					dirPath := filepath.Join(suite.testDir, strings.Join(parts[:len(parts)-1], "/"))
					filePath := filepath.Join(dirPath, fileID)

					_, err = os.Stat(filePath)
					assert.NoError(suite.T(), err)

					// Verify file content
					content, err := os.ReadFile(filePath)
					require.NoError(suite.T(), err)
					assert.Equal(suite.T(), tt.fileContent, string(content))
				}
			},
		)
	}
}

func TestFileUploadIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(FileUploadIntegrationTestSuite))
}
