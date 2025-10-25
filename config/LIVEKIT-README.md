# LiveKit Live Streaming Configuration

This directory contains the configuration for LiveKit SFU (Selective Forwarding Unit) used for low-latency live streaming in the RTA CCTV system.

## Overview

LiveKit provides WebRTC-based live streaming with sub-1-second latency for viewing camera feeds in real-time.

### Architecture

```
MediaMTX (RTSP) → LiveKit Ingress → LiveKit SFU → WebRTC Clients
                                           ↓
                                      Valkey (State)
                                           ↓
                                   coturn (TURN Server)
```

## Components

### 1. LiveKit SFU (`livekit.yaml`)
- **Purpose**: Main WebRTC server for real-time streaming
- **Port**: 7880 (HTTP), 7881 (TCP), 50000-50500/udp (WebRTC)
- **Resources**: 2-4 CPU cores, 2-4 GB RAM
- **Latency Target**: <800ms glass-to-glass

**Key Features**:
- Simulcast support (3 quality layers: 1080p, 720p, 360p)
- Dynamic quality switching (dynacast)
- Congestion control for bandwidth adaptation
- TCP fallback for restrictive networks
- Valkey-backed distributed state

### 2. LiveKit Ingress (`livekit-ingress.yaml`)
- **Purpose**: RTSP to WebRTC bridge
- **Port**: 8086 (health), 8087 (metrics)
- **Resources**: 1-2 CPU cores, 1-2 GB RAM

**Key Features**:
- RTSP ingestion from MediaMTX
- H.264 copy mode (NO transcoding)
- Automatic room creation
- Automatic reconnection

### 3. coturn TURN Server (`turnserver.conf`)
- **Purpose**: NAT traversal for WebRTC
- **Ports**: 3478 (UDP), 5349 (TLS)
- **Resources**: 0.5-1 CPU core, 256-512 MB RAM

**Key Features**:
- Long-term credential mechanism
- TLS support for secure connections
- Bandwidth quotas per user
- Prometheus metrics

## Configuration

### Environment Variables

```bash
# LiveKit API credentials
LIVEKIT_API_KEY=your-api-key-here
LIVEKIT_API_SECRET=your-secret-here
LIVEKIT_WEBHOOK_KEY=your-webhook-key

# TURN Server
TURN_DOMAIN=turn.rta.ae
EXTERNAL_IP=your-public-ip
TURN_USER=rta
TURN_PASSWORD=your-turn-password
TURN_SECRET=your-turn-secret

# Valkey
VALKEY_PASSWORD=your-valkey-password
```

### Simulcast Layers

The system uses 3 quality layers for adaptive streaming:

| Layer | Resolution | Bitrate | FPS | Use Case |
|-------|------------|---------|-----|----------|
| High | 1920x1080 | 3 Mbps | 25 | High-bandwidth, detail viewing |
| Medium | 1280x720 | 1.5 Mbps | 25 | Normal viewing |
| Low | 640x360 | 500 Kbps | 15 | Low-bandwidth, grid viewing |

LiveKit automatically switches between layers based on:
- Available bandwidth
- Network conditions
- Client request (manual quality selection)

## Room Management

### Room Naming Convention

```
camera_{camera_id}
```

Example: `camera_550e8400-e29b-41d4-a716-446655440000`

### Room Lifecycle

1. **Creation**: Room created when first viewer requests stream
2. **Active**: Room stays active while viewers are connected
3. **Cleanup**: Room destroyed 60s after last viewer leaves

### Token Generation

LiveKit uses JWT tokens for authentication:

```go
import (
    lksdk "github.com/livekit/server-sdk-go"
)

func GenerateToken(roomName, participantName string, canPublish bool) (string, error) {
    at := lksdk.NewAccessToken(apiKey, apiSecret)
    grant := &lksdk.VideoGrant{
        RoomJoin: true,
        Room:     roomName,
    }

    if canPublish {
        grant.CanPublish = &canPublish
    }

    at.AddGrant(grant).
        SetIdentity(participantName).
        SetValidFor(time.Hour)

    return at.ToJWT()
}
```

## Monitoring

### Prometheus Metrics

LiveKit exposes metrics on port 7882:

```bash
curl http://localhost:7882/metrics
```

**Key Metrics**:
- `livekit_room_total` - Total number of active rooms
- `livekit_participant_total` - Total participants across all rooms
- `livekit_packet_total` - Total packets sent/received
- `livekit_bytes_total` - Total bytes transferred
- `livekit_nack_total` - Packet retransmissions (higher = network issues)

### Health Checks

```bash
# LiveKit Server
curl http://localhost:7880/

# LiveKit Ingress
curl http://localhost:8086/health
```

## Performance Tuning

### CPU Optimization

For 500 concurrent cameras:
- **LiveKit SFU**: 2-4 cores (scales with viewer count)
- **LiveKit Ingress**: 1-2 cores (scales with camera count)
- **coturn**: 0.5-1 core

### Memory Optimization

Each active room consumes ~10 MB:
- 500 rooms × 10 MB = ~5 GB
- Add 1-2 GB for overhead
- **Total**: 6-7 GB for LiveKit SFU

### Network Optimization

**Bandwidth per stream**:
- High quality: 3 Mbps
- Medium quality: 1.5 Mbps
- Low quality: 500 Kbps

**Total bandwidth** (500 cameras, all high quality):
- Ingress: 500 × 3 Mbps = 1.5 Gbps
- Egress: Depends on viewer count and quality selection

**Recommended**: 10 Gbps network interface

## Troubleshooting

### Issue: High Latency

**Symptoms**: Glass-to-glass latency >1 second

**Solutions**:
1. Check network latency to LiveKit server
2. Reduce simulcast layers in `livekit.yaml`
3. Enable TCP fallback if UDP is blocked
4. Check congestion control settings

```bash
# Test network latency
ping -c 10 livekit-server-ip

# Check LiveKit logs
docker logs cctv-livekit
```

### Issue: Connection Failures

**Symptoms**: Clients cannot connect to LiveKit

**Solutions**:
1. Verify TURN server is running
2. Check firewall rules (ports 50000-50500/udp)
3. Verify external IP is correctly set
4. Test TURN server connectivity

```bash
# Check TURN server
docker logs cctv-coturn

# Test TURN connectivity
turnutils_uclient -v -u $TURN_USER -w $TURN_PASSWORD $TURN_DOMAIN
```

### Issue: Poor Video Quality

**Symptoms**: Pixelated or blurry video

**Solutions**:
1. Check source camera bitrate
2. Verify simulcast layers configuration
3. Monitor available bandwidth
4. Check packet loss (NACK metrics)

```bash
# Check LiveKit metrics
curl http://localhost:7882/metrics | grep nack

# Check ingress logs
docker logs cctv-livekit-ingress
```

### Issue: Rooms Not Cleaning Up

**Symptoms**: Memory usage keeps increasing

**Solutions**:
1. Check `empty_timeout` setting (default: 60s)
2. Verify Valkey connection
3. Check for stuck participants

```bash
# List active rooms
curl http://localhost:7880/twirp/livekit.RoomService/ListRooms \
  -H "Authorization: Bearer $LIVEKIT_TOKEN"
```

## Security

### JWT Token Security

- **Expiration**: Tokens expire after 1 hour
- **Scope**: Limited to specific room
- **Permissions**: Separate publish/subscribe grants

### TURN Server Security

- **Authentication**: Long-term credentials
- **TLS**: Enabled on port 5349
- **IP Filtering**: Allowed/denied peer IPs configured
- **Quotas**: Per-user bandwidth limits

### Network Security

- **Firewall**: Only required ports exposed
- **Internal**: Services communicate on private network
- **External**: Only LiveKit and TURN exposed to internet

## Development

### Local Testing

```bash
# Start LiveKit locally
docker-compose up livekit livekit-ingress coturn

# Generate test token
go run scripts/generate-token.go --room test-room --name test-user

# Test with browser
# Open: https://example.livekit.cloud/custom?url=ws://localhost:7880&token=YOUR_TOKEN
```

### Load Testing

Use LiveKit's load testing tool:

```bash
# Install
go install github.com/livekit/livekit-cli/lk@latest

# Load test: 100 rooms, 10 viewers each
lk load-test \
  --url ws://localhost:7880 \
  --api-key $LIVEKIT_API_KEY \
  --api-secret $LIVEKIT_API_SECRET \
  --room 100 \
  --publishers 1 \
  --subscribers 10 \
  --duration 5m
```

## Production Deployment

### Scaling Considerations

**Horizontal Scaling**:
- Deploy multiple LiveKit instances behind load balancer
- Use Valkey cluster for shared state
- TURN server can be scaled independently

**Vertical Scaling**:
- Increase CPU cores for more concurrent streams
- Increase memory for more active rooms
- Increase network bandwidth for higher quality

### High Availability

```yaml
# docker-compose.prod.yml
livekit:
  deploy:
    replicas: 3
    update_config:
      parallelism: 1
      delay: 30s
    rollback_config:
      parallelism: 1
    placement:
      constraints:
        - node.role == worker
```

### Backup & Recovery

**State Storage**:
- LiveKit state is stored in Valkey (ephemeral)
- No persistent data to backup
- Rooms automatically recreate on failure

**Configuration Backup**:
```bash
# Backup configurations
tar -czf livekit-config-backup.tar.gz \
  config/livekit.yaml \
  config/livekit-ingress.yaml \
  config/turnserver.conf
```

## Resources

- [LiveKit Documentation](https://docs.livekit.io/)
- [LiveKit Server SDK](https://github.com/livekit/server-sdk-go)
- [LiveKit Client SDK](https://docs.livekit.io/client-sdks/)
- [coturn Documentation](https://github.com/coturn/coturn)

## Support

For LiveKit issues:
1. Check logs: `docker logs cctv-livekit`
2. Review metrics: http://localhost:7882/metrics
3. Test connectivity: Use LiveKit CLI tools
4. Contact: LiveKit Community Slack
