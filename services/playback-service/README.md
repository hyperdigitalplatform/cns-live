# Playback Service

The Playback Service provides video playback capabilities for both recorded and live video streams from the RTA CCTV system.

## Features

- **HLS Playback**: Time-range playback with automatic segment stitching
- **Live Streaming**: HLS, RTSP, and WebRTC live streams via MediaMTX
- **On-Demand Transmuxing**: FFmpeg-based transmuxing for HLS compatibility
- **Session Management**: Cached playback sessions with TTL
- **Low CPU Usage**: H.264 copy (no transcoding)
- **Valkey Caching**: Redis-compatible caching for manifests and sessions

## Architecture

- **Clean Architecture**: Domain → Manager → Delivery layers
- **HLS Service**: Manifest generation and segment transmuxing
- **Valkey Cache**: Session and manifest caching
- **FFmpeg**: Video transmuxing (copy mode only)
- **MediaMTX Integration**: Live streaming proxy

## API Endpoints

### Start Playback (Recorded Video)

```bash
POST /api/v1/playback/start
{
  "camera_id": "uuid",
  "start_time": "2024-01-20T10:00:00Z",
  "end_time": "2024-01-20T11:00:00Z",
  "format": "hls"
}

# Response
{
  "session_id": "uuid",
  "camera_id": "uuid",
  "start_time": "2024-01-20T10:00:00Z",
  "end_time": "2024-01-20T11:00:00Z",
  "format": "hls",
  "url": "/api/v1/playback/sessions/{session_id}/playlist.m3u8",
  "expires_at": "2024-01-20T12:00:00Z",
  "segment_ids": ["seg1", "seg2", ...]
}
```

### Start Live Stream

```bash
POST /api/v1/playback/live
{
  "camera_id": "uuid",
  "format": "hls"  // hls, rtsp, webrtc
}

# Response
{
  "session_id": "uuid",
  "camera_id": "uuid",
  "format": "hls",
  "url": "http://mediamtx:8888/camera-id/index.m3u8",
  "expires_at": "2024-01-21T10:00:00Z"
}
```

### Get HLS Manifest

```bash
GET /api/v1/playback/sessions/{session_id}/playlist.m3u8

# Response (m3u8 playlist)
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-TARGETDURATION:3600
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:3600.000,
/api/v1/playback/sessions/{session_id}/segment/0.ts
#EXTINF:3600.000,
/api/v1/playback/sessions/{session_id}/segment/1.ts
#EXT-X-ENDLIST
```

### Get HLS Segment

```bash
GET /api/v1/playback/sessions/{session_id}/segment/{index}.ts

# Response: video/mp2t stream
```

### Session Management

```bash
# Get session details
GET /api/v1/playback/sessions/{session_id}

# Stop playback
DELETE /api/v1/playback/sessions/{session_id}

# Extend session TTL
POST /api/v1/playback/sessions/{session_id}/extend
{
  "duration_seconds": 3600
}
```

### Health & Metrics

```bash
GET /health
GET /metrics  # Prometheus metrics
```

## HLS Workflow

1. **Client requests playback** with time range
2. **Service fetches segments** from Storage Service
3. **Generates m3u8 manifest** with segment references
4. **Caches manifest** in Valkey (1 hour TTL)
5. **Client requests manifest** and starts playback
6. **Client requests segments** one by one
7. **Service downloads segment** from MinIO via presigned URL
8. **FFmpeg transmuxes** to HLS-compatible TS (H.264 copy)
9. **Streams segment** to client
10. **Cleanup temp files** after serving

## Live Streaming Workflow

1. **Client requests live stream**
2. **Service validates camera** via VMS Service
3. **Returns MediaMTX URL** for the camera
4. **Client connects directly** to MediaMTX
5. **MediaMTX proxies** from Milestone VMS

## FFmpeg Transmuxing

The service uses FFmpeg in copy mode for minimal CPU usage:

```bash
ffmpeg -i input.ts \
  -c copy \                       # Copy codecs (no transcoding)
  -bsf:v h264_mp4toannexb \      # Convert to Annex B for HLS
  -f mpegts \                     # Output format
  output-hls.ts
```

**Performance**: ~0.01 CPU per segment (transmuxing takes ~100ms for 1-hour segment)

## Caching Strategy

### Valkey Keys

- `playback:session:{session_id}` - Session metadata (1 hour TTL)
- `playback:hls:manifest:{session_id}` - HLS manifest (1 hour TTL)
- `playback:hls:segment:{session_id}:{index}` - Segment ID mapping (1 hour TTL)

### Cache Invalidation

- Sessions auto-expire after TTL
- Manual cleanup on DELETE /sessions/{id}
- Pattern-based deletion for all session-related keys

## Configuration

Environment variables:

```bash
# Service URLs
STORAGE_SERVICE_URL=http://storage-service:8082
VMS_SERVICE_URL=http://vms-service:8081
MEDIAMTX_URL=http://mediamtx:8888

# Valkey (Redis)
VALKEY_ADDR=valkey:6379
VALKEY_PASSWORD=
VALKEY_DB=1

# Service
PORT=8085
WORK_DIR=/tmp/playback
LOG_LEVEL=info
LOG_FORMAT=json
```

## Resource Usage

- **CPU**: 1-4 cores (depends on concurrent playback sessions)
- **Memory**: 512 MB - 2 GB
- **Disk**: Minimal temp files (auto-cleanup after serving)
- **Network**: High (streaming video to clients)

## Integration

### Storage Service
- Lists segments for time range
- Gets presigned download URLs

### VMS Service
- Validates camera existence
- Gets RTSP URLs for live streaming

### MediaMTX
- Proxies live RTSP streams
- Provides HLS/WebRTC endpoints

### Valkey
- Caches sessions and manifests
- Stores segment index mappings

## Video Player Integration

### Web (HLS.js)

```html
<video id="video" controls></video>
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<script>
  const video = document.getElementById('video');
  const hls = new Hls();

  // Start playback session
  fetch('/api/v1/playback/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      camera_id: 'uuid',
      start_time: '2024-01-20T10:00:00Z',
      end_time: '2024-01-20T11:00:00Z',
      format: 'hls'
    })
  })
  .then(res => res.json())
  .then(data => {
    hls.loadSource(data.url);
    hls.attachMedia(video);
  });
</script>
```

### Mobile (Native)

iOS/Android native players support HLS natively:

```swift
// iOS (AVPlayer)
let url = URL(string: "http://server:8085/api/v1/playback/sessions/uuid/playlist.m3u8")!
let player = AVPlayer(url: url)
```

## Development

```bash
# Build
go build -o playback-service ./cmd/main.go

# Run
./playback-service

# Docker build
docker build -t cctv-playback-service .

# Docker run
docker run -p 8085:8085 \
  -e STORAGE_SERVICE_URL=http://... \
  -e VALKEY_ADDR=valkey:6379 \
  cctv-playback-service
```

## Troubleshooting

### Segment not found
- Check Storage Service for segment availability
- Verify time range overlaps with recorded segments
- Check MinIO bucket for segment files

### FFmpeg errors
- Ensure FFmpeg is installed in container
- Check segment format compatibility (should be MPEG-TS)
- Review FFmpeg logs in service output

### Playback stuttering
- Increase client buffer size
- Check network bandwidth
- Verify segment availability in storage

### Cache misses
- Check Valkey connection
- Verify TTL settings
- Monitor Valkey memory usage
