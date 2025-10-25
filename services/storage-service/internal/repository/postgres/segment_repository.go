package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rta/cctv/storage-service/internal/domain"
	"github.com/rs/zerolog"
)

// SegmentRepository implements segment metadata storage in PostgreSQL
type SegmentRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewSegmentRepository creates a new PostgreSQL segment repository
func NewSegmentRepository(db *sql.DB, logger zerolog.Logger) *SegmentRepository {
	return &SegmentRepository{
		db:     db,
		logger: logger,
	}
}

// Create stores a segment metadata
func (r *SegmentRepository) Create(ctx context.Context, segment *domain.Segment) error {
	if segment.ID == "" {
		segment.ID = uuid.New().String()
	}

	query := `
		INSERT INTO segments (
			id, camera_id, start_time, end_time, duration_seconds,
			size_bytes, storage_backend, storage_path, checksum, thumbnail_path
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		segment.ID,
		segment.CameraID,
		segment.StartTime,
		segment.EndTime,
		segment.DurationSeconds,
		segment.SizeBytes,
		segment.StorageBackend,
		segment.StoragePath,
		segment.Checksum,
		segment.ThumbnailPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create segment: %w", err)
	}

	r.logger.Info().
		Str("segment_id", segment.ID).
		Str("camera_id", segment.CameraID).
		Msg("Segment metadata created")

	return nil
}

// Get retrieves a segment by ID
func (r *SegmentRepository) Get(ctx context.Context, id string) (*domain.Segment, error) {
	query := `
		SELECT id, camera_id, start_time, end_time, duration_seconds,
		       size_bytes, storage_backend, storage_path, checksum,
		       thumbnail_path, created_at
		FROM segments
		WHERE id = $1
	`

	segment := &domain.Segment{}
	var checksum, thumbnailPath sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&segment.ID,
		&segment.CameraID,
		&segment.StartTime,
		&segment.EndTime,
		&segment.DurationSeconds,
		&segment.SizeBytes,
		&segment.StorageBackend,
		&segment.StoragePath,
		&checksum,
		&thumbnailPath,
		&segment.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("segment not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get segment: %w", err)
	}

	if checksum.Valid {
		segment.Checksum = checksum.String
	}
	if thumbnailPath.Valid {
		segment.ThumbnailPath = thumbnailPath.String
	}

	return segment, nil
}

// List retrieves segments matching the query
func (r *SegmentRepository) List(ctx context.Context, query domain.ListSegmentsQuery) ([]*domain.Segment, error) {
	sqlQuery := `
		SELECT id, camera_id, start_time, end_time, duration_seconds,
		       size_bytes, storage_backend, storage_path, checksum,
		       thumbnail_path, created_at
		FROM segments
		WHERE camera_id = $1
		  AND start_time >= $2
		  AND end_time <= $3
		ORDER BY start_time ASC
		LIMIT $4 OFFSET $5
	`

	limit := query.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery,
		query.CameraID,
		query.StartTime,
		query.EndTime,
		limit,
		query.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}
	defer rows.Close()

	var segments []*domain.Segment

	for rows.Next() {
		segment := &domain.Segment{}
		var checksum, thumbnailPath sql.NullString

		err := rows.Scan(
			&segment.ID,
			&segment.CameraID,
			&segment.StartTime,
			&segment.EndTime,
			&segment.DurationSeconds,
			&segment.SizeBytes,
			&segment.StorageBackend,
			&segment.StoragePath,
			&checksum,
			&thumbnailPath,
			&segment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}

		if checksum.Valid {
			segment.Checksum = checksum.String
		}
		if thumbnailPath.Valid {
			segment.ThumbnailPath = thumbnailPath.String
		}

		segments = append(segments, segment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating segments: %w", err)
	}

	r.logger.Debug().
		Str("camera_id", query.CameraID).
		Int("count", len(segments)).
		Msg("Listed segments")

	return segments, nil
}

// Delete removes a segment metadata
func (r *SegmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM segments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete segment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("segment not found: %s", id)
	}

	r.logger.Info().
		Str("segment_id", id).
		Msg("Segment metadata deleted")

	return nil
}

// DeleteOlderThan removes segments older than the specified time
func (r *SegmentRepository) DeleteOlderThan(ctx context.Context, timestamp time.Time) (int64, error) {
	query := `DELETE FROM segments WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, timestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old segments: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info().
		Time("timestamp", timestamp).
		Int64("deleted", rowsAffected).
		Msg("Deleted old segments")

	return rowsAffected, nil
}
