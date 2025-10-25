package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rta/cctv/storage-service/internal/domain"
	"github.com/rs/zerolog"
)

// ExportRepository implements export metadata storage in PostgreSQL
type ExportRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewExportRepository creates a new PostgreSQL export repository
func NewExportRepository(db *sql.DB, logger zerolog.Logger) *ExportRepository {
	return &ExportRepository{
		db:     db,
		logger: logger,
	}
}

// Create stores an export metadata
func (r *ExportRepository) Create(ctx context.Context, export *domain.Export) error {
	if export.ID == "" {
		export.ID = uuid.New().String()
	}

	query := `
		INSERT INTO exports (
			id, camera_ids, start_time, end_time, format, reason,
			status, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		export.ID,
		pq.Array(export.CameraIDs),
		export.StartTime,
		export.EndTime,
		export.Format,
		export.Reason,
		export.Status,
		export.CreatedBy,
		export.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create export: %w", err)
	}

	r.logger.Info().
		Str("export_id", export.ID).
		Msg("Export metadata created")

	return nil
}

// Get retrieves an export by ID
func (r *ExportRepository) Get(ctx context.Context, id string) (*domain.Export, error) {
	query := `
		SELECT id, camera_ids, start_time, end_time, format, reason,
		       status, file_path, file_size, download_url, expires_at,
		       created_by, created_at, completed_at
		FROM exports
		WHERE id = $1
	`

	export := &domain.Export{}
	var filePath, downloadURL, createdBy, reason sql.NullString
	var fileSize sql.NullInt64
	var expiresAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&export.ID,
		pq.Array(&export.CameraIDs),
		&export.StartTime,
		&export.EndTime,
		&export.Format,
		&reason,
		&export.Status,
		&filePath,
		&fileSize,
		&downloadURL,
		&expiresAt,
		&createdBy,
		&export.CreatedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	if reason.Valid {
		export.Reason = reason.String
	}
	if filePath.Valid {
		export.FilePath = filePath.String
	}
	if fileSize.Valid {
		export.FileSize = fileSize.Int64
	}
	if downloadURL.Valid {
		export.DownloadURL = downloadURL.String
	}
	if expiresAt.Valid {
		export.ExpiresAt = &expiresAt.Time
	}
	if createdBy.Valid {
		export.CreatedBy = createdBy.String
	}
	if completedAt.Valid {
		export.CompletedAt = &completedAt.Time
	}

	return export, nil
}

// UpdateStatus updates the export status
func (r *ExportRepository) UpdateStatus(ctx context.Context, id string, status domain.ExportStatus, filePath string, fileSize int64, downloadURL string) error {
	query := `
		UPDATE exports
		SET status = $1, file_path = $2, file_size = $3, download_url = $4, completed_at = $5
		WHERE id = $6
	`

	completedAt := sql.NullTime{}
	if status == domain.ExportCompleted {
		completedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query, status, filePath, fileSize, downloadURL, completedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update export status: %w", err)
	}

	r.logger.Info().
		Str("export_id", id).
		Str("status", string(status)).
		Msg("Export status updated")

	return nil
}

// UpdateError updates the export with error information
func (r *ExportRepository) UpdateError(ctx context.Context, id string, errorMsg string) error {
	query := `UPDATE exports SET status = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, domain.ExportFailed, id)
	if err != nil {
		return fmt.Errorf("failed to update export error: %w", err)
	}

	r.logger.Error().
		Str("export_id", id).
		Str("error", errorMsg).
		Msg("Export failed")

	return nil
}

// DeleteExpired removes expired exports
func (r *ExportRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM exports WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired exports: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info().
		Int64("deleted", rowsAffected).
		Msg("Deleted expired exports")

	return rowsAffected, nil
}
