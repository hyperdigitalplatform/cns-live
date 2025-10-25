# RTA CCTV System - Deployment Guide

**Last Updated**: October 2025
**Version**: 1.0.0

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Development Deployment](#development-deployment)
3. [Production Deployment](#production-deployment)
4. [Environment Variables](#environment-variables)
5. [Service Ports](#service-ports)
6. [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements
- **OS**: Linux (Ubuntu 20.04+ recommended), macOS, Windows with WSL2
- **CPU**: 16-32 cores (minimum 11 cores)
- **Memory**: 16-32 GB RAM (minimum 10 GB)
- **Disk**: 200 GB+ SSD (for monitoring data + cache)
- **Network**: 1 Gbps NIC

### Software Requirements
- **Docker**: 24.0+ ([Install](https://docs.docker.com/engine/install/))
- **Docker Compose**: 2.20+ ([Install](https://docs.docker.com/compose/install/))
- **Git**: 2.30+

### Optional
- **Domain Names**: For production TLS
- **SMTP Server**: For email alerts
- **Backup Storage**: NFS/S3 for backups

## Development Deployment

### 1. Clone Repository
```bash
git clone https://github.com/rta/cns.git
cd cns
```

### 2. Configure Environment
```bash
# Copy example environment file
cp .env.example .env

# Edit configuration
nano .env
```

**Minimal `.env` for development**:
```bash
# Database
POSTGRES_PASSWORD=dev_password

# MinIO
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=dev_minio_pass

# LiveKit
LIVEKIT_API_KEY=devkey
LIVEKIT_API_SECRET=devsecret

# Grafana
GRAFANA_PASSWORD=admin_dev
```

### 3. Start Services
```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### 4. Verify Deployment
```bash
# Run health check
./scripts/health-check.sh

# Access services
open http://localhost:3000  # Dashboard
open http://localhost:3001  # Grafana
open http://localhost:9090  # Prometheus
```

### 5. Initial Setup
```bash
# Create MinIO buckets (done automatically by minio-init)
# Check MinIO console: http://localhost:9001

# Access Grafana
# Username: admin
# Password: (from GRAFANA_PASSWORD in .env)

# View Prometheus targets
curl http://localhost:9090/api/v1/targets
```

## Production Deployment

### 1. Prepare Production Environment

#### Create External Volumes
```bash
# Create persistent volumes on external storage
docker volume create --name rta_cctv_valkey_data --driver local \
  --opt type=nfs --opt o=addr=nfs.server,rw --opt device=:/mnt/valkey

docker volume create --name rta_cctv_postgres_data --driver local \
  --opt type=nfs --opt o=addr=nfs.server,rw --opt device=:/mnt/postgres

docker volume create --name rta_cctv_minio_data --driver local \
  --opt type=nfs --opt o=addr=nfs.server,rw --opt device=:/mnt/minio

docker volume create --name rta_cctv_prometheus_data --driver local \
  --opt type=nfs --opt o=addr=nfs.server,rw --opt device=:/mnt/prometheus

docker volume create --name rta_cctv_grafana_data --driver local \
  --opt type=nfs --opt o=addr=nfs.server,rw --opt device=:/mnt/grafana

docker volume create --name rta_cctv_loki_data --driver local \
  --opt type=nfs --opt o=addr=nfs.server,rw --opt device=:/mnt/loki

docker volume create --name rta_cctv_alertmanager_data --driver local \
  --opt type=nfs --opt o=addr=nfs.server,rw --opt device=:/mnt/alertmanager
```

#### Create External Network
```bash
docker network create rta_cctv_network
```

#### Create Secrets
```bash
# PostgreSQL password
echo "production_postgres_password" | docker secret create postgres_password -

# MinIO credentials
echo "minio_admin" | docker secret create minio_root_user -
echo "production_minio_password" | docker secret create minio_root_password -

# LiveKit credentials
echo "production_livekit_key" | docker secret create livekit_api_key -
echo "production_livekit_secret" | docker secret create livekit_api_secret -

# Grafana credentials
echo "production_grafana_password" | docker secret create grafana_admin_password -
echo "production_grafana_db_pass" | docker secret create grafana_db_password -

# SMTP password
echo "production_smtp_password" | docker secret create smtp_password -

# Other service passwords
echo "production_storage_pass" | docker secret create storage_service_password -
echo "production_valkey_pass" | docker secret create valkey_password -
echo "postgresql://cctv:production_pass@postgres:5432/cctv" | docker secret create metadata_db_url -
```

### 2. Configure Production Environment

Create `.env.prod`:
```bash
# Production Configuration
POSTGRES_PASSWORD=/run/secrets/postgres_password
MINIO_ROOT_USER=/run/secrets/minio_root_user
MINIO_ROOT_PASSWORD=/run/secrets/minio_root_password
LIVEKIT_API_KEY=/run/secrets/livekit_api_key
LIVEKIT_API_SECRET=/run/secrets/livekit_api_secret
GRAFANA_PASSWORD=/run/secrets/grafana_admin_password

# Production URLs
MILESTONE_SERVER=milestone.production.ae
GO_API_URL=https://api.rta-cctv.ae
TURN_DOMAIN=turn.rta-cctv.ae

# SMTP Configuration
ALERTMANAGER_SMTP_HOST=smtp.rta.ae:587
ALERTMANAGER_SMTP_USER=alerts@rta-cctv.ae

# Log Level
LOG_LEVEL=info
LOG_FORMAT=json
```

### 3. Deploy to Production
```bash
# Deploy with production overrides
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Check all services
docker-compose -f docker-compose.yml -f docker-compose.prod.yml ps

# Monitor logs
docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f
```

### 4. Configure TLS/SSL

#### Option A: Using Certificates
```bash
# Place certificates
mkdir -p certs/{grafana,minio}
cp your-grafana.crt certs/grafana/grafana.crt
cp your-grafana.key certs/grafana/grafana.key
cp your-minio.crt certs/minio/public.crt
cp your-minio.key certs/minio/private.key

# Restart services
docker-compose restart grafana minio
```

#### Option B: Using Let's Encrypt
```bash
# Install certbot
sudo apt-get install certbot

# Generate certificates
sudo certbot certonly --standalone -d api.rta-cctv.ae
sudo certbot certonly --standalone -d monitoring.rta-cctv.ae

# Copy to project
sudo cp /etc/letsencrypt/live/monitoring.rta-cctv.ae/fullchain.pem certs/grafana/grafana.crt
sudo cp /etc/letsencrypt/live/monitoring.rta-cctv.ae/privkey.pem certs/grafana/grafana.key
```

### 5. Configure Reverse Proxy (Nginx/Traefik)

Example Nginx configuration:
```nginx
# /etc/nginx/sites-available/rta-cctv

# Dashboard
server {
    listen 443 ssl http2;
    server_name cctv.rta.ae;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# API
server {
    listen 443 ssl http2;
    server_name api.rta-cctv.ae;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Monitoring
server {
    listen 443 ssl http2;
    server_name monitoring.rta-cctv.ae;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 6. Set Up Backups
```bash
# Configure backup schedule (cron)
crontab -e

# Add daily backup at 2 AM
0 2 * * * /path/to/cns/scripts/backup.sh

# Test backup
./scripts/backup.sh
```

## Environment Variables

### Core Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_PASSWORD` | changeme_postgres | PostgreSQL password |
| `MINIO_ROOT_USER` | admin | MinIO root username |
| `MINIO_ROOT_PASSWORD` | changeme_minio | MinIO root password |
| `VALKEY_PASSWORD` | (none) | Valkey password (optional) |

### LiveKit Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `LIVEKIT_API_KEY` | devkey | LiveKit API key |
| `LIVEKIT_API_SECRET` | devsecret | LiveKit API secret |
| `LIVEKIT_WEBHOOK_KEY` | webhookkey | LiveKit webhook key |
| `TURN_DOMAIN` | turn.rta.ae | TURN server domain |

### Service URLs
| Variable | Default | Description |
|----------|---------|-------------|
| `MILESTONE_SERVER` | milestone:554 | Milestone VMS server |
| `MILESTONE_USER` | admin | Milestone username |
| `MILESTONE_PASS` | password | Milestone password |
| `GO_API_URL` | http://go-api:8086 | Go API internal URL |

### Monitoring Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `GRAFANA_PASSWORD` | admin_changeme | Grafana admin password |
| `GRAFANA_DB_PASSWORD` | grafana_password | Grafana database password |
| `ALERTMANAGER_SMTP_HOST` | smtp.example.com:587 | SMTP server |
| `ALERTMANAGER_SMTP_USER` | alerts@rta-cctv.ae | SMTP username |

### Logging Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | json | Log format (json, text) |

## Service Ports

| Service | Port | Protocol | Public | Description |
|---------|------|----------|--------|-------------|
| **Dashboard** | 3000 | HTTP | ✅ | React frontend |
| **Grafana** | 3001 | HTTP/HTTPS | ✅ | Monitoring dashboard |
| **Go API** | 8088 | HTTP | ✅ | Main API service |
| **VMS Service** | 8081 | HTTP | ❌ | Internal only |
| **Storage Service** | 8082 | HTTP | ❌ | Internal only |
| **Recording Service** | 8083 | HTTP | ❌ | Internal only |
| **Metadata Service** | 8084 | HTTP | ❌ | Internal only |
| **Stream Counter** | 8087 | HTTP | ❌ | Internal only |
| **Playback Service** | 8090 | HTTP | ❌ | Internal only |
| **LiveKit** | 7880 | WS/WSS | ✅ | WebRTC streaming |
| **MediaMTX** | 8888 | HTTP | ❌ | HLS output |
| **MinIO** | 9000 | HTTP | ❌ | Object storage API |
| **MinIO Console** | 9001 | HTTP | ✅ | MinIO admin UI |
| **Prometheus** | 9090 | HTTP | ⚠️ | Metrics (restrict access) |
| **Alertmanager** | 9093 | HTTP | ⚠️ | Alerts (restrict access) |
| **Loki** | 3100 | HTTP | ❌ | Log ingestion |
| **PostgreSQL** | 5432 | TCP | ❌ | Database |
| **Valkey** | 6379 | TCP | ❌ | Cache |

**Legend**:
- ✅ Public: Accessible from internet/users
- ❌ Internal: Only accessible within Docker network
- ⚠️ Restricted: Should be behind firewall/VPN

## Troubleshooting

### Services Not Starting
```bash
# Check logs
docker-compose logs <service-name>

# Check resource usage
docker stats

# Restart individual service
docker-compose restart <service-name>
```

### Port Conflicts
```bash
# Check what's using a port
sudo lsof -i :3000

# Change port in docker-compose.yml or .env
```

### Database Connection Issues
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Verify connection
docker exec -it cctv-postgres psql -U cctv -c "SELECT 1"
```

### Storage Issues
```bash
# Check MinIO status
docker exec -it cctv-minio mc admin info local

# Check disk space
df -h
docker system df
```

### Monitoring Not Working
```bash
# Verify Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check Grafana datasource
curl http://localhost:3001/api/datasources

# Restart monitoring stack
docker-compose restart prometheus grafana loki
```

## Additional Resources

- **Architecture**: See `architecture.md`
- **Operations Manual**: See `operations.md`
- **Monitoring Guide**: See `monitoring/` directory
- **Health Check**: Run `./scripts/health-check.sh`
- **Backup**: Run `./scripts/backup.sh`
