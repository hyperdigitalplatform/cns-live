# RTA CCTV Monitoring - Quick Start Guide

## 🚀 Starting the Monitoring Stack

```bash
# Start all services including monitoring
cd /path/to/cns
docker-compose up -d

# Verify monitoring services are running
docker-compose ps | grep -E "prometheus|grafana|loki|alertmanager"

# Expected output (all healthy):
# cctv-prometheus        running
# cctv-grafana           running
# cctv-loki              running
# cctv-promtail          running
# cctv-alertmanager      running
# cctv-node-exporter     running
# cctv-cadvisor          running
# cctv-postgres-exporter running
# cctv-valkey-exporter   running
```

## 🌐 Accessing Monitoring Dashboards

### Grafana (Main Dashboard)
- **URL**: http://localhost:3001
- **Username**: `admin`
- **Password**: `admin_changeme`
- **Default Dashboard**: RTA CCTV System Overview

**First Login**:
1. Open http://localhost:3001
2. Login with credentials above
3. Navigate to: Dashboards → RTA CCTV → System Overview
4. Set time range (top-right): Last 6 hours
5. Enable auto-refresh: 30s

### Prometheus (Metrics Explorer)
- **URL**: http://localhost:9090
- **Use**: Query metrics, view targets, check alerts

**Quick Checks**:
1. View all targets: http://localhost:9090/targets
2. View active alerts: http://localhost:9090/alerts
3. Query metrics: http://localhost:9090/graph

### Alertmanager (Alert Management)
- **URL**: http://localhost:9093
- **Use**: View alerts, create silences, manage notifications

### Loki (Log Explorer)
- **Access**: Via Grafana → Explore → Select "Loki" datasource
- **Use**: Search logs, filter by service, analyze errors

## 📊 Key Metrics to Monitor

### System Health (Check First!)
```promql
# All services up? (should be 1)
up{job=~"go-api|vms-service|playback-service|livekit"}

# Any services down?
up == 0
```

### Live Streaming
```promql
# Active stream reservations by source
stream_reservations_active{source="DUBAI_POLICE"}

# Total reservations (rate)
rate(stream_reservations_total[5m])

# LiveKit rooms and participants
livekit_room_total
livekit_participant_total
```

### API Performance
```promql
# Request rate (req/s)
rate(http_requests_total{job="go-api"}[5m])

# Error rate (5xx errors)
rate(http_requests_total{job="go-api",status=~"5.."}[5m])

# API latency (95th percentile)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="go-api"}[5m]))
```

### Playback Performance
```promql
# Cache hit rate
rate(playback_cache_hits_total[5m]) / (rate(playback_cache_hits_total[5m]) + rate(playback_cache_misses_total[5m]))

# Transmux duration (95th percentile)
histogram_quantile(0.95, rate(playback_transmux_duration_seconds_bucket[5m]))
```

### Resource Usage
```promql
# CPU usage by container
rate(container_cpu_usage_seconds_total{name=~"cctv-.*"}[5m]) * 100

# Memory usage by container
container_memory_usage_bytes{name=~"cctv-.*"}

# Disk space free
node_filesystem_avail_bytes{fstype!~"tmpfs|fuse.lxcfs"}
```

## 🔍 Searching Logs

### Via Grafana Explore

1. Open Grafana → Explore (compass icon)
2. Select "Loki" datasource
3. Enter LogQL query
4. Set time range and run query

### Common LogQL Queries

```logql
# All errors across all services
{level="error"}

# Errors from specific service
{service="go-api", level="error"}

# API requests with 5xx status
{service="go-api"} | json | status_code >= 500

# Playback transmux operations
{service="playback-service"} |= "transmux"

# Slow database queries (>1 second)
{service="go-api"} | json | duration_ms > 1000

# Logs from specific container
{container="cctv-livekit"}

# Pattern search (case-insensitive)
{service="go-api"} |~ "(?i)stream.*reservation"

# Last 100 lines from playback service
{service="playback-service"} | limit 100
```

## 🚨 Understanding Alerts

### Alert Severity Levels

| Severity | Color | Response Time | Examples |
|----------|-------|---------------|----------|
| **Critical** | 🔴 Red | Immediate (0-15 min) | ServiceDown, LiveKitHighLatency, MinIOStorageFull |
| **Warning** | 🟡 Yellow | Soon (1-4 hours) | HighAPIErrorRate, PlaybackCacheLowHitRate |
| **Info** | 🔵 Blue | Informational | HighBandwidthUsage, CacheEvictionRate |

### Checking Active Alerts

**Via Prometheus**:
```bash
# Open browser
http://localhost:9090/alerts

# Or via API
curl http://localhost:9090/api/v1/alerts | jq
```

**Via Alertmanager**:
```bash
# Open browser
http://localhost:9093

# Or via API
curl http://localhost:9093/api/v1/alerts | jq
```

**Via Grafana**:
1. Navigate to: Alerting → Alert Rules
2. Filter by severity: Critical, Warning, Info
3. View recent notifications: Alerting → Notification History

### Creating Alert Silences

Sometimes you need to silence alerts during maintenance:

1. Open Alertmanager: http://localhost:9093
2. Click "Silences" → "New Silence"
3. Fill in:
   - **Matchers**: `alertname=ServiceDown`, `service=playback-service`
   - **Duration**: 2 hours
   - **Creator**: Your name
   - **Comment**: "Planned maintenance - upgrading FFmpeg"
4. Click "Create"

**Via CLI**:
```bash
# Silence alert for 2 hours
curl -X POST http://localhost:9093/api/v1/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [
      {"name": "alertname", "value": "ServiceDown", "isRegex": false},
      {"name": "service", "value": "playback-service", "isRegex": false}
    ],
    "startsAt": "2025-01-23T10:00:00Z",
    "endsAt": "2025-01-23T12:00:00Z",
    "createdBy": "ops-team",
    "comment": "Planned maintenance"
  }'
```

## 🛠️ Troubleshooting Common Issues

### Issue: Grafana shows "No data"

**Possible causes**:
1. Prometheus not scraping targets
2. Time range is wrong
3. Service not exposing metrics

**Solution**:
```bash
# 1. Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health != "up")'

# 2. Verify service metrics endpoint
curl http://localhost:8088/metrics  # Go API
curl http://localhost:8090/metrics  # Playback Service

# 3. Check Prometheus logs
docker-compose logs prometheus | tail -50

# 4. Reload Prometheus config
curl -X POST http://localhost:9090/-/reload
```

### Issue: Alerts not firing

**Solution**:
```bash
# 1. Check alert rules are loaded
curl http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.type=="alerting")'

# 2. Verify Alertmanager connection
curl http://localhost:9090/api/v1/alertmanagers | jq

# 3. Check alert evaluation
curl 'http://localhost:9090/api/v1/query?query=ALERTS{alertstate="firing"}' | jq

# 4. View Alertmanager logs
docker-compose logs alertmanager | tail -50
```

### Issue: Logs not appearing in Loki

**Solution**:
```bash
# 1. Check Promtail is running and connected
docker-compose logs promtail | tail -50

# 2. Verify Loki is receiving logs
curl http://localhost:3100/metrics | grep loki_ingester_streams_created_total

# 3. Check Loki logs
docker-compose logs loki | tail -50

# 4. Test log query directly
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={service="go-api"}' \
  --data-urlencode 'limit=10' | jq
```

### Issue: High memory usage in monitoring stack

**Solution**:
```bash
# Check container memory usage
docker stats --no-stream | grep -E "prometheus|grafana|loki"

# Reduce Prometheus retention
docker-compose exec prometheus promtool tsdb analyze /prometheus

# Adjust in docker-compose.yml:
# prometheus:
#   command:
#     - '--storage.tsdb.retention.time=15d'  # Instead of 30d
#     - '--storage.tsdb.retention.size=25GB' # Instead of 50GB

# Restart Prometheus
docker-compose restart prometheus
```

## 📈 Daily Monitoring Routine

### Morning Check (5 minutes)

1. **Open Grafana Dashboard**: http://localhost:3001
2. **Check System Health Panel**:
   - All services UP? (green = good)
   - Any red or yellow indicators?
3. **Check Active Alerts**: http://localhost:9090/alerts
   - Any critical alerts firing?
   - Follow runbook for resolution
4. **Review Key Metrics**:
   - Active stream reservations (normal range?)
   - API request rate (expected traffic?)
   - Cache hit rate (>70% = good)
   - Disk space (>20% free = good)

### Weekly Review (30 minutes)

1. **Performance Trends**:
   - Set time range to "Last 7 days"
   - Identify any degradation patterns
   - Review slow queries and high-latency endpoints
2. **Capacity Planning**:
   - Check CPU/memory trends
   - Review storage growth rate
   - Plan scaling if needed
3. **Alert Review**:
   - Were any alerts too noisy? Adjust thresholds
   - Were any incidents missed? Add new alerts
4. **Log Analysis**:
   - Search for errors in Loki
   - Identify recurring issues
   - Create tickets for fixes

## 🎯 Performance Targets

| Metric | Target | Warning | Critical |
|--------|--------|---------|----------|
| Service Uptime | 99.9% | <99% | <95% |
| API Latency (p95) | <500ms | >1s | >2s |
| API Error Rate | <0.1% | >1% | >5% |
| LiveKit Latency | <800ms | >1s | >2s |
| Cache Hit Rate | >70% | <50% | <30% |
| Playback Transmux (p95) | <10s | >20s | >30s |
| CPU Usage | <60% | >80% | >90% |
| Memory Usage | <80% | >90% | >95% |
| Disk Usage | <70% | >85% | >90% |

## 🔗 Quick Links

| Service | URL | Purpose |
|---------|-----|---------|
| **Grafana** | http://localhost:3001 | Main monitoring dashboard |
| **Prometheus** | http://localhost:9090 | Metrics query and exploration |
| **Alertmanager** | http://localhost:9093 | Alert management |
| **Go API** | http://localhost:8088 | Main API service |
| **Dashboard** | http://localhost:3000 | User-facing dashboard |
| **MinIO Console** | http://localhost:9001 | Storage management |

## 📚 Additional Resources

- **Prometheus Query Guide**: https://prometheus.io/docs/prometheus/latest/querying/basics/
- **LogQL Syntax**: https://grafana.com/docs/loki/latest/logql/
- **Grafana Dashboards**: https://grafana.com/docs/grafana/latest/dashboards/
- **Alerting Best Practices**: https://prometheus.io/docs/practices/alerting/

---

**Need Help?**
- Check logs: `docker-compose logs <service-name>`
- View documentation: `PHASE-6-MONITORING.md`
- Review alert runbooks: `config/prometheus/alerts/`
