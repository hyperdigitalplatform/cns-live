package soap

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
)

// Camera represents a Milestone camera device
type Camera struct {
	Id               string
	Name             string
	Enabled          bool
	Status           string // "ONLINE", "OFFLINE", "MAINTENANCE", "ERROR"
	RecordingServer  string
	Channel          int
	ShortName        string
	LiveStreamUrl    string
	RtspUrl          string // Constructed RTSP URL for streaming
	PtzEnabled       bool   // Whether PTZ is enabled for this camera
	PtzCapabilities  PtzCapabilities
}

// PtzCapabilities represents PTZ capabilities of a camera
type PtzCapabilities struct {
	Pan  bool
	Tilt bool
	Zoom bool
}

// GetItemsRequest represents a GetItems SOAP request
type GetItemsRequest struct {
	XMLName xml.Name `xml:"http://videoos.net/2/XProtectCSConfiguration GetItems"`
	Token   string   `xml:"token"`
	ItemTypes string `xml:"itemTypes"`
}

// GetItemsResponse represents a GetItems SOAP response
type GetItemsResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		GetItemsResult struct {
			Items []ConfigurationItem `xml:"Item"`
		} `xml:"GetItemsResponse>GetItemsResult"`
	} `xml:"Body"`
}

// ConfigurationItem represents a configuration item from Milestone
type ConfigurationItem struct {
	Type       string     `xml:"Type,attr"`
	Properties []Property `xml:"Properties>Property"`
}

// Property represents a property of a configuration item
type Property struct {
	Key        string `xml:"Key,attr"`
	Value      string `xml:"Value,attr"`
	ValueType  string `xml:"ValueType,attr"`
}

// GetCameras retrieves all cameras from Milestone
func (c *Client) GetCameras(ctx context.Context) ([]Camera, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, err
	}

	// Construct SOAP request
	request := GetItemsRequest{
		Token:     c.GetToken(),
		ItemTypes: "Camera",
	}

	var response GetItemsResponse
	url := fmt.Sprintf("%s/ConfigurationApiService.svc", c.baseURL)

	err := c.sendSOAPRequest(ctx, url, "http://videoos.net/2/XProtectCSConfiguration/IConfigurationApiService/GetItems", request, &response)
	if err != nil {
		return nil, fmt.Errorf("GetCameras SOAP request failed: %w", err)
	}

	// Parse cameras from response
	cameras := make([]Camera, 0)
	for _, item := range response.Body.GetItemsResult.Items {
		if item.Type == "Camera" {
			camera := parseCamera(item)
			cameras = append(cameras, camera)
		}
	}

	return cameras, nil
}

// parseCamera converts a ConfigurationItem to a Camera
func parseCamera(item ConfigurationItem) Camera {
	camera := Camera{
		Enabled: true,
		Status:  "ONLINE", // Default to ONLINE
		PtzEnabled: false,
		PtzCapabilities: PtzCapabilities{
			Pan:  false,
			Tilt: false,
			Zoom: false,
		},
	}

	for _, prop := range item.Properties {
		switch prop.Key {
		case "Id":
			camera.Id = prop.Value
		case "Name":
			camera.Name = prop.Value
		case "Enabled":
			camera.Enabled = prop.Value == "true" || prop.Value == "True"
			// Set status based on enabled state
			if camera.Enabled {
				camera.Status = "ONLINE"
			} else {
				camera.Status = "OFFLINE"
			}
		case "RecordingServer":
			camera.RecordingServer = prop.Value
		case "Channel":
			// Parse channel number if needed (future use)
		case "ShortName":
			camera.ShortName = prop.Value
		case "LiveDefaultStreamId":
			camera.LiveStreamUrl = prop.Value
		case "SupportsPTZ":
			hasPtz := prop.Value == "true" || prop.Value == "True"
			camera.PtzEnabled = hasPtz
			// If PTZ is supported, assume all capabilities for now
			if hasPtz {
				camera.PtzCapabilities.Pan = true
				camera.PtzCapabilities.Tilt = true
				camera.PtzCapabilities.Zoom = true
			}
		}
	}

	// Construct RTSP URL using Milestone RTSP port and camera ID
	// Format: rtsp://<server>:554/<camera_id>
	milestoneBaseURL := os.Getenv("MILESTONE_BASE_URL")
	if milestoneBaseURL == "" {
		milestoneBaseURL = "https://192.168.1.11"
	}

	// Extract hostname from base URL (remove https://)
	hostname := milestoneBaseURL
	if len(hostname) > 8 && hostname[:8] == "https://" {
		hostname = hostname[8:]
	} else if len(hostname) > 7 && hostname[:7] == "http://" {
		hostname = hostname[7:]
	}

	// Construct RTSP URL for live streaming
	// Milestone RTSP format: rtsp://<server>:554/<camera_id>
	if camera.Id != "" {
		camera.RtspUrl = fmt.Sprintf("rtsp://%s:554/%s", hostname, camera.Id)
	}

	return camera
}
