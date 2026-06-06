// Package workspace manages user code execution, workspace sandboxes, and integrated AI code reviews.
package workspace

import (
	"encoding/json"
	"skillbridge-backend/internal/ai_orchestrator"
	"time"

	"github.com/google/uuid"
)

// CodeSubmissionRequest defines the payload for compiling and reviewing user-submitted code in the sandbox.
type CodeSubmissionRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// WorkspaceSubmission stores metadata, run-time output, and external AI analysis of a user's code execution attempt.
type WorkspaceSubmission struct {
	ID         uuid.UUID                           `json:"id"`
	UserID     uuid.UUID                           `json:"user_id"`
	Language   string                              `json:"language"`
	Code       string                              `json:"code"`
	Status     string                              `json:"status"`
	Output     string                              `json:"output"`
	AIFeedback ai_orchestrator.CodeReviewResponse `json:"ai_feedback"`
	CreatedAt  time.Time                           `json:"created_at"`
}

// FeedbackJSON serializes the structured AI code review payload for relational database persistence.
func (w *WorkspaceSubmission) FeedbackJSON() ([]byte, error) {
	return json.Marshal(w.AIFeedback)
}

