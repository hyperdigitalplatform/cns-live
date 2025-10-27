package usecase

import (
	"fmt"

	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rs/zerolog"
)

// LayoutUseCase implements domain.LayoutUseCase
type LayoutUseCase struct {
	layoutRepo domain.LayoutRepository
	logger     zerolog.Logger
}

// NewLayoutUseCase creates a new layout use case
func NewLayoutUseCase(layoutRepo domain.LayoutRepository, logger zerolog.Logger) *LayoutUseCase {
	return &LayoutUseCase{
		layoutRepo: layoutRepo,
		logger:     logger,
	}
}

// CreateLayout creates a new layout preference
func (uc *LayoutUseCase) CreateLayout(request *domain.CreateLayoutRequest) (*domain.LayoutPreference, error) {
	// Validate request
	if err := uc.validateCreateRequest(request); err != nil {
		return nil, err
	}

	// Create layout entity
	layout := &domain.LayoutPreference{
		Name:        request.Name,
		Description: request.Description,
		LayoutType:  request.LayoutType,
		GridLayout:  request.GridLayout,
		Scope:       request.Scope,
		CreatedBy:   request.CreatedBy,
		IsActive:    true,
		Cameras:     request.Cameras,
	}

	// Create in repository
	if err := uc.layoutRepo.Create(layout); err != nil {
		uc.logger.Error().
			Err(err).
			Str("name", request.Name).
			Msg("Failed to create layout")
		return nil, fmt.Errorf("failed to create layout: %w", err)
	}

	uc.logger.Info().
		Str("layout_id", layout.ID).
		Str("name", layout.Name).
		Str("created_by", layout.CreatedBy).
		Msg("Layout created successfully")

	return layout, nil
}

// GetLayout retrieves a layout by ID
func (uc *LayoutUseCase) GetLayout(id string) (*domain.LayoutPreference, error) {
	if id == "" {
		return nil, fmt.Errorf("layout ID is required")
	}

	layout, err := uc.layoutRepo.GetByID(id)
	if err != nil {
		uc.logger.Error().
			Err(err).
			Str("layout_id", id).
			Msg("Failed to get layout")
		return nil, fmt.Errorf("failed to get layout: %w", err)
	}

	return layout, nil
}

// ListLayouts retrieves all layouts with optional filtering
func (uc *LayoutUseCase) ListLayouts(layoutType *domain.LayoutType, scope *domain.LayoutScope, createdBy *string) (*domain.LayoutListResponse, error) {
	layouts, err := uc.layoutRepo.List(layoutType, scope, createdBy)
	if err != nil {
		uc.logger.Error().
			Err(err).
			Msg("Failed to list layouts")
		return nil, fmt.Errorf("failed to list layouts: %w", err)
	}

	response := &domain.LayoutListResponse{
		Layouts: []domain.LayoutPreferenceSummary{},
		Total:   len(layouts),
	}

	// Convert pointers to values
	for _, layout := range layouts {
		response.Layouts = append(response.Layouts, *layout)
	}

	return response, nil
}

// UpdateLayout updates an existing layout
func (uc *LayoutUseCase) UpdateLayout(id string, request *domain.UpdateLayoutRequest) (*domain.LayoutPreference, error) {
	if id == "" {
		return nil, fmt.Errorf("layout ID is required")
	}

	// Validate request
	if err := uc.validateUpdateRequest(request); err != nil {
		return nil, err
	}

	// Update in repository
	layout, err := uc.layoutRepo.Update(id, request)
	if err != nil {
		uc.logger.Error().
			Err(err).
			Str("layout_id", id).
			Msg("Failed to update layout")
		return nil, fmt.Errorf("failed to update layout: %w", err)
	}

	uc.logger.Info().
		Str("layout_id", id).
		Str("name", request.Name).
		Msg("Layout updated successfully")

	return layout, nil
}

// DeleteLayout deletes a layout by ID
func (uc *LayoutUseCase) DeleteLayout(id string) error {
	if id == "" {
		return fmt.Errorf("layout ID is required")
	}

	if err := uc.layoutRepo.Delete(id); err != nil {
		uc.logger.Error().
			Err(err).
			Str("layout_id", id).
			Msg("Failed to delete layout")
		return fmt.Errorf("failed to delete layout: %w", err)
	}

	uc.logger.Info().
		Str("layout_id", id).
		Msg("Layout deleted successfully")

	return nil
}

// validateCreateRequest validates the create layout request
func (uc *LayoutUseCase) validateCreateRequest(request *domain.CreateLayoutRequest) error {
	if request.Name == "" {
		return fmt.Errorf("layout name is required")
	}

	if request.LayoutType != domain.LayoutTypeStandard && request.LayoutType != domain.LayoutTypeHotspot {
		return fmt.Errorf("invalid layout type: %s", request.LayoutType)
	}

	if request.Scope != domain.LayoutScopeGlobal && request.Scope != domain.LayoutScopeLocal {
		return fmt.Errorf("invalid scope: %s", request.Scope)
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	if len(request.Cameras) == 0 {
		return fmt.Errorf("at least one camera is required")
	}

	// Validate camera positions are unique and sequential
	positions := make(map[int]bool)
	for _, camera := range request.Cameras {
		if camera.CameraID == "" {
			return fmt.Errorf("camera ID is required")
		}

		if positions[camera.PositionIndex] {
			return fmt.Errorf("duplicate position index: %d", camera.PositionIndex)
		}

		positions[camera.PositionIndex] = true
	}

	return nil
}

// validateUpdateRequest validates the update layout request
func (uc *LayoutUseCase) validateUpdateRequest(request *domain.UpdateLayoutRequest) error {
	if request.Name == "" {
		return fmt.Errorf("layout name is required")
	}

	if len(request.Cameras) == 0 {
		return fmt.Errorf("at least one camera is required")
	}

	// Validate camera positions are unique
	positions := make(map[int]bool)
	for _, camera := range request.Cameras {
		if camera.CameraID == "" {
			return fmt.Errorf("camera ID is required")
		}

		if positions[camera.PositionIndex] {
			return fmt.Errorf("duplicate position index: %d", camera.PositionIndex)
		}

		positions[camera.PositionIndex] = true
	}

	return nil
}
