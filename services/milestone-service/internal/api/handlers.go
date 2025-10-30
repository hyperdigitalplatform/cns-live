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

// ============================================================================
// WebRTC Playback Handlers
// ============================================================================

// StartWebRTCPlayback handles POST /api/v1/cameras/:cameraId/playback/start
// Initiates a WebRTC playback session with Milestone
func (h *Handler) StartWebRTCPlayback(c *gin.Context) {
	cameraId := c.Param("cameraId")

	var req WebRTCPlaybackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Validate playback time
	if req.PlaybackTime == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "playbackTime is required",
		})
		return
	}

	// Default speed to 1.0 if not specified
	if req.Speed == 0 {
		req.Speed = 1.0
	}

	ctx := context.Background()

	// Create WebRTC session with Milestone
	session, err := h.restClient.CreateWebRTCPlaybackSession(ctx, rest.WebRTCPlaybackRequest{
		DeviceID:     cameraId,
		PlaybackTime: req.PlaybackTime,
		SkipGaps:     req.SkipGaps,
		Speed:        req.Speed,
		StreamID:     req.StreamID,
		IncludeAudio: false, // Audio not supported in playback for now
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "webrtc_session_failed",
			Message: fmt.Sprintf("Failed to create WebRTC playback session: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, WebRTCSessionResponse{
		SessionID: session.SessionID,
		OfferSDP:  session.OfferSDP,
	})
}

// UpdateWebRTCAnswer handles PUT /api/v1/playback/webrtc/answer
// Updates the WebRTC session with the answer SDP from client
func (h *Handler) UpdateWebRTCAnswer(c *gin.Context) {
	var req WebRTCAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	err := h.restClient.UpdateWebRTCAnswer(ctx, req.SessionID, req.AnswerSDP)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_failed",
			Message: fmt.Sprintf("Failed to update WebRTC answer: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// SendICECandidate handles POST /api/v1/playback/webrtc/ice
// Sends an ICE candidate from client to Milestone
func (h *Handler) SendICECandidate(c *gin.Context) {
	var req ICECandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	err := h.restClient.SendICECandidate(ctx, req.SessionID, req.Candidate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "ice_failed",
			Message: fmt.Sprintf("Failed to send ICE candidate: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// GetICECandidates handles GET /api/v1/playback/webrtc/ice/:sessionId
// Retrieves ICE candidates from Milestone
func (h *Handler) GetICECandidates(c *gin.Context) {
	sessionId := c.Param("sessionId")

	if sessionId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "sessionId is required",
		})
		return
	}

	ctx := context.Background()
	candidates, err := h.restClient.GetICECandidates(ctx, sessionId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "ice_failed",
			Message: fmt.Sprintf("Failed to get ICE candidates: %v", err),
		})
		return
	}

	// Convert []ICECandidate to []interface{} for JSON response
	candidateInterfaces := make([]interface{}, len(candidates))
	for i, c := range candidates {
		candidateInterfaces[i] = c
	}

	c.JSON(http.StatusOK, ICECandidatesResponse{
		Candidates: candidateInterfaces,
	})
}
