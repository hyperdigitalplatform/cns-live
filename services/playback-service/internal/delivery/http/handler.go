package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rta/cctv/playback-service/internal/hls"
	"github.com/rta/cctv/playback-service/internal/manager"
	"github.com/rs/zerolog"
)

type Handler struct {
	playbackMgr *manager.PlaybackManager
	hlsService  *hls.HLSService
	logger      zerolog.Logger
}

func NewHandler(
	playbackMgr *manager.PlaybackManager,
	hlsService *hls.HLSService,
	logger zerolog.Logger,
) *Handler {
	return &Handler{
		playbackMgr: playbackMgr,
		hlsService:  hlsService,
		logger:      logger,
	}
}

// StartPlayback initiates playback for a time range
func (h *Handler) StartPlayback(w http.ResponseWriter, r *http.Request) {
	var req domain.PlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate time range
	if req.EndTime.Before(req.StartTime) {
		h.respondError(w, http.StatusBadRequest, "End time must be after start time")
		return
	}

	response, err := h.playbackMgr.StartPlayback(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to start playback")
		h.respondError(w, http.StatusInternalServerError, "Failed to start playback")
		return
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// StartLiveStream initiates live streaming
func (h *Handler) StartLiveStream(w http.ResponseWriter, r *http.Request) {
	var req domain.LiveStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.playbackMgr.StartLiveStream(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to start live stream")
		h.respondError(w, http.StatusInternalServerError, "Failed to start live stream")
		return
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// GetSession retrieves a playback session
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	session, err := h.playbackMgr.GetSession(r.Context(), sessionID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	h.respondJSON(w, http.StatusOK, session)
}

// StopPlayback stops a playback session
func (h *Handler) StopPlayback(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	if err := h.playbackMgr.StopPlayback(r.Context(), sessionID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to stop playback")
		h.respondError(w, http.StatusInternalServerError, "Failed to stop playback")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// ExtendSession extends session TTL
func (h *Handler) ExtendSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		DurationSeconds int `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	duration := time.Duration(req.DurationSeconds) * time.Second
	if err := h.playbackMgr.ExtendSession(r.Context(), sessionID, duration); err != nil {
		h.logger.Error().Err(err).Msg("Failed to extend session")
		h.respondError(w, http.StatusInternalServerError, "Failed to extend session")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "extended"})
}

// GetHLSManifest serves the HLS m3u8 playlist
func (h *Handler) GetHLSManifest(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	manifest, err := h.hlsService.GetManifest(r.Context(), sessionID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get manifest")
		h.respondError(w, http.StatusNotFound, "Manifest not found")
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(manifest))
}

// GetHLSSegment serves an HLS video segment
func (h *Handler) GetHLSSegment(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	segmentStr := chi.URLParam(r, "segment")

	// Extract segment index from "123.ts"
	segmentIndex, err := strconv.Atoi(strings.TrimSuffix(segmentStr, ".ts"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid segment index")
		return
	}

	// Get segment data
	reader, err := h.hlsService.GetSegment(r.Context(), sessionID, segmentIndex)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get segment")
		h.respondError(w, http.StatusNotFound, "Segment not found")
		return
	}
	defer reader.Close()

	// Stream segment to client
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, reader); err != nil {
		h.logger.Error().Err(err).Msg("Failed to stream segment")
	}
}

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// Helper methods

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
