# RTA CCTV System - Complete Requirements (CORRECTED)

## **SYSTEM OVERVIEW**

**Project**: RTA (Roads and Transport Authority) CCTV Video Management System
**Purpose**: Multi-agency live streaming and video storage platform with AI analytics
**Scale**: 500 concurrent cameras, 1000+ simultaneous viewers
**Architecture**: Hybrid VMS with configurable storage (Local/Milestone/Both)

---

## **KEY CORRECTIONS FROM ORIGINAL DOCUMENT**

### **1. Storage Strategy**
- ❌ **ORIGINAL**: Assumed Milestone-only storage
- ✅ **CORRECTED**: Hybrid storage with admin configuration:
  - **Mode 1**: Local storage only
  - **Mode 2**: Milestone VMS only
  - **Mode 3**: Both (dual recording)
- **Requirement**: Sub-500ms playback latency for local storage

### **2. Cache/Database Technology**
- ❌ **ORIGINAL**: Mixed Redis/Valkey references
- ✅ **CORRECTED**: **Valkey ONLY** throughout entire system
  - No Redis dependencies
  - All caching, quotas, and queues use Valkey

### **3. Video Codec Strategy**
- ❌ **ORIGINAL**: Support H.264 and H.265
- ✅ **CORRECTED**: **H.264 ONLY**
  - All cameras standardized on H.264
  - No transcoding needed (copy codec)
  - No GPU required for video processing

### **4. Missing Services Identified**
- ✅ **VMS Service**: Dedicated Milestone integration layer (ADDED)
- ✅ **Video Storage Service**: Configurable storage orchestration (ADDED)
- ✅ **Recording Service**: Continuous recording with segmentation (ADDED)
- ✅ **Metadata Service**: Search, tags, annotations (ADDED)
- ✅ **Clip Management**: Incident extraction and evidence management (ADDED)
- ✅ **Object Detection Service**: AI analytics with small visual models (ADDED)

### **5. Hardware Footprint**
- ❌ **ORIGINAL**: No resource optimization focus
- ✅ **CORRECTED**: Optimized for low footprint:
  - Total RAM: <10GB
  - Total CPU: <12 cores
  - **No GPU required** (optional for AI acceleration)
  - Storage: 500TB+ for video (configurable based on retention)

---

## **PHASE 1: CORE INFRASTRUCTURE**

### **1.1: MediaMTX RTSP Ingest Service**

**Purpose**: Ingest 500 RTSP streams and distribute to multiple consumers

**FUNCTIONAL REQUIREMENTS**:
- Handle 500 concurrent RTSP streams (H.264 only)
- Track stream count per source in Valkey with atomic operations
- Exponential backoff reconnection (1s → 30s)
- **Tee output**:
  - Output 1: LiveKit Ingress (live viewing)
  - Output 2: Recording Service (storage)
  - Output 3: Object Detection Service (AI analytics)
- Circuit breaker pattern (3 failures → open 30s)
- Support on-demand stream activation (resource efficiency)

**PERFORMANCE**:
- Memory: 512MB per instance
- CPU: 1 core per 100 cameras
- Stream buffer: 256KB per camera (optimized)
- Connection pool: max 600 connections
- Latency: <50ms relay time

**CONFIGURATION**:
```yaml
# mediamtx.yml - OPTIMIZED FOR LOW LATENCY
logLevel: warn
readTimeout: 5s
writeTimeout: 5s
readBufferCount: 64  # Reduced for memory efficiency

protocols: [tcp]
encryption: no  # Internal network only

# On-demand activation
sourceOnDemand: yes
sourceOnDemandStartTimeout: 5s
sourceOnDemandCloseAfter: 10s

# Multiple outputs per path
paths:
  camera_*:
    source: "rtsp://vms-service:8554/camera_$1"

    # Output 1: LiveKit Ingress
    runOnReady: "http://livekit-ingress:8080/ingest/camera_$1"

    # Output 2: Recording Service
    runOnReady: "http://recording-service:8080/record/camera_$1"

    # Output 3: Object Detection (sampled)
    runOnReady: "http://object-detection:8080/analyze/camera_$1"
```

**DELIVERABLES**:
1. mediamtx.yml configuration
2. Dockerfile (Alpine, <100MB)
3. Health check script
4. Integration test with mock RTSP
5. Load test: 500 streams for 1 hour

---

### **1.2: Valkey Stream Counter Service**

**Purpose**: Distributed quota enforcement with atomic guarantees

**FUNCTIONAL REQUIREMENTS**:
- Track active streams per source (DUBAI_POLICE: 50, METRO: 30, BUS: 20, OTHER: 400)
- Atomic reserve/release operations via Lua scripts
- Auto-cleanup stale reservations (60s job)
- Real-time statistics via WebSocket

**DATA MODEL**:
```
Valkey Keys:
- stream:count:{source}        → INTEGER (current count)
- stream:limit:{source}        → INTEGER (max allowed)
- stream:reservation:{uuid}    → HASH (TTL: 3600s)
- stream:heartbeat:{uuid}      → STRING (TTL: 30s)
```

**LUA SCRIPTS** (CRITICAL FOR ATOMICITY):

**reserve_stream.lua**:
```lua
-- ATOMIC: Check limit + Increment + Create reservation
local source = KEYS[1]
local limit_key = "stream:limit:" .. source
local count_key = "stream:count:" .. source
local reservation_id = ARGV[1]
local camera_id = ARGV[2]
local user_id = ARGV[3]
local ttl = tonumber(ARGV[4])

local limit = tonumber(redis.call('GET', limit_key) or 0)
local current = tonumber(redis.call('GET', count_key) or 0)

if current >= limit then
    return {0, current, limit}  -- Reject
end

local new_count = redis.call('INCR', count_key)

if new_count <= limit then
    local reservation_key = "stream:reservation:" .. reservation_id
    redis.call('HSET', reservation_key, 'camera_id', camera_id, 'source', source, 'user_id', user_id)
    redis.call('EXPIRE', reservation_key, ttl)
    return {1, new_count, limit}  -- Success
else
    redis.call('DECR', count_key)  -- Rollback
    return {0, limit, limit}  -- Reject
end
```

**release_stream.lua**:
```lua
-- ATOMIC: Decrement + Delete reservation
local reservation_id = ARGV[1]
local reservation_key = "stream:reservation:" .. reservation_id
local source = redis.call('HGET', reservation_key, 'source')

if source then
    redis.call('DECR', "stream:count:" .. source)
    redis.call('DEL', reservation_key)
    return 1
end
return 0
```

**API ENDPOINTS**:
- POST /api/v1/stream/reserve
- DELETE /api/v1/stream/release/{id}
- POST /api/v1/stream/heartbeat/{id}
- GET /api/v1/stream/stats
- WebSocket /ws/stream/stats

**PERFORMANCE**:
- Operation latency: <10ms p99
- Throughput: 10,000 ops/second
- Memory: 256MB
- CPU: 0.5 core

**DELIVERABLES**:
1. Go service with chi router
2. Lua scripts with tests
3. Dockerfile (<50MB)
4. Valkey cluster setup guide
5. API documentation (OpenAPI)
6. Performance test results

---

### **1.3: Kong API Gateway**

**Purpose**: API gateway with Valkey-based rate limiting and JWT auth

**FUNCTIONAL REQUIREMENTS**:
- JWT validation (RS256 with RTA IAM public key)
- Source-based rate limiting using Valkey
- Bilingual error messages (Arabic/English)
- WebSocket support for real-time updates
- Request correlation ID propagation

**CUSTOM LUA PLUGIN**: Source-Based Rate Limiter
```lua
-- kong/plugins/source-rate-limit/handler.lua
local ValkeyCLient = require "resty.valkey"

function plugin:access(conf)
    local valkey = ValkeyCLient:new()
    valkey:connect("valkey", 6379)

    -- Extract source from JWT claims
    local claims = kong.ctx.shared.jwt_claims
    local source = claims.camera_source

    -- Check limit
    local script = [[
        local count_key = "stream:count:" .. ARGV[1]
        local limit_key = "stream:limit:" .. ARGV[1]
        local current = tonumber(redis.call('GET', count_key) or 0)
        local limit = tonumber(redis.call('GET', limit_key) or 0)
        return {current, limit}
    ]]

    local result = valkey:eval(script, 0, source)
    local current, limit = result[1], result[2]

    if current >= limit then
        return kong.response.exit(429, {
            error = {
                code = "RATE_LIMIT_EXCEEDED",
                message_en = string.format("Camera limit reached for %s (%d/%d)", source, current, limit),
                message_ar = string.format("تم الوصول إلى حد الكاميرات لـ %s (%d/%d)", source, current, limit),
                source = source,
                current = current,
                limit = limit
            }
        })
    end
end
```

**ROUTES**:
- /api/v1/stream/* → go-api (stream management)
- /api/v1/cameras/* → vms-service (camera list)
- /api/v1/playback/* → playback-service (recordings)
- /api/v1/clips/* → metadata-service (clip management)
- /api/v1/ptz/* → vms-service (PTZ control)
- /health → all services
- /metrics → Prometheus (IP whitelist)

**DELIVERABLES**:
1. kong.yaml declarative config
2. Custom Lua plugin source
3. Error message templates (AR/EN)
4. Integration tests
5. Load test scripts (K6)

---

## **PHASE 2: VMS & STORAGE LAYER**

### **2.1: VMS Service** (NEW - CRITICAL)

**Purpose**: Unified interface to Milestone XProtect VMS

**FUNCTIONAL REQUIREMENTS**:
- Connect to multiple Milestone Recording Servers
- Persistent connection pool (5 connections per server)
- In-memory camera metadata cache (5min TTL)
- RTSP URL generation for live streams
- PTZ control proxy with command validation
- Recording export API
- Background camera sync (every 10 minutes)

**ARCHITECTURE**:
```
VMS Service (Go)
├── Milestone SDK Wrapper
├── Connection Pool Manager
├── Camera Cache (in-memory)
├── PTZ Command Queue
└── Export Job Manager
```

**DATA MODEL**:
```go
type Camera struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    NameAr      string    `json:"name_ar"`
    Source      string    `json:"source"`  // DUBAI_POLICE, METRO, BUS, OTHER
    RTSPURL     string    `json:"rtsp_url"`
    PTZEnabled  bool      `json:"ptz_enabled"`
    Status      string    `json:"status"`  // ONLINE, OFFLINE, ERROR
    RecordingServer string `json:"recording_server"`
    Metadata    map[string]interface{} `json:"metadata"`
    LastUpdate  time.Time `json:"-"`
}
```

**API ENDPOINTS**:
- GET /vms/cameras → List all cameras (cached)
- GET /vms/cameras/{id} → Get camera details
- GET /vms/cameras/{id}/stream → Get RTSP URL
- POST /vms/cameras/{id}/ptz → PTZ control
- POST /vms/recordings/export → Export recording
- GET /vms/health → Health check

**PERFORMANCE**:
- Camera list: <100ms (from cache)
- RTSP URL generation: <50ms
- PTZ command: <100ms latency
- Memory: 256MB
- CPU: 0.5 core

**DELIVERABLES**:
1. Complete Go service
2. Milestone SDK wrapper
3. API documentation (OpenAPI)
4. Integration tests with Milestone simulator
5. Performance benchmarks

---

### **2.2: Video Storage Service** (NEW - CORE)

**Purpose**: Configurable storage orchestration (Local/Milestone/Both)

**FUNCTIONAL REQUIREMENTS**:
- Admin-configurable storage mode:
  - **LOCAL**: Record to MinIO/S3 only
  - **MILESTONE**: Use Milestone VMS only
  - **BOTH**: Dual recording to both systems
- Hot-reload configuration (no restart)
- Storage backend abstraction (MinIO/S3/Filesystem)
- Retention policy enforcement
- Storage capacity monitoring

**CONFIGURATION**:
```go
type StorageConfig struct {
    Mode            string  `json:"mode"`  // LOCAL, MILESTONE, BOTH
    LocalEnabled    bool    `json:"local_enabled"`
    MilestoneEnabled bool   `json:"milestone_enabled"`
    RetentionDays   int     `json:"retention_days"`
    StorageBackend  string  `json:"storage_backend"`  // MINIO, S3, FILESYSTEM

    MinIOConfig     *MinIOConfig     `json:"minio_config,omitempty"`
    S3Config        *S3Config        `json:"s3_config,omitempty"`
    FilesystemConfig *FilesystemConfig `json:"filesystem_config,omitempty"`
}

type MinIOConfig struct {
    Endpoint    string `json:"endpoint"`
    Bucket      string `json:"bucket"`
    AccessKey   string `json:"access_key"`
    SecretKey   string `json:"secret_key"`
    UseSSL      bool   `json:"use_ssl"`
}
```

**STORAGE PATH STRUCTURE**:
```
/{bucket}/recordings/{source}/{camera_id}/{year}/{month}/{day}/
  ├── segment_20240101_100000.mp4
  ├── segment_20240101_100500.mp4
  └── segment_20240101_101000.mp4

/{bucket}/clips/{clip_type}/{clip_id}/
  └── clip.mp4
```

**API ENDPOINTS**:
- GET /api/v1/admin/storage/config → Get current config
- PUT /api/v1/admin/storage/config → Update config (hot reload)
- GET /api/v1/admin/storage/stats → Storage usage statistics
- POST /api/v1/storage/recording/start → Start recording session
- POST /api/v1/storage/recording/stop → Stop recording session

**DELIVERABLES**:
1. Go storage orchestration service
2. Storage backend abstraction layer
3. Configuration management API
4. Admin UI component (React)
5. Retention policy scheduler

---

### **2.3: Recording Service** (NEW)

**Purpose**: Continuous video recording with segment-based storage

**FUNCTIONAL REQUIREMENTS**:
- Record from MediaMTX output
- Segment-based recording (5-minute segments)
- H.264 copy only (NO transcoding, NO GPU)
- Automatic upload to storage backend
- Local cleanup after upload
- Resume recording after failure

**RECORDING PROCESS**:
```
MediaMTX → FFmpeg (H.264 copy) → 5min segments → Upload to MinIO → Delete local
```

**FFMPEG COMMAND** (H.264 Copy - No GPU):
```bash
ffmpeg -rtsp_transport tcp \
  -i rtsp://mediamtx:8554/camera_{id} \
  -c:v copy \          # NO RE-ENCODING
  -c:a copy \          # NO RE-ENCODING
  -f segment \
  -segment_time 300 \  # 5 minutes
  -segment_format mp4 \
  -reset_timestamps 1 \
  -strftime 1 \
  /tmp/recordings/camera_{id}/segment_%Y%m%d_%H%M%S.mp4
```

**WORKER POOL**:
- Max 50 concurrent recordings
- Max 1 recording per camera
- Memory limit: 50MB per worker
- CPU limit: 2% per worker (copy only)

**PERFORMANCE**:
- CPU per camera: ~2% (H.264 copy)
- Memory per camera: ~50MB
- Disk write: ~200 KB/s per camera (2 Mbps)
- Upload bandwidth: ~200 KB/s per camera
- **NO GPU REQUIRED**

**DELIVERABLES**:
1. Go recording worker service
2. FFmpeg wrapper with error handling
3. Segment upload manager
4. Integration tests
5. Resource usage benchmarks

---

### **2.4: Metadata Service** (NEW)

**Purpose**: Video metadata, search, tags, and annotations

**DATABASE SCHEMA**:
```sql
-- Video recording sessions
CREATE TABLE video_sessions (
    id UUID PRIMARY KEY,
    camera_id UUID NOT NULL,
    camera_name VARCHAR(200),
    camera_name_ar NVARCHAR(200),
    source VARCHAR(50),  -- DUBAI_POLICE, METRO, BUS, OTHER
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    status VARCHAR(20),  -- RECORDING, STOPPED, ERROR
    storage_mode VARCHAR(20),  -- LOCAL, MILESTONE, BOTH
    total_size_bytes BIGINT DEFAULT 0,
    segment_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_camera_time (camera_id, start_time),
    INDEX idx_source_time (source, start_time)
);

-- Video segments (5-minute chunks)
CREATE TABLE video_segments (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES video_sessions(id) ON DELETE CASCADE,
    sequence_number INT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration_seconds INT NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    storage_backend VARCHAR(50),  -- MINIO, S3, MILESTONE
    codec VARCHAR(20) DEFAULT 'h264',
    resolution VARCHAR(20),  -- 1920x1080
    bitrate_kbps INT,
    has_motion BOOLEAN DEFAULT false,
    objects_detected JSONB,  -- AI detection summary
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (session_id, sequence_number),
    INDEX idx_session_seq (session_id, sequence_number),
    INDEX idx_time_range (start_time, end_time)
);

-- Video clips (extracted incidents/evidence)
CREATE TABLE video_clips (
    id UUID PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    name_ar NVARCHAR(200),
    description TEXT,
    description_ar TEXT,
    camera_id UUID NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration_seconds INT NOT NULL,
    clip_type VARCHAR(50),  -- INCIDENT, EVIDENCE, EXPORT, MANUAL
    tags TEXT[],  -- Searchable tags
    created_by VARCHAR(36),  -- User ID from RTA IAM
    storage_path TEXT NOT NULL,
    size_bytes BIGINT,
    thumbnail_path TEXT,
    metadata JSONB,  -- Custom key-value metadata
    retention_days INT DEFAULT 90,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_camera_time (camera_id, start_time),
    INDEX idx_tags (tags) USING GIN,
    INDEX idx_clip_type (clip_type),
    INDEX idx_created_by (created_by)
);

-- Annotations on clips (markers, notes, regions)
CREATE TABLE video_annotations (
    id UUID PRIMARY KEY,
    clip_id UUID REFERENCES video_clips(id) ON DELETE CASCADE,
    timestamp_offset_ms BIGINT NOT NULL,  -- Milliseconds from clip start
    annotation_type VARCHAR(50),  -- TEXT, MARKER, REGION, REDACTION
    content TEXT,
    content_ar TEXT,
    position JSONB,  -- {x, y, width, height} for regions
    created_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_clip (clip_id)
);

-- Object detection results (from AI service)
CREATE TABLE object_detections (
    id BIGSERIAL PRIMARY KEY,
    segment_id UUID REFERENCES video_segments(id) ON DELETE CASCADE,
    timestamp_offset_ms BIGINT NOT NULL,
    object_class VARCHAR(100),  -- person, vehicle, license_plate
    confidence FLOAT NOT NULL,
    bounding_box JSONB,  -- {x, y, width, height, rotation}
    attributes JSONB,  -- {color, type, direction, speed}
    detected_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_segment_class (segment_id, object_class),
    INDEX idx_class_time (object_class, detected_at),
    INDEX idx_confidence (confidence)
);
```

**SEARCH API**:
```go
type SearchRequest struct {
    CameraID      string    `json:"camera_id,omitempty"`
    Source        string    `json:"source,omitempty"`
    StartTime     time.Time `json:"start_time,omitempty"`
    EndTime       time.Time `json:"end_time,omitempty"`
    Tags          []string  `json:"tags,omitempty"`
    ClipType      string    `json:"clip_type,omitempty"`
    ObjectClasses []string  `json:"object_classes,omitempty"`  // AI search
    HasMotion     *bool     `json:"has_motion,omitempty"`
    CreatedBy     string    `json:"created_by,omitempty"`
    Limit         int       `json:"limit,omitempty"`
    Offset        int       `json:"offset,omitempty"`
}
```

**API ENDPOINTS**:
- GET /api/v1/metadata/sessions → Search recording sessions
- GET /api/v1/metadata/segments/{session_id} → Get segments
- POST /api/v1/metadata/clips → Create clip
- GET /api/v1/metadata/clips → Search clips
- POST /api/v1/metadata/clips/{id}/annotations → Add annotation
- GET /api/v1/metadata/search → Advanced search (AI objects, tags, etc.)

**DELIVERABLES**:
1. Go metadata service
2. PostgreSQL schema with migrations
3. Search API with full-text support
4. Tag management API
5. Performance-optimized indexes

---

### **2.5: Object Detection Service** (NEW - AI)

**Purpose**: Real-time object detection using lightweight visual models

**MODEL SELECTION** (Optimized for edge deployment):
- **Primary**: YOLOv8 Nano (yolov8n.onnx)
  - Size: ~6MB
  - Parameters: 3.2M
  - Speed: 80 FPS on CPU
  - Accuracy: mAP 37.3%
- **Alternative**: MobileNetV3 + SSDLite
  - Size: ~20MB
  - Speed: 60 FPS on CPU

**FUNCTIONAL REQUIREMENTS**:
- Analyze video segments at 1 FPS (sample rate)
- Detect: person, vehicle, license_plate, custom classes
- Confidence threshold: 0.5 (configurable)
- Batch processing: 10 frames per batch
- Save detections to metadata database
- **Optional GPU acceleration** (CUDA/TensorRT)

**DETECTION CLASSES**:
```python
COCO_CLASSES = [
    "person",           # 0
    "bicycle",          # 1
    "car",              # 2
    "motorcycle",       # 3
    "bus",              # 5
    "truck",            # 7
    # ... custom classes can be added via model retraining
]
```

**PROCESSING PIPELINE**:
```
Video Segment → Frame Extraction (1 FPS) → Batch (10 frames) →
YOLO Inference → Post-processing → Save to Metadata DB
```

**IMPLEMENTATION** (Go + ONNX Runtime):
```go
type ObjectDetectionService struct {
    model       *onnxruntime.Session
    sampleRate  int     // Frames per second to analyze (default: 1)
    confidence  float32 // Detection threshold (default: 0.5)
    batchSize   int     // Frames to process together (default: 10)
    classes     []string
    enableGPU   bool
}

func (s *ObjectDetectionService) AnalyzeSegment(ctx context.Context, segment *VideoSegment) error {
    // Extract frames at sample rate
    frames, err := s.extractFrames(segment.Path, s.sampleRate)
    if err != nil {
        return err
    }

    // Process in batches
    detections := []ObjectDetection{}
    for i := 0; i < len(frames); i += s.batchSize {
        batch := frames[i:min(i+s.batchSize, len(frames))]

        // Run inference
        results := s.model.Detect(batch)

        for j, result := range results {
            frameIdx := i + j
            timestamp := frameIdx * 1000 / s.sampleRate  // milliseconds

            for _, obj := range result.Objects {
                if obj.Confidence < s.confidence {
                    continue
                }

                detection := ObjectDetection{
                    SegmentID:       segment.ID,
                    TimestampOffset: timestamp,
                    ObjectClass:     s.classes[obj.ClassID],
                    Confidence:      obj.Confidence,
                    BoundingBox:     obj.BoundingBox,
                }
                detections = append(detections, detection)
            }
        }
    }

    // Save to database
    return s.metadataRepo.SaveDetections(ctx, detections)
}
```

**PERFORMANCE**:
- CPU per stream: ~10% (at 1 FPS analysis)
- Memory: 100MB for model + 50MB per worker
- Latency: ~12ms per frame (CPU), ~3ms per frame (GPU)
- Throughput: 80 FPS per worker (CPU)
- **GPU optional but recommended for >100 cameras**

**DELIVERABLES**:
1. Go object detection service
2. ONNX model integration
3. Frame extraction utility (FFmpeg)
4. Detection result API
5. Model retraining guide (custom classes)
6. GPU acceleration support (CUDA)

---

### **2.6: Playback Service** (UNIFIED)

**Purpose**: Unified playback from local storage or Milestone VMS

**FUNCTIONAL REQUIREMENTS**:
- Detect playback source based on storage config
- Prefer local storage (lower latency: <300ms)
- Fallback to Milestone VMS (latency: <800ms)
- H.264 transmux to HLS (NO transcoding, NO GPU)
- Signed URLs with 5-minute expiration
- Segment caching for reuse

**PLAYBACK FLOW**:
```
Request → Check storage config →
  IF LOCAL available: MinIO → FFmpeg (transmux) → HLS → Client (<300ms)
  ELSE IF MILESTONE: VMS Service → Milestone → FFmpeg (transmux) → HLS → Client (<800ms)
  ELSE: Error "No recording found"
```

**FFMPEG TRANSMUX** (H.264 Copy - No GPU):
```bash
# Local segments concatenation
ffmpeg -f concat -safe 0 -i concat.txt \
  -c:v copy \    # NO RE-ENCODING
  -c:a copy \    # NO RE-ENCODING
  -f hls \
  -hls_time 2 \
  -hls_list_size 0 \
  -hls_segment_filename /cache/{session}/segment_%03d.ts \
  /cache/{session}/playlist.m3u8

# Milestone recording transmux
ffmpeg -i rtsp://milestone:554/recording?id=... \
  -c:v copy \    # NO RE-ENCODING
  -c:a copy \
  -f hls \
  -hls_time 2 \
  /cache/{session}/playlist.m3u8
```

**API ENDPOINTS**:
- POST /api/v1/playback/prepare → Prepare playback session
- GET /api/v1/playback/status/{session_id} → Check status
- GET /api/v1/playback/playlist/{session_id}.m3u8 → Get playlist (signed)
- GET /api/v1/playback/segment/{session_id}/{segment}.ts → Get segment (signed)

**SIGNED URL GENERATION**:
```go
func (s *PlaybackService) GenerateSignedURL(sessionID string, expiresIn time.Duration) string {
    expires := time.Now().Add(expiresIn).Unix()
    message := fmt.Sprintf("%s:%d", sessionID, expires)
    signature := hmac.SHA256([]byte(s.secret), []byte(message))

    return fmt.Sprintf("/playback/playlist/%s.m3u8?expires=%d&sig=%s",
        sessionID, expires, base64.URLEncoding.EncodeToString(signature))
}
```

**CACHING**:
- Cache key: `{camera_id}:{start_time}:{end_time}`
- Max cache: 50GB
- LRU eviction at 80% capacity
- Segment-level caching (reuse across requests)

**PERFORMANCE**:
- CPU per session: ~5% (transmux only)
- Memory per session: ~50MB
- Latency (local): <300ms
- Latency (Milestone): <800ms
- **NO GPU REQUIRED**

**DELIVERABLES**:
1. Go playback service
2. FFmpeg transmux wrapper
3. Signed URL generator
4. Cache management
5. Nginx configuration for serving

---

### **2.7: LiveKit SFU (Low Latency Live Streaming)**

**Purpose**: WebRTC streaming with sub-1-second latency via WHIP ingestion

**ARCHITECTURE OVERVIEW**:

The system uses **WHIP (WebRTC HTTP Ingestion Protocol)** for camera ingestion, providing ultra-low latency (~450ms) vs traditional HLS (2-4 seconds).

**WHIP Ingestion Flow**:
```
Camera (RTSP) → MediaMTX → GStreamer WHIP Pusher Container → LiveKit WHIP Ingress → LiveKit SFU → Viewers
```

**Key Components**:
1. **MediaMTX**: Pulls RTSP streams from Milestone VMS, provides stable RTSP endpoints
2. **WHIP Pusher Container**: Per-camera Docker container running GStreamer pipeline with `whipsink`
3. **LiveKit Ingress**: Receives WHIP streams on port 8080 at endpoint `/w/{stream_key}`
4. **LiveKit SFU**: Distributes WebRTC streams to viewers with simulcast

**WHIP PUSHER IMPLEMENTATION**:

Each camera gets a dedicated WHIP pusher container spawned by go-api via Docker API:

**GStreamer Pipeline** (services/whip-pusher/pusher.sh):
```bash
gst-launch-1.0 -v \
  rtspsrc location="${RTSP_URL}" latency=0 protocols=tcp ! \
  application/x-rtp,media=video ! \
  rtpjitterbuffer latency=100 ! \
  decodebin ! \
  x264enc tune=zerolatency speed-preset=ultrafast bitrate=2000 key-int-max=60 ! \
  h264parse ! \
  rtph264pay config-interval=-1 pt=96 ! \
  application/x-rtp,media=video,encoding-name=H264,payload=96 ! \
  whipsink whip-endpoint="${WHIP_ENDPOINT}" auth-token="${STREAM_KEY}"
```

**Pipeline Features**:
- **Codec Support**: Handles both H.264 (passthrough) and H.265 (transcoded to H.264)
- **Audio Filtering**: Caps filter selects only video stream (ignores audio)
- **Stream Order Handling**: Works regardless of whether audio or video is stream 0
- **Low Latency**: rtpjitterbuffer with 100ms latency, zerolatency x264 tuning
- **Standardization**: All cameras transcoded to H.264 for consistent playback

**WHIP Pusher Docker Image** (services/whip-pusher/Dockerfile):
```dockerfile
FROM ubuntu:22.04

# Install GStreamer + dependencies
RUN apt-get update && apt-get install -y \
    gstreamer1.0-tools gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good gstreamer1.0-plugins-bad \
    gstreamer1.0-plugins-ugly gstreamer1.0-libav \
    gstreamer1.0-nice curl ca-certificates git \
    build-essential pkg-config libssl-dev

# Install Rust + build gst-plugins-rs (for whipsink)
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN git clone https://gitlab.freedesktop.org/gstreamer/gst-plugins-rs.git && \
    cd gst-plugins-rs && \
    cargo build --release --package gst-plugin-webrtchttp && \
    cp target/release/*.so /usr/lib/x86_64-linux-gnu/gstreamer-1.0/

COPY pusher.sh /app/pusher.sh
ENTRYPOINT ["/app/pusher.sh"]
```

**Container Management**:

Go-api spawns/manages WHIP pusher containers via Docker API:

```go
// services/go-api/internal/client/docker_client.go
func (d *DockerClient) StartWHIPPusher(ctx context.Context, config WHIPPusherConfig) error {
    containerConfig := &container.Config{
        Image: "whip-pusher:latest",
        Env: []string{
            fmt.Sprintf("RTSP_URL=%s", config.RTSPURL),
            fmt.Sprintf("WHIP_ENDPOINT=%s", config.WHIPEndpoint),
            fmt.Sprintf("STREAM_KEY=%s", config.StreamKey),
        },
    }
    // Container created and started with restart policy
}
```

**LiveKit Ingress Configuration**:

WHIP ingress created programmatically per camera:

```go
// services/go-api/internal/client/livekit_ingress_client.go
func (c *LiveKitIngressClient) CreateWHIPIngress(ctx context.Context, roomName, participantName string) (*livekit.IngressInfo, error) {
    req := &livekit.CreateIngressRequest{
        InputType:           livekit.IngressInput_WHIP_INPUT,
        Name:                fmt.Sprintf("whip_%s", roomName),
        RoomName:            roomName,
        ParticipantIdentity: participantName,
        ParticipantName:     participantName,
        // EnableTranscoding is false by default (bypass transcoding)
    }

    ingressInfo, err := ingressClient.CreateIngress(ctx, req)

    // Construct WHIP URL manually (LiveKit doesn't populate it)
    whipURL := fmt.Sprintf("http://livekit-ingress:8080/w/%s", ingressInfo.StreamKey)
    ingressInfo.Url = whipURL

    return ingressInfo, nil
}
```

**LIVEKIT SFU CONFIGURATION** (Optimized for low latency):
```yaml
# livekit.yaml
port: 7880

rtc:
  tcp_port: 7881
  port_range_start: 50000
  port_range_end: 50500  # 500 ports for 500 cameras
  use_external_ip: true
  congestion_control: true
  allow_tcp_fallback: true

redis:
  address: valkey:6379  # VALKEY ONLY
  password: ${VALKEY_PASSWORD}

room:
  auto_create: false
  empty_timeout: 60  # Aggressive cleanup
  max_participants: 100
  enable_simulcast: true

video:
  enable_dynacast: true  # Dynamic quality switching

# SIMULCAST LAYERS (bandwidth adaptation)
simulcast:
  layers:
    - quality: HIGH
      width: 1920
      height: 1080
      bitrate: 3000000  # 3 Mbps
    - quality: MEDIUM
      width: 1280
      height: 720
      bitrate: 1500000  # 1.5 Mbps
    - quality: LOW
      width: 640
      height: 360
      bitrate: 500000   # 500 Kbps

turn:
  enabled: true
  domain: turn.rta.ae
  port: 3478
  tls_port: 5349

webhook:
  urls:
    - http://go-api:8080/webhook/livekit

logging:
  level: warn  # Reduce overhead
```

**DOCKER COMPOSE INTEGRATION**:

```yaml
# docker-compose.yml
services:
  livekit-ingress:
    image: livekit/ingress:latest
    networks: [cctv-network]
    ports:
      - "8080:8080"  # WHIP endpoint
    environment:
      - LIVEKIT_URL=ws://livekit:7880
      - LIVEKIT_API_KEY=${LIVEKIT_API_KEY}
      - LIVEKIT_API_SECRET=${LIVEKIT_API_SECRET}
    depends_on: [livekit, mediamtx]

  go-api:
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # Docker API for WHIP pusher containers
```

**STREAM RESERVATION FLOW**:

1. User requests camera stream via go-api
2. go-api creates LiveKit WHIP ingress (gets stream key)
3. go-api constructs WHIP endpoint URL: `http://livekit-ingress:8080/w/{stream_key}`
4. go-api spawns WHIP pusher container with:
   - RTSP_URL: `rtsp://mediamtx:8554/camera_{id}`
   - WHIP_ENDPOINT: `http://livekit-ingress:8080/w/{stream_key}`
   - STREAM_KEY: `{stream_key}`
5. WHIP pusher connects to MediaMTX RTSP, pushes to LiveKit WHIP ingress
6. LiveKit ingress publishes to room
7. go-api returns LiveKit room token to user
8. User connects to LiveKit room via WebRTC

**PERFORMANCE**:
- **Latency**: ~450ms glass-to-glass (WHIP) vs 2-4s (HLS)
- **CPU per WHIP pusher**: ~15% (with transcoding), ~5% (passthrough)
- **Memory per WHIP pusher**: ~50MB
- **CPU per 100 rooms**: <50%
- **Memory per room**: ~10MB
- **Max concurrent streams**: 500 cameras
- **Bitrate per stream**: 0.5-2 Mbps (depends on codec and resolution)

**CODEC HANDLING**:
- **H.264 cameras**: Decoded and re-encoded with x264enc (for consistency)
- **H.265 cameras**: Decoded with decodebin, transcoded to H.264 with x264enc
- **Output**: Standardized H.264 for all cameras

**FAULT ISOLATION**:
- **Per-camera containers**: Each camera has dedicated WHIP pusher container
- **Automatic restart**: Docker restart policy ensures resilience
- **Independent failures**: One camera failure doesn't affect others
- **Resource limits**: CPU/memory limits per container prevent resource exhaustion

**DELIVERABLES**:
1. whip-pusher Docker image (custom build with gst-plugins-rs)
2. pusher.sh GStreamer pipeline script
3. livekit.yaml configuration
4. LiveKit Ingress setup
5. Docker API client in go-api for container management
6. Room management Go service
7. coturn (TURN server) setup
8. Load test: 500 WHIP pushers × 10 viewers per room

---

## **PHASE 3: APPLICATION LAYER**

### **3.1: Go API Backend**

**Purpose**: Central API orchestrating all backend services

**ARCHITECTURE** (Clean Architecture):
```
go-api/
├── cmd/api/main.go
├── internal/
│   ├── domain/          # Business entities
│   │   ├── camera.go
│   │   ├── stream.go
│   │   ├── clip.go
│   │   └── user.go
│   ├── usecase/         # Business logic
│   │   ├── stream_usecase.go
│   │   ├── playback_usecase.go
│   │   ├── clip_usecase.go
│   │   └── auth_usecase.go
│   ├── repository/      # Data access
│   │   ├── postgres/
│   │   ├── valkey/
│   │   └── interfaces.go
│   ├── delivery/        # HTTP/WebSocket
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── router.go
│   │   └── websocket/
│   └── infrastructure/  # External integrations
│       ├── vms_client.go
│       ├── storage_client.go
│       ├── livekit_client.go
│       └── metadata_client.go
└── pkg/                 # Shared utilities
    ├── errors/
    ├── logger/
    └── validator/
```

**KEY USE CASES**:

**StreamUseCase.RequestStream()**:
```go
func (u *StreamUseCase) RequestStream(ctx context.Context, cameraID string, userID string) (*StreamToken, error) {
    // 1. Get camera from VMS
    camera, err := u.vmsClient.GetCamera(ctx, cameraID)
    if err != nil {
        return nil, errors.Wrap(err, "camera not found")
    }

    // 2. Check user permissions (IAM)
    if !u.authUseCase.HasPermission(ctx, userID, camera.Source) {
        return nil, errors.Forbidden("no access to this camera source")
    }

    // 3. Check agency limit (Valkey Lua script)
    canReserve, current, limit := u.valkeyRepo.CheckAndReserve(ctx, camera.Source, userID, cameraID)
    if !canReserve {
        return nil, &errors.AgencyLimitError{
            Source:  camera.Source,
            Current: current,
            Limit:   limit,
        }
    }

    // 4. Create LiveKit room if needed
    roomName := fmt.Sprintf("camera_%s", cameraID)
    room, err := u.livekitClient.EnsureRoom(ctx, roomName)
    if err != nil {
        u.valkeyRepo.Release(ctx, camera.Source, userID)
        return nil, errors.Wrap(err, "failed to create room")
    }

    // 5. Generate LiveKit access token
    token, err := u.livekitClient.GenerateToken(ctx, roomName, userID, 3600)
    if err != nil {
        u.valkeyRepo.Release(ctx, camera.Source, userID)
        return nil, errors.Wrap(err, "failed to generate token")
    }

    // 6. Save stream reservation
    reservation := &domain.StreamReservation{
        ID:        uuid.New().String(),
        CameraID:  cameraID,
        UserID:    userID,
        Source:    camera.Source,
        RoomName:  roomName,
        ExpiresAt: time.Now().Add(1 * time.Hour),
    }
    if err := u.streamRepo.SaveReservation(ctx, reservation); err != nil {
        return nil, err
    }

    // 7. Audit log
    u.logger.Audit("stream_requested", map[string]interface{}{
        "user_id": userID,
        "camera_id": cameraID,
        "source": camera.Source,
    })

    return &StreamToken{
        ReservationID: reservation.ID,
        Token:         token,
        RoomName:      roomName,
        ExpiresAt:     reservation.ExpiresAt,
    }, nil
}
```

**API ENDPOINTS**:
```
# Live Streaming
POST   /api/v1/stream/reserve        # Reserve camera stream
DELETE /api/v1/stream/release/{id}   # Release stream
POST   /api/v1/stream/heartbeat/{id} # Keep-alive
GET    /api/v1/stream/stats          # Real-time statistics
WS     /ws/stream/stats               # WebSocket updates

# Cameras
GET    /api/v1/cameras                # List cameras
GET    /api/v1/cameras/{id}           # Get camera details
POST   /api/v1/cameras/{id}/ptz       # PTZ control

# Playback
POST   /api/v1/playback/prepare       # Prepare playback session
GET    /api/v1/playback/status/{id}   # Check status

# Clips
POST   /api/v1/clips                  # Create clip (extract)
GET    /api/v1/clips                  # Search clips
POST   /api/v1/clips/{id}/annotations # Add annotation
GET    /api/v1/clips/{id}/download    # Download clip

# Storage Admin
GET    /api/v1/admin/storage/config   # Get config
PUT    /api/v1/admin/storage/config   # Update config
GET    /api/v1/admin/storage/stats    # Storage statistics

# IAM
POST   /api/v1/iam/users              # Create user
PUT    /api/v1/iam/users/{id}         # Update user
POST   /api/v1/iam/users/{id}/groups  # Assign group

# Health & Metrics
GET    /health                        # Health check
GET    /metrics                       # Prometheus metrics
```

**WEBSOCKET HUB** (Real-time updates):
```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
}

// Message types:
// - AGENCY_LIMIT_UPDATE: {source, current, limit}
// - CAMERA_STATUS: {camera_id, status}
// - ALERT: {type, message, severity}
// - STREAM_STATS: {active_streams, per_source_count}
```

**DELIVERABLES**:
1. Complete Go API service
2. OpenAPI specification
3. Postman collection
4. Unit tests (>80% coverage)
5. Integration tests
6. Dockerfile (<50MB)

---

### **3.2: React Dashboard**

**Purpose**: Operator dashboard for viewing 64 simultaneous streams

**FEATURES**:
- Grid layouts: 2×2, 3×3, 4×4, 16-way hotspot, 64-way hotspot
- Drag-and-drop camera placement
- LiveKit WebRTC integration
- HLS playback for recordings
- PTZ controls with joystick
- Clip creation and management
- Search (time range, tags, AI objects)
- Bilingual UI (Arabic/English with RTL)
- Real-time agency quota display

**GRID LAYOUTS**:
```typescript
const GRID_LAYOUTS = {
  "2x2": { rows: 2, cols: 2, cells: 4 },
  "3x3": { rows: 3, cols: 3, cells: 9 },
  "4x4": { rows: 4, cols: 4, cells: 16 },
  "16-hotspot": {
    rows: 4, cols: 4,
    hotspot: { row: 0, col: 0, rowSpan: 3, colSpan: 3 },
    cells: 7 + 1  // 3×3 main + 7 small
  },
  "64-hotspot": {
    rows: 8, cols: 8,
    hotspot: { row: 0, col: 0, rowSpan: 7, colSpan: 7 },
    cells: 15 + 1  // 7×7 main + 15 small
  }
};
```

**PERFORMANCE OPTIMIZATIONS**:
1. **Viewport-based rendering** (Intersection Observer)
   - Only render visible cells
   - Pause off-screen streams
2. **Video element pooling** (reuse <video> elements)
3. **Web Workers** for grid calculations
4. **React.memo** for expensive components
5. **Lazy loading** with Suspense

**STATE MANAGEMENT** (Zustand):
```typescript
interface CameraStore {
  cameras: Camera[];
  activeCameras: Map<number, StreamState>;
  agencyLimits: Map<Source, LimitStatus>;
  layout: GridLayout;

  addCamera: (gridIndex: number, cameraId: string) => Promise<void>;
  removeCamera: (gridIndex: number) => void;
  changeLayout: (layout: string) => void;
  checkAgencyLimit: (source: Source) => boolean;
}
```

**LIVEKIT INTEGRATION**:
```typescript
// Custom hook for LiveKit
function useLiveKit(cameraId: string, quality: Quality) {
  const [room, setRoom] = useState<Room | null>(null);
  const videoRef = useRef<HTMLVideoElement>(null);

  // Only connect when in viewport
  const isVisible = useIntersectionObserver(videoRef);

  useEffect(() => {
    if (!isVisible) {
      room?.disconnect();
      return;
    }

    const connectRoom = async () => {
      const token = await api.reserveStream(cameraId);
      const r = new Room({
        adaptiveStream: true,
        dynacast: true,
        videoCaptureDefaults: {
          resolution: quality === 'HIGH' ? VideoPresets.h1080 : VideoPresets.h720
        }
      });

      await r.connect(LIVEKIT_URL, token);
      setRoom(r);
    };

    connectRoom();
  }, [isVisible, cameraId]);

  return { room, videoRef };
}
```

**CLIP MANAGEMENT UI**:
```typescript
// Clip extraction interface
function ClipCreator({ camera, startTime, endTime }: ClipCreatorProps) {
  const [name, setName] = useState('');
  const [nameAr, setNameAr] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [clipType, setClipType] = useState<ClipType>('INCIDENT');

  const handleCreate = async () => {
    const clip = await api.createClip({
      camera_id: camera.id,
      start_time: startTime,
      end_time: endTime,
      name,
      name_ar: nameAr,
      tags,
      clip_type: clipType
    });

    toast.success(`Clip created: ${clip.id}`);
  };

  return (
    <Dialog>
      <Input label="Clip Name (EN)" value={name} onChange={setName} />
      <Input label="Clip Name (AR)" value={nameAr} onChange={setNameAr} dir="rtl" />
      <TagInput tags={tags} onChange={setTags} />
      <Select value={clipType} onChange={setClipType}>
        <option value="INCIDENT">Incident</option>
        <option value="EVIDENCE">Evidence</option>
        <option value="EXPORT">Export</option>
      </Select>
      <Button onClick={handleCreate}>Create Clip</Button>
    </Dialog>
  );
}
```

**SEARCH INTERFACE**:
```typescript
// Advanced search with AI object detection
function VideoSearch() {
  const [filters, setFilters] = useState<SearchFilters>({
    cameras: [],
    sources: [],
    startTime: null,
    endTime: null,
    tags: [],
    clipType: null,
    objectClasses: [],  // AI detection: person, vehicle, etc.
    hasMotion: null
  });

  const { data: results, isLoading } = useQuery(
    ['search', filters],
    () => api.searchVideos(filters)
  );

  return (
    <div>
      <FilterPanel filters={filters} onChange={setFilters} />
      <ResultsGrid results={results} isLoading={isLoading} />
    </div>
  );
}
```

**ARABIC RTL SUPPORT**:
```typescript
// i18n configuration
i18n.use(initReactI18next).init({
  resources: { ar, en },
  lng: localStorage.getItem('language') || 'ar',
  fallbackLng: 'en',
  interpolation: { escapeValue: false }
});

// Apply RTL to document
document.dir = i18n.language === 'ar' ? 'rtl' : 'ltr';

// Component with RTL awareness
<div className={cn(styles.container, {
  [styles.rtl]: i18n.language === 'ar'
})}>
```

**DELIVERABLES**:
1. Complete React application
2. Component library (Storybook)
3. E2E tests (Playwright)
4. Performance benchmarks
5. Accessibility audit (WCAG 2.1 AA)
6. Docker build (<500KB gzipped bundle)

---

### **3.3: RTA IAM Integration**

**Purpose**: Integration with RTA Identity and Access Management system

**USER DATA MODEL**:
```go
type RTAUser struct {
    UserID          string     `json:"user_id"`
    LoginID         string     `json:"login_id"`
    EmployeeNumber  string     `json:"employee_number"`
    DisplayName     string     `json:"display_name"`
    Email           string     `json:"email"`
    FullNameArabic  string     `json:"full_name_arabic"`
    Agency          Agency     `json:"agency"`
    Department      Department `json:"department"`
    Section         Section    `json:"section"`
    Status          string     `json:"status"`  // ACTIVE, DISABLED
    Roles           []string   `json:"roles"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
}

type Agency struct {
    Code        string `json:"code"`
    Name        string `json:"name"`
    NameArabic  string `json:"name_arabic"`
}
```

**API ENDPOINTS**:
```
POST   /api/v1/iam/users              # Create user
GET    /api/v1/iam/users/{id}         # Get user
PUT    /api/v1/iam/users/{id}         # Update user
PATCH  /api/v1/iam/users/{id}/status  # Enable/Disable
GET    /api/v1/iam/groups             # List groups
POST   /api/v1/iam/users/{id}/groups  # Assign group
DELETE /api/v1/iam/users/{id}/groups/{groupId}  # Remove from group
```

**JWT TOKEN STRUCTURE**:
```json
{
  "iss": "rta-iam",
  "sub": "user123",
  "exp": 1735689600,
  "iat": 1735686000,
  "login_id": "ahmed.ali",
  "employee_number": "RTA-12345",
  "display_name": "Ahmed Ali",
  "email": "ahmed.ali@rta.ae",
  "agency": "DUBAI_POLICE",
  "department": "TRAFFIC",
  "permissions": [
    "camera:DUBAI_POLICE:view",
    "camera:DUBAI_POLICE:ptz",
    "clip:DUBAI_POLICE:create"
  ],
  "camera_sources": ["DUBAI_POLICE"]
}
```

**WEBHOOK HANDLER** (IAM events):
```go
func (h *WebhookHandler) HandleIAMEvent(w http.ResponseWriter, r *http.Request) {
    // Verify signature
    signature := r.Header.Get("X-IAM-Signature")
    if !h.verifySignature(r.Body, signature) {
        respondError(w, errors.Unauthorized("invalid signature"))
        return
    }

    var event IAMEvent
    json.NewDecoder(r.Body).Decode(&event)

    switch event.Type {
    case "USER_DISABLED":
        h.handleUserDisabled(event.Data["user_id"])
    case "GROUP_CHANGED":
        h.handleGroupChanged(event.Data["user_id"])
    }
}

func (h *WebhookHandler) handleUserDisabled(userID string) {
    // Revoke all active streams
    streams := h.streamRepo.GetUserStreams(userID)
    for _, stream := range streams {
        h.streamUseCase.ReleaseStream(context.Background(), stream.ID)
    }

    // Revoke JWT sessions
    h.sessionService.RevokeUserSessions(userID)
}
```

**AUDIT LOGGING**:
```sql
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    user_id VARCHAR(36),
    actor_id VARCHAR(36),
    timestamp TIMESTAMP NOT NULL,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    INDEX idx_user_timestamp (user_id, timestamp),
    INDEX idx_event_type (event_type)
);
```

**DELIVERABLES**:
1. IAM service implementation
2. JWT validation middleware
3. Webhook handler
4. Audit logging system
5. Group permission mapper

---

## **PHASE 4: DEPLOYMENT & MONITORING**

### **4.1: Docker Compose Stack**

**COMPLETE SERVICES**:
```yaml
version: '3.9'

networks:
  frontend:
  backend:
  storage:
  monitoring:

volumes:
  minio_data:      # Video storage (~500TB)
  postgres_data:   # Metadata (~100GB)
  valkey_data:     # Cache (~10GB)
  recording_temp:  # Temp recording (~100GB)
  playback_cache:  # HLS cache (~50GB)

services:
  # === INGEST LAYER ===
  mediamtx:
    image: bluenviron/mediamtx:latest
    networks: [backend]
    volumes:
      - ./config/mediamtx.yml:/mediamtx.yml:ro
    deploy:
      resources:
        limits: { cpus: '1', memory: 512M }
        reservations: { cpus: '0.5', memory: 256M }

  # === VMS LAYER ===
  vms-service:
    image: rta/vms-service:latest
    networks: [backend]
    environment:
      - MILESTONE_SERVER=${MILESTONE_SERVER}
    deploy:
      resources:
        limits: { cpus: '0.5', memory: 256M }

  # === STORAGE LAYER ===
  storage-service:
    image: rta/storage-service:latest
    networks: [backend, storage]
    environment:
      - STORAGE_MODE=${STORAGE_MODE:-BOTH}
      - MINIO_ENDPOINT=minio:9000
    deploy:
      resources:
        limits: { cpus: '0.5', memory: 256M }

  recording-service:
    image: rta/recording-service:latest
    networks: [backend, storage]
    volumes:
      - recording_temp:/tmp/recordings
    environment:
      - STORAGE_BACKEND=MINIO
      - MINIO_ENDPOINT=minio:9000
    deploy:
      resources:
        limits: { cpus: '1', memory: 512M }

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    networks: [storage]
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data
    environment:
      - MINIO_ROOT_USER=admin
      - MINIO_ROOT_PASSWORD=${MINIO_PASSWORD}
    deploy:
      resources:
        limits: { cpus: '1', memory: 1G }

  # === METADATA LAYER ===
  metadata-service:
    image: rta/metadata-service:latest
    networks: [backend]
    environment:
      - DATABASE_URL=postgres://cctv:${POSTGRES_PASSWORD}@postgres:5432/metadata
    deploy:
      resources:
        limits: { cpus: '0.5', memory: 256M }

  postgres:
    image: postgres:15-alpine
    networks: [backend]
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=cctv
      - POSTGRES_USER=cctv
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    deploy:
      resources:
        limits: { cpus: '1', memory: 2G }

  # === AI LAYER ===
  object-detection:
    image: rta/object-detection:latest
    networks: [backend]
    volumes:
      - ./models:/models:ro
    environment:
      - MODEL_PATH=/models/yolov8n.onnx
      - CONFIDENCE_THRESHOLD=0.5
    deploy:
      resources:
        limits: { cpus: '1', memory: 512M }
    # Optional GPU:
    # runtime: nvidia

  # === PLAYBACK LAYER ===
  playback-service:
    image: rta/playback-service:latest
    networks: [backend, storage]
    volumes:
      - playback_cache:/cache
    environment:
      - STORAGE_SERVICE_URL=http://storage-service:8080
      - VMS_SERVICE_URL=http://vms-service:8080
    deploy:
      resources:
        limits: { cpus: '1', memory: 512M }

  # === STREAMING LAYER ===
  livekit:
    image: livekit/livekit-server:latest
    networks: [backend]
    ports:
      - "7880:7880"
      - "7881:7881"
      - "50000-50500:50000-50500/udp"
    volumes:
      - ./config/livekit.yaml:/config/livekit.yaml:ro
    environment:
      - LIVEKIT_CONFIG=/config/livekit.yaml
    deploy:
      resources:
        limits: { cpus: '2', memory: 2G }

  livekit-ingress:
    image: livekit/ingress:latest
    networks: [backend]
    volumes:
      - ./config/ingress.yaml:/config/ingress.yaml:ro
    depends_on: [livekit, mediamtx]

  coturn:
    image: coturn/coturn:4.6
    network_mode: host
    volumes:
      - ./config/turnserver.conf:/etc/coturn/turnserver.conf:ro

  # === CACHE LAYER ===
  valkey:
    image: valkey/valkey:7.2-alpine
    networks: [backend]
    command: valkey-server /etc/valkey/valkey.conf --maxmemory 1gb
    volumes:
      - valkey_data:/data
      - ./config/valkey.conf:/etc/valkey/valkey.conf:ro
    deploy:
      resources:
        limits: { cpus: '0.5', memory: 1G }

  # === API LAYER ===
  kong:
    image: kong:3.4-alpine
    networks: [frontend, backend]
    ports:
      - "8000:8000"
      - "8443:8443"
    environment:
      - KONG_DATABASE=off
      - KONG_DECLARATIVE_CONFIG=/kong.yaml
    volumes:
      - ./config/kong.yaml:/kong.yaml:ro
    deploy:
      resources:
        limits: { cpus: '1', memory: 512M }

  go-api:
    image: rta/go-api:latest
    networks: [backend]
    environment:
      - DATABASE_URL=postgres://cctv:${POSTGRES_PASSWORD}@postgres:5432/cctv
      - VALKEY_ADDR=valkey:6379
      - VMS_SERVICE_URL=http://vms-service:8080
      - STORAGE_SERVICE_URL=http://storage-service:8080
      - PLAYBACK_SERVICE_URL=http://playback-service:8080
      - LIVEKIT_URL=http://livekit:7880
    deploy:
      resources:
        limits: { cpus: '1', memory: 512M }

  # === FRONTEND ===
  web-dashboard:
    image: rta/web-dashboard:latest
    networks: [frontend]
    environment:
      - REACT_APP_API_URL=https://api.cctv.rta.ae
      - REACT_APP_WS_URL=wss://api.cctv.rta.ae/ws

  nginx:
    image: nginx:alpine
    networks: [frontend]
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
      - playback_cache:/var/cache/segments:ro

  # === MONITORING ===
  prometheus:
    image: prom/prometheus:v2.45.0
    networks: [monitoring, backend]
    ports: ["9090:9090"]
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    deploy:
      resources:
        limits: { cpus: '1', memory: 1G }

  grafana:
    image: grafana/grafana:10.0.0
    networks: [monitoring]
    ports: ["3000:3000"]
    volumes:
      - grafana_data:/var/lib/grafana
      - ./dashboards:/etc/grafana/provisioning/dashboards:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}

  loki:
    image: grafana/loki:2.9.0
    networks: [monitoring, backend]
    ports: ["3100:3100"]
    volumes:
      - ./config/loki.yml:/etc/loki/config.yml:ro
      - loki_data:/loki

  promtail:
    image: grafana/promtail:2.9.0
    networks: [monitoring]
    volumes:
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - ./config/promtail.yml:/etc/promtail/config.yml:ro
```

**RESOURCE SUMMARY**:
| Component | CPU | Memory | Storage |
|-----------|-----|--------|---------|
| MediaMTX | 1 core | 512MB | - |
| VMS Service | 0.5 core | 256MB | - |
| Storage Service | 0.5 core | 256MB | - |
| Recording Service | 1 core | 512MB | - |
| Metadata Service | 0.5 core | 256MB | - |
| Object Detection | 1 core | 512MB | - |
| Playback Service | 1 core | 512MB | - |
| LiveKit | 2 cores | 2GB | - |
| Valkey | 0.5 core | 1GB | 10GB |
| PostgreSQL | 1 core | 2GB | 100GB |
| MinIO | 1 core | 1GB | 500TB+ |
| Go API | 1 core | 512MB | - |
| Kong | 1 core | 512MB | - |
| Monitoring | 1 core | 1GB | 50GB |
| **TOTAL** | **~13 cores** | **~11GB** | **~500TB** |

---

### **4.2: Monitoring & Observability**

**PROMETHEUS SCRAPE TARGETS**:
```yaml
scrape_configs:
  - job_name: 'mediamtx'
    static_configs:
      - targets: ['mediamtx:9090']

  - job_name: 'livekit'
    static_configs:
      - targets: ['livekit:7782']

  - job_name: 'go-services'
    static_configs:
      - targets:
        - 'go-api:8080'
        - 'vms-service:8080'
        - 'storage-service:8080'
        - 'recording-service:8080'
        - 'metadata-service:8080'
        - 'playback-service:8080'
        - 'object-detection:8080'
```

**KEY METRICS**:
```
# Stream Management
cctv_streams_active{source}
cctv_streams_limit{source}
cctv_stream_reservations_total{source,status}

# Storage
cctv_storage_bytes_used{backend}
cctv_recording_segments_total{camera_id}
cctv_storage_upload_duration_seconds

# Playback
cctv_playback_sessions_active
cctv_playback_cache_hit_ratio

# AI Detection
cctv_detections_total{object_class}
cctv_detection_latency_ms

# API
cctv_api_requests_total{method,endpoint,status}
cctv_api_request_duration_seconds
```

**GRAFANA DASHBOARDS**:
1. **Stream Overview**: Active streams per source, quota usage
2. **Storage Metrics**: Usage, upload rate, cache hits
3. **Playback Performance**: Latency, cache ratio, active sessions
4. **AI Analytics**: Detections per class, confidence distribution
5. **API Performance**: Request rate, latency percentiles, errors

**ALERTING RULES**:
```yaml
groups:
  - name: cctv_alerts
    rules:
      - alert: AgencyLimitReached
        expr: cctv_streams_active / cctv_streams_limit > 0.9
        for: 5m
        annotations:
          summary: "Agency {{ $labels.source }} at 90% capacity"

      - alert: StorageAlmostFull
        expr: cctv_storage_bytes_used / cctv_storage_bytes_total > 0.85
        for: 10m
        annotations:
          summary: "Storage at 85% capacity"

      - alert: HighPlaybackLatency
        expr: histogram_quantile(0.99, cctv_playback_latency_seconds) > 1.0
        for: 5m
        annotations:
          summary: "Playback latency p99 > 1 second"
```

**DELIVERABLES**:
1. Complete Docker Compose stack
2. Prometheus configuration
3. Grafana dashboards (JSON)
4. Alerting rules
5. Loki configuration for logs
6. Health check scripts
7. Backup/restore procedures

---

## **PERFORMANCE TARGETS**

| Metric | Target | How Achieved |
|--------|--------|--------------|
| **Live streaming latency (WHIP)** | ~450ms | GStreamer WHIP pusher → LiveKit WHIP Ingress → SFU |
| **Live streaming latency (WebRTC)** | <800ms | LiveKit SFU with dynacast, TCP transport |
| **Playback latency (local)** | <300ms | H.264 transmux (no GPU), segment caching |
| **Playback latency (Milestone)** | <800ms | Optimized RTSP fetch, transmux only |
| **Stream startup time** | <500ms | On-demand activation, connection pooling |
| **API response (p99)** | <100ms | Valkey Lua scripts, in-memory caching |
| **Object detection latency** | <20ms/frame | YOLOv8 Nano (6MB model), 1 FPS sampling |
| **Memory per stream** | <10MB | Video element pooling, viewport culling |
| **CPU per WHIP pusher** | ~15% | GStreamer with x264enc transcoding |
| **Memory per WHIP pusher** | ~50MB | GStreamer pipeline buffer optimization |
| **CPU per camera (recording)** | ~2% | H.264 copy (no transcoding) |
| **Storage per camera/day** | ~21.6GB | H.264 @ 2 Mbps, motion-based recording |
| **Total system footprint** | <11GB RAM | Aggressive optimization, minimal logging |

---

## **STORAGE CAPACITY PLANNING**

### **Scenario: 500 Cameras, 90-day Retention**

**Continuous Recording**:
```
Per camera: 2 Mbps × 86400 seconds / 8 = 21.6 GB/day
500 cameras: 500 × 21.6 GB = 10.8 TB/day
90 days: 10.8 TB × 90 = 972 TB ≈ 1 PB
```

**Motion-Based Recording** (30% activity):
```
90 days: 1 PB × 0.3 = ~300 TB
```

**Recommended Storage**:
- **Start**: 500TB NAS with motion detection
- **Growth**: Plan for 1PB within 2 years
- **Backend**: MinIO distributed mode (4+ nodes)

---

## **IMPLEMENTATION PHASES**

### **Phase 1: Foundation** (Weeks 1-2)
- MediaMTX RTSP ingest
- Valkey stream counter with Lua scripts
- Kong API gateway
- VMS service (Milestone integration)
- Basic Go API structure

### **Phase 2: Storage & Recording** (Weeks 3-4)
- Video Storage Service (configurable)
- Recording Service (segment-based)
- MinIO setup
- Metadata Service (PostgreSQL)
- Basic playback (transmux only)

### **Phase 3: Streaming & Playback** (Weeks 5-6)
- LiveKit SFU configuration
- LiveKit Ingress setup
- Unified Playback Service
- Clip extraction & management
- Search API

### **Phase 4: AI & Frontend** (Weeks 7-8)
- Object Detection Service (YOLOv8)
- React Dashboard (grid layouts)
- Clip management UI
- Search interface
- Arabic RTL support

### **Phase 5: IAM & Deployment** (Weeks 9-10)
- RTA IAM integration
- JWT authentication
- Audit logging
- Docker Compose stack
- Monitoring setup

### **Phase 6: Testing & Optimization** (Weeks 11-12)
- Load testing (500 cameras, 1000 viewers)
- Performance tuning
- Storage optimization
- Documentation
- Production deployment

---

## **TECHNOLOGY STACK**

| Layer | Technology | Justification |
|-------|------------|---------------|
| **Ingest** | MediaMTX | Lightweight, RTSP-native, stable endpoints |
| **WHIP Pusher** | GStreamer + gst-plugins-rs | whipsink for WHIP ingestion, codec transcoding |
| **WHIP Ingress** | LiveKit Ingress | WHIP protocol support, WebRTC bridge |
| **Cache** | Valkey 7.2 | Redis-compatible, Lua scripts, atomic ops |
| **VMS** | Milestone XProtect | External system (customer requirement) |
| **Storage** | MinIO | S3-compatible, self-hosted, scalable |
| **Database** | PostgreSQL 15 | JSONB, GIS support, reliability |
| **Streaming** | LiveKit SFU | WebRTC SFU, ultra-low latency, simulcast |
| **Container Mgmt** | Docker API (Go SDK) | Dynamic WHIP pusher container spawning |
| **AI** | YOLOv8 Nano (ONNX) | 6MB model, 80 FPS CPU, personalized |
| **API** | Go 1.23 | Performance, concurrency, low footprint |
| **Frontend** | React 18 + TypeScript | Performance, ecosystem, RTL support |
| **Gateway** | Kong CE | Valkey plugins, rate limiting, JWT |
| **Video** | FFmpeg | H.264 transmux, no GPU needed |
| **Monitoring** | Prometheus + Grafana + Loki | Industry standard, powerful |

---

## **SECURITY CONSIDERATIONS**

1. **Network Segmentation**: Separate networks for frontend/backend/storage/monitoring
2. **JWT Authentication**: RS256 with RTA IAM public key validation
3. **Signed URLs**: HMAC-SHA256 for playback URLs (5min expiration)
4. **TLS Everywhere**: Kong handles TLS termination
5. **No Hardcoded Secrets**: All credentials from environment/vault
6. **Audit Logging**: All actions logged to PostgreSQL (JSONB)
7. **IP Whitelisting**: /metrics endpoint restricted to monitoring network
8. **CORS**: Strict origin validation (*.rta.ae only)

---

## **COMPLIANCE & AUDIT**

1. **Audit Trail**: All user actions logged with timestamp, actor, details
2. **Evidence Chain**: Video clips with metadata, annotations, signatures
3. **Retention Policies**: Configurable per camera source
4. **Access Control**: Permission-based (IAM groups)
5. **Data Sovereignty**: All data stored in UAE (MinIO on-premises)
6. **Encryption**: At rest (MinIO SSE), in transit (TLS 1.3)

---

## **NEXT STEPS**

See `RTA-CCTV-Implementation-Plan.md` for detailed implementation roadmap.

**Priority Order**:
1. VMS Service (Milestone integration)
2. Valkey Counter Service (quota management)
3. Storage Service (configurable backends)
4. Recording Service (continuous recording)
5. Metadata Service (search & tags)
6. Playback Service (unified playback)
7. Object Detection (AI analytics)
8. Go API (orchestration)
9. React Dashboard (operator UI)
10. Deployment & Monitoring

---

**Document Version**: 2.0
**Last Updated**: 2025-01-XX
**Status**: APPROVED FOR IMPLEMENTATION
