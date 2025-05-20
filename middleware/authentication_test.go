package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/Caknoooo/go-gin-clean-starter/dto"
	"github.com/Caknoooo/go-gin-clean-starter/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) ValidateToken(token string) (*jwt.Token, error) {
	args := m.Called(token)
	// Handle nil case safely
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockJWTService) GetUserIDByToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateAccessToken(userId string, role string) string {
	args := m.Called(userId, role)
	return args.String(0)
}

func (m *MockJWTService) GenerateRefreshToken() (string, time.Time) {
	args := m.Called()
	return args.String(0), args.Get(1).(time.Time)
}

func TestAuthenticateMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		setupAuth        func() string
		mockJWTSetup     func(mock *MockJWTService)
		expectedStatus   int
		expectedResponse utils.Response
		checkContext     func(c *gin.Context)
	}{
		{
			name:           "No Authorization header",
			setupAuth:      func() string { return "" },
			mockJWTSetup:   func(mock *MockJWTService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.BuildResponseFailed(
				dto.MESSAGE_FAILED_PROSES_REQUEST,
				dto.MESSAGE_FAILED_TOKEN_NOT_FOUND,
				nil,
			),
		},
		{
			name:           "Invalid Authorization format (no Bearer)",
			setupAuth:      func() string { return "InvalidTokenFormat" },
			mockJWTSetup:   func(mock *MockJWTService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.BuildResponseFailed(
				dto.MESSAGE_FAILED_PROSES_REQUEST,
				dto.MESSAGE_FAILED_TOKEN_NOT_VALID,
				nil,
			),
		},
		{
			name:      "Invalid token",
			setupAuth: func() string { return "Bearer invalidtoken" },
			mockJWTSetup: func(mock *MockJWTService) {
				// Explicitly return nil as *jwt.Token
				mock.On("ValidateToken", "invalidtoken").Return((*jwt.Token)(nil), errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.BuildResponseFailed(
				dto.MESSAGE_FAILED_PROSES_REQUEST,
				dto.MESSAGE_FAILED_TOKEN_NOT_VALID,
				nil,
			),
		},
		{
			name:      "Valid token but invalid user ID",
			setupAuth: func() string { return "Bearer validtoken" },
			mockJWTSetup: func(mock *MockJWTService) {
				token := &jwt.Token{Valid: true}
				mock.On("ValidateToken", "validtoken").Return(token, nil)
				mock.On("GetUserIDByToken", "validtoken").Return("", errors.New("user not found"))
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "user not found", nil),
		},
		{
			name:      "Valid token and user ID",
			setupAuth: func() string { return "Bearer validtoken" },
			mockJWTSetup: func(mock *MockJWTService) {
				token := &jwt.Token{Valid: true}
				mock.On("ValidateToken", "validtoken").Return(token, nil)
				mock.On("GetUserIDByToken", "validtoken").Return("user123", nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: utils.BuildResponseSuccess("OK", nil),
			checkContext: func(c *gin.Context) {
				token, exists := c.Get("token")
				assert.True(t, exists)
				assert.Equal(t, "validtoken", token)

				userID, exists := c.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, "user123", userID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {

				gin.SetMode(gin.TestMode)
				router := gin.New()

				mockJWT := new(MockJWTService)
				tt.mockJWTSetup(mockJWT)

				router.Use(Authenticate(mockJWT))
				router.GET(
					"/test", func(c *gin.Context) {
						if tt.checkContext != nil {
							tt.checkContext(c)
						}
						c.JSON(http.StatusOK, utils.BuildResponseSuccess("OK", nil))
					},
				)

				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				authHeader := tt.setupAuth()
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}

				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req)

				assert.Equal(t, tt.expectedStatus, resp.Code)

				if tt.expectedStatus != http.StatusOK {
					var response utils.Response
					err := json.Unmarshal(resp.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedResponse.Status, response.Status)
					assert.Equal(t, tt.expectedResponse.Message, response.Message)
					assert.Equal(t, tt.expectedResponse.Error, response.Error)
				}

				mockJWT.AssertExpectations(t)
			},
		)
	}
}
