package domain

import "time"

// ExportStatus represents the status of an export
type ExportStatus string

const (
	ExportPending    ExportStatus = "PENDING"
	ExportProcessing ExportStatus = "PROCESSING"
	ExportCompleted  ExportStatus = "COMPLETED"
	ExportFailed     ExportStatus = "FAILED"
)

// ExportFormat represents the export format
type ExportFormat string

const (
	FormatMP4 ExportFormat = "mp4"
	FormatAVI ExportFormat = "avi"
	FormatMKV ExportFormat = "mkv"
)

// Export represents a video export
type Export struct {
	ID          string       `json:"id"`
	CameraIDs   []string     `json:"camera_ids"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	Format      ExportFormat `json:"format"`
	Reason      string       `json:"reason,omitempty"`
	Status      ExportStatus `json:"status"`
	FilePath    string       `json:"file_path,omitempty"`
	FileSize    int64        `json:"file_size,omitempty"`
	DownloadURL string       `json:"download_url,omitempty"`
	ExpiresAt   *time.Time   `json:"expires_at,omitempty"`
	CreatedBy   string       `json:"created_by,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// CreateExportRequest represents a request to create an export
type CreateExportRequest struct {
	CameraIDs []string     `json:"camera_ids" validate:"required,min=1"`
	StartTime time.Time    `json:"start_time" validate:"required"`
	EndTime   time.Time    `json:"end_time" validate:"required"`
	Format    ExportFormat `json:"format" validate:"required,oneof=mp4 avi mkv"`
	Reason    string       `json:"reason"`
	CreatedBy string       `json:"created_by"`
}
