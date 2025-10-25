# Phase 2: Storage & Recording - Implementation Plan

**Status**: 🚧 In Progress
**Progress**: MinIO Complete (20%), Storage Service In Progress
**Estimated Time**: ~32 hours

---

## **Overview**

Phase 2 adds video recording and storage capabilities to the RTA CCTV system, enabling:
- Continuous video recording from cameras
- Configurable storage modes (LOCAL/MILESTONE/BOTH)
- Video export and clip management
- Metadata tracking and search
- Playback of historical footage

---

## **Components**

### **1. MinIO Object Storage** ✅ COMPLETE

**Status**: 100% Complete
**Time Spent**: ~2 hours

**Delivered**:
- ✅ Docker configuration for MinIO server
- ✅ Automatic bucket initialization (4 buckets)
- ✅ Lifecycle policies (90d recordings, 7d exports, 30d thumbnails)
- ✅ Service user creation (recording, storage, playback)
- ✅ Prometheus metrics endpoint
- ✅ Web console (port 9001)
- ✅ Complete documentation

**Buckets Created**:
- `cctv-recordings`: Video segments (90-day retention)
- `cctv-exports`: User exports (7-day retention)
- `cctv-thumbnails`: Preview images (30-day retention)
- `cctv-clips`: Incident clips (indefinite retention)

**Storage Capacity**: ~1.9 PB for 500 cameras × 90 days

**Files**:
- `config/minio/init-buckets.sh`
- `config/minio/Dockerfile`
- `config/minio/README.md`
- `docker-compose.yml` (updated)

---

### **2. Video Storage Service** 🚧 IN PROGRESS

**Status**: 40% Complete (Planning & Documentation)
**Estimated Time**: ~8 hours remaining

**Purpose**: Orchestrates video storage across different backends

**Key Features**:
- Configurable storage modes:
  - **LOCAL**: Store in MinIO only
  - **MILESTONE**: Use Milestone VMS storage
  - **BOTH**: Dual recording (recommended)
- Multiple backend support (MinIO, S3, Filesystem)
- Segment metadata tracking in PostgreSQL
- Export generation (FFmpeg stitching)
- Automatic cleanup of expired data

**API Endpoints** (11 planned):
- `POST /api/v1/storage/segments` - Store segment metadata
- `GET /api/v1/storage/segments/{camera_id}` - List segments
- `DELETE /api/v1/storage/segments/{id}` - Delete segment
- `POST /api/v1/storage/exports` - Create export
- `GET /api/v1/storage/exports/{id}` - Export status
- `GET /api/v1/storage/exports/{id}/download` - Download export
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

**Database Tables**:
```sql
segments (
  id, camera_id, start_time, end_time,
  duration_seconds, size_bytes, storage_backend,
  storage_path, checksum, created_at
)

exports (
  id, camera_ids[], start_time, end_time,
  format, reason, status, file_path, file_size,
  download_url, expires_at, created_by, created_at
)
```

**Technology**: Go 1.21, Chi Router, MinIO SDK, PostgreSQL

**Resource Usage**: 0.5 CPU, 256MB RAM (idle), 2 CPU, 1GB (peak)

**Remaining Work**:
- [ ] Domain models (segments, exports)
- [ ] Storage backends (MinIO, S3, Filesystem)
- [ ] PostgreSQL repository
- [ ] HTTP handlers
- [ ] Export generation logic (FFmpeg)
- [ ] Background cleanup job
- [ ] Dockerfile
- [ ] Integration tests

---

### **3. Recording Service** ⏳ PENDING

**Status**: 0% Complete
**Estimated Time**: ~8 hours

**Purpose**: Continuous video recording from RTSP streams

**Key Features**:
- Pull RTSP streams from MediaMTX
- Record to 1-hour segments using FFmpeg
- H.264 copy (no transcoding)
- Automatic segment rotation
- Upload segments to Storage Service
- Thumbnail generation (every 5 minutes)
- Health monitoring per camera

**Recording Flow**:
```
1. Get camera RTSP URL from VMS Service
2. Reserve quota from Stream Counter
3. Pull RTSP stream from MediaMTX
4. Record to local disk (1-hour segments)
5. Upload segment to Storage Service (MinIO)
6. Generate thumbnails (JPEG @ 5-min intervals)
7. Delete local segment after upload
8. Repeat for next segment
```

**FFmpeg Command** (per camera):
```bash
ffmpeg -i rtsp://mediamtx:8554/camera_{id} \
  -c:v copy \
  -c:a copy \
  -f segment \
  -segment_time 3600 \
  -segment_format mpegts \
  -reset_timestamps 1 \
  /recordings/%Y-%m-%d-%H-00-00.ts
```

**Technology**: Go 1.21, FFmpeg wrapper, Storage Service client

**Resource Usage**:
- Per camera: 0.01 CPU (H.264 copy), ~10MB RAM
- 500 cameras: ~5 CPU, ~5GB RAM

**API Endpoints**:
- `POST /api/v1/recording/start/{camera_id}` - Start recording
- `POST /api/v1/recording/stop/{camera_id}` - Stop recording
- `GET /api/v1/recording/status` - Recording status
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

**Remaining Work**:
- [ ] FFmpeg wrapper library
- [ ] Recording manager (per-camera goroutines)
- [ ] Segment upload logic
- [ ] Thumbnail generation
- [ ] Storage Service client
- [ ] Health monitoring
- [ ] Dockerfile
- [ ] Integration tests

---

### **4. Metadata Service** ⏳ PENDING

**Status**: 0% Complete
**Estimated Time**: ~8 hours

**Purpose**: Rich metadata, search, tags, and annotations

**Key Features**:
- Full-text search for incidents
- Video tags and categories
- Annotations (markers on timeline)
- Incident tracking
- Evidence chain of custody
- Advanced filtering

**Database Tables**:
```sql
tags (
  id, name, category, color, created_at
)

video_tags (
  segment_id, tag_id, user_id, created_at
)

annotations (
  id, segment_id, timestamp, type, content,
  user_id, created_at
)

incidents (
  id, title, description, severity, status,
  camera_ids[], start_time, end_time,
  tags[], assigned_to, created_by, created_at
)
```

**API Endpoints**:
- `POST /api/v1/metadata/tags` - Create tag
- `POST /api/v1/metadata/segments/{id}/tags` - Tag segment
- `POST /api/v1/metadata/annotations` - Add annotation
- `GET /api/v1/metadata/search` - Search segments
- `POST /api/v1/metadata/incidents` - Create incident
- `GET /api/v1/metadata/incidents/{id}` - Get incident

**Search Capabilities**:
- Full-text search (PostgreSQL tsvector)
- Date/time range filtering
- Camera filtering (by ID or source)
- Tag filtering
- Incident severity filtering
- User filtering (created by, assigned to)

**Technology**: Go 1.21, PostgreSQL with full-text search

**Resource Usage**: 0.5 CPU, 256MB RAM

**Remaining Work**:
- [ ] Domain models (tags, annotations, incidents)
- [ ] PostgreSQL repository with full-text search
- [ ] Search query builder
- [ ] HTTP handlers
- [ ] Dockerfile
- [ ] Integration tests

---

### **5. Playback Service** ⏳ PENDING

**Status**: 0% Complete
**Estimated Time**: ~6 hours

**Purpose**: Unified playback from local storage and Milestone VMS

**Key Features**:
- Unified API for playback (local + Milestone)
- HLS transmux for web browsers
- Segment stitching across time ranges
- Adaptive bitrate (multiple qualities)
- Timeline scrubbing
- Frame-accurate seeking

**Playback Flow**:
```
1. Client requests playback for time range
2. Query Storage Service for segments
3. If LOCAL mode: Stream from MinIO
4. If MILESTONE mode: Proxy from Milestone
5. If BOTH mode: Prefer local, fallback to Milestone
6. Transmux to HLS (if needed)
7. Cache HLS segments
8. Return playlist URL to client
```

**API Endpoints**:
- `GET /api/v1/playback/{camera_id}/hls` - HLS playlist
- `GET /api/v1/playback/{camera_id}/segments` - List segments
- `GET /api/v1/playback/download` - Direct download
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

**HLS Transmux**:
```bash
# Convert MPEG-TS segment to HLS
ffmpeg -i segment.ts \
  -c:v copy \
  -c:a copy \
  -f hls \
  -hls_time 10 \
  -hls_list_size 0 \
  playlist.m3u8
```

**Caching Strategy**:
- Cache HLS playlists (5-minute TTL)
- Cache HLS segments (1-hour TTL)
- Cache frequently accessed segments in Valkey
- Serve from cache for repeated requests

**Technology**: Go 1.21, FFmpeg, HLS streaming

**Resource Usage**: 1 CPU (FFmpeg transmux), 512MB RAM

**Remaining Work**:
- [ ] Segment stitching logic
- [ ] HLS transmux wrapper
- [ ] Caching layer (Valkey)
- [ ] Milestone proxy
- [ ] HTTP handlers
- [ ] Dockerfile
- [ ] Integration tests

---

## **Database Schema (PostgreSQL)**

### Complete Schema for Phase 2

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Segments table
CREATE TABLE segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_seconds INTEGER NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_backend VARCHAR(50) NOT NULL,
    storage_path TEXT NOT NULL,
    checksum VARCHAR(64),
    thumbnail_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_segments_camera_time ON segments(camera_id, start_time, end_time);
CREATE INDEX idx_segments_created_at ON segments(created_at);

-- Exports table
CREATE TABLE exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_ids UUID[] NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    format VARCHAR(10) NOT NULL DEFAULT 'mp4',
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    file_path TEXT,
    file_size BIGINT,
    download_url TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_exports_status ON exports(status);
CREATE INDEX idx_exports_created_at ON exports(created_at);

-- Tags table
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50),
    color VARCHAR(7),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Video tags (many-to-many)
CREATE TABLE video_tags (
    segment_id UUID REFERENCES segments(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (segment_id, tag_id)
);

-- Annotations table
CREATE TABLE annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID REFERENCES segments(id) ON DELETE CASCADE,
    timestamp_offset INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_annotations_segment ON annotations(segment_id);

-- Incidents table
CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    camera_ids UUID[] NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    tags TEXT[],
    assigned_to VARCHAR(255),
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_incidents_status ON incidents(status);
CREATE INDEX idx_incidents_created_at ON incidents(created_at);
CREATE INDEX idx_incidents_severity ON incidents(severity);

-- Full-text search
ALTER TABLE incidents ADD COLUMN search_vector tsvector;
CREATE INDEX idx_incidents_search ON incidents USING gin(search_vector);

CREATE OR REPLACE FUNCTION incidents_search_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER incidents_search_update
    BEFORE INSERT OR UPDATE ON incidents
    FOR EACH ROW
    EXECUTE FUNCTION incidents_search_trigger();
```

---

## **Integration with Existing Services**

### Kong Routes (to be added)

```yaml
# Storage Service routes
- name: storage-segments
  paths: [/api/v1/storage/segments]
  service: storage-service

- name: storage-exports
  paths: [/api/v1/storage/exports]
  service: storage-service

# Recording Service routes
- name: recording-control
  paths: [/api/v1/recording]
  service: recording-service

# Metadata Service routes
- name: metadata-search
  paths: [/api/v1/metadata]
  service: metadata-service

# Playback Service routes
- name: playback-hls
  paths: [/api/v1/playback]
  service: playback-service
```

---

## **Timeline**

| Service | Status | Time Estimate | Dependencies |
|---------|--------|---------------|--------------|
| MinIO Setup | ✅ Complete | 2 hours | None |
| Storage Service | 🚧 40% | 8 hours | MinIO, PostgreSQL |
| Recording Service | ⏳ Pending | 8 hours | Storage Service, MediaMTX |
| Metadata Service | ⏳ Pending | 8 hours | PostgreSQL |
| Playback Service | ⏳ Pending | 6 hours | Storage Service, FFmpeg |
| **Total** | **20%** | **32 hours** | - |

---

## **Next Steps**

1. **Complete Storage Service** (~8 hours)
   - Implement storage backends
   - Add PostgreSQL repository
   - Create HTTP API
   - Write Dockerfile

2. **Implement Recording Service** (~8 hours)
   - FFmpeg wrapper
   - Recording manager
   - Segment upload logic

3. **Build Metadata Service** (~8 hours)
   - PostgreSQL schema
   - Full-text search
   - API handlers

4. **Create Playback Service** (~6 hours)
   - HLS transmux
   - Segment stitching
   - Caching layer

5. **Integration Testing** (~2 hours)
   - End-to-end recording test
   - Playback test
   - Export test

---

## **Success Criteria**

Phase 2 will be considered complete when:

✅ MinIO is running and initialized (DONE)
- [ ] Storage Service can store and retrieve segments
- [ ] Recording Service is recording from 1+ camera
- [ ] Metadata Service supports search and tagging
- [ ] Playback Service can generate HLS playlists
- [ ] End-to-end test: Record → Store → Search → Playback
- [ ] All services dockerized and in docker-compose.yml
- [ ] Documentation complete for all services
- [ ] Integration tests passing

---

**Status**: MinIO complete, Storage Service in progress. Continuing implementation...
