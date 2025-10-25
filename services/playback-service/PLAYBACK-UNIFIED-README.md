# Unified Playback Service (Phase 3 Week 6)

The unified playback service for the RTA CCTV system. This service handles video playback orchestration with intelligent source detection (local MinIO vs external Milestone), FFmpeg transmuxing to HLS, and segment caching with LRU eviction.

## Key Features

- ✅ **Intelligent Source Detection**: Automatically selects best source (local MinIO or Milestone VMS)
- ✅ **FFmpeg H.264 → HLS Transmuxing**: Stream copy mode (NO re-encoding, ~500x realtime)
- ✅ **LRU Segment Caching**: 10GB cache with automatic eviction
- ✅ **Nginx Static Delivery**: Efficient HLS/DASH serving with proper cache headers
- ✅ **MP4 Export**: Downloadable video clips
- ✅ **Multi-segment Concatenation**: Automatic stitching of multiple segments
- ✅ **Low Resource Footprint**: 512MB-2GB RAM, minimal CPU (stream copy)

## Architecture

```
Playback Request
       ↓
Source Detector ──→ MinIO (Primary) ──→ Found? → Use MinIO
       ↓                                    ↓
       └────────────→ Milestone (Fallback)  Not Found? → Use Milestone

Selected Source
       ↓
Download Segments (check cache first)
       ↓
FFmpeg Transmuxing (H.264 → HLS, stream copy)
       ↓
Generate HLS Manifest (playlist.m3u8)
       ↓
Nginx Serves Segments → Client HLS Player
```

## API Endpoints

### 1. Request Playback

Create an HLS playback session for recorded video.

**Request**:
```bash
POST /api/v1/playback/request
Content-Type: application/json

{
  "camera_id": "550e8400-e29b-41d4-a716-446655440000",
  "start_time": "2024-01-20T10:00:00Z",
  "end_time": "2024-01-20T11:00:00Z",
  "format": "hls",
  "user_id": "user123"
}
```

**Success Response** (200 OK):
```json
{
  "session_id": "abc-123-def",
  "camera_id": "550e8400-e29b-41d4-a716-446655440000",
  "start_time": "2024-01-20T10:00:00Z",
  "end_time": "2024-01-20T11:00:00Z",
  "format": "hls",
  "url": "http://localhost:8091/hls/abc-123-def/playlist.m3u8",
  "expires_at": "2024-01-20T12:00:00Z",
  "segment_ids": ["seg1", "seg2", "seg3"]
}
```

**Error Response** (404):
```json
{
  "error": {
    "message": "No recordings available",
    "detail": "No segments found in local storage",
    "timestamp": "2024-01-20T11:00:00Z"
  }
}
```

### 2. Create Video Export

Export recorded video to downloadable MP4.

**Request**:
```bash
POST /api/v1/playback/export
Content-Type: application/json

{
  "camera_id": "550e8400-e29b-41d4-a716-446655440000",
  "start_time": "2024-01-20T10:00:00Z",
  "end_time": "2024-01-20T10:30:00Z",
  "format": "mp4",
  "user_id": "user123",
  "title": "Incident Evidence"
}
```

**Response** (200 OK):
```json
{
  "export_id": "export-123",
  "camera_id": "550e8400-e29b-41d4-a716-446655440000",
  "format": "mp4",
  "status": "ready",
  "download_url": "http://localhost:8091/exports/export-123/export_export-123.mp4",
  "file_size": 524288000,
  "duration": 1800,
  "created_at": "2024-01-20T11:00:00Z"
}
```

### 3. Cache Statistics

Get segment cache performance metrics.

**Request**:
```bash
GET /api/v1/playback/cache/stats
```

**Response** (200 OK):
```json
{
  "total_entries": 45,
  "current_size": 5368709120,
  "max_size": 10737418240,
  "usage_percent": 50.0,
  "hit_count": 1234,
  "miss_count": 567,
  "hit_rate": 68.5
}
```

## Source Detection Algorithm

The service uses intelligent source detection with coverage analysis:

```go
func DetectSource(cameraID, startTime, endTime) {
    // 1. Check local MinIO storage
    segments := minioClient.ListSegments(cameraID, startTime, endTime)

    // 2. Calculate coverage
    requestedDuration := endTime - startTime
    coveredDuration := sum(segment.Duration for segment in segments)
    coveragePercent := (coveredDuration / requestedDuration) * 100

    // 3. Decision
    if coveragePercent >= 80% {
        return PlaybackSourceLocal  // Use MinIO
    }

    // 4. Fallback to Milestone
    available := milestoneClient.CheckAvailability(cameraID, startTime, endTime)
    if available {
        return PlaybackSourceMilestone
    }

    return Error("No recordings available")
}
```

**Coverage Threshold**: 80%
- Requested: 1 hour
- Found: 50 minutes = 83.3% → Use local MinIO ✅
- Found: 45 minutes = 75.0% → Use Milestone fallback

## FFmpeg Transmuxing

### HLS Transmuxing (H.264 → HLS)

**Command**:
```bash
ffmpeg -i input.mp4 \
  -c copy \                          # Stream copy (NO re-encoding!)
  -movflags +faststart \             # Web-optimized
  -f hls \                           # HLS output
  -hls_time 6 \                      # 6-second segments
  -hls_list_size 0 \                 # Include all segments
  -hls_flags independent_segments \  # Each segment playable independently
  -hls_segment_filename segment_%03d.ts \
  playlist.m3u8
```

**Performance**:
- **Speed**: ~500x realtime (1 hour video in 7 seconds)
- **CPU**: <5% (no transcoding)
- **Quality**: Lossless (stream copy)

### MP4 Export

**Command**:
```bash
ffmpeg -i input.mp4 \
  -c copy \                 # Stream copy
  -movflags +faststart \    # Move moov atom to beginning
  -f mp4 \
  output.mp4
```

### Multi-Segment Concatenation

**Command**:
```bash
# Create concat file
cat > concat.txt <<EOF
file 'segment_001.mp4'
file 'segment_002.mp4'
file 'segment_003.mp4'
EOF

# Concatenate
ffmpeg -f concat -safe 0 -i concat.txt -c copy output.mp4
```

## Segment Caching (LRU)

### Configuration

```go
cache := NewSegmentCache(
    cacheDir:     "/tmp/playback/cache",
    maxSizeBytes: 10 * 1024 * 1024 * 1024,  // 10 GB
    logger:       logger,
)

// Auto cleanup: every 1 hour, remove entries older than 2 hours
cache.StartCleanupWorker(1*time.Hour, 2*time.Hour)
```

### Cache Operations

**Get (with cache hit)**:
```go
segmentID := "cam1_1705761600"
if cachedPath, found := cache.Get(segmentID); found {
    // Cache HIT - use cached file (instant)
    return cachedPath
}

// Cache MISS - download from MinIO
downloadPath := minioClient.GetSegment(segment.StoragePath)
cache.Put(segmentID, downloadPath)  // Add to cache
return downloadPath
```

### LRU Eviction

**Scenario**: Cache is 9.8 GB, need 500 MB for new segment

1. Current: 9.8 GB / 10 GB (98% full)
2. Required: 500 MB
3. Available: 200 MB
4. **Evict LRU segments**: Remove 300 MB of least recently used
5. **Add new segment**: 500 MB cached

**LRU List**:
```
[Most Recent] → seg_20 → seg_19 → seg_18 → ... → seg_1 [Least Recent]
                                                   ↑
                                              Evict first
```

### Cache Statistics

```json
{
  "total_entries": 45,
  "current_size": 5368709120,      // 5 GB
  "max_size": 10737418240,         // 10 GB
  "usage_percent": 50.0,
  "hit_count": 1234,               // Cache hits
  "miss_count": 567,               // Cache misses
  "hit_rate": 68.5                 // 68.5% hit rate
}
```

## Nginx Configuration

### HLS Manifest (playlist.m3u8)

```nginx
location ~ \.m3u8$ {
    add_header Cache-Control 'no-cache, no-store, must-revalidate';
    add_header Pragma 'no-cache';
    add_header Expires '0';
    add_header 'Access-Control-Allow-Origin' '*';
}
```
- **No caching**: Playlist always fetched fresh
- **CORS enabled**: Browser playback allowed

### HLS Segments (segment_*.ts)

```nginx
location ~ \.ts$ {
    add_header Cache-Control 'public, max-age=3600, immutable';
    add_header 'Access-Control-Allow-Origin' '*';
}
```
- **1-hour cache**: Segments are immutable
- **CDN-friendly**: Can be cached by proxies

### Video Exports (*.mp4)

```nginx
location /exports/ {
    add_header Cache-Control 'public, max-age=86400';
    add_header Content-Disposition 'attachment';
    add_header Accept-Ranges bytes;
}
```
- **24-hour cache**: Long-lived exports
- **Force download**: Browser downloads instead of playing

## Client Integration

### React Component

```jsx
import React, { useEffect, useRef } from 'react';
import Hls from 'hls.js';

function PlaybackPlayer({ cameraId, startTime, endTime }) {
  const videoRef = useRef(null);

  useEffect(() => {
    const startPlayback = async () => {
      // Request playback session
      const response = await fetch('http://localhost:8090/api/v1/playback/request', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          camera_id: cameraId,
          start_time: startTime,
          end_time: endTime,
          format: 'hls',
          user_id: 'current-user'
        })
      });

      const data = await response.json();

      // Initialize HLS.js
      if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(data.url);
        hls.attachMedia(videoRef.current);
        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          videoRef.current.play();
        });
      } else if (videoRef.current.canPlayType('application/vnd.apple.mpegurl')) {
        // Native HLS support (Safari)
        videoRef.current.src = data.url;
      }
    };

    startPlayback();
  }, [cameraId, startTime, endTime]);

  return <video ref={videoRef} controls width="1280" height="720" />;
}

export default PlaybackPlayer;
```

### HTML + HLS.js

```html
<!DOCTYPE html>
<html>
<head>
  <script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
</head>
<body>
  <video id="video" controls width="1280"></video>
  <script>
    async function playVideo() {
      const response = await fetch('http://localhost:8090/api/v1/playback/request', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          camera_id: '550e8400-e29b-41d4-a716-446655440000',
          start_time: '2024-01-20T10:00:00Z',
          end_time: '2024-01-20T11:00:00Z',
          format: 'hls',
          user_id: 'user123'
        })
      });

      const data = await response.json();
      const video = document.getElementById('video');

      if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(data.url);
        hls.attachMedia(video);
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = data.url;
      }
    }

    playVideo();
  </script>
</body>
</html>
```

## Configuration

**Environment Variables**:

```bash
# MinIO Configuration
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=cctv-recordings
MINIO_USE_SSL=false

# Milestone VMS (optional - for fallback)
MILESTONE_URL=http://milestone:8081
MILESTONE_USERNAME=admin
MILESTONE_PASSWORD=password

# Service Configuration
PORT=8090
WORK_DIR=/tmp/playback
CACHE_DIR=/tmp/playback/cache
HLS_BASE_URL=http://localhost:8091/hls
LOG_LEVEL=info
```

## Resource Usage

| Component | CPU | Memory | Disk |
|-----------|-----|--------|------|
| Playback Service | 1-4 cores | 512MB-2GB | 10GB cache |
| Nginx | 0.25 cores | 128MB | Minimal |
| FFmpeg (per job) | 5-10% | 100MB | Temp files |

**Transmuxing Performance**:
- 1-hour video: ~7 seconds
- Throughput: ~500x realtime
- CPU usage: <5% (stream copy, no transcoding)

## Troubleshooting

### No recordings found

**Error**: `No recordings available`

**Causes**:
1. Camera didn't record during time range
2. Recording Service wasn't running
3. MinIO bucket missing segments

**Solutions**:
```bash
# Check MinIO bucket
mc ls minio/cctv-recordings/recordings/{camera_id}/

# Check Recording Service logs
docker logs cctv-recording-service

# Verify camera in VMS Service
curl http://localhost:8081/api/v1/cameras/{camera_id}
```

### FFmpeg not found

**Error**: `exec: "ffmpeg": executable file not found`

**Solution**:
```dockerfile
# Ensure Dockerfile includes FFmpeg
RUN apk add ffmpeg
```

### Cache full, cannot evict

**Error**: `insufficient cache space`

**Solution**:
```go
// Increase cache size (main.go)
segmentCache, err := cache.NewSegmentCache(
    config.CacheDir,
    20*1024*1024*1024,  // Increase from 10GB to 20GB
    logger,
)
```

### Playback stuttering

**Causes**:
1. Slow network
2. Cache misses
3. Nginx overloaded

**Solutions**:
- Enable CDN caching for segments
- Increase segment cache size
- Add more Nginx workers
- Pre-fetch segments

## Security

⚠️ **Current State**: Prototype (no authentication)

**Production TODO**:
- [ ] JWT authentication middleware
- [ ] Restrict CORS origins
- [ ] Signed URLs with expiration
- [ ] Rate limiting per user
- [ ] Audit logging

## Performance Optimization

### Current Optimizations

✅ Stream copy transmuxing (NO re-encoding)
✅ LRU segment caching (10GB)
✅ Nginx sendfile for efficient delivery
✅ Proper cache headers (immutable segments)
✅ Independent HLS segments (parallel downloads)

### Future Optimizations

- [ ] CDN integration (CloudFront, Cloudflare)
- [ ] Adaptive bitrate streaming (multiple qualities)
- [ ] Segment prefetching (ML-based prediction)
- [ ] Parallel transmuxing (multiple segments concurrently)
- [ ] Redis caching for manifests
- [ ] Distributed caching across nodes

## Monitoring

**Health Check**:
```bash
curl http://localhost:8090/health
# Response: {"status":"healthy"}
```

**Prometheus Metrics**:
```bash
curl http://localhost:8090/metrics

# Key metrics:
# - playback_requests_total{status="success|error"}
# - playback_transmux_duration_seconds
# - cache_hit_rate
# - cache_size_bytes
# - cache_evictions_total
```

## Development

```bash
# Build
go build -o playback-service ./cmd/main.go

# Run locally
PORT=8090 \
MINIO_ENDPOINT=localhost:9000 \
./playback-service

# Docker build
docker build -t cctv-playback-service .

# Docker run
docker run -p 8090:8090 \
  -v /tmp/playback:/tmp/playback \
  cctv-playback-service
```

## Phase 3 Week 6 Summary

**Completed**:
- ✅ Source detection (MinIO vs Milestone)
- ✅ FFmpeg transmuxer (H.264 → HLS)
- ✅ LRU segment cache (10GB)
- ✅ MinIO client (segment retrieval)
- ✅ Playback orchestration use case
- ✅ HTTP API endpoints
- ✅ Nginx HLS delivery
- ✅ Docker integration

**Next Phase** (Phase 4):
- Object detection service (YOLOv8 Nano)
- React dashboard (live streaming, playback UI)
- Grid layouts with LiveKit integration

**Overall Progress**: ~85% complete
