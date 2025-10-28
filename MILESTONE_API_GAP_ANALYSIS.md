# Milestone XProtect API Gap Analysis

**Date:** 2025-10-27
**Server:** 192.168.1.11
**Status:** ⚠️ CRITICAL GAPS IDENTIFIED

---

## Executive Summary

After comprehensive API discovery and testing against the actual Milestone server, **significant gaps exist** between what our implementation plan assumed and what the Milestone XProtect API actually provides.

### Overall Status

| Requirement | Assumed API | Actual API | Status |
|-------------|-------------|------------|--------|
| **Authentication** | Basic Auth / NTLM | OAuth 2.0 | ⚠️ Different method |
| **Camera Discovery** | GET `/cameras` | GET `/api/rest/v1/cameras` | ✅ Available |
| **Manual Recording Start** | POST `/cameras/{id}/recordings/start` | ❌ Does NOT exist | ❌ **CRITICAL GAP** |
| **Manual Recording Stop** | POST `/cameras/{id}/recordings/stop` | ❌ Does NOT exist | ❌ **CRITICAL GAP** |
| **Query Recordings** | GET `/cameras/{id}/recordings` | ❌ Does NOT exist | ❌ **CRITICAL GAP** |
| **Query Sequences** | GET `/cameras/{id}/sequences` | ❌ Does NOT exist | ❌ **CRITICAL GAP** |
| **Playback Stream** | GET `/cameras/{id}/playback/stream` | ❌ Does NOT exist | ❌ **CRITICAL GAP** |
| **Timeline Data** | GET `/cameras/{id}/timeline` | ❌ Does NOT exist | ❌ **CRITICAL GAP** |

---

## 1. Authentication - DIFFERENT METHOD

### ❌ What We Assumed (WRONG)

```http
POST /api/rest/v1/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}

Response: { "token": "...", "expires_in": 3600 }
```

### ✅ What Actually Exists (CORRECT)

```http
POST /API/IDP/connect/token
Content-Type: application/x-www-form-urlencoded

grant_type=password
username=raam
password=Ilove#123
client_id=GrantValidatorClient

Response:
{
  "access_token": "eyJhbGci...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "scope": "managementserver"
}
```

**Impact:** Medium - Different endpoint and request format
**Fix Required:** Update authentication code to use OAuth 2.0 password grant flow
**Status:** ✅ Verified working on server

---

## 2. Camera Discovery - AVAILABLE ✅

### ✅ What We Assumed

```http
GET /api/rest/v1/cameras
Authorization: Bearer <token>
```

### ✅ What Actually Exists

```http
GET https://192.168.1.11/api/rest/v1/cameras
Authorization: Bearer {token}
```

**Status:** ✅ **VERIFIED WORKING**
**Response:** Returns 2 cameras with all required fields including:
- `id` (camera UUID)
- `displayName`
- `ptzEnabled`
- `recordingEnabled`
- `manualRecordingTimeoutEnabled`
- `manualRecordingTimeoutMinutes` (15 minutes - matches requirement!)

**Fix Required:** None - works as expected

---

## 3. Manual Recording Control - CRITICAL GAP ❌

### ❌ What We Assumed (DOES NOT EXIST)

```http
POST /api/rest/v1/cameras/{cameraId}/recordings/start
Authorization: Bearer <token>
Content-Type: application/json

{
  "durationSeconds": 900,
  "triggerBy": "user_dashboard",
  "description": "Manual recording"
}
```

**Testing Result:** HTTP 404 - Endpoint does NOT exist

### ⚠️ What Actually Works (ALTERNATIVE METHODS)

#### Option A: Events API (Recommended)

```http
POST https://192.168.1.11/api/rest/v1/events
Authorization: Bearer {token}
Content-Type: application/json

{
  "source": {
    "id": "{cameraId}",
    "type": "Camera"
  },
  "type": {
    "id": "ManualRecordingTrigger"  // Need to verify correct event type
  },
  "message": "Start manual recording",
  "timestamp": "2025-10-27T12:00:00Z"
}
```

**Status:** ⚠️ Endpoint exists but needs correct event type ID
**Action Required:** Query `/api/rest/v1/eventTypes` to find recording event type

#### Option B: Bookmarks newReference (Alternative)

```http
POST https://192.168.1.11/api/rest/v1/bookmarks?task=newReference
Authorization: Bearer {token}
Content-Type: application/json

{
  "deviceId": "{cameraId}",
  "timeBegin": "2025-10-27T12:00:00Z"
}
```

**Status:** ⚠️ Endpoint exists, triggers rules without creating persistent bookmark
**Action Required:** Test if this actually starts recording

**Impact:** 🔴 **CRITICAL** - Core feature of our implementation
**Fix Required:** Implement Events API or Bookmarks newReference approach
**Estimated Effort:** 2-3 days

---

## 4. Query Recordings/Sequences - DOES NOT EXIST ❌

### ❌ What We Assumed (DOES NOT EXIST)

```http
GET /api/rest/v1/cameras/{cameraId}/recordings
Authorization: Bearer <token>

Query Parameters:
  - startTime: ISO8601
  - endTime: ISO8601

Response:
{
  "recordings": [
    {
      "id": "rec_123",
      "startTime": "2025-10-27T10:00:00Z",
      "endTime": "2025-10-27T10:15:00Z",
      "duration": 900
    }
  ]
}
```

**Testing Result:** HTTP 404 - Endpoint does NOT exist

### ⚠️ What Actually Works (ALTERNATIVE)

#### Use Bookmarks API to Search

```http
POST https://192.168.1.11/api/rest/v1/bookmarks?task=searchTime
Authorization: Bearer {token}
Content-Type: application/json

{
  "time": "2025-10-27T12:00:00Z",
  "timeSpanBefore": 3600,
  "timeSpanAfter": 3600,
  "deviceIds": ["{cameraId}"]
}

Response:
{
  "bookmarks": [
    {
      "id": "bookmark_123",
      "header": "Manual Recording",
      "timeBegin": "2025-10-27T10:00:00Z",
      "timeEnd": "2025-10-27T10:15:00Z",
      "deviceId": "{cameraId}"
    }
  ]
}
```

**Status:** ⚠️ Endpoint exists but returns bookmarks, not recording sequences
**Limitation:** Only returns recordings that have bookmarks - may miss unbookmarked recordings

**Impact:** 🔴 **CRITICAL** - Cannot query all recordings, only bookmarked ones
**Fix Required:** Use Bookmarks API for tagged recordings, may need alternative for continuous recordings
**Estimated Effort:** 1-2 days

---

## 5. Playback - COMPLETELY DIFFERENT APPROACH ❌

### ❌ What We Assumed (DOES NOT EXIST)

```http
GET /api/rest/v1/cameras/{cameraId}/playback/stream
Authorization: Bearer <token>

Query Parameters:
  - startTime: ISO8601
  - speed: float (1.0 = normal)

Response: HLS manifest or video stream
```

**Testing Result:** HTTP 404 - Endpoint does NOT exist

### ✅ What Actually Works (WebRTC - COMPLETELY DIFFERENT)

#### Step 1: Create WebRTC Session

```http
POST https://192.168.1.11/api/rest/v1/webRTC/session
Authorization: Bearer {token}
Content-Type: application/json

{
  "deviceId": "{cameraId}",
  "playbackTimeNode": {
    "playbackTime": "2025-10-27T12:00:00Z",
    "speed": 1.0,
    "skipGaps": true
  },
  "offer": {
    "type": "offer",
    "sdp": "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\n..."
  },
  "iceServers": [
    { "urls": ["stun:stun.l.google.com:19302"] }
  ]
}

Response:
{
  "sessionId": "7acf1727-5dfa-4fd2-8b47-a9e32d683ec6",
  "offerSDP": { ... },
  "answerSDP": "Uninitialised",
  "iceServers": []
}
```

#### Step 2: Establish WebRTC Peer Connection

**Requires:**
- WebRTC implementation in frontend (JavaScript)
- SDP offer/answer exchange
- ICE candidate handling
- Media stream handling

**Status:** ✅ Endpoint verified working, but requires full WebRTC implementation

**Impact:** 🔴 **CRITICAL** - Playback works completely differently than assumed
**Fix Required:** Implement full WebRTC client (frontend + backend)
**Estimated Effort:** 5-7 days (complex WebRTC implementation)

---

## 6. Timeline Data - DOES NOT EXIST ❌

### ❌ What We Assumed (DOES NOT EXIST)

```http
GET /api/rest/v1/cameras/{cameraId}/recordings/timeline
Authorization: Bearer <token>

Query Parameters:
  - startTime: ISO8601
  - endTime: ISO8601
  - interval: int (seconds)

Response:
{
  "timeline": [
    {
      "timestamp": "2025-10-27T10:00:00Z",
      "hasRecording": true,
      "duration": 300
    },
    {
      "timestamp": "2025-10-27T10:05:00Z",
      "hasRecording": false,
      "duration": 60
    }
  ]
}
```

**Testing Result:** HTTP 404 - Endpoint does NOT exist

### ⚠️ What Might Work (ALTERNATIVE)

#### Option A: Use Bookmarks Search

Query bookmarks and build timeline from results - but only shows bookmarked segments.

#### Option B: WebRTC Metadata

WebRTC playback might provide timeline data during playback session - needs investigation.

**Impact:** 🟠 **HIGH** - Timeline visualization may be limited
**Fix Required:** Build timeline from bookmark data or investigate WebRTC metadata
**Estimated Effort:** 2-3 days

---

## 7. Snapshot/Thumbnail - NOT IN PLAN BUT AVAILABLE

### ✅ Available (Bonus Feature)

```http
GET /api/rest/v1/alarms/{alarmId}/snapshots
Authorization: Bearer {token}
```

Alarms API supports attaching snapshots - could be adapted for recording thumbnails.

---

## Complete API Comparison Table

| Feature | Implementation Plan Assumption | Actual Milestone API | Verification Status | Fix Complexity |
|---------|-------------------------------|---------------------|-------------------|----------------|
| **Authentication** | POST `/api/rest/v1/login` (Basic) | POST `/API/IDP/connect/token` (OAuth 2.0) | ✅ Verified | 🟡 Medium |
| **List Cameras** | GET `/api/rest/v1/cameras` | GET `/api/rest/v1/cameras` | ✅ Verified | 🟢 None |
| **Get Camera** | GET `/api/rest/v1/cameras/{id}` | GET `/api/rest/v1/cameras/{id}` | ✅ Verified | 🟢 None |
| **Start Recording** | POST `/cameras/{id}/recordings/start` | ❌ Does NOT exist → Use Events or Bookmarks API | ⚠️ Alternative exists | 🔴 High |
| **Stop Recording** | POST `/cameras/{id}/recordings/stop` | ❌ Does NOT exist → Use Events API | ⚠️ Alternative exists | 🔴 High |
| **Get Recording Status** | GET `/cameras/{id}/recordings/status` | ❌ Does NOT exist | ❌ No alternative | 🔴 Critical |
| **List Recordings** | GET `/cameras/{id}/recordings` | ❌ Does NOT exist → Use Bookmarks API | ⚠️ Partial alternative | 🔴 High |
| **Query Sequences** | GET `/cameras/{id}/sequences` | ❌ Does NOT exist → Use Bookmarks API | ⚠️ Partial alternative | 🔴 High |
| **Playback Stream** | GET `/playback/stream` (HLS) | ❌ Does NOT exist → Use WebRTC | ✅ WebRTC verified | 🔴 Critical |
| **Timeline Data** | GET `/timeline` | ❌ Does NOT exist → Build from bookmarks | ⚠️ Workaround needed | 🟠 High |
| **Snapshots** | Not in plan | GET `/alarms/{id}/snapshots` | 📚 Documented | 🟢 Bonus |
| **Bookmarks** | Not in plan | POST `/bookmarks` | ⚠️ Tested (partial) | 🟢 Available |
| **Evidence Locks** | Not in plan | POST `/evidenceLocks` | 📚 Documented | 🟢 Available |
| **Alarms** | Not in plan | GET/POST `/alarms` | ⚠️ Exists | 🟢 Available |

**Legend:**
- ✅ Verified working
- ⚠️ Exists but needs more testing
- ❌ Does not exist
- 📚 Documented but not tested
- 🟢 Easy/None - 🟡 Medium - 🟠 High - 🔴 Critical

---

## Critical Gaps Summary

### 🔴 CRITICAL (Blocking Implementation)

1. **Manual Recording Control**
   - **Missing:** Direct start/stop endpoints
   - **Alternative:** Events API or Bookmarks newReference
   - **Status:** Not yet verified to work
   - **Action:** Test both approaches immediately

2. **Video Playback**
   - **Missing:** HLS/RTSP playback endpoints
   - **Alternative:** WebRTC (completely different architecture)
   - **Status:** Endpoint verified but requires full WebRTC implementation
   - **Action:** Implement WebRTC client (5-7 days effort)

3. **Recording Status/Progress**
   - **Missing:** No way to check if recording is active
   - **Alternative:** None found
   - **Status:** May need to track locally
   - **Action:** Implement local state tracking

### 🟠 HIGH (Significant Rework Needed)

4. **Recording Sequences Query**
   - **Missing:** Direct recordings query endpoint
   - **Alternative:** Bookmarks API (only tagged recordings)
   - **Status:** May miss untagged recordings
   - **Action:** Use bookmarks + investigate alternatives

5. **Timeline Data**
   - **Missing:** Timeline aggregation endpoint
   - **Alternative:** Build from bookmark search results
   - **Status:** Possible but limited
   - **Action:** Client-side timeline building

### 🟡 MEDIUM (Different but Manageable)

6. **Authentication**
   - **Missing:** Basic auth login endpoint
   - **Alternative:** OAuth 2.0 works fine
   - **Status:** Verified working
   - **Action:** Update auth code (1 day)

---

## Revised Implementation Requirements

### Phase 1: Core APIs (Must Have)

| API | Status | Action Required | Effort |
|-----|--------|----------------|--------|
| OAuth 2.0 Authentication | ✅ Working | Update implementation | 1 day |
| Camera Discovery | ✅ Working | No changes needed | 0 days |
| Events API (Recording) | ⚠️ Exists | Test event types | 2 days |
| Bookmarks newReference | ⚠️ Exists | Test recording trigger | 1 day |
| WebRTC Session Creation | ✅ Working | Implement full WebRTC | 7 days |

**Total Effort:** ~11 days

### Phase 2: Timeline & Search (Should Have)

| API | Status | Action Required | Effort |
|-----|--------|----------------|--------|
| Bookmarks Search | ⚠️ Exists | Test and integrate | 2 days |
| Timeline Building | ❌ Missing | Build client-side | 2 days |
| Sequence Display | Via Bookmarks | Adapt to bookmarks | 1 day |

**Total Effort:** ~5 days

### Phase 3: Advanced Features (Nice to Have)

| API | Status | Action Required | Effort |
|-----|--------|----------------|--------|
| Evidence Locks | 📚 Documented | Test and integrate | 1 day |
| Alarms Integration | ⚠️ Exists | Test and integrate | 2 days |
| Snapshots | ✅ Available | Implement | 1 day |

**Total Effort:** ~4 days

---

## Recommended Actions (Priority Order)

### IMMEDIATE (This Week)

1. ✅ **Test Events API for Recording Control**
   ```bash
   # Get event types
   curl -k "https://192.168.1.11/api/rest/v1/eventTypes" \
     -H "Authorization: Bearer $TOKEN"

   # Find recording event type and test
   ```

2. ✅ **Test Bookmarks newReference**
   ```bash
   curl -k -X POST "https://192.168.1.11/api/rest/v1/bookmarks?task=newReference" \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"deviceId":"a8a8b9dc-3995-49ed-9b00-62caac2ce74a"}'
   ```

3. ✅ **Test Bookmarks Search**
   ```bash
   curl -k -X POST "https://192.168.1.11/api/rest/v1/bookmarks?task=searchTime" \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"time":"2025-10-27T12:00:00Z","timeSpanBefore":3600,"timeSpanAfter":3600,"deviceIds":["a8a8b9dc-3995-49ed-9b00-62caac2ce74a"]}'
   ```

### SHORT TERM (Next Week)

4. **Implement WebRTC Client**
   - Study Milestone JavaScript sample
   - Implement SDP offer/answer
   - Handle ICE candidates
   - Display video stream

5. **Update Authentication**
   - Replace Basic Auth with OAuth 2.0
   - Update all API clients

6. **Implement Recording Control**
   - Use Events API or Bookmarks approach (whichever works)
   - Add local state tracking for recording status

### MEDIUM TERM (Next 2 Weeks)

7. **Implement Timeline**
   - Use Bookmarks search
   - Build client-side timeline visualization
   - Handle gaps

8. **Testing & Integration**
   - End-to-end testing
   - Performance optimization

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Events API doesn't start recording | Critical | Medium | Use Bookmarks newReference as backup |
| Bookmarks doesn't show all recordings | High | Medium | May need to accept limitation or find alternative |
| WebRTC too complex to implement | Critical | Low | Use Milestone GitHub samples as reference |
| Recording status unknown | Medium | High | Implement local state tracking |
| Timeline gaps incorrect | Medium | Medium | Use bookmark data + validation |

---

## Conclusion

### ❌ NOT All Required APIs Are Available

**Major Findings:**

1. **Authentication:** Different but works (OAuth 2.0 vs Basic Auth)
2. **Camera Discovery:** ✅ Works perfectly
3. **Manual Recording:** ❌ No direct endpoint - must use Events/Bookmarks
4. **Playback:** ❌ No HLS endpoint - must use WebRTC (major rework)
5. **Timeline:** ❌ No timeline endpoint - must build from bookmarks
6. **Recording Query:** ❌ No direct query - must use bookmarks (incomplete)

### Effort Estimate

- **Original Plan Estimate:** ~10 days
- **Revised Estimate:** ~20 days (2x more due to WebRTC + workarounds)

### Recommendation

**Proceed with implementation BUT:**

1. Accept that recording control will be Events/Bookmarks-based (not direct API)
2. Accept that playback requires WebRTC implementation (complex but doable)
3. Accept that timeline may only show bookmarked recordings (not all recordings)
4. Budget 2x more time than originally planned

**Alternative:** Consider if Milestone's native Smart Client web interface could be embedded instead of building custom integration.

---

**Status:** ⚠️ Significant gaps exist - Implementation possible but requires major adjustments

**Next Action:** Test Events API and Bookmarks newReference to confirm recording control approach works.
