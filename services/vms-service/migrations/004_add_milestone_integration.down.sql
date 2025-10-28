-- Rollback migration: Remove Milestone XProtect integration

-- Drop triggers
DROP TRIGGER IF EXISTS update_recording_sessions_updated_at ON milestone_recording_sessions;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS milestone_playback_cache;
DROP TABLE IF EXISTS milestone_sync_history;
DROP TABLE IF EXISTS milestone_recording_sessions;

-- Drop indexes from cameras table
DROP INDEX IF EXISTS idx_cameras_last_milestone_sync;
DROP INDEX IF EXISTS idx_cameras_milestone_server;
DROP INDEX IF EXISTS idx_cameras_milestone_device_id;

-- Remove Milestone columns from cameras table
ALTER TABLE cameras
DROP COLUMN IF EXISTS milestone_metadata,
DROP COLUMN IF EXISTS last_milestone_sync,
DROP COLUMN IF EXISTS milestone_server,
DROP COLUMN IF EXISTS milestone_device_id;
