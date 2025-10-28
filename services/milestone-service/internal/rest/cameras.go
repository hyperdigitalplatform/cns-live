package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// CameraResponse represents the REST API response for cameras
type CameraResponse struct {
	Array []RestCamera `json:"array"`
}

// RestCamera represents a camera from Milestone REST API
type RestCamera struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	DisplayName         string          `json:"displayName"`
	Enabled             bool            `json:"enabled"`
	Channel             int             `json:"channel"`
	Description         string          `json:"description"`
	ShortName           string          `json:"shortName"`
	PTZEnabled          bool            `json:"ptzEnabled"`
	RecordingEnabled    bool            `json:"recordingEnabled"`
	RecordingFramerate  int             `json:"recordingFramerate"`
	Relations           CameraRelations `json:"relations"`
}

// CameraRelations represents camera relationships
type CameraRelations struct {
	Parent ResourceRef `json:"parent"`
	Self   ResourceRef `json:"self"`
}

// ResourceRef represents a resource reference
type ResourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Camera represents a simplified camera for our API
type Camera struct {
	ID               string
	Name             string
	DisplayName      string
	Enabled          bool
	Status           string // "ONLINE", "OFFLINE"
	PTZEnabled       bool
	RecordingEnabled bool
	RecordingServer  string
	Channel          int
	ShortName        string
	RtspURL          string
	HardwareID       string
}

// GetCameras retrieves all cameras from Milestone REST API
func (c *Client) GetCameras(ctx context.Context) ([]Camera, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/rest/v1/cameras", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cameras: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get cameras failed with status %d: %s", resp.StatusCode, string(body))
	}

	var cameraResp CameraResponse
	if err := json.NewDecoder(resp.Body).Decode(&cameraResp); err != nil {
		return nil, fmt.Errorf("failed to decode camera response: %w", err)
	}

	// Convert REST cameras to our Camera model
	cameras := make([]Camera, 0, len(cameraResp.Array))
	for _, restCam := range cameraResp.Array {
		camera := Camera{
			ID:               restCam.ID,
			Name:             restCam.Name,
			DisplayName:      restCam.DisplayName,
			Enabled:          restCam.Enabled,
			Status:           mapStatus(restCam.Enabled),
			PTZEnabled:       restCam.PTZEnabled,
			RecordingEnabled: restCam.RecordingEnabled,
			Channel:          restCam.Channel,
			ShortName:        restCam.ShortName,
			HardwareID:       restCam.Relations.Parent.ID,
		}

		// Construct RTSP URL
		// Format: rtsp://<milestone-server>:554/<camera-id>
		camera.RtspURL = c.constructRTSPURL(restCam.ID)

		cameras = append(cameras, camera)
	}

	return cameras, nil
}

// mapStatus converts enabled boolean to status string
func mapStatus(enabled bool) string {
	if enabled {
		return "ONLINE"
	}
	return "OFFLINE"
}

// constructRTSPURL builds an RTSP URL for a camera
func (c *Client) constructRTSPURL(cameraID string) string {
	// Extract hostname from base URL
	hostname := c.baseURL
	if len(hostname) > 8 && hostname[:8] == "https://" {
		hostname = hostname[8:]
	} else if len(hostname) > 7 && hostname[:7] == "http://" {
		hostname = hostname[7:]
	}

	// Milestone RTSP format: rtsp://<server>:554/<camera-id>
	return fmt.Sprintf("rtsp://%s:554/%s", hostname, cameraID)
}
