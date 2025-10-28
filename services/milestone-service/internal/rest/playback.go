package rest

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GetPlaybackURL constructs a Milestone playback URL for a camera at a specific time
// Routes through nginx-playback proxy to handle SSL certificate issues
// Milestone supports several playback endpoints:
// 1. /RecorderApi/media/{cameraId} - HTTP media streaming
// 2. /RecorderApi/export - Create export and download
// 3. RTSP with time parameter
func (c *Client) GetPlaybackURL(ctx context.Context, cameraID string, timestamp time.Time) (string, error) {
	// Ensure we're authenticated
	if err := c.ensureAuthenticated(ctx); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Use nginx-playback proxy to handle Milestone's self-signed SSL cert
	// Format: http://localhost:8091/milestone/RecorderApi/media/<cameraId>?time=<ISO8601>&format=<format>&token=<token>

	// Build the playback URL through nginx proxy
	playbackURL := fmt.Sprintf("http://localhost:8091/milestone/RecorderApi/media/%s", cameraID)

	// Add query parameters
	params := url.Values{}
	params.Add("time", timestamp.Format(time.RFC3339))
	params.Add("format", "mp4") // Request MP4 format for better browser compatibility
	params.Add("speed", "1.0")   // Normal playback speed

	// Add authentication token to URL
	params.Add("token", c.getToken())

	fullURL := fmt.Sprintf("%s?%s", playbackURL, params.Encode())

	return fullURL, nil
}

// GetHLSPlaybackURL constructs a Milestone HLS playback URL
// Routes through nginx-playback proxy to handle SSL certificate issues
// Some Milestone versions support HLS streaming
func (c *Client) GetHLSPlaybackURL(ctx context.Context, cameraID string, timestamp time.Time) (string, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Use nginx-playback proxy to handle Milestone's self-signed SSL cert
	// Format: http://localhost:8091/milestone/RecorderApi/hls/<cameraId>/playlist.m3u8?time=<ISO8601>&token=<token>
	playbackURL := fmt.Sprintf("http://localhost:8091/milestone/RecorderApi/hls/%s/playlist.m3u8", cameraID)

	params := url.Values{}
	params.Add("time", timestamp.Format(time.RFC3339))
	params.Add("token", c.getToken())

	fullURL := fmt.Sprintf("%s?%s", playbackURL, params.Encode())

	return fullURL, nil
}

// GetRTSPPlaybackURL constructs an RTSP playback URL with time parameter
func (c *Client) GetRTSPPlaybackURL(cameraID string, timestamp time.Time) string {
	// RTSP playback URL format
	// rtsp://<server>/<cameraId>?time=<timestamp>
	// Note: RTSP doesn't support token auth in URL, requires RTSP authentication

	rtspURL := fmt.Sprintf("rtsp://%s:554/%s?time=%s",
		c.baseURL,
		cameraID,
		url.QueryEscape(timestamp.Format(time.RFC3339)))

	return rtspURL
}

// ProxyPlaybackStream fetches video stream from Milestone and streams it to the client
// This method handles authentication and proxies the stream directly
func (c *Client) ProxyPlaybackStream(ctx context.Context, w io.Writer, cameraID string, timestamp time.Time) error {
	// Ensure we're authenticated
	if err := c.ensureAuthenticated(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Build the Milestone REST API playback URL
	// Format: https://<server>/api/rest/v1/recordings/<cameraId>/playback?startTime=<ISO8601>
	playbackURL := fmt.Sprintf("%s/api/rest/v1/recordings/%s/playback", c.baseURL, cameraID)

	params := url.Values{}
	params.Add("startTime", timestamp.Format(time.RFC3339))

	fullURL := fmt.Sprintf("%s?%s", playbackURL, params.Encode())

	// Create HTTP client that accepts self-signed certificates
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 0, // No timeout for streaming
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add Bearer authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.getToken()))
	req.Header.Set("Accept", "video/*,application/octet-stream")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch playback stream: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("milestone returned status %d", resp.StatusCode)
	}

	// Stream the response to the writer
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to stream video: %w", err)
	}

	return nil
}
