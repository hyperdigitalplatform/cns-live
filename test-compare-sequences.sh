#!/bin/bash

# Compare Milestone sequences between working and non-working cameras
# Working: TP-Link Tapo (192.168.1.8) - d47fa4e9-8171-4cc2-a421-95a3194f6a1d
# Not Working: GUANGZHOU (192.168.1.13) - a8a8b9dc-3995-49ed-9b00-62caac2ce74a

USERNAME="raam"
PASSWORD="Ilove#123"
BASE_URL="https://192.168.1.9"

TAPO_ID="d47fa4e9-8171-4cc2-a421-95a3194f6a1d"
GUANGZHOU_ID="a8a8b9dc-3995-49ed-9b00-62caac2ce74a"

echo "=== Comparing Milestone Sequences ==="
echo ""

# Get SOAP token
echo "Step 1: Getting SOAP token..."
LOGIN_RESPONSE=$(curl -k -s -u "$USERNAME:$PASSWORD" -X POST \
  "$BASE_URL/ManagementServer/ServerCommandService.svc" \
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

# Set time range (last 7 days)
END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
START_TIME=$(date -u -d '7 days ago' +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-7d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null)

echo "Time Range: $START_TIME to $END_TIME"
echo ""

# Query TP-Link Tapo sequences
echo "=== TP-Link Tapo Camera ($TAPO_ID) ==="
echo "Step 2: Querying sequences for TP-Link Tapo..."
TAPO_RESPONSE=$(curl -k -s -X POST \
  "$BASE_URL:7563/RecorderCommandService/RecorderCommandService.asmx" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGet" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <SequencesGet xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$SOAP_TOKEN</token>
      <deviceId>$TAPO_ID</deviceId>
      <sequenceType>78289503-CBF4-43be-9DC3-9F34A8B60E6D</sequenceType>
      <minTime>$START_TIME</minTime>
      <maxTime>$END_TIME</maxTime>
      <maxCount>10</maxCount>
    </SequencesGet>
  </soap:Body>
</soap:Envelope>")

echo "Response:"
echo "$TAPO_RESPONSE" | xmllint --format - 2>/dev/null || echo "$TAPO_RESPONSE"
echo ""
echo "Sequence Count:"
echo "$TAPO_RESPONSE" | grep -o '<Sequence>' | wc -l
echo ""

# Query GUANGZHOU sequences
echo "=== GUANGZHOU Camera ($GUANGZHOU_ID) ==="
echo "Step 3: Querying sequences for GUANGZHOU..."
GUANGZHOU_RESPONSE=$(curl -k -s -X POST \
  "$BASE_URL:7563/RecorderCommandService/RecorderCommandService.asmx" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGet" \
  -d "<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <SequencesGet xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$SOAP_TOKEN</token>
      <deviceId>$GUANGZHOU_ID</deviceId>
      <sequenceType>78289503-CBF4-43be-9DC3-9F34A8B60E6D</sequenceType>
      <minTime>$START_TIME</minTime>
      <maxTime>$END_TIME</maxTime>
      <maxCount>10</maxCount>
    </SequencesGet>
  </soap:Body>
</soap:Envelope>")

echo "Response:"
echo "$GUANGZHOU_RESPONSE" | xmllint --format - 2>/dev/null || echo "$GUANGZHOU_RESPONSE"
echo ""
echo "Sequence Count:"
echo "$GUANGZHOU_RESPONSE" | grep -o '<Sequence>' | wc -l
echo ""

echo "=== Comparison Complete ==="
