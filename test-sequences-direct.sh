#!/bin/bash

CAMERA_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"
USERNAME="raam"
PASSWORD="Ilove#123"

echo "=== Direct SOAP Test: Start Recording, Query Sequences ==="
echo ""

# Get SOAP token
echo "Step 1: Getting SOAP token..."
LOGIN_RESPONSE=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST \
  "https://192.168.1.11/ManagementServer/ServerCommandService.svc" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSServerCommand/IServerCommandService/Login" \
  -d '<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <Login xmlns="http://videoos.net/2/XProtectCSServerCommand">
      <instanceId>00000000-0000-0000-0000-000000000000</instanceId>
    </Login>
  </soap:Body>
</soap:Envelope>')

SOAP_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '<Token>[^<]*</Token>' | sed 's/<Token>//;s/<\/Token>//')
echo "Token: ${SOAP_TOKEN:0:50}..."
echo ""

# Start manual recording
echo "Step 2: Starting 1-minute manual recording..."
START_RESPONSE=$(curl -k -s -X POST \
  "https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/StartManualRecording" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <StartManualRecording xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$SOAP_TOKEN</token>
      <deviceIds xmlns:a=\"http://microsoft.com/wsdl/types/\">
        <a:guid>$CAMERA_ID</a:guid>
      </deviceIds>
      <recordingTimeInMicroseconds>60000000</recordingTimeInMicroseconds>
    </StartManualRecording>
  </soap:Body>
</soap:Envelope>")

echo "Start Response:"
echo "$START_RESPONSE" | xmllint --format - 2>/dev/null || echo "$START_RESPONSE"
echo ""

# Wait 30 seconds
echo "Step 3: Waiting 40 seconds..."
sleep 40

# Stop recording
echo "Step 4: Stopping recording..."
STOP_RESPONSE=$(curl -k -s -X POST \
  "https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/StopManualRecording" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <StopManualRecording xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$SOAP_TOKEN</token>
      <deviceIds xmlns:a=\"http://microsoft.com/wsdl/types/\">
        <a:guid>$CAMERA_ID</a:guid>
      </deviceIds>
    </StopManualRecording>
  </soap:Body>
</soap:Envelope>")

echo "Stop Response:"
echo "$STOP_RESPONSE" | xmllint --format - 2>/dev/null || echo "$STOP_RESPONSE"
echo ""

# Wait a bit for recording to be saved
sleep 5

# Query sequences
echo "Step 5: Querying sequences..."
END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
START_TIME=$(date -u -d '5 minutes ago' +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-5M +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null)

SEQUENCES_RESPONSE=$(curl -k -s -X POST \
  "https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGet" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <SequencesGet xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$SOAP_TOKEN</token>
      <deviceId>$CAMERA_ID</deviceId>
      <sequenceType>78289503-CBF4-43be-9DC3-9F34A8B60E6D</sequenceType>
      <minTime>$START_TIME</minTime>
      <maxTime>$END_TIME</maxTime>
      <maxCount>100</maxCount>
    </SequencesGet>
  </soap:Body>
</soap:Envelope>")

echo "Sequences Response:"
echo "$SEQUENCES_RESPONSE" | xmllint --format - 2>/dev/null || echo "$SEQUENCES_RESPONSE"
echo ""

echo "=== Test Complete ==="
