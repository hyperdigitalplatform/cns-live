package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rs/zerolog"
)

// VMSClient handles communication with VMS Service
type VMSClient struct {
	baseURL    string
	httpClient *http.Client
	logger     zerolog.Logger
}

// NewVMSClient creates a new VMS service client
func NewVMSClient(baseURL string, logger zerolog.Logger) *VMSClient {
	return &VMSClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetCamera retrieves camera details from VMS
func (c *VMSClient) GetCamera(ctx context.Context, cameraID string) (*domain.Camera, error) {
	endpoint := fmt.Sprintf("%s/vms/cameras/%s", c.baseURL, cameraID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get camera: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("camera not found: %s", cameraID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vms service returned status %d", resp.StatusCode)
	}

	var camera domain.Camera
	if err := json.NewDecoder(resp.Body).Decode(&camera); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &camera, nil
}

// ListCameras retrieves all cameras from VMS
func (c *VMSClient) ListCameras(ctx context.Context, query domain.CameraQuery) ([]*domain.Camera, error) {
	endpoint := fmt.Sprintf("%s/vms/cameras", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if query.Source != "" {
		q.Add("source", query.Source)
	}
	if query.Status != "" {
		q.Add("status", query.Status)
	}
	if query.Search != "" {
		q.Add("search", query.Search)
	}
	if query.Limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", query.Limit))
	}
	if query.Offset > 0 {
		q.Add("offset", fmt.Sprintf("%d", query.Offset))
	}
	req.URL.RawQuery = q.Encode()

	c.logger.Debug().
		Str("endpoint", endpoint).
		Str("url", req.URL.String()).
		Msg("Requesting cameras from VMS")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to connect to VMS service")
		return nil, fmt.Errorf("failed to list cameras: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error().
			Int("status_code", resp.StatusCode).
			Str("endpoint", endpoint).
			Str("url", req.URL.String()).
			Msg("VMS service returned non-OK status")
		return nil, fmt.Errorf("vms service returned status %d", resp.StatusCode)
	}

	var response struct {
		Cameras []*domain.Camera `json:"cameras"`
		Total   int              `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Cameras, nil
}

// ControlPTZ sends PTZ control command to VMS
func (c *VMSClient) ControlPTZ(ctx context.Context, cmd domain.PTZCommand) error {
	endpoint := fmt.Sprintf("%s/vms/cameras/%s/ptz", c.baseURL, cmd.CameraID)

	reqBody, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to control PTZ: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vms service returned status %d", resp.StatusCode)
	}

	c.logger.Info().
		Str("camera_id", cmd.CameraID).
		Str("command", cmd.Command).
		Str("user_id", cmd.UserID).
		Msg("PTZ command executed")

	return nil
}
