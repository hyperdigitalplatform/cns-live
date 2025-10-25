# RTA CCTV System - Architecture Documentation

**Last Updated**: October 2025
**Version**: 1.0.0
**Status**: Production Ready (97% Complete)

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Technology Stack](#technology-stack)
4. [Microservices](#microservices)
5. [Data Flow](#data-flow)
6. [Scalability](#scalability)
7. [Security](#security)

## System Overview

The RTA CCTV Video Management System is a comprehensive, microservices-based solution for managing live streaming and recorded video from multiple camera sources (Dubai Police, Metro, Taxi, etc.). The system is built with:

- **Low resource footprint**: H.264 codec exclusively with stream copy (no transcoding)
- **High scalability**: Stateless microservices, horizontal scaling
- **Real-time streaming**: LiveKit SFU with <800ms latency
- **Efficient playback**: FFmpeg transmuxing at ~500x realtime
- **Complete observability**: Prometheus, Grafana, Loki monitoring stack

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    React Dashboard (3000)                        │
│          Live View | Playback | Analytics | Monitoring          │
└────────────┬─────────────────────────────────────────────────────┘
             │
┌────────────▼──────────────────────────────────────────────────┐
│                    Go API Service (8088/8086)                   │
│   Stream Management | Camera API | Playback Orchestration      │
└────┬──────────┬──────────┬──────────┬──────────┬──────────────┘
     │          │          │          │          │
┌────▼───┐ ┌───▼────┐ ┌───▼─────┐ ┌─▼────┐ ┌──▼──────┐
│LiveKit │ │ VMS    │ │ Stream  │ │Playback││ Metadata│
│ (7880) │ │Service │ │ Counter │ │Service ││ Service │
│        │ │ (8081) │ │ (8087)  │ │ (8090) ││ (8084)  │
└────────┘ └────┬───┘ └─────────┘ └───┬────┘ └─────────┘
                │                      │
           ┌────▼────┐          ┌──────▼────────┐
           │MediaMTX │          │ Storage       │
           │ (8888)  │          │ Service (8082)│
           └─────────┘          └───────┬───────┘
                                        │
                                 ┌──────▼──────┐
                                 │  MinIO      │
                                 │  (9000)     │
                                 └─────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Monitoring Stack                              │
│  Prometheus (9090) | Grafana (3001) | Loki (3100)              │
└──────────────────────────────────────────────────────────────────┘
```

## Technology Stack

### Backend Services
- **Language**: Go 1.21
- **Framework**: Chi router (lightweight, fast)
- **Architecture**: Clean Architecture (Domain → UseCase → Repository → Delivery)
- **Database**: PostgreSQL 15 (metadata, users)
- **Cache**: Valkey 7.2 (Redis-compatible, quota management)
- **Storage**: MinIO (S3-compatible object storage)

### Frontend
- **Framework**: React 18.2 + TypeScript 5.3
- **Build Tool**: Vite 5.0
- **Styling**: Tailwind CSS 3.4
- **State Management**: Zustand 4.4
- **Video Player**: HLS.js 1.4, LiveKit Client SDK 2.0

### Streaming
- **Live Streaming**: LiveKit SFU (Selective Forwarding Unit)
- **RTSP Bridge**: MediaMTX (RTSP → HLS/WebRTC)
- **NAT Traversal**: coturn (TURN server)
- **Video Processing**: FFmpeg 6.1 (H.264 stream copy)

### Monitoring
- **Metrics**: Prometheus 2.48.0 (30-day retention)
- **Visualization**: Grafana 10.2.3
- **Logs**: Loki 2.9.3 + Promtail 2.9.3 (31-day retention)
- **Alerting**: Alertmanager 0.26.0
- **Exporters**: Node, cAdvisor, PostgreSQL, Valkey

### Infrastructure
- **Containerization**: Docker 24.0
- **Orchestration**: Docker Compose 2.23
- **API Gateway**: Kong 3.4 (rate limiting, auth)
- **Load Balancing**: Nginx (HLS delivery)

## Microservices

### Core Services

#### 1. **VMS Service** (Port 8081)
- **Purpose**: Milestone VMS integration, camera management
- **Features**: Camera discovery, RTSP URL generation, PTZ control
- **Tech**: Go, Milestone SDK wrapper
- **Dependencies**: Milestone VMS

#### 2. **Stream Counter Service** (Port 8087)
- **Purpose**: Agency-based quota enforcement
- **Features**: Stream reservation, heartbeat monitoring, quota limits
- **Tech**: Go, Valkey (Lua scripts for atomic operations)
- **Dependencies**: Valkey

#### 3. **Storage Service** (Port 8082)
- **Purpose**: Storage orchestration (MinIO abstraction)
- **Features**: Segment upload, retrieval, lifecycle management
- **Tech**: Go, MinIO SDK
- **Dependencies**: MinIO, PostgreSQL

#### 4. **Recording Service** (Port 8083)
- **Purpose**: Scheduled video recording
- **Features**: FFmpeg-based recording, 60-second segments, upload to MinIO
- **Tech**: Go, FFmpeg wrapper
- **Dependencies**: Storage Service, VMS Service

#### 5. **Metadata Service** (Port 8084)
- **Purpose**: Video metadata, tags, search
- **Features**: Full-text search, incident tracking, annotations
- **Tech**: Go, PostgreSQL
- **Dependencies**: PostgreSQL

#### 6. **Playback Service** (Port 8090)
- **Purpose**: Unified playback (MinIO + Milestone)
- **Features**: Source detection, FFmpeg transmuxing, HLS generation, LRU cache
- **Tech**: Go, FFmpeg
- **Dependencies**: MinIO, Milestone VMS

#### 7. **Go API Service** (Port 8088)
- **Purpose**: Central API orchestrator
- **Features**: Stream management, LiveKit integration, PTZ control
- **Tech**: Go, LiveKit SDK
- **Dependencies**: All other services

### Frontend

#### 8. **React Dashboard** (Port 3000)
- **Purpose**: User interface
- **Features**: Live view (6 layouts), playback, PTZ controls, timeline
- **Tech**: React, TypeScript, Tailwind CSS
- **Dependencies**: Go API

### Infrastructure Services

#### 9. **LiveKit SFU** (Port 7880)
- **Purpose**: WebRTC streaming
- **Features**: Simulcast (3 quality layers), <800ms latency
- **Tech**: LiveKit Server
- **Dependencies**: Valkey (for room state)

#### 10. **MediaMTX** (Port 8888)
- **Purpose**: RTSP to HLS/WebRTC bridge
- **Features**: RTSP ingestion, HLS output
- **Tech**: MediaMTX
- **Dependencies**: None

## Data Flow

### Live Streaming Flow
```
Camera (RTSP) → MediaMTX → LiveKit Ingress → LiveKit SFU → Dashboard (WebRTC)
                                   ↓
                           Stream Counter (quota check)
                                   ↓
                              Go API (orchestration)
```

### Recording Flow
```
Camera (RTSP) → Recording Service → FFmpeg (H.264 copy) → 60s segments → Storage Service → MinIO
                                                                               ↓
                                                                      Metadata Service
```

### Playback Flow
```
Dashboard → Go API → Playback Service → Source Detection
                                            ├─→ MinIO (if ≥80% coverage)
                                            └─→ Milestone VMS (fallback)
                                                      ↓
                                                FFmpeg (transmux to HLS)
                                                      ↓
                                                LRU Cache (10GB)
                                                      ↓
                                                Nginx (delivery)
                                                      ↓
                                                Dashboard (HLS.js)
```

### Monitoring Flow
```
All Services (expose /metrics) → Prometheus (scrape 15s) → Grafana (visualize)
All Containers (stdout/stderr) → Promtail (collect) → Loki (aggregate) → Grafana (explore)
Prometheus (alert rules) → Alertmanager (route) → Email/Webhook
```

## Scalability

### Horizontal Scaling
- ✅ **Stateless Services**: Go API, Playback, Storage, Metadata can scale horizontally
- ✅ **Load Balancer**: Kong Gateway with round-robin
- ✅ **Database Read Replicas**: PostgreSQL supports read replicas
- ✅ **Distributed Cache**: Valkey cluster mode
- ✅ **Distributed Storage**: MinIO distributed mode (4+ nodes)

### Vertical Scaling
- ✅ **CPU**: Add more cores for FFmpeg parallel processing
- ✅ **Memory**: Increase cache sizes (Valkey, playback LRU cache)
- ✅ **Storage**: Add MinIO nodes, expand volumes

### Performance Targets
- **Live Streaming**: <800ms latency, 100 viewers per camera
- **Playback Transmux**: ~500x realtime (1hr video in 7s)
- **API Response**: p95 < 500ms
- **Cache Hit Rate**: >70%
- **System Resources**: 11-25 cores, 10-22GB RAM

## Security

### Current Implementation (Development)
- ✅ JWT tokens for LiveKit (1-hour expiration)
- ✅ Presigned URLs for MinIO (time-limited)
- ✅ Service-to-service API keys
- ✅ CORS configuration
- ✅ Health check endpoints

### Production Requirements (Phase 5 TODO)
- [ ] JWT authentication middleware for Dashboard
- [ ] RTA IAM integration
- [ ] Role-Based Access Control (RBAC)
- [ ] Rate limiting per user/agency
- [ ] Audit logging
- [ ] TLS/SSL certificates
- [ ] Network segmentation
- [ ] Secrets management (Vault)

## Additional Resources

- **API Documentation**: See `api.md`
- **Deployment Guide**: See `deployment.md`
- **Operations Manual**: See `operations.md`
- **Phase Documentation**: See `phases/` directory
- **Monitoring Guide**: See `monitoring/` directory
- **Project Status**: See `../PROJECT-STATUS.md`
