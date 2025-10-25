# 🎬 Playback vs Download vs Export - Explained

## 📋 Overview

When you want to view or retrieve recorded video from Milestone VMS for a specific timeline, there are **three different approaches**, each serving a different purpose.

---

## 🎥 1. PLAYBACK (Stream Playback)

### What is it?
**Real-time streaming** of recorded video from a specific time range. The video is **NOT downloaded** to your system - it streams directly like watching Netflix or YouTube.

### How it works:
```
You → Request playback URL → Milestone streams video → You watch in real-time
```

### API Endpoint:
```http
GET /api/rest/v1/recordings/{cameraId}/playback
  ?startTime=2025-10-24T10:00:00Z
  &endTime=2025-10-24T11:00:00Z
```

### Response:
```json
{
  "playbackUrl": "rtsp://192.168.1.9:554/playback/{cameraId}?start=2025-10-24T10:00:00Z&end=2025-10-24T11:00:00Z"
}
```

### Usage:
```bash
# Stream the video using VLC or ffplay
ffplay "rtsp://192.168.1.9:554/playback/{cameraId}?start=2025-10-24T10:00:00Z&duration=3600"

# Or integrate with your dashboard
# The dashboard plays the RTSP stream directly
```

### Characteristics:
✅ **Instant** - Starts playing immediately
✅ **No storage needed** - Doesn't save file to disk
✅ **Scrubbing support** - Can jump to any point in timeline
✅ **Speed control** - Can play at 1x, 2x, 4x, 8x speed
✅ **Live streaming** - Video is transmitted frame-by-frame
❌ **Requires connection** - Must stay connected to Milestone server
❌ **No offline viewing** - Can't watch without internet connection

### Use Cases:
- 👁️ **Dashboard playback viewer** - Watch recordings in real-time
- 🔍 **Investigation** - Scrub through timeline to find events
- 📊 **Live monitoring** - Review recent events quickly
- 🎮 **Interactive playback** - Pause, rewind, fast-forward

### Example Scenario:
```
Security operator wants to review what happened at 10:00 AM:
1. Opens dashboard playback view
2. Selects camera and time: 10:00 AM - 11:00 AM
3. System calls playback API to get RTSP stream URL
4. Video starts playing immediately in the browser
5. Operator can scrub timeline, pause, or speed up
```

---

## 💾 2. DOWNLOAD (Direct Download)

### What is it?
**Immediately downloads** the recorded video as a **complete file** (MP4, AVI, MKV) to your local storage.

### How it works:
```
You → Request download → Milestone packages video → Complete file downloaded → Saved to disk
```

### API Endpoint:
```http
GET /api/rest/v1/recordings/{cameraId}/download
  ?startTime=2025-10-24T10:00:00Z
  &endTime=2025-10-24T10:30:00Z
  &format=mp4
```

### Response:
```
Content-Type: video/mp4
Content-Disposition: attachment; filename="recording_2025-10-24.mp4"

<binary video file - downloaded directly>
```

### Usage:
```bash
# Download the video file
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/{cameraId}/download?startTime=2025-10-24T10:00:00Z&endTime=2025-10-24T10:30:00Z&format=mp4" \
  -o recording.mp4

# Now you have recording.mp4 saved locally
# You can play it offline with any media player
```

### Characteristics:
✅ **Offline viewing** - Can watch without server connection
✅ **Permanent copy** - File saved on your computer
✅ **Shareable** - Can send file to others
✅ **Any player** - Works with VLC, Windows Media Player, etc.
⏳ **Slower start** - Must download entire file before viewing
💾 **Storage required** - Takes up disk space
❌ **Not instant** - Wait time depends on video size
❌ **No scrubbing during download** - Can't seek until download completes

### Use Cases:
- 📧 **Evidence sharing** - Send video clip to police, insurance
- 💼 **Archival** - Save important footage locally
- 📱 **Offline review** - Watch video without internet
- 🎓 **Training** - Share clips with team members
- 📁 **Backup** - Keep local copy of critical footage

### Example Scenario:
```
An incident occurred at 10:00 AM. Manager needs video evidence:
1. Opens RTA system and selects time range: 10:00-10:30 AM
2. Clicks "Download Video Clip"
3. System calls download API
4. A 500MB MP4 file downloads to manager's computer
5. Manager attaches the file to an email and sends to police
```

---

## 📤 3. EXPORT (Asynchronous Export)

### What is it?
**Background job** that processes and packages video for later download. Used for **large video files** that take time to prepare.

### How it works:
```
You → Submit export request → Job queued → Milestone processes in background →
Job completes → Download link provided → You download when ready
```

### Why needed?
When you request a **large time range** (e.g., 2 hours of video = 14GB file), Milestone needs time to:
1. Find all recording segments
2. Stitch them together
3. Re-encode if needed
4. Add watermarks (optional)
5. Package into single file

This can take **minutes to hours**, so it runs as a background job.

### API Workflow:

**Step 1: Submit Export Request**
```http
POST /api/rest/v1/recordings/export
Content-Type: application/json

{
  "cameraId": "8b3a2c1d-4e5f-6789-abcd-ef0123456789",
  "startTime": "2025-10-24T08:00:00Z",
  "endTime": "2025-10-24T18:00:00Z",  // 10 hours of video!
  "format": "mp4",
  "quality": "high",
  "watermark": {
    "enabled": true,
    "text": "RTA CCTV - Exported {timestamp}"
  }
}
```

**Response:**
```json
{
  "exportId": "exp-12345-abcde",
  "status": "queued",
  "estimatedDurationSeconds": 36000,
  "estimatedSizeBytes": 262144000000,  // ~244 GB!
  "createdAt": "2025-10-25T10:00:00Z"
}
```

**Step 2: Poll Export Status**
```http
GET /api/rest/v1/recordings/export/exp-12345-abcde

Response:
{
  "exportId": "exp-12345-abcde",
  "status": "processing",  // queued → processing → completed
  "progress": 45,          // 45% complete
  "estimatedTimeRemaining": 600
}
```

**Step 3: Download When Ready**
```http
GET /api/rest/v1/recordings/export/exp-12345-abcde

Response:
{
  "exportId": "exp-12345-abcde",
  "status": "completed",
  "progress": 100,
  "downloadUrl": "http://192.168.1.9/api/rest/v1/recordings/export/exp-12345-abcde/download",
  "fileSize": 262144000000,
  "expiresAt": "2025-10-26T10:00:00Z"  // Available for 24 hours
}
```

**Step 4: Download the File**
```bash
curl -u username:password \
  "http://192.168.1.9/api/rest/v1/recordings/export/exp-12345-abcde/download" \
  -o exported_video.mp4
```

### Characteristics:
✅ **Large files** - Handle hours of video (10+ hours)
✅ **Background processing** - Don't block other operations
✅ **Progress tracking** - See percentage complete
✅ **Watermarking** - Add custom text to video
✅ **Optimization** - Can re-encode for smaller size
⏳ **Long wait** - Can take minutes to hours
⏰ **Time-limited** - Download link expires (24 hours)
🔄 **Asynchronous** - Must poll status

### Use Cases:
- 📚 **Long-term archival** - Export full day of footage
- 🎥 **Video evidence packages** - Comprehensive incident reports
- 📊 **Monthly reports** - Package entire month of footage
- 🔐 **Legal compliance** - Watermarked, certified exports
- 💿 **Offline storage** - Burn to DVD/Blu-ray

### Example Scenario:
```
Police request full day of footage for investigation:
1. Administrator submits export request for 24 hours of video
2. Export job ID: exp-12345 is returned
3. System shows "Processing: 15% complete"
4. Administrator goes to lunch
5. Returns an hour later, export is complete
6. Downloads 200GB MP4 file
7. Copies to external hard drive for police
```

---

## 📊 Side-by-Side Comparison

| Feature | Playback (Stream) | Download | Export |
|---------|-------------------|----------|--------|
| **Speed** | ⚡ Instant | 🐢 Minutes | 🐌 Minutes-Hours |
| **File Size** | N/A (streaming) | Small-Medium | Medium-Large |
| **Storage Required** | None | Local storage | Local storage |
| **Offline Viewing** | ❌ No | ✅ Yes | ✅ Yes |
| **Timeline Scrubbing** | ✅ Yes | ✅ Yes (after download) | ✅ Yes (after export) |
| **Best For** | Live review | Short clips | Large archives |
| **Max Duration** | Unlimited | ~30 min | Unlimited |
| **Processing Time** | None | Seconds-Minutes | Minutes-Hours |
| **Shareable** | ❌ No | ✅ Yes | ✅ Yes |
| **Watermarking** | ❌ No | ❌ No | ✅ Yes |
| **Progress Tracking** | N/A | ❌ No | ✅ Yes |

---

## 🎯 Which One Should You Use?

### Use **PLAYBACK** when:
- ✅ Viewing video in the **dashboard**
- ✅ Investigating incidents **interactively**
- ✅ Need **immediate** playback
- ✅ **Scrubbing** through timeline
- ✅ Don't need to save the file

### Use **DOWNLOAD** when:
- ✅ Need **short clip** (< 30 minutes)
- ✅ Want to **share** with someone
- ✅ Need **offline** viewing
- ✅ Creating **evidence** package
- ✅ File is **small enough** to download quickly

### Use **EXPORT** when:
- ✅ Need **long video** (> 1 hour)
- ✅ File is **very large** (> 10 GB)
- ✅ Need **watermarking**
- ✅ Creating **official** records
- ✅ Can **wait** for processing

---

## 🏗️ RTA CCTV System Architecture

### How Your System Uses These:

```
┌─────────────────────────────────────────────────────────────┐
│                    RTA CCTV Dashboard                        │
│  (User Interface - http://localhost:3000)                   │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ User Actions
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Use Case Decision                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  "Watch Recording"           "Download Clip"    "Export Day"│
│        │                            │                  │     │
│        ▼                            ▼                  ▼     │
│   PLAYBACK API             DOWNLOAD API         EXPORT API  │
│        │                            │                  │     │
│        ▼                            ▼                  ▼     │
│  Stream RTSP              Get MP4 File       Background Job  │
│        │                            │                  │     │
│        ▼                            ▼                  ▼     │
│  MediaMTX Re-stream      Save to Disk        Poll Status    │
│        │                            │                  │     │
│        ▼                            ▼                  ▼     │
│  Play in Browser        Share/Archive      Download File   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 💡 Real-World Examples

### Example 1: Security Guard Reviews Last Hour
```
Scenario: Guard wants to check who entered building at 2 PM

Solution: PLAYBACK
- Guard opens playback view in dashboard
- Selects camera: "Main Entrance"
- Selects time: 2:00 PM - 3:00 PM
- System streams RTSP playback
- Guard scrubs timeline, sees person at 2:15 PM
- No file saved, instant viewing
```

### Example 2: Manager Needs Evidence for Insurance
```
Scenario: Car accident in parking lot at 10:30 AM, need 5-minute clip

Solution: DOWNLOAD
- Manager opens RTA system
- Selects camera: "Parking Lot"
- Selects time: 10:25 AM - 10:35 AM (10 minutes)
- Clicks "Download Clip"
- 800MB MP4 file downloads in 2 minutes
- Manager emails file to insurance company
```

### Example 3: Police Request Full Week of Footage
```
Scenario: Investigation requires entire week of footage from 3 cameras

Solution: EXPORT
- Administrator selects 3 cameras
- Time range: October 18-25 (7 days × 24 hours = 168 hours)
- Submits export request
- System estimates: 1.5TB file, 4 hours processing time
- Administrator monitors progress (25%, 50%, 75%, 100%)
- Download complete after 4 hours
- Copies to external hard drive for police
```

---

## 🔄 Integration with RTA CCTV System

### Your Playback Service (Port 8090) Should:

1. **For Dashboard Playback** → Use PLAYBACK API
   ```python
   def get_playback_stream(camera_id, start_time, end_time):
       # Call Milestone playback API
       response = milestone_api.get_playback_url(camera_id, start_time, end_time)
       rtsp_url = response['playbackUrl']

       # Re-stream via MediaMTX for dashboard
       mediamtx_url = f"rtsp://mediamtx:8554/playback/{camera_id}"

       return {
           "stream_url": mediamtx_url,
           "protocol": "webrtc",  # For low latency in browser
           "hls_url": f"http://mediamtx:8888/playback/{camera_id}/index.m3u8"
       }
   ```

2. **For Clip Export** → Use DOWNLOAD API (short clips)
   ```python
   def export_clip(camera_id, start_time, end_time):
       if duration < 30_minutes:
           # Use direct download
           video_file = milestone_api.download_recording(camera_id, start_time, end_time)
           return save_to_minio(video_file, bucket='cctv-exports')
       else:
           # Use async export
           export_job = milestone_api.create_export(camera_id, start_time, end_time)
           return {"export_id": export_job.id, "status": "processing"}
   ```

---

## ✅ Summary

| Method | What It Does | When To Use | RTA System Usage |
|--------|--------------|-------------|------------------|
| **PLAYBACK** | Stream video in real-time | Interactive viewing | ✅ Dashboard playback view |
| **DOWNLOAD** | Get video file immediately | Short clips to share | ✅ Export clips (< 30 min) |
| **EXPORT** | Background job for large files | Long-term archives | ✅ Evidence packages (> 1 hour) |

**Key Insight:**
- **PLAYBACK** = Streaming (like Netflix) 🎬
- **DOWNLOAD** = Quick file download 📥
- **EXPORT** = Background processing for big files 📤

For your RTA CCTV dashboard, you'll primarily use **PLAYBACK API** for the timeline playback feature! 🎯

---

**Last Updated:** 2025-10-25
