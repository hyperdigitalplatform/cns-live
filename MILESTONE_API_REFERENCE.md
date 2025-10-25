# 🔌 Milestone XProtect Expert VMS - API Reference

## 📋 Overview

Complete API reference for integrating with Milestone XProtect Expert VMS server at **192.168.1.9**.

**Base URL:** `http://192.168.1.9`

---

## 🔐 Authentication

### Method 1: Basic Authentication
```http
Authorization: Basic <base64(username:password)>
```

### Method 2: Token-based Authentication (MIP SDK)
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "rta-integration",
  "password": "your-password"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}

# Use token in subsequent requests
Authorization: Bearer <token>
```

---

## 📹 1. Camera Management APIs

### 1.1 Get All Cameras

**Endpoint:** `GET /api/rest/v1/cameras`

**Description:** Retrieve list of all cameras configured in Milestone

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/cameras
Authorization: Basic <credentials>
```

**Response:**
```json
{
  "cameras": [
    {
      "id": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
      "name": "Main Entrance Camera",
      "enabled": true,
      "recordingEnabled": true,
      "channel": 1,
      "shortName": "Cam001",
      "recordingServer": {
        "id": "server-001",
        "name": "Recording Server 1",
        "address": "192.168.1.9"
      },
      "properties": {
        "manufacturer": "Axis",
        "model": "P3245-LVE",
        "macAddress": "00:40:8C:12:34:56",
        "ipAddress": "192.168.1.101",
        "firmware": "9.80.2"
      },
      "streams": [
        {
          "streamId": 1,
          "name": "High Quality",
          "resolution": "1920x1080",
          "codec": "H264",
          "fps": 25,
          "bitrate": 4096
        },
        {
          "streamId": 2,
          "name": "Low Quality",
          "resolution": "640x480",
          "codec": "H264",
          "fps": 15,
          "bitrate": 512
        }
      ],
      "ptzCapabilities": {
        "supportsPan": true,
        "supportsTilt": true,
        "supportsZoom": true,
        "presets": 10
      }
    }
  ],
  "totalCount": 500
}
```

**Query Parameters:**
- `limit` - Number of results per page (default: 100)
- `offset` - Pagination offset (default: 0)
- `enabled` - Filter by enabled status (true/false)
- `recordingEnabled` - Filter by recording status (true/false)

**Example:**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/cameras?limit=50&offset=0&enabled=true"
```

---

### 1.2 Get Camera by ID

**Endpoint:** `GET /api/rest/v1/cameras/{cameraId}`

**Description:** Get detailed information for a specific camera

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/cameras/8b3a2c1d-4e5f-6789-abcd-ef0123456789
Authorization: Basic <credentials>
```

**Response:**
```json
{
  "id": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "name": "Main Entrance Camera",
  "enabled": true,
  "recordingEnabled": true,
  "location": {
    "building": "Headquarters",
    "floor": "Ground Floor",
    "area": "Main Entrance"
  },
  "recordingServer": {
    "id": "server-001",
    "name": "Recording Server 1",
    "address": "192.168.1.9"
  },
  "streams": [
    {
      "streamId": 1,
      "liveUrl": "rtsp://192.168.1.9:554/live/8b3a2c1d-4e5f-6789-abcd-ef0123456789",
      "resolution": "1920x1080",
      "codec": "H264",
      "fps": 25
    }
  ],
  "recording": {
    "mode": "continuous",
    "retentionDays": 90,
    "storageSize": 1572864000,
    "oldestRecording": "2025-07-26T00:00:00Z",
    "newestRecording": "2025-10-25T23:59:59Z"
  }
}
```

---

### 1.3 Get Camera Groups

**Endpoint:** `GET /api/rest/v1/cameras/groups`

**Description:** Get camera groups/categories

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/cameras/groups
Authorization: Basic <credentials>
```

**Response:**
```json
{
  "groups": [
    {
      "id": "group-001",
      "name": "Dubai Police",
      "description": "Dubai Police Department Cameras",
      "cameraCount": 50,
      "cameras": [
        "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
        "9c4b3d2e-5f6a-7890-bcde-f12345678901"
      ]
    },
    {
      "id": "group-002",
      "name": "Metro",
      "description": "Metro Station Cameras",
      "cameraCount": 30,
      "cameras": ["..."]
    },
    {
      "id": "group-003",
      "name": "Bus",
      "description": "Bus Station Cameras",
      "cameraCount": 20,
      "cameras": ["..."]
    }
  ]
}
```

---

## 🎬 2. Live Streaming APIs

### 2.1 Get Live Stream URL (RTSP)

**Endpoint:** `GET /api/rest/v1/cameras/{cameraId}/live`

**Description:** Get RTSP URL for live streaming

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/cameras/8b3a2c1d-4e5f-6789-abcd-ef0123456789/live
Authorization: Basic <credentials>
```

**Response:**
```json
{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "streams": [
    {
      "streamId": 1,
      "quality": "high",
      "protocol": "rtsp",
      "url": "rtsp://192.168.1.9:554/live/8b3a2c1d-4e5f-6789-abcd-ef0123456789?stream=1",
      "resolution": "1920x1080",
      "fps": 25,
      "codec": "H264"
    },
    {
      "streamId": 2,
      "quality": "low",
      "protocol": "rtsp",
      "url": "rtsp://192.168.1.9:554/live/8b3a2c1d-4e5f-6789-abcd-ef0123456789?stream=2",
      "resolution": "640x480",
      "fps": 15,
      "codec": "H264"
    }
  ]
}
```

**RTSP URL Format:**
```
rtsp://192.168.1.9:554/live/{cameraId}
rtsp://192.168.1.9:554/live/{cameraId}?stream={streamId}

# With authentication
rtsp://username:password@192.168.1.9:554/live/{cameraId}
```

**Example Usage:**
```bash
# Using ffplay
ffplay rtsp://username:password@192.168.1.9:554/live/8b3a2c1d-4e5f-6789-abcd-ef0123456789

# Using VLC
vlc rtsp://192.168.1.9:554/live/8b3a2c1d-4e5f-6789-abcd-ef0123456789
```

---

### 2.2 Get Live Snapshot

**Endpoint:** `GET /api/rest/v1/cameras/{cameraId}/snapshot`

**Description:** Get current snapshot/thumbnail from camera

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/cameras/8b3a2c1d-4e5f-6789-abcd-ef0123456789/snapshot
Authorization: Basic <credentials>
```

**Response:**
```
Content-Type: image/jpeg
<binary image data>
```

**Query Parameters:**
- `width` - Image width (default: 640)
- `height` - Image height (default: 480)
- `quality` - JPEG quality 1-100 (default: 80)

**Example:**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/cameras/8b3a2c1d-4e5f-6789-abcd-ef0123456789/snapshot?width=1920&height=1080&quality=90" \
  -o snapshot.jpg
```

---

## 📼 3. Recording & Playback APIs

### 3.1 Get Recording Timeline

**Endpoint:** `GET /api/rest/v1/recordings/{cameraId}/timeline`

**Description:** Get timeline of available recordings for a camera

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/timeline
Authorization: Basic <credentials>
```

**Query Parameters (Required):**
- `startTime` - Start of time range (ISO 8601 format)
- `endTime` - End of time range (ISO 8601 format)

**Example:**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/timeline?startTime=2025-10-24T00:00:00Z&endTime=2025-10-25T23:59:59Z"
```

**Response:**
```json
{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "cameraName": "Main Entrance Camera",
  "timeRange": {
    "start": "2025-10-24T00:00:00Z",
    "end": "2025-10-25T23:59:59Z"
  },
  "sequences": [
    {
      "sequenceId": "seq-001",
      "startTime": "2025-10-24T00:00:00Z",
      "endTime": "2025-10-24T06:00:00Z",
      "durationSeconds": 21600,
      "sizeBytes": 15728640000,
      "hasVideo": true,
      "hasAudio": false,
      "codec": "H264",
      "resolution": "1920x1080",
      "fps": 25,
      "eventCount": 45,
      "gaps": []
    },
    {
      "sequenceId": "seq-002",
      "startTime": "2025-10-24T06:00:00Z",
      "endTime": "2025-10-24T12:00:00Z",
      "durationSeconds": 21600,
      "sizeBytes": 15728640000,
      "hasVideo": true,
      "hasAudio": false,
      "codec": "H264",
      "resolution": "1920x1080",
      "fps": 25,
      "eventCount": 67,
      "gaps": []
    }
  ],
  "totalDurationSeconds": 172800,
  "totalSizeBytes": 125829120000,
  "gapCount": 0
}
```

---

### 3.2 Get Playback Stream URL (RTSP)

**Endpoint:** `GET /api/rest/v1/recordings/{cameraId}/playback`

**Description:** Get RTSP URL for recorded video playback

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/playback
Authorization: Basic <credentials>
```

**Query Parameters:**
- `startTime` - Start timestamp (ISO 8601)
- `endTime` - End timestamp (ISO 8601) [Optional]
- `speed` - Playback speed: 1, 2, 4, 8 (default: 1)

**Response:**
```json
{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "playbackUrl": "rtsp://192.168.1.9:554/playback/8b3a2c1d-4e5f-6789-abcd-ef0123456789?start=2025-10-24T10:00:00Z&end=2025-10-24T11:00:00Z&speed=1",
  "startTime": "2025-10-24T10:00:00Z",
  "endTime": "2025-10-24T11:00:00Z",
  "durationSeconds": 3600,
  "speed": 1
}
```

**RTSP Playback URL Format:**
```
rtsp://192.168.1.9:554/playback/{cameraId}?start={timestamp}&end={timestamp}
rtsp://192.168.1.9:554/playback/{cameraId}?start={timestamp}&duration={seconds}
rtsp://192.168.1.9:554/playback/{cameraId}?start={timestamp}&speed={speed}

# With authentication
rtsp://username:password@192.168.1.9:554/playback/{cameraId}?start={timestamp}
```

**Example:**
```bash
# Playback 1 hour of recording starting at 10:00 AM
ffplay "rtsp://username:password@192.168.1.9:554/playback/8b3a2c1d-4e5f-6789-abcd-ef0123456789?start=2025-10-24T10:00:00Z&duration=3600"
```

---

### 3.3 Download Recording Segment

**Endpoint:** `GET /api/rest/v1/recordings/{cameraId}/download`

**Description:** Download recording segment as video file

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/download
Authorization: Basic <credentials>
```

**Query Parameters:**
- `startTime` - Start timestamp (ISO 8601) [Required]
- `endTime` - End timestamp (ISO 8601) [Required]
- `format` - Output format: mp4, avi, mkv (default: mp4)
- `codec` - Video codec: h264, h265 (default: h264)
- `quality` - Quality: low, medium, high (default: high)

**Response:**
```
Content-Type: video/mp4
Content-Disposition: attachment; filename="recording_8b3a2c1d_2025-10-24T10:00:00Z.mp4"
<binary video data>
```

**Example:**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/download?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T10:30:00Z&format=mp4" \
  -o recording.mp4
```

---

### 3.4 Export Recording (Async)

**Endpoint:** `POST /api/rest/v1/recordings/export`

**Description:** Create an export job for large recordings (async processing)

**Request:**
```http
POST http://192.168.1.9/api/rest/v1/recordings/export
Authorization: Basic <credentials>
Content-Type: application/json

{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "startTime": "2025-10-24T10:00:00Z",
  "endTime": "2025-10-24T12:00:00Z",
  "format": "mp4",
  "quality": "high",
  "includeAudio": false,
  "watermark": {
    "enabled": true,
    "text": "RTA CCTV - Exported {timestamp}"
  }
}
```

**Response:**
```json
{
  "exportId": "exp-12345-abcde",
  "status": "queued",
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "startTime": "2025-10-24T10:00:00Z",
  "endTime": "2025-10-24T12:00:00Z",
  "estimatedDurationSeconds": 7200,
  "estimatedSizeBytes": 52428800,
  "createdAt": "2025-10-25T14:30:00Z"
}
```

**Check Export Status:**
```http
GET /api/rest/v1/recordings/export/{exportId}

Response:
{
  "exportId": "exp-12345-abcde",
  "status": "completed",
  "progress": 100,
  "downloadUrl": "http://192.168.1.9/api/rest/v1/recordings/export/exp-12345-abcde/download",
  "expiresAt": "2025-10-26T14:30:00Z",
  "fileSize": 52428800
}
```

**Download Export:**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/export/exp-12345-abcde/download" \
  -o exported_video.mp4
```

---

## 📊 4. Metadata & Events APIs

### 4.1 Get Camera Events

**Endpoint:** `GET /api/rest/v1/events/{cameraId}`

**Description:** Get motion detection and other events for a camera

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/events/8b3a2c1d-4e5f-6789-abcd-ef0123456789
Authorization: Basic <credentials>
```

**Query Parameters:**
- `startTime` - Start of time range (ISO 8601)
- `endTime` - End of time range (ISO 8601)
- `eventTypes` - Filter by event types (comma-separated): motion,tampering,audio
- `limit` - Max results (default: 100)

**Response:**
```json
{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "events": [
    {
      "eventId": "evt-001",
      "type": "motion",
      "timestamp": "2025-10-24T10:15:30Z",
      "endTimestamp": "2025-10-24T10:16:00Z",
      "severity": "medium",
      "description": "Motion detected in zone 1",
      "metadata": {
        "zone": 1,
        "confidence": 0.85
      }
    },
    {
      "eventId": "evt-002",
      "type": "tampering",
      "timestamp": "2025-10-24T14:22:10Z",
      "severity": "high",
      "description": "Camera tampering detected"
    }
  ],
  "totalCount": 245
}
```

---

### 4.2 Get Recording Statistics

**Endpoint:** `GET /api/rest/v1/recordings/{cameraId}/statistics`

**Description:** Get recording statistics for a camera

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/statistics
Authorization: Basic <credentials>
```

**Query Parameters:**
- `startTime` - Start of time range (ISO 8601)
- `endTime` - End of time range (ISO 8601)

**Response:**
```json
{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "statistics": {
    "totalDurationSeconds": 7776000,
    "totalSizeBytes": 562949953421312,
    "averageBitrate": 4194304,
    "recordingGaps": 12,
    "totalGapDurationSeconds": 3600,
    "oldestRecording": "2025-07-26T00:00:00Z",
    "newestRecording": "2025-10-25T23:59:59Z",
    "retentionDays": 90,
    "usagePercentage": 75.5
  },
  "dailyBreakdown": [
    {
      "date": "2025-10-24",
      "durationSeconds": 86400,
      "sizeBytes": 6291456000,
      "gaps": 0
    }
  ]
}
```

---

## 🎮 5. PTZ Control APIs

### 5.1 PTZ Absolute Move

**Endpoint:** `POST /api/rest/v1/cameras/{cameraId}/ptz/move`

**Description:** Move PTZ camera to absolute position

**Request:**
```http
POST http://192.168.1.9/api/rest/v1/cameras/8b3a2c1d-4e5f-6789-abcd-ef0123456789/ptz/move
Authorization: Basic <credentials>
Content-Type: application/json

{
  "pan": 180.0,
  "tilt": 45.0,
  "zoom": 2.5,
  "speed": {
    "pan": 0.5,
    "tilt": 0.5,
    "zoom": 0.3
  }
}
```

**Response:**
```json
{
  "status": "success",
  "position": {
    "pan": 180.0,
    "tilt": 45.0,
    "zoom": 2.5
  }
}
```

---

### 5.2 PTZ Relative Move

**Endpoint:** `POST /api/rest/v1/cameras/{cameraId}/ptz/relative`

**Description:** Move PTZ camera relative to current position

**Request:**
```http
POST http://192.168.1.9/api/rest/v1/cameras/8b3a2c1d-4e5f-6789-abcd-ef0123456789/ptz/relative
Content-Type: application/json

{
  "pan": 10,      // degrees right (negative for left)
  "tilt": -5,     // degrees down (positive for up)
  "zoom": 0.5,    // zoom in (negative to zoom out)
  "speed": 0.5    // 0.0 to 1.0
}
```

---

### 5.3 PTZ Preset Management

**Get Presets:**
```http
GET /api/rest/v1/cameras/{cameraId}/ptz/presets

Response:
{
  "presets": [
    {
      "presetId": 1,
      "name": "Main Entrance View",
      "pan": 180.0,
      "tilt": 30.0,
      "zoom": 1.5
    },
    {
      "presetId": 2,
      "name": "Parking Area",
      "pan": 90.0,
      "tilt": 15.0,
      "zoom": 2.0
    }
  ]
}
```

**Go to Preset:**
```http
POST /api/rest/v1/cameras/{cameraId}/ptz/preset/{presetId}

Response:
{
  "status": "success",
  "presetId": 1,
  "presetName": "Main Entrance View"
}
```

---

## 🔄 6. System & Health APIs

### 6.1 System Health

**Endpoint:** `GET /api/rest/v1/system/health`

**Request:**
```http
GET http://192.168.1.9/api/rest/v1/system/health
Authorization: Basic <credentials>
```

**Response:**
```json
{
  "status": "healthy",
  "version": "2023 R3",
  "uptime": 7776000,
  "services": {
    "managementServer": "running",
    "recordingServer": "running",
    "eventServer": "running",
    "mobileServer": "running"
  },
  "recordingServers": [
    {
      "id": "server-001",
      "name": "Recording Server 1",
      "status": "running",
      "cpuUsage": 45.2,
      "memoryUsage": 62.8,
      "storageUsed": 75.5,
      "storageTotal": 1099511627776,
      "cameraCount": 500,
      "activeStreams": 125
    }
  ]
}
```

---

### 6.2 Get Server Capabilities

**Endpoint:** `GET /api/rest/v1/system/capabilities`

**Response:**
```json
{
  "maxCameras": 1000,
  "maxConcurrentStreams": 500,
  "supportedCodecs": ["H264", "H265", "MJPEG"],
  "supportedFormats": ["mp4", "avi", "mkv"],
  "features": {
    "ptzControl": true,
    "motionDetection": true,
    "audioRecording": true,
    "edgeStorage": true,
    "redundantRecording": false
  },
  "apiVersion": "v1",
  "sdkVersion": "2023.3.0"
}
```

---

## 📡 7. WebSocket APIs (Real-time Events)

### 7.1 Connect to Event Stream

**Endpoint:** `ws://192.168.1.9/api/ws/events`

**Connection:**
```javascript
const ws = new WebSocket('ws://192.168.1.9/api/ws/events?token=<auth-token>');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data);
};
```

**Event Messages:**
```json
{
  "type": "camera.motion",
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "timestamp": "2025-10-25T14:30:00Z",
  "data": {
    "zone": 1,
    "confidence": 0.92
  }
}

{
  "type": "camera.offline",
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "timestamp": "2025-10-25T14:35:00Z"
}

{
  "type": "recording.started",
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "timestamp": "2025-10-25T14:40:00Z",
  "sequenceId": "seq-123"
}
```

---

## 📋 8. Batch Operations

### 8.1 Bulk Camera Query

**Endpoint:** `POST /api/rest/v1/cameras/bulk`

**Request:**
```http
POST http://192.168.1.9/api/rest/v1/cameras/bulk
Content-Type: application/json

{
  "cameraIds": [
    "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
    "9c4b3d2e-5f6a-7890-bcde-f12345678901",
    "ad5c4e3f-6g7b-8901-cdef-g23456789012"
  ],
  "fields": ["id", "name", "enabled", "recordingEnabled", "streams"]
}
```

**Response:**
```json
{
  "cameras": [
    { "id": "...", "name": "...", "enabled": true, ... },
    { "id": "...", "name": "...", "enabled": true, ... }
  ]
}
```

---

## 🚨 Error Responses

### Standard Error Format

```json
{
  "error": {
    "code": "CAMERA_NOT_FOUND",
    "message": "Camera with ID 8b3a2c1d-4e5f-6789-abcd-ef0123456789 not found",
    "statusCode": 404,
    "timestamp": "2025-10-25T14:30:00Z"
  }
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `INVALID_REQUEST` | Invalid request parameters |
| 401 | `UNAUTHORIZED` | Authentication required |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `CAMERA_NOT_FOUND` | Camera not found |
| 404 | `RECORDING_NOT_FOUND` | Recording not available |
| 409 | `CAMERA_OFFLINE` | Camera is offline |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

---

## 📊 Rate Limits

| Endpoint Category | Requests per Minute |
|-------------------|---------------------|
| Camera List | 60 |
| Individual Camera | 300 |
| Live Streaming | Unlimited |
| Recording Timeline | 100 |
| Playback | 500 concurrent |
| Export | 10 concurrent |
| PTZ Control | 60 |
| Events | 120 |

---

## 🔧 Implementation Examples

### Example 1: Get All Cameras and Stream URLs

```python
import requests
from requests.auth import HTTPBasicAuth

base_url = "http://192.168.1.9/api/rest/v1"
auth = HTTPBasicAuth('rta-integration', 'your-password')

# Get all cameras
response = requests.get(f"{base_url}/cameras", auth=auth)
cameras = response.json()['cameras']

# Get RTSP URLs for each camera
for camera in cameras:
    camera_id = camera['id']
    live_response = requests.get(f"{base_url}/cameras/{camera_id}/live", auth=auth)
    rtsp_url = live_response.json()['streams'][0]['url']
    print(f"Camera: {camera['name']} - RTSP: {rtsp_url}")
```

### Example 2: Query and Download Recording

```python
import requests
from datetime import datetime, timedelta

# Get recording timeline
camera_id = "8b3a2c1d-4e5f-6789-abcd-ef0123456789"
start_time = datetime.now() - timedelta(hours=2)
end_time = datetime.now()

timeline_url = f"{base_url}/recordings/{camera_id}/timeline"
params = {
    'startTime': start_time.isoformat() + 'Z',
    'endTime': end_time.isoformat() + 'Z'
}

timeline = requests.get(timeline_url, auth=auth, params=params).json()

# Download recording
download_url = f"{base_url}/recordings/{camera_id}/download"
download_params = {
    'startTime': start_time.isoformat() + 'Z',
    'endTime': end_time.isoformat() + 'Z',
    'format': 'mp4'
}

response = requests.get(download_url, auth=auth, params=download_params, stream=True)
with open('recording.mp4', 'wb') as f:
    for chunk in response.iter_content(chunk_size=8192):
        f.write(chunk)
```

### Example 3: Monitor Camera Events via WebSocket

```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://192.168.1.9/api/ws/events', {
  headers: {
    'Authorization': 'Basic ' + Buffer.from('username:password').toString('base64')
  }
});

ws.on('open', () => {
  console.log('Connected to Milestone event stream');

  // Subscribe to specific camera events
  ws.send(JSON.stringify({
    action: 'subscribe',
    cameraIds: ['8b3a2c1d-4e5f-6789-abcd-ef0123456789'],
    eventTypes: ['motion', 'tampering']
  }));
});

ws.on('message', (data) => {
  const event = JSON.parse(data);
  console.log('Event received:', event);

  if (event.type === 'camera.motion') {
    console.log(`Motion detected on camera ${event.cameraId}`);
  }
});
```

---

## 📝 Notes

1. **API Versioning**: Use `/api/rest/v1` for stable API. Check server capabilities for supported versions.

2. **Time Formats**: All timestamps use ISO 8601 format with UTC timezone: `2025-10-25T14:30:00Z`

3. **RTSP Authentication**: Include credentials in URL or use separate authentication mechanism.

4. **Concurrent Streams**: Maximum 500 concurrent RTSP streams. Monitor active connections.

5. **Export Expiration**: Exported files expire after 24 hours by default.

6. **WebSocket Reconnection**: Implement reconnection logic with exponential backoff.

7. **Pagination**: Large result sets are paginated. Use `limit` and `offset` parameters.

---

**Last Updated:** 2025-10-25
**API Version:** v1
**Milestone Server:** 192.168.1.9
