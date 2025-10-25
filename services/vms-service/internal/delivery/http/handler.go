package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/vms-service/internal/domain"
	"github.com/rta/cctv/vms-service/internal/repository"
	"github.com/rs/zerolog"
)

// Handler handles HTTP requests for VMS Service
type Handler struct {
	cameraRepo repository.CameraRepository
	cache      repository.CacheRepository
	logger     zerolog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(cameraRepo repository.CameraRepository, cache repository.CacheRepository, logger zerolog.Logger) *Handler {
	return &Handler{
		cameraRepo: cameraRepo,
		cache:      cache,
		logger:     logger,
	}
}

// GetCameras retrieves all cameras
// GET /vms/cameras
func (h *Handler) GetCameras(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check cache first
	cacheKey := "cameras:all"
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
		h.logger.Debug().Msg("Serving cameras from cache")
		respondJSON(w, http.StatusOK, cached)
		return
	}

	// Fetch from Milestone VMS
	cameras, err := h.cameraRepo.GetAll(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get cameras")
		respondError(w, http.StatusInternalServerError, "Failed to retrieve cameras")
		return
	}

	// Cache for 5 minutes
	response := map[string]interface{}{
		"cameras":      cameras,
		"total":        len(cameras),
		"last_updated": time.Now(),
	}
	h.cache.Set(ctx, cacheKey, response, 5*time.Minute)

	respondJSON(w, http.StatusOK, response)
}

// GetCameraByID retrieves a single camera by ID
// GET /vms/cameras/{id}
func (h *Handler) GetCameraByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "id")

	if cameraID == "" {
		respondError(w, http.StatusBadRequest, "Camera ID is required")
		return
	}

	// Check cache
	cacheKey := "camera:" + cameraID
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
		h.logger.Debug().Str("camera_id", cameraID).Msg("Serving camera from cache")
		respondJSON(w, http.StatusOK, cached)
		return
	}

	// Fetch from Milestone VMS
	camera, err := h.cameraRepo.GetByID(ctx, cameraID)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Failed to get camera")
		respondError(w, http.StatusNotFound, "Camera not found")
		return
	}

	// Cache for 5 minutes
	h.cache.Set(ctx, cacheKey, camera, 5*time.Minute)

	respondJSON(w, http.StatusOK, camera)
}

// GetCamerasBySource retrieves cameras filtered by source
// GET /vms/cameras?source={source}
func (h *Handler) GetCamerasBySource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceParam := r.URL.Query().Get("source")

	if sourceParam == "" {
		// If no source filter, return all cameras
		h.GetCameras(w, r)
		return
	}

	source := domain.CameraSource(sourceParam)

	// Validate source
	validSources := map[domain.CameraSource]bool{
		domain.SourceDubaiPolice: true,
		domain.SourceMetro:       true,
		domain.SourceBus:         true,
		domain.SourceOther:       true,
	}

	if !validSources[source] {
		respondError(w, http.StatusBadRequest, "Invalid source. Must be one of: DUBAI_POLICE, METRO, BUS, OTHER")
		return
	}

	// Check cache
	cacheKey := "cameras:source:" + string(source)
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
		h.logger.Debug().Str("source", string(source)).Msg("Serving cameras from cache")
		respondJSON(w, http.StatusOK, cached)
		return
	}

	// Fetch from Milestone VMS
	cameras, err := h.cameraRepo.GetBySource(ctx, source)
	if err != nil {
		h.logger.Error().Err(err).Str("source", string(source)).Msg("Failed to get cameras by source")
		respondError(w, http.StatusInternalServerError, "Failed to retrieve cameras")
		return
	}

	response := map[string]interface{}{
		"cameras": cameras,
		"total":   len(cameras),
		"source":  source,
	}

	// Cache for 5 minutes
	h.cache.Set(ctx, cacheKey, response, 5*time.Minute)

	respondJSON(w, http.StatusOK, response)
}

// GetCameraStream retrieves RTSP URL for streaming
// GET /vms/cameras/{id}/stream
func (h *Handler) GetCameraStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "id")

	if cameraID == "" {
		respondError(w, http.StatusBadRequest, "Camera ID is required")
		return
	}

	// Get RTSP URL from Milestone
	rtspURL, err := h.cameraRepo.GetRTSPURL(ctx, cameraID)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Failed to get RTSP URL")
		respondError(w, http.StatusNotFound, "Camera not found or stream unavailable")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"camera_id": cameraID,
		"rtsp_url":  rtspURL,
		"transport": "tcp",
	})
}

// ExecutePTZ handles PTZ control commands
// POST /vms/cameras/{id}/ptz
func (h *Handler) ExecutePTZ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "id")

	if cameraID == "" {
		respondError(w, http.StatusBadRequest, "Camera ID is required")
		return
	}

	var cmd domain.PTZCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd.CameraID = cameraID

	// Execute PTZ command
	if err := h.cameraRepo.ExecutePTZCommand(ctx, &cmd); err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Failed to execute PTZ command")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Str("action", string(cmd.Action)).
		Msg("PTZ command executed")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"camera_id": cameraID,
		"action":    cmd.Action,
		"status":    "success",
	})
}

// GetRecordingSegments retrieves available recording segments
// GET /vms/recordings/{camera_id}/segments?start={start}&end={end}
func (h *Handler) GetRecordingSegments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "camera_id")

	if cameraID == "" {
		respondError(w, http.StatusBadRequest, "Camera ID is required")
		return
	}

	// Parse time parameters
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		respondError(w, http.StatusBadRequest, "Both start and end time parameters are required")
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid start time format (use RFC3339)")
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid end time format (use RFC3339)")
		return
	}

	// Get recording segments from Milestone
	segments, err := h.cameraRepo.GetRecordingSegments(ctx, cameraID, start, end)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Failed to get recording segments")
		respondError(w, http.StatusInternalServerError, "Failed to retrieve recording segments")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"camera_id": cameraID,
		"start":     start,
		"end":       end,
		"segments":  segments,
		"total":     len(segments),
	})
}

// ExportRecording creates an export job for recording
// POST /vms/recordings/export
func (h *Handler) ExportRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.RecordingExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.CameraID == "" {
		respondError(w, http.StatusBadRequest, "Camera ID is required")
		return
	}

	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		respondError(w, http.StatusBadRequest, "Start and end times are required")
		return
	}

	if req.EndTime.Before(req.StartTime) {
		respondError(w, http.StatusBadRequest, "End time must be after start time")
		return
	}

	// Create export job
	export, err := h.cameraRepo.ExportRecording(ctx, &req)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", req.CameraID).Msg("Failed to create export job")
		respondError(w, http.StatusInternalServerError, "Failed to create export job")
		return
	}

	h.logger.Info().
		Str("export_id", export.ID).
		Str("camera_id", req.CameraID).
		Msg("Export job created")

	respondJSON(w, http.StatusCreated, export)
}

// GetExportStatus retrieves status of an export job
// GET /vms/recordings/export/{export_id}
func (h *Handler) GetExportStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exportID := chi.URLParam(r, "export_id")

	if exportID == "" {
		respondError(w, http.StatusBadRequest, "Export ID is required")
		return
	}

	export, err := h.cameraRepo.GetExportStatus(ctx, exportID)
	if err != nil {
		h.logger.Error().Err(err).Str("export_id", exportID).Msg("Failed to get export status")
		respondError(w, http.StatusNotFound, "Export job not found")
		return
	}

	respondJSON(w, http.StatusOK, export)
}

// Health check endpoint
// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check Milestone connection
	if err := h.cameraRepo.Sync(ctx); err != nil {
		respondError(w, http.StatusServiceUnavailable, "Milestone VMS connection failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "vms-service",
		"timestamp": time.Now(),
	})
}

// Utility functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"code":    status,
		},
	})
}
