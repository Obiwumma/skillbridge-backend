// Package recruiter handles matchmaking calculations between candidate capabilities and enterprise job requirements.
package recruiter

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JobPosting represents an enterprise career opening outlining required core skills.
type JobPosting struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Company        string    `json:"company"`
	Description    string    `json:"description"`
	SkillsRequired []string  `json:"skills_required"`
	CreatedAt      time.Time `json:"created_at"`
}

// SkillsJSON serializes the collection of required skills into a JSON byte array for database persistence.
func (j *JobPosting) SkillsJSON() ([]byte, error) {
	if j.SkillsRequired == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(j.SkillsRequired)
}

// CandidateMatch holds computed matching scores, matching justifications, and profiles of candidate matches.
type CandidateMatch struct {
	UserID             uuid.UUID `json:"user_id"`
	Email              string    `json:"email"`
	University         string    `json:"university"`
	EmployabilityScore int       `json:"employability_score"`
	TotalXP            int       `json:"total_xp"`
	MatchScore         int       `json:"match_score"`
	MatchReasons       []string  `json:"match_reasons"`
}
