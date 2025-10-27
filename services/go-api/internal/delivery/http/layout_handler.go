package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rs/zerolog"
)

// LayoutHandler handles layout-related HTTP requests
type LayoutHandler struct {
	layoutUseCase domain.LayoutUseCase
	logger        zerolog.Logger
}

// NewLayoutHandler creates a new layout handler
func NewLayoutHandler(layoutUseCase domain.LayoutUseCase, logger zerolog.Logger) *LayoutHandler {
	return &LayoutHandler{
		layoutUseCase: layoutUseCase,
		logger:        logger,
	}
}

// CreateLayout handles layout creation request
// POST /api/v1/layouts
func (h *LayoutHandler) CreateLayout(w http.ResponseWriter, r *http.Request) {
	var request domain.CreateLayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode create layout request")
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Info().Str("name", request.Name).Str("layout_type", string(request.LayoutType)).Str("grid_layout", request.GridLayout).Msg("Received create layout request")

	layout, err := h.layoutUseCase.CreateLayout(&request)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create layout")
		respondError(w, http.StatusInternalServerError, "Failed to create layout")
		return
	}

	respondJSON(w, http.StatusCreated, layout)
}

// GetLayout handles layout retrieval request
// GET /api/v1/layouts/:id
func (h *LayoutHandler) GetLayout(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Layout ID is required")
		return
	}

	layout, err := h.layoutUseCase.GetLayout(id)
	if err != nil {
		h.logger.Error().Err(err).Str("layout_id", id).Msg("Failed to get layout")
		respondError(w, http.StatusNotFound, "Layout not found")
		return
	}

	respondJSON(w, http.StatusOK, layout)
}

// ListLayouts handles layout list request
// GET /api/v1/layouts
func (h *LayoutHandler) ListLayouts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	var layoutType *domain.LayoutType
	var scope *domain.LayoutScope
	var createdBy *string

	if lt := r.URL.Query().Get("layout_type"); lt != "" {
		t := domain.LayoutType(lt)
		layoutType = &t
	}

	if s := r.URL.Query().Get("scope"); s != "" {
		sc := domain.LayoutScope(s)
		scope = &sc
	}

	if cb := r.URL.Query().Get("created_by"); cb != "" {
		createdBy = &cb
	}

	response, err := h.layoutUseCase.ListLayouts(layoutType, scope, createdBy)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list layouts")
		respondError(w, http.StatusInternalServerError, "Failed to list layouts")
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// UpdateLayout handles layout update request
// PUT /api/v1/layouts/:id
func (h *LayoutHandler) UpdateLayout(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Layout ID is required")
		return
	}

	var request domain.UpdateLayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode update layout request")
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	layout, err := h.layoutUseCase.UpdateLayout(id, &request)
	if err != nil {
		h.logger.Error().Err(err).Str("layout_id", id).Msg("Failed to update layout")
		respondError(w, http.StatusInternalServerError, "Failed to update layout")
		return
	}

	respondJSON(w, http.StatusOK, layout)
}

// DeleteLayout handles layout deletion request
// DELETE /api/v1/layouts/:id
func (h *LayoutHandler) DeleteLayout(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Layout ID is required")
		return
	}

	if err := h.layoutUseCase.DeleteLayout(id); err != nil {
		// Check if it's a "not found" error
		if err.Error() == "layout not found: "+id ||
		   err.Error() == "failed to delete layout: layout not found: "+id {
			h.logger.Warn().Str("layout_id", id).Msg("Layout not found for deletion")
			respondError(w, http.StatusNotFound, "Layout not found")
			return
		}
		h.logger.Error().Err(err).Str("layout_id", id).Msg("Failed to delete layout")
		respondError(w, http.StatusInternalServerError, "Failed to delete layout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions for consistent response format

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
