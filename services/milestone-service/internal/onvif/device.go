package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

// DeviceInformation represents device information from ONVIF
type DeviceInformation struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareId      string
}

// GetDeviceInformationResponse represents the SOAP response
type GetDeviceInformationResponse struct {
	XMLName         xml.Name `xml:"Envelope"`
	Manufacturer    string   `xml:"Body>GetDeviceInformationResponse>Manufacturer"`
	Model           string   `xml:"Body>GetDeviceInformationResponse>Model"`
	FirmwareVersion string   `xml:"Body>GetDeviceInformationResponse>FirmwareVersion"`
	SerialNumber    string   `xml:"Body>GetDeviceInformationResponse>SerialNumber"`
	HardwareId      string   `xml:"Body>GetDeviceInformationResponse>HardwareId"`
}

// GetDeviceInformation retrieves device information from ONVIF camera
func (c *Client) GetDeviceInformation(ctx context.Context) (*DeviceInformation, error) {
	// SOAP body for GetDeviceInformation
	soapBody := `<tds:GetDeviceInformation/>`

	// Send SOAP request
	responseBody, err := c.sendSOAPRequest(ctx, c.endpoint, "", soapBody)
	if err != nil {
		return nil, fmt.Errorf("GetDeviceInformation request failed: %w", err)
	}

	// Parse response
	var response GetDeviceInformationResponse
	if err := xml.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse GetDeviceInformation response: %w", err)
	}

	return &DeviceInformation{
		Manufacturer:    response.Manufacturer,
		Model:           response.Model,
		FirmwareVersion: response.FirmwareVersion,
		SerialNumber:    response.SerialNumber,
		HardwareId:      response.HardwareId,
	}, nil
}
