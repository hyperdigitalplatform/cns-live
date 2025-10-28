# Milestone XProtect Integration - Implementation Plan

**Version:** 1.0
**Date:** 2025-10-27
**Status:** In Progress

## Executive Summary

This document outlines the comprehensive implementation plan for integrating Milestone XProtect VMS with the CCTV Management System (CNS) dashboard. The integration enables:

1. **Camera Discovery & Addition** - Discover and add cameras from Milestone XProtect to the CNS system
2. **Manual Recording Control** - Start/stop on-demand recording with configurable duration (default 15 minutes)
3. **Recording Query & Playback** - Query recorded video for specific time ranges and playback with full VCR controls
4. **Timeline Visualization** - Display recording availability on interactive timeline with gaps identification

---

## Table of Contents

1. [Current System Architecture](#current-system-architecture)
2. [Milestone XProtect APIs](#milestone-xprotect-apis)
3. [Architecture Overview](#architecture-overview)
4. [Implementation Phases](#implementation-phases)
5. [Detailed Component Specifications](#detailed-component-specifications)
6. [Data Flow Diagrams](#data-flow-diagrams)
7. [Database Schema Updates](#database-schema-updates)
8. [API Specifications](#api-specifications)
9. [Frontend Components](#frontend-components)
10. [Testing Strategy](#testing-strategy)
11. [Deployment Plan](#deployment-plan)
12. [Risk Assessment & Mitigation](#risk-assessment--mitigation)

---

## 1. Current System Architecture

### Existing Services

```
├── go-api (Port: 8080)
│   ├── Camera Management (via VMS Client)
│   ├── Stream Management (MediaMTX, LiveKit)
│   ├── Layout Management
│   └── WebSocket Hub
│
├── vms-service (Port: 8081)
│   ├── Camera CRUD Operations
│   ├── PTZ Control (ONVIF)
│   ├── Milestone Repository (Stub)
│   └── PostgreSQL Integration
│
├── playback-service (Port: 8084)
│   ├── HLS Streaming
│   ├── Storage Client (MinIO)
│   ├── Milestone Client (Stub - TODO)
│   └── FFmpeg Transmuxer
│
├── recording-service (Port: 8083)
│   ├── Continuous Recording
│   ├── FFmpeg Recorder
│   └── Storage Integration
│
└── storage-service (Port: 8085)
    ├── Segment Management
    ├── Export Jobs
    └── MinIO Integration
```

### Current Camera Model

```go
type Camera struct {
    ID                string                 // UUID
    Name              string                 // Camera name
    NameAr            string                 // Arabic name
    Source            string                 // DUBAI_POLICE, METRO, BUS, OTHER
    RTSPURL           string                 // RTSP stream URL
    Status            string                 // ONLINE, OFFLINE, ERROR
    PTZEnabled        bool                   // PTZ capability
    RecordingServer   string                 // Recording server identifier
    MilestoneDeviceID string                 // Milestone camera ID (exists!)
    Metadata          map[string]interface{} // Additional metadata
    Location          *Location              // GPS coordinates
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

**Key Finding:** `MilestoneDeviceID` field already exists in vms-service domain model!

---

## 2. Milestone XProtect APIs

### 2.1 Authentication APIs

#### Login (Basic Auth)
```http
POST /api/rest/v1/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

#### Windows Authentication (NTLM)
```http
POST /api/rest/v1/login/negotiate
Authorization: Negotiate <base64-encoded-token>
```

#### Refresh Token
```http
POST /api/rest/v1/login/refresh
Authorization: Bearer <token>
```

### 2.2 Camera Discovery APIs

#### List All Cameras
```http
GET /api/rest/v1/cameras
Authorization: Bearer <token>

Query Parameters:
  - limit: int (default: 100)
  - offset: int (default: 0)
  - enabled: bool (filter by enabled status)
  - recording: bool (filter by recording status)

Response:
{
  "cameras": [
    {
      "id": "3772df00-6ea9-4893-b4b3-6aa944f38861",
      "name": "Camera 01 - Main Entrance",
      "enabled": true,
      "recording": true,
      "recordingServer": "Recording Server 1",
      "liveStreamUrl": "rtsp://milestone:7563/...",
      "ptzCapabilities": {
        "pan": true,
        "tilt": true,
        "zoom": true
      },
      "metadata": {
        "manufacturer": "Axis",
        "model": "P3245-LVE"
      }
    }
  ],
  "total": 150
}
```

#### Get Single Camera
```http
GET /api/rest/v1/cameras/{cameraId}
Authorization: Bearer <token>

Response:
{
  "id": "3772df00-6ea9-4893-b4b3-6aa944f38861",
  "name": "Camera 01 - Main Entrance",
  "enabled": true,
  "recording": true,
  "streamUrls": {
    "live": "rtsp://...",
    "recorded": "http://..."
  }
}
```

### 2.3 Recording Control APIs

#### Start Manual Recording
```http
POST /api/rest/v1/cameras/{cameraId}/recordings/start
Authorization: Bearer <token>
Content-Type: application/json

{
  "durationSeconds": 900,
  "triggerBy": "user_dashboard",
  "description": "Manual recording triggered from CNS dashboard"
}

Response:
{
  "recordingId": "rec_abc123",
  "cameraId": "3772df00-6ea9-4893-b4b3-6aa944f38861",
  "startTime": "2025-10-27T14:30:00Z",
  "estimatedEndTime": "2025-10-27T14:45:00Z",
  "status": "recording"
}
```

#### Stop Manual Recording
```http
POST /api/rest/v1/cameras/{cameraId}/recordings/stop
Authorization: Bearer <token>
Content-Type: application/json

{
  "recordingId": "rec_abc123"
}

Response:
{
  "recordingId": "rec_abc123",
  "actualDuration": 847,
  "status": "stopped"
}
```

#### Get Recording Status
```http
GET /api/rest/v1/cameras/{cameraId}/recordings/status
Authorization: Bearer <token>

Response:
{
  "isRecording": true,
  "currentRecording": {
    "recordingId": "rec_abc123",
    "startTime": "2025-10-27T14:30:00Z",
    "elapsedSeconds": 245,
    "remainingSeconds": 655
  }
}
```

### 2.4 Recording Query APIs

#### Query Recording Sequences
```http
GET /api/rest/v1/cameras/{cameraId}/sequences
Authorization: Bearer <token>

Query Parameters:
  - startTime: ISO 8601 datetime
  - endTime: ISO 8601 datetime

Response:
{
  "cameraId": "3772df00-6ea9-4893-b4b3-6aa944f38861",
  "sequences": [
    {
      "sequenceId": "seq_001",
      "startTime": "2025-10-27T10:00:00Z",
      "endTime": "2025-10-27T11:30:00Z",
      "durationSeconds": 5400,
      "available": true,
      "sizeBytes": 2147483648
    },
    {
      "sequenceId": "seq_002",
      "startTime": "2025-10-27T12:00:00Z",
      "endTime": "2025-10-27T14:00:00Z",
      "durationSeconds": 7200,
      "available": true,
      "sizeBytes": 2863311360
    }
  ],
  "gaps": [
    {
      "startTime": "2025-10-27T11:30:00Z",
      "endTime": "2025-10-27T12:00:00Z"
    }
  ]
}
```

### 2.5 Playback APIs

#### Get Recorded Video Stream
```http
GET /api/rest/v1/cameras/{cameraId}/video
Authorization: Bearer <token>

Query Parameters:
  - time: ISO 8601 datetime (playback start time)
  - speed: float (-8, -4, -2, -1, 1, 2, 4, 8) - playback speed
  - format: string (mjpeg, h264, webrtc)

Response:
- Binary video stream (MJPEG, H.264, etc.)
- Or WebRTC signaling for WebRTC format
```

#### Get Snapshot at Timestamp
```http
GET /api/rest/v1/cameras/{cameraId}/snapshots
Authorization: Bearer <token>

Query Parameters:
  - time: ISO 8601 datetime

Response:
- Binary image (JPEG)
```

---

## 3. Architecture Overview

### 3.1 Service Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        Dashboard (React)                     │
│  ┌────────────────┐  ┌──────────────────┐  ┌──────────────┐ │
│  │ Camera         │  │ Recording        │  │ Timeline     │ │
│  │ Discovery UI   │  │ Control UI       │  │ Playback UI  │ │
│  └────────────────┘  └──────────────────┘  └──────────────┘ │
└──────────────────────────────────────────────────────────────┘
                             ↕ HTTP/WS
┌──────────────────────────────────────────────────────────────┐
│                      go-api (API Gateway)                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Camera Handler | Recording Handler | Playback Handler │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
       ↕                      ↕                     ↕
┌─────────────┐    ┌──────────────────┐    ┌────────────────┐
│ vms-service │    │ recording-service│    │ playback-      │
│             │    │                  │    │ service        │
│ ┌─────────┐ │    │ ┌──────────────┐ │    │ ┌────────────┐ │
│ │Milestone│ │    │ │ Recording    │ │    │ │ Milestone  │ │
│ │Repository│ │    │ │ Manager      │ │    │ │ Client     │ │
│ └─────────┘ │    │ └──────────────┘ │    │ └────────────┘ │
└─────────────┘    └──────────────────┘    └────────────────┘
       ↕                      ↕                     ↕
┌──────────────────────────────────────────────────────────────┐
│                   Milestone XProtect VMS                     │
│   Management Server | Recording Server | Event Server       │
└──────────────────────────────────────────────────────────────┘
```

### 3.2 Data Flow Architecture

```
1. Camera Discovery Flow:
   Dashboard → go-api → vms-service → Milestone API → PostgreSQL

2. Recording Control Flow:
   Dashboard → go-api → recording-service → Milestone API
                    ↓
              WebSocket notification to Dashboard

3. Recording Query Flow:
   Dashboard → go-api → playback-service → Milestone API
                                         → Storage (MinIO)

4. Playback Flow:
   Dashboard → go-api → playback-service → Milestone API
                                         → FFmpeg → HLS
```

---

## 4. Implementation Phases

### Phase 1: Milestone API Client Foundation (Week 1)
**Goal:** Create robust Milestone API client with authentication

**Tasks:**
1. ✅ Implement Milestone API client in vms-service
   - Authentication (Basic Auth, NTLM)
   - Session management with token refresh
   - Error handling and retry logic
   - Connection pooling

2. ✅ Implement camera discovery endpoints
   - List all Milestone cameras
   - Get single camera details
   - Sync Milestone cameras to PostgreSQL
   - Map Milestone device IDs

3. ✅ Create configuration management
   - Milestone server URL
   - Credentials (environment variables)
   - Connection timeouts
   - Retry policies

**Deliverables:**
- `services/vms-service/internal/client/milestone_client.go`
- `services/vms-service/internal/repository/milestone/milestone_repository.go`
- Configuration updates in docker-compose.yml
- Unit tests for Milestone client

### Phase 2: Camera Discovery & Management (Week 1-2)
**Goal:** Enable users to discover and add Milestone cameras

**Tasks:**
1. ✅ Implement backend API endpoints
   - `GET /api/v1/milestone/cameras` - List Milestone cameras
   - `POST /api/v1/cameras/import` - Import camera from Milestone
   - `PUT /api/v1/cameras/{id}/sync` - Sync camera with Milestone
   - `POST /api/v1/milestone/sync-all` - Bulk sync all cameras

2. ✅ Update database schema
   - Add milestone-specific fields to cameras table
   - Create milestone_sync_history table
   - Add indexes for performance

3. ✅ Implement frontend UI
   - Camera discovery modal/page
   - List available Milestone cameras
   - Bulk import functionality
   - Sync status indicators

**Deliverables:**
- Camera import API endpoints
- Database migration scripts
- Frontend camera discovery component
- Integration tests

### Phase 3: Recording Control (Week 2)
**Goal:** Enable manual recording start/stop with duration

**Tasks:**
1. ✅ Implement recording control in recording-service
   - Milestone recording API integration
   - Recording session management
   - Auto-stop scheduler
   - State persistence (Redis/PostgreSQL)

2. ✅ Create API endpoints in go-api
   - `POST /api/v1/cameras/{id}/recordings/start`
   - `POST /api/v1/cameras/{id}/recordings/stop`
   - `GET /api/v1/cameras/{id}/recordings/status`

3. ✅ Implement frontend recording controls
   - Recording control buttons in sidebar
   - Duration selector (5/15/30/60 min, custom)
   - Recording status indicator
   - Countdown timer
   - WebSocket notifications

4. ✅ Add recording state management
   - Active recordings tracking
   - Recovery on service restart
   - User permission checks

**Deliverables:**
- Recording control service
- API endpoints
- Frontend sidebar enhancements
- WebSocket event handlers

### Phase 4: Recording Query & Timeline (Week 3)
**Goal:** Query recordings and display timeline

**Tasks:**
1. ✅ Implement recording query in playback-service
   - Query sequences from Milestone
   - Identify recording gaps
   - Aggregate data for timeline UI
   - Cache results for performance

2. ✅ Create query API endpoints
   - `POST /api/v1/cameras/{id}/recordings/query`
   - `GET /api/v1/cameras/{id}/recordings/timeline`
   - `GET /api/v1/cameras/{id}/recordings/sequences`

3. ✅ Build timeline UI component
   - Interactive timeline bar
   - Recording availability visualization
   - Gap highlighting
   - Zoom controls
   - Time markers

4. ✅ Integrate timeline with sidebar
   - "View Recordings" button
   - Modal/drawer with timeline
   - Date range picker

**Deliverables:**
- Recording query service
- Timeline API endpoints
- Timeline React component
- Query interface in sidebar

### Phase 5: Playback Implementation (Week 3-4)
**Goal:** Enable video playback with VCR controls

**Tasks:**
1. ✅ Implement playback streaming
   - Milestone video stream retrieval
   - FFmpeg transmuxing to HLS
   - Speed control (-8x to 8x)
   - Seek functionality

2. ✅ Create playback API endpoints
   - `GET /api/v1/cameras/{id}/playback/stream`
   - `GET /api/v1/cameras/{id}/playback/snapshot`
   - WebSocket for playback control

3. ✅ Build playback UI
   - Video player component
   - VCR controls (play/pause/forward/backward)
   - Speed selector
   - Playhead synchronization
   - Timestamp overlay

4. ✅ Integrate playback with timeline
   - Click timeline to seek
   - Update playhead position
   - Show current timestamp

**Deliverables:**
- Playback streaming service
- Playback API endpoints
- Video player component
- VCR control UI

### Phase 6: Testing & Optimization (Week 4)
**Goal:** Ensure stability and performance

**Tasks:**
1. ✅ Unit testing
   - Milestone client tests
   - API endpoint tests
   - Frontend component tests

2. ✅ Integration testing
   - End-to-end recording workflow
   - Playback functionality
   - Error scenarios

3. ✅ Performance optimization
   - API response caching
   - Timeline data aggregation
   - Video streaming optimization
   - Connection pooling

4. ✅ Error handling & recovery
   - Network failures
   - Milestone server downtime
   - Session expiration
   - Graceful degradation

**Deliverables:**
- Test suite (unit + integration)
- Performance benchmarks
- Error handling documentation
- Optimization report

### Phase 7: Deployment & Documentation (Week 4-5)
**Goal:** Deploy to production and document

**Tasks:**
1. ✅ Update deployment configuration
   - Docker compose updates
   - Environment variables
   - Kong gateway routes
   - Health checks

2. ✅ Create documentation
   - API documentation
   - User guide
   - Admin setup guide
   - Troubleshooting guide

3. ✅ Deploy to staging
   - Test with real Milestone server
   - User acceptance testing
   - Performance monitoring

4. ✅ Production deployment
   - Blue-green deployment
   - Monitoring setup
   - Alerting configuration

**Deliverables:**
- Deployment scripts
- API documentation
- User documentation
- Production deployment

---

## 5. Detailed Component Specifications

### 5.1 Milestone Client Service

**Location:** `services/vms-service/internal/client/milestone_client.go`

```go
package client

import (
    "context"
    "time"
)

type MilestoneClient struct {
    baseURL      string
    username     string
    password     string
    token        string
    tokenExpiry  time.Time
    httpClient   *http.Client
    mu           sync.RWMutex
}

// Authentication methods
func (m *MilestoneClient) Login(ctx context.Context) error
func (m *MilestoneClient) Logout(ctx context.Context) error
func (m *MilestoneClient) RefreshToken(ctx context.Context) error
func (m *MilestoneClient) ensureAuthenticated(ctx context.Context) error

// Camera discovery methods
func (m *MilestoneClient) ListCameras(ctx context.Context, opts ListCamerasOptions) (*CameraList, error)
func (m *MilestoneClient) GetCamera(ctx context.Context, cameraID string) (*MilestoneCamera, error)

// Recording control methods
func (m *MilestoneClient) StartRecording(ctx context.Context, req StartRecordingRequest) (*RecordingSession, error)
func (m *MilestoneClient) StopRecording(ctx context.Context, cameraID, recordingID string) error
func (m *MilestoneClient) GetRecordingStatus(ctx context.Context, cameraID string) (*RecordingStatus, error)

// Recording query methods
func (m *MilestoneClient) QuerySequences(ctx context.Context, req SequenceQueryRequest) (*SequenceList, error)
func (m *MilestoneClient) GetRecordingMetadata(ctx context.Context, cameraID string, timeRange TimeRange) (*RecordingMetadata, error)

// Playback methods
func (m *MilestoneClient) GetVideoStream(ctx context.Context, req VideoStreamRequest) (io.ReadCloser, error)
func (m *MilestoneClient) GetSnapshot(ctx context.Context, cameraID string, timestamp time.Time) ([]byte, error)
```

### 5.2 Recording Manager Service

**Location:** `services/recording-service/internal/manager/milestone_recording_manager.go`

```go
package manager

type MilestoneRecordingManager struct {
    milestoneClient *client.MilestoneClient
    activeRecordings map[string]*ActiveRecording
    mu sync.RWMutex
}

type ActiveRecording struct {
    RecordingID   string
    CameraID      string
    StartTime     time.Time
    DurationSec   int
    TriggeredBy   string
    StopTimer     *time.Timer
}

func (m *MilestoneRecordingManager) StartRecording(ctx context.Context, req StartRecordingRequest) (*RecordingSession, error)
func (m *MilestoneRecordingManager) StopRecording(ctx context.Context, cameraID string) error
func (m *MilestoneRecordingManager) GetStatus(ctx context.Context, cameraID string) (*RecordingStatus, error)
func (m *MilestoneRecordingManager) scheduleAutoStop(recording *ActiveRecording)
func (m *MilestoneRecordingManager) RestoreActiveSessions(ctx context.Context) error
```

### 5.3 Playback Service Integration

**Location:** `services/playback-service/internal/usecase/milestone_playback_usecase.go`

```go
package usecase

type MilestonePlaybackUsecase struct {
    milestoneClient *client.MilestoneClient
    transmuxer      *transmux.FFmpegTransmuxer
    cache           *cache.SegmentCache
}

func (u *MilestonePlaybackUsecase) QueryRecordings(ctx context.Context, req QueryRequest) (*TimelineData, error)
func (u *MilestonePlaybackUsecase) StartPlayback(ctx context.Context, req PlaybackRequest) (*PlaybackSession, error)
func (u *MilestonePlaybackUsecase) ControlPlayback(ctx context.Context, sessionID string, cmd PlaybackCommand) error
func (u *MilestonePlaybackUsecase) GetSnapshot(ctx context.Context, cameraID string, timestamp time.Time) ([]byte, error)
```

---

## 6. Data Flow Diagrams

### 6.1 Camera Discovery Flow

```
┌─────────┐     ┌────────┐     ┌───────────┐     ┌──────────┐     ┌──────────┐
│Dashboard│     │go-api  │     │vms-service│     │Milestone │     │PostgreSQL│
└────┬────┘     └───┬────┘     └─────┬─────┘     └────┬─────┘     └────┬─────┘
     │              │                │                 │                 │
     │ GET /milestone/cameras        │                 │                 │
     ├─────────────>│                │                 │                 │
     │              │ GET /vms/milestone/cameras       │                 │
     │              ├───────────────>│                 │                 │
     │              │                │ GET /api/rest/v1/cameras          │
     │              │                ├────────────────>│                 │
     │              │                │                 │                 │
     │              │                │  Camera List    │                 │
     │              │                │<────────────────┤                 │
     │              │                │                 │                 │
     │              │  Camera List   │                 │                 │
     │              │<───────────────┤                 │                 │
     │              │                │                 │                 │
     │  Camera List │                │                 │                 │
     │<─────────────┤                │                 │                 │
     │              │                │                 │                 │
     │ POST /cameras/import          │                 │                 │
     │ {milestoneDeviceID}           │                 │                 │
     ├─────────────>│                │                 │                 │
     │              │ POST /vms/cameras                │                 │
     │              ├───────────────>│                 │                 │
     │              │                │ INSERT INTO cameras              │
     │              │                ├─────────────────────────────────>│
     │              │                │                 │                 │
     │              │                │  Success        │                 │
     │              │                │<─────────────────────────────────┤
     │              │  Camera        │                 │                 │
     │              │<───────────────┤                 │                 │
     │  Camera      │                │                 │                 │
     │<─────────────┤                │                 │                 │
```

### 6.2 Recording Control Flow

```
┌─────────┐     ┌────────┐     ┌─────────────────┐     ┌──────────┐
│Dashboard│     │go-api  │     │recording-service│     │Milestone │
└────┬────┘     └───┬────┘     └────────┬────────┘     └────┬─────┘
     │              │                    │                    │
     │ POST /cameras/{id}/recordings/start                   │
     │ {duration: 900}                   │                    │
     ├─────────────>│                    │                    │
     │              │ POST /recordings/start                  │
     │              ├───────────────────>│                    │
     │              │                    │ POST /api/rest/v1/cameras/{id}/recordings/start
     │              │                    ├───────────────────>│
     │              │                    │                    │
     │              │                    │  RecordingSession  │
     │              │                    │<───────────────────┤
     │              │                    │                    │
     │              │                    │ Schedule Stop      │
     │              │                    │ (Timer: 15 min)    │
     │              │                    │                    │
     │              │  RecordingSession  │                    │
     │              │<───────────────────┤                    │
     │              │                    │                    │
     │ RecordingSession                  │                    │
     │<─────────────┤                    │                    │
     │              │                    │                    │
     │ WS: recording_started             │                    │
     │<─────────────┤                    │                    │
     │              │                    │                    │
     │  [15 min elapses]                 │                    │
     │              │                    │ Timer Fires        │
     │              │                    │                    │
     │              │                    │ POST /api/rest/v1/cameras/{id}/recordings/stop
     │              │                    ├───────────────────>│
     │              │                    │                    │
     │              │                    │  Success           │
     │              │                    │<───────────────────┤
     │              │                    │                    │
     │ WS: recording_stopped             │                    │
     │<─────────────┤<───────────────────┤                    │
```

### 6.3 Playback Flow

```
┌─────────┐     ┌────────┐     ┌────────────────┐     ┌──────────┐
│Dashboard│     │go-api  │     │playback-service│     │Milestone │
└────┬────┘     └───┬────┘     └───────┬────────┘     └────┬─────┘
     │              │                   │                    │
     │ POST /cameras/{id}/recordings/query                  │
     │ {startTime, endTime}             │                    │
     ├─────────────>│                   │                    │
     │              │ GET /playback/sequences               │
     │              ├──────────────────>│                    │
     │              │                   │ GET /api/rest/v1/cameras/{id}/sequences
     │              │                   ├───────────────────>│
     │              │                   │                    │
     │              │                   │  SequenceList      │
     │              │                   │<───────────────────┤
     │              │                   │                    │
     │              │  TimelineData     │                    │
     │              │<──────────────────┤                    │
     │              │                   │                    │
     │  TimelineData│                   │                    │
     │<─────────────┤                   │                    │
     │              │                   │                    │
     │ [User clicks play at timestamp]  │                    │
     │              │                   │                    │
     │ GET /cameras/{id}/playback/stream?time=...&speed=1   │
     ├─────────────>│                   │                    │
     │              │ GET /playback/stream                   │
     │              ├──────────────────>│                    │
     │              │                   │ GET /api/rest/v1/cameras/{id}/video
     │              │                   ├───────────────────>│
     │              │                   │                    │
     │              │                   │  Video Stream      │
     │              │                   │<───────────────────┤
     │              │                   │                    │
     │              │                   │ FFmpeg Transmux    │
     │              │                   │ to HLS             │
     │              │                   │                    │
     │              │  HLS Playlist     │                    │
     │              │<──────────────────┤                    │
     │              │                   │                    │
     │  HLS Stream  │                   │                    │
     │<─────────────┤                   │                    │
```

---

## 7. Database Schema Updates

### 7.1 Cameras Table Update

```sql
-- Add Milestone-specific fields to cameras table
ALTER TABLE cameras ADD COLUMN IF NOT EXISTS milestone_device_id VARCHAR(255);
ALTER TABLE cameras ADD COLUMN IF NOT EXISTS milestone_server VARCHAR(255);
ALTER TABLE cameras ADD COLUMN IF NOT EXISTS last_milestone_sync TIMESTAMP;
ALTER TABLE cameras ADD COLUMN IF NOT EXISTS milestone_metadata JSONB;

-- Create index for Milestone device ID lookups
CREATE INDEX IF NOT EXISTS idx_cameras_milestone_device_id ON cameras(milestone_device_id);
CREATE INDEX IF NOT EXISTS idx_cameras_milestone_server ON cameras(milestone_server);
```

### 7.2 New Tables

#### milestone_recording_sessions
```sql
CREATE TABLE IF NOT EXISTS milestone_recording_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id UUID NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
    milestone_recording_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INT NOT NULL,
    triggered_by VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- recording, stopped, failed
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recording_sessions_camera_id ON milestone_recording_sessions(camera_id);
CREATE INDEX idx_recording_sessions_status ON milestone_recording_sessions(status);
CREATE INDEX idx_recording_sessions_start_time ON milestone_recording_sessions(start_time);
```

#### milestone_sync_history
```sql
CREATE TABLE IF NOT EXISTS milestone_sync_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sync_type VARCHAR(50) NOT NULL, -- camera_discovery, camera_import, full_sync
    cameras_discovered INT,
    cameras_imported INT,
    cameras_updated INT,
    errors INT,
    error_details JSONB,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(50) NOT NULL, -- in_progress, completed, failed
    initiated_by VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sync_history_sync_type ON milestone_sync_history(sync_type);
CREATE INDEX idx_sync_history_started_at ON milestone_sync_history(started_at);
```

#### milestone_playback_cache
```sql
CREATE TABLE IF NOT EXISTS milestone_playback_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id UUID NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
    query_hash VARCHAR(64) NOT NULL, -- MD5 of query parameters
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    sequence_data JSONB NOT NULL,
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    hit_count INT DEFAULT 0
);

CREATE UNIQUE INDEX idx_playback_cache_query_hash ON milestone_playback_cache(query_hash);
CREATE INDEX idx_playback_cache_camera_id ON milestone_playback_cache(camera_id);
CREATE INDEX idx_playback_cache_expires_at ON milestone_playback_cache(expires_at);
```

---

## 8. API Specifications

### 8.1 Camera Discovery APIs

#### List Milestone Cameras
```http
GET /api/v1/milestone/cameras
Authorization: Bearer <token>

Query Parameters:
  - limit: int (default: 100)
  - offset: int (default: 0)
  - recording: bool (filter by recording status)
  - ptz: bool (filter by PTZ capability)
  - search: string (search by name)

Response 200 OK:
{
  "cameras": [
    {
      "id": "3772df00-6ea9-4893-b4b3-6aa944f38861",
      "name": "Camera 01 - Main Entrance",
      "enabled": true,
      "recording": true,
      "ptzEnabled": true,
      "recordingServer": "Recording Server 1",
      "imported": false,
      "liveStreamUrl": "rtsp://milestone:7563/..."
    }
  ],
  "total": 150,
  "imported": 45
}
```

#### Import Camera from Milestone
```http
POST /api/v1/cameras/import
Authorization: Bearer <token>
Content-Type: application/json

{
  "milestoneDeviceId": "3772df00-6ea9-4893-b4b3-6aa944f38861",
  "name": "Custom Name (optional)",
  "nameAr": "Arabic Name (optional)",
  "source": "DUBAI_POLICE",
  "location": {
    "latitude": 25.2048,
    "longitude": 55.2708,
    "address": "Sheikh Zayed Road"
  }
}

Response 201 Created:
{
  "id": "a1b2c3d4-...",
  "name": "Camera 01 - Main Entrance",
  "milestone_device_id": "3772df00-6ea9-4893-b4b3-6aa944f38861",
  "rtsp_url": "rtsp://milestone:7563/...",
  "ptz_enabled": true,
  "status": "ONLINE",
  "created_at": "2025-10-27T14:30:00Z"
}
```

#### Bulk Sync Cameras
```http
POST /api/v1/milestone/sync-all
Authorization: Bearer <token>
Content-Type: application/json

{
  "sourceFilter": "DUBAI_POLICE", // optional
  "autoImport": false // if true, auto-import all discovered cameras
}

Response 202 Accepted:
{
  "syncJobId": "sync_abc123",
  "status": "in_progress",
  "message": "Camera synchronization started"
}
```

### 8.2 Recording Control APIs

#### Start Recording
```http
POST /api/v1/cameras/{cameraId}/recordings/start
Authorization: Bearer <token>
Content-Type: application/json

{
  "duration": 900 // seconds, default: 900 (15 min)
}

Response 200 OK:
{
  "recordingId": "rec_abc123",
  "cameraId": "a1b2c3d4-...",
  "startTime": "2025-10-27T14:30:00Z",
  "estimatedEndTime": "2025-10-27T14:45:00Z",
  "durationSeconds": 900,
  "status": "recording"
}

Response 409 Conflict (if already recording):
{
  "error": "Camera is already recording",
  "currentRecording": {
    "recordingId": "rec_xyz789",
    "startTime": "2025-10-27T14:20:00Z",
    "remainingSeconds": 300
  }
}
```

#### Stop Recording
```http
POST /api/v1/cameras/{cameraId}/recordings/stop
Authorization: Bearer <token>

Response 200 OK:
{
  "recordingId": "rec_abc123",
  "cameraId": "a1b2c3d4-...",
  "startTime": "2025-10-27T14:30:00Z",
  "endTime": "2025-10-27T14:42:15Z",
  "actualDuration": 735,
  "status": "stopped"
}

Response 404 Not Found:
{
  "error": "No active recording found for camera"
}
```

#### Get Recording Status
```http
GET /api/v1/cameras/{cameraId}/recordings/status
Authorization: Bearer <token>

Response 200 OK (recording):
{
  "isRecording": true,
  "recording": {
    "recordingId": "rec_abc123",
    "startTime": "2025-10-27T14:30:00Z",
    "elapsedSeconds": 245,
    "remainingSeconds": 655,
    "durationSeconds": 900
  }
}

Response 200 OK (not recording):
{
  "isRecording": false,
  "lastRecording": {
    "recordingId": "rec_xyz789",
    "startTime": "2025-10-27T13:00:00Z",
    "endTime": "2025-10-27T13:15:00Z"
  }
}
```

### 8.3 Recording Query APIs

#### Query Recordings
```http
POST /api/v1/cameras/{cameraId}/recordings/query
Authorization: Bearer <token>
Content-Type: application/json

{
  "startTime": "2025-10-27T10:00:00Z",
  "endTime": "2025-10-27T14:00:00Z"
}

Response 200 OK:
{
  "cameraId": "a1b2c3d4-...",
  "queryRange": {
    "start": "2025-10-27T10:00:00Z",
    "end": "2025-10-27T14:00:00Z"
  },
  "sequences": [
    {
      "sequenceId": "seq_001",
      "startTime": "2025-10-27T10:00:00Z",
      "endTime": "2025-10-27T11:30:00Z",
      "durationSeconds": 5400,
      "available": true,
      "sizeBytes": 2147483648
    },
    {
      "sequenceId": "seq_002",
      "startTime": "2025-10-27T12:00:00Z",
      "endTime": "2025-10-27T14:00:00Z",
      "durationSeconds": 7200,
      "available": true,
      "sizeBytes": 2863311360
    }
  ],
  "gaps": [
    {
      "startTime": "2025-10-27T11:30:00Z",
      "endTime": "2025-10-27T12:00:00Z",
      "durationSeconds": 1800
    }
  ],
  "totalRecordingSeconds": 12600,
  "totalGapSeconds": 1800,
  "coverage": 0.875 // 87.5% coverage
}
```

#### Get Timeline Data
```http
GET /api/v1/cameras/{cameraId}/recordings/timeline
Authorization: Bearer <token>

Query Parameters:
  - startTime: ISO 8601 datetime
  - endTime: ISO 8601 datetime
  - resolution: string (minute, hour, day) - aggregation level

Response 200 OK:
{
  "cameraId": "a1b2c3d4-...",
  "resolution": "minute",
  "timeline": [
    {
      "timestamp": "2025-10-27T10:00:00Z",
      "hasRecording": true,
      "segmentCount": 1
    },
    {
      "timestamp": "2025-10-27T10:01:00Z",
      "hasRecording": true,
      "segmentCount": 1
    },
    // ... more timestamps
  ]
}
```

### 8.4 Playback APIs

#### Start Playback Stream
```http
GET /api/v1/cameras/{cameraId}/playback/stream
Authorization: Bearer <token>

Query Parameters:
  - time: ISO 8601 datetime (playback start time)
  - speed: float (default: 1.0, range: -8 to 8)
  - format: string (hls, mjpeg) - default: hls

Response 200 OK:
Content-Type: application/vnd.apple.mpegurl

#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
segment_0.ts
#EXTINF:10.0,
segment_1.ts
...
```

#### Get Snapshot
```http
GET /api/v1/cameras/{cameraId}/playback/snapshot
Authorization: Bearer <token>

Query Parameters:
  - time: ISO 8601 datetime

Response 200 OK:
Content-Type: image/jpeg

[Binary JPEG image data]
```

#### Control Playback
```http
POST /api/v1/cameras/{cameraId}/playback/control
Authorization: Bearer <token>
Content-Type: application/json

{
  "action": "play" | "pause" | "seek" | "speed",
  "timestamp": "2025-10-27T10:30:00Z", // for seek
  "speed": 2.0 // for speed change
}

Response 200 OK:
{
  "status": "playing",
  "currentTime": "2025-10-27T10:30:00Z",
  "speed": 2.0
}
```

---

## 9. Frontend Components

### 9.1 Camera Discovery Component

**Location:** `dashboard/src/components/MilestoneCameraDiscovery.tsx`

```typescript
interface MilestoneCameraDiscoveryProps {
  onClose: () => void;
  onImport: (cameras: MilestoneCamera[]) => void;
}

export function MilestoneCameraDiscovery({
  onClose,
  onImport
}: MilestoneCameraDiscoveryProps) {
  const [cameras, setCameras] = useState<MilestoneCamera[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedCameras, setSelectedCameras] = useState<Set<string>>(new Set());

  // Features:
  // - List all Milestone cameras
  // - Filter by recording/PTZ/imported status
  // - Search by name
  // - Bulk select and import
  // - Show import status
  // - Real-time sync progress
}
```

### 9.2 Recording Control Component

**Location:** `dashboard/src/components/RecordingControl.tsx`

```typescript
interface RecordingControlProps {
  cameraId: string;
  isRecording: boolean;
  recordingStatus?: RecordingStatus;
  onStartRecording: (duration: number) => void;
  onStopRecording: () => void;
}

export function RecordingControl({
  cameraId,
  isRecording,
  recordingStatus,
  onStartRecording,
  onStopRecording
}: RecordingControlProps) {
  const [duration, setDuration] = useState(900); // 15 min default

  // Features:
  // - Duration selector dropdown (5/15/30/60 min, custom)
  // - Start/Stop button with state management
  // - Countdown timer with progress bar
  // - Remaining time display
  // - Status notifications
}
```

### 9.3 Timeline Component

**Location:** `dashboard/src/components/RecordingTimeline.tsx`

```typescript
interface RecordingTimelineProps {
  cameraId: string;
  startTime: Date;
  endTime: Date;
  sequences: RecordingSequence[];
  gaps: RecordingGap[];
  onSeek: (timestamp: Date) => void;
  currentPlaybackTime?: Date;
}

export function RecordingTimeline({
  cameraId,
  startTime,
  endTime,
  sequences,
  gaps,
  onSeek,
  currentPlaybackTime
}: RecordingTimelineProps) {
  const [zoomLevel, setZoomLevel] = useState(1);

  // Features:
  // - Interactive timeline bar (SVG-based)
  // - Recording availability blocks (green)
  // - Gap visualization (red/gray)
  // - Playhead indicator (synced with video)
  // - Time markers (hour/minute labels)
  // - Zoom controls (1x, 2x, 4x, 8x)
  // - Click to seek
  // - Hover tooltip with timestamp
}
```

### 9.4 Playback Player Component

**Location:** `dashboard/src/components/RecordingPlayer.tsx`

```typescript
interface RecordingPlayerProps {
  cameraId: string;
  startTime: Date;
  endTime: Date;
  initialPlaybackTime?: Date;
  onPlaybackTimeChange: (time: Date) => void;
}

export function RecordingPlayer({
  cameraId,
  startTime,
  endTime,
  initialPlaybackTime,
  onPlaybackTimeChange
}: RecordingPlayerProps) {
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState(1.0);
  const [currentTime, setCurrentTime] = useState(initialPlaybackTime);

  // Features:
  // - HLS video player (using hls.js)
  // - VCR controls (play/pause/forward/backward)
  // - Speed selector (-8x to 8x)
  // - Frame-by-frame navigation
  // - Timestamp overlay
  // - Fullscreen mode
  // - Volume control
}
```

### 9.5 Sidebar Integration

**Location:** Update `dashboard/src/components/CameraSidebarNew.tsx`

Add recording control section:

```typescript
// Add state
const [showRecordingControl, setShowRecordingControl] = useState(false);
const [recordingStatus, setRecordingStatus] = useState<RecordingStatus | null>(null);

// Add UI section
<div className="border-t border-gray-200 p-3">
  <h3 className="text-sm font-semibold mb-2">Recording</h3>

  {recordingStatus?.isRecording ? (
    <RecordingControl
      cameraId={selectedCameraId}
      isRecording={true}
      recordingStatus={recordingStatus}
      onStopRecording={handleStopRecording}
    />
  ) : (
    <button onClick={handleStartRecording}>
      <RecordCircle className="w-4 h-4 mr-2" />
      Start Recording
    </button>
  )}

  <button onClick={handleViewRecordings}>
    <Film className="w-4 h-4 mr-2" />
    View Recordings
  </button>
</div>
```

---

## 10. Testing Strategy

### 10.1 Unit Tests

**Milestone Client Tests:**
```go
// services/vms-service/internal/client/milestone_client_test.go
func TestMilestoneClient_Login(t *testing.T)
func TestMilestoneClient_RefreshToken(t *testing.T)
func TestMilestoneClient_ListCameras(t *testing.T)
func TestMilestoneClient_StartRecording(t *testing.T)
func TestMilestoneClient_QuerySequences(t *testing.T)
```

**Recording Manager Tests:**
```go
// services/recording-service/internal/manager/milestone_recording_manager_test.go
func TestRecordingManager_StartRecording(t *testing.T)
func TestRecordingManager_AutoStop(t *testing.T)
func TestRecordingManager_RecoverySessions(t *testing.T)
```

**Frontend Component Tests:**
```typescript
// dashboard/src/components/__tests__/RecordingControl.test.tsx
describe('RecordingControl', () => {
  it('should start recording with default duration', () => {});
  it('should show countdown timer when recording', () => {});
  it('should stop recording when button clicked', () => {});
});
```

### 10.2 Integration Tests

**API Integration Tests:**
```go
// services/go-api/tests/integration/milestone_test.go
func TestMilestoneIntegration_CameraDiscovery(t *testing.T)
func TestMilestoneIntegration_RecordingControl(t *testing.T)
func TestMilestoneIntegration_RecordingQuery(t *testing.T)
func TestMilestoneIntegration_Playback(t *testing.T)
```

### 10.3 End-to-End Tests

**User Workflows:**
```typescript
// dashboard/e2e/milestone.spec.ts
describe('Milestone Integration', () => {
  it('should discover and import cameras', () => {});
  it('should start and stop recording', () => {});
  it('should query recordings and play video', () => {});
});
```

### 10.4 Performance Tests

**Load Testing:**
- 100 concurrent camera queries
- 50 simultaneous playback sessions
- 20 concurrent recording starts
- Timeline query response time < 500ms
- Playback stream start time < 2s

---

## 11. Deployment Plan

### 11.1 Environment Variables

Add to `docker-compose.yml`:

```yaml
services:
  vms-service:
    environment:
      - MILESTONE_BASE_URL=http://milestone-server:80
      - MILESTONE_USERNAME=${MILESTONE_USERNAME}
      - MILESTONE_PASSWORD=${MILESTONE_PASSWORD}
      - MILESTONE_AUTH_TYPE=basic # or ntlm
      - MILESTONE_SESSION_TIMEOUT=3600
      - MILESTONE_RETRY_ATTEMPTS=3
      - MILESTONE_RETRY_DELAY=5s

  recording-service:
    environment:
      - MILESTONE_BASE_URL=http://milestone-server:80
      - MILESTONE_USERNAME=${MILESTONE_USERNAME}
      - MILESTONE_PASSWORD=${MILESTONE_PASSWORD}
      - RECORDING_AUTO_STOP_ENABLED=true
      - RECORDING_MAX_DURATION=7200 # 2 hours

  playback-service:
    environment:
      - MILESTONE_BASE_URL=http://milestone-server:80
      - MILESTONE_USERNAME=${MILESTONE_USERNAME}
      - MILESTONE_PASSWORD=${MILESTONE_PASSWORD}
      - PLAYBACK_CACHE_TTL=300 # 5 minutes
      - PLAYBACK_MAX_CONCURRENT_STREAMS=50
```

### 11.2 Kong Gateway Routes

Add routes in `config/kong/kong.yml`:

```yaml
services:
  - name: milestone-cameras
    url: http://vms-service:8081/vms/milestone
    routes:
      - name: milestone-camera-discovery
        paths:
          - /api/v1/milestone/cameras
        methods:
          - GET
      - name: milestone-camera-import
        paths:
          - /api/v1/cameras/import
        methods:
          - POST

  - name: milestone-recordings
    url: http://recording-service:8083/recordings
    routes:
      - name: recording-start
        paths:
          - /api/v1/cameras/~/recordings/start
        methods:
          - POST
      - name: recording-stop
        paths:
          - /api/v1/cameras/~/recordings/stop
        methods:
          - POST
      - name: recording-status
        paths:
          - /api/v1/cameras/~/recordings/status
        methods:
          - GET

  - name: milestone-playback
    url: http://playback-service:8084/playback
    routes:
      - name: playback-query
        paths:
          - /api/v1/cameras/~/recordings/query
        methods:
          - POST
      - name: playback-stream
        paths:
          - /api/v1/cameras/~/playback/stream
        methods:
          - GET
```

### 11.3 Database Migrations

Create migration files:

```bash
# services/vms-service/migrations/
000008_add_milestone_fields.up.sql
000008_add_milestone_fields.down.sql

000009_create_recording_sessions.up.sql
000009_create_recording_sessions.down.sql

000010_create_sync_history.up.sql
000010_create_sync_history.down.sql

000011_create_playback_cache.up.sql
000011_create_playback_cache.down.sql
```

### 11.4 Deployment Steps

1. **Pre-deployment:**
   - Backup PostgreSQL database
   - Test Milestone connectivity
   - Review environment variables
   - Run database migrations on staging

2. **Deployment:**
   - Build new Docker images
   - Update docker-compose.yml
   - Deploy services (rolling update)
   - Run database migrations
   - Verify service health checks

3. **Post-deployment:**
   - Test camera discovery
   - Test recording control
   - Test playback functionality
   - Monitor logs and metrics
   - Performance testing

4. **Rollback Plan:**
   - Revert to previous Docker images
   - Rollback database migrations
   - Restore configuration

---

## 12. Risk Assessment & Mitigation

### 12.1 Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Milestone API changes | High | Low | Version checking, API contracts, fallback mechanisms |
| Authentication failures | High | Medium | Token refresh, retry logic, health checks |
| Network latency | Medium | High | Caching, connection pooling, timeout configuration |
| Concurrent recording limits | Medium | Medium | Queue management, user notifications, limit checks |
| Video streaming performance | High | Medium | CDN integration, adaptive bitrate, HLS optimization |
| Session expiration | Medium | High | Auto-refresh tokens, graceful degradation |
| Database performance | Medium | Medium | Indexes, query optimization, connection pooling |

### 12.2 Operational Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Milestone server downtime | High | Low | Fallback to local recordings, status monitoring |
| Credential leakage | Critical | Low | Secrets management, environment variables, access control |
| Insufficient storage | High | Medium | Storage monitoring, auto-cleanup, alerts |
| High concurrent load | Medium | High | Load balancing, rate limiting, horizontal scaling |
| Data sync issues | Medium | Medium | Sync verification, conflict resolution, manual sync |

### 12.3 Security Considerations

1. **Authentication:**
   - Store Milestone credentials in secrets manager (e.g., HashiCorp Vault)
   - Use encrypted environment variables
   - Implement token rotation
   - Audit authentication attempts

2. **Authorization:**
   - Implement user permissions for recording control
   - Camera-level access control
   - API rate limiting per user
   - Audit logs for sensitive operations

3. **Data Protection:**
   - Encrypt video streams (HTTPS/TLS)
   - Secure database connections
   - PII protection in logs
   - GDPR compliance for recordings

4. **Network Security:**
   - VPN/private network for Milestone connection
   - Firewall rules
   - DDoS protection
   - Network segmentation

---

## Appendix A: Configuration Examples

### A.1 Complete Environment File

```env
# Milestone Configuration
MILESTONE_BASE_URL=http://milestone-mgmt-server:80
MILESTONE_USERNAME=api_user
MILESTONE_PASSWORD=secure_password
MILESTONE_AUTH_TYPE=basic
MILESTONE_SESSION_TIMEOUT=3600
MILESTONE_MAX_CONNECTIONS=50

# Recording Configuration
RECORDING_DEFAULT_DURATION=900
RECORDING_MAX_DURATION=7200
RECORDING_AUTO_STOP=true
RECORDING_NOTIFICATIONS=true

# Playback Configuration
PLAYBACK_CACHE_ENABLED=true
PLAYBACK_CACHE_TTL=300
PLAYBACK_MAX_SPEED=8
PLAYBACK_MIN_SPEED=-8

# Performance Tuning
DB_MAX_CONNECTIONS=100
CACHE_SIZE_MB=512
WORKER_POOL_SIZE=20
```

### A.2 Nginx Reverse Proxy (Optional)

```nginx
upstream milestone {
    server milestone-server:80;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name milestone.example.com;

    location /api/rest/v1/ {
        proxy_pass http://milestone;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_connect_timeout 60s;
        proxy_send_timeout 600s;
        proxy_read_timeout 600s;
    }
}
```

---

## Appendix B: API Response Examples

### B.1 Full Recording Query Response

```json
{
  "cameraId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "cameraName": "Camera 01 - Main Entrance",
  "queryRange": {
    "start": "2025-10-27T00:00:00Z",
    "end": "2025-10-27T23:59:59Z"
  },
  "sequences": [
    {
      "sequenceId": "seq_20251027_000000",
      "startTime": "2025-10-27T00:00:00Z",
      "endTime": "2025-10-27T06:00:00Z",
      "durationSeconds": 21600,
      "available": true,
      "sizeBytes": 8589934592,
      "quality": "high",
      "fps": 25,
      "resolution": "1920x1080"
    },
    {
      "sequenceId": "seq_20251027_080000",
      "startTime": "2025-10-27T08:00:00Z",
      "endTime": "2025-10-27T18:00:00Z",
      "durationSeconds": 36000,
      "available": true,
      "sizeBytes": 14316557653,
      "quality": "high",
      "fps": 25,
      "resolution": "1920x1080"
    },
    {
      "sequenceId": "seq_20251027_200000",
      "startTime": "2025-10-27T20:00:00Z",
      "endTime": "2025-10-27T23:59:59Z",
      "durationSeconds": 14399,
      "available": true,
      "sizeBytes": 5726623061,
      "quality": "high",
      "fps": 25,
      "resolution": "1920x1080"
    }
  ],
  "gaps": [
    {
      "startTime": "2025-10-27T06:00:00Z",
      "endTime": "2025-10-27T08:00:00Z",
      "durationSeconds": 7200,
      "reason": "scheduled_maintenance"
    },
    {
      "startTime": "2025-10-27T18:00:00Z",
      "endTime": "2025-10-27T20:00:00Z",
      "durationSeconds": 7200,
      "reason": "recording_server_offline"
    }
  ],
  "statistics": {
    "totalRecordingSeconds": 71999,
    "totalGapSeconds": 14400,
    "coverage": 0.833,
    "totalSizeBytes": 28633115306,
    "averageBitrate": 3182346
  },
  "metadata": {
    "recordingServer": "Recording Server 1",
    "storageLocation": "SAN-01",
    "retentionDays": 30
  }
}
```

---

## Appendix C: Frontend State Management

### C.1 Zustand Store for Recording

```typescript
// dashboard/src/stores/recordingStore.ts
interface RecordingStore {
  activeRecordings: Map<string, RecordingStatus>;
  recordingHistory: RecordingHistoryItem[];

  startRecording: (cameraId: string, duration: number) => Promise<void>;
  stopRecording: (cameraId: string) => Promise<void>;
  getRecordingStatus: (cameraId: string) => Promise<RecordingStatus>;

  updateRecordingStatus: (cameraId: string, status: RecordingStatus) => void;
  clearRecording: (cameraId: string) => void;
}

export const useRecordingStore = create<RecordingStore>((set, get) => ({
  activeRecordings: new Map(),
  recordingHistory: [],

  startRecording: async (cameraId, duration) => {
    const response = await api.post(`/cameras/${cameraId}/recordings/start`, {
      duration
    });
    set((state) => ({
      activeRecordings: new Map(state.activeRecordings).set(cameraId, response)
    }));
  },

  // ... more actions
}));
```

### C.2 WebSocket Event Handling

```typescript
// dashboard/src/hooks/useRecordingEvents.ts
export function useRecordingEvents(cameraId: string) {
  const { updateRecordingStatus, clearRecording } = useRecordingStore();

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws');

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);

      if (data.type === 'recording_started' && data.cameraId === cameraId) {
        updateRecordingStatus(cameraId, data.status);
        toast.success('Recording started successfully');
      }

      if (data.type === 'recording_stopped' && data.cameraId === cameraId) {
        clearRecording(cameraId);
        toast.info('Recording stopped');
      }

      if (data.type === 'recording_progress' && data.cameraId === cameraId) {
        updateRecordingStatus(cameraId, data.status);
      }
    };

    return () => ws.close();
  }, [cameraId]);
}
```

---

## Document Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-27 | System | Initial comprehensive plan created |

---

## Next Steps

1. ✅ Review and approve this implementation plan
2. ⏳ Set up Milestone test environment
3. ⏳ Begin Phase 1: Milestone API Client Foundation
4. ⏳ Create GitHub issues for each phase
5. ⏳ Set up CI/CD pipeline updates

---

**End of Document**
