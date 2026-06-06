// Package roadmap implements branching graph architectures representing adaptive student learning pathways.
package roadmap

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UnlockCondition captures specific skill level thresholds required to unlock a learning node.
type UnlockCondition struct {
	SkillName        string `json:"skill_name"`
	MinLevelRequired int    `json:"min_level_required"`
}

// RoadmapNode represents a milestone node in a dynamic, non-linear Directed Acyclic Graph (DAG) learning track.
type RoadmapNode struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Prerequisites    []string          `json:"prerequisites"`
	UnlockConditions []UnlockCondition `json:"unlock_conditions"`
	XPReward         int               `json:"xp_reward"`
	SkillsUnlocked   []string          `json:"skills_unlocked"`
	Children         []string          `json:"children"`
	Status           string            `json:"status"` // Can be "locked", "unlocked", or "completed"
}

// Roadmap represents the student learning pathways DAG containing active nodes and edges.
type Roadmap struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Title     string                 `json:"title"`
	Nodes     map[string]*RoadmapNode `json:"nodes"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NodesJSON marshals the roadmap DAG nodes to JSON bytes for PostgreSQL JSONB storage.
func (r *Roadmap) NodesJSON() ([]byte, error) {
	if r.Nodes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(r.Nodes)
}
