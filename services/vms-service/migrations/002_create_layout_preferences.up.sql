-- Migration: Create layout preferences tables
-- Description: Store user-saved camera layout configurations

-- Layout preferences table
CREATE TABLE IF NOT EXISTS layout_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    layout_type VARCHAR(50) NOT NULL CHECK (layout_type IN ('standard', 'hotspot')),
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('global', 'local')),
    created_by VARCHAR(255) NOT NULL, -- User identifier (for demo, simple string)
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Layout camera assignments table (stores which cameras go where)
CREATE TABLE IF NOT EXISTS layout_camera_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    layout_id UUID NOT NULL REFERENCES layout_preferences(id) ON DELETE CASCADE,
    camera_id VARCHAR(255) NOT NULL,
    position_index INT NOT NULL, -- Position in the layout (0-based)
    cell_size VARCHAR(50), -- 'small', 'medium', 'large', 'hotspot'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Ensure no duplicate camera positions in same layout
    UNIQUE (layout_id, position_index)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_layout_prefs_scope ON layout_preferences(scope);
CREATE INDEX IF NOT EXISTS idx_layout_prefs_created_by ON layout_preferences(created_by);
CREATE INDEX IF NOT EXISTS idx_layout_prefs_type ON layout_preferences(layout_type);
CREATE INDEX IF NOT EXISTS idx_layout_prefs_active ON layout_preferences(is_active);
CREATE INDEX IF NOT EXISTS idx_layout_assignments_layout ON layout_camera_assignments(layout_id);
CREATE INDEX IF NOT EXISTS idx_layout_assignments_camera ON layout_camera_assignments(camera_id);

-- Comments for documentation
COMMENT ON TABLE layout_preferences IS 'Stores user-saved camera layout configurations';
COMMENT ON TABLE layout_camera_assignments IS 'Stores camera-to-position assignments for each layout';
COMMENT ON COLUMN layout_preferences.scope IS 'global: visible to all users, local: personal to creator';
COMMENT ON COLUMN layout_preferences.layout_type IS 'standard: grid layout, hotspot: priority/custom layout';
COMMENT ON COLUMN layout_camera_assignments.position_index IS 'Zero-based position index in the layout grid';
