package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(playbackHandler *PlaybackHandler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second)) // Longer timeout for video streaming

	// CORS middleware
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))
	r.Use(middleware.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS"))
	r.Use(middleware.SetHeader("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization"))

	// Health and metrics
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Playback management
		r.Route("/playback", func(r chi.Router) {
			r.Post("/request", playbackHandler.RequestPlayback)
			r.Post("/export", playbackHandler.CreateExport)

			// Cache management
			r.Route("/cache", func(r chi.Router) {
				r.Get("/stats", playbackHandler.GetCacheStats)
			})
		})
	})

	return r
}
