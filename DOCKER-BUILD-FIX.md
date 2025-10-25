# Docker Build Fix - Missing go.sum Files

## Issue

When running `docker-compose up -d`, you may encounter this error:

```
ERROR [service builder] COPY go.mod go.sum ./
failed to solve: failed to compute cache key: "/go.sum": not found
```

## Root Cause

The Dockerfiles expected both `go.mod` and `go.sum` files, but only `go.mod` files exist in the service directories.

## Solution Applied

Updated all 7 service Dockerfiles to handle missing `go.sum` files:

### Before (Fails if go.sum missing)
```dockerfile
COPY go.mod go.sum ./
RUN go mod download
```

### After (Works with or without go.sum)
```dockerfile
COPY go.mod ./
COPY go.sum* ./

# Copy source code (needed for go mod tidy to scan imports)
COPY . .

# Download dependencies and tidy
RUN go mod download && go mod tidy
```

## Explanation

- `COPY go.sum* ./` - The `*` makes it optional (won't fail if missing)
- `COPY . .` - Copy source code **before** running go mod tidy (tidy needs to scan imports)
- `go mod download` - Downloads dependencies from go.mod
- `go mod tidy` - Scans source code imports, generates go.sum with checksums

**Important**: `go mod tidy` must run **after** copying source code because it needs to scan all `.go` files to detect required dependencies.

## Files Updated

1. `services/vms-service/Dockerfile`
2. `services/storage-service/Dockerfile`
3. `services/recording-service/Dockerfile`
4. `services/metadata-service/Dockerfile`
5. `services/playback-service/Dockerfile`
6. `services/stream-counter/Dockerfile`
7. `services/go-api/Dockerfile`

## Try Again

Now run:

```bash
# Clear any cached layers
docker-compose build --no-cache

# Or just start everything
docker-compose up -d
```

The build should now succeed and generate `go.sum` files automatically during the build process.

## Why This Works

The `go mod tidy` command:
1. Verifies all dependencies in go.mod
2. Adds missing modules
3. Removes unused modules
4. **Generates go.sum with checksums**

This happens **inside the Docker build**, so you don't need go.sum files in your source code.

## Alternative: Pre-generate go.sum (Optional)

If you have Go installed locally, you can pre-generate go.sum files:

```bash
# For each service
cd services/vms-service
go mod download
go mod tidy
cd ../..

# Repeat for all services
```

But this is **not required** with the Dockerfile fix applied.
