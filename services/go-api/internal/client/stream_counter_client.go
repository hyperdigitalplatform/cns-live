package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// StreamCounterClient handles communication with Stream Counter Service
type StreamCounterClient struct {
	baseURL    string
	httpClient *http.Client
	logger     zerolog.Logger
}

// NewStreamCounterClient creates a new stream counter client
func NewStreamCounterClient(baseURL string, logger zerolog.Logger) *StreamCounterClient {
	return &StreamCounterClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// ReserveStreamRequest represents a stream reservation request
type ReserveStreamRequest struct {
	CameraID string `json:"camera_id"`
	Source   string `json:"source"`
	UserID   string `json:"user_id"`
	Duration int    `json:"duration"` // seconds (1 min to 2 hours)
}

// ReserveStreamResponse represents a stream reservation response
type ReserveStreamResponse struct {
	ReservationID string    `json:"reservation_id"`
	CameraID      string    `json:"camera_id"`
	UserID        string    `json:"user_id"`
	Source        string    `json:"source"`
	ExpiresAt     time.Time `json:"expires_at"`
	CurrentUsage  int       `json:"current_usage"`
	Limit         int       `json:"limit"`
}

// ReserveStream reserves a stream slot
func (c *StreamCounterClient) ReserveStream(ctx context.Context, cameraID, source, userID string) (*ReserveStreamResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stream/reserve", c.baseURL)

	reqBody := ReserveStreamRequest{
		CameraID: cameraID,
		Source:   source,
		UserID:   userID,
		Duration: 3600, // Default 1 hour
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve stream: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors (stream-counter returns 400 on limit exceeded)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errData, ok := errorResp["error"].(map[string]interface{}); ok {
				return nil, fmt.Errorf("stream reservation failed: %v", errData["message"])
			}
		}
		return nil, fmt.Errorf("stream reservation failed with status %d", resp.StatusCode)
	}

	var result ReserveStreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info().
		Str("reservation_id", result.ReservationID).
		Str("source", source).
		Int("current_usage", result.CurrentUsage).
		Int("limit", result.Limit).
		Msg("Stream reserved successfully")

	return &result, nil
}

// ReleaseStream releases a stream reservation
func (c *StreamCounterClient) ReleaseStream(ctx context.Context, reservationID string) error {
	endpoint := fmt.Sprintf("%s/api/v1/stream/release/%s", c.baseURL, reservationID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to release stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to release stream, status: %d", resp.StatusCode)
	}

	c.logger.Info().Str("reservation_id", reservationID).Msg("Stream released")
	return nil
}

// SendHeartbeat sends a heartbeat for a reservation
func (c *StreamCounterClient) SendHeartbeat(ctx context.Context, reservationID string) error {
	endpoint := fmt.Sprintf("%s/api/v1/stream/heartbeat/%s", c.baseURL, reservationID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed, status: %d", resp.StatusCode)
	}

	return nil
}

// GetStats retrieves stream statistics
func (c *StreamCounterClient) GetStats(ctx context.Context) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stream/stats", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stats, nil
}
