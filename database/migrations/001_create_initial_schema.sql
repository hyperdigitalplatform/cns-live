-- RTA CCTV System - Initial Database Schema
-- Version: 1.0.0
-- Description: Creates base tables for cameras, streams, and recordings

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For text search

-- ============================================
-- CAMERAS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS cameras (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Location information
    location VARCHAR(255),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    agency VARCHAR(100) NOT NULL CHECK (agency IN ('dubai_police', 'metro', 'bus', 'other')),

    -- Technical details
    stream_url TEXT NOT NULL,
    stream_type VARCHAR(20) NOT NULL DEFAULT 'RTSP' CHECK (stream_type IN ('RTSP', 'HTTP', 'RTMP')),
    resolution VARCHAR(20),
    fps INTEGER DEFAULT 25,

    -- Capabilities
    has_ptz BOOLEAN DEFAULT FALSE,
    has_audio BOOLEAN DEFAULT FALSE,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'MAINTENANCE', 'ERROR')),
    last_seen TIMESTAMP WITH TIME ZONE,

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Full-text search
    search_vector tsvector
);

-- Indexes for cameras
CREATE INDEX IF NOT EXISTS idx_cameras_camera_id ON cameras(camera_id);
CREATE INDEX IF NOT EXISTS idx_cameras_agency ON cameras(agency);
CREATE INDEX IF NOT EXISTS idx_cameras_status ON cameras(status);
CREATE INDEX IF NOT EXISTS idx_cameras_location ON cameras(location);
CREATE INDEX IF NOT EXISTS idx_cameras_search ON cameras USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_cameras_name_trgm ON cameras USING GIN(name gin_trgm_ops);

-- Trigger function for camera search vector
CREATE OR REPLACE FUNCTION cameras_search_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.location, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cameras_search_update
    BEFORE INSERT OR UPDATE ON cameras
    FOR EACH ROW
    EXECUTE FUNCTION cameras_search_trigger();

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_cameras_updated_at
    BEFORE UPDATE ON cameras
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE cameras IS 'Camera registry with metadata and status';
COMMENT ON COLUMN cameras.agency IS 'Agency that owns the camera: dubai_police, metro, bus, other';
COMMENT ON COLUMN cameras.status IS 'Camera operational status';

-- ============================================
-- STREAMS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id UUID NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
    agency VARCHAR(100) NOT NULL,

    -- LiveKit information
    livekit_room_name VARCHAR(255) NOT NULL,
    livekit_token TEXT,

    -- Stream details
    viewer_id VARCHAR(255) NOT NULL,
    viewer_ip VARCHAR(50),

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'ERROR')),

    -- Timestamps
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    last_heartbeat TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for streams
CREATE INDEX IF NOT EXISTS idx_streams_camera_id ON streams(camera_id);
CREATE INDEX IF NOT EXISTS idx_streams_agency ON streams(agency);
CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);
CREATE INDEX IF NOT EXISTS idx_streams_viewer_id ON streams(viewer_id);
CREATE INDEX IF NOT EXISTS idx_streams_started_at ON streams(started_at);

COMMENT ON TABLE streams IS 'Active stream reservations and viewer sessions';
COMMENT ON COLUMN streams.agency IS 'Agency quota this stream counts against';

-- ============================================
-- RECORDINGS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id UUID NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,

    -- Recording details
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,

    -- Storage
    storage_backend VARCHAR(50) NOT NULL CHECK (storage_backend IN ('MINIO', 'S3', 'MILESTONE', 'FILESYSTEM')),
    storage_path TEXT NOT NULL,
    total_size_bytes BIGINT DEFAULT 0,
    segment_count INTEGER DEFAULT 0,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'RECORDING' CHECK (status IN ('RECORDING', 'COMPLETED', 'FAILED', 'STOPPED')),

    -- Metadata
    recorded_by VARCHAR(255),
    reason TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for recordings
CREATE INDEX IF NOT EXISTS idx_recordings_camera_id ON recordings(camera_id);
CREATE INDEX IF NOT EXISTS idx_recordings_status ON recordings(status);
CREATE INDEX IF NOT EXISTS idx_recordings_time ON recordings(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_recordings_created_at ON recordings(created_at);

COMMENT ON TABLE recordings IS 'Recording sessions for cameras';
COMMENT ON COLUMN recordings.storage_backend IS 'Storage backend where recording is stored';

-- ============================================
-- VIDEO SEGMENTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS video_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id UUID NOT NULL REFERENCES recordings(id) ON DELETE CASCADE,
    camera_id UUID NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,

    -- Segment timing
    segment_index INTEGER NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_seconds INTEGER NOT NULL,

    -- Storage
    storage_backend VARCHAR(50) NOT NULL,
    storage_path TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL,

    -- Video properties
    codec VARCHAR(20),
    resolution VARCHAR(20),
    fps INTEGER,
    bitrate_kbps INTEGER,

    -- Verification
    checksum VARCHAR(64),
    is_corrupted BOOLEAN DEFAULT FALSE,

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for video_segments
CREATE INDEX IF NOT EXISTS idx_video_segments_recording_id ON video_segments(recording_id);
CREATE INDEX IF NOT EXISTS idx_video_segments_camera_id ON video_segments(camera_id);
CREATE INDEX IF NOT EXISTS idx_video_segments_time ON video_segments(camera_id, start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_video_segments_segment_index ON video_segments(recording_id, segment_index);
CREATE INDEX IF NOT EXISTS idx_video_segments_created_at ON video_segments(created_at);

-- Unique constraint for segments
CREATE UNIQUE INDEX IF NOT EXISTS idx_video_segments_unique
    ON video_segments(recording_id, segment_index);

COMMENT ON TABLE video_segments IS 'Individual video segments that make up recordings';
COMMENT ON COLUMN video_segments.segment_index IS 'Sequential index within recording (starts at 0)';

-- ============================================
-- STREAM STATISTICS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS stream_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Time bucket (for time-series aggregation)
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Metrics by agency
    agency VARCHAR(100) NOT NULL,
    active_streams INTEGER NOT NULL DEFAULT 0,
    total_streams_started INTEGER NOT NULL DEFAULT 0,
    total_streams_ended INTEGER NOT NULL DEFAULT 0,

    -- Bandwidth metrics (in bytes)
    total_bytes_sent BIGINT DEFAULT 0,

    -- Average metrics
    avg_duration_seconds INTEGER,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for stream_stats
CREATE INDEX IF NOT EXISTS idx_stream_stats_timestamp ON stream_stats(timestamp);
CREATE INDEX IF NOT EXISTS idx_stream_stats_agency ON stream_stats(agency);

-- Unique constraint for time buckets
CREATE UNIQUE INDEX IF NOT EXISTS idx_stream_stats_unique
    ON stream_stats(timestamp, agency);

COMMENT ON TABLE stream_stats IS 'Time-series statistics for stream usage by agency';

-- ============================================
-- SYSTEM SETTINGS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_system_settings_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default settings
INSERT INTO system_settings (key, value, description) VALUES
    ('retention_days', '90', 'Number of days to retain recordings'),
    ('max_concurrent_streams', '500', 'Maximum total concurrent streams'),
    ('recording_segment_duration', '300', 'Recording segment duration in seconds (5 minutes)'),
    ('enable_object_detection', 'false', 'Enable AI object detection on recordings')
ON CONFLICT (key) DO NOTHING;

COMMENT ON TABLE system_settings IS 'System-wide configuration settings';

-- ============================================
-- VIEWS FOR COMMON QUERIES
-- ============================================

-- Active cameras view
CREATE OR REPLACE VIEW v_active_cameras AS
SELECT
    c.id,
    c.camera_id,
    c.name,
    c.location,
    c.agency,
    c.status,
    c.has_ptz,
    COUNT(s.id) AS active_streams
FROM cameras c
LEFT JOIN streams s ON c.id = s.camera_id AND s.status = 'ACTIVE'
WHERE c.status = 'ACTIVE'
GROUP BY c.id;

COMMENT ON VIEW v_active_cameras IS 'Active cameras with their current stream counts';

-- Recording summary view
CREATE OR REPLACE VIEW v_recording_summary AS
SELECT
    c.id AS camera_id,
    c.camera_id AS camera_code,
    c.name AS camera_name,
    c.agency,
    COUNT(DISTINCT r.id) AS total_recordings,
    COUNT(vs.id) AS total_segments,
    SUM(vs.file_size_bytes) AS total_size_bytes,
    SUM(vs.duration_seconds) AS total_duration_seconds,
    MAX(r.end_time) AS last_recording_time
FROM cameras c
LEFT JOIN recordings r ON c.id = r.camera_id
LEFT JOIN video_segments vs ON r.id = vs.recording_id
GROUP BY c.id, c.camera_id, c.name, c.agency;

COMMENT ON VIEW v_recording_summary IS 'Recording statistics per camera';

-- Agency quota view
CREATE OR REPLACE VIEW v_agency_quotas AS
SELECT
    agency,
    COUNT(*) FILTER (WHERE status = 'ACTIVE') AS active_streams,
    CASE agency
        WHEN 'dubai_police' THEN 50
        WHEN 'metro' THEN 30
        WHEN 'bus' THEN 20
        ELSE 400
    END AS stream_limit
FROM streams
GROUP BY agency
UNION ALL
SELECT
    'dubai_police' AS agency,
    0 AS active_streams,
    50 AS stream_limit
WHERE NOT EXISTS (SELECT 1 FROM streams WHERE agency = 'dubai_police')
UNION ALL
SELECT
    'metro' AS agency,
    0 AS active_streams,
    30 AS stream_limit
WHERE NOT EXISTS (SELECT 1 FROM streams WHERE agency = 'metro')
UNION ALL
SELECT
    'bus' AS agency,
    0 AS active_streams,
    20 AS stream_limit
WHERE NOT EXISTS (SELECT 1 FROM streams WHERE agency = 'bus')
UNION ALL
SELECT
    'other' AS agency,
    0 AS active_streams,
    400 AS stream_limit
WHERE NOT EXISTS (SELECT 1 FROM streams WHERE agency = 'other');

COMMENT ON VIEW v_agency_quotas IS 'Current stream usage vs limits by agency';

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to cleanup old stream records
CREATE OR REPLACE FUNCTION cleanup_old_streams(days_old INTEGER DEFAULT 7)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM streams
    WHERE status = 'INACTIVE'
    AND ended_at < NOW() - INTERVAL '1 day' * days_old;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_streams IS 'Cleanup inactive stream records older than specified days';

-- Function to get camera recording availability
CREATE OR REPLACE FUNCTION get_camera_availability(
    p_camera_id UUID,
    p_start_time TIMESTAMP WITH TIME ZONE,
    p_end_time TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE(
    segment_start TIMESTAMP WITH TIME ZONE,
    segment_end TIMESTAMP WITH TIME ZONE,
    available BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        vs.start_time AS segment_start,
        vs.end_time AS segment_end,
        TRUE AS available
    FROM video_segments vs
    WHERE vs.camera_id = p_camera_id
    AND vs.start_time <= p_end_time
    AND vs.end_time >= p_start_time
    AND vs.is_corrupted = FALSE
    ORDER BY vs.start_time;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_camera_availability IS 'Get available recording segments for a camera in a time range';

-- ============================================
-- INITIAL DATA
-- ============================================

-- Insert sample agencies (for reference)
-- Actual camera data will be populated by VMS service

COMMENT ON SCHEMA public IS 'RTA CCTV System - Version 1.0.0';

-- Grant permissions to cctv user (if needed)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cctv;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cctv;
-- GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO cctv;
