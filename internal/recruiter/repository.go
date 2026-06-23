// Package recruiter handles matchmaking calculations between candidate capabilities and enterprise job requirements.
package recruiter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"skillbridge-backend/internal/profiles"
	"time"

	"github.com/google/uuid"
)

// RecruiterRepository provides database persistence interfaces for storing job openings and matchmaking pipelines.
type RecruiterRepository interface {
	CreateJob(ctx context.Context, job *JobPosting) error
	GetJobs(ctx context.Context) ([]JobPosting, error)
	GetJobByID(ctx context.Context, jobID uuid.UUID) (*JobPosting, error)
	GetCandidatesForMatch(ctx context.Context) ([]CandidateMatch, error)
}

type sqlRecruiterRepository struct {
	db *sql.DB
}

// NewRecruiterRepository creates a postgreSQL execution client implementation for RecruiterRepository.
func NewRecruiterRepository(db *sql.DB) RecruiterRepository {
	return &sqlRecruiterRepository{db: db}
}

// CreateJob persists a new job opening specifying mandatory skill criteria.
func (r *sqlRecruiterRepository) CreateJob(ctx context.Context, job *JobPosting) error {
	query := `
		INSERT INTO jobs (id, title, company, description, skills_required, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	skillsJSON, err := job.SkillsJSON()
	if err != nil {
		return err
	}

	job.CreatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		job.ID,
		job.Title,
		job.Company,
		job.Description,
		skillsJSON,
		job.CreatedAt,
	)
	return err
}

// GetJobs fetches chronological listings of active corporate job postings.
func (r *sqlRecruiterRepository) GetJobs(ctx context.Context) ([]JobPosting, error) {
	query := `
		SELECT id, title, company, description, skills_required, created_at
		FROM jobs
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobPosting
	for rows.Next() {
		var job JobPosting
		var skillsBytes []byte

		err := rows.Scan(
			&job.ID,
			&job.Title,
			&job.Company,
			&job.Description,
			&skillsBytes,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		job.SkillsRequired = []string{}
		if len(skillsBytes) > 0 {
			_ = json.Unmarshal(skillsBytes, &job.SkillsRequired)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobByID retrieves a specific job posting's details and skills prerequisites.
func (r *sqlRecruiterRepository) GetJobByID(ctx context.Context, jobID uuid.UUID) (*JobPosting, error) {
	query := `
		SELECT id, title, company, description, skills_required, created_at
		FROM jobs
		WHERE id = $1
	`
	var job JobPosting
	var skillsBytes []byte

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID,
		&job.Title,
		&job.Company,
		&job.Description,
		&skillsBytes,
		&job.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	job.SkillsRequired = []string{}
	if len(skillsBytes) > 0 {
		_ = json.Unmarshal(skillsBytes, &job.SkillsRequired)
	}

	return &job, nil
}

// GetCandidatesForMatch scans profiles of students eligible for enterprise recruitment matching.
func (r *sqlRecruiterRepository) GetCandidatesForMatch(ctx context.Context) ([]CandidateMatch, error) {
	query := `
		WITH LatestLedgers AS (
			SELECT DISTINCT ON (user_id) 
				user_id, architectural_velocity_score, debugging_loop_efficiency_score, communication_clarity_score
			FROM assessment_ledgers
			WHERE status = 'passed'
			ORDER BY user_id, completed_at DESC
		)
		SELECT 
			u.id, u.email, p.university, p.employability_score, p.total_xp, p.skills, p.premium_vetting_passed,
			COALESCE(l.architectural_velocity_score, 0),
			COALESCE(l.debugging_loop_efficiency_score, 0),
			COALESCE(l.communication_clarity_score, 0)
		FROM users u
		INNER JOIN profiles p ON u.id = p.user_id
		LEFT JOIN LatestLedgers l ON u.id = l.user_id
		WHERE u.role = 'student'
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []CandidateMatch
	for rows.Next() {
		var c CandidateMatch
		var skillsBytes []byte
		var userSkills []profiles.SkillItem

		err := rows.Scan(
			&c.UserID,
			&c.Email,
			&c.University,
			&c.EmployabilityScore,
			&c.TotalXP,
			&skillsBytes,
			&c.PremiumVettingPassed,
			&c.ArchitecturalVelocity,
			&c.DebuggingEfficiency,
			&c.CommunicationClarity,
		)
		if err != nil {
			return nil, err
		}

		if len(skillsBytes) > 0 {
			_ = json.Unmarshal(skillsBytes, &userSkills)
		}
		
		candidates = append(candidates, c)
	}

	return candidates, nil
}
