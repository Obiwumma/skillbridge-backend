// Package recruiter handles matchmaking calculations between candidate capabilities and enterprise job requirements.
package recruiter

import (
	"net/http"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RecruiterHandler handles recruitment matching requests and corporate postings.
type RecruiterHandler struct {
	service RecruiterService
}

// NewRecruiterHandler constructs a new RecruiterHandler instance.
func NewRecruiterHandler(service RecruiterService) *RecruiterHandler {
	return &RecruiterHandler{service: service}
}

// GetJobs lists all active corporate career postings.
func (h *RecruiterHandler) GetJobs(c *gin.Context) {
	jobs, err := h.service.GetJobs(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, jobs)
}

// MatchCandidates compares candidate profiles against job requirements to compute match rankings.
func (h *RecruiterHandler) MatchCandidates(c *gin.Context) {
	jobIDStr := c.Query("job_id")
	if jobIDStr == "" {
		response.Error(c, errors.NewValidationError("job_id query parameter is required", nil))
		return
	}

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		response.Error(c, errors.NewValidationError("invalid job_id format", err))
		return
	}

	matches, err := h.service.MatchCandidates(c.Request.Context(), jobID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, matches)
}

// PostJob creates a new career opening with mandatory skill sets.
func (h *RecruiterHandler) PostJob(c *gin.Context) {
	var req struct {
		Title          string   `json:"title" binding:"required"`
		Company        string   `json:"company" binding:"required"`
		Description    string   `json:"description" binding:"required"`
		SkillsRequired []string `json:"skills_required" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewValidationError("title, company, description, and skills_required are required", err))
		return
	}

	job, err := h.service.PostJob(c.Request.Context(), req.Title, req.Company, req.Description, req.SkillsRequired)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, job)
}
