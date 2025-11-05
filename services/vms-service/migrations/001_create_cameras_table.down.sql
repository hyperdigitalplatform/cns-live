-- Rollback cameras table creation

-- Drop tables in reverse order (child tables first)
DROP TABLE IF EXISTS recording_exports;
DROP TABLE IF EXISTS cameras;

-- Drop indexes (they will be dropped with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_recording_exports_created_at;
DROP INDEX IF EXISTS idx_recording_exports_status;
DROP INDEX IF EXISTS idx_recording_exports_camera_id;
DROP INDEX IF EXISTS idx_cameras_last_update;
DROP INDEX IF EXISTS idx_cameras_milestone_device_id;
DROP INDEX IF EXISTS idx_cameras_status;
DROP INDEX IF EXISTS idx_cameras_source;
