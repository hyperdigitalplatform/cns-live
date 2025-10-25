# RTA CCTV System - Operations Manual

**Last Updated**: October 2025
**Version**: 1.0.0

## Table of Contents
1. [Daily Operations](#daily-operations)
2. [Monitoring](#monitoring)
3. [Backup & Recovery](#backup--recovery)
4. [Incident Response](#incident-response)
5. [Maintenance](#maintenance)
6. [Performance Tuning](#performance-tuning)

## Daily Operations

### Morning Checklist (5 minutes)
```bash
# 1. Run health check
./scripts/health-check.sh

# 2. Check Grafana dashboard
open http://localhost:3001
# Verify all panels are green

# 3. Check active alerts
curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.state=="firing")'

# 4. Review logs for errors
docker-compose logs --since 24h | grep -i error | tail -50
```

### Weekly Tasks
- Review performance trends in Grafana (last 7 days)
- Check disk space: `df -h`
- Review alert thresholds and adjust if needed
- Update documentation for any changes
- Review and cleanup old logs/backups

### Monthly Tasks
- Update Docker images: `docker-compose pull`
- Review capacity planning (CPU, memory, storage)
- Test backup/restore procedures
- Security audit (check for updates, CVEs)
- Performance testing

## Monitoring

### Key Metrics to Watch

**System Health**:
```promql
# All services up?
up{job=~"go-api|vms-service|playback-service|livekit"}

# Any services down?
up == 0
```

**Performance**:
```promql
# API latency (95th percentile)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="go-api"}[5m]))

# Cache hit rate
rate(playback_cache_hits_total[5m]) / (rate(playback_cache_hits_total[5m]) + rate(playback_cache_misses_total[5m]))

# Active streams
stream_reservations_active
```

**Resources**:
```promql
# CPU usage
rate(container_cpu_usage_seconds_total{name=~"cctv-.*"}[5m]) * 100

# Memory usage
container_memory_usage_bytes{name=~"cctv-.*"}

# Disk space
node_filesystem_avail_bytes
```

### Grafana Dashboards
- **RTA CCTV System Overview**: http://localhost:3001/d/rta-overview
- **Service Health**: Check all service panels
- **Alerts**: http://localhost:9090/alerts

### Log Analysis
```bash
# Search for errors in last hour
docker-compose logs --since 1h | grep -i error

# Count errors by service
docker-compose logs --since 24h | grep -i error | cut -d'|' -f1 | sort | uniq -c

# Search Loki (via Grafana Explore)
{service="go-api", level="error"}
{service="playback-service"} |= "transmux" |= "failed"
```

## Backup & Recovery

### Automated Backups
```bash
# Daily backup (configured in cron)
0 2 * * * /path/to/cns/scripts/backup.sh

# Manual backup
./scripts/backup.sh

# Backups stored in: /mnt/backups/rta-cctv/
# Retention: 30 days (configurable via RETENTION_DAYS)
```

### What Gets Backed Up
- ✅ PostgreSQL database (all schemas)
- ✅ Prometheus data (metrics, 30 days)
- ✅ Grafana data (dashboards, users, settings)
- ✅ Loki data (logs, 31 days)
- ✅ MinIO metadata (configs)
- ✅ Configuration files
- ❌ Video recordings (too large, use separate backup)

### Restore from Backup
```bash
# List available backups
ls -lh /mnt/backups/rta-cctv/

# Restore (WARNING: overwrites current data)
./scripts/restore.sh 20250124_120000

# The script will:
# 1. Stop all services
# 2. Restore all data
# 3. Restart services
# 4. Verify restoration
```

### Video Storage Backup
```bash
# MinIO replication (for recordings)
# Configure in MinIO console or via mc

# Example: Replicate to remote MinIO
docker exec cctv-minio mc alias set remote https://backup-minio.example.com access_key secret_key
docker exec cctv-minio mc mirror local/cctv-recordings remote/cctv-recordings-backup
```

## Incident Response

### Critical Alert Response

#### ServiceDown
```bash
# 1. Identify which service
docker-compose ps | grep Exit

# 2. Check logs
docker-compose logs <service-name> | tail -100

# 3. Restart service
docker-compose restart <service-name>

# 4. If still failing, check dependencies
docker-compose logs postgres valkey minio
```

#### HighAPIErrorRate
```bash
# 1. Check error logs
docker-compose logs go-api | grep -i error | tail -50

# 2. Check specific error
curl http://localhost:8088/health

# 3. Query error metrics
curl 'http://localhost:9090/api/v1/query?query=rate(http_requests_total{job="go-api",status=~"5.."}[5m])'

# 4. If database-related, check PostgreSQL
docker-compose logs postgres
```

#### MinIOStorageFull
```bash
# 1. Check disk usage
docker exec cctv-minio df -h

# 2. Check MinIO storage
docker exec cctv-minio mc du local/cctv-recordings

# 3. Delete old recordings (if configured)
# MinIO lifecycle policies handle this automatically

# 4. Add more storage or expand volume
docker volume inspect cctv_minio_data
```

### Service Recovery Procedures

#### PostgreSQL Crash
```bash
# 1. Check logs
docker-compose logs postgres

# 2. Try restart
docker-compose restart postgres

# 3. If corrupted, restore from backup
./scripts/restore.sh <timestamp>

# 4. Verify data integrity
docker exec cctv-postgres pg_isready
docker exec cctv-postgres psql -U cctv -c "SELECT COUNT(*) FROM cameras"
```

#### LiveKit Connection Issues
```bash
# 1. Check LiveKit logs
docker-compose logs livekit

# 2. Verify TURN server
docker-compose logs coturn

# 3. Test WebRTC connection
# Use browser console to check ICE candidates

# 4. Restart LiveKit stack
docker-compose restart livekit livekit-ingress coturn
```

## Maintenance

### Updating Services
```bash
# 1. Pull latest images
docker-compose pull

# 2. Stop services (rolling update recommended)
docker-compose stop go-api
docker-compose up -d go-api

# 3. Verify
./scripts/health-check.sh

# 4. Repeat for other services
```

### Database Maintenance
```bash
# Vacuum PostgreSQL
docker exec cctv-postgres psql -U cctv -c "VACUUM ANALYZE"

# Reindex
docker exec cctv-postgres psql -U cctv -c "REINDEX DATABASE cctv"

# Check database size
docker exec cctv-postgres psql -U cctv -c "SELECT pg_size_pretty(pg_database_size('cctv'))"
```

### Cache Maintenance
```bash
# Check Valkey memory
docker exec cctv-valkey valkey-cli INFO memory

# Clear cache if needed (caution!)
docker exec cctv-valkey valkey-cli FLUSHALL

# Check keys
docker exec cctv-valkey valkey-cli KEYS "*"
```

### Log Rotation
```bash
# Docker logs (configured in docker-compose.prod.yml)
logging:
  driver: "json-file"
  options:
    max-size: "100m"
    max-file: "10"

# Manual cleanup if needed
docker system prune -af --volumes
```

## Performance Tuning

### Identify Bottlenecks
```bash
# Check slowest endpoints
# Query Prometheus for p95 latency by endpoint

# Check database slow queries
docker exec cctv-postgres psql -U cctv -c "SELECT query, calls, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10"

# Check resource usage
docker stats --no-stream
```

### Optimization Tips

**Database**:
```sql
-- Add indexes for frequently queried columns
CREATE INDEX idx_segments_camera_time ON video_segments(camera_id, start_time);

-- Analyze query plans
EXPLAIN ANALYZE SELECT * FROM video_segments WHERE camera_id = '...';
```

**Cache**:
```bash
# Increase Valkey memory (in docker-compose.yml)
command: >
  valkey-server
  --maxmemory 2gb  # Increase from 1gb

# Increase playback cache (in docker-compose.yml)
environment:
  PLAYBACK_CACHE_SIZE: 20GB  # Increase from 10GB
```

**FFmpeg**:
```yaml
# Increase playback service resources
playback-service:
  deploy:
    resources:
      limits:
        cpus: '8'  # Increase for more parallel jobs
        memory: 4G
```

### Scaling Recommendations

**When to scale**:
- CPU >80% for extended periods
- Memory >90%
- Disk >85%
- API latency >500ms (p95)
- Cache hit rate <50%

**How to scale**:
1. **Vertical**: Increase resources in docker-compose.yml
2. **Horizontal**: Run multiple instances of stateless services
3. **Storage**: Add MinIO nodes, expand volumes
4. **Database**: Add read replicas, partition tables

## Support & Escalation

### Issue Severity Levels

| Severity | Response Time | Examples |
|----------|---------------|----------|
| **P1 - Critical** | 15 minutes | Complete service outage, data loss |
| **P2 - High** | 1 hour | Single service down, degraded performance |
| **P3 - Medium** | 4 hours | Non-critical feature broken |
| **P4 - Low** | 24 hours | Minor bugs, documentation issues |

### Escalation Path
1. On-call engineer (PagerDuty)
2. Team lead
3. Senior architect
4. Management

### Contact Information
- **On-call**: ops-team@rta-cctv.ae
- **Slack**: #rta-cctv-alerts
- **Phone**: +971-XXX-XXXX

## Additional Resources
- **Architecture**: See `architecture.md`
- **Deployment**: See `deployment.md`
- **Monitoring Guide**: See `monitoring/quick-start.md`
- **Runbooks**: See `runbooks/` (to be created)
