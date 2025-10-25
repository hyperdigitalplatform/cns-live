package repository

import (
	"context"
	"io"
	"time"

	"github.com/rta/cctv/storage-service/internal/domain"
)

// StorageRepository defines the interface for storage operations
type StorageRepository interface {
	// Store uploads a file to storage
	Store(ctx context.Context, path string, reader io.Reader, size int64) error

	// Get retrieves a file from storage
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, path string) error

	// GenerateDownloadURL creates a temporary download URL
	GenerateDownloadURL(ctx context.Context, path string, expirySeconds int) (string, error)

	// List lists objects with a prefix
	List(ctx context.Context, prefix string) ([]string, error)
}

// SegmentRepository defines the interface for segment metadata operations
type SegmentRepository interface {
	// Create stores a segment metadata
	Create(ctx context.Context, segment *domain.Segment) error

	// Get retrieves a segment by ID
	Get(ctx context.Context, id string) (*domain.Segment, error)

	// List retrieves segments matching the query
	List(ctx context.Context, query domain.ListSegmentsQuery) ([]*domain.Segment, error)

	// Delete removes a segment metadata
	Delete(ctx context.Context, id string) error

	// DeleteOlderThan removes segments older than the specified time
	DeleteOlderThan(ctx context.Context, timestamp time.Time) (int64, error)
}

// ExportRepository defines the interface for export metadata operations
type ExportRepository interface {
	// Create stores an export metadata
	Create(ctx context.Context, export *domain.Export) error

	// Get retrieves an export by ID
	Get(ctx context.Context, id string) (*domain.Export, error)

	// UpdateStatus updates the export status
	UpdateStatus(ctx context.Context, id string, status domain.ExportStatus, filePath string, fileSize int64, downloadURL string) error

	// UpdateError updates the export with error information
	UpdateError(ctx context.Context, id string, errorMsg string) error

	// DeleteExpired removes expired exports
	DeleteExpired(ctx context.Context) (int64, error)
}
