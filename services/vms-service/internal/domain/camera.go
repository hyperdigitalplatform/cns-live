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

// CameraStatus represents the operational status of a camera
type CameraStatus string

const (
	StatusOnline  CameraStatus = "ONLINE"
	StatusOffline CameraStatus = "OFFLINE"
	StatusError   CameraStatus = "ERROR"
)

// Camera represents a CCTV camera from Milestone VMS
type Camera struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	NameAr            string                 `json:"name_ar"`
	Source            CameraSource           `json:"source"`
	RTSPURL           string                 `json:"rtsp_url"`
	PTZEnabled        bool                   `json:"ptz_enabled"`
	Status            CameraStatus           `json:"status"`
	RecordingServer   string                 `json:"recording_server"`
	MilestoneDeviceID string                 `json:"milestone_device_id"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	LastUpdate        time.Time              `json:"last_update"`
	CreatedAt         time.Time              `json:"created_at"`
}

// PTZCapabilities represents PTZ control capabilities
type PTZCapabilities struct {
	Pan         bool    `json:"pan"`
	Tilt        bool    `json:"tilt"`
	Zoom        bool    `json:"zoom"`
	Focus       bool    `json:"focus"`
	MaxPanSpeed float64 `json:"max_pan_speed"`
	MaxTiltSpeed float64 `json:"max_tilt_speed"`
	MaxZoom     int     `json:"max_zoom"`
}

// PTZCommand represents a PTZ control command
type PTZCommand struct {
	CameraID string      `json:"camera_id"`
	Action   PTZAction   `json:"action"`
	Pan      float64     `json:"pan,omitempty"`       // -1.0 to 1.0
	Tilt     float64     `json:"tilt,omitempty"`      // -1.0 to 1.0
	Zoom     float64     `json:"zoom,omitempty"`      // 0.0 to 1.0
	Speed    float64     `json:"speed,omitempty"`     // 0.0 to 1.0
	Preset   int         `json:"preset,omitempty"`    // Preset number
}

// PTZAction represents PTZ command type
type PTZAction string

const (
	PTZActionMove           PTZAction = "MOVE"
	PTZActionStop           PTZAction = "STOP"
	PTZActionGoToPreset     PTZAction = "GO_TO_PRESET"
	PTZActionSetPreset      PTZAction = "SET_PRESET"
	PTZActionClearPreset    PTZAction = "CLEAR_PRESET"
)

// RecordingExportRequest represents a request to export recording
type RecordingExportRequest struct {
	CameraID  string    `json:"camera_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Format    string    `json:"format"`     // MP4, AVI, MKV
	Quality   string    `json:"quality"`    // HIGH, MEDIUM, LOW
}

// RecordingExport represents an export job
type RecordingExport struct {
	ID         string    `json:"id"`
	CameraID   string    `json:"camera_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Format     string    `json:"format"`
	Status     string    `json:"status"`      // PENDING, PROCESSING, COMPLETED, FAILED
	FilePath   string    `json:"file_path,omitempty"`
	FileSize   int64     `json:"file_size,omitempty"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// RecordingSegment represents a time segment with recording
type RecordingSegment struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
	SizeBytes int64     `json:"size_bytes,omitempty"`
}
