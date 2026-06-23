// Package assessment handles the Premium Vetting Gate, tracking ledgers and AI-driven interview sessions.
package assessment

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AssessmentLedger tracks a candidate's attempt at the premium vetting gate (The Retake Ledger).
type AssessmentLedger struct {
	ID                            uuid.UUID `json:"id"`
	UserID                        uuid.UUID `json:"user_id"`
	AttemptNumber                 int       `json:"attempt_number"`
	Status                        string    `json:"status"` // 'in_progress', 'passed', 'failed'
	OverallScore                  int       `json:"overall_score"`
	ArchitecturalVelocityScore    int       `json:"architectural_velocity_score"`
	CodeAuditingScore             int       `json:"code_auditing_score"`
	DebuggingLoopEfficiencyScore  int       `json:"debugging_loop_efficiency_score"`
	SLAPunctualityScore           int       `json:"sla_punctuality_score"`
	ToolingFluencyScore           int       `json:"tooling_fluency_score"`
	CommunicationClarityScore     int       `json:"communication_clarity_score"`
	IdentifiedGaps                []string  `json:"identified_gaps"`
	CreatedAt                     time.Time `json:"created_at"`
	CompletedAt                   *time.Time `json:"completed_at,omitempty"`
}

// IdentifiedGapsJSON serializes the specific gaps into a JSON byte array for database persistence.
func (a *AssessmentLedger) IdentifiedGapsJSON() ([]byte, error) {
	if a.IdentifiedGaps == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(a.IdentifiedGaps)
}

// AssessmentSession tracks a specific 60-minute simulated pair-programming session.
type AssessmentSession struct {
	ID                        uuid.UUID `json:"id"`
	LedgerID                  uuid.UUID `json:"ledger_id"`
	SessionType               string    `json:"session_type"` // 'ai_technical', 'hr_behavioral'
	ProblemStatement          string    `json:"problem_statement"`
	TimeToFirstSolutionSec    int       `json:"time_to_first_solution_seconds"`
	BugsIntroducedByAI        int       `json:"bugs_introduced_by_ai"`
	BugsCaughtByCandidate     int       `json:"bugs_caught_by_candidate"`
	CompilationAttempts       int       `json:"compilation_attempts"`
	AIAnalysisReferenceID     string    `json:"ai_analysis_reference_id"`
	CreatedAt                 time.Time `json:"created_at"`
	EndedAt                   *time.Time `json:"ended_at,omitempty"`
}

// AIInterviewPayload represents the precise JSON contract sent to the Python AI service.
type AIInterviewPayload struct {
	SessionID         string `json:"session_id"`
	UserID            string `json:"user_id"`
	ProblemStatement  string `json:"problem_statement"`
	CandidateCode     string `json:"candidate_code"`
	ExecutionOutput   string `json:"execution_output"`
	TimeTakenSeconds  int    `json:"time_taken_seconds"`
	CompilationCount  int    `json:"compilation_count"`
}
