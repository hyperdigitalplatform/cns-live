#!/bin/bash
# Test Lua scripts directly with Valkey CLI

set -e

VALKEY_HOST=${VALKEY_HOST:-localhost}
VALKEY_PORT=${VALKEY_PORT:-6379}
VALKEY_CLI="redis-cli -h $VALKEY_HOST -p $VALKEY_PORT"

echo "=== Testing Valkey Stream Counter Lua Scripts ==="
echo "Valkey: $VALKEY_HOST:$VALKEY_PORT"
echo ""

# Initialize limits
echo "1. Initializing limits..."
$VALKEY_CLI SET stream:limit:DUBAI_POLICE 50
$VALKEY_CLI SET stream:limit:METRO 30
$VALKEY_CLI SET stream:limit:BUS 20
$VALKEY_CLI SET stream:limit:OTHER 400
$VALKEY_CLI SET stream:count:DUBAI_POLICE 0
$VALKEY_CLI SET stream:count:METRO 0
$VALKEY_CLI SET stream:count:BUS 0
$VALKEY_CLI SET stream:count:OTHER 0
echo "✓ Limits initialized"
echo ""

# Test reserve_stream.lua
echo "2. Testing reserve_stream.lua..."
RESULT=$($VALKEY_CLI --eval scripts/lua/reserve_stream.lua 0 \
    "DUBAI_POLICE" \
    "test-reservation-1" \
    "camera-123" \
    "user-123" \
    "3600")
echo "Result: $RESULT"

if [[ $RESULT == *"1"* ]]; then
    echo "✓ Reserve successful"
else
    echo "✗ Reserve failed"
    exit 1
fi
echo ""

# Check count increased
COUNT=$($VALKEY_CLI GET stream:count:DUBAI_POLICE)
echo "3. Current count: $COUNT"
if [ "$COUNT" == "1" ]; then
    echo "✓ Count incremented correctly"
else
    echo "✗ Count incorrect (expected 1, got $COUNT)"
    exit 1
fi
echo ""

# Test heartbeat_stream.lua
echo "4. Testing heartbeat_stream.lua..."
RESULT=$($VALKEY_CLI --eval scripts/lua/heartbeat_stream.lua 0 \
    "test-reservation-1" \
    "60")
echo "Result: $RESULT"
if [[ $RESULT == *"1"* ]]; then
    echo "✓ Heartbeat successful"
else
    echo "✗ Heartbeat failed"
    exit 1
fi
echo ""

# Test get_stats.lua
echo "5. Testing get_stats.lua..."
RESULT=$($VALKEY_CLI --eval scripts/lua/get_stats.lua 0 \
    "DUBAI_POLICE,METRO,BUS,OTHER")
echo "Stats: $RESULT"
echo "✓ Stats retrieved"
echo ""

# Test release_stream.lua
echo "6. Testing release_stream.lua..."
RESULT=$($VALKEY_CLI --eval scripts/lua/release_stream.lua 0 \
    "test-reservation-1")
echo "Result: $RESULT"
if [[ $RESULT == *"1"* ]]; then
    echo "✓ Release successful"
else
    echo "✗ Release failed"
    exit 1
fi
echo ""

# Check count decreased
COUNT=$($VALKEY_CLI GET stream:count:DUBAI_POLICE)
echo "7. Current count after release: $COUNT"
if [ "$COUNT" == "0" ]; then
    echo "✓ Count decremented correctly"
else
    echo "✗ Count incorrect (expected 0, got $COUNT)"
    exit 1
fi
echo ""

# Test limit enforcement
echo "8. Testing limit enforcement..."
# Set limit to 2 for testing
$VALKEY_CLI SET stream:limit:DUBAI_POLICE 2

# Reserve 2 streams
$VALKEY_CLI --eval scripts/lua/reserve_stream.lua 0 \
    "DUBAI_POLICE" "test-res-1" "cam-1" "user-1" "3600" > /dev/null
$VALKEY_CLI --eval scripts/lua/reserve_stream.lua 0 \
    "DUBAI_POLICE" "test-res-2" "cam-2" "user-2" "3600" > /dev/null

# Try to reserve 3rd stream (should fail)
RESULT=$($VALKEY_CLI --eval scripts/lua/reserve_stream.lua 0 \
    "DUBAI_POLICE" "test-res-3" "cam-3" "user-3" "3600")

if [[ $RESULT == *"0"* ]]; then
    echo "✓ Limit enforcement working (correctly rejected 3rd stream)"
else
    echo "✗ Limit enforcement failed (should have rejected 3rd stream)"
    exit 1
fi
echo ""

# Cleanup
echo "9. Cleaning up..."
$VALKEY_CLI DEL stream:reservation:test-res-1
$VALKEY_CLI DEL stream:reservation:test-res-2
$VALKEY_CLI DEL stream:heartbeat:test-res-1
$VALKEY_CLI DEL stream:heartbeat:test-res-2
$VALKEY_CLI SET stream:count:DUBAI_POLICE 0
$VALKEY_CLI SET stream:limit:DUBAI_POLICE 50
echo "✓ Cleanup complete"
echo ""

echo "=== All Lua script tests passed! ==="
