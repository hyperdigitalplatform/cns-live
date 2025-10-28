#!/bin/bash

CAMERA_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"
SERVICE_URL="http://localhost:8085"

echo "=== Test: Start Recording, Wait, Query Sequences ==="
echo ""

# Step 1: Start 1-minute recording
echo "Step 1: Starting 1-minute manual recording..."
curl -s -X POST "$SERVICE_URL/api/v1/recordings/start" \
  -H "Content-Type: application/json" \
  -d "{\"cameraId\": \"$CAMERA_ID\", \"durationMinutes\": 1}" | python -m json.tool
echo ""

# Step 2: Check recording status
echo "Step 2: Checking recording status after 3 seconds..."
sleep 3
curl -s -X GET "$SERVICE_URL/api/v1/recordings/status/$CAMERA_ID" | python -m json.tool
echo ""

# Step 3: Wait 30 seconds
echo "Step 3: Waiting 30 seconds for recording to accumulate..."
sleep 30
echo "30 seconds elapsed"
echo ""

# Step 4: Stop recording
echo "Step 4: Stopping manual recording..."
curl -s -X POST "$SERVICE_URL/api/v1/recordings/stop" \
  -H "Content-Type: application/json" \
  -d "{\"cameraId\": \"$CAMERA_ID\"}" | python -m json.tool
echo ""

# Step 5: Check status after stopping
echo "Step 5: Checking recording status after stopping..."
sleep 2
curl -s -X GET "$SERVICE_URL/api/v1/recordings/status/$CAMERA_ID" | python -m json.tool
echo ""

# Step 6: Query sequences from last 5 minutes
echo "Step 6: Querying sequences from last 5 minutes..."
START_TIME=$(date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v-5M +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "Time range: $START_TIME to $END_TIME"
echo ""

curl -s -X POST "$SERVICE_URL/api/v1/sequences" \
  -H "Content-Type: application/json" \
  -d "{
    \"cameraId\": \"$CAMERA_ID\",
    \"startTime\": \"$START_TIME\",
    \"endTime\": \"$END_TIME\"
  }" | python -m json.tool
echo ""

# Step 7: Query timeline from last 5 minutes
echo "Step 7: Querying timeline from last 5 minutes..."
curl -s -X POST "$SERVICE_URL/api/v1/timeline" \
  -H "Content-Type: application/json" \
  -d "{
    \"cameraId\": \"$CAMERA_ID\",
    \"startTime\": \"$START_TIME\",
    \"endTime\": \"$END_TIME\"
  }" | python -m json.tool
echo ""

echo "=== Test Complete ==="
