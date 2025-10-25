# Storage Service

## Overview

The Storage Service orchestrates video storage across different backends based on configuration:
- **LOCAL**: Store only in MinIO/S3
- **MILESTONE**: Store only in Milestone VMS
- **BOTH**: Store in both MinIO and Milestone (dual recording)

## Features

- ✅ Configurable storage modes (LOCAL/MILESTONE/BOTH)
- ✅ Multiple backend support (MinIO, S3, Filesystem)
- ✅ Automatic segment management
- ✅ Metadata tracking in PostgreSQL
- ✅ Export generation (clips, incidents)
- ✅ Thumbnail generation
- ✅ Cleanup jobs for expired data
- ✅ Prometheus metrics

## Architecture

```
┌────────────────────────────────────────┐
│      Storage Service (Go)              │
├────────────────────────────────────────┤
│  HTTP API Layer                        │
│  ├── Store Segment                     │
│  ├── Get Segment                       │
│  ├── Delete Segment                    │
│  ├── Create Export                     │
│  └── List Segments                     │
├────────────────────────────────────────┤
│  Storage Orchestrator                  │
│  ├── Mode: LOCAL/MILESTONE/BOTH        │
│  ├── Backend Selection                 │
│  └── Dual Write Logic                  │
├────────────────────────────────────────┤
│  Backend Implementations               │
│  ├── MinIO Backend                     │
│  ├── S3 Backend                        │
│  ├── Filesystem Backend                │
│  └── Milestone Backend (Proxy)         │
├────────────────────────────────────────┤
│  Metadata Repository                   │
│  └── PostgreSQL (Segments, Exports)    │
└────────────────────────────────────────┘
         │              │
         ↓              ↓
    ┌────────┐    ┌──────────┐
    │ MinIO  │    │PostgreSQL│
    └────────┘    └──────────┘
```

## Storage Modes

### LOCAL Mode
- Stores all recordings in MinIO
- Milestone is read-only (live streaming only)
- Full control over retention
- Best for: Maximum storage capacity, custom retention

### MILESTONE Mode
- Uses Milestone VMS for all storage
- No local recording
- Milestone handles retention
- Best for: Existing Milestone infrastructure, minimal local storage

### BOTH Mode (Recommended)
- Records to both MinIO and Milestone simultaneously
- Provides redundancy
- Allows custom retention on local storage
- Best for: High availability, compliance requirements

## API Endpoints

### Store Segment
```http
POST /api/v1/storage/segments
Content-Type: application/json

{
  "camera_id": "uuid",
  "start_time": "2025-01-23T10:00:00Z",
  "end_time": "2025-01-23T11:00:00Z",
  "file_path": "/recordings/segment.ts",
  "size_bytes": 157286400,
  "duration_seconds": 3600
}
```

### Get Segment
```http
GET /api/v1/storage/segments/{camera_id}?start={timestamp}&end={timestamp}
```

### Create Export
```http
POST /api/v1/storage/exports
Content-Type: application/json

{
  "camera_ids": ["uuid1", "uuid2"],
  "start_time": "2025-01-23T10:00:00Z",
  "end_time": "2025-01-23T11:00:00Z",
  "format": "mp4",
  "reason": "incident"
}
```

### Get Export Status
```http
GET /api/v1/storage/exports/{export_id}
```

### Download Export
```http
GET /api/v1/storage/exports/{export_id}/download
```

## Database Schema

### segments Table
```sql
CREATE TABLE segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_seconds INTEGER NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_backend VARCHAR(50) NOT NULL,  -- MINIO, S3, FILESYSTEM, MILESTONE
    storage_path TEXT NOT NULL,
    checksum VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX idx_camera_time (camera_id, start_time, end_time),
    INDEX idx_created_at (created_at)
);
```

### exports Table
```sql
CREATE TABLE exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_ids UUID[] NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    format VARCHAR(10) NOT NULL,
    reason TEXT,
    status VARCHAR(20) NOT NULL,  -- PENDING, PROCESSING, COMPLETED, FAILED
    file_path TEXT,
    file_size BIGINT,
    download_url TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

## Configuration

### Environment Variables

```bash
# Storage Mode
STORAGE_MODE=BOTH  # LOCAL, MILESTONE, BOTH

# Backend Selection
STORAGE_BACKEND=MINIO  # MINIO, S3, FILESYSTEM

# MinIO Configuration
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=storage-service
MINIO_SECRET_KEY=changeme_storage
MINIO_USE_SSL=false
MINIO_BUCKET_RECORDINGS=cctv-recordings
MINIO_BUCKET_EXPORTS=cctv-exports
MINIO_BUCKET_THUMBNAILS=cctv-thumbnails

# Database
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=cctv
POSTGRES_USER=cctv
POSTGRES_PASSWORD=changeme

# Service
PORT=8082
LOG_LEVEL=info
```

## Quick Start

### Development
```bash
# Install dependencies
go mod download

# Run service
go run cmd/main.go
```

### Docker
```bash
# Build image
docker build -t rta/storage-service:latest .

# Run container
docker run -d \
  --name storage-service \
  -p 8082:8082 \
  -e STORAGE_MODE=BOTH \
  -e MINIO_ENDPOINT=minio:9000 \
  rta/storage-service:latest
```

## Usage Examples

### Store Video Segment

```bash
curl -X POST http://localhost:8082/api/v1/storage/segments \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "123e4567-e89b-12d3-a456-426614174000",
    "start_time": "2025-01-23T10:00:00Z",
    "end_time": "2025-01-23T11:00:00Z",
    "file_path": "/tmp/segment.ts",
    "size_bytes": 157286400,
    "duration_seconds": 3600
  }'
```

### Query Segments

```bash
curl "http://localhost:8082/api/v1/storage/segments/123e4567-e89b-12d3-a456-426614174000?start=2025-01-23T00:00:00Z&end=2025-01-23T23:59:59Z"
```

### Create Export

```bash
curl -X POST http://localhost:8082/api/v1/storage/exports \
  -H "Content-Type: application/json" \
  -d '{
    "camera_ids": ["123e4567-e89b-12d3-a456-426614174000"],
    "start_time": "2025-01-23T10:00:00Z",
    "end_time": "2025-01-23T10:30:00Z",
    "format": "mp4",
    "reason": "Traffic incident investigation"
  }'
```

### Get Export Status

```bash
curl http://localhost:8082/api/v1/storage/exports/{export_id}
```

Response:
```json
{
  "id": "export-uuid",
  "status": "COMPLETED",
  "download_url": "http://minio:9000/cctv-exports/export-uuid.mp4?...",
  "file_size": 52428800,
  "expires_at": "2025-01-30T10:00:00Z",
  "created_at": "2025-01-23T10:00:00Z",
  "completed_at": "2025-01-23T10:05:00Z"
}
```

## Metrics

Prometheus metrics at `/metrics`:

```
# Storage operations
storage_segments_stored_total{backend,mode}
storage_segments_retrieved_total{backend}
storage_segments_deleted_total{backend}

# Exports
storage_exports_created_total{status}
storage_export_duration_seconds{status}

# Storage usage
storage_bytes_used{backend,bucket}
storage_segments_count{camera_id}

# Errors
storage_errors_total{operation,backend}
```

## Performance

| Operation | Latency (p50) | Latency (p99) |
|-----------|---------------|---------------|
| Store Segment (metadata) | <10ms | <50ms |
| Get Segment List | <50ms | <200ms |
| Create Export | <100ms | <500ms |
| Upload to MinIO | Depends on size | N/A |

## Resource Usage

- **CPU**: 0.5 core (idle), 2 cores (peak during exports)
- **RAM**: 256MB (idle), 1GB (peak)
- **Network**: Depends on video upload/download rates

## Troubleshooting

### Segment Upload Fails

```bash
# Check MinIO connection
curl http://minio:9000/minio/health/live

# Check credentials
mc alias set test http://minio:9000 storage-service password
mc ls test/cctv-recordings

# Check service logs
docker logs storage-service | grep ERROR
```

### Export Generation Slow

```bash
# Check FFmpeg progress
docker exec storage-service ps aux | grep ffmpeg

# Check disk I/O
iostat -x 1

# Check database queries
docker exec cctv-postgres psql -U cctv -c "SELECT * FROM pg_stat_activity WHERE state = 'active';"
```

## References

- MinIO Go SDK: https://min.io/docs/minio/linux/developers/go/minio-go.html
- PostgreSQL with Go: https://github.com/lib/pq
- FFmpeg Documentation: https://ffmpeg.org/documentation.html
