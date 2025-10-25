# 🚀 Milestone XProtect Integration - Quick Start Guide

## 📋 Overview

This guide provides a streamlined path to integrate your Milestone XProtect Expert VMS (192.168.1.9) with the RTA CCTV System.

---

## 📚 Documentation Index

Your complete Milestone integration documentation:

1. **[MILESTONE_INTEGRATION_REQUIREMENTS.md](./MILESTONE_INTEGRATION_REQUIREMENTS.md)**
   - What needs to be configured on the Milestone server
   - Required services, ports, and permissions
   - Pre-integration checklist

2. **[MILESTONE_API_REFERENCE.md](./MILESTONE_API_REFERENCE.md)**
   - Complete API endpoint reference
   - 8 categories of APIs with examples
   - Camera management, streaming, playback, PTZ, events

3. **[PLAYBACK_VS_DOWNLOAD_VS_EXPORT.md](./PLAYBACK_VS_DOWNLOAD_VS_EXPORT.md)**
   - Understanding three methods of video retrieval
   - When to use each method
   - Technical differences explained

4. **[MILESTONE_PLAYBACK_API_EXPLAINED.md](./MILESTONE_PLAYBACK_API_EXPLAINED.md)**
   - Detailed breakdown of playback API
   - Every component explained
   - Python implementation examples

5. **[VERIFICATION_GUIDE.md](./VERIFICATION_GUIDE.md)**
   - Step-by-step testing procedures
   - Camera registration, live streaming, playback
   - End-to-end verification scenarios

---

## ⚡ Quick Start (5 Steps)

### Step 1: Configure Milestone Server (15 minutes)

On your Milestone XProtect Expert server (192.168.1.9):

1. **Create Integration User**
   ```
   Navigate to: Security > Users and Roles
   Create user: rta-integration
   Permissions:
     - View all cameras
     - Playback recordings
     - Export video
     - Access via API
   ```

2. **Enable API Access**
   ```
   Navigate to: Tools > Options
   ☑ Enable Mobile Server (port 8081)
   ☑ Enable MIP SDK Access
   ☑ Enable RTSP Streaming (port 554)
   ```

3. **Configure Recording Server**
   ```
   Navigate to: Recording Server settings
   ☑ Enable RTSP Streaming
   Port: 554
   Authentication: Basic
   ```

4. **Verify Camera Settings**
   - All cameras recording continuously
   - Retention: 90 days minimum
   - Codec: H.264
   - Resolution: 1920x1080
   - Frame Rate: 25 FPS

### Step 2: Update RTA System Configuration (5 minutes)

Edit `.env` file in RTA CCTV project:

```bash
# Milestone VMS Configuration
MILESTONE_SERVER_URL=http://192.168.1.9
MILESTONE_RTSP_URL=rtsp://192.168.1.9:554
MILESTONE_USERNAME=rta-integration
MILESTONE_PASSWORD=your-secure-password
MILESTONE_API_PORT=8081
MILESTONE_RECORDING_PORT=7563
```

### Step 3: Test Milestone Connection (5 minutes)

```bash
# Test 1: Verify network connectivity
ping 192.168.1.9

# Test 2: Test API access
curl -u rta-integration:your-password http://192.168.1.9:8081/cameras

# Test 3: Test RTSP stream (replace {camera-id} with actual camera ID)
ffplay rtsp://rta-integration:your-password@192.168.1.9:554/camera/{camera-id}
```

**Expected Results:**
- ✅ Ping responds
- ✅ API returns JSON list of cameras
- ✅ ffplay shows live video stream

### Step 4: Register Cameras in RTA System (2 minutes per camera)

```bash
# Register your first camera
curl -X POST http://localhost:8081/api/v1/cameras \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "cam-main-entrance",
    "name": "Main Entrance Camera",
    "location": "Building A - Main Entrance",
    "rtsp_url": "rtsp://192.168.1.9:554/camera/{milestone-camera-id}",
    "agency": "dubai_police",
    "enabled": true,
    "metadata": {
      "resolution": "1920x1080",
      "fps": 25,
      "manufacturer": "Axis",
      "model": "P3245-LVE"
    }
  }'
```

**Response:**
```json
{
  "camera_id": "cam-main-entrance",
  "status": "registered",
  "stream_url": "rtsp://localhost:8554/dubai_police_cam-main-entrance"
}
```

### Step 5: Verify Live Streaming (2 minutes)

1. **Open RTA Dashboard**
   ```
   http://localhost:3000
   ```

2. **Navigate to Live View**
   - Click "Live View" in navigation menu
   - Select grid size (2x2, 3x3, or 4x4)

3. **Add Camera to Grid**
   - Click empty grid cell
   - Select "Main Entrance Camera"
   - Stream should start playing immediately

---

## 🎥 Common Usage Scenarios

### Scenario 1: View Live Stream

**Dashboard:**
```
1. Open http://localhost:3000
2. Click "Live View"
3. Select camera from dropdown
4. Stream plays automatically
```

**Direct RTSP:**
```bash
ffplay rtsp://localhost:8554/dubai_police_cam-main-entrance
```

**HLS (Browser):**
```
http://localhost:8888/dubai_police_cam-main-entrance/index.m3u8
```

---

### Scenario 2: Playback Recording from Timeline

**API Call:**
```bash
curl "http://localhost:8081/api/v1/recordings/cam-main-entrance/playback?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T11:00:00Z"
```

**Response:**
```json
{
  "playbackUrl": "rtsp://192.168.1.9:554/playback/cam-main-entrance?start=2025-10-24T10:00:00Z&end=2025-10-24T11:00:00Z",
  "durationSeconds": 3600
}
```

**Play Recording:**
```bash
ffplay "rtsp://192.168.1.9:554/playback/cam-main-entrance?start=2025-10-24T10:00:00Z&end=2025-10-24T11:00:00Z"
```

**Dashboard:**
```
1. Click "Playback" in navigation
2. Select camera: "Main Entrance Camera"
3. Select date/time range using timeline picker
4. Click "Play"
5. Use controls: Play/Pause, Scrub, Speed (1x, 2x, 4x, 8x)
```

---

### Scenario 3: Export Video Clip

**Short Clip (< 30 minutes) - Direct Download:**
```bash
curl "http://localhost:8081/api/v1/recordings/cam-main-entrance/download?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T10:30:00Z&format=mp4" \
  -o incident_video.mp4
```

**Long Clip (> 1 hour) - Async Export:**
```bash
# Step 1: Submit export request
curl -X POST http://localhost:8090/api/v1/export \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "cam-main-entrance",
    "start_time": "2025-10-24T08:00:00Z",
    "end_time": "2025-10-24T18:00:00Z",
    "format": "mp4",
    "quality": "high"
  }'

# Response:
# {"export_id": "exp-12345", "status": "processing"}

# Step 2: Check export status
curl http://localhost:8090/api/v1/export/exp-12345/status

# Response when ready:
# {"export_id": "exp-12345", "status": "completed", "download_url": "..."}

# Step 3: Download exported file
curl -O http://localhost:9000/cctv-exports/exp-12345.mp4
```

---

### Scenario 4: Search Recording Timeline

**Get Recording Availability:**
```bash
curl "http://localhost:8081/api/v1/recordings/cam-main-entrance/timeline?startTime=2025-10-24T00:00:00Z&endTime=2025-10-25T23:59:59Z"
```

**Response:**
```json
{
  "camera_id": "cam-main-entrance",
  "timeline": [
    {
      "sequence_id": "seq-001",
      "start_time": "2025-10-24T00:00:00Z",
      "end_time": "2025-10-24T06:00:00Z",
      "duration_seconds": 21600,
      "has_video": true,
      "gaps": []
    },
    {
      "sequence_id": "seq-002",
      "start_time": "2025-10-24T06:00:00Z",
      "end_time": "2025-10-24T12:00:00Z",
      "duration_seconds": 21600,
      "has_video": true,
      "gaps": []
    }
  ]
}
```

---

## 🔧 Troubleshooting

### Issue: Cannot connect to Milestone server

**Check:**
```bash
# 1. Network connectivity
ping 192.168.1.9

# 2. Port accessibility
telnet 192.168.1.9 8081
telnet 192.168.1.9 554

# 3. Firewall rules (on Milestone server)
# Ensure ports 8081, 554, 7563 are open
```

---

### Issue: Authentication failed

**Verify:**
```bash
# Test credentials with curl
curl -u rta-integration:your-password http://192.168.1.9:8081/cameras

# If fails with 401 Unauthorized:
# - Check username/password are correct
# - Verify user exists in Milestone
# - Confirm user has API access permission
```

---

### Issue: RTSP stream not accessible

**Debug:**
```bash
# 1. Check if RTSP is enabled on Milestone
# Navigate to Recording Server settings
# Verify "Enable RTSP Streaming" is checked

# 2. Test RTSP with VLC
vlc rtsp://192.168.1.9:554/camera/{camera-id}

# 3. Check camera ID is correct
curl -u rta-integration:password http://192.168.1.9:8081/cameras | grep camera-id
```

---

### Issue: No recordings found

**Check:**
```bash
# 1. Verify camera is recording
# In Milestone Smart Client, check if camera shows recording indicator

# 2. Check retention period
# Ensure requested time range is within 90-day retention

# 3. Verify time format
# Use ISO 8601: 2025-10-24T10:00:00Z
# NOT: 2025-10-24 10:00:00

# 4. Check storage
# Ensure recording server has adequate disk space
```

---

## 📊 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   RTA CCTV Dashboard                         │
│                 http://localhost:3000                        │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ User Actions
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    VMS Service (Port 8081)                   │
│  - Camera registration                                       │
│  - Recording queries                                         │
│  - Milestone API integration                                 │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ API Calls
                           ▼
┌─────────────────────────────────────────────────────────────┐
│            Milestone XProtect Expert VMS                     │
│                   192.168.1.9                                │
├─────────────────────────────────────────────────────────────┤
│  - Mobile Server API (8081)                                  │
│  - RTSP Streaming (554)                                      │
│  - Recording Server (7563)                                   │
│  - Camera Management                                         │
│  - 90-day Recording Storage                                  │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ RTSP Streams
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              MediaMTX (RTSP Re-streamer)                     │
│                 rtsp://localhost:8554                        │
│  - HLS output: http://localhost:8888                         │
│  - WebRTC output: ws://localhost:8889                        │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ Low-latency Streams
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   LiveKit (WebRTC SFU)                       │
│                 ws://localhost:7880                          │
│  - Sub-500ms latency                                         │
│  - Dashboard streaming                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 📝 Configuration Files Reference

### 1. Environment Variables (.env)
```bash
# Milestone VMS
MILESTONE_SERVER_URL=http://192.168.1.9
MILESTONE_RTSP_URL=rtsp://192.168.1.9:554
MILESTONE_USERNAME=rta-integration
MILESTONE_PASSWORD=your-secure-password
MILESTONE_API_PORT=8081
MILESTONE_RECORDING_PORT=7563
```

### 2. VMS Service Config (services/vms-service/config/config.yaml)
```yaml
milestone:
  server_url: "http://192.168.1.9"
  rtsp_base_url: "rtsp://192.168.1.9:554"
  username: "rta-integration"
  password: "your-secure-password"
  api:
    cameras: "/api/rest/v1/cameras"
    recordings: "/api/rest/v1/recordings"
    playback: "/api/rest/v1/recordings/{cameraId}/playback"
```

---

## ✅ Pre-Integration Checklist

Before starting integration:

- [ ] Milestone XProtect Expert is running on 192.168.1.9
- [ ] All cameras are configured and recording
- [ ] Retention set to 90 days minimum
- [ ] Integration user `rta-integration` created
- [ ] User has permissions: View cameras, Playback, Export
- [ ] Mobile Server API enabled (port 8081)
- [ ] RTSP streaming enabled (port 554)
- [ ] Network connectivity verified (ping test)
- [ ] API access tested (curl test)
- [ ] Sample RTSP stream tested (ffplay/VLC test)

---

## 🎯 API Quick Reference

### Camera Management
```bash
# List all cameras
GET http://192.168.1.9/api/rest/v1/cameras

# Get camera details
GET http://192.168.1.9/api/rest/v1/cameras/{cameraId}
```

### Live Streaming
```bash
# Get live stream URL
GET http://192.168.1.9/api/rest/v1/cameras/{cameraId}/live

# Direct RTSP
rtsp://192.168.1.9:554/camera/{cameraId}
```

### Recording Playback
```bash
# Get recording timeline
GET http://192.168.1.9/api/rest/v1/recordings/{cameraId}/timeline?startTime={iso8601}&endTime={iso8601}

# Get playback stream
GET http://192.168.1.9/api/rest/v1/recordings/{cameraId}/playback?startTime={iso8601}&endTime={iso8601}&speed=1
```

### Download/Export
```bash
# Download (short clips)
GET http://192.168.1.9/api/rest/v1/recordings/{cameraId}/download?startTime={iso8601}&endTime={iso8601}&format=mp4

# Export (long clips)
POST http://192.168.1.9/api/rest/v1/recordings/export
Body: {"cameraId": "...", "startTime": "...", "endTime": "...", "format": "mp4"}
```

---

## 🔐 Security Considerations

**Production Deployment:**

1. **Change Default Passwords**
   - Milestone integration user
   - All RTA system default passwords (Grafana, MinIO, PostgreSQL)

2. **Enable HTTPS**
   - Configure SSL/TLS on Milestone server
   - Update RTA configuration: `https://192.168.1.9`

3. **Restrict Network Access**
   - Use firewall rules to limit access to required ports only
   - Implement IP whitelisting for API access

4. **Authentication**
   - Use token-based authentication instead of Basic Auth
   - Rotate credentials regularly

5. **Audit Logging**
   - Enable audit logs on Milestone
   - Monitor API access patterns

---

## 📞 Support Resources

**Milestone Documentation:**
- MIP SDK: https://doc.developer.milestonesys.com/
- API Reference: Check Milestone documentation portal
- Community Forum: https://forum.milestonesys.com/

**RTA CCTV System:**
- Dashboard: http://localhost:3000
- Grafana Monitoring: http://localhost:3001
- MinIO Storage: http://localhost:9001
- API Gateway: http://localhost:8000

**Testing Tools:**
- VLC Media Player (RTSP testing)
- ffmpeg/ffplay (Stream analysis)
- Postman (API testing)
- Milestone Smart Client (Verification)

---

## 🚀 Next Steps

After completing this quick start:

1. **Register All Cameras**
   - Use the camera registration API for each camera
   - Organize by agency (dubai_police, metro, bus, other)

2. **Configure Stream Quotas**
   - Set limits per agency in stream-counter service
   - Monitor usage via Grafana dashboard

3. **Set Up Monitoring**
   - Configure Grafana alerts for stream failures
   - Monitor storage usage in MinIO
   - Track API response times

4. **Test Playback Features**
   - Verify timeline navigation
   - Test export functionality
   - Validate speed controls (1x, 2x, 4x, 8x)

5. **User Training**
   - Dashboard navigation
   - Camera selection and viewing
   - Playback and export procedures

---

## 📖 Documentation Reading Order

**For Administrators:**
1. MILESTONE_INTEGRATION_REQUIREMENTS.md (15 min)
2. MILESTONE_QUICK_START.md (this file) (10 min)
3. VERIFICATION_GUIDE.md (20 min)

**For Developers:**
1. MILESTONE_QUICK_START.md (this file) (10 min)
2. MILESTONE_API_REFERENCE.md (30 min)
3. MILESTONE_PLAYBACK_API_EXPLAINED.md (20 min)
4. PLAYBACK_VS_DOWNLOAD_VS_EXPORT.md (15 min)

**For Operators:**
1. VERIFICATION_GUIDE.md (20 min)
2. MILESTONE_QUICK_START.md - "Common Usage Scenarios" section (5 min)

---

**Last Updated:** 2025-10-25
**Milestone Server:** 192.168.1.9
**RTA CCTV System Version:** 1.0.0
**Integration Status:** Ready for Configuration
