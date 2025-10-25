# Go API Service

The central API orchestration service for the RTA CCTV system. This service coordinates stream requests, manages LiveKit rooms, enforces agency quotas, and provides real-time updates via WebSocket.

## Features

- **Stream Management**: Reserve and release camera streams with quota enforcement
- **LiveKit Integration**: Automated room creation and JWT token generation
- **Camera Management**: List cameras and control PTZ
- **Real-time Updates**: WebSocket hub for live stream statistics
- **Agency Quota Enforcement**: Integration with Stream Counter service
- **Clean Architecture**: Domain-driven design with clear separation of concerns

## Architecture

```
go-api/
├── cmd/
│   └── main.go                    # Service entry point
├── internal/
│   ├── domain/                    # Business entities
│   │   ├── stream.go              # Stream models
│   │   └── camera.go              # Camera models
│   ├── usecase/                   # Business logic
│   │   └── stream_usecase.go     # Stream request logic
│   ├── client/                    # External service clients
│   │   ├── livekit_client.go     # LiveKit SDK wrapper
│   │   ├── vms_client.go         # VMS Service client
│   │   └── stream_counter_client.go  # Stream Counter client
│   ├── repository/                # Data access
│   │   ├── stream_repository.go  # Repository interface
│   │   └── valkey/               # Valkey implementation
│   │       └── stream_repository.go
│   └── delivery/                  # HTTP/WebSocket handlers
│       ├── http/
│       │   ├── stream_handler.go
│       │   ├── camera_handler.go
│       │   └── router.go
│       └── websocket/
│           ├── hub.go
│           └── handler.go
```

## API Endpoints

### Stream Management

#### Reserve Stream
Request a camera stream with automatic quota checking and LiveKit room creation.

```bash
POST /api/v1/stream/reserve
Content-Type: application/json

{
  "camera_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user123",
  "quality": "medium"  // optional: high, medium, low
}

# Success Response (201 Created)
{
  "reservation_id": "uuid",
  "camera_id": "550e8400-e29b-41d4-a716-446655440000",
  "camera_name": "Camera 1 - Main Entrance",
  "room_name": "camera_550e8400-e29b-41d4-a716-446655440000",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "livekit_url": "ws://livekit:7880",
  "expires_at": "2024-01-20T12:00:00Z",
  "quality": "medium"
}

# Error Response - Agency Limit Exceeded (429 Too Many Requests)
{
  "error": {
    "code": "AGENCY_LIMIT_EXCEEDED",
    "message_en": "Agency limit reached for DUBAI_POLICE (50/50)",
    "message_ar": "تم الوصول إلى حد الوكالة لـ DUBAI_POLICE (50/50)",
    "source": "DUBAI_POLICE",
    "current": 50,
    "limit": 50
  }
}
```

#### Release Stream
Release a stream reservation and free up quota.

```bash
DELETE /api/v1/stream/release/{reservation_id}

# Response (200 OK)
{
  "status": "released",
  "message": "Stream reservation released successfully"
}
```

#### Send Heartbeat
Keep the reservation alive (required every 30 seconds).

```bash
POST /api/v1/stream/heartbeat/{reservation_id}

# Response (200 OK)
{
  "status": "ok"
}
```

#### Get Stream Statistics
Retrieve real-time stream statistics.

```bash
GET /api/v1/stream/stats

# Response (200 OK)
{
  "active_streams": 45,
  "total_viewers": 123,
  "source_stats": {
    "DUBAI_POLICE": {
      "source": "DUBAI_POLICE",
      "current": 25,
      "limit": 50,
      "usage_percent": 50.0,
      "active_cameras": 25
    },
    "METRO": {
      "source": "METRO",
      "current": 15,
      "limit": 30,
      "usage_percent": 50.0,
      "active_cameras": 15
    }
  },
  "camera_stats": [
    {
      "camera_id": "uuid",
      "camera_name": "Camera 1",
      "viewer_count": 5,
      "source": "DUBAI_POLICE",
      "active_since": "2024-01-20T10:00:00Z"
    }
  ],
  "timestamp": "2024-01-20T11:00:00Z"
}
```

### Camera Management

#### List Cameras
```bash
GET /api/v1/cameras?source=DUBAI_POLICE&status=ONLINE&limit=100&offset=0

# Response (200 OK)
{
  "cameras": [
    {
      "id": "uuid",
      "name": "Camera 1 - Main Entrance",
      "name_ar": "كاميرا 1 - المدخل الرئيسي",
      "source": "DUBAI_POLICE",
      "rtsp_url": "rtsp://...",
      "status": "ONLINE",
      "ptz_enabled": true,
      "recording_server": "recorder-1",
      "metadata": {},
      "location": {
        "latitude": 25.2048,
        "longitude": 55.2708,
        "address": "Sheikh Zayed Road, Dubai"
      }
    }
  ],
  "count": 1
}
```

#### Get Camera
```bash
GET /api/v1/cameras/{camera_id}

# Response (200 OK)
{
  "id": "uuid",
  "name": "Camera 1",
  ...
}
```

#### Control PTZ
```bash
POST /api/v1/cameras/{camera_id}/ptz
Content-Type: application/json

{
  "command": "pan_left",     // pan_left, pan_right, tilt_up, tilt_down, zoom_in, zoom_out, preset, home
  "speed": 0.5,              // 0.0 - 1.0 (optional)
  "preset_id": 1,            // required for "preset" command
  "user_id": "user123"
}

# Response (200 OK)
{
  "status": "success",
  "message": "PTZ command executed"
}
```

### WebSocket

#### Stream Statistics (Real-time)
Connect to receive real-time stream statistics every 5 seconds.

```javascript
// Client-side JavaScript
const ws = new WebSocket('ws://localhost:8088/ws/stream/stats');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Message type:', message.type);
  console.log('Data:', message.data);
  console.log('Timestamp:', message.timestamp);
};

// Message types:
// - STREAM_STATS: Stream statistics update
// - CAMERA_STATUS: Camera status change
// - AGENCY_LIMIT_UPDATE: Agency limit update
// - ALERT: System alert
```

### Health & Metrics

```bash
GET /health        # Health check
GET /metrics       # Prometheus metrics
```

## Configuration

Environment variables:

```bash
# Service URLs
STREAM_COUNTER_URL=http://stream-counter:8087
VMS_SERVICE_URL=http://vms-service:8081
LIVEKIT_URL=ws://livekit:7880

# LiveKit credentials
LIVEKIT_API_KEY=your-api-key
LIVEKIT_API_SECRET=your-secret

# Valkey (Redis)
VALKEY_ADDR=valkey:6379
VALKEY_PASSWORD=
VALKEY_DB=0

# Service
PORT=8086
LOG_LEVEL=info
LOG_FORMAT=json
```

## Stream Request Flow

1. **Client** requests stream via `POST /api/v1/stream/reserve`
2. **Go API** validates camera exists (VMS Service)
3. **Go API** checks camera is online
4. **Go API** reserves quota (Stream Counter Service)
5. **Go API** creates LiveKit room (if doesn't exist)
6. **Go API** generates LiveKit JWT token
7. **Go API** saves reservation to Valkey
8. **Go API** returns token and room name to client
9. **Client** connects to LiveKit with token
10. **Client** sends heartbeat every 30s to keep reservation alive

## Integration with Other Services

### Stream Counter Service
- Reserve stream quota before granting access
- Release quota when stream ends
- Send heartbeats to keep reservation alive

### VMS Service
- Get camera details (name, source, status)
- Verify camera is online
- Control PTZ

### LiveKit
- Create rooms for cameras
- Generate access tokens for viewers
- List active rooms and participants

### Valkey (Redis)
- Store stream reservations (1 hour TTL)
- Track active streams per user
- Fast lookup for heartbeat validation

## Development

```bash
# Build
go build -o go-api ./cmd/main.go

# Run
./go-api

# Docker build
docker build -t cctv-go-api .

# Docker run
docker run -p 8086:8086 \
  -e LIVEKIT_URL=ws://livekit:7880 \
  -e VALKEY_ADDR=valkey:6379 \
  cctv-go-api
```

## Resource Usage

- **CPU**: 0.5-2 cores
- **Memory**: 512 MB - 1 GB
- **Network**: Minimal (JSON API only)

## Monitoring

### Prometheus Metrics

```bash
curl http://localhost:8086/metrics
```

Key metrics:
- `http_requests_total{method,endpoint,status}`
- `http_request_duration_seconds{method,endpoint}`
- `websocket_connections_total`
- `stream_reservations_total{source,status}`

## Troubleshooting

### Issue: Agency limit reached

**Error**: `AGENCY_LIMIT_EXCEEDED`

**Solution**:
- Check current quota usage: `GET /api/v1/stream/stats`
- Release unused streams
- Contact admin to increase agency limit

### Issue: LiveKit connection failed

**Error**: Failed to create LiveKit room

**Solution**:
- Check LiveKit service is running
- Verify LiveKit URL and credentials
- Check LiveKit logs: `docker logs cctv-livekit`

### Issue: Heartbeat timeout

**Error**: Reservation not found

**Solution**:
- Ensure heartbeats are sent every 30 seconds
- Check network connectivity
- Verify reservation ID is correct

## Security

- **JWT Tokens**: 1-hour expiration, scoped to specific room
- **CORS**: Configurable allowed origins
- **Quota Enforcement**: Atomic operations via Stream Counter
- **Audit Logging**: All stream requests logged with user ID

## Testing

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test -tags=integration ./...

# Load test
k6 run tests/load/stream-reserve.js
```

## Future Enhancements

- [ ] JWT authentication middleware
- [ ] Rate limiting per user
- [ ] Stream quality analytics
- [ ] Auto-scaling based on demand
- [ ] Multi-region LiveKit support
