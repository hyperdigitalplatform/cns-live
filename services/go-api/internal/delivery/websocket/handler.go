package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub    *Hub
	logger zerolog.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, logger zerolog.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// ServeWS handles WebSocket requests from clients
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id")
	if userID == nil {
		userID = "anonymous"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upgrade connection")
		return
	}

	client := h.hub.NewClient(conn, userID.(string))
	h.hub.register <- client

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()
}
