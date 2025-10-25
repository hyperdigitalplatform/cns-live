#!/bin/bash
# RTA CCTV System - Backup Script
# Backs up all critical data (database, Prometheus, Grafana, Loki, MinIO)

set -e

BACKUP_DIR="${BACKUP_DIR:-/mnt/backups/rta-cctv}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS="${RETENTION_DAYS:-30}"

echo "========================================="
echo "RTA CCTV System Backup"
echo "Timestamp: $TIMESTAMP"
echo "Backup Directory: $BACKUP_DIR"
echo "========================================="

# Create backup directory
mkdir -p "$BACKUP_DIR"

# =========================================
# 1. BACKUP POSTGRESQL DATABASE
# =========================================
echo ""
echo "[1/6] Backing up PostgreSQL database..."
docker exec cctv-postgres pg_dumpall -U cctv | gzip > "$BACKUP_DIR/postgres_${TIMESTAMP}.sql.gz"
echo "✓ PostgreSQL backup complete: postgres_${TIMESTAMP}.sql.gz"

# =========================================
# 2. BACKUP PROMETHEUS DATA
# =========================================
echo ""
echo "[2/6] Backing up Prometheus data..."
docker run --rm \
  -v cctv_prometheus_data:/data \
  -v "$BACKUP_DIR:/backup" \
  alpine tar czf "/backup/prometheus_${TIMESTAMP}.tar.gz" /data
echo "✓ Prometheus backup complete: prometheus_${TIMESTAMP}.tar.gz"

# =========================================
# 3. BACKUP GRAFANA DATA
# =========================================
echo ""
echo "[3/6] Backing up Grafana data (dashboards, users, settings)..."
docker run --rm \
  -v cctv_grafana_data:/data \
  -v "$BACKUP_DIR:/backup" \
  alpine tar czf "/backup/grafana_${TIMESTAMP}.tar.gz" /data
echo "✓ Grafana backup complete: grafana_${TIMESTAMP}.tar.gz"

# =========================================
# 4. BACKUP LOKI DATA
# =========================================
echo ""
echo "[4/6] Backing up Loki data (logs)..."
docker run --rm \
  -v cctv_loki_data:/data \
  -v "$BACKUP_DIR:/backup" \
  alpine tar czf "/backup/loki_${TIMESTAMP}.tar.gz" /data
echo "✓ Loki backup complete: loki_${TIMESTAMP}.tar.gz"

# =========================================
# 5. BACKUP MINIO METADATA (not videos)
# =========================================
echo ""
echo "[5/6] Backing up MinIO metadata..."
docker exec cctv-minio mc alias set local http://localhost:9000 ${MINIO_ROOT_USER:-admin} ${MINIO_ROOT_PASSWORD:-changeme_minio}
docker exec cctv-minio mc admin config export local > "$BACKUP_DIR/minio_config_${TIMESTAMP}.json"
echo "✓ MinIO metadata backup complete: minio_config_${TIMESTAMP}.json"
echo "⚠️  Note: Video recordings NOT backed up (too large). Use separate storage backup."

# =========================================
# 6. BACKUP CONFIGURATION FILES
# =========================================
echo ""
echo "[6/6] Backing up configuration files..."
cd /d/armed/github/cns
tar czf "$BACKUP_DIR/config_${TIMESTAMP}.tar.gz" \
  docker-compose.yml \
  docker-compose.prod.yml \
  .env.example \
  config/ \
  --exclude='config/kong/Dockerfile' \
  --exclude='config/minio/Dockerfile'
echo "✓ Configuration backup complete: config_${TIMESTAMP}.tar.gz"

# =========================================
# CLEANUP OLD BACKUPS
# =========================================
echo ""
echo "Cleaning up backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -name "*.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "*.json" -mtime +$RETENTION_DAYS -delete
echo "✓ Cleanup complete"

# =========================================
# BACKUP SUMMARY
# =========================================
echo ""
echo "========================================="
echo "Backup Complete!"
echo "========================================="
echo "Backup files:"
ls -lh "$BACKUP_DIR"/*${TIMESTAMP}*
echo ""
echo "Total backup size:"
du -sh "$BACKUP_DIR"/*${TIMESTAMP}* | awk '{sum+=$1} END {print sum " MB"}'
echo ""
echo "Retention: Keeping backups for $RETENTION_DAYS days"
echo "========================================="
