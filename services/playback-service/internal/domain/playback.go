package domain

import "time"

// PlaybackRequest represents a request to play video
type PlaybackRequest struct {
	CameraID  string    `json:"camera_id" validate:"required,uuid"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
	Format    string    `json:"format" validate:"required,oneof=hls rtsp"`
}

// PlaybackResponse represents the playback response
type PlaybackResponse struct {
	SessionID  string    `json:"session_id"`
	CameraID   string    `json:"camera_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Format     string    `json:"format"`
	URL        string    `json:"url"`
	ExpiresAt  time.Time `json:"expires_at"`
	SegmentIDs []string  `json:"segment_ids"`
}

// LiveStreamRequest represents a request for live streaming
type LiveStreamRequest struct {
	CameraID string `json:"camera_id" validate:"required,uuid"`
	Format   string `json:"format" validate:"required,oneof=hls rtsp webrtc"`
}

// LiveStreamResponse represents the live stream response
type LiveStreamResponse struct {
	SessionID string    `json:"session_id"`
	CameraID  string    `json:"camera_id"`
	Format    string    `json:"format"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ExportRequest represents a request to export video
type ExportRequest struct {
	CameraID  string    `json:"camera_id" validate:"required,uuid"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
	Format    string    `json:"format" validate:"required,oneof=mp4 avi mkv"`
	UserID    string    `json:"user_id" validate:"required"`
}

// ExportResponse represents the export response
type ExportResponse struct {
	ExportID   string    `json:"export_id"`
	CameraID   string    `json:"camera_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Format     string    `json:"format"`
	Status     string    `json:"status"` // pending, processing, completed, failed
	DownloadURL string   `json:"download_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Segment represents a video segment
type Segment struct {
	ID              string    `json:"id"`
	CameraID        string    `json:"camera_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationSeconds int       `json:"duration_seconds"`
	SizeBytes       int64     `json:"size_bytes"`
	StoragePath     string    `json:"storage_path"`
}

// PlaybackSource represents where video is stored
type PlaybackSource string

const (
	PlaybackSourceLocal     PlaybackSource = "LOCAL"      // MinIO (our storage)
	PlaybackSourceMilestone PlaybackSource = "MILESTONE"  // External Milestone VMS
)

// SourceDetectionResult contains information about detected video source
type SourceDetectionResult struct {
	Source       PlaybackSource `json:"source"`
	Available    bool           `json:"available"`
	SegmentCount int            `json:"segment_count"`
	TotalSize    int64          `json:"total_size"`    // bytes
	Reason       string         `json:"reason,omitempty"`
}

// PlaybackSession tracks an active playback session
type PlaybackSession struct {
	ID            string         `json:"id"`
	CameraID      string         `json:"camera_id"`
	UserID        string         `json:"user_id"`
	Source        PlaybackSource `json:"source"`
	Format        string         `json:"format"`
	StartTime     time.Time      `json:"start_time"`
	EndTime       time.Time      `json:"end_time"`
	ManifestPath  string         `json:"manifest_path"`  // Path to generated manifest
	SegmentPaths  []string       `json:"segment_paths"`  // Paths to video segments
	CreatedAt     time.Time      `json:"created_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
	LastAccessed  time.Time      `json:"last_accessed"`
}

// HLSSegment represents a single HLS segment
type HLSSegment struct {
	SegmentID    string    `json:"segment_id"`
	SequenceNum  int       `json:"sequence_num"`
	Duration     float64   `json:"duration"`       // seconds
	Path         string    `json:"path"`
	Size         int64     `json:"size"`           // bytes
	Cached       bool      `json:"cached"`
	CachedAt     time.Time `json:"cached_at,omitempty"`
}
