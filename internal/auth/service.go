// Package auth coordinates system session generation, verification gates, and token structures.
package auth

import (
	"context"
	"errors"
	"skillbridge-backend/internal/cache"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/internal/users"
	"skillbridge-backend/pkg/config"
	appErrors "skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/logger"
	"skillbridge-backend/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthService specifies functions managing authentication handshakes, sessions, and token rotation.
type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error)
	Login(ctx context.Context, email, password string) (*TokenResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error)
	ValidateToken(tokenStr string) (*JWTClaims, error)
}

type authService struct {
	userService users.UserService
}

// NewAuthService returns an AuthService instance linked to its dependencies.
func NewAuthService(userService users.UserService) AuthService {
	return &authService{userService: userService}
}

// Register creates a new user account, broadcasts the event to NATS, and generates standard tokens.
func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	user, err := s.userService.CreateUser(ctx, req.Email, req.Password, req.Role)
	if err != nil {
		return nil, err
	}

	eventPayload := map[string]interface{}{
		"user_id":       user.ID.String(),
		"role":          string(user.Role),
		"email":         user.Email,
		"university":    req.University,
		"current_level": req.CurrentLevel,
	}
	_ = events.Publish(events.UserRegistered, eventPayload)

	return s.generateTokenPair(ctx, user.ID.String(), string(user.Role))
}

// Login authenticates credentials, returns signed tokens, and creates a secure session.
func (s *authService) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to verify login credentials", err)
	}
	if user == nil {
		return nil, appErrors.NewUnauthorizedError("invalid email or password", nil)
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, appErrors.NewUnauthorizedError("invalid email or password", nil)
	}

	return s.generateTokenPair(ctx, user.ID.String(), string(user.Role))
}

// Refresh rotates active refresh tokens in Redis to secure sessions and generate fresh token pairs.
func (s *authService) Refresh(ctx context.Context, refreshTokenStr string) (*TokenResponse, error) {
	cfg := config.AppConfig

	token, err := jwt.ParseWithClaims(refreshTokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTRefreshSec), nil
	})
	if err != nil || !token.Valid {
		return nil, appErrors.NewUnauthorizedError("invalid refresh token", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, appErrors.NewUnauthorizedError("invalid refresh token claims", nil)
	}

	redisKey := "refresh_token:" + claims.UserID
	storedToken, err := cache.Get(ctx, redisKey)
	if err != nil || storedToken != refreshTokenStr {
		_ = cache.Delete(ctx, redisKey)
		return nil, appErrors.NewUnauthorizedError("refresh token reuse detected or expired", nil)
	}

	return s.generateTokenPair(ctx, claims.UserID, claims.Role)
}

// ValidateToken parses, decrypts, and confirms the signature validity of an access token string.
func (s *authService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	cfg := config.AppConfig
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid jwt token")
}

func (s *authService) generateTokenPair(ctx context.Context, userID, role string) (*TokenResponse, error) {
	cfg := config.AppConfig

	accessClaims := &JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		logger.Error("Failed to sign access token", err)
		return nil, appErrors.NewInternalError("failed to generate access token", err)
	}

	refreshClaims := &JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(cfg.JWTRefreshSec))
	if err != nil {
		logger.Error("Failed to sign refresh token", err)
		return nil, appErrors.NewInternalError("failed to generate refresh token", err)
	}

	redisKey := "refresh_token:" + userID
	err = cache.Set(ctx, redisKey, refreshStr, 7*24*time.Hour)
	if err != nil {
		logger.Error("Failed to cache refresh token in Redis", err)
		return nil, appErrors.NewInternalError("failed to secure session cache", err)
	}

	return &TokenResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}
