package domain

import "time"

// AnnotationType represents the type of annotation
type AnnotationType string

const (
	AnnotationNote     AnnotationType = "NOTE"
	AnnotationMarker   AnnotationType = "MARKER"
	AnnotationWarning  AnnotationType = "WARNING"
	AnnotationEvidence AnnotationType = "EVIDENCE"
)

// Annotation represents a time-based annotation on a video segment
type Annotation struct {
	ID              string         `json:"id"`
	SegmentID       string         `json:"segment_id"`
	TimestampOffset int            `json:"timestamp_offset"` // Seconds from segment start
	Type            AnnotationType `json:"type"`
	Content         string         `json:"content"`
	UserID          string         `json:"user_id"`
	CreatedAt       time.Time      `json:"created_at"`
}

// CreateAnnotationRequest represents a request to create an annotation
type CreateAnnotationRequest struct {
	SegmentID       string         `json:"segment_id" validate:"required,uuid"`
	TimestampOffset int            `json:"timestamp_offset" validate:"min=0"`
	Type            AnnotationType `json:"type" validate:"required,oneof=NOTE MARKER WARNING EVIDENCE"`
	Content         string         `json:"content" validate:"required,min=1"`
	UserID          string         `json:"user_id" validate:"required"`
}
