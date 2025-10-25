package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/rta/cctv/vms-service/internal/domain"
	"github.com/rs/zerolog"
)

// PostgresRepository implements repository.CameraRepository using PostgreSQL
type PostgresRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB, logger zerolog.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// GetAll retrieves all cameras from the database
func (r *PostgresRepository) GetAll(ctx context.Context) ([]*domain.Camera, error) {
	query := `
		SELECT id, name, name_ar, source, rtsp_url, ptz_enabled, status,
		       recording_server, milestone_device_id, metadata, last_update, created_at
		FROM cameras
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cameras: %w", err)
	}
	defer rows.Close()

	cameras := make([]*domain.Camera, 0)
	for rows.Next() {
		camera, err := r.scanCamera(rows)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan camera")
			continue
		}
		cameras = append(cameras, camera)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cameras: %w", err)
	}

	return cameras, nil
}

// GetByID retrieves a single camera by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Camera, error) {
	query := `
		SELECT id, name, name_ar, source, rtsp_url, ptz_enabled, status,
		       recording_server, milestone_device_id, metadata, last_update, created_at
		FROM cameras
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	camera, err := r.scanCameraRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("camera not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get camera: %w", err)
	}

	return camera, nil
}

// GetBySource retrieves cameras filtered by source
func (r *PostgresRepository) GetBySource(ctx context.Context, source domain.CameraSource) ([]*domain.Camera, error) {
	query := `
		SELECT id, name, name_ar, source, rtsp_url, ptz_enabled, status,
		       recording_server, milestone_device_id, metadata, last_update, created_at
		FROM cameras
		WHERE source = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(source))
	if err != nil {
		return nil, fmt.Errorf("failed to query cameras by source: %w", err)
	}
	defer rows.Close()

	cameras := make([]*domain.Camera, 0)
	for rows.Next() {
		camera, err := r.scanCamera(rows)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan camera")
			continue
		}
		cameras = append(cameras, camera)
	}

	return cameras, nil
}

// GetRTSPURL generates RTSP URL for live streaming
func (r *PostgresRepository) GetRTSPURL(ctx context.Context, cameraID string) (string, error) {
	camera, err := r.GetByID(ctx, cameraID)
	if err != nil {
		return "", err
	}
	return camera.RTSPURL, nil
}

// GetPTZCapabilities retrieves PTZ capabilities for a camera
func (r *PostgresRepository) GetPTZCapabilities(ctx context.Context, cameraID string) (*domain.PTZCapabilities, error) {
	camera, err := r.GetByID(ctx, cameraID)
	if err != nil {
		return nil, err
	}

	if !camera.PTZEnabled {
		return &domain.PTZCapabilities{
			Pan:  false,
			Tilt: false,
			Zoom: false,
		}, nil
	}

	// Default PTZ capabilities for enabled cameras
	return &domain.PTZCapabilities{
		Pan:          true,
		Tilt:         true,
		Zoom:         true,
		Focus:        true,
		MaxPanSpeed:  1.0,
		MaxTiltSpeed: 1.0,
		MaxZoom:      20,
	}, nil
}

// ExecutePTZCommand sends PTZ command to camera
func (r *PostgresRepository) ExecutePTZCommand(ctx context.Context, cmd *domain.PTZCommand) error {
	// Verify camera exists and has PTZ enabled
	camera, err := r.GetByID(ctx, cmd.CameraID)
	if err != nil {
		return err
	}

	if !camera.PTZEnabled {
		return fmt.Errorf("camera %s does not support PTZ", cmd.CameraID)
	}

	// TODO: Implement actual Milestone SDK PTZ command
	r.logger.Info().
		Str("camera_id", cmd.CameraID).
		Str("action", string(cmd.Action)).
		Msg("PTZ command executed (mock)")

	return nil
}

// GetRecordingSegments retrieves available recording segments for time range
func (r *PostgresRepository) GetRecordingSegments(ctx context.Context, cameraID string, start, end time.Time) ([]*domain.RecordingSegment, error) {
	// Verify camera exists
	_, err := r.GetByID(ctx, cameraID)
	if err != nil {
		return nil, err
	}

	// TODO: Implement actual Milestone SDK recording query
	// For now, return mock data
	segments := []*domain.RecordingSegment{
		{
			StartTime: start,
			EndTime:   end,
			Available: true,
			SizeBytes: 1024 * 1024 * 1024, // 1GB mock
		},
	}

	return segments, nil
}

// ExportRecording creates an export job for recording
func (r *PostgresRepository) ExportRecording(ctx context.Context, req *domain.RecordingExportRequest) (*domain.RecordingExport, error) {
	// Verify camera exists
	_, err := r.GetByID(ctx, req.CameraID)
	if err != nil {
		return nil, err
	}

	// Generate export ID
	exportID := fmt.Sprintf("exp-%d", time.Now().Unix())

	// Insert export job
	query := `
		INSERT INTO recording_exports (id, camera_id, start_time, end_time, format, quality, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'PENDING', NOW())
		RETURNING id, camera_id, start_time, end_time, format, status, created_at
	`

	export := &domain.RecordingExport{}
	err = r.db.QueryRowContext(ctx, query,
		exportID, req.CameraID, req.StartTime, req.EndTime, req.Format, req.Quality,
	).Scan(&export.ID, &export.CameraID, &export.StartTime, &export.EndTime,
		&export.Format, &export.Status, &export.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	r.logger.Info().
		Str("export_id", exportID).
		Str("camera_id", req.CameraID).
		Msg("Export job created")

	return export, nil
}

// GetExportStatus retrieves status of an export job
func (r *PostgresRepository) GetExportStatus(ctx context.Context, exportID string) (*domain.RecordingExport, error) {
	query := `
		SELECT id, camera_id, start_time, end_time, format, status, file_path, file_size,
		       error, created_at, completed_at
		FROM recording_exports
		WHERE id = $1
	`

	export := &domain.RecordingExport{}
	var filePath, errorMsg sql.NullString
	var fileSize sql.NullInt64
	var completedAt pq.NullTime

	err := r.db.QueryRowContext(ctx, query, exportID).Scan(
		&export.ID, &export.CameraID, &export.StartTime, &export.EndTime,
		&export.Format, &export.Status, &filePath, &fileSize, &errorMsg,
		&export.CreatedAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export not found: %s", exportID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export status: %w", err)
	}

	if filePath.Valid {
		export.FilePath = filePath.String
	}
	if fileSize.Valid {
		export.FileSize = fileSize.Int64
	}
	if errorMsg.Valid {
		export.Error = errorMsg.String
	}
	if completedAt.Valid {
		export.CompletedAt = &completedAt.Time
	}

	return export, nil
}

// Sync performs full synchronization with Milestone VMS
func (r *PostgresRepository) Sync(ctx context.Context) error {
	// TODO: Implement actual Milestone SDK sync
	// This would:
	// 1. Connect to Milestone VMS
	// 2. Fetch all cameras
	// 3. Upsert to database
	r.logger.Info().Msg("Sync with Milestone VMS (not implemented)")
	return nil
}

// Helper functions

func (r *PostgresRepository) scanCamera(rows *sql.Rows) (*domain.Camera, error) {
	camera := &domain.Camera{}
	var metadataJSON []byte

	err := rows.Scan(
		&camera.ID,
		&camera.Name,
		&camera.NameAr,
		&camera.Source,
		&camera.RTSPURL,
		&camera.PTZEnabled,
		&camera.Status,
		&camera.RecordingServer,
		&camera.MilestoneDeviceID,
		&metadataJSON,
		&camera.LastUpdate,
		&camera.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &camera.Metadata); err != nil {
			r.logger.Warn().Err(err).Str("camera_id", camera.ID).Msg("Failed to unmarshal metadata")
		}
	}

	return camera, nil
}

func (r *PostgresRepository) scanCameraRow(row *sql.Row) (*domain.Camera, error) {
	camera := &domain.Camera{}
	var metadataJSON []byte

	err := row.Scan(
		&camera.ID,
		&camera.Name,
		&camera.NameAr,
		&camera.Source,
		&camera.RTSPURL,
		&camera.PTZEnabled,
		&camera.Status,
		&camera.RecordingServer,
		&camera.MilestoneDeviceID,
		&metadataJSON,
		&camera.LastUpdate,
		&camera.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &camera.Metadata); err != nil {
			r.logger.Warn().Err(err).Str("camera_id", camera.ID).Msg("Failed to unmarshal metadata")
		}
	}

	return camera, nil
}
