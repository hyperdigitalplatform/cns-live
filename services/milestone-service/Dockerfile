# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy source code first
COPY . .

# Download dependencies and build
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o milestone-service ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/milestone-service .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./milestone-service"]
