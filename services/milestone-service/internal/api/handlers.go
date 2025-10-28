package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"milestone-service/internal/rest"
	"milestone-service/internal/service"
	"milestone-service/internal/soap"
)

// Handler holds the API handlers and dependencies
type Handler struct {
	soapClient       *soap.Client
	restClient       *rest.Client
	discoveryService *service.DiscoveryService
}

// NewHandler creates a new API handler
func NewHandler(soapClient *soap.Client, restClient *rest.Client, discoveryService *service.DiscoveryService) *Handler {
	return &Handler{
		soapClient:       soapClient,
		restClient:       restClient,
		discoveryService: discoveryService,
	}
}

// StartRecording handles POST /api/recordings/start
func (h *Handler) StartRecording(c *gin.Context) {
	var req StartRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	err := h.soapClient.StartManualRecording(ctx, []string{req.CameraId}, req.DurationMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recording_start_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Recording started successfully",
	})
}

// StopRecording handles POST /api/recordings/stop
func (h *Handler) StopRecording(c *gin.Context) {
	var req StopRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	err := h.soapClient.StopManualRecording(ctx, []string{req.CameraId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recording_stop_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Recording stopped successfully",
	})
}

// GetRecordingStatus handles GET /api/recordings/status/:cameraId
func (h *Handler) GetRecordingStatus(c *gin.Context) {
	cameraId := c.Param("cameraId")
	if cameraId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "cameraId is required",
		})
		return
	}

	ctx := context.Background()
	isRecording, err := h.soapClient.IsManualRecording(ctx, []string{cameraId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "status_check_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecordingStatusResponse{
		CameraId:    cameraId,
		IsRecording: isRecording,
	})
}

// GetSequenceTypes handles GET /api/sequences/types/:cameraId
func (h *Handler) GetSequenceTypes(c *gin.Context) {
	cameraId := c.Param("cameraId")
	if cameraId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "cameraId is required",
		})
		return
	}

	ctx := context.Background()
	types, err := h.soapClient.SequencesGetTypes(ctx, []string{cameraId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "sequence_types_failed",
			Message: err.Error(),
		})
		return
	}

	// Convert SOAP types to API types
	apiTypes := make([]SequenceType, len(types))
	for i, t := range types {
		apiTypes[i] = SequenceType{
			Id:   t.Id,
			Name: t.Name,
		}
	}

	c.JSON(http.StatusOK, SequenceTypesResponse{
		CameraId: cameraId,
		Types:    apiTypes,
	})
}

// GetSequences handles POST /api/sequences
func (h *Handler) GetSequences(c *gin.Context) {
	var req SequencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Parse time strings
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time_format",
			Message: "startTime must be in ISO 8601 format",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time_format",
			Message: "endTime must be in ISO 8601 format",
		})
		return
	}

	ctx := context.Background()
	sequences, err := h.soapClient.SequencesGet(ctx, []string{req.CameraId}, startTime, endTime, req.SequenceTypes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "sequences_get_failed",
			Message: err.Error(),
		})
		return
	}

	// Convert SOAP sequences to API sequences
	apiSequences := make([]SequenceEntry, len(sequences))
	for i, s := range sequences {
		apiSequences[i] = SequenceEntry{
			TimeBegin:   s.TimeBegin,
			TimeTrigged: s.TimeTrigged,
			TimeEnd:     s.TimeEnd,
		}
	}

	c.JSON(http.StatusOK, SequencesResponse{
		CameraId:  req.CameraId,
		Sequences: apiSequences,
	})
}

// GetTimeline handles POST /api/timeline
func (h *Handler) GetTimeline(c *gin.Context) {
	var req TimelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Parse time strings
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time_format",
			Message: "startTime must be in ISO 8601 format",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time_format",
			Message: "endTime must be in ISO 8601 format",
		})
		return
	}

	ctx := context.Background()
	timeline, err := h.soapClient.TimeLineInformationGet(ctx, []string{req.CameraId}, startTime, endTime, req.SequenceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "timeline_get_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TimelineResponse{
		CameraId: req.CameraId,
		Timeline: TimelineInformation{
			Count: timeline.Count,
			Data:  timeline.Data,
		},
	})
}

// GetCameras handles GET /api/v1/cameras
func (h *Handler) GetCameras(c *gin.Context) {
	ctx := context.Background()
	cameras, err := h.restClient.GetCameras(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "cameras_get_failed",
			Message: err.Error(),
		})
		return
	}

	// Convert REST cameras to API cameras
	apiCameras := make([]MilestoneCamera, len(cameras))
	for i, cam := range cameras {
		apiCameras[i] = MilestoneCamera{
			Id:              cam.ID,
			Name:            cam.Name,
			Enabled:         cam.Enabled,
			Status:          cam.Status,
			RecordingServer: cam.RecordingServer,
			ShortName:       cam.ShortName,
			LiveStreamUrl:   "",  // Not provided in REST API
			RtspUrl:         cam.RtspURL,
			PtzEnabled:      cam.PTZEnabled,
			PtzCapabilities: PtzCapabilities{
				Pan:  cam.PTZEnabled,  // Assume all capabilities if PTZ enabled
				Tilt: cam.PTZEnabled,
				Zoom: cam.PTZEnabled,
			},
		}
	}

	c.JSON(http.StatusOK, CamerasResponse{
		Cameras: apiCameras,
		Total:   len(apiCameras),
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "milestone-service",
	})
}

// DiscoverCameras handles GET /api/v1/cameras/discover
func (h *Handler) DiscoverCameras(c *gin.Context) {
	ctx := context.Background()

	// Discover cameras with ONVIF enrichment
	cameras, err := h.discoveryService.DiscoverCameras(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "discovery_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CameraDiscoveryResponse{
		Cameras: cameras,
		Total:   len(cameras),
	})
}

// StreamPlayback handles GET /api/v1/cameras/:cameraId/playback/stream
// Proxies video stream from Milestone with authentication
func (h *Handler) StreamPlayback(c *gin.Context) {
	cameraId := c.Param("cameraId")
	timestampStr := c.Query("time")

	if cameraId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "cameraId is required",
		})
		return
	}

	if timestampStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "time parameter is required",
		})
		return
	}

	// Parse the timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time_format",
			Message: "time must be in ISO 8601 format",
		})
		return
	}

	ctx := context.Background()

	// Set headers for video streaming
	c.Header("Content-Type", "video/mp4")
	c.Header("Cache-Control", "no-cache")
	c.Header("Accept-Ranges", "bytes")

	// Proxy the stream from Milestone REST API
	err = h.restClient.ProxyPlaybackStream(ctx, c.Writer, cameraId, timestamp)
	if err != nil {
		// Only send error if headers haven't been sent yet
		if !c.Writer.Written() {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "playback_stream_failed",
				Message: fmt.Sprintf("Failed to stream playback: %v", err),
			})
		}
		return
	}
}
