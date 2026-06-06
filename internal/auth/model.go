// Package auth coordinates system session generation, verification gates, and token structures.
package auth

import (
	"skillbridge-backend/internal/users"

	"github.com/golang-jwt/jwt/v5"
)

// RegisterRequest defines input constraints when submitting structural user registrations.
type RegisterRequest struct {
	Email        string     `json:"email" binding:"required,email"`
	Password     string     `json:"password" binding:"required,min=6"`
	Role         users.Role `json:"role" binding:"required,oneof=student recruiter admin"`
	University   string     `json:"university"`
	CurrentLevel string     `json:"current_level"`
}

// LoginRequest defines credentials parameters required during account authentication handshakes.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest defines parameters received to rotate expired user access sessions.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse aggregates signed Access and Refresh token strings returned to successful clients.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JWTClaims defines payload metadata properties encrypted inside signed JSON Web Tokens.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
