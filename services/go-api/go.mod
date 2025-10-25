module github.com/rta/cctv/go-api

go 1.23

require (
	github.com/docker/docker v27.4.1+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/go-chi/chi/v5 v5.0.11
	github.com/google/uuid v1.5.0
	github.com/gorilla/websocket v1.5.1
	github.com/lib/pq v1.10.9
	github.com/livekit/server-sdk-go/v2 v2.1.0
	github.com/moby/docker-image-spec v1.3.1
	github.com/prometheus/client_golang v1.18.0
	github.com/redis/go-redis/v9 v9.4.0
	github.com/rs/zerolog v1.31.0
	golang.org/x/time v0.5.0  // Use compatible version for Go 1.23
)
