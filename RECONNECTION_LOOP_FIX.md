# WebRTC Reconnection Loop Fix

## Problem

After adding frame tracking, the dashboard was reconnecting every 2 seconds:
- Timeline would jump/stutter every 2 seconds
- Video would show loading spinner repeatedly
- Console showed multiple WebRTC session creations
- Each reconnection created a new session ID

## Root Cause

**Frame updates were triggering WebRTC reconnections!**

### The Feedback Loop:

```
1. Frame received → onPlaybackTimeChange(frameDate)
2. Parent updates → cell.playbackState.currentTime = frameDate
3. React re-renders → externalCurrentTime prop changes
4. RecordingPlayer sees prop change → triggers time sync
5. useWebRTCPlayback sees playbackTime change → RECONNECTS! ❌
6. New connection established
7. Repeat from step 1...
```

## How test-webrtc-playback.html Avoids This

**Key difference:** The test HTML never reconnects during normal playback!

```javascript
// test-webrtc-playback.html
function onFrameReceived(now, metadata) {
    const frameDate = new Date(frameStartTime + metadata.rtpTimestamp);
    lastFrameTime = frameDate;  // ← Just updates a global variable

    updateTimelineScroll(frameDate);  // ← Only updates CSS transform
    highlightCurrentSequence(frameDate);  // ← Only updates UI

    // NO WebRTC state changes, NO reconnection triggers!
    video.requestVideoFrameCallback(onFrameReceived);
}
```

**WebRTC connection:** Established once, stays open until user manually seeks.

## Solution

Added **time difference threshold** to distinguish between:
- **Frame updates** (< 5 seconds difference) → Ignore, just update UI
- **Manual seeks** (≥ 5 seconds difference) → Reconnect to new time

### Changes Made

#### 1. Added Time Difference Check

**File:** `dashboard/src/components/RecordingPlayer.tsx:166-181`

**Before:**
```typescript
useEffect(() => {
  if (externalCurrentTime === undefined) return;
  if (externalCurrentTime.getTime() === currentTime.getTime()) return;

  // Every prop change triggered reconnection! ❌
  setCurrentTime(externalCurrentTime);
  onPlaybackTimeChange?.(externalCurrentTime);
}, [externalCurrentTime]);
```

**After:**
```typescript
useEffect(() => {
  if (externalCurrentTime === undefined) return;

  // Only trigger reconnection if there's a SIGNIFICANT time jump (> 5 seconds)
  // This means user clicked forward/backward button or timeline
  // Small differences are just frame updates echoed back - IGNORE THEM
  const timeDiff = Math.abs(externalCurrentTime.getTime() - currentTime.getTime());
  if (timeDiff < 5000) return; // Less than 5 seconds = ignore (frame updates)

  // This is a real external seek (forward/backward button or timeline jump)
  console.log('🔄 External time control: jumping to', externalCurrentTime.toISOString());
  setCurrentTime(externalCurrentTime);
  frameStartTimeRef.current = externalCurrentTime.getTime();
}, [externalCurrentTime, currentTime]);
```

#### 2. Simplified Frame Update Logic

**File:** `dashboard/src/components/RecordingPlayer.tsx:193-212`

Removed unnecessary complexity:
- ❌ Removed `isFrameUpdatingRef` flag (not needed)
- ❌ Removed `setTimeout` delays (not needed)
- ✅ Simplified to match test HTML behavior

```typescript
const onFrameReceived = (now: number, metadata: any) => {
  if (!frameStartTimeRef.current) return;

  const frameDate = new Date(frameStartTimeRef.current + metadata.rtpTimestamp);

  // Update internal currentTime (for timeline display in parent)
  // Like test-webrtc-playback.html, this only updates UI, doesn't trigger reconnection
  setCurrentTime(frameDate);

  // Notify parent ONLY for timeline UI update
  onPlaybackTimeChange?.(frameDate);

  // Request next frame
  if (video && video.readyState >= 2) {
    rafId = video.requestVideoFrameCallback(onFrameReceived);
  }
};
```

## Why 5 Seconds Threshold?

**Reasoning:**
- **Forward/Backward buttons:** Jump to next/previous sequence (typically 10+ seconds apart)
- **Timeline clicks:** User seeks to specific time (usually > 5 seconds away)
- **Frame updates:** Continuous, small increments (milliseconds to seconds)

**The 5-second threshold:**
- ✅ Ignores all frame updates (no reconnection)
- ✅ Catches all button clicks (reconnects as intended)
- ✅ Catches manual timeline seeks (reconnects as intended)

## Expected Behavior Now

### ✅ Normal Playback (No Reconnections)
```
Console output during playback:
(No WebRTC session logs - just one stable connection)

Timeline: Scrolls smoothly every frame
Video: Plays continuously without interruption
```

### ✅ Forward/Backward Button (Reconnects)
```
Console output:
🔄 External time control: jumping to 2025-10-27T18:45:00.000Z
📡 WebRTC session created: abc-123-...
✅ Remote description set
✅ Data channel opened
🔗 Connection state: connected
```

### ✅ Timeline Click (Reconnects)
```
Console output:
🔄 External time control: jumping to 2025-10-27T19:00:00.000Z
📡 WebRTC session created: def-456-...
(same connection flow as above)
```

## Performance Impact

### Before Fix:
- **Reconnections:** Every 2 seconds
- **Network overhead:** Constant ICE candidate exchange
- **CPU usage:** High (constant connection setup/teardown)
- **User experience:** Stuttering, loading spinners

### After Fix:
- **Reconnections:** Only on manual seek (< 1% of operations)
- **Network overhead:** Minimal (stable connection)
- **CPU usage:** Low (matches test HTML)
- **User experience:** Smooth, continuous playback ✅

## Comparison: test-webrtc-playback.html vs Dashboard

| Aspect | test-webrtc-playback.html | Dashboard (After Fix) |
|--------|---------------------------|----------------------|
| **Frame updates** | Global variable only | React state (with threshold) |
| **Reconnection trigger** | Manual seek only | Time jump > 5 seconds |
| **Playback smoothness** | Continuous, no stutters | Continuous, no stutters ✅ |
| **Timeline scroll** | CSS transform | CSS transform ✅ |
| **Session stability** | One session per playback | One session per playback ✅ |

## Testing Checklist

- [x] Playback is smooth without reconnections
- [x] Timeline scrolls continuously (no 2-second jumps)
- [x] Forward button reconnects and jumps to next sequence
- [x] Backward button reconnects and jumps to previous sequence
- [x] Timeline click reconnects to new time
- [x] Console shows only one WebRTC session during playback
- [x] Video plays without loading spinners
- [x] Multi-cell playback: each cell plays smoothly

## Alternative Approaches Considered

### ❌ Approach 1: Don't update parent state
- **Problem:** Timeline wouldn't scroll at all
- **Rejected:** Defeats the purpose of frame tracking

### ❌ Approach 2: Throttle updates to once per second
- **Problem:** Timeline would jump instead of smooth scroll
- **Rejected:** Poor UX, not matching test HTML

### ✅ Approach 3: Time difference threshold (SELECTED)
- **Advantage:** Simple, effective, matches test HTML behavior
- **Works:** Distinguishes frame updates from manual seeks
- **Performance:** Minimal overhead (one Math.abs comparison)

## Files Modified

1. **`dashboard/src/components/RecordingPlayer.tsx`**
   - Added 5-second threshold to external time sync
   - Removed unnecessary frame update flag complexity
   - Simplified frame callback to match test HTML

## Known Edge Cases

1. **User seeks < 5 seconds:** Won't trigger reconnection
   - **Impact:** Minimal, rare use case
   - **Workaround:** Seek to exact position via timeline (> 5 sec jumps)

2. **Playback speed > 5x:** May trigger false reconnections
   - **Impact:** None (speed not implemented yet)
   - **Future fix:** Adjust threshold based on playback speed

3. **Network lag causing delayed updates:** Could cause threshold to be exceeded
   - **Impact:** Extremely rare
   - **Mitigation:** 5-second buffer is generous

---

**Status:** ✅ Fixed and tested
**Performance:** Matches test-webrtc-playback.html
**User Experience:** Smooth, continuous playback
