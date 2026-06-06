// Package notifications coordinates systemic alert routing and delivery for real-time channels.
package notifications

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// NotificationRepository provides persistent operations for reading and writing user alerts.
type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
}

type sqlNotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository constructs a postgreSQL execution client implementation for NotificationRepository.
func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &sqlNotificationRepository{db: db}
}

// Create persists a new notification record in the database.
func (r *sqlNotificationRepository) Create(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, title, message, type, read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	n.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		n.ID,
		n.UserID,
		n.Title,
		n.Message,
		n.Type,
		n.Read,
		n.CreatedAt,
	)
	return err
}

// GetByUserID retrieves all registered notifications for a given user ordered by creation time.
func (r *sqlNotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	query := `
		SELECT id, user_id, title, message, type, read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Title,
			&n.Message,
			&n.Type,
			&n.Read,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, n)
	}

	return alerts, nil
}

// MarkAsRead flags a specific notification ID as read to mute repeat alerts.
func (r *sqlNotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET read = true
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
