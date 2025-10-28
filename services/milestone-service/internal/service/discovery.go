package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"milestone-service/internal/onvif"
	"milestone-service/internal/rest"
	"milestone-service/internal/types"
)

// DiscoveryService handles camera discovery
type DiscoveryService struct {
	restClient *rest.Client
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(restClient *rest.Client) *DiscoveryService {
	return &DiscoveryService{
		restClient: restClient,
	}
}

// Credentials represents ONVIF credentials to try
type Credentials struct {
	Username string
	Password string
}

// Default credentials pool for ONVIF discovery (including demo credentials)
var defaultCredentials = []Credentials{
	{"admin", "pass"},           // Common default
	{"raammohan", "Ilove123"},   // Demo credential for tp-link Tapo
	{"admin", "admin"},          // Common default
	{"admin", ""},               // Common default (blank password)
}

// DiscoverCameras discovers all cameras from Milestone with ONVIF enrichment
func (s *DiscoveryService) DiscoverCameras(ctx context.Context) ([]types.DiscoveredCamera, error) {
	log.Println("Starting camera discovery...")

	// Step 1: Get cameras from Milestone REST API
	cameras, err := s.restClient.GetCameras(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cameras from Milestone: %w", err)
	}

	log.Printf("Found %d cameras from Milestone", len(cameras))

	// Step 2: Enrich each camera with ONVIF data
	discoveredCameras := make([]types.DiscoveredCamera, 0, len(cameras))
	for _, cam := range cameras {
		log.Printf("Processing camera: %s (%s)", cam.Name, cam.ID)

		// Extract camera IP from name (format: "ModelName (IP) - Camera 1")
		cameraIP := extractIPFromName(cam.Name)

		discovered := types.DiscoveredCamera{
			MilestoneId:      cam.ID,
			Name:             cam.Name,
			DisplayName:      cam.DisplayName,
			Enabled:          cam.Enabled,
			Status:           cam.Status,
			PtzEnabled:       cam.PTZEnabled,
			RecordingEnabled: cam.RecordingEnabled,
			ShortName:        cam.ShortName,
			Device: types.DeviceInfo{
				IP:    cameraIP,
				Model: extractModelFromName(cam.Name),
			},
			Streams: []types.StreamProfile{},
			PtzCapabilities: types.PtzCapabilities{
				Pan:  cam.PTZEnabled,
				Tilt: cam.PTZEnabled,
				Zoom: cam.PTZEnabled,
			},
		}

		// Step 3: Try to get ONVIF data
		if discovered.Device.IP != "" {
			s.enrichWithONVIF(ctx, &discovered)
		} else {
			log.Printf("  No IP address found for camera %s, skipping ONVIF discovery", cam.Name)
		}

		discoveredCameras = append(discoveredCameras, discovered)
	}

	log.Printf("Camera discovery completed. Total: %d cameras", len(discoveredCameras))
	return discoveredCameras, nil
}

// enrichWithONVIF attempts to enrich camera data with ONVIF information
func (s *DiscoveryService) enrichWithONVIF(ctx context.Context, camera *types.DiscoveredCamera) {
	// Try common ONVIF ports
	ports := []int{8888, 2020, 80, 8080}

	for _, port := range ports {
		camera.Device.Port = port
		endpoint := fmt.Sprintf("http://%s:%d/onvif/device_service", camera.Device.IP, port)

		// Try different credentials
		for _, cred := range defaultCredentials {
			log.Printf("  Trying ONVIF %s with %s:%s", endpoint, cred.Username, "***")

			onvifClient := onvif.NewClient(endpoint, cred.Username, cred.Password)

			// Try to get device information
			deviceInfo, err := onvifClient.GetDeviceInformation(ctx)
			if err != nil {
				log.Printf("    ONVIF failed: %v", err)
				continue
			}

			// Success! Fill in device info
			log.Printf("  ✓ ONVIF connected with %s:%s on port %d", cred.Username, "***", port)
			camera.Device.Manufacturer = deviceInfo.Manufacturer
			camera.Device.Model = deviceInfo.Model
			camera.Device.FirmwareVersion = deviceInfo.FirmwareVersion
			camera.Device.SerialNumber = deviceInfo.SerialNumber
			camera.Device.HardwareId = deviceInfo.HardwareId
			camera.OnvifEndpoint = endpoint
			camera.OnvifUsername = cred.Username

			// Get stream profiles with RTSP URLs
			profiles, err := onvifClient.GetProfilesWithRtspUrls(ctx)
			if err != nil {
				log.Printf("  Warning: Failed to get profiles: %v", err)
				return
			}

			log.Printf("  Found %d stream profiles", len(profiles))

			// Convert to stream profiles
			for _, p := range profiles {
				streamProfile := types.StreamProfile{
					ProfileToken: p.Token,
					Name:         p.Name,
					Encoding:     p.Encoding,
					Resolution:   fmt.Sprintf("%dx%d", p.Width, p.Height),
					Width:        p.Width,
					Height:       p.Height,
					FrameRate:    p.FrameRate,
					Bitrate:      p.Bitrate,
					RtspUrl:      p.RtspUrl,
				}
				camera.Streams = append(camera.Streams, streamProfile)
				log.Printf("    - %s: %s (%s)", p.Name, p.RtspUrl, streamProfile.Resolution)
			}

			return // Success, no need to try other credentials or ports
		}
	}

	log.Printf("  ONVIF discovery failed for camera %s", camera.Name)
}

// extractIPFromURL extracts IP address from RTSP URL or hardware address
func extractIPFromURL(rtspURL string) string {
	if rtspURL == "" {
		return ""
	}

	// Parse URL
	u, err := url.Parse(rtspURL)
	if err != nil {
		return ""
	}

	// Extract host (may include port)
	host := u.Hostname()
	return host
}

// extractIPFromName extracts IP address from camera name
// Format: "ModelName (192.168.1.13) - Camera 1"
func extractIPFromName(name string) string {
	// Find IP in parentheses
	start := strings.Index(name, "(")
	end := strings.Index(name, ")")

	if start != -1 && end != -1 && end > start {
		ipCandidate := strings.TrimSpace(name[start+1 : end])
		// Basic IP validation (contains dots)
		if strings.Contains(ipCandidate, ".") {
			return ipCandidate
		}
	}

	return ""
}

// extractModelFromName attempts to extract model name from camera name
func extractModelFromName(name string) string {
	// Format usually: "ModelName (IP) - Camera 1"
	parts := strings.Split(name, "(")
	if len(parts) > 0 {
		model := strings.TrimSpace(parts[0])
		return model
	}
	return ""
}
