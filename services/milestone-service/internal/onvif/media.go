package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
)

// Profile represents an ONVIF media profile
type Profile struct {
	Token      string
	Name       string
	Encoding   string
	Width      int
	Height     int
	Quality    int
	FrameRate  int
	Bitrate    int
	RtspUrl    string
}

// GetProfilesResponse represents the SOAP response for GetProfiles
type GetProfilesResponse struct {
	XMLName  xml.Name        `xml:"Envelope"`
	Profiles []ProfileData   `xml:"Body>GetProfilesResponse>Profiles"`
}

// ProfileData represents profile data from SOAP response
type ProfileData struct {
	Token                    string                   `xml:"token,attr"`
	Name                     string                   `xml:"Name"`
	VideoEncoderConfiguration VideoEncoderConfiguration `xml:"VideoEncoderConfiguration"`
}

// VideoEncoderConfiguration represents video encoder configuration
type VideoEncoderConfiguration struct {
	Encoding   string     `xml:"Encoding"`
	Resolution Resolution `xml:"Resolution"`
	Quality    int        `xml:"Quality"`
	RateControl RateControl `xml:"RateControl"`
}

// Resolution represents video resolution
type Resolution struct {
	Width  int `xml:"Width"`
	Height int `xml:"Height"`
}

// RateControl represents rate control settings
type RateControl struct {
	FrameRateLimit int `xml:"FrameRateLimit"`
	BitrateLimit   int `xml:"BitrateLimit"`
}

// GetStreamUriResponse represents the SOAP response for GetStreamUri
type GetStreamUriResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Uri     string   `xml:"Body>GetStreamUriResponse>MediaUri>Uri"`
}

// GetProfiles retrieves all media profiles from ONVIF camera
func (c *Client) GetProfiles(ctx context.Context) ([]Profile, error) {
	// Determine media endpoint (usually /onvif/Media)
	mediaEndpoint := strings.Replace(c.endpoint, "/device_service", "/Media", 1)

	// SOAP body for GetProfiles
	soapBody := `<trt:GetProfiles/>`

	// Send SOAP request
	responseBody, err := c.sendSOAPRequest(ctx, mediaEndpoint, "", soapBody)
	if err != nil {
		return nil, fmt.Errorf("GetProfiles request failed: %w", err)
	}

	// Parse response
	var response GetProfilesResponse
	if err := xml.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse GetProfiles response: %w", err)
	}

	// Convert to Profile structs
	profiles := make([]Profile, 0, len(response.Profiles))
	for _, p := range response.Profiles {
		profile := Profile{
			Token:     p.Token,
			Name:      p.Name,
			Encoding:  p.VideoEncoderConfiguration.Encoding,
			Width:     p.VideoEncoderConfiguration.Resolution.Width,
			Height:    p.VideoEncoderConfiguration.Resolution.Height,
			Quality:   p.VideoEncoderConfiguration.Quality,
			FrameRate: p.VideoEncoderConfiguration.RateControl.FrameRateLimit,
			Bitrate:   p.VideoEncoderConfiguration.RateControl.BitrateLimit,
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// GetStreamUri retrieves the RTSP URL for a specific profile
func (c *Client) GetStreamUri(ctx context.Context, profileToken string) (string, error) {
	// Determine media endpoint
	mediaEndpoint := strings.Replace(c.endpoint, "/device_service", "/Media", 1)

	// SOAP body for GetStreamUri
	soapBody := fmt.Sprintf(`
		<trt:GetStreamUri>
			<trt:StreamSetup>
				<tt:Stream>RTP-Unicast</tt:Stream>
				<tt:Transport><tt:Protocol>RTSP</tt:Protocol></tt:Transport>
			</trt:StreamSetup>
			<trt:ProfileToken>%s</trt:ProfileToken>
		</trt:GetStreamUri>
	`, profileToken)

	// Send SOAP request
	responseBody, err := c.sendSOAPRequest(ctx, mediaEndpoint, "", soapBody)
	if err != nil {
		return "", fmt.Errorf("GetStreamUri request failed: %w", err)
	}

	// Parse response
	var response GetStreamUriResponse
	if err := xml.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse GetStreamUri response: %w", err)
	}

	return response.Uri, nil
}

// GetProfilesWithRtspUrls retrieves all profiles with their RTSP URLs
func (c *Client) GetProfilesWithRtspUrls(ctx context.Context) ([]Profile, error) {
	// Get all profiles
	profiles, err := c.GetProfiles(ctx)
	if err != nil {
		return nil, err
	}

	// Get RTSP URL for each profile
	for i := range profiles {
		rtspUrl, err := c.GetStreamUri(ctx, profiles[i].Token)
		if err != nil {
			// Log error but continue with other profiles
			continue
		}
		profiles[i].RtspUrl = rtspUrl
	}

	return profiles, nil
}
