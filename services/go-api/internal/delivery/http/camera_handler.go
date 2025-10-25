package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/go-api/internal/client"
	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rs/zerolog"
)

// CameraHandler handles camera-related HTTP requests
type CameraHandler struct {
	vmsClient *client.VMSClient
	logger    zerolog.Logger
}

// NewCameraHandler creates a new camera handler
func NewCameraHandler(vmsClient *client.VMSClient, logger zerolog.Logger) *CameraHandler {
	return &CameraHandler{
		vmsClient: vmsClient,
		logger:    logger,
	}
}

// ListCameras handles camera list request
// GET /api/v1/cameras
func (h *CameraHandler) ListCameras(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := domain.CameraQuery{
		Source: r.URL.Query().Get("source"),
		Status: r.URL.Query().Get("status"),
		Search: r.URL.Query().Get("search"),
	}

	// Parse pagination
	if limit := r.URL.Query().Get("limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &query.Limit)
	} else {
		query.Limit = 100 // Default
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		fmt.Sscanf(offset, "%d", &query.Offset)
	}

	cameras, err := h.vmsClient.ListCameras(r.Context(), query)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list cameras")
		h.respondError(w, http.StatusInternalServerError, "Failed to list cameras")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"cameras": cameras,
		"count":   len(cameras),
	})
}

// GetCamera handles single camera request
// GET /api/v1/cameras/{id}
func (h *CameraHandler) GetCamera(w http.ResponseWriter, r *http.Request) {
	cameraID := chi.URLParam(r, "id")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	camera, err := h.vmsClient.GetCamera(r.Context(), cameraID)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Failed to get camera")
		h.respondError(w, http.StatusNotFound, "Camera not found")
		return
	}

	h.respondJSON(w, http.StatusOK, camera)
}

// ControlPTZ handles PTZ control request
// POST /api/v1/cameras/{id}/ptz
func (h *CameraHandler) ControlPTZ(w http.ResponseWriter, r *http.Request) {
	cameraID := chi.URLParam(r, "id")

	var cmd domain.PTZCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd.CameraID = cameraID

	// Validate command
	validCommands := map[string]bool{
		"pan_left":  true,
		"pan_right": true,
		"tilt_up":   true,
		"tilt_down": true,
		"zoom_in":   true,
		"zoom_out":  true,
		"preset":    true,
		"home":      true,
	}

	if !validCommands[cmd.Command] {
		h.respondError(w, http.StatusBadRequest, "Invalid PTZ command")
		return
	}

	err := h.vmsClient.ControlPTZ(r.Context(), cmd)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to control PTZ")
		h.respondError(w, http.StatusInternalServerError, "Failed to control PTZ")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "PTZ command executed",
	})
}

// Helper methods

func (h *CameraHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *CameraHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
