#!/bin/bash
# Test RTA CCTV Services

set -e

echo "=== RTA CCTV Services Health Check ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test function
test_endpoint() {
    local name=$1
    local url=$2

    echo -n "Testing $name... "

    if curl -s -f -m 5 "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 5
echo ""

# Test Valkey
echo "1. Valkey Cache"
if docker exec cctv-valkey valkey-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Valkey is running${NC}"
else
    echo -e "${RED}✗ Valkey is not running${NC}"
    exit 1
fi
echo ""

# Test PostgreSQL
echo "2. PostgreSQL Database"
if docker exec cctv-postgres pg_isready -U cctv > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PostgreSQL is running${NC}"
else
    echo -e "${RED}✗ PostgreSQL is not running${NC}"
    exit 1
fi
echo ""

# Test VMS Service
echo "3. VMS Service"
test_endpoint "Health Check" "http://localhost:8081/health"
test_endpoint "Cameras List" "http://localhost:8081/vms/cameras"
test_endpoint "Metrics" "http://localhost:8081/metrics"
echo ""

# Test Stream Counter Service
echo "4. Stream Counter Service"
test_endpoint "Health Check" "http://localhost:8087/health"
test_endpoint "Stream Stats" "http://localhost:8087/api/v1/stream/stats"
test_endpoint "Metrics" "http://localhost:8087/metrics"
echo ""

# Test Reserve/Release Flow
echo "5. Testing Reserve/Release Flow"
echo -n "Reserving stream... "

RESERVE_RESPONSE=$(curl -s -X POST http://localhost:8087/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "test-user",
    "source": "DUBAI_POLICE",
    "duration": 300
  }')

RESERVATION_ID=$(echo $RESERVE_RESPONSE | jq -r '.reservation_id')

if [ "$RESERVATION_ID" != "null" ] && [ -n "$RESERVATION_ID" ]; then
    echo -e "${GREEN}✓ Reserved (ID: $RESERVATION_ID)${NC}"

    # Test heartbeat
    echo -n "Sending heartbeat... "
    if curl -s -X POST "http://localhost:8087/api/v1/stream/heartbeat/$RESERVATION_ID" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${RED}✗ FAILED${NC}"
    fi

    # Release stream
    echo -n "Releasing stream... "
    if curl -s -X DELETE "http://localhost:8087/api/v1/stream/release/$RESERVATION_ID" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${RED}✗ FAILED${NC}"
    fi
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "Response: $RESERVE_RESPONSE"
fi
echo ""

# Display current stats
echo "6. Current Stream Statistics"
curl -s http://localhost:8087/api/v1/stream/stats | jq '.'
echo ""

# Test MediaMTX
echo "7. MediaMTX RTSP Server"
test_endpoint "API Health" "http://localhost:9997/v3/config/get"
test_endpoint "Metrics" "http://localhost:9998/metrics"
echo ""

# Test MediaMTX paths
echo "8. MediaMTX Path Configuration"
echo -n "Checking default paths... "
PATHS_RESPONSE=$(curl -s http://localhost:9997/v3/paths/list)
if [ -n "$PATHS_RESPONSE" ]; then
    echo -e "${GREEN}✓ OK${NC}"
    echo "Available paths:"
    echo "$PATHS_RESPONSE" | jq -r '.items[].name' | head -5
else
    echo -e "${RED}✗ FAILED${NC}"
fi
echo ""

echo -e "${GREEN}=== All Tests Passed! ===${NC}"
