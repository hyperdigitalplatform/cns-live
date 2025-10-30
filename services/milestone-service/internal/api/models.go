package api

import (
	"time"

	"milestone-service/internal/types"
)

// StartRecordingRequest represents the request to start manual recording
type StartRecordingRequest struct {
	CameraId        string `json:"cameraId" binding:"required"`
	DurationMinutes int    `json:"durationMinutes,omitempty"` // Default 15 if not specified
}

// StopRecordingRequest represents the request to stop manual recording
type StopRecordingRequest struct {
	CameraId string `json:"cameraId" binding:"required"`
}

// RecordingStatusRequest represents the request to check recording status
type RecordingStatusRequest struct {
	CameraId string `json:"cameraId" binding:"required"`
}

// RecordingStatusResponse represents the response with recording status
type RecordingStatusResponse struct {
	CameraId    string `json:"cameraId"`
	IsRecording bool   `json:"isRecording"`
}

// SequenceTypesRequest represents the request to get sequence types
type SequenceTypesRequest struct {
	CameraId string `json:"cameraId" binding:"required"`
}

// SequenceTypesResponse represents the response with sequence types
type SequenceTypesResponse struct {
	CameraId string         `json:"cameraId"`
	Types    []SequenceType `json:"types"`
}

// SequenceType represents a recording sequence type
type SequenceType struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// SequencesRequest represents the request to get recording sequences
type SequencesRequest struct {
	CameraId      string   `json:"cameraId" binding:"required"`
	StartTime     string   `json:"startTime" binding:"required"` // ISO 8601 format
	EndTime       string   `json:"endTime" binding:"required"`   // ISO 8601 format
	SequenceTypes []string `json:"sequenceTypes,omitempty"`      // Optional filter
}

// SequencesResponse represents the response with recording sequences
type SequencesResponse struct {
	CameraId  string          `json:"cameraId"`
	Sequences []SequenceEntry `json:"sequences"`
}

// SequenceEntry represents a single recording sequence
type SequenceEntry struct {
	TimeBegin   time.Time `json:"timeBegin"`
	TimeTrigged time.Time `json:"timeTrigged"`
	TimeEnd     time.Time `json:"timeEnd"`
}

// TimelineRequest represents the request to get timeline data
type TimelineRequest struct {
	CameraId     string `json:"cameraId" binding:"required"`
	StartTime    string `json:"startTime" binding:"required"` // ISO 8601 format
	EndTime      string `json:"endTime" binding:"required"`   // ISO 8601 format
	SequenceType string `json:"sequenceType,omitempty"`       // Optional, defaults to RecordedDataAvailable
}

// TimelineResponse represents the response with timeline data
type TimelineResponse struct {
	CameraId string              `json:"cameraId"`
	Timeline TimelineInformation `json:"timeline"`
}

// TimelineInformation represents timeline bitmap data
type TimelineInformation struct {
	Count int    `json:"count"` // Number of time intervals
	Data  string `json:"data"`  // Base64 encoded bitmap
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// CamerasResponse represents the response with camera list
type CamerasResponse struct {
	Cameras []MilestoneCamera `json:"cameras"`
	Total   int               `json:"total"`
}

// MilestoneCamera represents a camera from Milestone XProtect
type MilestoneCamera struct {
	Id              string          `json:"id"`
	Name            string          `json:"name"`
	Enabled         bool            `json:"enabled"`
	Status          string          `json:"status"`          // "ONLINE", "OFFLINE", "MAINTENANCE", "ERROR"
	RecordingServer string          `json:"recordingServer"` // Recording server name
	ShortName       string          `json:"shortName"`       // Short display name
	LiveStreamUrl   string          `json:"liveStreamUrl"`   // Stream ID from Milestone
	RtspUrl         string          `json:"rtspUrl"`         // Constructed RTSP URL for streaming
	PtzEnabled      bool            `json:"ptzEnabled"`      // Whether PTZ is enabled
	PtzCapabilities PtzCapabilities `json:"ptzCapabilities"` // PTZ capabilities
}

// PtzCapabilities represents PTZ capabilities
type PtzCapabilities struct {
	Pan  bool `json:"pan"`  // Pan capability
	Tilt bool `json:"tilt"` // Tilt capability
	Zoom bool `json:"zoom"` // Zoom capability
}

// CameraDiscoveryResponse represents the response from camera discovery
type CameraDiscoveryResponse struct {
	Cameras []types.DiscoveredCamera `json:"cameras"`
	Total   int                      `json:"total"`
}

// ============================================================================
// WebRTC Playback Models
// ============================================================================

// WebRTCPlaybackRequest represents the request to start WebRTC playback
type WebRTCPlaybackRequest struct {
	PlaybackTime string  `json:"playbackTime" binding:"required"` // ISO 8601 format
	SkipGaps     bool    `json:"skipGaps"`                        // Skip gaps between recordings
	Speed        float64 `json:"speed"`                           // Playback speed (default 1.0)
	StreamID     string  `json:"streamId,omitempty"`              // Optional stream ID
}

// WebRTCAnswerRequest represents the answer SDP from client
type WebRTCAnswerRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
	AnswerSDP string `json:"answerSDP" binding:"required"`
}

// ICECandidateRequest represents an ICE candidate from client
type ICECandidateRequest struct {
	SessionID string      `json:"sessionId" binding:"required"`
	Candidate interface{} `json:"candidate" binding:"required"`
}

// WebRTCSessionResponse represents the WebRTC session response
type WebRTCSessionResponse struct {
	SessionID string `json:"sessionId"`
	OfferSDP  string `json:"offerSDP"`
}

// ICECandidatesResponse represents ICE candidates from server
type ICECandidatesResponse struct {
	Candidates []interface{} `json:"candidates"`
}
