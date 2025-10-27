package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rs/zerolog"
)

// LayoutRepository implements domain.LayoutRepository using PostgreSQL
type LayoutRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewLayoutRepository creates a new PostgreSQL layout repository
func NewLayoutRepository(db *sql.DB, logger zerolog.Logger) *LayoutRepository {
	return &LayoutRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new layout preference with camera assignments
func (r *LayoutRepository) Create(layout *domain.LayoutPreference) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate UUID for layout if not provided
	if layout.ID == "" {
		layout.ID = uuid.New().String()
	}

	// Insert layout preference
	query := `
		INSERT INTO layout_preferences (
			id, name, description, layout_type, grid_layout, scope, created_by, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	_, err = tx.Exec(query,
		layout.ID,
		layout.Name,
		layout.Description,
		layout.LayoutType,
		layout.GridLayout,
		layout.Scope,
		layout.CreatedBy,
		layout.IsActive,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert layout preference: %w", err)
	}

	// Insert camera assignments
	cameraQuery := `
		INSERT INTO layout_camera_assignments (
			id, layout_id, camera_id, position_index, cell_size, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, camera := range layout.Cameras {
		cameraID := uuid.New().String()
		_, err = tx.Exec(cameraQuery,
			cameraID,
			layout.ID,
			camera.CameraID,
			camera.PositionIndex,
			camera.CellSize,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert camera assignment: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	layout.CreatedAt = now
	layout.UpdatedAt = now

	r.logger.Info().
		Str("layout_id", layout.ID).
		Str("name", layout.Name).
		Str("created_by", layout.CreatedBy).
		Msg("Layout created successfully")

	return nil
}

// GetByID retrieves a layout by ID with camera assignments
func (r *LayoutRepository) GetByID(id string) (*domain.LayoutPreference, error) {
	// Get layout preference
	query := `
		SELECT id, name, description, layout_type, grid_layout, scope, created_by, is_active, created_at, updated_at
		FROM layout_preferences
		WHERE id = $1 AND is_active = true
	`

	layout := &domain.LayoutPreference{}
	err := r.db.QueryRow(query, id).Scan(
		&layout.ID,
		&layout.Name,
		&layout.Description,
		&layout.LayoutType,
		&layout.GridLayout,
		&layout.Scope,
		&layout.CreatedBy,
		&layout.IsActive,
		&layout.CreatedAt,
		&layout.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("layout not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get layout: %w", err)
	}

	// Get camera assignments
	cameraQuery := `
		SELECT id, camera_id, position_index, cell_size, created_at
		FROM layout_camera_assignments
		WHERE layout_id = $1
		ORDER BY position_index ASC
	`

	rows, err := r.db.Query(cameraQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get camera assignments: %w", err)
	}
	defer rows.Close()

	cameras := []domain.LayoutCameraAssignment{}
	for rows.Next() {
		var camera domain.LayoutCameraAssignment
		var cellSize sql.NullString

		err := rows.Scan(
			&camera.ID,
			&camera.CameraID,
			&camera.PositionIndex,
			&cellSize,
			&camera.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan camera assignment: %w", err)
		}

		camera.LayoutID = id
		if cellSize.Valid {
			camera.CellSize = cellSize.String
		}

		cameras = append(cameras, camera)
	}

	layout.Cameras = cameras
	return layout, nil
}

// List retrieves layouts with optional filtering
func (r *LayoutRepository) List(layoutType *domain.LayoutType, scope *domain.LayoutScope, createdBy *string) ([]*domain.LayoutPreferenceSummary, error) {
	query := `
		SELECT
			lp.id,
			lp.name,
			lp.description,
			lp.layout_type,
			lp.grid_layout,
			lp.scope,
			lp.created_by,
			lp.created_at,
			lp.updated_at,
			COUNT(lca.id) as camera_count
		FROM layout_preferences lp
		LEFT JOIN layout_camera_assignments lca ON lp.id = lca.layout_id
		WHERE lp.is_active = true
	`

	args := []interface{}{}
	argCount := 1

	if layoutType != nil {
		query += fmt.Sprintf(" AND lp.layout_type = $%d", argCount)
		args = append(args, *layoutType)
		argCount++
	}

	if scope != nil {
		query += fmt.Sprintf(" AND lp.scope = $%d", argCount)
		args = append(args, *scope)
		argCount++
	}

	if createdBy != nil {
		query += fmt.Sprintf(" AND lp.created_by = $%d", argCount)
		args = append(args, *createdBy)
		argCount++
	}

	query += " GROUP BY lp.id, lp.name, lp.description, lp.layout_type, lp.scope, lp.created_by, lp.created_at, lp.updated_at"
	query += " ORDER BY lp.updated_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list layouts: %w", err)
	}
	defer rows.Close()

	layouts := []*domain.LayoutPreferenceSummary{}
	for rows.Next() {
		var layout domain.LayoutPreferenceSummary
		var description sql.NullString

		err := rows.Scan(
			&layout.ID,
			&layout.Name,
			&description,
			&layout.LayoutType,
		&layout.GridLayout,
			&layout.Scope,
			&layout.CreatedBy,
			&layout.CreatedAt,
			&layout.UpdatedAt,
			&layout.CameraCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan layout: %w", err)
		}

		if description.Valid {
			layout.Description = description.String
		}

		layouts = append(layouts, &layout)
	}

	return layouts, nil
}

// Update updates an existing layout
func (r *LayoutRepository) Update(id string, request *domain.UpdateLayoutRequest) (*domain.LayoutPreference, error) {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update layout preference
	query := `
		UPDATE layout_preferences
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND is_active = true
		RETURNING created_by, layout_type, scope, is_active, created_at
	`

	now := time.Now()
	layout := &domain.LayoutPreference{
		ID:          id,
		Name:        request.Name,
		Description: request.Description,
		UpdatedAt:   now,
	}

	err = tx.QueryRow(query, request.Name, request.Description, now, id).Scan(
		&layout.CreatedBy,
		&layout.LayoutType,
		&layout.GridLayout,
		&layout.Scope,
		&layout.IsActive,
		&layout.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("layout not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update layout: %w", err)
	}

	// Delete existing camera assignments
	_, err = tx.Exec("DELETE FROM layout_camera_assignments WHERE layout_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete camera assignments: %w", err)
	}

	// Insert new camera assignments
	cameraQuery := `
		INSERT INTO layout_camera_assignments (
			id, layout_id, camera_id, position_index, cell_size, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, camera := range request.Cameras {
		cameraID := uuid.New().String()
		_, err = tx.Exec(cameraQuery,
			cameraID,
			id,
			camera.CameraID,
			camera.PositionIndex,
			camera.CellSize,
			now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert camera assignment: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Set cameras in response
	layout.Cameras = request.Cameras

	r.logger.Info().
		Str("layout_id", id).
		Str("name", request.Name).
		Msg("Layout updated successfully")

	return layout, nil
}

// Delete deletes a layout by ID (soft delete)
func (r *LayoutRepository) Delete(id string) error {
	query := `
		UPDATE layout_preferences
		SET is_active = false, updated_at = $1
		WHERE id = $2 AND is_active = true
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete layout: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("layout not found: %s", id)
	}

	r.logger.Info().
		Str("layout_id", id).
		Msg("Layout deleted successfully")

	return nil
}
