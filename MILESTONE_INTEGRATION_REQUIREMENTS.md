# 🎬 Milestone XProtect Expert VMS - Integration Requirements

## 📋 Overview

This document outlines the requirements and expectations from the Milestone XProtect Expert VMS server running at **192.168.1.9** for integration with the RTA CCTV System.

---

## 🔧 Milestone XProtect Expert Server Details

**Server Information:**
- **IP Address**: 192.168.1.9
- **VMS Type**: Milestone XProtect Expert
- **Purpose**: Camera management, configuration storage, and video recording storage

---

## 📡 Required Milestone Services & Ports

### 1. Management Server
**Port**: 80/443 (HTTP/HTTPS)
- Used for: Camera configuration retrieval, management API access
- Required for: Fetching camera list, configurations, metadata

### 2. Recording Server
**Port**: 7563 (default)
- Used for: Video data retrieval
- Required for: Playback of recorded video

### 3. RTSP Streaming (per camera)
**Port**: 554 (default RTSP)
- Used for: Live video streaming from cameras
- Required for: Real-time video feed distribution
- Format: `rtsp://192.168.1.9:554/camera/<camera-id>`

### 4. Mobile Server (if available)
**Port**: 8081 (default)
- Used for: Mobile/web API access
- Optional but recommended for easier integration

---

## 🎥 Camera Information Required

The Milestone server must provide the following information for each camera:

### 1. Camera Metadata
```json
{
  "camera_id": "unique-camera-identifier",
  "camera_name": "Camera Name/Description",
  "location": "Physical location",
  "recording_server": "192.168.1.9",
  "enabled": true,
  "properties": {
    "manufacturer": "Camera manufacturer",
    "model": "Camera model",
    "resolution": "1920x1080",
    "fps": 25,
    "codec": "H.264"
  }
}
```

### 2. Camera Streaming URLs
Each camera should expose:

**RTSP Live Stream:**
```
rtsp://192.168.1.9:554/camera/<camera-id>
```

**Alternative formats (if available):**
```
rtsp://192.168.1.9:554/<camera-name>
rtsp://192.168.1.9:554/<camera-guid>
```

### 3. Camera Groups/Categories
Cameras should be organized by agency:
- **Dubai Police** - Up to 50 cameras
- **Metro** - Up to 30 cameras
- **Bus** - Up to 20 cameras
- **Other Agencies** - Up to 400 cameras

---

## 📼 Recording & Playback Requirements

### 1. Recording Storage Structure

Milestone should maintain recordings with:
- **Retention Period**: Minimum 90 days
- **Recording Format**: H.264 video, AAC audio (if available)
- **Frame Rate**: 25 FPS (configurable per camera)
- **Resolution**: 1080p (1920x1080) preferred, minimum 720p

### 2. Recording Metadata

For each recording segment:
```json
{
  "recording_id": "unique-recording-id",
  "camera_id": "camera-identifier",
  "start_time": "2025-10-24T10:00:00Z",
  "end_time": "2025-10-24T11:00:00Z",
  "duration_seconds": 3600,
  "file_size_bytes": 1073741824,
  "codec": "H.264",
  "resolution": "1920x1080",
  "fps": 25,
  "has_audio": false
}
```

### 3. Time-based Playback

The system must support:
- **Start Time Query**: Get recording starting from specific timestamp
- **Time Range Query**: Get all recordings within date/time range
- **Continuous Playback**: Stream recordings across multiple segments
- **Seek Operations**: Jump to specific timestamp within recording

---

## 🔌 Integration Methods

### Method 1: MIP SDK (Recommended)

**Milestone Integration Platform SDK**

Requirements:
- MIP SDK .NET libraries
- SDK credentials and license
- Access to Milestone Management Server API

**API Endpoints Needed:**
```
# Get all cameras
GET http://192.168.1.9/api/cameras

# Get camera details
GET http://192.168.1.9/api/cameras/{camera-id}

# Get recording timeline
GET http://192.168.1.9/api/recordings/{camera-id}?start={timestamp}&end={timestamp}

# Export recording
POST http://192.168.1.9/api/export
```

### Method 2: RTSP Direct Streaming

**For Live Streams:**
```bash
# Connect directly to camera RTSP stream via Milestone
rtsp://192.168.1.9:554/camera/<camera-id>
```

**For Playback:**
```bash
# Time-shifted RTSP playback
rtsp://192.168.1.9:554/playback/<camera-id>?start=<timestamp>&end=<timestamp>
```

### Method 3: Mobile Server API

**If Mobile Server is enabled:**
```
# Camera list
GET http://192.168.1.9:8081/cameras

# Live stream
GET http://192.168.1.9:8081/stream/{camera-id}

# Playback
GET http://192.168.1.9:8081/playback/{camera-id}?start={timestamp}&duration={seconds}
```

---

## 🔐 Authentication & Authorization

### Required Credentials

**Management Server Access:**
- Username: `<milestone-admin-user>`
- Password: `<milestone-admin-password>`
- Domain: (if Active Directory integrated)

**RTSP Access:**
- May require Basic Authentication
- Format: `rtsp://username:password@192.168.1.9:554/camera/<camera-id>`

**API Access:**
- Token-based authentication (preferred)
- Basic authentication (fallback)
- Windows Authentication (if AD integrated)

### Permissions Required

The integration account needs:
- **Read access** to all camera configurations
- **Streaming access** to all camera live feeds
- **Playback access** to recorded video
- **Export permissions** for video clips
- **Metadata read access** for camera properties

---

## 📊 Expected Data Formats

### Camera List Response (Example)

```json
{
  "cameras": [
    {
      "id": "cam-001",
      "name": "Main Entrance - Camera 1",
      "enabled": true,
      "recording_enabled": true,
      "live_stream_url": "rtsp://192.168.1.9:554/camera/cam-001",
      "location": {
        "building": "Headquarters",
        "floor": "Ground",
        "area": "Main Entrance"
      },
      "properties": {
        "manufacturer": "Axis",
        "model": "P3245-LVE",
        "ip_address": "192.168.1.101",
        "resolution": "1920x1080",
        "fps": 25,
        "ptz_capable": true
      },
      "recording_server": "192.168.1.9",
      "retention_days": 90
    }
  ]
}
```

### Recording Timeline Response (Example)

```json
{
  "camera_id": "cam-001",
  "timeline": [
    {
      "sequence_id": "seq-001",
      "start_time": "2025-10-24T00:00:00Z",
      "end_time": "2025-10-24T06:00:00Z",
      "duration_seconds": 21600,
      "has_video": true,
      "has_audio": false,
      "codec": "H.264",
      "resolution": "1920x1080",
      "fps": 25,
      "gaps": []
    },
    {
      "sequence_id": "seq-002",
      "start_time": "2025-10-24T06:00:00Z",
      "end_time": "2025-10-24T12:00:00Z",
      "duration_seconds": 21600,
      "has_video": true,
      "has_audio": false,
      "codec": "H.264",
      "resolution": "1920x1080",
      "fps": 25,
      "gaps": []
    }
  ]
}
```

---

## 🛠️ Configuration Steps on Milestone Server

### 1. Enable External API Access

In Milestone Management Client:
1. Navigate to **Tools** > **Options**
2. Enable **Mobile Server** (if using)
3. Enable **MIP SDK Access**
4. Configure **External Ports** (7563 for recording, 554 for RTSP)

### 2. Create Integration User Account

1. Go to **Security** > **Users and Roles**
2. Create new user: `rta-integration`
3. Assign role with permissions:
   - View all cameras
   - Playback recordings
   - Export video
   - Access via API

### 3. Configure Camera Access

For each camera:
1. Enable **Recording**
2. Set **Retention** to 90 days minimum
3. Enable **Live Streaming**
4. Configure **Stream Settings**:
   - Codec: H.264
   - Resolution: 1920x1080
   - Frame Rate: 25 FPS
   - Bitrate: 2-4 Mbps

### 4. Enable RTSP Streaming

1. Navigate to **Recording Server** settings
2. Enable **RTSP Streaming**
3. Configure **RTSP Port**: 554 (or custom)
4. Set **Authentication**: Basic or None (based on security requirements)

### 5. Configure Recording Settings

1. Set **Recording Mode**: Continuous or Motion Detection
2. Configure **Pre/Post Recording**: 5 seconds before/after events
3. Set **Storage Location**: Adequate storage for 90-day retention
4. Enable **Archiving** if needed

---

## 🔍 Testing Milestone Integration

### Test 1: Verify Camera List Access

```bash
# Using MIP SDK or Mobile Server API
curl -u username:password http://192.168.1.9:8081/cameras

# Expected: JSON list of all cameras
```

### Test 2: Test RTSP Live Stream

```bash
# Using ffplay or VLC
ffplay rtsp://username:password@192.168.1.9:554/camera/cam-001

# Or with VLC
vlc rtsp://192.168.1.9:554/camera/cam-001
```

### Test 3: Verify Recording Playback

```bash
# Query recordings for a specific time range
curl -u username:password \
  "http://192.168.1.9:8081/recordings/cam-001?start=2025-10-24T00:00:00Z&end=2025-10-24T23:59:59Z"
```

### Test 4: Test Time-based Playback

```bash
# Play recording from specific timestamp
ffplay "rtsp://192.168.1.9:554/playback/cam-001?start=2025-10-24T10:00:00Z&duration=3600"
```

---

## 📝 Configuration in RTA CCTV System

### Update VMS Service Configuration

Edit `services/vms-service/config/config.yaml`:

```yaml
milestone:
  # Milestone server details
  server_url: "http://192.168.1.9"
  rtsp_base_url: "rtsp://192.168.1.9:554"
  recording_server: "192.168.1.9:7563"

  # Authentication
  username: "rta-integration"
  password: "your-secure-password"
  domain: ""  # Leave empty if not using AD

  # API endpoints
  api:
    cameras: "/api/cameras"
    recordings: "/api/recordings"
    export: "/api/export"

  # Streaming settings
  streaming:
    rtsp_port: 554
    rtsp_path_template: "/camera/{camera_id}"
    playback_path_template: "/playback/{camera_id}"

  # Integration method
  method: "mip_sdk"  # Options: mip_sdk, mobile_server, rtsp_direct
```

### Environment Variables

Update `.env` file:

```bash
# Milestone VMS Configuration
MILESTONE_SERVER_URL=http://192.168.1.9
MILESTONE_RTSP_URL=rtsp://192.168.1.9:554
MILESTONE_USERNAME=rta-integration
MILESTONE_PASSWORD=your-secure-password
MILESTONE_API_PORT=8081
MILESTONE_RECORDING_PORT=7563
```

---

## 🚨 Common Issues & Solutions

### Issue 1: Cannot Connect to Milestone Server

**Symptoms:**
- Connection timeout
- "Unable to reach server" errors

**Solutions:**
1. Verify network connectivity: `ping 192.168.1.9`
2. Check firewall rules on Milestone server
3. Ensure required ports are open (80/443, 554, 7563, 8081)
4. Verify Milestone services are running

### Issue 2: Authentication Failed

**Symptoms:**
- 401 Unauthorized errors
- "Invalid credentials" messages

**Solutions:**
1. Verify username and password are correct
2. Check user has required permissions in Milestone
3. Confirm API access is enabled
4. Test credentials using Milestone Smart Client

### Issue 3: RTSP Stream Not Accessible

**Symptoms:**
- RTSP connection refused
- Stream timeout

**Solutions:**
1. Verify RTSP is enabled on Recording Server
2. Check RTSP port (default 554)
3. Test stream using VLC player directly
4. Verify camera is recording and online

### Issue 4: No Recordings Found

**Symptoms:**
- Empty recording timeline
- "No recordings available" errors

**Solutions:**
1. Verify camera is set to record continuously
2. Check retention period hasn't expired
3. Confirm storage has adequate space
4. Verify time range in query is correct

---

## 📊 Performance Considerations

### Network Bandwidth

**Per Camera:**
- Live Stream: 2-4 Mbps (H.264 @ 1080p, 25fps)
- Playback Stream: 2-4 Mbps

**Total for 500 Cameras:**
- Peak Bandwidth: 2 Gbps (if all streaming)
- Typical Usage: 200-500 Mbps (50-100 concurrent streams)

### Storage Requirements

**Per Camera (90-day retention):**
- Continuous Recording: ~1.5-2 TB per camera
- Motion-Based: ~300-500 GB per camera

**Total for 500 Cameras:**
- Continuous: 750-1000 TB
- Motion-Based: 150-250 TB

### API Request Limits

Milestone server should support:
- **Camera List Queries**: 1 request per minute
- **Recording Queries**: 100 requests per minute
- **Stream Requests**: 500 concurrent streams
- **Export Requests**: 10 concurrent exports

---

## ✅ Pre-Integration Checklist

Before integrating with RTA CCTV System:

- [ ] Milestone XProtect Expert is installed and running on 192.168.1.9
- [ ] All cameras are added and configured in Milestone
- [ ] Recording is enabled for all cameras
- [ ] Retention is set to minimum 90 days
- [ ] Integration user account created with proper permissions
- [ ] RTSP streaming is enabled on Recording Server
- [ ] Mobile Server API is enabled (if using)
- [ ] Network connectivity verified between RTA system and Milestone server
- [ ] Required ports are open in firewall
- [ ] Tested RTSP stream access from RTA system network
- [ ] Verified recording playback works
- [ ] Documented all camera IDs and RTSP URLs

---

## 📞 Support & Documentation

**Milestone Documentation:**
- MIP SDK Documentation: https://doc.developer.milestonesys.com/
- XProtect Expert Manual: Check Milestone documentation portal
- API Reference: Available in MIP SDK installation

**Testing Tools:**
- VLC Media Player: For RTSP stream testing
- ffmpeg/ffplay: For stream analysis and testing
- Postman: For API endpoint testing
- Milestone Smart Client: For verification

---

## 🔄 Next Steps

1. **Configure Milestone Server** according to specifications above
2. **Create Integration User** with required permissions
3. **Document Camera URLs** for all cameras in the system
4. **Update RTA Configuration** with Milestone server details
5. **Test Integration** using verification steps
6. **Monitor Performance** during initial deployment
7. **Adjust Settings** based on performance metrics

---

## 🎯 WHIP Streaming Integration

### Overview

The RTA CCTV System uses **WHIP (WebRTC HTTP Ingestion Protocol)** for ultra-low latency camera streaming (~450ms vs 2-4s with HLS).

### Architecture Flow

```
Milestone VMS → MediaMTX (RTSP) → GStreamer WHIP Pusher → LiveKit WHIP Ingress → LiveKit SFU → Viewers
```

### How WHIP Integration Works with Milestone

1. **MediaMTX pulls RTSP from Milestone**:
   - MediaMTX connects to Milestone RTSP endpoints
   - Provides stable, buffered RTSP streams
   - Acts as RTSP proxy between Milestone and WHIP pushers

2. **WHIP Pusher Containers**:
   - One container per active camera stream
   - Pulls RTSP from MediaMTX
   - Pushes to LiveKit via WHIP protocol
   - Handles H.264 and H.265 codecs (transcodes H.265 to H.264)

3. **LiveKit WHIP Ingress**:
   - Receives WHIP streams from GStreamer
   - Publishes to LiveKit rooms
   - Distributes to viewers via WebRTC

### Testing WHIP with Milestone Cameras

#### Test 1: Verify MediaMTX can access Milestone RTSP

```bash
# Test direct RTSP connection from MediaMTX to Milestone
docker exec cctv-mediamtx-1 ffprobe -v error -show_entries stream=codec_name,width,height \
  rtsp://username:password@192.168.1.9:554/camera/cam-001
```

**Expected Output**:
```
[STREAM]
codec_name=h264
width=1920
height=1080
[/STREAM]
```

#### Test 2: Reserve a Stream and Verify WHIP Pusher Container

```bash
# Reserve camera stream
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "cam-001-sheikh-zayed",
    "user_id": "test-user",
    "quality": "medium"
  }'
```

**Expected Response**:
```json
{
  "reservation_id": "res-xyz",
  "ingress_id": "IN_abc123",
  "whip_url": "http://livekit-ingress:8080/w/stream_key_xyz",
  "livekit_token": "eyJhbGc...",
  "room_name": "camera_cam-001-sheikh-zayed",
  "expires_at": "2025-10-26T..."
}
```

**Verify WHIP Pusher Container Started**:
```bash
docker ps --filter "name=whip-pusher-cam-001"
```

**Expected**: Container running with name like `whip-pusher-cam-001-sheikh-zayed`

#### Test 3: Check WHIP Pusher Logs

```bash
# View WHIP pusher logs
docker logs whip-pusher-cam-001-sheikh-zayed
```

**Expected Output (Successful)**:
```
Starting WHIP Pusher...
RTSP Source: rtsp://mediamtx:8554/camera_cam-001-sheikh-zayed
WHIP Endpoint: http://livekit-ingress:8080/w/stream_key_xyz
Stream Key: stream_key_xyz
Setting pipeline to PAUSED ...
Pipeline is PREROLLING ...
Pipeline is PREROLLED ...
Setting pipeline to PLAYING ...
New clock: GstSystemClock
Redistribute latency...
```

**Expected Output (After Connection)**:
```
Got message #XX from element "whipsink0" (state-changed)
whipsink: Pushing packets to LiveKit at 1.79 Mbps
```

#### Test 4: Verify Different Cameras Show Different Feeds

```bash
# Reserve Camera 1
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{"camera_id": "cam-001-sheikh-zayed", "user_id": "user1", "quality": "medium"}'

# Reserve Camera 2
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{"camera_id": "cam-002-metro-station", "user_id": "user2", "quality": "medium"}'
```

**Verify**:
1. Two separate WHIP pusher containers running
2. Each connected to different RTSP endpoint in MediaMTX
3. Each pushing to different LiveKit room
4. Viewers see different video feeds for each camera

#### Test 5: Release Stream and Verify Cleanup

```bash
# Release stream
curl -X DELETE "http://localhost:8088/api/v1/stream/release/{reservation_id}"
```

**Expected**:
- WHIP pusher container stops and is removed
- LiveKit ingress is deleted
- Resources freed

### Common WHIP Integration Issues

#### Issue: WHIP Pusher Container Fails with "no element 'whipsink'"

**Cause**: WHIP pusher image doesn't have gst-plugins-rs installed

**Solution**:
```bash
# Rebuild WHIP pusher image
cd services/whip-pusher
docker build -t whip-pusher:latest .
```

#### Issue: MediaMTX Cannot Connect to Milestone RTSP

**Symptoms**: WHIP pusher logs show RTSP connection errors

**Solution**:
1. Verify Milestone RTSP URL format is correct
2. Check Milestone credentials
3. Ensure Milestone RTSP port 554 is accessible
4. Test with VLC: `vlc rtsp://192.168.1.9:554/camera/cam-001`

#### Issue: Both Cameras Show Same Feed

**Symptoms**: Multiple cameras display identical video

**Root Cause**: MediaMTX path configuration or WHIP pusher using wrong RTSP URL

**Solution**:
1. Check MediaMTX configuration: each camera should have unique path
2. Verify WHIP pusher environment variable `RTSP_URL` is different per container
3. Check go-api stream reservation logic assigns correct camera ID

#### Issue: H.265 Cameras Not Streaming

**Symptoms**: WHIP pusher fails with "delayed linking failed" or "streaming stopped, reason not-linked"

**Root Cause**: Camera uses H.265 codec, pipeline only supports H.264

**Solution**: The current GStreamer pipeline already handles this:
```bash
# Pipeline uses decodebin + x264enc for universal codec support
rtspsrc ! application/x-rtp,media=video ! rtpjitterbuffer ! decodebin ! x264enc ! ...
```

Verify `decodebin` and `x264enc` are present in pipeline.

### WHIP Performance Metrics from Milestone

**Typical Performance** (tested with Milestone cameras):

| Camera Codec | Resolution | Bitrate | CPU per Pusher | Latency |
|--------------|------------|---------|----------------|---------|
| H.264 | 1920x1080 | 2 Mbps | ~15% | ~450ms |
| H.265 | 1920x1080 | 1 Mbps | ~20% | ~500ms |
| H.264 | 1280x720 | 1.5 Mbps | ~10% | ~400ms |

**Network Requirements**:
- MediaMTX ↔ Milestone: 2-4 Mbps per camera
- WHIP Pusher ↔ LiveKit: 2-4 Mbps per camera
- Total per camera: 4-8 Mbps

### Milestone RTSP Configuration for WHIP

**Recommended Milestone Settings**:

1. **Enable RTSP Streaming** on Recording Server
2. **Set Stream Profile**:
   - Codec: H.264 (preferred) or H.265
   - Resolution: 1920x1080 or 1280x720
   - Frame Rate: 25 FPS
   - Bitrate: 2-4 Mbps (constant bitrate preferred)
3. **Enable TCP Transport** (more reliable than UDP)
4. **Set Authentication**: Basic Auth or None (internal network)

**RTSP URL Format for MediaMTX**:
```
rtsp://milestone-username:password@192.168.1.9:554/camera/<camera-id>
```

---

**Last Updated:** 2025-10-26
**Milestone Server IP:** 192.168.1.9
**Integration Status:** WHIP Integration Active
**Tested Cameras:** cam-001-sheikh-zayed (H.264), cam-002-metro-station (H.265)
