// Package roadmap implements branching graph architectures representing adaptive student learning pathways.
package roadmap

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// RoadmapRepository defines datastore access specifications for mapping active student roadmaps.
type RoadmapRepository interface {
	Create(ctx context.Context, roadmap *Roadmap) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Roadmap, error)
	Update(ctx context.Context, roadmap *Roadmap) error
}

type sqlRoadmapRepository struct {
	db *sql.DB
}

// NewRoadmapRepository instantiates a new PostgreSQL repository mapped to DB pool connection registers.
func NewRoadmapRepository(db *sql.DB) RoadmapRepository {
	return &sqlRoadmapRepository{db: db}
}

// Create inserts a new student dynamic learning path graph record in PostgreSQL.
func (r *sqlRoadmapRepository) Create(ctx context.Context, roadmap *Roadmap) error {
	query := `
		INSERT INTO roadmaps (id, user_id, title, nodes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	nodesJSON, err := roadmap.NodesJSON()
	if err != nil {
		return err
	}

	now := time.Now()
	roadmap.CreatedAt = now
	roadmap.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		roadmap.ID,
		roadmap.UserID,
		roadmap.Title,
		nodesJSON,
		roadmap.CreatedAt,
		roadmap.UpdatedAt,
	)
	return err
}

// GetByUserID retrieves student learning paths DAG matching the user ID, parsing the JSONB graph payload.
func (r *sqlRoadmapRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Roadmap, error) {
	query := `
		SELECT id, user_id, title, nodes, created_at, updated_at
		FROM roadmaps
		WHERE user_id = $1
	`
	var rm Roadmap
	var nodesBytes []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&rm.ID,
		&rm.UserID,
		&rm.Title,
		&nodesBytes,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	rm.Nodes = make(map[string]*RoadmapNode)
	if len(nodesBytes) > 0 {
		err = json.Unmarshal(nodesBytes, &rm.Nodes)
		if err != nil {
			return nil, err
		}
	}

	return &rm, nil
}

// Update persists structural mutations inside student learning graph DAGs.
func (r *sqlRoadmapRepository) Update(ctx context.Context, roadmap *Roadmap) error {
	query := `
		UPDATE roadmaps
		SET title = $1, nodes = $2, updated_at = $3
		WHERE id = $4
	`
	nodesJSON, err := roadmap.NodesJSON()
	if err != nil {
		return err
	}

	roadmap.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		roadmap.Title,
		nodesJSON,
		roadmap.UpdatedAt,
		roadmap.ID,
	)
	return err
}
