package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rs/zerolog"
)

// StorageClient handles communication with the Storage Service
type StorageClient struct {
	baseURL    string
	httpClient *http.Client
	logger     zerolog.Logger
}

// NewStorageClient creates a new storage service client
func NewStorageClient(baseURL string, logger zerolog.Logger) *StorageClient {
	return &StorageClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ListSegments retrieves segments for a camera within a time range
func (c *StorageClient) ListSegments(ctx context.Context, cameraID string, startTime, endTime time.Time) ([]*domain.Segment, error) {
	endpoint := fmt.Sprintf("%s/api/v1/storage/segments", c.baseURL)

	// Build query parameters
	params := url.Values{}
	params.Add("camera_id", cameraID)
	params.Add("start_time", startTime.Format(time.RFC3339))
	params.Add("end_time", endTime.Format(time.RFC3339))
	params.Add("limit", "1000")

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storage service returned status %d", resp.StatusCode)
	}

	var result struct {
		Segments []*domain.Segment `json:"segments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug().
		Str("camera_id", cameraID).
		Int("count", len(result.Segments)).
		Msg("Retrieved segments from storage service")

	return result.Segments, nil
}

// GetSegmentDownloadURL retrieves a presigned download URL for a segment
func (c *StorageClient) GetSegmentDownloadURL(ctx context.Context, segmentID string, expirySeconds int) (string, error) {
	endpoint := fmt.Sprintf("%s/api/v1/storage/segments/%s/download", c.baseURL, segmentID)

	params := url.Values{}
	params.Add("expiry_seconds", fmt.Sprintf("%d", expirySeconds))

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("storage service returned status %d", resp.StatusCode)
	}

	var result struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.URL, nil
}
