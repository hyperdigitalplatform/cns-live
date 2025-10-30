# 🎯 Milestone WebRTC Playback Implementation Plan

**Status**: 🔴 Not Started
**Start Date**: 2025-10-29
**Target Completion**: 2025-11-04 (6 days)
**Priority**: HIGH

---

## 📋 Executive Summary

### Problem
- Current playback implementation uses non-existent REST API endpoint
- Trying to proxy raw video streams (incorrect approach)
- Frontend expects HLS but Milestone provides WebRTC
- Getting 500 Internal Server Error on playback attempts

### Solution
- Use Milestone's WebRTC API with `PlaybackTimeNode` parameter
- Implement WebRTC signaling (offer/answer/ICE) on backend
- Use WebRTC peer connection on frontend (similar to live streaming)
- Leverage existing WebRTC infrastructure from live streams

### Benefits
- ✅ Seamless playback across multiple sequences (Milestone handles it)
- ✅ Automatic gap skipping server-side
- ✅ No transcoding needed on our side
- ✅ Low latency playback
- ✅ Reuse existing WebRTC knowledge

---

## 🎯 Progress Overview

| Phase | Tasks | Status | Completion | Duration |
|-------|-------|--------|------------|----------|
| **Phase 1** | Backend WebRTC API | ✅ Complete | 4/4 | 2 days |
| **Phase 2** | Frontend WebRTC Player | ✅ Complete | 3/3 | 2 days |
| **Phase 3** | Timeline Integration | ✅ Complete | 2/2 | 1 day |
| **Phase 4** | Testing & Polish | 🔴 Not Started | 0/3 | 1 day |
| **TOTAL** | | 🟡 In Progress | **9/12** | **6 days** |

---

## 📅 Phase 1: Backend - Milestone WebRTC Playback API (2 days)

**Status**: 🔴 Not Started
**Owner**: Backend Team
**Dependencies**: Milestone VMS (version >= 2023 R1)

### Task 1.1: Implement Milestone WebRTC Client ⏱️ 4 hours

**Status**: ⬜ Not Started

**File**: `services/milestone-service/internal/client/webrtc_client.go` (NEW)

**Checklist**:
- [ ] Create `webrtc_client.go` file
- [ ] Implement `WebRTCPlaybackRequest` struct
- [ ] Implement `WebRTCSession` struct
- [ ] Implement `CreateWebRTCPlaybackSession()` method
  - [ ] Build request body with `playbackTimeNode`
  - [ ] POST to `/API/REST/v1/WebRTC/Session`
  - [ ] Parse and return sessionId and offerSDP
- [ ] Implement `UpdateWebRTCAnswer()` method
  - [ ] PUT to `/API/REST/v1/WebRTC/Session`
  - [ ] Send answerSDP from client
- [ ] Implement `SendICECandidate()` method
  - [ ] POST to `/API/REST/v1/WebRTC/IceCandidate`
  - [ ] Send ICE candidate to Milestone
- [ ] Implement `GetICECandidates()` method
  - [ ] GET from `/API/REST/v1/WebRTC/IceCandidate`
  - [ ] Retrieve server ICE candidates
- [ ] Add proper error handling
- [ ] Add logging for debugging
- [ ] Test with curl/Postman

**Acceptance Criteria**:
- ✅ Can create WebRTC playback session with Milestone
- ✅ Receives valid sessionId and offerSDP
- ✅ Can exchange ICE candidates bidirectionally
- ✅ Proper error messages on failure

**Testing**:
```bash
# Test WebRTC session creation
curl -X POST http://localhost:8085/api/v1/cameras/{cameraId}/playback/start \
  -H "Content-Type: application/json" \
  -d '{
    "playbackTime": "2025-10-29T10:00:00Z",
    "skipGaps": true,
    "speed": 1.0
  }'

# Expected response:
# {
#   "sessionId": "some-session-id",
#   "offerSDP": "{\"type\":\"offer\",\"sdp\":\"...\"}"
# }
```

---

### Task 1.2: Add Playback Handler ⏱️ 2 hours

**Status**: ⬜ Not Started

**File**: `services/milestone-service/internal/api/handlers.go`

**Checklist**:
- [ ] Remove broken `StreamPlayback()` function
- [ ] Create `WebRTCPlaybackRequest` struct
- [ ] Create `WebRTCAnswerRequest` struct
- [ ] Create `ICECandidateRequest` struct
- [ ] Implement `StartWebRTCPlayback()` handler
  - [ ] Parse request body
  - [ ] Validate playbackTime
  - [ ] Call `CreateWebRTCPlaybackSession()`
  - [ ] Return session data
- [ ] Implement `UpdateWebRTCAnswer()` handler
  - [ ] Parse answerSDP
  - [ ] Forward to Milestone
- [ ] Implement `SendICECandidate()` handler
  - [ ] Parse ICE candidate
  - [ ] Forward to Milestone
- [ ] Implement `GetICECandidates()` handler
  - [ ] Retrieve candidates from Milestone
  - [ ] Return as JSON array
- [ ] Add proper HTTP status codes
- [ ] Add error responses

**Acceptance Criteria**:
- ✅ All handlers respond correctly
- ✅ Proper error handling and status codes
- ✅ Request validation works
- ✅ Can be tested with curl

---

### Task 1.3: Update Routes ⏱️ 30 minutes

**Status**: ⬜ Not Started

**File**: `services/milestone-service/internal/api/router.go`

**Checklist**:
- [ ] Remove old broken route: `/cameras/:cameraId/playback/stream`
- [ ] Add new route: `POST /cameras/:cameraId/playback/start`
- [ ] Add new route: `PUT /playback/webrtc/answer`
- [ ] Add new route: `POST /playback/webrtc/ice`
- [ ] Add new route: `GET /playback/webrtc/ice/:sessionId`
- [ ] Test routes with `go run` or `go test`
- [ ] Verify no route conflicts

**Acceptance Criteria**:
- ✅ Old broken route removed
- ✅ New routes registered correctly
- ✅ Service starts without errors
- ✅ Routes accessible via curl

**Testing**:
```bash
# After starting service
curl http://localhost:8085/api/v1/cameras/test-id/playback/start
# Should return 400 (validation error) or 401 (auth required), NOT 404
```

---

### Task 1.4: Update Kong Routes ⏱️ 15 minutes

**Status**: ⬜ Not Started

**File**: `config/kong/kong.yml`

**Checklist**:
- [ ] Remove old route: `milestone-playback-stream`
- [ ] Add route: `milestone-playback-start`
  - [ ] Path: `~/api/v1/cameras/[^/]+/playback/start$`
  - [ ] Method: POST, OPTIONS
  - [ ] Tags: playback, webrtc
- [ ] Add route: `milestone-playback-webrtc-answer`
  - [ ] Path: `/api/v1/playback/webrtc/answer`
  - [ ] Method: PUT, OPTIONS
- [ ] Add route: `milestone-playback-webrtc-ice`
  - [ ] Path: `/api/v1/playback/webrtc/ice`
  - [ ] Methods: GET, POST, OPTIONS
- [ ] Reload Kong configuration
- [ ] Test through Kong gateway

**Acceptance Criteria**:
- ✅ Kong validates configuration (no syntax errors)
- ✅ Routes accessible through Kong (port 8000)
- ✅ CORS headers work for OPTIONS requests

**Testing**:
```bash
# Validate Kong config
docker exec cctv-kong kong config parse /etc/kong/kong.yml

# Reload Kong
docker exec cctv-kong kong reload

# Test through Kong
curl http://localhost:8000/api/v1/cameras/test-id/playback/start
```

---

### Phase 1 Completion Checklist

**Before moving to Phase 2, verify**:
- [ ] All 4 tasks completed
- [ ] Backend service starts without errors
- [ ] Can create WebRTC session via API
- [ ] ICE candidate exchange works
- [ ] Kong routes traffic correctly
- [ ] No 500 errors in logs
- [ ] Postman/curl tests pass

**Phase 1 Sign-off**: _________________ Date: _________

---

## 🎨 Phase 2: Frontend - WebRTC Playback Player (2 days)

**Status**: 🔴 Not Started
**Owner**: Frontend Team
**Dependencies**: Phase 1 completed

### Task 2.1: Create WebRTC Playback Hook ⏱️ 3 hours

**Status**: ⬜ Not Started

**File**: `dashboard/src/hooks/useWebRTCPlayback.ts` (NEW)

**Checklist**:
- [ ] Create `useWebRTCPlayback.ts` file
- [ ] Define `WebRTCPlaybackOptions` interface
- [ ] Create hook function with state management
  - [ ] State: `idle`, `connecting`, `connected`, `failed`, `disconnected`
  - [ ] Refs: `videoRef`, `pcRef`, `sessionIdRef`
- [ ] Implement WebRTC connection flow:
  - [ ] Step 1: Create RTCPeerConnection
  - [ ] Step 2: Request session from backend
  - [ ] Step 3: Set remote description (offerSDP)
  - [ ] Step 4: Create and set local description (answer)
  - [ ] Step 5: Send answer to backend
  - [ ] Step 6: Handle onicecandidate event
  - [ ] Step 7: Poll server ICE candidates
  - [ ] Step 8: Handle ontrack event (video stream)
  - [ ] Step 9: Monitor connection state
- [ ] Implement cleanup on unmount
- [ ] Add error handling
- [ ] Return `{ videoRef, state, error, stop }`
- [ ] Test hook with sample component

**Acceptance Criteria**:
- ✅ Hook establishes WebRTC connection
- ✅ Video stream attached to video element
- ✅ State changes work correctly
- ✅ Cleanup prevents memory leaks
- ✅ Errors handled gracefully

**Testing**:
```typescript
// Test component
const TestPlayer = () => {
  const { videoRef, state, error } = useWebRTCPlayback({
    cameraId: 'd47fa4e9-8171-4cc2-a421-95a3194f6a1d',
    playbackTime: new Date('2025-10-29T10:00:00Z'),
    skipGaps: true,
  });

  return (
    <div>
      <p>State: {state}</p>
      {error && <p>Error: {error}</p>}
      <video ref={videoRef} autoPlay />
    </div>
  );
};
```

---

### Task 2.2: Update RecordingPlayer Component ⏱️ 3 hours

**Status**: ⬜ Not Started

**File**: `dashboard/src/components/RecordingPlayer.tsx`

**Checklist**:
- [ ] Remove HLS.js imports and code
- [ ] Import `useWebRTCPlayback` hook
- [ ] Remove `hlsRef` and HLS-related state
- [ ] Use `useWebRTCPlayback` instead of HLS
- [ ] Update play/pause logic for WebRTC
- [ ] Update skip logic (change currentTime prop to trigger reconnect)
- [ ] Keep UI controls (play, pause, skip, fullscreen)
- [ ] Update loading/error states
- [ ] Remove volume controls (no audio in playback)
- [ ] Test with actual camera

**Acceptance Criteria**:
- ✅ Video plays using WebRTC
- ✅ Play/pause works
- ✅ Skip forward/backward works
- ✅ Loading state shows during connection
- ✅ Error state shows on failure
- ✅ No HLS.js errors in console
- ✅ Controls responsive and intuitive

**Testing**:
```typescript
// In StreamGridEnhanced or standalone test
<RecordingPlayer
  cameraId="d47fa4e9-8171-4cc2-a421-95a3194f6a1d"
  startTime={new Date('2025-10-29T00:00:00Z')}
  endTime={new Date('2025-10-29T23:59:59Z')}
  initialPlaybackTime={new Date('2025-10-29T10:00:00Z')}
  onPlaybackTimeChange={(time) => console.log('Time:', time)}
/>
```

---

### Task 2.3: Update StreamGridEnhanced ⏱️ 2 hours

**Status**: ⬜ Not Started

**File**: `dashboard/src/components/StreamGridEnhanced.tsx`

**Checklist**:
- [ ] Fix `handleModeChange()` function (line ~304)
  - [ ] Remove custom stream URL construction
  - [ ] Set `streamUrl: undefined` (not needed for WebRTC)
  - [ ] Keep sequence query logic
- [ ] Update `handleSeek()` function (line ~397)
  - [ ] Update currentTime in state
  - [ ] Let useWebRTCPlayback hook handle reconnection
- [ ] Update rendering section (line ~753)
  - [ ] Pass correct props to RecordingPlayer
  - [ ] Remove streamUrl prop
- [ ] Test mode switching (live ↔ playback)
- [ ] Test with multiple cells
- [ ] Verify no console errors

**Acceptance Criteria**:
- ✅ Can switch to playback mode
- ✅ RecordingPlayer receives correct props
- ✅ Timeline data loads correctly
- ✅ Can switch back to live mode
- ✅ Multiple cells work independently
- ✅ No errors in browser console

**Testing**:
```typescript
// Test scenarios:
// 1. Switch cell 1 to playback
// 2. Switch cell 2 to playback
// 3. Both should work independently
// 4. Switch cell 1 back to live
// 5. Cell 2 should continue playback
```

---

### Phase 2 Completion Checklist

**Before moving to Phase 3, verify**:
- [ ] All 3 tasks completed
- [ ] Video plays in browser using WebRTC
- [ ] Play/pause works correctly
- [ ] Skip forward/backward works
- [ ] No HLS.js errors in console
- [ ] Can switch between live and playback modes
- [ ] Multiple cells work simultaneously
- [ ] Loading and error states display correctly

**Phase 2 Sign-off**: _________________ Date: _________

---

## 📊 Phase 3: Timeline Integration (1 day)

**Status**: 🔴 Not Started
**Owner**: Frontend Team
**Dependencies**: Phase 2 completed

### Task 3.1: Implement Seek Functionality ⏱️ 4 hours

**Status**: ⬜ Not Started

**File**: `dashboard/src/components/StreamGridEnhanced.tsx`

**Checklist**:
- [ ] Update `handleSeek()` to trigger WebRTC reconnection
- [ ] Add debouncing to prevent rapid reconnections
- [ ] Update playback state properly
- [ ] Test seeking to different times
- [ ] Test seeking to gaps (should skip if skipGaps=true)
- [ ] Test seeking to future (should show error)
- [ ] Verify timeline cursor updates

**Acceptance Criteria**:
- ✅ Can seek to any time in timeline
- ✅ Video plays from new time correctly
- ✅ Debouncing prevents connection spam
- ✅ Gaps skipped automatically if enabled
- ✅ Timeline cursor shows current position

**Testing**:
```typescript
// Test scenarios:
// 1. Seek to 10:00 AM → plays from 10:00
// 2. Seek to 2:00 PM → plays from 2:00
// 3. Seek to gap → skips to next recording
// 4. Rapid seeking → debounced, doesn't crash
```

---

### Task 3.2: Add Timeline Scrubbing ⏱️ 4 hours

**Status**: ⬜ Not Started

**File**: `dashboard/src/components/playback/PlaybackControlBar.tsx`

**Checklist**:
- [ ] Add click handler to timeline
- [ ] Calculate time from click position
- [ ] Call `onSeek` with new time
- [ ] Add hover preview (optional)
- [ ] Add drag support for scrubbing (optional)
- [ ] Test clicking at different positions
- [ ] Verify seeks to correct time

**Acceptance Criteria**:
- ✅ Can click timeline to seek
- ✅ Calculates correct time from click position
- ✅ Visual feedback on hover
- ✅ Smooth scrubbing experience

**Testing**:
```typescript
// Test timeline scrubbing:
// 1. Click at 25% → should seek to startTime + 25% duration
// 2. Click at 50% → should seek to middle
// 3. Click at 75% → should seek to 75% through range
// 4. Hover → shows timestamp preview
```

---

### Phase 3 Completion Checklist

**Before moving to Phase 4, verify**:
- [ ] All 2 tasks completed
- [ ] Can seek to any time via controls
- [ ] Can seek by clicking timeline
- [ ] Debouncing works correctly
- [ ] Timeline updates reflect playback position
- [ ] Scrubbing is smooth and responsive

**Phase 3 Sign-off**: _________________ Date: _________

---

## 🧪 Phase 4: Testing & Polish (1 day)

**Status**: 🔴 Not Started
**Owner**: Full Team
**Dependencies**: Phase 3 completed

### Task 4.1: Integration Testing ⏱️ 4 hours

**Status**: ⬜ Not Started

**Test Scenarios**:

#### Scenario 1: Basic Playback
- [ ] Open dashboard
- [ ] Add camera to grid
- [ ] Switch to playback mode
- [ ] Video plays from current time
- [ ] Can see timeline with sequences
- [ ] Play/pause works
- [ ] Skip forward/backward works

#### Scenario 2: Timeline Navigation
- [ ] Switch to playback mode
- [ ] Click timeline at different positions
- [ ] Video seeks to clicked time
- [ ] Timeline cursor follows playback
- [ ] Sequence gaps shown correctly

#### Scenario 3: Multiple Cells
- [ ] Add 4 cameras to grid
- [ ] Switch cells 1 & 2 to playback
- [ ] Both play independently
- [ ] Cells 3 & 4 remain live
- [ ] Can control each separately

#### Scenario 4: Mode Switching
- [ ] Start in live mode
- [ ] Switch to playback
- [ ] Playback starts correctly
- [ ] Switch back to live
- [ ] Live stream resumes

#### Scenario 5: Error Handling
- [ ] Request playback at time with no recording
- [ ] Should show error message
- [ ] Request playback with invalid camera ID
- [ ] Should show error message
- [ ] Network disconnection during playback
- [ ] Should show disconnected state

#### Scenario 6: Performance
- [ ] 4 cells in playback mode simultaneously
- [ ] CPU usage acceptable (< 80%)
- [ ] Memory usage stable
- [ ] No memory leaks after 10 minutes
- [ ] Switching modes doesn't cause lag

#### Scenario 7: Edge Cases
- [ ] Seek to exact gap boundary
- [ ] Seek to end of recording
- [ ] Very fast seeking (spam clicks)
- [ ] Switch mode during loading
- [ ] Close browser during playback (cleanup check)

**Acceptance Criteria**:
- ✅ All 7 scenarios pass
- ✅ No console errors
- ✅ No memory leaks
- ✅ Performance acceptable

---

### Task 4.2: Error Handling ⏱️ 2 hours

**Status**: ⬜ Not Started

**Checklist**:
- [ ] Add error handling for no recordings
  - [ ] Show message: "No recording available at this time"
  - [ ] Suggest seeking to different time
- [ ] Add error handling for WebRTC connection failure
  - [ ] Show message: "Failed to connect to playback stream"
  - [ ] Add retry button
- [ ] Add error handling for authentication issues
  - [ ] Redirect to login if needed
- [ ] Add error handling for network timeouts
  - [ ] Show timeout message
  - [ ] Auto-retry with backoff
- [ ] Add error logging
  - [ ] Log errors to console with context
  - [ ] Send critical errors to backend (optional)
- [ ] Test all error scenarios

**Acceptance Criteria**:
- ✅ All error cases handled gracefully
- ✅ User-friendly error messages
- ✅ Errors logged for debugging
- ✅ Can recover from errors

---

### Task 4.3: Performance Optimization ⏱️ 2 hours

**Status**: ⬜ Not Started

**Checklist**:
- [ ] Add debouncing for seek operations
  - [ ] 500ms delay before triggering new connection
- [ ] Implement connection reuse
  - [ ] Don't reconnect if already at requested time
- [ ] Add WebRTC statistics logging
  - [ ] Log connection quality
  - [ ] Log bandwidth usage
  - [ ] Log packet loss
- [ ] Optimize ICE candidate polling
  - [ ] Stop polling after connection established
  - [ ] Use exponential backoff
- [ ] Add memory cleanup
  - [ ] Properly close peer connections
  - [ ] Clear video element srcObject
- [ ] Test performance with Chrome DevTools
  - [ ] Profile memory usage
  - [ ] Check for leaks
  - [ ] Verify cleanup

**Acceptance Criteria**:
- ✅ Seek operations debounced properly
- ✅ No unnecessary reconnections
- ✅ Statistics logged for monitoring
- ✅ No memory leaks
- ✅ Good performance with 4+ cells

---

### Phase 4 Completion Checklist

**Before considering complete, verify**:
- [ ] All 3 tasks completed
- [ ] All integration tests pass
- [ ] Error handling works correctly
- [ ] Performance is acceptable
- [ ] No memory leaks detected
- [ ] Code reviewed by team
- [ ] Documentation updated

**Phase 4 Sign-off**: _________________ Date: _________

---

## ✅ Final Acceptance Criteria

### Functional Requirements
- [ ] Can switch from live to playback mode
- [ ] Video plays at requested time
- [ ] Can seek to any time in timeline
- [ ] Timeline shows recording sequences and gaps
- [ ] Play/pause controls work
- [ ] Skip forward/backward works
- [ ] Multiple cells work independently
- [ ] Can switch back to live mode

### Non-Functional Requirements
- [ ] No 500 errors in playback
- [ ] WebRTC connection establishes within 3 seconds
- [ ] Seek operation completes within 2 seconds
- [ ] CPU usage < 80% with 4 playback cells
- [ ] Memory stable (no leaks)
- [ ] Works in Chrome, Firefox, Safari
- [ ] Responsive on 1920x1080 and 2560x1440

### Code Quality
- [ ] TypeScript types properly defined
- [ ] Error handling comprehensive
- [ ] Logging added for debugging
- [ ] No console warnings/errors
- [ ] Code follows project conventions
- [ ] Comments added for complex logic

---

## 🔧 Technical Reference

### Milestone WebRTC API

**Session Creation**:
```
POST /API/REST/v1/WebRTC/Session
Body:
{
  "deviceId": "camera-uuid",
  "playbackTimeNode": {
    "playbackTime": "2025-10-29T10:00:00Z",
    "skipGaps": true,
    "speed": 1.0
  }
}

Response:
{
  "sessionId": "session-uuid",
  "offerSDP": "{\"type\":\"offer\",\"sdp\":\"...\"}"
}
```

**Answer Update**:
```
PUT /API/REST/v1/WebRTC/Session
Body:
{
  "sessionId": "session-uuid",
  "answerSDP": "{\"type\":\"answer\",\"sdp\":\"...\"}"
}
```

**ICE Candidates**:
```
POST /API/REST/v1/WebRTC/IceCandidate
Body:
{
  "sessionId": "session-uuid",
  "candidate": { ... }
}

GET /API/REST/v1/WebRTC/IceCandidate?sessionId=session-uuid
Response:
{
  "candidates": [ ... ]
}
```

---

## 📚 Documentation References

- Milestone WebRTC JavaScript Sample: `mipsdk-samples-protocol/WebRTC_JavaScript/README.md`
- WebRTC API Docs: `mipsdk-samples-protocol/WebRTC_JavaScript/js/rest.js`
- Existing Implementation: `MILESTONE_PLAYBACK_API_EXPLAINED.md`
- Complete Solution: `MILESTONE_COMPLETE_SOLUTION.md`

---

## 🚨 Rollback Plan

If critical issues occur:

1. **Immediate Rollback** (< 5 minutes)
   ```bash
   # Revert to previous Kong config
   git checkout HEAD~1 config/kong/kong.yml
   docker exec cctv-kong kong reload

   # Disable playback mode in frontend
   # Set REACT_APP_ENABLE_PLAYBACK=false in .env
   ```

2. **Partial Rollback** (Phase-specific)
   - Phase 1 issues: Remove Kong routes, keep code
   - Phase 2 issues: Hide playback toggle in UI
   - Phase 3 issues: Disable timeline seeking
   - Phase 4 issues: Show "under maintenance" message

3. **Backup Files**
   - Keep `*.backup` files for all modified code
   - Tag git commit before each phase
   - Document rollback procedure per phase

---

## 📞 Support & Escalation

### Development Issues
- Backend issues: Check milestone-service logs
- Frontend issues: Check browser console
- WebRTC issues: Enable WebRTC debug logging

### Milestone VMS Issues
- Check Milestone version: Must be >= 2023 R1
- Verify WebRTC API enabled in API Gateway
- Check certificate issues (self-signed SSL)

### Escalation Path
1. Check this document for troubleshooting
2. Review Milestone documentation
3. Check milestone-service logs
4. Review browser network tab
5. Contact team lead if blocked > 2 hours

---

## 📊 Success Metrics

### Before Implementation
- ❌ Playback: 0% working (500 errors)
- ❌ User satisfaction: N/A (not usable)

### After Implementation (Target)
- ✅ Playback: 100% working
- ✅ Connection time: < 3 seconds
- ✅ Seek time: < 2 seconds
- ✅ User satisfaction: >= 90%

---

## 🎉 Completion Sign-off

### Development Team
- [ ] Backend Lead: _________________ Date: _________
- [ ] Frontend Lead: _________________ Date: _________
- [ ] QA Lead: _________________ Date: _________

### Product Team
- [ ] Product Owner: _________________ Date: _________
- [ ] Stakeholder: _________________ Date: _________

### Deployment
- [ ] Deployed to Dev: _________
- [ ] Deployed to Staging: _________
- [ ] Deployed to Production: _________

---

**Document Version**: 1.0
**Last Updated**: 2025-10-29
**Next Review**: After Phase 1 completion
