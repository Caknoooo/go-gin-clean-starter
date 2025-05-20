package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/Caknoooo/go-gin-clean-starter/dto"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/provider"
	"github.com/Caknoooo/go-gin-clean-starter/routes"
	"github.com/Caknoooo/go-gin-clean-starter/service"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"github.com/Caknoooo/go-gin-clean-starter/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	injector *do.Injector
)

func TestMain(m *testing.M) {

	testContainer, err := container.StartTestContainer()
	if err != nil {
		panic(fmt.Sprintf("Failed to start test container: %v", err))
	}

	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_PORT", testContainer.Port)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")

	db = container.SetUpDatabaseConnection()

	if err := db.AutoMigrate(&entity.User{}, &entity.RefreshToken{}); err != nil {
		panic(fmt.Sprintf("Failed to migrate tables: %v", err))
	}

	injector = do.New()

	do.ProvideNamed(
		injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
			return db, nil
		},
	)

	jwtService := service.NewJWTService()
	do.ProvideNamed(
		injector, constants.JWTService, func(i *do.Injector) (service.JWTService, error) {
			return jwtService, nil
		},
	)

	provider.ProvideUserDependencies(injector)

	code := m.Run()

	if err := container.CloseDatabaseConnection(db); err != nil {
		fmt.Printf("Failed to close database connection: %v\n", err)
	}
	if err := testContainer.Stop(); err != nil {
		fmt.Printf("Failed to stop test container: %v\n", err)
	}

	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM refresh_tokens")

	os.Exit(code)
}

func TestUserRoutes(t *testing.T) {

	router := gin.Default()

	routes.User(router, injector)

	createTestUserAndGetToken := func(t *testing.T) (string, string) {
		registerReq := dto.UserCreateRequest{
			Name:       fmt.Sprintf("Test User %d", time.Now().UnixNano()),
			TelpNumber: "08123456789",
			Email:      fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
			Password:   "password123",
		}

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", registerReq.Name)
		writer.WriteField("email", registerReq.Email)
		writer.WriteField("password", registerReq.Password)
		writer.Close()

		regReq, err := http.NewRequest("POST", "/api/user", body)
		if err != nil {
			t.Fatal(err)
		}
		regReq.Header.Set("Content-Type", writer.FormDataContentType())

		regRec := httptest.NewRecorder()
		router.ServeHTTP(regRec, regReq)

		if regRec.Code != http.StatusOK {
			t.Fatalf("Failed to register test user: status %d, body: %s", regRec.Code, regRec.Body.String())
		}

		loginReq := dto.UserLoginRequest{
			Email:    registerReq.Email,
			Password: registerReq.Password,
		}
		loginBytes, err := json.Marshal(loginReq)
		if err != nil {
			t.Fatal(err)
		}

		loginReqHttp, err := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(loginBytes))
		if err != nil {
			t.Fatal(err)
		}
		loginReqHttp.Header.Set("Content-Type", "application/json")

		loginRec := httptest.NewRecorder()
		router.ServeHTTP(loginRec, loginReqHttp)

		if loginRec.Code != http.StatusOK {
			t.Fatalf("Failed to login test user: status %d, body: %s", loginRec.Code, loginRec.Body.String())
		}

		var loginRes struct {
			Status  bool              `json:"status"`
			Message string            `json:"message"`
			Data    dto.TokenResponse `json:"data"`
		}
		err = json.Unmarshal(loginRec.Body.Bytes(), &loginRes)
		if err != nil {
			t.Fatalf("Failed to parse login response: %v", err)
		}
		accessToken := loginRes.Data.AccessToken
		if accessToken == "" {
			t.Fatal("Access token is empty in login response")
		}

		return accessToken, registerReq.Email
	}

	// Test cases
	tests := []struct {
		name         string
		method       string
		path         string
		body         interface{}
		contentType  string
		authToken    string
		expectedCode int
		setupUser    bool
	}{
		{
			name:   "Register - Invalid input",
			method: "POST",
			path:   "/api/user",
			body: dto.UserCreateRequest{
				Name:       "",
				Email:      "invalid",
				Password:   "short",
				TelpNumber: "",
			},
			contentType:  "multipart/form-data",
			expectedCode: http.StatusBadRequest,
			setupUser:    false,
		},
		{
			name:         "GetAllUser - Valid request",
			method:       "GET",
			path:         "/api/user?page=1&per_page=10",
			body:         nil,
			contentType:  "",
			expectedCode: http.StatusOK,
			setupUser:    false,
		},
		{
			name:         "Login - Invalid credentials",
			method:       "POST",
			path:         "/api/user/login",
			body:         dto.UserLoginRequest{Email: "wrong@example.com", Password: "wrong"},
			contentType:  "application/json",
			expectedCode: http.StatusBadRequest,
			setupUser:    false,
		},
		{
			name:         "Refresh - Invalid token",
			method:       "POST",
			path:         "/api/user/refresh",
			body:         dto.RefreshTokenRequest{RefreshToken: "G6rAsWzMbYTKyzw0g/YgINPPxPWF9PWEOQBUp/4g1VM="},
			contentType:  "application/json", // Changed from multipart/form-data
			expectedCode: http.StatusUnauthorized,
			setupUser:    false,
		},
		{
			name:         "Delete - With auth",
			method:       "DELETE",
			path:         "/api/user",
			body:         nil,
			contentType:  "",
			authToken:    "",
			expectedCode: http.StatusOK,
			setupUser:    true,
		},
		{
			name:         "Update - With auth",
			method:       "PATCH",
			path:         "/api/user",
			body:         dto.UserUpdateRequest{Name: "Updated Name", Email: "updated@example.com"},
			contentType:  "application/json",
			authToken:    "",
			expectedCode: http.StatusOK,
			setupUser:    true,
		},
		{
			name:         "Me - With auth",
			method:       "GET",
			path:         "/api/user/me",
			body:         nil,
			contentType:  "",
			authToken:    "",
			expectedCode: http.StatusOK,
			setupUser:    true,
		},
		{
			name:         "Delete - No auth",
			method:       "DELETE",
			path:         "/api/user",
			body:         nil,
			contentType:  "",
			authToken:    "",
			expectedCode: http.StatusUnauthorized,
			setupUser:    false,
		},
		{
			name:         "Update - No auth",
			method:       "PATCH",
			path:         "/api/user",
			body:         dto.UserUpdateRequest{Name: "Updated Name", Email: "updated@example.com"},
			contentType:  "application/json",
			authToken:    "",
			expectedCode: http.StatusUnauthorized,
			setupUser:    false,
		},
		{
			name:         "Me - No auth",
			method:       "GET",
			path:         "/api/user/me",
			body:         nil,
			contentType:  "",
			authToken:    "",
			expectedCode: http.StatusUnauthorized,
			setupUser:    false,
		},
		{
			name:         "VerifyEmail - Invalid token",
			method:       "POST",
			path:         "/api/user/verify_email",
			body:         dto.VerifyEmailRequest{Token: "invalid"},
			contentType:  "application/json",
			expectedCode: http.StatusBadRequest,
			setupUser:    false,
		},
		{
			name:         "SendVerificationEmail - Invalid email",
			method:       "POST",
			path:         "/api/user/send_verification_email",
			body:         dto.SendVerificationEmailRequest{Email: "invalid"},
			contentType:  "application/json",
			expectedCode: http.StatusBadRequest,
			setupUser:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var accessToken string
				var email string
				if tt.setupUser {
					accessToken, email = createTestUserAndGetToken(t)
					defer db.Exec("DELETE FROM users WHERE email = ?", email)
				} else {
					accessToken = tt.authToken
				}

				var req *http.Request
				var err error

				if tt.contentType == "multipart/form-data" {
					body := new(bytes.Buffer)
					writer := multipart.NewWriter(body)

					if formData, ok := tt.body.(map[string]string); ok {
						for key, value := range formData {
							if err := writer.WriteField(key, value); err != nil {
								t.Fatalf("Error writing field %s: %v", key, err)
							}
						}
					}
					writer.Close()

					req, err = http.NewRequest(tt.method, tt.path, body)
					if err != nil {
						t.Fatal(err)
					}
					req.Header.Set("Content-Type", writer.FormDataContentType())
				} else if tt.contentType == "application/json" {
					bodyBytes, err := json.Marshal(tt.body)
					if err != nil {
						t.Fatal(err)
					}
					req, err = http.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
					if err != nil {
						t.Fatal(err)
					}
					req.Header.Set("Content-Type", "application/json")
				} else {
					req, err = http.NewRequest(tt.method, tt.path, nil)
					if err != nil {
						t.Fatal(err)
					}
				}

				if accessToken != "" {
					req.Header.Set("Authorization", "Bearer "+accessToken)
				}

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code, "Status code mismatch for %s", tt.name)

				var response utils.Response
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err, "Failed to parse response for %s", tt.name)
				if tt.expectedCode == http.StatusOK {
					assert.True(t, response.Status, "Response status should be true for %s", tt.name)
				} else {
					assert.False(t, response.Status, "Response status should be false for %s", tt.name)
				}
			},
		)
	}
}
