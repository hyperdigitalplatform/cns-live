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
	r.Use(middleware.SetHeader("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS"))
	r.Use(middleware.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization"))

	// Health check
	r.Get("/health", handler.Health)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1/recording", func(r chi.Router) {
		r.Post("/start/{camera_id}", handler.StartRecording)
		r.Post("/stop/{camera_id}", handler.StopRecording)
		r.Get("/status/{camera_id}", handler.GetRecording)
		r.Get("/status", handler.ListRecordings)
	})

	return r
}
