package domain

import "time"

// CameraSource represents the agency/source of the camera
type CameraSource string

const (
	SourceDubaiPolice CameraSource = "DUBAI_POLICE"
	SourceMetro       CameraSource = "METRO"
	SourceBus         CameraSource = "BUS"
	SourceOther       CameraSource = "OTHER"
)

// IsValid checks if camera source is valid
func (s CameraSource) IsValid() bool {
	switch s {
	case SourceDubaiPolice, SourceMetro, SourceBus, SourceOther:
		return true
	}
	return false
}

// StreamReservation represents a reserved stream slot
type StreamReservation struct {
	ID        string       `json:"id"`
	CameraID  string       `json:"camera_id"`
	UserID    string       `json:"user_id"`
	Source    CameraSource `json:"source"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// ReserveRequest represents a request to reserve a stream
type ReserveRequest struct {
	CameraID string       `json:"camera_id" validate:"required,uuid"`
	UserID   string       `json:"user_id" validate:"required"`
	Source   CameraSource `json:"source" validate:"required"`
	Duration int          `json:"duration" validate:"required,min=60,max=7200"` // seconds (1 min to 2 hours)
}

// ReserveResponse represents the response to a reserve request
type ReserveResponse struct {
	ReservationID string       `json:"reservation_id"`
	CameraID      string       `json:"camera_id"`
	UserID        string       `json:"user_id"`
	Source        CameraSource `json:"source"`
	ExpiresAt     time.Time    `json:"expires_at"`
	CurrentUsage  int          `json:"current_usage"`
	Limit         int          `json:"limit"`
}

// ReleaseRequest represents a request to release a stream
type ReleaseRequest struct {
	ReservationID string `json:"reservation_id" validate:"required,uuid"`
}

// ReleaseResponse represents the response to a release request
type ReleaseResponse struct {
	ReservationID string       `json:"reservation_id"`
	Source        CameraSource `json:"source"`
	Released      bool         `json:"released"`
	NewCount      int          `json:"new_count"`
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	ReservationID string `json:"reservation_id" validate:"required,uuid"`
	ExtendTTL     int    `json:"extend_ttl,omitempty"` // seconds (optional, default 60)
}

// HeartbeatResponse represents the response to a heartbeat
type HeartbeatResponse struct {
	ReservationID string `json:"reservation_id"`
	RemainingTTL  int    `json:"remaining_ttl"` // seconds
	Updated       bool   `json:"updated"`
}

// StreamStats represents stream usage statistics
type StreamStats struct {
	Source     CameraSource `json:"source"`
	Current    int          `json:"current"`
	Limit      int          `json:"limit"`
	Percentage int          `json:"percentage"`
	Available  int          `json:"available"`
}

// StatsResponse represents the response to a stats request
type StatsResponse struct {
	Stats     []StreamStats `json:"stats"`
	Total     StatsInfo     `json:"total"`
	Timestamp time.Time     `json:"timestamp"`
}

// StatsInfo represents aggregate statistics
type StatsInfo struct {
	Current    int `json:"current"`
	Limit      int `json:"limit"`
	Percentage int `json:"percentage"`
	Available  int `json:"available"`
}

// LimitConfig represents stream limits configuration
type LimitConfig struct {
	DubaiPolice int `json:"dubai_police"`
	Metro       int `json:"metro"`
	Bus         int `json:"bus"`
	Other       int `json:"other"`
	Total       int `json:"total"`
}

// GetLimit returns the limit for a given source
func (c *LimitConfig) GetLimit(source CameraSource) int {
	switch source {
	case SourceDubaiPolice:
		return c.DubaiPolice
	case SourceMetro:
		return c.Metro
	case SourceBus:
		return c.Bus
	case SourceOther:
		return c.Other
	default:
		return 0
	}
}
