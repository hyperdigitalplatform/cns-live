# Dashboard Playback Feature - Implementation Plan

**Date:** October 31, 2025
**Status:** Playback Foundation Complete - Enhancement Phase
**Author:** Claude AI Assistant

---

## Executive Summary

The dashboard **already has working playback functionality** using WebRTC from Milestone VMS. This plan focuses on **completing polish, adding advanced features, and ensuring production readiness**.

**Current Status:**
- ✅ **Phase 1:** Backend WebRTC API - COMPLETE
- ✅ **Phase 2:** Frontend WebRTC Player - COMPLETE
- 🟡 **Phase 3:** Timeline Integration - PARTIALLY COMPLETE (80%)
- 🔴 **Phase 4:** Testing & Polish - NOT STARTED (0%)

---

## 1. CURRENT PLAYBACK CAPABILITIES

### ✅ What's Already Working

#### **A. Core Playback**
- WebRTC-based playback from Milestone VMS
- Per-cell playback mode (independent of other cells)
- Play/pause/seek controls
- Timeline with recording sequences
- Automatic sequence progression
- Connection state monitoring
- Performance statistics (bandwidth, packet loss, jitter)

**Location:** `dashboard/src/hooks/useWebRTCPlayback.ts` (Lines 168-408)

#### **B. UI Components**
- `PlaybackModeToggle` - Switch between LIVE/PLAYBACK per cell
- `PlaybackControlBar` - Collapsed/Expanded timeline controls
- `RecordingPlayer` - Video player with WebRTC integration
- `NavigationSlider` - Timeline scrubbing
- `TimePickerDialog` - Date/time selection

**Location:** `dashboard/src/components/playback/`

#### **C. Integration**
- Integrated into `StreamGridEnhanced` (multi-cell grid)
- Camera sidebar with recording section
- Milestone sequence querying
- Timeline visualization with gaps
- Seek with debouncing

**Location:** `dashboard/src/components/StreamGridEnhanced.tsx` (Lines 305-857)

---

## 2. IMPLEMENTATION PHASES

### Phase 4: Testing & Polish (HIGH PRIORITY) ⏱️ 2-3 days

**Goal:** Make current playback features production-ready and bug-free.

#### **Task 4.1: Integration Testing** ⏱️ 1 day

**Scenarios to Test:**

1. **Single Cell Playback**
   - [ ] Start playback from sidebar
   - [ ] Play/pause functionality
   - [ ] Seek forward/backward
   - [ ] Timeline scrubbing
   - [ ] Zoom level changes
   - [ ] Switch back to live mode

2. **Multi-Cell Playback**
   - [ ] Play different cameras in different cells simultaneously
   - [ ] Mix live and playback modes in same grid
   - [ ] Verify independent playback state per cell
   - [ ] No interference between cells

3. **Edge Cases**
   - [ ] No recordings available (show graceful message)
   - [ ] Network interruption during playback
   - [ ] Seek to non-existent time
   - [ ] Rapid mode switching
   - [ ] Connection timeout handling

4. **Memory Leaks**
   - [ ] Start/stop playback 50+ times
   - [ ] Monitor memory usage in DevTools
   - [ ] Verify proper cleanup on unmount
   - [ ] Check ICE candidate polling stops

5. **Performance**
   - [ ] Measure time to first frame (< 3 seconds)
   - [ ] Verify smooth playback (no stuttering)
   - [ ] Check CPU usage (< 15% per stream)
   - [ ] Monitor bandwidth consumption

**Implementation:**
- Create test cases in `dashboard/tests/playback.test.tsx`
- Use React Testing Library + Vitest
- Add Cypress E2E tests for critical paths

**Files to Create:**
```
dashboard/tests/
  playback.test.tsx           # Unit tests
  integration/
    playback-flow.cy.tsx      # E2E tests
```

#### **Task 4.2: Error Handling Enhancement** ⏱️ 1 day

**Current Gaps:**
- No retry logic for failed connections
- Limited user feedback on errors
- No fallback mechanisms

**Improvements:**

1. **Add Retry Logic** (useWebRTCPlayback.ts)
```typescript
// Add to useWebRTCPlayback.ts around line 193
const startPlaybackWithRetry = async (maxRetries = 3) => {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      await startPlayback();
      return;
    } catch (error) {
      if (attempt === maxRetries) throw error;
      await new Promise(resolve => setTimeout(resolve, 2000 * attempt));
    }
  }
};
```

2. **User-Friendly Error Messages**
```typescript
const ERROR_MESSAGES = {
  'NO_RECORDINGS': 'No recordings found for this camera at the selected time.',
  'CONNECTION_FAILED': 'Failed to connect to playback server. Please try again.',
  'NETWORK_ERROR': 'Network error. Please check your connection.',
  'TIMEOUT': 'Connection timed out. The recording server may be busy.',
  'UNSUPPORTED_CAMERA': 'This camera does not support playback.',
};
```

3. **Fallback Mechanisms**
- If WebRTC fails, show option to use test-webrtc-playback.html
- Provide "Report Issue" button
- Auto-retry with exponential backoff

**Files to Modify:**
- `dashboard/src/hooks/useWebRTCPlayback.ts` (Add retry logic)
- `dashboard/src/components/RecordingPlayer.tsx` (Better error UI)
- `dashboard/src/components/StreamGridEnhanced.tsx` (Fallback handling)

#### **Task 4.3: Performance Optimization** ⏱️ 0.5 days

**Optimizations:**

1. **Connection Pooling**
```typescript
// Reuse ICE servers across connections
const ICE_SERVER_POOL = {
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' },
  ],
};
```

2. **Lazy Timeline Data Loading**
```typescript
// Only load timeline when expanded
const [timelineExpanded, setTimelineExpanded] = useState(false);

useEffect(() => {
  if (timelineExpanded && !timelineData) {
    fetchTimelineData();
  }
}, [timelineExpanded]);
```

3. **Debounce ICE Polling**
- Current: Fixed 1-second polling
- Improved: Exponential backoff (1s, 2s, 4s, 8s)

**Files to Modify:**
- `dashboard/src/hooks/useWebRTCPlayback.ts` (Lines 286-326)
- `dashboard/src/components/playback/PlaybackControlBar.tsx` (Lazy loading)

#### **Task 4.4: Documentation** ⏱️ 0.5 days

**Create:**
- `docs/PLAYBACK_USER_GUIDE.md` - End-user instructions
- `docs/PLAYBACK_DEVELOPER_GUIDE.md` - Technical documentation
- Inline JSDoc comments for key functions

---

### Phase 5: Speed Control (MEDIUM PRIORITY) ⏱️ 1 day

**Goal:** Add playback speed control (0.25x, 0.5x, 1x, 2x, 4x)

**Current Status:**
- ✅ Backend supports speed parameter (milestone-service)
- ❌ Frontend UI not implemented
- ❌ Hook doesn't accept speed changes

#### **Task 5.1: Update WebRTC Hook** ⏱️ 2 hours

**File:** `dashboard/src/hooks/useWebRTCPlayback.ts`

**Changes:**

1. Add speed to hook interface (line 25):
```typescript
export const useWebRTCPlayback = (cameraId: string | null, speed: number = 1.0) => {
  // ...
}
```

2. Include speed in API request (line 200):
```typescript
const response = await fetch(`/api/v1/cameras/${cameraId}/playback/start`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    startTime,
    endTime,
    speed, // Add this
  }),
});
```

3. Add speed change function (after line 407):
```typescript
const changeSpeed = useCallback(async (newSpeed: number) => {
  if (!sessionIdRef.current) return;

  try {
    await fetch(`/api/v1/playback/speed/${sessionIdRef.current}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ speed: newSpeed }),
    });
  } catch (error) {
    console.error('Failed to change speed:', error);
  }
}, []);
```

#### **Task 5.2: Add Speed Control UI** ⏱️ 3 hours

**File:** `dashboard/src/components/playback/PlaybackControlBar.tsx`

**Design:**
```
[0.25x] [0.5x] [1x] [2x] [4x]  ← Speed selector buttons
```

**Implementation:**

1. Add speed state (line 30):
```typescript
const [playbackSpeed, setPlaybackSpeed] = useState<number>(1.0);
```

2. Add speed selector UI (after play/pause button, ~line 248):
```tsx
<div className="flex items-center gap-1 px-2 border-l border-gray-700">
  {[0.25, 0.5, 1, 2, 4].map((speed) => (
    <button
      key={speed}
      onClick={() => {
        setPlaybackSpeed(speed);
        onSpeedChange?.(speed);
      }}
      className={`px-2 py-1 text-xs rounded transition-colors ${
        playbackSpeed === speed
          ? 'bg-blue-500 text-white'
          : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
      }`}
    >
      {speed}x
    </button>
  ))}
</div>
```

3. Add callback prop (line 20):
```typescript
interface PlaybackControlBarProps {
  // ... existing props
  onSpeedChange?: (speed: number) => void;
}
```

#### **Task 5.3: Integrate with Grid** ⏱️ 1 hour

**File:** `dashboard/src/components/StreamGridEnhanced.tsx`

**Changes:**

1. Add speed to playback state (line 88):
```typescript
interface GridCell {
  // ...
  playbackState?: {
    // ... existing fields
    speed: number;  // Add this
  }
}
```

2. Pass speed to hook (line 772):
```typescript
const {
  videoRef,
  connect,
  disconnect,
  connectionState,
  error,
  stats,
} = useWebRTCPlayback(
  cell.camera?.id || null,
  cell.playbackState?.speed || 1.0  // Add this
);
```

3. Handle speed change callback (line 850):
```typescript
<PlaybackControlBar
  // ... existing props
  onSpeedChange={(speed) => {
    updateCellPlaybackState(index, { speed });
  }}
/>
```

**Files to Modify:**
- `dashboard/src/hooks/useWebRTCPlayback.ts`
- `dashboard/src/components/playback/PlaybackControlBar.tsx`
- `dashboard/src/components/StreamGridEnhanced.tsx`
- `dashboard/src/types/playback.ts` (add speed to PlaybackState)

---

### Phase 6: Export & Download (MEDIUM PRIORITY) ⏱️ 1.5 days

**Goal:** Allow users to export/download recording segments

**Current Status:**
- ✅ Backend endpoint exists: `POST /api/v1/playback/export`
- ❌ Frontend not implemented

#### **Task 6.1: Export API Integration** ⏱️ 2 hours

**File:** `dashboard/src/services/api.ts`

**Changes:**

1. Update export endpoint (line 147):
```typescript
export async function exportRecording(
  cameraId: string,
  startTime: string,
  endTime: string,
  format: 'mp4' | 'avi' = 'mp4'
): Promise<{ exportId: string; downloadUrl: string }> {
  const response = await apiRequest<{ exportId: string; downloadUrl: string }>(
    `/api/v1/playback/export`,
    {
      method: 'POST',
      body: JSON.stringify({ cameraId, startTime, endTime, format }),
    }
  );
  return response;
}
```

#### **Task 6.2: Export UI Component** ⏱️ 4 hours

**Create:** `dashboard/src/components/playback/ExportDialog.tsx`

**Features:**
- Time range selection (or use current playback range)
- Format selection (MP4, AVI)
- Progress indicator
- Download button when ready

**Design:**
```
┌─────────────────────────────────────┐
│ Export Recording                  × │
├─────────────────────────────────────┤
│ Camera: Camera 1                    │
│ Start: 2025-10-31 10:00:00         │
│ End:   2025-10-31 10:30:00         │
│                                     │
│ Format: [MP4 ▼]                    │
│                                     │
│ [■■■■■■■□□□] 75% - Exporting...    │
│                                     │
│ [Cancel] [Download (when ready)]   │
└─────────────────────────────────────┘
```

**Implementation:**
```typescript
export const ExportDialog: React.FC<ExportDialogProps> = ({
  camera,
  startTime,
  endTime,
  onClose,
}) => {
  const [format, setFormat] = useState<'mp4' | 'avi'>('mp4');
  const [exporting, setExporting] = useState(false);
  const [progress, setProgress] = useState(0);
  const [downloadUrl, setDownloadUrl] = useState<string | null>(null);

  const handleExport = async () => {
    setExporting(true);
    try {
      const result = await exportRecording(
        camera.id,
        startTime,
        endTime,
        format
      );

      // Poll for progress
      const interval = setInterval(async () => {
        const status = await fetch(`/api/v1/playback/export/${result.exportId}/status`);
        const data = await status.json();
        setProgress(data.progress);

        if (data.status === 'complete') {
          clearInterval(interval);
          setDownloadUrl(data.downloadUrl);
          setExporting(false);
        }
      }, 1000);
    } catch (error) {
      console.error('Export failed:', error);
      setExporting(false);
    }
  };

  return (
    <Dialog open onClose={onClose}>
      {/* UI implementation */}
    </Dialog>
  );
};
```

#### **Task 6.3: Add Export Button** ⏱️ 2 hours

**File:** `dashboard/src/components/playback/PlaybackControlBar.tsx`

**Changes:**

1. Add export button (after zoom selector, ~line 282):
```tsx
<button
  onClick={() => setShowExportDialog(true)}
  className="p-2 hover:bg-gray-700 rounded transition-colors"
  title="Export Recording"
>
  <Download className="w-4 h-4" />
</button>
```

2. Add export dialog (end of component):
```tsx
{showExportDialog && (
  <ExportDialog
    camera={camera}
    startTime={startTime}
    endTime={endTime}
    onClose={() => setShowExportDialog(false)}
  />
)}
```

**Files to Create:**
- `dashboard/src/components/playback/ExportDialog.tsx`

**Files to Modify:**
- `dashboard/src/services/api.ts`
- `dashboard/src/components/playback/PlaybackControlBar.tsx`

---

### Phase 7: Bookmarks (LOW PRIORITY) ⏱️ 2 days

**Goal:** Add bookmarks to timeline for quick navigation

#### **Task 7.1: Backend API** ⏱️ 4 hours

**Create Endpoints:**
```
POST   /api/v1/bookmarks
GET    /api/v1/bookmarks?cameraId=xxx
PUT    /api/v1/bookmarks/:id
DELETE /api/v1/bookmarks/:id
```

**Database Schema:**
```sql
CREATE TABLE bookmarks (
  id UUID PRIMARY KEY,
  camera_id UUID REFERENCES cameras(id),
  user_id UUID REFERENCES users(id),
  timestamp TIMESTAMPTZ NOT NULL,
  label VARCHAR(255),
  color VARCHAR(7) DEFAULT '#FFD700',
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### **Task 7.2: Frontend UI** ⏱️ 6 hours

**Features:**
- Add bookmark button on PlaybackControlBar
- Render bookmarks as markers on timeline
- Click bookmark to jump to that time
- Right-click to edit/delete bookmark
- Color coding for different types

**Implementation:**

1. **Bookmark API Client** (`dashboard/src/services/api.ts`):
```typescript
export async function createBookmark(
  cameraId: string,
  timestamp: string,
  label: string,
  color: string = '#FFD700'
): Promise<Bookmark> {
  return apiRequest<Bookmark>(`/api/v1/bookmarks`, {
    method: 'POST',
    body: JSON.stringify({ cameraId, timestamp, label, color }),
  });
}
```

2. **Bookmark Component** (`dashboard/src/components/playback/BookmarkMarker.tsx`):
```tsx
export const BookmarkMarker: React.FC<BookmarkMarkerProps> = ({
  bookmark,
  position,
  onClick,
  onEdit,
  onDelete,
}) => {
  return (
    <div
      className="absolute top-0 bottom-0 w-1 cursor-pointer"
      style={{
        left: `${position}%`,
        backgroundColor: bookmark.color,
      }}
      onClick={() => onClick(bookmark)}
      onContextMenu={(e) => {
        e.preventDefault();
        // Show context menu
      }}
      title={bookmark.label}
    />
  );
};
```

3. **Integration in PlaybackControlBar**:
```tsx
{bookmarks.map((bookmark) => {
  const position = calculatePosition(bookmark.timestamp);
  return (
    <BookmarkMarker
      key={bookmark.id}
      bookmark={bookmark}
      position={position}
      onClick={handleBookmarkClick}
      onEdit={handleBookmarkEdit}
      onDelete={handleBookmarkDelete}
    />
  );
})}
```

**Files to Create:**
- `dashboard/src/components/playback/BookmarkMarker.tsx`
- `dashboard/src/components/playback/BookmarkDialog.tsx`
- `dashboard/src/types/bookmark.ts`

**Files to Modify:**
- `dashboard/src/services/api.ts`
- `dashboard/src/components/playback/PlaybackControlBar.tsx`

---

### Phase 8: Multi-Camera Sync (LOW PRIORITY) ⏱️ 2.5 days

**Goal:** Synchronize playback across multiple cells

#### **Task 8.1: Sync State Management** ⏱️ 4 hours

**Create:** `dashboard/src/stores/playbackSyncStore.ts`

```typescript
interface PlaybackSyncState {
  syncEnabled: boolean;
  syncedCells: Set<number>;
  masterCellIndex: number | null;
  syncTime: Date | null;

  enableSync: (cellIndices: number[]) => void;
  disableSync: () => void;
  updateSyncTime: (time: Date) => void;
  setMasterCell: (index: number) => void;
}

export const usePlaybackSyncStore = create<PlaybackSyncState>((set, get) => ({
  syncEnabled: false,
  syncedCells: new Set(),
  masterCellIndex: null,
  syncTime: null,

  enableSync: (cellIndices) => {
    set({
      syncEnabled: true,
      syncedCells: new Set(cellIndices),
      masterCellIndex: cellIndices[0],
    });
  },

  disableSync: () => {
    set({
      syncEnabled: false,
      syncedCells: new Set(),
      masterCellIndex: null,
      syncTime: null,
    });
  },

  updateSyncTime: (time) => {
    set({ syncTime: time });
  },

  setMasterCell: (index) => {
    set({ masterCellIndex: index });
  },
}));
```

#### **Task 8.2: Sync UI Controls** ⏱️ 4 hours

**Create:** `dashboard/src/components/playback/PlaybackSyncPanel.tsx`

**Features:**
- Select cells to sync
- Visual indicator of synced cells
- Master cell selection
- Sync timeline
- Global play/pause

**Design:**
```
┌──────────────────────────────────┐
│ Multi-Camera Sync               │
├──────────────────────────────────┤
│ Synced Cameras:                 │
│ ☑ Cell 1 - Camera A (Master)   │
│ ☑ Cell 2 - Camera B             │
│ ☑ Cell 3 - Camera C             │
│ ☐ Cell 4 - Camera D             │
│                                  │
│ [▶ Play All] [⏸ Pause All]     │
│                                  │
│ Timeline: [■■■■■■□□□□]          │
│ Time: 10:30:45                  │
└──────────────────────────────────┘
```

#### **Task 8.3: Sync Logic Implementation** ⏱️ 8 hours

**File:** `dashboard/src/components/StreamGridEnhanced.tsx`

**Changes:**

1. Watch for sync state changes:
```typescript
const syncState = usePlaybackSyncStore();

useEffect(() => {
  if (!syncState.syncEnabled) return;

  // When master cell time changes, sync all cells
  const masterCell = cells[syncState.masterCellIndex!];
  if (!masterCell?.playbackState) return;

  const masterTime = masterCell.playbackState.currentTime;

  // Update all synced cells
  syncState.syncedCells.forEach((cellIndex) => {
    if (cellIndex !== syncState.masterCellIndex) {
      updateCellPlaybackState(cellIndex, {
        currentTime: masterTime,
      });
      // Trigger seek on that cell's player
      seekToTime(cellIndex, masterTime);
    }
  });
}, [syncState.syncTime]);
```

2. Handle global play/pause:
```typescript
const handleGlobalPlay = () => {
  syncState.syncedCells.forEach((cellIndex) => {
    const cell = cells[cellIndex];
    if (cell?.playbackState) {
      updateCellPlaybackState(cellIndex, { isPlaying: true });
    }
  });
};

const handleGlobalPause = () => {
  syncState.syncedCells.forEach((cellIndex) => {
    const cell = cells[cellIndex];
    if (cell?.playbackState) {
      updateCellPlaybackState(cellIndex, { isPlaying: false });
    }
  });
};
```

3. Visual sync indicator:
```tsx
{syncState.syncEnabled && syncState.syncedCells.has(index) && (
  <div className="absolute top-2 right-2 z-20">
    <div className="bg-blue-500 text-white px-2 py-1 rounded text-xs flex items-center gap-1">
      <Link className="w-3 h-3" />
      {syncState.masterCellIndex === index ? 'Master' : 'Synced'}
    </div>
  </div>
)}
```

**Files to Create:**
- `dashboard/src/stores/playbackSyncStore.ts`
- `dashboard/src/components/playback/PlaybackSyncPanel.tsx`

**Files to Modify:**
- `dashboard/src/components/StreamGridEnhanced.tsx`
- `dashboard/src/pages/LiveViewEnhanced.tsx` (add sync panel)

---

### Phase 9: Advanced Timeline Features (LOW PRIORITY) ⏱️ 3 days

**Goal:** Add advanced visualization and interactions

#### **Task 9.1: Motion Detection Overlay** ⏱️ 1 day

**Features:**
- Query motion events from Milestone
- Render as heatmap on timeline
- Color intensity based on motion level
- Click to jump to motion event

**Implementation:**

1. **API Endpoint** (backend):
```
GET /api/v1/milestone/motion?cameraId=xxx&startTime=xxx&endTime=xxx
```

Response:
```json
{
  "events": [
    {
      "timestamp": "2025-10-31T10:15:23Z",
      "intensity": 0.85,
      "duration": 12
    }
  ]
}
```

2. **Frontend Rendering** (`PlaybackControlBar.tsx`):
```tsx
<div className="absolute inset-0 pointer-events-none">
  {motionEvents.map((event) => {
    const position = calculatePosition(event.timestamp);
    const width = (event.duration / totalDuration) * 100;
    return (
      <div
        key={event.timestamp}
        className="absolute h-full"
        style={{
          left: `${position}%`,
          width: `${width}%`,
          backgroundColor: `rgba(255, 0, 0, ${event.intensity * 0.3})`,
        }}
      />
    );
  })}
</div>
```

#### **Task 9.2: Event Markers** ⏱️ 1 day

**Features:**
- Alarm events from VMS
- Analytics events (line crossing, loitering, etc.)
- Custom user events
- Click to see event details

**Types of Events:**
- 🚨 Alarms (red)
- 🎯 Analytics (yellow)
- 👤 User Events (blue)
- 📹 Recording Start/Stop (gray)

#### **Task 9.3: Thumbnail Previews** ⏱️ 1 day

**Features:**
- Hover over timeline to see thumbnail preview
- Shows frame at that timestamp
- Tooltip with exact time
- Smooth thumbnail loading

**Implementation:**

1. **Thumbnail API** (backend):
```
GET /api/v1/playback/thumbnail?cameraId=xxx&timestamp=xxx&width=160&height=90
```

2. **Frontend** (`PlaybackControlBar.tsx`):
```tsx
const [hoverTime, setHoverTime] = useState<Date | null>(null);
const [thumbnailUrl, setThumbnailUrl] = useState<string | null>(null);

useEffect(() => {
  if (!hoverTime) return;

  const loadThumbnail = async () => {
    const url = `/api/v1/playback/thumbnail?cameraId=${camera.id}&timestamp=${hoverTime.toISOString()}&width=160&height=90`;
    setThumbnailUrl(url);
  };

  const timeout = setTimeout(loadThumbnail, 300); // Debounce
  return () => clearTimeout(timeout);
}, [hoverTime]);

return (
  <div
    onMouseMove={(e) => {
      const rect = e.currentTarget.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const percent = x / rect.width;
      const time = new Date(startTime.getTime() + percent * totalDuration);
      setHoverTime(time);
    }}
    onMouseLeave={() => setHoverTime(null)}
  >
    {/* Timeline */}

    {hoverTime && thumbnailUrl && (
      <div
        className="absolute bottom-full mb-2 pointer-events-none"
        style={{ left: `${calculatePosition(hoverTime)}%` }}
      >
        <img
          src={thumbnailUrl}
          alt="Preview"
          className="w-40 h-auto rounded shadow-lg border-2 border-white"
        />
        <div className="text-xs text-center mt-1 text-white bg-black bg-opacity-75 px-2 py-1 rounded">
          {formatTime(hoverTime)}
        </div>
      </div>
    )}
  </div>
);
```

---

## 3. PRIORITY ROADMAP

### Immediate (This Week)
1. ✅ **Complete Phase 4:** Testing & Polish
   - Integration testing
   - Error handling
   - Performance optimization
   - Documentation

### Short Term (Next 2 Weeks)
2. **Phase 5:** Speed Control
3. **Phase 6:** Export & Download

### Medium Term (Next Month)
4. **Phase 7:** Bookmarks
5. **Phase 8:** Multi-Camera Sync

### Long Term (2-3 Months)
6. **Phase 9:** Advanced Timeline Features
7. Additional features based on user feedback

---

## 4. ESTIMATED EFFORT

| Phase | Tasks | Estimated Time | Priority |
|-------|-------|----------------|----------|
| **Phase 4: Testing & Polish** | 4 | 2-3 days | 🔴 HIGH |
| **Phase 5: Speed Control** | 3 | 1 day | 🟡 MEDIUM |
| **Phase 6: Export & Download** | 3 | 1.5 days | 🟡 MEDIUM |
| **Phase 7: Bookmarks** | 2 | 2 days | 🟢 LOW |
| **Phase 8: Multi-Camera Sync** | 3 | 2.5 days | 🟢 LOW |
| **Phase 9: Advanced Features** | 3 | 3 days | 🟢 LOW |
| **TOTAL** | 18 tasks | **12.5 days** | |

---

## 5. TECHNICAL DECISIONS

### Architecture Choices

1. **Per-Cell Playback State**
   - ✅ **Decision:** Keep playback state local to each cell
   - **Rationale:** Allows mixed live/playback grids, independent controls
   - **Trade-off:** More complex than global state, but more flexible

2. **WebRTC vs HLS**
   - ✅ **Decision:** Use WebRTC for playback (already implemented)
   - **Rationale:** Lower latency, better quality, Milestone native support
   - **Trade-off:** More complex than HLS, but better performance

3. **Timeline Rendering**
   - ✅ **Decision:** Canvas-based timeline with CSS overlays
   - **Rationale:** Better performance for large recordings
   - **Trade-off:** More complex than pure DOM, but scalable

4. **State Management**
   - ✅ **Decision:** Zustand for global state, local React state for UI
   - **Rationale:** Lightweight, easy to use, good TypeScript support
   - **Trade-off:** Not as powerful as Redux, but simpler

### Technology Stack

- **Frontend:** React 18 + TypeScript + Vite
- **State:** Zustand stores
- **Styling:** Tailwind CSS
- **WebRTC:** Native RTCPeerConnection
- **Testing:** Vitest + React Testing Library + Cypress
- **Backend:** Go (milestone-service) + Node.js (API gateway)

---

## 6. RISKS & MITIGATION

### Risk 1: Performance with Multiple Playback Streams
**Impact:** High
**Probability:** Medium
**Mitigation:**
- Limit concurrent playback streams (max 4)
- Add warning when approaching limit
- Optimize WebRTC connection pooling

### Risk 2: Milestone VMS Availability
**Impact:** High
**Probability:** Low
**Mitigation:**
- Add connection health checks
- Implement retry logic with exponential backoff
- Graceful degradation (show cached data)

### Risk 3: Browser Compatibility
**Impact:** Medium
**Probability:** Low
**Mitigation:**
- Test on Chrome, Firefox, Edge, Safari
- Add browser detection and warnings
- Provide fallback for unsupported browsers

### Risk 4: Network Bandwidth
**Impact:** Medium
**Probability:** Medium
**Mitigation:**
- Add quality selection (SD, HD)
- Monitor bandwidth usage
- Warn users when bandwidth is insufficient

---

## 7. SUCCESS METRICS

### Phase 4 (Testing & Polish)
- ✅ All integration tests pass (100%)
- ✅ Zero memory leaks in 50+ playback cycles
- ✅ Time to first frame < 3 seconds
- ✅ CPU usage < 15% per stream
- ✅ User-reported bugs < 5 in first week

### Phase 5 (Speed Control)
- ✅ Speed changes within 500ms
- ✅ Smooth playback at all speeds
- ✅ UI responds correctly
- ✅ Backend handles speed changes

### Phase 6 (Export & Download)
- ✅ Export completes within 2x recording duration
- ✅ Downloaded files play correctly
- ✅ Progress indicator accurate
- ✅ Error handling for failed exports

### Phase 7+ (Advanced Features)
- ✅ Features work as designed
- ✅ User satisfaction > 80%
- ✅ Performance remains acceptable

---

## 8. CONCLUSION

The dashboard **already has a solid playback foundation** (Phases 1-2 complete, Phase 3 at 80%). The focus should be on:

1. **Immediate:** Complete testing and polish (Phase 4) to ensure production readiness
2. **Short-term:** Add user-requested features (speed control, export)
3. **Long-term:** Enhance with advanced features (bookmarks, sync, motion detection)

**Recommended Next Steps:**
1. Start Phase 4 immediately (2-3 days)
2. Get user feedback on current playback
3. Prioritize Phase 5 and 6 based on user needs
4. Plan Phase 7+ for future iterations

---

## 9. APPENDIX: KEY FILE REFERENCE

### Critical Files for Playback Development

| File | Lines | Purpose |
|------|-------|---------|
| `StreamGridEnhanced.tsx` | 305-381 | Mode switching |
| | 393-428 | Seek handling |
| | 753-857 | Playback rendering |
| `useWebRTCPlayback.ts` | 168-408 | WebRTC flow |
| | 84-119 | Statistics |
| `RecordingPlayer.tsx` | 42-99 | Player controls |
| `PlaybackControlBar.tsx` | 216-457 | Timeline controls |
| `api.ts` | 221-243 | Milestone API |
| `streamStore.ts` | 32-111 | Stream management |

### New Files to Create

- `dashboard/tests/playback.test.tsx`
- `dashboard/tests/integration/playback-flow.cy.tsx`
- `dashboard/src/components/playback/ExportDialog.tsx`
- `dashboard/src/components/playback/BookmarkMarker.tsx`
- `dashboard/src/components/playback/BookmarkDialog.tsx`
- `dashboard/src/components/playback/PlaybackSyncPanel.tsx`
- `dashboard/src/stores/playbackSyncStore.ts`
- `dashboard/src/types/bookmark.ts`
- `docs/PLAYBACK_USER_GUIDE.md`
- `docs/PLAYBACK_DEVELOPER_GUIDE.md`

---

**Document Version:** 1.0
**Last Updated:** October 31, 2025
**Status:** Ready for Implementation
