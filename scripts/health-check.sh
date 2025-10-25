#!/bin/bash
# RTA CCTV System - Health Check Script
# Validates all services are running and healthy

set -e

echo "========================================="
echo "RTA CCTV System Health Check"
echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="

FAILED_CHECKS=0
TOTAL_CHECKS=0

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_service() {
  local service_name=$1
  local endpoint=$2
  local expected_status=${3:-200}

  TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

  echo -n "Checking $service_name... "

  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$endpoint" --max-time 10 || echo "000")

  if [ "$status_code" = "$expected_status" ]; then
    echo -e "${GREEN}✓ OK${NC} (HTTP $status_code)"
  else
    echo -e "${RED}✗ FAILED${NC} (HTTP $status_code, expected $expected_status)"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
  fi
}

check_docker_service() {
  local container_name=$1

  TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

  echo -n "Checking Docker container $container_name... "

  if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
    status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "no-health")

    if [ "$status" = "healthy" ] || [ "$status" = "no-health" ]; then
      echo -e "${GREEN}✓ Running${NC}"
    else
      echo -e "${YELLOW}⚠ Running but unhealthy (status: $status)${NC}"
      FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
  else
    echo -e "${RED}✗ NOT RUNNING${NC}"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
  fi
}

# =========================================
# 1. CHECK DOCKER CONTAINERS
# =========================================
echo ""
echo "=== Docker Containers ==="
check_docker_service "cctv-valkey"
check_docker_service "cctv-postgres"
check_docker_service "cctv-minio"
check_docker_service "cctv-vms-service"
check_docker_service "cctv-stream-counter"
check_docker_service "cctv-storage-service"
check_docker_service "cctv-recording-service"
check_docker_service "cctv-metadata-service"
check_docker_service "cctv-playback-service"
check_docker_service "cctv-livekit"
check_docker_service "cctv-go-api"
check_docker_service "cctv-dashboard"
check_docker_service "cctv-prometheus"
check_docker_service "cctv-grafana"
check_docker_service "cctv-loki"
check_docker_service "cctv-alertmanager"

# =========================================
# 2. CHECK HTTP ENDPOINTS
# =========================================
echo ""
echo "=== HTTP Health Endpoints ==="
check_service "VMS Service" "http://localhost:8081/health"
check_service "Stream Counter" "http://localhost:8087/health"
check_service "Storage Service" "http://localhost:8082/health"
check_service "Recording Service" "http://localhost:8083/health"
check_service "Metadata Service" "http://localhost:8084/health"
check_service "Playback Service" "http://localhost:8090/health"
check_service "Go API" "http://localhost:8088/health"
check_service "Dashboard" "http://localhost:3000/"
check_service "LiveKit" "http://localhost:7880/"
check_service "Prometheus" "http://localhost:9090/-/healthy"
check_service "Grafana" "http://localhost:3001/api/health"
check_service "Loki" "http://localhost:3100/ready"
check_service "Alertmanager" "http://localhost:9093/-/healthy"

# =========================================
# 3. CHECK CRITICAL METRICS
# =========================================
echo ""
echo "=== Critical Metrics ==="

# Check if Prometheus is scraping targets
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
echo -n "Checking Prometheus scraping... "
up_targets=$(curl -s "http://localhost:9090/api/v1/query?query=up" | jq -r '.data.result | length')
if [ "$up_targets" -gt 10 ]; then
  echo -e "${GREEN}✓ OK${NC} ($up_targets targets)"
else
  echo -e "${YELLOW}⚠ Low target count${NC} ($up_targets targets, expected >10)"
  FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check if any services are down
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
echo -n "Checking for down services... "
down_services=$(curl -s "http://localhost:9090/api/v1/query?query=up==0" | jq -r '.data.result | length')
if [ "$down_services" -eq 0 ]; then
  echo -e "${GREEN}✓ All services up${NC}"
else
  echo -e "${RED}✗ $down_services service(s) down${NC}"
  FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check active alerts
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
echo -n "Checking for active alerts... "
firing_alerts=$(curl -s "http://localhost:9090/api/v1/query?query=ALERTS{alertstate=\"firing\"}" | jq -r '.data.result | length')
if [ "$firing_alerts" -eq 0 ]; then
  echo -e "${GREEN}✓ No alerts firing${NC}"
else
  echo -e "${YELLOW}⚠ $firing_alerts alert(s) firing${NC}"
  # Don't count as failure, just warning
fi

# =========================================
# 4. CHECK DISK SPACE
# =========================================
echo ""
echo "=== System Resources ==="

TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
echo -n "Checking disk space... "
disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$disk_usage" -lt 80 ]; then
  echo -e "${GREEN}✓ OK${NC} ($disk_usage% used)"
elif [ "$disk_usage" -lt 90 ]; then
  echo -e "${YELLOW}⚠ Warning${NC} ($disk_usage% used)"
else
  echo -e "${RED}✗ Critical${NC} ($disk_usage% used)"
  FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check Docker volumes
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
echo -n "Checking Docker volumes... "
volume_count=$(docker volume ls | grep cctv_ | wc -l)
if [ "$volume_count" -ge 7 ]; then
  echo -e "${GREEN}✓ OK${NC} ($volume_count volumes)"
else
  echo -e "${YELLOW}⚠ Missing volumes${NC} ($volume_count/7 found)"
fi

# =========================================
# SUMMARY
# =========================================
echo ""
echo "========================================="
echo "Health Check Summary"
echo "========================================="
echo "Total Checks: $TOTAL_CHECKS"
echo -e "Passed: ${GREEN}$((TOTAL_CHECKS - FAILED_CHECKS))${NC}"
if [ "$FAILED_CHECKS" -gt 0 ]; then
  echo -e "Failed: ${RED}$FAILED_CHECKS${NC}"
else
  echo -e "Failed: ${GREEN}0${NC}"
fi
echo "========================================="

if [ "$FAILED_CHECKS" -eq 0 ]; then
  echo -e "${GREEN}✓ All systems operational!${NC}"
  exit 0
else
  echo -e "${RED}✗ Some checks failed. Review logs above.${NC}"
  echo ""
  echo "Troubleshooting:"
  echo "  1. Check logs: docker-compose logs <service-name>"
  echo "  2. Restart failed services: docker-compose restart <service-name>"
  echo "  3. Check Grafana dashboard: http://localhost:3001"
  echo "  4. Check Prometheus alerts: http://localhost:9090/alerts"
  exit 1
fi
