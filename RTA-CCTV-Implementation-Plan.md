# RTA CCTV System - Implementation Plan

## **PROJECT OVERVIEW**

**System**: Hybrid Video Management System with AI Analytics
**Timeline**: 12 weeks (3 months)
**Team Size**: 5-7 developers
**Architecture**: Microservices with Docker Compose deployment

---

## **TEAM STRUCTURE**

| Role | Responsibility | Count |
|------|---------------|-------|
| **Backend Lead** | Go services, architecture decisions | 1 |
| **Backend Developer** | VMS, Storage, Recording services | 2 |
| **Frontend Developer** | React dashboard, UI/UX | 1 |
| **DevOps Engineer** | Docker, monitoring, deployment | 1 |
| **AI Engineer** | Object detection, model optimization | 1 |
| **QA Engineer** | Testing, performance benchmarks | 1 |

---

## **DEVELOPMENT PHASES**

### **PHASE 1: FOUNDATION** (Weeks 1-2)

#### **Week 1: Core Infrastructure**

**Deliverables**:
1. **MediaMTX RTSP Ingest Service**
   - mediamtx.yml configuration with tee output
   - Docker container setup
   - Health check endpoint
   - Load test: 100 concurrent streams

2. **Valkey Deployment**
   - valkey.conf optimized for low latency
   - Docker container with persistence
   - Connection testing

3. **VMS Service - Basic Structure**
   - Go project scaffolding (Clean Architecture)
   - Milestone SDK wrapper (connection pool)
   - In-memory camera cache
   - GET /vms/cameras endpoint

**Tasks**:
```
├── Setup project repository structure
├── Docker network configuration
├── Environment variable management (.env)
├── MediaMTX configuration and testing
├── Valkey setup with AOF persistence
├── VMS Service: Milestone connection
└── Basic integration test
```

**Success Criteria**:
- [ ] MediaMTX ingests 100 RTSP streams
- [ ] Valkey responds to SET/GET in <5ms
- [ ] VMS Service connects to Milestone
- [ ] VMS Service lists cameras (cached)

---

#### **Week 2: Quota Management & API Gateway**

**Deliverables**:
1. **Valkey Stream Counter Service**
   - Lua scripts (reserve_stream.lua, release_stream.lua, heartbeat.lua)
   - Go service with chi router
   - API: POST /stream/reserve, DELETE /stream/release
   - Unit tests for Lua scripts
   - Integration tests with Valkey

2. **Kong API Gateway**
   - kong.yaml declarative configuration
   - Custom Lua plugin: source-based rate limiter
   - Routes to VMS service
   - Bilingual error messages (AR/EN)

3. **Go API - Initial Structure**
   - Clean Architecture scaffolding
   - Database connection (PostgreSQL)
   - Valkey connection pool
   - Basic health check

**Tasks**:
```
├── Write and test Lua scripts (reserve/release/heartbeat)
├── Stream Counter Service implementation
├── API endpoint testing (Postman collection)
├── Kong configuration with custom plugin
├── Kong rate limiting integration with Valkey
├── Go API project structure
└── Database schema initialization
```

**Success Criteria**:
- [ ] Lua scripts execute atomically
- [ ] Reserve/release operations <10ms p99
- [ ] Kong enforces source-based rate limits
- [ ] Kong returns bilingual error messages
- [ ] Load test: 10,000 reserve/release ops/sec

---

### **PHASE 2: STORAGE & RECORDING** (Weeks 3-4)

#### **Week 3: Storage Backend & Metadata**

**Deliverables**:
1. **MinIO Object Storage**
   - MinIO cluster setup (4 nodes for distributed mode)
   - Bucket creation: recordings, clips, thumbnails
   - Access policy configuration
   - S3 SDK integration testing

2. **Video Storage Service**
   - Go service with storage backend abstraction
   - Configuration: LOCAL/MILESTONE/BOTH modes
   - Storage path generation
   - Admin API: GET/PUT /admin/storage/config

3. **Metadata Service - Database Schema**
   - PostgreSQL schema (video_sessions, video_segments, video_clips, etc.)
   - Database migration scripts
   - Connection pool configuration
   - Basic CRUD operations

**Tasks**:
```
├── MinIO deployment and bucket setup
├── Storage Service: backend abstraction layer
├── Storage Service: configuration hot-reload
├── Metadata database schema design
├── PostgreSQL migrations (using golang-migrate)
├── Metadata Service: session management
└── Integration test: storage config switch
```

**Success Criteria**:
- [ ] MinIO stores and retrieves 100GB test data
- [ ] Storage Service switches modes without restart
- [ ] Database schema supports 1M+ segments
- [ ] Metadata queries execute in <100ms

---

#### **Week 4: Recording & Clip Management**

**Deliverables**:
1. **Recording Service**
   - Worker pool for concurrent recordings
   - FFmpeg wrapper (H.264 copy, segmentation)
   - Segment upload to MinIO
   - Local cleanup after upload
   - Job queue (Valkey-based)

2. **Clip Extraction**
   - Segment finder (by time range)
   - FFmpeg concatenation
   - Clip metadata creation
   - API: POST /clips, GET /clips/{id}

3. **Metadata Service - Search API**
   - Search by camera, time range, tags
   - Full-text search on descriptions
   - Pagination support
   - Performance optimization (indexes)

**Tasks**:
```
├── Recording Service: FFmpeg integration
├── Recording Service: segment upload pipeline
├── Recording Service: job queue implementation
├── Clip extraction logic
├── Metadata Service: search endpoint
├── Metadata Service: tag management
└── Load test: 50 concurrent recordings
```

**Success Criteria**:
- [ ] Recording Service handles 50 cameras
- [ ] CPU usage <5% per camera (H.264 copy)
- [ ] Segments uploaded within 10s of creation
- [ ] Clip extraction completes in <30s
- [ ] Search query returns results in <200ms

---

### **PHASE 3: STREAMING & PLAYBACK** (Weeks 5-6)

#### **Week 5: Live Streaming with LiveKit**

**Deliverables**:
1. **LiveKit SFU**
   - livekit.yaml configuration (low latency)
   - Valkey integration for room state
   - Simulcast configuration (3 layers)
   - TURN server (coturn) setup

2. **LiveKit Ingress**
   - Ingress configuration (RTSP → LiveKit)
   - Connection to MediaMTX
   - Room creation automation

3. **Go API - Stream Management**
   - RequestStream use case
   - LiveKit room creation
   - Token generation
   - Stream reservation tracking
   - WebSocket hub for real-time updates

**Tasks**:
```
├── LiveKit deployment and configuration
├── TURN server setup (coturn)
├── LiveKit Ingress: RTSP ingestion
├── Go API: StreamUseCase implementation
├── Go API: LiveKit SDK integration
├── WebSocket hub for stream stats
└── E2E test: reserve stream → view in browser
```

**Success Criteria**:
- [ ] LiveKit streams 100 cameras
- [ ] Glass-to-glass latency <800ms
- [ ] Simulcast switches quality smoothly
- [ ] TURN works behind NAT
- [ ] WebSocket updates in real-time

---

#### **Week 6: Unified Playback**

**Deliverables**:
1. **Playback Service**
   - Source detection (local vs Milestone)
   - FFmpeg transmux (H.264 → HLS)
   - Segment caching (LRU eviction)
   - Signed URL generation
   - Nginx configuration for serving

2. **VMS Service - Recording Export**
   - Export API: POST /vms/recordings/export
   - Milestone recording fetch
   - Integration with Playback Service

3. **Go API - Playback Orchestration**
   - PlaybackUseCase implementation
   - Source preference logic (local first)
   - Playback session management

**Tasks**:
```
├── Playback Service: source detection
├── Playback Service: FFmpeg transmux wrapper
├── Playback Service: cache management
├── Nginx: HLS serving configuration
├── VMS Service: Milestone export integration
├── Go API: playback orchestration
└── Latency benchmark: local vs Milestone
```

**Success Criteria**:
- [ ] Local playback latency <300ms
- [ ] Milestone playback latency <800ms
- [ ] Cache hit ratio >70%
- [ ] Signed URLs expire correctly
- [ ] Concurrent playback: 20 sessions

---

### **PHASE 4: AI & FRONTEND** (Weeks 7-8)

#### **Week 7: Object Detection**

**Deliverables**:
1. **Object Detection Service**
   - YOLOv8 Nano model integration (ONNX)
   - Frame extraction (1 FPS sampling)
   - Batch inference (10 frames)
   - Detection result storage
   - Worker queue for segment analysis

2. **Metadata Service - AI Integration**
   - object_detections table population
   - Search by object class
   - Detection summary on segments

**Tasks**:
```
├── YOLOv8 Nano model download and conversion (ONNX)
├── Object Detection Service: ONNX Runtime integration
├── Frame extraction utility (FFmpeg)
├── Batch inference implementation
├── Detection result saving
├── Metadata Service: AI search endpoint
└── Benchmark: detection latency and accuracy
```

**Success Criteria**:
- [ ] Detection latency <20ms per frame (CPU)
- [ ] Accuracy: mAP >35% on test dataset
- [ ] Memory usage <200MB per worker
- [ ] AI search returns results in <300ms
- [ ] 10 concurrent detection workers

---

#### **Week 8: React Dashboard - Grid Layouts**

**Deliverables**:
1. **React Dashboard - Core**
   - Project setup (Vite, TypeScript)
   - Grid layout system (2×2, 3×3, 4×4, hotspots)
   - LiveKit SDK integration
   - Camera selection and placement

2. **Performance Optimizations**
   - Viewport-based rendering (Intersection Observer)
   - Video element pooling
   - Web Workers for grid calculations
   - React.memo for expensive components

3. **State Management**
   - Zustand store setup
   - Agency quota tracking
   - Layout persistence (localStorage)

**Tasks**:
```
├── React project scaffolding
├── Grid layout algorithm implementation
├── LiveKit SDK: useLiveKit hook
├── Viewport detection with Intersection Observer
├── Video element pooling
├── Zustand store for cameras and layout
└── E2E test: place 64 cameras in grid
```

**Success Criteria**:
- [ ] 64 cameras render without lag
- [ ] Memory usage <500MB for 64 streams
- [ ] Layout switch in <100ms
- [ ] Only visible cameras consume bandwidth
- [ ] Drag-and-drop camera placement works

---

### **PHASE 5: UI FEATURES & IAM** (Weeks 9-10)

#### **Week 9: Dashboard Features**

**Deliverables**:
1. **Playback UI**
   - Timeline component
   - HLS.js integration
   - Playback controls (play/pause/seek)
   - Time range selector

2. **Clip Management UI**
   - Clip creation dialog
   - Tag input component
   - Annotation interface
   - Clip search UI

3. **Arabic RTL Support**
   - i18n configuration (react-i18next)
   - RTL stylesheet
   - Arabic translations
   - Language switcher

**Tasks**:
```
├── Timeline component development
├── HLS.js integration for playback
├── Clip creation workflow
├── Tag management UI
├── Annotation tools (markers, regions)
├── i18n setup with Arabic translations
└── RTL CSS adjustments
```

**Success Criteria**:
- [ ] Playback timeline shows accurate segments
- [ ] HLS playback starts in <2s
- [ ] Clip creation extracts correct time range
- [ ] Tags are searchable
- [ ] Arabic text renders correctly (RTL)

---

#### **Week 10: IAM Integration**

**Deliverables**:
1. **IAM Service**
   - User management API
   - Group management API
   - JWT token generation (RS256)
   - Webhook handler for IAM events

2. **Go API - Authentication**
   - JWT validation middleware
   - Permission checking
   - Audit logging
   - Session management

3. **Dashboard - Auth Flow**
   - Login page
   - JWT storage and refresh
   - Permission-based UI (hide unavailable features)
   - Agency quota display

**Tasks**:
```
├── IAM Service: user CRUD operations
├── IAM Service: JWT token generation
├── IAM Service: webhook handler
├── Go API: JWT validation middleware
├── Go API: permission checker
├── Audit logging implementation
├── Dashboard: login and token management
└── E2E test: login → view camera → logout
```

**Success Criteria**:
- [ ] JWT tokens validate correctly
- [ ] Permissions enforced on API
- [ ] Disabled users cannot access system
- [ ] Audit log captures all actions
- [ ] Dashboard shows agency quotas

---

### **PHASE 6: DEPLOYMENT & TESTING** (Weeks 11-12)

#### **Week 11: Docker Compose & Monitoring**

**Deliverables**:
1. **Docker Compose Stack**
   - Complete docker-compose.yml
   - Network segmentation
   - Volume management
   - Environment variable configuration
   - Health checks for all services

2. **Monitoring Stack**
   - Prometheus: scrape configuration
   - Grafana: dashboards
   - Loki: log aggregation
   - Alerting rules

3. **CI/CD Pipeline**
   - GitHub Actions or GitLab CI
   - Build Docker images
   - Run tests
   - Deploy to staging

**Tasks**:
```
├── docker-compose.yml: all services
├── Network isolation configuration
├── Volume and storage planning
├── Prometheus: scrape targets
├── Grafana: dashboards creation (5 dashboards)
├── Loki: log shipping configuration
├── Alert rules for critical metrics
└── CI/CD: build and test pipeline
```

**Success Criteria**:
- [ ] All services start with `docker-compose up`
- [ ] Health checks pass for all services
- [ ] Prometheus scrapes all metrics
- [ ] Grafana dashboards show live data
- [ ] Alerts fire on test conditions
- [ ] CI/CD pipeline builds and tests

---

#### **Week 12: Load Testing & Optimization**

**Deliverables**:
1. **Load Testing**
   - K6 scripts: 500 concurrent streams
   - Stress test: 1000 concurrent viewers
   - Recording stress test: 100 cameras
   - Playback stress test: 50 sessions
   - AI detection benchmark: 500 segments

2. **Performance Tuning**
   - Database query optimization
   - Valkey memory tuning
   - FFmpeg process limits
   - LiveKit bandwidth optimization
   - React bundle size optimization

3. **Documentation**
   - Architecture diagram
   - API documentation (OpenAPI/Swagger)
   - Deployment guide
   - Operations runbook
   - Troubleshooting guide

**Tasks**:
```
├── K6 load test scripts
├── Run load tests and collect metrics
├── Identify bottlenecks
├── Database index optimization
├── Valkey maxmemory tuning
├── React bundle analysis and optimization
├── Architecture diagram (draw.io)
├── API documentation generation
├── Deployment guide writing
└── Operations runbook creation
```

**Success Criteria**:
- [ ] 500 concurrent streams at <800ms latency
- [ ] 1000 viewers without degradation
- [ ] 100 cameras recording at <3% CPU each
- [ ] 50 playback sessions at <500ms latency
- [ ] AI detection: 100 frames/sec throughput
- [ ] All documentation complete

---

## **PROJECT STRUCTURE**

```
rta-cctv/
├── services/
│   ├── vms-service/             # Milestone integration
│   ├── storage-service/         # Storage orchestration
│   ├── recording-service/       # Video recording
│   ├── metadata-service/        # Search & tags
│   ├── object-detection/        # AI analytics
│   ├── playback-service/        # Unified playback
│   ├── stream-counter/          # Valkey quota management
│   └── go-api/                  # Central API
├── web/
│   └── dashboard/               # React frontend
├── config/
│   ├── mediamtx.yml
│   ├── livekit.yaml
│   ├── kong.yaml
│   ├── valkey.conf
│   ├── prometheus.yml
│   └── nginx.conf
├── deploy/
│   ├── docker-compose.yml
│   ├── docker-compose.prod.yml
│   └── .env.example
├── scripts/
│   ├── init-db.sql
│   ├── backup.sh
│   └── health-check.sh
├── tests/
│   ├── integration/
│   ├── load/
│   └── e2e/
├── docs/
│   ├── architecture.md
│   ├── api.md
│   ├── deployment.md
│   └── operations.md
└── README.md
```

---

## **DEVELOPMENT WORKFLOW**

### **Git Branching Strategy**

```
main
  ├── develop
  │   ├── feature/vms-service
  │   ├── feature/storage-service
  │   ├── feature/recording-service
  │   ├── feature/object-detection
  │   ├── feature/dashboard-grid
  │   └── feature/iam-integration
  ├── release/v1.0
  └── hotfix/critical-bug
```

### **Code Review Process**

1. Developer creates feature branch
2. Implements feature with tests
3. Creates Pull Request
4. CI runs tests and builds
5. Code review by 2 team members
6. Merge to develop
7. Weekly merge to release branch
8. Deploy to staging for QA
9. Production deployment (bi-weekly)

### **Testing Strategy**

| Test Type | Coverage | Tools |
|-----------|----------|-------|
| Unit Tests | >80% | Go: testify, React: Jest |
| Integration Tests | Key flows | Testcontainers, Docker |
| E2E Tests | Critical paths | Playwright |
| Load Tests | Performance targets | K6 |
| Security Tests | OWASP Top 10 | OWASP ZAP |

---

## **DAILY STANDUP TEMPLATE**

```
Team Member: [Name]
Date: [YYYY-MM-DD]

✅ Yesterday:
- Completed task X
- Fixed bug Y

🚧 Today:
- Working on task Z
- Code review for PR #123

🚫 Blockers:
- Waiting for Milestone test credentials
- Need clarification on storage config behavior
```

---

## **WEEKLY SPRINT PLANNING**

**Every Monday, 10:00 AM**

Agenda:
1. Review previous week's progress (15 min)
2. Demo completed features (30 min)
3. Identify blockers (15 min)
4. Plan current week's tasks (30 min)
5. Update project board (10 min)

---

## **RISK MANAGEMENT**

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Milestone SDK integration issues | Medium | High | Early integration test, contact vendor support |
| Storage capacity underestimated | Low | High | Monitor usage, plan for expansion |
| LiveKit performance bottleneck | Medium | Medium | Load test early, plan for horizontal scaling |
| YOLOv8 accuracy insufficient | Low | Medium | Test with real data, retrain if needed |
| Team member unavailability | Medium | Medium | Cross-training, documentation |
| Security vulnerability | Low | High | Regular security audits, penetration testing |

---

## **DEPENDENCIES**

### **External Systems**
- **Milestone VMS**: RTSP access, SDK credentials, test environment
- **RTA IAM**: Public key for JWT validation, webhook endpoint
- **Network**: 10Gbps internal, firewall rules for RTSP/WebRTC

### **Hardware Requirements**
- **Compute**: 16-core CPU, 32GB RAM (development)
- **Storage**: 1TB SSD (development), 500TB NAS (production)
- **Network**: 10Gbps NIC

### **Software Licenses**
- Milestone SDK (provided by customer)
- LiveKit (open source, free)
- All other components: open source

---

## **DEPLOYMENT STRATEGY**

### **Environments**

| Environment | Purpose | Specs |
|-------------|---------|-------|
| **Development** | Local testing | Docker Compose on laptops |
| **Staging** | QA and demos | 16-core, 32GB, 2TB SSD |
| **Production** | Live system | 32-core, 64GB, 500TB NAS |

### **Deployment Process**

1. **Staging Deployment** (Weekly)
   - Merge develop → release/vX.Y
   - Build Docker images
   - Deploy to staging
   - Run smoke tests
   - QA testing (2 days)

2. **Production Deployment** (Bi-weekly)
   - Tag release/vX.Y
   - Build production images
   - Backup production database
   - Deploy with zero-downtime (blue-green)
   - Health check verification
   - Rollback plan ready

### **Rollback Procedure**

```bash
# Stop new version
docker-compose -f docker-compose.prod.yml down

# Restore previous version
docker-compose -f docker-compose.prod.yml up -d --build

# Restore database (if needed)
psql -U cctv < backup_YYYYMMDD.sql
```

---

## **MONITORING & ALERTING**

### **Key Metrics**

1. **Availability**: >99.9% uptime
2. **Latency**:
   - Live streaming: <800ms p99
   - Playback: <500ms p99
   - API: <100ms p99
3. **Throughput**:
   - 500 concurrent streams
   - 1000 concurrent viewers
4. **Storage**:
   - <85% capacity
   - Upload rate tracking

### **Alert Thresholds**

| Alert | Threshold | Action |
|-------|-----------|--------|
| Service down | Any service unhealthy | Page on-call engineer |
| High latency | p99 > 2× target | Investigate performance |
| Storage full | >90% capacity | Add storage capacity |
| Agency limit reached | >95% quota | Notify operations team |
| AI detection lag | Queue depth >100 | Scale workers |

---

## **SUCCESS METRICS**

### **Technical KPIs**

- [ ] All services deployed and operational
- [ ] 500 concurrent camera streams supported
- [ ] 1000 concurrent viewers supported
- [ ] <800ms live streaming latency
- [ ] <300ms playback latency (local storage)
- [ ] Object detection accuracy >35% mAP
- [ ] 99.9% uptime SLA
- [ ] <11GB RAM total footprint
- [ ] <13 CPU cores total usage

### **Functional KPIs**

- [ ] 100% of cameras discoverable from Milestone
- [ ] Storage mode switchable without downtime
- [ ] Clip creation completes in <30s
- [ ] Search returns results in <200ms
- [ ] Arabic UI fully functional (RTL)
- [ ] All IAM operations working
- [ ] Audit log captures all actions

### **Project KPIs**

- [ ] Delivered on time (12 weeks)
- [ ] All phases completed
- [ ] >80% test coverage
- [ ] Zero critical bugs in production
- [ ] Documentation complete
- [ ] Training completed for operations team

---

## **POST-LAUNCH SUPPORT**

### **Week 13-14: Hypercare**

- On-call support 24/7
- Daily health checks
- Monitor alerts closely
- Quick bug fixes
- User feedback collection

### **Month 2-3: Optimization**

- Performance tuning based on real usage
- Storage optimization
- UI/UX improvements
- Feature enhancements

### **Month 4+: Maintenance**

- Regular security updates
- Database maintenance
- Storage cleanup
- Feature additions based on user requests

---

## **TRAINING PLAN**

### **Operations Team Training** (Week 12)

**Day 1: System Overview**
- Architecture walkthrough
- Service responsibilities
- Data flow explanation

**Day 2: Operations**
- Deployment procedures
- Backup and restore
- Monitoring dashboards
- Alert handling

**Day 3: Troubleshooting**
- Common issues and solutions
- Log analysis
- Performance tuning
- Rollback procedures

**Day 4: Hands-on Lab**
- Deploy to staging
- Simulate failures
- Practice recovery
- Q&A session

---

## **DOCUMENTATION DELIVERABLES**

1. **Architecture Documentation**
   - System architecture diagram
   - Service interaction diagram
   - Data flow diagram
   - Network topology

2. **API Documentation**
   - OpenAPI/Swagger spec
   - Postman collection
   - Example requests/responses
   - Error code reference

3. **Deployment Documentation**
   - Deployment guide
   - Configuration reference
   - Environment setup
   - Docker Compose reference

4. **Operations Documentation**
   - Operations runbook
   - Troubleshooting guide
   - Backup/restore procedures
   - Monitoring guide
   - Alert response playbook

5. **User Documentation**
   - User manual (English & Arabic)
   - Video tutorials
   - FAQ
   - Training materials

---

## **BUDGET ESTIMATION**

### **Development Costs** (12 weeks)

| Resource | Cost |
|----------|------|
| Backend Lead (1) | $X |
| Backend Developers (2) | $X |
| Frontend Developer (1) | $X |
| DevOps Engineer (1) | $X |
| AI Engineer (1) | $X |
| QA Engineer (1) | $X |
| **Total Development** | **$X** |

### **Infrastructure Costs** (Annual)

| Resource | Cost |
|----------|------|
| Staging Server | $X/year |
| Production Servers | $X/year |
| Storage (500TB NAS) | $X/year |
| Network Bandwidth | $X/year |
| **Total Infrastructure** | **$X/year** |

### **Licenses** (Annual)

| License | Cost |
|---------|------|
| Milestone SDK | $0 (provided) |
| All others | $0 (open source) |

---

## **CONCLUSION**

This implementation plan provides a structured approach to building the RTA CCTV system over 12 weeks. The plan is designed to:

1. **Deliver incrementally**: Each phase produces working features
2. **Minimize risk**: Critical components built early
3. **Enable testing**: Continuous integration and testing
4. **Ensure quality**: Code reviews, automated tests, performance benchmarks
5. **Facilitate operations**: Comprehensive documentation and training

**Next Steps**:
1. Approve this plan
2. Assemble the team
3. Set up development environment
4. Begin Phase 1 (Week 1)

---

**Document Version**: 1.0
**Last Updated**: 2025-01-XX
**Status**: READY FOR APPROVAL

---

## **APPENDIX: USEFUL COMMANDS**

### **Development**

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f [service-name]

# Restart service
docker-compose restart [service-name]

# Run tests
docker-compose exec go-api go test ./...

# Database migration
docker-compose exec go-api migrate up

# Load test
k6 run tests/load/stream-test.js
```

### **Monitoring**

```bash
# Check service health
curl http://localhost:8080/health

# View metrics
curl http://localhost:8080/metrics

# Check Valkey
docker-compose exec valkey valkey-cli ping

# Check PostgreSQL
docker-compose exec postgres psql -U cctv -c "SELECT 1"
```

### **Backup**

```bash
# Backup database
./scripts/backup.sh

# Backup MinIO
mc mirror minio/recordings /backup/recordings

# Backup configurations
tar -czf config-backup.tar.gz config/
```

---

**End of Implementation Plan**
