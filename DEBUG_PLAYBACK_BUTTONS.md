# Debug Playback Buttons

## Step 1: Open Browser Console

1. Open http://localhost:3000
2. Press F12 to open Developer Tools
3. Click "Console" tab
4. Keep it open while testing

## Step 2: Set Up Playback

1. Add a camera with Milestone recordings to a grid cell
2. Switch to PLAYBACK mode
3. Wait for timeline to load (you should see orange sequence bars)
4. Click the Play button to start playback

## Step 3: Test Buttons

### Test Forward Button (|>)

1. Click the Forward button
2. **Check console for:**
   - ✅ `Jumping from sequence X to Y at <timestamp>`
   - ✅ `📡 WebRTC session created: {sessionId}`
   - ❌ Any red error messages

### Test Backward Button (<|)

1. First click Forward a few times to get to sequence 2 or 3
2. Click the Backward button
3. **Check console for same logs as above**

## Step 4: Manual Debug

Paste this into the browser console to inspect the state:

```javascript
// Find React root
const root = document.querySelector('#root').__reactInternalInstance$;

// This might not work directly - React DevTools is better
// Install React DevTools extension if you don't have it
```

## Better Approach: Use React DevTools

1. Install React DevTools extension for Chrome
2. Open DevTools → Components tab
3. Find `StreamGridEnhanced` component
4. Look at the state:
   - `gridCells[0].playbackState.currentSequenceIndex` - should be a number (0, 1, 2, etc.)
   - `gridCells[0].playbackState.timelineData.sequences` - should be an array
   - `gridCells[0].playbackState.isPlaying` - should be true/false

## Step 5: Check Network Tab

1. DevTools → Network tab
2. Click Forward button
3. **You should see:**
   - POST to `/api/v1/cameras/{id}/playback/start` with NEW playbackTime
   - PUT to `/api/v1/playback/webrtc/answer`
   - GET to `/api/v1/playback/webrtc/ice/{sessionId}`

## Common Issues

### Issue 1: Buttons Click But Nothing Happens

**Possible Cause:** No sequences loaded or currentSequenceIndex not set

**Check:**
```javascript
// In React DevTools, check:
gridCells[0].playbackState.timelineData.sequences.length // Should be > 0
gridCells[0].playbackState.currentSequenceIndex // Should be >= 0
```

### Issue 2: Console Shows "Cannot jump: already at first/last sequence"

**This is CORRECT** - You're at the boundary. Try clicking the opposite button.

### Issue 3: No Console Logs At All

**Possible Cause:** Event handler not attached

**Check:** Hard refresh the page (Ctrl+Shift+R) to clear cache

### Issue 4: Error: "Cannot read property 'sequences' of null"

**Possible Cause:** Timeline not loaded yet

**Fix:** Wait for orange sequence bars to appear before clicking buttons

## Expected Console Output

### When Forward button works correctly:

```
Jumping from sequence 2 to 3 at 2025-10-27T18:45:00.000Z
🔗 Connection state: connecting
📡 WebRTC session created: abc123-def456-...
✅ Remote description set
✅ Data channel created
✅ Local description set
✅ Answer SDP sent
🔗 Connection state: connected
📊 WebRTC Stats: { bandwidth: '1.23 Mbps', ... }
```

### When Backward button works correctly:

```
Jumping from sequence 3 to 2 at 2025-10-27T18:30:00.000Z
(same WebRTC logs as above)
```

### When at boundary:

```
Cannot jump right: already at last sequence
(no WebRTC logs - this is correct)
```

## What to Report Back

Please share:

1. **Console output** when clicking buttons (copy/paste the logs)
2. **Network tab** - screenshot of requests after clicking Forward
3. **React DevTools** - screenshot of `gridCells[0].playbackState` object
4. **Specific behavior** - what exactly happens (or doesn't happen) when you click

This will help me understand what's not working!
