package assessment

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AssessmentRepository defines persistence operations for vetting ledgers and interview sessions.
type AssessmentRepository interface {
	CreateLedger(ctx context.Context, ledger *AssessmentLedger) error
	GetLedgerByID(ctx context.Context, id uuid.UUID) (*AssessmentLedger, error)
	GetActiveLedger(ctx context.Context, userID uuid.UUID) (*AssessmentLedger, error)
	UpdateLedger(ctx context.Context, ledger *AssessmentLedger) error
	CreateSession(ctx context.Context, session *AssessmentSession) error
	UpdateSession(ctx context.Context, session *AssessmentSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*AssessmentSession, error)
	GetAttemptCount(ctx context.Context, userID uuid.UUID) (int, error)
}

type sqlAssessmentRepository struct {
	db *sql.DB
}

// NewAssessmentRepository constructs an SQL-backed implementation of AssessmentRepository.
func NewAssessmentRepository(db *sql.DB) AssessmentRepository {
	return &sqlAssessmentRepository{db: db}
}

func (r *sqlAssessmentRepository) CreateLedger(ctx context.Context, ledger *AssessmentLedger) error {
	query := `
		INSERT INTO assessment_ledgers 
		(id, user_id, attempt_number, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	ledger.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		ledger.ID, ledger.UserID, ledger.AttemptNumber, ledger.Status, ledger.CreatedAt,
	)
	return err
}

func (r *sqlAssessmentRepository) GetActiveLedger(ctx context.Context, userID uuid.UUID) (*AssessmentLedger, error) {
	query := `
		SELECT id, user_id, attempt_number, status, overall_score,
		architectural_velocity_score, code_auditing_score, debugging_loop_efficiency_score,
		sla_punctuality_score, tooling_fluency_score, communication_clarity_score,
		identified_gaps, created_at, completed_at
		FROM assessment_ledgers
		WHERE user_id = $1 AND status = 'in_progress'
		ORDER BY created_at DESC LIMIT 1
	`
	var l AssessmentLedger
	var gapsBytes []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&l.ID, &l.UserID, &l.AttemptNumber, &l.Status, &l.OverallScore,
		&l.ArchitecturalVelocityScore, &l.CodeAuditingScore, &l.DebuggingLoopEfficiencyScore,
		&l.SLAPunctualityScore, &l.ToolingFluencyScore, &l.CommunicationClarityScore,
		&gapsBytes, &l.CreatedAt, &l.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(gapsBytes) > 0 {
		_ = json.Unmarshal(gapsBytes, &l.IdentifiedGaps)
	}

	return &l, nil
}

func (r *sqlAssessmentRepository) GetLedgerByID(ctx context.Context, id uuid.UUID) (*AssessmentLedger, error) {
	query := `
		SELECT id, user_id, attempt_number, status, overall_score,
		architectural_velocity_score, code_auditing_score, debugging_loop_efficiency_score,
		sla_punctuality_score, tooling_fluency_score, communication_clarity_score,
		identified_gaps, created_at, completed_at
		FROM assessment_ledgers
		WHERE id = $1
	`
	var l AssessmentLedger
	var gapsBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.UserID, &l.AttemptNumber, &l.Status, &l.OverallScore,
		&l.ArchitecturalVelocityScore, &l.CodeAuditingScore, &l.DebuggingLoopEfficiencyScore,
		&l.SLAPunctualityScore, &l.ToolingFluencyScore, &l.CommunicationClarityScore,
		&gapsBytes, &l.CreatedAt, &l.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(gapsBytes) > 0 {
		_ = json.Unmarshal(gapsBytes, &l.IdentifiedGaps)
	}

	return &l, nil
}

func (r *sqlAssessmentRepository) UpdateLedger(ctx context.Context, ledger *AssessmentLedger) error {
	query := `
		UPDATE assessment_ledgers SET
		status = $1, overall_score = $2, architectural_velocity_score = $3,
		code_auditing_score = $4, debugging_loop_efficiency_score = $5,
		sla_punctuality_score = $6, tooling_fluency_score = $7,
		communication_clarity_score = $8, identified_gaps = $9, completed_at = $10
		WHERE id = $11
	`
	gapsJSON, _ := ledger.IdentifiedGapsJSON()
	_, err := r.db.ExecContext(ctx, query,
		ledger.Status, ledger.OverallScore, ledger.ArchitecturalVelocityScore,
		ledger.CodeAuditingScore, ledger.DebuggingLoopEfficiencyScore,
		ledger.SLAPunctualityScore, ledger.ToolingFluencyScore,
		ledger.CommunicationClarityScore, gapsJSON, ledger.CompletedAt,
		ledger.ID,
	)
	return err
}

func (r *sqlAssessmentRepository) CreateSession(ctx context.Context, session *AssessmentSession) error {
	query := `
		INSERT INTO assessment_sessions 
		(id, ledger_id, session_type, problem_statement, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	session.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.LedgerID, session.SessionType, session.ProblemStatement, session.CreatedAt,
	)
	return err
}

func (r *sqlAssessmentRepository) UpdateSession(ctx context.Context, session *AssessmentSession) error {
	query := `
		UPDATE assessment_sessions SET
		time_to_first_solution_seconds = $1, bugs_introduced_by_ai = $2,
		bugs_caught_by_candidate = $3, compilation_attempts = $4,
		ai_analysis_reference_id = $5, ended_at = $6
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		session.TimeToFirstSolutionSec, session.BugsIntroducedByAI,
		session.BugsCaughtByCandidate, session.CompilationAttempts,
		session.AIAnalysisReferenceID, session.EndedAt, session.ID,
	)
	return err
}

func (r *sqlAssessmentRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*AssessmentSession, error) {
	query := `
		SELECT id, ledger_id, session_type, problem_statement,
		time_to_first_solution_seconds, bugs_introduced_by_ai, bugs_caught_by_candidate,
		compilation_attempts, ai_analysis_reference_id, created_at, ended_at
		FROM assessment_sessions
		WHERE id = $1
	`
	var s AssessmentSession
	var timeSec, bugsAI, bugsCand, comps sql.NullInt64
	var aiRef sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.LedgerID, &s.SessionType, &s.ProblemStatement,
		&timeSec, &bugsAI, &bugsCand, &comps, &aiRef, &s.CreatedAt, &s.EndedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if timeSec.Valid { s.TimeToFirstSolutionSec = int(timeSec.Int64) }
	if bugsAI.Valid { s.BugsIntroducedByAI = int(bugsAI.Int64) }
	if bugsCand.Valid { s.BugsCaughtByCandidate = int(bugsCand.Int64) }
	if comps.Valid { s.CompilationAttempts = int(comps.Int64) }
	if aiRef.Valid { s.AIAnalysisReferenceID = aiRef.String }

	return &s, nil
}

func (r *sqlAssessmentRepository) GetAttemptCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM assessment_ledgers WHERE user_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}
