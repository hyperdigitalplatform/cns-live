-- Rollback ONVIF discovery fields migration

-- Drop camera_streams table
DROP TABLE IF EXISTS camera_streams;

-- Remove ONVIF fields from cameras table
ALTER TABLE cameras
DROP COLUMN IF EXISTS ip_address,
DROP COLUMN IF EXISTS onvif_port,
DROP COLUMN IF EXISTS manufacturer,
DROP COLUMN IF EXISTS model,
DROP COLUMN IF EXISTS firmware_version,
DROP COLUMN IF EXISTS serial_number,
DROP COLUMN IF EXISTS hardware_id,
DROP COLUMN IF EXISTS onvif_endpoint,
DROP COLUMN IF EXISTS onvif_username,
DROP COLUMN IF EXISTS onvif_password_encrypted;

-- Drop index
DROP INDEX IF EXISTS idx_cameras_ip_address;
