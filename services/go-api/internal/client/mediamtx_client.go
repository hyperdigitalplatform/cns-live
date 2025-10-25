package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

// MediaMTXClient handles communication with MediaMTX server
type MediaMTXClient struct {
	apiURL string
	logger zerolog.Logger
}

// NewMediaMTXClient creates a new MediaMTX client
func NewMediaMTXClient(apiURL string, logger zerolog.Logger) *MediaMTXClient {
	return &MediaMTXClient{
		apiURL: apiURL,
		logger: logger,
	}
}

// PathConfig represents MediaMTX path configuration
type PathConfig struct {
	Name                     string `json:"name"`
	Source                   string `json:"source"`
	SourceOnDemand           bool   `json:"sourceOnDemand"`
	SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout"`
	SourceOnDemandCloseAfter string `json:"sourceOnDemandCloseAfter"`
	RunOnDemand              string `json:"runOnDemand,omitempty"`
	RunOnDemandRestart       bool   `json:"runOnDemandRestart"`
	RunOnDemandStartTimeout  string `json:"runOnDemandStartTimeout"`
	RunOnDemandCloseAfter    string `json:"runOnDemandCloseAfter"`
}

// ConfigurePath configures a MediaMTX path to pull RTSP stream
// Tries to add the path first, if it exists then patches it
func (c *MediaMTXClient) ConfigurePath(ctx context.Context, pathName, rtspURL string) error {
	config := PathConfig{
		Name:                     pathName,
		Source:                   rtspURL,
		SourceOnDemand:           true,
		SourceOnDemandStartTimeout: "15s",
		SourceOnDemandCloseAfter: "10s",
		RunOnDemandRestart:       true,
		RunOnDemandStartTimeout:  "15s",
		RunOnDemandCloseAfter:    "10s",
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Try ADD first
	addURL := fmt.Sprintf("%s/v3/config/paths/add/%s", c.apiURL, pathName)
	req, err := http.NewRequestWithContext(ctx, "POST", addURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// If ADD succeeded, we're done
	if resp.StatusCode == http.StatusOK {
		c.logger.Info().
			Str("path", pathName).
			Str("rtsp_url", rtspURL).
			Msg("Configured MediaMTX path")
		return nil
	}

	// If path already exists (400), try PATCH instead
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusBadRequest && bytes.Contains(body, []byte("already exists")) {
		// Try PATCH
		patchURL := fmt.Sprintf("%s/v3/config/paths/patch/%s", c.apiURL, pathName)
		req, err := http.NewRequestWithContext(ctx, "PATCH", patchURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create patch request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send patch request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("MediaMTX PATCH API returned status %d: %s", resp.StatusCode, string(body))
		}

		c.logger.Info().
			Str("path", pathName).
			Str("rtsp_url", rtspURL).
			Msg("Updated MediaMTX path")
		return nil
	}

	// Other error
	return fmt.Errorf("MediaMTX API returned status %d: %s", resp.StatusCode, string(body))
}

// DeletePath removes a path configuration from MediaMTX
func (c *MediaMTXClient) DeletePath(ctx context.Context, pathName string) error {
	url := fmt.Sprintf("%s/v3/config/paths/delete/%s", c.apiURL, pathName)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MediaMTX API returned status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info().Str("path", pathName).Msg("Deleted MediaMTX path")

	return nil
}

// GetPath retrieves path information from MediaMTX
func (c *MediaMTXClient) GetPath(ctx context.Context, pathName string) (*PathInfo, error) {
	url := fmt.Sprintf("%s/v3/paths/get/%s", c.apiURL, pathName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found: %s", pathName)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MediaMTX API returned status %d: %s", resp.StatusCode, string(body))
	}

	var pathInfo PathInfo
	if err := json.NewDecoder(resp.Body).Decode(&pathInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pathInfo, nil
}

// PathInfo represents MediaMTX path information
type PathInfo struct {
	Name          string `json:"name"`
	ConfName      string `json:"confName"`
	Source        string `json:"source"`
	Ready         bool   `json:"ready"`
	ReadyTime     string `json:"readyTime,omitempty"`
	Tracks        int    `json:"tracks"`
	BytesReceived int64  `json:"bytesReceived"`
	Readers       int    `json:"readers"`
}
