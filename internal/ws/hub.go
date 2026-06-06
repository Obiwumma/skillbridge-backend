// Package ws manages client WebSocket connection lifetimes and coordinates real-time message routing.
package ws

import (
	"encoding/json"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/pkg/logger"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected user WebSocket connection and its outbound message buffer.
type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub coordinates the registration, unregistration, and broadcasting of real-time WebSocket messages.
type Hub struct {
	clients    map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// GlobalHub is the singleton instance managing all socket transactions across the server application.
var GlobalHub *Hub

// Init bootstraps the dynamic GlobalHub and registers NATS event listeners.
func Init() {
	GlobalHub = &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go GlobalHub.Run()
	go GlobalHub.SubscribeNatsBroadcasts()
}

// Run executes the main select loop handling client registrations and unregistrations.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if _, exists := h.clients[client.UserID]; !exists {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mutex.Unlock()
			logger.Info("WebSocket client registered successfully", "user_id", client.UserID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if clients, exists := h.clients[client.UserID]; exists {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mutex.Unlock()
			logger.Info("WebSocket client unregistered successfully", "user_id", client.UserID)
		}
	}
}

// BroadcastToUser serializes and delivers JSON messages to all socket connections of a specific user.
func (h *Hub) BroadcastToUser(userID string, message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients, exists := h.clients[userID]
	if !exists {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal websocket broadcast", err)
		return
	}

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// SubscribeNatsBroadcasts binds NATS event bus subscriptions to push real-time updates directly to clients.
func (h *Hub) SubscribeNatsBroadcasts() {
	_, err := events.Subscribe(events.WorkspaceExecutionCompleted, func(data []byte) {
		var payload struct {
			UserID string `json:"user_id"`
		}
		err := json.Unmarshal(data, &payload)
		if err != nil {
			return
		}

		wsMsg := map[string]interface{}{
			"event": "workspace.execution.completed",
			"data":  json.RawMessage(data),
		}
		h.BroadcastToUser(payload.UserID, wsMsg)
	})
	if err != nil {
		logger.Error("WebSocket Hub failed NATS execution completed sub", err)
	}

	_, err = events.Subscribe("push.notification", func(data []byte) {
		var payload struct {
			UserID string `json:"user_id"`
		}
		err := json.Unmarshal(data, &payload)
		if err != nil {
			return
		}

		wsMsg := map[string]interface{}{
			"event": "notification.received",
			"data":  json.RawMessage(data),
		}
		h.BroadcastToUser(payload.UserID, wsMsg)
	})
	if err != nil {
		logger.Error("WebSocket Hub failed NATS push.notification sub", err)
	}

	_, err = events.Subscribe("resume.analysis.completed", func(data []byte) {
		var payload struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			h.BroadcastRawToUser(payload.UserID, data)
		}
	})
	if err != nil {
		logger.Error("WebSocket Hub failed NATS resume.analysis.completed sub", err)
	}

	_, err = events.Subscribe("code.review.completed", func(data []byte) {
		var payload struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			h.BroadcastRawToUser(payload.UserID, data)
		}
	})
	if err != nil {
		logger.Error("WebSocket Hub failed NATS code.review.completed sub", err)
	}

	_, err = events.Subscribe("recruiter.alert", func(data []byte) {
		var payload struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			h.BroadcastRawToUser(payload.UserID, data)
		}
	})
	if err != nil {
		logger.Error("WebSocket Hub failed NATS recruiter.alert sub", err)
	}
}

// BroadcastRawToUser transmits raw bytes directly to all active sockets of a specific user.
func (h *Hub) BroadcastRawToUser(userID string, data []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients, exists := h.clients[userID]
	if !exists {
		return
	}

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}
