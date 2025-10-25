package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/livekit/protocol/livekit"
	"github.com/rta/cctv/go-api/internal/client"
	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rta/cctv/go-api/internal/repository"
	"github.com/rs/zerolog"
)

// StreamUseCase handles stream business logic
type StreamUseCase struct {
	streamCounterClient  *client.StreamCounterClient
	vmsClient            *client.VMSClient
	livekitClient        *client.LiveKitClient
	mediaMTXClient       *client.MediaMTXClient
	livekitIngressClient *client.LiveKitIngressClient
	dockerClient         *client.DockerClient
	streamRepo           repository.StreamRepository
	livekitURL           string
	logger               zerolog.Logger
}

// NewStreamUseCase creates a new stream use case
func NewStreamUseCase(
	streamCounterClient *client.StreamCounterClient,
	vmsClient *client.VMSClient,
	livekitClient *client.LiveKitClient,
	mediaMTXClient *client.MediaMTXClient,
	livekitIngressClient *client.LiveKitIngressClient,
	dockerClient *client.DockerClient,
	streamRepo repository.StreamRepository,
	livekitURL string,
	logger zerolog.Logger,
) *StreamUseCase {
	return &StreamUseCase{
		streamCounterClient:  streamCounterClient,
		vmsClient:            vmsClient,
		livekitClient:        livekitClient,
		mediaMTXClient:       mediaMTXClient,
		livekitIngressClient: livekitIngressClient,
		dockerClient:         dockerClient,
		streamRepo:           streamRepo,
		livekitURL:           livekitURL,
		logger:               logger,
	}
}

// RequestStream handles stream request from user
func (u *StreamUseCase) RequestStream(ctx context.Context, req domain.StreamRequest) (*domain.StreamResponse, error) {
	// 1. Get camera details from VMS
	camera, err := u.vmsClient.GetCamera(ctx, req.CameraID)
	if err != nil {
		return nil, fmt.Errorf("camera not found: %w", err)
	}

	// Check if camera is online
	if camera.Status != "ONLINE" {
		return nil, fmt.Errorf("camera is not online: %s", camera.Status)
	}

	// 2. Check agency limit (Stream Counter)
	reservation, err := u.streamCounterClient.ReserveStream(ctx, req.CameraID, camera.Source, req.UserID)
	if err != nil {
		// If reservation fails, it's likely due to limit exceeded
		return nil, fmt.Errorf("failed to reserve stream: %w", err)
	}

	// 3. Configure MediaMTX to pull RTSP stream from camera
	mediaMTXPath := fmt.Sprintf("camera_%s", req.CameraID)
	err = u.mediaMTXClient.ConfigurePath(ctx, mediaMTXPath, camera.RTSPURL)
	if err != nil {
		// Rollback reservation
		u.streamCounterClient.ReleaseStream(ctx, reservation.ReservationID)
		u.logger.Error().Err(err).Str("camera_id", req.CameraID).Msg("Failed to configure MediaMTX")
		return nil, fmt.Errorf("failed to configure stream source: %w", err)
	}

	// 4. Create LiveKit room
	roomName := fmt.Sprintf("camera_%s", req.CameraID)
	err = u.livekitClient.CreateRoom(ctx, roomName, 100)
	if err != nil {
		// Rollback
		u.streamCounterClient.ReleaseStream(ctx, reservation.ReservationID)
		u.mediaMTXClient.DeletePath(ctx, mediaMTXPath)
		return nil, fmt.Errorf("failed to create LiveKit room: %w", err)
	}

	// 5. Create LiveKit WHIP Ingress (accepts WebRTC push from GStreamer)
	// This approach bypasses HLS transcoding and provides ~450ms latency vs 2-4s
	ingressInfo, err := u.livekitIngressClient.CreateWHIPIngress(
		ctx,
		roomName,
		fmt.Sprintf("camera_%s_publisher", req.CameraID),
	)
	if err != nil {
		// Rollback
		u.streamCounterClient.ReleaseStream(ctx, reservation.ReservationID)
		u.mediaMTXClient.DeletePath(ctx, mediaMTXPath)
		u.logger.Error().Err(err).Str("camera_id", req.CameraID).Msg("Failed to create LiveKit WHIP Ingress")
		return nil, fmt.Errorf("failed to create WHIP ingress: %w", err)
	}

	// 6. Start GStreamer WHIP pusher (Docker container)
	// Pulls RTSP from MediaMTX and pushes to LiveKit WHIP endpoint
	// No transcoding - just RTP repackaging for low latency
	pusherContainerName := fmt.Sprintf("whip-pusher-%s", req.CameraID)
	err = u.startWHIPPusher(ctx, pusherContainerName, camera.RTSPURL, ingressInfo.Url, ingressInfo.StreamKey)
	if err != nil {
		// Rollback
		u.streamCounterClient.ReleaseStream(ctx, reservation.ReservationID)
		u.mediaMTXClient.DeletePath(ctx, mediaMTXPath)
		u.livekitIngressClient.DeleteIngress(ctx, ingressInfo.IngressId)
		u.logger.Error().Err(err).Str("camera_id", req.CameraID).Msg("Failed to start WHIP pusher")
		return nil, fmt.Errorf("failed to start stream pusher: %w", err)
	}

	ingressID := ingressInfo.IngressId

	// 6. Generate LiveKit access token
	quality := req.Quality
	if quality == "" {
		quality = "medium"
	}

	token, err := u.livekitClient.GenerateToken(
		roomName,
		req.UserID,
		false, // viewers cannot publish
		time.Hour,
	)
	if err != nil {
		// Rollback
		u.streamCounterClient.ReleaseStream(ctx, reservation.ReservationID)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 6. Save stream reservation
	streamReservation := &domain.StreamReservation{
		ID:            reservation.ReservationID,
		CameraID:      req.CameraID,
		CameraName:    camera.Name,
		UserID:        req.UserID,
		Source:        camera.Source,
		RoomName:      roomName,
		Token:         token,
		IngressID:     ingressID,
		ReservedAt:    time.Now(),
		ExpiresAt:     time.Now().Add(time.Hour),
		LastHeartbeat: time.Now(),
	}

	if err := u.streamRepo.SaveReservation(ctx, streamReservation); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to save reservation to repository")
		// Don't fail the request, just log the warning
	}

	// 7. Audit log
	u.logger.Info().
		Str("reservation_id", reservation.ReservationID).
		Str("user_id", req.UserID).
		Str("camera_id", req.CameraID).
		Str("source", camera.Source).
		Str("ingress_id", ingressID).
		Str("rtsp_url", camera.RTSPURL).
		Msg("Stream requested successfully")

	return &domain.StreamResponse{
		ReservationID: reservation.ReservationID,
		CameraID:      req.CameraID,
		CameraName:    camera.Name,
		RoomName:      roomName,
		Token:         token,
		LiveKitURL:    u.livekitURL,
		ExpiresAt:     streamReservation.ExpiresAt,
		Quality:       quality,
	}, nil
}

// ReleaseStream releases a stream reservation
func (u *StreamUseCase) ReleaseStream(ctx context.Context, reservationID string) error {
	// Get reservation details
	reservation, err := u.streamRepo.GetReservation(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("reservation not found: %w", err)
	}

	// Stop WHIP pusher container
	pusherContainerName := fmt.Sprintf("whip-pusher-%s", reservation.CameraID)
	if err := u.dockerClient.StopWHIPPusher(ctx, pusherContainerName); err != nil {
		u.logger.Error().Err(err).Str("container", pusherContainerName).Msg("Failed to stop WHIP pusher")
		// Continue anyway
	}

	// Delete LiveKit Ingress
	if reservation.IngressID != "" {
		if err := u.livekitIngressClient.DeleteIngress(ctx, reservation.IngressID); err != nil {
			u.logger.Error().Err(err).Str("ingress_id", reservation.IngressID).Msg("Failed to delete LiveKit Ingress")
			// Continue anyway - ingress might already be deleted
		}
	}

	// Delete MediaMTX path to stop pulling RTSP stream
	mediaMTXPath := fmt.Sprintf("camera_%s", reservation.CameraID)
	if err := u.mediaMTXClient.DeletePath(ctx, mediaMTXPath); err != nil {
		u.logger.Error().Err(err).Str("path", mediaMTXPath).Msg("Failed to delete MediaMTX path")
		// Continue anyway
	}

	// Release from Stream Counter
	if err := u.streamCounterClient.ReleaseStream(ctx, reservationID); err != nil {
		u.logger.Error().Err(err).Msg("Failed to release stream from counter")
		// Continue anyway
	}

	// Remove from repository
	if err := u.streamRepo.DeleteReservation(ctx, reservationID); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to delete reservation from repository")
	}

	// Note: We don't delete the LiveKit room as other viewers might still be watching
	// Rooms auto-cleanup after empty_timeout (60s)

	u.logger.Info().
		Str("reservation_id", reservationID).
		Str("camera_id", reservation.CameraID).
		Str("user_id", reservation.UserID).
		Str("ingress_id", reservation.IngressID).
		Msg("Stream released")

	return nil
}

// SendHeartbeat updates the heartbeat for a reservation
func (u *StreamUseCase) SendHeartbeat(ctx context.Context, reservationID string) error {
	// Send heartbeat to Stream Counter
	if err := u.streamCounterClient.SendHeartbeat(ctx, reservationID); err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	// Update last heartbeat in repository
	if err := u.streamRepo.UpdateHeartbeat(ctx, reservationID); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to update heartbeat in repository")
	}

	return nil
}

// GetStreamStats retrieves real-time stream statistics
func (u *StreamUseCase) GetStreamStats(ctx context.Context) (*domain.StreamStats, error) {
	// Get stats from Stream Counter
	stats, err := u.streamCounterClient.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Get active reservations from repository
	reservations, err := u.streamRepo.GetActiveReservations(ctx)
	if err != nil {
		u.logger.Warn().Err(err).Msg("Failed to get reservations from repository")
		reservations = []*domain.StreamReservation{}
	}

	// Get LiveKit room information
	rooms, err := u.livekitClient.ListRooms(ctx)
	if err != nil {
		u.logger.Warn().Err(err).Msg("Failed to list LiveKit rooms")
		rooms = nil
	}

	// Build response
	sourceStats := make(map[string]domain.SourceStat)
	if limitsData, ok := stats["limits"].(map[string]interface{}); ok {
		for source, data := range limitsData {
			if sourceData, ok := data.(map[string]interface{}); ok {
				current := int(sourceData["current"].(float64))
				limit := int(sourceData["limit"].(float64))
				sourceStats[source] = domain.SourceStat{
					Source:       source,
					Current:      current,
					Limit:        limit,
					UsagePercent: float64(current) / float64(limit) * 100,
				}
			}
		}
	}

	// Build camera stats
	cameraStats := []domain.CameraStat{}
	for _, reservation := range reservations {
		viewerCount := 0

		// Find corresponding LiveKit room
		if rooms != nil {
			for _, room := range rooms {
				if room.Name == reservation.RoomName {
					viewerCount = int(room.NumParticipants)
					break
				}
			}
		}

		cameraStats = append(cameraStats, domain.CameraStat{
			CameraID:    reservation.CameraID,
			CameraName:  reservation.CameraName,
			ViewerCount: viewerCount,
			Source:      reservation.Source,
			ActiveSince: reservation.ReservedAt,
		})
	}

	return &domain.StreamStats{
		ActiveStreams: len(reservations),
		TotalViewers:  u.countTotalViewers(rooms),
		SourceStats:   sourceStats,
		CameraStats:   cameraStats,
		Timestamp:     time.Now(),
	}, nil
}

// countTotalViewers counts total viewers across all rooms
func (u *StreamUseCase) countTotalViewers(rooms []*livekit.Room) int {
	total := 0
	for _, room := range rooms {
		total += int(room.NumParticipants)
	}
	return total
}

// startWHIPPusher starts a GStreamer WHIP pusher container
// This spawns a separate Docker container that pulls RTSP and pushes to WHIP
func (u *StreamUseCase) startWHIPPusher(ctx context.Context, containerName, rtspURL, whipEndpoint, streamKey string) error {
	// Configure WHIP pusher container
	config := client.WHIPPusherConfig{
		ContainerName: containerName,
		RTSPURL:       rtspURL,
		WHIPEndpoint:  whipEndpoint,
		StreamKey:     streamKey,
		NetworkName:   "cns_cctv-network", // Docker Compose prefixes network names with project name
	}

	// Start the WHIP pusher container
	err := u.dockerClient.StartWHIPPusher(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to start WHIP pusher container: %w", err)
	}

	u.logger.Info().
		Str("container", containerName).
		Str("rtsp_url", rtspURL).
		Msg("Successfully started WHIP pusher container")

	return nil
}
