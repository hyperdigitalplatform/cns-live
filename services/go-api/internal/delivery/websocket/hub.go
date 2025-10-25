package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rta/cctv/go-api/internal/usecase"
	"github.com/rs/zerolog"
)

// MessageType represents WebSocket message types
type MessageType string

const (
	MessageTypeStreamStats    MessageType = "STREAM_STATS"
	MessageTypeCameraStatus   MessageType = "CAMERA_STATUS"
	MessageTypeAgencyLimit    MessageType = "AGENCY_LIMIT_UPDATE"
	MessageTypeAlert          MessageType = "ALERT"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan Message
	userID   string
	logger   zerolog.Logger
}

// Hub manages WebSocket connections and broadcasts
type Hub struct {
	clients       map[*Client]bool
	broadcast     chan Message
	register      chan *Client
	unregister    chan *Client
	streamUseCase *usecase.StreamUseCase
	logger        zerolog.Logger
	mu            sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(streamUseCase *usecase.StreamUseCase, logger zerolog.Logger) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		broadcast:     make(chan Message, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		streamUseCase: streamUseCase,
		logger:        logger,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	// Start stats broadcaster
	go h.broadcastStats(ctx)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info().Msg("Hub shutting down")
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info().Str("user_id", client.userID).Msg("Client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Info().Str("user_id", client.userID).Msg("Client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send channel is full, disconnect
					h.mu.RUnlock()
					h.unregister <- client
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// broadcastStats periodically broadcasts stream statistics
func (h *Hub) broadcastStats(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			stats, err := h.streamUseCase.GetStreamStats(ctx)
			if err != nil {
				h.logger.Error().Err(err).Msg("Failed to get stream stats")
				continue
			}

			h.broadcast <- Message{
				Type:      MessageTypeStreamStats,
				Data:      stats,
				Timestamp: time.Now(),
			}
		}
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *Hub) BroadcastMessage(msgType MessageType, data interface{}) {
	h.broadcast <- Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// BroadcastAgencyLimitUpdate broadcasts agency limit update
func (h *Hub) BroadcastAgencyLimitUpdate(source string, current, limit int) {
	h.BroadcastMessage(MessageTypeAgencyLimit, map[string]interface{}{
		"source":        source,
		"current":       current,
		"limit":         limit,
		"usage_percent": float64(current) / float64(limit) * 100,
	})
}

// BroadcastCameraStatus broadcasts camera status update
func (h *Hub) BroadcastCameraStatus(cameraID, status string) {
	h.BroadcastMessage(MessageTypeCameraStatus, map[string]interface{}{
		"camera_id": cameraID,
		"status":    status,
	})
}

// BroadcastAlert broadcasts an alert
func (h *Hub) BroadcastAlert(alertType, message string, severity string) {
	h.BroadcastMessage(MessageTypeAlert, map[string]interface{}{
		"alert_type": alertType,
		"message":    message,
		"severity":   severity,
	})
}

// NewClient creates a new WebSocket client
func (h *Hub) NewClient(conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan Message, 256),
		userID: userID,
		logger: h.logger.With().Str("user_id", userID).Logger(),
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error().Err(err).Msg("WebSocket error")
			}
			break
		}

		// Handle client messages (e.g., subscriptions, filters)
		c.handleMessage(message)
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Write message
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(message); err != nil {
				c.logger.Error().Err(err).Msg("Failed to encode message")
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles messages from the client
func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal client message")
		return
	}

	// Handle different message types
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "subscribe":
		// Handle subscription (e.g., subscribe to specific cameras)
		c.logger.Debug().Msg("Client subscribed")
	case "unsubscribe":
		// Handle unsubscription
		c.logger.Debug().Msg("Client unsubscribed")
	default:
		c.logger.Warn().Str("type", msgType).Msg("Unknown message type")
	}
}
