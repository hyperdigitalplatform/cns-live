# 🎬 Milestone Playback API - Detailed Explanation

## 📋 Overview

This document explains **exactly how** the Milestone Playback API works for retrieving recorded video for a specific timeline.

---

## 🔍 The API Call - Broken Down

### **Complete Example:**

```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/playback?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T11:00:00Z&speed=1"
```

---

## 📖 Part-by-Part Explanation

### **1. `curl`**
```bash
curl
```
**What it is:** Command-line tool for making HTTP requests

**Why we use it:** To send a request to the Milestone server

**In production:** Your application uses Python `requests` library or similar HTTP client instead of curl

---

### **2. `-u username:password`**
```bash
-u username:password
```
**What it is:** Authentication credentials

**Format:** `-u username:password`

**How it works:**
- curl automatically encodes this as HTTP Basic Authentication
- Sends header: `Authorization: Basic base64(username:password)`

**Real example:**
```bash
-u rta-integration:MySecurePass123
```

**What happens:**
1. curl converts `rta-integration:MySecurePass123` to Base64: `cnRhLWludGVncmF0aW9uOk15U2VjdXJlUGFzczEyMw==`
2. Sends HTTP header: `Authorization: Basic cnRhLWludGVncmF0aW9uOk15U2VjdXJlUGFzczEyMw==`
3. Milestone server decodes and validates credentials

---

### **3. Base URL**
```bash
http://192.168.1.9
```
**What it is:** Milestone server address

**Components:**
- `http://` - Protocol (use `https://` in production for security)
- `192.168.1.9` - IP address of Milestone server
- Could also be: `http://milestone.rta.gov.ae` (domain name)

---

### **4. API Path**
```bash
/api/rest/v1/recordings/{cameraId}/playback
```

**Breaking it down:**
- `/api` - Base API path
- `/rest` - RESTful API
- `/v1` - API version 1
- `/recordings` - Recordings module
- `/{cameraId}` - Specific camera identifier (replaced with actual GUID)
- `/playback` - Playback endpoint

**With actual camera ID:**
```bash
/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/playback
```

**What this means:**
"Give me the playback URL for recordings from camera with ID `8b3a2c1d-4e5f-6789-abcd-ef0123456789`"

---

### **5. Camera ID**
```bash
8b3a2c1d-4e5f-6789-abcd-ef0123456789
```

**What it is:** Unique identifier (GUID) for a camera in Milestone

**Format:** UUID/GUID (Globally Unique Identifier)
- 32 hexadecimal characters
- Separated by hyphens: `8-4-4-4-12` format

**How to get it:**
1. Call camera list API: `GET /api/rest/v1/cameras`
2. Response includes camera IDs:
```json
{
  "cameras": [
    {
      "id": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",  ← This is the camera ID
      "name": "Main Entrance Camera"
    }
  ]
}
```

**Important:** Each camera has a unique GUID assigned by Milestone

---

### **6. Query Parameters**
```bash
?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T11:00:00Z&speed=1
```

**Breaking it down:**

#### **a) `?` - Query String Start**
Indicates the start of query parameters

#### **b) `startTime=2025-10-24T10:00:00Z`**
**Parameter:** Start of playback timeline

**Format:** ISO 8601 timestamp
- `2025-10-24` - Date (YYYY-MM-DD)
- `T` - Separator between date and time
- `10:00:00` - Time (HH:MM:SS)
- `Z` - UTC timezone indicator

**Meaning:** "Start playing from October 24, 2025 at 10:00:00 AM UTC"

#### **c) `&` - Parameter Separator**
Separates multiple query parameters

#### **d) `endTime=2025-10-24T11:00:00Z`**
**Parameter:** End of playback timeline

**Meaning:** "Stop playing at October 24, 2025 at 11:00:00 AM UTC"

**Result:** Plays exactly 1 hour of video (10:00 AM to 11:00 AM)

#### **e) `&speed=1`**
**Parameter:** Playback speed multiplier

**Values:**
- `1` - Normal speed (real-time)
- `2` - 2x speed (double fast)
- `4` - 4x speed
- `8` - 8x speed
- `0.5` - Slow motion (half speed)

---

## 📥 What the API Returns

### **Response (JSON):**

```json
{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "cameraName": "Main Entrance Camera",
  "playbackUrl": "rtsp://192.168.1.9:554/playback/8b3a2c1d-4e5f-6789-abcd-ef0123456789?start=2025-10-24T10:00:00Z&end=2025-10-24T11:00:00Z&speed=1",
  "startTime": "2025-10-24T10:00:00Z",
  "endTime": "2025-10-24T11:00:00Z",
  "durationSeconds": 3600,
  "speed": 1,
  "protocol": "rtsp"
}
```

### **Response Fields Explained:**

#### **1. `cameraId`**
```json
"cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789"
```
**What it is:** Echo of the camera ID from the request

**Purpose:** Confirm which camera this playback URL is for

---

#### **2. `cameraName`**
```json
"cameraName": "Main Entrance Camera"
```
**What it is:** Human-readable camera name

**Purpose:** Display to users in dashboard (easier to read than GUID)

---

#### **3. `playbackUrl` ⭐ MOST IMPORTANT**
```json
"playbackUrl": "rtsp://192.168.1.9:554/playback/8b3a2c1d-4e5f-6789-abcd-ef0123456789?start=2025-10-24T10:00:00Z&end=2025-10-24T11:00:00Z&speed=1"
```

**What it is:** RTSP streaming URL for the recorded video

**Purpose:** This is what you actually use to stream the video

**Format breakdown:**
```
rtsp://192.168.1.9:554/playback/{cameraId}?start={timestamp}&end={timestamp}&speed={speed}
```

**Components:**
- `rtsp://` - RTSP protocol (Real-Time Streaming Protocol)
- `192.168.1.9` - Milestone server IP
- `:554` - RTSP port (standard RTSP port)
- `/playback/` - Playback path
- `{cameraId}` - Camera identifier
- `?start=...` - Start timestamp
- `&end=...` - End timestamp
- `&speed=1` - Playback speed

**This URL is what you pass to:**
- VLC player
- ffmpeg/ffplay
- MediaMTX for re-streaming
- Your video player in dashboard

---

#### **4. `startTime`**
```json
"startTime": "2025-10-24T10:00:00Z"
```
**What it is:** Confirmed start time of playback

**Purpose:** Echo of what you requested, confirms Milestone accepted the timestamp

---

#### **5. `endTime`**
```json
"endTime": "2025-10-24T11:00:00Z"
```
**What it is:** Confirmed end time of playback

**Purpose:** Confirms the timeline range

---

#### **6. `durationSeconds`**
```json
"durationSeconds": 3600
```
**What it is:** Total duration in seconds

**Calculation:**
```
endTime - startTime = 1 hour = 3600 seconds
```

**Purpose:** Know how long the video will play

**Examples:**
- 5 minutes = 300 seconds
- 30 minutes = 1800 seconds
- 1 hour = 3600 seconds
- 2 hours = 7200 seconds

---

#### **7. `speed`**
```json
"speed": 1
```
**What it is:** Playback speed multiplier

**Purpose:** Confirms the speed setting

---

#### **8. `protocol`**
```json
"protocol": "rtsp"
```
**What it is:** Streaming protocol type

**Purpose:** Indicates this is an RTSP stream (could also be HTTP, HLS, etc.)

---

## 🐍 Python Implementation Explained

### **The Code:**

```python
response = requests.get(
    f"http://192.168.1.9/api/rest/v1/recordings/{camera_id}/playback",
    params={
        "startTime": start_time,
        "endTime": end_time,
        "speed": 1
    },
    auth=('username', 'password')
)
```

### **Breaking It Down:**

#### **1. `requests.get()`**
```python
requests.get(...)
```
**What it is:** Python function to make HTTP GET request

**Library:** `import requests`

**Purpose:** Same as `curl` but in Python

---

#### **2. URL with f-string**
```python
f"http://192.168.1.9/api/rest/v1/recordings/{camera_id}/playback"
```

**What it is:** Python f-string (formatted string)

**How it works:**
```python
camera_id = "8b3a2c1d-4e5f-6789-abcd-ef0123456789"

# f-string replaces {camera_id} with actual value
url = f"http://192.168.1.9/api/rest/v1/recordings/{camera_id}/playback"

# Result:
# "http://192.168.1.9/api/rest/v1/recordings/8b3a2c1d-4e5f-6789-abcd-ef0123456789/playback"
```

---

#### **3. `params` Dictionary**
```python
params={
    "startTime": start_time,
    "endTime": end_time,
    "speed": 1
}
```

**What it is:** Query parameters as Python dictionary

**How it works:**
```python
start_time = "2025-10-24T10:00:00Z"
end_time = "2025-10-24T11:00:00Z"

params = {
    "startTime": "2025-10-24T10:00:00Z",
    "endTime": "2025-10-24T11:00:00Z",
    "speed": 1
}

# requests library automatically converts to:
# ?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T11:00:00Z&speed=1

# And handles URL encoding automatically!
```

**Benefits:**
- ✅ Cleaner code
- ✅ Automatic URL encoding
- ✅ Handles special characters
- ✅ Easy to modify parameters

---

#### **4. `auth` Tuple**
```python
auth=('username', 'password')
```

**What it is:** Authentication credentials as Python tuple

**How it works:**
```python
auth = ('rta-integration', 'MySecurePass123')

# requests library automatically:
# 1. Combines username:password
# 2. Encodes to Base64
# 3. Adds Authorization header
```

**Equivalent to curl's:**
```bash
-u username:password
```

---

### **Complete Working Example:**

```python
import requests
from datetime import datetime, timedelta

# Configuration
MILESTONE_SERVER = "http://192.168.1.9"
MILESTONE_USER = "rta-integration"
MILESTONE_PASS = "MySecurePass123"

# Camera and timeline
camera_id = "8b3a2c1d-4e5f-6789-abcd-ef0123456789"
start_time = "2025-10-24T10:00:00Z"
end_time = "2025-10-24T11:00:00Z"

# Make API request
response = requests.get(
    f"{MILESTONE_SERVER}/api/rest/v1/recordings/{camera_id}/playback",
    params={
        "startTime": start_time,
        "endTime": end_time,
        "speed": 1
    },
    auth=(MILESTONE_USER, MILESTONE_PASS)
)

# Check if request was successful
if response.status_code == 200:
    data = response.json()

    # Extract the RTSP URL
    rtsp_url = data['playbackUrl']

    print(f"Camera: {data['cameraName']}")
    print(f"Duration: {data['durationSeconds']} seconds")
    print(f"RTSP URL: {rtsp_url}")

    # Now you can:
    # 1. Stream via MediaMTX
    # 2. Play in VLC
    # 3. Process with ffmpeg
    # 4. Display in dashboard

else:
    print(f"Error: {response.status_code}")
    print(response.text)
```

**Output:**
```
Camera: Main Entrance Camera
Duration: 3600 seconds
RTSP URL: rtsp://192.168.1.9:554/playback/8b3a2c1d-4e5f-6789-abcd-ef0123456789?start=2025-10-24T10:00:00Z&end=2025-10-24T11:00:00Z&speed=1
```

---

## 🔄 Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│  Step 1: Your Application (RTA CCTV System)                 │
│                                                              │
│  User selects:                                               │
│  - Camera: "Main Entrance"                                   │
│  - Timeline: Oct 24, 10:00 AM - 11:00 AM                    │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ 1. Make API request
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 2: Call Milestone API                                 │
│                                                              │
│  GET http://192.168.1.9/api/rest/v1/recordings/             │
│      8b3a2c1d-4e5f-6789-abcd-ef0123456789/playback          │
│      ?startTime=2025-10-24T10:00:00Z                        │
│      &endTime=2025-10-24T11:00:00Z                          │
│      &speed=1                                                │
│                                                              │
│  Authorization: Basic <credentials>                          │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ 2. Milestone processes request
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 3: Milestone Server Response                          │
│                                                              │
│  {                                                           │
│    "playbackUrl": "rtsp://192.168.1.9:554/playback/...",   │
│    "durationSeconds": 3600,                                  │
│    ...                                                       │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ 3. Extract RTSP URL
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 4: Use RTSP URL                                        │
│                                                              │
│  rtsp://192.168.1.9:554/playback/...                        │
│                                                              │
│  Options:                                                    │
│  A) Stream directly to dashboard                            │
│  B) Re-stream via MediaMTX                                  │
│  C) Convert to HLS/WebRTC                                   │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ 4. Stream video
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 5: User Watches Playback                              │
│                                                              │
│  Dashboard displays video timeline:                          │
│  ────────────────▶──────────────────                        │
│  10:00 AM    10:30 AM    11:00 AM                           │
│                                                              │
│  [Play] [Pause] [Speed: 1x ▼] [Timeline Scrubber]          │
└─────────────────────────────────────────────────────────────┘
```

---

## 💡 Real-World Usage in RTA System

### **Your Playback Service (Port 8090) Implementation:**

```python
from flask import Flask, jsonify, request
import requests

app = Flask(__name__)

@app.route('/api/v1/playback', methods=['GET'])
def get_playback():
    """
    RTA CCTV Playback API endpoint

    Query Parameters:
        camera_id: Milestone camera GUID
        start_time: ISO 8601 timestamp
        end_time: ISO 8601 timestamp

    Returns:
        Stream URL for dashboard
    """
    # 1. Get parameters from RTA dashboard
    camera_id = request.args.get('camera_id')
    start_time = request.args.get('start_time')
    end_time = request.args.get('end_time')

    # 2. Call Milestone API
    milestone_response = requests.get(
        f"http://192.168.1.9/api/rest/v1/recordings/{camera_id}/playback",
        params={
            "startTime": start_time,
            "endTime": end_time,
            "speed": 1
        },
        auth=('rta-integration', 'password')
    )

    if milestone_response.status_code != 200:
        return jsonify({"error": "Failed to get playback from Milestone"}), 500

    data = milestone_response.json()
    rtsp_url = data['playbackUrl']

    # 3. Re-stream via MediaMTX for dashboard
    # MediaMTX will convert RTSP to WebRTC for low-latency browser playback
    mediamtx_stream_id = f"playback_{camera_id}_{start_time}"

    # 4. Return stream URLs for dashboard
    return jsonify({
        "camera_id": camera_id,
        "camera_name": data['cameraName'],
        "timeline": {
            "start": start_time,
            "end": end_time,
            "duration": data['durationSeconds']
        },
        "streams": {
            "webrtc": f"http://mediamtx:8889/{mediamtx_stream_id}/whep",
            "hls": f"http://mediamtx:8888/{mediamtx_stream_id}/index.m3u8",
            "rtsp": rtsp_url  # Original Milestone RTSP URL
        }
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8090)
```

### **Dashboard Usage:**

```javascript
// In your React dashboard
async function playbackVideo(cameraId, startTime, endTime) {
  // Call RTA Playback Service
  const response = await fetch(
    `http://localhost:8090/api/v1/playback?camera_id=${cameraId}&start_time=${startTime}&end_time=${endTime}`
  );

  const data = await response.json();

  // Use WebRTC stream for low latency
  const streamUrl = data.streams.webrtc;

  // Initialize video player with stream URL
  videoPlayer.play(streamUrl);

  // Show timeline
  timeline.setDuration(data.timeline.duration);
  timeline.setStartTime(data.timeline.start);
}
```

---

## ✅ Summary

**What the API does:**
1. You ask Milestone: "Give me playback URL for camera X from time A to time B"
2. Milestone responds: "Here's the RTSP URL to stream that recorded video"
3. You use that RTSP URL to stream the video

**Key Points:**
- ✅ API returns a **URL**, not the video itself
- ✅ You need to stream from that RTSP URL
- ✅ Timeline is specified by `startTime` and `endTime`
- ✅ Response tells you duration and camera details
- ✅ RTSP URL can be used with any RTSP client

**The RTSP URL is like a Netflix link** - it's a streaming URL that plays video, not a downloaded file.

---

**Last Updated:** 2025-10-25
