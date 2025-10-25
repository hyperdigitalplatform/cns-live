#!/bin/bash
set -e

# GStreamer WHIP Pusher Script
# Pulls RTSP stream from MediaMTX and pushes to LiveKit WHIP endpoint
# No transcoding for H.264 - just RTP repackaging

# Required environment variables
: ${RTSP_URL:?Error: RTSP_URL is required}
: ${WHIP_ENDPOINT:?Error: WHIP_ENDPOINT is required}
: ${STREAM_KEY:?Error: STREAM_KEY is required}

echo "Starting WHIP Pusher..."
echo "RTSP Source: ${RTSP_URL}"
echo "WHIP Endpoint: ${WHIP_ENDPOINT}"
echo "Stream Key: ${STREAM_KEY}"

# GStreamer pipeline for video streaming via WHIP
# Using whipsink from gst-plugins-rs
# Supports both H.264 (passthrough) and H.265 (transcode to H.264)
# Caps filter selects only video stream (ignores audio)
gst-launch-1.0 -v \
  rtspsrc location="${RTSP_URL}" latency=0 protocols=tcp ! \
  application/x-rtp,media=video ! \
  rtpjitterbuffer latency=100 ! \
  decodebin ! \
  x264enc tune=zerolatency speed-preset=ultrafast bitrate=2000 key-int-max=60 ! \
  h264parse ! \
  rtph264pay config-interval=-1 pt=96 ! \
  application/x-rtp,media=video,encoding-name=H264,payload=96 ! \
  whipsink whip-endpoint="${WHIP_ENDPOINT}" auth-token="${STREAM_KEY}"

# If GStreamer exits, the container will restart (handled by Docker restart policy)
echo "GStreamer pipeline exited"
