package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rta/cctv/playback-service/internal/usecase"
	"github.com/rs/zerolog"
)

// PlaybackHandler handles HTTP requests for playback
type PlaybackHandler struct {
	playbackUseCase *usecase.PlaybackUseCase
	logger          zerolog.Logger
}

func NewPlaybackHandler(playbackUseCase *usecase.PlaybackUseCase, logger zerolog.Logger) *PlaybackHandler {
	return &PlaybackHandler{
		playbackUseCase: playbackUseCase,
		logger:          logger,
	}
}

// RequestPlayback handles playback requests
// POST /api/v1/playback/request
func (h *PlaybackHandler) RequestPlayback(w http.ResponseWriter, r *http.Request) {
	var req domain.PlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set defaults
	if req.Format == "" {
		req.Format = "hls"
	}

	response, err := h.playbackUseCase.RequestPlayback(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to request playback")
		h.respondError(w, http.StatusInternalServerError, "Failed to create playback session", err)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// CreateExport handles export requests
// POST /api/v1/playback/export
func (h *PlaybackHandler) CreateExport(w http.ResponseWriter, r *http.Request) {
	var req domain.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set defaults
	if req.Format == "" {
		req.Format = "mp4"
	}

	response, err := h.playbackUseCase.CreateExport(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create export")
		h.respondError(w, http.StatusInternalServerError, "Failed to create export", err)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// GetCacheStats returns cache statistics
// GET /api/v1/playback/cache/stats
func (h *PlaybackHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.playbackUseCase.GetCacheStats()
	h.respondJSON(w, http.StatusOK, stats)
}

// respondJSON sends a JSON response
func (h *PlaybackHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// respondError sends an error response
func (h *PlaybackHandler) respondError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"message":   message,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	if err != nil {
		errorResponse["error"].(map[string]interface{})["detail"] = err.Error()
	}

	json.NewEncoder(w).Encode(errorResponse)
}
