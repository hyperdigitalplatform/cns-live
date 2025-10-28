package domain

import "time"

// Camera represents a camera from VMS
type Camera struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	NameAr            string                 `json:"name_ar"`
	Source            string                 `json:"source"`  // DUBAI_POLICE, METRO, BUS, OTHER
	RTSPURL           string                 `json:"rtsp_url"`
	Status            string                 `json:"status"`  // ONLINE, OFFLINE, ERROR
	PTZEnabled        bool                   `json:"ptz_enabled"`
	RecordingServer   string                 `json:"recording_server"`
	MilestoneDeviceID string                 `json:"milestone_device_id,omitempty"`
	Metadata          map[string]interface{} `json:"metadata"`
	Location          *Location              `json:"location,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// Location represents camera geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
	AddressAr string  `json:"address_ar,omitempty"`
}

// PTZCommand represents a PTZ control command
type PTZCommand struct {
	CameraID  string  `json:"camera_id" validate:"required,uuid"`
	Command   string  `json:"command" validate:"required,oneof=pan_left pan_right tilt_up tilt_down zoom_in zoom_out preset home"`
	Speed     float64 `json:"speed,omitempty"`     // 0.0 - 1.0
	PresetID  int     `json:"preset_id,omitempty"` // For preset command
	UserID    string  `json:"user_id" validate:"required"`
}

// CameraQuery represents a camera search query
type CameraQuery struct {
	Source   string `json:"source,omitempty"`
	Status   string `json:"status,omitempty"`
	Search   string `json:"search,omitempty"` // Search by name
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// ImportCameraRequest represents a request to import a discovered camera
type ImportCameraRequest struct {
	MilestoneID             string              `json:"milestoneId" validate:"required"`
	Name                    string              `json:"name" validate:"required"`
	NameAr                  string              `json:"name_ar"`
	Source                  string              `json:"source" validate:"required,oneof=DUBAI_POLICE METRO BUS OTHER"`
	Status                  string              `json:"status" validate:"required,oneof=ONLINE OFFLINE ERROR"`
	PtzEnabled              bool                `json:"ptzEnabled"`
	RecordingServer         string              `json:"recordingServer"`
	Device                  DeviceInfo          `json:"device"`
	OnvifEndpoint           string              `json:"onvifEndpoint"`
	OnvifUsername           string              `json:"onvifUsername"`
	OnvifPassword           string              `json:"onvifPassword"`            // Plain text password for RTSP/ONVIF
	OnvifPasswordEncrypted  string              `json:"onvifPasswordEncrypted"`   // Encrypted password for storage
	Streams                 []StreamProfile     `json:"streams"`
	PtzCapabilities         PtzCapabilitiesInfo `json:"ptzCapabilities"`
}

// DeviceInfo represents ONVIF device information
type DeviceInfo struct {
	IP              string `json:"ip"`
	Port            int    `json:"port"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmwareVersion"`
	SerialNumber    string `json:"serialNumber"`
	HardwareID      string `json:"hardwareId"`
}

// StreamProfile represents a video stream profile
type StreamProfile struct {
	ProfileToken string `json:"profileToken"`
	Name         string `json:"name"`
	Encoding     string `json:"encoding"`
	Resolution   string `json:"resolution"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FrameRate    int    `json:"frameRate"`
	Bitrate      int    `json:"bitrate"`
	RtspURL      string `json:"rtspUrl"`
	IsPrimary    bool   `json:"isPrimary"`
}

// PtzCapabilitiesInfo represents PTZ capabilities
type PtzCapabilitiesInfo struct {
	Pan  bool `json:"pan"`
	Tilt bool `json:"tilt"`
	Zoom bool `json:"zoom"`
}

// ImportCamerasRequest represents a batch import request
type ImportCamerasRequest struct {
	Cameras []ImportCameraRequest `json:"cameras" validate:"required,min=1"`
}

// ImportCamerasResponse represents the import response
type ImportCamerasResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}
