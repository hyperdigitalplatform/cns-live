-- Create cameras table for VMS service
-- This table stores camera metadata synced from Milestone VMS

CREATE TABLE IF NOT EXISTS cameras (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    name_ar VARCHAR(500),
    source VARCHAR(50) NOT NULL,
    rtsp_url TEXT NOT NULL,
    ptz_enabled BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'OFFLINE',
    recording_server VARCHAR(255),
    milestone_device_id VARCHAR(255) UNIQUE,
    metadata JSONB,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT valid_source CHECK (source IN ('DUBAI_POLICE', 'METRO', 'BUS', 'OTHER')),
    CONSTRAINT valid_status CHECK (status IN ('ONLINE', 'OFFLINE', 'ERROR'))
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_cameras_source ON cameras(source);
CREATE INDEX IF NOT EXISTS idx_cameras_status ON cameras(status);
CREATE INDEX IF NOT EXISTS idx_cameras_milestone_device_id ON cameras(milestone_device_id);
CREATE INDEX IF NOT EXISTS idx_cameras_last_update ON cameras(last_update DESC);

-- Create recording exports table
CREATE TABLE IF NOT EXISTS recording_exports (
    id VARCHAR(255) PRIMARY KEY,
    camera_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    format VARCHAR(10) NOT NULL,
    quality VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    file_path TEXT,
    file_size BIGINT,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT valid_export_status CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED')),
    CONSTRAINT valid_time_range CHECK (end_time > start_time),
    CONSTRAINT fk_camera FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE
);

-- Create indexes for exports
CREATE INDEX IF NOT EXISTS idx_recording_exports_camera_id ON recording_exports(camera_id);
CREATE INDEX IF NOT EXISTS idx_recording_exports_status ON recording_exports(status);
CREATE INDEX IF NOT EXISTS idx_recording_exports_created_at ON recording_exports(created_at DESC);

-- Insert sample cameras (for testing)
INSERT INTO cameras (id, name, name_ar, source, rtsp_url, ptz_enabled, status, recording_server, milestone_device_id, metadata, created_at)
VALUES
    ('cam-001-sheikh-zayed', 'Camera 1 - 192.168.1.8', 'كاميرا 1', 'DUBAI_POLICE',
     'rtsp://raammohan:Ilove123@192.168.1.8:554/stream1', true, 'ONLINE', '192.168.1.8:554', 'camera_device_001',
     '{"fps": 25, "location": {"lat": 25.2048, "lon": 55.2708}, "resolution": "1920x1080"}', NOW() - INTERVAL '30 days'),

    ('cam-002-metro-station', 'Camera 2 - 192.168.1.13', 'كاميرا 2', 'METRO',
     'rtsp://admin:pass@192.168.1.13:554/ch0_0.264', true, 'ONLINE', '192.168.1.13:554', 'camera_device_002',
     '{"fps": 25, "location": {"lat": 25.2697, "lon": 55.3095}, "resolution": "1920x1080"}', NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;
