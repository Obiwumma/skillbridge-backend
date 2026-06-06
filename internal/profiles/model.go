// Package profiles implements student profile structures, skill states, and XP progression logic.
package profiles

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SkillItem aggregates experience points and level metrics for a specific skill node.
type SkillItem struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
	XP    int    `json:"xp"`
}

// Profile represents a student profile, capturing university tier and dynamic employability score ranks.
type Profile struct {
	UserID             uuid.UUID   `json:"user_id"`
	University         string      `json:"university"`
	CurrentLevel       string      `json:"current_level"`
	EmployabilityScore int         `json:"employability_score"`
	TotalXP            int         `json:"total_xp"`
	Skills             []SkillItem `json:"skills"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// SkillsJSON marshals the skill item array to a JSON slice for PostgreSQL JSONB persistence.
func (p *Profile) SkillsJSON() ([]byte, error) {
	if p.Skills == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(p.Skills)
}
