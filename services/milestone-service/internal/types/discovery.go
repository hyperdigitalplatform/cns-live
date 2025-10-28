package types

// DiscoveredCamera represents a fully discovered camera with all details
type DiscoveredCamera struct {
	// Milestone metadata
	MilestoneId      string `json:"milestoneId"`
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	Enabled          bool   `json:"enabled"`
	Status           string `json:"status"` // "ONLINE", "OFFLINE"
	PtzEnabled       bool   `json:"ptzEnabled"`
	RecordingEnabled bool   `json:"recordingEnabled"`
	RecordingServer  string `json:"recordingServer,omitempty"`
	ShortName        string `json:"shortName"`

	// Device information (from ONVIF)
	Device DeviceInfo `json:"device"`

	// ONVIF credentials used (for reference)
	OnvifEndpoint string `json:"onvifEndpoint,omitempty"`
	OnvifUsername string `json:"onvifUsername,omitempty"`

	// Stream profiles with RTSP URLs
	Streams []StreamProfile `json:"streams"`

	// PTZ capabilities
	PtzCapabilities PtzCapabilities `json:"ptzCapabilities"`
}

// DeviceInfo represents device information from ONVIF
type DeviceInfo struct {
	IP              string `json:"ip"`
	Port            int    `json:"port"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmwareVersion"`
	SerialNumber    string `json:"serialNumber"`
	HardwareId      string `json:"hardwareId"`
}

// StreamProfile represents a video stream profile
type StreamProfile struct {
	ProfileToken string `json:"profileToken"`
	Name         string `json:"name"`
	Encoding     string `json:"encoding"`   // H264, H265, MJPEG, etc.
	Resolution   string `json:"resolution"` // e.g., "1920x1080"
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FrameRate    int    `json:"frameRate"`
	Bitrate      int    `json:"bitrate"` // in kbps
	RtspUrl      string `json:"rtspUrl"`
}

// PtzCapabilities represents PTZ capabilities
type PtzCapabilities struct {
	Pan  bool `json:"pan"`  // Pan capability
	Tilt bool `json:"tilt"` // Tilt capability
	Zoom bool `json:"zoom"` // Zoom capability
}
