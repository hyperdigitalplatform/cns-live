# Multi-stage build for Stream Counter Service
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Copy source code (including embedded Lua scripts, needed for go mod tidy)
COPY . .

# Download dependencies and tidy
RUN go mod download && go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o stream-counter \
    ./cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 stream && \
    adduser -D -u 1000 -G stream stream

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/stream-counter .

# Change ownership
RUN chown -R stream:stream /app

# Switch to non-root user
USER stream

# Expose port
EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8087/health || exit 1

# Run the binary
CMD ["./stream-counter"]
