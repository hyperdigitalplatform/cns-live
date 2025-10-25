package client

import (
	"context"
	"fmt"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/rs/zerolog"
)

// LiveKitIngressClient handles communication with LiveKit Ingress service
type LiveKitIngressClient struct {
	apiURL    string
	apiKey    string
	apiSecret string
	logger    zerolog.Logger
}

// NewLiveKitIngressClient creates a new LiveKit Ingress client
func NewLiveKitIngressClient(apiURL, apiKey, apiSecret string, logger zerolog.Logger) *LiveKitIngressClient {
	return &LiveKitIngressClient{
		apiURL:    apiURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		logger:    logger,
	}
}

// CreateWHIPIngress creates a WHIP ingress that accepts WebRTC streams
// This approach bypasses HLS transcoding issues and provides lower latency (~450ms vs 2-4s)
func (c *LiveKitIngressClient) CreateWHIPIngress(ctx context.Context, roomName, participantName string) (*livekit.IngressInfo, error) {
	// Create ingress client
	ingressClient := lksdk.NewIngressClient(c.apiURL, c.apiKey, c.apiSecret)

	// Create WHIP ingress request
	// No transcoding needed since cameras are standardized to H.264
	req := &livekit.CreateIngressRequest{
		InputType:           livekit.IngressInput_WHIP_INPUT,
		Name:                fmt.Sprintf("whip_%s", roomName),
		RoomName:            roomName,
		ParticipantIdentity: participantName,
		ParticipantName:     participantName,
		// EnableTranscoding is false by default for WHIP (bypass transcoding)
	}

	// Create ingress
	ingressInfo, err := ingressClient.CreateIngress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create WHIP ingress: %w", err)
	}

	// Construct WHIP URL manually since LiveKit doesn't populate it
	// The livekit-ingress container exposes WHIP on port 8080
	whipURL := fmt.Sprintf("http://livekit-ingress:8080/w/%s", ingressInfo.StreamKey)
	ingressInfo.Url = whipURL

	c.logger.Info().
		Str("ingress_id", ingressInfo.IngressId).
		Str("room", roomName).
		Str("whip_url", ingressInfo.Url).
		Str("stream_key", ingressInfo.StreamKey).
		Msg("Created LiveKit WHIP Ingress")

	return ingressInfo, nil
}

// DeleteIngress deletes an ingress
func (c *LiveKitIngressClient) DeleteIngress(ctx context.Context, ingressID string) error {
	// Create ingress client
	ingressClient := lksdk.NewIngressClient(c.apiURL, c.apiKey, c.apiSecret)

	// Delete ingress
	_, err := ingressClient.DeleteIngress(ctx, &livekit.DeleteIngressRequest{
		IngressId: ingressID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete ingress: %w", err)
	}

	c.logger.Info().Str("ingress_id", ingressID).Msg("Deleted LiveKit Ingress")

	return nil
}
