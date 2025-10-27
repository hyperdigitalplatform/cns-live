package milestone

import (
	"fmt"
	"net/url"

	"github.com/use-go/onvif"
	"github.com/use-go/onvif/ptz"
)

type ONVIFClient struct {
	device *onvif.Device
}

// NewONVIFClient creates a new ONVIF client from RTSP URL
func NewONVIFClient(rtspURL string) (*ONVIFClient, error) {
	// Parse RTSP URL to get camera IP and credentials
	u, err := url.Parse(rtspURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RTSP URL: %w", err)
	}

	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	// Create ONVIF device
	// Most cameras use /onvif/device_service endpoint
	device, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    fmt.Sprintf("http://%s/onvif/device_service", u.Host),
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ONVIF device: %w", err)
	}

	return &ONVIFClient{device: device}, nil
}

// ContinuousMove sends continuous PTZ movement command
func (c *ONVIFClient) ContinuousMove(pan, tilt, zoom float64) error {
	request := ptz.ContinuousMove{
		ProfileToken: "profile_1", // Usually profile_1, may need to query
		Velocity: ptz.PTZSpeed{
			PanTilt: ptz.Vector2D{
				X: pan,
				Y: tilt,
			},
			Zoom: ptz.Vector1D{
				X: zoom,
			},
		},
	}

	_, err := c.device.CallMethod(request)
	if err != nil {
		return fmt.Errorf("ONVIF continuous move failed: %w", err)
	}
	return nil
}

// Stop sends PTZ stop command
func (c *ONVIFClient) Stop() error {
	request := ptz.Stop{
		ProfileToken: "profile_1",
		PanTilt:      true,
		Zoom:         true,
	}

	_, err := c.device.CallMethod(request)
	if err != nil {
		return fmt.Errorf("ONVIF stop failed: %w", err)
	}
	return nil
}

// GotoHomePosition sends camera to home position
func (c *ONVIFClient) GotoHomePosition() error {
	request := ptz.GotoHomePosition{
		ProfileToken: "profile_1",
	}

	_, err := c.device.CallMethod(request)
	if err != nil {
		return fmt.Errorf("ONVIF goto home failed: %w", err)
	}
	return nil
}

// GotoPreset moves camera to a preset position
func (c *ONVIFClient) GotoPreset(presetToken string) error {
	request := ptz.GotoPreset{
		ProfileToken: "profile_1",
		PresetToken:  presetToken,
	}

	_, err := c.device.CallMethod(request)
	if err != nil {
		return fmt.Errorf("ONVIF goto preset failed: %w", err)
	}
	return nil
}
