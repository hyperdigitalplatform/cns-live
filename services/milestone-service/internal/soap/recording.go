package soap

import (
	"context"
	"encoding/xml"
	"fmt"
)

// RecordingResult represents the common result structure for recording operations
type RecordingResult struct {
	ResultCode int    `xml:"ResultCode"`
	Message    string `xml:"Message"`
}

// StartManualRecordingRequest represents the SOAP request to start manual recording
type StartManualRecordingRequest struct {
	XMLName   xml.Name `xml:"http://videoos.net/2/XProtectCSRecorderCommand StartManualRecording"`
	Token     string   `xml:"token"`
	DeviceIds struct {
		Guids []string `xml:"guid"`
	} `xml:"deviceIds"`
	RecordingTimeInMicroseconds int64 `xml:"recordingTimeInMicroseconds,omitempty"`
}

// StartManualRecordingResponse represents the SOAP response from starting manual recording
type StartManualRecordingResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Result struct {
			RecordingResult RecordingResult `xml:"StartManualRecordingResult"`
		} `xml:"StartManualRecordingResponse"`
	} `xml:"Body"`
}

// StopManualRecordingRequest represents the SOAP request to stop manual recording
type StopManualRecordingRequest struct {
	XMLName   xml.Name `xml:"http://videoos.net/2/XProtectCSRecorderCommand StopManualRecording"`
	Token     string   `xml:"token"`
	DeviceIds struct {
		Guids []string `xml:"guid"`
	} `xml:"deviceIds"`
}

// StopManualRecordingResponse represents the SOAP response from stopping manual recording
type StopManualRecordingResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Result struct {
			RecordingResult RecordingResult `xml:"StopManualRecordingResult"`
		} `xml:"StopManualRecordingResponse"`
	} `xml:"Body"`
}

// IsManualRecordingRequest represents the SOAP request to check recording status
type IsManualRecordingRequest struct {
	XMLName   xml.Name `xml:"http://videoos.net/2/XProtectCSRecorderCommand IsManualRecording"`
	Token     string   `xml:"token"`
	DeviceIds struct {
		Guids []string `xml:"guid"`
	} `xml:"deviceIds"`
}

// IsManualRecordingResponse represents the SOAP response for recording status
type IsManualRecordingResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Result struct {
			IsRecording bool `xml:"IsManualRecordingResult"`
		} `xml:"IsManualRecordingResponse"`
	} `xml:"Body"`
}

// StartManualRecording starts manual recording for the specified camera device(s)
// durationMinutes: recording duration in minutes (default 15 if 0)
func (c *Client) StartManualRecording(ctx context.Context, deviceIds []string, durationMinutes int) error {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Default to 15 minutes if not specified
	if durationMinutes <= 0 {
		durationMinutes = 15
	}

	// Convert minutes to microseconds
	durationMicroseconds := int64(durationMinutes) * 60 * 1000000

	recorderURL := c.baseURL + ":7563/RecorderCommandService/RecorderCommandService.asmx"
	soapAction := "http://videoos.net/2/XProtectCSRecorderCommand/StartManualRecording"

	request := StartManualRecordingRequest{
		Token:                         c.GetToken(),
		RecordingTimeInMicroseconds: durationMicroseconds,
	}
	request.DeviceIds.Guids = deviceIds

	var response StartManualRecordingResponse
	if err := c.sendSOAPRequest(ctx, recorderURL, soapAction, request, &response); err != nil {
		return fmt.Errorf("start manual recording: %w", err)
	}

	result := response.Body.Result.RecordingResult
	if result.ResultCode != 0 {
		return fmt.Errorf("start recording failed: code=%d, message=%s", result.ResultCode, result.Message)
	}

	return nil
}

// StopManualRecording stops manual recording for the specified camera device(s)
func (c *Client) StopManualRecording(ctx context.Context, deviceIds []string) error {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	recorderURL := c.baseURL + ":7563/RecorderCommandService/RecorderCommandService.asmx"
	soapAction := "http://videoos.net/2/XProtectCSRecorderCommand/StopManualRecording"

	request := StopManualRecordingRequest{
		Token: c.GetToken(),
	}
	request.DeviceIds.Guids = deviceIds

	var response StopManualRecordingResponse
	if err := c.sendSOAPRequest(ctx, recorderURL, soapAction, request, &response); err != nil {
		return fmt.Errorf("stop manual recording: %w", err)
	}

	result := response.Body.Result.RecordingResult
	if result.ResultCode != 0 {
		return fmt.Errorf("stop recording failed: code=%d, message=%s", result.ResultCode, result.Message)
	}

	return nil
}

// IsManualRecording checks if manual recording is active for the specified camera device(s)
func (c *Client) IsManualRecording(ctx context.Context, deviceIds []string) (bool, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return false, fmt.Errorf("authentication failed: %w", err)
	}

	recorderURL := c.baseURL + ":7563/RecorderCommandService/RecorderCommandService.asmx"
	soapAction := "http://videoos.net/2/XProtectCSRecorderCommand/IsManualRecording"

	request := IsManualRecordingRequest{
		Token: c.GetToken(),
	}
	request.DeviceIds.Guids = deviceIds

	var response IsManualRecordingResponse
	if err := c.sendSOAPRequest(ctx, recorderURL, soapAction, request, &response); err != nil {
		return false, fmt.Errorf("check recording status: %w", err)
	}

	return response.Body.Result.IsRecording, nil
}
