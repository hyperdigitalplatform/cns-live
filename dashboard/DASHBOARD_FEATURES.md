# RTA CCTV Dashboard - Features Documentation

## Overview

The RTA CCTV Dashboard provides an advanced camera management interface with tree-based organization, drag-and-drop functionality, and flexible grid layouts for viewing multiple camera streams simultaneously.

**Last Updated**: 2025-10-26

---

## Table of Contents

1. [Camera Organization](#camera-organization)
2. [Folder Management](#folder-management)
3. [Drag and Drop](#drag-and-drop)
4. [Grid View](#grid-view)
5. [Search and Filtering](#search-and-filtering)
6. [User Guide](#user-guide)
7. [Keyboard Shortcuts](#keyboard-shortcuts)

---

## Camera Organization

### Tree-Style Folder Structure

Cameras can be organized in a hierarchical folder structure similar to file systems:

- **Root Folders**: Top-level categories (Dubai Police, Metro, Taxi, etc.)
- **Subfolders**: Unlimited nesting depth for granular organization
- **Unorganized**: Cameras not assigned to any folder

### Default Folders

The system initializes with predefined folders:

| Folder Name | Arabic Name | Purpose |
|-------------|-------------|---------|
| Dubai Police | شرطة دبي | Dubai Police cameras |
| Sharjah Police | شرطة الشارقة | Sharjah Police cameras |
| Metro | المترو | Metro station cameras |
| Taxi | التاكسي | Taxi monitoring cameras |
| Parking | مواقف السيارات | Parking area cameras |
| Unorganized | غير منظم | Cameras not in folders |

---

## Folder Management

### Creating Folders

**Create Root Folder:**
1. Click the **+ Folder** button in sidebar header
2. Enter folder name (English)
3. Optionally enter Arabic name
4. Folder appears at root level

**Create Subfolder:**
1. Right-click on a parent folder
2. Select "Add Subfolder"
3. Enter subfolder name
4. Subfolder appears nested under parent

### Renaming Folders

**Method 1: Context Menu**
1. Right-click folder
2. Select "Rename"
3. Edit name inline
4. Press Enter to save or Escape to cancel

**Method 2: Double-click (Admin only)**
1. Double-click folder name
2. Edit directly
3. Click outside or press Enter to save

### Deleting Folders

1. Right-click folder
2. Select "Delete"
3. Confirm deletion
4. **Note**: Cameras in folder move to parent folder
5. **Note**: Subfolders move to parent folder

### Moving Folders

**Drag and Drop:**
1. Click and hold folder
2. Drag to new parent folder
3. Drop to move
4. **Note**: Cannot create circular references

---

## Drag and Drop

### Drag Camera to Folder

**Purpose**: Organize cameras into folders

**Steps**:
1. Click and drag camera from:
   - Camera list
   - Another folder
   - Grid cell
2. Hover over target folder (folder highlights in blue)
3. Release to drop
4. Camera moves to new folder

### Drag Camera to Grid Cell

**Purpose**: Display camera in grid

**Steps**:
1. Click and drag camera from sidebar
2. Hover over empty grid cell (cell highlights)
3. Release to drop
4. Camera stream begins playing

### Drag Folder to Folder

**Purpose**: Reorganize folder hierarchy

**Steps**:
1. Click and drag folder
2. Hover over target parent folder
3. Release to drop
4. Folder becomes subfolder of target

### Visual Feedback

- **Dragging**: Item appears semi-transparent
- **Valid Drop**: Target highlights in blue
- **Invalid Drop**: Red border or no highlight
- **Drop Indicator**: Blue ring around target

---

## Grid View

### Grid Layouts

Available layouts for multi-camera viewing:

| Layout | Cells | Best For |
|--------|-------|----------|
| 1×1 | 1 | Single camera focus |
| 2×2 | 4 | Small control room |
| 3×3 | 9 | **Default** - Balanced view |
| 4×4 | 16 | Medium control room |
| 2×3 | 6 | Portrait displays |
| 3×4 | 12 | Widescreen displays |
| 4×5 | 20 | Large control room |
| 5×5 | 25 | Command center |
| 6×6 | 36 | Operations center |

### Assigning Cameras to Grid

**Method 1: Drag and Drop**
1. Drag camera from sidebar
2. Drop onto grid cell
3. Stream starts automatically

**Method 2: Double-Click**
1. Double-click camera in sidebar
2. Camera auto-assigns to next empty cell
3. If grid full, camera added to end

### Managing Grid Cells

**Remove Camera from Cell:**
- Hover over cell
- Click **×** button in top-right
- Cell becomes empty

**Fullscreen Mode:**
- Hover over cell
- Click **⛶** (maximize) button
- Press Escape or click **×** to exit

**Clear All Cells:**
- Click "Clear All" in toolbar
- Confirm deletion
- All cells become empty

### Cell Information

Each cell displays:
- Camera name (English)
- Camera name (Arabic, if available)
- Online/offline status indicator
- Cell number (on hover)

---

## Search and Filtering

### Global Search

**Search Box** (top of sidebar):
- Search by camera name (English or Arabic)
- Search by camera ID
- Search by folder name
- **Real-time filtering** as you type

**Search Behavior**:
- Folders with matching cameras automatically expand
- Matching cameras highlighted
- Non-matching items hidden

### Filters

**Source Filter:**
- Filter by agency/department
- Options: Dubai Police, Sharjah Police, Metro, Taxi, Parking, All

**Status Filter:**
- Filter by camera status
- Options: Online, Offline, Maintenance, Error, All

**Combined Filtering:**
- All filters work together (AND logic)
- Example: Dubai Police + Online = Only online Dubai Police cameras

### Filter Actions

**Expand All:**
- Expands all folders in tree
- Shows all cameras
- Useful after search

**Collapse All:**
- Collapses all folders
- Shows only root folders
- Useful for navigation

---

## User Guide

### Quick Start

**1. Initial Setup**
```
Dashboard loads → Default folders created → Cameras fetch from API
```

**2. Organize Cameras**
```
Create folders → Drag cameras into folders → Rename/organize as needed
```

**3. View Streams**
```
Select grid layout → Drag/drop cameras to cells → Streams start playing
```

### Common Workflows

#### Workflow 1: Create Department Structure

```
1. Create "Traffic Department" folder
2. Create subfolders: "Highway", "Intersection", "Tunnel"
3. Drag cameras from sidebar to appropriate subfolders
4. Rename folders as needed
```

#### Workflow 2: Set Up Multi-Camera View

```
1. Select 4×4 layout (16 cells)
2. Expand "Dubai Police" folder
3. Double-click first camera → Auto-assigns to Cell 1
4. Double-click next cameras → Auto-fill remaining cells
5. OR drag specific cameras to specific cells
```

#### Workflow 3: Quick Camera Search

```
1. Type "metro" in search box
2. Metro folder auto-expands
3. All metro cameras visible
4. Double-click camera to add to grid
5. Clear search to see all cameras again
```

---

## View Modes

### Tree View (Default)

**Features**:
- Hierarchical folder structure
- Expandable/collapsible folders
- Drag-and-drop organization
- Visual hierarchy with indentation

**Best For**:
- Large camera deployments (100+ cameras)
- Organized camera management
- Department-based access

### List View

**Features**:
- Flat list of all cameras
- Faster scrolling
- Simpler interface
- Still supports drag-to-grid

**Best For**:
- Small deployments (<50 cameras)
- Quick camera selection
- Simple operations

**Switch Views:**
- Click grid icon (🗂️) for tree view
- Click list icon (☰) for list view

---

## Sidebar Features

### Collapsible Sidebar

**Collapse:**
- Click **▼** (minimize) button
- Sidebar collapses to thin bar
- More screen space for grid

**Expand:**
- Click **▲** (maximize) button
- Sidebar expands to full width

### Sidebar Stats

**Footer displays**:
- Total cameras visible: `X cameras`
- Total folders: `Y folders`
- Updates based on search/filters

---

## Keyboard Shortcuts

### Navigation

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + F` | Focus search box |
| `Escape` | Clear search / Exit fullscreen |
| `Ctrl/Cmd + E` | Expand all folders |
| `Ctrl/Cmd + Shift + E` | Collapse all folders |

### Grid Control

| Shortcut | Action |
|----------|--------|
| `1` | Switch to 1×1 layout |
| `2` | Switch to 2×2 layout |
| `3` | Switch to 3×3 layout |
| `4` | Switch to 4×4 layout |
| `Ctrl/Cmd + Enter` | Fullscreen selected camera |
| `Ctrl/Cmd + Shift + C` | Clear all grid cells |

### Folder Management (Admin)

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + N` | Create new folder |
| `F2` | Rename selected folder |
| `Delete` | Delete selected folder |

---

## Context Menus

### Folder Context Menu

**Right-click folder to access**:
- 📁 **Add Subfolder** - Create child folder
- ✏️ **Rename** - Edit folder name
- 🗑️ **Delete** - Remove folder (cameras move to parent)

### Camera Context Menu

**Right-click camera to access**:
- 🗑️ **Remove from Folder** - Remove camera assignment
- (More options in future updates)

---

## Data Persistence

### Local Storage

The following data is saved in browser:
- ✅ Folder structure
- ✅ Folder names (English & Arabic)
- ✅ Camera-to-folder assignments
- ✅ Expanded/collapsed folder states
- ✅ Sidebar view mode (tree/list)

### Session Storage

The following data is temporary:
- ❌ Grid cell assignments
- ❌ Selected cameras
- ❌ Search queries
- ❌ Filter selections

**Note**: Refresh page to reset grid but keep folder structure.

---

## Admin vs Operator Permissions

### Admin Permissions

Admins can:
- ✅ Create folders
- ✅ Rename folders
- ✅ Delete folders
- ✅ Move folders
- ✅ Organize cameras
- ✅ All operator permissions

### Operator Permissions

Operators can:
- ✅ View folder structure
- ✅ Search cameras
- ✅ Drag cameras to grid
- ✅ View streams
- ❌ Cannot modify folders

**Permission Check**: Based on user role from JWT token

---

## Performance Optimization

### Large Deployments

**For 500+ cameras**:
- Tree view uses virtualization
- Cameras load on-demand
- Folders lazy-load children
- Streams only load for visible cells

### Best Practices

1. **Organize into folders**: Improves performance and navigation
2. **Use search**: Faster than scrolling
3. **Collapse unused folders**: Reduces DOM nodes
4. **Limit grid size**: Use appropriate layout for monitor size
5. **Close unused streams**: Remove from grid when not needed

---

## Troubleshooting

### Camera Not Appearing in Folder

**Problem**: Dragged camera to folder but it's not showing

**Solutions**:
1. Check if folder is expanded
2. Check if search filter is active
3. Check if status filter excludes camera
4. Refresh page and try again

### Drag and Drop Not Working

**Problem**: Cannot drag cameras or folders

**Solutions**:
1. Ensure camera/folder is not being edited
2. Check browser compatibility (modern browsers only)
3. Try reloading page
4. Check console for JavaScript errors

### Grid Not Updating

**Problem**: Dropped camera but grid cell is empty

**Solutions**:
1. Check if camera is online
2. Wait for stream to load (may take 2-5 seconds)
3. Check network connection
4. Check browser console for errors

### Folders Lost After Refresh

**Problem**: Created folders disappeared after page reload

**Solutions**:
1. Check browser local storage not disabled
2. Check incognito/private mode (doesn't persist)
3. Check browser storage quota
4. Try different browser

---

## API Integration

### Folder Storage API (Future)

Folders will sync with backend API:

```typescript
// Create folder
POST /api/v1/folders
Body: {
  name: "Traffic Cameras",
  name_ar: "كاميرات المرور",
  parent_id: null,
  camera_ids: ["cam-001", "cam-002"]
}

// Update folder
PUT /api/v1/folders/{id}
Body: {
  name: "Updated Name",
  camera_ids: ["cam-001", "cam-003"]
}

// Delete folder
DELETE /api/v1/folders/{id}

// Get folder tree
GET /api/v1/folders/tree
```

**Current**: Local storage only (browser)
**Future**: API sync with database

---

## Future Enhancements

### Planned Features

1. **Shared Folder Views**
   - Save folder layouts
   - Share with team members
   - Department-specific views

2. **Advanced Grid**
   - Hotspot layouts (1 large + multiple small)
   - Custom cell spanning
   - Grid templates/presets

3. **Camera Groups**
   - Bulk operations
   - Group streaming
   - Sequential viewing

4. **Smart Organization**
   - Auto-organize by location
   - Auto-organize by agency
   - AI-suggested folders

5. **Mobile Support**
   - Touch-optimized drag-and-drop
   - Responsive grid layouts
   - Mobile-friendly tree view

---

## Component Architecture

### React Components

```
LiveViewEnhanced (Page)
├── CameraSidebarNew
│   ├── CameraTreeView
│   │   ├── FolderNode (recursive)
│   │   └── CameraNode
│   └── SearchBar
└── StreamGridEnhanced
    └── GridCell[]
```

### State Management

```
Zustand Stores:
├── cameraStore
│   ├── cameras: Camera[]
│   ├── fetchCameras()
│   └── selectCamera()
├── folderStore
│   ├── folders: CameraFolder[]
│   ├── expandedFolders: Set<string>
│   ├── createFolder()
│   ├── updateFolder()
│   ├── deleteFolder()
│   ├── moveCameraBetweenFolders()
│   └── buildFolderTree()
└── streamStore
    ├── activeStreams: Map
    └── reserveStream()
```

### Data Flow

```
User Action → Component Event Handler → Store Action → State Update → UI Re-render
```

---

## Browser Compatibility

| Browser | Version | Support |
|---------|---------|---------|
| Chrome | 90+ | ✅ Full |
| Firefox | 88+ | ✅ Full |
| Safari | 14+ | ✅ Full |
| Edge | 90+ | ✅ Full |
| Opera | 76+ | ✅ Full |
| IE 11 | - | ❌ Not supported |

**Required Features**:
- HTML5 Drag and Drop API
- ES6+ JavaScript
- CSS Grid
- Local Storage API
- WebRTC (for streams)

---

## Accessibility

### Keyboard Navigation

- ✅ All actions accessible via keyboard
- ✅ Tab navigation through tree
- ✅ Arrow keys for folder navigation
- ✅ Enter/Space to select
- ✅ Context menu via keyboard

### Screen Readers

- ✅ ARIA labels on all interactive elements
- ✅ Semantic HTML structure
- ✅ Focus indicators
- ✅ Descriptive button text

### Visual

- ✅ High contrast mode support
- ✅ Customizable text size
- ✅ Color-blind friendly status indicators

---

## Development

### Running Locally

```bash
cd dashboard
npm install
npm run dev
```

### Building for Production

```bash
npm run build
npm run preview  # Test build
```

### Environment Variables

```bash
# .env
VITE_API_URL=http://localhost:8088
VITE_LIVEKIT_URL=ws://localhost:7880
```

---

## Support

For issues or feature requests:
- **Email**: support@rta.ae
- **GitHub**: https://github.com/rta/cctv-dashboard/issues
- **Documentation**: https://docs.rta.ae/cctv

---

**Document Version**: 1.0
**Last Updated**: 2025-10-26
**Author**: RTA CCTV Development Team
