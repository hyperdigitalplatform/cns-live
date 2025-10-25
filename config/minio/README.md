# MinIO Object Storage Configuration

## Overview

MinIO provides S3-compatible object storage for the RTA CCTV system, storing:
- **Recordings**: Continuous video recordings from cameras (90-day retention)
- **Exports**: User-requested video exports (7-day retention)
- **Thumbnails**: Preview images for video segments (30-day retention)
- **Clips**: Incident clips and evidence (kept indefinitely)

## Architecture

```
┌──────────────────────────────────────────┐
│         MinIO Object Storage             │
│  ┌────────────────────────────────────┐  │
│  │  Bucket: cctv-recordings           │  │
│  │  • 90-day lifecycle                │  │
│  │  • Write: recording-service        │  │
│  │  • Read: playback-service          │  │
│  │  • Size: ~500TB (500 cams × 90d)   │  │
│  └────────────────────────────────────┘  │
│  ┌────────────────────────────────────┐  │
│  │  Bucket: cctv-exports              │  │
│  │  • 7-day lifecycle                 │  │
│  │  • Temporary download URLs         │  │
│  │  • Size: ~1TB                      │  │
│  └────────────────────────────────────┘  │
│  ┌────────────────────────────────────┐  │
│  │  Bucket: cctv-thumbnails           │  │
│  │  • 30-day lifecycle                │  │
│  │  • JPEG preview images             │  │
│  │  • Size: ~100GB                    │  │
│  └────────────────────────────────────┘  │
│  ┌────────────────────────────────────┐  │
│  │  Bucket: cctv-clips                │  │
│  │  • No expiration                   │  │
│  │  • Manual deletion only            │  │
│  │  • Size: ~10TB                     │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
         │                    │
         ↓                    ↓
  ┌──────────┐        ┌──────────────┐
  │Recording │        │   Playback   │
  │ Service  │        │   Service    │
  └──────────┘        └──────────────┘
```

## Buckets

### 1. cctv-recordings

**Purpose**: Continuous video recordings from all cameras

**Path Structure**:
```
cctv-recordings/
├── {camera_id}/
│   ├── {year}/
│   │   ├── {month}/
│   │   │   ├── {day}/
│   │   │   │   ├── {hour}-00-00.ts    # 1-hour segment
│   │   │   │   ├── {hour}-00-00.json  # Segment metadata
│   │   │   │   └── ...
```

**Example**:
```
cctv-recordings/
├── 123e4567-e89b-12d3-a456-426614174000/
│   ├── 2025/
│   │   ├── 01/
│   │   │   ├── 23/
│   │   │   │   ├── 00-00-00.ts      # Jan 23, 2025, 00:00-01:00
│   │   │   │   ├── 00-00-00.json    # Metadata
│   │   │   │   ├── 01-00-00.ts      # Jan 23, 2025, 01:00-02:00
│   │   │   │   └── ...
```

**Lifecycle**: Delete after 90 days

**Access**:
- Write: `recording-service` user
- Read: `storage-service`, `playback-service` users

### 2. cctv-exports

**Purpose**: User-requested video exports (clips, incidents)

**Path Structure**:
```
cctv-exports/
├── {export_id}.mp4
├── {export_id}.json    # Export metadata
```

**Example**:
```
cctv-exports/
├── export-2025-01-23-abc123.mp4
├── export-2025-01-23-abc123.json
```

**Lifecycle**: Delete after 7 days

**Access**:
- Write: `storage-service` user
- Read: Public (with temporary signed URLs)

### 3. cctv-thumbnails

**Purpose**: JPEG preview images for video segments

**Path Structure**:
```
cctv-thumbnails/
├── {camera_id}/
│   ├── {year}/
│   │   ├── {month}/
│   │   │   ├── {day}/
│   │   │   │   ├── {timestamp}.jpg
```

**Example**:
```
cctv-thumbnails/
├── 123e4567-e89b-12d3-a456-426614174000/
│   ├── 2025/
│   │   ├── 01/
│   │   │   ├── 23/
│   │   │   │   ├── 2025-01-23-00-00-00.jpg
│   │   │   │   ├── 2025-01-23-00-05-00.jpg
│   │   │   │   └── ...
```

**Lifecycle**: Delete after 30 days

**Access**:
- Write: `recording-service` user
- Read: `playback-service`, `storage-service` users

### 4. cctv-clips

**Purpose**: Incident clips and evidence (permanent storage)

**Path Structure**:
```
cctv-clips/
├── {incident_id}/
│   ├── {clip_id}.mp4
│   ├── {clip_id}.json    # Clip metadata
│   └── {clip_id}.jpg     # Thumbnail
```

**Example**:
```
cctv-clips/
├── incident-2025-01-23-traffic-001/
│   ├── clip-cam1.mp4
│   ├── clip-cam1.json
│   ├── clip-cam1.jpg
│   ├── clip-cam2.mp4
│   └── ...
```

**Lifecycle**: No expiration (manual deletion only)

**Access**:
- Write: `storage-service` user
- Read: `playback-service`, authorized users

## Service Users

MinIO uses IAM-style users for access control:

### 1. recording-service

**Permissions**: Read/Write to `cctv-recordings`, `cctv-thumbnails`
**Policy**: `readwrite`
**Usage**: Recording Service writes video segments

### 2. storage-service

**Permissions**: Read/Write to all buckets
**Policy**: `readwrite`
**Usage**: Storage Service orchestrates video storage

### 3. playback-service

**Permissions**: Read-only to all buckets
**Policy**: `readonly`
**Usage**: Playback Service reads videos for streaming

## Lifecycle Policies

Automatic cleanup to manage storage costs:

| Bucket | Retention | Policy |
|--------|-----------|--------|
| cctv-recordings | 90 days | Delete old segments |
| cctv-exports | 7 days | Delete temporary exports |
| cctv-thumbnails | 30 days | Delete old previews |
| cctv-clips | Indefinite | Manual deletion only |

## Storage Capacity Planning

### Per Camera

**Assumptions**:
- Resolution: 1080p (1920×1080)
- Codec: H.264
- Bitrate: 4 Mbps average
- Recording: 24/7 continuous

**Daily Storage**:
```
4 Mbps × 3600 seconds × 24 hours = 345,600 Mb/day
345,600 Mb ÷ 8 = 43,200 MB/day ≈ 42 GB/day
```

**90-Day Storage (per camera)**:
```
42 GB/day × 90 days = 3,780 GB ≈ 3.7 TB
```

### Total System (500 cameras)

**90-Day Recordings**:
```
3.7 TB × 500 cameras = 1,850 TB ≈ 1.85 PB
```

**With overhead (thumbnails, exports, clips)**:
```
Recordings: 1,850 TB
Thumbnails: 50 TB
Exports: 1 TB
Clips: 10 TB
──────────────────
Total: ~1,911 TB ≈ 1.9 PB
```

**Recommended Deployment**:
- MinIO distributed mode (4+ nodes)
- Erasure coding (EC:4)
- Total raw capacity: ~2.5 PB (with 25% overhead)

## Deployment

### Development (Single Node)

```yaml
services:
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"
```

### Production (Distributed Mode)

**Minimum**: 4 nodes (for erasure coding EC:4)

```bash
# Node 1
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=admin \
  -e MINIO_ROOT_PASSWORD=password \
  minio/minio server \
    http://node{1...4}/data{1...4} \
    --console-address ":9001"

# Repeat for nodes 2-4
```

**With Docker Swarm**:
```yaml
services:
  minio:
    image: minio/minio:latest
    command: server http://minio{1...4}/data --console-address ":9001"
    deploy:
      replicas: 4
      placement:
        constraints:
          - node.labels.storage == true
    volumes:
      - /mnt/data:/data
```

## Initialization

Buckets and policies are initialized automatically via `minio-init` container:

```bash
# Runs automatically on first startup
docker-compose up -d minio minio-init

# Check initialization logs
docker-compose logs minio-init
```

**Manual initialization**:
```bash
# Install MinIO client
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc

# Configure alias
mc alias set minio http://localhost:9000 admin password

# Create buckets
mc mb minio/cctv-recordings
mc mb minio/cctv-exports
mc mb minio/cctv-thumbnails
mc mb minio/cctv-clips

# Set lifecycle policy
mc ilm add --expiry-days 90 minio/cctv-recordings
```

## Access Patterns

### Recording Service (Write)

```go
// Upload segment
client.PutObject(ctx, "cctv-recordings",
    "camera-123/2025/01/23/00-00-00.ts",
    file, fileSize,
    minio.PutObjectOptions{ContentType: "video/mp2t"})
```

### Playback Service (Read)

```go
// Get segment
object := client.GetObject(ctx, "cctv-recordings",
    "camera-123/2025/01/23/00-00-00.ts",
    minio.GetObjectOptions{})
```

### Temporary Export URL

```go
// Generate signed URL (valid for 1 hour)
url := client.PresignedGetObject(ctx, "cctv-exports",
    "export-abc123.mp4",
    1*time.Hour,
    url.Values{})
```

## Monitoring

### MinIO Metrics (Prometheus)

Available at `http://minio:9000/minio/v2/metrics/cluster`

Key metrics:
```
# Storage
minio_cluster_capacity_usable_total_bytes
minio_cluster_capacity_usable_free_bytes

# Performance
minio_s3_requests_total{api}
minio_s3_requests_errors_total{api}
minio_s3_time_ttfb_seconds_distribution{api}

# Objects
minio_bucket_objects_count{bucket}
minio_bucket_usage_total_bytes{bucket}
```

### MinIO Console

Web UI available at `http://localhost:9001`

Features:
- Bucket browser
- Object explorer
- User management
- Metrics dashboard
- Configuration

## Security

### Production Recommendations

1. **Enable TLS**:
   ```bash
   # Generate certificates
   openssl req -new -x509 -nodes -days 365 \
     -keyout /certs/private.key \
     -out /certs/public.crt

   # Mount certificates
   -v /path/to/certs:/root/.minio/certs
   ```

2. **Rotate credentials**:
   ```bash
   # Change root password
   mc admin user passwd minio admin

   # Rotate service keys
   mc admin user svcacct add minio recording-service
   ```

3. **Enable versioning**:
   ```bash
   mc version enable minio/cctv-clips
   ```

4. **Enable encryption**:
   ```bash
   mc encrypt set sse-s3 minio/cctv-recordings
   ```

## Troubleshooting

### Bucket Not Found

```bash
# List buckets
mc ls minio

# Check bucket exists
mc stat minio/cctv-recordings
```

### Access Denied

```bash
# Check user policy
mc admin user info minio recording-service

# Test credentials
mc alias set test http://localhost:9000 recording-service password
mc ls test/cctv-recordings
```

### Slow Uploads

```bash
# Check network bandwidth
iperf3 -s  # On MinIO server
iperf3 -c minio-server  # On client

# Check disk I/O
iostat -x 1

# Check MinIO metrics
curl http://localhost:9000/minio/v2/metrics/cluster
```

### Storage Full

```bash
# Check usage
mc du minio/cctv-recordings

# Force cleanup
mc ilm edit minio/cctv-recordings --expiry-days 60

# Manually delete old objects
mc rm --recursive --force --older-than 90d minio/cctv-recordings
```

## References

- **MinIO Documentation**: https://min.io/docs
- **MinIO Client (mc)**: https://min.io/docs/minio/linux/reference/minio-mc.html
- **Distributed Mode**: https://min.io/docs/minio/linux/operations/install-deploy-manage/deploy-minio-multi-node-multi-drive.html
- **Lifecycle Management**: https://min.io/docs/minio/linux/administration/object-management/object-lifecycle-management.html
