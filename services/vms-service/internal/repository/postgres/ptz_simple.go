package postgres

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rta/cctv/vms-service/internal/domain"
)

// getHTTPClientWithTimeout returns an HTTP client with timeout
func getHTTPClientWithTimeout() *http.Client {
	return &http.Client{
		Timeout: 3 * time.Second, // 3 second timeout per camera request
	}
}

// ExecutePTZCommand sends PTZ command to camera using HTTP control URLs
func (r *PostgresRepository) ExecutePTZCommand(ctx context.Context, cmd *domain.PTZCommand) error {
	// Verify camera exists and has PTZ enabled
	camera, err := r.GetByID(ctx, cmd.CameraID)
	if err != nil {
		return err
	}

	if !camera.PTZEnabled {
		return fmt.Errorf("camera %s does not support PTZ", cmd.CameraID)
	}

	// Validate command
	if err := r.validatePTZCommand(cmd); err != nil {
		return err
	}

	// Parse RTSP URL to get camera IP and credentials
	u, err := url.Parse(camera.RTSPURL)
	if err != nil {
		r.logger.Error().Err(err).Str("camera_id", cmd.CameraID).Msg("Failed to parse RTSP URL")
		return fmt.Errorf("failed to parse camera URL: %w", err)
	}

	// Try different PTZ control methods
	var lastErr error

	// Method 1: Try ONVIF SOAP protocol (TP-Link Tapo, most modern cameras)
	lastErr = r.sendPTZViaONVIFSOAP(u, cmd)
	if lastErr == nil {
		r.logger.Info().
			Str("camera_id", cmd.CameraID).
			Str("action", string(cmd.Action)).
			Str("method", "ONVIF-SOAP").
			Msg("PTZ command executed successfully")
		return nil
	}
	r.logger.Debug().Err(lastErr).Str("camera_id", cmd.CameraID).Msg("ONVIF-SOAP method failed")

	// Method 2: Try ONVIF-style HTTP endpoint
	lastErr = r.sendPTZViaONVIFHTTP(u, cmd)
	if lastErr == nil {
		r.logger.Info().
			Str("camera_id", cmd.CameraID).
			Str("action", string(cmd.Action)).
			Str("method", "ONVIF-HTTP").
			Msg("PTZ command executed successfully")
		return nil
	}
	r.logger.Debug().Err(lastErr).Str("camera_id", cmd.CameraID).Msg("ONVIF-HTTP method failed")

	// Method 3: Try generic CGI endpoint
	lastErr = r.sendPTZViaCGI(u, cmd)
	if lastErr == nil {
		r.logger.Info().
			Str("camera_id", cmd.CameraID).
			Str("action", string(cmd.Action)).
			Str("method", "CGI").
			Msg("PTZ command executed successfully")
		return nil
	}
	r.logger.Debug().Err(lastErr).Str("camera_id", cmd.CameraID).Msg("CGI method failed")

	// Method 4: Try Hikvision ISAPI
	lastErr = r.sendPTZViaISAPI(u, cmd)
	if lastErr == nil {
		r.logger.Info().
			Str("camera_id", cmd.CameraID).
			Str("action", string(cmd.Action)).
			Str("method", "ISAPI").
			Msg("PTZ command executed successfully")
		return nil
	}
	r.logger.Debug().Err(lastErr).Str("camera_id", cmd.CameraID).Msg("ISAPI method failed")

	// All methods failed
	r.logger.Warn().
		Str("camera_id", cmd.CameraID).
		Str("action", string(cmd.Action)).
		Msg("All PTZ methods failed - camera may not support PTZ control")

	return fmt.Errorf("PTZ command failed: all methods exhausted for camera %s", cmd.CameraID)
}

// validatePTZCommand validates PTZ command parameters
func (r *PostgresRepository) validatePTZCommand(cmd *domain.PTZCommand) error {
	switch cmd.Action {
	case domain.PTZActionMove:
		if cmd.Pan < -1.0 || cmd.Pan > 1.0 {
			return fmt.Errorf("invalid pan value: %f (must be -1.0 to 1.0)", cmd.Pan)
		}
		if cmd.Tilt < -1.0 || cmd.Tilt > 1.0 {
			return fmt.Errorf("invalid tilt value: %f (must be -1.0 to 1.0)", cmd.Tilt)
		}
		if cmd.Zoom < -1.0 || cmd.Zoom > 1.0 {
			return fmt.Errorf("invalid zoom value: %f (must be -1.0 to 1.0)", cmd.Zoom)
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

// sendPTZViaONVIFHTTP tries ONVIF HTTP endpoint
func (r *PostgresRepository) sendPTZViaONVIFHTTP(u *url.URL, cmd *domain.PTZCommand) error {
	// Try different ONVIF ports (2020 for TP-Link Tapo, 80 for others)
	host := u.Hostname()
	ports := []string{"2020", "80", "8080"}

	for _, port := range ports {
		ptzURL := fmt.Sprintf("http://%s:%s/onvif/ptz_service", host, port)
		if err := r.tryONVIFEndpoint(ptzURL, u, cmd); err == nil {
			return nil
		}
	}

	return fmt.Errorf("all ONVIF ports failed")
}

func (r *PostgresRepository) tryONVIFEndpoint(ptzURL string, u *url.URL, cmd *domain.PTZCommand) error {

	// Build simple HTTP request (ONVIF usually requires SOAP, but some cameras support simple HTTP)
	req, err := http.NewRequest("POST", ptzURL, nil)
	if err != nil {
		return err
	}

	// Add authentication from RTSP URL
	if u.User != nil {
		username := u.User.Username()
		password, _ := u.User.Password()
		req.SetBasicAuth(username, password)
	}

	resp, err := getHTTPClientWithTimeout().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("camera returned status %d", resp.StatusCode)
	}

	return nil
}

// sendPTZViaCGI tries generic CGI endpoint (works with many IP cameras)
func (r *PostgresRepository) sendPTZViaCGI(u *url.URL, cmd *domain.PTZCommand) error {
	var ptzURL string

	switch cmd.Action {
	case domain.PTZActionMove:
		// Determine direction
		var direction string
		if cmd.Pan < 0 {
			direction = "left"
		} else if cmd.Pan > 0 {
			direction = "right"
		} else if cmd.Tilt > 0 {
			direction = "up"
		} else if cmd.Tilt < 0 {
			direction = "down"
		} else if cmd.Zoom > 0 {
			direction = "zoomin"
		} else if cmd.Zoom < 0 {
			direction = "zoomout"
		} else {
			return fmt.Errorf("no movement specified")
		}

		speed := int(cmd.Speed * 100)
		ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=start&direction=%s&speed=%d",
			u.Host, direction, speed)

	case domain.PTZActionStop:
		ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=stop", u.Host)

	case domain.PTZActionGoToPreset:
		if cmd.Preset == 1 {
			ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=gohome", u.Host)
		} else {
			ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=preset&id=%d", u.Host, cmd.Preset)
		}

	default:
		return fmt.Errorf("unsupported action: %s", cmd.Action)
	}

	// Send HTTP request
	req, err := http.NewRequest("GET", ptzURL, nil)
	if err != nil {
		return err
	}

	// Add authentication
	if u.User != nil {
		username := u.User.Username()
		password, _ := u.User.Password()
		req.SetBasicAuth(username, password)
	}

	resp, err := getHTTPClientWithTimeout().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("camera returned status %d", resp.StatusCode)
	}

	return nil
}

// sendPTZViaISAPI tries Hikvision ISAPI endpoint
func (r *PostgresRepository) sendPTZViaISAPI(u *url.URL, cmd *domain.PTZCommand) error {
	var ptzURL string

	switch cmd.Action {
	case domain.PTZActionMove:
		// Hikvision ISAPI format
		var direction string
		if cmd.Pan < 0 {
			direction = "LEFT"
		} else if cmd.Pan > 0 {
			direction = "RIGHT"
		} else if cmd.Tilt > 0 {
			direction = "UP"
		} else if cmd.Tilt < 0 {
			direction = "DOWN"
		} else if cmd.Zoom > 0 {
			direction = "ZOOM_IN"
		} else if cmd.Zoom < 0 {
			direction = "ZOOM_OUT"
		} else {
			return fmt.Errorf("no movement specified")
		}

		speed := int(cmd.Speed * 100)
		ptzURL = fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/continuous?%s=%d",
			u.Host, direction, speed)

	case domain.PTZActionStop:
		ptzURL = fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/momentary", u.Host)

	default:
		return fmt.Errorf("unsupported action for ISAPI: %s", cmd.Action)
	}

	// Send HTTP request
	req, err := http.NewRequest("PUT", ptzURL, nil)
	if err != nil {
		return err
	}

	// Add authentication
	if u.User != nil {
		username := u.User.Username()
		password, _ := u.User.Password()
		req.SetBasicAuth(username, password)
	}

	resp, err := getHTTPClientWithTimeout().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("camera returned status %d", resp.StatusCode)
	}

	return nil
}
