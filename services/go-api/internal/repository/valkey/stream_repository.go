package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rta/cctv/go-api/internal/domain"
	"github.com/rs/zerolog"
)

// StreamRepository implements stream repository using Valkey
type StreamRepository struct {
	client *redis.Client
	logger zerolog.Logger
}

// NewStreamRepository creates a new Valkey stream repository
func NewStreamRepository(client *redis.Client, logger zerolog.Logger) *StreamRepository {
	return &StreamRepository{
		client: client,
		logger: logger,
	}
}

// SaveReservation saves a stream reservation
func (r *StreamRepository) SaveReservation(ctx context.Context, reservation *domain.StreamReservation) error {
	key := fmt.Sprintf("stream:reservation:%s", reservation.ID)

	data, err := json.Marshal(reservation)
	if err != nil {
		return fmt.Errorf("failed to marshal reservation: %w", err)
	}

	// Set with TTL (1 hour)
	err = r.client.Set(ctx, key, data, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save reservation: %w", err)
	}

	// Add to active set
	err = r.client.SAdd(ctx, "stream:active", reservation.ID).Err()
	if err != nil {
		r.logger.Warn().Err(err).Msg("Failed to add to active set")
	}

	// Add to user's reservations set
	userKey := fmt.Sprintf("stream:user:%s", reservation.UserID)
	err = r.client.SAdd(ctx, userKey, reservation.ID).Err()
	if err != nil {
		r.logger.Warn().Err(err).Msg("Failed to add to user set")
	}

	return nil
}

// GetReservation retrieves a stream reservation
func (r *StreamRepository) GetReservation(ctx context.Context, reservationID string) (*domain.StreamReservation, error) {
	key := fmt.Sprintf("stream:reservation:%s", reservationID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("reservation not found: %s", reservationID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	var reservation domain.StreamReservation
	if err := json.Unmarshal(data, &reservation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reservation: %w", err)
	}

	return &reservation, nil
}

// GetActiveReservations retrieves all active reservations
func (r *StreamRepository) GetActiveReservations(ctx context.Context) ([]*domain.StreamReservation, error) {
	// Get all reservation IDs from active set
	reservationIDs, err := r.client.SMembers(ctx, "stream:active").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active reservations: %w", err)
	}

	reservations := make([]*domain.StreamReservation, 0, len(reservationIDs))
	for _, id := range reservationIDs {
		reservation, err := r.GetReservation(ctx, id)
		if err != nil {
			// Reservation might have expired, remove from active set
			r.client.SRem(ctx, "stream:active", id)
			continue
		}
		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// GetUserReservations retrieves all reservations for a user
func (r *StreamRepository) GetUserReservations(ctx context.Context, userID string) ([]*domain.StreamReservation, error) {
	userKey := fmt.Sprintf("stream:user:%s", userID)

	reservationIDs, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user reservations: %w", err)
	}

	reservations := make([]*domain.StreamReservation, 0, len(reservationIDs))
	for _, id := range reservationIDs {
		reservation, err := r.GetReservation(ctx, id)
		if err != nil {
			// Reservation expired, remove from user set
			r.client.SRem(ctx, userKey, id)
			continue
		}
		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// UpdateHeartbeat updates the last heartbeat time
func (r *StreamRepository) UpdateHeartbeat(ctx context.Context, reservationID string) error {
	reservation, err := r.GetReservation(ctx, reservationID)
	if err != nil {
		return err
	}

	reservation.LastHeartbeat = time.Now()

	// Save updated reservation
	return r.SaveReservation(ctx, reservation)
}

// DeleteReservation deletes a stream reservation
func (r *StreamRepository) DeleteReservation(ctx context.Context, reservationID string) error {
	// Get reservation to find user ID
	reservation, err := r.GetReservation(ctx, reservationID)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("stream:reservation:%s", reservationID)

	// Delete reservation
	err = r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	// Remove from active set
	r.client.SRem(ctx, "stream:active", reservationID)

	// Remove from user set
	userKey := fmt.Sprintf("stream:user:%s", reservation.UserID)
	r.client.SRem(ctx, userKey, reservationID)

	return nil
}
