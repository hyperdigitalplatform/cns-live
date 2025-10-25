# 🎥 RTA CCTV System - Verification Guide

## 📋 System Overview

**Key Services & Ports:**
- **Dashboard (Frontend)**: http://localhost:3000
- **API Gateway (Kong)**: http://localhost:8000 (Proxy), http://localhost:8001 (Admin)
- **VMS Service**: http://localhost:8081
- **Storage Service**: http://localhost:8082
- **Recording Service**: http://localhost:8083
- **Metadata Service**: http://localhost:8084
- **Stream Counter**: http://localhost:8087
- **Go API**: http://localhost:8088
- **Playback Service**: http://localhost:8092 ⚠️ (Changed from 8090)
- **MediaMTX (RTSP Server)**: rtsp://localhost:8554
- **LiveKit (WebRTC)**: ws://localhost:7880
- **Grafana (Monitoring)**: http://localhost:3001
- **MinIO (Storage)**: http://localhost:9001 (Console)
- **Prometheus**: http://localhost:9090

---

## 📌 Important API Endpoint Notes

### Kong Gateway vs Direct Access

**Via Kong Gateway (Port 8000) - Recommended for external clients:**
- VMS APIs: `http://localhost:8000/vms/*`
- Stream APIs: `http://localhost:8000/api/v1/stream/*`
- RTSP APIs: `http://localhost:8000/api/v1/rtsp/*`

**Direct Service Access - For testing/debugging:**
- VMS Service: `http://localhost:8081/vms/*`
- Stream Counter: `http://localhost:8087/api/v1/stream/*`
- Other services: Use their respective ports

### ⚠️ Common Mistakes to Avoid
1. **Wrong:** `http://localhost:8081/api/v1/cameras` → **Correct:** `http://localhost:8081/vms/cameras`
2. **Wrong:** `http://localhost:8087/api/v1/health` → **Correct:** `http://localhost:8087/health`
3. **Wrong:** `http://localhost:8000/health` → **Correct:** `http://localhost:8001/status` (Kong Admin API)
4. **Wrong:** Port 8090 for playback → **Correct:** Port 8092

---

## 🔧 Phase 1: System Health Check

### 1.1 Verify All Services are Running
```bash
docker ps --filter "name=cctv-" --format "table {{.Names}}\t{{.Status}}" | grep healthy
```

**Expected**: All core services should show "(healthy)" status

### 1.2 Check API Gateway Health
```bash
# Kong Admin API Status (detailed)
curl http://localhost:8001/status

# Kong Admin API (simple check)
curl http://localhost:8001
```

**Expected Response:**
```json
{
  "server": {
    "connections_active": 21,
    "connections_accepted": 49,
    "total_requests": 49
  },
  "configuration_hash": "..."
}
```

### 1.3 Check Stream Counter Service
```bash
curl http://localhost:8087/health
```

**Expected Response:**
```json
{
  "service": "stream-counter",
  "status": "healthy",
  "timestamp": "2025-10-25T..."
}
```

### 1.4 Check All Service Health Endpoints
```bash
# VMS Service
curl http://localhost:8081/health

# Storage Service
curl http://localhost:8082/health

# Recording Service
curl http://localhost:8083/health

# Metadata Service
curl http://localhost:8084/health

# Stream Counter
curl http://localhost:8087/health

# Go API
curl http://localhost:8088/health

# Playback Service (Note: Port changed from 8090 to 8092)
curl http://localhost:8092/health
```

---

## 🎬 Phase 2: Camera Configuration & Registration

### 2.1 List Cameras from VMS Service

The VMS Service connects to Milestone VMS and retrieves camera information.

**API Endpoints:**

**Direct Access:**
```bash
# Get all cameras
curl http://localhost:8081/vms/cameras

# Get all cameras (via Kong Gateway)
curl http://localhost:8000/vms/cameras

# Filter by source
curl http://localhost:8081/vms/cameras?source=DUBAI_POLICE
```

**Expected Response:**
```json
{
  "cameras": [
    {
      "id": "a14c5b2b-c315-4f68-a87b-dffbfb60917b",
      "name": "Camera 001 - Sheikh Zayed Road",
      "name_ar": "كاميرا 001 - شارع الشيخ زايد",
      "source": "DUBAI_POLICE",
      "rtsp_url": "rtsp://milestone.rta.ae:554/camera_001",
      "ptz_enabled": true,
      "status": "ONLINE",
      "recording_server": "milestone.rta.ae:554",
      "metadata": {
        "fps": 25,
        "location": {
          "lat": 25.2048,
          "lon": 55.2708
        },
        "resolution": "1920x1080"
      }
    }
  ],
  "total": 2,
  "last_updated": "2025-10-25T..."
}
```

### 2.2 Get Specific Camera Details

✅ **FIXED:** The VMS service now uses PostgreSQL database for persistence, so camera IDs are stable.

**Test with sample cameras:**
```bash
# Get a camera ID from the list
curl http://localhost:8081/vms/cameras

# Use one of the sample IDs
curl http://localhost:8081/vms/cameras/cam-001-sheikh-zayed

# Or the metro camera
curl http://localhost:8081/vms/cameras/cam-002-metro-station
```

**Expected Response:**
```json
{
  "id": "cam-001-sheikh-zayed",
  "name": "Camera 001 - Sheikh Zayed Road",
  "name_ar": "كاميرا 001 - شارع الشيخ زايد",
  "source": "DUBAI_POLICE",
  "rtsp_url": "rtsp://milestone.rta.ae:554/camera_001",
  "ptz_enabled": true,
  "status": "ONLINE",
  "recording_server": "milestone.rta.ae:554",
  "milestone_device_id": "milestone_device_001",
  "metadata": {
    "fps": 25,
    "location": {
      "lat": 25.2048,
      "lon": 55.2708
    },
    "resolution": "1920x1080"
  },
  "last_update": "2025-10-25T...",
  "created_at": "2025-09-25T..."
}
```

### 2.3 Get Camera Stream URL
```bash
# Direct access
curl http://localhost:8081/vms/cameras/{camera_id}/stream

# Via Kong Gateway
curl http://localhost:8000/vms/cameras/{camera_id}/stream
```

---

## 📺 Phase 3: Live Stream Viewing

### 3.1 Access the Dashboard
Open your browser and navigate to:
```
http://localhost:3000
```

### 3.2 View Live Streams in Grid

**Dashboard Features:**
- **Grid View**: Display multiple cameras in 2x2, 3x3, or 4x4 grid
- **Single View**: Full-screen view of a single camera
- **PTZ Controls**: Pan/Tilt/Zoom controls (if camera supports)

**Using the Dashboard:**

1. **Navigate to Live View**
   - Click on "Live View" in the navigation menu
   - You should see the camera grid layout

2. **Select Grid Size**
   - Use the grid selector (2x2, 3x3, 4x4) to choose layout
   - Available in the top toolbar

3. **Add Camera to Grid**
   - Click on an empty grid cell
   - Select camera from the dropdown list
   - Camera stream should start playing

### 3.3 Stream via RTSP (Direct Access)

**Test with VLC or ffplay:**
```bash
# Using VLC
vlc rtsp://localhost:8554/dubai_police_test-camera-001

# Using ffplay
ffplay rtsp://localhost:8554/dubai_police_test-camera-001
```

### 3.4 Stream via HLS (Web Browser)

**HLS Stream URL:**
```
http://localhost:8888/dubai_police_test-camera-001/index.m3u8
```

**Test in browser or with curl:**
```bash
curl -I http://localhost:8888/dubai_police_test-camera-001/index.m3u8
```

### 3.5 Stream via WebRTC/WHIP (Ultra Low Latency)

**WHIP (WebRTC HTTP Ingestion Protocol) Architecture:**

The system uses WHIP for camera ingestion, providing ~450ms latency vs 2-4 seconds with HLS.

**Architecture Flow:**
```
Camera (RTSP) → MediaMTX → GStreamer WHIP Pusher → LiveKit WHIP Ingress → LiveKit SFU → Viewers
```

**Key Components:**
- **WHIP Pusher Containers**: Separate Docker containers per camera
- **GStreamer Pipeline**: H.264 passthrough (no transcoding) or H.265→H.264 transcoding
- **LiveKit Ingress**: WHIP endpoint accepting WebRTC push
- **LiveKit SFU**: Selective Forwarding Unit for distribution to viewers

**Testing WHIP Stream:**

1. **Reserve a stream via Go API:**
```bash
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "cam-001-sheikh-zayed",
    "user_id": "test-user",
    "quality": "medium"
  }'
```

2. **Expected Response:**
```json
{
  "reservation_id": "uuid-here",
  "camera_id": "cam-001-sheikh-zayed",
  "camera_name": "Camera 1 - 192.168.1.8",
  "room_name": "camera_cam-001-sheikh-zayed",
  "token": "jwt-token-here",
  "livekit_url": "ws://localhost:7880",
  "expires_at": "2025-10-25T22:00:00Z",
  "quality": "medium"
}
```

3. **Verify WHIP Pusher Container:**
```bash
# Check container is running
docker ps | grep whip-pusher-cam-001-sheikh-zayed

# Check GStreamer pipeline logs
docker logs whip-pusher-cam-001-sheikh-zayed --tail 50
```

4. **Expected Log Output:**
```
Starting WHIP Pusher...
RTSP Source: rtsp://raammohan:Ilove123@192.168.1.8:554/stream1
WHIP Endpoint: http://livekit-ingress:8080/w/<stream-key>
[GStreamer pipeline negotiation logs...]
packets-sent=(guint64)1000+
bitrate=(guint64)1700000+  // ~1.7 Mbps for H.264
```

5. **Check LiveKit Room:**
```bash
# View LiveKit logs for participant connection
docker logs cctv-livekit --tail 30 | grep "camera_cam-001"
```

**WHIP Pusher Technical Details:**
- **Base Image**: Ubuntu 22.04 with GStreamer 1.0 + gst-plugins-rs
- **Pipeline**: `rtspsrc → caps(video) → rtpjitterbuffer → decodebin → x264enc → rtph264pay → whipsink`
- **Codec Support**:
  - H.264: Passthrough (no transcoding)
  - H.265: Transcoded to H.264 for standardization
- **Audio**: Filtered out (video-only streams)
- **Latency**: ~450ms end-to-end

**Multiple Camera Testing:**
```bash
# Reserve Camera 1 (H.264)
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{"camera_id":"cam-001-sheikh-zayed","user_id":"user1","quality":"medium"}'

# Reserve Camera 2 (H.265 → H.264 transcoded)
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{"camera_id":"cam-002-metro-station","user_id":"user2","quality":"medium"}'

# Verify both containers running
docker ps | grep whip-pusher

# Should show:
# whip-pusher-cam-001-sheikh-zayed
# whip-pusher-cam-002-metro-station
```

**Release Stream:**
```bash
curl -X DELETE "http://localhost:8088/api/v1/stream/release/{reservation_id}"
```

---

## 🎞️ Phase 4: Recording Playback from Milestone

### 4.1 Request Recording Segments from Milestone VMS

**API Endpoint:** `GET http://localhost:8081/vms/recordings/{camera_id}/segments`

**Get available recording segments:**
```bash
# Direct access
curl "http://localhost:8081/vms/recordings/{camera_id}/segments?start=2025-10-24T00:00:00Z&end=2025-10-25T23:59:59Z"

# Via Kong Gateway
curl "http://localhost:8000/vms/recordings/{camera_id}/segments?start=2025-10-24T00:00:00Z&end=2025-10-25T23:59:59Z"
```

**Expected Response:**
```json
{
  "camera_id": "a14c5b2b-c315-4f68-a87b-dffbfb60917b",
  "segments": [
    {
      "start_time": "2025-10-24T10:00:00Z",
      "end_time": "2025-10-24T11:00:00Z",
      "duration_seconds": 3600,
      "recording_server": "milestone.rta.ae:554"
    }
  ],
  "total_segments": 1
}
```

### 4.2 Stream Recorded Video via Playback Service

**API Endpoint:** `GET http://localhost:8092/playback` (Note: Port changed from 8090 to 8092)

**Stream a recording:**
```bash
curl "http://localhost:8092/playback?camera_id={camera_id}&start=2025-10-24T10:00:00Z&duration=3600" \
  -H "Accept: video/mp4"
```

### 4.3 Use Dashboard Playback View

1. **Navigate to Playback View**
   - Click "Playback" in the navigation menu

2. **Select Camera & Time Range**
   - Choose camera from dropdown
   - Select date and time range using the timeline picker

3. **Control Playback**
   - Play/Pause button
   - Timeline scrubbing
   - Speed controls (1x, 2x, 4x, 8x)
   - Frame-by-frame stepping

### 4.4 Export Recordings

**API Endpoint:** `POST http://localhost:8081/vms/recordings/export` (Note: Export handled by VMS service)

```bash
# Direct access
curl -X POST http://localhost:8081/vms/recordings/export \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "{camera_id}",
    "start_time": "2025-10-24T10:00:00Z",
    "end_time": "2025-10-24T10:30:00Z",
    "format": "mp4"
  }'

# Via Kong Gateway
curl -X POST http://localhost:8000/vms/recordings/export \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "{camera_id}",
    "start_time": "2025-10-24T10:00:00Z",
    "end_time": "2025-10-24T10:30:00Z",
    "format": "mp4"
  }'
```

**Expected Response:**
```json
{
  "export_id": "exp-12345",
  "status": "processing",
  "camera_id": "{camera_id}",
  "estimated_completion": "2025-10-24T10:35:00Z"
}
```

**Check export status:**
```bash
# Direct access
curl http://localhost:8081/vms/recordings/export/{export_id}

# Via Kong Gateway
curl http://localhost:8000/vms/recordings/export/{export_id}
```

**Download exported video (from MinIO):**
```bash
# Access via MinIO S3 API
curl -O http://localhost:9000/cctv-recordings/exports/{export_id}.mp4
```

---

## 🔍 Phase 5: Advanced Testing

### 5.1 Stream Counter & Quota Management

**Check current stream statistics:**
```bash
# Direct access
curl http://localhost:8087/api/v1/stream/stats

# Via Kong Gateway
curl http://localhost:8000/api/v1/stream/stats
```

**Expected Response:**
```json
{
  "stats": [
    {
      "source": "DUBAI_POLICE",
      "current": 0,
      "limit": 50,
      "percentage": 0,
      "available": 50
    },
    {
      "source": "METRO",
      "current": 0,
      "limit": 30,
      "percentage": 0,
      "available": 30
    },
    {
      "source": "BUS",
      "current": 0,
      "limit": 20,
      "percentage": 0,
      "available": 20
    },
    {
      "source": "OTHER",
      "current": 0,
      "limit": 400,
      "percentage": 0,
      "available": 400
    }
  ],
  "total": {
    "current": 0,
    "limit": 500,
    "percentage": 0,
    "available": 500
  },
  "timestamp": "2025-10-25T..."
}
```

**Reserve a new stream:**
```bash
# Direct access
curl -X POST http://localhost:8087/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "cameraID": "{camera_id}",
    "userID": "user-123",
    "source": "DUBAI_POLICE",
    "duration": 3600
  }'

# Via Kong Gateway
curl -X POST http://localhost:8000/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "cameraID": "{camera_id}",
    "userID": "user-123",
    "source": "DUBAI_POLICE",
    "duration": 3600
  }'
```

**Expected Response:**
```json
{
  "reservationID": "res-abc123",
  "cameraID": "{camera_id}",
  "expiresAt": "2025-10-25T11:00:00Z",
  "currentUsage": {
    "source": "DUBAI_POLICE",
    "current": 1,
    "limit": 50
  }
}
```

**Release a stream:**
```bash
# Direct access
curl -X DELETE http://localhost:8087/api/v1/stream/release/{reservation_id}

# Via Kong Gateway
curl -X DELETE http://localhost:8000/api/v1/stream/release/{reservation_id}
```

**Send heartbeat to keep reservation alive:**
```bash
# Direct access
curl -X POST http://localhost:8087/api/v1/stream/heartbeat/{reservation_id}

# Via Kong Gateway
curl -X POST http://localhost:8000/api/v1/stream/heartbeat/{reservation_id}
```

### 5.2 Monitoring & Metrics

**Access Grafana Dashboard:**
```
http://localhost:3001
```

**Default Credentials:**
- Username: `admin`
- Password: `admin_changeme`

**Key Dashboards:**
- RTA CCTV Overview
- Stream Performance Metrics
- Storage Utilization
- Network Bandwidth

**Access Prometheus:**
```
http://localhost:9090
```

**Example Queries:**
- Total active streams: `sum(mediamtx_paths_active)`
- CPU usage: `rate(container_cpu_usage_seconds_total[5m])`
- Memory usage: `container_memory_usage_bytes`

### 5.3 Storage Management (MinIO)

**Access MinIO Console:**
```
http://localhost:9001
```

**Default Credentials:**
- Username: `admin`
- Password: `changeme_minio`

**Verify Buckets:**
- `cctv-recordings` - Long-term recordings (90-day retention)
- `cctv-exports` - Exported clips (7-day retention)
- `cctv-thumbnails` - Preview thumbnails (30-day retention)
- `cctv-clips` - Saved clips (manual deletion)

---

## 🧪 Phase 6: End-to-End Testing Scenario

### Complete Workflow Test:

```bash
# 1. Get list of available cameras from Milestone
curl http://localhost:8081/vms/cameras

# 2. Get details of a specific camera (use a camera ID from step 1)
CAMERA_ID="a14c5b2b-c315-4f68-a87b-dffbfb60917b"
curl "http://localhost:8081/vms/cameras/${CAMERA_ID}"

# 3. Check stream quota/statistics
curl http://localhost:8087/api/v1/stream/stats

# 4. Reserve stream access
curl -X POST http://localhost:8087/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d "{
    \"cameraID\": \"${CAMERA_ID}\",
    \"userID\": \"test-user\",
    \"source\": \"DUBAI_POLICE\",
    \"duration\": 3600
  }"

# Save the reservationID from response
RESERVATION_ID="<reservation_id_from_response>"

# 5. View live stream in browser
# Open: http://localhost:3000

# 6. Query available recording segments
curl "http://localhost:8081/vms/recordings/${CAMERA_ID}/segments?start=2025-10-24T00:00:00Z&end=2025-10-25T23:59:59Z"

# 7. Export a recording clip
curl -X POST http://localhost:8081/vms/recordings/export \
  -H "Content-Type: application/json" \
  -d "{
    \"camera_id\": \"${CAMERA_ID}\",
    \"start_time\": \"2025-10-24T10:00:00Z\",
    \"end_time\": \"2025-10-24T10:05:00Z\",
    \"format\": \"mp4\"
  }"

# Save export_id from response
EXPORT_ID="<export_id_from_response>"

# 8. Check export status
curl "http://localhost:8081/vms/recordings/export/${EXPORT_ID}"

# 9. Send heartbeat to keep stream reservation alive
curl -X POST "http://localhost:8087/api/v1/stream/heartbeat/${RESERVATION_ID}"

# 10. Release stream when done
curl -X DELETE "http://localhost:8087/api/v1/stream/release/${RESERVATION_ID}"
```

---

## ✅ Success Criteria Checklist

- [ ] All 25+ services are running and healthy
- [ ] Can register cameras via VMS API
- [ ] Dashboard loads successfully at http://localhost:3000
- [ ] Live streams display in grid view (2x2, 3x3, 4x4)
- [ ] Can switch between cameras in grid
- [ ] RTSP streams accessible via MediaMTX
- [ ] HLS streams work in web browser
- [ ] WebRTC streams provide low-latency playback
- [ ] Can query recordings from Milestone
- [ ] Playback timeline shows available recordings
- [ ] Can export video clips
- [ ] Stream quota management works correctly
- [ ] Grafana dashboards show metrics
- [ ] MinIO storage buckets accessible
- [ ] No stream quota violations occur

---

## 🐛 Troubleshooting

### Issue: Dashboard not loading
```bash
docker logs cctv-dashboard --tail 50
docker logs cctv-go-api --tail 50
```

### Issue: Streams not playing
```bash
docker logs cctv-mediamtx --tail 50
docker logs cctv-livekit --tail 50
docker logs cctv-livekit-ingress --tail 50
```

### Issue: WHIP Pusher Container Failing

**Symptoms:**
- Container restarting constantly
- No video in LiveKit room
- "no element whipsink" error

**Diagnosis:**
```bash
# Check container status
docker ps -a | grep whip-pusher

# View logs
docker logs whip-pusher-cam-<camera-id> --tail 100

# Check if container is restarting
docker inspect whip-pusher-cam-<camera-id> | grep RestartCount
```

**Common Issues:**

1. **Missing gst-plugins-rs**
```bash
# Verify whipsink element exists in image
docker run --rm --entrypoint sh whip-pusher:latest -c "gst-inspect-1.0 | grep whipsink"
# Should output: webrtchttp:  whipsink: WHIP Sink Bin
```

2. **RTSP URL Not Reachable**
```bash
# Test RTSP connection from pusher container
docker run --rm --network cns_cctv-network --entrypoint sh whip-pusher:latest \
  -c "timeout 5 gst-launch-1.0 rtspsrc location='rtsp://192.168.1.8:554/stream1' ! fakesink"
```

3. **LiveKit Ingress Not Ready**
```bash
# Check ingress health
docker logs cctv-livekit-ingress --tail 30
curl http://localhost:8080/  # Should respond
```

4. **Codec Mismatch (H.265 Camera)**
- Camera 2 uses H.265 and requires transcoding
- Check pipeline includes: `decodebin ! x264enc`
- Verify logs show transcoding activity

5. **Audio Stream Interference**
```bash
# Check if pipeline has video-only caps filter
docker logs whip-pusher-cam-<camera-id> | grep "application/x-rtp,media=video"
```

**Solutions:**

1. **Rebuild WHIP Pusher Image:**
```bash
cd services/whip-pusher
docker build -t whip-pusher:latest .
```

2. **Manually Stop/Remove Failed Container:**
```bash
docker stop whip-pusher-cam-<camera-id>
docker rm whip-pusher-cam-<camera-id>
```

3. **Check Network Connectivity:**
```bash
# Verify Docker network exists
docker network ls | grep cns_cctv-network

# Check container can reach LiveKit ingress
docker exec whip-pusher-cam-<camera-id> ping -c 3 livekit-ingress
```

### Issue: Both Cameras Showing Same Feed

**Root Cause:** Each camera must have its own unique WHIP pusher container with different RTSP URLs.

**Verification:**
```bash
# Check each container has different RTSP URL
docker inspect whip-pusher-cam-001-sheikh-zayed | grep RTSP_URL
docker inspect whip-pusher-cam-002-metro-station | grep RTSP_URL

# Should show different IPs:
# Camera 1: rtsp://...@192.168.1.8:554/...
# Camera 2: rtsp://...@192.168.1.13:554/...
```

**Solution:**
- Ensure VMS service returns correct camera details
- Verify MediaMTX path configuration is per-camera
- Check go-api spawns separate containers per camera ID

### Issue: Recordings not accessible
```bash
docker logs cctv-vms-service --tail 50
docker logs cctv-playback-service --tail 50
```

### Check service health
```bash
docker ps --filter "name=cctv-" --format "table {{.Names}}\t{{.Status}}"
```

### View service logs in real-time
```bash
# Follow logs for a specific service
docker logs -f cctv-<service-name>

# View last 100 lines
docker logs cctv-<service-name> --tail 100
```

### Restart a specific service
```bash
docker-compose restart <service-name>
```

### Rebuild a service after code changes
```bash
docker-compose build <service-name>
docker-compose up -d <service-name>
```

---

## 📊 Service Architecture

### Core Services:
1. **VMS Service** - Integrates with Milestone VMS
2. **Stream Counter** - Manages quota for 500 concurrent streams
3. **MediaMTX** - RTSP server for stream distribution
4. **LiveKit** - WebRTC SFU for ultra-low latency
5. **Recording Service** - Handles recording to MinIO
6. **Playback Service** - Retrieves and streams recordings
7. **Storage Service** - Manages MinIO storage
8. **Metadata Service** - Camera metadata and indexing

### Supporting Services:
- **Kong** - API Gateway
- **PostgreSQL** - Metadata database
- **Valkey** - Redis-compatible cache
- **MinIO** - S3-compatible object storage
- **Prometheus/Grafana** - Monitoring and visualization
- **Loki** - Log aggregation

---

## 🔐 Default Credentials

**Grafana:**
- URL: http://localhost:3001
- Username: `admin`
- Password: `admin_changeme`

**MinIO Console:**
- URL: http://localhost:9001
- Username: `admin`
- Password: `changeme_minio`

**PostgreSQL:**
- Host: localhost:5432
- Database: `cctv`
- Username: `cctv`
- Password: `changeme_db`

**Valkey (Redis):**
- Host: localhost:6379
- No password (development mode)

---

## 📝 Notes

1. **Production Deployment**: Change all default passwords before deploying to production
2. **Milestone Integration**: Update `rtsp_url` in camera registration to point to your actual Milestone VMS server
3. **Network Configuration**: Ensure firewall rules allow access to required ports
4. **Storage Capacity**: Monitor MinIO storage usage as recordings accumulate
5. **Stream Limits**: The system is configured for 500 concurrent streams. Adjust quotas in stream-counter service if needed
6. **Security**: All services are currently configured for development. Enable authentication and SSL/TLS for production

---

## 🚀 Quick Start Commands

```bash
# Start all services
docker-compose up -d

# Check service status
docker ps --filter "name=cctv-"

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v

# Rebuild specific service
docker-compose build <service-name>
docker-compose up -d <service-name>
```

---

## 📞 Support

For issues or questions:
1. Check service logs using `docker logs cctv-<service-name>`
2. Review the troubleshooting section above
3. Verify all services are healthy using health check commands
4. Check network connectivity between services

---

**Last Updated:** 2025-10-25
**Version:** 1.0.0
