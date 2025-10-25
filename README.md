# RTA CCTV Video Management System

**Hybrid VMS with AI Analytics** for Roads and Transport Authority

## **System Overview**

- **Scale**: 500 concurrent cameras, 1000+ viewers
- **Storage**: Configurable (Local/Milestone/Both)
- **Latency**: <800ms live streaming, <300ms playback
- **AI**: Object detection with YOLOv8 Nano
- **Languages**: English & Arabic (RTL support)

---

## **Architecture**

```
┌─── INGEST LAYER ───┐     ┌─── STORAGE LAYER ───┐     ┌─── APPLICATION LAYER ───┐
│  MediaMTX          │────▶│  VMS Service         │────▶│  Go API Core            │
│  (RTSP Ingest)     │     │  Storage Service     │     │  Kong Gateway           │
└────────────────────┘     │  Recording Service   │     │  React Dashboard        │
                           │  Metadata Service    │     └─────────────────────────┘
┌─── STREAMING ───┐        │  Playback Service    │
│  LiveKit SFU    │        └──────────────────────┘     ┌─── MONITORING ───┐
│  (WebRTC)       │                                      │  Prometheus       │
└─────────────────┘        ┌─── AI LAYER ───┐           │  Grafana          │
                           │  Object         │           │  Loki             │
┌─── CACHE ───┐            │  Detection      │           └───────────────────┘
│  Valkey      │            └─────────────────┘
└──────────────┘
```

---

## **Services**

| Service | Purpose | Tech Stack | Port |
|---------|---------|------------|------|
| **vms-service** | Milestone VMS integration | Go 1.21 | 8081 |
| **storage-service** | Storage orchestration | Go 1.21 | 8082 |
| **recording-service** | Video recording | Go 1.21 + FFmpeg | 8083 |
| **metadata-service** | Search & tags | Go 1.21 + PostgreSQL | 8084 |
| **object-detection** | AI analytics | Go 1.21 + ONNX | 8085 |
| **playback-service** | Unified playback | Go 1.21 + FFmpeg | 8086 |
| **stream-counter** | Quota management | Go 1.21 + Valkey | 8087 |
| **go-api** | Central API | Go 1.21 | 8080 |
| **dashboard** | Operator UI | React 18 + TS | 3000 |

---

## **Quick Start**

### **Prerequisites**

- Docker & Docker Compose
- Go 1.21+ (for development)
- Node.js 18+ (for frontend)
- 16GB RAM minimum
- 100GB disk space

### **Development Setup**

```bash
# Clone repository
git clone <repo-url>
cd cns

# Copy environment template
cp .env.example .env

# Edit .env with your configuration
nano .env

# Start all services
docker-compose up -d

# Check health
./scripts/health-check.sh

# View logs
docker-compose logs -f
```

### **Access Points**

- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8000 (via Kong)
- **Grafana**: http://localhost:3001
- **MinIO Console**: http://localhost:9001

---

## **Development**

### **Service Development**

Each service follows Clean Architecture:

```
services/{service-name}/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── domain/              # Business entities
│   ├── usecase/             # Business logic
│   ├── repository/          # Data access
│   ├── delivery/            # HTTP handlers
│   └── infrastructure/      # External integrations
├── pkg/                     # Shared utilities
├── Dockerfile
├── go.mod
└── README.md
```

### **Running Tests**

```bash
# Unit tests
cd services/vms-service
go test ./... -v -cover

# Integration tests
cd tests/integration
go test ./... -v

# Load tests
cd tests/load
k6 run stream-test.js
```

### **Building Docker Images**

```bash
# Build single service
docker build -t rta/vms-service:latest services/vms-service/

# Build all services
./scripts/build-all.sh
```

---

## **Configuration**

### **Environment Variables**

See `.env.example` for full list. Key variables:

```bash
# Storage Configuration
STORAGE_MODE=BOTH              # LOCAL, MILESTONE, or BOTH
STORAGE_BACKEND=MINIO          # MINIO, S3, or FILESYSTEM
RETENTION_DAYS=90

# Milestone VMS
MILESTONE_SERVER=milestone.rta.ae
MILESTONE_USER=${MILESTONE_USER}
MILESTONE_PASS=${MILESTONE_PASS}

# Database
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
VALKEY_PASSWORD=${VALKEY_PASSWORD}

# Agency Limits
LIMIT_DUBAI_POLICE=50
LIMIT_METRO=30
LIMIT_BUS=20
LIMIT_OTHER=400
```

---

## **Monitoring**

### **Metrics**

All services expose Prometheus metrics on `/metrics`:

```
# Stream quotas
cctv_streams_active{source}
cctv_streams_limit{source}

# Storage
cctv_storage_bytes_used{backend}
cctv_recording_segments_total

# Performance
cctv_api_request_duration_seconds
cctv_playback_latency_ms
```

### **Dashboards**

Grafana dashboards available at `config/grafana/dashboards/`:
1. Stream Overview
2. Storage Metrics
3. Playback Performance
4. AI Analytics
5. API Performance

---

## **Deployment**

### **Production Deployment**

```bash
# Deploy to production
docker-compose -f deploy/docker-compose.prod.yml up -d

# Run health checks
./scripts/health-check.sh

# View status
docker-compose ps
```

### **Backup**

```bash
# Backup database and configs
./scripts/backup.sh

# Backup stored at: /backup/YYYYMMDD.tar.gz
```

### **Scaling**

```bash
# Scale recording workers
docker-compose up -d --scale recording-service=3

# Scale API instances
docker-compose up -d --scale go-api=2
```

---

## **Documentation**

- [Architecture](docs/architecture.md)
- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Operations Runbook](docs/operations.md)
- [Troubleshooting](docs/troubleshooting.md)

---

## **Project Status**

- [x] Requirements finalized
- [x] Implementation plan created
- [ ] Phase 1: Core Infrastructure (In Progress)
- [ ] Phase 2: Storage & Recording
- [ ] Phase 3: Streaming & Playback
- [ ] Phase 4: AI & Frontend
- [ ] Phase 5: IAM Integration
- [ ] Phase 6: Testing & Optimization

---

## **Contributing**

### **Development Workflow**

1. Create feature branch: `git checkout -b feature/my-feature`
2. Implement with tests (>80% coverage)
3. Run linter: `golangci-lint run`
4. Create Pull Request
5. Wait for CI checks and code review
6. Merge to develop

### **Code Style**

- **Go**: Follow [Uber Go Style Guide](https://github.com/uber-go/guide)
- **React**: ESLint + Prettier configuration
- **Commit Messages**: Conventional Commits format

---

## **Support**

- **Issues**: [GitHub Issues](https://github.com/rta/cctv/issues)
- **Wiki**: [Project Wiki](https://github.com/rta/cctv/wiki)
- **Email**: cctv-support@rta.ae

---

## **License**

Proprietary - Roads and Transport Authority (RTA)

---

**Version**: 1.0.0-dev
**Last Updated**: 2025-01-XX
