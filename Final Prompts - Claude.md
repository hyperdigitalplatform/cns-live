## **Phase 1: Core Infrastructure Setup**

### **Prompt 1.1: MediaMTX RTSP Ingest Service (Enhanced)**

Create a production-ready RTSP camera ingestion service using MediaMTX for RTA CCTV system.

FUNCTIONAL REQUIREMENTS:
- Handle up to 500 concurrent RTSP camera streams from Milestone Recording Servers
- Track stream count per source in Valkey with atomic operations
- Implement exponential backoff reconnection (initial: 1s, max: 30s, multiplier: 2)
- Support H.264 and H.265 codec passthrough
- Multicast support for local distribution

ERROR HANDLING:
- Connection timeout: 10 seconds per camera
- Retry failed connections: max 5 attempts with exponential backoff
- Circuit breaker: open after 3 consecutive failures, half-open after 30s
- Alert on: >10% cameras failing, any source at 0 streams for >60s
- Fallback: return cached last frame if stream temporarily unavailable
- Graceful shutdown with connection draining

SECURITY:
- Validate RTSP URLs against regex: ^rtsp://[a-zA-Z0-9\.\-]+:[0-9]+/[a-zA-Z0-9/\-_]+$
- Use service account credentials from environment (no hardcoding)
- TLS 1.3 for API endpoints
- No direct external access (behind Kong only)
- Sanitize logs to prevent credential leakage

DATA STRUCTURES:
StreamState {
  CameraID string (UUID)
  Source string (enum: DUBAI_POLICE|METRO|BUS|OTHER)
  Status enum (CONNECTING|ACTIVE|FAILED|CIRCUIT_OPEN)
  LastFrame []byte (max 1MB)
  ConnectTime timestamp
  LastError string (max 500 chars)
  RetryCount int32
  Bitrate int64
  Resolution string (e.g., "1920x1080")
  Codec string (H264|H265)
}

PERFORMANCE:
- Connection pool: max 600 connections
- Memory limit: 4GB per instance
- Stream buffer: 256KB per camera
- Health check response: <100ms
- Metrics scrape: <500ms
- CPU cores: 8
- Network: 10Gbps NIC

OBSERVABILITY:
Metrics (Prometheus format):
- mediamtx_streams_active{source="",status=""}
- mediamtx_streams_failed_total{source="",reason=""}
- mediamtx_reconnect_attempts_total{source=""}
- mediamtx_stream_bitrate_bytes{camera_id=""}
- mediamtx_stream_latency_ms{camera_id=""}

Logs (JSON to stdout):
{
  "timestamp": "ISO8601",
  "level": "info|warn|error",
  "camera_id": "uuid",
  "source": "DUBAI_POLICE",
  "event": "stream_connected",
  "latency_ms": 1823,
  "trace_id": "uuid"
}

CONFIGURATION:
Environment variables:
MILESTONE_SERVERS=server1:554,server2:554
VALKEY_ADDR=valkey:6379
VALKEY_PASSWORD=${VALKEY_PASSWORD}
MAX_STREAMS_POLICE=50
MAX_STREAMS_METRO=30
MAX_STREAMS_BUS=20
MAX_STREAMS_OTHER=400
LOG_LEVEL=info
METRICS_PORT=9090
RTSP_TRANSPORT=tcp
BUFFER_SIZE_KB=256

TESTING REQUIREMENTS:
- Unit tests: >80% coverage
- Integration test with mock RTSP server
- Load test: 500 concurrent streams for 1 hour
- Chaos test: random disconnections
- Memory leak test: 24-hour run

DELIVERABLES:
1. mediamtx.yml with all settings
2. Dockerfile (Alpine base, <100MB, multi-stage build)
3. docker-compose.yml segment
4. Health check script (/health endpoint)
5. Integration test suite
6. Runbook with troubleshooting guide
7. Capacity planning calculator (Excel/Sheets)
8. Kubernetes manifests (optional)
```

### **Prompt 1.2: Kong API Gateway Configuration (Enhanced)**
```
Configure Kong CE in DB-less mode as API gateway for RTA CCTV system with complete security and rate limiting.

FUNCTIONAL REQUIREMENTS:
- JWT validation using RS256 with RTA IAM public key
- Rate limiting using Valkey backend with sliding window algorithm
- Return bilingual messages (Arabic/English) for all errors
- Request correlation ID generation and propagation
- WebSocket support for live stats

RATE LIMITING RULES:
Per-source limits (sliding 60-second window):
- Dubai Police: 50 concurrent streams
- Metro: 30 concurrent streams
- Bus: 20 concurrent streams
- Other: 400 concurrent streams
Global limits:
- 500 total concurrent streams
- 1000 requests/minute per user
- 50 requests/second per IP

ERROR RESPONSES:
HTTP 429 Rate Limit Exceeded:
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message_en": "Camera limit reached for Dubai Police (50/50)",
    "message_ar": "تم الوصول إلى حد الكاميرات لشرطة دبي (50/50)",
    "retry_after": 30,
    "current_usage": 50,
    "limit": 50,
    "source": "DUBAI_POLICE"
  }
}

SECURITY CONFIGURATION:
- CORS: Allow only RTA domains
- Rate limit by: user_id, source, ip_address
- JWT validation: RS256, check exp, iat, nbf
- IP Whitelist for /metrics endpoint
- Request size limit: 10MB
- Header injection: X-Request-ID, X-User-ID

ROUTES CONFIGURATION:
routes:
  - name: live-stream
    paths: ["/api/v1/stream/live"]
    methods: ["POST", "DELETE"]
    plugins:
      - jwt: {key_claim_name: "iss"}
      - rate-limiting-advanced: {redis_host: "valkey"}
      - cors: {origins: ["https://*.rta.ae"]}
      - request-id: {header_name: "X-Request-ID"}
      
  - name: playback
    paths: ["/api/v1/playback"]
    methods: ["GET", "POST"]
    plugins:
      - jwt: {key_claim_name: "iss"}
      - request-size-limiting: {allowed_payload_size: 128}
      
  - name: cameras
    paths: ["/api/v1/cameras"]
    methods: ["GET"]
    plugins:
      - jwt: {key_claim_name: "iss"}
      - proxy-cache: {strategy: "redis", redis_host: "valkey"}
      
  - name: ptz-control
    paths: ["/api/v1/ptz"]
    methods: ["POST"]
    plugins:
      - jwt: {key_claim_name: "iss"}
      - acl: {allow: ["operator", "admin"]}
      
  - name: health
    paths: ["/health"]
    methods: ["GET"]
    strip_path: true
    
  - name: metrics
    paths: ["/metrics"]
    methods: ["GET"]
    plugins:
      - ip-restriction: {allow: ["10.0.0.0/8"]}

CUSTOM PLUGINS:
Create Lua plugin for source-based rate limiting:
- Check source from JWT claims or request body
- Atomic increment/decrement in Valkey
- Return appropriate error message in AR/EN
- Emit metrics for monitoring

OBSERVABILITY:
- Prometheus plugin on all routes
- Log all 4xx/5xx responses
- Trace sampling: 1% of requests
- Custom metrics: requests_by_source, rate_limit_hits

DELIVERABLES:
1. kong.yaml declarative configuration
2. Dockerfile with Kong CE + custom plugins
3. Custom Lua plugins source code
4. Error message templates (AR/EN)
5. Integration tests with mock backends
6. Load test scripts (K6)
7. TLS certificate configuration guide
8. Backup and restore procedures
```

### **Prompt 1.3: Valkey Stream Counter Service (Enhanced)**
```
Implement Valkey-based distributed stream counting service with high availability for RTA agency limits.

FUNCTIONAL REQUIREMENTS:
- Track active streams with atomic guarantees
- Support concurrent operations from multiple services
- Implement distributed locking for critical sections
- Provide real-time statistics via WebSocket
- Auto-cleanup of stale reservations

DATA MODEL:
Redis/Valkey keys structure:
- stream:count:{source} - Current active count (INTEGER)
- stream:limit:{source} - Maximum allowed (INTEGER)
- stream:reservation:{uuid} - Reservation details (HASH)
- stream:heartbeat:{uuid} - Last heartbeat timestamp (STRING)
- stream:stats:{source}:history - Time series data (SORTED SET)

Reservation Hash:
{
  "camera_id": "uuid",
  "source": "DUBAI_POLICE",
  "user_id": "user123",
  "reserved_at": "2024-01-01T10:00:00Z",
  "expires_at": "2024-01-01T10:00:30Z",
  "client_ip": "10.0.0.1"
}

LUA SCRIPTS:
1. reserve_stream.lua:
   - Check current count vs limit
   - Increment counter atomically
   - Create reservation with TTL
   - Return success/failure with details

2. release_stream.lua:
   - Verify reservation exists
   - Decrement counter atomically
   - Delete reservation
   - Log release event

3. heartbeat_stream.lua:
   - Update reservation TTL
   - Update heartbeat timestamp
   - Return remaining TTL

API ENDPOINTS:
POST /api/v1/stream/reserve
Request:
{
  "camera_id": "uuid",
  "source": "DUBAI_POLICE",
  "user_id": "user123",
  "duration_seconds": 3600
}
Response (200):
{
  "reservation_id": "uuid",
  "expires_at": "2024-01-01T11:00:00Z",
  "current_usage": 45,
  "limit": 50
}
Response (429):
{
  "error": "LIMIT_REACHED",
  "source": "DUBAI_POLICE",
  "current": 50,
  "limit": 50,
  "retry_after": 30
}

POST /api/v1/stream/heartbeat/{reservation_id}
DELETE /api/v1/stream/release/{reservation_id}
GET /api/v1/stream/stats
WebSocket /ws/stream/stats

ERROR HANDLING:
- Valkey connection loss: Use local cache for 30 seconds
- Reservation cleanup: Background job every 60 seconds
- Distributed lock timeout: 5 seconds max wait
- Circuit breaker: Open after 5 consecutive failures

PERFORMANCE:
- Operation latency: <10ms p99
- Throughput: 10,000 ops/second
- Connection pool: min=10, max=100
- Pipeline batch size: 100 operations
- Lua script caching: enabled

HIGH AVAILABILITY:
- Valkey Cluster: 3 nodes minimum
- Persistence: AOF with fsync every second
- Replication: 1 replica per master
- Failover: automatic via Sentinel
- Backup: Daily snapshots to object storage

OBSERVABILITY:
Metrics:
- stream_reservations_total{source,status}
- stream_current_count{source}
- stream_limit_hits_total{source}
- valkey_operation_duration_ms{operation}
- circuit_breaker_state{state}

Logs:
- All reservation attempts
- Limit exceeded events
- Cleanup operations
- Connection failures

TESTING:
- Unit tests with miniredis
- Integration tests with real Valkey
- Load test: 10,000 concurrent operations
- Chaos test: node failures during operations
- Race condition tests

DELIVERABLES:
1. Go service with chi router
2. Lua scripts with documentation
3. Dockerfile (multi-stage, <50MB)
4. Valkey cluster setup guide
5. API documentation (OpenAPI)
6. Performance test results
7. Monitoring dashboard (Grafana)
8. Disaster recovery procedures
```

## **Phase 2: Video Streaming Core**

### **Prompt 2.1: LiveKit SFU Configuration (Enhanced)**
```
Deploy and configure LiveKit Server with LiveKit Ingress for RTA CCTV WebRTC streaming.

FUNCTIONAL REQUIREMENTS:
- 500 unique camera rooms with unlimited viewers per room
- RTSP ingestion from MediaMTX
- Automatic quality adaptation based on network
- Recording capability for compliance
- Sub-2-second glass-to-glass latency

LIVEKIT CONFIGURATION:
livekit.yaml:
port: 7880
rtc:
  tcp_port: 7881
  port_range_start: 50000
  port_range_end: 60000
  use_external_ip: true
  enable_loopback_candidate: false
  
redis:
  address: valkey:6379
  password: ${REDIS_PASSWORD}
  db: 0
  
room:
  auto_create: false
  empty_timeout: 300
  max_participants: 100
  enable_recording: true
  
turn:
  enabled: true
  domain: turn.rta.ae
  port: 3478
  tls_port: 5349
  external_tls: true
  
webhook:
  api_key: ${WEBHOOK_KEY}
  urls:
    - http://go-api:8080/webhook/livekit
    
keys:
  APIKey1: ${LIVEKIT_API_KEY}
  
logging:
  level: info
  json: true
  
prometheus_port: 7782

INGRESS CONFIGURATION:
For each camera stream:
{
  "input": {
    "type": "RTSP_INPUT",
    "url": "rtsp://mediamtx:8554/camera_{id}",
    "transport": "TCP"
  },
  "room_name": "camera_{id}",
  "participant_identity": "camera_{id}_feed",
  "participant_name": "Camera {id}",
  "video": {
    "source": "SCREEN_SHARE",
    "preset": "H264_1080P_30FPS_3_LAYERS",
    "codec": "H264"
  },
  "audio": {
    "source": "MICROPHONE",
    "preset": "OPUS_STEREO_48KHZ"
  }
}

ROOM MANAGEMENT SERVICE (Go):
POST /rooms/create
{
  "camera_id": "uuid",
  "source": "DUBAI_POLICE",
  "stream_url": "rtsp://...",
  "max_participants": 100
}

GET /rooms/{camera_id}
POST /rooms/{camera_id}/token
DELETE /rooms/{camera_id}
GET /rooms/list?source=DUBAI_POLICE

Token generation with claims:
{
  "video": { "room": "camera_123", "canSubscribe": true },
  "metadata": { "source": "DUBAI_POLICE", "user_role": "operator" },
  "exp": "1 hour from now"
}

TURN SERVER (coturn):
listening-port=3478
tls-listening-port=5349
external-ip=${PUBLIC_IP}/${PRIVATE_IP}
realm=turn.rta.ae
server-name=turn.rta.ae
fingerprint
no-tcp-relay
user=rta:${TURN_PASSWORD}
cert=/etc/coturn/cert.pem
pkey=/etc/coturn/key.pem
log-file=stdout
simple-log

QUALITY ADAPTATION:
- Simulcast: 3 layers (1080p, 720p, 360p)
- Auto-switch based on bandwidth
- Prioritize video over audio
- Max bitrate: 6 Mbps per stream
- Min bitrate: 500 Kbps per stream

ERROR HANDLING:
- Ingress failure: Retry with exponential backoff
- Room creation failure: Queue and retry
- Participant limit reached: Return friendly error
- TURN allocation failure: Fallback to STUN
- Recording failure: Alert but don't block stream

PERFORMANCE:
- Room creation: <500ms
- Token generation: <50ms
- Participant join: <2s
- Switch quality: <500ms
- CPU per 100 rooms: <50%
- Memory per room: <10MB

MONITORING:
Metrics to track:
- livekit_room_count{source}
- livekit_participant_count{room}
- livekit_ingress_active
- livekit_packet_loss_rate
- livekit_jitter_ms
- turn_allocations_active

TESTING:
- Load test: 500 rooms with 10 viewers each
- Network simulation: packet loss, jitter
- Failover test: LiveKit node failure
- Quality adaptation test
- TURN connectivity test

DELIVERABLES:
1. livekit.yaml configuration
2. LiveKit Ingress setup scripts
3. Room management Go service
4. coturn configuration
5. Docker Compose stack
6. Client SDK wrapper (JavaScript)
7. Load testing scripts
8. Monitoring dashboard
9. Troubleshooting guide
```

### **Prompt 2.2: Milestone .NET Bridge Service (Enhanced)**
```
Create production-ready .NET 6 microservice for Milestone XProtect VMS integration.

FUNCTIONAL REQUIREMENTS:
- Connect to multiple Recording Servers simultaneously
- Maintain persistent connection with automatic reconnection
- Cache camera metadata with smart invalidation
- Support batch operations for efficiency
- Handle PTZ with validation and queueing

ARCHITECTURE:
- Clean Architecture with Domain, Application, Infrastructure, API layers
- CQRS pattern for commands and queries
- Repository pattern with Unit of Work
- Dependency injection throughout
- Background services for auto-refresh

PROJECT STRUCTURE:
RTA.CCTV.MilestoneBridge/
├── Domain/
│   ├── Entities/
│   │   ├── Camera.cs
│   │   ├── RecordingServer.cs
│   │   └── Recording.cs
│   ├── Enums/
│   │   └── CameraSource.cs
│   └── Events/
│       └── CameraUpdatedEvent.cs
├── Application/
│   ├── Commands/
│   │   ├── PTZCommand.cs
│   │   └── ExportRecordingCommand.cs
│   ├── Queries/
│   │   ├── GetCamerasQuery.cs
│   │   └── GetRecordingsQuery.cs
│   └── Services/
│       └── IMilestoneService.cs
├── Infrastructure/
│   ├── Milestone/
│   │   ├── MilestoneConnection.cs
│   │   └── MilestoneRepository.cs
│   ├── Persistence/
│   │   ├── ApplicationDbContext.cs
│   │   └── Configurations/
│   └── Caching/
│       └── RedisCacheService.cs
└── API/
    ├── Controllers/
    ├── Middleware/
    └── Program.cs

DATA MODELS:
Camera entity:
public class Camera
{
    public Guid Id { get; set; }
    public string MilestoneId { get; set; }
    public string Name { get; set; }
    public string NameAr { get; set; }
    public CameraSource Source { get; set; }
    public string RtspUrl { get; set; }
    public bool SupportsPtz { get; set; }
    public string RecordingServer { get; set; }
    public CameraStatus Status { get; set; }
    public DateTime LastSeen { get; set; }
    public Dictionary<string, object> Metadata { get; set; }
}

API ENDPOINTS:
GET /api/v1/cameras
Response:
{
  "cameras": [
    {
      "id": "uuid",
      "name": "Camera 001",
      "name_ar": "كاميرا 001",
      "source": "DUBAI_POLICE",
      "rtsp_url": "rtsp://...",
      "supports_ptz": true,
      "status": "ONLINE",
      "metadata": {...}
    }
  ],
  "total": 500,
  "last_updated": "2024-01-01T10:00:00Z"
}

POST /api/v1/cameras/{id}/ptz
Request:
{
  "action": "MOVE",
  "pan": 45.0,
  "tilt": 30.0,
  "zoom": 2.0
}

POST /api/v1/recordings/export
Request:
{
  "camera_id": "uuid",
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": "2024-01-01T11:00:00Z",
  "format": "MP4"
}

MILESTONE SDK INTEGRATION:
Connection management:
- Connection pool: 5 connections per Recording Server
- Heartbeat every 30 seconds
- Reconnection with exponential backoff
- Credential management via Azure Key Vault

Camera discovery:
- Full sync every 10 minutes
- Incremental updates via events
- Parallel processing for multiple servers
- Batch size: 50 cameras per query

PTZ operations:
- Command queue with priority
- Validation against camera capabilities
- Rate limiting: 10 commands/second/camera
- Audit logging of all commands

ERROR HANDLING:
- Polly policies for resilience
- Circuit breaker: 5 failures in 60 seconds
- Retry: 3 attempts with exponential backoff
- Timeout: 30 seconds for long operations
- Fallback: Return cached data when available

CACHING:
- Redis for camera metadata (TTL: 5 minutes)
- In-memory for frequently accessed cameras
- PostgreSQL for persistent storage
- Cache-aside pattern
- Invalidation on updates

SECURITY:
- Service account authentication
- Certificate validation for Milestone
- API key for internal services
- Audit logging for all operations
- No sensitive data in logs

PERFORMANCE:
- Startup time: <30 seconds
- Camera list: <500ms for 500 cameras
- PTZ command: <100ms latency
- Memory usage: <500MB
- Database connections: max 20

CONFIGURATION:
appsettings.json:
{
  "Milestone": {
    "ManagementServer": "milestone.rta.ae",
    "Username": "${MILESTONE_USER}",
    "Password": "${MILESTONE_PASS}",
    "AuthenticationType": "WindowsDefault"
  },
  "ConnectionStrings": {
    "PostgreSQL": "Host=postgres;Database=milestone;",
    "Redis": "valkey:6379"
  },
  "Caching": {
    "CameraTTL": 300,
    "RecordingTTL": 60
  }
}

TESTING:
- Unit tests: >80% coverage
- Integration tests with Milestone simulator
- Performance tests: 1000 req/s
- Memory leak detection
- Stress test: 24-hour run

DELIVERABLES:
1. Complete .NET 6 solution
2. Dockerfile (multi-stage, <200MB)
3. Database migration scripts
4. OpenAPI specification
5. Postman collection
6. Integration test suite
7. Performance test results
8. Deployment guide
9. Milestone SDK wrapper NuGet package
```

### **Prompt 2.3: FFmpeg Playback Service (Enhanced)**
```
Build production FFmpeg-based transmux service for Milestone recordings with HLS/DASH output.

FUNCTIONAL REQUIREMENTS:
- On-demand conversion of Milestone recordings to HLS
- Support H.264 and H.265 input without transcoding
- Generate 2-second segments for smooth scrubbing
- Implement signed URLs with 5-minute expiration
- Cache generated segments for reuse

SERVICE ARCHITECTURE:
Components:
- Job Queue (Redis-based)
- Worker Pool (concurrent FFmpeg processes)
- Segment Storage (shared volume)
- URL Signer (HMAC-SHA256)
- Cache Manager (LRU eviction)
- Nginx for serving

JOB PROCESSING FLOW:
1. Receive conversion request
2. Check cache for existing segments
3. Queue job if not cached
4. Worker pulls from Milestone
5. FFmpeg transmuxes to HLS
6. Store segments with metadata
7. Generate signed playlist URL
8. Return to client

DATA MODELS:
type TranscodeJob struct {
    ID           string    `json:"id"`
    CameraID     string    `json:"camera_id"`
    StartTime    time.Time `json:"start_time"`
    EndTime      time.Time `json:"end_time"`
    Status       JobStatus `json:"status"`
    Progress     int       `json:"progress"`
    OutputPath   string    `json:"output_path"`
    PlaylistURL  string    `json:"playlist_url"`
    Error        string    `json:"error,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    CompletedAt  time.Time `json:"completed_at,omitempty"`
}

API ENDPOINTS:
POST /api/v1/playback/prepare
Request:
{
  "camera_id": "uuid",
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": "2024-01-01T10:30:00Z",
  "format": "HLS"
}
Response:
{
  "job_id": "uuid",
  "status": "PROCESSING",
  "estimated_time": 30
}

GET /api/v1/playback/status/{job_id}
Response:
{
  "job_id": "uuid",
  "status": "COMPLETED",
  "progress": 100,
  "playlist_url": "https://cdn.rta.ae/playback/uuid/playlist.m3u8?token=...",
  "expires_at": "2024-01-01T10:35:00Z"
}

GET /api/v1/playback/playlist/{session_id}.m3u8
GET /api/v1/playback/segment/{session_id}/{segment}.ts

FFMPEG COMMANDS:
Transmux only (no transcoding):
ffmpeg -i input.mp4 \
  -c:v copy \
  -c:a copy \
  -f hls \
  -hls_time 2 \
  -hls_list_size 0 \
  -hls_segment_filename 'segment_%03d.ts' \
  -hls_flags independent_segments \
  -master_pl_name master.m3u8 \
  playlist.m3u8

With encoding (if required):
ffmpeg -i input.mp4 \
  -c:v libx264 \
  -preset fast \
  -crf 23 \
  -c:a aac \
  -b:a 128k \
  -f hls \
  -hls_time 2 \
  -hls_list_size 0 \
  -hls_segment_filename 'segment_%03d.ts' \
  playlist.m3u8

WORKER POOL:
Configuration:
- Max workers: 10
- Max concurrent per camera: 1
- Job timeout: 5 minutes
- Memory limit per worker: 2GB
- CPU limit per worker: 2 cores

Job prioritization:
- Live playback requests: Priority 1
- Export requests: Priority 2
- Pregeneration: Priority 3

Resource management:
- Monitor FFmpeg memory usage
- Kill stuck processes after timeout
- Cleanup temp files on failure
- Rate limit per user

CACHING STRATEGY:
- Cache key: {camera_id}:{start_time}:{end_time}:{format}
- Storage: Local NVMe SSD
- Max cache size: 500GB
- LRU eviction when 80% full
- Popular content pinning
- Segment-level caching

SIGNED URL GENERATION:
Format: /playlist.m3u8?token={base64}&expires={timestamp}
Token contains:
- Session ID
- User ID
- Camera ID
- Expiration timestamp
- HMAC signature

Validation:
- Check signature validity
- Verify timestamp not expired
- Validate user permissions
- Log access for audit

NGINX CONFIGURATION:
location /playback/ {
    root /var/cache/segments;
    add_header Cache-Control "private, max-age=300";
    add_header X-Content-Type-Options nosniff;
    
    # CORS headers
    add_header Access-Control-Allow-Origin "$allowed_origin";
    add_header Access-Control-Allow-Methods "GET, OPTIONS";
    
    # Rate limiting
    limit_req zone=playback burst=100 nodelay;
    limit_rate 10m;
    
    # Token validation via auth_request
    auth_request /validate_token;
}

ERROR HANDLING:
- Milestone unavailable: Return 503, retry later
- FFmpeg failure: Retry with encoding fallback
- Disk full: Clean old segments, alert ops
- Invalid time range: Return 400 with details
- Corrupt video: Skip segments, log issue

MONITORING:
Metrics:
- Jobs queued/processing/completed
- Average processing time
- Cache hit ratio
- Segment generation rate
- Storage usage
- FFmpeg CPU/memory usage

Alerts:
- Queue depth > 100
- Processing time > 2 minutes
- Cache hit ratio < 50%
- Storage usage > 80%
- Failed jobs > 10%

PERFORMANCE OPTIMIZATION:
- Use FFmpeg hardware acceleration when available
- Segment prefetching for sequential playback
- Parallel segment generation
- Connection pooling to Milestone
- Compress playlist files

TESTING:
- Unit tests for job queue logic
- Integration tests with mock video files
- Load test: 100 concurrent conversions
- Various codec/resolution inputs
- Network failure scenarios
- Storage exhaustion handling

DELIVERABLES:
1. Go service with job queue
2. FFmpeg Docker image with codecs
3. Nginx configuration
4. Segment storage management scripts
5. API documentation
6. Performance tuning guide
7. Monitoring dashboard
8. Troubleshooting playbook
9. Cache analysis tools
```

## **Phase 3: Web Application**

### **Prompt 3.1: React Operator Dashboard (Enhanced)**
```
Create production React 18 TypeScript operator dashboard for RTA CCTV with RTL Arabic support.

FUNCTIONAL REQUIREMENTS:
- Support 64 simultaneous video streams
- Grid layouts with drag-and-drop
- LiveKit WebRTC integration
- HLS.js playback with timeline
- PTZ controls with joystick
- Bilingual UI (Arabic/English)

PROJECT STRUCTURE:
rta-cctv-dashboard/
├── src/
│   ├── components/
│   │   ├── Camera/
│   │   │   ├── CameraGrid.tsx
│   │   │   ├── CameraStream.tsx
│   │   │   └── PTZControls.tsx
│   │   ├── Layout/
│   │   │   ├── GridSelector.tsx
│   │   │   ├── LayoutManager.tsx
│   │   │   └── GridTemplates.ts
│   │   ├── Playback/
│   │   │   ├── Timeline.tsx
│   │   │   ├── PlaybackControls.tsx
│   │   │   └── SegmentViewer.tsx
│   │   └── Common/
│   │       ├── ErrorBoundary.tsx
│   │       ├── LoadingSpinner.tsx
│   │       └── Toast.tsx
│   ├── hooks/
│   │   ├── useLiveKit.ts
│   │   ├── usePlayback.ts
│   │   └── useAgencyLimits.ts
│   ├── services/
│   │   ├── api.ts
│   │   ├── auth.ts
│   │   └── websocket.ts
│   ├── store/
│   │   ├── cameraStore.ts
│   │   ├── layoutStore.ts
│   │   └── userStore.ts
│   ├── i18n/
│   │   ├── ar.json
│   │   └── en.json
│   └── utils/
│       ├── gridCalculator.ts
│       └── videoOptimizer.ts

GRID LAYOUT SYSTEM:
Grid configurations:
const GRID_LAYOUTS = {
  "2x2": { rows: 2, cols: 2, cells: 4 },
  "3x3": { rows: 3, cols: 3, cells: 9 },
  "4x4": { rows: 4, cols: 4, cells: 16 },
  "16-way-hotspot": {
    rows: 4, cols: 4,
    hotspot: { row: 0, col: 0, rowSpan: 3, colSpan: 3 },
    cells: 7 + 1
  },
  "64-way-hotspot": {
    rows: 8, cols: 8,
    hotspot: { row: 0, col: 0, rowSpan: 7, colSpan: 7 },
    cells: 15 + 1
  }
};

Responsive calculations:
function calculateGridDimensions(
  containerWidth: number,
  containerHeight: number,
  layout: GridLayout
): CellDimensions[] {
  // Account for gaps and borders
  // Maintain 16:9 aspect ratio
  // Handle hotspot cells
  // Return array of cell positions and sizes
}

CAMERA STREAM COMPONENT:
interface CameraStreamProps {
  cameraId: string;
  source: CameraSource;
  isLive: boolean;
  quality: 'AUTO' | 'HIGH' | 'MEDIUM' | 'LOW';
  onError: (error: Error) => void;
}

const CameraStream: React.FC<CameraStreamProps> = ({
  cameraId,
  source,
  isLive,
  quality
}) => {
  const { room, participant, connectionQuality } = useLiveKit(cameraId);
  const videoRef = useRef<HTMLVideoElement>(null);
  
  // Intersection Observer for viewport detection
  const isInViewport = useIntersectionObserver(videoRef);
  
  // Only render video if in viewport (performance)
  // Show loading state
  // Handle connection failures
  // Display agency limit warnings
  // Implement reconnection logic
  
  return (
    <div className={styles.cameraContainer}>
      {isLive ? (
        <LiveKitVideo
          room={room}
          participant={participant}
          quality={quality}
          ref={videoRef}
        />
      ) : (
        <HLSPlayer
          url={playbackUrl}
          ref={videoRef}
        />
      )}
      {showPTZ && <PTZControls cameraId={cameraId} />}
      <CameraOverlay
        name={cameraName}
        source={source}
        quality={connectionQuality}
        recording={isRecording}
      />
    </div>
  );
};

STATE MANAGEMENT (Zustand):
interface CameraStore {
  cameras: Camera[];
  activeCameras: Map<string, StreamState>;
  agencyLimits: Map<CameraSource, LimitStatus>;
  
  addCamera: (gridIndex: number, cameraId: string) => Promise<void>;
  removeCamera: (gridIndex: number) => void;
  checkAgencyLimit: (source: CameraSource) => boolean;
  updateQuality: (cameraId: string, quality: Quality) => void;
}

const useCameraStore = create<CameraStore>((set, get) => ({
  cameras: [],
  activeCameras: new Map(),
  agencyLimits: new Map(),
  
  addCamera: async (gridIndex, cameraId) => {
    const camera = get().cameras.find(c => c.id === cameraId);
    if (!camera) return;
    
    // Check agency limit
    if (!get().checkAgencyLimit(camera.source)) {
      throw new AgencyLimitError(camera.source);
    }
    
    // Reserve stream slot
    const reservation = await api.reserveStream(cameraId, camera.source);
    
    // Start stream
    const stream = await startStream(cameraId, reservation);
    
    set(state => ({
      activeCameras: new Map(state.activeCameras).set(gridIndex, stream)
    }));
  }
}));

PERFORMANCE OPTIMIZATIONS:
1. Video element pooling:
const VideoPool = {
  available: [],
  inUse: new Map(),
  
  acquire(): HTMLVideoElement {
    return this.available.pop() || document.createElement('video');
  },
  
  release(video: HTMLVideoElement) {
    video.src = '';
    this.available.push(video);
  }
};

2. Viewport-based rendering:
- Only render visible cameras
- Pause invisible streams
- Reduce quality for small cells

3. Web Workers for heavy computations:
- Grid layout calculations
- Timeline data processing
- Metrics aggregation

4. Memoization and lazy loading:
const CameraGrid = React.memo(({ layout, cameras }) => {
  // Render grid
}, (prevProps, nextProps) => {
  // Custom comparison
});

ARABIC RTL SUPPORT:
// i18n configuration
i18n.use(initReactI18next).init({
  resources: { ar, en },
  lng: 'ar',
  fallbackLng: 'en',
  interpolation: { escapeValue: false }
});

// RTL detection and application
document.dir = i18n.language === 'ar' ? 'rtl' : 'ltr';

// Component RTL awareness
<div className={cn(styles.container, {
  [styles.rtl]: isRTL
})}>

WEBSOCKET REAL-TIME UPDATES:
const useRealtimeUpdates = () => {
  const ws = useWebSocket('/ws/updates');
  
  useEffect(() => {
    ws.on('agency-limit-update', (data) => {
      updateAgencyLimits(data);
    });
    
    ws.on('camera-status', (data) => {
      updateCameraStatus(data);
    });
    
    ws.on('alert', (data) => {
      showAlert(data);
    });
  }, []);
};

ERROR HANDLING:
class ErrorBoundary extends React.Component {
  componentDidCatch(error, errorInfo) {
    // Log to monitoring service
    logger.error('React error boundary', { error, errorInfo });
    
    // Show user-friendly error
    this.setState({
      hasError: true,
      error: this.translateError(error)
    });
  }
  
  translateError(error: Error): UserMessage {
    if (error instanceof AgencyLimitError) {
      return {
        ar: `تم الوصول إلى حد الكاميرات لـ ${error.source}`,
        en: `Camera limit reached for ${error.source}`
      };
    }
    // ... other error types
  }
}

TESTING:
// Component tests
describe('CameraGrid', () => {
  it('should handle 64 cameras without performance degradation', async () => {
    const cameras = generateMockCameras(64);
    const { container } = render(<CameraGrid cameras={cameras} />);
    
    // Measure rendering time
    const renderTime = await measureRenderTime();
    expect(renderTime).toBeLessThan(1000);
    
    // Check all cameras rendered
    expect(container.querySelectorAll('video')).toHaveLength(64);
  });
  
  it('should enforce agency limits', async () => {
    // Set Dubai Police limit to 50
    mockAgencyLimit('DUBAI_POLICE', 50);
    
    // Try to add 51st camera
    const result = await addCamera('DUBAI_POLICE', 51);
    expect(result).toThrow(AgencyLimitError);
  });
});

// E2E tests with Playwright
test('operator can save and load layouts', async ({ page }) => {
  await page.goto('/dashboard');
  
  // Arrange cameras in grid
  await page.dragAndDrop('#camera-1', '#grid-cell-0');
  await page.dragAndDrop('#camera-2', '#grid-cell-1');
  
  // Save layout
  await page.click('#save-layout');
  await page.fill('#layout-name', 'Morning Setup');
  await page.click('#save-confirm');
  
  // Load layout
  await page.click('#load-layout');
  await page.click('[data-layout="Morning Setup"]');
  
  // Verify cameras in correct positions
  await expect(page.locator('#grid-cell-0')).toContainText('Camera 1');
});

ACCESSIBILITY:
- WCAG 2.1 Level AA compliance
- Keyboard navigation for all controls
- Screen reader support
- High contrast mode
- Focus indicators
- ARIA labels in both languages

DELIVERABLES:
1. Complete React application
2. Component library documentation
3. Storybook for components
4. E2E test suite (Playwright)
5. Performance benchmarks
6. Accessibility audit report
7. Docker build for production
8. CI/CD pipeline configuration
9. Deployment guide
```

### **Prompt 3.2: Go API Backend Service (Enhanced)**
```
Create production Go API service as central backend for RTA CCTV system with clean architecture.

FUNCTIONAL REQUIREMENTS:
- Central API for all frontend operations
- Integration with all backend services
- JWT authentication with RTA IAM
- Real-time WebSocket updates
- Audit logging for compliance

PROJECT STRUCTURE:
rta-cctv-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── camera.go
│   │   ├── layout.go
│   │   ├── stream.go
│   │   └── user.go
│   ├── usecase/
│   │   ├── camera_usecase.go
│   │   ├── stream_usecase.go
│   │   └── auth_usecase.go
│   ├── repository/
│   │   ├── postgres/
│   │   ├── valkey/
│   │   └── interfaces.go
│   ├── delivery/
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── router.go
│   │   └── websocket/
│   │       └── hub.go
│   └── infrastructure/
│       ├── config/
│       ├── logger/
│       └── telemetry/
├── pkg/
│   ├── errors/
│   ├── validator/
│   └── utils/

DOMAIN MODELS:
// domain/camera.go
type Camera struct {
    ID          uuid.UUID      `json:"id"`
    Name        string         `json:"name"`
    NameAr      string         `json:"name_ar"`
    Source      CameraSource   `json:"source"`
    FolderID    uuid.UUID      `json:"folder_id"`
    RTSPURL     string         `json:"-"`
    SupportsPTZ bool          `json:"supports_ptz"`
    Status      CameraStatus   `json:"status"`
    Metadata    map[string]any `json:"metadata"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

type CameraSource string
const (
    SourceDubaiPolice CameraSource = "DUBAI_POLICE"
    SourceMetro      CameraSource = "METRO"
    SourceBus        CameraSource = "BUS"
    SourceOther      CameraSource = "OTHER"
)

// domain/stream.go
type StreamReservation struct {
    ID         uuid.UUID     `json:"id"`
    CameraID   uuid.UUID     `json:"camera_id"`
    UserID     string        `json:"user_id"`
    Source     CameraSource  `json:"source"`
    SessionID  string        `json:"session_id"`
    ExpiresAt  time.Time     `json:"expires_at"`
    CreatedAt  time.Time     `json:"created_at"`
}

// domain/layout.go
type Layout struct {
    ID        uuid.UUID         `json:"id"`
    Name      string            `json:"name"`
    UserID    string            `json:"user_id"`
    IsGlobal  bool             `json:"is_global"`
    GridType  string           `json:"grid_type"`
    Cameras   []LayoutCamera   `json:"cameras"`
    CreatedAt time.Time        `json:"created_at"`
}

USE CASES:
// usecase/stream_usecase.go
type StreamUseCase interface {
    RequestStream(ctx context.Context, cameraID uuid.UUID, userID string) (*StreamToken, error)
    ReleaseStream(ctx context.Context, reservationID uuid.UUID) error
    GetStreamStats(ctx context.Context) (map[CameraSource]StreamStats, error)
}

type streamUseCase struct {
    cameraRepo  repository.CameraRepository
    streamRepo  repository.StreamRepository
    valkeyRepo  repository.ValkeyRepository
    livekitSvc  service.LiveKitService
    logger      logger.Logger
}

func (u *streamUseCase) RequestStream(ctx context.Context, cameraID uuid.UUID, userID string) (*StreamToken, error) {
    // 1. Get camera details
    camera, err := u.cameraRepo.GetByID(ctx, cameraID)
    if err != nil {
        return nil, errors.Wrap(err, "camera not found")
    }
    
    // 2. Check user permissions
    if err := u.checkUserPermission(ctx, userID, camera); err != nil {
        return nil, errors.Wrap(err, "permission denied")
    }
    
    // 3. Check agency limits
    canReserve, current, limit := u.valkeyRepo.CheckAndReserve(ctx, camera.Source)
    if !canReserve {
        return nil, &errors.AgencyLimitError{
            Source:  camera.Source,
            Current: current,
            Limit:   limit,
        }
    }
    
    // 4. Create LiveKit room if needed
    room, err := u.livekitSvc.EnsureRoom(ctx, cameraID.String())
    if err != nil {
        u.valkeyRepo.Release(ctx, camera.Source)
        return nil, errors.Wrap(err, "failed to create room")
    }
    
    // 5. Generate access token
    token, err := u.livekitSvc.GenerateToken(ctx, room, userID)
    if err != nil {
        u.valkeyRepo.Release(ctx, camera.Source)
        return nil, errors.Wrap(err, "failed to generate token")
    }
    
    // 6. Save reservation
    reservation := &StreamReservation{
        ID:        uuid.New(),
        CameraID:  cameraID,
        UserID:    userID,
        Source:    camera.Source,
        SessionID: token.SessionID,
        ExpiresAt: time.Now().Add(1 * time.Hour),
    }
    
    if err := u.streamRepo.Create(ctx, reservation); err != nil {
        u.valkeyRepo.Release(ctx, camera.Source)
        return nil, errors.Wrap(err, "failed to save reservation")
    }
    
    // 7. Audit log
    u.logger.Audit("stream_requested", map[string]any{
        "user_id":   userID,
        "camera_id": cameraID,
        "source":    camera.Source,
    })
    
    return token, nil
}

HTTP HANDLERS:
// delivery/http/handlers/stream_handler.go
type StreamHandler struct {
    useCase usecase.StreamUseCase
    logger  logger.Logger
}

func (h *StreamHandler) RequestStream(w http.ResponseWriter, r *http.Request) {
    var req RequestStreamRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.BadRequest("invalid request body"))
        return
    }
    
    // Validate request
    if err := validator.Validate(req); err != nil {
        respondError(w, errors.BadRequest(err.Error()))
        return
    }
    
    // Get user from context (set by auth middleware)
    userID := middleware.UserIDFromContext(r.Context())
    
    // Process request
    token, err := h.useCase.RequestStream(r.Context(), req.CameraID, userID)
    if err != nil {
        switch e := err.(type) {
        case *errors.AgencyLimitError:
            respondJSON(w, http.StatusTooManyRequests, map[string]any{
                "error": map[string]any{
                    "code":       "AGENCY_LIMIT_EXCEEDED",
                    "message_en": fmt.Sprintf("Camera limit reached for %s (%d/%d)", e.Source, e.Current, e.Limit),
                    "message_ar": fmt.Sprintf("تم الوصول إلى حد الكاميرات لـ %s (%d/%d)", e.Source, e.Current, e.Limit),
                    "source":     e.Source,
                    "current":    e.Current,
                    "limit":      e.Limit,
                },
            })
        default:
            h.logger.Error("stream request failed", "error", err)
            respondError(w, errors.Internal("failed to process request"))
        }
        return
    }
    
    respondJSON(w, http.StatusOK, token)
}

MIDDLEWARE:
// delivery/http/middleware/auth.go
func JWTAuth(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                respondError(w, errors.Unauthorized("missing token"))
                return
            }
            
            claims, err := validateToken(token, secret)
            if err != nil {
                respondError(w, errors.Unauthorized("invalid token"))
                return
            }
            
            ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
            ctx = context.WithValue(ctx, userRoleKey, claims.Role)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// delivery/http/middleware/ratelimit.go
func RateLimit(limiter *rate.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                respondError(w, errors.TooManyRequests("rate limit exceeded"))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

WEBSOCKET HUB:
// delivery/websocket/hub.go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

type Client struct {
    hub    *Hub
    conn   *websocket.Conn
    send   chan Message
    userID string
    role   string
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
            
        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                if h.shouldReceive(client, message) {
                    select {
                    case client.send <- message:
                    default:
                        close(client.send)
                        delete(h.clients, client)
                    }
                }
            }
            h.mu.RUnlock()
        }
    }
}

REPOSITORY IMPLEMENTATIONS:
// repository/postgres/camera_repository.go
type cameraRepository struct {
    db *sql.DB
}

func (r *cameraRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Camera, error) {
    query := `
        SELECT id, name, name_ar, source, folder_id, rtsp_url, 
               supports_ptz, status, metadata, created_at, updated_at
        FROM cameras
        WHERE id = $1
    `
    
    var camera domain.Camera
    var metadata json.RawMessage
    
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &camera.ID, &camera.Name, &camera.NameAr, &camera.Source,
        &camera.FolderID, &camera.RTSPURL, &camera.SupportsPTZ,
        &camera.Status, &metadata, &camera.CreatedAt, &camera.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, errors.NotFound("camera not found")
    }
    if err != nil {
        return nil, errors.Wrap(err, "failed to get camera")
    }
    
    if err := json.Unmarshal(metadata, &camera.Metadata); err != nil {
        return nil, errors.Wrap(err, "failed to unmarshal metadata")
    }
    
    return &camera, nil
}

CONFIGURATION:
// internal/infrastructure/config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Valkey   ValkeyConfig
    LiveKit  LiveKitConfig
    Auth     AuthConfig
    Limits   LimitsConfig
}

type LimitsConfig struct {
    DubaiPolice int `env:"LIMIT_DUBAI_POLICE" envDefault:"50"`
    Metro       int `env:"LIMIT_METRO" envDefault:"30"`
    Bus         int `env:"LIMIT_BUS" envDefault:"20"`
    Other       int `env:"LIMIT_OTHER" envDefault:"400"`
    Total       int `env:"LIMIT_TOTAL" envDefault:"500"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, err
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}

OBSERVABILITY:
// internal/infrastructure/telemetry/metrics.go
var (
    streamRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cctv_stream_requests_total",
            Help: "Total number of stream requests",
        },
        []string{"source", "status"},
    )
    
    activeStreamsGauge = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cctv_active_streams",
            Help: "Number of active streams",
        },
        []string{"source"},
    )
    
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "cctv_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
)

TESTING:
// internal/usecase/stream_usecase_test.go
func TestRequestStream(t *testing.T) {
    // Setup
    mockCameraRepo := mocks.NewMockCameraRepository(t)
    mockStreamRepo := mocks.NewMockStreamRepository(t)
    mockValkeyRepo := mocks.NewMockValkeyRepository(t)
    mockLiveKitSvc := mocks.NewMockLiveKitService(t)
    
    useCase := NewStreamUseCase(
        mockCameraRepo,
        mockStreamRepo,
        mockValkeyRepo,
        mockLiveKitSvc,
        logger.NewNoop(),
    )
    
    t.Run("successful stream request", func(t *testing.T) {
        // Arrange
        cameraID := uuid.New()
        userID := "user123"
        camera := &domain.Camera{
            ID:     cameraID,
            Source: domain.SourceDubaiPolice,
        }
        
        mockCameraRepo.On("GetByID", mock.Anything, cameraID).Return(camera, nil)
        mockValkeyRepo.On("CheckAndReserve", mock.Anything, domain.SourceDubaiPolice).Return(true, 45, 50)
        mockLiveKitSvc.On("EnsureRoom", mock.Anything, cameraID.String()).Return("room123", nil)
        mockLiveKitSvc.On("GenerateToken", mock.Anything, "room123", userID).Return(&StreamToken{}, nil)
        mockStreamRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.StreamReservation")).Return(nil)
        
        // Act
        token, err := useCase.RequestStream(context.Background(), cameraID, userID)
        
        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, token)
        mockCameraRepo.AssertExpectations(t)
    })
    
    t.Run("agency limit exceeded", func(t *testing.T) {
        // Test limit exceeded scenario
    })
}

// Integration test
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Start test containers
    ctx := context.Background()
    postgres, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "postgres:15",
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
            },
        },
    })
    defer postgres.Terminate(ctx)
    
    // Run integration tests
}

DELIVERABLES:
1. Complete Go API service
2. Database migration scripts
3. OpenAPI specification
4. Postman collection
5. Unit test suite (>80% coverage)
6. Integration test suite
7. Load test scripts
8. Dockerfile (multi-stage, <50MB)
9. Kubernetes manifests
10. API documentation


### **Prompt 3.3: RTA IAM Integration Service (Enhanced)**

```
Implement complete RTA IAM integration service matching REQ10 specifications with audit and compliance.

FUNCTIONAL REQUIREMENTS:
- Implement all required IAM endpoints
- Map RTA organizational structure
- Handle Arabic names and data
- Provide complete audit trail
- Support webhook events from IAM

DATA MODELS:
// RTA User Structure (as per requirements)
type RTAUser struct {
    UserID          string         `json:"user_id" validate:"required"`
    LoginID         string         `json:"login_id" validate:"required"`
    EmployeeNumber  string         `json:"employee_number" validate:"required"`
    DisplayName     string         `json:"display_name" validate:"required"`
    Email           string         `json:"email" validate:"required,email"`
    FullNameArabic  string         `json:"full_name_arabic" validate:"required"`
    Agency          Agency         `json:"agency" validate:"required"`
    Department      Department     `json:"department" validate:"required"`
    Section         Section        `json:"section" validate:"required"`
    Status          UserStatus     `json:"status" validate:"required"`
    Roles           []string       `json:"roles"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
}

type Agency struct {
    Code        string `json:"code" validate:"required"`
    Name        string `json:"name" validate:"required"`
    NameArabic  string `json:"name_arabic" validate:"required"`
}

type Department struct {
    Code        string `json:"code" validate:"required"`
    Name        string `json:"name" validate:"required"`
    NameArabic  string `json:"name_arabic" validate:"required"`
}

type Section struct {
    Code        string `json:"code" validate:"required"`
    Name        string `json:"name" validate:"required"`
    NameArabic  string `json:"name_arabic" validate:"required"`
}

type UserStatus string
const (
    StatusActive   UserStatus = "ACTIVE"
    StatusDisabled UserStatus = "DISABLED"
    StatusPending  UserStatus = "PENDING"
)

// Groups and Permissions
type Group struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    NameArabic  string       `json:"name_arabic"`
    Description string       `json:"description"`
    Permissions []Permission `json:"permissions"`
    CreatedAt   time.Time    `json:"created_at"`
}

type Permission struct {
    ID          string `json:"id"`
    Resource    string `json:"resource"`
    Action      string `json:"action"`
    Description string `json:"description"`
}

REQUIRED API ENDPOINTS:
// POST /iam/users - Create User
func (h *IAMHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.BadRequest("invalid request"))
        return
    }
    
    // Validate required fields
    if err := validator.Validate(req); err != nil {
        respondError(w, errors.BadRequest(err.Error()))
        return
    }
    
    // Check if user already exists
    existing, _ := h.service.GetUserByLoginID(r.Context(), req.LoginID)
    if existing != nil {
        respondError(w, errors.Conflict("user already exists"))
        return
    }
    
    // Create user
    user := &RTAUser{
        UserID:         uuid.New().String(),
        LoginID:        req.LoginID,
        EmployeeNumber: req.EmployeeNumber,
        DisplayName:    req.DisplayName,
        Email:          req.Email,
        FullNameArabic: req.FullNameArabic,
        Agency:         req.Agency,
        Department:     req.Department,
        Section:        req.Section,
        Status:         StatusPending,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    
    // Save to database
    if err := h.service.CreateUser(r.Context(), user); err != nil {
        h.logger.Error("failed to create user", "error", err)
        respondError(w, errors.Internal("failed to create user"))
        return
    }
    
    // Audit log
    h.auditLog.Log(AuditEvent{
        Type:      "USER_CREATED",
        UserID:    user.UserID,
        Actor:     getActorFromContext(r.Context()),
        Timestamp: time.Now(),
        Details:   map[string]any{"login_id": user.LoginID},
    })
    
    respondJSON(w, http.StatusCreated, user)
}

// PUT /iam/users/{id} - Update User Profile
func (h *IAMHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    
    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.BadRequest("invalid request"))
        return
    }
    
    // Get existing user
    user, err := h.service.GetUserByID(r.Context(), userID)
```

Retry

R

Continue

````
(continued from IAMHandler.UpdateUser)

    if err != nil {
        if errors.IsNotFound(err) {
            respondError(w, errors.NotFound("user not found"))
            return
        }
        respondError(w, errors.Internal("failed to get user"))
        return
    }
    
    // Update fields
    if req.Department != nil {
        user.Department = *req.Department
    }
    if req.Agency != nil {
        user.Agency = *req.Agency
    }
    if req.Section != nil {
        user.Section = *req.Section
    }
    user.UpdatedAt = time.Now()
    
    // Save changes
    if err := h.service.UpdateUser(r.Context(), user); err != nil {
        respondError(w, errors.Internal("failed to update user"))
        return
    }
    
    // Audit log
    h.auditLog.Log(AuditEvent{
        Type:      "USER_UPDATED",
        UserID:    userID,
        Actor:     getActorFromContext(r.Context()),
        Timestamp: time.Now(),
        Changes:   req,
    })
    
    respondJSON(w, http.StatusOK, user)
}

// PATCH /iam/users/{id}/status - Enable/Disable User
func (h *IAMHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    
    var req UpdateStatusRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.BadRequest("invalid request"))
        return
    }
    
    if req.Status != StatusActive && req.Status != StatusDisabled {
        respondError(w, errors.BadRequest("invalid status"))
        return
    }
    
    // Update status
    if err := h.service.UpdateUserStatus(r.Context(), userID, req.Status); err != nil {
        respondError(w, errors.Internal("failed to update status"))
        return
    }
    
    // Revoke active sessions if disabling
    if req.Status == StatusDisabled {
        if err := h.sessionService.RevokeUserSessions(r.Context(), userID); err != nil {
            h.logger.Error("failed to revoke sessions", "error", err)
        }
    }
    
    // Audit log
    h.auditLog.Log(AuditEvent{
        Type:      "USER_STATUS_CHANGED",
        UserID:    userID,
        Actor:     getActorFromContext(r.Context()),
        Timestamp: time.Now(),
        Details:   map[string]any{"status": req.Status},
    })
    
    respondJSON(w, http.StatusOK, map[string]any{
        "user_id": userID,
        "status":  req.Status,
    })
}

// GET /iam/groups - Get All Groups
func (h *IAMHandler) GetAllGroups(w http.ResponseWriter, r *http.Request) {
    groups, err := h.service.GetAllGroups(r.Context())
    if err != nil {
        respondError(w, errors.Internal("failed to get groups"))
        return
    }
    
    // Include meaningful descriptions as required
    response := make([]GroupResponse, len(groups))
    for i, g := range groups {
        response[i] = GroupResponse{
            ID:          g.ID,
            Name:        g.Name,
            NameArabic:  g.NameArabic,
            Description: g.Description,
            Permissions: g.Permissions,
        }
    }
    
    respondJSON(w, http.StatusOK, map[string]any{
        "groups": response,
        "total":  len(response),
    })
}

// POST /iam/users/{id}/groups - Assign User to Group
func (h *IAMHandler) AssignUserToGroup(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    
    var req AssignGroupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.BadRequest("invalid request"))
        return
    }
    
    // Validate group exists
    group, err := h.service.GetGroup(r.Context(), req.GroupID)
    if err != nil {
        respondError(w, errors.NotFound("group not found"))
        return
    }
    
    // Assign user to group
    if err := h.service.AssignUserToGroup(r.Context(), userID, req.GroupID); err != nil {
        respondError(w, errors.Internal("failed to assign group"))
        return
    }
    
    // Update camera access based on new permissions
    if err := h.updateCameraAccess(r.Context(), userID, group.Permissions); err != nil {
        h.logger.Error("failed to update camera access", "error", err)
    }
    
    // Audit log
    h.auditLog.Log(AuditEvent{
        Type:      "GROUP_ASSIGNED",
        UserID:    userID,
        Actor:     getActorFromContext(r.Context()),
        Timestamp: time.Now(),
        Details:   map[string]any{"group_id": req.GroupID, "group_name": group.Name},
    })
    
    respondJSON(w, http.StatusOK, map[string]any{
        "user_id":  userID,
        "group_id": req.GroupID,
        "assigned": true,
    })
}

// DELETE /iam/users/{id}/groups/{groupId} - Remove User from Group
func (h *IAMHandler) RemoveUserFromGroup(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    groupID := chi.URLParam(r, "groupId")
    
    // Remove user from group
    if err := h.service.RemoveUserFromGroup(r.Context(), userID, groupID); err != nil {
        respondError(w, errors.Internal("failed to remove from group"))
        return
    }
    
    // Update camera access
    if err := h.updateCameraAccess(r.Context(), userID, nil); err != nil {
        h.logger.Error("failed to update camera access", "error", err)
    }
    
    // Audit log
    h.auditLog.Log(AuditEvent{
        Type:      "GROUP_REMOVED",
        UserID:    userID,
        Actor:     getActorFromContext(r.Context()),
        Timestamp: time.Now(),
        Details:   map[string]any{"group_id": groupID},
    })
    
    respondJSON(w, http.StatusOK, map[string]any{
        "user_id":  userID,
        "group_id": groupID,
        "removed":  true,
    })
}

SERVICE LAYER:
// service/iam_service.go
type IAMService struct {
    userRepo  repository.UserRepository
    groupRepo repository.GroupRepository
    cache     cache.Cache
    logger    logger.Logger
}

func (s *IAMService) CreateUser(ctx context.Context, user *RTAUser) error {
    // Begin transaction
    tx, err := s.userRepo.BeginTx(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Create user
    if err := tx.CreateUser(user); err != nil {
        return err
    }
    
    // Assign default group based on agency
    defaultGroup := s.getDefaultGroupForAgency(user.Agency.Code)
    if defaultGroup != "" {
        if err := tx.AssignGroup(user.UserID, defaultGroup); err != nil {
            return err
        }
    }
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        return err
    }
    
    // Invalidate cache
    s.cache.Delete(fmt.Sprintf("user:%s", user.UserID))
    
    // Send welcome email
    go s.sendWelcomeEmail(user)
    
    return nil
}

WEBHOOK HANDLER:
// handlers/webhook_handler.go
func (h *WebhookHandler) HandleIAMEvent(w http.ResponseWriter, r *http.Request) {
    // Verify webhook signature
    signature := r.Header.Get("X-IAM-Signature")
    if !h.verifySignature(r.Body, signature) {
        respondError(w, errors.Unauthorized("invalid signature"))
        return
    }
    
    var event IAMEvent
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        respondError(w, errors.BadRequest("invalid event"))
        return
    }
    
    switch event.Type {
    case "USER_CREATED":
        h.handleUserCreated(event)
    case "USER_UPDATED":
        h.handleUserUpdated(event)
    case "USER_DISABLED":
        h.handleUserDisabled(event)
    case "GROUP_CHANGED":
        h.handleGroupChanged(event)
    default:
        h.logger.Warn("unknown event type", "type", event.Type)
    }
    
    w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) handleUserDisabled(event IAMEvent) {
    userID := event.Data["user_id"].(string)
    
    // Revoke all active streams
    streams, _ := h.streamService.GetUserStreams(context.Background(), userID)
    for _, stream := range streams {
        h.streamService.ReleaseStream(context.Background(), stream.ID)
    }
    
    // Revoke all sessions
    h.sessionService.RevokeUserSessions(context.Background(), userID)
    
    // Audit log
    h.auditLog.Log(AuditEvent{
        Type:      "USER_DISABLED_WEBHOOK",
        UserID:    userID,
        Timestamp: time.Now(),
    })
}

JWT TOKEN GENERATION:
// auth/token_generator.go
type TokenGenerator struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    issuer     string
}

func (g *TokenGenerator) GenerateToken(user *RTAUser, groups []Group) (string, error) {
    now := time.Now()
    
    // Build permissions from groups
    permissions := make([]string, 0)
    sources := make([]string, 0)
    
    for _, group := range groups {
        for _, perm := range group.Permissions {
            permissions = append(permissions, fmt.Sprintf("%s:%s", perm.Resource, perm.Action))
            
            // Extract camera sources from permissions
            if strings.HasPrefix(perm.Resource, "camera:") {
                source := strings.TrimPrefix(perm.Resource, "camera:")
                sources = append(sources, source)
            }
        }
    }
    
    claims := jwt.MapClaims{
        "iss":            g.issuer,
        "sub":            user.UserID,
        "exp":            now.Add(1 * time.Hour).Unix(),
        "iat":            now.Unix(),
        "nbf":            now.Unix(),
        "login_id":       user.LoginID,
        "employee_number": user.EmployeeNumber,
        "display_name":   user.DisplayName,
        "email":          user.Email,
        "agency":         user.Agency.Code,
        "department":     user.Department.Code,
        "permissions":    permissions,
        "camera_sources": sources,
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(g.privateKey)
}

DATABASE SCHEMA:
-- users table
CREATE TABLE users (
    user_id VARCHAR(36) PRIMARY KEY,
    login_id VARCHAR(100) UNIQUE NOT NULL,
    employee_number VARCHAR(50) NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    email VARCHAR(200) NOT NULL,
    full_name_arabic NVARCHAR(200) NOT NULL,
    agency_code VARCHAR(50) NOT NULL,
    agency_name VARCHAR(200) NOT NULL,
    agency_name_arabic NVARCHAR(200) NOT NULL,
    department_code VARCHAR(50) NOT NULL,
    department_name VARCHAR(200) NOT NULL,
    department_name_arabic NVARCHAR(200) NOT NULL,
    section_code VARCHAR(50) NOT NULL,
    section_name VARCHAR(200) NOT NULL,
    section_name_arabic NVARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_login TIMESTAMP,
    INDEX idx_login_id (login_id),
    INDEX idx_agency (agency_code),
    INDEX idx_status (status)
);

-- groups table
CREATE TABLE groups (
    group_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    name_arabic NVARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL
);

-- user_groups junction table
CREATE TABLE user_groups (
    user_id VARCHAR(36) NOT NULL,
    group_id VARCHAR(36) NOT NULL,
    assigned_at TIMESTAMP NOT NULL,
    assigned_by VARCHAR(36),
    PRIMARY KEY (user_id, group_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE
);

-- audit_log table
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    user_id VARCHAR(36),
    actor_id VARCHAR(36),
    timestamp TIMESTAMP NOT NULL,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    INDEX idx_user_id (user_id),
    INDEX idx_timestamp (timestamp),
    INDEX idx_event_type (event_type)
);

TESTING:
// iam_service_test.go
func TestIAMService(t *testing.T) {
    t.Run("create user with Arabic data", func(t *testing.T) {
        user := &RTAUser{
            LoginID:        "ahmed.ali",
            DisplayName:    "Ahmed Ali",
            FullNameArabic: "أحمد علي",
            Agency: Agency{
                Code:       "RTA_TRAFFIC",
                Name:       "Traffic Department",
                NameArabic: "إدارة المرور",
            },
        }
        
        err := service.CreateUser(context.Background(), user)
        assert.NoError(t, err)
        assert.NotEmpty(t, user.UserID)
    })
    
    t.Run("enforce group permissions", func(t *testing.T) {
        // Test permission enforcement
    })
}

DELIVERABLES:
1. Complete IAM service implementation
2. Database schema with migrations
3. OpenAPI specification for IAM endpoints
4. Mock IAM server for testing
5. Integration test suite
6. Webhook handler documentation
7. JWT token validation library
8. Group permission mapping guide
9. Audit log analyzer tool
10. Deployment configuration
```

## **Phase 4: Deployment & Monitoring**

### **Prompt 4.1: Complete Docker Compose Stack (Enhanced)**
```
Create production-ready Docker Compose configuration for complete RTA CCTV system with HA and monitoring.

COMPOSE FILE STRUCTURE:
version: '3.9'

networks:
  frontend:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/24
  backend:
    driver: bridge
    ipam:
      config:
        - subnet: 172.21.0.0/24
  data:
    driver: bridge
    ipam:
      config:
        - subnet: 172.22.0.0/24
  monitoring:
    driver: bridge
    ipam:
      config:
        - subnet: 172.23.0.0/24

volumes:
  postgres_data:
  valkey_data:
  segment_cache:
  prometheus_data:
  grafana_data:
  loki_data:

services:
  # INGEST LAYER
  mediamtx-1:
    image: rta/mediamtx:1.0.0
    container_name: mediamtx-1
    restart: always
    networks:
      - backend
    ports:
      - "8554:8554"  # RTSP
      - "8889:8889"  # WebRTC
    environment:
      - RTSP_PROTOCOLS=tcp
      - RTSP_READTIMEOUT=10s
      - METRICS_ADDRESS=:9090
    volumes:
      - ./config/mediamtx.yml:/mediamtx.yml:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/metrics"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 4G
        reservations:
          cpus: '2'
          memory: 2G

  mediamtx-2:
    image: rta/mediamtx:1.0.0
    container_name: mediamtx-2
    # Same config as mediamtx-1
    
  # STREAMING LAYER
  livekit:
    image: livekit/livekit-server:v1.5.0
    container_name: livekit
    restart: always
    networks:
      - backend
    ports:
      - "7880:7880"  # HTTP
      - "7881:7881"  # RTC/TCP
      - "50000-50100:50000-50100/udp"  # RTC/UDP
    environment:
      - LIVEKIT_CONFIG=/config/livekit.yaml
    volumes:
      - ./config/livekit.yaml:/config/livekit.yaml:ro
    depends_on:
      - valkey
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7880/"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '8'
          memory: 8G

  livekit-ingress:
    image: livekit/ingress:latest
    container_name: livekit-ingress
    restart: always
    networks:
      - backend
    environment:
      - INGRESS_CONFIG_FILE=/config/ingress.yaml
    volumes:
      - ./config/ingress.yaml:/config/ingress.yaml:ro
    depends_on:
      - livekit
      - mediamtx-1

  coturn:
    image: coturn/coturn:4.6
    container_name: coturn
    restart: always
    network_mode: host
    volumes:
      - ./config/turnserver.conf:/etc/coturn/turnserver.conf:ro
      - ./certs:/etc/coturn/certs:ro
    
  # API GATEWAY
  kong:
    image: kong:3.4-alpine
    container_name: kong
    restart: always
    networks:
      - frontend
      - backend
    ports:
      - "8000:8000"  # HTTP
      - "8443:8443"  # HTTPS
    environment:
      - KONG_DATABASE=off
      - KONG_DECLARATIVE_CONFIG=/kong.yaml
      - KONG_PROXY_ACCESS_LOG=/dev/stdout
      - KONG_ADMIN_ACCESS_LOG=/dev/stdout
      - KONG_PROXY_ERROR_LOG=/dev/stderr
      - KONG_ADMIN_ERROR_LOG=/dev/stderr
    volumes:
      - ./config/kong.yaml:/kong.yaml:ro
    healthcheck:
      test: ["CMD", "kong", "health"]
      interval: 30s
      timeout: 10s
      retries: 3
      
  # DATA STORES
  postgres:
    image: postgres:15-alpine
    container_name: postgres
    restart: always
    networks:
      - data
    environment:
      - POSTGRES_DB=cctv
      - POSTGRES_USER=cctv
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_MAX_CONNECTIONS=200
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cctv"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G

  valkey:
    image: valkey/valkey:7.2-alpine
    container_name: valkey
    restart: always
    networks:
      - data
    ports:
      - "6379:6379"
    environment:
      - VALKEY_PASSWORD=${VALKEY_PASSWORD}
    volumes:
      - valkey_data:/data
      - ./config/valkey.conf:/usr/local/etc/valkey/valkey.conf:ro
    command: valkey-server /usr/local/etc/valkey/valkey.conf
    healthcheck:
      test: ["CMD", "valkey-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      
  # APPLICATION SERVICES
  go-api:
    image: rta/go-api:1.0.0
    container_name: go-api
    restart: always
    networks:
      - backend
      - data
    environment:
      - DATABASE_URL=postgres://cctv:${POSTGRES_PASSWORD}@postgres:5432/cctv
      - VALKEY_ADDR=valkey:6379
      - LIVEKIT_URL=http://livekit:7880
      - MILESTONE_URL=http://milestone-bridge:8080
    depends_on:
      - postgres
      - valkey
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '2'
          memory: 2G

  milestone-bridge:
    image: rta/milestone-bridge:1.0.0
    container_name: milestone-bridge
    restart: always
    networks:
      - backend
      - data
    environment:
      - ConnectionStrings__PostgreSQL=Host=postgres;Database=milestone
      - ConnectionStrings__Redis=valkey:6379
      - Milestone__ManagementServer=${MILESTONE_SERVER}
    depends_on:
      - postgres
      - valkey
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  ffmpeg-service:
    image: rta/ffmpeg-service:1.0.0
    container_name: ffmpeg-service
    restart: always
    networks:
      - backend
    volumes:
      - segment_cache:/cache
    environment:
      - MILESTONE_URL=http://milestone-bridge:8080
      - CACHE_PATH=/cache
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 4G

  # WEB FRONTEND
  web-dashboard:
    image: rta/web-dashboard:1.0.0
    container_name: web-dashboard
    restart: always
    networks:
      - frontend
    environment:
      - REACT_APP_API_URL=https://api.cctv.rta.ae
      - REACT_APP_WS_URL=wss://api.cctv.rta.ae/ws
    
  nginx:
    image: nginx:alpine
    container_name: nginx
    restart: always
    networks:
      - frontend
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
      - segment_cache:/var/cache/segments:ro
    depends_on:
      - web-dashboard
      - kong

  # MONITORING STACK
  prometheus:
    image: prom/prometheus:v2.45.0
    container_name: prometheus
    restart: always
    networks:
      - monitoring
      - backend
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

  grafana:
    image: grafana/grafana:10.0.0
    container_name: grafana
    restart: always
    networks:
      - monitoring
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
      - GF_AUTH_ANONYMOUS_ENABLED=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./datasources:/etc/grafana/provisioning/datasources:ro
    depends_on:
      - prometheus
      - loki

  loki:
    image: grafana/loki:2.9.0
    container_name: loki
    restart: always
    networks:
      - monitoring
      - backend
    ports:
      - "3100:3100"
    volumes:
      - ./config/loki.yml:/etc/loki/config.yml:ro
      - loki_data:/loki
    command: -config.file=/etc/loki/config.yml

  promtail:
    image: grafana/promtail:2.9.0
    container_name: promtail
    restart: always
    networks:
      - monitoring
      - backend
    volumes:
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - ./config/promtail.yml:/etc/promtail/config.yml:ro
    command: -config.file=/etc/promtail/config.yml
    depends_on:
      - loki

HEALTH CHECK SCRIPT:
#!/bin/bash
# health-check.sh

services=(
  "mediamtx-1:8554"
  "mediamtx-2:8554"
  "livekit:7880"
  "kong:8000"
  "postgres:5432"
  "valkey:6379"
  "go-api:8080"
  "prometheus:9090"
  "grafana:3000"
)

for service in "${services[@]}"; do
  name="${service%:*}"
  port="${service#*:}"
  
  if nc -z "$name" "$port" 2>/dev/null; then
    echo "✓ $name is healthy"
  else
    echo "✗ $name is unhealthy"
    exit 1
  fi
done

echo "All services healthy!"

SCALING SCRIPT:
#!/bin/bash
# scale.sh

case "$1" in
  up)
    docker-compose up -d --scale go-api=3 --scale mediamtx=2
    ;;
  down)
    docker-compose up -d --scale go-api=1 --scale mediamtx=1
    ;;
  *)
    echo "Usage: $0 {up|down}"
    exit 1
    ;;
esac

BACKUP SCRIPT:
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Backup PostgreSQL
docker exec postgres pg_dumpall -U cctv > "$BACKUP_DIR/postgres.sql"

# Backup Valkey
docker exec valkey valkey-cli --rdb "$BACKUP_DIR/valkey.rdb"

# Backup configurations
cp -r ./config "$BACKUP_DIR/config"

# Create archive
tar -czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
rm -rf "$BACKUP_DIR"

echo "Backup completed: $BACKUP_DIR.tar.gz"

DELIVERABLES:
1. docker-compose.yml with all services
2. Environment variable template (.env.example)
3. Configuration files for each service
4. Health check scripts
5. Scaling scripts
6. Backup and restore procedures
7. Network topology diagram
8. SSL/TLS certificate setup guide
9. Deployment checklist
10. Disaster recovery plan
````

This completes the full set of production-ready implementation prompts for the RTA CCTV system. Each prompt now includes:

- Complete functional requirements
- Detailed error handling
- Security considerations
- Performance specifications
- Data models and schemas
- Testing requirements
- Monitoring and observability
- Full code examples where needed

The prompts are designed to be executed sequentially but can also work independently for specific components.
