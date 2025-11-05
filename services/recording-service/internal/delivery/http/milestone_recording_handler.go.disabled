package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/recording-service/internal/manager"
	"github.com/rs/zerolog"
)

// MilestoneRecordingHandler handles recording-related HTTP requests
type MilestoneRecordingHandler struct {
	recordingManager *manager.MilestoneRecordingManager
	logger           zerolog.Logger
}

// NewMilestoneRecordingHandler creates a new recording handler
func NewMilestoneRecordingHandler(recordingManager *manager.MilestoneRecordingManager, logger zerolog.Logger) *MilestoneRecordingHandler {
	return &MilestoneRecordingHandler{
		recordingManager: recordingManager,
		logger:           logger,
	}
}

// StartRecordingRequest represents the request to start recording
type StartRecordingRequest struct {
	Duration int `json:"duration"` // Duration in seconds
}

// StartRecording starts manual recording for a camera
// POST /recordings/cameras/{cameraId}/start
func (h *MilestoneRecordingHandler) StartRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	var req StartRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Default duration: 15 minutes (900 seconds)
	if req.Duration == 0 {
		req.Duration = 900
	}

	// Validate duration limits
	if req.Duration < 60 {
		h.respondError(w, http.StatusBadRequest, "Duration must be at least 60 seconds")
		return
	}

	if req.Duration > 7200 {
		h.respondError(w, http.StatusBadRequest, "Duration cannot exceed 7200 seconds (2 hours)")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Int("duration", req.Duration).
		Msg("Starting manual recording")

	// Get user ID from context (if available)
	userID := "system"
	if uid, ok := ctx.Value("user_id").(string); ok && uid != "" {
		userID = uid
	}

	managerReq := manager.StartRecordingRequest{
		CameraID:        cameraID,
		DurationSeconds: req.Duration,
		TriggeredBy:     userID,
		Description:     "Manual recording from CNS dashboard",
	}

	session, err := h.recordingManager.StartRecording(ctx, managerReq)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to start recording")

		// Check if it's a conflict (already recording)
		if err.Error() == "camera "+cameraID+" is already recording" {
			h.respondError(w, http.StatusConflict, err.Error())
			return
		}

		h.respondError(w, http.StatusInternalServerError, "Failed to start recording")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Str("recording_id", session.RecordingID).
		Msg("Recording started successfully")

	h.respondJSON(w, http.StatusOK, session)
}

// StopRecording stops manual recording for a camera
// POST /recordings/cameras/{cameraId}/stop
func (h *MilestoneRecordingHandler) StopRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Msg("Stopping manual recording")

	err := h.recordingManager.StopRecording(ctx, cameraID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to stop recording")

		// Check if recording not found
		if err.Error() == "no active recording found for camera "+cameraID {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}

		h.respondError(w, http.StatusInternalServerError, "Failed to stop recording")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Msg("Recording stopped successfully")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Recording stopped successfully",
	})
}

// GetRecordingStatus retrieves current recording status for a camera
// GET /recordings/cameras/{cameraId}/status
func (h *MilestoneRecordingHandler) GetRecordingStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "cameraId")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	h.logger.Debug().
		Str("camera_id", cameraID).
		Msg("Getting recording status")

	status, err := h.recordingManager.GetStatus(ctx, cameraID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to get recording status")
		h.respondError(w, http.StatusInternalServerError, "Failed to get recording status")
		return
	}

	h.respondJSON(w, http.StatusOK, status)
}

// GetActiveRecordings retrieves all active recordings
// GET /recordings/active
func (h *MilestoneRecordingHandler) GetActiveRecordings(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("Getting all active recordings")

	activeRecordings := h.recordingManager.GetActiveRecordings()

	// Convert to response format
	type ActiveRecordingResponse struct {
		CameraID        string `json:"cameraId"`
		RecordingID     string `json:"recordingId"`
		StartTime       string `json:"startTime"`
		DurationSeconds int    `json:"durationSeconds"`
		ElapsedSeconds  int    `json:"elapsedSeconds"`
		RemainingSeconds int   `json:"remainingSeconds"`
		Status          string `json:"status"`
	}

	recordings := make([]ActiveRecordingResponse, 0, len(activeRecordings))
	for _, rec := range activeRecordings {
		elapsed := int(rec.StartTime.Sub(rec.StartTime).Seconds())
		remaining := rec.DurationSeconds - elapsed
		if remaining < 0 {
			remaining = 0
		}

		recordings = append(recordings, ActiveRecordingResponse{
			CameraID:        rec.CameraID,
			RecordingID:     rec.RecordingID,
			StartTime:       rec.StartTime.Format("2006-01-02T15:04:05Z"),
			DurationSeconds: rec.DurationSeconds,
			ElapsedSeconds:  elapsed,
			RemainingSeconds: remaining,
			Status:          rec.Status,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"recordings": recordings,
		"count":      len(recordings),
	})
}

// Helper methods

func (h *MilestoneRecordingHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *MilestoneRecordingHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
