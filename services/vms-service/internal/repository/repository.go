package repository

import (
	"context"
	"time"

	"github.com/rta/cctv/vms-service/internal/domain"
)

// CameraRepository defines operations for camera data
type CameraRepository interface {
	// GetAll retrieves all cameras from Milestone VMS
	GetAll(ctx context.Context) ([]*domain.Camera, error)

	// GetByID retrieves a single camera by ID
	GetByID(ctx context.Context, id string) (*domain.Camera, error)

	// GetBySource retrieves cameras filtered by source
	GetBySource(ctx context.Context, source domain.CameraSource) ([]*domain.Camera, error)

	// GetRTSPURL generates RTSP URL for live streaming
	GetRTSPURL(ctx context.Context, cameraID string) (string, error)

	// GetPTZCapabilities retrieves PTZ capabilities for a camera
	GetPTZCapabilities(ctx context.Context, cameraID string) (*domain.PTZCapabilities, error)

	// ExecutePTZCommand sends PTZ command to camera
	ExecutePTZCommand(ctx context.Context, cmd *domain.PTZCommand) error

	// GetRecordingSegments retrieves available recording segments for time range
	GetRecordingSegments(ctx context.Context, cameraID string, start, end time.Time) ([]*domain.RecordingSegment, error)

	// ExportRecording creates an export job for recording
	ExportRecording(ctx context.Context, req *domain.RecordingExportRequest) (*domain.RecordingExport, error)

	// GetExportStatus retrieves status of an export job
	GetExportStatus(ctx context.Context, exportID string) (*domain.RecordingExport, error)

	// Sync performs full synchronization with Milestone VMS
	Sync(ctx context.Context) error
}

// CacheRepository defines operations for caching camera data
type CacheRepository interface {
	// Set stores camera data in cache
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get retrieves camera data from cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete removes data from cache
	Delete(ctx context.Context, key string) error

	// Invalidate clears all cache entries
	Invalidate(ctx context.Context) error
}
