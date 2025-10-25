package domain

import "time"

// Camera represents a camera from VMS
type Camera struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	NameAr          string                 `json:"name_ar"`
	Source          string                 `json:"source"`  // DUBAI_POLICE, METRO, BUS, OTHER
	RTSPURL         string                 `json:"rtsp_url"`
	Status          string                 `json:"status"`  // ONLINE, OFFLINE, ERROR
	PTZEnabled      bool                   `json:"ptz_enabled"`
	RecordingServer string                 `json:"recording_server"`
	Metadata        map[string]interface{} `json:"metadata"`
	Location        *Location              `json:"location,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Location represents camera geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
	AddressAr string  `json:"address_ar,omitempty"`
}

// PTZCommand represents a PTZ control command
type PTZCommand struct {
	CameraID  string  `json:"camera_id" validate:"required,uuid"`
	Command   string  `json:"command" validate:"required,oneof=pan_left pan_right tilt_up tilt_down zoom_in zoom_out preset home"`
	Speed     float64 `json:"speed,omitempty"`     // 0.0 - 1.0
	PresetID  int     `json:"preset_id,omitempty"` // For preset command
	UserID    string  `json:"user_id" validate:"required"`
}

// CameraQuery represents a camera search query
type CameraQuery struct {
	Source   string `json:"source,omitempty"`
	Status   string `json:"status,omitempty"`
	Search   string `json:"search,omitempty"` // Search by name
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}
