#!/bin/bash

# Milestone Recording & Playback API Testing Script
# Server: 192.168.1.11
# User: raam

BASE_URL="https://192.168.1.11"

echo "=== Milestone XProtect Recording & Playback API Testing ==="
echo ""

# Step 1: Authenticate
echo "1. Authentication - Getting OAuth 2.0 token"
echo "POST $BASE_URL/API/IDP/connect/token"
echo ""

AUTH_RESPONSE=$(curl -k -s -X POST "$BASE_URL/API/IDP/connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "grant_type=password" \
  --data-urlencode "username=raam" \
  --data-urlencode "password=Ilove#123" \
  --data-urlencode "client_id=GrantValidatorClient")

echo "Response:"
echo "$AUTH_RESPONSE" | python -m json.tool 2>/dev/null || echo "$AUTH_RESPONSE"
echo ""

# Extract token
TOKEN=$(echo "$AUTH_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to get access token"
  exit 1
fi

echo "Token obtained successfully (length: ${#TOKEN} chars)"
echo ""

# Get camera IDs for testing
echo "2. Getting Camera IDs"
echo "GET $BASE_URL/api/rest/v1/cameras"
echo ""

CAMERAS_RESPONSE=$(curl -k -s "$BASE_URL/api/rest/v1/cameras" \
  -H "Authorization: Bearer $TOKEN")

echo "Response:"
echo "$CAMERAS_RESPONSE" | python -m json.tool 2>/dev/null || echo "$CAMERAS_RESPONSE"
echo ""

# Extract first camera ID - use python for reliable JSON parsing
CAMERA_ID=$(echo "$CAMERAS_RESPONSE" | python -c "import sys, json; data = json.load(sys.stdin); print(data['array'][0]['id'] if data.get('array') and len(data['array']) > 0 else '')" 2>/dev/null)

if [ -z "$CAMERA_ID" ]; then
  echo "ERROR: No cameras found or unable to extract camera ID"
  exit 1
fi

echo "Using Camera ID: $CAMERA_ID"
echo ""

# Test 3: Try recordings endpoints (common patterns)
echo "3. Testing Recordings Endpoints"
echo ""

# Try /api/rest/v1/recordings
echo "3a. GET $BASE_URL/api/rest/v1/recordings"
RECORDINGS_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/recordings" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$RECORDINGS_RESPONSE"
echo ""

# Try /api/rest/v1/recordings/cameras/{cameraId}
echo "3b. GET $BASE_URL/api/rest/v1/recordings/cameras/$CAMERA_ID"
CAMERA_RECORDINGS_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/recordings/cameras/$CAMERA_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$CAMERA_RECORDINGS_RESPONSE"
echo ""

# Try /api/rest/v1/cameras/{cameraId}/recordings
echo "3c. GET $BASE_URL/api/rest/v1/cameras/$CAMERA_ID/recordings"
CAMERA_RECORDINGS2_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/cameras/$CAMERA_ID/recordings" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$CAMERA_RECORDINGS2_RESPONSE"
echo ""

# Test 4: Try sequences endpoints
echo "4. Testing Sequences Endpoints"
echo ""

# Try /api/rest/v1/sequences
echo "4a. GET $BASE_URL/api/rest/v1/sequences"
SEQUENCES_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/sequences" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$SEQUENCES_RESPONSE"
echo ""

# Try /api/rest/v1/cameras/{cameraId}/sequences
echo "4b. GET $BASE_URL/api/rest/v1/cameras/$CAMERA_ID/sequences"
CAMERA_SEQUENCES_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/cameras/$CAMERA_ID/sequences" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$CAMERA_SEQUENCES_RESPONSE"
echo ""

# Test 5: Try manual recording control endpoints
echo "5. Testing Manual Recording Control Endpoints"
echo ""

# Try start recording - POST /api/rest/v1/cameras/{cameraId}/recording/start
echo "5a. POST $BASE_URL/api/rest/v1/cameras/$CAMERA_ID/recording/start"
START_RECORDING_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/api/rest/v1/cameras/$CAMERA_ID/recording/start" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")
echo "Response:"
echo "$START_RECORDING_RESPONSE"
echo ""

# Try /api/rest/v1/recordings/start
echo "5b. POST $BASE_URL/api/rest/v1/recordings/start (with camera ID in body)"
START_RECORDING2_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/api/rest/v1/recordings/start" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"cameraId\":\"$CAMERA_ID\"}")
echo "Response:"
echo "$START_RECORDING2_RESPONSE"
echo ""

# Test 6: Try output control (which includes recording control)
echo "6. Testing Output/Control Endpoints"
echo ""

# Try /api/rest/v1/cameras/{cameraId}/outputs
echo "6a. GET $BASE_URL/api/rest/v1/cameras/$CAMERA_ID/outputs"
OUTPUTS_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/cameras/$CAMERA_ID/outputs" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$OUTPUTS_RESPONSE"
echo ""

# Try /api/rest/v1/outputs
echo "6b. GET $BASE_URL/api/rest/v1/outputs"
OUTPUTS2_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/outputs" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$OUTPUTS2_RESPONSE"
echo ""

# Test 7: Try bookmarks endpoint
echo "7. Testing Bookmarks Endpoint"
echo ""

# Try /api/rest/v1/bookmarks
echo "7a. GET $BASE_URL/api/rest/v1/bookmarks"
BOOKMARKS_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/bookmarks" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$BOOKMARKS_RESPONSE"
echo ""

# Test 8: Try playback endpoints
echo "8. Testing Playback Endpoints"
echo ""

# Try /api/rest/v1/playback
echo "8a. GET $BASE_URL/api/rest/v1/playback"
PLAYBACK_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/playback" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$PLAYBACK_RESPONSE"
echo ""

# Try /api/rest/v1/cameras/{cameraId}/playback
echo "8b. GET $BASE_URL/api/rest/v1/cameras/$CAMERA_ID/playback"
CAMERA_PLAYBACK_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/cameras/$CAMERA_ID/playback" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$CAMERA_PLAYBACK_RESPONSE"
echo ""

# Test 9: Try live stream endpoint
echo "9. Testing Live Stream Endpoints"
echo ""

# Try /api/rest/v1/cameras/{cameraId}/live
echo "9a. GET $BASE_URL/api/rest/v1/cameras/$CAMERA_ID/live"
LIVE_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/cameras/$CAMERA_ID/live" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$LIVE_RESPONSE"
echo ""

# Test 10: Try investigations/evidence locks
echo "10. Testing Investigations/Evidence Endpoints"
echo ""

# Try /api/rest/v1/evidenceLocks
echo "10a. GET $BASE_URL/api/rest/v1/evidenceLocks"
EVIDENCE_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/evidenceLocks" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$EVIDENCE_RESPONSE"
echo ""

# Try /api/rest/v1/investigations
echo "10b. GET $BASE_URL/api/rest/v1/investigations"
INVESTIGATIONS_RESPONSE=$(curl -k -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/api/rest/v1/investigations" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$INVESTIGATIONS_RESPONSE"
echo ""

echo "=== Testing Complete ==="
echo ""
echo "Summary:"
echo "- Camera ID used for testing: $CAMERA_ID"
echo "- Check HTTP status codes above"
echo "- 200 = Success, 404 = Not Found, 405 = Method Not Allowed, 401 = Unauthorized"
