// Package profiles implements student profile structures, skill states, and XP progression logic.
package profiles

import (
	"net/http"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProfileHandler translates incoming HTTP requests to profile services.
type ProfileHandler struct {
	service ProfileService
}

// NewProfileHandler returns a ProfileHandler instance.
func NewProfileHandler(service ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

// GetProfile extracts user context session details and resolves student profile models.
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, errors.NewUnauthorizedError("missing session info", nil))
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, errors.NewValidationError("invalid session format", err))
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, profile)
}

// UpdateProfile decodes update details and mutates biographical profile parameters.
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, errors.NewUnauthorizedError("missing session info", nil))
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, errors.NewValidationError("invalid session format", err))
		return
	}

	var req struct {
		University   string `json:"university" binding:"required"`
		CurrentLevel string `json:"current_level" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("invalid update request body", err))
		return
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), userID, req.University, req.CurrentLevel)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, profile)
}
