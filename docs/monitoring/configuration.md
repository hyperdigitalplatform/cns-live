# Monitoring Configuration

This directory contains all configuration files for the RTA CCTV monitoring stack.

## Directory Structure

```
config/
├── prometheus/
│   ├── prometheus.yml           # Main Prometheus config
│   └── alerts/
│       ├── critical.yml         # Critical alert rules (15 rules)
│       └── performance.yml      # Performance alert rules (10 rules)
├── grafana/
│   ├── grafana.ini              # Grafana server config
│   └── provisioning/
│       ├── datasources/
│       │   └── datasources.yml  # Prometheus, Loki, PostgreSQL
│       └── dashboards/
│           ├── dashboards.yml   # Dashboard provisioning
│           └── rta-overview.json # Main dashboard
├── loki/
│   └── loki-config.yml          # Loki configuration
├── promtail/
│   └── promtail-config.yml      # Log collection config
└── alertmanager/
    ├── alertmanager.yml         # Alert routing config
    └── templates/
        └── email.tmpl           # Email notification templates
```

## Configuration Files

### Prometheus (`prometheus/prometheus.yml`)

**Scrape Targets** (13 total):
- Application services: go-api, vms-service, storage-service, recording-service, metadata-service, stream-counter, playback-service, livekit
- Infrastructure: postgres-exporter, valkey-exporter, minio, node-exporter, cadvisor

**Key Settings**:
- Scrape interval: 15s (10s for critical services)
- Evaluation interval: 15s
- Retention: 30 days or 50GB
- External labels: cluster=rta-cctv-prod, environment=production

**To modify scrape interval**:
```yaml
global:
  scrape_interval: 15s  # Change this value
```

**To add a new service**:
```yaml
scrape_configs:
  - job_name: 'my-new-service'
    static_configs:
      - targets: ['my-service:8080']
        labels:
          service: 'my-service'
          tier: 'application'
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Alert Rules (`prometheus/alerts/`)

#### Critical Alerts (`critical.yml`)
15 rules for immediate response:
- ServiceDown, HighAPIErrorRate, APIHighLatency
- LiveKitHighLatency, LiveKitRoomFailures
- StreamReservationFailures, QuotaExceeded
- PlaybackCacheLowHitRate, PlaybackTransmuxFailures
- MinIOHighErrorRate, MinIOStorageFull
- PostgreSQLDown, PostgreSQLHighConnections, PostgreSQLSlowQueries
- ValkeyDown, ValkeyHighMemory
- HighCPUUsage, HighMemoryUsage, HighDiskUsage
- ContainerHighMemory, ContainerRestarting

#### Performance Alerts (`performance.yml`)
10 rules for degradation detection:
- StreamCountDropping, HighBandwidthUsage
- RecordingQueueBacklog, SlowRecordingProcessing
- SlowPlaybackTransmux, HighPlaybackConcurrency
- HighRequestRate, SlowDatabaseQueries
- CacheEvictionRate, HighFFmpegCPU
- HighNetworkErrors, SlowMetadataIndexing

**To add a new alert**:
```yaml
groups:
  - name: my_alerts
    interval: 30s
    rules:
      - alert: MyNewAlert
        expr: my_metric > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "My alert is firing"
          description: "Metric value is {{ $value }}"
```

**Alert Severity Levels**:
- `critical`: Immediate action required, repeat every 3 hours
- `warning`: Action needed soon, repeat every 12 hours
- `info`: Informational, daily digest

### Grafana (`grafana/grafana.ini`)

**Key Settings**:
- HTTP port: 3001
- Database: PostgreSQL
- Default credentials: admin / admin_changeme
- Default dashboard: rta-overview.json

**Security Settings**:
```ini
[security]
admin_user = admin
admin_password = admin_changeme  # CHANGE IN PRODUCTION!
secret_key = SW2YcwTIb9zpOOhoPsMm  # CHANGE IN PRODUCTION!
```

**To change admin password**:
1. Edit `grafana.ini`:
   ```ini
   admin_password = your_secure_password
   ```
2. Set environment variable:
   ```bash
   export GRAFANA_PASSWORD=your_secure_password
   ```
3. Restart Grafana:
   ```bash
   docker-compose restart grafana
   ```

### Datasources (`grafana/provisioning/datasources/datasources.yml`)

**Pre-configured datasources**:
1. **Prometheus**: Default datasource for metrics
   - URL: http://prometheus:9090
   - Scrape interval: 15s
2. **Loki**: Log aggregation
   - URL: http://loki:3100
   - Max lines: 1000
3. **PostgreSQL**: Direct database queries
   - URL: postgres:5432
   - Database: rta_cctv

### Loki (`loki/loki-config.yml`)

**Key Settings**:
- Retention: 31 days (744 hours)
- Ingestion rate: 50 MB/s (burst: 100 MB/s)
- Max query parallelism: 32
- Storage: Filesystem (production should use S3/GCS)

**To change retention**:
```yaml
limits_config:
  retention_period: 744h  # Change this (hours)

compactor:
  retention_enabled: true
  retention_delete_delay: 2h
```

**To use S3 storage (production)**:
```yaml
storage_config:
  aws:
    s3: s3://region/bucket-name
    dynamodb:
      dynamodb_url: dynamodb://region
```

### Promtail (`promtail/promtail-config.yml`)

**Log Sources**:
- Docker containers (via Docker socket)
- System logs (/var/log)
- Nginx access/error logs
- PostgreSQL logs

**Pipeline Stages**:
1. JSON parsing for structured logs
2. Label extraction (level, service, status_code)
3. Timestamp extraction (RFC3339)
4. Drop rules for noisy logs (health checks, debug logs)

**To add a new log source**:
```yaml
scrape_configs:
  - job_name: my-service
    static_configs:
      - targets:
          - localhost
        labels:
          job: my-service
          __path__: /var/log/my-service/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            message: msg
      - labels:
          level:
```

### Alertmanager (`alertmanager/alertmanager.yml`)

**Notification Channels**:
1. **Email (SMTP)**:
   ```yaml
   global:
     smtp_from: 'alerts@rta-cctv.ae'
     smtp_smarthost: 'smtp.example.com:587'
     smtp_auth_username: 'alerts@rta-cctv.ae'
     smtp_auth_password: 'changeme'  # Use secrets in production!
   ```

2. **Webhook** (Go API integration):
   ```yaml
   webhook_configs:
     - url: 'http://go-api:8088/api/v1/alerts/webhook'
       send_resolved: true
   ```

3. **Slack** (optional, commented):
   ```yaml
   slack_configs:
     - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
       channel: '#alerts-critical'
   ```

4. **PagerDuty** (optional, commented):
   ```yaml
   pagerduty_configs:
     - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
   ```

**Alert Routing**:
- Critical alerts → ops-team + oncall (email), immediate
- Warning alerts → ops-team (email), 12h repeat
- Info alerts → ops-team (email), daily digest
- Service-specific routes → streaming-team, playback-team, infrastructure-team

**Inhibition Rules**:
- ServiceDown inhibits all other alerts for that service
- HighDiskUsage inhibits DiskFilling
- PostgreSQLDown inhibits PostgreSQLSlowQueries

**To configure email notifications**:
1. Edit `alertmanager.yml`:
   ```yaml
   global:
     smtp_smarthost: 'smtp.gmail.com:587'
     smtp_from: 'your-email@gmail.com'
     smtp_auth_username: 'your-email@gmail.com'
     smtp_auth_password: 'your-app-password'
   ```
2. For Gmail, use an [App Password](https://support.google.com/accounts/answer/185833)
3. Restart Alertmanager:
   ```bash
   docker-compose restart alertmanager
   ```

### Email Templates (`alertmanager/templates/email.tmpl`)

**Four templates**:
1. `email.default.html`: Standard alerts
2. `email.critical.html`: Critical alerts (red banner, urgent)
3. `email.warning.html`: Warning alerts (yellow banner)
4. `email.info.html`: Informational alerts (blue, digest)

**To customize templates**:
Edit `email.tmpl` and add your custom HTML. Variables available:
- `.GroupLabels` - Alert labels (alertname, cluster, service)
- `.Alerts` - Array of alerts
- `.Annotations` - Alert annotations (summary, description)
- `.StartsAt` - Alert start time

## Environment Variables

Create a `.env` file in the project root:

```bash
# Grafana
GRAFANA_PASSWORD=admin_changeme
GRAFANA_DB_PASSWORD=grafana_password

# PostgreSQL (for services + Grafana)
POSTGRES_PASSWORD=changeme_postgres

# Email (Alertmanager)
ALERTMANAGER_SMTP_HOST=smtp.example.com:587
ALERTMANAGER_SMTP_USER=alerts@rta-cctv.ae
ALERTMANAGER_SMTP_PASSWORD=changeme

# Monitoring (optional)
PROMETHEUS_RETENTION_TIME=30d
PROMETHEUS_RETENTION_SIZE=50GB
LOKI_RETENTION_PERIOD=744h
```

## Starting the Monitoring Stack

```bash
# Start all services
docker-compose up -d

# Verify monitoring services
docker-compose ps | grep -E "prometheus|grafana|loki|alertmanager"

# Check logs
docker-compose logs -f prometheus grafana loki

# Access dashboards
# Grafana:      http://localhost:3001 (admin/admin_changeme)
# Prometheus:   http://localhost:9090
# Alertmanager: http://localhost:9093
```

## Maintenance

### Backup Configurations

```bash
# Backup all config files
tar czf monitoring-config-backup-$(date +%Y%m%d).tar.gz config/

# Backup Prometheus data
docker run --rm -v cctv_prometheus_data:/data -v $(pwd)/backups:/backup \
  alpine tar czf /backup/prometheus-data-$(date +%Y%m%d).tar.gz /data

# Backup Grafana dashboards
curl -X GET http://admin:admin@localhost:3001/api/search?type=dash-db | \
  jq -r '.[].uri' | \
  xargs -I {} curl -X GET http://admin:admin@localhost:3001/api/dashboards/{} \
  > grafana-dashboards-backup-$(date +%Y%m%d).json
```

### Reload Configurations

```bash
# Reload Prometheus config (without restart)
curl -X POST http://localhost:9090/-/reload

# Reload Alertmanager config
curl -X POST http://localhost:9093/-/reload

# Restart Grafana (required for config changes)
docker-compose restart grafana

# Restart Loki
docker-compose restart loki
```

### Update Monitoring Stack

```bash
# Update to latest versions
docker-compose pull prometheus grafana loki alertmanager

# Restart services with new images
docker-compose up -d --force-recreate prometheus grafana loki alertmanager
```

## Troubleshooting

### Prometheus Not Scraping

```bash
# Check targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health != "up")'

# Check Prometheus logs
docker-compose logs prometheus | grep -i error

# Validate config
docker-compose exec prometheus promtool check config /etc/prometheus/prometheus.yml
```

### Alerts Not Firing

```bash
# Check alert rules
docker-compose exec prometheus promtool check rules /etc/prometheus/alerts/*.yml

# View rule status
curl http://localhost:9090/api/v1/rules | jq

# Check Alertmanager connection
curl http://localhost:9090/api/v1/alertmanagers | jq
```

### Grafana Connection Issues

```bash
# Test Prometheus datasource
curl -X POST http://localhost:3001/api/datasources/proxy/1/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query":"up"}'

# Check Grafana logs
docker-compose logs grafana | grep -i error

# Restart Grafana
docker-compose restart grafana
```

## Security Best Practices

1. **Change Default Passwords**:
   - Grafana admin password
   - SMTP credentials
   - PostgreSQL passwords

2. **Use Secrets Management**:
   - Docker secrets
   - HashiCorp Vault
   - AWS Secrets Manager

3. **Enable TLS**:
   - Use HTTPS for Grafana
   - Enable TLS for SMTP
   - Use TLS for Alertmanager webhook

4. **Network Isolation**:
   - Put monitoring in separate Docker network
   - Use firewall rules
   - Restrict access to monitoring ports

5. **Authentication**:
   - Enable OAuth for Grafana
   - Use API keys for Prometheus
   - Implement basic auth for Alertmanager

## Documentation

- **Full Documentation**: `../PHASE-6-MONITORING.md`
- **Quick Start Guide**: `../MONITORING-QUICK-START.md`
- **Project Status**: `../PROJECT-STATUS.md`
