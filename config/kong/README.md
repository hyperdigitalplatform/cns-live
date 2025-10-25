# Kong API Gateway Configuration

## Overview

Kong serves as the API Gateway for the RTA CCTV system, providing:
- **Unified API endpoint** for all services
- **Rate limiting** integration with Stream Counter
- **Request routing** to backend microservices
- **Authentication** and authorization (future)
- **Metrics** and monitoring
- **Request/Response transformation**

## Architecture

```
┌──────────────────────────────────────┐
│      Clients (Web/Mobile/APIs)      │
└────────────────┬─────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│           Kong API Gateway :8000                │
│  ┌───────────────────────────────────────────┐  │
│  │  Global Plugins                           │  │
│  │  • CORS                                   │  │
│  │  • Request ID (Correlation)              │  │
│  │  • Response Transformer                   │  │
│  │  • Prometheus Metrics                     │  │
│  └───────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────┐  │
│  │  Custom Plugins                           │  │
│  │  • Quota Validator (checks Stream Counter)│  │
│  └───────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────┐  │
│  │  Routes & Services                        │  │
│  │  • /api/v1/vms/*      → vms-service       │  │
│  │  • /api/v1/stream/*   → stream-counter    │  │
│  │  • /api/v1/rtsp/*     → mediamtx          │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
         │                │              │
         ↓                ↓              ↓
  ┌──────────┐   ┌──────────────┐   ┌──────────┐
  │   VMS    │   │    Stream    │   │ MediaMTX │
  │ Service  │   │   Counter    │   │   API    │
  └──────────┘   └──────────────┘   └──────────┘
```

## API Routes

All client requests go through Kong at `http://kong:8000`

### VMS Service Routes

| Method | Path | Backend | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/vms/cameras` | vms-service:8081 | List all cameras |
| GET | `/api/v1/vms/cameras/{id}` | vms-service:8081 | Get camera details |
| GET | `/api/v1/vms/cameras/{id}/stream` | vms-service:8081 | Get RTSP URL |
| POST | `/api/v1/vms/cameras/{id}/ptz` | vms-service:8081 | PTZ control |
| GET/POST | `/api/v1/vms/recordings` | vms-service:8081 | Recording operations |

### Stream Counter Routes

| Method | Path | Backend | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/stream/reserve` | stream-counter:8087 | Reserve stream quota |
| DELETE | `/api/v1/stream/release/{id}` | stream-counter:8087 | Release quota |
| POST | `/api/v1/stream/heartbeat/{id}` | stream-counter:8087 | Send heartbeat |
| GET | `/api/v1/stream/stats` | stream-counter:8087 | Get statistics |

### MediaMTX Routes (Admin)

| Method | Path | Backend | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/rtsp/paths` | mediamtx:9997 | List stream paths |
| GET/POST/DELETE | `/api/v1/rtsp/paths/{name}` | mediamtx:9997 | Manage paths |

## Custom Quota Validator Plugin

Located at: `config/kong/plugins/quota-validator/`

### Purpose
Validates stream quota **before** proxying reserve requests to Stream Counter. This provides:
- **Early rejection** of quota-exceeded requests
- **Reduced load** on Stream Counter
- **Caching** of quota statistics (5-second TTL)
- **Bilingual error messages** (Arabic/English)

### How It Works

1. **Client sends reserve request**:
   ```bash
   POST http://kong:8000/api/v1/stream/reserve
   {
     "camera_id": "123e4567-e89b-12d3-a456-426614174000",
     "user_id": "user123",
     "source": "DUBAI_POLICE",
     "duration": 3600
   }
   ```

2. **Quota Validator checks stats** (cached for 5s):
   ```bash
   GET http://stream-counter:8087/api/v1/stream/stats
   ```

3. **If quota exceeded**, return 429 immediately:
   ```json
   HTTP/1.1 429 Too Many Requests
   X-RateLimit-Limit: 50
   X-RateLimit-Remaining: 0
   Retry-After: 30

   {
     "error": {
       "code": "RATE_LIMIT_EXCEEDED",
       "message_en": "Camera limit reached for Dubai Police",
       "message_ar": "تم الوصول إلى حد الكاميرات لشرطة دبي",
       "source": "DUBAI_POLICE",
       "current": 50,
       "limit": 50,
       "available": 0
     }
   }
   ```

4. **If quota available**, add headers and proxy to Stream Counter:
   ```
   X-RateLimit-Limit: 50
   X-RateLimit-Remaining: 15
   X-Quota-Percentage: 70
   ```

### Configuration

Edit `config/kong.yml` to enable on specific routes:

```yaml
routes:
  - name: stream-reserve
    paths:
      - /api/v1/stream/reserve
    plugins:
      - name: quota-validator
        config:
          stream_counter_url: http://stream-counter:8087
          validate_before_proxy: true
          cache_ttl: 5
          timeout: 1000
          enabled_routes:
            - /api/v1/stream/reserve
```

## Global Plugins

### 1. CORS
Allows cross-origin requests from RTA domains:
- `https://rta.ae`
- `https://*.rta.ae`
- `http://localhost:3000` (development only)

### 2. Correlation ID
Adds `X-Request-ID` header to all requests for tracing.

### 3. Request Size Limiting
Max request size: 10MB

### 4. Response Transformer
Adds custom headers:
- `X-Powered-By: RTA CCTV System`
- `X-Kong-Upstream-Status: {status}`

### 5. Prometheus
Exports metrics at `http://kong:8001/metrics`:
```
# Request metrics
kong_http_requests_total{service,route,code}
kong_latency_ms{service,route,type}

# Upstream health
kong_upstream_health{upstream,address}

# Bandwidth
kong_bandwidth_bytes{service,route,type}
```

## Testing

### Test Kong Health
```bash
# Health check
curl http://localhost:8100/status

# Admin API
curl http://localhost:8001

# List routes
curl http://localhost:8001/routes | jq .

# List services
curl http://localhost:8001/services | jq .

# List plugins
curl http://localhost:8001/plugins | jq .
```

### Test VMS Routes via Kong
```bash
# List cameras (proxied through Kong)
curl http://localhost:8000/api/v1/vms/cameras

# Get camera stream URL
curl http://localhost:8000/api/v1/vms/cameras/{id}/stream
```

### Test Stream Counter via Kong
```bash
# Get stats
curl http://localhost:8000/api/v1/stream/stats

# Reserve stream
curl -X POST http://localhost:8000/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "camera_id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "user123",
    "source": "DUBAI_POLICE",
    "duration": 3600
  }'
```

### Test Quota Validator Plugin
```bash
# Reserve 50 streams (up to limit)
for i in {1..50}; do
  curl -s -X POST http://localhost:8000/api/v1/stream/reserve \
    -H "Content-Type: application/json" \
    -d "{\"camera_id\":\"cam-$i\",\"user_id\":\"user123\",\"source\":\"DUBAI_POLICE\",\"duration\":3600}" \
    | jq '.reservation_id'
done

# Try 51st (should get 429)
curl -i -X POST http://localhost:8000/api/v1/stream/reserve \
  -H "Content-Type: application/json" \
  -d '{"camera_id":"cam-51","user_id":"user123","source":"DUBAI_POLICE","duration":3600}'
```

Expected response:
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 50
X-RateLimit-Remaining: 0
Retry-After: 30

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message_en": "Camera limit reached for Dubai Police",
    "message_ar": "تم الوصول إلى حد الكاميرات لشرطة دبي"
  }
}
```

## Metrics

Prometheus metrics available at `http://localhost:8001/metrics`:

```bash
# Scrape metrics
curl http://localhost:8001/metrics

# Filter Kong metrics
curl http://localhost:8001/metrics | grep kong_
```

Key metrics:
- `kong_http_requests_total` - Total requests
- `kong_latency_ms` - Request latency (proxy, upstream, total)
- `kong_bandwidth_bytes` - Bandwidth usage
- `kong_datastore_reachable` - Database health (not used in DB-less mode)

## Production Deployment

### Enable PostgreSQL Database Mode

For production, use database mode instead of declarative config:

1. **Update docker-compose.yml**:
   ```yaml
   kong-database:
     image: postgres:15-alpine
     environment:
       POSTGRES_USER: kong
       POSTGRES_PASSWORD: ${KONG_DB_PASSWORD}
       POSTGRES_DB: kong

   kong:
     environment:
       KONG_DATABASE: postgres
       KONG_PG_HOST: kong-database
       KONG_PG_USER: kong
       KONG_PG_PASSWORD: ${KONG_DB_PASSWORD}
     depends_on:
       - kong-database
   ```

2. **Run migrations**:
   ```bash
   docker-compose run --rm kong kong migrations bootstrap
   docker-compose run --rm kong kong migrations up
   ```

3. **Apply declarative config**:
   ```bash
   # Install deck CLI
   curl -sL https://github.com/kong/deck/releases/download/v1.28.0/deck_1.28.0_linux_amd64.tar.gz | tar -xz

   # Sync configuration
   ./deck sync --config config/kong.yml
   ```

### Enable Authentication

Add Key Auth plugin:

```yaml
plugins:
  - name: key-auth
    enabled: true
    config:
      key_names:
        - apikey
      key_in_header: true
      key_in_query: true
      hide_credentials: false

consumers:
  - username: dubai-police-operator
    keyauth_credentials:
      - key: ${DP_API_KEY}
```

Test with API key:
```bash
curl -H "apikey: ${DP_API_KEY}" http://kong:8000/api/v1/vms/cameras
```

### Enable SSL/TLS

1. **Generate certificates**:
   ```bash
   openssl req -new -x509 -nodes -newkey rsa:4096 \
     -keyout rta.key -out rta.crt -days 365 \
     -subj "/C=AE/ST=Dubai/L=Dubai/O=RTA/CN=api.rta.ae"
   ```

2. **Update Kong config**:
   ```
   ssl_cert = /etc/kong/ssl/rta.crt
   ssl_cert_key = /etc/kong/ssl/rta.key
   ```

3. **Use HTTPS**:
   ```bash
   curl https://kong:8443/api/v1/vms/cameras
   ```

## Troubleshooting

### Kong Won't Start
```bash
# Check config syntax
docker-compose run --rm kong kong config parse /etc/kong/kong.yml

# Check logs
docker-compose logs kong
```

### Plugin Not Loading
```bash
# Verify plugin path
docker exec cctv-kong ls -la /etc/kong/plugins/quota-validator/

# Check plugin list
curl http://localhost:8001/plugins | jq '.data[] | select(.name=="quota-validator")'
```

### Routes Not Working
```bash
# List all routes
curl http://localhost:8001/routes | jq .

# Test specific route
curl -v http://localhost:8000/api/v1/vms/cameras

# Check upstream health
curl http://localhost:8001/upstreams | jq .
```

## References

- **Kong Documentation**: https://docs.konghq.com/
- **Declarative Config**: https://docs.konghq.com/gateway/latest/production/deployment-topologies/db-less-and-declarative-config/
- **Custom Plugins**: https://docs.konghq.com/gateway/latest/plugin-development/
- **deck CLI**: https://docs.konghq.com/deck/latest/
