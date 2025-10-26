# Multi-Viewer Streaming Troubleshooting

**Quick Reference Guide**

## Common Issues and Solutions

### Issue 1: Browser Stuck on "Connecting to Camera..."

**Symptoms:**
- Circular progress indicator indefinitely
- No video appears
- Other browsers may work fine

**Possible Causes & Solutions:**

#### A. WHIP Container Not Running
```bash
# Check if container exists
docker ps --filter "name=whip-pusher-cam-001"

# If not running, check why it was stopped
docker logs cctv-go-api --tail 100 | grep "whip-pusher"

# Manual restart (temporary fix)
# Better to close and reopen camera in dashboard
```

#### B. Token Expired
```bash
# Check token expiry in browser console (F12)
# Tokens valid for 1 hour by default

# Solution: Close and reopen camera
```

#### C. LiveKit Connection Issues
```bash
# Check LiveKit logs
docker logs cctv-livekit --tail 50

# Check if room exists
curl http://localhost:7880/rooms | jq '.rooms[] | select(.name == "camera_cam-001-sheikh-zayed")'

# Solution: Restart go-api service
docker-compose restart go-api
```

---

### Issue 2: Video Stops When Second Viewer Joins

**Symptoms:**
- Browser 1 shows video perfectly
- Browser 2 opens same camera
- Browser 1's video stops and shows "Connecting..."

**Root Cause:** DUPLICATE_IDENTITY error (should be fixed in v1.0+)

**Verification:**
```bash
# Check LiveKit logs for DUPLICATE_IDENTITY
docker logs cctv-livekit 2>&1 | grep "DUPLICATE_IDENTITY"

# Check if fix is applied
grep "viewer_%s" /d/armed/github/cns/services/go-api/internal/usecase/stream_usecase.go

# Should show:
# participantIdentity := fmt.Sprintf("viewer_%s", reservation.ReservationID)
```

**Solution:**
If you see DUPLICATE_IDENTITY errors, the fix is not applied:
```bash
# Rebuild and deploy go-api
cd /d/armed/github/cns
docker-compose build go-api
docker-compose up -d --force-recreate go-api
```

---

### Issue 3: Viewer Count Shows Wrong Number

**Expected Behavior:**
- Viewer count should exclude the publisher (WHIP container)
- Only count actual dashboard viewers

**Debug:**
```bash
# Check LiveKit participants
curl http://localhost:7880/rooms/camera_cam-001-sheikh-zayed | jq '.rooms[].num_participants'

# Check stream stats
curl http://localhost:8086/api/v1/stream/stats | jq '.camera_stats'
```

**If viewer count = num_participants:**
The fix to subtract publisher is not applied.

**Solution:**
Check line ~303-322 in `services/go-api/internal/usecase/stream_usecase.go`:
```go
viewerCount := int(room.NumParticipants)
if viewerCount > 0 {
    viewerCount-- // Subtract 1 for the publisher
}
```

---

### Issue 4: Container Not Cleaned Up After Last Viewer Leaves

**Symptoms:**
- All browsers closed
- WHIP container still running
- Reservation still in Valkey

**Check Reference Counting:**
```bash
# Check active reservations for camera
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs -I {} sh -c 'echo "{}"; docker exec cctv-valkey valkey-cli HGET {} camera_id'

# Check running containers
docker ps --filter "name=whip-pusher"
```

**If container should be stopped but isn't:**
```bash
# Check go-api logs for cleanup logic
docker logs cctv-go-api 2>&1 | grep "Last viewer disconnected"

# Manual cleanup
docker stop whip-pusher-cam-001-sheikh-zayed
docker exec cctv-valkey valkey-cli DEL "stream:reservation:<id>"
```

---

### Issue 5: Stream Counter Not Decreasing

**Symptoms:**
- Viewers close browsers
- Counter stays at higher number
- Prevents new streams due to quota

**Diagnosis:**
```bash
# Check counter vs actual reservations
echo "Counter:"
docker exec cctv-valkey valkey-cli GET stream:count:DUBAI_POLICE

echo "Actual reservations:"
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | wc -l

# If mismatch, check release_stream.lua logs
docker logs cctv-stream-counter --tail 100
```

**Solution:**
```bash
# Reset counter to match reality
ACTUAL=$(docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs -I {} sh -c 'docker exec cctv-valkey valkey-cli HGET {} source' | \
  grep "DUBAI_POLICE" | wc -l)

docker exec cctv-valkey valkey-cli SET stream:count:DUBAI_POLICE $ACTUAL
```

---

### Issue 6: Multiple WHIP Containers for Same Camera

**Symptoms:**
- `docker ps` shows multiple whip-pusher containers with similar names
- Resource waste
- Possible conflicts

**This Should Never Happen** (indicates backend resource sharing bug)

**Diagnosis:**
```bash
# Check for duplicate containers
docker ps --filter "name=whip-pusher" --format "{{.Names}}"

# Check reservations
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs -I {} sh -c 'echo "=== {} ==="; docker exec cctv-valkey valkey-cli HGETALL {}'
```

**Emergency Fix:**
```bash
# Stop all WHIP containers
docker ps --filter "name=whip-pusher" -q | xargs docker stop

# Clear all reservations
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs docker exec cctv-valkey valkey-cli DEL

# Reset counters
docker exec cctv-valkey valkey-cli SET stream:count:DUBAI_POLICE 0
docker exec cctv-valkey valkey-cli SET stream:count:METRO 0

# Reopen cameras from dashboard
```

---

## Monitoring Commands

### Quick Health Check
```bash
# One-liner status
echo "Containers: $(docker ps --filter 'name=whip-pusher' -q | wc -l) | \
Reservations: $(docker exec cctv-valkey valkey-cli --scan --pattern 'stream:reservation:*' | wc -l) | \
DUBAI_POLICE: $(docker exec cctv-valkey valkey-cli GET stream:count:DUBAI_POLICE)"
```

### Detailed Status
```bash
# Stream stats API
curl -s http://localhost:8086/api/v1/stream/stats | jq

# Active reservations with details
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs -I {} sh -c 'echo "=== {} ==="; docker exec cctv-valkey valkey-cli HGETALL {}'

# LiveKit rooms
curl -s http://localhost:7880/rooms | jq '.rooms[] | {name, num_participants}'
```

### Watch Mode (Auto-refresh)
```bash
# Stream stats
watch -n 2 'curl -s http://localhost:8086/api/v1/stream/stats | jq'

# Container status
watch -n 2 'docker ps --filter "name=whip-pusher" --format "table {{.Names}}\t{{.Status}}"'
```

---

## Log Analysis

### Finding Errors
```bash
# Go-API errors
docker logs cctv-go-api 2>&1 | grep -i "error" | tail -20

# LiveKit issues
docker logs cctv-livekit 2>&1 | grep -E "(error|ERROR|DUPLICATE)" | tail -20

# Stream counter issues
docker logs cctv-stream-counter 2>&1 | grep -i "error" | tail -20
```

### Tracing a Specific Request
```bash
# Find reservation by camera ID
RESERVATION_ID=$(docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs -I {} sh -c 'ID={}; CAM=$(docker exec cctv-valkey valkey-cli HGET {} camera_id); \
  if [ "$CAM" = "cam-001-sheikh-zayed" ]; then echo ${ID#stream:reservation:}; fi')

echo "Reservation ID: $RESERVATION_ID"

# Follow the lifecycle
docker logs cctv-go-api 2>&1 | grep "$RESERVATION_ID"
```

---

## Performance Issues

### High CPU Usage
```bash
# Check WHIP containers
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep whip-pusher

# If any container >50% CPU, check camera RTSP
# High motion scenes = higher encoding CPU
```

### Memory Leaks
```bash
# Monitor over time
watch -n 5 'docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}" | grep -E "(cctv|whip)"'

# If memory grows continuously, restart affected service
docker-compose restart go-api
```

### Network Bandwidth
```bash
# Check LiveKit bandwidth
docker exec cctv-livekit wget -qO- http://localhost:7880/debug/pprof/allocs

# Monitor network usage
docker stats --no-stream --format "table {{.Name}}\t{{.NetIO}}" | grep -E "(livekit|whip)"
```

---

## Emergency Procedures

### Complete System Reset
```bash
# WARNING: Disconnects all viewers!

# 1. Stop all WHIP containers
docker ps --filter "name=whip-pusher" -q | xargs docker stop 2>/dev/null

# 2. Clear Valkey reservations
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs docker exec cctv-valkey valkey-cli DEL

# 3. Clear Valkey metadata
docker exec cctv-valkey valkey-cli --scan --pattern "stream:metadata:*" | \
  xargs docker exec cctv-valkey valkey-cli DEL

# 4. Reset counters
for SOURCE in DUBAI_POLICE METRO PARKING TAXI; do
  docker exec cctv-valkey valkey-cli SET "stream:count:$SOURCE" 0
done

# 5. Restart services
docker-compose restart go-api stream-counter livekit

# 6. Verify clean state
echo "Reservations: $(docker exec cctv-valkey valkey-cli --scan --pattern 'stream:reservation:*' | wc -l)"
echo "Containers: $(docker ps --filter 'name=whip-pusher' -q | wc -l)"
echo "Counters: $(docker exec cctv-valkey valkey-cli MGET stream:count:DUBAI_POLICE stream:count:METRO)"
```

### Specific Camera Reset
```bash
CAMERA_ID="cam-001-sheikh-zayed"

# 1. Stop WHIP container
docker stop "whip-pusher-$CAMERA_ID" 2>/dev/null

# 2. Find and delete reservations
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs -I {} sh -c 'CAM=$(docker exec cctv-valkey valkey-cli HGET {} camera_id); \
  if [ "$CAM" = "'$CAMERA_ID'" ]; then docker exec cctv-valkey valkey-cli DEL {}; fi'

# 3. Delete LiveKit ingress (if needed)
# Find ingress ID from go-api logs, then:
# curl -X DELETE http://localhost:7880/ingress/<ingress_id>

echo "Camera $CAMERA_ID reset complete"
```

---

## Validation After Fix

After applying any fix, validate with this checklist:

### ✅ Single Viewer Test
```bash
# 1. Open camera in Browser 1
# 2. Verify video plays
# 3. Check stats:
curl -s http://localhost:8086/api/v1/stream/stats | jq '.camera_stats[] | select(.camera_id == "cam-001-sheikh-zayed")'
# Expected: viewer_count = 1
```

### ✅ Multi-Viewer Test
```bash
# 1. Open same camera in Browser 2 (incognito)
# 2. Verify both browsers show video
# 3. Check stats:
curl -s http://localhost:8086/api/v1/stream/stats | jq '.camera_stats[] | select(.camera_id == "cam-001-sheikh-zayed")'
# Expected: viewer_count = 2

# 4. Verify single WHIP container:
docker ps --filter "name=whip-pusher-cam-001" --format "{{.Names}}"
# Expected: One container only
```

### ✅ Cleanup Test
```bash
# 1. Close Browser 2
# 2. Verify Browser 1 still works
# 3. Close Browser 1
# 4. Wait 5 seconds
# 5. Check cleanup:
docker ps --filter "name=whip-pusher-cam-001"
# Expected: No containers

docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*"
# Expected: Empty or unrelated cameras only
```

---

## Getting Help

### Collect Debug Info
```bash
# Create debug bundle
mkdir -p /tmp/cctv-debug
cd /tmp/cctv-debug

# Collect logs
docker-compose logs --tail 200 go-api > go-api.log
docker-compose logs --tail 200 livekit > livekit.log
docker-compose logs --tail 200 stream-counter > stream-counter.log

# Collect state
docker ps > containers.txt
curl -s http://localhost:8086/api/v1/stream/stats | jq > stream-stats.json
curl -s http://localhost:7880/rooms | jq > livekit-rooms.json
docker exec cctv-valkey valkey-cli --scan --pattern "stream:reservation:*" | \
  xargs -I {} sh -c 'echo "=== {} ==="; docker exec cctv-valkey valkey-cli HGETALL {}' > valkey-reservations.txt

# Create archive
tar -czf ../cctv-debug-$(date +%Y%m%d-%H%M%S).tar.gz .

echo "Debug bundle created: /tmp/cctv-debug-*.tar.gz"
```

### Contact Support
- Include debug bundle
- Describe symptoms and steps to reproduce
- Note any recent changes to system

---

**Last Updated**: 2025-10-26
**Version**: 1.0
**Related**: [MULTI_VIEWER_STREAMING.md](MULTI_VIEWER_STREAMING.md)
