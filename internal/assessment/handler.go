package assessment

import (
	"net/http"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AssessmentHandler handles REST endpoints for the premium vetting gate.
type AssessmentHandler struct {
	service AssessmentService
}

// NewAssessmentHandler constructs the handler.
func NewAssessmentHandler(service AssessmentService) *AssessmentHandler {
	return &AssessmentHandler{service: service}
}

// StartAssessment initiates the rigorous vetting process for an eligible candidate.
func (h *AssessmentHandler) StartAssessment(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, errors.NewValidationError("invalid session format", err))
		return
	}

	ledger, err := h.service.StartAssessment(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, ledger)
}

// StartSession begins a specific simulated interview stage.
func (h *AssessmentHandler) StartSession(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, _ := uuid.Parse(userIDStr.(string))

	var req struct {
		SessionType string `json:"session_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("session_type is required", err))
		return
	}

	session, err := h.service.StartSession(c.Request.Context(), userID, req.SessionType)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, session)
}

// SubmitSession processes the results of a simulated pair-programming session via the Python AI service.
func (h *AssessmentHandler) SubmitSession(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, _ := uuid.Parse(userIDStr.(string))

	var req struct {
		SessionID            string `json:"session_id" binding:"required"`
		Code                 string `json:"code" binding:"required"`
		ExecutionOutput      string `json:"execution_output"`
		TimeTakenSeconds     int    `json:"time_taken_seconds" binding:"required"`
		CompilationAttempts  int    `json:"compilation_attempts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("invalid request payload", err))
		return
	}

	sessionID, _ := uuid.Parse(req.SessionID)

	session, err := h.service.SubmitSession(c.Request.Context(), userID, sessionID, req.Code, req.ExecutionOutput, req.TimeTakenSeconds, req.CompilationAttempts)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, session)
}
