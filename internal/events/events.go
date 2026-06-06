// Package events orchestrates asynchronous, event-driven integrations powered by NATS.
package events

import (
	"encoding/json"
	"skillbridge-backend/pkg/config"
	"skillbridge-backend/pkg/logger"

	"github.com/nats-io/nats.go"
)

// Active event subject channels in our modular monolith.
const (
	UserRegistered              = "user.registered"
	CVUploaded                  = "cv.uploaded"
	SkillAnalysisCompleted      = "skill.analysis.completed"
	RoadmapGenerated            = "roadmap.generated"
	WorkspaceExecutionCompleted = "workspace.execution.completed"
	JobMatchUpdated             = "job.match.updated"
)

// NatsConn is the active NATS messaging server connection pointer.
var NatsConn *nats.Conn

// Init establishes connection registers to the distributed event broker.
func Init() {
	cfg := config.AppConfig
	var err error
	NatsConn, err = nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS event bus", err)
		if cfg.Env == "production" {
			panic(err)
		}
	} else {
		logger.Info("Successfully connected to NATS event bus")
	}
}

// Publish serializes and transmits a payload data object over a NATS channel topic.
func Publish(subject string, payload interface{}) error {
	if NatsConn == nil {
		logger.Warn("NATS connection not initialized, skipping publish: " + subject)
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal event payload", err)
		return err
	}

	err = NatsConn.Publish(subject, data)
	if err != nil {
		logger.Error("Failed to publish event to subject "+subject, err)
		return err
	}

	logger.Info("Published event to NATS", "subject", subject)
	return nil
}

// Subscribe registers a listener to execute a callback function concurrently in a background Goroutine when an event fires.
func Subscribe(subject string, handler func([]byte)) (*nats.Subscription, error) {
	if NatsConn == nil {
		logger.Warn("NATS connection not initialized, skipping subscribe: " + subject)
		return nil, nil
	}

	sub, err := NatsConn.Subscribe(subject, func(msg *nats.Msg) {
		logger.Info("Received NATS event", "subject", msg.Subject)
		go handler(msg.Data)
	})
	if err != nil {
		logger.Error("Failed to subscribe to subject "+subject, err)
		return nil, err
	}

	logger.Info("Subscribed to NATS subject", "subject", subject)
	return sub, nil
}
