# RTA CCTV System - Database

This directory contains database migrations and setup scripts for the RTA CCTV system.

## Quick Start

### With Docker Compose (Recommended)

```bash
# 1. Start PostgreSQL
docker-compose up -d postgres

# 2. Wait for PostgreSQL to be ready
sleep 15

# 3. Run migrations
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/001_create_initial_schema.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/002_create_storage_tables.sql
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/003_create_metadata_tables.sql

# 4. Verify setup
chmod +x database/verify-setup.sh
./database/verify-setup.sh
```

### With Standalone PostgreSQL

See [DATABASE-SETUP.md](DATABASE-SETUP.md) for complete instructions.

## Directory Structure

```
database/
├── README.md                           # This file
├── DATABASE-SETUP.md                   # Complete setup guide
├── verify-setup.sh                     # Verification script
└── migrations/                         # SQL migration files
    ├── 001_create_initial_schema.sql   # Core tables
    ├── 002_create_storage_tables.sql   # Storage tables
    └── 003_create_metadata_tables.sql  # Metadata tables
```

## Migration Files

| File | Purpose | Tables Created |
|------|---------|----------------|
| `001_create_initial_schema.sql` | Core system tables | cameras, streams, recordings, video_segments, stream_stats, system_settings |
| `002_create_storage_tables.sql` | Storage management | segments, exports |
| `003_create_metadata_tables.sql` | Metadata & search | tags, video_tags, annotations, incidents |

## Database Schema

### Core Tables

- **cameras** - Camera registry with metadata
- **streams** - Active stream reservations
- **recordings** - Recording sessions
- **video_segments** - Individual video segments
- **segments** - Storage segment metadata
- **exports** - Video export requests

### Metadata Tables

- **tags** - Tag definitions
- **video_tags** - Segment tags (many-to-many)
- **annotations** - Timeline annotations
- **incidents** - Incident reports

### System Tables

- **stream_stats** - Usage statistics (time-series)
- **system_settings** - System configuration

### Views

- **v_active_cameras** - Active cameras with stream counts
- **v_recording_summary** - Recording statistics per camera
- **v_agency_quotas** - Stream usage vs limits by agency

### Functions

- **cleanup_old_streams(days_old)** - Cleanup old stream records
- **get_camera_availability(camera_id, start_time, end_time)** - Check recording availability

## Connection Details

### Docker Compose

- **Host**: `postgres` (from Docker network) or `localhost` (from host)
- **Port**: `5432`
- **Database**: `cctv`
- **User**: `cctv`
- **Password**: From `POSTGRES_PASSWORD` in `.env`

### Connection String

```bash
# From Docker network
postgresql://cctv:${POSTGRES_PASSWORD}@postgres:5432/cctv

# From host
postgresql://cctv:${POSTGRES_PASSWORD}@localhost:5432/cctv
```

## Verification

Run the verification script:

```bash
./database/verify-setup.sh
```

**Expected output**:
```
✅ PostgreSQL is ready
✅ Database 'cctv' exists
✅ Required extensions installed
✅ All tables created (12 tables)
✅ Views created (3 views)
✅ Indexes created
✅ Functions created
✅ Can query tables successfully
```

## Troubleshooting

### Migrations Failed

```bash
# Check if database exists
docker exec cctv-postgres psql -U cctv -c "\l"

# Check if tables exist
docker exec cctv-postgres psql -U cctv -d cctv -c "\dt"

# Re-run specific migration
docker exec -i cctv-postgres psql -U cctv -d cctv < database/migrations/001_create_initial_schema.sql
```

### Connection Refused

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Permission Denied

```bash
# Connect as postgres superuser and grant permissions
docker exec -it cctv-postgres psql -U postgres -d cctv

# Grant all privileges
GRANT ALL PRIVILEGES ON SCHEMA public TO cctv;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cctv;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cctv;
```

## Backup & Restore

### Backup

```bash
# Automated backup (recommended)
./scripts/backup.sh

# Manual backup
docker exec cctv-postgres pg_dump -U cctv -d cctv -F c > backups/cctv_$(date +%Y%m%d_%H%M%S).dump
```

### Restore

```bash
# Automated restore (recommended)
./scripts/restore.sh <timestamp>

# Manual restore
docker exec -i cctv-postgres pg_restore -U cctv -d cctv -c < backups/cctv_20251024_120000.dump
```

## Performance

### Recommended PostgreSQL Settings

```ini
shared_buffers = 4GB
effective_cache_size = 12GB
maintenance_work_mem = 1GB
work_mem = 64MB
max_connections = 200
```

### Maintenance

```bash
# Vacuum and analyze
docker exec cctv-postgres psql -U cctv -d cctv -c "VACUUM ANALYZE;"

# Reindex
docker exec cctv-postgres psql -U cctv -d cctv -c "REINDEX DATABASE cctv;"

# Check database size
docker exec cctv-postgres psql -U cctv -d cctv -c "SELECT pg_size_pretty(pg_database_size('cctv'));"
```

## Documentation

- **[DATABASE-SETUP.md](DATABASE-SETUP.md)** - Complete setup guide
- **[verify-setup.sh](verify-setup.sh)** - Verification script
- **[../docs/deployment.md](../docs/deployment.md)** - Deployment guide
- **[../docs/operations.md](../docs/operations.md)** - Operations manual

## Support

For issues or questions:
1. Check [DATABASE-SETUP.md](DATABASE-SETUP.md) troubleshooting section
2. Run `./database/verify-setup.sh` for diagnostics
3. Check logs: `docker-compose logs postgres`
4. Review [../docs/operations.md](../docs/operations.md) for incident response

---

**Database Version**: 1.0.0
**PostgreSQL Version**: 15
**Last Updated**: October 24, 2025
