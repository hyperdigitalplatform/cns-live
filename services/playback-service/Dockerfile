# Build stage
FROM golang:1.21-alpine AS builder

# Install git (required for go mod download)
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Copy source code (needed for go mod tidy to scan imports)
COPY . .

# Download dependencies and tidy
RUN go mod download && go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o playback-service ./cmd/main.go

# Final stage
FROM alpine:latest

# Install FFmpeg for HLS transmuxing
RUN apk --no-cache add ca-certificates wget ffmpeg

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/playback-service .

# Create directories for work, cache, HLS, and exports
RUN mkdir -p /tmp/playback /tmp/playback/cache /tmp/playback/hls /tmp/playback/exports /tmp/playback/temp

# Create non-root user
RUN addgroup -S app && adduser -S app -G app
RUN chown -R app:app /tmp/playback

USER app

EXPOSE 8090

CMD ["./playback-service"]
