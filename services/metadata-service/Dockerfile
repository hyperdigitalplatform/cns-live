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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o metadata-service ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/metadata-service .

# Create non-root user
RUN addgroup -S app && adduser -S app -G app
USER app

EXPOSE 8084

CMD ["./metadata-service"]
