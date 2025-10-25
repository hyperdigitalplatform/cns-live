package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rta/cctv/go-api/internal/usecase"
	"github.com/rs/zerolog"
)

// StreamHandler handles stream-related HTTP requests
type StreamHandler struct {
	streamUseCase *usecase.StreamUseCase
	logger        zerolog.Logger
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(streamUseCase *usecase.StreamUseCase, logger zerolog.Logger) *StreamHandler {
	return &StreamHandler{
		streamUseCase: streamUseCase,
		logger:        logger,
	}
}

// RequestStream handles stream request
// POST /api/v1/stream/reserve
func (h *StreamHandler) RequestStream(w http.ResponseWriter, r *http.Request) {
	var req domain.StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.CameraID == "" || req.UserID == "" {
		h.respondError(w, http.StatusBadRequest, "camera_id and user_id are required")
		return
	}

	// Request stream
	response, err := h.streamUseCase.RequestStream(r.Context(), req)
	if err != nil {
		// Check if it's an agency limit error
		if limitErr, ok := err.(*domain.AgencyLimitError); ok {
			h.respondJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"error": map[string]interface{}{
					"code":        "AGENCY_LIMIT_EXCEEDED",
					"message_en":  limitErr.Message,
					"message_ar":  h.translateAgencyLimitError(limitErr),
					"source":      limitErr.Source,
					"current":     limitErr.Current,
					"limit":       limitErr.Limit,
				},
			})
			return
		}

		h.logger.Error().Err(err).Msg("Failed to request stream")
		h.respondError(w, http.StatusInternalServerError, "Failed to request stream")
		return
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// ReleaseStream handles stream release
// DELETE /api/v1/stream/release/{id}
func (h *StreamHandler) ReleaseStream(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "id")

	if reservationID == "" {
		h.respondError(w, http.StatusBadRequest, "reservation_id is required")
		return
	}

	err := h.streamUseCase.ReleaseStream(r.Context(), reservationID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to release stream")
		h.respondError(w, http.StatusInternalServerError, "Failed to release stream")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "released",
		"message": "Stream reservation released successfully",
	})
}

// SendHeartbeat handles heartbeat request
// POST /api/v1/stream/heartbeat/{id}
func (h *StreamHandler) SendHeartbeat(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "id")

	if reservationID == "" {
		h.respondError(w, http.StatusBadRequest, "reservation_id is required")
		return
	}

	err := h.streamUseCase.SendHeartbeat(r.Context(), reservationID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to send heartbeat")
		h.respondError(w, http.StatusInternalServerError, "Failed to send heartbeat")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// GetStreamStats handles stream statistics request
// GET /api/v1/stream/stats
func (h *StreamHandler) GetStreamStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.streamUseCase.GetStreamStats(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get stream stats")
		h.respondError(w, http.StatusInternalServerError, "Failed to get stream stats")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// Helper methods

func (h *StreamHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *StreamHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *StreamHandler) translateAgencyLimitError(err *domain.AgencyLimitError) string {
	// Arabic translation
	return fmt.Sprintf("تم الوصول إلى حد الوكالة لـ %s (%d/%d)", err.Source, err.Current, err.Limit)
}
