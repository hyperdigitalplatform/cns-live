# 🎥 RTA CCTV System - Verification Guide

## 📋 System Overview

**Key Services & Ports:**
- **Dashboard (Frontend)**: http://localhost:3000
- **API Gateway (Kong)**: http://localhost:8000 (Proxy), http://localhost:8001 (Admin)
- **VMS Service**: http://localhost:8081
- **Storage Service**: http://localhost:8082
- **Recording Service**: http://localhost:8083
- **Metadata Service**: http://localhost:8084
- **Stream Counter**: http://localhost:8087
- **Go API**: http://localhost:8088
- **Playback Service**: http://localhost:8092 ⚠️ (Changed from 8090)
- **MediaMTX (RTSP Server)**: rtsp://localhost:8554
- **LiveKit (WebRTC)**: ws://localhost:7880
- **Grafana (Monitoring)**: http://localhost:3001
- **MinIO (Storage)**: http://localhost:9001 (Console)
- **Prometheus**: http://localhost:9090

---

## 📌 Important API Endpoint Notes

### Kong Gateway vs Direct Access

**Via Kong Gateway (Port 8000) - Recommended for external clients:**
- VMS APIs: `http://localhost:8000/vms/*`
- Stream APIs: `http://localhost:8000/api/v1/stream/*`
- RTSP APIs: `http://localhost:8000/api/v1/rtsp/*`

**Direct Service Access - For testing/debugging:**
- VMS Service: `http://localhost:8081/vms/*`
- Stream Counter: `http://localhost:8087/api/v1/stream/*`
- Other services: Use their respective ports

### ⚠️ Common Mistakes to Avoid
1. **Wrong:** `http://localhost:8081/api/v1/cameras` → **Correct:** `http://localhost:8081/vms/cameras`
2. **Wrong:** `http://localhost:8087/api/v1/health` → **Correct:** `http://localhost:8087/health`
3. **Wrong:** `http://localhost:8000/health` → **Correct:** `http://localhost:8001/status` (Kong Admin API)
4. **Wrong:** Port 8090 for playback → **Correct:** Port 8092

---

## 🔧 Phase 1: System Health Check

### 1.1 Verify All Services are Running
```bash
docker ps --filter "name=cctv-" --format "table {{.Names}}\t{{.Status}}" | grep healthy
```

**Expected**: All core services should show "(healthy)" status

### 1.2 Check API Gateway Health
```bash
# Kong Admin API Status (detailed)
curl http://localhost:8001/status

# Kong Admin API (simple check)
curl http://localhost:8001
```

**Expected Response:**
```json
{
  "server": {
    "connections_active": 21,
    "connections_accepted": 49,
    "total_requests": 49
  },
  "configuration_hash": "..."
}
```

### 1.3 Check Stream Counter Service
```bash
curl http://localhost:8087/health
```

**Expected Response:**
```json
{
  "service": "stream-counter",
  "status": "healthy",
  "timestamp": "2025-10-25T..."
}
```

### 1.4 Check All Service Health Endpoints
```bash
# VMS Service
curl http://localhost:8081/health

# Storage Service
curl http://localhost:8082/health

# Recording Service
curl http://localhost:8083/health

# Metadata Service
curl http://localhost:8084/health

# Stream Counter
curl http://localhost:8087/health

# Go API
curl http://localhost:8088/health

# Playback Service (Note: Port changed from 8090 to 8092)
curl http://localhost:8092/health
```

---

## 🎬 Phase 2: Camera Configuration & Registration

### 2.1 List Cameras from VMS Service

The VMS Service connects to Milestone VMS and retrieves camera information.

**API Endpoints:**

**Direct Access:**
```bash
# Get all cameras
curl http://localhost:8081/vms/cameras

# Get all cameras (via Kong Gateway)
curl http://localhost:8000/vms/cameras

# Filter by source
curl http://localhost:8081/vms/cameras?source=DUBAI_POLICE
```

**Expected Response:**
```json
{
  "cameras": [
    {
      "id": "a14c5b2b-c315-4f68-a87b-dffbfb60917b",
      "name": "Camera 001 - Sheikh Zayed Road",
      "name_ar": "كاميرا 001 - شارع الشيخ زايد",
      "source": "DUBAI_POLICE",
      "rtsp_url": "rtsp://milestone.rta.ae:554/camera_001",
      "ptz_enabled": true,
      "status": "ONLINE",
      "recording_server": "milestone.rta.ae:554",
      "metadata": {
        "fps": 25,
        "location": {
          "lat": 25.2048,
          "lon": 55.2708
        },
        "resolution": "1920x1080"
      }
    }
  ],
  "total": 2,
  "last_updated": "2025-10-25T..."
}
```

### 2.2 Get Specific Camera Details

✅ **FIXED:** The VMS service now uses PostgreSQL database for persistence, so camera IDs are stable.

**Test with sample cameras:**
```bash
# Get a camera ID from the list
curl http://localhost:8081/vms/cameras

# Use one of the sample IDs
curl http://localhost:8081/vms/cameras/cam-001-sheikh-zayed

# Or the metro camera
curl http://localhost:8081/vms/cameras/cam-002-metro-station
```

**Expected Response:**
```json
{
  "id": "cam-001-sheikh-zayed",
  "name": "Camera 001 - Sheikh Zayed Road",
  "name_ar": "كاميرا 001 - شارع الشيخ زايد",
  "source": "DUBAI_POLICE",
  "rtsp_url": "rtsp://milestone.rta.ae:554/camera_001",
  "ptz_enabled": true,
  "status": "ONLINE",
  "recording_server": "milestone.rta.ae:554",
  "milestone_device_id": "milestone_device_001",
  "metadata": {
    "fps": 25,
    "location": {
      "lat": 25.2048,
      "lon": 55.2708
    },
    "resolution": "1920x1080"
  },
  "last_update": "2025-10-25T...",
  "created_at": "2025-09-25T..."
}
```

### 2.3 Get Camera Stream URL
```bash
# Direct access
curl http://localhost:8081/vms/cameras/{camera_id}/stream

# Via Kong Gateway
curl http://localhost:8000/vms/cameras/{camera_id}/stream
```

---

## 📺 Phase 3: Live Stream Viewing & Dashboard Features

### 3.1 Access the Dashboard
Open your browser and navigate to:
```
http://localhost:3000
```

### 3.2 Dashboard Enhanced Features

**Dashboard Capabilities:**
- ✅ **Tree-style folder organization** with unlimited nesting depth
- ✅ **Drag-and-drop** cameras between folders and to grid cells
- ✅ **9 grid layouts** from 1×1 to 6×6 (up to 36 cameras)
- ✅ **Search and filtering** across all cameras and folders
- ✅ **Persistent folder structure** saved in browser storage
- ✅ **View modes**: Tree view (hierarchical) and List view (flat)
- ✅ **PTZ Controls**: Pan/Tilt/Zoom controls (if camera supports)

---

## 📂 Phase 3A: Folder Management Testing

### 3A.1 Verify Default Folders

**Test Steps:**
1. Open dashboard at `http://localhost:3000`
2. Navigate to "Live View Enhanced" page
3. Observe the sidebar on the left

**Expected Results:**
- [ ] Sidebar displays default folders:
  - Dubai Police (شرطة دبي)
  - Sharjah Police (شرطة الشارقة)
  - Metro (المترو)
  - Taxi (التاكسي)
  - Parking (مواقف السيارات)
  - Unorganized (غير منظم)
- [ ] Each folder shows camera count badge (e.g., "5 cameras")
- [ ] Folders have expand/collapse arrows
- [ ] Tree view icon is selected by default

---

### 3A.2 Create Root Folder

**Test Steps:**
1. Click the **"+ Folder"** button in sidebar header
2. Enter folder name: `Traffic Department`
3. Enter Arabic name (optional): `قسم المرور`
4. Press Enter or click outside to save

**Expected Results:**
- [ ] New folder appears in the folder tree
- [ ] Folder is expanded by default
- [ ] Folder shows "0 cameras" badge
- [ ] Folder has expand/collapse arrow
- [ ] Folder persists after page refresh

**Verification:**
```bash
# Check browser local storage
# Open DevTools → Application → Local Storage → http://localhost:3000
# Look for key: folder-storage
# Should contain new folder with name "Traffic Department"
```

---

### 3A.3 Create Subfolder

**Test Steps:**
1. Right-click on "Dubai Police" folder
2. Select **"Add Subfolder"** from context menu
3. Enter subfolder name: `Highway Cameras`
4. Press Enter to save

**Expected Results:**
- [ ] Subfolder appears nested under "Dubai Police"
- [ ] Subfolder is indented to show hierarchy
- [ ] Subfolder shows "0 cameras" badge
- [ ] Parent folder auto-expands to show subfolder
- [ ] Subfolder persists after page refresh

---

### 3A.4 Rename Folder

**Test Steps (Method 1 - Context Menu):**
1. Right-click on "Traffic Department" folder
2. Select **"Rename"** from context menu
3. Edit name to: `Traffic Control`
4. Press Enter to save

**Test Steps (Method 2 - Double-click):**
1. Double-click on "Traffic Control" folder name
2. Edit name to: `Traffic Management`
3. Press Enter or click outside to save

**Expected Results:**
- [ ] Folder name updates immediately
- [ ] Edit mode has blue outline
- [ ] Can cancel edit with Escape key
- [ ] Renamed folder persists after refresh
- [ ] Camera assignments remain intact

---

### 3A.5 Delete Folder

**Test Steps:**
1. Create a test folder: "Test Folder"
2. Add 2 cameras to "Test Folder" (via drag-and-drop)
3. Right-click on "Test Folder"
4. Select **"Delete"** from context menu
5. Confirm deletion

**Expected Results:**
- [ ] Folder is removed from tree
- [ ] Cameras from deleted folder move to parent folder
- [ ] If root folder: cameras move to "Unorganized"
- [ ] Subfolders (if any) move to parent folder
- [ ] Deletion persists after refresh

---

### 3A.6 Move Folder via Drag-and-Drop

**Test Steps:**
1. Create folder: "Intersection Cameras"
2. Drag "Intersection Cameras" folder
3. Drop onto "Traffic Management" folder

**Expected Results:**
- [ ] "Intersection Cameras" becomes subfolder of "Traffic Management"
- [ ] Visual feedback during drag (semi-transparent drag preview)
- [ ] Target folder highlights in blue on hover
- [ ] Folder hierarchy updates correctly
- [ ] Cannot create circular reference (folder cannot be its own ancestor)

**Test Circular Reference Prevention:**
1. Try to drag "Traffic Management" folder
2. Drop onto its own subfolder "Intersection Cameras"

**Expected Results:**
- [ ] Drop is prevented/rejected
- [ ] No hierarchy change occurs
- [ ] Console may show warning (optional)

---

### 3A.7 Expand/Collapse Folders

**Test Steps:**
1. Click arrow icon next to "Dubai Police" folder
2. Folder should collapse (hide subfolders and cameras)
3. Click arrow icon again
4. Folder should expand (show contents)

**Expected Results:**
- [ ] Arrow rotates to indicate state (▶ collapsed, ▼ expanded)
- [ ] Contents appear/disappear smoothly
- [ ] Expanded state persists after refresh
- [ ] Nested folders can be independently collapsed

**Bulk Operations:**
1. Click **"Expand All"** button in toolbar

**Expected Results:**
- [ ] All folders expand recursively
- [ ] All cameras become visible

2. Click **"Collapse All"** button in toolbar

**Expected Results:**
- [ ] All folders collapse to root level
- [ ] Only root folders visible

---

## 🖱️ Phase 3B: Drag-and-Drop Testing

### 3B.1 Drag Camera to Folder

**Test Steps:**
1. Ensure "Unorganized" folder has cameras
2. Click and hold a camera from "Unorganized"
3. Drag over "Dubai Police" folder
4. Release mouse button to drop

**Expected Results:**
- [ ] Camera appears semi-transparent while dragging
- [ ] Target folder highlights in **blue** on hover
- [ ] Camera moves from "Unorganized" to "Dubai Police"
- [ ] Camera count badges update (Unorganized -1, Dubai Police +1)
- [ ] Camera assignment persists after refresh

**Visual Feedback Verification:**
- [ ] Drag cursor shows item being dragged
- [ ] Drop target has blue ring/highlight
- [ ] Invalid drop targets show no highlight or red border

---

### 3B.2 Drag Camera Between Folders

**Test Steps:**
1. Drag a camera from "Dubai Police" folder
2. Drop onto "Metro" folder

**Expected Results:**
- [ ] Camera moves from source to target folder
- [ ] Both folder badges update counts
- [ ] Camera disappears from source folder
- [ ] Camera appears in target folder
- [ ] Change persists after refresh

---

### 3B.3 Drag Camera to Grid Cell

**Test Steps:**
1. Select **3×3** grid layout (9 cells)
2. Drag a camera from sidebar
3. Hover over empty grid cell (Cell 1)
4. Release to drop

**Expected Results:**
- [ ] Grid cell highlights on hover (border or background change)
- [ ] Camera stream starts loading immediately after drop
- [ ] Cell shows camera name (English and Arabic)
- [ ] Cell shows online/offline status indicator
- [ ] Loading spinner appears briefly before stream starts
- [ ] Stream plays within 2-5 seconds (WHIP latency ~450ms)

**Test Multiple Cells:**
1. Drag 4 different cameras to Cells 1, 2, 3, 4

**Expected Results:**
- [ ] Each cell displays different camera stream
- [ ] All streams play simultaneously
- [ ] No camera feed duplication
- [ ] Performance remains smooth (no lag)

---

### 3B.4 Drag Camera to Occupied Grid Cell

**Test Steps:**
1. Drag a new camera to a grid cell already displaying a stream
2. Drop the camera

**Expected Results:**
- [ ] New camera **replaces** the existing camera in that cell
- [ ] Old stream stops immediately
- [ ] New stream starts loading
- [ ] Previous camera can still be found in sidebar

---

### 3B.5 Remove Camera from Grid Cell

**Test Steps:**
1. Hover over a grid cell with an active camera stream
2. Click the **×** (close) button in top-right corner of cell

**Expected Results:**
- [ ] Camera stream stops immediately
- [ ] Cell becomes empty placeholder
- [ ] Cell displays "Drop camera here" message
- [ ] Cell accepts new camera drops
- [ ] Camera remains in sidebar folder (not deleted)

---

### 3B.6 Drag Folder to Folder (Reorganize Hierarchy)

**Test Steps:**
1. Create folders: "Location A" and "Location B"
2. Drag "Location B" folder
3. Drop onto "Location A" folder

**Expected Results:**
- [ ] "Location B" becomes subfolder of "Location A"
- [ ] All cameras in "Location B" remain assigned
- [ ] Folder hierarchy visual indentation updates
- [ ] Change persists after refresh

---

## 🎛️ Phase 3C: Grid Layout Testing

### 3C.1 Test All Grid Layouts

**Test Each Layout:**

| Layout | Cells | Test Steps |
|--------|-------|------------|
| 1×1 | 1 | Select layout, drag 1 camera, verify single large cell |
| 2×2 | 4 | Select layout, drag 4 cameras, verify 2x2 arrangement |
| 3×3 | 9 | Select layout, drag 9 cameras, verify 3x3 arrangement |
| 4×4 | 16 | Select layout, drag 16 cameras, verify 4x4 arrangement |
| 2×3 | 6 | Select layout, drag 6 cameras, verify 2 rows × 3 columns |
| 3×4 | 12 | Select layout, drag 12 cameras, verify 3 rows × 4 columns |
| 4×5 | 20 | Select layout, drag 20 cameras, verify 4 rows × 5 columns |
| 5×5 | 25 | Select layout, drag 25 cameras, verify 5x5 arrangement |
| 6×6 | 36 | Select layout, drag 36 cameras, verify 6x6 arrangement (max) |

**Expected Results for Each Layout:**
- [ ] Grid cells render in correct rows × columns
- [ ] Each cell is equal size
- [ ] Grid fills available screen space
- [ ] No cell overlap or gaps
- [ ] Responsive to window resize

---

### 3C.2 Switch Grid Layouts with Active Streams

**Test Steps:**
1. Select **2×2** layout
2. Add 4 cameras to all cells (streams playing)
3. Switch to **3×3** layout
4. Switch back to **2×2** layout

**Expected Results:**
- [ ] Layout changes immediately
- [ ] Cameras remain assigned to their cells
- [ ] Streams continue playing without restart
- [ ] New empty cells appear when increasing layout size
- [ ] Grid cell assignments reset when switching layouts (Note: expected behavior - not persisted)

---

### 3C.3 Double-Click Auto-Assignment

**Test Steps:**
1. Select **3×3** layout (9 cells, all empty)
2. Double-click on a camera in the sidebar

**Expected Results:**
- [ ] Camera auto-assigns to Cell 1 (first empty cell)
- [ ] Stream starts playing immediately

3. Double-click on 5 more cameras in sidebar

**Expected Results:**
- [ ] Cameras fill Cells 2, 3, 4, 5, 6 sequentially
- [ ] Each camera goes to next available empty cell
- [ ] If grid is full, camera may append (or show "grid full" message)

---

### 3C.4 Fullscreen Mode

**Test Steps:**
1. Add camera to a grid cell
2. Hover over the cell
3. Click **⛶** (maximize/fullscreen) button

**Expected Results:**
- [ ] Cell expands to fill entire screen
- [ ] Other cells hidden
- [ ] Stream continues playing
- [ ] Close button (×) or Escape key exits fullscreen
- [ ] Grid returns to previous layout after exiting

---

### 3C.5 Clear All Cells

**Test Steps:**
1. Fill multiple grid cells with cameras
2. Click **"Clear All"** button in toolbar
3. Confirm action if prompted

**Expected Results:**
- [ ] All grid cells become empty
- [ ] All streams stop immediately
- [ ] Cells show empty placeholders
- [ ] Cameras remain in sidebar folders (not deleted)
- [ ] Grid accepts new camera drops

---

## 🔍 Phase 3D: Search and Filtering Testing

### 3D.1 Search by Camera Name (English)

**Test Steps:**
1. Type `"Sheikh Zayed"` in search box at top of sidebar
2. Observe results in real-time

**Expected Results:**
- [ ] Folders containing matching cameras auto-expand
- [ ] Only cameras with "Sheikh Zayed" in name are visible
- [ ] Non-matching cameras are hidden
- [ ] Matching text may be highlighted (optional)
- [ ] Clear search (×) button restores all cameras

---

### 3D.2 Search by Camera Name (Arabic)

**Test Steps:**
1. Type `"كاميرا"` (camera in Arabic) in search box
2. Observe results

**Expected Results:**
- [ ] Arabic text input works correctly (RTL support)
- [ ] Cameras with Arabic names containing "كاميرا" appear
- [ ] Arabic text is displayed correctly in results
- [ ] Search is case-insensitive

---

### 3D.3 Search by Camera ID

**Test Steps:**
1. Get a camera ID from camera list (e.g., `cam-001-sheikh-zayed`)
2. Type camera ID in search box
3. Observe results

**Expected Results:**
- [ ] Camera with matching ID appears
- [ ] Folder containing camera auto-expands
- [ ] Only matching camera is visible

---

### 3D.4 Search by Folder Name

**Test Steps:**
1. Type `"Metro"` in search box
2. Observe results

**Expected Results:**
- [ ] "Metro" folder appears in results
- [ ] Folder auto-expands to show cameras
- [ ] All cameras in Metro folder are visible
- [ ] Other folders are hidden

---

### 3D.5 Filter by Source

**Test Steps:**
1. Click **Source Filter** dropdown
2. Select `"Dubai Police"`
3. Observe results

**Expected Results:**
- [ ] Only Dubai Police cameras visible
- [ ] Dubai Police folder auto-expands
- [ ] Other agency folders are hidden
- [ ] Filter can be cleared with "All" option

---

### 3D.6 Filter by Status

**Test Steps:**
1. Click **Status Filter** dropdown
2. Select `"Online"`
3. Observe results

**Expected Results:**
- [ ] Only online cameras visible
- [ ] Offline cameras are hidden
- [ ] Status indicator shows green dot for online
- [ ] Filter can be cleared with "All" option

---

### 3D.7 Combined Search and Filtering

**Test Steps:**
1. Type `"Camera"` in search box
2. Select Source: `"Metro"`
3. Select Status: `"Online"`

**Expected Results:**
- [ ] Only online Metro cameras with "Camera" in name are visible
- [ ] All three filters work together (AND logic)
- [ ] Clearing any filter restores those results
- [ ] Clearing all filters/search shows all cameras

---

## 🔄 Phase 3E: View Modes Testing

### 3E.1 Switch to List View

**Test Steps:**
1. Click **List View** icon (☰) in sidebar header
2. Observe layout change

**Expected Results:**
- [ ] Folder tree collapses to flat list
- [ ] All cameras shown in single scrollable list
- [ ] No folder hierarchy visible
- [ ] Cameras still draggable to grid
- [ ] Search and filters still work
- [ ] View mode preference persists after refresh

---

### 3E.2 Switch Back to Tree View

**Test Steps:**
1. Click **Tree View** icon (🗂️) in sidebar header
2. Observe layout change

**Expected Results:**
- [ ] Folder hierarchy reappears
- [ ] Previously expanded folders remain expanded
- [ ] Cameras organized back into folders
- [ ] Drag-and-drop folder operations available again

---

## 💾 Phase 3F: Data Persistence Testing

### 3F.1 Folder Structure Persistence

**Test Steps:**
1. Create folders: "Test A", "Test B" (subfolder of A)
2. Add 3 cameras to "Test B"
3. Refresh the page (F5)

**Expected Results:**
- [ ] "Test A" and "Test B" folders still exist
- [ ] "Test B" is still subfolder of "Test A"
- [ ] All 3 cameras still assigned to "Test B"
- [ ] Folder names unchanged

---

### 3F.2 Expanded/Collapsed State Persistence

**Test Steps:**
1. Expand "Dubai Police" folder
2. Collapse "Metro" folder
3. Refresh the page

**Expected Results:**
- [ ] "Dubai Police" remains expanded
- [ ] "Metro" remains collapsed
- [ ] Other folders maintain their states

---

### 3F.3 Grid Cell Assignments (Session Only)

**Test Steps:**
1. Add 4 cameras to grid cells
2. Refresh the page

**Expected Results:**
- [ ] Grid cells are empty (not persisted - expected behavior)
- [ ] Cameras still in their folders
- [ ] Must re-add cameras to grid after refresh
- [ ] This is by design (session-only grid state)

---

### 3F.4 View Mode Persistence

**Test Steps:**
1. Switch to List View
2. Refresh the page

**Expected Results:**
- [ ] Dashboard loads in List View (last selected)
- [ ] View mode preference saved in local storage

---

## 🌐 Phase 3G: Browser Compatibility Testing

### 3G.1 Test Drag-and-Drop on Supported Browsers

**Test on each browser:**

| Browser | Min Version | Test Results |
|---------|-------------|--------------|
| Chrome | 90+ | [ ] Drag-and-drop works<br>[ ] Streams play<br>[ ] No console errors |
| Firefox | 88+ | [ ] Drag-and-drop works<br>[ ] Streams play<br>[ ] No console errors |
| Safari | 14+ | [ ] Drag-and-drop works<br>[ ] Streams play<br>[ ] No console errors |
| Edge | 90+ | [ ] Drag-and-drop works<br>[ ] Streams play<br>[ ] No console errors |

**Test Procedure for Each Browser:**
1. Open dashboard
2. Drag camera to folder → Verify works
3. Drag camera to grid → Verify works
4. Drag folder to folder → Verify works
5. Verify stream plays in grid
6. Check browser console for errors

---

### 3G.2 Test on Mobile/Touch Devices (Limited Support)

**Test Steps:**
1. Open dashboard on iPad/Android tablet
2. Try to drag camera with touch gesture

**Expected Results:**
- [ ] Touch drag-and-drop may not work (known limitation)
- [ ] Dashboard layout is responsive
- [ ] Can view streams (no drag-and-drop required)
- [ ] Future enhancement planned for touch support

---

## 🎨 Phase 3H: UI/UX Visual Testing

### 3H.1 Visual Feedback During Drag Operations

**Test Steps:**
1. Drag a camera from sidebar
2. Hover over valid drop target (folder or grid cell)
3. Hover over invalid drop target

**Expected Results:**
- [ ] Dragged item appears semi-transparent
- [ ] Valid drop target highlights in **blue**
- [ ] Invalid target shows no highlight or red border
- [ ] Cursor changes to indicate drag operation
- [ ] Drop indicator (ring or outline) appears

---

### 3H.2 Sidebar Collapse/Expand

**Test Steps:**
1. Click **▼** (minimize) button in sidebar header
2. Sidebar should collapse to thin bar
3. Click **▲** (maximize) button
4. Sidebar should expand

**Expected Results:**
- [ ] Sidebar animates smoothly
- [ ] Grid area expands to fill space when sidebar collapsed
- [ ] Sidebar width returns to normal when expanded
- [ ] Camera tree state preserved during collapse/expand

---

### 3H.3 Context Menu Positioning

**Test Steps:**
1. Right-click on folder near bottom of sidebar
2. Observe context menu position

**Expected Results:**
- [ ] Context menu appears near cursor
- [ ] Menu does not overflow screen bottom (repositions if needed)
- [ ] Menu items are readable and clickable
- [ ] Clicking outside menu closes it

---

### 3H.4 Loading States and Animations

**Test Steps:**
1. Add camera to grid cell
2. Observe loading process

**Expected Results:**
- [ ] Loading spinner appears immediately
- [ ] Spinner is centered in cell
- [ ] Stream replaces spinner when loaded
- [ ] No blank/black cell before stream starts
- [ ] Smooth transition from loading to playing

---

## ⚠️ Phase 3I: Error Handling & Edge Cases

### 3I.1 Offline Camera in Grid

**Test Steps:**
1. Add a camera with status "OFFLINE" to grid cell
2. Observe cell behavior

**Expected Results:**
- [ ] Cell shows offline indicator (red dot or icon)
- [ ] Error message or placeholder appears instead of stream
- [ ] User can remove camera from cell
- [ ] No infinite loading spinner

---

### 3I.2 Network Disconnection During Stream

**Test Steps:**
1. Add camera to grid (stream playing)
2. Disconnect network (disable Wi-Fi or unplug ethernet)
3. Wait 10 seconds
4. Reconnect network

**Expected Results:**
- [ ] Stream shows error or "connection lost" message
- [ ] Stream attempts to reconnect automatically
- [ ] Stream resumes after network restored
- [ ] No crash or frozen UI

---

### 3I.3 Exceed Maximum Grid Cells (36 Cameras)

**Test Steps:**
1. Select **6×6** layout (36 cells max)
2. Fill all 36 cells with cameras
3. Try to add 37th camera via double-click

**Expected Results:**
- [ ] System shows "Grid full" message, or
- [ ] 37th camera replaces oldest/first camera, or
- [ ] Camera is not added (graceful rejection)
- [ ] No console errors

---

### 3I.4 Delete Folder with Cameras in Grid

**Test Steps:**
1. Add 2 cameras from "Test Folder" to grid cells
2. Delete "Test Folder" from sidebar
3. Observe grid cells

**Expected Results:**
- [ ] Cameras remain playing in grid cells
- [ ] Cameras move to parent folder (or "Unorganized") in sidebar
- [ ] Grid cells continue showing streams
- [ ] No broken references or errors

---

### 3I.5 Browser Local Storage Full

**Test Steps:**
1. Fill browser local storage with dummy data (simulate quota exceeded)
2. Try to create new folder

**Expected Results:**
- [ ] Error message: "Storage quota exceeded" or similar
- [ ] Folder is not created
- [ ] Existing folders remain intact
- [ ] Dashboard remains functional

---

## 📊 Phase 3J: Performance Testing

### 3J.1 Large Folder Tree (100+ Cameras)

**Preparation:**
```bash
# Use Go API to create 100 test cameras
for i in {1..100}; do
  curl -X POST http://localhost:8088/api/v1/cameras \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"Test Camera $i\",\"source\":\"TEST\"}"
done
```

**Test Steps:**
1. Open dashboard
2. Observe sidebar load time
3. Expand all folders
4. Search for camera
5. Scroll through camera list

**Expected Results:**
- [ ] Sidebar loads within 2-3 seconds
- [ ] Expand all completes within 1 second
- [ ] Search results appear instantly (<500ms)
- [ ] Smooth scrolling (no lag)
- [ ] Memory usage stays under 200MB

---

### 3J.2 Multiple Simultaneous Streams (16 Cameras)

**Test Steps:**
1. Select **4×4** layout (16 cells)
2. Add 16 cameras to all cells
3. Observe performance

**Expected Results:**
- [ ] All 16 streams play simultaneously
- [ ] No stuttering or frame drops
- [ ] CPU usage under 70%
- [ ] Network bandwidth: ~27 Mbps (16 × ~1.7 Mbps per H.264 stream)
- [ ] Dashboard remains responsive

**Monitor with DevTools:**
```bash
# Open Chrome DevTools → Performance tab
# Record 10 seconds of playback
# Check for:
# - Frame rate: 60 FPS (UI)
# - Long tasks: <50ms
# - Memory: No continuous growth (memory leak)
```

---

### 3J.3 Rapid Folder Operations

**Test Steps:**
1. Create 10 folders rapidly (click + Folder 10 times)
2. Rename 5 folders in quick succession
3. Delete 3 folders quickly

**Expected Results:**
- [ ] All operations complete successfully
- [ ] No race conditions or duplicate IDs
- [ ] Folder counts update correctly
- [ ] No console errors
- [ ] State remains consistent

---

## 🔐 Phase 3K: Security & Permissions Testing

### 3K.1 Admin Permissions

**Test Steps (logged in as Admin):**
1. Try to create folder
2. Try to rename folder
3. Try to delete folder
4. Try to move folder

**Expected Results:**
- [ ] All operations allowed
- [ ] Context menus show all options
- [ ] No permission errors

---

### 3K.2 Operator Permissions (Future - Role-based Access)

**Test Steps (logged in as Operator):**
1. Try to create folder (should be disabled/hidden)
2. Try to rename folder (should be disabled)
3. Try to delete folder (should be disabled)
4. Can view cameras and drag to grid (should work)

**Expected Results:**
- [ ] Folder management options hidden or disabled
- [ ] Can view folder structure (read-only)
- [ ] Can drag cameras to grid
- [ ] Can search and filter
- [ ] Cannot modify folder structure

**Note:** Role-based permissions require JWT token authentication (future enhancement)

---

## ✅ Dashboard Testing Success Checklist

### Folder Management
- [ ] Create root folder
- [ ] Create subfolder
- [ ] Rename folder (context menu)
- [ ] Rename folder (double-click)
- [ ] Delete folder (cameras move to parent)
- [ ] Move folder via drag-and-drop
- [ ] Prevent circular folder references
- [ ] Expand/Collapse individual folders
- [ ] Expand All folders
- [ ] Collapse All folders

### Drag and Drop
- [ ] Drag camera to folder
- [ ] Drag camera between folders
- [ ] Drag camera to grid cell
- [ ] Drag camera to occupied cell (replace)
- [ ] Drag folder to folder (reorganize)
- [ ] Visual feedback (blue highlight on drop target)
- [ ] Drop validation (prevent invalid drops)

### Grid Operations
- [ ] Select all 9 grid layouts (1×1 to 6×6)
- [ ] Assign camera via drag-and-drop
- [ ] Assign camera via double-click (auto-assign)
- [ ] Remove camera from cell (× button)
- [ ] Fullscreen mode (⛶ button)
- [ ] Exit fullscreen (Escape key)
- [ ] Clear all grid cells
- [ ] Switch layouts with active streams

### Search and Filtering
- [ ] Search by camera name (English)
- [ ] Search by camera name (Arabic)
- [ ] Search by camera ID
- [ ] Search by folder name
- [ ] Filter by source (Dubai Police, Metro, etc.)
- [ ] Filter by status (Online, Offline)
- [ ] Combined search + filters (AND logic)
- [ ] Clear search/filters

### View Modes
- [ ] Switch to List View
- [ ] Switch to Tree View
- [ ] View mode persists after refresh

### Persistence
- [ ] Folders persist after refresh
- [ ] Camera assignments persist
- [ ] Expanded/collapsed states persist
- [ ] View mode persists
- [ ] Grid cells reset after refresh (expected - session only)

### Browser Compatibility
- [ ] Chrome 90+ works
- [ ] Firefox 88+ works
- [ ] Safari 14+ works
- [ ] Edge 90+ works

### Performance
- [ ] Large folder tree (100+ cameras) loads smoothly
- [ ] 16 simultaneous streams play without lag
- [ ] Search is responsive (<500ms)
- [ ] No memory leaks during extended use

### Error Handling
- [ ] Offline camera shows error state
- [ ] Network disconnection handled gracefully
- [ ] Full grid (36 cameras) prevents overflow
- [ ] Storage quota errors handled

---

### 3.3 Legacy Grid View (Original Implementation)

### 3.3 Stream via RTSP (Direct Access)

**Test with VLC or ffplay:**
```bash
# Using VLC
vlc rtsp://localhost:8554/dubai_police_test-camera-001

# Using ffplay
ffplay rtsp://localhost:8554/dubai_police_test-camera-001
```

### 3.4 Stream via HLS (Web Browser)

**HLS Stream URL:**
```
http://localhost:8888/dubai_police_test-camera-001/index.m3u8
```

**Test in browser or with curl:**
```bash
curl -I http://localhost:8888/dubai_police_test-camera-001/index.m3u8
```

### 3.5 Stream via WebRTC/WHIP (Ultra Low Latency)

**WHIP (WebRTC HTTP Ingestion Protocol) Architecture:**

The system uses WHIP for camera ingestion, providing ~450ms latency vs 2-4 seconds with HLS.

**Architecture Flow:**
```
Camera (RTSP) → MediaMTX → GStreamer WHIP Pusher → LiveKit WHIP Ingress → LiveKit SFU → Viewers
```

**Key Components:**
- **WHIP Pusher Containers**: Separate Docker containers per camera
- **GStreamer Pipeline**: H.264 passthrough (no transcoding) or H.265→H.264 transcoding
- **LiveKit Ingress**: WHIP endpoint accepting WebRTC push
- **LiveKit SFU**: Selective Forwarding Unit for distribution to viewers

**Testing WHIP Stream:**

1. **Reserve a stream via Go API:**
```bash
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "cam-001-sheikh-zayed",
    "user_id": "test-user",
    "quality": "medium"
  }'
```

2. **Expected Response:**
```json
{
  "reservation_id": "uuid-here",
  "camera_id": "cam-001-sheikh-zayed",
  "camera_name": "Camera 1 - 192.168.1.8",
  "room_name": "camera_cam-001-sheikh-zayed",
  "token": "jwt-token-here",
  "livekit_url": "ws://localhost:7880",
  "expires_at": "2025-10-25T22:00:00Z",
  "quality": "medium"
}
```

3. **Verify WHIP Pusher Container:**
```bash
# Check container is running
docker ps | grep whip-pusher-cam-001-sheikh-zayed

# Check GStreamer pipeline logs
docker logs whip-pusher-cam-001-sheikh-zayed --tail 50
```

4. **Expected Log Output:**
```
Starting WHIP Pusher...
RTSP Source: rtsp://raammohan:Ilove123@192.168.1.8:554/stream1
WHIP Endpoint: http://livekit-ingress:8080/w/<stream-key>
[GStreamer pipeline negotiation logs...]
packets-sent=(guint64)1000+
bitrate=(guint64)1700000+  // ~1.7 Mbps for H.264
```

5. **Check LiveKit Room:**
```bash
# View LiveKit logs for participant connection
docker logs cctv-livekit --tail 30 | grep "camera_cam-001"
```

**WHIP Pusher Technical Details:**
- **Base Image**: Ubuntu 22.04 with GStreamer 1.0 + gst-plugins-rs
- **Pipeline**: `rtspsrc → caps(video) → rtpjitterbuffer → decodebin → x264enc → rtph264pay → whipsink`
- **Codec Support**:
  - H.264: Passthrough (no transcoding)
  - H.265: Transcoded to H.264 for standardization
- **Audio**: Filtered out (video-only streams)
- **Latency**: ~450ms end-to-end

**Multiple Camera Testing:**
```bash
# Reserve Camera 1 (H.264)
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{"camera_id":"cam-001-sheikh-zayed","user_id":"user1","quality":"medium"}'

# Reserve Camera 2 (H.265 → H.264 transcoded)
curl -X POST "http://localhost:8088/api/v1/stream/reserve" \
  -H "Content-Type: application/json" \
  -d '{"camera_id":"cam-002-metro-station","user_id":"user2","quality":"medium"}'

# Verify both containers running
docker ps | grep whip-pusher

# Should show:
# whip-pusher-cam-001-sheikh-zayed
# whip-pusher-cam-002-metro-station
```

**Release Stream:**
```bash
curl -X DELETE "http://localhost:8088/api/v1/stream/release/{reservation_id}"
```

---

## 🎞️ Phase 4: Recording Playback from Milestone

### 4.1 Request Recording Segments from Milestone VMS

**API Endpoint:** `GET http://localhost:8081/vms/recordings/{camera_id}/segments`

**Get available recording segments:**
```bash
# Direct access
curl "http://localhost:8081/vms/recordings/{camera_id}/segments?start=2025-10-24T00:00:00Z&end=2025-10-25T23:59:59Z"

# Via Kong Gateway
curl "http://localhost:8000/vms/recordings/{camera_id}/segments?start=2025-10-24T00:00:00Z&end=2025-10-25T23:59:59Z"
```

**Expected Response:**
```json
{
  "camera_id": "a14c5b2b-c315-4f68-a87b-dffbfb60917b",
  "segments": [
    {
      "start_time": "2025-10-24T10:00:00Z",
      "end_time": "2025-10-24T11:00:00Z",
      "duration_seconds": 3600,
      "recording_server": "milestone.rta.ae:554"
    }
  ],
  "total_segments": 1
}
```

### 4.2 Stream Recorded Video via Playback Service

**API Endpoint:** `GET http://localhost:8092/playback` (Note: Port changed from 8090 to 8092)

**Stream a recording:**
```bash
curl "http://localhost:8092/playback?camera_id={camera_id}&start=2025-10-24T10:00:00Z&duration=3600" \
  -H "Accept: video/mp4"
```

### 4.3 Use Dashboard Playback View

1. **Navigate to Playback View**
   - Click "Playback" in the navigation menu

2. **Select Camera & Time Range**
   - Choose camera from dropdown
   - Select date and time range using the timeline picker

3. **Control Playback**
   - Play/Pause button
   - Timeline scrubbing
   - Speed controls (1x, 2x, 4x, 8x)
   - Frame-by-frame stepping

### 4.4 Export Recordings

**API Endpoint:** `POST http://localhost:8081/vms/recordings/export` (Note: Export handled by VMS service)

```bash
# Direct access
curl -X POST http://localhost:8081/vms/recordings/export \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "{camera_id}",
    "start_time": "2025-10-24T10:00:00Z",
    "end_time": "2025-10-24T10:30:00Z",
    "format": "mp4"
  }'

# Via Kong Gateway
curl -X POST http://localhost:8000/vms/recordings/export \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "{camera_id}",
    "start_time": "2025-10-24T10:00:00Z",
    "end_time": "2025-10-24T10:30:00Z",
    "format": "mp4"
  }'
```

**Expected Response:**
```json
{
  "export_id": "exp-12345",
  "status": "processing",
  "camera_id": "{camera_id}",
  "estimated_completion": "2025-10-24T10:35:00Z"
}
```

**Check export status:**
```bash
# Direct access
curl http://localhost:8081/vms/recordings/export/{export_id}

# Via Kong Gateway
curl http://localhost:8000/vms/recordings/export/{export_id}
```

**Download exported video (from MinIO):**
```bash
# Access via MinIO S3 API
curl -O http://localhost:9000/cctv-recordings/exports/{export_id}.mp4
```

---

## 🔍 Phase 5: Advanced Testing

### 5.1 Stream Counter & Quota Management

**Check current stream statistics:**
```bash
# Direct access
curl http://localhost:8087/api/v1/stream/stats

# Via Kong Gateway
curl http://localhost:8000/api/v1/stream/stats
```

**Expected Response:**
```json
{
  "stats": [
    {
      "source": "DUBAI_POLICE",
      "current": 0,
      "limit": 50,
      "percentage": 0,
      "available": 50
    },
    {
      "source": "METRO",
      "current": 0,
      "limit": 30,
      "percentage": 0,
      "available": 30
    },
    {
      "source": "BUS",
      "current": 0,
      "limit": 20,
      "percentage": 0,
      "available": 20
    },
    {
      "source": "OTHER",
      "current": 0,
      "limit": 400,
      "percentage": 0,
      "available": 400
    }
  ],
  "total": {
    "current": 0,
    "limit": 500,
    "percentage": 0,
    "available": 500
  },
  "timestamp": "2025-10-25T..."
}
```

**Reserve a new stream:**
```bash
# Direct access
curl -X POST http://localhost:8087/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "cameraID": "{camera_id}",
    "userID": "user-123",
    "source": "DUBAI_POLICE",
    "duration": 3600
  }'

# Via Kong Gateway
curl -X POST http://localhost:8000/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "cameraID": "{camera_id}",
    "userID": "user-123",
    "source": "DUBAI_POLICE",
    "duration": 3600
  }'
```

**Expected Response:**
```json
{
  "reservationID": "res-abc123",
  "cameraID": "{camera_id}",
  "expiresAt": "2025-10-25T11:00:00Z",
  "currentUsage": {
    "source": "DUBAI_POLICE",
    "current": 1,
    "limit": 50
  }
}
```

**Release a stream:**
```bash
# Direct access
curl -X DELETE http://localhost:8087/api/v1/stream/release/{reservation_id}

# Via Kong Gateway
curl -X DELETE http://localhost:8000/api/v1/stream/release/{reservation_id}
```

**Send heartbeat to keep reservation alive:**
```bash
# Direct access
curl -X POST http://localhost:8087/api/v1/stream/heartbeat/{reservation_id}

# Via Kong Gateway
curl -X POST http://localhost:8000/api/v1/stream/heartbeat/{reservation_id}
```

### 5.2 Monitoring & Metrics

**Access Grafana Dashboard:**
```
http://localhost:3001
```

**Default Credentials:**
- Username: `admin`
- Password: `admin_changeme`

**Key Dashboards:**
- RTA CCTV Overview
- Stream Performance Metrics
- Storage Utilization
- Network Bandwidth

**Access Prometheus:**
```
http://localhost:9090
```

**Example Queries:**
- Total active streams: `sum(mediamtx_paths_active)`
- CPU usage: `rate(container_cpu_usage_seconds_total[5m])`
- Memory usage: `container_memory_usage_bytes`

### 5.3 Storage Management (MinIO)

**Access MinIO Console:**
```
http://localhost:9001
```

**Default Credentials:**
- Username: `admin`
- Password: `changeme_minio`

**Verify Buckets:**
- `cctv-recordings` - Long-term recordings (90-day retention)
- `cctv-exports` - Exported clips (7-day retention)
- `cctv-thumbnails` - Preview thumbnails (30-day retention)
- `cctv-clips` - Saved clips (manual deletion)

---

## 🧪 Phase 6: End-to-End Testing Scenario

### Complete Workflow Test:

```bash
# 1. Get list of available cameras from Milestone
curl http://localhost:8081/vms/cameras

# 2. Get details of a specific camera (use a camera ID from step 1)
CAMERA_ID="a14c5b2b-c315-4f68-a87b-dffbfb60917b"
curl "http://localhost:8081/vms/cameras/${CAMERA_ID}"

# 3. Check stream quota/statistics
curl http://localhost:8087/api/v1/stream/stats

# 4. Reserve stream access
curl -X POST http://localhost:8087/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d "{
    \"cameraID\": \"${CAMERA_ID}\",
    \"userID\": \"test-user\",
    \"source\": \"DUBAI_POLICE\",
    \"duration\": 3600
  }"

# Save the reservationID from response
RESERVATION_ID="<reservation_id_from_response>"

# 5. View live stream in browser
# Open: http://localhost:3000

# 6. Query available recording segments
curl "http://localhost:8081/vms/recordings/${CAMERA_ID}/segments?start=2025-10-24T00:00:00Z&end=2025-10-25T23:59:59Z"

# 7. Export a recording clip
curl -X POST http://localhost:8081/vms/recordings/export \
  -H "Content-Type: application/json" \
  -d "{
    \"camera_id\": \"${CAMERA_ID}\",
    \"start_time\": \"2025-10-24T10:00:00Z\",
    \"end_time\": \"2025-10-24T10:05:00Z\",
    \"format\": \"mp4\"
  }"

# Save export_id from response
EXPORT_ID="<export_id_from_response>"

# 8. Check export status
curl "http://localhost:8081/vms/recordings/export/${EXPORT_ID}"

# 9. Send heartbeat to keep stream reservation alive
curl -X POST "http://localhost:8087/api/v1/stream/heartbeat/${RESERVATION_ID}"

# 10. Release stream when done
curl -X DELETE "http://localhost:8087/api/v1/stream/release/${RESERVATION_ID}"
```

---

## ✅ Success Criteria Checklist

- [ ] All 25+ services are running and healthy
- [ ] Can register cameras via VMS API
- [ ] Dashboard loads successfully at http://localhost:3000
- [ ] Live streams display in grid view (2x2, 3x3, 4x4)
- [ ] Can switch between cameras in grid
- [ ] RTSP streams accessible via MediaMTX
- [ ] HLS streams work in web browser
- [ ] WebRTC streams provide low-latency playback
- [ ] Can query recordings from Milestone
- [ ] Playback timeline shows available recordings
- [ ] Can export video clips
- [ ] Stream quota management works correctly
- [ ] Grafana dashboards show metrics
- [ ] MinIO storage buckets accessible
- [ ] No stream quota violations occur

---

## 🐛 Troubleshooting

### Issue: Dashboard not loading
```bash
docker logs cctv-dashboard --tail 50
docker logs cctv-go-api --tail 50
```

### Issue: Streams not playing
```bash
docker logs cctv-mediamtx --tail 50
docker logs cctv-livekit --tail 50
docker logs cctv-livekit-ingress --tail 50
```

### Issue: WHIP Pusher Container Failing

**Symptoms:**
- Container restarting constantly
- No video in LiveKit room
- "no element whipsink" error

**Diagnosis:**
```bash
# Check container status
docker ps -a | grep whip-pusher

# View logs
docker logs whip-pusher-cam-<camera-id> --tail 100

# Check if container is restarting
docker inspect whip-pusher-cam-<camera-id> | grep RestartCount
```

**Common Issues:**

1. **Missing gst-plugins-rs**
```bash
# Verify whipsink element exists in image
docker run --rm --entrypoint sh whip-pusher:latest -c "gst-inspect-1.0 | grep whipsink"
# Should output: webrtchttp:  whipsink: WHIP Sink Bin
```

2. **RTSP URL Not Reachable**
```bash
# Test RTSP connection from pusher container
docker run --rm --network cns_cctv-network --entrypoint sh whip-pusher:latest \
  -c "timeout 5 gst-launch-1.0 rtspsrc location='rtsp://192.168.1.8:554/stream1' ! fakesink"
```

3. **LiveKit Ingress Not Ready**
```bash
# Check ingress health
docker logs cctv-livekit-ingress --tail 30
curl http://localhost:8080/  # Should respond
```

4. **Codec Mismatch (H.265 Camera)**
- Camera 2 uses H.265 and requires transcoding
- Check pipeline includes: `decodebin ! x264enc`
- Verify logs show transcoding activity

5. **Audio Stream Interference**
```bash
# Check if pipeline has video-only caps filter
docker logs whip-pusher-cam-<camera-id> | grep "application/x-rtp,media=video"
```

**Solutions:**

1. **Rebuild WHIP Pusher Image:**
```bash
cd services/whip-pusher
docker build -t whip-pusher:latest .
```

2. **Manually Stop/Remove Failed Container:**
```bash
docker stop whip-pusher-cam-<camera-id>
docker rm whip-pusher-cam-<camera-id>
```

3. **Check Network Connectivity:**
```bash
# Verify Docker network exists
docker network ls | grep cns_cctv-network

# Check container can reach LiveKit ingress
docker exec whip-pusher-cam-<camera-id> ping -c 3 livekit-ingress
```

### Issue: Both Cameras Showing Same Feed

**Root Cause:** Each camera must have its own unique WHIP pusher container with different RTSP URLs.

**Verification:**
```bash
# Check each container has different RTSP URL
docker inspect whip-pusher-cam-001-sheikh-zayed | grep RTSP_URL
docker inspect whip-pusher-cam-002-metro-station | grep RTSP_URL

# Should show different IPs:
# Camera 1: rtsp://...@192.168.1.8:554/...
# Camera 2: rtsp://...@192.168.1.13:554/...
```

**Solution:**
- Ensure VMS service returns correct camera details
- Verify MediaMTX path configuration is per-camera
- Check go-api spawns separate containers per camera ID

### Issue: Recordings not accessible
```bash
docker logs cctv-vms-service --tail 50
docker logs cctv-playback-service --tail 50
```

### Check service health
```bash
docker ps --filter "name=cctv-" --format "table {{.Names}}\t{{.Status}}"
```

### View service logs in real-time
```bash
# Follow logs for a specific service
docker logs -f cctv-<service-name>

# View last 100 lines
docker logs cctv-<service-name> --tail 100
```

### Restart a specific service
```bash
docker-compose restart <service-name>
```

### Rebuild a service after code changes
```bash
docker-compose build <service-name>
docker-compose up -d <service-name>
```

---

## 📊 Service Architecture

### Core Services:
1. **VMS Service** - Integrates with Milestone VMS
2. **Stream Counter** - Manages quota for 500 concurrent streams
3. **MediaMTX** - RTSP server for stream distribution
4. **LiveKit** - WebRTC SFU for ultra-low latency
5. **Recording Service** - Handles recording to MinIO
6. **Playback Service** - Retrieves and streams recordings
7. **Storage Service** - Manages MinIO storage
8. **Metadata Service** - Camera metadata and indexing

### Supporting Services:
- **Kong** - API Gateway
- **PostgreSQL** - Metadata database
- **Valkey** - Redis-compatible cache
- **MinIO** - S3-compatible object storage
- **Prometheus/Grafana** - Monitoring and visualization
- **Loki** - Log aggregation

---

## 🔐 Default Credentials

**Grafana:**
- URL: http://localhost:3001
- Username: `admin`
- Password: `admin_changeme`

**MinIO Console:**
- URL: http://localhost:9001
- Username: `admin`
- Password: `changeme_minio`

**PostgreSQL:**
- Host: localhost:5432
- Database: `cctv`
- Username: `cctv`
- Password: `changeme_db`

**Valkey (Redis):**
- Host: localhost:6379
- No password (development mode)

---

## 📝 Notes

1. **Production Deployment**: Change all default passwords before deploying to production
2. **Milestone Integration**: Update `rtsp_url` in camera registration to point to your actual Milestone VMS server
3. **Network Configuration**: Ensure firewall rules allow access to required ports
4. **Storage Capacity**: Monitor MinIO storage usage as recordings accumulate
5. **Stream Limits**: The system is configured for 500 concurrent streams. Adjust quotas in stream-counter service if needed
6. **Security**: All services are currently configured for development. Enable authentication and SSL/TLS for production

---

## 🚀 Quick Start Commands

```bash
# Start all services
docker-compose up -d

# Check service status
docker ps --filter "name=cctv-"

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v

# Rebuild specific service
docker-compose build <service-name>
docker-compose up -d <service-name>
```

---

## 📞 Support

For issues or questions:
1. Check service logs using `docker logs cctv-<service-name>`
2. Review the troubleshooting section above
3. Verify all services are healthy using health check commands
4. Check network connectivity between services

---

**Last Updated:** 2025-10-25
**Version:** 1.0.0
