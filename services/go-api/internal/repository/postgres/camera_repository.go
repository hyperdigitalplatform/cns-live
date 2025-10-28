package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rta/cctv/go-api/internal/domain"
)

// CameraRepository handles camera database operations
type CameraRepository struct {
	db *sql.DB
}

// NewCameraRepository creates a new camera repository
func NewCameraRepository(db *sql.DB) *CameraRepository {
	return &CameraRepository{db: db}
}

// ImportCamera saves a discovered camera to the database
func (r *CameraRepository) ImportCamera(ctx context.Context, camera *domain.ImportCameraRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare metadata JSON
	metadataJSON, err := json.Marshal(map[string]interface{}{
		"ptzCapabilities": camera.PtzCapabilities,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert or update camera
	query := `
		INSERT INTO cameras (
			id, name, name_ar, source, rtsp_url, ptz_enabled, status,
			recording_server, milestone_device_id, metadata,
			ip_address, onvif_port, manufacturer, model, firmware_version,
			serial_number, hardware_id, onvif_endpoint, onvif_username,
			onvif_password_encrypted, created_at, last_update
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			name_ar = EXCLUDED.name_ar,
			rtsp_url = EXCLUDED.rtsp_url,
			ptz_enabled = EXCLUDED.ptz_enabled,
			status = EXCLUDED.status,
			ip_address = EXCLUDED.ip_address,
			onvif_port = EXCLUDED.onvif_port,
			manufacturer = EXCLUDED.manufacturer,
			model = EXCLUDED.model,
			firmware_version = EXCLUDED.firmware_version,
			serial_number = EXCLUDED.serial_number,
			hardware_id = EXCLUDED.hardware_id,
			onvif_endpoint = EXCLUDED.onvif_endpoint,
			onvif_username = EXCLUDED.onvif_username,
			onvif_password_encrypted = EXCLUDED.onvif_password_encrypted,
			metadata = EXCLUDED.metadata,
			last_update = NOW()
	`

	// Get primary stream RTSP URL (use first stream or main stream)
	primaryRtspURL := ""
	if len(camera.Streams) > 0 {
		// Try to find main/primary stream
		for _, stream := range camera.Streams {
			if stream.Name == "mainStream" || stream.Name == "PROFILE_000" || stream.IsPrimary {
				primaryRtspURL = stream.RtspURL
				break
			}
		}
		// If no primary found, use first stream
		if primaryRtspURL == "" {
			primaryRtspURL = camera.Streams[0].RtspURL
		}

		// Determine credentials to use
		username := camera.OnvifUsername
		password := camera.OnvifPassword

		// If password not provided, try to get demo credentials by IP
		if password == "" && camera.Device.IP != "" {
			if demoUser, demoPass, found := getDemoCredentials(camera.Device.IP); found {
				username = demoUser
				password = demoPass
			}
		}

		// Add credentials to RTSP URL if we have them
		if username != "" && password != "" {
			primaryRtspURL = addRTSPCredentials(primaryRtspURL, username, password)
		}
	}

	_, err = tx.ExecContext(ctx, query,
		camera.MilestoneID,      // id
		camera.Name,             // name
		camera.NameAr,           // name_ar
		camera.Source,           // source
		primaryRtspURL,          // rtsp_url
		camera.PtzEnabled,       // ptz_enabled
		camera.Status,           // status
		camera.RecordingServer,  // recording_server
		camera.MilestoneID,      // milestone_device_id
		metadataJSON,            // metadata
		camera.Device.IP,        // ip_address
		camera.Device.Port,      // onvif_port
		camera.Device.Manufacturer,    // manufacturer
		camera.Device.Model,            // model
		camera.Device.FirmwareVersion,  // firmware_version
		camera.Device.SerialNumber,     // serial_number
		camera.Device.HardwareID,       // hardware_id
		camera.OnvifEndpoint,           // onvif_endpoint
		camera.OnvifUsername,           // onvif_username
		camera.OnvifPasswordEncrypted,  // onvif_password_encrypted
	)
	if err != nil {
		return fmt.Errorf("failed to insert camera: %w", err)
	}

	// Delete existing streams for this camera
	_, err = tx.ExecContext(ctx, "DELETE FROM camera_streams WHERE camera_id = $1", camera.MilestoneID)
	if err != nil {
		return fmt.Errorf("failed to delete old streams: %w", err)
	}

	// Insert stream profiles
	streamQuery := `
		INSERT INTO camera_streams (
			camera_id, profile_token, profile_name, encoding, resolution,
			width, height, frame_rate, bitrate, rtsp_url, is_primary,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
	`

	for _, stream := range camera.Streams {
		_, err = tx.ExecContext(ctx, streamQuery,
			camera.MilestoneID,
			stream.ProfileToken,
			stream.Name,
			stream.Encoding,
			stream.Resolution,
			stream.Width,
			stream.Height,
			stream.FrameRate,
			stream.Bitrate,
			stream.RtspURL,
			stream.IsPrimary,
		)
		if err != nil {
			return fmt.Errorf("failed to insert stream profile: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCamera retrieves a camera by ID with its streams
func (r *CameraRepository) GetCamera(ctx context.Context, id string) (*domain.Camera, error) {
	query := `
		SELECT id, name, name_ar, source, rtsp_url, ptz_enabled, status,
		       recording_server, milestone_device_id, metadata, created_at, last_update,
		       ip_address, onvif_port, manufacturer, model, firmware_version
		FROM cameras
		WHERE id = $1
	`

	var camera domain.Camera
	var metadataJSON []byte
	var ipAddress, manufacturer, model, firmwareVersion sql.NullString
	var milestoneDeviceID sql.NullString
	var onvifPort sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&camera.ID, &camera.Name, &camera.NameAr, &camera.Source, &camera.RTSPURL,
		&camera.PTZEnabled, &camera.Status, &camera.RecordingServer, &milestoneDeviceID, &metadataJSON,
		&camera.CreatedAt, &camera.UpdatedAt,
		&ipAddress, &onvifPort, &manufacturer, &model, &firmwareVersion,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("camera not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get camera: %w", err)
	}

	// Set milestone device ID if present
	if milestoneDeviceID.Valid {
		camera.MilestoneDeviceID = milestoneDeviceID.String
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &camera.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &camera, nil
}

// ListCameras retrieves cameras with optional filters
func (r *CameraRepository) ListCameras(ctx context.Context, query domain.CameraQuery) ([]*domain.Camera, error) {
	sqlQuery := `
		SELECT id, name, name_ar, source, rtsp_url, ptz_enabled, status,
		       recording_server, milestone_device_id, metadata, created_at, last_update
		FROM cameras
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if query.Source != "" {
		sqlQuery += fmt.Sprintf(" AND source = $%d", argCount)
		args = append(args, query.Source)
		argCount++
	}

	if query.Status != "" {
		sqlQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, query.Status)
		argCount++
	}

	if query.Search != "" {
		sqlQuery += fmt.Sprintf(" AND (name ILIKE $%d OR name_ar ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+query.Search+"%")
		argCount++
	}

	sqlQuery += " ORDER BY created_at DESC"

	if query.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, query.Limit)
		argCount++
	}

	if query.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, query.Offset)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list cameras: %w", err)
	}
	defer rows.Close()

	cameras := []*domain.Camera{}
	for rows.Next() {
		var camera domain.Camera
		var metadataJSON []byte
		var milestoneDeviceID sql.NullString

		err := rows.Scan(
			&camera.ID, &camera.Name, &camera.NameAr, &camera.Source, &camera.RTSPURL,
			&camera.PTZEnabled, &camera.Status, &camera.RecordingServer, &milestoneDeviceID, &metadataJSON,
			&camera.CreatedAt, &camera.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan camera: %w", err)
		}

		// Set milestone device ID if present
		if milestoneDeviceID.Valid {
			camera.MilestoneDeviceID = milestoneDeviceID.String
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &camera.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		cameras = append(cameras, &camera)
	}

	return cameras, nil
}

// demoCredentials maps camera IP addresses to their credentials
// These are the predefined demo credentials for development/testing
var demoCredentials = map[string]struct {
	username string
	password string
}{
	"192.168.1.8":  {"raammohan", "Ilove123"}, // TP-Link Tapo camera
	"192.168.1.13": {"admin", "pass"},          // Guangzhou PTZ camera
	// Add more demo camera credentials here as needed
}

// getDemoCredentials returns the demo credentials for a given IP address
func getDemoCredentials(ipAddress string) (username, password string, found bool) {
	if creds, ok := demoCredentials[ipAddress]; ok {
		return creds.username, creds.password, true
	}
	return "", "", false
}

// addRTSPCredentials adds username:password to an RTSP URL
// Example: rtsp://192.168.1.13:554/path -> rtsp://admin:pass@192.168.1.13:554/path
func addRTSPCredentials(rtspURL, username, password string) string {
	// Parse the RTSP URL to insert credentials
	// Format: rtsp://[username:password@]host[:port]/path

	// Check if URL already has credentials
	if len(rtspURL) < 7 || rtspURL[:7] != "rtsp://" {
		return rtspURL // Invalid URL, return as-is
	}

	// Check if credentials already exist (has @ before the first /)
	for i := 7; i < len(rtspURL); i++ {
		if rtspURL[i] == '@' {
			// Credentials already exist
			return rtspURL
		}
		if rtspURL[i] == '/' {
			// Reached path without finding @, no credentials
			break
		}
	}

	// Insert credentials after rtsp://
	return fmt.Sprintf("rtsp://%s:%s@%s", username, password, rtspURL[7:])
}

// DeleteCamera deletes a camera and its associated streams from the database
func (r *CameraRepository) DeleteCamera(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete associated camera streams first (due to foreign key constraint)
	_, err = tx.ExecContext(ctx, "DELETE FROM camera_streams WHERE camera_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete camera streams: %w", err)
	}

	// Delete the camera
	result, err := tx.ExecContext(ctx, "DELETE FROM cameras WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete camera: %w", err)
	}

	// Check if camera was found and deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("camera not found")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
