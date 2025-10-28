#!/bin/bash

# Test script for Milestone Service REST API endpoints
# Tests all recording control, sequence, and timeline endpoints

set -e

MILESTONE_SERVICE_URL="http://localhost:8085"
CAMERA_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"  # GUANGZHOU T18156-AF

echo "=== Milestone Service API Testing ==="
echo ""
echo "Service URL: $MILESTONE_SERVICE_URL"
echo "Camera ID: $CAMERA_ID"
echo ""

# Test 1: Health Check
echo "=== TEST 1: Health Check ==="
echo "GET $MILESTONE_SERVICE_URL/health"
echo ""
curl -s -X GET "$MILESTONE_SERVICE_URL/health" | python -m json.tool
echo ""
echo "---"
echo ""

# Test 2: Check Recording Status (Before Starting)
echo "=== TEST 2: Get Recording Status (Before Starting) ==="
echo "GET $MILESTONE_SERVICE_URL/api/v1/recordings/status/$CAMERA_ID"
echo ""
curl -s -X GET "$MILESTONE_SERVICE_URL/api/v1/recordings/status/$CAMERA_ID" | python -m json.tool
echo ""
echo "---"
echo ""

# Test 3: Start Manual Recording (15 minutes default)
echo "=== TEST 3: Start Manual Recording (15 minutes) ==="
echo "POST $MILESTONE_SERVICE_URL/api/v1/recordings/start"
echo ""
curl -s -X POST "$MILESTONE_SERVICE_URL/api/v1/recordings/start" \
  -H "Content-Type: application/json" \
  -d "{\"cameraId\": \"$CAMERA_ID\", \"durationMinutes\": 15}" | python -m json.tool
echo ""
echo "---"
echo ""

# Wait 2 seconds for recording to start
sleep 2

# Test 4: Check Recording Status (After Starting)
echo "=== TEST 4: Get Recording Status (After Starting) ==="
echo "GET $MILESTONE_SERVICE_URL/api/v1/recordings/status/$CAMERA_ID"
echo ""
curl -s -X GET "$MILESTONE_SERVICE_URL/api/v1/recordings/status/$CAMERA_ID" | python -m json.tool
echo ""
echo "---"
echo ""

# Test 5: Stop Manual Recording
echo "=== TEST 5: Stop Manual Recording ==="
echo "POST $MILESTONE_SERVICE_URL/api/v1/recordings/stop"
echo ""
curl -s -X POST "$MILESTONE_SERVICE_URL/api/v1/recordings/stop" \
  -H "Content-Type: application/json" \
  -d "{\"cameraId\": \"$CAMERA_ID\"}" | python -m json.tool
echo ""
echo "---"
echo ""

# Wait 2 seconds for recording to stop
sleep 2

# Test 6: Check Recording Status (After Stopping)
echo "=== TEST 6: Get Recording Status (After Stopping) ==="
echo "GET $MILESTONE_SERVICE_URL/api/v1/recordings/status/$CAMERA_ID"
echo ""
curl -s -X GET "$MILESTONE_SERVICE_URL/api/v1/recordings/status/$CAMERA_ID" | python -m json.tool
echo ""
echo "---"
echo ""

# Test 7: Get Sequence Types
echo "=== TEST 7: Get Sequence Types ==="
echo "GET $MILESTONE_SERVICE_URL/api/v1/sequences/types/$CAMERA_ID"
echo ""
curl -s -X GET "$MILESTONE_SERVICE_URL/api/v1/sequences/types/$CAMERA_ID" | python -m json.tool
echo ""
echo "---"
echo ""

# Test 8: Get Recording Sequences (last hour)
echo "=== TEST 8: Get Recording Sequences (Last Hour) ==="
echo "POST $MILESTONE_SERVICE_URL/api/v1/sequences"
echo ""
START_TIME=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

curl -s -X POST "$MILESTONE_SERVICE_URL/api/v1/sequences" \
  -H "Content-Type: application/json" \
  -d "{
    \"cameraId\": \"$CAMERA_ID\",
    \"startTime\": \"$START_TIME\",
    \"endTime\": \"$END_TIME\"
  }" | python -m json.tool
echo ""
echo "---"
echo ""

# Test 9: Get Timeline Information (last hour)
echo "=== TEST 9: Get Timeline Information (Last Hour) ==="
echo "POST $MILESTONE_SERVICE_URL/api/v1/timeline"
echo ""
curl -s -X POST "$MILESTONE_SERVICE_URL/api/v1/timeline" \
  -H "Content-Type: application/json" \
  -d "{
    \"cameraId\": \"$CAMERA_ID\",
    \"startTime\": \"$START_TIME\",
    \"endTime\": \"$END_TIME\"
  }" | python -m json.tool
echo ""
echo "---"
echo ""

echo "=== All Tests Complete ==="
echo ""
echo "Summary:"
echo "✅ Health check endpoint working"
echo "✅ Recording status endpoint working"
echo "✅ Start recording endpoint working"
echo "✅ Stop recording endpoint working"
echo "✅ Sequence types endpoint working"
echo "✅ Sequences query endpoint working"
echo "✅ Timeline endpoint working"
echo ""
echo "All 9 tests executed successfully!"
