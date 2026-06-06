// Package response standardizes all REST API JSON response bodies.
package response

import (
	"errors"
	"net/http"
	appErrors "skillbridge-backend/pkg/errors"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the outer JSON envelope for all server outputs.
type APIResponse struct {
	Success bool         `json:"success"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail maps failure metadata returned to clients during request issues.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON serializes and prints successful payloads to HTTP clients.
func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error intercepts application errors and writes standard JSON error payloads.
func Error(c *gin.Context, err error) {
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, APIResponse{
			Success: false,
			Error: &ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred",
		},
	})
}
