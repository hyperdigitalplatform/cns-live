package domain

import "time"

// LayoutType represents the type of layout
type LayoutType string

const (
	LayoutTypeStandard LayoutType = "standard"
	LayoutTypeHotspot  LayoutType = "hotspot"
)

// LayoutScope represents the visibility scope of a layout
type LayoutScope string

const (
	LayoutScopeGlobal LayoutScope = "global"
	LayoutScopeLocal  LayoutScope = "local"
)

// LayoutPreference represents a saved camera layout configuration
type LayoutPreference struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	LayoutType  LayoutType              `json:"layout_type"`
	GridLayout  string                  `json:"grid_layout"` // Specific grid: "2x2", "3x3", "9-way-1-hotspot", etc.
	Scope       LayoutScope             `json:"scope"`
	CreatedBy   string                  `json:"created_by"`
	IsActive    bool                    `json:"is_active"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Cameras     []LayoutCameraAssignment `json:"cameras,omitempty"`
}

// LayoutCameraAssignment represents a camera-to-position assignment
type LayoutCameraAssignment struct {
	ID            string    `json:"id,omitempty"`
	LayoutID      string    `json:"layout_id,omitempty"`
	CameraID      string    `json:"camera_id"`
	PositionIndex int       `json:"position_index"`
	CellSize      string    `json:"cell_size,omitempty"` // 'small', 'medium', 'large', 'hotspot'
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

// CreateLayoutRequest represents the request to create a new layout
type CreateLayoutRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	LayoutType  LayoutType              `json:"layout_type" binding:"required,oneof=standard hotspot"`
	GridLayout  string                  `json:"grid_layout" binding:"required"` // e.g., "2x2", "3x3", "9-way-1-hotspot"
	Scope       LayoutScope             `json:"scope" binding:"required,oneof=global local"`
	CreatedBy   string                  `json:"created_by" binding:"required"`
	Cameras     []LayoutCameraAssignment `json:"cameras" binding:"required,min=1"`
}

// UpdateLayoutRequest represents the request to update a layout
type UpdateLayoutRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Cameras     []LayoutCameraAssignment `json:"cameras" binding:"required,min=1"`
}

// LayoutListResponse represents the response for listing layouts
type LayoutListResponse struct {
	Layouts []LayoutPreferenceSummary `json:"layouts"`
	Total   int                       `json:"total"`
}

// LayoutPreferenceSummary represents a summary of a layout (without camera details)
type LayoutPreferenceSummary struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	LayoutType  LayoutType  `json:"layout_type"`
	GridLayout  string      `json:"grid_layout"`
	Scope       LayoutScope `json:"scope"`
	CreatedBy   string      `json:"created_by"`
	CameraCount int         `json:"camera_count"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// LayoutRepository defines the interface for layout data access
type LayoutRepository interface {
	// Create creates a new layout preference
	Create(layout *LayoutPreference) error

	// GetByID retrieves a layout by ID with camera assignments
	GetByID(id string) (*LayoutPreference, error)

	// List retrieves layouts with optional filtering
	List(layoutType *LayoutType, scope *LayoutScope, createdBy *string) ([]*LayoutPreferenceSummary, error)

	// Update updates an existing layout
	Update(id string, request *UpdateLayoutRequest) (*LayoutPreference, error)

	// Delete deletes a layout by ID
	Delete(id string) error
}

// LayoutUseCase defines the interface for layout business logic
type LayoutUseCase interface {
	// CreateLayout creates a new layout preference
	CreateLayout(request *CreateLayoutRequest) (*LayoutPreference, error)

	// GetLayout retrieves a layout by ID
	GetLayout(id string) (*LayoutPreference, error)

	// ListLayouts retrieves all layouts with optional filtering
	ListLayouts(layoutType *LayoutType, scope *LayoutScope, createdBy *string) (*LayoutListResponse, error)

	// UpdateLayout updates an existing layout
	UpdateLayout(id string, request *UpdateLayoutRequest) (*LayoutPreference, error)

	// DeleteLayout deletes a layout by ID
	DeleteLayout(id string) error
}
