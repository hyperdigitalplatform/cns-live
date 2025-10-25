package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter creates a new HTTP router with all routes
func NewRouter(handler *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*.rta.ae", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", handler.Health)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1/stream", func(r chi.Router) {
		r.Post("/reserve", handler.ReserveStream)                        // POST /api/v1/stream/reserve
		r.Delete("/release/{reservation_id}", handler.ReleaseStream)     // DELETE /api/v1/stream/release/{id}
		r.Post("/heartbeat/{reservation_id}", handler.HeartbeatStream)   // POST /api/v1/stream/heartbeat/{id}
		r.Get("/stats", handler.GetStats)                                // GET /api/v1/stream/stats
	})

	return r
}
