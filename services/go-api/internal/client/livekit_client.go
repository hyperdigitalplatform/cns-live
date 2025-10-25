package client

import (
	"context"
	"fmt"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/rs/zerolog"
)

// LiveKitClient handles communication with LiveKit server
type LiveKitClient struct {
	url       string
	apiKey    string
	apiSecret string
	logger    zerolog.Logger
}

// NewLiveKitClient creates a new LiveKit client
func NewLiveKitClient(url, apiKey, apiSecret string, logger zerolog.Logger) *LiveKitClient {
	return &LiveKitClient{
		url:       url,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		logger:    logger,
	}
}

// CreateRoom creates a new LiveKit room
func (c *LiveKitClient) CreateRoom(ctx context.Context, roomName string, maxParticipants int32) error {
	roomClient := lksdk.NewRoomServiceClient(c.url, c.apiKey, c.apiSecret)

	_, err := roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    60, // 60 seconds
		MaxParticipants: uint32(maxParticipants),
	})

	if err != nil {
		// Room might already exist, which is fine
		c.logger.Debug().Str("room", roomName).Msg("Room might already exist")
		return nil
	}

	c.logger.Info().Str("room", roomName).Msg("LiveKit room created")
	return nil
}

// DeleteRoom deletes a LiveKit room
func (c *LiveKitClient) DeleteRoom(ctx context.Context, roomName string) error {
	roomClient := lksdk.NewRoomServiceClient(c.url, c.apiKey, c.apiSecret)

	_, err := roomClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: roomName,
	})

	if err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	c.logger.Info().Str("room", roomName).Msg("LiveKit room deleted")
	return nil
}

// GenerateToken generates a LiveKit access token
func (c *LiveKitClient) GenerateToken(roomName, participantName string, canPublish bool, validFor time.Duration) (string, error) {
	at := auth.NewAccessToken(c.apiKey, c.apiSecret)

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}

	if canPublish {
		grant.CanPublish = &canPublish
	}

	at.AddGrant(grant).
		SetIdentity(participantName).
		SetValidFor(validFor)

	token, err := at.ToJWT()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	c.logger.Debug().
		Str("room", roomName).
		Str("participant", participantName).
		Msg("Generated LiveKit token")

	return token, nil
}

// ListRooms lists all active rooms
func (c *LiveKitClient) ListRooms(ctx context.Context) ([]*livekit.Room, error) {
	roomClient := lksdk.NewRoomServiceClient(c.url, c.apiKey, c.apiSecret)

	resp, err := roomClient.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	return resp.Rooms, nil
}

// GetRoomInfo gets information about a specific room
func (c *LiveKitClient) GetRoomInfo(ctx context.Context, roomName string) (*livekit.Room, error) {
	rooms, err := c.ListRooms(ctx)
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		if room.Name == roomName {
			return room, nil
		}
	}

	return nil, fmt.Errorf("room not found: %s", roomName)
}

// ListParticipants lists participants in a room
func (c *LiveKitClient) ListParticipants(ctx context.Context, roomName string) ([]*livekit.ParticipantInfo, error) {
	roomClient := lksdk.NewRoomServiceClient(c.url, c.apiKey, c.apiSecret)

	resp, err := roomClient.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: roomName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}

	return resp.Participants, nil
}

// RemoveParticipant removes a participant from a room
func (c *LiveKitClient) RemoveParticipant(ctx context.Context, roomName, participantIdentity string) error {
	roomClient := lksdk.NewRoomServiceClient(c.url, c.apiKey, c.apiSecret)

	_, err := roomClient.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: participantIdentity,
	})

	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	c.logger.Info().
		Str("room", roomName).
		Str("participant", participantIdentity).
		Msg("Removed participant from room")

	return nil
}

