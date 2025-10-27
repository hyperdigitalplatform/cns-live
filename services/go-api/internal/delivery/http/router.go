package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	wsDelivery "github.com/rta/cctv/go-api/internal/delivery/websocket"
)

// Router creates the HTTP router
func NewRouter(
	streamHandler *StreamHandler,
	cameraHandler *CameraHandler,
	wsHandler *wsDelivery.Handler,
	layoutHandler *LayoutHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware - handle preflight OPTIONS requests
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health and metrics
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	r.Handle("/metrics", promhttp.Handler())

	// WebSocket endpoint
	r.Get("/ws/stream/stats", wsHandler.ServeWS)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Stream management
		r.Route("/stream", func(r chi.Router) {
			r.Post("/reserve", streamHandler.RequestStream)
			r.Delete("/release/{id}", streamHandler.ReleaseStream)
			r.Post("/heartbeat/{id}", streamHandler.SendHeartbeat)
			r.Get("/stats", streamHandler.GetStreamStats)
		})

		// Camera management
		r.Route("/cameras", func(r chi.Router) {
			r.Get("/", cameraHandler.ListCameras)
			r.Get("/{id}", cameraHandler.GetCamera)
			r.Post("/{id}/ptz", cameraHandler.ControlPTZ)
		})

		// Layout management
		r.Route("/layouts", func(r chi.Router) {
			r.Post("/", layoutHandler.CreateLayout)
			r.Get("/", layoutHandler.ListLayouts)
			r.Get("/{id}", layoutHandler.GetLayout)
			r.Put("/{id}", layoutHandler.UpdateLayout)
			r.Delete("/{id}", layoutHandler.DeleteLayout)
		})
	})

	return r
}
