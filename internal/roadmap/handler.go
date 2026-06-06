// Package roadmap implements branching graph architectures representing adaptive student learning pathways.
package roadmap

import (
	"net/http"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RoadmapHandler translates standard incoming HTTP calls into roadmap domain operations.
type RoadmapHandler struct {
	service RoadmapService
}

// NewRoadmapHandler compiles a RoadmapHandler instance.
func NewRoadmapHandler(service RoadmapService) *RoadmapHandler {
	return &RoadmapHandler{service: service}
}

// GetRoadmap extracts request credentials context and returns student learning paths DAG graphs.
func (h *RoadmapHandler) GetRoadmap(c *gin.Context) {
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

	rm, err := h.service.GetRoadmap(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, rm)
}

// CompleteNode decodes inputs and triggers dynamic progression node completion and XP evaluations.
func (h *RoadmapHandler) CompleteNode(c *gin.Context) {
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
		NodeID string `json:"node_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("node_id is required", err))
		return
	}

	rm, err := h.service.CompleteNode(c.Request.Context(), userID, req.NodeID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, rm)
}
