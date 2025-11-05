package http

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/playback-service/internal/usecase"
	"github.com/rs/zerolog"
)

// MilestonePlaybackHandler handles playback-related HTTP requests
type MilestonePlaybackHandler struct {
	playbackUsecase *usecase.MilestonePlaybackUsecase
	logger          zerolog.Logger
}

// NewMilestonePlaybackHandler creates a new playback handler
func NewMilestonePlaybackHandler(playbackUsecase *usecase.MilestonePlaybackUsecase, logger zerolog.Logger) *MilestonePlaybackHandler {
	return &MilestonePlaybackHandler{
		playbackUsecase: playbackUsecase,
		logger:          logger,
	}
}

// QueryRecordingsRequest represents the request to query recordings
type QueryRecordingsRequest struct {
	StartTime string `json:"startTime"` // ISO 8601
	EndTime   string `json:"endTime"`   // ISO 8601
}

// QueryRecordings queries available recordings for a time range
// POST /playback/cameras/{cameraId}/query
func (h *MilestonePlaybackHandler) QueryRecordings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	var req QueryRecordingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse timestamps
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid startTime format (use ISO 8601)")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid endTime format (use ISO 8601)")
		return
	}

	// Validate time range
	if endTime.Before(startTime) {
		h.respondError(w, http.StatusBadRequest, "endTime must be after startTime")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Time("start_time", startTime).
		Time("end_time", endTime).
		Msg("Querying recordings")

	queryReq := usecase.QueryRequest{
		CameraID:  cameraID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	timelineData, err := h.playbackUsecase.QueryRecordings(ctx, queryReq)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to query recordings")
		h.respondError(w, http.StatusInternalServerError, "Failed to query recordings")
		return
	}

	h.respondJSON(w, http.StatusOK, timelineData)
}

// GetTimelineData retrieves aggregated timeline data
// GET /playback/cameras/{cameraId}/timeline
func (h *MilestonePlaybackHandler) GetTimelineData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("startTime")
	endTimeStr := r.URL.Query().Get("endTime")
	resolution := r.URL.Query().Get("resolution")

	if startTimeStr == "" || endTimeStr == "" {
		h.respondError(w, http.StatusBadRequest, "startTime and endTime are required")
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid startTime format")
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid endTime format")
		return
	}

	if resolution == "" {
		resolution = "hour"
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Time("start_time", startTime).
		Time("end_time", endTime).
		Str("resolution", resolution).
		Msg("Getting timeline data")

	queryReq := usecase.QueryRequest{
		CameraID:  cameraID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	timelineData, err := h.playbackUsecase.GetTimelineData(ctx, queryReq, resolution)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to get timeline data")
		h.respondError(w, http.StatusInternalServerError, "Failed to get timeline data")
		return
	}

	h.respondJSON(w, http.StatusOK, timelineData)
}

// StartPlaybackRequest represents the request to start playback
type StartPlaybackRequest struct {
	Timestamp string  `json:"timestamp"` // ISO 8601
	Speed     float64 `json:"speed"`     // Default: 1.0
	Format    string  `json:"format"`    // hls, mjpeg - Default: hls
}

// StartPlayback initiates playback of recorded video
// POST /playback/cameras/{cameraId}/start
func (h *MilestonePlaybackHandler) StartPlayback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	var req StartPlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid timestamp format")
		return
	}

	// Set defaults
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	if req.Format == "" {
		req.Format = "hls"
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Time("timestamp", timestamp).
		Float64("speed", req.Speed).
		Str("format", req.Format).
		Msg("Starting playback")

	playbackReq := usecase.PlaybackRequest{
		CameraID:  cameraID,
		Timestamp: timestamp,
		Speed:     req.Speed,
		Format:    req.Format,
	}

	session, err := h.playbackUsecase.StartPlayback(ctx, playbackReq)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to start playback")
		h.respondError(w, http.StatusInternalServerError, "Failed to start playback")
		return
	}

	h.respondJSON(w, http.StatusOK, session)
}

// GetSnapshot retrieves a snapshot at a specific timestamp
// GET /playback/cameras/{cameraId}/snapshot
func (h *MilestonePlaybackHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	timestampStr := r.URL.Query().Get("time")
	if timestampStr == "" {
		h.respondError(w, http.StatusBadRequest, "time parameter is required")
		return
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid time format")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Time("timestamp", timestamp).
		Msg("Getting snapshot")

	snapshot, err := h.playbackUsecase.GetSnapshot(ctx, cameraID, timestamp)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to get snapshot")
		h.respondError(w, http.StatusInternalServerError, "Failed to get snapshot")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", string(len(snapshot)))
	w.WriteHeader(http.StatusOK)
	w.Write(snapshot)
}

// StreamPlayback streams video for playback
// GET /playback/cameras/{cameraId}/stream
func (h *MilestonePlaybackHandler) StreamPlayback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	// Parse query parameters
	timestampStr := r.URL.Query().Get("time")
	speedStr := r.URL.Query().Get("speed")
	format := r.URL.Query().Get("format")

	if timestampStr == "" {
		h.respondError(w, http.StatusBadRequest, "time parameter is required")
		return
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid time format")
		return
	}

	speed := 1.0
	if speedStr != "" {
		if _, err := fmt.Sscanf(speedStr, "%f", &speed); err != nil {
			h.respondError(w, http.StatusBadRequest, "Invalid speed format")
			return
		}
	}

	if format == "" {
		format = "hls"
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Time("timestamp", timestamp).
		Float64("speed", speed).
		Str("format", format).
		Msg("Streaming playback")

	playbackReq := usecase.PlaybackRequest{
		CameraID:  cameraID,
		Timestamp: timestamp,
		Speed:     speed,
		Format:    format,
	}

	session, err := h.playbackUsecase.StartPlayback(ctx, playbackReq)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to start playback stream")
		h.respondError(w, http.StatusInternalServerError, "Failed to start playback stream")
		return
	}

	// For HLS, return the playlist URL
	if format == "hls" {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		h.respondJSON(w, http.StatusOK, map[string]string{
			"streamUrl": session.StreamURL,
			"sessionId": session.SessionID,
		})
		return
	}

	// For other formats, stream directly (TODO: implement)
	h.respondError(w, http.StatusNotImplemented, "Direct streaming not yet implemented")
}

// ControlPlaybackRequest represents playback control commands
type ControlPlaybackRequest struct {
	Action    string  `json:"action"`    // play, pause, seek, speed
	Timestamp string  `json:"timestamp,omitempty"` // For seek
	Speed     float64 `json:"speed,omitempty"`     // For speed
}

// ControlPlayback controls an active playback session
// POST /playback/cameras/{cameraId}/control
func (h *MilestonePlaybackHandler) ControlPlayback(w http.ResponseWriter, r *http.Request) {
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	var req ControlPlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Str("action", req.Action).
		Msg("Controlling playback")

	// TODO: Implement playback control logic
	// This would manage active playback sessions and send control commands

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"action":  req.Action,
		"cameraId": cameraID,
	})
}

// Helper methods

func (h *MilestonePlaybackHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *MilestonePlaybackHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
