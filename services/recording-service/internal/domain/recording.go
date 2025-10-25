package domain

import "time"

// RecordingStatus represents the status of a recording
type RecordingStatus string

const (
	StatusStopped    RecordingStatus = "STOPPED"
	StatusStarting   RecordingStatus = "STARTING"
	StatusRecording  RecordingStatus = "RECORDING"
	StatusStopping   RecordingStatus = "STOPPING"
	StatusError      RecordingStatus = "ERROR"
)

// Recording represents an active recording session
type Recording struct {
	CameraID       string          `json:"camera_id"`
	CameraName     string          `json:"camera_name"`
	RTSPURL        string          `json:"rtsp_url"`
	Status         RecordingStatus `json:"status"`
	StartedAt      time.Time       `json:"started_at"`
	LastSegmentAt  time.Time       `json:"last_segment_at,omitempty"`
	SegmentCount   int             `json:"segment_count"`
	TotalBytes     int64           `json:"total_bytes"`
	Error          string          `json:"error,omitempty"`
	ReservationID  string          `json:"reservation_id,omitempty"`
}

// StartRecordingRequest represents a request to start recording
type StartRecordingRequest struct {
	CameraID string `json:"camera_id" validate:"required,uuid"`
}

// RecordingStats represents recording statistics
type RecordingStats struct {
	TotalRecordings    int                       `json:"total_recordings"`
	ActiveRecordings   int                       `json:"active_recordings"`
	TotalSegments      int64                     `json:"total_segments"`
	TotalBytes         int64                     `json:"total_bytes"`
	RecordingsByCamera map[string]*Recording `json:"recordings_by_camera"`
}

// SegmentInfo represents information about a recorded segment
type SegmentInfo struct {
	CameraID        string    `json:"camera_id"`
	FilePath        string    `json:"file_path"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationSeconds int       `json:"duration_seconds"`
	SizeBytes       int64     `json:"size_bytes"`
	ThumbnailPath   string    `json:"thumbnail_path,omitempty"`
}
