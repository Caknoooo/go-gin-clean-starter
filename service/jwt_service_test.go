package service

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTService(t *testing.T) {
	// Test with environment variable
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	service := NewJWTService()
	assert.NotNil(t, service)
	jwtService, ok := service.(*jwtService)
	assert.True(t, ok)
	assert.Equal(t, "test-secret", jwtService.secretKey)
	assert.Equal(t, "Template", jwtService.issuer)
	assert.Equal(t, time.Minute*15, jwtService.accessExpiry)
	assert.Equal(t, time.Hour*24*7, jwtService.refreshExpiry)

	// Test without environment variable
	os.Unsetenv("JWT_SECRET")
	service = NewJWTService()
	assert.Equal(t, "test-secret", jwtService.secretKey)
}

func TestGenerateAccessToken(t *testing.T) {
	service := NewJWTService()
	userID := "test-user"
	role := "admin"

	token := service.GenerateAccessToken(userID, role)
	assert.NotEmpty(t, token)

	// Verify token contents
	parsedToken, err := jwt.ParseWithClaims(
		token, &jwtCustomClaim{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("Template"), nil
		},
	)
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*jwtCustomClaim)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "Template", claims.Issuer)
	assert.WithinDuration(t, time.Now().Add(time.Minute*15), claims.ExpiresAt.Time, time.Second)
}

func TestGenerateRefreshToken(t *testing.T) {
	service := NewJWTService()

	token, expiresAt := service.GenerateRefreshToken()
	assert.NotEmpty(t, token)
	assert.False(t, expiresAt.IsZero())
	assert.WithinDuration(t, time.Now().Add(time.Hour*24*7), expiresAt, time.Second)

	// Test token length (base64 encoded 32 bytes should be ~44 characters)
	assert.Len(t, token, 44)
}

func TestValidateToken(t *testing.T) {
	service := NewJWTService()

	// Test valid token
	validToken := service.GenerateAccessToken("test-user", "admin")
	parsedToken, err := service.ValidateToken(validToken)
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Test invalid token
	_, err = service.ValidateToken("invalid.token.string")
	assert.Error(t, err)

	// Test token with wrong signing method
	wrongMethodToken := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		&jwtCustomClaim{
			UserID: "test-user",
			Role:   "admin",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
				Issuer:    "Template",
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	)
	signedWrong, err := wrongMethodToken.SignedString([]byte("Template"))
	require.NoError(t, err, "Failed to sign token with wrong method")

	// This should now fail with the expected error
	_, err = service.ValidateToken(signedWrong)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

func TestGetUserIDByToken(t *testing.T) {
	service := NewJWTService()
	userID := "test-user"

	// Test valid token
	validToken := service.GenerateAccessToken(userID, "admin")
	retrievedID, err := service.GetUserIDByToken(validToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, retrievedID)

	// Test invalid token
	_, err = service.GetUserIDByToken("invalid.token.string")
	assert.Error(t, err)
}

func TestGetSecretKey(t *testing.T) {
	// Test with environment variable
	os.Setenv("JWT_SECRET", "custom-secret")
	defer os.Unsetenv("JWT_SECRET")
	assert.Equal(t, "custom-secret", getSecretKey())

	// Test without environment variable
	os.Unsetenv("JWT_SECRET")
	assert.Equal(t, "Template", getSecretKey())
}
