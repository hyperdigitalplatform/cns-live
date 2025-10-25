#!/bin/sh
# MinIO Bucket Initialization Script
# Creates buckets and sets policies for RTA CCTV System

set -e

echo "Waiting for MinIO to be ready..."
until mc alias set minio http://minio:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD}; do
    echo "MinIO not ready yet, waiting..."
    sleep 2
done

echo "MinIO is ready. Creating buckets..."

# Create buckets
mc mb --ignore-existing minio/cctv-recordings
mc mb --ignore-existing minio/cctv-exports
mc mb --ignore-existing minio/cctv-thumbnails
mc mb --ignore-existing minio/cctv-clips

echo "Buckets created successfully"

# Set lifecycle policies for automatic cleanup
echo "Setting lifecycle policies..."

# Recordings: Delete after 90 days
cat > /tmp/recordings-lifecycle.json <<EOF
{
    "Rules": [
        {
            "ID": "DeleteOldRecordings",
            "Status": "Enabled",
            "Expiration": {
                "Days": 90
            }
        }
    ]
}
EOF

mc ilm import minio/cctv-recordings < /tmp/recordings-lifecycle.json

# Exports: Delete after 7 days
cat > /tmp/exports-lifecycle.json <<EOF
{
    "Rules": [
        {
            "ID": "DeleteOldExports",
            "Status": "Enabled",
            "Expiration": {
                "Days": 7
            }
        }
    ]
}
EOF

mc ilm import minio/cctv-exports < /tmp/exports-lifecycle.json

# Thumbnails: Delete after 30 days
cat > /tmp/thumbnails-lifecycle.json <<EOF
{
    "Rules": [
        {
            "ID": "DeleteOldThumbnails",
            "Status": "Enabled",
            "Expiration": {
                "Days": 30
            }
        }
    ]
}
EOF

mc ilm import minio/cctv-thumbnails < /tmp/thumbnails-lifecycle.json

# Clips: Keep indefinitely (manual deletion only)

echo "Lifecycle policies set successfully"

# Set public download policy for exports (temporary URLs)
cat > /tmp/export-policy.json <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": ["*"]
            },
            "Action": [
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::cctv-exports/*"
            ],
            "Condition": {
                "StringLike": {
                    "s3:ExistingObjectTag/temporary": "true"
                }
            }
        }
    ]
}
EOF

mc anonymous set-json /tmp/export-policy.json minio/cctv-exports

echo "Bucket policies configured"

# Create access keys for services
echo "Creating service access keys..."

# Create user for recording service
mc admin user add minio recording-service ${RECORDING_SERVICE_PASSWORD:-changeme_recording}
mc admin policy attach minio readwrite --user recording-service

# Create user for storage service
mc admin user add minio storage-service ${STORAGE_SERVICE_PASSWORD:-changeme_storage}
mc admin policy attach minio readwrite --user storage-service

# Create user for playback service (readwrite needed for bucket operations)
mc admin user add minio playback-service ${PLAYBACK_SERVICE_PASSWORD:-changeme_playback}
mc admin policy attach minio readwrite --user playback-service

echo "Service users created"

# Display bucket information
echo ""
echo "=== MinIO Buckets Created ==="
mc ls minio

echo ""
echo "=== Lifecycle Policies ==="
mc ilm ls minio/cctv-recordings
mc ilm ls minio/cctv-exports
mc ilm ls minio/cctv-thumbnails

echo ""
echo "MinIO initialization complete!"
