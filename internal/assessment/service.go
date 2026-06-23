package assessment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"skillbridge-backend/internal/profiles"
	"skillbridge-backend/pkg/config"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/logger"
	"time"

	"github.com/google/uuid"
)

// AssessmentService coordinates the MVP rigorous assessment flow.
type AssessmentService interface {
	StartAssessment(ctx context.Context, userID uuid.UUID) (*AssessmentLedger, error)
	StartSession(ctx context.Context, userID uuid.UUID, sessionType string) (*AssessmentSession, error)
	SubmitSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, code string, output string, timeSec int, comps int) (*AssessmentSession, error)
}

type assessmentService struct {
	repo           AssessmentRepository
	profileService profiles.ProfileService
	client         *http.Client
}

// NewAssessmentService constructs the assessment service.
func NewAssessmentService(repo AssessmentRepository, profileService profiles.ProfileService) AssessmentService {
	return &assessmentService{
		repo:           repo,
		profileService: profileService,
		client: &http.Client{
			Timeout: 30 * time.Second, // Allow more time for AI processing
		},
	}
}

func (s *assessmentService) StartAssessment(ctx context.Context, userID uuid.UUID) (*AssessmentLedger, error) {
	// Check employability score
	profile, err := s.profileService.GetProfile(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get profile", err)
	}
	// Let's say 85 is the minimum for the gate
	if profile.EmployabilityScore < 85 {
		return nil, errors.NewForbiddenError("employability score too low for premium vetting", nil)
	}

	// Check if already in progress
	active, _ := s.repo.GetActiveLedger(ctx, userID)
	if active != nil {
		return active, nil
	}

	count, _ := s.repo.GetAttemptCount(ctx, userID)

	ledger := &AssessmentLedger{
		ID:            uuid.New(),
		UserID:        userID,
		AttemptNumber: count + 1,
		Status:        "in_progress",
	}

	err = s.repo.CreateLedger(ctx, ledger)
	if err != nil {
		return nil, errors.NewInternalError("failed to create assessment ledger", err)
	}

	return ledger, nil
}

func (s *assessmentService) StartSession(ctx context.Context, userID uuid.UUID, sessionType string) (*AssessmentSession, error) {
	ledger, err := s.repo.GetActiveLedger(ctx, userID)
	if err != nil || ledger == nil {
		return nil, errors.NewValidationError("no active assessment ledger", nil)
	}

	session := &AssessmentSession{
		ID:               uuid.New(),
		LedgerID:         ledger.ID,
		SessionType:      sessionType,
		ProblemStatement: "Refactor this poorly optimized data pipeline to handle 10k req/sec securely.",
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, errors.NewInternalError("failed to create assessment session", err)
	}

	return session, nil
}

func (s *assessmentService) SubmitSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, code string, output string, timeSec int, comps int) (*AssessmentSession, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil || session == nil {
		return nil, errors.NewNotFoundError("session not found", nil)
	}

	// Make HTTP call to external Python AI service (ADR-003)
	payload := AIInterviewPayload{
		SessionID:        sessionID.String(),
		UserID:           userID.String(),
		ProblemStatement: session.ProblemStatement,
		CandidateCode:    code,
		ExecutionOutput:  output,
		TimeTakenSeconds: timeSec,
		CompilationCount: comps,
	}

	reqBytes, _ := json.Marshal(payload)
	aiURL := config.AppConfig.ExternalAIURL + "/assess"
	
	req, _ := http.NewRequestWithContext(ctx, "POST", aiURL, bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	if config.AppConfig.ExternalAIKey != "" {
		req.Header.Set("Authorization", "Bearer "+config.AppConfig.ExternalAIKey)
	}

	resp, err := s.client.Do(req)
	var bugsAI, bugsCand int
	var aiRef string
	
	if err != nil {
		logger.Error("Failed to reach Python AI service, using fallback metrics", err)
		bugsAI = 2
		bugsCand = 1
		aiRef = "fallback-offline"
	} else {
		defer resp.Body.Close()
		var aiResp struct {
			BugsIntroduced int    `json:"bugs_introduced"`
			BugsCaught     int    `json:"bugs_caught"`
			ReferenceID    string `json:"reference_id"`
		}
		if json.NewDecoder(resp.Body).Decode(&aiResp) == nil {
			bugsAI = aiResp.BugsIntroduced
			bugsCand = aiResp.BugsCaught
			aiRef = aiResp.ReferenceID
		}
	}

	now := time.Now()
	session.TimeToFirstSolutionSec = timeSec
	session.CompilationAttempts = comps
	session.BugsIntroducedByAI = bugsAI
	session.BugsCaughtByCandidate = bugsCand
	session.AIAnalysisReferenceID = aiRef
	session.EndedAt = &now

	err = s.repo.UpdateSession(ctx, session)
	if err != nil {
		return nil, errors.NewInternalError("failed to update session metrics", err)
	}

	return session, nil
}
