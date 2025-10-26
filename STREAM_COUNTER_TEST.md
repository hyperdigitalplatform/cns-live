# Stream Counter Verification Guide

## Quick Commands

### 1. Check Current Stream Stats (API)
```bash
curl -s http://localhost:8000/api/v1/stream/stats | python -m json.tool
```

### 2. Check Stream Counts in Valkey (Direct)
```bash
# All sources
docker exec cctv-valkey valkey-cli MGET stream:count:DUBAI_POLICE stream:count:METRO stream:count:PARKING stream:count:TAXI

# Individual source
docker exec cctv-valkey valkey-cli GET "stream:count:DUBAI_POLICE"
```

### 3. Use Monitoring Script
```bash
# Windows
.\check-streams.bat

# Linux/Mac
./monitor-streams.sh
```

### 4. Watch in Real-Time
```bash
watch -n 2 'curl -s http://localhost:8000/api/v1/stream/stats | python -m json.tool | grep -E "active_streams|total_viewers|camera_id"'
```

## Manual Test Scenarios

### Test 1: Add Camera to Grid
**Steps:**
1. Check initial count:
   ```bash
   curl -s http://localhost:8000/api/v1/stream/stats | python -m json.tool | grep active_streams
   ```
2. Open dashboard: http://localhost:3000
3. Double-click a camera to add to grid
4. Check count again - **should INCREASE by 1**

**Expected:**
- `active_streams` increases by 1
- `stream:count:<SOURCE>` increases by 1
- Camera appears in `camera_stats` array

### Test 2: Remove Camera from Grid
**Steps:**
1. Note current count
2. Hover over camera in grid
3. Click the X button to remove camera
4. Check count - **should DECREASE by 1**

**Expected:**
- `active_streams` decreases by 1
- `stream:count:<SOURCE>` decreases by 1
- Camera removed from `camera_stats` array

### Test 3: Clear All Cameras
**Steps:**
1. Add multiple cameras to grid (2-4 cameras)
2. Note the count (e.g., active_streams: 4)
3. Click "Clear All" button
4. Check count - **should DROP to 0**

**Expected:**
- `active_streams` = 0
- All `stream:count:<SOURCE>` = 0
- `camera_stats` array is empty

### Test 4: Page Refresh
**Steps:**
1. Add 2-3 cameras to grid
2. Note the count (e.g., active_streams: 3)
3. Press F5 or Ctrl+R to refresh page
4. **Wait 2-3 seconds** for cleanup
5. Check count - **should DROP to 0**

**Expected:**
- `beforeunload` event triggers
- All streams released
- Count drops to 0 after page refresh completes

### Test 5: Browser Close/Tab Close
**Steps:**
1. Add 2-3 cameras to grid
2. Note the count
3. Close the browser tab or window
4. From terminal, check count - **should DROP to 0**

**Expected:**
- Streams are released on tab close
- Count decreases within 1-2 seconds

### Test 6: Change Grid Layout
**Steps:**
1. Add cameras in 2×2 layout
2. Note the count
3. Switch to 3×3 layout
4. Check count - **should STAY SAME or ADJUST**

**Expected:**
- Cameras that remain visible: streams stay active
- Cameras that are removed due to layout change: streams released
- New count = number of cameras still in grid

### Test 7: Multiple Users
**Steps:**
1. Open dashboard in Browser 1
2. Open dashboard in Browser 2 (incognito)
3. Add camera in Browser 1 → count +1
4. Add camera in Browser 2 → count +1
5. Remove camera in Browser 1 → count -1
6. Total count should reflect both users

**Expected:**
- Each browser has independent reservations
- Total count = sum of all active streams
- Removing from one browser doesn't affect the other

## Monitoring Stream Counter Logs

### View Stream Counter Service Logs
```bash
# Last 50 lines
docker logs cctv-stream-counter --tail 50

# Follow in real-time
docker logs cctv-stream-counter -f

# Filter for reserve/release
docker logs cctv-stream-counter --tail 100 | grep -E "reserve|release|limit"
```

### Watch for Issues
Look for these patterns:

**Good (Normal):**
```json
{"level":"info","message":"Stream reserved","camera_id":"cam-001","source":"DUBAI_POLICE","new_count":1}
{"level":"info","message":"Stream released","reservation_id":"abc-123","source":"DUBAI_POLICE","new_count":0}
```

**Bad (Stream Limit):**
```json
{"level":"warn","message":"Stream limit reached","source":"DUBAI_POLICE","current":50,"limit":50}
```

**Bad (Failed Release):**
```
POST /api/v1/stream/reserve - 500
```

## Debugging Commands

### Reset All Stream Counts (Emergency)
```bash
docker exec cctv-valkey valkey-cli SET "stream:count:DUBAI_POLICE" 0
docker exec cctv-valkey valkey-cli SET "stream:count:METRO" 0
docker exec cctv-valkey valkey-cli SET "stream:count:PARKING" 0
docker exec cctv-valkey valkey-cli SET "stream:count:TAXI" 0
```

### List All Active Reservations
```bash
docker exec cctv-valkey valkey-cli KEYS "stream:user:*"
```

### Check Specific Reservation
```bash
docker exec cctv-valkey valkey-cli GET "stream:user:<reservation-id>"
```

### View Stream Limit Configuration
```bash
docker exec cctv-valkey valkey-cli MGET stream:limit:DUBAI_POLICE stream:limit:METRO stream:limit:PARKING
```

## Grafana Dashboard (If Available)

If Grafana is configured, view real-time metrics:

1. Open: http://localhost:3001
2. Login: admin / admin
3. Look for "Stream Counter" dashboard
4. Metrics to watch:
   - `stream_count_total` - Total active streams
   - `stream_count_by_source` - Streams per source
   - `stream_reserve_total` - Total reserve requests
   - `stream_release_total` - Total release requests

## Expected Behavior Summary

| Action | Stream Count | Valkey Count | Notes |
|--------|-------------|--------------|-------|
| Add camera | +1 | +1 | Immediate |
| Remove camera | -1 | -1 | Immediate |
| Clear all | -N | -N | Immediate |
| Page refresh | -N | -N | Within 1-2 seconds |
| Tab close | -N | -N | Within 1-2 seconds |
| Heartbeat timeout | -1 | -1 | After ~30 seconds |
| Layout change | ±N | ±N | Depends on cameras kept |

## Troubleshooting

### Count Not Decreasing?
1. Check browser console for errors (F12)
2. Check stream-counter logs: `docker logs cctv-stream-counter --tail 50`
3. Verify release API is being called
4. Check network tab in browser DevTools

### Count Stuck at Limit?
1. Manually reset: `docker exec cctv-valkey valkey-cli SET "stream:count:<SOURCE>" 0`
2. Restart stream-counter: `docker-compose restart stream-counter`
3. Check for orphaned ingress containers: `docker ps | grep ingress`

### Counts Mismatch?
If API stats show different count than Valkey:
1. Restart stream-counter service
2. The service rebuilds counts from active reservations on startup

## Test Result Template

Use this to document your test results:

```
Date: _______________
Test: Add Camera to Grid

Before:
- active_streams: ___
- DUBAI_POLICE count: ___

Action: Double-clicked cam-001-sheikh-zayed

After:
- active_streams: ___ (expected +1)
- DUBAI_POLICE count: ___ (expected +1)

✅ PASS / ❌ FAIL

Notes: _________________________________
```
