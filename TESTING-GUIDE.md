# RTA CCTV System - Testing & Verification Guide

**Version**: 1.0.0
**Last Updated**: October 24, 2025
**Purpose**: Complete testing guide to start and verify the full solution

---

## Table of Contents

1. [Quick Start Test (10 minutes)](#quick-start-test-10-minutes)
2. [Full System Verification (30 minutes)](#full-system-verification-30-minutes)
3. [Service-by-Service Testing](#service-by-service-testing)
4. [Monitoring Stack Verification](#monitoring-stack-verification)
5. [Functional Feature Testing](#functional-feature-testing)
6. [Performance Testing](#performance-testing)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before starting, ensure you have:

- ✅ Docker 24.0+ installed
- ✅ Docker Compose 2.20+ installed
- ✅ 16GB RAM available
- ✅ 50GB disk space available
- ✅ Ports 3000-9093 available (no conflicts)

Check prerequisites:
```bash
docker --version                    # Should be 24.0+
docker compose version              # Should be 2.20+
free -h                             # Check available RAM
df -h                               # Check disk space
```

### Database Setup

The PostgreSQL database is **automatically created** when using Docker Compose. See [`database/DATABASE-SETUP.md`](database/DATABASE-SETUP.md) for:
- Automatic setup with Docker Compose (recommended)
- Manual setup for standalone PostgreSQL
- Database schema details
- Migration instructions
- Backup & restore procedures
- Troubleshooting guide

---

## Quick Start Test (10 minutes)

This is the fastest way to verify the platform works.

### Step 1: Clone and Configure (2 minutes)

```bash
# Clone repository (if not already done)
git clone <repository-url>
cd cns

# Copy environment template
cp .env.example .env

# Optional: Edit passwords (or use defaults for testing)
nano .env
```

**Minimal `.env` for testing** (use defaults from `.env.example`):
```bash
POSTGRES_PASSWORD=test_postgres_pass
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=test_minio_pass
LIVEKIT_API_KEY=testkey
LIVEKIT_API_SECRET=testsecret
GRAFANA_ADMIN_PASSWORD=admin_test
```

### Step 2: Start All Services (3 minutes)

```bash
# Start all 28 services in background
docker-compose up -d

# Wait for services to initialize (2 minutes)
echo "Waiting for services to start..."
sleep 120

# Run database migrations (first time only)
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/001_create_initial_schema.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/002_create_storage_tables.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/003_create_metadata_tables.sql

# Check all containers are running
docker-compose ps
```

**Expected output**: All services should show "Up" status (28 containers).

**Note**: Database migrations create all required tables. See [`database/DATABASE-SETUP.md`](database/DATABASE-SETUP.md) for details.

### Step 3: Run Automated Health Check (1 minute)

```bash
# Run comprehensive health check
./scripts/health-check.sh
```

**Expected output**:
```
===========================================
RTA CCTV SYSTEM - HEALTH CHECK
===========================================

[1/3] DOCKER CONTAINERS
✓ OK   cctv-valkey (Up)
✓ OK   cctv-postgres (Up)
✓ OK   cctv-vms-service (Up)
✓ OK   cctv-storage-service (Up)
✓ OK   cctv-recording-service (Up)
✓ OK   cctv-metadata-service (Up)
✓ OK   cctv-playback-service (Up)
✓ OK   cctv-stream-counter (Up)
✓ OK   cctv-go-api (Up)
✓ OK   cctv-dashboard (Up)
✓ OK   cctv-minio (Up)
✓ OK   cctv-livekit (Up)
✓ OK   cctv-prometheus (Up)
✓ OK   cctv-grafana (Up)
✓ OK   cctv-loki (Up)
✓ OK   cctv-alertmanager (Up)

[2/3] HTTP ENDPOINTS
✓ OK   Go API (http://localhost:8088/health)
✓ OK   VMS Service (http://localhost:8081/health)
✓ OK   Storage Service (http://localhost:8082/health)
✓ OK   Recording Service (http://localhost:8083/health)
✓ OK   Metadata Service (http://localhost:8084/health)
✓ OK   Playback Service (http://localhost:8090/health)
✓ OK   Stream Counter (http://localhost:8087/health)
✓ OK   Dashboard (http://localhost:3000)
✓ OK   Grafana (http://localhost:3001)
✓ OK   Prometheus (http://localhost:9090)
✓ OK   MinIO (http://localhost:9001)
✓ OK   Loki (http://localhost:3100/ready)
✓ OK   Alertmanager (http://localhost:9093/-/healthy)

[3/3] PROMETHEUS TARGETS
✓ OK   All Prometheus targets are up (13/13)

===========================================
HEALTH CHECK PASSED ✅
All 16 containers running, all 13 endpoints healthy
===========================================
```

### Step 4: Access Web Interfaces (2 minutes)

```bash
# Open dashboard
open http://localhost:3000

# Open monitoring (Grafana)
open http://localhost:3001

# Open MinIO console
open http://localhost:9001
```

**Manual verification**:
- ✅ **Dashboard** loads without errors
- ✅ **Grafana** shows login page (login: `admin` / password from `.env`)
- ✅ **MinIO Console** shows login page

### Step 5: Quick API Test (2 minutes)

```bash
# Test Go API health
curl -s http://localhost:8088/health | jq

# Expected output:
{
  "status": "healthy",
  "timestamp": "2025-10-24T...",
  "services": {
    "database": "up",
    "cache": "up",
    "storage": "up"
  }
}

# Test Prometheus targets
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[].health' | grep -c up

# Expected output: 13 (all targets up)
```

### ✅ Quick Test Complete

If all steps passed:
- ✅ All 28 services are running
- ✅ All health checks passed
- ✅ Web interfaces are accessible
- ✅ APIs are responding

**Result**: Platform is operational! Proceed to full verification for detailed testing.

---

## Full System Verification (30 minutes)

Comprehensive verification of all components.

### Part 1: Infrastructure Layer (5 minutes)

#### 1.1 Database (PostgreSQL)

```bash
# Check PostgreSQL is ready
docker exec cctv-postgres pg_isready -U cctv

# Expected: accepting connections

# Verify database exists
docker exec cctv-postgres psql -U cctv -c "\l"

# Expected: "cctv" database listed

# Check tables (from migrations)
docker exec cctv-postgres psql -U cctv -d cctv -c "\dt"

# Expected: cameras, video_segments, incidents, tags, annotations, etc.

# Test query
docker exec cctv-postgres psql -U cctv -d cctv -c "SELECT COUNT(*) FROM cameras;"

# Expected: Row count (0 or more)
```

**Status**: ✅ PostgreSQL operational

#### 1.2 Cache (Valkey)

```bash
# Check Valkey connection
docker exec cctv-valkey valkey-cli ping

# Expected: PONG

# Check memory usage
docker exec cctv-valkey valkey-cli INFO memory | grep used_memory_human

# Expected: Memory usage shown

# Test set/get
docker exec cctv-valkey valkey-cli SET test_key "test_value"
docker exec cctv-valkey valkey-cli GET test_key

# Expected: "test_value"

# Cleanup
docker exec cctv-valkey valkey-cli DEL test_key
```

**Status**: ✅ Valkey cache operational

#### 1.3 Storage (MinIO)

```bash
# Check MinIO health
curl -s http://localhost:9000/minio/health/live

# Expected: Empty response (200 OK)

# Access MinIO console
open http://localhost:9001
# Login: admin / <MINIO_ROOT_PASSWORD from .env>

# Using mc client (if installed)
docker exec cctv-minio mc alias set local http://localhost:9000 admin <password>
docker exec cctv-minio mc ls local/

# Expected: List of buckets (cctv-recordings, cctv-exports, cctv-thumbnails, cctv-clips)
```

**Status**: ✅ MinIO storage operational

---

### Part 2: Core Services (10 minutes)

#### 2.1 VMS Service (Port 8081)

```bash
# Check health
curl -s http://localhost:8081/health | jq

# Expected:
{
  "status": "healthy",
  "milestone_connection": "ready"
}

# Check metrics
curl -s http://localhost:8081/metrics | grep vms_

# Expected: Prometheus metrics (vms_cameras_total, etc.)

# View logs
docker-compose logs vms-service | tail -20

# Expected: No errors, service started successfully
```

**Status**: ✅ VMS Service operational

#### 2.2 Storage Service (Port 8082)

```bash
# Check health
curl -s http://localhost:8082/health | jq

# Check storage status
curl -s http://localhost:8082/api/v1/storage/status | jq

# Expected:
{
  "mode": "BOTH",
  "backends": {
    "minio": "available",
    "milestone": "available"
  }
}

# View logs
docker-compose logs storage-service | tail -20
```

**Status**: ✅ Storage Service operational

#### 2.3 Recording Service (Port 8083)

```bash
# Check health
curl -s http://localhost:8083/health | jq

# Check recording stats
curl -s http://localhost:8083/api/v1/recordings/stats | jq

# Expected:
{
  "active_recordings": 0,
  "total_segments": 0,
  "storage_used_gb": 0
}

# View logs
docker-compose logs recording-service | tail -20
```

**Status**: ✅ Recording Service operational

#### 2.4 Metadata Service (Port 8084)

```bash
# Check health
curl -s http://localhost:8084/health | jq

# Test search (empty result expected if no data)
curl -s http://localhost:8084/api/v1/search?q=test | jq

# Check tags
curl -s http://localhost:8084/api/v1/tags | jq

# View logs
docker-compose logs metadata-service | tail -20
```

**Status**: ✅ Metadata Service operational

#### 2.5 Playback Service (Port 8090)

```bash
# Check health
curl -s http://localhost:8090/health | jq

# Check cache stats
curl -s http://localhost:8090/api/v1/cache/stats | jq

# Expected:
{
  "cache_size_gb": 50,
  "cache_used_gb": 0,
  "cache_hit_rate": 0
}

# View logs
docker-compose logs playback-service | tail -20
```

**Status**: ✅ Playback Service operational

#### 2.6 Stream Counter (Port 8087)

```bash
# Check health
curl -s http://localhost:8087/health | jq

# Check stream limits
curl -s http://localhost:8087/api/v1/limits | jq

# Expected:
{
  "dubai_police": {"limit": 50, "active": 0},
  "metro": {"limit": 30, "active": 0},
  "bus": {"limit": 20, "active": 0},
  "other": {"limit": 400, "active": 0},
  "total": {"limit": 500, "active": 0}
}

# View logs
docker-compose logs stream-counter | tail -20
```

**Status**: ✅ Stream Counter operational

#### 2.7 Go API (Port 8088)

```bash
# Check health
curl -s http://localhost:8088/health | jq

# Check version/info
curl -s http://localhost:8088/api/v1/info | jq

# List cameras (empty if no data)
curl -s http://localhost:8088/api/v1/cameras | jq

# Check WebSocket endpoint
curl -i http://localhost:8088/ws

# Expected: 400 or connection upgrade message

# View logs
docker-compose logs go-api | tail -20
```

**Status**: ✅ Go API operational

---

### Part 3: Frontend & Streaming (5 minutes)

#### 3.1 Dashboard (Port 3000)

```bash
# Check if dashboard is serving
curl -s http://localhost:3000 | grep -c "<!doctype html>"

# Expected: 1 (HTML served)

# Open in browser
open http://localhost:3000

# Manual verification:
# ✅ Dashboard loads
# ✅ Camera grid visible
# ✅ Navigation menu works
# ✅ No console errors (F12)

# View logs
docker-compose logs dashboard | tail -20
```

**Status**: ✅ Dashboard operational

#### 3.2 LiveKit (Port 7880)

```bash
# Check LiveKit health
curl -s http://localhost:7880 | grep -c "LiveKit"

# Expected: 1 (LiveKit response)

# Check LiveKit rooms (requires API key)
# This will fail authentication but confirms service is up
curl -i http://localhost:7880/rooms

# Expected: 401 Unauthorized (service responding)

# View logs
docker-compose logs livekit | tail -20

# Expected: LiveKit server started, WebRTC endpoints ready
```

**Status**: ✅ LiveKit operational

#### 3.3 MediaMTX (Port 8888)

```bash
# Check MediaMTX is running
docker-compose logs mediamtx | tail -10

# Expected: "mediamtx ready"

# Check metrics
curl -s http://localhost:9998/metrics | grep mediamtx_

# Expected: Prometheus metrics

# Note: Streaming tests require actual RTSP sources
```

**Status**: ✅ MediaMTX operational

---

### Part 4: Monitoring Stack (10 minutes)

#### 4.1 Prometheus (Port 9090)

```bash
# Check Prometheus health
curl -s http://localhost:9090/-/healthy

# Expected: Prometheus is Healthy

# Check all targets are up
curl -s http://localhost:9090/api/v1/targets | \
  jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Expected: All 13 targets with health: "up"

# Query active containers
curl -s 'http://localhost:9090/api/v1/query?query=up' | jq

# Check for firing alerts
curl -s http://localhost:9090/api/v1/alerts | \
  jq '.data.alerts[] | select(.state=="firing")'

# Expected: Empty (no firing alerts)

# Open in browser
open http://localhost:9090

# Manual checks:
# ✅ Status > Targets: All 13 targets UP
# ✅ Alerts: No firing alerts
# ✅ Graph: Query "up" shows 1 for all services
```

**Status**: ✅ Prometheus operational

#### 4.2 Grafana (Port 3001)

```bash
# Check Grafana health
curl -s http://localhost:3001/api/health | jq

# Expected:
{
  "database": "ok",
  "version": "10.2.3"
}

# Check datasources
curl -s -u admin:<password> http://localhost:3001/api/datasources | jq

# Expected: Prometheus, Loki datasources configured

# Open in browser
open http://localhost:3001

# Login: admin / <GRAFANA_ADMIN_PASSWORD from .env>

# Manual checks:
# ✅ Login successful
# ✅ Home dashboard loads
# ✅ Datasources: Prometheus (green), Loki (green)
# ✅ Dashboards: "RTA CCTV System Overview" exists
# ✅ Panels showing data (metrics from last 5 minutes)
```

**Status**: ✅ Grafana operational

#### 4.3 Loki (Port 3100)

```bash
# Check Loki health
curl -s http://localhost:3100/ready

# Expected: ready

# Check if logs are being received
curl -s http://localhost:3100/loki/api/v1/label

# Expected: List of log labels

# Query recent logs
curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={job="cctv"}' | jq

# Test in Grafana Explore:
# 1. Open Grafana
# 2. Go to Explore
# 3. Select Loki datasource
# 4. Query: {service="go-api"}
# Expected: Recent logs from go-api
```

**Status**: ✅ Loki operational

#### 4.4 Alertmanager (Port 9093)

```bash
# Check Alertmanager health
curl -s http://localhost:9093/-/healthy

# Expected: OK

# Check current alerts
curl -s http://localhost:9093/api/v2/alerts | jq

# Expected: [] (empty array, no active alerts)

# Open in browser
open http://localhost:9093

# Manual checks:
# ✅ No active alerts shown
# ✅ Silences tab accessible
```

**Status**: ✅ Alertmanager operational

#### 4.5 Exporters

```bash
# Node Exporter (port 9100)
curl -s http://localhost:9100/metrics | grep node_

# cAdvisor (port 8080)
curl -s http://localhost:8080/metrics | grep container_

# PostgreSQL Exporter (port 9187)
curl -s http://localhost:9187/metrics | grep pg_

# Valkey Exporter (port 9121)
curl -s http://localhost:9121/metrics | grep redis_
```

**Status**: ✅ All exporters operational

---

## Service-by-Service Testing

### Test 1: VMS Service - Camera Discovery

```bash
# Trigger camera discovery (requires Milestone VMS)
curl -X POST http://localhost:8081/api/v1/discover

# Check discovered cameras
curl -s http://localhost:8088/api/v1/cameras | jq

# Expected: List of cameras from Milestone (or empty if not connected)
```

### Test 2: Stream Reservation

```bash
# Reserve a stream
curl -X POST http://localhost:8088/api/v1/streams/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "camera-001",
    "agency": "dubai_police"
  }' | jq

# Expected:
{
  "stream_id": "...",
  "livekit_token": "...",
  "room_name": "..."
}

# Check active streams
curl -s http://localhost:8087/api/v1/limits | jq

# Expected: dubai_police.active = 1

# Release stream
curl -X DELETE http://localhost:8088/api/v1/streams/{stream_id}
```

### Test 3: Recording

```bash
# Start recording (requires active stream)
curl -X POST http://localhost:8083/api/v1/recordings/start \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "camera-001",
    "duration": 300
  }' | jq

# Check recording status
curl -s http://localhost:8083/api/v1/recordings/active | jq

# List recordings
curl -s http://localhost:8088/api/v1/recordings?camera_id=camera-001 | jq
```

### Test 4: Playback

```bash
# Request playback URL
curl -X POST http://localhost:8090/api/v1/playback \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "camera-001",
    "start_time": "2025-10-24T00:00:00Z",
    "end_time": "2025-10-24T01:00:00Z"
  }' | jq

# Expected:
{
  "playback_url": "http://localhost:8090/hls/...",
  "duration": 3600,
  "segments": 12
}
```

### Test 5: Metadata & Search

```bash
# Create incident
curl -X POST http://localhost:8084/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "camera-001",
    "title": "Test Incident",
    "description": "Testing incident creation",
    "severity": "medium",
    "timestamp": "2025-10-24T10:00:00Z"
  }' | jq

# Add tag
curl -X POST http://localhost:8084/api/v1/tags \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "camera-001",
    "tag": "test-tag",
    "timestamp": "2025-10-24T10:00:00Z"
  }'

# Search incidents
curl -s "http://localhost:8084/api/v1/search?q=test&type=incident" | jq
```

---

## Monitoring Stack Verification

### Grafana Dashboard Test

1. Open Grafana: http://localhost:3001
2. Login: `admin` / `<password>`
3. Go to **Dashboards** > **RTA CCTV System Overview**

**Verify these panels**:

| Panel | What to Check | Expected |
|-------|---------------|----------|
| **Service Health** | All services shown | Green (up=1) |
| **API Request Rate** | Requests per second | Graph with data |
| **API Latency (p95)** | 95th percentile latency | <500ms |
| **Active Streams** | Current stream count | 0 or actual count |
| **Storage Used** | MinIO storage | Shows usage % |
| **CPU Usage** | Container CPU % | <80% |
| **Memory Usage** | Container memory | <90% |
| **Database Connections** | Active connections | Shows count |
| **Cache Hit Rate** | Playback cache hits | 0-100% |
| **Recording Rate** | Segments per minute | 0 or actual rate |

### Prometheus Query Test

Open Prometheus: http://localhost:9090/graph

**Test these queries**:

```promql
# All services up?
up{job=~"go-api|vms-service|playback-service"}

# API request rate
rate(http_requests_total{job="go-api"}[5m])

# Active streams
stream_reservations_active

# Storage usage
minio_storage_bytes_used / minio_storage_bytes_total * 100

# Database connections
pg_stat_database_numbackends{datname="cctv"}

# Cache hit rate
rate(playback_cache_hits_total[5m]) /
(rate(playback_cache_hits_total[5m]) + rate(playback_cache_misses_total[5m]))
```

### Alert Testing

**Trigger a test alert**:

```bash
# Stop a service to trigger ServiceDown alert
docker-compose stop vms-service

# Wait 2 minutes (alert threshold)
sleep 120

# Check firing alerts
curl -s http://localhost:9090/api/v1/alerts | \
  jq '.data.alerts[] | select(.state=="firing")'

# Expected: ServiceDown alert for vms-service

# Check Alertmanager
curl -s http://localhost:9093/api/v2/alerts | jq

# Restart service
docker-compose start vms-service

# Wait 2 minutes for alert to resolve
sleep 120

# Verify alert resolved
curl -s http://localhost:9090/api/v1/alerts | \
  jq '.data.alerts[] | select(.labels.alertname=="ServiceDown")'

# Expected: Empty or state="resolved"
```

---

## Functional Feature Testing

### Feature 1: Live Streaming Workflow

**Test complete live streaming flow**:

```bash
# 1. Get available cameras
curl -s http://localhost:8088/api/v1/cameras | jq

# 2. Check agency quota
curl -s http://localhost:8087/api/v1/limits/dubai_police | jq

# 3. Reserve stream
STREAM_RESPONSE=$(curl -s -X POST http://localhost:8088/api/v1/streams/reserve \
  -H "Content-Type: application/json" \
  -d '{"camera_id":"camera-001","agency":"dubai_police"}')

echo $STREAM_RESPONSE | jq

STREAM_ID=$(echo $STREAM_RESPONSE | jq -r '.stream_id')
LIVEKIT_TOKEN=$(echo $STREAM_RESPONSE | jq -r '.livekit_token')

# 4. Verify stream is active
curl -s http://localhost:8087/api/v1/limits | jq '.dubai_police.active'
# Expected: 1

# 5. (Optional) Open dashboard and use LiveKit token to connect

# 6. Release stream
curl -X DELETE http://localhost:8088/api/v1/streams/$STREAM_ID

# 7. Verify stream released
curl -s http://localhost:8087/api/v1/limits | jq '.dubai_police.active'
# Expected: 0
```

### Feature 2: Recording & Playback Workflow

```bash
# 1. Start recording
RECORDING_RESPONSE=$(curl -s -X POST http://localhost:8083/api/v1/recordings/start \
  -H "Content-Type: application/json" \
  -d '{"camera_id":"camera-001","duration":60}')

echo $RECORDING_RESPONSE | jq

RECORDING_ID=$(echo $RECORDING_RESPONSE | jq -r '.recording_id')

# 2. Check recording status
curl -s http://localhost:8083/api/v1/recordings/$RECORDING_ID | jq

# 3. Wait for some segments to be recorded (30 seconds)
sleep 30

# 4. List recorded segments
curl -s "http://localhost:8088/api/v1/recordings?camera_id=camera-001" | jq

# 5. Request playback
PLAYBACK_RESPONSE=$(curl -s -X POST http://localhost:8090/api/v1/playback \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id":"camera-001",
    "start_time":"'$(date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ)'",
    "end_time":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }')

echo $PLAYBACK_RESPONSE | jq

PLAYBACK_URL=$(echo $PLAYBACK_RESPONSE | jq -r '.playback_url')

# 6. Test playback URL
curl -I $PLAYBACK_URL
# Expected: 200 OK

# 7. Stop recording
curl -X POST http://localhost:8083/api/v1/recordings/$RECORDING_ID/stop
```

### Feature 3: Metadata & Incident Management

```bash
# 1. Create incident
INCIDENT_RESPONSE=$(curl -s -X POST http://localhost:8084/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id":"camera-001",
    "title":"Traffic Accident",
    "description":"Vehicle collision at intersection",
    "severity":"high",
    "timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }')

echo $INCIDENT_RESPONSE | jq

INCIDENT_ID=$(echo $INCIDENT_RESPONSE | jq -r '.id')

# 2. Add tags to incident
curl -X POST http://localhost:8084/api/v1/incidents/$INCIDENT_ID/tags \
  -H "Content-Type: application/json" \
  -d '{"tags":["accident","intersection","high-priority"]}'

# 3. Add annotation
curl -X POST http://localhost:8084/api/v1/annotations \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id":"camera-001",
    "incident_id":"'$INCIDENT_ID'",
    "timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
    "note":"Ambulance arrived",
    "created_by":"operator-001"
  }'

# 4. Search incidents
curl -s "http://localhost:8084/api/v1/search?q=traffic&type=incident" | jq

# 5. Get incident details
curl -s http://localhost:8084/api/v1/incidents/$INCIDENT_ID | jq

# 6. Update incident status
curl -X PATCH http://localhost:8084/api/v1/incidents/$INCIDENT_ID \
  -H "Content-Type: application/json" \
  -d '{"status":"resolved"}'

# 7. List all incidents
curl -s http://localhost:8084/api/v1/incidents | jq
```

---

## Performance Testing

### Load Test Preparation

```bash
# Install k6 (if not installed)
# https://k6.io/docs/getting-started/installation/

# Or use Docker
alias k6="docker run --rm -i --network=host grafana/k6"
```

### Test 1: API Endpoint Load

Create `tests/load/api-load-test.js`:

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 50 },  // Ramp to 50 users
    { duration: '3m', target: 50 },  // Stay at 50 users
    { duration: '1m', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% under 500ms
    http_req_failed: ['rate<0.01'],   // <1% errors
  },
};

export default function() {
  // Test health endpoint
  let healthRes = http.get('http://localhost:8088/health');
  check(healthRes, {
    'health status 200': (r) => r.status === 200,
  });

  // Test cameras endpoint
  let camerasRes = http.get('http://localhost:8088/api/v1/cameras');
  check(camerasRes, {
    'cameras status 200': (r) => r.status === 200,
  });

  sleep(1);
}
```

Run the test:
```bash
k6 run tests/load/api-load-test.js
```

**Expected results**:
- ✅ p(95) latency < 500ms
- ✅ Error rate < 1%
- ✅ All requests successful

### Test 2: Stream Reservation Load

Create `tests/load/stream-reservation-test.js`:

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 10 },
    { duration: '30s', target: 0 },
  ],
};

export default function() {
  // Check limits
  let limitsRes = http.get('http://localhost:8087/api/v1/limits');
  check(limitsRes, { 'limits 200': (r) => r.status === 200 });

  // Reserve stream
  let reserveRes = http.post(
    'http://localhost:8088/api/v1/streams/reserve',
    JSON.stringify({
      camera_id: `camera-${__VU}`,
      agency: 'other'
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  let streamId = null;
  let success = check(reserveRes, {
    'reserve 200': (r) => r.status === 200 || r.status === 429,
  });

  if (reserveRes.status === 200) {
    streamId = reserveRes.json('stream_id');
  }

  sleep(5);

  // Release stream
  if (streamId) {
    http.del(`http://localhost:8088/api/v1/streams/${streamId}`);
  }

  sleep(1);
}
```

Run:
```bash
k6 run tests/load/stream-reservation-test.js
```

### Monitor Performance During Load

While running load tests, monitor in separate terminals:

```bash
# Terminal 1: Watch resource usage
docker stats

# Terminal 2: Watch Prometheus metrics
watch -n 2 'curl -s "http://localhost:9090/api/v1/query?query=rate(http_requests_total{job=\"go-api\"}[1m])" | jq'

# Terminal 3: Watch logs for errors
docker-compose logs -f go-api | grep -i error

# Terminal 4: Open Grafana dashboard
open http://localhost:3001
```

---

## Troubleshooting

### Issue 1: Services Not Starting

**Symptom**: `docker-compose ps` shows services as "Exit 1" or "Restarting"

**Debug**:
```bash
# Check logs
docker-compose logs <service-name>

# Common issues:
# - Port conflict: Change port in docker-compose.yml
# - Memory: Increase Docker memory limit
# - Dependencies: Check if postgres/valkey started first

# Fix: Restart specific service
docker-compose restart <service-name>

# Or restart all
docker-compose down && docker-compose up -d
```

### Issue 2: Health Check Failing

**Symptom**: `./scripts/health-check.sh` shows red ✗ FAILED

**Debug**:
```bash
# Check which endpoint failed
curl -v http://localhost:<port>/health

# Check service logs
docker-compose logs <service-name> | tail -50

# Check if port is accessible
netstat -tuln | grep <port>

# Restart service
docker-compose restart <service-name>
```

### Issue 3: Grafana Showing "No Data"

**Debug**:
```bash
# Check Prometheus is scraping
curl http://localhost:9090/api/v1/targets | jq

# Check if metrics exist
curl http://localhost:8088/metrics | head -20

# Check Grafana datasource
curl -u admin:<pass> http://localhost:3001/api/datasources

# Restart monitoring stack
docker-compose restart prometheus grafana
```

### Issue 4: MinIO Access Denied

**Debug**:
```bash
# Check MinIO is running
docker-compose logs minio | tail -20

# Verify credentials
echo $MINIO_ROOT_USER
echo $MINIO_ROOT_PASSWORD

# Test connection
curl -u admin:<password> http://localhost:9000/minio/health/live

# Recreate buckets
docker-compose restart minio-init
```

### Issue 5: Database Connection Error

**Debug**:
```bash
# Check PostgreSQL
docker exec cctv-postgres pg_isready -U cctv

# Check password
echo $POSTGRES_PASSWORD

# Test connection
docker exec cctv-postgres psql -U cctv -c "SELECT 1"

# Check service environment
docker-compose exec go-api env | grep POSTGRES

# Restart database and dependent services
docker-compose restart postgres
docker-compose restart go-api vms-service storage-service
```

---

## Complete Verification Checklist

Use this checklist to verify complete system functionality:

### Infrastructure
- [ ] PostgreSQL accepting connections
- [ ] Valkey responding to ping
- [ ] MinIO buckets created (4 buckets)
- [ ] All Docker containers running (28 containers)

### Core Services
- [ ] VMS Service health check passing
- [ ] Storage Service health check passing
- [ ] Recording Service health check passing
- [ ] Metadata Service health check passing
- [ ] Playback Service health check passing
- [ ] Stream Counter health check passing
- [ ] Go API health check passing

### Frontend & Streaming
- [ ] Dashboard accessible and loading
- [ ] LiveKit server responding
- [ ] MediaMTX running

### Monitoring
- [ ] Prometheus scraping all 13 targets
- [ ] Grafana dashboards showing data
- [ ] Loki receiving logs
- [ ] Alertmanager healthy
- [ ] All 4 exporters providing metrics

### Functional Tests
- [ ] Camera list API working
- [ ] Stream reservation working
- [ ] Stream quota tracking working
- [ ] Recording start/stop working
- [ ] Playback URL generation working
- [ ] Incident creation working
- [ ] Search functionality working
- [ ] Tag management working

### Performance
- [ ] API latency <500ms (p95)
- [ ] CPU usage <80%
- [ ] Memory usage <90%
- [ ] Disk space >15% free

### Operations
- [ ] Health check script passing
- [ ] Backup script executes successfully
- [ ] Logs accessible via Loki
- [ ] Alerts configured and testable

---

## Next Steps

After completing all tests:

1. **For Development**:
   - Add test cameras to database
   - Configure Milestone VMS connection
   - Test live streaming with real cameras
   - Develop additional features

2. **For Production**:
   - Follow `docs/deployment.md` for production setup
   - Configure TLS certificates
   - Set up external volumes (NFS/S3)
   - Configure production secrets
   - Set up backup cron job
   - Configure email alerts (SMTP)
   - Perform security audit
   - Load testing with production data

3. **For Testing**:
   - Implement unit tests (see `tests/README.md`)
   - Write integration tests
   - Create E2E test suite
   - Set up CI/CD pipeline

---

## Support

If you encounter issues not covered in this guide:

1. Check logs: `docker-compose logs -f`
2. Review documentation: `docs/`
3. Run health check: `./scripts/health-check.sh`
4. Check PROJECT-STATUS.md for known issues
5. Consult `docs/operations.md` for incident response

---

**Testing Guide Version**: 1.0.0
**Platform Version**: 1.0.0
**Completion**: 97%
**Status**: Production Ready (except Phase 5 Auth & Phase 7 AI)
