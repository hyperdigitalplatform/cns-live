# Playback Controls Testing Guide

## Quick Start Test

### Prerequisites
1. Dashboard is running: `npm run dev` (in dashboard directory)
2. Backend API is running with WebRTC playback support
3. At least one camera with Milestone recording available

### Basic Functionality Test (Single Cell)

1. **Setup Grid**
   - Open dashboard
   - Use a 2x2 or 3x3 grid layout
   - Drag a camera with recordings to Cell 0

2. **Switch to Playback Mode**
   - Hover over Cell 0
   - Click the "PLAYBACK" mode toggle button
   - Wait for timeline to load (shows orange sequence bars)
   - Verify: Control bar appears at bottom with play/pause and forward/backward buttons

3. **Test Play/Pause**
   - Click Play button (▶)
   - Verify: Video starts playing
   - Verify: Button changes to Pause (⏸)
   - Click Pause button
   - Verify: Video pauses
   - Verify: Button changes back to Play (▶)

4. **Test Forward Button (|>)**
   - Ensure playback is active
   - Note current time display (e.g., "6:10:46 AM, 2025-10-28")
   - Click Forward button (|>)
   - **Expected:**
     - Video jumps to start of next recording sequence
     - Time display updates to new sequence start time
     - WebRTC connection briefly shows "Connecting..." then "Connected"
     - Video plays from new sequence
   - Open browser console (F12)
   - **Expected log:** `Jumping from sequence X to X+1 at <timestamp>`

5. **Test Backward Button (<|)**
   - Click Forward a few times to get to sequence 3 or 4
   - Click Backward button (<|)
   - **Expected:**
     - Video jumps to start of previous recording sequence
     - Time display updates to previous sequence start time
     - WebRTC reconnects
   - **Expected log:** `Jumping from sequence X to X-1 at <timestamp>`

6. **Test Boundary Conditions**
   - Keep clicking Backward until you reach first sequence
   - Click Backward again
   - **Expected:** No action (stays at first sequence)
   - **Expected log:** `Cannot jump left: already at first sequence`
   - Click Forward repeatedly until last sequence
   - Click Forward again
   - **Expected:** No action (stays at last sequence)
   - **Expected log:** `Cannot jump right: already at last sequence`

7. **Test Timeline Click**
   - Click anywhere on the orange sequence bar
   - **Expected:**
     - Video jumps to clicked time
     - WebRTC reconnects
     - Current sequence index updates

---

## Cell Isolation Test (Multi-Cell)

### Test 1: Independent Forward/Backward

1. **Setup:**
   - Use 2x2 grid
   - Add same camera to Cell 0 and Cell 1
   - Switch both to playback mode
   - Start playback on both cells

2. **Test:**
   - Cell 0: Click Forward → jumps to sequence N+1
   - Cell 1: Should still be at original sequence (NO CHANGE)
   - Cell 1: Click Backward → jumps to sequence M-1
   - Cell 0: Should still be at sequence N+1 (NO CHANGE)

3. **Verify:**
   - Each cell maintains independent sequence index
   - Time displays show different times
   - No visual glitches or interference

### Test 2: Simultaneous Playback

1. **Setup:**
   - Use 3x3 grid
   - Add different cameras to Cells 0, 1, 2
   - Switch all to playback mode
   - Start playback on all cells

2. **Test:**
   - All cells play simultaneously
   - Click Forward on Cell 0
   - Cells 1 and 2 continue playing uninterrupted

3. **Verify:**
   - No audio/video glitches in other cells
   - Network tab shows separate WebRTC sessions
   - CPU/memory usage is reasonable

### Test 3: Play/Pause Independence

1. **Setup:**
   - 2x2 grid with 4 different cameras
   - All in playback mode

2. **Test:**
   - Play all 4 cells
   - Pause Cell 0 → Only Cell 0 pauses, others continue
   - Play Cell 0 → Cell 0 resumes
   - Pause Cell 2 → Only Cell 2 pauses
   - Jump sequence on Cell 1 → Only Cell 1 jumps

3. **Verify:**
   - Each cell's play/pause state is independent
   - No cross-cell interference

---

## Advanced Tests

### Test 4: Rapid Sequence Jumping

1. Click Forward button 5 times rapidly (< 1 second apart)
2. **Expected:**
   - All jumps are queued and executed
   - No WebRTC session errors
   - Final position is 5 sequences ahead
3. **Verify in console:**
   - 5 "Jumping from sequence..." logs
   - No connection errors

### Test 5: Timeline Drag While Playing

1. Start playback
2. Click and drag on timeline (horizontal movement)
3. **Expected:**
   - Debounced seek (updates after 500ms of no drag)
   - WebRTC reconnects to new time
   - Sequence index updates correctly
4. **Verify:**
   - Smooth seeking without excessive reconnections

### Test 6: Speed Change + Sequence Jump

1. Start playback at 1x speed
2. Change speed to 2x (via speed dropdown)
3. Click Forward to jump sequence
4. **Expected:**
   - Speed setting preserved
   - Sequence jump works normally
   - Playback continues at 2x speed

### Test 7: Zoom + Sequence Jump

1. Set zoom to 10 minutes
2. Jump to next sequence
3. **Expected:**
   - Zoom level preserved (still 10 minutes)
   - Timeline re-centers on new sequence
   - Sequence bars still visible

### Test 8: Mode Switch During Playback

1. Start playback in Cell 0
2. Switch Cell 0 back to LIVE mode
3. Switch Cell 0 back to PLAYBACK mode
4. **Expected:**
   - Timeline reloads
   - Sequence index resets to initial
   - No WebRTC connection leaks
5. **Verify in console:**
   - Old session closes cleanly
   - New session starts fresh

---

## Performance Tests

### Test 9: Memory Leak Check

1. Open Chrome DevTools → Memory tab
2. Take heap snapshot (Snapshot 1)
3. Perform 20 sequence jumps (forward/backward alternating)
4. Take heap snapshot (Snapshot 2)
5. **Expected:**
   - RTCPeerConnection count: 1 (only current session)
   - No accumulated MediaStream objects
   - Memory increase < 50MB

### Test 10: Network Efficiency

1. Open DevTools → Network tab
2. Filter: WS (WebSocket) + WebRTC
3. Perform sequence jump
4. **Expected:**
   - Old WebRTC connection closes (status: closed)
   - New WebRTC connection opens
   - No lingering connections
   - ICE candidate exchange completes < 2 seconds

### Test 11: CPU Usage

1. Open Task Manager / Activity Monitor
2. Play 4 cells simultaneously
3. Jump sequences on all cells
4. **Expected:**
   - CPU usage < 80% (on modern hardware)
   - No sustained 100% usage
   - Browser remains responsive

---

## Error Handling Tests

### Test 12: No Recording at Time

1. Switch to playback mode
2. Seek to a time with no recording (gap in sequences)
3. Click Forward
4. **Expected:**
   - Jumps to next available sequence
   - Error message if no sequences exist
   - UI remains functional

### Test 13: Network Interruption

1. Start playback
2. Disconnect network (Wi-Fi off)
3. Click Forward
4. **Expected:**
   - WebRTC connection fails gracefully
   - Error message displayed
   - Reconnect network → Can retry

### Test 14: Invalid Sequence Data

1. Use browser console to corrupt sequence data:
   ```javascript
   // In React DevTools, modify gridCells[0].playbackState.timelineData.sequences = []
   ```
2. Click Forward
3. **Expected:**
   - No crash
   - Button does nothing (logs "No sequences")
   - UI remains stable

---

## Browser Compatibility

Test on:
- ✅ Chrome/Edge (WebRTC native)
- ✅ Firefox (WebRTC native)
- ✅ Safari (WebRTC native on macOS/iOS)

---

## Known Limitations (Expected Behavior)

1. **No animation on jump** - Intentional (matches test-webrtc-playback.html)
2. **500ms debounce on timeline drag** - Performance optimization
3. **Buttons disabled at boundaries** - NOT implemented (just returns early)
4. **Sequence counter in UI** - NOT implemented (could be added)

---

## Debugging Tips

### Enable Verbose Logging

In `useWebRTCPlayback.ts`, all console logs are already present:
- `📡 WebRTC session created: {sessionId}`
- `✅ Remote description set`
- `✅ Data channel opened`
- `🔗 Connection state: {state}`
- `📊 WebRTC Stats: {bandwidth, packetLoss, ...}`

### Check Sequence Tracking

In browser console:
```javascript
// Get current state
const grid = document.querySelector('[data-grid]'); // Adjust selector
// Use React DevTools to inspect gridCells[index].playbackState.currentSequenceIndex
```

### Monitor WebRTC Sessions

In DevTools → Console, filter by "WebRTC":
- Look for session creation/cleanup
- Verify no duplicate sessions
- Check connection state transitions

### Inspect Timeline Data

In React DevTools → Components → StreamGridEnhanced:
```
gridCells[0].playbackState.timelineData.sequences = [
  { sequenceId, startTime, endTime, durationSeconds },
  ...
]
```

---

## Success Criteria

✅ **All tests pass** without errors
✅ **No cell interference** - each cell operates independently
✅ **No memory leaks** - heap snapshots show cleanup
✅ **No WebRTC errors** - connections open/close cleanly
✅ **UI remains responsive** - no freezing or lag
✅ **Sequence jumps work** - forward/backward navigate recordings
✅ **Boundary handling** - graceful stops at first/last sequence

---

## Regression Test Checklist

Before marking complete, verify:

- [ ] Single cell forward/backward works
- [ ] Multi-cell isolation verified (no interference)
- [ ] Play/pause state independent per cell
- [ ] Timeline click updates sequence index
- [ ] Boundary conditions handled gracefully
- [ ] WebRTC sessions close cleanly on jump
- [ ] No memory leaks after 20+ jumps
- [ ] Network connections don't accumulate
- [ ] Browser console shows expected logs
- [ ] No React warnings or errors
- [ ] Performance acceptable with 4+ cells
- [ ] Error handling works (no recordings, network failure)

---

**Last Updated:** 2025-01-01
**Test Coverage:** 14 test scenarios
**Status:** Ready for testing
