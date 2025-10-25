package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rta/cctv/stream-counter/internal/domain"
	"github.com/rta/cctv/stream-counter/pkg/valkey"
	"github.com/rs/zerolog"
)

// Handler handles HTTP requests for Stream Counter Service
type Handler struct {
	valkey *valkey.Client
	logger zerolog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(valkeyClient *valkey.Client, logger zerolog.Logger) *Handler {
	return &Handler{
		valkey: valkeyClient,
		logger: logger,
	}
}

// ReserveStream reserves a stream slot for a camera
// POST /api/v1/stream/reserve
func (h *Handler) ReserveStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode reserve request")
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if req.CameraID == "" {
		respondError(w, http.StatusBadRequest, "Camera ID is required", "MISSING_CAMERA_ID")
		return
	}

	if req.UserID == "" {
		respondError(w, http.StatusBadRequest, "User ID is required", "MISSING_USER_ID")
		return
	}

	if !req.Source.IsValid() {
		respondError(w, http.StatusBadRequest, "Invalid source. Must be one of: DUBAI_POLICE, METRO, BUS, OTHER", "INVALID_SOURCE")
		return
	}

	if req.Duration < 60 || req.Duration > 7200 {
		respondError(w, http.StatusBadRequest, "Duration must be between 60 and 7200 seconds", "INVALID_DURATION")
		return
	}

	// Generate reservation ID
	reservationID := uuid.New().String()

	// Attempt to reserve stream
	success, current, limit, err := h.valkey.ReserveStream(
		ctx,
		string(req.Source),
		reservationID,
		req.CameraID,
		req.UserID,
		req.Duration,
	)

	if err != nil {
		h.logger.Error().Err(err).Str("source", string(req.Source)).Msg("Failed to reserve stream")
		respondError(w, http.StatusInternalServerError, "Failed to reserve stream", "INTERNAL_ERROR")
		return
	}

	if !success {
		// Limit reached - return 429 with bilingual message
		h.logger.Warn().
			Str("source", string(req.Source)).
			Int("current", current).
			Int("limit", limit).
			Msg("Stream limit reached")

		respondLimitExceeded(w, req.Source, current, limit)
		return
	}

	// Success - return reservation details
	expiresAt := time.Now().Add(time.Duration(req.Duration) * time.Second)

	response := domain.ReserveResponse{
		ReservationID: reservationID,
		CameraID:      req.CameraID,
		UserID:        req.UserID,
		Source:        req.Source,
		ExpiresAt:     expiresAt,
		CurrentUsage:  current,
		Limit:         limit,
	}

	h.logger.Info().
		Str("reservation_id", reservationID).
		Str("camera_id", req.CameraID).
		Str("source", string(req.Source)).
		Int("current", current).
		Int("limit", limit).
		Msg("Stream reserved successfully")

	respondJSON(w, http.StatusOK, response)
}

// ReleaseStream releases a stream reservation
// DELETE /api/v1/stream/release/{reservation_id}
func (h *Handler) ReleaseStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reservationID := chi.URLParam(r, "reservation_id")

	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "Reservation ID is required", "MISSING_RESERVATION_ID")
		return
	}

	// Attempt to release stream
	success, newCount, source, err := h.valkey.ReleaseStream(ctx, reservationID)

	if err != nil {
		h.logger.Error().Err(err).Str("reservation_id", reservationID).Msg("Failed to release stream")
		respondError(w, http.StatusInternalServerError, "Failed to release stream", "INTERNAL_ERROR")
		return
	}

	if !success {
		h.logger.Warn().Str("reservation_id", reservationID).Msg("Reservation not found")
		respondError(w, http.StatusNotFound, "Reservation not found", "RESERVATION_NOT_FOUND")
		return
	}

	// Success
	response := domain.ReleaseResponse{
		ReservationID: reservationID,
		Source:        domain.CameraSource(source),
		Released:      true,
		NewCount:      newCount,
	}

	h.logger.Info().
		Str("reservation_id", reservationID).
		Str("source", source).
		Int("new_count", newCount).
		Msg("Stream released successfully")

	respondJSON(w, http.StatusOK, response)
}

// HeartbeatStream sends heartbeat to keep reservation alive
// POST /api/v1/stream/heartbeat/{reservation_id}
func (h *Handler) HeartbeatStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reservationID := chi.URLParam(r, "reservation_id")

	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "Reservation ID is required", "MISSING_RESERVATION_ID")
		return
	}

	// Parse request body (optional)
	var req domain.HeartbeatRequest
	req.ReservationID = reservationID
	req.ExtendTTL = 60 // Default 60 seconds

	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// Validate TTL extension
	if req.ExtendTTL < 30 || req.ExtendTTL > 300 {
		req.ExtendTTL = 60 // Reset to default if invalid
	}

	// Send heartbeat
	success, remainingTTL, err := h.valkey.HeartbeatStream(ctx, reservationID, req.ExtendTTL)

	if err != nil {
		h.logger.Error().Err(err).Str("reservation_id", reservationID).Msg("Failed to send heartbeat")
		respondError(w, http.StatusInternalServerError, "Failed to send heartbeat", "INTERNAL_ERROR")
		return
	}

	if !success {
		h.logger.Warn().Str("reservation_id", reservationID).Msg("Reservation not found for heartbeat")
		respondError(w, http.StatusNotFound, "Reservation not found", "RESERVATION_NOT_FOUND")
		return
	}

	// Success
	response := domain.HeartbeatResponse{
		ReservationID: reservationID,
		RemainingTTL:  remainingTTL,
		Updated:       true,
	}

	h.logger.Debug().
		Str("reservation_id", reservationID).
		Int("remaining_ttl", remainingTTL).
		Msg("Heartbeat updated")

	respondJSON(w, http.StatusOK, response)
}

// GetStats retrieves current stream statistics
// GET /api/v1/stream/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get stats for all sources
	sources := []string{
		string(domain.SourceDubaiPolice),
		string(domain.SourceMetro),
		string(domain.SourceBus),
		string(domain.SourceOther),
	}

	statsData, err := h.valkey.GetStats(ctx, sources)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get stats")
		respondError(w, http.StatusInternalServerError, "Failed to retrieve statistics", "INTERNAL_ERROR")
		return
	}

	// Parse stats
	stats := make([]domain.StreamStats, 0, len(statsData))
	totalCurrent := 0
	totalLimit := 0

	for _, item := range statsData {
		if len(item) < 4 {
			continue
		}

		source, _ := item[0].(string)
		current, _ := item[1].(int64)
		limit, _ := item[2].(int64)
		percentage, _ := item[3].(int64)

		available := int(limit) - int(current)
		if available < 0 {
			available = 0
		}

		stats = append(stats, domain.StreamStats{
			Source:     domain.CameraSource(source),
			Current:    int(current),
			Limit:      int(limit),
			Percentage: int(percentage),
			Available:  available,
		})

		totalCurrent += int(current)
		totalLimit += int(limit)
	}

	// Calculate total percentage
	totalPercentage := 0
	if totalLimit > 0 {
		totalPercentage = (totalCurrent * 100) / totalLimit
	}

	response := domain.StatsResponse{
		Stats: stats,
		Total: domain.StatsInfo{
			Current:    totalCurrent,
			Limit:      totalLimit,
			Percentage: totalPercentage,
			Available:  totalLimit - totalCurrent,
		},
		Timestamp: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// Health check endpoint
// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check Valkey connection
	if err := h.valkey.Ping(ctx); err != nil {
		respondError(w, http.StatusServiceUnavailable, "Valkey connection failed", "VALKEY_UNAVAILABLE")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "stream-counter",
		"timestamp": time.Now(),
	})
}

// Utility functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string, code string) {
	respondJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"code":    code,
		},
	})
}

func respondLimitExceeded(w http.ResponseWriter, source domain.CameraSource, current int, limit int) {
	// Bilingual error messages
	messages := map[domain.CameraSource]map[string]string{
		domain.SourceDubaiPolice: {
			"en": "Camera limit reached for Dubai Police",
			"ar": "تم الوصول إلى حد الكاميرات لشرطة دبي",
		},
		domain.SourceMetro: {
			"en": "Camera limit reached for Metro",
			"ar": "تم الوصول إلى حد الكاميرات للمترو",
		},
		domain.SourceBus: {
			"en": "Camera limit reached for Bus",
			"ar": "تم الوصول إلى حد الكاميرات للحافلات",
		},
		domain.SourceOther: {
			"en": "Camera limit reached for Other",
			"ar": "تم الوصول إلى حد الكاميرات للأخرى",
		},
	}

	msg := messages[source]
	if msg == nil {
		msg = map[string]string{
			"en": "Camera limit reached",
			"ar": "تم الوصول إلى حد الكاميرات",
		}
	}

	respondJSON(w, http.StatusTooManyRequests, map[string]interface{}{
		"error": map[string]interface{}{
			"code":       "RATE_LIMIT_EXCEEDED",
			"message_en": msg["en"],
			"message_ar": msg["ar"],
			"source":     source,
			"current":    current,
			"limit":      limit,
			"retry_after": 30,
		},
	})
}
