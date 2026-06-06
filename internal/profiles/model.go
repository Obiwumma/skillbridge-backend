// Package profiles implements student profile structures, skill states, and XP progression logic.
package profiles

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// academiclevelconstants
type AcademicLevel string

const (
	Level100 AcademicLevel = "100"
	Level200 AcademicLevel = "200"
	Level300 AcademicLevel = "300"
	Level400 AcademicLevel = "400"
	Level500 AcademicLevel = "500"
)

// IsValid checks if the academic level is one of the predefined values.
func (al AcademicLevel) IsValid() bool {
	switch al {
	case Level100, Level200, Level300, Level400, Level500:
		return true
	}
	return false
}

// SkillItem aggregates experience points and level metrics for a specific skill node.
type SkillItem struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
	XP    int    `json:"xp"`
}

// Profile represents a student profile, capturing university tier and dynamic employability score ranks.
type Profile struct {
	UserID             uuid.UUID     `json:"user_id"`
	University         string        `json:"university"`
	CurrentLevel       AcademicLevel `json:"current_level"`
	EmployabilityScore int           `json:"employability_score"`
	TotalXP            int           `json:"total_xp"`
	Skills             []SkillItem   `json:"skills"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// SkillsJSON marshals the skill item array to a JSON slice for PostgreSQL JSONB persistence.
func (p *Profile) SkillsJSON() ([]byte, error) {
	if p.Skills == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(p.Skills)
}
