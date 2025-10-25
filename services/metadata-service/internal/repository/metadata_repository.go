package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rta/cctv/metadata-service/internal/domain"
	"github.com/rs/zerolog"
)

// MetadataRepository handles all metadata operations
type MetadataRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewMetadataRepository creates a new metadata repository
func NewMetadataRepository(db *sql.DB, logger zerolog.Logger) *MetadataRepository {
	return &MetadataRepository{
		db:     db,
		logger: logger,
	}
}

// Tags

func (r *MetadataRepository) CreateTag(ctx context.Context, tag *domain.Tag) error {
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}

	query := `INSERT INTO tags (id, name, category, color) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, tag.ID, tag.Name, tag.Category, tag.Color)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	r.logger.Info().Str("tag_id", tag.ID).Str("name", tag.Name).Msg("Tag created")
	return nil
}

func (r *MetadataRepository) GetTags(ctx context.Context) ([]*domain.Tag, error) {
	query := `SELECT id, name, category, color, created_at FROM tags ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		tag := &domain.Tag{}
		var category, color sql.NullString
		err := rows.Scan(&tag.ID, &tag.Name, &category, &color, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		if category.Valid {
			tag.Category = category.String
		}
		if color.Valid {
			tag.Color = color.String
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *MetadataRepository) TagSegment(ctx context.Context, segmentID, tagID, userID string) error {
	query := `INSERT INTO video_tags (segment_id, tag_id, user_id) VALUES ($1, $2, $3)
	          ON CONFLICT (segment_id, tag_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, segmentID, tagID, userID)
	if err != nil {
		return fmt.Errorf("failed to tag segment: %w", err)
	}

	r.logger.Info().Str("segment_id", segmentID).Str("tag_id", tagID).Msg("Segment tagged")
	return nil
}

func (r *MetadataRepository) GetSegmentTags(ctx context.Context, segmentID string) ([]*domain.VideoTag, error) {
	query := `
		SELECT vt.segment_id, vt.tag_id, t.name, vt.user_id, vt.created_at
		FROM video_tags vt
		JOIN tags t ON vt.tag_id = t.id
		WHERE vt.segment_id = $1
		ORDER BY vt.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, segmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.VideoTag
	for rows.Next() {
		tag := &domain.VideoTag{}
		err := rows.Scan(&tag.SegmentID, &tag.TagID, &tag.TagName, &tag.UserID, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// Annotations

func (r *MetadataRepository) CreateAnnotation(ctx context.Context, annotation *domain.Annotation) error {
	if annotation.ID == "" {
		annotation.ID = uuid.New().String()
	}

	query := `
		INSERT INTO annotations (id, segment_id, timestamp_offset, type, content, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		annotation.ID,
		annotation.SegmentID,
		annotation.TimestampOffset,
		annotation.Type,
		annotation.Content,
		annotation.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to create annotation: %w", err)
	}

	r.logger.Info().Str("annotation_id", annotation.ID).Msg("Annotation created")
	return nil
}

func (r *MetadataRepository) GetSegmentAnnotations(ctx context.Context, segmentID string) ([]*domain.Annotation, error) {
	query := `
		SELECT id, segment_id, timestamp_offset, type, content, user_id, created_at
		FROM annotations
		WHERE segment_id = $1
		ORDER BY timestamp_offset ASC
	`

	rows, err := r.db.QueryContext(ctx, query, segmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotations: %w", err)
	}
	defer rows.Close()

	var annotations []*domain.Annotation
	for rows.Next() {
		annotation := &domain.Annotation{}
		err := rows.Scan(
			&annotation.ID,
			&annotation.SegmentID,
			&annotation.TimestampOffset,
			&annotation.Type,
			&annotation.Content,
			&annotation.UserID,
			&annotation.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		annotations = append(annotations, annotation)
	}

	return annotations, nil
}

// Incidents

func (r *MetadataRepository) CreateIncident(ctx context.Context, incident *domain.Incident) error {
	if incident.ID == "" {
		incident.ID = uuid.New().String()
	}

	query := `
		INSERT INTO incidents (
			id, title, description, severity, status, camera_ids,
			start_time, end_time, tags, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		incident.ID,
		incident.Title,
		incident.Description,
		incident.Severity,
		domain.IncidentOpen,
		pq.Array(incident.CameraIDs),
		incident.StartTime,
		incident.EndTime,
		pq.Array(incident.Tags),
		incident.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	r.logger.Info().Str("incident_id", incident.ID).Str("title", incident.Title).Msg("Incident created")
	return nil
}

func (r *MetadataRepository) GetIncident(ctx context.Context, id string) (*domain.Incident, error) {
	query := `
		SELECT id, title, description, severity, status, camera_ids,
		       start_time, end_time, tags, assigned_to, created_by,
		       created_at, updated_at, closed_at
		FROM incidents
		WHERE id = $1
	`

	incident := &domain.Incident{}
	var assignedTo sql.NullString
	var closedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&incident.ID,
		&incident.Title,
		&incident.Description,
		&incident.Severity,
		&incident.Status,
		pq.Array(&incident.CameraIDs),
		&incident.StartTime,
		&incident.EndTime,
		pq.Array(&incident.Tags),
		&assignedTo,
		&incident.CreatedBy,
		&incident.CreatedAt,
		&incident.UpdatedAt,
		&closedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("incident not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	if assignedTo.Valid {
		incident.AssignedTo = assignedTo.String
	}
	if closedAt.Valid {
		incident.ClosedAt = &closedAt.Time
	}

	return incident, nil
}

func (r *MetadataRepository) UpdateIncident(ctx context.Context, id string, req domain.UpdateIncidentRequest) error {
	updates := []string{}
	args := []interface{}{}
	argNum := 1

	if req.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *req.Status)
		argNum++

		if *req.Status == domain.IncidentClosed {
			updates = append(updates, fmt.Sprintf("closed_at = $%d", argNum))
			args = append(args, time.Now())
			argNum++
		}
	}

	if req.AssignedTo != nil {
		updates = append(updates, fmt.Sprintf("assigned_to = $%d", argNum))
		args = append(args, *req.AssignedTo)
		argNum++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argNum))
	args = append(args, time.Now())
	argNum++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE incidents SET %s WHERE id = $%d", strings.Join(updates, ", "), argNum)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("incident not found: %s", id)
	}

	r.logger.Info().Str("incident_id", id).Msg("Incident updated")
	return nil
}

func (r *MetadataRepository) SearchIncidents(ctx context.Context, query domain.SearchQuery) ([]*domain.Incident, error) {
	sqlQuery := `
		SELECT id, title, description, severity, status, camera_ids,
		       start_time, end_time, tags, assigned_to, created_by,
		       created_at, updated_at, closed_at
		FROM incidents
		WHERE 1=1
	`

	args := []interface{}{}
	argNum := 1

	// Full-text search
	if query.Query != "" {
		sqlQuery += fmt.Sprintf(" AND search_vector @@ plainto_tsquery('english', $%d)", argNum)
		args = append(args, query.Query)
		argNum++
	}

	// Camera IDs filter
	if len(query.CameraIDs) > 0 {
		sqlQuery += fmt.Sprintf(" AND camera_ids && $%d", argNum)
		args = append(args, pq.Array(query.CameraIDs))
		argNum++
	}

	// Severity filter
	if query.Severity != "" {
		sqlQuery += fmt.Sprintf(" AND severity = $%d", argNum)
		args = append(args, query.Severity)
		argNum++
	}

	// Status filter
	if query.Status != "" {
		sqlQuery += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, query.Status)
		argNum++
	}

	// Time range filter
	if query.StartTime != nil {
		sqlQuery += fmt.Sprintf(" AND start_time >= $%d", argNum)
		args = append(args, *query.StartTime)
		argNum++
	}
	if query.EndTime != nil {
		sqlQuery += fmt.Sprintf(" AND end_time <= $%d", argNum)
		args = append(args, *query.EndTime)
		argNum++
	}

	sqlQuery += " ORDER BY created_at DESC"

	// Pagination
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}
	sqlQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, query.Offset)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		incident := &domain.Incident{}
		var assignedTo sql.NullString
		var closedAt sql.NullTime

		err := rows.Scan(
			&incident.ID,
			&incident.Title,
			&incident.Description,
			&incident.Severity,
			&incident.Status,
			pq.Array(&incident.CameraIDs),
			&incident.StartTime,
			&incident.EndTime,
			pq.Array(&incident.Tags),
			&assignedTo,
			&incident.CreatedBy,
			&incident.CreatedAt,
			&incident.UpdatedAt,
			&closedAt,
		)
		if err != nil {
			return nil, err
		}

		if assignedTo.Valid {
			incident.AssignedTo = assignedTo.String
		}
		if closedAt.Valid {
			incident.ClosedAt = &closedAt.Time
		}

		incidents = append(incidents, incident)
	}

	return incidents, nil
}
