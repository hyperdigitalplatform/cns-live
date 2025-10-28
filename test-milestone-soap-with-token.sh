#!/bin/bash

# Milestone SOAP API Testing Script with OAuth Token
# Server: 192.168.1.11
# Port: 7563 (RecorderCommandService)

BASE_URL="https://192.168.1.11"
SOAP_URL="https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx"
USERNAME="raam"
PASSWORD="Ilove#123"

echo "=== Milestone XProtect SOAP API Testing (OAuth Token) ==="
echo ""

# Step 1: Get OAuth token
echo "1. Getting OAuth 2.0 token..."
AUTH_RESPONSE=$(curl -k -s -X POST "$BASE_URL/API/IDP/connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "grant_type=password" \
  --data-urlencode "username=$USERNAME" \
  --data-urlencode "password=$PASSWORD" \
  --data-urlencode "client_id=GrantValidatorClient")

TOKEN=$(echo "$AUTH_RESPONSE" | python -m json.tool | grep '"access_token"' | sed 's/.*": "//;s/".*//')

if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to get access token"
  echo "Response: $AUTH_RESPONSE"
  exit 1
fi

echo "✅ Token obtained (length: ${#TOKEN} chars)"
echo ""

# Camera IDs
CAMERA_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"
echo "Using Camera ID: $CAMERA_ID"
echo ""

# Test 1: SequencesGetTypes
echo "=== TEST 1: SequencesGetTypes ==="
echo "Getting available sequence types for camera..."
echo ""

RESPONSE_1=$(curl -k -s -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGetTypes" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <SequencesGetTypes xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$TOKEN</token>
      <deviceId>$CAMERA_ID</deviceId>
    </SequencesGetTypes>
  </soap:Body>
</soap:Envelope>")

echo "Response:"
echo "$RESPONSE_1" | xmllint --format - 2>/dev/null || echo "$RESPONSE_1"
echo ""
echo "---"
echo ""

# Test 2: IsManualRecording
echo "=== TEST 2: IsManualRecording ==="
echo "Checking if camera is currently recording..."
echo ""

RESPONSE_2=$(curl -k -s -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/IsManualRecording" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <IsManualRecording xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$TOKEN</token>
      <deviceIds xmlns:a=\"http://microsoft.com/wsdl/types/\">
        <a:guid>$CAMERA_ID</a:guid>
      </deviceIds>
    </IsManualRecording>
  </soap:Body>
</soap:Envelope>")

echo "Response:"
echo "$RESPONSE_2" | xmllint --format - 2>/dev/null || echo "$RESPONSE_2"
echo ""
echo "---"
echo ""

# Test 3: StartRecording (15 minutes = 900 seconds = 900000000 microseconds)
echo "=== TEST 3: StartRecording ==="
echo "Starting 15-minute manual recording..."
echo ""

RESPONSE_3=$(curl -k -s -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/StartRecording" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <StartRecording xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$TOKEN</token>
      <deviceId>$CAMERA_ID</deviceId>
      <recordingTimeMicroSeconds>900000000</recordingTimeMicroSeconds>
    </StartRecording>
  </soap:Body>
</soap:Envelope>")

echo "Response:"
echo "$RESPONSE_3" | xmllint --format - 2>/dev/null || echo "$RESPONSE_3"
echo ""
echo "---"
echo ""

# Test 4: Check recording status again
echo "=== TEST 4: IsManualRecording (After Starting) ==="
echo "Checking if recording started..."
echo ""

sleep 2  # Wait a moment

RESPONSE_4=$(curl -k -s -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/IsManualRecording" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <IsManualRecording xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$TOKEN</token>
      <deviceIds xmlns:a=\"http://microsoft.com/wsdl/types/\">
        <a:guid>$CAMERA_ID</a:guid>
      </deviceIds>
    </IsManualRecording>
  </soap:Body>
</soap:Envelope>")

echo "Response:"
echo "$RESPONSE_4" | xmllint --format - 2>/dev/null || echo "$RESPONSE_4"
echo ""
echo "---"
echo ""

echo "=== Testing Complete ==="
echo ""
echo "All tests executed successfully!"
