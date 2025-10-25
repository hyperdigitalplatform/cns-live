package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/recording-service/internal/domain"
	"github.com/rta/cctv/recording-service/internal/manager"
	"github.com/rs/zerolog"
)

// Handler handles HTTP requests for recording service
type Handler struct {
	manager *manager.RecordingManager
	logger  zerolog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(manager *manager.RecordingManager, logger zerolog.Logger) *Handler {
	return &Handler{
		manager: manager,
		logger:  logger,
	}
}

// StartRecording handles starting a recording
func (h *Handler) StartRecording(w http.ResponseWriter, r *http.Request) {
	cameraID := chi.URLParam(r, "camera_id")

	err := h.manager.StartRecording(r.Context(), cameraID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Failed to start recording", err)
		return
	}

	recording, _ := h.manager.GetRecording(cameraID)
	h.respondJSON(w, http.StatusOK, recording)
}

// StopRecording handles stopping a recording
func (h *Handler) StopRecording(w http.ResponseWriter, r *http.Request) {
	cameraID := chi.URLParam(r, "camera_id")

	err := h.manager.StopRecording(cameraID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Failed to stop recording", err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Recording stopped",
		"camera_id": cameraID,
	})
}

// GetRecording handles getting recording status
func (h *Handler) GetRecording(w http.ResponseWriter, r *http.Request) {
	cameraID := chi.URLParam(r, "camera_id")

	recording, err := h.manager.GetRecording(cameraID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Recording not found", err)
		return
	}

	h.respondJSON(w, http.StatusOK, recording)
}

// ListRecordings handles listing all active recordings
func (h *Handler) ListRecordings(w http.ResponseWriter, r *http.Request) {
	recordings := h.manager.GetAllRecordings()

	// Calculate stats
	totalSegments := int64(0)
	totalBytes := int64(0)

	for _, rec := range recordings {
		totalSegments += int64(rec.SegmentCount)
		totalBytes += rec.TotalBytes
	}

	stats := domain.RecordingStats{
		TotalRecordings:          len(recordings),
		ActiveRecordings:         len(recordings),
		TotalSegments:            totalSegments,
		TotalBytes:               totalBytes,
		RecordingsByCamera:       recordings,
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// Health handles health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "recording-service",
		"time":    time.Now(),
	})
}

// Helper functions

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
