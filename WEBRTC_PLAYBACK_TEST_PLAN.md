# WebRTC Playback - Integration Test Plan

**Date**: 2025-10-29
**Version**: 1.0
**Status**: Ready for Testing

---

## 🎯 Test Objectives

1. Verify WebRTC playback works end-to-end
2. Ensure timeline integration functions correctly
3. Validate error handling and edge cases
4. Confirm performance meets requirements
5. Test multi-cell playback scenarios

---

## 🧪 Test Environment Setup

### Prerequisites
```bash
# 1. Ensure all services are running
docker-compose ps

# Expected services:
# - milestone-service (port 8085)
# - kong (port 8000)
# - dashboard (port 3000)
# - Other supporting services

# 2. Check Milestone VMS is accessible
curl -k https://192.168.1.11/IDP/connect/token

# 3. Verify you have test recordings
# - Camera ID with recordings
# - Recording date range
```

### Test Data Required
- **Camera ID**: `d47fa4e9-8171-4cc2-a421-95a3194f6a1d` (or your camera)
- **Recording Date Range**: Last 24 hours
- **Test Timestamps**: Times with known recordings

---

## 📋 Test Scenarios

### ✅ Scenario 1: Basic Playback

**Objective**: Verify basic WebRTC playback functionality

**Steps**:
1. [ ] Open dashboard at `http://localhost:3000`
2. [ ] Add a camera to the grid (drag or click)
3. [ ] Camera should show live stream
4. [ ] Click "Playback" toggle button
5. [ ] Wait for mode to switch
6. [ ] Timeline should appear with recording sequences (green bars)
7. [ ] Video should start playing from current time
8. [ ] Connection status shows "🟢 Connected"

**Expected Results**:
- ✅ Mode switches to playback
- ✅ Timeline loads within 2 seconds
- ✅ Recording sequences visible on timeline
- ✅ Video plays smoothly
- ✅ No console errors

**Pass Criteria**:
- All steps complete successfully
- Video playback quality is acceptable
- No JavaScript errors in console

---

### ✅ Scenario 2: Timeline Navigation

**Objective**: Test seeking and scrubbing functionality

**Steps**:
1. [ ] With playback active (from Scenario 1)
2. [ ] Click "Show" to expand timeline
3. [ ] Click on a different time on the timeline
4. [ ] Video should seek to that time within 3 seconds
5. [ ] Current time indicator (blue line) should move
6. [ ] Click and drag on the timeline (scrubbing)
7. [ ] Video should seek as you drag
8. [ ] Release mouse - video continues from new position

**Expected Results**:
- ✅ Timeline clicks seek immediately
- ✅ Scrubbing is smooth (no lag)
- ✅ Debouncing prevents rapid reconnections
- ✅ Timeline cursor updates correctly
- ✅ Time display updates to match position

**Pass Criteria**:
- Seeking completes within 3 seconds
- Scrubbing feels responsive
- No connection errors during rapid seeking

---

### ✅ Scenario 3: Playback Controls

**Objective**: Verify all playback controls work

**Steps**:
1. [ ] With playback active
2. [ ] Click **Play/Pause** button
3. [ ] Video should pause
4. [ ] Click again - video resumes
5. [ ] Click **Skip Back** (-10s)
6. [ ] Video seeks backward 10 seconds
7. [ ] Click **Skip Forward** (+10s)
8. [ ] Video seeks forward 10 seconds
9. [ ] Click **Fullscreen** button
10. [ ] Player enters fullscreen mode
11. [ ] Press ESC or click minimize
12. [ ] Player exits fullscreen

**Expected Results**:
- ✅ Play/Pause toggles correctly
- ✅ Skip buttons seek by 10 seconds
- ✅ Fullscreen works in both directions
- ✅ Controls remain functional in fullscreen

**Pass Criteria**:
- All controls respond within 500ms
- No control becomes unresponsive
- Video state persists correctly

---

### ✅ Scenario 4: Mode Switching

**Objective**: Test switching between live and playback modes

**Steps**:
1. [ ] Start with camera in **Live** mode
2. [ ] Note current live stream is playing
3. [ ] Click "Playback" toggle
4. [ ] Mode switches to playback
5. [ ] Timeline appears
6. [ ] Video plays from current time
7. [ ] Click "Live" toggle
8. [ ] Mode switches back to live
9. [ ] Timeline disappears
10. [ ] Live stream resumes

**Expected Results**:
- ✅ Smooth transition between modes
- ✅ Live stream reconnects properly
- ✅ Playback state is cleared when switching to live
- ✅ Can switch multiple times without issues

**Pass Criteria**:
- Mode switches complete within 5 seconds
- No memory leaks (check DevTools)
- Live stream quality maintained after switch

---

### ✅ Scenario 5: Multiple Cells Playback

**Objective**: Test concurrent playback in multiple grid cells

**Steps**:
1. [ ] Add **4 cameras** to grid (2x2 or similar)
2. [ ] Switch **Cell 1** to playback mode
3. [ ] Wait for Cell 1 to start playing
4. [ ] Switch **Cell 2** to playback mode
5. [ ] Both should play independently
6. [ ] Switch **Cell 3** to playback mode
7. [ ] All three should play independently
8. [ ] Switch **Cell 4** to playback mode
9. [ ] All four should play independently
10. [ ] Seek on Cell 1 - others should not be affected
11. [ ] Switch Cell 1 back to live - others continue playback

**Expected Results**:
- ✅ All cells can be in playback mode simultaneously
- ✅ Each cell operates independently
- ✅ Seeking in one cell doesn't affect others
- ✅ CPU usage remains reasonable (< 80%)
- ✅ Memory usage stable (no leaks)

**Pass Criteria**:
- 4 concurrent playback streams work
- System remains responsive
- No cross-cell interference

---

### ✅ Scenario 6: Error Handling

**Objective**: Verify graceful error handling

#### Test 6.1: No Recording at Time
**Steps**:
1. [ ] Switch to playback mode
2. [ ] Expand timeline
3. [ ] Click on a time with **no recording** (gap between sequences)
4. [ ] Observe behavior

**Expected**:
- ✅ Shows error message: "No recording available at this time"
- ✅ Offers to retry or select different time
- ✅ No crash or infinite loading

#### Test 6.2: Invalid Camera ID
**Steps**:
1. [ ] Manually trigger playback with invalid camera ID
2. [ ] (Developer test - modify code temporarily)

**Expected**:
- ✅ Returns 400 or 404 error
- ✅ Shows user-friendly error message
- ✅ Can recover by selecting valid camera

#### Test 6.3: Network Disconnection
**Steps**:
1. [ ] Start playback successfully
2. [ ] Disconnect network (WiFi/Ethernet)
3. [ ] Wait 10 seconds
4. [ ] Reconnect network

**Expected**:
- ✅ Shows "Disconnected" state
- ✅ Attempts to reconnect automatically
- ✅ Resumes playback when network returns

#### Test 6.4: Milestone Service Down
**Steps**:
1. [ ] Stop milestone-service: `docker-compose stop milestone-service`
2. [ ] Try to start playback

**Expected**:
- ✅ Shows error: "Service unavailable"
- ✅ Provides retry option
- ✅ Doesn't hang indefinitely

**Pass Criteria**:
- All error scenarios handled gracefully
- User receives clear feedback
- System can recover from errors

---

### ✅ Scenario 7: Edge Cases

**Objective**: Test boundary conditions and edge cases

#### Test 7.1: Very Old Recording
**Steps**:
1. [ ] Select timestamp from 7+ days ago
2. [ ] Start playback

**Expected**:
- ✅ Works if recording exists
- ✅ Shows appropriate error if too old

#### Test 7.2: Future Time
**Steps**:
1. [ ] Try to seek to future time
2. [ ] (Via timeline or time picker)

**Expected**:
- ✅ Prevents seeking beyond current time
- ✅ Shows message: "Cannot play future recordings"

#### Test 7.3: Very Long Session
**Steps**:
1. [ ] Start playback
2. [ ] Leave running for 10+ minutes
3. [ ] Interact periodically

**Expected**:
- ✅ Connection remains stable
- ✅ No memory leaks
- ✅ Token refresh works (if session > 1 hour)

#### Test 7.4: Rapid Mode Switching
**Steps**:
1. [ ] Rapidly toggle Live ↔ Playback 10 times
2. [ ] Click toggle quickly

**Expected**:
- ✅ Handles rapid switches gracefully
- ✅ No race conditions
- ✅ Ends in correct state

#### Test 7.5: Browser Compatibility
**Steps**:
1. [ ] Test in **Chrome** (latest)
2. [ ] Test in **Firefox** (latest)
3. [ ] Test in **Edge** (latest)
4. [ ] Test in **Safari** (if available)

**Expected**:
- ✅ Works in all major browsers
- ✅ WebRTC supported
- ✅ UI renders correctly

**Pass Criteria**:
- No crashes on edge cases
- Graceful degradation where needed
- Cross-browser compatibility

---

## 🔍 Performance Testing

### Test 8: Performance Metrics

**Objective**: Measure and validate performance

#### Metrics to Collect:
1. [ ] **Initial Connection Time**
   - Time from "Start Playback" to "Connected"
   - Target: < 3 seconds

2. [ ] **Seek Time**
   - Time from timeline click to video playing
   - Target: < 2 seconds

3. [ ] **CPU Usage**
   - With 1 cell in playback: ____%
   - With 4 cells in playback: ____%
   - Target: < 80%

4. [ ] **Memory Usage**
   - Initial: _____ MB
   - After 10 minutes: _____ MB
   - Increase: _____ MB
   - Target: < 200MB increase

5. [ ] **Network Bandwidth**
   - Per stream: _____ Mbps
   - 4 streams total: _____ Mbps

**Tools**:
- Chrome DevTools → Performance tab
- Chrome DevTools → Memory tab
- Chrome DevTools → Network tab

**Pass Criteria**:
- All metrics within targets
- No memory leaks over 10 minutes
- Smooth 30 FPS playback

---

## 📊 Test Results Template

### Test Execution Log

| Scenario | Date | Tester | Result | Notes |
|----------|------|--------|--------|-------|
| 1. Basic Playback | | | ⬜ Pass / ❌ Fail | |
| 2. Timeline Navigation | | | ⬜ Pass / ❌ Fail | |
| 3. Playback Controls | | | ⬜ Pass / ❌ Fail | |
| 4. Mode Switching | | | ⬜ Pass / ❌ Fail | |
| 5. Multiple Cells | | | ⬜ Pass / ❌ Fail | |
| 6. Error Handling | | | ⬜ Pass / ❌ Fail | |
| 7. Edge Cases | | | ⬜ Pass / ❌ Fail | |
| 8. Performance | | | ⬜ Pass / ❌ Fail | |

---

## 🐛 Bug Tracking

### Known Issues

| ID | Severity | Description | Status | Fix |
|----|----------|-------------|--------|-----|
| BUG-001 | | | | |
| BUG-002 | | | | |

### Severity Levels:
- **Critical**: Blocks functionality, no workaround
- **High**: Major feature broken, workaround exists
- **Medium**: Minor feature affected
- **Low**: Cosmetic or edge case

---

## ✅ Acceptance Criteria

Before marking Phase 4 complete, verify:

### Functionality
- [ ] All 8 test scenarios pass
- [ ] No critical or high severity bugs
- [ ] All core features work as designed

### Performance
- [ ] Connection time < 3 seconds
- [ ] Seek time < 2 seconds
- [ ] CPU usage < 80% with 4 cells
- [ ] No memory leaks

### Quality
- [ ] No console errors during normal use
- [ ] Error messages are user-friendly
- [ ] UI is responsive and intuitive

### Documentation
- [ ] Test results documented
- [ ] Known issues logged
- [ ] User documentation updated

---

## 🚀 Post-Testing Actions

### If All Tests Pass:
1. Update implementation plan to "Complete"
2. Create deployment checklist
3. Prepare user documentation
4. Plan production rollout

### If Issues Found:
1. Log all bugs with severity
2. Prioritize critical/high issues
3. Create fix tasks
4. Re-test after fixes

---

## 📝 Testing Notes

### Tips for Effective Testing:
- Test in a clean browser session (incognito mode)
- Clear cache between major test runs
- Keep browser DevTools open to catch errors
- Test with realistic network conditions
- Document exact steps to reproduce bugs

### Common Issues to Watch For:
- WebRTC connection failures
- Memory leaks from unclosed connections
- Race conditions in mode switching
- Timeline synchronization issues
- CORS errors in console

---

## 🎯 Next Steps After Testing

1. **Review Results**: Analyze all test data
2. **Fix Critical Bugs**: Address blocking issues
3. **Optimize**: Improve based on performance data
4. **Document**: Update user guides
5. **Deploy**: Roll out to production

---

**Test Coordinator**: _________________
**Test Date**: _________________
**Sign-off**: _________________
