# PTZ Controls - Detailed Requirements

**Last Updated**: 2025-10-26
**Status**: ✅ Implemented
**Component**: `dashboard/src/components/PTZControls.tsx`

---

## Overview

PTZ (Pan-Tilt-Zoom) controls provide operators with an intuitive interface to remotely control camera positioning and zoom. The controls are designed as a compact vertical sidebar that appears on-demand without obstructing the video feed or overlays.

---

## Layout & Position

### Sidebar Design
- **Position**: Left edge of video player
- **Layout**: Single vertical column with all controls stacked
- **Design Philosophy**: Compact and non-intrusive
- **Z-index**: 10 (above video, below modals)

### Responsive Sizing

Controls automatically resize based on grid cell size:

| Cell Type | Container Width | Button Size | Icon Size | Use Case |
|-----------|----------------|-------------|-----------|----------|
| Hotspot/Fullscreen | 64px | 44px × 44px | 20px | Maximum detail view |
| Large (4×4+) | 56px | 36px × 36px | 16px | Large control rooms |
| Medium (3×3) | 48px | 32px × 32px | 14px | Default balanced view |
| Small (2×2) | 40px | 24px × 24px | 12px | Compact monitoring |

### Non-Obstruction Requirements
✅ **Camera name overlay** (top-left) remains visible
✅ **LIVE indicator** (bottom-right) remains visible
✅ **Fullscreen button** (top-right) remains accessible
✅ **Close button** (top-right) remains accessible

---

## Control Layout

### Vertical Stack Order (Top to Bottom)

```
┌─────────────┐
│      ↑      │  1. Tilt Up
│      ←      │  2. Pan Left
│      ⌂      │  3. Home Position (blue highlight)
│      →      │  4. Pan Right
│      ↓      │  5. Tilt Down
│ ─────────── │  6. Divider
│      +      │  7. Zoom In
│      -      │  8. Zoom Out
│             │  9. Spacer (flex-1)
│      📌     │ 10. Pin/Unpin Button
└─────────────┘
```

### Control Details

**Directional Controls**:
- Up/Down: Tilt camera vertically
- Left/Right: Pan camera horizontally
- Home: Return to preset home position (highlighted in primary blue)

**Zoom Controls**:
- Zoom In (+): Increase camera zoom
- Zoom Out (-): Decrease camera zoom

**Pin Button**:
- Icon changes based on state:
  - Unpinned: `PinOff` icon (white/translucent background)
  - Pinned: `Pin` icon (primary blue background)
- Tooltip adapts: "Pin Controls" or "Unpin Controls"

---

## Interaction Behavior

### State Machine

```
┌─────────────┐
│   Hidden    │ ◄─┐
└──────┬──────┘   │
       │ Hover    │ Mouse Leave (unpinned)
       ▼          │
┌─────────────┐   │
│   Visible   │───┘
│ (Unpinned)  │
└──────┬──────┘
       │ Click Video / Click Pin
       ▼
┌─────────────┐
│   Visible   │
│  (Pinned)   │ ◄─┐
└──────┬──────┘   │ Click Pin
       │          │
       └──────────┘
       │ Click Video / Click Pin (toggle off)
       ▼
    Hidden
```

### User Actions

| Action | Condition | Result |
|--------|-----------|--------|
| Hover over video | PTZ enabled, not pinned | Sidebar appears |
| Mouse leaves video | PTZ enabled, not pinned | Sidebar disappears |
| Click video canvas | PTZ enabled, not pinned | Sidebar pins (stays visible) |
| Click video canvas | PTZ enabled, pinned | Sidebar unpins and hides |
| Click Pin button | Sidebar visible | Toggles pin state |
| Click PTZ control | Any | Command sent, event doesn't bubble |

### PTZ Command Execution

**Continuous Movement** (Pan/Tilt/Zoom):
- `onMouseDown`: Start movement command
- `onMouseUp`: Stop movement
- `onMouseLeave`: Stop movement (safety)
- Speed parameter: 0.5 (configurable)

**Instant Commands**:
- Home position: Single click executes command

### Event Propagation
- All clicks on PTZ controls call `e.stopPropagation()`
- Prevents sidebar clicks from triggering video canvas click handler
- Ensures pin/unpin behavior works correctly

---

## Visual Design

### Color Scheme

**Sidebar**:
- Background: `bg-black/80` (80% black opacity)
- Backdrop filter: `backdrop-blur-md` (medium blur)
- Border: `border-r border-white/10` (subtle right border)

**Button States**:

| State | Background | Additional | Use Case |
|-------|-----------|------------|----------|
| Normal | `bg-white/10` | - | Default state |
| Hover | `bg-white/20` | - | Mouse over |
| Active | `bg-white/30` | `ring-1 ring-white/50` | Currently pressed |
| Home (special) | `bg-primary-600/80` | - | Home button always highlighted |
| Pin (pinned) | `bg-primary-600/80` | - | Pin button when active |

**Icons**:
- Color: White (`text-white`)
- Size: Responsive based on cell size (see table above)
- Library: Lucide React
  - `ChevronUp`, `ChevronDown`, `ChevronLeft`, `ChevronRight`
  - `ZoomIn`, `ZoomOut`
  - `Home`
  - `Pin`, `PinOff`

### Animations
- All state transitions: `transition-colors` (smooth color changes)
- Opacity transitions: `transition-opacity` for hover hints

---

## Technical Implementation

### Component Structure

**File**: `dashboard/src/components/PTZControls.tsx`

**Props Interface**:
```typescript
interface PTZControlsProps {
  camera: Camera;              // Camera object with PTZ capabilities
  onTogglePin: () => void;     // Callback to toggle pin state
  isPinned: boolean;           // Current pin state
  cellSize?: 'hotspot' | 'large' | 'medium' | 'small'; // Responsive sizing
}
```

**Responsive Size Classes**:
```typescript
const sizeClasses = {
  hotspot: {
    container: 'w-16',   // 64px
    button: 'h-11 w-11', // 44px
    icon: 'w-5 h-5',     // 20px
    gap: 'gap-1.5',
    padding: 'p-2',
  },
  large: {
    container: 'w-14',   // 56px
    button: 'h-9 w-9',   // 36px
    icon: 'w-4 h-4',     // 16px
    gap: 'gap-1',
    padding: 'p-1.5',
  },
  medium: {
    container: 'w-12',   // 48px
    button: 'h-8 w-8',   // 32px
    icon: 'w-3.5 h-3.5', // 14px
    gap: 'gap-1',
    padding: 'p-1',
  },
  small: {
    container: 'w-10',   // 40px
    button: 'h-6 w-6',   // 24px
    icon: 'w-3 h-3',     // 12px
    gap: 'gap-0.5',
    padding: 'p-1',
  },
};
```

### PTZ Commands

**API Endpoint**: `/api/v1/cameras/{id}/ptz`

**Command Types**:
```typescript
type PTZCommand =
  | 'tilt_up'
  | 'tilt_down'
  | 'pan_left'
  | 'pan_right'
  | 'zoom_in'
  | 'zoom_out'
  | 'home';

interface PTZParams {
  speed?: number;      // 0.0 - 1.0 (default: 0.5)
  preset_id?: number;  // For preset commands (not used in current implementation)
}
```

**Command Handler**:
```typescript
const handlePTZCommand = async (
  command: string,
  params?: { speed?: number; preset_id?: number }
) => {
  try {
    await api.controlPTZ(camera.id, command, params);
  } catch (error) {
    console.error('PTZ command failed:', error);
  }
};
```

### Integration with LiveStreamPlayer

**Parent Component**: `dashboard/src/components/LiveStreamPlayer.tsx`

**State Management**:
```typescript
const [showPTZ, setShowPTZ] = useState(false);     // Controls visibility
const [ptzPinned, setPTZPinned] = useState(false); // Pin state
```

**Event Handlers**:
```typescript
// Pin controls (called on first click)
const handlePTZClick = () => {
  setPTZPinned(true);
  setShowPTZ(true);
};

// Toggle pin state (called by Pin button or click-outside-to-unpin)
const handleTogglePin = () => {
  setPTZPinned(!ptzPinned);
  if (ptzPinned) {
    setShowPTZ(false); // Hide when unpinning
  }
};

// Video canvas click handler
const handleVideoClick = () => {
  if (!camera.ptz_enabled) return;

  if (ptzPinned) {
    // Click outside PTZ controls to unpin
    handleTogglePin();
  } else {
    // Click to pin
    handlePTZClick();
  }
};
```

**Rendering**:
```typescript
<div
  onMouseEnter={() => !ptzPinned && camera.ptz_enabled && setShowPTZ(true)}
  onMouseLeave={() => !ptzPinned && setShowPTZ(false)}
  onClick={handleVideoClick}
>
  {/* Video player */}

  {showPTZ && (
    <PTZControls
      camera={camera}
      onTogglePin={handleTogglePin}
      isPinned={ptzPinned}
      cellSize={cellSize}
    />
  )}
</div>
```

---

## Requirements Checklist

### Functional Requirements
- [x] Vertical sidebar layout on left edge
- [x] Single column with all controls stacked
- [x] Responsive sizing based on grid cell size
- [x] Pan/Tilt/Zoom controls with mouse interaction
- [x] Home position button
- [x] Pin/Unpin functionality
- [x] Hover to show (when unpinned)
- [x] Click to pin
- [x] Click-outside-to-unpin
- [x] Event propagation handling

### Visual Requirements
- [x] Semi-transparent black background with blur
- [x] Subtle right border
- [x] White icons with proper sizing
- [x] Button state transitions (normal/hover/active)
- [x] Home button highlighted in primary blue
- [x] Pin button highlighted when pinned
- [x] Smooth color transitions

### Integration Requirements
- [x] Integrated into LiveStreamPlayer component
- [x] State managed in parent component
- [x] Only renders for PTZ-enabled cameras
- [x] Doesn't obstruct camera name overlay
- [x] Doesn't obstruct LIVE indicator
- [x] Doesn't obstruct fullscreen/close buttons

### Performance Requirements
- [x] Minimal re-renders (React.memo considerations)
- [x] Event handlers optimized
- [x] No memory leaks on mount/unmount
- [x] Smooth 60fps animations

---

## Future Enhancements

### Potential Improvements
- [ ] Preset position support (save/recall favorite positions)
- [ ] Keyboard shortcuts (arrow keys for pan/tilt, +/- for zoom)
- [ ] Speed control slider (adjust movement speed)
- [ ] Auto-patrol mode (automated scanning patterns)
- [ ] Tour mode (cycle through preset positions)
- [ ] PTZ limits configuration (restrict movement range)
- [ ] Multi-camera PTZ sync (control multiple cameras simultaneously)

### Accessibility Improvements
- [ ] ARIA labels for screen readers
- [ ] Focus indicators for keyboard navigation
- [ ] Keyboard-only operation support
- [ ] High contrast mode support

---

## Testing Checklist

### Manual Testing
- [x] PTZ controls appear on hover (unpinned state)
- [x] PTZ controls disappear on mouse leave (unpinned state)
- [x] Click video pins controls
- [x] Click video while pinned unpins controls
- [x] Click Pin button toggles pin state
- [x] Click PTZ control doesn't unpin
- [x] All directional commands work (up/down/left/right)
- [x] Zoom commands work (in/out)
- [x] Home command works
- [x] Responsive sizing works across grid layouts
- [x] Camera name overlay not obstructed
- [x] LIVE indicator not obstructed
- [x] Fullscreen/close buttons not obstructed

### Browser Compatibility
- [x] Chrome/Edge (Chromium)
- [x] Firefox
- [ ] Safari (not tested - Windows environment)

### Grid Layout Testing
- [x] 1×1 (hotspot) - Large controls
- [x] 2×2 - Small controls
- [x] 3×3 - Medium controls (default)
- [x] 4×4 - Large controls
- [x] 6×6 - Small controls

---

## Related Documentation

- [Multi-Viewer Streaming](MULTI_VIEWER_STREAMING.md) - Multi-viewer architecture
- [Troubleshooting Multi-Viewer](TROUBLESHOOTING_MULTI_VIEWER.md) - Debugging streaming issues
- [Dashboard Features](phases/PHASE-4-DASHBOARD-COMPLETE.md) - Complete dashboard features
- [Architecture](architecture.md) - System architecture overview

---

## Change History

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-10-26 | 1.0 | Initial implementation - Vertical sidebar PTZ controls | System |

---

**Document Version**: 1.0
**Component Version**: Implemented in dashboard v1.0.0
**API Version**: v1
