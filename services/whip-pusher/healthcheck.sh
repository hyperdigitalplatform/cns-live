#!/bin/bash
# Simple health check - verify GStreamer process is running
if pgrep -x "gst-launch-1.0" > /dev/null; then
    exit 0
else
    exit 1
fi
