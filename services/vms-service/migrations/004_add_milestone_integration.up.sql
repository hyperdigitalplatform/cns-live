-- Migration: Add Milestone XProtect integration fields
-- Description: Adds fields to support Milestone camera synchronization and tracking

-- Add Milestone-specific columns to cameras table
ALTER TABLE cameras
ADD COLUMN IF NOT EXISTS milestone_device_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS milestone_server VARCHAR(255),
ADD COLUMN IF NOT EXISTS last_milestone_sync TIMESTAMP,
ADD COLUMN IF NOT EXISTS milestone_metadata JSONB;

-- Create indexes for Milestone device ID lookups
CREATE INDEX IF NOT EXISTS idx_cameras_milestone_device_id ON cameras(milestone_device_id);
CREATE INDEX IF NOT EXISTS idx_cameras_milestone_server ON cameras(milestone_server);
CREATE INDEX IF NOT EXISTS idx_cameras_last_milestone_sync ON cameras(last_milestone_sync);

-- Create milestone_recording_sessions table
CREATE TABLE IF NOT EXISTS milestone_recording_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id VARCHAR(255) NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
    milestone_recording_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INT NOT NULL,
    triggered_by VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'recording',
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_recording_status CHECK (status IN ('recording', 'stopped', 'failed', 'completed'))
);

-- Create indexes for recording sessions
CREATE INDEX IF NOT EXISTS idx_recording_sessions_camera_id ON milestone_recording_sessions(camera_id);
CREATE INDEX IF NOT EXISTS idx_recording_sessions_status ON milestone_recording_sessions(status);
CREATE INDEX IF NOT EXISTS idx_recording_sessions_start_time ON milestone_recording_sessions(start_time);
CREATE INDEX IF NOT EXISTS idx_recording_sessions_milestone_recording_id ON milestone_recording_sessions(milestone_recording_id);

-- Create milestone_sync_history table
CREATE TABLE IF NOT EXISTS milestone_sync_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sync_type VARCHAR(50) NOT NULL,
    cameras_discovered INT DEFAULT 0,
    cameras_imported INT DEFAULT 0,
    cameras_updated INT DEFAULT 0,
    errors INT DEFAULT 0,
    error_details JSONB,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    initiated_by VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_sync_status CHECK (status IN ('in_progress', 'completed', 'failed')),
    CONSTRAINT chk_sync_type CHECK (sync_type IN ('camera_discovery', 'camera_import', 'full_sync', 'single_camera'))
);

-- Create indexes for sync history
CREATE INDEX IF NOT EXISTS idx_sync_history_sync_type ON milestone_sync_history(sync_type);
CREATE INDEX IF NOT EXISTS idx_sync_history_started_at ON milestone_sync_history(started_at);
CREATE INDEX IF NOT EXISTS idx_sync_history_status ON milestone_sync_history(status);

-- Create milestone_playback_cache table
CREATE TABLE IF NOT EXISTS milestone_playback_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id VARCHAR(255) NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
    query_hash VARCHAR(64) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    sequence_data JSONB NOT NULL,
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    hit_count INT DEFAULT 0,

    CONSTRAINT chk_cache_time_range CHECK (end_time > start_time),
    CONSTRAINT chk_cache_expiry CHECK (expires_at > cached_at)
);

-- Create unique index on query hash
CREATE UNIQUE INDEX IF NOT EXISTS idx_playback_cache_query_hash ON milestone_playback_cache(query_hash);
CREATE INDEX IF NOT EXISTS idx_playback_cache_camera_id ON milestone_playback_cache(camera_id);
CREATE INDEX IF NOT EXISTS idx_playback_cache_expires_at ON milestone_playback_cache(expires_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for milestone_recording_sessions
CREATE TRIGGER update_recording_sessions_updated_at
    BEFORE UPDATE ON milestone_recording_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE milestone_recording_sessions IS 'Tracks manual recording sessions initiated through Milestone XProtect integration';
COMMENT ON TABLE milestone_sync_history IS 'Logs synchronization operations with Milestone XProtect VMS';
COMMENT ON TABLE milestone_playback_cache IS 'Caches recording sequence queries to improve playback performance';

COMMENT ON COLUMN cameras.milestone_device_id IS 'Milestone XProtect device/camera ID for API operations';
COMMENT ON COLUMN cameras.milestone_server IS 'Milestone recording server hostname/address';
COMMENT ON COLUMN cameras.last_milestone_sync IS 'Timestamp of last successful sync with Milestone';
COMMENT ON COLUMN cameras.milestone_metadata IS 'Additional Milestone-specific metadata (JSON)';
