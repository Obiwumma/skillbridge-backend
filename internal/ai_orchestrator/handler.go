// Package ai_orchestrator manages downstream communication with external machine learning pipelines.
package ai_orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/logger"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AIHandler translates REST request streams into machine intelligence pipeline orchestrations.
type AIHandler struct {
	service AIOrchestratorService
}

// NewAIHandler compiles an AIHandler instance.
func NewAIHandler(service AIOrchestratorService) *AIHandler {
	return &AIHandler{service: service}
}

// CVAnalyze triggers asynchronous resume parsing workflows, emitting cv.uploaded to the NATS bus.
func (h *AIHandler) CVAnalyze(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, errors.NewUnauthorizedError("missing session info", nil))
		return
	}

	var req struct {
		ResumeURL  string `json:"resume_url" binding:"required"`
		TargetRole string `json:"target_role" binding:"required"`
		Region     string `json:"region" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("resume_url, target_role, and region are required", err))
		return
	}

	eventPayload := CVUploadedRequest{
		Event:      events.CVUploaded,
		RequestID:  uuid.New().String(),
		UserID:     userIDStr.(string),
		ResumeURL:  req.ResumeURL,
		TargetRole: req.TargetRole,
		Region:     req.Region,
	}

	err := events.Publish(events.CVUploaded, eventPayload)
	if err != nil {
		response.Error(c, errors.NewInternalError("failed to initiate async CV processing", err))
		return
	}

	response.JSON(c, http.StatusAccepted, gin.H{
		"status":     "processing",
		"request_id": eventPayload.RequestID,
		"message":    "CV uploaded successfully. Asynchronous skill analysis is underway.",
	})
}

// SkillGap evaluates current skill metrics against target role requirements to return deficits.
func (h *AIHandler) SkillGap(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, errors.NewUnauthorizedError("missing session info", nil))
		return
	}

	var req struct {
		CurrentSkills []string `json:"current_skills" binding:"required"`
		TargetRole    string   `json:"target_role" binding:"required"`
		Region        string   `json:"region" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("current_skills, target_role, and region are required", err))
		return
	}

	aiReq := &SkillGapRequest{
		RequestID:     uuid.New().String(),
		UserID:        userIDStr.(string),
		CurrentSkills: req.CurrentSkills,
		TargetRole:    req.TargetRole,
		Region:        req.Region,
	}

	result, err := h.service.AnalyzeSkillGap(c.Request.Context(), aiReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, result)
}

// SubscribeToEvents binds NATS subscriptions to coordinate out-of-band resume parsing tasks.
func (h *AIHandler) SubscribeToEvents() {
	_, err := events.Subscribe(events.CVUploaded, func(data []byte) {
		logger.Info("AI Orchestrator received cv.uploaded event")

		var payload CVUploadedRequest
		err := json.Unmarshal(data, &payload)
		if err != nil {
			logger.Error("Failed to parse cv.uploaded payload", err)
			return
		}

		ctx := context.Background()
		analysis, err := h.service.AnalyzeCV(ctx, &payload)
		if err != nil {
			logger.Error("Failed to perform external AI CV analysis", err)
			return
		}

		var parsedSkills []map[string]interface{}
		for _, skill := range analysis.Skills {
			parsedSkills = append(parsedSkills, map[string]interface{}{
				"name":  skill.Name,
				"level": 1,
			})
		}

		analysisCompletedPayload := map[string]interface{}{
			"user_id": payload.UserID,
			"skills":  parsedSkills,
		}
		
		err = events.Publish(events.SkillAnalysisCompleted, analysisCompletedPayload)
		if err != nil {
			logger.Error("Failed to publish skill.analysis.completed", err)
			return
		}

		// Emit the authoritative WebSocket notification event contract
		wsEvent := map[string]string{
			"event":   "resume.analysis.completed",
			"user_id": payload.UserID,
			"status":  "SUCCESS",
		}
		_ = events.Publish("resume.analysis.completed", wsEvent)

		logger.Info("Async CV processing complete; skill analysis emitted", "user_id", payload.UserID)
	})

	if err != nil {
		logger.Error("AI Orchestrator failed to subscribe to cv.uploaded", err)
	}
}
