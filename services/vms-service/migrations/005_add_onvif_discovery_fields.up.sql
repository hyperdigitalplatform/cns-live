-- Add ONVIF discovery fields to cameras table
-- This migration extends the cameras table with fields from ONVIF discovery

-- Add ONVIF device information fields
ALTER TABLE cameras
ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45),
ADD COLUMN IF NOT EXISTS onvif_port INTEGER,
ADD COLUMN IF NOT EXISTS manufacturer VARCHAR(255),
ADD COLUMN IF NOT EXISTS model VARCHAR(255),
ADD COLUMN IF NOT EXISTS firmware_version VARCHAR(255),
ADD COLUMN IF NOT EXISTS serial_number VARCHAR(255),
ADD COLUMN IF NOT EXISTS hardware_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS onvif_endpoint TEXT,
ADD COLUMN IF NOT EXISTS onvif_username VARCHAR(255),
ADD COLUMN IF NOT EXISTS onvif_password_encrypted TEXT;

-- Create camera_streams table for storing stream profiles
CREATE TABLE IF NOT EXISTS camera_streams (
    id SERIAL PRIMARY KEY,
    camera_id VARCHAR(255) NOT NULL,
    profile_token VARCHAR(255) NOT NULL,
    profile_name VARCHAR(255) NOT NULL,
    encoding VARCHAR(50) NOT NULL,
    resolution VARCHAR(50) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    frame_rate INTEGER NOT NULL,
    bitrate INTEGER NOT NULL,
    rtsp_url TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT fk_camera_stream FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE,
    CONSTRAINT unique_camera_profile UNIQUE (camera_id, profile_token)
);

-- Create indexes for camera_streams
CREATE INDEX IF NOT EXISTS idx_camera_streams_camera_id ON camera_streams(camera_id);
CREATE INDEX IF NOT EXISTS idx_camera_streams_is_primary ON camera_streams(camera_id, is_primary) WHERE is_primary = true;

-- Add index for IP address for faster lookups
CREATE INDEX IF NOT EXISTS idx_cameras_ip_address ON cameras(ip_address);
