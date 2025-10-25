package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(handler *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health and metrics
	r.Get("/health", handler.Health)
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1/metadata", func(r chi.Router) {
		// Tags
		r.Post("/tags", handler.CreateTag)
		r.Get("/tags", handler.GetTags)
		r.Post("/segments/{id}/tags", handler.TagSegment)
		r.Get("/segments/{id}/tags", handler.GetSegmentTags)

		// Annotations
		r.Post("/annotations", handler.CreateAnnotation)
		r.Get("/segments/{id}/annotations", handler.GetSegmentAnnotations)

		// Incidents
		r.Post("/incidents", handler.CreateIncident)
		r.Get("/incidents/{id}", handler.GetIncident)
		r.Patch("/incidents/{id}", handler.UpdateIncident)
		r.Post("/search", handler.SearchIncidents)
	})

	return r
}
