package domain

import "time"

// StorageBackend represents the storage backend type
type StorageBackend string

const (
	BackendMinIO      StorageBackend = "MINIO"
	BackendS3         StorageBackend = "S3"
	BackendFilesystem StorageBackend = "FILESYSTEM"
	BackendMilestone  StorageBackend = "MILESTONE"
)

// StorageMode represents the storage mode
type StorageMode string

const (
	ModeLocal     StorageMode = "LOCAL"
	ModeMilestone StorageMode = "MILESTONE"
	ModeBoth      StorageMode = "BOTH"
)

// Segment represents a video segment stored in the system
type Segment struct {
	ID              string         `json:"id"`
	CameraID        string         `json:"camera_id"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	DurationSeconds int            `json:"duration_seconds"`
	SizeBytes       int64          `json:"size_bytes"`
	StorageBackend  StorageBackend `json:"storage_backend"`
	StoragePath     string         `json:"storage_path"`
	Checksum        string         `json:"checksum,omitempty"`
	ThumbnailPath   string         `json:"thumbnail_path,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

// StoreSegmentRequest represents a request to store a segment
type StoreSegmentRequest struct {
	CameraID        string    `json:"camera_id" validate:"required,uuid"`
	StartTime       time.Time `json:"start_time" validate:"required"`
	EndTime         time.Time `json:"end_time" validate:"required"`
	FilePath        string    `json:"file_path" validate:"required"`
	SizeBytes       int64     `json:"size_bytes" validate:"required,min=1"`
	DurationSeconds int       `json:"duration_seconds" validate:"required,min=1"`
	ThumbnailPath   string    `json:"thumbnail_path,omitempty"`
}

// ListSegmentsQuery represents query parameters for listing segments
type ListSegmentsQuery struct {
	CameraID  string    `json:"camera_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}
