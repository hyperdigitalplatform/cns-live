# Layout Management - Feature Documentation

**Last Updated**: 2025-10-27

---

## Overview

The Layout Management system allows users to save, load, and manage custom camera grid configurations. Each layout preserves the exact grid type and camera positions for quick recall.

### Key Features

- **Save Custom Layouts**: Save current camera arrangements with descriptive names
- **Load Layouts**: Instantly restore saved camera configurations
- **Manage Layouts**: Edit, delete, and organize saved layouts
- **Layout Scopes**: Choose between personal (local) or shared (global) layouts
- **Automatic Grid Switching**: System automatically switches to the correct grid when loading a layout

---

## Saving Layouts

### How to Save a Layout

1. Arrange cameras in desired grid positions
2. Select appropriate grid layout (2×2, 3×3, 9-way-1-hotspot, etc.)
3. Click **"Save Layout"** button in toolbar
4. Enter layout details:
   - **Name** (required): Descriptive name for the layout
   - **Description** (optional): Additional notes about the layout
   - **Scope**: Choose layout visibility
     - **Local (Personal)**: Only visible to you
     - **Global**: Visible to all users
5. Click **"Save"**

### What Gets Saved

- ✅ Grid layout type (e.g., "3×3", "9-way-1-hotspot")
- ✅ Camera IDs and positions
- ✅ Layout type (Standard or Hotspot)
- ✅ Created by user
- ✅ Creation and modification timestamps

### Example Use Cases

- "Control Room - Day Shift" (3×4 grid with primary traffic cameras)
- "Metro Monitoring" (9-way-1-hotspot with station entrances)
- "Emergency Response" (4×4 grid with critical incident cameras)

---

## Loading Layouts

### How to Load a Layout

1. Click **"Load Layout"** dropdown in toolbar
2. Browse saved layouts grouped by type:
   - **Standard Layouts**: Regular grid configurations
   - **Hotspot Layouts**: Priority view configurations
3. Click on a layout to load it
4. System automatically:
   - Switches to the saved grid type
   - Clears current grid
   - Loads cameras in exact saved positions
   - Starts camera streams

### Layout Information Display

- Layout name and description
- Grid type (Standard/Hotspot)
- Number of cameras
- Scope indicator (Global/Personal)
- Last updated timestamp

### Loading Behavior

```
User selects layout → Grid switches to saved type → Cameras load in positions → Streams start
```

---

## Managing Layouts

### How to Manage Layouts

1. Click **"Manage"** button in toolbar
2. Layout Manager dialog opens showing all layouts
3. Available actions:
   - **Search**: Filter layouts by name or description
   - **Edit**: Modify layout name and description
   - **Delete**: Remove layout with confirmation
   - **View Details**: See full layout metadata

### Edit Layout

1. Click **Edit** icon on layout card
2. Modify name or description
3. Click **Save** to update
4. Changes apply immediately

### Delete Layout

1. Click **Delete** icon on layout card
2. Confirm deletion in dialog
3. Layout removed from system
4. **Note**: Deletion is permanent and cannot be undone

### Layout Cards Display

- Layout name (bold)
- Description (if provided)
- Grid type badge
- Scope indicator (Global 🌐 or Personal 👤)
- Camera count
- Created by username
- Last updated time
- Action buttons (Edit, Delete)

---

## Layout Types

### Standard Layouts

Regular grid configurations for balanced viewing:

| Grid Type | Cells | Use Case |
|-----------|-------|----------|
| 2×2 | 4 | Small displays, focused monitoring |
| 2×3 | 6 | Portrait orientation, vertical displays |
| 3×3 | 9 | Default balanced view |
| 3×4 | 12 | Widescreen displays, control rooms |

**Features:**
- Equal-sized cells
- Uniform camera views
- Best for general monitoring
- Easy camera scanning

### Hotspot Layouts

Priority-based configurations with one large view:

| Grid Type | Total Cells | Hotspot Size | Regular Cells |
|-----------|-------------|--------------|---------------|
| 9-Way-1-Hotspot | 9 | 2×2 | 5 |
| 12-Way-1-Hotspot | 12 | 2×3 | 6 |
| 16-Way-1-Hotspot | 16 | 3×3 | 7 |
| 25-Way-1-Hotspot | 25 | 4×4 | 9 |
| 64-Way-1-Hotspot | 64 | 7×7 | 15 |

**Features:**
- One large "hotspot" cell (4× normal size)
- Multiple smaller surrounding cells
- Best for priority camera monitoring
- Quick context switching

**Hotspot Use Cases:**
- Main entrance (large) + perimeter cameras (small)
- Incident focus (large) + context views (small)
- VIP monitoring (large) + general surveillance (small)

---

## Layout Scopes

### Local (Personal) Layouts

**Characteristics:**
- 👤 Visible only to the creator
- Saved per user account
- Perfect for personal preferences
- Cannot be seen or edited by others

**Best For:**
- Personal monitoring preferences
- Temporary configurations
- Experimental setups
- Individual shift patterns

**Example:**
```
Name: "My Night Shift View"
Scope: Local
Cameras: Personal preference for night monitoring
```

### Global (Shared) Layouts

**Characteristics:**
- 🌐 Visible to all users
- Shared across team
- Standardized views
- Can be used by multiple operators

**Best For:**
- Department standards
- Shift handovers
- Training new operators
- Emergency procedures

**Example:**
```
Name: "Traffic Control Room - Standard"
Scope: Global
Cameras: Standard traffic monitoring layout for all operators
```

---

## Layout Workflows

### Workflow 1: Create Morning Shift Layout

```
1. Switch to 3×4 grid layout
2. Drag 12 high-priority cameras to grid
3. Position cameras by importance (top-left = highest priority)
4. Click "Save Layout"
5. Name: "Morning Shift - High Traffic"
6. Scope: Global
7. Description: "Main highways and intersections"
8. Save
```

### Workflow 2: Load Saved Layout

```
1. Start of shift
2. Click "Load Layout" dropdown
3. Select "Morning Shift - High Traffic"
4. Grid automatically switches to 3×4
5. All 12 cameras load in saved positions
6. Begin monitoring
```

### Workflow 3: Update Existing Layout

```
1. Load layout "Control Room Standard"
2. Make camera adjustments (add/remove cameras)
3. Delete old layout via "Manage"
4. Save new version with same name
5. All users get updated layout
```

### Workflow 4: Emergency Response Layout

```
1. Create 9-way-1-hotspot layout
2. Assign incident camera to large hotspot cell
3. Assign context cameras to surrounding cells
4. Save as "Emergency Response - Template"
5. Scope: Global
6. During incidents: Load layout, replace hotspot camera
```

---

## Best Practices

### Naming Conventions

- Use clear, descriptive names
- Include purpose: "Shift", "Department", "Event"
- Add time context: "Day Shift", "Night Shift"
- Examples:
  - ✅ "Traffic Control - Morning Peak"
  - ✅ "Metro Stations - Night Surveillance"
  - ❌ "Layout 1" (not descriptive)
  - ❌ "Test" (not meaningful)

### Organization Tips

- Create standardized layouts for each shift
- Use global layouts for shared operations
- Keep personal layouts for experimenting
- Regularly review and delete unused layouts
- Update layout descriptions when changing cameras

### Performance Considerations

- Limit number of cameras in hotspot layouts (10-16 recommended)
- Standard layouts perform better for large grids
- Consider monitor resolution when choosing grid size
- Test layouts before sharing globally

---

## Layout Data Structure

### Saved Layout Contains

```json
{
  "id": "uuid",
  "name": "Control Room Standard",
  "description": "Main traffic monitoring layout",
  "layout_type": "standard",
  "grid_layout": "3x3",
  "scope": "global",
  "created_by": "operator@rta.ae",
  "cameras": [
    {
      "camera_id": "cam-001",
      "position_index": 0
    },
    {
      "camera_id": "cam-002",
      "position_index": 1
    }
  ],
  "created_at": "2025-10-27T10:00:00Z",
  "updated_at": "2025-10-27T10:00:00Z"
}
```

### Field Descriptions

- `layout_type`: "standard" or "hotspot"
- `grid_layout`: Exact grid configuration (e.g., "3x3", "9-way-1-hotspot")
- `scope`: "local" (personal) or "global" (shared)
- `cameras`: Array of camera assignments with positions
- `position_index`: Zero-based position in grid (0 = top-left)

---

## Troubleshooting

### Layout Won't Save

**Problem**: "Save Layout" fails with error

**Solutions:**
1. Ensure at least one camera is assigned to grid
2. Check layout name is not empty
3. Verify network connection
4. Check browser console for errors
5. Try refreshing page and saving again

### Layout Loads But Grid Wrong

**Problem**: Cameras load but grid layout is incorrect

**Solutions:**
1. Check if grid_layout was saved correctly
2. Verify monitor resolution supports grid size
3. Hard refresh browser (Ctrl+Shift+R)
4. Clear browser cache
5. Contact administrator if issue persists

### Cameras Missing After Loading Layout

**Problem**: Some cameras don't appear after loading

**Solutions:**
1. Check if cameras are still online
2. Verify camera IDs haven't changed
3. Check if cameras were deleted from system
4. Load layout again (may be temporary network issue)
5. Contact administrator to verify camera status

### Cannot Delete Layout

**Problem**: Delete button doesn't work or shows error

**Solutions:**
1. Check permissions (may require admin role)
2. Refresh page and try again
3. Check if layout is being used by another user
4. Contact administrator if error persists

### Layout List Empty

**Problem**: No layouts visible in Load Layout dropdown

**Solutions:**
1. Create a new layout first
2. Check if layouts exist in database
3. Verify API connection
4. Clear browser cache and reload
5. Check browser console for API errors

---

## API Endpoints

### Layout Management API

```typescript
// Create layout
POST /api/v1/layouts
Body: {
  name: string,
  description?: string,
  layout_type: "standard" | "hotspot",
  grid_layout: string,
  scope: "global" | "local",
  created_by: string,
  cameras: Array<{camera_id: string, position_index: number}>
}

// List layouts
GET /api/v1/layouts
Query: ?layout_type=standard&scope=global

// Get specific layout
GET /api/v1/layouts/:id

// Update layout
PUT /api/v1/layouts/:id
Body: {
  name: string,
  description?: string,
  cameras: Array<{camera_id: string, position_index: number}>
}

// Delete layout
DELETE /api/v1/layouts/:id
```

---

## Database Schema

### Tables

```sql
-- Layout metadata
layout_preferences (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  layout_type VARCHAR(50) NOT NULL,
  grid_layout VARCHAR(50) NOT NULL,
  scope VARCHAR(50) NOT NULL,
  created_by VARCHAR(255) NOT NULL,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Camera assignments
layout_camera_assignments (
  id UUID PRIMARY KEY,
  layout_id UUID REFERENCES layout_preferences(id),
  camera_id VARCHAR(255) NOT NULL,
  position_index INT NOT NULL,
  cell_size VARCHAR(50),
  created_at TIMESTAMP,
  UNIQUE(layout_id, position_index)
);
```

### Indexes

```sql
CREATE INDEX idx_layout_prefs_scope ON layout_preferences(scope);
CREATE INDEX idx_layout_prefs_created_by ON layout_preferences(created_by);
CREATE INDEX idx_layout_prefs_type ON layout_preferences(layout_type);
CREATE INDEX idx_layout_prefs_active ON layout_preferences(is_active);
CREATE INDEX idx_layout_assignments_layout ON layout_camera_assignments(layout_id);
CREATE INDEX idx_layout_assignments_camera ON layout_camera_assignments(camera_id);
```

---

## Implementation Details

### Frontend Components

**Components Created:**
- `SaveLayoutDialog.tsx` - Dialog for saving new layouts
- `LoadLayoutDropdown.tsx` - Dropdown for browsing/loading layouts
- `LayoutManagerDialog.tsx` - Full layout management interface

**Integration:**
- Integrated into `StreamGridEnhanced.tsx`
- Toolbar buttons for Save/Load/Manage
- State management for layout operations

### Backend Services

**Go API Endpoints:**
- Layout Handler (`layout_handler.go`)
- Layout Use Case (`layout_usecase.go`)
- Layout Repository (`layout_repository.go`)
- Domain Models (`layout.go`)

**Features:**
- Full CRUD operations
- Input validation
- Transaction support
- Soft delete (is_active flag)
- Error handling

---

## Future Enhancements

### Planned Features

1. **Layout Templates**
   - Pre-configured layouts for common scenarios
   - Department-specific templates
   - Industry best practices templates

2. **Layout Sharing**
   - Share layouts via link
   - Import/export layout configurations
   - Copy layouts from other users

3. **Layout Scheduling**
   - Auto-switch layouts by time of day
   - Event-based layout changes
   - Calendar integration

4. **Advanced Filtering**
   - Filter layouts by department
   - Filter by camera count
   - Sort by usage frequency
   - Favorite/bookmark layouts

5. **Layout Analytics**
   - Track layout usage
   - Popular layouts dashboard
   - Load time metrics
   - User adoption statistics

---

**Document Version**: 1.0
**Last Updated**: 2025-10-27
**Author**: RTA CCTV Development Team
