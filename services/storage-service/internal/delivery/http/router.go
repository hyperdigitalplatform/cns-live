package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter creates a new HTTP router
func NewRouter(handler *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))
	r.Use(middleware.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS"))
	r.Use(middleware.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization"))

	// Health check
	r.Get("/health", handler.Health)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1/storage", func(r chi.Router) {
		// Segments
		r.Post("/segments", handler.StoreSegment)
		r.Get("/segments/{camera_id}", handler.ListSegments)

		// Exports
		r.Post("/exports", handler.CreateExport)
		r.Get("/exports/{export_id}", handler.GetExport)
		r.Get("/exports/{export_id}/download", handler.DownloadExport)
	})

	return r
}
