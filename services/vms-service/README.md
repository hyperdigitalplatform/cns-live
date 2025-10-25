# VMS Service

Milestone XProtect VMS Integration Service for RTA CCTV System.

## **Overview**

This service provides a unified interface to Milestone XProtect Video Management System, handling:
- Camera discovery and metadata retrieval
- RTSP URL generation for live streaming
- PTZ (Pan-Tilt-Zoom) control
- Recording segment queries
- Video export functionality

## **Features**

- ✅ Connection pool management for Milestone Recording Servers
- ✅ In-memory caching (5-minute TTL) for performance
- ✅ Background synchronization every 10 minutes
- ✅ RESTful API with OpenAPI compatibility
- ✅ Prometheus metrics export
- ✅ Health check endpoint
- ✅ Graceful shutdown
- ✅ Low resource footprint (<256MB RAM, 0.5 CPU)

## **API Endpoints**

### **Cameras**

```
GET  /vms/cameras                  - List all cameras
GET  /vms/cameras?source=METRO     - Filter cameras by source
GET  /vms/cameras/{id}             - Get camera details
GET  /vms/cameras/{id}/stream      - Get RTSP URL for streaming
POST /vms/cameras/{id}/ptz         - Execute PTZ command
```

### **Recordings**

```
GET  /vms/recordings/{camera_id}/segments?start={iso8601}&end={iso8601}  - Get recording segments
POST /vms/recordings/export                                               - Create export job
GET  /vms/recordings/export/{export_id}                                   - Get export status
```

### **System**

```
GET /health    - Health check
GET /metrics   - Prometheus metrics
```

## **Environment Variables**

```bash
# Milestone VMS Configuration
MILESTONE_SERVER=milestone.rta.ae:554
MILESTONE_USER=vms_service_account
MILESTONE_PASS=your_password
MILESTONE_AUTH_TYPE=WindowsDefault

# Service Configuration
PORT=8081
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=json         # json, text

# Cache Configuration
CACHE_TTL=5m            # Cache time-to-live
CACHE_CLEANUP=10m       # Cache cleanup interval
```

## **Quick Start**

### **Development**

```bash
# Install dependencies
go mod download

# Run service
go run cmd/main.go

# Run with environment file
cp .env.example .env
# Edit .env with your configuration
go run cmd/main.go
```

### **Docker**

```bash
# Build image
docker build -t rta/vms-service:latest .

# Run container
docker run -d \
  --name vms-service \
  -p 8081:8081 \
  -e MILESTONE_SERVER=milestone:554 \
  -e MILESTONE_USER=admin \
  -e MILESTONE_PASS=password \
  rta/vms-service:latest

# View logs
docker logs -f vms-service

# Check health
curl http://localhost:8081/health
```

## **API Examples**

### **List All Cameras**

```bash
curl http://localhost:8081/vms/cameras
```

**Response:**
```json
{
  "cameras": [
    {
      "id": "uuid",
      "name": "Camera 001 - Sheikh Zayed Road",
      "name_ar": "كاميرا 001 - شارع الشيخ زايد",
      "source": "DUBAI_POLICE",
      "rtsp_url": "rtsp://milestone:554/camera_001",
      "ptz_enabled": true,
      "status": "ONLINE",
      "recording_server": "milestone:554"
    }
  ],
  "total": 100,
  "last_updated": "2024-01-01T10:00:00Z"
}
```

### **Get Camera Stream URL**

```bash
curl http://localhost:8081/vms/cameras/{id}/stream
```

**Response:**
```json
{
  "camera_id": "uuid",
  "rtsp_url": "rtsp://milestone:554/camera_001",
  "transport": "tcp"
}
```

### **Execute PTZ Command**

```bash
curl -X POST http://localhost:8081/vms/cameras/{id}/ptz \
  -H "Content-Type: application/json" \
  -d '{
    "action": "MOVE",
    "pan": 0.5,
    "tilt": 0.3,
    "zoom": 0.8,
    "speed": 0.5
  }'
```

### **Get Recording Segments**

```bash
curl "http://localhost:8081/vms/recordings/{camera_id}/segments?start=2024-01-01T00:00:00Z&end=2024-01-01T23:59:59Z"
```

**Response:**
```json
{
  "camera_id": "uuid",
  "start": "2024-01-01T00:00:00Z",
  "end": "2024-01-01T23:59:59Z",
  "segments": [
    {
      "start_time": "2024-01-01T00:00:00Z",
      "end_time": "2024-01-01T01:00:00Z",
      "available": true,
      "size_bytes": 524288000
    }
  ],
  "total": 24
}
```

## **Architecture**

```
┌─────────────────────────────────────────┐
│         VMS Service (Go)                │
├─────────────────────────────────────────┤
│  HTTP API Layer                         │
│  ├── Camera Endpoints                   │
│  ├── Recording Endpoints                │
│  └── PTZ Endpoints                      │
├─────────────────────────────────────────┤
│  Business Logic                         │
│  ├── In-Memory Cache (5min TTL)        │
│  └── Background Sync (10min interval)   │
├─────────────────────────────────────────┤
│  Milestone Integration Layer            │
│  ├── Connection Pool (5 per server)    │
│  ├── SDK Wrapper                        │
│  └── Error Handling                     │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│    Milestone XProtect VMS               │
│    Recording Servers                    │
└─────────────────────────────────────────┘
```

## **Metrics**

Prometheus metrics available at `/metrics`:

```
# Connection status
vms_milestone_connected{server} gauge

# Camera counts
vms_cameras_total{source} gauge
vms_cameras_online{source} gauge

# API requests
vms_api_requests_total{method,endpoint,status} counter
vms_api_request_duration_seconds{method,endpoint} histogram

# Cache statistics
vms_cache_hits_total counter
vms_cache_misses_total counter
vms_cache_size gauge
```

## **Development**

### **Project Structure**

```
vms-service/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── domain/
│   │   └── camera.go              # Business entities
│   ├── repository/
│   │   ├── repository.go          # Interfaces
│   │   ├── milestone/
│   │   │   └── milestone_repository.go  # Milestone integration
│   │   └── cache/
│   │       └── memory_cache.go    # In-memory cache
│   └── delivery/
│       └── http/
│           ├── handler.go         # HTTP handlers
│           └── router.go          # Route definitions
├── Dockerfile
├── go.mod
└── README.md
```

### **Running Tests**

```bash
# Unit tests
go test ./... -v

# With coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### **Building**

```bash
# Build binary
go build -o vms-service cmd/main.go

# Build with optimizations
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags='-w -s' \
  -o vms-service \
  cmd/main.go
```

## **Troubleshooting**

### **Connection Issues**

```bash
# Check Milestone connectivity
telnet milestone.rta.ae 554

# View logs
docker logs vms-service

# Check health endpoint
curl http://localhost:8081/health
```

### **Cache Issues**

```bash
# View cache statistics
curl http://localhost:8081/metrics | grep vms_cache
```

### **Performance Issues**

```bash
# Check resource usage
docker stats vms-service

# Enable debug logging
docker run -e LOG_LEVEL=debug rta/vms-service:latest
```

## **Integration with Other Services**

### **MediaMTX Integration**

VMS Service provides RTSP URLs that MediaMTX uses for ingestion:

```bash
# Get RTSP URL
RTSP_URL=$(curl -s http://vms-service:8081/vms/cameras/{id}/stream | jq -r '.rtsp_url')

# MediaMTX uses this URL for streaming
```

### **Go API Integration**

Go API calls VMS Service for camera operations:

```bash
# List cameras for streaming
curl http://vms-service:8081/vms/cameras?source=DUBAI_POLICE

# Get PTZ capabilities before allowing control
curl http://vms-service:8081/vms/cameras/{id}
```

## **Production Considerations**

1. **Security**:
   - Store Milestone credentials in secrets management (e.g., Kubernetes Secrets)
   - Use TLS for Milestone connections
   - Implement authentication middleware

2. **Performance**:
   - Adjust cache TTL based on camera change frequency
   - Scale horizontally by running multiple instances
   - Use connection pooling efficiently

3. **Monitoring**:
   - Set up Prometheus alerts for connection failures
   - Monitor cache hit ratio
   - Track API latency

4. **Backup**:
   - No persistent data to backup (uses Milestone as source of truth)
   - Document Milestone server configuration

## **License**

Proprietary - Roads and Transport Authority (RTA)

## **Support**

- **Issues**: [GitHub Issues](https://github.com/rta/cctv/issues)
- **Email**: vms-support@rta.ae
