package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/controller"
	"github.com/Caknoooo/go-gin-clean-starter/dto"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/middleware"
	"github.com/Caknoooo/go-gin-clean-starter/repository"
	"github.com/Caknoooo/go-gin-clean-starter/service"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"github.com/Caknoooo/go-gin-clean-starter/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db             *gorm.DB
	userController controller.UserController
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

	if err := db.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
	); err != nil {
		panic(fmt.Sprintf("Failed to migrate tables: %v", err))
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)
	userController = controller.NewUserController(userService)

	code := m.Run()

	if err := container.CloseDatabaseConnection(db); err != nil {
		fmt.Printf("Failed to close database connection: %v\n", err)
	}
	if err := testContainer.Stop(); err != nil {
		fmt.Printf("Failed to stop test container: %v\n", err)
	}

	os.Exit(code)
}

func TestRegister(t *testing.T) {
	// Test cases
	tests := []struct {
		name         string
		payload      dto.UserCreateRequest
		expectedCode int
		checkData    bool
	}{
		{
			name: "Success register",
			payload: dto.UserCreateRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusOK,
			checkData:    true,
		},
		{
			name: "Invalid email format",
			payload: dto.UserCreateRequest{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
			checkData:    false,
		},
		{
			name: "Password too short",
			payload: dto.UserCreateRequest{
				Name:     "Test User",
				Email:    "test2@example.com",
				Password: "short",
			},
			expectedCode: http.StatusBadRequest,
			checkData:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {

				router := gin.Default()
				router.POST("/user", userController.Register)

				body := new(bytes.Buffer)
				writer := multipart.NewWriter(body)

				writer.WriteField("name", tt.payload.Name)
				writer.WriteField("email", tt.payload.Email)
				writer.WriteField("password", tt.payload.Password)
				if tt.payload.TelpNumber != "" {
					writer.WriteField("telp_number", tt.payload.TelpNumber)
				}

				if tt.payload.Image != nil {
					part, err := writer.CreateFormFile("image", filepath.Base(tt.payload.Image.Filename))
					if err != nil {
						t.Fatal(err)
					}
					_, err = part.Write([]byte("fake image content"))
					if err != nil {
						t.Fatal(err)
					}
				}

				writer.Close()

				req, err := http.NewRequest("POST", "/user", body)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())

				rr := httptest.NewRecorder()

				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)

				if tt.checkData {
					var response struct {
						Status  bool             `json:"status"`
						Message string           `json:"message"`
						Data    dto.UserResponse `json:"data"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.True(t, response.Status)
					assert.Equal(t, tt.payload.Name, response.Data.Name)
					assert.Equal(t, tt.payload.Email, response.Data.Email)
					assert.False(t, response.Data.IsVerified)
				}

				if tt.checkData {
					db.Exec("DELETE FROM users WHERE email = ?", tt.payload.Email)
				}
			},
		)
	}
}

func TestGetAllUser(t *testing.T) {
	// First, create some test users
	testUsers := []dto.UserCreateRequest{
		{
			Name:     "Alice Johnson",
			Email:    "alice@example.com",
			Password: "password123",
		},
		{
			Name:     "Bob Smith",
			Email:    "bob@example.com",
			Password: "password123",
		},
		{
			Name:     "Charlie Brown",
			Email:    "charlie@example.com",
			Password: "password123",
		},
	}

	router := gin.Default()
	router.POST("/user", userController.Register)

	for _, user := range testUsers {

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", user.Name)
		writer.WriteField("email", user.Email)
		writer.WriteField("password", user.Password)
		writer.Close()

		req, err := http.NewRequest("POST", "/user", body)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to create test user %s: status %d, body: %s", user.Email, rr.Code, rr.Body.String())
		}
	}

	// Test cases
	tests := []struct {
		name         string
		queryParams  string
		expectedCode int
		expectedLen  int
		checkMeta    bool
	}{
		{
			name:         "Default pagination",
			queryParams:  "",
			expectedCode: http.StatusOK,
			expectedLen:  3,
			checkMeta:    true,
		},
		{
			name:         "Page 1 with 2 items per page",
			queryParams:  "page=1&per_page=2",
			expectedCode: http.StatusOK,
			expectedLen:  2,
			checkMeta:    true,
		},
		{
			name:         "Search by name",
			queryParams:  "search=Alice",
			expectedCode: http.StatusOK,
			expectedLen:  1,
			checkMeta:    false,
		},
		{
			name:         "Invalid page number",
			queryParams:  "page=abc",
			expectedCode: http.StatusBadRequest,
			expectedLen:  0,
			checkMeta:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {

				router := gin.Default()
				router.GET("/user", userController.GetAllUser)

				url := "/user"
				if tt.queryParams != "" {
					url = url + "?" + tt.queryParams
				}
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					t.Fatal(err)
				}

				rr := httptest.NewRecorder()

				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)

				if tt.expectedCode == http.StatusOK {
					var response struct {
						Status  bool                   `json:"status"`
						Message string                 `json:"message"`
						Data    []dto.UserResponse     `json:"data"`
						Meta    dto.PaginationResponse `json:"meta"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.True(t, response.Status)
					assert.Equal(t, dto.MESSAGE_SUCCESS_GET_LIST_USER, response.Message)
					assert.Len(t, response.Data, tt.expectedLen)

					if tt.checkMeta {
						assert.NotNil(t, response.Meta)
						if tt.queryParams == "" {

							assert.Equal(t, 1, response.Meta.Page)
							assert.Equal(t, 10, response.Meta.PerPage)
						} else if strings.Contains(tt.queryParams, "page=1&per_page=2") {
							assert.Equal(t, 1, response.Meta.Page)
							assert.Equal(t, 2, response.Meta.PerPage)
							assert.Equal(t, int64(2), response.Meta.MaxPage)
						}
					}
				}
			},
		)
	}

	for _, user := range testUsers {
		db.Exec("DELETE FROM users WHERE email = ?", user.Email)
	}
}

func TestMe(t *testing.T) {

	registerPayload := dto.UserCreateRequest{
		Name:     "Me Test User",
		Email:    "me_test@example.com",
		Password: "password123",
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)
	registeredUser, err := userService.Register(context.Background(), registerPayload)
	assert.NoError(t, err)

	token := jwtService.GenerateAccessToken(registeredUser.ID, registeredUser.Role)
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name         string
		setupAuth    func(t *testing.T, request *http.Request)
		expectedCode int
		checkData    bool
	}{
		{
			name: "Success get current user",
			setupAuth: func(t *testing.T, request *http.Request) {
				request.Header.Set("Authorization", "Bearer "+token)
			},
			expectedCode: http.StatusOK,
			checkData:    true,
		},
		{
			name: "Unauthorized - no token",
			setupAuth: func(t *testing.T, request *http.Request) {
				// No auth header set
			},
			expectedCode: http.StatusUnauthorized,
			checkData:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {

				router := gin.Default()

				router.Use(middleware.Authenticate(jwtService))

				router.GET("/user/me", userController.Me)

				req, err := http.NewRequest("GET", "/user/me", nil)
				if err != nil {
					t.Fatal(err)
				}

				tt.setupAuth(t, req)

				rr := httptest.NewRecorder()

				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)

				if tt.checkData {
					var response struct {
						Status  bool             `json:"status"`
						Message string           `json:"message"`
						Data    dto.UserResponse `json:"data"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.True(t, response.Status)
					assert.Equal(t, dto.MESSAGE_SUCCESS_GET_USER, response.Message)
					assert.Equal(t, registeredUser.ID, response.Data.ID)
					assert.Equal(t, registerPayload.Name, response.Data.Name)
					assert.Equal(t, registerPayload.Email, response.Data.Email)
				}
			},
		)
	}

	db.Exec("DELETE FROM users WHERE email = ?", registerPayload.Email)
}

func TestLogin(t *testing.T) {
	userService := service.NewUserService(
		repository.NewUserRepository(db),
		repository.NewRefreshTokenRepository(db),
		service.NewJWTService(),
		db,
	)
	userController := controller.NewUserController(userService)

	testUser := dto.UserCreateRequest{
		Name:     "Login Test User",
		Email:    "login_test@example.com",
		Password: "password123",
	}

	userBytes, err := json.Marshal(testUser)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.Default()
	router.POST("/user/register", userController.Register)

	registerReq, err := http.NewRequest("POST", "/user/register", bytes.NewBuffer(userBytes))
	if err != nil {
		t.Fatal(err)
	}
	registerReq.Header.Set("Content-Type", "application/json")

	registerRec := httptest.NewRecorder()

	router.ServeHTTP(registerRec, registerReq)

	if registerRec.Code != http.StatusOK {
		t.Fatalf("Failed to register test user: %v", registerRec.Body.String())
	}

	// Test cases
	tests := []struct {
		name         string
		payload      dto.UserLoginRequest
		expectedCode int
		checkTokens  bool
	}{
		{
			name: "Success login",
			payload: dto.UserLoginRequest{
				Email:    "login_test@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusOK,
			checkTokens:  true,
		},
		{
			name: "Invalid email",
			payload: dto.UserLoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
			checkTokens:  false,
		},
		{
			name: "Wrong password",
			payload: dto.UserLoginRequest{
				Email:    "login_test@example.com",
				Password: "wrongpassword",
			},
			expectedCode: http.StatusBadRequest,
			checkTokens:  false,
		},
		{
			name: "Missing email",
			payload: dto.UserLoginRequest{
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
			checkTokens:  false,
		},
		{
			name: "Missing password",
			payload: dto.UserLoginRequest{
				Email: "login_test@example.com",
			},
			expectedCode: http.StatusBadRequest,
			checkTokens:  false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {

				router := gin.Default()
				router.POST("/user/login", userController.Login)

				payloadBytes, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest("POST", "/user/login", bytes.NewBuffer(payloadBytes))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()

				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)

				if tt.checkTokens {
					var response struct {
						Status  bool              `json:"status"`
						Message string            `json:"message"`
						Data    dto.TokenResponse `json:"data"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.True(t, response.Status)
					assert.Equal(t, dto.MESSAGE_SUCCESS_LOGIN, response.Message)
					assert.NotEmpty(t, response.Data.AccessToken)
					assert.NotEmpty(t, response.Data.RefreshToken)
				} else if tt.expectedCode == http.StatusBadRequest {

					var response struct {
						Status  bool   `json:"status"`
						Message string `json:"message"`
						Error   string `json:"error"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.False(t, response.Status)
					assert.NotEmpty(t, response.Message)
				}
			},
		)
	}

	db.Exec("DELETE FROM users WHERE email = ?", "login_test@example.com")
}

func TestSendVerificationEmail(t *testing.T) {
	testUser := dto.UserCreateRequest{
		Name:     "Verification Test User",
		Email:    "verification_test@example.com",
		Password: "password123",
	}

	userRepo := repository.NewUserRepository(db)
	_, err := userRepo.Register(
		context.Background(), nil, entity.User{
			Name:       testUser.Name,
			Email:      testUser.Email,
			Password:   testUser.Password,
			Role:       "user",
			IsVerified: false,
		},
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name         string
		payload      dto.SendVerificationEmailRequest
		expectedCode int
		wantSuccess  bool
	}{
		{
			name: "Success send verification email",
			payload: dto.SendVerificationEmailRequest{
				Email: "verification_test@example.com",
			},
			expectedCode: http.StatusOK,
			wantSuccess:  true,
		},
		{
			name: "Invalid email format",
			payload: dto.SendVerificationEmailRequest{
				Email: "invalid-email",
			},
			expectedCode: http.StatusBadRequest,
			wantSuccess:  false,
		},
		{
			name: "Email not registered",
			payload: dto.SendVerificationEmailRequest{
				Email: "not_registered@example.com",
			},
			expectedCode: http.StatusBadRequest,
			wantSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				router.POST("/user/send_verification_email", userController.SendVerificationEmail)

				payloadBytes, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest("POST", "/user/send_verification_email", bytes.NewBuffer(payloadBytes))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)

				var response struct {
					Status  bool        `json:"status"`
					Message string      `json:"message"`
					Data    interface{} `json:"data"`
				}
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, tt.wantSuccess, response.Status)

				if tt.wantSuccess {
					assert.Equal(t, dto.MESSAGE_SEND_VERIFICATION_EMAIL_SUCCESS, response.Message)
				} else {
					assert.NotEmpty(t, response.Message)
				}
			},
		)
	}

	db.Exec("DELETE FROM users WHERE email = ?", "verification_test@example.com")
}

func TestVerifyEmail(t *testing.T) {
	testContainer, err := container.StartTestContainer()
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer func() {
		if err := testContainer.Stop(); err != nil {
			t.Fatalf("Failed to stop test container: %v", err)
		}
	}()

	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_PORT", testContainer.Port)

	db := container.SetUpDatabaseConnection()
	defer func() {
		if err := container.CloseDatabaseConnection(db); err != nil {
			t.Fatalf("Failed to close database connection: %v", err)
		}
	}()

	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)
	userController := controller.NewUserController(userService)

	registerReq := dto.UserCreateRequest{
		Name:     "Test Verify User",
		Email:    "verify@example.com",
		Password: "password123",
	}

	registeredUser, err := userService.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	expired := time.Now().Add(time.Hour * 24).Format("2006-01-02 15:04:05")
	plainText := registeredUser.Email + "_" + expired
	token, err := utils.AESEncrypt(plainText)
	if err != nil {
		t.Fatalf("Failed to encrypt verification token: %v", err)
	}

	tests := []struct {
		name         string
		payload      dto.VerifyEmailRequest
		expectedCode int
		checkData    bool
	}{
		{
			name: "Success verify email",
			payload: dto.VerifyEmailRequest{
				Token: token,
			},
			expectedCode: http.StatusOK,
			checkData:    true,
		},
		{
			name: "Empty token",
			payload: dto.VerifyEmailRequest{
				Token: "",
			},
			expectedCode: http.StatusBadRequest,
			checkData:    false,
		},
		{
			name: "Invalid token",
			payload: dto.VerifyEmailRequest{
				Token: "invalid-token",
			},
			expectedCode: http.StatusBadRequest,
			checkData:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				router.POST("/user/verify_email", userController.VerifyEmail)

				reqBody, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}

				req, err := http.NewRequest("POST", "/user/verify_email", bytes.NewBuffer(reqBody))
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code, "Status code mismatch for test: %s", tt.name)

				if tt.checkData {
					var response struct {
						Status  bool                    `json:"status"`
						Message string                  `json:"message"`
						Data    dto.VerifyEmailResponse `json:"data"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err, "Failed to unmarshal response for test: %s", tt.name)
					assert.True(t, response.Status, "Response status should be true for test: %s", tt.name)
					assert.Equal(
						t,
						dto.MESSAGE_SUCCESS_VERIFY_EMAIL,
						response.Message,
						"Response message mismatch for test: %s",
						tt.name,
					)
					assert.True(
						t,
						response.Data.IsVerified,
						"User should be verified in response for test: %s",
						tt.name,
					)

					user, err := userService.GetUserById(context.Background(), registeredUser.ID)
					assert.NoError(t, err, "Failed to fetch user from database for test: %s", tt.name)
					assert.True(t, user.IsVerified, "User should be verified in database for test: %s", tt.name)
				}
			},
		)
	}

	err = db.Exec("DELETE FROM users WHERE email = ?", registerReq.Email).Error
	if err != nil {
		t.Fatalf("Failed to clean up test user: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	testContainer, err := container.StartTestContainer()
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer func() {
		if err := testContainer.Stop(); err != nil {
			t.Fatalf("Failed to stop test container: %v", err)
		}
	}()

	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_PORT", testContainer.Port)

	db := container.SetUpDatabaseConnection()
	defer func() {
		if err := container.CloseDatabaseConnection(db); err != nil {
			t.Fatalf("Failed to close database connection: %v", err)
		}
	}()

	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)
	userController := controller.NewUserController(userService)

	registerReq := dto.UserCreateRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	registeredUser, err := userService.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	token := jwtService.GenerateAccessToken(registeredUser.ID, registeredUser.Role)

	tests := []struct {
		name         string
		payload      dto.UserUpdateRequest
		userID       string
		token        string
		expectedCode int
		checkData    bool
	}{
		{
			name: "Success update user",
			payload: dto.UserUpdateRequest{
				Name:       "Updated User",
				TelpNumber: "1234567890",
				Email:      "updated@example.com",
			},
			userID:       registeredUser.ID,
			token:        token,
			expectedCode: http.StatusOK,
			checkData:    true,
		},
		{
			name: "Invalid email format",
			payload: dto.UserUpdateRequest{
				Name:       "Updated User",
				TelpNumber: "1234567890",
				Email:      "invalid-email",
			},
			userID:       registeredUser.ID,
			token:        token,
			expectedCode: http.StatusBadRequest,
			checkData:    false,
		},
		{
			name: "Unauthorized - missing token",
			payload: dto.UserUpdateRequest{
				Name:       "Updated User",
				TelpNumber: "1234567890",
				Email:      "updated@example.com",
			},
			userID:       registeredUser.ID,
			token:        "",
			expectedCode: http.StatusUnauthorized,
			checkData:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				router.Use(middleware.Authenticate(jwtService))
				router.PATCH("/user", userController.Update)

				reqBody, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}

				req, err := http.NewRequest("PATCH", "/user", bytes.NewBuffer(reqBody))
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
				req.Header.Set("Content-Type", "application/json")
				if tt.token != "" {
					req.Header.Set("Authorization", "Bearer "+tt.token)
				}

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code, "Status code mismatch for test: %s", tt.name)

				if tt.checkData {
					var response struct {
						Status  bool                   `json:"status"`
						Message string                 `json:"message"`
						Data    dto.UserUpdateResponse `json:"data"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err, "Failed to unmarshal response for test: %s", tt.name)
					assert.True(t, response.Status, "Response status should be true for test: %s", tt.name)
					assert.Equal(
						t,
						dto.MESSAGE_SUCCESS_UPDATE_USER,
						response.Message,
						"Response message mismatch for test: %s",
						tt.name,
					)
					assert.Equal(
						t,
						tt.payload.Name,
						response.Data.Name,
						"Name mismatch in response for test: %s",
						tt.name,
					)
					assert.Equal(
						t,
						tt.payload.TelpNumber,
						response.Data.TelpNumber,
						"TelpNumber mismatch in response for test: %s",
						tt.name,
					)
					assert.Equal(
						t,
						tt.payload.Email,
						response.Data.Email,
						"Email mismatch in response for test: %s",
						tt.name,
					)

					user, err := userService.GetUserById(context.Background(), registeredUser.ID)
					assert.NoError(t, err, "Failed to fetch user from database for test: %s", tt.name)
					assert.Equal(t, tt.payload.Name, user.Name, "Name mismatch in database for test: %s", tt.name)
					assert.Equal(
						t,
						tt.payload.TelpNumber,
						user.TelpNumber,
						"TelpNumber mismatch in database for test: %s",
						tt.name,
					)
					assert.Equal(t, tt.payload.Email, user.Email, "Email mismatch in database for test: %s", tt.name)
				}
			},
		)
	}

	err = db.Exec("DELETE FROM users WHERE email = ?", registerReq.Email).Error
	if err != nil {
		t.Fatalf("Failed to clean up test user: %v", err)
	}
}

func TestDelete(t *testing.T) {
	registerReq := dto.UserCreateRequest{
		Name:     "Delete Test User",
		Email:    "delete_test@example.com",
		Password: "password123",
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)
	registeredUser, err := userService.Register(context.Background(), registerReq)
	assert.NoError(t, err)

	token := jwtService.GenerateAccessToken(registeredUser.ID, registeredUser.Role)

	tests := []struct {
		name         string
		setupAuth    func(t *testing.T, request *http.Request)
		expectedCode int
		checkData    bool
	}{
		{
			name: "Success delete user",
			setupAuth: func(t *testing.T, request *http.Request) {
				request.Header.Set("Authorization", "Bearer "+token)
			},
			expectedCode: http.StatusOK,
			checkData:    true,
		},
		{
			name: "Unauthorized - no token",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			expectedCode: http.StatusUnauthorized,
			checkData:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				router.Use(middleware.Authenticate(jwtService))
				router.DELETE("/user", userController.Delete)

				req, err := http.NewRequest("DELETE", "/user", nil)
				if err != nil {
					t.Fatal(err)
				}

				tt.setupAuth(t, req)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)

				if tt.checkData {
					var response struct {
						Status  bool        `json:"status"`
						Message string      `json:"message"`
						Data    interface{} `json:"data"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.True(t, response.Status)
					assert.Equal(t, dto.MESSAGE_SUCCESS_DELETE_USER, response.Message)

					_, err := userService.GetUserById(context.Background(), registeredUser.ID)
					assert.Error(t, err)
				}
			},
		)
	}

	db.Exec("DELETE FROM users WHERE email = ?", registerReq.Email)
}

func TestRefreshToken(t *testing.T) {
	registerReq := dto.UserCreateRequest{
		Name:     "Refresh Test User",
		Email:    "refresh_test@example.com",
		Password: "password123",
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	_, err := userService.Register(context.Background(), registerReq)
	assert.NoError(t, err)

	loginReq := dto.UserLoginRequest{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}
	loginRes, err := userService.Verify(context.Background(), loginReq)
	assert.NoError(t, err)
	refreshToken := loginRes.RefreshToken

	tests := []struct {
		name         string
		payload      dto.RefreshTokenRequest
		expectedCode int
		checkData    bool
	}{
		{
			name: "Success refresh token",
			payload: dto.RefreshTokenRequest{
				RefreshToken: refreshToken,
			},
			expectedCode: http.StatusOK,
			checkData:    true,
		},
		{
			name: "Invalid refresh token",
			payload: dto.RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			expectedCode: http.StatusUnauthorized,
			checkData:    false,
		},
		{
			name:         "Empty refresh token",
			payload:      dto.RefreshTokenRequest{},
			expectedCode: http.StatusBadRequest,
			checkData:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				router.POST("/user/refresh", userController.Refresh)

				payloadBytes, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest("POST", "/user/refresh", bytes.NewBuffer(payloadBytes))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)

				if tt.checkData {
					var response struct {
						Status  bool              `json:"status"`
						Message string            `json:"message"`
						Data    dto.TokenResponse `json:"data"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.True(t, response.Status)
					assert.Equal(t, dto.MESSAGE_SUCCESS_REFRESH_TOKEN, response.Message)
					assert.NotEmpty(t, response.Data.AccessToken)
					assert.NotEmpty(t, response.Data.RefreshToken)
				} else if tt.expectedCode == http.StatusBadRequest {
					var response struct {
						Status  bool   `json:"status"`
						Message string `json:"message"`
					}
					err = json.Unmarshal(rr.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.False(t, response.Status)
					assert.NotEmpty(t, response.Message)
				}
			},
		)
	}

	db.Exec("DELETE FROM users WHERE email = ?", registerReq.Email)
	db.Exec("DELETE FROM refresh_tokens WHERE token = ?", refreshToken)
}
