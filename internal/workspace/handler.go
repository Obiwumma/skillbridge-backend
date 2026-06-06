// Package workspace manages user code execution, workspace sandboxes, and integrated AI code reviews.
package workspace

import (
	"net/http"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WorkspaceHandler exposes REST endpoints for submitting and compiling code sandbox requests.
type WorkspaceHandler struct {
	service WorkspaceService
}

// NewWorkspaceHandler constructs a new WorkspaceHandler with the given service dependency.
func NewWorkspaceHandler(service WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: service}
}

// Execute compiles and executes the submitted user code sandbox and coordinates real-time feedback.
func (h *WorkspaceHandler) Execute(c *gin.Context) {
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

	var req CodeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("language and code are required fields", err))
		return
	}

	result, err := h.service.ExecuteCode(c.Request.Context(), userID, req.Language, req.Code)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, result)
}

// GetHistory fetches the chronological log of user workspace submissions and reviews.
func (h *WorkspaceHandler) GetHistory(c *gin.Context) {
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

	history, err := h.service.GetHistory(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, history)
}
