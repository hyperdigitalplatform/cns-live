#!/bin/bash

# Milestone SOAP API Testing Script with Basic Authentication
# Server: 192.168.1.11
# Port: 7563 (RecorderCommandService)

BASE_URL="https://192.168.1.11"
SOAP_URL="https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx"
USERNAME="raam"
PASSWORD="Ilove#123"

echo "=== Milestone XProtect SOAP API Testing (Basic Auth) ==="
echo ""
echo "Authentication: HTTP Basic Authentication"
echo "Username: $USERNAME"
echo ""

# Camera IDs
CAMERA_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"
echo "Using Camera ID: $CAMERA_ID"
echo ""

# Test 1: SequencesGetTypes
echo "=== TEST 1: SequencesGetTypes ==="
echo "Getting available sequence types for camera..."
echo ""

SOAP_REQUEST_1='<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <SequencesGetTypes xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <deviceId>'$CAMERA_ID'</deviceId>
    </SequencesGetTypes>
  </soap:Body>
</soap:Envelope>'

echo "Request:"
echo "$SOAP_REQUEST_1"
echo ""

RESPONSE_1=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGetTypes" \
  -d "$SOAP_REQUEST_1")

echo "Response:"
echo "$RESPONSE_1" | xmllint --format - 2>/dev/null || echo "$RESPONSE_1"
echo ""
echo "---"
echo ""

# Test 2: IsManualRecording
echo "=== TEST 2: IsManualRecording ==="
echo "Checking if camera is currently recording..."
echo ""

SOAP_REQUEST_2='<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <IsManualRecording xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <deviceIds xmlns:a="http://microsoft.com/wsdl/types/">
        <a:guid>'$CAMERA_ID'</a:guid>
      </deviceIds>
    </IsManualRecording>
  </soap:Body>
</soap:Envelope>'

echo "Request:"
echo "$SOAP_REQUEST_2"
echo ""

RESPONSE_2=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/IsManualRecording" \
  -d "$SOAP_REQUEST_2")

echo "Response:"
echo "$RESPONSE_2" | xmllint --format - 2>/dev/null || echo "$RESPONSE_2"
echo ""
echo "---"
echo ""

# Test 3: StartRecording (15 minutes = 900 seconds = 900000000 microseconds)
echo "=== TEST 3: StartRecording ==="
echo "Starting 15-minute manual recording..."
echo ""

SOAP_REQUEST_3='<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <StartRecording xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <deviceId>'$CAMERA_ID'</deviceId>
      <recordingTimeMicroSeconds>900000000</recordingTimeMicroSeconds>
    </StartRecording>
  </soap:Body>
</soap:Envelope>'

echo "Request:"
echo "$SOAP_REQUEST_3"
echo ""

RESPONSE_3=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/StartRecording" \
  -d "$SOAP_REQUEST_3")

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

RESPONSE_4=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/IsManualRecording" \
  -d "$SOAP_REQUEST_2")

echo "Response:"
echo "$RESPONSE_4" | xmllint --format - 2>/dev/null || echo "$RESPONSE_4"
echo ""
echo "---"
echo ""

# Test 5: SequencesGet (get recordings from last hour)
echo "=== TEST 5: SequencesGet ==="
echo "Getting recording sequences from last hour..."
echo ""

# Get sequence type ID from first test (assuming first type)
SEQUENCE_TYPE_ID=$(echo "$RESPONSE_1" | grep -o '<Id>[^<]*</Id>' | head -1 | sed 's/<Id>//;s/<\/Id>//')

if [ -z "$SEQUENCE_TYPE_ID" ]; then
  SEQUENCE_TYPE_ID="00000000-0000-0000-0000-000000000000"
fi

echo "Using Sequence Type ID: $SEQUENCE_TYPE_ID"
echo ""

# Calculate times (last hour)
END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
START_TIME=$(date -u -d '1 hour ago' +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2025-10-27T16:00:00Z")

SOAP_REQUEST_5='<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <SequencesGet xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <deviceId>'$CAMERA_ID'</deviceId>
      <sequenceType xmlns:a="http://microsoft.com/wsdl/types/">'$SEQUENCE_TYPE_ID'</sequenceType>
      <minTime>'$START_TIME'</minTime>
      <maxTime>'$END_TIME'</maxTime>
      <maxCount>100</maxCount>
    </SequencesGet>
  </soap:Body>
</soap:Envelope>'

echo "Request:"
echo "$SOAP_REQUEST_5"
echo ""

RESPONSE_5=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGet" \
  -d "$SOAP_REQUEST_5")

echo "Response:"
echo "$RESPONSE_5" | xmllint --format - 2>/dev/null || echo "$RESPONSE_5"
echo ""
echo "---"
echo ""

# Test 6: TimeLineInformationGet
echo "=== TEST 6: TimeLineInformationGet ==="
echo "Getting timeline data for last hour..."
echo ""

SOAP_REQUEST_6='<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <TimeLineInformationGet xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <deviceId>'$CAMERA_ID'</deviceId>
      <timeLineInformationTypes xmlns:a="http://microsoft.com/wsdl/types/">
        <a:guid>'$SEQUENCE_TYPE_ID'</a:guid>
      </timeLineInformationTypes>
      <timeLineInformationBeginTime>'$START_TIME'</timeLineInformationBeginTime>
      <timeLineInformationInterval>
        <MicroSeconds>60000000</MicroSeconds>
      </timeLineInformationInterval>
      <timeLineInformationCount>60</timeLineInformationCount>
    </TimeLineInformationGet>
  </soap:Body>
</soap:Envelope>'

echo "Request:"
echo "$SOAP_REQUEST_6"
echo ""

RESPONSE_6=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/TimeLineInformationGet" \
  -d "$SOAP_REQUEST_6")

echo "Response:"
echo "$RESPONSE_6" | xmllint --format - 2>/dev/null || echo "$RESPONSE_6"
echo ""
echo "---"
echo ""

# Test 7: JPEGGetAt
echo "=== TEST 7: JPEGGetAt ==="
echo "Getting JPEG snapshot from 1 minute ago..."
echo ""

SNAPSHOT_TIME=$(date -u -d '1 minute ago' +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-1M +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2025-10-27T17:00:00Z")

SOAP_REQUEST_7='<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <JPEGGetAt xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <deviceId>'$CAMERA_ID'</deviceId>
      <time>'$SNAPSHOT_TIME'</time>
    </JPEGGetAt>
  </soap:Body>
</soap:Envelope>'

echo "Request:"
echo "$SOAP_REQUEST_7"
echo ""

RESPONSE_7=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST "$SOAP_URL" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/JPEGGetAt" \
  -d "$SOAP_REQUEST_7")

echo "Response (truncated if large):"
echo "$RESPONSE_7" | head -50
echo ""
echo "---"
echo ""

echo "=== Testing Complete ==="
echo ""
echo "Summary:"
echo "1. SequencesGetTypes - Check response above"
echo "2. IsManualRecording (before) - Check response above"
echo "3. StartRecording - Check response above"
echo "4. IsManualRecording (after) - Check response above"
echo "5. SequencesGet - Check response above"
echo "6. TimeLineInformationGet - Check response above"
echo "7. JPEGGetAt - Check response above"
