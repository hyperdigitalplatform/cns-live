package repository

import (
	"context"

	"github.com/rta/cctv/go-api/internal/domain"
)

// StreamRepository handles stream reservation persistence
type StreamRepository interface {
	SaveReservation(ctx context.Context, reservation *domain.StreamReservation) error
	SaveReservationMetadata(ctx context.Context, reservation *domain.StreamReservation) error
	GetReservation(ctx context.Context, reservationID string) (*domain.StreamReservation, error)
	GetReservationFromHash(ctx context.Context, reservationID string) (*domain.StreamReservation, error)
	GetActiveReservations(ctx context.Context) ([]*domain.StreamReservation, error)
	GetReservationByCameraID(ctx context.Context, cameraID string) (*domain.StreamReservation, error)
	GetUserReservations(ctx context.Context, userID string) ([]*domain.StreamReservation, error)
	UpdateHeartbeat(ctx context.Context, reservationID string) error
	DeleteReservation(ctx context.Context, reservationID string) error
	DeleteReservationMetadata(ctx context.Context, reservationID string) error
}
