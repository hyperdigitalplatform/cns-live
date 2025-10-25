# Build stage
FROM golang:1.23-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-api ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/go-api .

# Note: Running as root to access Docker socket
# In production, consider using Docker socket proxy or adding user to docker group

EXPOSE 8086

CMD ["./go-api"]
