#!/bin/bash

# Milestone API Testing Script
# Server: 192.168.1.11
# User: raam

BASE_URL="https://192.168.1.11"

echo "=== Milestone XProtect API Testing ==="
echo ""

# Step 1: Authenticate and get token
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

# Step 2: Test Sites API
echo "2. Testing Sites API"
echo "GET $BASE_URL/api/rest/v1/sites"
echo ""

SITES_RESPONSE=$(curl -k -s "$BASE_URL/api/rest/v1/sites" \
  -H "Authorization: Bearer $TOKEN")

echo "Response:"
echo "$SITES_RESPONSE" | python -m json.tool 2>/dev/null || echo "$SITES_RESPONSE"
echo ""

# Step 3: Test Cameras API (try different endpoints)
echo "3. Testing Cameras API"
echo "GET $BASE_URL/api/rest/v1/cameras"
echo ""

CAMERAS_RESPONSE=$(curl -k -s "$BASE_URL/api/rest/v1/cameras" \
  -H "Authorization: Bearer $TOKEN")

echo "Response:"
echo "$CAMERAS_RESPONSE" | python -m json.tool 2>/dev/null || echo "$CAMERAS_RESPONSE"
echo ""

# Step 4: Try hardware endpoints
echo "4. Testing Hardware API"
echo "GET $BASE_URL/api/rest/v1/hardware"
echo ""

HARDWARE_RESPONSE=$(curl -k -s "$BASE_URL/api/rest/v1/hardware" \
  -H "Authorization: Bearer $TOKEN")

echo "Response:"
echo "$HARDWARE_RESPONSE" | python -m json.tool 2>/dev/null || echo "$HARDWARE_RESPONSE"
echo ""

# Step 5: Try devices endpoints
echo "5. Testing Devices API"
echo "GET $BASE_URL/api/rest/v1/devices"
echo ""

DEVICES_RESPONSE=$(curl -k -s "$BASE_URL/api/rest/v1/devices" \
  -H "Authorization: Bearer $TOKEN")

echo "Response:"
echo "$DEVICES_RESPONSE" | python -m json.tool 2>/dev/null || echo "$DEVICES_RESPONSE"
echo ""

echo "=== Testing Complete ==="
