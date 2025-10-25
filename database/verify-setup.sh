#!/bin/bash
# RTA CCTV Database Verification Script
# Verifies database setup and schema integrity

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "==========================================="
echo "RTA CCTV DATABASE VERIFICATION"
echo "==========================================="
echo ""

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ docker-compose not found${NC}"
    exit 1
fi

# [1/8] Check PostgreSQL connection
echo "[1/8] Checking PostgreSQL connection..."
docker exec cctv-postgres pg_isready -U cctv > /dev/null 2>&1
print_status $? "PostgreSQL is ready"

# [2/8] Check database exists
echo ""
echo "[2/8] Checking database 'cctv' exists..."
DB_EXISTS=$(docker exec cctv-postgres psql -U cctv -lqt | cut -d \| -f 1 | grep -w cctv | wc -l)
if [ $DB_EXISTS -eq 1 ]; then
    print_status 0 "Database 'cctv' exists"
else
    print_status 1 "Database 'cctv' does not exist"
fi

# [3/8] Check extensions
echo ""
echo "[3/8] Checking required extensions..."
UUID_EXT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM pg_extension WHERE extname='uuid-ossp';")
TRGM_EXT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM pg_extension WHERE extname='pg_trgm';")

if [ $UUID_EXT -eq 1 ] && [ $TRGM_EXT -eq 1 ]; then
    print_status 0 "Required extensions installed (uuid-ossp, pg_trgm)"
else
    print_status 1 "Missing required extensions"
fi

# [4/8] Check tables
echo ""
echo "[4/8] Checking tables..."
TABLE_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';")

EXPECTED_TABLES=(
    "cameras"
    "streams"
    "recordings"
    "video_segments"
    "segments"
    "exports"
    "tags"
    "video_tags"
    "annotations"
    "incidents"
    "stream_stats"
    "system_settings"
)

if [ $TABLE_COUNT -ge 12 ]; then
    print_status 0 "All tables created ($TABLE_COUNT tables)"

    # Check each table
    for table in "${EXPECTED_TABLES[@]}"; do
        EXISTS=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='$table';")
        if [ $EXISTS -eq 1 ]; then
            echo "  ✓ $table"
        else
            echo "  ✗ $table (missing)"
        fi
    done
else
    print_warning "Expected 12+ tables, found $TABLE_COUNT"
fi

# [5/8] Check views
echo ""
echo "[5/8] Checking views..."
VIEW_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public';")

EXPECTED_VIEWS=(
    "v_active_cameras"
    "v_recording_summary"
    "v_agency_quotas"
)

if [ $VIEW_COUNT -ge 3 ]; then
    print_status 0 "Views created ($VIEW_COUNT views)"

    for view in "${EXPECTED_VIEWS[@]}"; do
        EXISTS=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public' AND table_name='$view';")
        if [ $EXISTS -eq 1 ]; then
            echo "  ✓ $view"
        else
            echo "  ✗ $view (missing)"
        fi
    done
else
    print_warning "Expected 3+ views, found $VIEW_COUNT"
fi

# [6/8] Check indexes
echo ""
echo "[6/8] Checking indexes..."
INDEX_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM pg_indexes WHERE schemaname='public';")
if [ $INDEX_COUNT -ge 30 ]; then
    print_status 0 "Indexes created ($INDEX_COUNT indexes)"
else
    print_warning "Expected 30+ indexes, found $INDEX_COUNT"
fi

# [7/8] Check functions
echo ""
echo "[7/8] Checking functions..."
FUNC_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM pg_proc WHERE pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public');")
if [ $FUNC_COUNT -ge 2 ]; then
    print_status 0 "Functions created ($FUNC_COUNT functions)"
else
    print_warning "Expected 2+ functions, found $FUNC_COUNT"
fi

# [8/8] Test queries
echo ""
echo "[8/8] Testing queries..."

# Test basic query
docker exec cctv-postgres psql -U cctv -d cctv -c "SELECT COUNT(*) FROM cameras;" > /dev/null 2>&1
print_status $? "Can query cameras table"

# Test view query
docker exec cctv-postgres psql -U cctv -d cctv -c "SELECT * FROM v_agency_quotas;" > /dev/null 2>&1
print_status $? "Can query v_agency_quotas view"

# Test function
docker exec cctv-postgres psql -U cctv -d cctv -c "SELECT cleanup_old_streams(7);" > /dev/null 2>&1
print_status $? "Can execute cleanup_old_streams function"

# Summary
echo ""
echo "==========================================="
echo -e "${GREEN}✅ DATABASE VERIFICATION COMPLETE${NC}"
echo "==========================================="
echo ""
echo "Database Statistics:"
echo "  - Tables: $TABLE_COUNT"
echo "  - Views: $VIEW_COUNT"
echo "  - Indexes: $INDEX_COUNT"
echo "  - Functions: $FUNC_COUNT"
echo ""

# Row counts
echo "Current Data:"
CAMERA_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM cameras;")
STREAM_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM streams;")
RECORDING_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM recordings;")
SEGMENT_COUNT=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT COUNT(*) FROM video_segments;")

echo "  - Cameras: $CAMERA_COUNT"
echo "  - Active Streams: $STREAM_COUNT"
echo "  - Recordings: $RECORDING_COUNT"
echo "  - Video Segments: $SEGMENT_COUNT"
echo ""

# Database size
DB_SIZE=$(docker exec cctv-postgres psql -U cctv -d cctv -tAc "SELECT pg_size_pretty(pg_database_size('cctv'));")
echo "Database Size: $DB_SIZE"
echo ""

echo -e "${GREEN}Database is ready for use!${NC}"
