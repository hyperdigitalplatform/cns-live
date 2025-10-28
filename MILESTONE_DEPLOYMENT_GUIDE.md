# Milestone XProtect Integration - Deployment Guide

**Version:** 1.0
**Date:** 2025-10-27
**Status:** Production Ready

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Database Migration](#database-migration)
4. [Service Configuration](#service-configuration)
5. [Kong Gateway Setup](#kong-gateway-setup)
6. [Frontend Integration](#frontend-integration)
7. [Deployment Steps](#deployment-steps)
8. [Verification & Testing](#verification--testing)
9. [Monitoring & Troubleshooting](#monitoring--troubleshooting)
10. [Rollback Procedure](#rollback-procedure)

---

## Prerequisites

### System Requirements
- Docker 20.10+ and Docker Compose 1.29+
- PostgreSQL 13+ with existing CNS database
- Kong Gateway 3.0+
- Milestone XProtect VMS 2020 R3+ (or compatible version)
- Node.js 18+ (for frontend build)
- Go 1.21+ (for backend services)

### Network Requirements
- Network connectivity between CNS services and Milestone server
- Firewall rules allowing HTTP/HTTPS traffic to Milestone Management Server
- Port access: Milestone API (typically port 80/443)

### Access Requirements
- Milestone API user account with permissions:
  - List cameras
  - Control recording
  - Query recording sequences
  - Access video streams
- Database admin access for migrations
- Docker deployment permissions

---

## Environment Setup

### 1. Create Environment File

Create `.env.milestone` file in the project root:

```bash
# Milestone XProtect Server Configuration
MILESTONE_BASE_URL=http://10.0.1.100:80
MILESTONE_USERNAME=api_user
MILESTONE_PASSWORD=YOUR_SECURE_PASSWORD_HERE
MILESTONE_AUTH_TYPE=basic
MILESTONE_SESSION_TIMEOUT=3600
MILESTONE_RETRY_ATTEMPTS=3
MILESTONE_RETRY_DELAY=5s
MILESTONE_MAX_CONNECTIONS=50

# Recording Configuration
RECORDING_DEFAULT_DURATION=900
RECORDING_MAX_DURATION=7200
RECORDING_MIN_DURATION=60
RECORDING_AUTO_STOP=true
RECORDING_NOTIFICATIONS=true
RECORDING_MAX_CONCURRENT=50

# Playback Configuration
PLAYBACK_CACHE_ENABLED=true
PLAYBACK_CACHE_TTL=300
PLAYBACK_MAX_CONCURRENT_STREAMS=50
PLAYBACK_MAX_SPEED=8
PLAYBACK_MIN_SPEED=-8

# HLS Streaming
HLS_SEGMENT_DURATION=4
HLS_PLAYLIST_LENGTH=5
HLS_OUTPUT_PATH=/tmp/hls

# FFmpeg Configuration
FFMPEG_THREADS=4
FFMPEG_VIDEO_CODEC=libx264
FFMPEG_AUDIO_CODEC=aac
FFMPEG_PRESET=fast

# Feature Flags
MILESTONE_ENABLED=true
MILESTONE_AUTO_SYNC=false
MILESTONE_SYNC_INTERVAL=3600

# Service URLs (internal)
VMS_SERVICE_URL=http://vms-service:8081
RECORDING_SERVICE_URL=http://recording-service:8083
PLAYBACK_SERVICE_URL=http://playback-service:8084
```

### 2. Secure Environment Variables

```bash
# Encrypt sensitive values (recommended)
# Option 1: Use Docker secrets
echo "YOUR_MILESTONE_PASSWORD" | docker secret create milestone_password -

# Option 2: Use HashiCorp Vault
vault kv put secret/cns/milestone password=YOUR_MILESTONE_PASSWORD

# Option 3: Use encrypted .env file
ansible-vault encrypt .env.milestone
```

---

## Database Migration

### 1. Backup Current Database

```bash
# Create backup
docker exec cctv-postgres pg_dump -U postgres cctv > backup_$(date +%Y%m%d_%H%M%S).sql

# Verify backup
ls -lh backup_*.sql
```

### 2. Run Migrations

```bash
# Navigate to vms-service
cd services/vms-service

# Check migration status
migrate -path migrations -database "postgres://postgres:password@localhost:5432/cctv?sslmode=disable" version

# Run up migrations
migrate -path migrations -database "postgres://postgres:password@localhost:5432/cctv?sslmode=disable" up

# Verify migration
psql -U postgres -d cctv -c "\dt milestone*"
```

Expected tables:
- `milestone_recording_sessions`
- `milestone_sync_history`
- `milestone_playback_cache`

### 3. Verify Schema

```bash
# Check cameras table has new columns
psql -U postgres -d cctv -c "\d cameras" | grep milestone

# Should show:
# milestone_device_id    | character varying(255)
# milestone_server       | character varying(255)
# last_milestone_sync    | timestamp without time zone
# milestone_metadata     | jsonb
```

---

## Service Configuration

### 1. VMS Service

Update `services/vms-service/cmd/main.go`:

```go
import (
    "github.com/rta/cctv/vms-service/internal/client"
    milestoneHttp "github.com/rta/cctv/vms-service/internal/delivery/http"
)

// Initialize Milestone client
milestoneConfig := client.MilestoneConfig{
    BaseURL:        os.Getenv("MILESTONE_BASE_URL"),
    Username:       os.Getenv("MILESTONE_USERNAME"),
    Password:       os.Getenv("MILESTONE_PASSWORD"),
    AuthType:       os.Getenv("MILESTONE_AUTH_TYPE"),
    SessionTimeout: 1 * time.Hour,
}

milestoneClient := client.NewMilestoneClient(milestoneConfig, logger)

// Login to Milestone
if err := milestoneClient.Login(context.Background()); err != nil {
    logger.Fatal().Err(err).Msg("Failed to connect to Milestone")
}

// Create handler
milestoneHandler := milestoneHttp.NewMilestoneHandler(milestoneClient, cameraRepo, logger)

// Register routes
r.Get("/vms/milestone/cameras", milestoneHandler.ListMilestoneCameras)
r.Get("/vms/milestone/cameras/{id}", milestoneHandler.GetMilestoneCamera)
r.Post("/vms/cameras/import", milestoneHandler.ImportCamera)
r.Post("/vms/milestone/sync-all", milestoneHandler.BulkSyncCameras)
r.Put("/vms/cameras/{id}/sync", milestoneHandler.SyncCameraWithMilestone)
```

### 2. Recording Service

Update `services/recording-service/cmd/main.go`:

```go
import (
    "github.com/rta/cctv/recording-service/internal/manager"
    recordingHttp "github.com/rta/cctv/recording-service/internal/delivery/http"
)

// Initialize Milestone client (same as VMS service)
milestoneClient := client.NewMilestoneClient(milestoneConfig, logger)

// Create recording manager
recordingManager := manager.NewMilestoneRecordingManager(milestoneClient, logger)

// Restore active sessions on startup
cameraIDs := []string{} // Load from database
recordingManager.RestoreActiveSessions(context.Background(), cameraIDs)

// Graceful shutdown
defer recordingManager.Shutdown(context.Background())

// Create handler
recordingHandler := recordingHttp.NewMilestoneRecordingHandler(recordingManager, logger)

// Register routes
r.Post("/recordings/cameras/{cameraId}/start", recordingHandler.StartRecording)
r.Post("/recordings/cameras/{cameraId}/stop", recordingHandler.StopRecording)
r.Get("/recordings/cameras/{cameraId}/status", recordingHandler.GetRecordingStatus)
r.Get("/recordings/active", recordingHandler.GetActiveRecordings)
```

### 3. Playback Service

Update `services/playback-service/cmd/main.go`:

```go
import (
    "github.com/rta/cctv/playback-service/internal/usecase"
    playbackHttp "github.com/rta/cctv/playback-service/internal/delivery/http"
)

// Initialize Milestone client
milestoneClient := client.NewMilestoneClient(milestoneConfig, logger)

// Create cache (simple in-memory cache)
cache := NewSimpleCache()

// Create playback usecase
playbackUsecase := usecase.NewMilestonePlaybackUsecase(milestoneClient, cache, logger)

// Create handler
playbackHandler := playbackHttp.NewMilestonePlaybackHandler(playbackUsecase, logger)

// Register routes
r.Post("/playback/cameras/{cameraId}/query", playbackHandler.QueryRecordings)
r.Get("/playback/cameras/{cameraId}/timeline", playbackHandler.GetTimelineData)
r.Post("/playback/cameras/{cameraId}/start", playbackHandler.StartPlayback)
r.Get("/playback/cameras/{cameraId}/stream", playbackHandler.StreamPlayback)
r.Get("/playback/cameras/{cameraId}/snapshot", playbackHandler.GetSnapshot)
r.Post("/playback/cameras/{cameraId}/control", playbackHandler.ControlPlayback)
```

---

## Kong Gateway Setup

### 1. Add Milestone Routes

```bash
# Merge milestone routes with main Kong config
cd config/kong

# Option 1: Include in main kong.yml
cat milestone-routes.yml >> kong.yml

# Option 2: Load separately
deck sync -s kong.yml
deck sync -s milestone-routes.yml
```

### 2. Verify Kong Configuration

```bash
# Check Kong configuration
curl -i http://localhost:8001/services | jq '.data[] | select(.name | contains("milestone"))'

# Check routes
curl -i http://localhost:8001/routes | jq '.data[] | select(.name | contains("milestone"))'
```

### 3. Test Kong Proxy

```bash
# Test camera discovery through Kong
curl -X GET "http://localhost:8000/api/v1/milestone/cameras?limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Test recording start
curl -X POST "http://localhost:8000/api/v1/cameras/CAMERA_ID/recordings/start" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"duration": 300}'
```

---

## Frontend Integration

### 1. Install Dependencies

```bash
cd dashboard

# Install HLS.js for video playback
npm install hls.js

# Install date-time utilities (if not already installed)
npm install date-fns
```

### 2. Update Sidebar Component

Edit `dashboard/src/components/CameraSidebarNew.tsx`:

```typescript
import { CameraSidebarRecordingSection } from './CameraSidebarRecordingSection';

// Add to sidebar after camera tree view
<CameraSidebarRecordingSection selectedCamera={selectedCamera} />
```

### 3. Build Frontend

```bash
cd dashboard
npm run build

# Verify build output
ls -la dist/
```

### 4. Update Frontend Configuration

Edit `dashboard/src/config/api.ts`:

```typescript
export const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8000';

export const API_ENDPOINTS = {
  // Existing endpoints...

  // Milestone endpoints
  milestone: {
    cameras: '/api/v1/milestone/cameras',
    import: '/api/v1/cameras/import',
    sync: '/api/v1/milestone/sync-all',
  },
  recordings: {
    start: (cameraId: string) => `/api/v1/cameras/${cameraId}/recordings/start`,
    stop: (cameraId: string) => `/api/v1/cameras/${cameraId}/recordings/stop`,
    status: (cameraId: string) => `/api/v1/cameras/${cameraId}/recordings/status`,
    query: (cameraId: string) => `/api/v1/cameras/${cameraId}/recordings/query`,
  },
  playback: {
    stream: (cameraId: string) => `/api/v1/cameras/${cameraId}/playback/stream`,
    snapshot: (cameraId: string) => `/api/v1/cameras/${cameraId}/playback/snapshot`,
  },
};
```

---

## Deployment Steps

### Step 1: Pre-Deployment Checklist

- [ ] Milestone server is accessible from Docker network
- [ ] Database backup completed
- [ ] Environment variables configured
- [ ] Kong gateway is running
- [ ] Frontend build completed
- [ ] All services build successfully

### Step 2: Deploy Services

```bash
# Stop existing services
docker-compose down

# Pull latest images (if using registry)
docker-compose pull

# Build services with Milestone integration
docker-compose -f docker-compose.yml -f docker-compose.milestone.yml build

# Start services
docker-compose -f docker-compose.yml -f docker-compose.milestone.yml up -d

# Check service status
docker-compose ps
```

### Step 3: Run Migrations

```bash
# Migrations run automatically on service startup
# Verify migrations
docker-compose logs vms-service | grep migration
```

### Step 4: Verify Service Health

```bash
# Check each service
curl http://localhost:8081/health  # VMS Service
curl http://localhost:8083/health  # Recording Service
curl http://localhost:8084/health  # Playback Service

# Check Milestone connection
docker-compose logs vms-service | grep "Milestone"
```

### Step 5: Test Integration

```bash
# Test camera discovery
curl http://localhost:8081/vms/milestone/cameras | jq '.'

# Test recording status
curl http://localhost:8083/recordings/cameras/CAMERA_ID/status | jq '.'
```

---

## Verification & Testing

### Functional Tests

1. **Camera Discovery**
```bash
# List Milestone cameras
curl -X GET "http://localhost:8000/api/v1/milestone/cameras?limit=5"

# Expected: JSON array of cameras with import status
```

2. **Camera Import**
```bash
# Import a camera
curl -X POST "http://localhost:8000/api/v1/cameras/import" \
  -H "Content-Type: application/json" \
  -d '{
    "milestoneDeviceId": "MILESTONE_CAMERA_ID",
    "name": "Test Camera",
    "source": "DUBAI_POLICE"
  }'

# Expected: 201 Created with camera details
```

3. **Start Recording**
```bash
# Start 5-minute recording
curl -X POST "http://localhost:8000/api/v1/cameras/CAMERA_ID/recordings/start" \
  -H "Content-Type: application/json" \
  -d '{"duration": 300}'

# Expected: 200 OK with recording session details
```

4. **Query Recordings**
```bash
# Query last 24 hours
curl -X POST "http://localhost:8000/api/v1/cameras/CAMERA_ID/recordings/query" \
  -H "Content-Type: application/json" \
  -d '{
    "startTime": "2025-10-26T00:00:00Z",
    "endTime": "2025-10-27T00:00:00Z"
  }'

# Expected: Timeline data with sequences and gaps
```

### Frontend Tests

1. **Open Dashboard**
   - Navigate to `http://localhost:3000`
   - Login with credentials

2. **Test Camera Discovery**
   - Click "Import from Milestone" button
   - Verify camera list loads
   - Select and import a camera
   - Check camera appears in sidebar

3. **Test Recording Control**
   - Select imported camera
   - Click "Start Recording"
   - Verify countdown timer appears
   - Wait or stop manually
   - Verify recording stops

4. **Test Playback**
   - Click "View Recordings"
   - Select date range
   - Verify timeline displays
   - Click on timeline to seek
   - Verify video plays

---

## Monitoring & Troubleshooting

### Log Monitoring

```bash
# Watch all Milestone-related logs
docker-compose logs -f vms-service recording-service playback-service | grep -i milestone

# Check for errors
docker-compose logs --tail=100 vms-service | grep -i error
docker-compose logs --tail=100 recording-service | grep -i error
docker-compose logs --tail=100 playback-service | grep -i error
```

### Common Issues

#### 1. Connection to Milestone Failed

**Symptoms:**
- "Failed to connect to Milestone" in logs
- 500 errors on discovery endpoints

**Solutions:**
```bash
# Check network connectivity
docker exec cctv-vms-service ping milestone-server

# Check Milestone credentials
docker exec cctv-vms-service env | grep MILESTONE

# Test Milestone API directly
curl -u username:password http://milestone-server/api/rest/v1/cameras
```

#### 2. Recording Not Starting

**Symptoms:**
- Start recording returns error
- Recording status shows not recording

**Solutions:**
```bash
# Check recording service logs
docker-compose logs recording-service | tail -50

# Verify camera has Milestone device ID
psql -U postgres -d cctv -c "SELECT id, name, milestone_device_id FROM cameras WHERE id='CAMERA_ID';"

# Test Milestone recording API directly
curl -X POST http://milestone-server/api/rest/v1/cameras/DEVICE_ID/recordings/start
```

#### 3. Timeline Not Loading

**Symptoms:**
- Query recordings returns empty
- Frontend shows "No recordings"

**Solutions:**
```bash
# Check playback service
docker-compose logs playback-service | grep query

# Verify time range
# Milestone may not have recordings for requested period

# Check cache
psql -U postgres -d cctv -c "SELECT * FROM milestone_playback_cache;"
```

### Health Checks

```bash
# Create health check script
cat > health-check.sh << 'EOF'
#!/bin/bash

echo "=== CNS Milestone Integration Health Check ==="

# VMS Service
echo -n "VMS Service: "
curl -sf http://localhost:8081/health > /dev/null && echo "✓ OK" || echo "✗ FAIL"

# Recording Service
echo -n "Recording Service: "
curl -sf http://localhost:8083/health > /dev/null && echo "✓ OK" || echo "✗ FAIL"

# Playback Service
echo -n "Playback Service: "
curl -sf http://localhost:8084/health > /dev/null && echo "✓ OK" || echo "✗ FAIL"

# Milestone Connection
echo -n "Milestone Connection: "
docker-compose logs vms-service | grep -q "Successfully logged into Milestone" && echo "✓ OK" || echo "✗ FAIL"

# Database Tables
echo -n "Database Migrations: "
psql -U postgres -d cctv -tc "SELECT COUNT(*) FROM milestone_recording_sessions" > /dev/null 2>&1 && echo "✓ OK" || echo "✗ FAIL"

echo "=== Health Check Complete ==="
EOF

chmod +x health-check.sh
./health-check.sh
```

---

## Rollback Procedure

### If Issues Occur

1. **Stop Services**
```bash
docker-compose down
```

2. **Restore Database**
```bash
# Restore from backup
docker exec -i cctv-postgres psql -U postgres -d cctv < backup_YYYYMMDD_HHMMSS.sql
```

3. **Revert Docker Compose**
```bash
# Start without Milestone integration
docker-compose up -d
```

4. **Rollback Database Migration**
```bash
cd services/vms-service
migrate -path migrations -database "postgres://..." down 1
```

---

## Post-Deployment Tasks

### 1. Performance Tuning

- Monitor query response times
- Adjust cache TTL based on usage
- Scale services if needed

### 2. Security Hardening

- Enable JWT authentication
- Configure HTTPS/TLS
- Set up IP whitelisting
- Review and tighten rate limits

### 3. Documentation

- Document custom configurations
- Update team on new features
- Create user guides

### 4. Monitoring Setup

- Configure Prometheus metrics
- Set up Grafana dashboards
- Create alerts for failures

---

## Support & Maintenance

### Regular Maintenance

- **Daily:** Monitor logs for errors
- **Weekly:** Review active recordings, check disk usage
- **Monthly:** Database cleanup, performance review
- **Quarterly:** Security audit, dependency updates

### Contact & Resources

- **Documentation:** `MILESTONE_INTEGRATION_PLAN.md`
- **API Reference:** `http://localhost:8001/docs`
- **Milestone Docs:** https://www.milestonesys.com/developers/

---

**Deployment Guide Version:** 1.0
**Last Updated:** 2025-10-27
**Status:** ✅ Complete and Ready for Production
