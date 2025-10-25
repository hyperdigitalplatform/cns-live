package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rta/cctv/playback-service/internal/client"
	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rta/cctv/playback-service/internal/hls"
	"github.com/rta/cctv/playback-service/internal/repository"
	"github.com/rs/zerolog"
)

// PlaybackManager orchestrates playback operations
type PlaybackManager struct {
	hlsService    *hls.HLSService
	vmsClient     *client.VMSClient
	storageClient *client.StorageClient
	cacheRepo     *repository.CacheRepository
	mediamtxURL   string
	logger        zerolog.Logger
}

// NewPlaybackManager creates a new playback manager
func NewPlaybackManager(
	hlsService *hls.HLSService,
	vmsClient *client.VMSClient,
	storageClient *client.StorageClient,
	cacheRepo *repository.CacheRepository,
	mediamtxURL string,
	logger zerolog.Logger,
) *PlaybackManager {
	return &PlaybackManager{
		hlsService:    hlsService,
		vmsClient:     vmsClient,
		storageClient: storageClient,
		cacheRepo:     cacheRepo,
		mediamtxURL:   mediamtxURL,
		logger:        logger,
	}
}

// StartPlayback initiates playback for a time range
func (m *PlaybackManager) StartPlayback(
	ctx context.Context,
	req domain.PlaybackRequest,
) (*domain.PlaybackResponse, error) {
	// Validate camera exists
	camera, err := m.vmsClient.GetCamera(ctx, req.CameraID)
	if err != nil {
		return nil, fmt.Errorf("failed to get camera: %w", err)
	}

	m.logger.Info().
		Str("camera_id", req.CameraID).
		Str("camera_name", camera.Name).
		Time("start_time", req.StartTime).
		Time("end_time", req.EndTime).
		Str("format", req.Format).
		Msg("Starting playback")

	// Handle based on format
	switch req.Format {
	case "hls":
		return m.hlsService.CreatePlaybackSession(ctx, req.CameraID, req.StartTime, req.EndTime)
	case "rtsp":
		return nil, fmt.Errorf("RTSP playback not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}
}

// StartLiveStream initiates live streaming for a camera
func (m *PlaybackManager) StartLiveStream(
	ctx context.Context,
	req domain.LiveStreamRequest,
) (*domain.LiveStreamResponse, error) {
	// Validate camera exists
	camera, err := m.vmsClient.GetCamera(ctx, req.CameraID)
	if err != nil {
		return nil, fmt.Errorf("failed to get camera: %w", err)
	}

	sessionID := uuid.New().String()

	m.logger.Info().
		Str("camera_id", req.CameraID).
		Str("camera_name", camera.Name).
		Str("format", req.Format).
		Msg("Starting live stream")

	var streamURL string

	switch req.Format {
	case "hls":
		// MediaMTX provides HLS at /camera-id/index.m3u8
		streamURL = fmt.Sprintf("%s/%s/index.m3u8", m.mediamtxURL, req.CameraID)
	case "rtsp":
		// Get RTSP URL from VMS service
		rtspURL, err := m.vmsClient.GetRTSPURL(ctx, req.CameraID)
		if err != nil {
			return nil, fmt.Errorf("failed to get RTSP URL: %w", err)
		}
		streamURL = rtspURL
	case "webrtc":
		// MediaMTX provides WebRTC at /camera-id/whep
		streamURL = fmt.Sprintf("%s/%s/whep", m.mediamtxURL, req.CameraID)
	default:
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}

	response := &domain.LiveStreamResponse{
		SessionID: sessionID,
		CameraID:  req.CameraID,
		Format:    req.Format,
		URL:       streamURL,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Cache session (24 hour TTL for live streams)
	if err := m.cacheRepo.SetSession(ctx, sessionID, response, 24*time.Hour); err != nil {
		m.logger.Warn().Err(err).Msg("Failed to cache live session")
	}

	return response, nil
}

// StopPlayback stops a playback session
func (m *PlaybackManager) StopPlayback(ctx context.Context, sessionID string) error {
	// Check if session exists
	var session domain.PlaybackResponse
	if err := m.cacheRepo.GetSession(ctx, sessionID, &session); err != nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Cleanup HLS resources
	if session.Format == "hls" {
		if err := m.hlsService.CleanupSession(ctx, sessionID); err != nil {
			m.logger.Error().Err(err).Msg("Failed to cleanup HLS session")
		}
	}

	m.logger.Info().Str("session_id", sessionID).Msg("Stopped playback session")
	return nil
}

// GetSession retrieves a playback session
func (m *PlaybackManager) GetSession(ctx context.Context, sessionID string) (*domain.PlaybackResponse, error) {
	var session domain.PlaybackResponse
	if err := m.cacheRepo.GetSession(ctx, sessionID, &session); err != nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return &session, nil
}

// ExtendSession extends the TTL of a playback session
func (m *PlaybackManager) ExtendSession(ctx context.Context, sessionID string, duration time.Duration) error {
	// Verify session exists first
	var session domain.PlaybackResponse
	if err := m.cacheRepo.GetSession(ctx, sessionID, &session); err != nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if err := m.cacheRepo.ExtendSessionTTL(ctx, sessionID, duration); err != nil {
		return fmt.Errorf("failed to extend session: %w", err)
	}

	m.logger.Info().
		Str("session_id", sessionID).
		Dur("duration", duration).
		Msg("Extended session TTL")

	return nil
}
