# Metadata Service

The Metadata Service manages video segment metadata including tags, annotations, and incident tracking for the RTA CCTV system.

## Features

- **Tag Management**: Create and apply tags to video segments for categorization
- **Annotations**: Time-based annotations on video segments (notes, markers, warnings, evidence)
- **Incident Tracking**: Comprehensive incident management with full-text search
- **Full-Text Search**: PostgreSQL-powered search across incident titles and descriptions
- **Multi-Camera Support**: Link incidents to multiple cameras

## Architecture

- **Clean Architecture**: Domain → Repository → Delivery layers
- **PostgreSQL**: Metadata storage with full-text search (tsvector)
- **RESTful API**: HTTP/JSON endpoints
- **Structured Logging**: Zerolog for JSON logs

## API Endpoints

### Tags

```bash
# Create a new tag
POST /api/v1/metadata/tags
{
  "name": "suspicious-activity",
  "category": "security",
  "color": "#ff0000"
}

# List all tags
GET /api/v1/metadata/tags

# Tag a video segment
POST /api/v1/metadata/segments/{segment_id}/tags
{
  "tag_id": "uuid",
  "user_id": "user123"
}

# Get segment tags
GET /api/v1/metadata/segments/{segment_id}/tags
```

### Annotations

```bash
# Create annotation
POST /api/v1/metadata/annotations
{
  "segment_id": "uuid",
  "timestamp_offset": 120,  // seconds from segment start
  "type": "WARNING",        // NOTE, MARKER, WARNING, EVIDENCE
  "content": "Vehicle detected running red light",
  "user_id": "user123"
}

# Get segment annotations
GET /api/v1/metadata/segments/{segment_id}/annotations
```

### Incidents

```bash
# Create incident
POST /api/v1/metadata/incidents
{
  "title": "Traffic Violation - Red Light",
  "description": "Vehicle ran red light at intersection",
  "severity": "MEDIUM",     // LOW, MEDIUM, HIGH, CRITICAL
  "camera_ids": ["cam-001", "cam-002"],
  "start_time": "2024-01-20T10:30:00Z",
  "end_time": "2024-01-20T10:31:00Z",
  "tags": ["traffic", "violation"],
  "created_by": "user123"
}

# Get incident
GET /api/v1/metadata/incidents/{incident_id}

# Update incident (status, assignment)
PATCH /api/v1/metadata/incidents/{incident_id}
{
  "status": "IN_PROGRESS",  // OPEN, IN_PROGRESS, RESOLVED, CLOSED
  "assigned_to": "officer456"
}

# Search incidents (full-text search)
POST /api/v1/metadata/search
{
  "query": "red light violation",
  "camera_ids": ["cam-001"],
  "severity": "MEDIUM",
  "status": "OPEN",
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-31T23:59:59Z",
  "limit": 50,
  "offset": 0
}
```

### Health & Metrics

```bash
GET /health
GET /metrics  # Prometheus metrics
```

## Database Schema

### Tables

- **tags**: Tag definitions (id, name, category, color)
- **video_tags**: Segment-to-tag relationships (many-to-many)
- **annotations**: Time-based segment annotations
- **incidents**: Incident records with full-text search

### Indexes

- GIN indexes on array fields (camera_ids, tags)
- GIN index on tsvector for full-text search
- B-tree indexes on time ranges and status fields

## Configuration

Environment variables:

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/cctv

# Service
PORT=8084
LOG_LEVEL=info
LOG_FORMAT=json
```

## Full-Text Search

The service uses PostgreSQL's full-text search capabilities:

- Automatic tsvector generation via triggers
- Weighted search (title: 'A', description: 'B')
- Support for English language stemming
- GIN index for fast searches

Example search:
```sql
SELECT * FROM incidents
WHERE search_vector @@ plainto_tsquery('english', 'traffic violation')
ORDER BY created_at DESC;
```

## Development

```bash
# Build
go build -o metadata-service ./cmd/main.go

# Run
./metadata-service

# Docker build
docker build -t cctv-metadata-service .

# Docker run
docker run -p 8084:8084 \
  -e DATABASE_URL=postgresql://... \
  cctv-metadata-service
```

## Resource Usage

- **CPU**: 0.25-1 core
- **Memory**: 128-512 MB
- **Disk**: Minimal (metadata only, no video files)

## Integration

The Metadata Service integrates with:

- **Storage Service**: Links metadata to video segments
- **VMS Service**: References camera IDs
- **Playback Service**: Provides metadata for playback UI

## API Response Examples

### Tag Response
```json
{
  "id": "uuid",
  "name": "suspicious-activity",
  "category": "security",
  "color": "#ff0000",
  "created_at": "2024-01-20T10:00:00Z"
}
```

### Annotation Response
```json
{
  "id": "uuid",
  "segment_id": "segment-uuid",
  "timestamp_offset": 120,
  "type": "WARNING",
  "content": "Vehicle detected running red light",
  "user_id": "user123",
  "created_at": "2024-01-20T10:00:00Z"
}
```

### Incident Response
```json
{
  "id": "uuid",
  "title": "Traffic Violation - Red Light",
  "description": "Vehicle ran red light at intersection",
  "severity": "MEDIUM",
  "status": "OPEN",
  "camera_ids": ["cam-001", "cam-002"],
  "start_time": "2024-01-20T10:30:00Z",
  "end_time": "2024-01-20T10:31:00Z",
  "tags": ["traffic", "violation"],
  "assigned_to": null,
  "created_by": "user123",
  "created_at": "2024-01-20T10:00:00Z",
  "updated_at": "2024-01-20T10:00:00Z",
  "closed_at": null
}
```

### Search Response
```json
{
  "incidents": [...],
  "count": 15
}
```
