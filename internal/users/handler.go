// Package users manages account profiles, registration records, and permissions.
package users

import (
	"net/http"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler translates standard incoming HTTP calls into user domain workflows.
type UserHandler struct {
	service UserService
}

// NewUserHandler compiles a UserHandler instance.
func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetProfile extracts request session contexts and serializes user account data profiles.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, errors.NewUnauthorizedError("missing authorization session", nil))
		return
	}

	userId, err := uuid.Parse(userIdStr.(string))
	if err != nil {
		response.Error(c, errors.NewValidationError("invalid session identity", err))
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), userId)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, user)
}
