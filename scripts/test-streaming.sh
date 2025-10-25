#!/bin/bash
# Test RTSP streaming with MediaMTX
# This script demonstrates how to publish and view RTSP streams

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=== MediaMTX RTSP Streaming Test ==="
echo ""

# Check if MediaMTX is running
echo -n "Checking MediaMTX status... "
if curl -s -f http://localhost:9997/v3/config/get > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo "MediaMTX is not running. Please start docker-compose first."
    exit 1
fi
echo ""

# Check if test video file exists
TEST_VIDEO="test.mp4"
if [ ! -f "$TEST_VIDEO" ]; then
    echo -e "${YELLOW}Test video not found. You can create one with:${NC}"
    echo "  ffmpeg -f lavfi -i testsrc=duration=60:size=1280x720:rate=30 \\"
    echo "         -f lavfi -i sine=frequency=1000:duration=60 \\"
    echo "         -c:v libx264 -preset ultrafast -tune zerolatency \\"
    echo "         -c:a aac test.mp4"
    echo ""
fi

echo "=== Test 1: Publish Test Stream ==="
echo ""
echo "You can publish a test stream with FFmpeg:"
echo -e "${GREEN}ffmpeg -re -stream_loop -1 -i test.mp4 -c copy -f rtsp rtsp://localhost:8554/test${NC}"
echo ""
echo "Or use a webcam (if available):"
echo -e "${GREEN}ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -preset ultrafast -f rtsp rtsp://localhost:8554/webcam${NC}"
echo ""

echo "=== Test 2: View Stream ==="
echo ""
echo "View with VLC:"
echo -e "${GREEN}vlc rtsp://localhost:8554/test${NC}"
echo ""
echo "View with FFplay:"
echo -e "${GREEN}ffplay -rtsp_transport tcp rtsp://localhost:8554/test${NC}"
echo ""

echo "=== Test 3: HLS Playback (for browsers) ==="
echo ""
echo "Once stream is published, access HLS at:"
echo -e "${GREEN}http://localhost:8888/test/index.m3u8${NC}"
echo ""
echo "View in browser with hls.js or native HLS support (Safari)"
echo ""

echo "=== Test 4: WebRTC Playback (ultra-low latency) ==="
echo ""
echo "WebRTC WHEP endpoint:"
echo -e "${GREEN}http://localhost:8889/test/whep${NC}"
echo ""
echo "Use a WebRTC client to connect to this endpoint"
echo ""

echo "=== Test 5: Simulate Milestone Camera Stream ==="
echo ""
echo "Simulate pulling from Milestone VMS with on-demand:"
echo ""
echo "1. First, start a mock Milestone stream:"
echo -e "${GREEN}   ffmpeg -re -stream_loop -1 -i test.mp4 -c copy -f rtsp rtsp://localhost:8554/milestone_source${NC}"
echo ""
echo "2. Configure MediaMTX to pull from it on demand:"
echo -e "${GREEN}   curl -X POST http://localhost:9997/v3/config/paths/add/milestone_123 \\${NC}"
echo -e "${GREEN}     -H 'Content-Type: application/json' \\${NC}"
echo -e "${GREEN}     -d '{${NC}"
echo -e "${GREEN}       \"source\": \"publisher\",${NC}"
echo -e "${GREEN}       \"runOnDemand\": \"ffmpeg -i rtsp://localhost:8554/milestone_source -c copy -f rtsp rtsp://localhost:8554/milestone_123\",${NC}"
echo -e "${GREEN}       \"runOnDemandRestart\": true,${NC}"
echo -e "${GREEN}       \"runOnDemandCloseAfter\": \"10s\"${NC}"
echo -e "${GREEN}     }'${NC}"
echo ""
echo "3. Now view the stream (will auto-start FFmpeg pull):"
echo -e "${GREEN}   vlc rtsp://localhost:8554/milestone_123${NC}"
echo ""

echo "=== Test 6: Check Active Streams ==="
echo ""
echo "List all active paths:"
echo -e "${GREEN}curl http://localhost:9997/v3/paths/list | jq '.items[] | {name: .name, ready: .ready, readers: .readers}'${NC}"
echo ""
echo "Check RTSP connections:"
echo -e "${GREEN}curl http://localhost:9997/v3/rtspconns/list | jq .${NC}"
echo ""

echo "=== Test 7: Multi-Viewer Test ==="
echo ""
echo "Test multiple viewers on same stream:"
echo "1. Publish once: ffmpeg -re -i test.mp4 -c copy -f rtsp rtsp://localhost:8554/multitest"
echo "2. Open multiple VLC instances:"
echo "   - vlc rtsp://localhost:8554/multitest"
echo "   - vlc rtsp://localhost:8554/multitest"
echo "   - vlc rtsp://localhost:8554/multitest"
echo ""
echo "3. Check viewer count:"
echo -e "${GREEN}curl http://localhost:9997/v3/paths/get/multitest | jq .readers${NC}"
echo ""

echo "=== Test 8: Performance Metrics ==="
echo ""
echo "View Prometheus metrics:"
echo -e "${GREEN}curl http://localhost:9998/metrics | grep mediamtx${NC}"
echo ""
echo "Key metrics to watch:"
echo "  - mediamtx_paths_count: Number of active streams"
echo "  - mediamtx_rtsp_conns_count: Active RTSP connections"
echo "  - mediamtx_hls_sessions_count: Active HLS viewers"
echo "  - mediamtx_path_bytes_received: Bytes received per stream"
echo "  - mediamtx_path_bytes_sent: Bytes sent per stream"
echo ""

echo "=== Integration Flow Example ==="
echo ""
echo "Complete workflow (VMS Service → Stream Counter → MediaMTX):"
echo ""
echo "1. Get camera RTSP URL from VMS Service:"
echo -e "${GREEN}   RTSP_URL=\$(curl -s http://localhost:8081/vms/cameras/123e4567-e89b-12d3-a456-426614174000/stream | jq -r .rtsp_url)${NC}"
echo ""
echo "2. Reserve stream quota:"
echo -e "${GREEN}   RESERVATION=\$(curl -s -X POST http://localhost:8087/api/v1/stream/reserve \\${NC}"
echo -e "${GREEN}     -H 'Content-Type: application/json' \\${NC}"
echo -e "${GREEN}     -d '{\"camera_id\":\"123e4567-e89b-12d3-a456-426614174000\",\"user_id\":\"user123\",\"source\":\"DUBAI_POLICE\",\"duration\":3600}')${NC}"
echo ""
echo "3. Create MediaMTX path to pull from Milestone:"
echo -e "${GREEN}   curl -X POST http://localhost:9997/v3/config/paths/add/camera_123 \\${NC}"
echo -e "${GREEN}     -H 'Content-Type: application/json' \\${NC}"
echo -e "${GREEN}     -d \"{\\\"source\\\":\\\"publisher\\\",\\\"runOnDemand\\\":\\\"ffmpeg -i \$RTSP_URL -c copy -f rtsp rtsp://localhost:8554/camera_123\\\"}\"${NC}"
echo ""
echo "4. Return stream URL to client:"
echo -e "${GREEN}   {${NC}"
echo -e "${GREEN}     \"rtsp\": \"rtsp://mediamtx:8554/camera_123\",${NC}"
echo -e "${GREEN}     \"hls\": \"http://mediamtx:8888/camera_123/index.m3u8\",${NC}"
echo -e "${GREEN}     \"webrtc\": \"http://mediamtx:8889/camera_123/whep\"${NC}"
echo -e "${GREEN}   }${NC}"
echo ""

echo "For more information, see:"
echo "  - config/README.md (MediaMTX integration guide)"
echo "  - https://github.com/bluenviron/mediamtx (MediaMTX docs)"
echo ""
