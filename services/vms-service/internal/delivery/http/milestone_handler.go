package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/vms-service/internal/client"
	"github.com/rta/cctv/vms-service/internal/domain"
	"github.com/rta/cctv/vms-service/internal/repository"
	"github.com/rs/zerolog"
)

// MilestoneHandler handles Milestone-specific HTTP requests
type MilestoneHandler struct {
	milestoneClient *client.MilestoneClient
	cameraRepo      repository.CameraRepository
	logger          zerolog.Logger
}

// NewMilestoneHandler creates a new Milestone handler
func NewMilestoneHandler(milestoneClient *client.MilestoneClient, cameraRepo repository.CameraRepository, logger zerolog.Logger) *MilestoneHandler {
	return &MilestoneHandler{
		milestoneClient: milestoneClient,
		cameraRepo:      cameraRepo,
		logger:          logger,
	}
}

// ListMilestoneCameras lists all cameras from Milestone XProtect
// GET /vms/milestone/cameras
func (h *MilestoneHandler) ListMilestoneCameras(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	var enabledFilter *bool
	if enabledStr := r.URL.Query().Get("enabled"); enabledStr != "" {
		enabled := enabledStr == "true"
		enabledFilter = &enabled
	}

	var recordingFilter *bool
	if recordingStr := r.URL.Query().Get("recording"); recordingStr != "" {
		recording := recordingStr == "true"
		recordingFilter = &recording
	}

	h.logger.Info().
		Int("limit", limit).
		Int("offset", offset).
		Msg("Listing cameras from Milestone")

	opts := client.ListCamerasOptions{
		Limit:     limit,
		Offset:    offset,
		Enabled:   enabledFilter,
		Recording: recordingFilter,
	}

	cameraList, err := h.milestoneClient.ListCameras(ctx, opts)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list cameras from Milestone")
		h.respondError(w, http.StatusInternalServerError, "Failed to list cameras from Milestone")
		return
	}

	// Check which cameras are already imported
	importedCameras, err := h.cameraRepo.GetAll(ctx)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to get imported cameras")
		importedCameras = []*domain.Camera{} // Continue with empty list
	}

	// Create a map of imported Milestone device IDs
	importedMap := make(map[string]bool)
	for _, cam := range importedCameras {
		if cam.MilestoneDeviceID != "" {
			importedMap[cam.MilestoneDeviceID] = true
		}
	}

	// Build response with import status
	type CameraResponse struct {
		*client.MilestoneCamera
		Imported bool `json:"imported"`
	}

	cameras := make([]CameraResponse, len(cameraList.Cameras))
	importedCount := 0
	for i, cam := range cameraList.Cameras {
		imported := importedMap[cam.ID]
		if imported {
			importedCount++
		}
		cameras[i] = CameraResponse{
			MilestoneCamera: cam,
			Imported:        imported,
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"cameras":  cameras,
		"total":    cameraList.Total,
		"imported": importedCount,
	})
}

// GetMilestoneCamera retrieves a single camera from Milestone
// GET /vms/milestone/cameras/{id}
func (h *MilestoneHandler) GetMilestoneCamera(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "id")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	h.logger.Info().Str("camera_id", cameraID).Msg("Getting camera from Milestone")

	camera, err := h.milestoneClient.GetCamera(ctx, cameraID)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Failed to get camera from Milestone")
		h.respondError(w, http.StatusNotFound, "Camera not found in Milestone")
		return
	}

	h.respondJSON(w, http.StatusOK, camera)
}

// ImportCameraRequest represents a request to import a camera from Milestone
type ImportCameraRequest struct {
	MilestoneDeviceID string                 `json:"milestoneDeviceId" validate:"required"`
	Name              string                 `json:"name,omitempty"`
	NameAr            string                 `json:"nameAr,omitempty"`
	Source            string                 `json:"source" validate:"required"`
	Location          *domain.Location       `json:"location,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ImportCamera imports a camera from Milestone into the CNS system
// POST /vms/cameras/import
func (h *MilestoneHandler) ImportCamera(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ImportCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.MilestoneDeviceID == "" {
		h.respondError(w, http.StatusBadRequest, "milestoneDeviceId is required")
		return
	}

	if req.Source == "" {
		h.respondError(w, http.StatusBadRequest, "source is required")
		return
	}

	h.logger.Info().
		Str("milestone_device_id", req.MilestoneDeviceID).
		Str("source", req.Source).
		Msg("Importing camera from Milestone")

	// Get camera details from Milestone
	milestoneCamera, err := h.milestoneClient.GetCamera(ctx, req.MilestoneDeviceID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("milestone_device_id", req.MilestoneDeviceID).
			Msg("Failed to get camera from Milestone")
		h.respondError(w, http.StatusNotFound, "Camera not found in Milestone")
		return
	}

	// Check if camera is already imported
	existingCameras, err := h.cameraRepo.GetAll(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check existing cameras")
		h.respondError(w, http.StatusInternalServerError, "Failed to check existing cameras")
		return
	}

	for _, existing := range existingCameras {
		if existing.MilestoneDeviceID == req.MilestoneDeviceID {
			h.logger.Warn().
				Str("milestone_device_id", req.MilestoneDeviceID).
				Str("existing_camera_id", existing.ID).
				Msg("Camera already imported")
			h.respondError(w, http.StatusConflict, "Camera already imported")
			return
		}
	}

	// Use provided name or default to Milestone name
	name := req.Name
	if name == "" {
		name = milestoneCamera.Name
	}

	// Build camera object
	camera := &domain.Camera{
		Name:              name,
		NameAr:            req.NameAr,
		Source:            domain.CameraSource(req.Source),
		RTSPURL:           milestoneCamera.LiveStreamURL,
		PTZEnabled:        milestoneCamera.PTZCapabilities != nil && (milestoneCamera.PTZCapabilities.Pan || milestoneCamera.PTZCapabilities.Tilt || milestoneCamera.PTZCapabilities.Zoom),
		Status:            domain.StatusOnline,
		RecordingServer:   milestoneCamera.RecordingServer,
		MilestoneDeviceID: req.MilestoneDeviceID,
		Metadata:          req.Metadata,
	}

	if camera.Metadata == nil {
		camera.Metadata = make(map[string]interface{})
	}

	// Add Milestone metadata
	camera.Metadata["milestone_enabled"] = milestoneCamera.Enabled
	camera.Metadata["milestone_recording"] = milestoneCamera.Recording
	if milestoneCamera.Metadata != nil {
		for k, v := range milestoneCamera.Metadata {
			camera.Metadata[k] = v
		}
	}

	// Create camera in database
	createdCamera, err := h.cameraRepo.Create(ctx, camera)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create camera in database")
		h.respondError(w, http.StatusInternalServerError, "Failed to import camera")
		return
	}

	h.logger.Info().
		Str("camera_id", createdCamera.ID).
		Str("milestone_device_id", req.MilestoneDeviceID).
		Msg("Camera imported successfully")

	h.respondJSON(w, http.StatusCreated, createdCamera)
}

// BulkSyncRequest represents a request to sync cameras from Milestone
type BulkSyncRequest struct {
	SourceFilter string `json:"sourceFilter,omitempty"`
	AutoImport   bool   `json:"autoImport"`
}

// BulkSyncResponse represents the response for bulk sync
type BulkSyncResponse struct {
	SyncJobID         string `json:"syncJobId"`
	Status            string `json:"status"`
	CamerasDiscovered int    `json:"camerasDiscovered"`
	CamerasImported   int    `json:"camerasImported"`
	Message           string `json:"message"`
}

// BulkSyncCameras synchronizes cameras from Milestone
// POST /vms/milestone/sync-all
func (h *MilestoneHandler) BulkSyncCameras(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BulkSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Info().
		Str("source_filter", req.SourceFilter).
		Bool("auto_import", req.AutoImport).
		Msg("Starting bulk camera sync from Milestone")

	// List all cameras from Milestone
	opts := client.ListCamerasOptions{
		Limit:  1000, // Get all cameras
		Offset: 0,
	}

	cameraList, err := h.milestoneClient.ListCameras(ctx, opts)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list cameras from Milestone")
		h.respondError(w, http.StatusInternalServerError, "Failed to sync cameras from Milestone")
		return
	}

	discovered := len(cameraList.Cameras)
	imported := 0

	if req.AutoImport {
		// Get existing cameras
		existingCameras, err := h.cameraRepo.GetAll(ctx)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get existing cameras")
			h.respondError(w, http.StatusInternalServerError, "Failed to get existing cameras")
			return
		}

		existingMap := make(map[string]bool)
		for _, cam := range existingCameras {
			if cam.MilestoneDeviceID != "" {
				existingMap[cam.MilestoneDeviceID] = true
			}
		}

		// Import new cameras
		for _, milestoneCamera := range cameraList.Cameras {
			if existingMap[milestoneCamera.ID] {
				continue // Already imported
			}

			// Auto-import with default source
			source := domain.SourceOther
			if req.SourceFilter != "" {
				source = domain.CameraSource(req.SourceFilter)
			}

			camera := &domain.Camera{
				Name:              milestoneCamera.Name,
				Source:            source,
				RTSPURL:           milestoneCamera.LiveStreamURL,
				PTZEnabled:        milestoneCamera.PTZCapabilities != nil,
				Status:            domain.StatusOnline,
				RecordingServer:   milestoneCamera.RecordingServer,
				MilestoneDeviceID: milestoneCamera.ID,
				Metadata: map[string]interface{}{
					"milestone_enabled":   milestoneCamera.Enabled,
					"milestone_recording": milestoneCamera.Recording,
				},
			}

			_, err := h.cameraRepo.Create(ctx, camera)
			if err != nil {
				h.logger.Warn().
					Err(err).
					Str("milestone_device_id", milestoneCamera.ID).
					Msg("Failed to import camera during bulk sync")
				continue
			}

			imported++
		}
	}

	h.logger.Info().
		Int("discovered", discovered).
		Int("imported", imported).
		Msg("Bulk sync completed")

	response := BulkSyncResponse{
		SyncJobID:         "sync_" + strconv.FormatInt(r.Context().Value("request_id").(int64), 10),
		Status:            "completed",
		CamerasDiscovered: discovered,
		CamerasImported:   imported,
		Message:           "Camera synchronization completed successfully",
	}

	h.respondJSON(w, http.StatusOK, response)
}

// SyncCameraWithMilestone syncs a single camera with Milestone
// PUT /vms/cameras/{id}/sync
func (h *MilestoneHandler) SyncCameraWithMilestone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cameraID := chi.URLParam(r, "id")

	if cameraID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id is required")
		return
	}

	h.logger.Info().Str("camera_id", cameraID).Msg("Syncing camera with Milestone")

	// Get camera from database
	camera, err := h.cameraRepo.GetByID(ctx, cameraID)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Camera not found")
		h.respondError(w, http.StatusNotFound, "Camera not found")
		return
	}

	if camera.MilestoneDeviceID == "" {
		h.respondError(w, http.StatusBadRequest, "Camera is not linked to Milestone")
		return
	}

	// Get latest data from Milestone
	milestoneCamera, err := h.milestoneClient.GetCamera(ctx, camera.MilestoneDeviceID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Str("milestone_device_id", camera.MilestoneDeviceID).
			Msg("Failed to get camera from Milestone")
		h.respondError(w, http.StatusInternalServerError, "Failed to sync with Milestone")
		return
	}

	// Update camera with latest Milestone data
	camera.RTSPURL = milestoneCamera.LiveStreamURL
	camera.RecordingServer = milestoneCamera.RecordingServer
	camera.PTZEnabled = milestoneCamera.PTZCapabilities != nil

	if camera.Metadata == nil {
		camera.Metadata = make(map[string]interface{})
	}
	camera.Metadata["milestone_enabled"] = milestoneCamera.Enabled
	camera.Metadata["milestone_recording"] = milestoneCamera.Recording
	camera.Metadata["last_sync"] = r.Context().Value("request_time")

	// Update in database
	updatedCamera, err := h.cameraRepo.Update(ctx, camera)
	if err != nil {
		h.logger.Error().Err(err).Str("camera_id", cameraID).Msg("Failed to update camera")
		h.respondError(w, http.StatusInternalServerError, "Failed to update camera")
		return
	}

	h.logger.Info().
		Str("camera_id", cameraID).
		Msg("Camera synced successfully")

	h.respondJSON(w, http.StatusOK, updatedCamera)
}

// Helper methods

func (h *MilestoneHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *MilestoneHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
