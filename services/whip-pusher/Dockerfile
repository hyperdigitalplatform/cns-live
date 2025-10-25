# GStreamer WHIP Pusher with Rust plugins
# Pulls RTSP streams and pushes to LiveKit WHIP endpoints
# No transcoding - just RTP repackaging for low latency

FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install GStreamer and dependencies
RUN apt-get update && apt-get install -y \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    gstreamer1.0-plugins-ugly \
    gstreamer1.0-libav \
    gstreamer1.0-nice \
    libgstreamer1.0-dev \
    libgstreamer-plugins-base1.0-dev \
    libgstreamer-plugins-bad1.0-dev \
    curl \
    ca-certificates \
    git \
    build-essential \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Clone and build gst-plugins-rs (WebRTC plugins including whipclientsink)
WORKDIR /tmp
RUN git clone https://gitlab.freedesktop.org/gstreamer/gst-plugins-rs.git && \
    cd gst-plugins-rs && \
    cargo build --release --package gst-plugin-webrtchttp && \
    cp target/release/*.so /usr/lib/x86_64-linux-gnu/gstreamer-1.0/ && \
    cd .. && \
    rm -rf gst-plugins-rs /root/.cargo/registry /root/.cargo/git

# Create app directory
WORKDIR /app

# Copy pusher script
COPY pusher.sh /app/pusher.sh
RUN chmod +x /app/pusher.sh

# Health check script
COPY healthcheck.sh /app/healthcheck.sh
RUN chmod +x /app/healthcheck.sh

# Expose health check port
EXPOSE 8090

# Run the pusher
ENTRYPOINT ["/app/pusher.sh"]
