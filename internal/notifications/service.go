// Package notifications coordinates systemic alert routing and delivery for real-time channels.
package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/logger"

	"github.com/google/uuid"
)

// NotificationService coordinates creation, read status updates, and event-driven delivery of user notifications.
type NotificationService interface {
	SendNotification(ctx context.Context, userID uuid.UUID, title, message, alertType string) (*Notification, error)
	GetNotifications(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	SubscribeToEvents()
}

type notificationService struct {
	repo NotificationRepository
}

// NewNotificationService constructs an instance of NotificationService.
func NewNotificationService(repo NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

// SendNotification writes an alert and publishes it to the WebSocket real-time push network.
func (s *notificationService) SendNotification(ctx context.Context, userID uuid.UUID, title, message, alertType string) (*Notification, error) {
	n := &Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    alertType,
		Read:    false,
	}

	err := s.repo.Create(ctx, n)
	if err != nil {
		logger.Error("Failed to write in-app notification", err)
		return nil, errors.NewInternalError("failed to create notification", err)
	}

	eventPayload := map[string]interface{}{
		"user_id":         userID.String(),
		"notification_id": n.ID.String(),
		"title":           n.Title,
		"message":         n.Message,
		"type":            n.Type,
	}
	_ = events.Publish("push.notification", eventPayload)

	logger.Info("Dispatched notification to user", "user_id", userID.String(), "title", title)
	return n, nil
}

// GetNotifications fetches chronological notifications for a specific user.
func (s *notificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// MarkRead flags a specific notification as read.
func (s *notificationService) MarkRead(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id)
}

// SubscribeToEvents registers event listeners to trigger automatic system notifications.
func (s *notificationService) SubscribeToEvents() {
	_, err := events.Subscribe(events.UserRegistered, func(data []byte) {
		logger.Info("Notifications module received user.registered event")

		var payload struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		}

		_ = json.Unmarshal(data, &payload)
		uid, _ := uuid.Parse(payload.UserID)

		ctx := context.Background()
		_, _ = s.SendNotification(ctx, uid,
			"Welcome to Skillbridge!",
			"Congratulations on setting up your account. Upload your CV to begin adaptive learning pathway generation!",
			"milestone",
		)
	})
	if err != nil {
		logger.Error("Notifications module failed welcome subscription", err)
	}

	_, err = events.Subscribe(events.WorkspaceExecutionCompleted, func(data []byte) {
		logger.Info("Notifications module received workspace.execution.completed event")

		var payload struct {
			UserID   string `json:"user_id"`
			Language string `json:"language"`
			Status   string `json:"status"`
			Score    int    `json:"score"`
		}

		_ = json.Unmarshal(data, &payload)
		uid, _ := uuid.Parse(payload.UserID)

		ctx := context.Background()
		_, _ = s.SendNotification(ctx, uid,
			"Code Execution Analyzed",
			fmt.Sprintf("Your %s execution yielded status: %s with an AI review score of %d%%.", payload.Language, payload.Status, payload.Score),
			"workspace",
		)
	})
	if err != nil {
		logger.Error("Notifications module failed execution subscription", err)
	}
}
