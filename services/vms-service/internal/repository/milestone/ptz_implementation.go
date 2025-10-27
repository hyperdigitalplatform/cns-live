package milestone

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rta/cctv/vms-service/internal/domain"
)

// ExecutePTZCommand sends PTZ command to camera using multiple methods
func (r *MilestoneRepository) ExecutePTZCommand(ctx context.Context, cmd *domain.PTZCommand) error {
	camera, err := r.GetByID(ctx, cmd.CameraID)
	if err != nil {
		return err
	}

	if !camera.PTZEnabled {
		return fmt.Errorf("camera does not support PTZ: %s", cmd.CameraID)
	}

	// Validate command first
	if err := r.validatePTZCommand(cmd); err != nil {
		return err
	}

	// Try Method 1: ONVIF (most modern cameras)
	err = r.executePTZViaONVIF(camera, cmd)
	if err == nil {
		r.logger.Info().
			Str("camera_id", cmd.CameraID).
			Str("action", string(cmd.Action)).
			Str("method", "ONVIF").
			Msg("PTZ command executed successfully")
		return nil
	}
	r.logger.Debug().
		Err(err).
		Str("camera_id", cmd.CameraID).
		Msg("ONVIF method failed, trying RTSP")

	// Try Method 2: RTSP Query Parameters (generic IP cameras)
	err = r.executePTZViaRTSP(camera, cmd)
	if err == nil {
		r.logger.Info().
			Str("camera_id", cmd.CameraID).
			Str("action", string(cmd.Action)).
			Str("method", "RTSP").
			Msg("PTZ command executed successfully")
		return nil
	}
	r.logger.Debug().
		Err(err).
		Str("camera_id", cmd.CameraID).
		Msg("RTSP method failed, trying Milestone SDK")

	// Try Method 3: Milestone SDK (enterprise VMS)
	err = r.executePTZViaMilestone(camera, cmd)
	if err == nil {
		r.logger.Info().
			Str("camera_id", cmd.CameraID).
			Str("action", string(cmd.Action)).
			Str("method", "Milestone SDK").
			Msg("PTZ command executed successfully")
		return nil
	}

	// All methods failed
	return fmt.Errorf("PTZ command failed: all methods (ONVIF, RTSP, Milestone) failed for camera %s", cmd.CameraID)
}

// validatePTZCommand validates PTZ command parameters
func (r *MilestoneRepository) validatePTZCommand(cmd *domain.PTZCommand) error {
	switch cmd.Action {
	case domain.PTZActionMove:
		if cmd.Pan < -1.0 || cmd.Pan > 1.0 {
			return fmt.Errorf("invalid pan value: %f (must be -1.0 to 1.0)", cmd.Pan)
		}
		if cmd.Tilt < -1.0 || cmd.Tilt > 1.0 {
			return fmt.Errorf("invalid tilt value: %f (must be -1.0 to 1.0)", cmd.Tilt)
		}
		if cmd.Zoom < 0.0 || cmd.Zoom > 1.0 {
			return fmt.Errorf("invalid zoom value: %f (must be 0.0 to 1.0)", cmd.Zoom)
		}
	case domain.PTZActionGoToPreset:
		if cmd.Preset < 1 || cmd.Preset > 256 {
			return fmt.Errorf("invalid preset number: %d (must be 1-256)", cmd.Preset)
		}
	case domain.PTZActionStop:
		// No validation needed
	default:
		return fmt.Errorf("unsupported PTZ action: %s", cmd.Action)
	}
	return nil
}

// executePTZViaONVIF executes PTZ command using ONVIF protocol (Method 1)
func (r *MilestoneRepository) executePTZViaONVIF(camera *domain.Camera, cmd *domain.PTZCommand) error {
	// Get or create ONVIF client
	r.clientsMutex.Lock()
	client, exists := r.onvifClients[camera.ID]
	if !exists {
		var err error
		client, err = NewONVIFClient(camera.RTSPURL)
		if err != nil {
			r.clientsMutex.Unlock()
			return fmt.Errorf("failed to create ONVIF client: %w", err)
		}
		r.onvifClients[camera.ID] = client
	}
	r.clientsMutex.Unlock()

	// Execute command
	switch cmd.Action {
	case domain.PTZActionMove:
		return client.ContinuousMove(cmd.Pan, cmd.Tilt, cmd.Zoom)

	case domain.PTZActionStop:
		return client.Stop()

	case domain.PTZActionGoToPreset:
		if cmd.Preset == 1 {
			return client.GotoHomePosition()
		}
		return client.GotoPreset(fmt.Sprintf("preset_%d", cmd.Preset))

	default:
		return fmt.Errorf("unsupported PTZ action for ONVIF: %s", cmd.Action)
	}
}

// executePTZViaRTSP executes PTZ command using RTSP control URLs (Method 2)
func (r *MilestoneRepository) executePTZViaRTSP(camera *domain.Camera, cmd *domain.PTZCommand) error {
	// Parse RTSP URL to get camera IP and credentials
	u, err := url.Parse(camera.RTSPURL)
	if err != nil {
		return fmt.Errorf("failed to parse RTSP URL: %w", err)
	}

	// Build PTZ control URL (camera-specific, may need customization)
	var ptzURL string

	switch cmd.Action {
	case domain.PTZActionMove:
		// Example for common IP cameras (Hikvision, Dahua, generic)
		ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=start", u.Host)

		if cmd.Pan < 0 {
			ptzURL += "&direction=left"
		} else if cmd.Pan > 0 {
			ptzURL += "&direction=right"
		} else if cmd.Tilt > 0 {
			ptzURL += "&direction=up"
		} else if cmd.Tilt < 0 {
			ptzURL += "&direction=down"
		} else if cmd.Zoom > 0 {
			ptzURL += "&direction=zoomin"
		} else if cmd.Zoom < 0 {
			ptzURL += "&direction=zoomout"
		}

		ptzURL += fmt.Sprintf("&speed=%d", int(cmd.Speed*100))

	case domain.PTZActionStop:
		ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=stop", u.Host)

	case domain.PTZActionGoToPreset:
		if cmd.Preset == 1 {
			ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=gohome", u.Host)
		} else {
			ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=preset&id=%d", u.Host, cmd.Preset)
		}

	default:
		return fmt.Errorf("unsupported PTZ action for RTSP: %s", cmd.Action)
	}

	// Send HTTP request
	req, err := http.NewRequest("GET", ptzURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add authentication from RTSP URL
	if u.User != nil {
		username := u.User.Username()
		password, _ := u.User.Password()
		req.SetBasicAuth(username, password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("camera returned status %d", resp.StatusCode)
	}

	return nil
}

// executePTZViaMilestone executes PTZ command using Milestone SDK (Method 3)
func (r *MilestoneRepository) executePTZViaMilestone(camera *domain.Camera, cmd *domain.PTZCommand) error {
	// TODO: Implement Milestone SDK PTZ control
	// This will be implemented later when Milestone SDK is integrated
	
	r.logger.Debug().
		Str("camera_id", camera.ID).
		Str("milestone_device_id", camera.MilestoneDeviceID).
		Msg("Milestone SDK PTZ not yet implemented")

	return fmt.Errorf("Milestone SDK PTZ not yet implemented")
}
