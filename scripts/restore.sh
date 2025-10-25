#!/bin/bash
# RTA CCTV System - Restore Script
# Restores data from backups

set -e

if [ -z "$1" ]; then
  echo "Usage: ./restore.sh <backup_timestamp>"
  echo "Example: ./restore.sh 20250124_120000"
  echo ""
  echo "Available backups:"
  ls -lh /mnt/backups/rta-cctv/ | grep -E "postgres_|prometheus_|grafana_|loki_|config_" | tail -10
  exit 1
fi

BACKUP_DIR="${BACKUP_DIR:-/mnt/backups/rta-cctv}"
TIMESTAMP=$1

echo "========================================="
echo "RTA CCTV System Restore"
echo "Timestamp: $TIMESTAMP"
echo "Backup Directory: $BACKUP_DIR"
echo "========================================="

# Verify backup files exist
if [ ! -f "$BACKUP_DIR/postgres_${TIMESTAMP}.sql.gz" ]; then
  echo "❌ Error: Backup files not found for timestamp $TIMESTAMP"
  exit 1
fi

read -p "⚠️  WARNING: This will OVERWRITE current data. Continue? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
  echo "Restore cancelled."
  exit 0
fi

# =========================================
# 1. STOP ALL SERVICES
# =========================================
echo ""
echo "[1/7] Stopping all services..."
cd /d/armed/github/cns
docker-compose down
echo "✓ Services stopped"

# =========================================
# 2. RESTORE POSTGRESQL DATABASE
# =========================================
echo ""
echo "[2/7] Restoring PostgreSQL database..."
docker-compose up -d postgres
sleep 10  # Wait for PostgreSQL to start
gunzip -c "$BACKUP_DIR/postgres_${TIMESTAMP}.sql.gz" | docker exec -i cctv-postgres psql -U cctv
echo "✓ PostgreSQL restore complete"

# =========================================
# 3. RESTORE PROMETHEUS DATA
# =========================================
echo ""
echo "[3/7] Restoring Prometheus data..."
docker volume rm cctv_prometheus_data || true
docker volume create cctv_prometheus_data
docker run --rm \
  -v cctv_prometheus_data:/data \
  -v "$BACKUP_DIR:/backup" \
  alpine sh -c "cd /data && tar xzf /backup/prometheus_${TIMESTAMP}.tar.gz --strip-components=1"
echo "✓ Prometheus restore complete"

# =========================================
# 4. RESTORE GRAFANA DATA
# =========================================
echo ""
echo "[4/7] Restoring Grafana data..."
docker volume rm cctv_grafana_data || true
docker volume create cctv_grafana_data
docker run --rm \
  -v cctv_grafana_data:/data \
  -v "$BACKUP_DIR:/backup" \
  alpine sh -c "cd /data && tar xzf /backup/grafana_${TIMESTAMP}.tar.gz --strip-components=1"
echo "✓ Grafana restore complete"

# =========================================
# 5. RESTORE LOKI DATA
# =========================================
echo ""
echo "[5/7] Restoring Loki data..."
docker volume rm cctv_loki_data || true
docker volume create cctv_loki_data
docker run --rm \
  -v cctv_loki_data:/data \
  -v "$BACKUP_DIR:/backup" \
  alpine sh -c "cd /data && tar xzf /backup/loki_${TIMESTAMP}.tar.gz --strip-components=1"
echo "✓ Loki restore complete"

# =========================================
# 6. RESTORE CONFIGURATION FILES
# =========================================
echo ""
echo "[6/7] Restoring configuration files..."
tar xzf "$BACKUP_DIR/config_${TIMESTAMP}.tar.gz" -C /d/armed/github/cns
echo "✓ Configuration restore complete"

# =========================================
# 7. START ALL SERVICES
# =========================================
echo ""
echo "[7/7] Starting all services..."
docker-compose up -d
echo "✓ Services started"

# =========================================
# VERIFY RESTORATION
# =========================================
echo ""
echo "Waiting 30 seconds for services to initialize..."
sleep 30

echo ""
echo "========================================="
echo "Restore Complete!"
echo "========================================="
echo ""
echo "Verifying services..."
docker-compose ps
echo ""
echo "Check service health:"
echo "  Grafana:    http://localhost:3001"
echo "  Prometheus: http://localhost:9090"
echo "  Go API:     http://localhost:8088/health"
echo ""
echo "========================================="
