// Package profiles implements student profile structures, skill states, and XP progression logic.
package profiles

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProfileRepository defines datastore access specifications for mapping student profile entities.
type ProfileRepository interface {
	Create(ctx context.Context, profile *Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	Update(ctx context.Context, profile *Profile) error
}

type sqlProfileRepository struct {
	db *sql.DB
}

// NewProfileRepository instantiates a new PostgreSQL profile repository dependency.
func NewProfileRepository(db *sql.DB) ProfileRepository {
	return &sqlProfileRepository{db: db}
}

// Create inserts a new default student profile row in PostgreSQL.
func (r *sqlProfileRepository) Create(ctx context.Context, profile *Profile) error {
	query := `
		INSERT INTO profiles (user_id, university, current_level, employability_score, total_xp, skills, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	skillsJSON, err := profile.SkillsJSON()
	if err != nil {
		return err
	}

	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		profile.UserID,
		profile.University,
		profile.CurrentLevel,
		profile.EmployabilityScore,
		profile.TotalXP,
		skillsJSON,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	return err
}

// GetByUserID retrieves student details by user UUID, unmarshalling the JSONB skills payload.
func (r *sqlProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	query := `
		SELECT user_id, university, current_level, employability_score, total_xp, skills, created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`
	var p Profile
	var skillsBytes []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.UserID,
		&p.University,
		&p.CurrentLevel,
		&p.EmployabilityScore,
		&p.TotalXP,
		&skillsBytes,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	p.Skills = []SkillItem{}
	if len(skillsBytes) > 0 {
		err = json.Unmarshal(skillsBytes, &p.Skills)
		if err != nil {
			return nil, err
		}
	}

	return &p, nil
}

// Update mutates student profile fields, total XP ranks, and active skills arrays.
func (r *sqlProfileRepository) Update(ctx context.Context, profile *Profile) error {
	query := `
		UPDATE profiles
		SET university = $1, current_level = $2, employability_score = $3, total_xp = $4, skills = $5, updated_at = $6
		WHERE user_id = $7
	`
	skillsJSON, err := profile.SkillsJSON()
	if err != nil {
		return err
	}

	profile.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		profile.University,
		profile.CurrentLevel,
		profile.EmployabilityScore,
		profile.TotalXP,
		skillsJSON,
		profile.UpdatedAt,
		profile.UserID,
	)
	return err
}
