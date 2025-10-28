package usecase

import (
	"context"
	"fmt"

	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rs/zerolog"
)

// CameraRepository defines the interface for camera database operations
type CameraRepository interface {
	ImportCamera(ctx context.Context, camera *domain.ImportCameraRequest) error
	GetCamera(ctx context.Context, id string) (*domain.Camera, error)
	ListCameras(ctx context.Context, query domain.CameraQuery) ([]*domain.Camera, error)
	DeleteCamera(ctx context.Context, id string) error
}

// CameraUsecase handles camera business logic
type CameraUsecase struct {
	repo   CameraRepository
	logger zerolog.Logger
}

// NewCameraUsecase creates a new camera usecase
func NewCameraUsecase(repo CameraRepository, logger zerolog.Logger) *CameraUsecase {
	return &CameraUsecase{
		repo:   repo,
		logger: logger,
	}
}

// ImportCameras imports multiple discovered cameras
func (u *CameraUsecase) ImportCameras(ctx context.Context, req domain.ImportCamerasRequest) (*domain.ImportCamerasResponse, error) {
	response := &domain.ImportCamerasResponse{
		Imported: 0,
		Failed:   0,
		Errors:   []string{},
	}

	for _, camera := range req.Cameras {
		// Validate camera data
		if camera.MilestoneID == "" {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Camera %s: missing milestone ID", camera.Name))
			continue
		}

		if camera.Name == "" {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Camera %s: missing name", camera.MilestoneID))
			continue
		}

		// Set default source if not provided
		if camera.Source == "" {
			camera.Source = "OTHER"
		}

		// Set default status if not provided
		if camera.Status == "" {
			camera.Status = "ONLINE"
		}

		// Import camera
		err := u.repo.ImportCamera(ctx, &camera)
		if err != nil {
			u.logger.Error().
				Err(err).
				Str("camera_id", camera.MilestoneID).
				Str("camera_name", camera.Name).
				Msg("Failed to import camera")
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Camera %s: %v", camera.Name, err))
			continue
		}

		u.logger.Info().
			Str("camera_id", camera.MilestoneID).
			Str("camera_name", camera.Name).
			Msg("Successfully imported camera")
		response.Imported++
	}

	return response, nil
}

// GetCamera retrieves a camera by ID
func (u *CameraUsecase) GetCamera(ctx context.Context, id string) (*domain.Camera, error) {
	return u.repo.GetCamera(ctx, id)
}

// ListCameras retrieves cameras with filters
func (u *CameraUsecase) ListCameras(ctx context.Context, query domain.CameraQuery) ([]*domain.Camera, error) {
	return u.repo.ListCameras(ctx, query)
}

// DeleteCamera deletes a camera by ID
func (u *CameraUsecase) DeleteCamera(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("camera ID is required")
	}

	u.logger.Info().Str("camera_id", id).Msg("Deleting camera")

	err := u.repo.DeleteCamera(ctx, id)
	if err != nil {
		u.logger.Error().Err(err).Str("camera_id", id).Msg("Failed to delete camera")
		return fmt.Errorf("failed to delete camera: %w", err)
	}

	u.logger.Info().Str("camera_id", id).Msg("Successfully deleted camera")
	return nil
}
