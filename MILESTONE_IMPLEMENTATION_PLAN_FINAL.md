# Milestone XProtect Integration - FINAL Implementation Plan

**Date:** 2025-10-27
**Server:** 192.168.1.11
**Status:** ✅ ALL SOAP APIs VERIFIED AND WORKING

---

## Executive Summary

All required Milestone XProtect SOAP APIs have been successfully verified and are working:

✅ **Authentication** - ServerCommandService Login with HTTP Basic Auth
✅ **Manual Recording Control** - StartManualRecording, StopManualRecording
✅ **Recording Status** - IsManualRecording
✅ **Recording Sequences** - SequencesGetTypes, SequencesGet
✅ **Timeline Data** - TimeLineInformationGet

**Implementation Approach:** Create a REST API facade in our Go backend that calls Milestone SOAP services internally.

---

## Architecture Overview

```
┌─────────────────┐
│  React Frontend │
│   (Dashboard)   │
└────────┬────────┘
         │ REST API calls
         │
         ▼
┌─────────────────────────────────────────┐
│          Kong API Gateway                │
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│      Go Backend Service                  │
│  (REST Facade for Milestone SOAP)       │
│                                          │
│  ┌────────────────────────────────┐    │
│  │  MilestoneSOAPClient           │    │
│  │  - Login()                     │    │
│  │  - StartManualRecording()      │    │
│  │  - StopManualRecording()       │    │
│  │  - IsManualRecording()         │    │
│  │  - GetSequenceTypes()          │    │
│  │  - GetSequences()              │    │
│  │  - GetTimelineData()           │    │
│  └────────────────────────────────┘    │
└────────┬────────────────────────────────┘
         │ SOAP calls
         │
         ▼
┌─────────────────────────────────────────┐
│    Milestone XProtect Server             │
│    192.168.1.11                          │
│                                          │
│  ServerCommandService (Port 443)        │
│  - Login (Basic Auth)                   │
│                                          │
│  RecorderCommandService (Port 7563)     │
│  - StartManualRecording                 │
│  - StopManualRecording                  │
│  - IsManualRecording                    │
│  - SequencesGetTypes                    │
│  - SequencesGet                         │
│  - TimeLineInformationGet               │
└─────────────────────────────────────────┘
```

---

## Phase 1: Go SOAP Client Implementation (3-4 days)

### 1.1 Create SOAP Client Package Structure

**File:** `services/milestone-service/internal/soap/client.go`

```go
package soap

import (
    "bytes"
    "context"
    "crypto/tls"
    "encoding/xml"
    "fmt"
    "io"
    "net/http"
    "sync"
    "time"
)

type Client struct {
    baseURL        string
    username       string
    password       string
    token          string
    tokenExpiry    time.Time
    httpClient     *http.Client
    mu             sync.RWMutex
}

// SOAP envelope structures
type Envelope struct {
    XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
    Body    Body
}

type Body struct {
    XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
    Content interface{}
}

type SOAPFault struct {
    XMLName     xml.Name `xml:"Fault"`
    FaultCode   string   `xml:"faultcode"`
    FaultString string   `xml:"faultstring"`
    Detail      struct {
        ErrorNumber    int `xml:"ErrorNumber"`
        SubErrorNumber int `xml:"SubErrorNumber"`
    } `xml:"detail"`
}
```

### 1.2 Implement Authentication

**File:** `services/milestone-service/internal/soap/auth.go`

```go
package soap

// Login request/response structures
type LoginRequest struct {
    XMLName    xml.Name `xml:"http://videoos.net/2/XProtectCSServerCommand Login"`
    InstanceID string   `xml:"instanceId"`
}

type LoginResponse struct {
    XMLName     xml.Name `xml:"LoginResponse"`
    LoginResult struct {
        RegistrationTime time.Time `xml:"RegistrationTime"`
        TimeToLive       struct {
            MicroSeconds int64 `xml:"MicroSeconds"`
        } `xml:"TimeToLive"`
        TimeToLiveLimited bool   `xml:"TimeToLiveLimited"`
        Token             string `xml:"Token"`
    } `xml:"LoginResult"`
}

func (c *Client) Login(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Check if token is still valid
    if c.token != "" && time.Now().Before(c.tokenExpiry) {
        return nil
    }

    soapRequest := LoginRequest{
        InstanceID: "00000000-0000-0000-0000-000000000000",
    }

    envelope := Envelope{
        Body: Body{Content: soapRequest},
    }

    xmlData, err := xml.Marshal(envelope)
    if err != nil {
        return fmt.Errorf("marshal login request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/ManagementServer/ServerCommandService.svc",
        bytes.NewReader(append([]byte(xml.Header), xmlData...)))
    if err != nil {
        return err
    }

    req.SetBasicAuth(c.username, c.password)
    req.Header.Set("Content-Type", "text/xml; charset=utf-8")
    req.Header.Set("SOAPAction", "http://videoos.net/2/XProtectCSServerCommand/IServerCommandService/Login")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }

    var respEnvelope struct {
        Body struct {
            LoginResponse LoginResponse
            Fault         SOAPFault
        } `xml:"Body"`
    }

    if err := xml.Unmarshal(body, &respEnvelope); err != nil {
        return err
    }

    if respEnvelope.Body.Fault.FaultString != "" {
        return fmt.Errorf("SOAP fault: %s (error %d:%d)",
            respEnvelope.Body.Fault.FaultString,
            respEnvelope.Body.Fault.Detail.ErrorNumber,
            respEnvelope.Body.Fault.Detail.SubErrorNumber)
    }

    c.token = respEnvelope.Body.LoginResponse.LoginResult.Token
    ttlSeconds := respEnvelope.Body.LoginResponse.LoginResult.TimeToLive.MicroSeconds / 1000000
    c.tokenExpiry = time.Now().Add(time.Duration(ttlSeconds) * time.Second)

    return nil
}

func (c *Client) ensureAuthenticated(ctx context.Context) error {
    c.mu.RLock()
    needsAuth := c.token == "" || time.Now().After(c.tokenExpiry.Add(-5*time.Minute))
    c.mu.RUnlock()

    if needsAuth {
        return c.Login(ctx)
    }
    return nil
}
```

### 1.3 Implement Manual Recording Methods

**File:** `services/milestone-service/internal/soap/recording.go`

```go
package soap

// StartManualRecording structures
type StartManualRecordingRequest struct {
    XMLName   xml.Name `xml:"http://videoos.net/2/XProtectCSRecorderCommand StartManualRecording"`
    Token     string   `xml:"token"`
    DeviceIDs struct {
        XMLName xml.Name `xml:"deviceIds"`
        GUIDs   []string `xml:"http://microsoft.com/wsdl/types/ guid"`
    }
}

type ManualRecordingResult struct {
    DeviceID   string `xml:"DeviceId"`
    ResultCode int    `xml:"ResultCode"`
    Message    string `xml:"Message"`
}

type StartManualRecordingResponse struct {
    XMLName xml.Name `xml:"StartManualRecordingResponse"`
    Result  struct {
        Results []ManualRecordingResult `xml:"ManualRecordingResult"`
    } `xml:"StartManualRecordingResult"`
}

func (c *Client) StartManualRecording(ctx context.Context, deviceIDs []string) ([]ManualRecordingResult, error) {
    if err := c.ensureAuthenticated(ctx); err != nil {
        return nil, err
    }

    c.mu.RLock()
    token := c.token
    c.mu.RUnlock()

    req := StartManualRecordingRequest{
        Token: token,
    }
    req.DeviceIDs.GUIDs = deviceIDs

    envelope := Envelope{
        Body: Body{Content: req},
    }

    xmlData, err := xml.Marshal(envelope)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+":7563/RecorderCommandService/RecorderCommandService.asmx",
        bytes.NewReader(append([]byte(xml.Header), xmlData...)))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
    httpReq.Header.Set("SOAPAction", "http://videoos.net/2/XProtectCSRecorderCommand/StartManualRecording")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var respEnvelope struct {
        Body struct {
            Response StartManualRecordingResponse
            Fault    SOAPFault
        } `xml:"Body"`
    }

    if err := xml.Unmarshal(body, &respEnvelope); err != nil {
        return nil, err
    }

    if respEnvelope.Body.Fault.FaultString != "" {
        return nil, fmt.Errorf("SOAP fault: %s", respEnvelope.Body.Fault.FaultString)
    }

    return respEnvelope.Body.Response.Result.Results, nil
}

// StopManualRecording - similar structure to StartManualRecording
func (c *Client) StopManualRecording(ctx context.Context, deviceIDs []string) ([]ManualRecordingResult, error) {
    // Implementation similar to StartManualRecording
    // Just change SOAPAction and request/response types
    // ...
}

// IsManualRecording structures
type IsManualRecordingRequest struct {
    XMLName   xml.Name `xml:"http://videoos.net/2/XProtectCSRecorderCommand IsManualRecording"`
    Token     string   `xml:"token"`
    DeviceIDs struct {
        XMLName xml.Name `xml:"deviceIds"`
        GUIDs   []string `xml:"http://microsoft.com/wsdl/types/ guid"`
    }
}

type ManualRecordingInfo struct {
    DeviceID          string `xml:"DeviceId"`
    IsManualRecording bool   `xml:"IsManualRecording"`
}

type IsManualRecordingResponse struct {
    XMLName xml.Name `xml:"IsManualRecordingResponse"`
    Result  struct {
        Infos []ManualRecordingInfo `xml:"ManualRecordingInfo"`
    } `xml:"IsManualRecordingResult"`
}

func (c *Client) IsManualRecording(ctx context.Context, deviceIDs []string) (map[string]bool, error) {
    if err := c.ensureAuthenticated(ctx); err != nil {
        return nil, err
    }

    // Implementation similar to StartManualRecording
    // Returns map[deviceID]isRecording
    // ...
}
```

### 1.4 Implement Timeline and Sequence Methods

**File:** `services/milestone-service/internal/soap/timeline.go`

```go
package soap

// SequenceType structures
type SequenceType struct {
    ID   string `xml:"Id"`
    Name string `xml:"Name"`
}

type SequencesGetTypesResponse struct {
    XMLName xml.Name `xml:"SequencesGetTypesResponse"`
    Result  struct {
        Types []SequenceType `xml:"SequenceType"`
    } `xml:"SequencesGetTypesResult"`
}

func (c *Client) GetSequenceTypes(ctx context.Context, deviceID string) ([]SequenceType, error) {
    // Implementation...
}

// SequenceEntry structures
type SequenceEntry struct {
    TimeBegin    time.Time `xml:"TimeBegin"`
    TimeTriggered time.Time `xml:"TimeTrigged"`
    TimeEnd      time.Time `xml:"TimeEnd"`
}

type SequencesGetResponse struct {
    XMLName xml.Name `xml:"SequencesGetResponse"`
    Result  struct {
        Entries []SequenceEntry `xml:"SequenceEntry"`
    } `xml:"SequencesGetResult"`
}

func (c *Client) GetSequences(ctx context.Context, deviceID, sequenceTypeID string, minTime, maxTime time.Time, maxCount int) ([]SequenceEntry, error) {
    // Implementation...
}

// TimeLineInformationData structures
type TimeLineInformationData struct {
    DeviceID  string    `xml:"DeviceId"`
    Type      string    `xml:"Type"`
    BeginTime time.Time `xml:"BeginTime"`
    Interval  struct {
        MicroSeconds int64 `xml:"MicroSeconds"`
    } `xml:"Interval"`
    Count int    `xml:"Count"`
    Data  string `xml:"Data"` // Base64 encoded bitmap
}

type TimeLineInformationGetResponse struct {
    XMLName xml.Name `xml:"TimeLineInformationGetResponse"`
    Result  struct {
        Data []TimeLineInformationData `xml:"TimeLineInformationData"`
    } `xml:"TimeLineInformationGetResult"`
}

func (c *Client) GetTimelineData(ctx context.Context, deviceID string, typeIDs []string, beginTime time.Time, intervalMicros int64, count int) ([]TimeLineInformationData, error) {
    // Implementation...
}
```

### 1.5 Initialize SOAP Client

**File:** `services/milestone-service/internal/soap/client.go` (continued)

```go
func NewClient(baseURL, username, password string) *Client {
    return &Client{
        baseURL:  baseURL,
        username: username,
        password: password,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: true, // For self-signed certs
                },
            },
        },
    }
}
```

---

## Phase 2: REST API Facade (2-3 days)

### 2.1 Define REST API Endpoints

**File:** `services/milestone-service/internal/api/routes.go`

```go
package api

import (
    "github.com/gin-gonic/gin"
)

func SetupMilestoneRoutes(router *gin.Engine, handler *MilestoneHandler) {
    milestone := router.Group("/api/v1/milestone")
    {
        // Manual recording control
        milestone.POST("/cameras/:id/recording/start", handler.StartRecording)
        milestone.POST("/cameras/:id/recording/stop", handler.StopRecording)
        milestone.GET("/cameras/:id/recording/status", handler.GetRecordingStatus)

        // Batch operations
        milestone.POST("/cameras/recording/start", handler.BatchStartRecording)
        milestone.POST("/cameras/recording/stop", handler.BatchStopRecording)

        // Recording sequences and timeline
        milestone.GET("/cameras/:id/sequences", handler.GetSequences)
        milestone.GET("/cameras/:id/sequence-types", handler.GetSequenceTypes)
        milestone.GET("/cameras/:id/timeline", handler.GetTimeline)
    }
}
```

### 2.2 Implement REST Handlers

**File:** `services/milestone-service/internal/api/milestone_handler.go`

```go
package api

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "your-project/services/milestone-service/internal/soap"
)

type MilestoneHandler struct {
    soapClient *soap.Client
}

func NewMilestoneHandler(soapClient *soap.Client) *MilestoneHandler {
    return &MilestoneHandler{
        soapClient: soapClient,
    }
}

// StartRecording godoc
// @Summary Start manual recording
// @Description Start manual recording for a camera (default 15 minutes)
// @Tags milestone
// @Accept json
// @Produce json
// @Param id path string true "Camera ID (GUID)"
// @Param request body StartRecordingRequest false "Recording options"
// @Success 200 {object} StartRecordingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/milestone/cameras/{id}/recording/start [post]
func (h *MilestoneHandler) StartRecording(c *gin.Context) {
    cameraID := c.Param("id")

    var req StartRecordingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Use default 15 minutes if not specified
        req.DurationMinutes = 15
    }

    results, err := h.soapClient.StartManualRecording(c.Request.Context(), []string{cameraID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: err.Error(),
        })
        return
    }

    if len(results) == 0 {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "No result returned",
        })
        return
    }

    result := results[0]
    if result.ResultCode != 0 {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: result.Message,
            Code:  result.ResultCode,
        })
        return
    }

    c.JSON(http.StatusOK, StartRecordingResponse{
        CameraID:        result.DeviceID,
        Success:         true,
        Message:         result.Message,
        DurationMinutes: req.DurationMinutes,
    })
}

// StopRecording godoc
// @Summary Stop manual recording
// @Description Stop manual recording for a camera
// @Tags milestone
// @Produce json
// @Param id path string true "Camera ID (GUID)"
// @Success 200 {object} StopRecordingResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/milestone/cameras/{id}/recording/stop [post]
func (h *MilestoneHandler) StopRecording(c *gin.Context) {
    cameraID := c.Param("id")

    results, err := h.soapClient.StopManualRecording(c.Request.Context(), []string{cameraID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: err.Error(),
        })
        return
    }

    // Handle response...
}

// GetRecordingStatus godoc
// @Summary Get recording status
// @Description Check if camera is currently in manual recording mode
// @Tags milestone
// @Produce json
// @Param id path string true "Camera ID (GUID)"
// @Success 200 {object} RecordingStatusResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/milestone/cameras/{id}/recording/status [get]
func (h *MilestoneHandler) GetRecordingStatus(c *gin.Context) {
    cameraID := c.Param("id")

    statusMap, err := h.soapClient.IsManualRecording(c.Request.Context(), []string{cameraID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: err.Error(),
        })
        return
    }

    isRecording, exists := statusMap[cameraID]
    if !exists {
        c.JSON(http.StatusNotFound, ErrorResponse{
            Error: "Camera not found",
        })
        return
    }

    c.JSON(http.StatusOK, RecordingStatusResponse{
        CameraID:    cameraID,
        IsRecording: isRecording,
    })
}

// GetTimeline godoc
// @Summary Get timeline data
// @Description Get recording timeline bitmap for playback UI
// @Tags milestone
// @Produce json
// @Param id path string true "Camera ID (GUID)"
// @Param from query string true "Start time (RFC3339)"
// @Param to query string true "End time (RFC3339)"
// @Param granularity query int false "Granularity in seconds (default 60)"
// @Success 200 {object} TimelineResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/milestone/cameras/{id}/timeline [get]
func (h *MilestoneHandler) GetTimeline(c *gin.Context) {
    cameraID := c.Param("id")

    fromStr := c.Query("from")
    toStr := c.Query("to")
    granularityStr := c.DefaultQuery("granularity", "60")

    // Parse times and call h.soapClient.GetTimelineData()
    // Convert bitmap to JSON-friendly format
    // ...
}
```

### 2.3 Define Request/Response Models

**File:** `services/milestone-service/internal/api/models.go`

```go
package api

import "time"

type StartRecordingRequest struct {
    DurationMinutes int `json:"durationMinutes" example:"15"`
}

type StartRecordingResponse struct {
    CameraID        string `json:"cameraId"`
    Success         bool   `json:"success"`
    Message         string `json:"message"`
    DurationMinutes int    `json:"durationMinutes"`
}

type StopRecordingResponse struct {
    CameraID string `json:"cameraId"`
    Success  bool   `json:"success"`
    Message  string `json:"message"`
}

type RecordingStatusResponse struct {
    CameraID    string `json:"cameraId"`
    IsRecording bool   `json:"isRecording"`
}

type SequenceType struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type SequenceEntry struct {
    TimeBegin    time.Time `json:"timeBegin"`
    TimeTriggered time.Time `json:"timeTriggered"`
    TimeEnd      time.Time `json:"timeEnd"`
}

type SequencesResponse struct {
    CameraID  string          `json:"cameraId"`
    Sequences []SequenceEntry `json:"sequences"`
}

type TimelineInterval struct {
    Time      time.Time `json:"time"`
    HasData   bool      `json:"hasData"`
    HasMotion bool      `json:"hasMotion"`
}

type TimelineResponse struct {
    CameraID    string             `json:"cameraId"`
    BeginTime   time.Time          `json:"beginTime"`
    Granularity int                `json:"granularity"` // seconds
    Intervals   []TimelineInterval `json:"intervals"`
}

type ErrorResponse struct {
    Error string `json:"error"`
    Code  int    `json:"code,omitempty"`
}
```

---

## Phase 3: Frontend Integration (2-3 days)

### 3.1 Create Milestone API Service

**File:** `dashboard/src/services/milestoneApi.ts`

```typescript
import axios from 'axios';

const API_BASE = '/api/v1/milestone';

export interface StartRecordingRequest {
  durationMinutes?: number;
}

export interface RecordingStatusResponse {
  cameraId: string;
  isRecording: boolean;
}

export interface TimelineInterval {
  time: string;
  hasData: boolean;
  hasMotion: boolean;
}

export interface TimelineResponse {
  cameraId: string;
  beginTime: string;
  granularity: number;
  intervals: TimelineInterval[];
}

export const milestoneApi = {
  // Start manual recording (default 15 min)
  startRecording: async (cameraId: string, durationMinutes: number = 15) => {
    const response = await axios.post(
      `${API_BASE}/cameras/${cameraId}/recording/start`,
      { durationMinutes }
    );
    return response.data;
  },

  // Stop manual recording
  stopRecording: async (cameraId: string) => {
    const response = await axios.post(
      `${API_BASE}/cameras/${cameraId}/recording/stop`
    );
    return response.data;
  },

  // Get recording status
  getRecordingStatus: async (cameraId: string): Promise<RecordingStatusResponse> => {
    const response = await axios.get(
      `${API_BASE}/cameras/${cameraId}/recording/status`
    );
    return response.data;
  },

  // Get timeline data for playback
  getTimeline: async (
    cameraId: string,
    from: Date,
    to: Date,
    granularitySeconds: number = 60
  ): Promise<TimelineResponse> => {
    const response = await axios.get(
      `${API_BASE}/cameras/${cameraId}/timeline`,
      {
        params: {
          from: from.toISOString(),
          to: to.toISOString(),
          granularity: granularitySeconds,
        },
      }
    );
    return response.data;
  },

  // Get recording sequences
  getSequences: async (
    cameraId: string,
    from: Date,
    to: Date,
    sequenceType?: string
  ) => {
    const response = await axios.get(
      `${API_BASE}/cameras/${cameraId}/sequences`,
      {
        params: {
          from: from.toISOString(),
          to: to.toISOString(),
          sequenceType,
        },
      }
    );
    return response.data;
  },
};
```

### 3.2 Add Recording Control to Camera UI

**File:** `dashboard/src/components/CameraRecordingControl.tsx`

```typescript
import React, { useState, useEffect } from 'react';
import { Button, Tooltip, Badge } from 'antd';
import { VideoCameraOutlined, StopOutlined } from '@ant-design/icons';
import { milestoneApi } from '../services/milestoneApi';

interface CameraRecordingControlProps {
  cameraId: string;
}

export const CameraRecordingControl: React.FC<CameraRecordingControlProps> = ({
  cameraId,
}) => {
  const [isRecording, setIsRecording] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    checkRecordingStatus();
    const interval = setInterval(checkRecordingStatus, 10000); // Check every 10s
    return () => clearInterval(interval);
  }, [cameraId]);

  const checkRecordingStatus = async () => {
    try {
      const status = await milestoneApi.getRecordingStatus(cameraId);
      setIsRecording(status.isRecording);
    } catch (error) {
      console.error('Failed to check recording status:', error);
    }
  };

  const handleStartRecording = async () => {
    setLoading(true);
    try {
      await milestoneApi.startRecording(cameraId, 15); // 15 minutes
      setIsRecording(true);
    } catch (error) {
      console.error('Failed to start recording:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleStopRecording = async () => {
    setLoading(true);
    try {
      await milestoneApi.stopRecording(cameraId);
      setIsRecording(false);
    } catch (error) {
      console.error('Failed to stop recording:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      {isRecording ? (
        <Badge status="processing" text="Recording">
          <Tooltip title="Stop recording">
            <Button
              type="primary"
              danger
              icon={<StopOutlined />}
              onClick={handleStopRecording}
              loading={loading}
            >
              Stop
            </Button>
          </Tooltip>
        </Badge>
      ) : (
        <Tooltip title="Start 15-min recording">
          <Button
            type="primary"
            icon={<VideoCameraOutlined />}
            onClick={handleStartRecording}
            loading={loading}
          >
            Record
          </Button>
        </Tooltip>
      )}
    </div>
  );
};
```

### 3.3 Timeline Visualization Component

**File:** `dashboard/src/components/TimelineViewer.tsx`

```typescript
import React, { useEffect, useState } from 'react';
import { milestoneApi, TimelineInterval } from '../services/milestoneApi';

interface TimelineViewerProps {
  cameraId: string;
  startTime: Date;
  endTime: Date;
  onTimeSelect?: (time: Date) => void;
}

export const TimelineViewer: React.FC<TimelineViewerProps> = ({
  cameraId,
  startTime,
  endTime,
  onTimeSelect,
}) => {
  const [intervals, setIntervals] = useState<TimelineInterval[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadTimeline();
  }, [cameraId, startTime, endTime]);

  const loadTimeline = async () => {
    setLoading(true);
    try {
      const response = await milestoneApi.getTimeline(
        cameraId,
        startTime,
        endTime,
        60 // 1-minute granularity
      );
      setIntervals(response.intervals);
    } catch (error) {
      console.error('Failed to load timeline:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="timeline-viewer">
      <canvas
        width={1000}
        height={50}
        ref={(canvas) => {
          if (canvas && intervals.length > 0) {
            drawTimeline(canvas, intervals);
          }
        }}
      />
    </div>
  );
};

function drawTimeline(canvas: HTMLCanvasElement, intervals: TimelineInterval[]) {
  const ctx = canvas.getContext('2d')!;
  const width = canvas.width;
  const height = canvas.height;
  const barWidth = width / intervals.length;

  ctx.clearRect(0, 0, width, height);

  intervals.forEach((interval, index) => {
    const x = index * barWidth;

    if (interval.hasData) {
      ctx.fillStyle = interval.hasMotion ? '#ff4d4f' : '#52c41a'; // Red for motion, green for data
      ctx.fillRect(x, 0, barWidth, height);
    } else {
      ctx.fillStyle = '#d9d9d9'; // Gray for no data
      ctx.fillRect(x, 0, barWidth, height);
    }
  });
}
```

---

## Phase 4: Configuration & Deployment (1 day)

### 4.1 Environment Variables

**File:** `services/milestone-service/.env`

```bash
# Milestone XProtect Configuration
MILESTONE_BASE_URL=https://192.168.1.11
MILESTONE_USERNAME=raam
MILESTONE_PASSWORD=Ilove#123

# Recording defaults
MILESTONE_DEFAULT_RECORDING_MINUTES=15

# Camera configuration
MILESTONE_CAMERA_1_ID=a8a8b9dc-3995-49ed-9b00-62caac2ce74a
MILESTONE_CAMERA_1_NAME=GUANGZHOU T18156-AF

MILESTONE_CAMERA_2_ID=d47fa4e9-8171-4cc2-a421-95a3194f6a1d
MILESTONE_CAMERA_2_NAME=tp-link Tapo C225

# Service configuration
PORT=8083
```

### 4.2 Docker Compose Service

**File:** `docker-compose.yml` (add milestone-service)

```yaml
services:
  milestone-service:
    build:
      context: ./services/milestone-service
      dockerfile: Dockerfile
    container_name: cctv-milestone-service
    environment:
      - MILESTONE_BASE_URL=${MILESTONE_BASE_URL}
      - MILESTONE_USERNAME=${MILESTONE_USERNAME}
      - MILESTONE_PASSWORD=${MILESTONE_PASSWORD}
      - PORT=8083
    ports:
      - "8083:8083"
    networks:
      - cctv-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8083/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 4.3 Kong Route Configuration

**File:** `config/kong/kong.yml` (add milestone routes)

```yaml
services:
  - name: milestone-service
    url: http://milestone-service:8083
    routes:
      - name: milestone-api
        paths:
          - /api/v1/milestone
        strip_path: false
```

---

## Testing Strategy

### Unit Tests

```go
// services/milestone-service/internal/soap/client_test.go
func TestSOAPClient_Login(t *testing.T) {
    client := soap.NewClient("https://192.168.1.11", "raam", "Ilove#123")
    err := client.Login(context.Background())
    assert.NoError(t, err)
    assert.NotEmpty(t, client.token)
}

func TestSOAPClient_StartManualRecording(t *testing.T) {
    client := soap.NewClient("https://192.168.1.11", "raam", "Ilove#123")
    results, err := client.StartManualRecording(context.Background(),
        []string{"a8a8b9dc-3995-49ed-9b00-62caac2ce74a"})
    assert.NoError(t, err)
    assert.Len(t, results, 1)
    assert.Equal(t, 0, results[0].ResultCode)
}
```

### Integration Tests

**File:** `test-milestone-integration.sh`

```bash
#!/bin/bash

BASE_URL="http://localhost:8083/api/v1/milestone"
CAMERA_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"

echo "=== Testing Milestone Integration ==="

# Test 1: Start recording
echo "1. Starting recording..."
curl -X POST "$BASE_URL/cameras/$CAMERA_ID/recording/start" \
  -H "Content-Type: application/json" \
  -d '{"durationMinutes": 15}'

# Test 2: Check status
echo "2. Checking status..."
curl "$BASE_URL/cameras/$CAMERA_ID/recording/status"

# Test 3: Get timeline
echo "3. Getting timeline..."
FROM=$(date -u -d '1 hour ago' +"%Y-%m-%dT%H:%M:%SZ")
TO=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
curl "$BASE_URL/cameras/$CAMERA_ID/timeline?from=$FROM&to=$TO&granularity=60"

# Test 4: Stop recording
echo "4. Stopping recording..."
curl -X POST "$BASE_URL/cameras/$CAMERA_ID/recording/stop"

echo "=== Tests Complete ==="
```

---

## Implementation Timeline

| Phase | Task | Estimate | Dependencies |
|-------|------|----------|--------------|
| **Phase 1** | SOAP Client Package | 1 day | - |
| | Authentication | 0.5 day | SOAP Client |
| | Manual Recording Methods | 1 day | Authentication |
| | Timeline & Sequence Methods | 1 day | Authentication |
| | **Phase 1 Total** | **3.5 days** | |
| **Phase 2** | REST API Routes | 0.5 day | Phase 1 |
| | REST Handlers | 1 day | Phase 1 |
| | Request/Response Models | 0.5 day | - |
| | **Phase 2 Total** | **2 days** | |
| **Phase 3** | Frontend API Service | 0.5 day | Phase 2 |
| | Recording Control UI | 1 day | API Service |
| | Timeline Viewer | 1 day | API Service |
| | **Phase 3 Total** | **2.5 days** | |
| **Phase 4** | Configuration | 0.5 day | All phases |
| | Docker & Kong Setup | 0.5 day | Configuration |
| | **Phase 4 Total** | **1 day** | |
| **Testing** | Unit Tests | 1 day | Phase 1 |
| | Integration Tests | 1 day | Phase 2 |
| | **Testing Total** | **2 days** | |
| **GRAND TOTAL** | | **11 days** | |

---

## Success Criteria

✅ **Authentication**
- [ ] SOAP Login successful with Basic Auth
- [ ] Token caching and auto-refresh working
- [ ] Handle token expiration gracefully

✅ **Manual Recording**
- [ ] Start recording returns success (ResultCode=0)
- [ ] Stop recording returns success
- [ ] IsManualRecording returns correct status
- [ ] Multiple cameras can be controlled simultaneously

✅ **Timeline & Sequences**
- [ ] Timeline data returns bitmap for time range
- [ ] Bitmap correctly decoded to intervals
- [ ] Sequence types retrieved (Recording, Motion)
- [ ] Sequences queried for date range

✅ **REST API**
- [ ] All endpoints return proper HTTP status codes
- [ ] Error handling for SOAP faults
- [ ] Request validation working
- [ ] API documentation (Swagger) generated

✅ **Frontend**
- [ ] Recording buttons work in live view
- [ ] Recording status indicator updates
- [ ] Timeline visualization shows recording gaps
- [ ] User can query playback by time range

✅ **Deployment**
- [ ] Service builds and runs in Docker
- [ ] Kong routes traffic correctly
- [ ] Health checks passing
- [ ] Environment variables configured

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Token expiration during operation | High | Implement auto-refresh 5 min before expiry |
| SOAP XML parsing errors | Medium | Comprehensive error handling and logging |
| Network timeouts to Milestone | Medium | Set appropriate timeouts, retry logic |
| Multiple users controlling same camera | Low | Check status before operations |
| Timeline bitmap decoding complexity | Medium | Reference Milestone docs, add unit tests |

---

## Next Steps After Approval

1. Create milestone-service directory structure
2. Implement SOAP client package (Phase 1)
3. Add unit tests for SOAP client
4. Implement REST API facade (Phase 2)
5. Test with Postman/curl
6. Implement frontend components (Phase 3)
7. End-to-end testing
8. Deploy and configure Kong routes

---

**Status:** ✅ Ready for implementation - awaiting approval

**Last Updated:** 2025-10-27
