package soap

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"
)

// SequenceType represents a recording sequence type
type SequenceType struct {
	Id   string `xml:"Id"`
	Name string `xml:"Name"`
}

// SequencesGetTypesRequest represents the SOAP request to get sequence types
type SequencesGetTypesRequest struct {
	XMLName   xml.Name `xml:"http://videoos.net/2/XProtectCSRecorderCommand SequencesGetTypes"`
	Token     string   `xml:"token"`
	DeviceIds struct {
		Guids []string `xml:"guid"`
	} `xml:"deviceIds"`
}

// SequencesGetTypesResponse represents the SOAP response with sequence types
type SequencesGetTypesResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Result struct {
			Types []SequenceType `xml:"SequencesGetTypesResult>SequenceType"`
		} `xml:"SequencesGetTypesResponse"`
	} `xml:"Body"`
}

// SequenceEntry represents a single recording sequence entry
type SequenceEntry struct {
	TimeBegin    time.Time `xml:"TimeBegin"`
	TimeTrigged  time.Time `xml:"TimeTrigged"`
	TimeEnd      time.Time `xml:"TimeEnd"`
}

// SequencesGetRequest represents the SOAP request to get recording sequences
type SequencesGetRequest struct {
	XMLName      xml.Name  `xml:"http://videoos.net/2/XProtectCSRecorderCommand SequencesGet"`
	Token        string    `xml:"token"`
	DeviceId     string    `xml:"deviceId"`
	SequenceType string    `xml:"sequenceType,omitempty"`
	MinTime      time.Time `xml:"minTime"`
	MaxTime      time.Time `xml:"maxTime"`
	MaxCount     int       `xml:"maxCount"`
}

// SequencesGetResponse represents the SOAP response with recording sequences
type SequencesGetResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Result struct {
			Sequences []SequenceEntry `xml:"SequencesGetResult>SequenceEntry"`
		} `xml:"SequencesGetResponse"`
	} `xml:"Body"`
}

// TimeLineInformation represents timeline data for a camera
type TimeLineInformation struct {
	Count int    `xml:"Count"`
	Data  string `xml:"Data"` // Base64 encoded bitmap
}

// TimeLineInformationGetRequest represents the SOAP request to get timeline data
type TimeLineInformationGetRequest struct {
	XMLName                       xml.Name  `xml:"http://videoos.net/2/XProtectCSRecorderCommand TimeLineInformationGet"`
	Token                         string    `xml:"token"`
	DeviceId                      string    `xml:"deviceId"`
	TimeLineInformationTypes      struct {
		Guid string `xml:"http://microsoft.com/wsdl/types/ guid"`
	} `xml:"timeLineInformationTypes"`
	TimeLineInformationBeginTime  time.Time `xml:"timeLineInformationBeginTime"`
	TimeLineInformationInterval   struct {
		MicroSeconds int64 `xml:"MicroSeconds"`
	} `xml:"timeLineInformationInterval"`
	TimeLineInformationCount      int    `xml:"timeLineInformationCount"`
}

// TimeLineInformationGetResponse represents the SOAP response with timeline data
type TimeLineInformationGetResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Result struct {
			Timeline TimeLineInformation `xml:"TimeLineInformationGetResult"`
		} `xml:"TimeLineInformationGetResponse"`
	} `xml:"Body"`
}

// SequencesGetTypes retrieves available sequence types for the camera(s)
func (c *Client) SequencesGetTypes(ctx context.Context, deviceIds []string) ([]SequenceType, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	recorderURL := c.baseURL + ":7563/RecorderCommandService/RecorderCommandService.asmx"
	soapAction := "http://videoos.net/2/XProtectCSRecorderCommand/SequencesGetTypes"

	request := SequencesGetTypesRequest{
		Token: c.GetToken(),
	}
	request.DeviceIds.Guids = deviceIds

	var response SequencesGetTypesResponse
	if err := c.sendSOAPRequest(ctx, recorderURL, soapAction, request, &response); err != nil {
		return nil, fmt.Errorf("get sequence types: %w", err)
	}

	return response.Body.Result.Types, nil
}

// SequencesGet retrieves recording sequences for the specified time range
func (c *Client) SequencesGet(ctx context.Context, deviceIds []string, startTime, endTime time.Time, sequenceTypes []string) ([]SequenceEntry, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	recorderURL := c.baseURL + ":7563/RecorderCommandService/RecorderCommandService.asmx"
	soapAction := "http://videoos.net/2/XProtectCSRecorderCommand/SequencesGet"

	// Use the first device ID (Milestone API only supports one at a time)
	deviceId := deviceIds[0]

	// Use first sequence type if specified, otherwise use RecordedDataAvailable
	sequenceType := "78289503-CBF4-43be-9DC3-9F34A8B60E6D" // RecordedDataAvailable
	if len(sequenceTypes) > 0 && sequenceTypes[0] != "" {
		sequenceType = sequenceTypes[0]
	}

	request := SequencesGetRequest{
		Token:        c.GetToken(),
		DeviceId:     deviceId,
		SequenceType: sequenceType,
		MinTime:      startTime,
		MaxTime:      endTime,
		MaxCount:     1000, // Max sequences to return
	}

	var response SequencesGetResponse
	if err := c.sendSOAPRequest(ctx, recorderURL, soapAction, request, &response); err != nil {
		return nil, fmt.Errorf("get sequences: %w", err)
	}

	return response.Body.Result.Sequences, nil
}

// TimeLineInformationGet retrieves timeline data for the specified time range
// sequenceType: typically "78289503-CBF4-43be-9DC3-9F34A8B60E6D" for RecordedDataAvailable
func (c *Client) TimeLineInformationGet(ctx context.Context, deviceIds []string, startTime, endTime time.Time, sequenceType string) (*TimeLineInformation, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Use RecordedDataAvailable type if not specified
	if sequenceType == "" {
		sequenceType = "78289503-CBF4-43be-9DC3-9F34A8B60E6D"
	}

	recorderURL := c.baseURL + ":7563/RecorderCommandService/RecorderCommandService.asmx"
	soapAction := "http://videoos.net/2/XProtectCSRecorderCommand/TimeLineInformationGet"

	// Use the first device ID (Milestone API only supports one at a time)
	deviceId := deviceIds[0]

	// Calculate duration in microseconds and determine interval
	duration := endTime.Sub(startTime)
	intervalMicroseconds := int64(60000000) // 60 seconds default
	count := int(duration.Microseconds() / intervalMicroseconds)
	if count == 0 {
		count = 1
	}

	request := TimeLineInformationGetRequest{
		Token:                        c.GetToken(),
		DeviceId:                     deviceId,
		TimeLineInformationBeginTime: startTime,
		TimeLineInformationCount:     count,
	}
	request.TimeLineInformationTypes.Guid = sequenceType
	request.TimeLineInformationInterval.MicroSeconds = intervalMicroseconds

	var response TimeLineInformationGetResponse
	if err := c.sendSOAPRequest(ctx, recorderURL, soapAction, request, &response); err != nil {
		return nil, fmt.Errorf("get timeline information: %w", err)
	}

	return &response.Body.Result.Timeline, nil
}
