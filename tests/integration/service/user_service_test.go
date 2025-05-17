package service_test

import (
	"context"
	"errors"
	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/Caknoooo/go-gin-clean-starter/helpers"
	"github.com/Caknoooo/go-gin-clean-starter/service"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/dto"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/repository"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"github.com/Caknoooo/go-gin-clean-starter/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/gomail.v2"
)

// MockJWTService for testing
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(userID, role string) string {
	args := m.Called(userID, role)
	return args.String(0)
}

func (m *MockJWTService) GenerateRefreshToken() (string, time.Time) {
	args := m.Called()
	return args.String(0), args.Get(1).(time.Time)
}

func (m *MockJWTService) ValidateToken(token string) (*jwt.Token, error) {
	args := m.Called(token)
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockJWTService) GetUserIDByToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

// MockDialer for email sending
type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialAndSend(messages ...*gomail.Message) error {
	args := m.Called(messages)
	return args.Error(0)
}

func TestUserService_Register(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	err = os.Setenv("SMTP_HOST", "smtp.example.com")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_EMAIL", "user@example.com")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_PASSWORD", "password123")
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Mock JWT service
	jwtService := &MockJWTService{}

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Create temporary email template file
	tempDir := t.TempDir()
	emailTemplatePath := filepath.Join(tempDir, "base_mail.html")
	err = os.WriteFile(
		emailTemplatePath, []byte(`
		<html>
			<body>
				<p>Hello {{.Email}}</p>
				<a href="{{.Verify}}">Verify Email</a>
			</body>
		</html>
	`), 0644,
	)
	assert.NoError(t, err)

	// Mock email sending by overriding newDialer
	originalNewDialer := utils.NewDialer
	utils.NewDialer = func(host string, port int, username, password string) utils.Dialer {
		dialer := &MockDialer{}
		dialer.On("DialAndSend", mock.Anything).Return(nil)
		return dialer
	}
	defer func() { utils.NewDialer = originalNewDialer }()

	// Mock UploadFile
	originalPath := utils.PATH
	utils.PATH = tempDir
	defer func() { utils.PATH = originalPath }()

	// Test cases
	tests := []struct {
		name          string
		input         dto.UserCreateRequest
		setup         func()
		expectedError error
		validateUser  func(t *testing.T, user dto.UserResponse)
	}{
		{
			name: "Successful Registration",
			input: dto.UserCreateRequest{
				Name:       "John Doe",
				Email:      "john.doe@example.com",
				Password:   "password123",
				TelpNumber: "1234567890",
			},
			setup:         func() {},
			expectedError: nil,
			validateUser: func(t *testing.T, user dto.UserResponse) {
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, "John Doe", user.Name)
				assert.Equal(t, "john.doe@example.com", user.Email)
				assert.Equal(t, "1234567890", user.TelpNumber)
				assert.Equal(t, constants.ENUM_ROLE_USER, user.Role)
				assert.False(t, user.IsVerified)

				// Verify user exists in database
				var dbUser entity.User
				err := db.Where("email = ?", user.Email).First(&dbUser).Error
				assert.NoError(t, err)
				assert.Equal(t, user.Name, dbUser.Name)
				assert.Equal(t, user.Email, dbUser.Email)
			},
		},
		{
			name: "Email Already Exists",
			input: dto.UserCreateRequest{
				Name:       "Jane Doe",
				Email:      "jane.doe@example.com",
				Password:   "password123",
				TelpNumber: "0987654321",
			},
			setup: func() {
				// Create existing user
				existingUser := entity.User{
					ID:         uuid.New(),
					Name:       "Existing User",
					Email:      "jane.doe@example.com",
					Password:   "hashedpassword",
					TelpNumber: "1234567890",
					Role:       constants.ENUM_ROLE_USER,
				}
				db.Create(&existingUser)
			},
			expectedError: dto.ErrEmailAlreadyExists,
			validateUser: func(t *testing.T, user dto.UserResponse) {
				assert.Empty(t, user.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Run setup
				tt.setup()

				// Execute registration
				user, err := userService.Register(context.Background(), tt.input)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.NoError(t, err)
				}

				// Validate user response
				tt.validateUser(t, user)
			},
		)
	}
}

func TestUserService_GetAllUserWithPagination(t *testing.T) {
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	err = os.Setenv("SMTP_HOST", "smtp.example.com")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_EMAIL", "user@example.com")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_PASSWORD", "password123")
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Clean up database after test
	defer func() {
		db.Exec("DELETE FROM users WHERE TRUE")
	}()

	// Create test users
	ctx := context.Background()
	testUsers := []entity.User{
		{
			ID:         uuid.New(),
			Name:       "John Doe",
			Email:      "john@test.com",
			Password:   "password123",
			Role:       "user",
			TelpNumber: "1234567890",
			IsVerified: true,
		},
		{
			ID:         uuid.New(),
			Name:       "Jane Smith",
			Email:      "jane@test.com",
			Password:   "password123",
			Role:       "user",
			TelpNumber: "1234567891",
			IsVerified: true,
		},
		{
			ID:         uuid.New(),
			Name:       "Admin User",
			Email:      "admin@test.com",
			Password:   "password123",
			Role:       "admin",
			TelpNumber: "1234567892",
			IsVerified: true,
		},
	}

	for _, user := range testUsers {
		_, err := userRepo.Register(ctx, nil, user)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

	tests := []struct {
		name          string
		req           dto.PaginationRequest
		expectedCount int
		expectedError bool
	}{
		{
			name: "Get all users with default pagination",
			req: dto.PaginationRequest{
				Page:    1,
				PerPage: 10,
			},
			expectedCount: 3,
			expectedError: false,
		},
		{
			name: "Get first page with 2 users per page",
			req: dto.PaginationRequest{
				Page:    1,
				PerPage: 2,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "Get second page with 2 users per page",
			req: dto.PaginationRequest{
				Page:    2,
				PerPage: 2,
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "Search by name - exact match",
			req: dto.PaginationRequest{
				Page:    1,
				PerPage: 10,
				Search:  "John Doe",
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "Search by name - partial match",
			req: dto.PaginationRequest{
				Page:    1,
				PerPage: 10,
				Search:  "Jane",
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "Search by name - no match",
			req: dto.PaginationRequest{
				Page:    1,
				PerPage: 10,
				Search:  "Nonexistent",
			},
			expectedCount: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result, err := userService.GetAllUserWithPagination(ctx, tt.req)

				if tt.expectedError {
					assert.Error(t, err)
					return
				}

				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result.Data))
				assert.Equal(t, tt.req.Page, result.Page)
				assert.Equal(t, tt.req.PerPage, result.PerPage)

				// Verify that the returned users match our test data
				for _, user := range result.Data {
					found := false
					for _, testUser := range testUsers {
						if user.ID == testUser.ID.String() {
							found = true
							assert.Equal(t, testUser.Name, user.Name)
							assert.Equal(t, testUser.Email, user.Email)
							assert.Equal(t, testUser.Role, user.Role)
							break
						}
					}
					assert.True(t, found, "User not found in test data")
				}
			},
		)
	}
}

func TestUserService_GetUserById(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Clean up database after test
	defer func() {
		db.Exec("DELETE FROM users WHERE TRUE")
	}()

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		setup         func() string // returns user ID
		expectedError error
		validate      func(t *testing.T, user dto.UserResponse)
	}{
		{
			name: "Successfully get user by ID",
			setup: func() string {
				// Create test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)
				return createdUser.ID.String()
			},
			expectedError: nil,
			validate: func(t *testing.T, user dto.UserResponse) {
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, "Test User", user.Name)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "1234567890", user.TelpNumber)
				assert.Equal(t, "user", user.Role)
				assert.True(t, user.IsVerified)
			},
		},
		{
			name: "User not found",
			setup: func() string {
				return uuid.New().String() // non-existent ID
			},
			expectedError: dto.ErrGetUserById,
			validate: func(t *testing.T, user dto.UserResponse) {
				assert.Empty(t, user.ID)
			},
		},
		{
			name: "Invalid UUID format",
			setup: func() string {
				return "invalid-uuid-format"
			},
			expectedError: dto.ErrGetUserById,
			validate: func(t *testing.T, user dto.UserResponse) {
				assert.Empty(t, user.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Setup test data and get user ID
				userId := tt.setup()

				// Execute the function
				user, err := userService.GetUserById(ctx, userId)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.NoError(t, err)
				}

				// Validate user response
				tt.validate(t, user)
			},
		)
	}
}

func TestUserService_GetUserByEmail(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Clean up database after test
	defer func() {
		db.Exec("DELETE FROM users WHERE TRUE")
	}()

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		setup         func() string // returns email to test
		expectedError error
		validate      func(t *testing.T, user dto.UserResponse)
	}{
		{
			name: "Successfully get user by email",
			setup: func() string {
				// Create test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)
				return createdUser.Email
			},
			expectedError: nil,
			validate: func(t *testing.T, user dto.UserResponse) {
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, "Test User", user.Name)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "1234567890", user.TelpNumber)
				assert.Equal(t, "user", user.Role)
				assert.True(t, user.IsVerified)
			},
		},
		{
			name: "User not found by email",
			setup: func() string {
				return "nonexistent@example.com" // non-existent email
			},
			expectedError: dto.ErrGetUserByEmail,
			validate: func(t *testing.T, user dto.UserResponse) {
				assert.Empty(t, user.ID)
			},
		},
		{
			name: "Empty email",
			setup: func() string {
				return "" // empty email
			},
			expectedError: dto.ErrGetUserByEmail,
			validate: func(t *testing.T, user dto.UserResponse) {
				assert.Empty(t, user.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Setup test data and get email
				email := tt.setup()

				// Execute the function
				user, err := userService.GetUserByEmail(ctx, email)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.NoError(t, err)
				}

				// Validate user response
				tt.validate(t, user)
			},
		)
	}
}

func TestUserService_SendVerificationEmail(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Create temporary email template file
	tempDir := t.TempDir()
	emailTemplatePath := filepath.Join(tempDir, "base_mail.html")
	err = os.WriteFile(
		emailTemplatePath, []byte(`
		<html>
			<body>
				<p>Hello {{.Email}}</p>
				<a href="{{.Verify}}">Verify Email</a>
			</body>
		</html>
	`), 0644,
	)
	assert.NoError(t, err)

	// Mock email sending by overriding newDialer
	originalNewDialer := utils.NewDialer
	mockDialer := &MockDialer{}
	utils.NewDialer = func(host string, port int, username, password string) utils.Dialer {
		return mockDialer
	}
	defer func() { utils.NewDialer = originalNewDialer }()

	// Mock template path
	originalPath := utils.PATH
	utils.PATH = tempDir
	defer func() { utils.PATH = originalPath }()

	// Test cases
	tests := []struct {
		name          string
		setup         func() dto.SendVerificationEmailRequest
		mockEmail     func(*MockDialer)
		expectedError error
	}{
		{
			name: "Successfully send verification email",
			setup: func() dto.SendVerificationEmailRequest {
				// Create test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: false,
				}
				_, err := userRepo.Register(context.Background(), nil, user)
				assert.NoError(t, err)
				return dto.SendVerificationEmailRequest{Email: "test@example.com"}
			},
			mockEmail: func(m *MockDialer) {
				m.On("DialAndSend", mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Email not found",
			setup: func() dto.SendVerificationEmailRequest {
				return dto.SendVerificationEmailRequest{Email: "nonexistent@example.com"}
			},
			mockEmail:     func(m *MockDialer) {},
			expectedError: dto.ErrEmailNotFound,
		},
		{
			name: "Email sending fails",
			setup: func() dto.SendVerificationEmailRequest {
				// Create test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test2@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: false,
				}
				_, err := userRepo.Register(context.Background(), nil, user)
				assert.NoError(t, err)
				return dto.SendVerificationEmailRequest{Email: "test2@example.com"}
			},
			mockEmail: func(m *MockDialer) {
				m.On("DialAndSend", mock.Anything).Return(errors.New("smtp error")).Once()
			},
			expectedError: errors.New("smtp error"),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Reset the template file to valid state before each test
				err := os.WriteFile(
					emailTemplatePath, []byte(`
                <html>
                    <body>
                        <p>Hello {{.Email}}</p>
                        <a href="{{.Verify}}">Verify Email</a>
                    </body>
                </html>
            `), 0644,
				)
				assert.NoError(t, err)

				// Create fresh mock for each test
				mockDialer := &MockDialer{}
				utils.NewDialer = func(host string, port int, username, password string) utils.Dialer {
					return mockDialer
				}

				// Setup test data
				req := tt.setup()

				// Setup mock expectations
				tt.mockEmail(mockDialer)

				// Execute the function
				err = userService.SendVerificationEmail(context.Background(), req)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					if tt.name == "Template parsing fails" {
						assert.ErrorContains(t, err, "template")
					} else {
						assert.Equal(t, tt.expectedError, err)
					}
				} else {
					assert.NoError(t, err)
				}

				// Verify mock expectations
				mockDialer.AssertExpectations(t)
			},
		)
	}
}

func TestUserService_VerifyEmail(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Set AES encryption key for testing
	originalKey := utils.KEY
	utils.KEY = "6368616e676520746869732070617373776f726420746f206120736563726574" // test key
	defer func() { utils.KEY = originalKey }()

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Clean up database after test
	defer func() {
		db.Exec("DELETE FROM users WHERE TRUE")
	}()

	// Create test context
	ctx := context.Background()

	// Helper function to create a test token
	createTestToken := func(email string, hoursToAdd time.Duration) string {
		expired := time.Now().Add(hoursToAdd).Format("2006-01-02 15:04:05")
		plainText := email + "_" + expired
		token, err := utils.AESEncrypt(plainText)
		assert.NoError(t, err)
		return token
	}

	// Test cases
	tests := []struct {
		name          string
		setup         func() (string, string) // returns (email, token)
		expectedError error
		validate      func(t *testing.T, response dto.VerifyEmailResponse, email string)
	}{
		{
			name: "Successfully verify email",
			setup: func() (string, string) {
				// Create unverified test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: false,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				// Create valid token
				token := createTestToken(createdUser.Email, 24*time.Hour)
				return createdUser.Email, token
			},
			expectedError: nil,
			validate: func(t *testing.T, response dto.VerifyEmailResponse, email string) {
				assert.Equal(t, email, response.Email)
				assert.True(t, response.IsVerified)

				// Verify user is updated in database
				dbUser, err := userRepo.GetUserByEmail(ctx, nil, email)
				assert.NoError(t, err)
				assert.True(t, dbUser.IsVerified)
			},
		},
		{
			name: "Expired token",
			setup: func() (string, string) {
				// Create unverified test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: false,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				// Create expired token
				token := createTestToken(createdUser.Email, -24*time.Hour)
				return createdUser.Email, token
			},
			expectedError: dto.ErrTokenExpired,
			validate: func(t *testing.T, response dto.VerifyEmailResponse, email string) {
				assert.Equal(t, email, response.Email)
				assert.False(t, response.IsVerified)

				// Verify user is not updated in database
				dbUser, err := userRepo.GetUserByEmail(ctx, nil, email)
				assert.NoError(t, err)
				assert.False(t, dbUser.IsVerified)
			},
		},
		{
			name: "Invalid token format",
			setup: func() (string, string) {
				return "test@example.com", "invalid_token_format"
			},
			expectedError: dto.ErrTokenInvalid,
			validate: func(t *testing.T, response dto.VerifyEmailResponse, email string) {
				assert.Empty(t, response.Email)
				assert.False(t, response.IsVerified)
			},
		},
		{
			name: "Already verified account",
			setup: func() (string, string) {
				// Create verified test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				// Create valid token
				token := createTestToken(createdUser.Email, 24*time.Hour)
				return createdUser.Email, token
			},
			expectedError: dto.ErrAccountAlreadyVerified,
			validate: func(t *testing.T, response dto.VerifyEmailResponse, email string) {
				assert.Empty(t, response.Email)
				assert.False(t, response.IsVerified)
			},
		},
		{
			name: "User not found",
			setup: func() (string, string) {
				email := "nonexistent@example.com"
				token := createTestToken(email, 24*time.Hour)
				return email, token
			},
			expectedError: dto.ErrUserNotFound,
			validate: func(t *testing.T, response dto.VerifyEmailResponse, email string) {
				assert.Empty(t, response.Email)
				assert.False(t, response.IsVerified)
			},
		},
		{
			name: "Malformed token content",
			setup: func() (string, string) {
				// Create a token without the expected format (email_expiry)
				plainText := "malformed_content"
				token, err := utils.AESEncrypt(plainText)
				assert.NoError(t, err)
				return "", token
			},
			expectedError: dto.ErrTokenInvalid,
			validate: func(t *testing.T, response dto.VerifyEmailResponse, email string) {
				assert.Empty(t, response.Email)
				assert.False(t, response.IsVerified)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Setup test data and get email and token
				email, token := tt.setup()

				// Execute the function
				response, err := userService.VerifyEmail(
					ctx, dto.VerifyEmailRequest{
						Token: token,
					},
				)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.NoError(t, err)
				}

				// Validate response
				tt.validate(t, response, email)
			},
		)
	}
}

func TestUserService_Update(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Clean up database after test
	defer func() {
		db.Exec("DELETE FROM users WHERE TRUE")
	}()

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		setup         func() (string, dto.UserUpdateRequest) // returns user ID and update request
		expectedError error
		validate      func(t *testing.T, response dto.UserUpdateResponse, db *gorm.DB)
	}{
		{
			name: "Successfully update user",
			setup: func() (string, dto.UserUpdateRequest) {
				// Create test user
				user := entity.User{
					Name:       "Original Name",
					Email:      "original@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				updateReq := dto.UserUpdateRequest{
					Name:       "Updated Name",
					Email:      "updated@example.com",
					TelpNumber: "0987654321",
				}

				return createdUser.ID.String(), updateReq
			},
			expectedError: nil,
			validate: func(t *testing.T, response dto.UserUpdateResponse, db *gorm.DB) {
				assert.Equal(t, "Updated Name", response.Name)
				assert.Equal(t, "updated@example.com", response.Email)
				assert.Equal(t, "0987654321", response.TelpNumber)
				assert.Equal(t, "user", response.Role) // Role shouldn't change
				assert.True(t, response.IsVerified)    // IsVerified shouldn't change

				// Verify the changes were persisted in the database
				var dbUser entity.User
				err := db.First(&dbUser, "id = ?", response.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, "Updated Name", dbUser.Name)
				assert.Equal(t, "updated@example.com", dbUser.Email)
				assert.Equal(t, "0987654321", dbUser.TelpNumber)
			},
		},
		{
			name: "Update non-existent user",
			setup: func() (string, dto.UserUpdateRequest) {
				return uuid.New().String(), dto.UserUpdateRequest{
					Name:       "Should Fail",
					Email:      "fail@example.com",
					TelpNumber: "0000000000",
				}
			},
			expectedError: dto.ErrUserNotFound,
			validate: func(t *testing.T, response dto.UserUpdateResponse, db *gorm.DB) {
				assert.Empty(t, response.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Setup test data and get user ID and update request
				userId, updateReq := tt.setup()

				// Execute the function
				response, err := userService.Update(ctx, updateReq, userId)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.NoError(t, err)
				}

				// Validate response and database state
				tt.validate(t, response, db)
			},
		)
	}
}

func TestUserService_Delete(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Create test context
	ctx := context.Background()

	tests := []struct {
		name          string
		setup         func() string // returns user ID to delete
		expectedError error
		verify        func(t *testing.T, userId string)
	}{
		{
			name: "Successfully delete user",
			setup: func() string {
				// Create test user
				user := entity.User{
					Name:       "User to Delete",
					Email:      "delete@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)
				return createdUser.ID.String()
			},
			expectedError: nil,
			verify: func(t *testing.T, userId string) {
				// Verify user is deleted
				_, err := userRepo.GetUserById(ctx, nil, userId)
				assert.Error(t, err)
				assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

				// Verify transaction was committed by checking if refresh tokens are gone
				var count int64
				db.Model(&entity.RefreshToken{}).Where("user_id = ?", userId).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name: "User not found",
			setup: func() string {
				return uuid.New().String() // non-existent ID
			},
			expectedError: dto.ErrUserNotFound,
			verify: func(t *testing.T, userId string) {
				// No verification needed for this case
			},
		},
		{
			name: "Invalid UUID format",
			setup: func() string {
				return "invalid-uuid" // malformed ID
			},
			expectedError: dto.ErrUserNotFound,
			verify: func(t *testing.T, userId string) {
				// No verification needed for this case
			},
		},
		{
			name: "Delete user with refresh tokens",
			setup: func() string {
				// Create test user
				user := entity.User{
					Name:       "User With Tokens",
					Email:      "withtokens@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				// Create refresh tokens for user
				tokens := []entity.RefreshToken{
					{
						UserID:    createdUser.ID,
						Token:     "token1",
						ExpiresAt: time.Now().Add(time.Hour),
					},
					{
						UserID:    createdUser.ID,
						Token:     "token2",
						ExpiresAt: time.Now().Add(time.Hour),
					},
				}

				for _, token := range tokens {
					_, err := refreshTokenRepo.Create(ctx, nil, token)
					assert.NoError(t, err)
				}

				return createdUser.ID.String()
			},
			expectedError: nil,
			verify: func(t *testing.T, userId string) {
				// Verify user is deleted
				_, err := userRepo.GetUserById(ctx, nil, userId)
				assert.Error(t, err)
				assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

				// Verify refresh tokens are also deleted
				var count int64
				db.Model(&entity.RefreshToken{}).Where("user_id = ?", userId).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
				db.Exec("TRUNCATE TABLE refresh_tokens RESTART IDENTITY CASCADE")

				// Setup test data and get user ID
				userId := tt.setup()

				// Execute the function
				err := userService.Delete(ctx, userId)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.NoError(t, err)
				}

				// Run verification
				tt.verify(t, userId)
			},
		)
	}
}

func TestUserService_Verify(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Mock JWT service
	mockJWTService := &MockJWTService{}
	mockJWTService.On("GenerateAccessToken", mock.Anything, mock.Anything).Return("mock-access-token")
	mockJWTService.On("GenerateRefreshToken").Return("mock-refresh-token", time.Now().Add(24*time.Hour))

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, mockJWTService, db)

	// Clean up database after test
	defer func() {
		db.Exec("DELETE FROM refresh_tokens WHERE TRUE")
		db.Exec("DELETE FROM users WHERE TRUE")
	}()

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		setup         func() dto.UserLoginRequest
		expectedError string
		validate      func(t *testing.T, tokens dto.TokenResponse)
	}{
		{
			name: "Successful verification",
			setup: func() dto.UserLoginRequest {
				// Create test user
				password := "correctpassword"
				assert.NoError(t, err)

				user := entity.User{
					Name:       "Verified User",
					Email:      "verified@example.com",
					Password:   password,
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				_, err = userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				return dto.UserLoginRequest{
					Email:    "verified@example.com",
					Password: "correctpassword",
				}
			},
			expectedError: "",
			validate: func(t *testing.T, tokens dto.TokenResponse) {
				assert.NotEmpty(t, tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
				assert.Equal(t, "user", tokens.Role)

				// Verify refresh token was stored in database
				var refreshToken entity.RefreshToken
				err := db.First(&refreshToken).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, refreshToken.Token)
				assert.True(t, refreshToken.ExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Invalid email",
			setup: func() dto.UserLoginRequest {
				// No setup needed - user doesn't exist
				return dto.UserLoginRequest{
					Email:    "nonexistent@example.com",
					Password: "anypassword",
				}
			},
			expectedError: "invalid email or password",
			validate: func(t *testing.T, tokens dto.TokenResponse) {
				assert.Empty(t, tokens.AccessToken)
				assert.Empty(t, tokens.RefreshToken)
			},
		},
		{
			name: "Invalid password",
			setup: func() dto.UserLoginRequest {
				// Create test user
				password := "correctpassword"
				hashedPassword, err := helpers.HashPassword(password)
				assert.NoError(t, err)

				user := entity.User{
					Name:       "Verified User",
					Email:      "user@example.com",
					Password:   hashedPassword,
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				_, err = userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				return dto.UserLoginRequest{
					Email:    "user@example.com",
					Password: "wrongpassword",
				}
			},
			expectedError: "invalid email or password",
			validate: func(t *testing.T, tokens dto.TokenResponse) {
				assert.Empty(t, tokens.AccessToken)
				assert.Empty(t, tokens.RefreshToken)
			},
		},
		{
			name: "Unverified account",
			setup: func() dto.UserLoginRequest {
				// Create test user
				password := "correctpassword"
				hashedPassword, err := helpers.HashPassword(password)
				assert.NoError(t, err)

				user := entity.User{
					Name:       "Unverified User",
					Email:      "unverified@example.com",
					Password:   hashedPassword,
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: false,
				}
				_, err = userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				return dto.UserLoginRequest{
					Email:    "unverified@example.com",
					Password: "correctpassword",
				}
			},
			expectedError: "invalid email or password",
			validate: func(t *testing.T, tokens dto.TokenResponse) {
				assert.Empty(t, tokens.AccessToken)
				assert.Empty(t, tokens.RefreshToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE refresh_tokens RESTART IDENTITY CASCADE")
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Setup test data
				loginRequest := tt.setup()

				// Execute the function
				tokens, err := userService.Verify(ctx, loginRequest)

				// Validate results
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				} else {
					assert.NoError(t, err)
				}

				// Validate response
				tt.validate(t, tokens)
			},
		)
	}
}

func TestUserService_RefreshToken(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
	if err != nil {
		panic(err)
	}

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Clean up database after test
	defer func() {
		db.Exec("DELETE FROM refresh_tokens WHERE TRUE")
		db.Exec("DELETE FROM users WHERE TRUE")
	}()

	// Create test context
	ctx := context.Background()

	// Helper function to create test user with refresh token
	createTestUserWithToken := func() (entity.User, string) {
		// Create test user
		user := entity.User{
			Name:       "Test User",
			Email:      "test@example.com",
			Password:   "password123",
			TelpNumber: "1234567890",
			Role:       "user",
			IsVerified: true,
		}
		createdUser, err := userRepo.Register(ctx, nil, user)
		assert.NoError(t, err)

		// Create refresh token
		refreshTokenString, expiresAt := jwtService.GenerateRefreshToken()

		refreshToken := entity.RefreshToken{
			UserID:    createdUser.ID,
			Token:     refreshTokenString,
			ExpiresAt: expiresAt,
		}
		_, err = refreshTokenRepo.Create(ctx, nil, refreshToken)
		assert.NoError(t, err)

		return createdUser, refreshTokenString
	}

	// Test cases
	tests := []struct {
		name          string
		setup         func() (dto.RefreshTokenRequest, string) // returns request and expected role
		expectedError string
		validate      func(t *testing.T, response dto.TokenResponse, originalRefreshToken string)
	}{
		{
			name: "Successfully refresh token",
			setup: func() (dto.RefreshTokenRequest, string) {
				_, refreshToken := createTestUserWithToken()
				return dto.RefreshTokenRequest{
					RefreshToken: refreshToken,
				}, "user"
			},
			expectedError: "",
			validate: func(t *testing.T, response dto.TokenResponse, originalRefreshToken string) {
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "user", response.Role)
				assert.NotEqual(
					t,
					originalRefreshToken,
					response.RefreshToken,
					"Refresh token should be different after refresh",
				)

				// Verify old token was deleted
				_, err := refreshTokenRepo.FindByToken(ctx, nil, originalRefreshToken)
				assert.Error(t, err)
				assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

				// Verify new token exists
				_, err = refreshTokenRepo.FindByToken(ctx, nil, response.RefreshToken)
				assert.NoError(t, err)
			},
		},
		{
			name: "Invalid refresh token",
			setup: func() (dto.RefreshTokenRequest, string) {
				return dto.RefreshTokenRequest{
					RefreshToken: "invalid-token",
				}, ""
			},
			expectedError: dto.MESSAGE_FAILED_INVALID_REFRESH_TOKEN,
			validate:      func(t *testing.T, response dto.TokenResponse, originalRefreshToken string) {},
		},
		{
			name: "Expired refresh token",
			setup: func() (dto.RefreshTokenRequest, string) {
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				// Create expired refresh token
				refreshTokenString, _ := jwtService.GenerateRefreshToken()
				// Do not hash the token
				refreshToken := entity.RefreshToken{
					UserID:    createdUser.ID,
					Token:     refreshTokenString,             // Store raw token
					ExpiresAt: time.Now().Add(-1 * time.Hour), // expired 1 hour ago
				}
				_, err = refreshTokenRepo.Create(ctx, nil, refreshToken)
				assert.NoError(t, err)

				return dto.RefreshTokenRequest{
					RefreshToken: refreshTokenString,
				}, ""
			},
			expectedError: dto.MESSAGE_FAILED_EXPIRED_REFRESH_TOKEN,
			validate:      func(t *testing.T, response dto.TokenResponse, originalRefreshToken string) {},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE refresh_tokens RESTART IDENTITY CASCADE")
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

				// Setup test data
				req, expectedRole := tt.setup()

				// Execute the function
				response, err := userService.RefreshToken(ctx, req)

				// Validate results
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expectedRole, response.Role)
				}

				// Run additional validations
				tt.validate(t, response, req.RefreshToken)
			},
		)
	}
}

func TestUserService_RevokeRefreshToken(t *testing.T) {
	// Start test container
	dbContainer, err := container.StartTestContainer()
	assert.NoError(t, err)
	defer func(dbContainer *container.TestDatabaseContainer) {
		err := dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(dbContainer)

	// Set environment variables for database connection
	err = os.Setenv("DB_HOST", dbContainer.Host)
	assert.NoError(t, err)
	err = os.Setenv("DB_USER", "testuser")
	assert.NoError(t, err)
	err = os.Setenv("DB_PASS", "testpassword")
	assert.NoError(t, err)
	err = os.Setenv("DB_NAME", "testdb")
	assert.NoError(t, err)
	err = os.Setenv("DB_PORT", dbContainer.Port)
	assert.NoError(t, err)

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		assert.NoError(t, err)
	}(db)

	// Migrate database schema
	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	assert.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := service.NewJWTService()

	// Create user service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		setup         func() (string, int) // returns user ID and expected token count
		expectedError error
	}{
		{
			name: "Successfully revoke refresh tokens",
			setup: func() (string, int) {
				// Create test user
				user := entity.User{
					Name:       "Test User",
					Email:      "test@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)

				// Create refresh tokens for the user
				refreshTokens := []entity.RefreshToken{
					{
						UserID:    createdUser.ID,
						Token:     "token1",
						ExpiresAt: time.Now().Add(time.Hour),
					},
					{
						UserID:    createdUser.ID,
						Token:     "token2",
						ExpiresAt: time.Now().Add(time.Hour),
					},
				}

				for _, rt := range refreshTokens {
					_, err := refreshTokenRepo.Create(ctx, nil, rt)
					assert.NoError(t, err)
				}

				return createdUser.ID.String(), 0 // Expect 0 tokens after revocation
			},
			expectedError: nil,
		},
		{
			name: "User not found",
			setup: func() (string, int) {
				return uuid.New().String(), 0 // non-existent user ID
			},
			expectedError: dto.ErrUserNotFound,
		},
		{
			name: "No tokens to revoke",
			setup: func() (string, int) {
				// Create test user with no refresh tokens
				user := entity.User{
					Name:       "Test User No Tokens",
					Email:      "notokens@example.com",
					Password:   "password123",
					TelpNumber: "1234567890",
					Role:       "user",
					IsVerified: true,
				}
				createdUser, err := userRepo.Register(ctx, nil, user)
				assert.NoError(t, err)
				return createdUser.ID.String(), 0 // Expect 0 tokens (none existed)
			},
			expectedError: nil,
		},
		{
			name: "Invalid user ID format",
			setup: func() (string, int) {
				return "invalid-uuid-format", 0
			},
			expectedError: dto.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Clean database before each test
				db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
				db.Exec("TRUNCATE TABLE refresh_tokens RESTART IDENTITY CASCADE")

				// Setup test data and get user ID and expected token count
				userID, expectedTokenCount := tt.setup()

				// Execute the function
				err := userService.RevokeRefreshToken(ctx, userID)

				// Validate results
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.NoError(t, err)
				}

				// Verify token count in database if no error was expected
				if tt.expectedError == nil {
					var count int64
					err := db.Model(&entity.RefreshToken{}).Where("user_id = ?", userID).Count(&count).Error
					assert.NoError(t, err)
					assert.Equal(t, int64(expectedTokenCount), count)
				}
			},
		)
	}
}
