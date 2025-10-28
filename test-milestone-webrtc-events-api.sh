#!/bin/bash

# Milestone WebRTC & Events API Testing Script
# Server: 192.168.1.11
# User: raam

BASE_URL="https://192.168.1.11"

echo "=== Milestone XProtect WebRTC & Events API Testing ==="
echo ""

# Step 1: Authenticate
echo "1. Authentication - Getting OAuth 2.0 token"
AUTH_RESPONSE=$(curl -k -s -X POST "$BASE_URL/API/IDP/connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "grant_type=password" \
  --data-urlencode "username=raam" \
  --data-urlencode "password=Ilove#123" \
  --data-urlencode "client_id=GrantValidatorClient")

TOKEN=$(echo "$AUTH_RESPONSE" | python -c "import sys, json; data = json.load(sys.stdin); print(data.get('access_token', ''))" 2>/dev/null)

if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to get access token"
  exit 1
fi

echo "✅ Token obtained (length: ${#TOKEN} chars)"
echo ""

# Get camera ID
CAMERA_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"
echo "Using Camera ID: $CAMERA_ID"
echo ""

# Test 1: Events API - List Events
echo "=== EVENTS API ==="
echo ""
echo "1. GET /api/rest/v1/events - List past events"
curl -k -s -w "\nHTTP_STATUS:%{http_code}\n" "$BASE_URL/api/rest/v1/events" \
  -H "Authorization: Bearer $TOKEN" | python -m json.tool 2>/dev/null
echo ""

# Test 2: Events API - Trigger Event (POST)
echo "2. POST /api/rest/v1/events - Trigger new event"
EVENT_PAYLOAD=$(cat <<EOF
{
  "source": {
    "id": "$CAMERA_ID",
    "type": "Camera"
  },
  "type": {
    "id": "UserDefinedEvent"
  },
  "message": "Test event from API",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)

echo "Request Body:"
echo "$EVENT_PAYLOAD" | python -m json.tool
echo ""

curl -k -s -w "\nHTTP_STATUS:%{http_code}\n" -X POST "$BASE_URL/api/rest/v1/events" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$EVENT_PAYLOAD" | python -m json.tool 2>/dev/null
echo ""

# Test 3: WebRTC API - Check Session Endpoint
echo "=== WEBRTC API ==="
echo ""
echo "3. POST /webRTC/session - Create WebRTC session (check if endpoint exists)"
WEBRTC_PAYLOAD=$(cat <<EOF
{
  "cameraId": "$CAMERA_ID",
  "offer": {
    "type": "offer",
    "sdp": "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\n"
  }
}
EOF
)

curl -k -s -w "\nHTTP_STATUS:%{http_code}\n" -X POST "$BASE_URL/webRTC/session" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$WEBRTC_PAYLOAD" 2>&1
echo ""

# Try with /api/rest/v1 prefix
echo "4. POST /api/rest/v1/webRTC/session - Create WebRTC session (alternative path)"
curl -k -s -w "\nHTTP_STATUS:%{http_code}\n" -X POST "$BASE_URL/api/rest/v1/webRTC/session" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$WEBRTC_PAYLOAD" 2>&1
echo ""

# Test 4: Bookmarks API - Create Bookmark
echo "=== BOOKMARKS API ==="
echo ""
echo "5. POST /api/rest/v1/bookmarks - Create bookmark"
BOOKMARK_PAYLOAD=$(cat <<EOF
{
  "header": "Test Recording",
  "description": "API test recording bookmark",
  "timeBegin": "$(date -u -d '10 minutes ago' +%Y-%m-%dT%H:%M:%SZ)",
  "timeEnd": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "reference": "TEST-$(date +%s)",
  "deviceId": "$CAMERA_ID"
}
EOF
)

echo "Request Body:"
echo "$BOOKMARK_PAYLOAD" | python -m json.tool
echo ""

curl -k -s -w "\nHTTP_STATUS:%{http_code}\n" -X POST "$BASE_URL/api/rest/v1/bookmarks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$BOOKMARK_PAYLOAD" | python -m json.tool 2>/dev/null
echo ""

# Test 5: Search bookmarks
echo "6. POST /api/rest/v1/bookmarks?task=searchTime - Search bookmarks"
SEARCH_PAYLOAD=$(cat <<EOF
{
  "time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "timeSpanBefore": 3600,
  "timeSpanAfter": 3600,
  "deviceIds": ["$CAMERA_ID"]
}
EOF
)

curl -k -s -w "\nHTTP_STATUS:%{http_code}\n" -X POST "$BASE_URL/api/rest/v1/bookmarks?task=searchTime" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$SEARCH_PAYLOAD" | python -m json.tool 2>/dev/null
echo ""

# Test 6: Alarms API
echo "=== ALARMS API ==="
echo ""
echo "7. GET /api/rest/v1/alarms - List alarms"
curl -k -s -w "\nHTTP_STATUS:%{http_code}\n" "$BASE_URL/api/rest/v1/alarms" \
  -H "Authorization: Bearer $TOKEN" | python -m json.tool 2>/dev/null
echo ""

echo "=== Testing Complete ==="
