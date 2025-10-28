package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter creates a new HTTP router with all routes
func NewRouter(handler *Handler, milestoneHandler *MilestoneHandler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*.rta.ae", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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
	r.Route("/vms", func(r chi.Router) {
		// Camera routes
		r.Route("/cameras", func(r chi.Router) {
			r.Get("/", handler.GetCameras)              // GET /vms/cameras (with optional ?source= filter)
			r.Get("/{id}", handler.GetCameraByID)       // GET /vms/cameras/{id}
			r.Get("/{id}/stream", handler.GetCameraStream) // GET /vms/cameras/{id}/stream
			r.Post("/{id}/ptz", handler.ExecutePTZ)     // POST /vms/cameras/{id}/ptz
		})

		// Recording routes
		r.Route("/recordings", func(r chi.Router) {
			r.Get("/{camera_id}/segments", handler.GetRecordingSegments) // GET /vms/recordings/{camera_id}/segments
			r.Post("/export", handler.ExportRecording)                   // POST /vms/recordings/export
			r.Get("/export/{export_id}", handler.GetExportStatus)        // GET /vms/recordings/export/{export_id}
		})

	})

	return r
}
