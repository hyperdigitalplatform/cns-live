-- Migration: Add grid_layout column to layout_preferences
-- Description: Store the specific grid configuration (2x2, 3x3, 9-way-1-hotspot, etc.)

-- Add grid_layout column
ALTER TABLE layout_preferences
ADD COLUMN grid_layout VARCHAR(50);

-- Add comment
COMMENT ON COLUMN layout_preferences.grid_layout IS 'Specific grid layout configuration (e.g., 2x2, 3x3, 9-way-1-hotspot)';

-- For existing rows, set a default based on layout_type
-- This is just for backward compatibility with any existing data
UPDATE layout_preferences
SET grid_layout = CASE
    WHEN layout_type = 'standard' THEN '3x3'
    WHEN layout_type = 'hotspot' THEN '9-way-1-hotspot'
    ELSE '2x2'
END
WHERE grid_layout IS NULL;

-- Now make it NOT NULL since all rows have values
ALTER TABLE layout_preferences
ALTER COLUMN grid_layout SET NOT NULL;
