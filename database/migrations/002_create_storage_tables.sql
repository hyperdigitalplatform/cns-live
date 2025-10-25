-- Storage Service Tables
-- Segments and Exports for video storage

-- Enable UUID extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Segments table: Video segment metadata
CREATE TABLE IF NOT EXISTS segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_seconds INTEGER NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_backend VARCHAR(50) NOT NULL,
    storage_path TEXT NOT NULL,
    checksum VARCHAR(64),
    thumbnail_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_segments_camera_time
    ON segments(camera_id, start_time, end_time);

CREATE INDEX IF NOT EXISTS idx_segments_created_at
    ON segments(created_at);

CREATE INDEX IF NOT EXISTS idx_segments_storage_backend
    ON segments(storage_backend);

-- Exports table: Video export requests
CREATE TABLE IF NOT EXISTS exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_ids UUID[] NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    format VARCHAR(10) NOT NULL DEFAULT 'mp4',
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    file_path TEXT,
    file_size BIGINT,
    download_url TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for exports
CREATE INDEX IF NOT EXISTS idx_exports_status
    ON exports(status);

CREATE INDEX IF NOT EXISTS idx_exports_created_at
    ON exports(created_at);

CREATE INDEX IF NOT EXISTS idx_exports_expires_at
    ON exports(expires_at);

-- Comments for documentation
COMMENT ON TABLE segments IS 'Video segment metadata for stored recordings';
COMMENT ON TABLE exports IS 'Video export requests and their status';

COMMENT ON COLUMN segments.storage_backend IS 'Storage backend type: MINIO, S3, FILESYSTEM, MILESTONE';
COMMENT ON COLUMN segments.storage_path IS 'Path to segment file in storage backend';
COMMENT ON COLUMN segments.checksum IS 'SHA-256 checksum of segment file';

COMMENT ON COLUMN exports.status IS 'Export status: PENDING, PROCESSING, COMPLETED, FAILED';
COMMENT ON COLUMN exports.format IS 'Export format: mp4, avi, mkv';
COMMENT ON COLUMN exports.download_url IS 'Temporary signed URL for downloading export';
