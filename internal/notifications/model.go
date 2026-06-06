// Package notifications coordinates systemic alert routing and delivery for real-time channels.
package notifications

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a system or transactional alert routed to a specific user.
type Notification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}
