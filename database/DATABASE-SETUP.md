# RTA CCTV System - Database Setup Guide

**Version**: 1.0.0
**Database**: PostgreSQL 15
**Last Updated**: October 24, 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Automatic Setup (Docker Compose)](#automatic-setup-docker-compose)
3. [Manual Setup (Standalone PostgreSQL)](#manual-setup-standalone-postgresql)
4. [Database Schema](#database-schema)
5. [Running Migrations](#running-migrations)
6. [Verification](#verification)
7. [Backup & Restore](#backup--restore)
8. [Troubleshooting](#troubleshooting)

---

## Overview

The RTA CCTV system uses **PostgreSQL 15** as its primary database for storing:
- Camera metadata and configuration
- Active stream reservations
- Recording metadata and segments
- Incidents, tags, and annotations
- Search indexes and statistics

### Database Details

- **Database Name**: `cctv`
- **Default User**: `cctv`
- **Default Password**: Set via `POSTGRES_PASSWORD` environment variable
- **Port**: `5432`
- **Encoding**: UTF-8
- **Extensions**: `uuid-ossp`, `pg_trgm`

---

## Automatic Setup (Docker Compose)

**This is the recommended method** for both development and production.

### Step 1: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit database password
nano .env
```

Set the following in `.env`:
```bash
POSTGRES_PASSWORD=your_secure_password_here
```

### Step 2: Start PostgreSQL Container

```bash
# Start only PostgreSQL
docker-compose up -d postgres

# Wait for PostgreSQL to be ready (10-15 seconds)
sleep 15

# Check status
docker-compose ps postgres
```

**Expected output**:
```
NAME              IMAGE               STATUS
cctv-postgres     postgres:15-alpine  Up (healthy)
```

### Step 3: Verify Database Creation

```bash
# Check database exists
docker exec cctv-postgres psql -U cctv -c "\l"

# Expected output: database "cctv" should be listed
```

### Step 4: Run Migrations

```bash
# Run all migration files
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/001_create_initial_schema.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/002_create_storage_tables.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/003_create_metadata_tables.sql
```

**Expected output**:
```
CREATE EXTENSION
CREATE TABLE
CREATE INDEX
...
(migrations complete successfully)
```

### Step 5: Verify Schema

```bash
# List all tables
docker exec cctv-postgres psql -U cctv -d cctv -c "\dt"

# Expected tables:
# - cameras
# - streams
# - recordings
# - video_segments
# - segments
# - exports
# - tags
# - video_tags
# - annotations
# - incidents
# - stream_stats
# - system_settings
```

### ✅ Automatic Setup Complete

Your database is now ready! All services can connect using:
- **Host**: `postgres` (from Docker network) or `localhost` (from host)
- **Port**: `5432`
- **Database**: `cctv`
- **User**: `cctv`
- **Password**: From `POSTGRES_PASSWORD` in `.env`

---

## Manual Setup (Standalone PostgreSQL)

Use this method if you have an existing PostgreSQL instance (not using Docker).

### Prerequisites

- PostgreSQL 15+ installed
- `psql` command-line tool available
- Superuser or database creation privileges

### Step 1: Create Database and User

```bash
# Connect as postgres superuser
sudo -u postgres psql

# Or on Windows:
psql -U postgres
```

Run the following SQL:

```sql
-- Create database
CREATE DATABASE cctv
    WITH
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TEMPLATE = template0;

-- Create user with password
CREATE USER cctv WITH PASSWORD 'your_secure_password_here';

-- Grant all privileges on database
GRANT ALL PRIVILEGES ON DATABASE cctv TO cctv;

-- Connect to cctv database
\c cctv

-- Grant schema privileges
GRANT ALL PRIVILEGES ON SCHEMA public TO cctv;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cctv;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cctv;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO cctv;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL PRIVILEGES ON TABLES TO cctv;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL PRIVILEGES ON SEQUENCES TO cctv;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT EXECUTE ON FUNCTIONS TO cctv;

-- Exit
\q
```

### Step 2: Enable Required Extensions

```bash
# Connect as cctv user
psql -U cctv -d cctv

# Or if authentication fails, connect as postgres first:
sudo -u postgres psql -d cctv
```

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable trigram extension for text search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Verify extensions
\dx

-- Exit
\q
```

**Expected output**:
```
                                   List of installed extensions
   Name    | Version |   Schema   |                         Description
-----------+---------+------------+--------------------------------------------------------------
 pg_trgm   | 1.6     | public     | text similarity measurement and index searching using trigrams
 plpgsql   | 1.0     | pg_catalog | PL/pgSQL procedural language
 uuid-ossp | 1.1     | public     | generate universally unique identifiers (UUIDs)
```

### Step 3: Run Migrations

```bash
# Navigate to project directory
cd /path/to/cns

# Run migrations in order
psql -U cctv -d cctv -f database/migrations/001_create_initial_schema.sql
psql -U cctv -d cctv -f database/migrations/002_create_storage_tables.sql
psql -U cctv -d cctv -f database/migrations/003_create_metadata_tables.sql
```

**If authentication fails**, use:
```bash
psql -U cctv -d cctv -h localhost -W -f database/migrations/001_create_initial_schema.sql
# Enter password when prompted
```

### Step 4: Configure Services

Update `.env` file with your database connection:

```bash
POSTGRES_HOST=localhost          # Or your PostgreSQL server IP
POSTGRES_PORT=5432
POSTGRES_DB=cctv
POSTGRES_USER=cctv
POSTGRES_PASSWORD=your_secure_password_here
```

### Step 5: Test Connection

```bash
# Test connection from command line
psql -U cctv -d cctv -h localhost -c "SELECT NOW();"

# Expected output: current timestamp
```

---

## Database Schema

### Tables Overview

| Table | Purpose | Row Count (Est.) |
|-------|---------|------------------|
| `cameras` | Camera registry | 500-1000 |
| `streams` | Active stream sessions | 0-500 (active only) |
| `recordings` | Recording sessions | Growing (thousands) |
| `video_segments` | Individual video segments | Growing (millions) |
| `segments` | Storage segment metadata | Growing (millions) |
| `exports` | Export requests | Growing (hundreds) |
| `tags` | Tag definitions | 50-200 |
| `video_tags` | Segment tags (many-to-many) | Growing (thousands) |
| `annotations` | Timeline annotations | Growing (thousands) |
| `incidents` | Incident reports | Growing (thousands) |
| `stream_stats` | Usage statistics | Growing (time-series) |
| `system_settings` | System configuration | ~10 (static) |

### Key Relationships

```
cameras (1) ──< (many) streams
cameras (1) ──< (many) recordings
recordings (1) ──< (many) video_segments
video_segments (1) ──< (many) video_tags ──> (1) tags
video_segments (1) ──< (many) annotations
```

### Views

- `v_active_cameras` - Active cameras with stream counts
- `v_recording_summary` - Recording statistics per camera
- `v_agency_quotas` - Stream usage vs limits by agency

### Functions

- `cleanup_old_streams(days_old)` - Cleanup old stream records
- `get_camera_availability(camera_id, start_time, end_time)` - Check recording availability

---

## Running Migrations

### Migration Files

Located in `database/migrations/`:

1. **`001_create_initial_schema.sql`** - Core tables (cameras, streams, recordings)
2. **`002_create_storage_tables.sql`** - Storage and export tables
3. **`003_create_metadata_tables.sql`** - Metadata (tags, annotations, incidents)

### Migration Order

**IMPORTANT**: Migrations must be run in numerical order!

```bash
# Method 1: Docker (if using docker-compose)
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/001_create_initial_schema.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/002_create_storage_tables.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/003_create_metadata_tables.sql

# Method 2: Direct psql (standalone PostgreSQL)
psql -U cctv -d cctv -f database/migrations/001_create_initial_schema.sql
psql -U cctv -d cctv -f database/migrations/002_create_storage_tables.sql
psql -U cctv -d cctv -f database/migrations/003_create_metadata_tables.sql

# Method 3: All at once (bash)
for file in database/migrations/*.sql; do
    echo "Running $file..."
    docker exec -i cctv-postgres psql -U cctv -d cctv < "$file"
done
```

### Verify Migrations

```bash
# Check tables created
docker exec cctv-postgres psql -U cctv -d cctv -c "\dt"

# Count tables (should be 12)
docker exec cctv-postgres psql -U cctv -d cctv -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';"

# Check extensions
docker exec cctv-postgres psql -U cctv -d cctv -c "\dx"

# Check views (should be 3)
docker exec cctv-postgres psql -U cctv -d cctv -c "\dv"

# Check functions
docker exec cctv-postgres psql -U cctv -d cctv -c "\df"
```

---

## Verification

### Quick Verification Script

```bash
#!/bin/bash
# database/verify-setup.sh

echo "=== RTA CCTV Database Verification ==="
echo ""

# Check PostgreSQL is running
echo "[1/6] Checking PostgreSQL connection..."
docker exec cctv-postgres pg_isready -U cctv
if [ $? -eq 0 ]; then
    echo "✅ PostgreSQL is ready"
else
    echo "❌ PostgreSQL is not ready"
    exit 1
fi

# Check database exists
echo ""
echo "[2/6] Checking database 'cctv' exists..."
DB_EXISTS=$(docker exec cctv-postgres psql -U cctv -lqt | cut -d \| -f 1 | grep -w cctv | wc -l)
if [ $DB_EXISTS -eq 1 ]; then
    echo "✅ Database 'cctv' exists"
else
    echo "❌ Database 'cctv' does not exist"
    exit 1
fi

# Check extensions
echo ""
echo "[3/6] Checking required extensions..."
EXTENSIONS=$(docker exec cctv-postgres psql -U cctv -d cctv -c "\dx" | grep -E "(uuid-ossp|pg_trgm)" | wc -l)
if [ $EXTENSIONS -eq 2 ]; then
    echo "✅ Required extensions installed (uuid-ossp, pg_trgm)"
else
    echo "❌ Missing required extensions"
    exit 1
fi

# Check tables
echo ""
echo "[4/6] Checking tables..."
TABLE_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';")
if [ $TABLE_COUNT -ge 12 ]; then
    echo "✅ All tables created ($TABLE_COUNT tables)"
else
    echo "⚠️  Expected 12+ tables, found $TABLE_COUNT"
fi

# Check views
echo ""
echo "[5/6] Checking views..."
VIEW_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public';")
if [ $VIEW_COUNT -ge 3 ]; then
    echo "✅ Views created ($VIEW_COUNT views)"
else
    echo "⚠️  Expected 3+ views, found $VIEW_COUNT"
fi

# Test query
echo ""
echo "[6/6] Testing query..."
docker exec cctv-postgres psql -U cctv -d cctv -c "SELECT COUNT(*) FROM cameras;" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Can query tables successfully"
else
    echo "❌ Cannot query tables"
    exit 1
fi

echo ""
echo "=== ✅ Database verification complete ==="
```

Make it executable and run:
```bash
chmod +x database/verify-setup.sh
./database/verify-setup.sh
```

### Manual Verification

```sql
-- Connect to database
docker exec -it cctv-postgres psql -U cctv -d cctv

-- Check version
SELECT version();

-- List all tables
\dt

-- Check table structure
\d cameras
\d streams
\d recordings
\d video_segments

-- Check indexes
\di

-- Check views
\dv

-- Query system settings
SELECT * FROM system_settings;

-- Check empty tables (should return 0 for all)
SELECT 'cameras' AS table_name, COUNT(*) FROM cameras
UNION ALL
SELECT 'streams', COUNT(*) FROM streams
UNION ALL
SELECT 'recordings', COUNT(*) FROM recordings;

-- Exit
\q
```

---

## Backup & Restore

### Backup Database

```bash
# Using docker-compose (recommended)
./scripts/backup.sh

# Or manual backup
docker exec cctv-postgres pg_dump -U cctv -d cctv -F c -f /tmp/cctv_backup.dump
docker cp cctv-postgres:/tmp/cctv_backup.dump ./backups/cctv_$(date +%Y%m%d_%H%M%S).dump

# Or SQL format
docker exec cctv-postgres pg_dumpall -U cctv | gzip > backups/cctv_$(date +%Y%m%d_%H%M%S).sql.gz
```

### Restore Database

```bash
# Using docker-compose (recommended)
./scripts/restore.sh <timestamp>

# Or manual restore from custom format
docker cp backups/cctv_20251024_120000.dump cctv-postgres:/tmp/restore.dump
docker exec cctv-postgres pg_restore -U cctv -d cctv -c /tmp/restore.dump

# Or from SQL format
gunzip -c backups/cctv_20251024_120000.sql.gz | docker exec -i cctv-postgres psql -U cctv
```

### Backup Schedule (Production)

Add to crontab:
```bash
# Daily backup at 2 AM
0 2 * * * /path/to/cns/scripts/backup.sh

# Weekly full backup at Sunday 3 AM
0 3 * * 0 docker exec cctv-postgres pg_dumpall -U cctv | gzip > /mnt/backups/weekly/cctv_$(date +%Y%m%d).sql.gz
```

---

## Troubleshooting

### Issue 1: "role 'cctv' does not exist"

**Solution**:
```bash
# Docker: The role is created automatically by POSTGRES_USER environment variable
# Check docker-compose.yml has POSTGRES_USER=cctv

# Standalone: Create user manually
sudo -u postgres psql -c "CREATE USER cctv WITH PASSWORD 'password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE cctv TO cctv;"
```

### Issue 2: "database 'cctv' does not exist"

**Solution**:
```bash
# Docker: Database is created automatically by POSTGRES_DB environment variable
# Restart postgres container:
docker-compose restart postgres

# Standalone: Create database manually
sudo -u postgres psql -c "CREATE DATABASE cctv OWNER cctv;"
```

### Issue 3: "extension 'uuid-ossp' does not exist"

**Solution**:
```bash
# Connect as superuser and enable extension
docker exec -it cctv-postgres psql -U cctv -d cctv -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"

# Or as postgres user:
docker exec -it cctv-postgres psql -U postgres -d cctv -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
```

### Issue 4: "permission denied for schema public"

**Solution**:
```sql
-- Connect as postgres superuser
docker exec -it cctv-postgres psql -U postgres -d cctv

-- Grant permissions
GRANT ALL PRIVILEGES ON SCHEMA public TO cctv;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cctv;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cctv;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO cctv;
```

### Issue 5: "relation 'cameras' does not exist"

**Solution**:
```bash
# Migrations were not run. Run them:
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/001_create_initial_schema.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/002_create_storage_tables.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/003_create_metadata_tables.sql
```

### Issue 6: "connection refused"

**Solution**:
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check port is exposed
docker-compose port postgres 5432

# Check from host
psql -U cctv -d cctv -h localhost -p 5432 -c "SELECT 1;"

# Check logs
docker-compose logs postgres
```

### Issue 7: Migration fails with duplicate key error

**Solution**:
```bash
# Drop and recreate database (WARNING: loses all data)
docker exec -it cctv-postgres psql -U postgres -c "DROP DATABASE IF EXISTS cctv;"
docker exec -it cctv-postgres psql -U postgres -c "CREATE DATABASE cctv OWNER cctv;"

# Re-run migrations
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/001_create_initial_schema.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/002_create_storage_tables.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/003_create_metadata_tables.sql
```

---

## Performance Tuning

### Recommended Settings for Production

Edit `postgresql.conf` or set via Docker environment:

```ini
# Memory Settings
shared_buffers = 4GB                # 25% of total RAM
effective_cache_size = 12GB         # 75% of total RAM
maintenance_work_mem = 1GB
work_mem = 64MB

# Connection Settings
max_connections = 200
superuser_reserved_connections = 3

# Query Planning
random_page_cost = 1.1              # For SSD storage
effective_io_concurrency = 200

# Write Performance
wal_buffers = 16MB
checkpoint_completion_target = 0.9
max_wal_size = 4GB
min_wal_size = 1GB

# Logging
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_min_duration_statement = 1000   # Log queries > 1 second
```

### Index Maintenance

```sql
-- Vacuum and analyze regularly
VACUUM ANALYZE;

-- Reindex if performance degrades
REINDEX DATABASE cctv;

-- Check index usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;

-- Check table bloat
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## Monitoring

### Key Metrics to Watch

```sql
-- Active connections
SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active';

-- Database size
SELECT pg_size_pretty(pg_database_size('cctv'));

-- Table sizes
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Slow queries (requires pg_stat_statements extension)
SELECT
    query,
    calls,
    mean_exec_time,
    max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Cache hit ratio (should be >99%)
SELECT
    sum(heap_blks_read) as heap_read,
    sum(heap_blks_hit)  as heap_hit,
    sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as ratio
FROM pg_statio_user_tables;
```

### Prometheus Exporter

The system includes `postgres_exporter` for Prometheus monitoring:

```bash
# Check exporter metrics
curl http://localhost:9187/metrics | grep pg_
```

---

## Next Steps

After database setup:

1. ✅ **Verify Setup**: Run `./database/verify-setup.sh`
2. ✅ **Start Services**: `docker-compose up -d`
3. ✅ **Run Health Check**: `./scripts/health-check.sh`
4. ✅ **Configure Backups**: Add cron job for daily backups
5. ✅ **Monitor Performance**: Check Grafana dashboards

---

## Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/15/)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Docker PostgreSQL](https://hub.docker.com/_/postgres)
- [pg_stat_statements](https://www.postgresql.org/docs/current/pgstatstatements.html)

---

**Database Setup Version**: 1.0.0
**Platform Version**: 1.0.0
**PostgreSQL Version**: 15
