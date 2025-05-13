package dto

const (
	MESSAGE_SUCCESS_REFRESH_TOKEN        = "Successfully refreshed token"
	MESSAGE_FAILED_REFRESH_TOKEN         = "Failed to refresh token"
	MESSAGE_FAILED_INVALID_REFRESH_TOKEN = "Invalid refresh token"
	MESSAGE_FAILED_EXPIRED_REFRESH_TOKEN = "Refresh token has expired"
)

// TokenResponse represents a token response
// @Description Response body for access and refresh tokens
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Role         string `json:"role"`
}

// RefreshTokenRequest represents a request to refresh a token
// @Description Request body for refreshing a token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
