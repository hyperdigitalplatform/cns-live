package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rta/cctv/metadata-service/internal/domain"
	"github.com/rta/cctv/metadata-service/internal/repository"
	"github.com/rs/zerolog"
)

type Handler struct {
	metadataRepo *repository.MetadataRepository
	logger       zerolog.Logger
}

func NewHandler(metadataRepo *repository.MetadataRepository, logger zerolog.Logger) *Handler {
	return &Handler{
		metadataRepo: metadataRepo,
		logger:       logger,
	}
}

// Tag handlers

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tag := &domain.Tag{
		Name:     req.Name,
		Category: req.Category,
		Color:    req.Color,
	}

	if err := h.metadataRepo.CreateTag(r.Context(), tag); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create tag")
		h.respondError(w, http.StatusInternalServerError, "Failed to create tag")
		return
	}

	h.respondJSON(w, http.StatusCreated, tag)
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.metadataRepo.GetTags(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get tags")
		h.respondError(w, http.StatusInternalServerError, "Failed to get tags")
		return
	}

	h.respondJSON(w, http.StatusOK, tags)
}

func (h *Handler) TagSegment(w http.ResponseWriter, r *http.Request) {
	segmentID := chi.URLParam(r, "id")

	var req domain.TagSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.metadataRepo.TagSegment(r.Context(), segmentID, req.TagID, req.UserID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to tag segment")
		h.respondError(w, http.StatusInternalServerError, "Failed to tag segment")
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]string{"status": "tagged"})
}

func (h *Handler) GetSegmentTags(w http.ResponseWriter, r *http.Request) {
	segmentID := chi.URLParam(r, "id")

	tags, err := h.metadataRepo.GetSegmentTags(r.Context(), segmentID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get segment tags")
		h.respondError(w, http.StatusInternalServerError, "Failed to get segment tags")
		return
	}

	h.respondJSON(w, http.StatusOK, tags)
}

// Annotation handlers

func (h *Handler) CreateAnnotation(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAnnotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	annotation := &domain.Annotation{
		SegmentID:       req.SegmentID,
		TimestampOffset: req.TimestampOffset,
		Type:            req.Type,
		Content:         req.Content,
		UserID:          req.UserID,
	}

	if err := h.metadataRepo.CreateAnnotation(r.Context(), annotation); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create annotation")
		h.respondError(w, http.StatusInternalServerError, "Failed to create annotation")
		return
	}

	h.respondJSON(w, http.StatusCreated, annotation)
}

func (h *Handler) GetSegmentAnnotations(w http.ResponseWriter, r *http.Request) {
	segmentID := chi.URLParam(r, "id")

	annotations, err := h.metadataRepo.GetSegmentAnnotations(r.Context(), segmentID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get annotations")
		h.respondError(w, http.StatusInternalServerError, "Failed to get annotations")
		return
	}

	h.respondJSON(w, http.StatusOK, annotations)
}

// Incident handlers

func (h *Handler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	incident := &domain.Incident{
		Title:       req.Title,
		Description: req.Description,
		Severity:    req.Severity,
		CameraIDs:   req.CameraIDs,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Tags:        req.Tags,
		CreatedBy:   req.CreatedBy,
	}

	if err := h.metadataRepo.CreateIncident(r.Context(), incident); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create incident")
		h.respondError(w, http.StatusInternalServerError, "Failed to create incident")
		return
	}

	h.respondJSON(w, http.StatusCreated, incident)
}

func (h *Handler) GetIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	incident, err := h.metadataRepo.GetIncident(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get incident")
		h.respondError(w, http.StatusNotFound, "Incident not found")
		return
	}

	h.respondJSON(w, http.StatusOK, incident)
}

func (h *Handler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req domain.UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.metadataRepo.UpdateIncident(r.Context(), id, req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to update incident")
		h.respondError(w, http.StatusInternalServerError, "Failed to update incident")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) SearchIncidents(w http.ResponseWriter, r *http.Request) {
	var query domain.SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	incidents, err := h.metadataRepo.SearchIncidents(r.Context(), query)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to search incidents")
		h.respondError(w, http.StatusInternalServerError, "Failed to search incidents")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"incidents": incidents,
		"count":     len(incidents),
	})
}

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// Helper methods

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
