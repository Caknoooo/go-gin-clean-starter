package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTService interface {
	GenerateAccessToken(userId string, role string) string
	GenerateRefreshToken() (string, time.Time)
	ValidateToken(token string) (*jwt.Token, error)
	GetUserIDByToken(token string) (string, error)
}

type jwtCustomClaim struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey     string
	issuer        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTService() JWTService {
	return &jwtService{
		secretKey:     getSecretKey(),
		issuer:        "Template",
		accessExpiry:  time.Minute * 15,
		refreshExpiry: time.Hour * 24 * 7,
	}
}

func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "Template"
	}
	return secretKey
}

func (j *jwtService) GenerateAccessToken(userId string, role string) string {
	claims := jwtCustomClaim{
		userId,
		role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpiry)),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tx, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		log.Println(err)
	}
	return tx
}

func (j *jwtService) GenerateRefreshToken() (string, time.Time) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
		return "", time.Time{}
	}

	refreshToken := base64.StdEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(j.refreshExpiry)

	return refreshToken, expiresAt
}

func (j *jwtService) parseToken(t_ *jwt.Token) (any, error) {
	if _, ok := t_.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method %v", t_.Header["alg"])
	}
	return []byte(j.secretKey), nil
}

func (j *jwtService) ValidateToken(token string) (*jwt.Token, error) {
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	// First check if the token has the expected number of parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token contains an invalid number of segments")
	}

	parsedToken, err := jwt.ParseWithClaims(
		token, &jwtCustomClaim{}, func(token *jwt.Token) (interface{}, error) {
			// Specifically check for SigningMethodHS256
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
			}
			return []byte(j.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return parsedToken, nil
}

func (j *jwtService) GetUserIDByToken(token string) (string, error) {
	tToken, err := j.ValidateToken(token)
	if err != nil {
		return "", err
	}

	// Get the claims from the token
	claims, ok := tToken.Claims.(*jwtCustomClaim)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.UserID, nil
}
