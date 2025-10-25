# MediaMTX RTSP Server Configuration

## Overview

MediaMTX is the RTSP server that handles video stream ingestion from Milestone VMS and distributes them to viewers with minimal latency.

## Architecture Integration

```
┌─────────────────┐
│  Milestone VMS  │ (External)
│  RTSP Sources   │
└────────┬────────┘
         │ RTSP Pull
         ↓
┌─────────────────────────────────────┐
│         MediaMTX RTSP Server        │
│  ┌──────────────────────────────┐  │
│  │  RTSP Server :8554           │  │ ← Ingest from Milestone
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  HLS Server :8888            │  │ ← Web browser playback
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  WebRTC Server :8889         │  │ ← Ultra-low latency (<500ms)
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  API Server :9997            │  │ ← Control API
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  Metrics :9998               │  │ ← Prometheus metrics
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
         │
         ├─────► React Dashboard (HLS/WebRTC)
         ├─────► Mobile Apps (RTSP/HLS)
         └─────► Recording Service (RTSP)
```

## Protocols Supported

### 1. RTSP (Input/Output)
- **Port**: 8554
- **Use**: Primary protocol for camera ingestion and backend services
- **Latency**: ~1-2 seconds
- **URL Format**: `rtsp://mediamtx:8554/{camera_id}`

### 2. HLS (Output Only)
- **Port**: 8888
- **Use**: Web browser playback (Safari, Chrome, Firefox)
- **Latency**: ~2-4 seconds (lowLatency mode: <2s)
- **URL Format**: `http://mediamtx:8888/{camera_id}/index.m3u8`
- **Variant**: lowLatency (fMP4) for sub-2s latency

### 3. WebRTC (Output Only)
- **Port**: 8889
- **Use**: Ultra-low latency streaming (<500ms)
- **URL Format**: `http://mediamtx:8889/{camera_id}/whep`
- **Protocol**: WHEP (WebRTC HTTP Egress Protocol)

## Stream Path Patterns

MediaMTX supports regex-based path matching:

### Dubai Police Cameras
- **Pattern**: `dubai_police_~id`
- **Example**: `dubai_police_123e4567-e89b-12d3-a456-426614174000`
- **Max Concurrent**: 50 (enforced by Stream Counter)

### Metro Cameras
- **Pattern**: `metro_~id`
- **Example**: `metro_987f6543-e21b-12d3-a456-426614174000`
- **Max Concurrent**: 30

### Bus Cameras
- **Pattern**: `bus_~id`
- **Example**: `bus_456a7890-e89b-12d3-a456-426614174000`
- **Max Concurrent**: 20

### Other Cameras
- **Pattern**: `other_~id`
- **Example**: `other_789b1234-e89b-12d3-a456-426614174000`
- **Max Concurrent**: 400

### Milestone Proxy Streams
- **Pattern**: `milestone_~cameraId`
- **Example**: `milestone_123e4567-e89b-12d3-a456-426614174000`
- **Purpose**: On-demand RTSP pull from Milestone VMS
- **Behavior**: Starts stream when first viewer connects, stops after 10s idle

## On-Demand Stream Pull

MediaMTX can pull streams from Milestone VMS on-demand when the first viewer connects:

```yaml
paths:
  milestone_~cameraId:
    runOnDemand: ffmpeg -i rtsp://milestone:554/${cameraId} -c copy -f rtsp rtsp://localhost:8554/milestone_${cameraId}
    runOnDemandRestart: yes
    runOnDemandCloseAfter: 10s
```

**How it works:**
1. Viewer requests `rtsp://mediamtx:8554/milestone_{camera_id}`
2. MediaMTX detects no active publisher
3. Runs `runOnDemand` command to start FFmpeg pull from Milestone
4. FFmpeg copies H.264 stream (no transcoding) to MediaMTX
5. MediaMTX distributes to all viewers
6. When last viewer disconnects, waits 10s then stops FFmpeg

**Benefits:**
- Saves bandwidth: Only pulls streams that are being viewed
- Saves CPU: No unnecessary transcoding
- Automatic cleanup: Stops streams when not in use

## API Usage

MediaMTX provides an HTTP API for runtime control:

### Get All Paths
```bash
curl http://mediamtx:9997/v3/paths/list
```

### Get Path Details
```bash
curl http://mediamtx:9997/v3/paths/get/milestone_{camera_id}
```

### Add Dynamic Path
```bash
curl -X POST http://mediamtx:9997/v3/config/paths/add/{name} \
  -H "Content-Type: application/json" \
  -d '{
    "source": "publisher",
    "runOnDemand": "ffmpeg -i rtsp://milestone:554/{camera_id} -c copy -f rtsp rtsp://localhost:8554/{name}",
    "runOnDemandRestart": true,
    "runOnDemandCloseAfter": "10s"
  }'
```

### Remove Path
```bash
curl -X POST http://mediamtx:9997/v3/config/paths/remove/{name}
```

### Get Active Connections
```bash
curl http://mediamtx:9997/v3/rtspconns/list
```

## Integration with Go API

The Go API will use MediaMTX API to:

1. **Create stream paths dynamically**:
   ```go
   // When user requests camera stream
   // 1. Check quota with Stream Counter
   // 2. Get RTSP URL from VMS Service
   // 3. Create MediaMTX path with runOnDemand
   // 4. Return MediaMTX URL to client
   ```

2. **Monitor active streams**:
   ```go
   // Poll MediaMTX API to get active connections
   // Update Stream Counter heartbeats
   // Clean up stale reservations
   ```

3. **Control stream lifecycle**:
   ```go
   // When user stops viewing:
   // 1. Release quota in Stream Counter
   // 2. Let MediaMTX auto-close after timeout
   ```

## Prometheus Metrics

MediaMTX exports metrics at `:9998/metrics`:

```
# Active paths (streams)
mediamtx_paths_count

# Active RTSP connections
mediamtx_rtsp_conns_count

# Active HLS sessions
mediamtx_hls_sessions_count

# Active WebRTC sessions
mediamtx_webrtc_sessions_count

# Bytes received/sent per path
mediamtx_path_bytes_received{path="milestone_xxx"}
mediamtx_path_bytes_sent{path="milestone_xxx"}
```

## Performance Tuning

### Low Latency Mode (HLS)
- **Setting**: `hlsVariant: lowLatency`
- **Segment Duration**: 1 second
- **Part Duration**: 200ms
- **Result**: ~1-2 second latency (vs 6-10s standard HLS)

### WebRTC for Ultra-Low Latency
- **Latency**: <500ms
- **Use Case**: Live monitoring dashboards, critical operations
- **Limitation**: Higher CPU usage per viewer

### RTSP TCP Mode
- **Setting**: `protocols: [tcp]`
- **Reason**: Better firewall compatibility, no UDP packet loss
- **Tradeoff**: Slightly higher latency vs UDP

### Connection Limits
- **Max Streams**: 500 (matches quota system)
- **Max Viewers per Stream**: 10-20 (typical)
- **Total Connections**: ~5000-10000

## Resource Usage

### Per Stream (RTSP Pass-Through)
- **CPU**: 0.01 core (no transcoding)
- **RAM**: ~5MB
- **Network**: 2-4 Mbps (H.264 720p/1080p)

### Total Footprint (500 streams)
- **CPU**: ~5 cores
- **RAM**: ~2.5 GB
- **Network**: 1-2 Gbps

### With HLS Muxing (500 streams)
- **CPU**: ~10 cores (HLS segmentation overhead)
- **RAM**: ~4 GB
- **Network**: Same as RTSP

## Testing

### Test Stream Publishing
```bash
# Publish test stream with FFmpeg
ffmpeg -re -i test.mp4 -c copy -f rtsp rtsp://localhost:8554/test

# View with VLC
vlc rtsp://localhost:8554/test

# View HLS in browser
open http://localhost:8888/test/index.m3u8
```

### Test Milestone Proxy
```bash
# Create test path
curl -X POST http://localhost:9997/v3/config/paths/add/test_milestone \
  -H "Content-Type: application/json" \
  -d '{
    "source": "publisher",
    "runOnDemand": "ffmpeg -re -i test.mp4 -c copy -f rtsp rtsp://localhost:8554/test_milestone",
    "runOnDemandRestart": true,
    "runOnDemandCloseAfter": "10s"
  }'

# Connect viewer (will trigger runOnDemand)
vlc rtsp://localhost:8554/test_milestone
```

### Test WebRTC
```bash
# View with browser (requires WebRTC client)
# See: https://github.com/bluenviron/mediamtx#webrtc
```

## Docker Integration

```yaml
services:
  mediamtx:
    image: bluenviron/mediamtx:latest
    container_name: cctv-mediamtx
    restart: unless-stopped
    networks:
      - cctv-network
    ports:
      - "8554:8554"   # RTSP
      - "8888:8888"   # HLS
      - "8889:8889"   # WebRTC
      - "9997:9997"   # API
      - "9998:9998"   # Metrics
    volumes:
      - ./config/mediamtx.yml:/mediamtx.yml:ro
    environment:
      MTX_LOGDESTINATIONS: "stdout"
    deploy:
      resources:
        limits:
          cpus: '10'
          memory: 4G
        reservations:
          cpus: '2'
          memory: 1G
```

## Security Considerations

### Production Deployment
1. **Enable authentication**:
   ```yaml
   rtspAuthMethods: [basic]
   paths:
     milestone_~id:
       publishUser: milestone
       publishPass: ${MILESTONE_RTSP_PASSWORD}
       readUser: viewer
       readPass: ${VIEWER_RTSP_PASSWORD}
   ```

2. **Enable TLS/SSL**:
   ```yaml
   encryption: "yes"
   serverKey: /certs/server.key
   serverCert: /certs/server.crt
   ```

3. **Restrict origins**:
   ```yaml
   hlsAllowOrigin: "https://rta.ae"
   webrtcAllowOrigin: "https://rta.ae"
   ```

4. **Internal network only**:
   - Deploy MediaMTX on internal network
   - Only expose via Kong API Gateway with authentication

## Troubleshooting

### Stream Not Starting
```bash
# Check MediaMTX logs
docker logs cctv-mediamtx

# Check if path exists
curl http://localhost:9997/v3/paths/get/milestone_{camera_id}

# Test RTSP connection manually
ffplay rtsp://localhost:8554/milestone_{camera_id}
```

### High CPU Usage
```bash
# Check if transcoding is happening (should NOT be)
docker exec cctv-mediamtx ps aux | grep ffmpeg

# Verify -c copy in runOnDemand commands
cat config/mediamtx.yml | grep "runOnDemand"
```

### Latency Issues
```bash
# Check HLS variant (should be lowLatency)
curl http://localhost:9997/v3/config/get | jq .hlsVariant

# Test with WebRTC instead (lower latency)
# Use WebRTC URL: http://localhost:8889/{path}/whep
```

## References

- **MediaMTX Documentation**: https://github.com/bluenviron/mediamtx
- **RTSP Protocol**: RFC 2326
- **HLS Low Latency**: Apple HLS Authoring Specification
- **WebRTC WHEP**: IETF Draft
