package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rta/cctv/storage-service/internal/domain"
	"github.com/rta/cctv/storage-service/internal/repository"
	"github.com/rs/zerolog"
)

// Handler handles HTTP requests for storage service
type Handler struct {
	storageRepo repository.StorageRepository
	segmentRepo repository.SegmentRepository
	exportRepo  repository.ExportRepository
	logger      zerolog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(
	storageRepo repository.StorageRepository,
	segmentRepo repository.SegmentRepository,
	exportRepo repository.ExportRepository,
	logger zerolog.Logger,
) *Handler {
	return &Handler{
		storageRepo: storageRepo,
		segmentRepo: segmentRepo,
		exportRepo:  exportRepo,
		logger:      logger,
	}
}

// StoreSegment handles storing a video segment
func (h *Handler) StoreSegment(w http.ResponseWriter, r *http.Request) {
	var req domain.StoreSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Open the file
	file, err := os.Open(req.FilePath)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Failed to open file", err)
		return
	}
	defer file.Close()

	// Generate storage path
	storagePath := h.generateStoragePath(req.CameraID, req.StartTime)

	// Upload to storage
	err = h.storageRepo.Store(r.Context(), storagePath, file, req.SizeBytes)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to upload segment", err)
		return
	}

	// Create segment metadata
	segment := &domain.Segment{
		ID:              uuid.New().String(),
		CameraID:        req.CameraID,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		DurationSeconds: req.DurationSeconds,
		SizeBytes:       req.SizeBytes,
		StorageBackend:  domain.BackendMinIO,
		StoragePath:     storagePath,
		ThumbnailPath:   req.ThumbnailPath,
		CreatedAt:       time.Now(),
	}

	err = h.segmentRepo.Create(r.Context(), segment)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to store segment metadata", err)
		return
	}

	h.respondJSON(w, http.StatusCreated, segment)
}

// ListSegments handles listing video segments
func (h *Handler) ListSegments(w http.ResponseWriter, r *http.Request) {
	cameraID := chi.URLParam(r, "camera_id")

	startTimeStr := r.URL.Query().Get("start")
	endTimeStr := r.URL.Query().Get("end")

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid start time", err)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid end time", err)
		return
	}

	query := domain.ListSegmentsQuery{
		CameraID:  cameraID,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     100,
	}

	segments, err := h.segmentRepo.List(r.Context(), query)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list segments", err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"segments": segments,
		"count":    len(segments),
	})
}

// CreateExport handles creating a video export
func (h *Handler) CreateExport(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create export record
	export := &domain.Export{
		ID:        uuid.New().String(),
		CameraIDs: req.CameraIDs,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Format:    req.Format,
		Reason:    req.Reason,
		Status:    domain.ExportPending,
		CreatedBy: req.CreatedBy,
		CreatedAt: time.Now(),
	}

	err := h.exportRepo.Create(r.Context(), export)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create export", err)
		return
	}

	// Start export processing in background
	go h.processExport(export)

	h.respondJSON(w, http.StatusAccepted, export)
}

// GetExport handles getting export status
func (h *Handler) GetExport(w http.ResponseWriter, r *http.Request) {
	exportID := chi.URLParam(r, "export_id")

	export, err := h.exportRepo.Get(r.Context(), exportID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Export not found", err)
		return
	}

	h.respondJSON(w, http.StatusOK, export)
}

// DownloadExport handles downloading an export
func (h *Handler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	exportID := chi.URLParam(r, "export_id")

	export, err := h.exportRepo.Get(r.Context(), exportID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Export not found", err)
		return
	}

	if export.Status != domain.ExportCompleted {
		h.respondError(w, http.StatusBadRequest, "Export not completed", fmt.Errorf("status: %s", export.Status))
		return
	}

	// Redirect to download URL
	http.Redirect(w, r, export.DownloadURL, http.StatusTemporaryRedirect)
}

// Health handles health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "storage-service",
		"time":    time.Now(),
	})
}

// Helper functions

func (h *Handler) generateStoragePath(cameraID string, startTime time.Time) string {
	return filepath.Join(
		cameraID,
		startTime.Format("2006"),
		startTime.Format("01"),
		startTime.Format("02"),
		startTime.Format("15-04-05")+".ts",
	)
}

func (h *Handler) processExport(export *domain.Export) {
	ctx := context.Background()

	h.logger.Info().
		Str("export_id", export.ID).
		Msg("Starting export processing")

	// Update status to processing
	h.exportRepo.UpdateStatus(ctx, export.ID, domain.ExportProcessing, "", 0, "")

	// TODO: Implement actual export logic with FFmpeg
	// For now, simulate processing
	time.Sleep(5 * time.Second)

	// Generate download URL
	exportPath := fmt.Sprintf("exports/%s.%s", export.ID, export.Format)
	downloadURL, err := h.storageRepo.GenerateDownloadURL(ctx, exportPath, 7*24*3600) // 7 days
	if err != nil {
		h.logger.Error().Err(err).Str("export_id", export.ID).Msg("Failed to generate download URL")
		h.exportRepo.UpdateError(ctx, export.ID, err.Error())
		return
	}

	// Update export as completed
	err = h.exportRepo.UpdateStatus(ctx, export.ID, domain.ExportCompleted, exportPath, 0, downloadURL)
	if err != nil {
		h.logger.Error().Err(err).Str("export_id", export.ID).Msg("Failed to update export status")
		return
	}

	h.logger.Info().
		Str("export_id", export.ID).
		Msg("Export processing completed")
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error().Err(err).Str("message", message).Msg("Request error")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"details": err.Error(),
		},
	})
}
