package domain

import "time"

// Tag represents a metadata tag
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category,omitempty"`
	Color     string    `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTagRequest represents a request to create a tag
type CreateTagRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Category string `json:"category,omitempty"`
	Color    string `json:"color,omitempty"`
}

// VideoTag represents a tag applied to a video segment
type VideoTag struct {
	SegmentID string    `json:"segment_id"`
	TagID     string    `json:"tag_id"`
	TagName   string    `json:"tag_name"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TagSegmentRequest represents a request to tag a segment
type TagSegmentRequest struct {
	TagID  string `json:"tag_id" validate:"required,uuid"`
	UserID string `json:"user_id" validate:"required"`
}
