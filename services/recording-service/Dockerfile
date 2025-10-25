# Recording Service Dockerfile
# Multi-stage build with FFmpeg

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Copy source code (needed for go mod tidy to scan imports)
COPY . .

# Download dependencies and tidy
RUN go mod download && go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o recording-service ./cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies (FFmpeg is critical)
RUN apk --no-cache add \
    ca-certificates \
    wget \
    ffmpeg

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# Set working directory
WORKDIR /app

# Create recordings directory
RUN mkdir -p /tmp/recordings && chown -R app:app /tmp/recordings

# Copy binary from builder
COPY --from=builder /build/recording-service .

# Change ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=10s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# Run service
CMD ["./recording-service"]
