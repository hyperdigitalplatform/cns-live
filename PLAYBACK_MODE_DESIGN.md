# Playback Mode Design Specification

## Overview

This document describes the playback mode implementation for the CCTV dashboard. Each grid cell can toggle between **LIVE** and **PLAYBACK** modes independently, allowing operators to review recorded footage while monitoring other live cameras.

---

## Design Principles

1. **Reuse Common Controls** - Fullscreen, camera name, and other controls work in both LIVE and PLAYBACK modes
2. **In-Cell Playback** - Playback happens within each grid cell, not in separate dialogs
3. **Minimal UI Clutter** - Controls are collapsible and auto-hide during playback
4. **No Audio** - CCTV cameras have unreliable/noisy audio, so all audio controls are removed
5. **Visual Timeline** - Sequences and gaps shown visually, not as separate lists
6. **Three Navigation Methods**:
   - **Slider** (fast, long-distance jumps)
   - **Scroll arrows** (controlled, systematic navigation)
   - **Timeline click** (precise, within visible window)

---

## Mode Toggle

### Location
Top of each grid cell, integrated with existing header

### States
- `[🔴 LIVE]` - Active when viewing live stream
- `[⏯️ PLAYBACK ✓]` - Active when viewing recorded footage

### Behavior
- Click to switch between modes
- "Jump to Live" button appears when in playback mode
- Mode is per-cell independent (Cell 1 can be LIVE while Cell 2 is PLAYBACK)

---

## Playback Control Bar

### Two States

#### **COLLAPSED (Minimal Display)**
```
┌───────────────────────────────────────────────────────────┐
│ [PLAYBACK] Oct 24, 2025 14:32:15 • ✓ Recording           │
│                                                           │
│ 00:00   06:00   12:00   18:00   24:00                   │
│ ━━━━━━━━╸╸━━━━━━╸╸━━━━━━╸╸╸╸╸╸╸╸╸╸╸                     │
│ ░░░░░░░████░░░████░░░░░░░░░░                             │
│                                                           │
│ 00:00 ├──────────────■──────────────────┤ 24:00          │
│       │░░░░░░███░░███░░░░░░░░░░░░░░░│                   │
│                                                           │
│                   [▲ Show]                                │
└───────────────────────────────────────────────────────────┘
```

**Shows:**
- PLAYBACK badge
- Current timestamp (clickable)
- Recording status (✓/✗)
- Time markers
- Thin timeline (recording bars only, no playhead)
- Navigation slider with handle (shows current position)
- Show button [▲]

#### **EXPANDED (Full Controls)**
```
┌───────────────────────────────────────────────────────────┐
│ [📅 Oct 24] [⏰14:32:15] [⏸️] [◀1hr] [1hr▶] [🔍±]        │
│                                                           │
│ Timeline (Detail):                                        │
│  12:00    13:00    14:00    15:00    16:00               │
│  ░░░░ ████████ ░░░ ████████ ░░░░                         │
│         ▲Seq 1       ▲Seq 2                              │
│                      🔴 14:32                             │
│                                                           │
│ Navigation Slider:                                        │
│ 00:00 ├──────────────■──────────────────┤ 24:00          │
│       │░░░░░░███░░███░░░░░░░░░░░░░░░│                   │
│                                                           │
│                     [▼ Hide]                              │
└───────────────────────────────────────────────────────────┘
```

**Shows:**
- Date selector (clickable → calendar)
- Time selector (clickable → time picker)
- Play/Pause button (toggles)
- Scroll left button [◀ 1hr]
- Scroll right button [1hr ▶]
- Zoom control [🔍±]
- Detailed timeline with playhead 🔴
- Sequence markers (▲)
- Navigation slider
- Hide button [▼]

---

## Control Components

### 1. Date Selector
**Display:** `[📅 Oct 24, 2025]`

**Click Action:** Opens calendar dialog
```
┌─ SELECT DATE ──────────┐
│  📅 Calendar:          │
│  ┌─────────────────┐   │
│  │ Oct 2025        │   │
│  │ Su Mo Tu We Th  │   │
│  │        1  2  3  │   │
│  │ 21 22 23 [24]  │   │
│  └─────────────────┘   │
│  Quick:                │
│  [Today] [Yesterday]   │
│        [Cancel] [OK]   │
└────────────────────────┘
```

### 2. Time Selector
**Display:** `[⏰ 14:32:15]`

**Click Action:** Opens time picker dialog
```
┌─ SELECT TIME ──────────┐
│  Time: [14]:[32]:[15]  │
│        HH   MM   SS    │
│                        │
│  Quick Jump:           │
│  [00:00] [06:00]      │
│  [12:00] [18:00]      │
│        [Cancel] [OK]   │
└────────────────────────┘
```

### 3. Play/Pause Button
**States:**
- Paused: `[▶️ Play]`
- Playing: `[⏸️ Pause]`

**Behavior:** Single button that toggles between states

### 4. Timeline Scroll Controls
**Buttons:**
- `[◀ 1hr]` - Shift timeline view 1 hour earlier
- `[1hr ▶]` - Shift timeline view 1 hour later

**Adaptive Scrolling:** Scroll amount adapts to zoom level
- 1 hour view → Scroll by 15 min
- 6 hour view → Scroll by 1 hour
- 24 hour view → Scroll by 6 hours
- 7 day view → Scroll by 1 day

### 5. Zoom Control
**Display:** `[🔍±]`

**Click Action:** Opens zoom dropdown
```
┌─ ZOOM LEVEL ────────┐
│ ○ 1 hour            │
│ ○ 4 hours           │
│ ● 12 hours (active) │
│ ○ 24 hours (1 day)  │
│ ○ 7 days            │
│ ○ 30 days           │
└─────────────────────┘
```

### 6. Timeline Component
**Visual:**
```
12:00    13:00    14:00    15:00    16:00
░░░░ ████████ ░░░ ████████ ░░░░
        ▲Seq 1       ▲Seq 2
                     🔴 14:32
```

**Legend:**
- `████` Green - Recording available
- `░░░░` Gray - Gap (no recording)
- `▲` Sequence start markers (clickable)
- `🔴` Playhead (current position) - **Only in EXPANDED state**

**Interactions:**
- Click anywhere → Seek to that time
- Hover → Show tooltip with timestamp and recording status
- Click sequence marker → Jump to sequence start
- Drag timeline → Pan left/right

### 7. Navigation Slider
**Purpose:** Quick navigation across full query range

**Visual:**
```
00:00 ├──────────────■──────────────────┤ 24:00
      │░░░░░░███░░███░░░░░░░░░░░░░░░│
      │      ▲  ▲         ■            │
      │      │  │         └─ Current position
      │      │  └─ Seq 2
      │      └─ Seq 1
      └────────────────────────────────┘
```

**Features:**
- Shows full query range (e.g., 24 hours)
- Mini recording bars overlaid on track
- Draggable handle (■) shows current position
- Click track → Jump to that time instantly
- See all sequences at a glance

**Interactions:**
- Drag handle → Jump to any time
- Click track → Jump to clicked position

### 8. Show/Hide Toggle
**Position:** Bottom center of control bar

**States:**
- Expanded: `[▼ Hide]` - Arrow points down
- Collapsed: `[▲ Show]` - Arrow points up

**Behavior:**
- Auto-collapse after 5 seconds of inactivity (during playback)
- Stays expanded when paused
- User can manually toggle anytime

---

## Navigation Methods

### Method 1: Slider (Fast, Long Distance)
**Use Case:** Jump from morning to evening quickly

**Example:** Drag slider handle from 08:00 to 20:00 instantly

### Method 2: Scroll Arrows (Controlled, Systematic)
**Use Case:** Move forward/backward hour by hour

**Example:** Click [1hr ▶] multiple times to advance systematically

### Method 3: Timeline Click (Precise, Within View)
**Use Case:** Exact time within visible window

**Example:** Click timeline at 14:32:15 precisely

---

## Timeline Behavior

### Visible Window vs Full Range

**Timeline (Detail View):** Shows 4-hour window
```
12:00     13:00     14:00     15:00     16:00
░░░░ ████████ ░░░ ████████ ░░░░
```

**Slider (Overview):** Shows full 24-hour range
```
00:00 ├─────────────■─────────────────┤ 24:00
      │░░░░░░░░░████░███░░░░░░░░░░░│
      │         └─┬─┘                │
      │     Visible window           │
      │     (12:00-16:00)            │
```

### Scrolling Example

**Current View:** 12:00 - 16:00
```
12:00     13:00     14:00     15:00     16:00
░░░░ ████████ ░░░ ████████ ░░░
```

**After clicking [◀ 1hr]:** 11:00 - 15:00
```
11:00     12:00     13:00     14:00     15:00
░░░░░ ████████ ░░░ ████
```

**After clicking [1hr ▶]:** 13:00 - 17:00
```
13:00     14:00     15:00     16:00     17:00
░░░ ████████ ░░░░░░░░░░░
```

---

## Color Coding

### Timeline Colors
- **Green (`#10B981`)** - Recording available
- **Gray (`#E5E7EB`)** - Gap (no recording)
- **Blue (`#3B82F6`)** - Playhead position
- **Blue (`#3B82F6`)** - Sequence markers

### Status Indicators
- **✓ Recording Available** - Green checkmark
- **✗ No Recording** - Red X

---

## Speed Control

### Available Speeds (Forward Only)
- 0.25x - Very slow
- 0.5x - Slow
- 1x - Normal (default)
- 2x - Fast
- 4x - Very fast
- 8x - Ultra fast

**Note:** Reverse playback removed (not logical for CCTV review)

---

## Audio Handling

**Decision:** All audio controls removed

**Rationale:**
- CCTV cameras have unreliable/noisy audio
- Not essential for video review
- Simplifies UI

**Removed:**
- Volume slider
- Mute/unmute button
- Audio status indicators

---

## Playback Flow

### 1. Switch to Playback Mode
```
User clicks [⏯️ PLAYBACK] in grid cell
↓
System queries recordings for last 24 hours
↓
Timeline shows available sequences and gaps
↓
Video paused at most recent recording
↓
Controls shown in EXPANDED state
```

### 2. Select Time Range
```
User clicks date [📅 Oct 24]
↓
Calendar dialog opens
↓
User selects date and clicks [OK]
↓
System queries recordings for selected date
↓
Timeline updates with new sequences
```

### 3. Navigate to Specific Time
```
User drags slider handle to 14:00
↓
Video seeks to 14:00:00
↓
Timeline view centers on 14:00
↓
Playback starts automatically (if recordings available)
```

### 4. Review Multiple Sequences
```
User clicks [▶️ Play] at Sequence 1
↓
Video plays forward from 08:00
↓
Reaches end of Sequence 1 at 09:15
↓
Gap encountered (no recording)
↓
User clicks [1hr ▶] to scroll timeline
↓
Sequence 2 comes into view at 14:00
↓
User clicks Sequence 2 marker (▲)
↓
Video jumps to 14:00 and plays
```

---

## API Integration

### Required Endpoints

#### 1. Query Recordings
```
GET /api/v1/playback/cameras/{cameraId}/sequences
Query Params:
  - startTime: ISO8601 datetime
  - endTime: ISO8601 datetime

Response:
{
  "sequences": [
    {
      "sequenceId": "seq-1",
      "startTime": "2025-10-24T08:00:00Z",
      "endTime": "2025-10-24T09:15:23Z",
      "durationSeconds": 4523,
      "available": true
    }
  ],
  "gaps": [
    {
      "startTime": "2025-10-24T09:15:23Z",
      "endTime": "2025-10-24T14:30:00Z",
      "durationSeconds": 18877
    }
  ],
  "coverage": 0.255
}
```

#### 2. Start Playback
```
POST /api/v1/playback/cameras/{cameraId}/start
Body:
{
  "timestamp": "2025-10-24T14:32:15Z",
  "speed": 1.0,
  "format": "hls"
}

Response:
{
  "playbackId": "pb-12345",
  "streamUrl": "/api/v1/playback/stream/pb-12345/playlist.m3u8"
}
```

#### 3. Control Playback
```
POST /api/v1/playback/{playbackId}/seek
Body:
{
  "timestamp": "2025-10-24T15:00:00Z"
}

POST /api/v1/playback/{playbackId}/speed
Body:
{
  "speed": 2.0
}
```

---

## Component Structure

```
GridCell
├── CellHeader
│   ├── CameraName
│   ├── ModeToggle (LIVE/PLAYBACK)
│   ├── JumpToLiveButton (if playback)
│   └── FullscreenButton
│
├── VideoArea
│   ├── LiveStreamPlayer (if mode === 'live')
│   └── PlaybackPlayer (if mode === 'playback')
│       ├── HLS Video Player
│       └── PlaybackOverlay
│
└── PlaybackControlBar (if mode === 'playback')
    ├── CollapsedView
    │   ├── StatusBadge
    │   ├── Timestamp (clickable)
    │   ├── RecordingStatus
    │   ├── ThinTimeline
    │   ├── NavigationSlider
    │   └── ShowButton [▲]
    │
    └── ExpandedView
        ├── ControlButtons
        │   ├── DatePicker
        │   ├── TimePicker
        │   ├── PlayPauseButton
        │   ├── ScrollLeftButton
        │   ├── ScrollRightButton
        │   └── ZoomControl
        ├── DetailedTimeline
        │   ├── TimeMarkers
        │   ├── RecordingBars
        │   ├── SequenceMarkers
        │   └── Playhead
        ├── NavigationSlider
        └── HideButton [▼]
```

---

## File Structure

```
dashboard/src/
├── components/
│   ├── playback/
│   │   ├── PlaybackModeToggle.tsx
│   │   ├── PlaybackControlBar.tsx
│   │   ├── PlaybackTimeline.tsx
│   │   ├── NavigationSlider.tsx
│   │   ├── TimePickerDialog.tsx
│   │   ├── DatePickerDialog.tsx
│   │   └── PlaybackPlayer.tsx
│   │
│   ├── StreamGridEnhanced.tsx (updated)
│   └── LiveStreamPlayer.tsx
│
├── hooks/
│   ├── usePlayback.ts
│   └── useTimelineNavigation.ts
│
├── types/
│   └── playback.ts
│
└── services/
    └── playbackApi.ts
```

---

## Summary of Key Features

✅ **Per-cell independent playback** - Each grid cell operates independently
✅ **Collapsible controls** - Minimal UI when collapsed, full controls when expanded
✅ **Three navigation methods** - Slider, scroll arrows, timeline click
✅ **No audio controls** - Simplified for CCTV use case
✅ **Visual timeline** - Sequences and gaps shown graphically
✅ **No redundant playhead** - Collapsed state shows position only on slider
✅ **Reused common controls** - Fullscreen and other controls work in both modes
✅ **Auto-collapse** - Controls hide automatically during playback
✅ **Touch-friendly** - Works on tablets and touch screens

---

## Future Enhancements (Not in MVP)

- [ ] Keyboard shortcuts (Space, arrows, etc.)
- [ ] Frame-by-frame stepping
- [ ] Time selection brackets (IN/OUT markers)
- [ ] Export with custom range
- [ ] Snapshot capture
- [ ] Multiple sequence download
- [ ] Motion search within recordings
- [ ] Bookmarks on timeline
- [ ] Multi-camera synchronized playback

---

## References

- Milestone XProtect Web Client Documentation
- Original design discussions
- Existing RecordingPlayer component (dashboard/src/components/RecordingPlayer.tsx)
- Existing RecordingTimeline component (dashboard/src/components/RecordingTimeline.tsx)

---

**Document Version:** 1.0
**Last Updated:** 2025-10-29
**Author:** Design discussion with stakeholders
