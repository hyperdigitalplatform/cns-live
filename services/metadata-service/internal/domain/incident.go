package domain

import "time"

// IncidentSeverity represents the severity level of an incident
type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "LOW"
	SeverityMedium   IncidentSeverity = "MEDIUM"
	SeverityHigh     IncidentSeverity = "HIGH"
	SeverityCritical IncidentSeverity = "CRITICAL"
)

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	IncidentOpen       IncidentStatus = "OPEN"
	IncidentInProgress IncidentStatus = "IN_PROGRESS"
	IncidentResolved   IncidentStatus = "RESOLVED"
	IncidentClosed     IncidentStatus = "CLOSED"
)

// Incident represents a tracked incident
type Incident struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Severity    IncidentSeverity `json:"severity"`
	Status      IncidentStatus   `json:"status"`
	CameraIDs   []string         `json:"camera_ids"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
	Tags        []string         `json:"tags,omitempty"`
	AssignedTo  string           `json:"assigned_to,omitempty"`
	CreatedBy   string           `json:"created_by"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	ClosedAt    *time.Time       `json:"closed_at,omitempty"`
}

// CreateIncidentRequest represents a request to create an incident
type CreateIncidentRequest struct {
	Title       string           `json:"title" validate:"required,min=1,max=255"`
	Description string           `json:"description"`
	Severity    IncidentSeverity `json:"severity" validate:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
	CameraIDs   []string         `json:"camera_ids" validate:"required,min=1"`
	StartTime   time.Time        `json:"start_time" validate:"required"`
	EndTime     time.Time        `json:"end_time" validate:"required"`
	Tags        []string         `json:"tags,omitempty"`
	CreatedBy   string           `json:"created_by" validate:"required"`
}

// UpdateIncidentRequest represents a request to update an incident
type UpdateIncidentRequest struct {
	Status     *IncidentStatus `json:"status,omitempty"`
	AssignedTo *string         `json:"assigned_to,omitempty"`
}

// SearchQuery represents a search query
type SearchQuery struct {
	Query      string           `json:"query,omitempty"`
	CameraIDs  []string         `json:"camera_ids,omitempty"`
	Severity   IncidentSeverity `json:"severity,omitempty"`
	Status     IncidentStatus   `json:"status,omitempty"`
	StartTime  *time.Time       `json:"start_time,omitempty"`
	EndTime    *time.Time       `json:"end_time,omitempty"`
	Limit      int              `json:"limit,omitempty"`
	Offset     int              `json:"offset,omitempty"`
}
