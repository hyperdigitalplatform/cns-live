# RTA CCTV System - Documentation

Welcome to the RTA CCTV Video Management System documentation.

## 📚 Documentation Structure

### Core Documentation
- **[Architecture](architecture.md)** - System architecture, technology stack, data flow
- **[Deployment](deployment.md)** - Installation, configuration, production setup
- **[Operations](operations.md)** - Daily operations, monitoring, incident response
- **[API Reference](api.md)** - REST API endpoints and usage (TODO)

### Phase Documentation
- **[phases/](phases/)** - Detailed phase-by-phase implementation documentation
  - Phase 1: Core Infrastructure
  - Phase 2: Storage & Recording
  - Phase 3: Live Streaming (Week 5 & 6)
  - Phase 4: Dashboard & Enhancements
  - Phase 6: Monitoring & Operations

### Monitoring
- **[monitoring/quick-start.md](monitoring/quick-start.md)** - Quick reference for daily monitoring
- **[monitoring/configuration.md](monitoring/configuration.md)** - Detailed monitoring configuration

## 🚀 Quick Start

### For Developers
1. Read [Architecture](architecture.md) to understand the system
2. Follow [Deployment](deployment.md) for local setup
3. Review phase docs in [phases/](phases/) for implementation details

### For Operations
1. Read [Operations](operations.md) for daily tasks
2. Use [monitoring/quick-start.md](monitoring/quick-start.md) for monitoring
3. Review [Deployment](deployment.md) for production setup

### For Management
1. Read [../PROJECT-STATUS.md](../PROJECT-STATUS.md) for current status
2. Review [Architecture](architecture.md) for system overview
3. Check [phases/](phases/) for completed work

## 📖 Key Documents

| Document | Purpose | Audience |
|----------|---------|----------|
| [Architecture](architecture.md) | System design, components, data flow | Developers, Architects |
| [Multi-Viewer Streaming](MULTI_VIEWER_STREAMING.md) | Multi-viewer architecture, LiveKit SFU | Developers, Architects |
| [PTZ Controls Requirements](PTZ_CONTROLS_REQUIREMENTS.md) | PTZ controls UI/UX specifications | Developers, UI/UX |
| [Deployment](deployment.md) | Installation, configuration | DevOps, Developers |
| [Operations](operations.md) | Daily operations, troubleshooting | Operations, SRE |
| [API Reference](api.md) | API endpoints, usage | Developers, Integrators |
| [Project Status](../PROJECT-STATUS.md) | Current status, progress | Management, Team |
| [Monitoring Quick Start](monitoring/quick-start.md) | Daily monitoring tasks | Operations, SRE |

## 🎯 Documentation by Role

### Software Developer
- Start: [Architecture](architecture.md) → [phases/](phases/)
- Reference: API docs, service READMEs
- Testing: [../tests/README.md](../tests/README.md)

### DevOps Engineer
- Start: [Deployment](deployment.md) → [Operations](operations.md)
- Reference: [monitoring/configuration.md](monitoring/configuration.md)
- Scripts: [../scripts/](../scripts/)

### Operations / SRE
- Daily: [monitoring/quick-start.md](monitoring/quick-start.md)
- Incidents: [Operations](operations.md#incident-response)
- Maintenance: [Operations](operations.md#maintenance)

### System Architect
- Overview: [Architecture](architecture.md)
- Status: [../PROJECT-STATUS.md](../PROJECT-STATUS.md)
- Planning: [phases/](phases/)

## 📊 System Status

**Current Version**: 1.0.0
**Completion**: 97%
**Status**: Production Ready (except Phase 5 Auth & Phase 7 Object Detection)

**Completed**:
- ✅ Core Infrastructure
- ✅ Storage & Recording
- ✅ Live Streaming (LiveKit)
- ✅ Playback Service
- ✅ React Dashboard
- ✅ Monitoring Stack

**TODO**:
- ⏸️ Phase 5: Security & Auth (RTA IAM, JWT, RBAC)
- ⏸️ Phase 7: Object Detection (YOLOv8 Nano)
- ⏸️ Automated Testing (Unit, Integration, E2E)

## 🔗 External Resources

- **LiveKit Documentation**: https://docs.livekit.io/
- **MinIO Documentation**: https://min.io/docs/
- **Prometheus Documentation**: https://prometheus.io/docs/
- **Grafana Documentation**: https://grafana.com/docs/
- **Docker Compose**: https://docs.docker.com/compose/

## 📝 Contributing to Documentation

When updating documentation:
1. Keep it concise and actionable
2. Use examples and code snippets
3. Update the table of contents
4. Add diagrams where helpful
5. Link related documents
6. Update the last modified date

## 📞 Support

- **Technical Issues**: ops-team@rta-cctv.ae
- **Documentation Issues**: Create GitHub issue
- **Emergency**: See [Operations](operations.md#support--escalation)
