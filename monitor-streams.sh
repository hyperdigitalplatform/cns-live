#!/bin/bash

# Stream Counter Monitor
# Usage: ./monitor-streams.sh

echo "=========================================="
echo "   RTA CCTV Stream Counter Monitor"
echo "=========================================="
echo ""

while true; do
    clear
    echo "=========================================="
    echo "   Stream Counter Status"
    echo "   $(date '+%Y-%m-%d %H:%M:%S')"
    echo "=========================================="
    echo ""

    # Get stream stats from API
    stats=$(curl -s http://localhost:8000/api/v1/stream/stats)

    # Parse and display
    active=$(echo $stats | grep -o '"active_streams":[0-9]*' | grep -o '[0-9]*')
    viewers=$(echo $stats | grep -o '"total_viewers":[0-9]*' | grep -o '[0-9]*')

    echo "📊 Overview:"
    echo "   Active Streams: $active"
    echo "   Total Viewers:  $viewers"
    echo ""

    # Get counts from Valkey
    echo "📈 Source Counts (from Valkey):"
    dubai=$(docker exec cctv-valkey valkey-cli GET "stream:count:DUBAI_POLICE" 2>/dev/null || echo "0")
    metro=$(docker exec cctv-valkey valkey-cli GET "stream:count:METRO" 2>/dev/null || echo "0")
    parking=$(docker exec cctv-valkey valkey-cli GET "stream:count:PARKING" 2>/dev/null || echo "0")
    taxi=$(docker exec cctv-valkey valkey-cli GET "stream:count:TAXI" 2>/dev/null || echo "0")

    echo "   DUBAI_POLICE:    $dubai"
    echo "   METRO:           $metro"
    echo "   PARKING:         $parking"
    echo "   TAXI:            $taxi"
    echo ""

    # Active cameras
    echo "📹 Active Cameras:"
    echo "$stats" | python -m json.tool 2>/dev/null | grep -A 4 "camera_id" | head -20
    echo ""

    echo "=========================================="
    echo "Press Ctrl+C to stop monitoring"
    echo "Refreshing in 3 seconds..."

    sleep 3
done
