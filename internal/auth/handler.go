// Package auth coordinates system session generation, verification gates, and token structures.
package auth

import (
	"net/http"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler translates HTTP routes into authentication actions.
type AuthHandler struct {
	service AuthService
}

// NewAuthHandler returns an AuthHandler instance.
func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register processes requests to register new user accounts.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("invalid registration input parameters", err))
		return
	}

	res, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, res)
}

// Login processes requests to authenticate existing credentials.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("invalid credentials structure", err))
		return
	}

	res, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, res)
}

// Refresh processes requests to rotate expired access token sessions.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("refresh token is required", err))
		return
	}

	res, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, res)
}
