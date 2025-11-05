-- Rollback layout preferences tables

-- Drop indexes
DROP INDEX IF EXISTS idx_layout_assignments_camera;
DROP INDEX IF EXISTS idx_layout_assignments_layout;
DROP INDEX IF EXISTS idx_layout_prefs_active;
DROP INDEX IF EXISTS idx_layout_prefs_type;
DROP INDEX IF EXISTS idx_layout_prefs_created_by;
DROP INDEX IF EXISTS idx_layout_prefs_scope;

-- Drop tables (child table first)
DROP TABLE IF EXISTS layout_camera_assignments;
DROP TABLE IF EXISTS layout_preferences;
