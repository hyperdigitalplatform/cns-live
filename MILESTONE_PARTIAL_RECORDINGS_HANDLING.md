# 🎬 Milestone XProtect - Handling Partial Recordings & Gaps

## 📋 Overview

This document explains how Milestone XProtect VMS handles scenarios where recordings are **partial**, **incomplete**, or have **gaps** in the timeline - and how the RTA CCTV System should handle these cases when integrating with Milestone.

---

## ❓ The Problem: User Requests Full Day, Only 2 Hours Available

### Example Scenario:
```
User Request:
  Camera: "Main Entrance"
  Start Time: 2025-10-24T00:00:00Z (midnight)
  End Time: 2025-10-24T23:59:59Z (end of day)
  Duration Requested: 24 hours

Actual Recording Availability:
  Sequence 1: 2025-10-24T08:00:00Z → 2025-10-24T09:00:00Z (1 hour)
  Sequence 2: 2025-10-24T14:00:00Z → 2025-10-24T15:00:00Z (1 hour)
  Total Available: 2 hours out of 24 hours requested
```

**Question**: How does the Milestone API respond to this request?

---

## 🔍 How Milestone Handles This

### 1. Timeline/Sequences API Response

Milestone provides a **separate API** to query recording availability **before** attempting playback.

**API Endpoint**: `/api/rest/v1/recordings/{cameraId}/timeline`

**Example Request:**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/cam-main-entrance/timeline?startTime=2025-10-24T00:00:00Z&endTime=2025-10-24T23:59:59Z"
```

**Response (Showing Gaps):**
```json
{
  "cameraId": "cam-main-entrance",
  "cameraName": "Main Entrance Camera",
  "requestedRange": {
    "startTime": "2025-10-24T00:00:00Z",
    "endTime": "2025-10-24T23:59:59Z",
    "durationSeconds": 86400
  },
  "availableDuration": 7200,  // Only 2 hours available!
  "coveragePercent": 8.33,     // Only 8.33% of requested time
  "sequences": [
    {
      "sequenceId": "seq-001",
      "startTime": "2025-10-24T08:00:00Z",
      "endTime": "2025-10-24T09:00:00Z",
      "durationSeconds": 3600,
      "hasVideo": true,
      "hasAudio": false,
      "recordingType": "continuous",
      "gaps": []
    },
    {
      "sequenceId": "seq-002",
      "startTime": "2025-10-24T14:00:00Z",
      "endTime": "2025-10-24T15:00:00Z",
      "durationSeconds": 3600,
      "hasVideo": true,
      "hasAudio": false,
      "recordingType": "motion",
      "gaps": []
    }
  ],
  "gaps": [
    {
      "gapId": "gap-001",
      "startTime": "2025-10-24T00:00:00Z",
      "endTime": "2025-10-24T08:00:00Z",
      "durationSeconds": 28800,  // 8 hours gap
      "reason": "no_recording"
    },
    {
      "gapId": "gap-002",
      "startTime": "2025-10-24T09:00:00Z",
      "endTime": "2025-10-24T14:00:00Z",
      "durationSeconds": 18000,  // 5 hours gap
      "reason": "no_recording"
    },
    {
      "gapId": "gap-003",
      "startTime": "2025-10-24T15:00:00Z",
      "endTime": "2025-10-24T23:59:59Z",
      "durationSeconds": 32399,  // 9 hours gap
      "reason": "no_recording"
    }
  ]
}
```

### Key Fields Explained:

- **`sequences[]`**: Array of **available** recording segments
  - Each sequence has exact start/end times
  - Tells you what **IS** available

- **`gaps[]`**: Array of **missing** recording segments
  - Each gap shows where recordings **don't exist**
  - Includes duration and reason (no_recording, camera_offline, disk_full, etc.)

- **`availableDuration`**: Total seconds of actual video
- **`coveragePercent`**: What percentage of requested time has recordings

---

## 🎥 Playback API Behavior with Partial Recordings

### 2. Playback Request with Gaps

**When you request playback** for the full day, Milestone provides a playback URL that **only plays available sequences**.

**API Endpoint**: `/api/rest/v1/recordings/{cameraId}/playback`

**Example Request (Full Day):**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/cam-main-entrance/playback?startTime=2025-10-24T00:00:00Z&endTime=2025-10-24T23:59:59Z&skipGaps=true"
```

**Response:**
```json
{
  "cameraId": "cam-main-entrance",
  "playbackUrl": "rtsp://192.168.1.9:554/playback/cam-main-entrance?start=2025-10-24T00:00:00Z&end=2025-10-24T23:59:59Z&skipGaps=true",
  "requestedRange": {
    "startTime": "2025-10-24T00:00:00Z",
    "endTime": "2025-10-24T23:59:59Z"
  },
  "actualPlayback": {
    "sequences": [
      {
        "sequenceId": "seq-001",
        "startTime": "2025-10-24T08:00:00Z",
        "endTime": "2025-10-24T09:00:00Z"
      },
      {
        "sequenceId": "seq-002",
        "startTime": "2025-10-24T14:00:00Z",
        "endTime": "2025-10-24T15:00:00Z"
      }
    ],
    "totalDuration": 7200,  // 2 hours
    "skipGaps": true
  },
  "warning": "Requested 24 hours, only 2 hours available. Gaps will be skipped during playback."
}
```

---

## 🔧 Gap Handling Options

### Option 1: `skipGaps=true` (Default)

**Behavior**: Playback **automatically jumps** over gaps

```
Timeline:
┌────────┬─────────────┬──────────┬─────────────┬─────────────┐
│  Gap   │  Sequence 1 │   Gap    │  Sequence 2 │    Gap      │
│ 8 hrs  │  08:00-09:00│  5 hrs   │ 14:00-15:00 │   9 hrs     │
└────────┴─────────────┴──────────┴─────────────┴─────────────┘

Playback with skipGaps=true:
  ↓ Skip     ▶ Play 1hr    ↓ Skip    ▶ Play 1hr    ↓ Skip

User Experience:
  - Playback starts at 08:00 (first available frame)
  - Plays 08:00-09:00 continuously
  - JUMPS to 14:00 instantly (gap skipped)
  - Plays 14:00-15:00 continuously
  - Playback ends
```

**Pros:**
- ✅ Fast - only plays actual video
- ✅ No wasted time on empty periods
- ✅ Good for investigations

**Cons:**
- ❌ User might not notice time jumps
- ❌ Can be confusing if not displayed clearly

---

### Option 2: `skipGaps=false`

**Behavior**: Shows **black screen** or **last frame** during gaps

```
Timeline:
┌────────┬─────────────┬──────────┬─────────────┬─────────────┐
│  Gap   │  Sequence 1 │   Gap    │  Sequence 2 │    Gap      │
│ 8 hrs  │  08:00-09:00│  5 hrs   │ 14:00-15:00 │   9 hrs     │
└────────┴─────────────┴──────────┴─────────────┴─────────────┘

Playback with skipGaps=false:
  ⬛ Black 8hr  ▶ Play 1hr  ⬛ Black 5hr  ▶ Play 1hr  ⬛ Black 9hr

User Experience:
  - Playback starts at 00:00 with black screen
  - Black screen for 8 hours (or fast-forward)
  - Plays 08:00-09:00 with real video
  - Black screen for 5 hours
  - Plays 14:00-15:00 with real video
  - Black screen for 9 hours
  - Playback ends at 23:59:59
```

**Pros:**
- ✅ Timeline accuracy maintained
- ✅ User sees exact time position
- ✅ Clear indication of missing data

**Cons:**
- ❌ Very slow playback (24 hours to watch 2 hours of video)
- ❌ Requires fast-forward controls

---

## 🎯 Recommended Approach for RTA CCTV System

### Step 1: Query Timeline FIRST

Before attempting playback, **always query the timeline** to check recording availability.

```python
def check_recording_availability(camera_id, start_time, end_time):
    """
    Check if recordings are available for requested time range.
    Returns availability info with sequences and gaps.
    """
    response = requests.get(
        f"http://192.168.1.9/api/rest/v1/recordings/{camera_id}/timeline",
        params={
            "startTime": start_time,
            "endTime": end_time
        },
        auth=('username', 'password')
    )

    timeline = response.json()

    return {
        "requested_duration": calculate_duration(start_time, end_time),
        "available_duration": timeline['availableDuration'],
        "coverage_percent": timeline['coveragePercent'],
        "sequences": timeline['sequences'],
        "gaps": timeline['gaps']
    }
```

### Step 2: Display Warning to User

If coverage is less than 100%, warn the user:

```python
availability = check_recording_availability(
    "cam-main-entrance",
    "2025-10-24T00:00:00Z",
    "2025-10-24T23:59:59Z"
)

if availability['coverage_percent'] < 100:
    print(f"⚠️  Warning: Only {availability['coverage_percent']:.1f}% of requested time has recordings")
    print(f"   Requested: {availability['requested_duration']} seconds")
    print(f"   Available: {availability['available_duration']} seconds")
    print(f"   Missing: {availability['requested_duration'] - availability['available_duration']} seconds")
    print()
    print("Available sequences:")
    for seq in availability['sequences']:
        print(f"   - {seq['startTime']} → {seq['endTime']}")
```

**Output:**
```
⚠️  Warning: Only 8.3% of requested time has recordings
   Requested: 86400 seconds (24 hours)
   Available: 7200 seconds (2 hours)
   Missing: 79200 seconds (22 hours)

Available sequences:
   - 2025-10-24T08:00:00Z → 2025-10-24T09:00:00Z
   - 2025-10-24T14:00:00Z → 2025-10-24T15:00:00Z
```

### Step 3: Let User Choose Playback Mode

Provide options to the user:

```
┌─────────────────────────────────────────────────────────┐
│  Recording Availability Warning                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Requested: 24 hours (Full day)                         │
│  Available: 2 hours (8.3% coverage)                     │
│                                                          │
│  Recordings available at:                               │
│    • 08:00 - 09:00 (1 hour)                            │
│    • 14:00 - 15:00 (1 hour)                            │
│                                                          │
│  How would you like to proceed?                         │
│                                                          │
│  [ Play Available Segments Only (Skip Gaps) ]           │
│  [ Play Full Timeline (Show Gaps as Black Screen) ]     │
│  [ Download Available Segments as Separate Files ]      │
│  [ Cancel ]                                             │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Step 4: Implement Playback Based on User Choice

**Option A: Skip Gaps (Recommended)**
```python
def playback_skip_gaps(camera_id, start_time, end_time):
    response = requests.get(
        f"http://192.168.1.9/api/rest/v1/recordings/{camera_id}/playback",
        params={
            "startTime": start_time,
            "endTime": end_time,
            "skipGaps": True,
            "speed": 1
        },
        auth=('username', 'password')
    )

    playback_url = response.json()['playbackUrl']

    # Stream to dashboard
    return stream_rtsp_to_dashboard(playback_url)
```

**Option B: Show Gaps**
```python
def playback_with_gaps(camera_id, start_time, end_time):
    response = requests.get(
        f"http://192.168.1.9/api/rest/v1/recordings/{camera_id}/playback",
        params={
            "startTime": start_time,
            "endTime": end_time,
            "skipGaps": False,
            "speed": 1
        },
        auth=('username', 'password')
    )

    playback_url = response.json()['playbackUrl']

    # Show timeline with gap indicators
    return stream_with_timeline_overlay(playback_url)
```

---

## 🖥️ Dashboard Implementation

### Timeline Visualization

Display the timeline with visual indicators for gaps:

```
RTA CCTV Dashboard - Playback View

Camera: Main Entrance
Date: 2025-10-24

Timeline (24 hours):
┌───────────────────────────────────────────────────────────┐
│ 00:00                12:00                23:59           │
├───────────────────────────────────────────────────────────┤
│ ░░░░░░░░ ████ ░░░░░ ████ ░░░░░░░░░░░░░░░░░░              │
│          ↑           ↑                                     │
│       08:00-09:00  14:00-15:00                            │
└───────────────────────────────────────────────────────────┘

Legend:
  ████  Video available
  ░░░░  No recording (gap)

Coverage: 8.3% (2 hours out of 24 hours)

[▶ Play] [⏸ Pause] [⏩ Fast Forward] [⏪ Rewind]

Playback Mode:
  ◉ Skip gaps (fast playback)
  ○ Show gaps (full timeline)
```

---

## 📊 API Response Examples

### Example 1: Perfect Coverage (No Gaps)

**Request:**
```bash
GET /api/rest/v1/recordings/cam-001/timeline?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T11:00:00Z
```

**Response:**
```json
{
  "cameraId": "cam-001",
  "coveragePercent": 100.0,
  "availableDuration": 3600,
  "sequences": [
    {
      "sequenceId": "seq-001",
      "startTime": "2025-10-24T10:00:00Z",
      "endTime": "2025-10-24T11:00:00Z",
      "durationSeconds": 3600
    }
  ],
  "gaps": []  // No gaps!
}
```

---

### Example 2: Partial Coverage (Multiple Gaps)

**Request:**
```bash
GET /api/rest/v1/recordings/cam-001/timeline?startTime=2025-10-24T00:00:00Z&endTime=2025-10-24T23:59:59Z
```

**Response:**
```json
{
  "cameraId": "cam-001",
  "coveragePercent": 45.8,
  "availableDuration": 39600,  // ~11 hours
  "sequences": [
    {
      "sequenceId": "seq-001",
      "startTime": "2025-10-24T06:00:00Z",
      "endTime": "2025-10-24T09:00:00Z",
      "durationSeconds": 10800  // 3 hours
    },
    {
      "sequenceId": "seq-002",
      "startTime": "2025-10-24T12:00:00Z",
      "endTime": "2025-10-24T20:00:00Z",
      "durationSeconds": 28800  // 8 hours
    }
  ],
  "gaps": [
    {
      "gapId": "gap-001",
      "startTime": "2025-10-24T00:00:00Z",
      "endTime": "2025-10-24T06:00:00Z",
      "durationSeconds": 21600,  // 6 hours
      "reason": "camera_offline"
    },
    {
      "gapId": "gap-002",
      "startTime": "2025-10-24T09:00:00Z",
      "endTime": "2025-10-24T12:00:00Z",
      "durationSeconds": 10800,  // 3 hours
      "reason": "motion_recording_no_motion"
    },
    {
      "gapId": "gap-003",
      "startTime": "2025-10-24T20:00:00Z",
      "endTime": "2025-10-24T23:59:59Z",
      "durationSeconds": 14399,  // ~4 hours
      "reason": "no_recording"
    }
  ]
}
```

---

### Example 3: No Coverage (Complete Gap)

**Request:**
```bash
GET /api/rest/v1/recordings/cam-001/timeline?startTime=2025-10-23T00:00:00Z&endTime=2025-10-23T23:59:59Z
```

**Response:**
```json
{
  "cameraId": "cam-001",
  "coveragePercent": 0.0,
  "availableDuration": 0,
  "sequences": [],  // No sequences!
  "gaps": [
    {
      "gapId": "gap-001",
      "startTime": "2025-10-23T00:00:00Z",
      "endTime": "2025-10-23T23:59:59Z",
      "durationSeconds": 86400,  // Full 24 hours
      "reason": "retention_expired"
    }
  ],
  "error": "NO_RECORDINGS_AVAILABLE",
  "message": "No recordings found for the requested time range. Recordings may have been deleted due to retention policy."
}
```

---

## 🔄 Gap Reasons

Common reasons for gaps in recordings:

| Reason Code | Description | User Message |
|-------------|-------------|--------------|
| `no_recording` | Recording was disabled | "Recording was not enabled during this period" |
| `camera_offline` | Camera was disconnected | "Camera was offline or disconnected" |
| `disk_full` | Storage full, couldn't record | "Recording server storage was full" |
| `motion_recording_no_motion` | Motion-based recording, no motion | "No motion detected (motion-based recording)" |
| `retention_expired` | Recordings deleted (past retention) | "Recordings deleted - past retention period (90 days)" |
| `network_error` | Network issues during recording | "Network error prevented recording" |
| `database_error` | Database failure | "Database error - recordings may be corrupted" |
| `manual_deletion` | Admin deleted recordings | "Recordings manually deleted by administrator" |

---

## 🚀 Complete Integration Example

### Python: Check Availability + Playback

```python
import requests
from datetime import datetime, timedelta

class MilestonePlaybackService:
    def __init__(self, server_url, username, password):
        self.server_url = server_url
        self.auth = (username, password)

    def check_availability(self, camera_id, start_time, end_time):
        """
        Check recording availability for a time range.
        Returns timeline with sequences and gaps.
        """
        response = requests.get(
            f"{self.server_url}/api/rest/v1/recordings/{camera_id}/timeline",
            params={
                "startTime": start_time,
                "endTime": end_time
            },
            auth=self.auth
        )
        response.raise_for_status()
        return response.json()

    def get_playback_url(self, camera_id, start_time, end_time, skip_gaps=True):
        """
        Get playback URL for a time range.

        Args:
            skip_gaps: If True, playback skips over gaps automatically
        """
        # First, check availability
        timeline = self.check_availability(camera_id, start_time, end_time)

        # Warn if coverage is low
        if timeline['coveragePercent'] < 100:
            print(f"⚠️  Warning: Only {timeline['coveragePercent']:.1f}% coverage")
            print(f"   {len(timeline['sequences'])} sequences available")
            print(f"   {len(timeline['gaps'])} gaps detected")

        # Get playback URL
        response = requests.get(
            f"{self.server_url}/api/rest/v1/recordings/{camera_id}/playback",
            params={
                "startTime": start_time,
                "endTime": end_time,
                "skipGaps": str(skip_gaps).lower(),
                "speed": 1
            },
            auth=self.auth
        )
        response.raise_for_status()

        return response.json()

    def export_available_sequences(self, camera_id, start_time, end_time):
        """
        Export only the available sequences (skip gaps entirely).
        Returns list of export jobs.
        """
        # Get timeline
        timeline = self.check_availability(camera_id, start_time, end_time)

        if not timeline['sequences']:
            raise ValueError("No recordings available for requested time range")

        # Export each sequence separately
        export_jobs = []
        for seq in timeline['sequences']:
            response = requests.post(
                f"{self.server_url}/api/rest/v1/recordings/export",
                json={
                    "cameraId": camera_id,
                    "startTime": seq['startTime'],
                    "endTime": seq['endTime'],
                    "format": "mp4",
                    "quality": "high"
                },
                auth=self.auth
            )
            response.raise_for_status()
            export_jobs.append(response.json())

        return export_jobs


# Usage Example
service = MilestonePlaybackService(
    server_url="http://192.168.1.9",
    username="rta-integration",
    password="your-password"
)

# Check availability for full day
availability = service.check_availability(
    camera_id="cam-main-entrance",
    start_time="2025-10-24T00:00:00Z",
    end_time="2025-10-24T23:59:59Z"
)

print(f"Coverage: {availability['coveragePercent']:.1f}%")
print(f"Sequences: {len(availability['sequences'])}")
print(f"Gaps: {len(availability['gaps'])}")

# If acceptable coverage, get playback URL
if availability['coveragePercent'] > 5:  # At least 5% available
    playback = service.get_playback_url(
        camera_id="cam-main-entrance",
        start_time="2025-10-24T00:00:00Z",
        end_time="2025-10-24T23:59:59Z",
        skip_gaps=True  # Skip gaps for faster playback
    )

    print(f"Playback URL: {playback['playbackUrl']}")
else:
    print("Insufficient recording coverage. Cannot playback.")
```

---

## ✅ Best Practices

### 1. Always Query Timeline First
- Don't attempt playback without checking availability
- Show coverage percentage to users
- Display sequence/gap information

### 2. Provide Clear User Feedback
- Show warnings when coverage < 100%
- Display visual timeline with gaps
- Let users choose skip/show gaps mode

### 3. Handle Edge Cases
- **No recordings**: Show clear error message
- **Partial recordings**: Warn user before playback
- **Expired recordings**: Explain retention policy

### 4. Optimize for User Experience
- Default to `skipGaps=true` for investigations
- Provide timeline scrubber showing available periods
- Allow downloading only available sequences

---

## 📝 Summary

| Scenario | API Response | Recommended Action |
|----------|--------------|-------------------|
| **100% Coverage** | All requested time has recordings | Proceed with normal playback |
| **Partial Coverage** | Some gaps in timeline | Show warning, let user choose gap handling |
| **No Coverage** | No recordings available | Display error, suggest different time range |
| **Low Coverage (< 10%)** | Very few recordings | Warn user, offer to download segments instead |

**Key Takeaway**: Milestone provides comprehensive timeline/sequences API that lets you **check availability before playback**, giving you full control over how to handle partial recordings and gaps in the RTA CCTV System.

---

**Last Updated:** 2025-10-25
**Milestone Server:** 192.168.1.9
**Related Documents:**
- [MILESTONE_PLAYBACK_API_EXPLAINED.md](./MILESTONE_PLAYBACK_API_EXPLAINED.md)
- [PLAYBACK_VS_DOWNLOAD_VS_EXPORT.md](./PLAYBACK_VS_DOWNLOAD_VS_EXPORT.md)
- [MILESTONE_API_REFERENCE.md](./MILESTONE_API_REFERENCE.md)
