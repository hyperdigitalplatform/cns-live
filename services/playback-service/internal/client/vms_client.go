package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// VMSClient handles communication with the VMS Service
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

// Camera represents a camera from VMS
type Camera struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

// GetCamera retrieves camera details from VMS
func (c *VMSClient) GetCamera(ctx context.Context, cameraID string) (*Camera, error) {
	endpoint := fmt.Sprintf("%s/api/v1/vms/cameras/%s", c.baseURL, cameraID)

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

	var camera Camera
	if err := json.NewDecoder(resp.Body).Decode(&camera); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &camera, nil
}

// GetRTSPURL retrieves the RTSP URL for live streaming
func (c *VMSClient) GetRTSPURL(ctx context.Context, cameraID string) (string, error) {
	endpoint := fmt.Sprintf("%s/api/v1/vms/cameras/%s/rtsp", c.baseURL, cameraID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get RTSP URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vms service returned status %d", resp.StatusCode)
	}

	var result struct {
		URL string `json:"rtsp_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.URL, nil
}
