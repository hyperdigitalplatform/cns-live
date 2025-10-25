# Recording Service

## Overview

The Recording Service handles continuous video recording from RTSP streams using FFmpeg. It records video segments, generates thumbnails, and uploads them to the Storage Service.

## Features

- ✅ Continuous RTSP recording with FFmpeg
- ✅ 1-hour segment rotation
- ✅ H.264 copy (no transcoding for low CPU usage)
- ✅ Automatic thumbnail generation (every 5 minutes)
- ✅ Upload to Storage Service
- ✅ Quota management via Stream Counter
- ✅ Heartbeat mechanism
- ✅ Graceful shutdown
- ✅ Per-camera recording management
- ✅ Prometheus metrics

## Architecture

```
┌────────────────────────────────────────┐
│     Recording Service (Go)             │
├────────────────────────────────────────┤
│  HTTP API Layer                        │
│  ├── Start Recording                   │
│  ├── Stop Recording                    │
│  ├── Get Status                        │
│  └── List Recordings                   │
├────────────────────────────────────────┤
│  Recording Manager                     │
│  ├── Multiple Camera Management        │
│  ├── Quota Reservation                 │
│  ├── Heartbeat Sender                  │
│  └── Segment Upload                    │
├────────────────────────────────────────┤
│  FFmpeg Recorder (per camera)          │
│  ├── RTSP Input                        │
│  ├── Segment Output (1-hour)           │
│  ├── Thumbnail Generation              │
│  └── Process Monitoring                │
└────────────────────────────────────────┘
         │           │            │
         ↓           ↓            ↓
    ┌────────┐  ┌─────────┐  ┌──────────┐
    │  VMS   │  │ Stream  │  │ Storage  │
    │Service │  │ Counter │  │ Service  │
    └────────┘  └─────────┘  └──────────┘
```

## Recording Flow

1. **Start Recording Request** → API call with camera_id
2. **Get Camera Details** → Call VMS Service
3. **Get RTSP URL** → Call VMS Service
4. **Reserve Quota** → Call Stream Counter
5. **Start FFmpeg** → Begin recording
6. **Record Segments** → 1-hour MPEG-TS files
7. **Generate Thumbnails** → Every 5 minutes
8. **Upload Segment** → Call Storage Service
9. **Send Heartbeats** → Every 30 seconds
10. **Delete Local File** → After successful upload

## API Endpoints

### Start Recording
```http
POST /api/v1/recording/start/{camera_id}
```

**Response (200 OK)**:
```json
{
  "camera_id": "uuid",
  "camera_name": "Camera 001",
  "rtsp_url": "rtsp://...",
  "status": "RECORDING",
  "started_at": "2025-01-23T10:00:00Z",
  "segment_count": 0,
  "total_bytes": 0,
  "reservation_id": "uuid"
}
```

### Stop Recording
```http
POST /api/v1/recording/stop/{camera_id}
```

**Response (200 OK)**:
```json
{
  "message": "Recording stopped",
  "camera_id": "uuid"
}
```

### Get Recording Status
```http
GET /api/v1/recording/status/{camera_id}
```

**Response (200 OK)**:
```json
{
  "camera_id": "uuid",
  "camera_name": "Camera 001",
  "status": "RECORDING",
  "started_at": "2025-01-23T10:00:00Z",
  "last_segment_at": "2025-01-23T11:00:00Z",
  "segment_count": 2,
  "total_bytes": 314572800
}
```

### List All Recordings
```http
GET /api/v1/recording/status
```

**Response (200 OK)**:
```json
{
  "total_recordings": 150,
  "active_recordings": 150,
  "total_segments": 3600,
  "total_bytes": 56778342400,
  "recordings_by_camera": {
    "uuid1": { ... },
    "uuid2": { ... }
  }
}
```

## FFmpeg Configuration

### Recording Command
```bash
ffmpeg \
  -rtsp_transport tcp \
  -i rtsp://mediamtx:8554/camera_{id} \
  -c:v copy \
  -c:a copy \
  -f segment \
  -segment_time 3600 \
  -segment_format mpegts \
  -segment_atclocktime 1 \
  -strftime 1 \
  -reset_timestamps 1 \
  /recordings/{camera_id}/%Y-%m-%d-%H-%M-%S.ts
```

**Parameters Explained**:
- `-rtsp_transport tcp`: Use TCP for RTSP (better reliability)
- `-i`: Input RTSP stream
- `-c:v copy`: Copy video codec (no transcoding)
- `-c:a copy`: Copy audio codec (no transcoding)
- `-f segment`: Output segmented files
- `-segment_time 3600`: 1-hour segments
- `-segment_format mpegts`: MPEG-TS container
- `-segment_atclocktime 1`: Align segments to clock time
- `-strftime 1`: Use strftime for filenames
- `-reset_timestamps 1`: Reset timestamps per segment

### Thumbnail Generation
```bash
ffmpeg \
  -i segment.ts \
  -ss 5 \
  -vframes 1 \
  -vf scale=640:360 \
  -y \
  thumbnail.jpg
```

## Environment Variables

```bash
# Service Configuration
PORT=8083
LOG_LEVEL=info
LOG_FORMAT=json

# Directories
OUTPUT_DIR=/tmp/recordings

# Service URLs
STORAGE_SERVICE_URL=http://storage-service:8082
VMS_SERVICE_URL=http://vms-service:8081
STREAM_COUNTER_URL=http://stream-counter:8087

# Recording Settings
SEGMENT_SECONDS=3600  # 1 hour
```

## Resource Usage

### Per Camera Recording
- **CPU**: 0.01 core (H.264 copy, no transcoding)
- **RAM**: ~10MB
- **Disk I/O**: 2-4 MB/s (writing segments)
- **Network**: 2-4 Mbps (RTSP stream)

### Total for 500 Cameras
- **CPU**: ~5 cores
- **RAM**: ~5 GB
- **Disk I/O**: 1-2 GB/s
- **Network**: 1-2 Gbps

### Temporary Storage
Each segment is ~150 MB (1 hour @ 4 Mbps).
With 500 cameras and 1-hour retention before upload:
- **Disk Space**: 500 × 150 MB = ~75 GB temporary storage

## Docker Setup

```yaml
recording-service:
  image: rta/recording-service:latest
  ports:
    - "8083:8083"
  volumes:
    - /tmp/recordings:/tmp/recordings
  environment:
    OUTPUT_DIR: /tmp/recordings
    STORAGE_SERVICE_URL: http://storage-service:8082
    VMS_SERVICE_URL: http://vms-service:8081
    STREAM_COUNTER_URL: http://stream-counter:8087
  depends_on:
    - storage-service
    - vms-service
    - stream-counter
```

## Testing

### Start Recording
```bash
curl -X POST http://localhost:8083/api/v1/recording/start/123e4567-e89b-12d3-a456-426614174000
```

### Check Status
```bash
curl http://localhost:8083/api/v1/recording/status/123e4567-e89b-12d3-a456-426614174000
```

### List All Recordings
```bash
curl http://localhost:8083/api/v1/recording/status
```

### Stop Recording
```bash
curl -X POST http://localhost:8083/api/v1/recording/stop/123e4567-e89b-12d3-a456-426614174000
```

## Metrics

Prometheus metrics at `/metrics`:

```
# Recordings
recording_active_count
recording_total_segments
recording_total_bytes

# Per camera
recording_segments_count{camera_id}
recording_bytes_total{camera_id}

# Uploads
recording_uploads_total{status}
recording_upload_duration_seconds

# Errors
recording_errors_total{type}
```

## Troubleshooting

### FFmpeg Not Recording

```bash
# Check FFmpeg is installed
docker exec recording-service ffmpeg -version

# Check FFmpeg process
docker exec recording-service ps aux | grep ffmpeg

# Check logs
docker logs recording-service | grep ERROR
```

### Segment Upload Fails

```bash
# Check Storage Service connectivity
curl http://storage-service:8082/health

# Check disk space
df -h /tmp/recordings

# Check service logs
docker logs recording-service | grep "upload"
```

### High CPU Usage

```bash
# Verify H.264 copy (should NOT see "Encoding" in logs)
docker logs recording-service | grep -i encoding

# Check FFmpeg command (should have -c:v copy)
docker logs recording-service | grep ffmpeg
```

## Production Considerations

1. **Storage**: Use fast SSD/NVMe for temporary segment storage
2. **Network**: Ensure sufficient bandwidth (2 Gbps+ for 500 cameras)
3. **Monitoring**: Alert on failed uploads, high error rates
4. **Cleanup**: Automatic cleanup of old temporary files
5. **Redundancy**: Run multiple instances with load balancing
6. **Graceful Shutdown**: Service stops all recordings before exit

## References

- FFmpeg Documentation: https://ffmpeg.org/documentation.html
- FFmpeg Segmenting: https://ffmpeg.org/ffmpeg-formats.html#segment
- RTSP Protocol: RFC 2326
