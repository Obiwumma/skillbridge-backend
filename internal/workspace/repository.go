package workspace

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WorkspaceRepository defines the persistence operations for user code submissions and AI analysis feedback.
type WorkspaceRepository interface {
	Create(ctx context.Context, sub *WorkspaceSubmission) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]WorkspaceSubmission, error)
}

type sqlWorkspaceRepository struct {
	db *sql.DB
}

// NewWorkspaceRepository constructs an SQL-backed implementation of the WorkspaceRepository interface.
func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &sqlWorkspaceRepository{db: db}
}

func (r *sqlWorkspaceRepository) Create(ctx context.Context, sub *WorkspaceSubmission) error {
	query := `
		INSERT INTO workspace_submissions (id, user_id, language, code, status, output, ai_feedback, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	feedbackJSON, err := sub.FeedbackJSON()
	if err != nil {
		return err
	}

	sub.CreatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		sub.ID,
		sub.UserID,
		sub.Language,
		sub.Code,
		sub.Status,
		sub.Output,
		feedbackJSON,
		sub.CreatedAt,
	)
	return err
}

func (r *sqlWorkspaceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]WorkspaceSubmission, error) {
	query := `
		SELECT id, user_id, language, code, status, output, ai_feedback, created_at
		FROM workspace_submissions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []WorkspaceSubmission
	for rows.Next() {
		var sub WorkspaceSubmission
		var feedbackBytes []byte

		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.Language,
			&sub.Code,
			&sub.Status,
			&sub.Output,
			&feedbackBytes,
			&sub.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(feedbackBytes) > 0 {
			_ = json.Unmarshal(feedbackBytes, &sub.AIFeedback)
		}

		submissions = append(submissions, sub)
	}

	return submissions, nil
}
