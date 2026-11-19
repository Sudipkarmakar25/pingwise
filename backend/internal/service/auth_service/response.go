package auth_service

import "pingwise/internal/models"

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"` // access token lifetime in seconds
	User         *models.User `json:"user"`
}
