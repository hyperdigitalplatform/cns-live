package domain

import "time"

// StreamReservation represents an active stream reservation
type StreamReservation struct {
	ID            string    `json:"id"`
	CameraID      string    `json:"camera_id"`
	CameraName    string    `json:"camera_name"`
	UserID        string    `json:"user_id"`
	Source        string    `json:"source"`        // DUBAI_POLICE, METRO, BUS, OTHER
	RoomName      string    `json:"room_name"`     // LiveKit room name
	Token         string    `json:"token"`         // LiveKit access token
	IngressID     string    `json:"ingress_id"`    // LiveKit ingress ID for RTSP stream
	ReservedAt    time.Time `json:"reserved_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// StreamRequest represents a request to reserve a stream
type StreamRequest struct {
	CameraID string `json:"camera_id" validate:"required,uuid"`
	UserID   string `json:"user_id" validate:"required"`
	Quality  string `json:"quality,omitempty"` // high, medium, low (default: medium)
}

// StreamResponse represents the response after stream reservation
type StreamResponse struct {
	ReservationID string    `json:"reservation_id"`
	CameraID      string    `json:"camera_id"`
	CameraName    string    `json:"camera_name"`
	RoomName      string    `json:"room_name"`
	Token         string    `json:"token"`
	LiveKitURL    string    `json:"livekit_url"`
	ExpiresAt     time.Time `json:"expires_at"`
	Quality       string    `json:"quality"`
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	ReservationID string `json:"reservation_id" validate:"required,uuid"`
}

// StreamStats represents real-time stream statistics
type StreamStats struct {
	ActiveStreams  int                 `json:"active_streams"`
	TotalViewers   int                 `json:"total_viewers"`
	SourceStats    map[string]SourceStat `json:"source_stats"`
	CameraStats    []CameraStat        `json:"camera_stats"`
	Timestamp      time.Time           `json:"timestamp"`
}

// SourceStat represents statistics per source
type SourceStat struct {
	Source        string  `json:"source"`
	Current       int     `json:"current"`
	Limit         int     `json:"limit"`
	UsagePercent  float64 `json:"usage_percent"`
	ActiveCameras int     `json:"active_cameras"`
}

// CameraStat represents statistics per camera
type CameraStat struct {
	CameraID      string    `json:"camera_id"`
	CameraName    string    `json:"camera_name"`
	ViewerCount   int       `json:"viewer_count"`
	Source        string    `json:"source"`
	ActiveSince   time.Time `json:"active_since"`
}

// AgencyLimitError represents an agency limit exceeded error
type AgencyLimitError struct {
	Source  string `json:"source"`
	Current int    `json:"current"`
	Limit   int    `json:"limit"`
	Message string `json:"message"`
}

func (e *AgencyLimitError) Error() string {
	return e.Message
}
