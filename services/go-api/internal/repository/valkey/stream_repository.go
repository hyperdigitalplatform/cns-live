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
// NOTE: Disabled - stream-counter service manages reservations via HASH
// Use SaveReservationMetadata to save go-api specific metadata separately
func (r *StreamRepository) SaveReservation(ctx context.Context, reservation *domain.StreamReservation) error {
	// Disabled to prevent overwriting stream-counter's HASH with JSON STRING
	return nil
}

// SaveReservationMetadata saves go-api specific metadata for a reservation
// This stores data in a separate key to avoid conflicting with stream-counter's HASH
func (r *StreamRepository) SaveReservationMetadata(ctx context.Context, reservation *domain.StreamReservation) error {
	// Use a separate key for go-api metadata
	key := fmt.Sprintf("stream:metadata:%s", reservation.ID)

	// Store as HASH to be consistent with stream-counter
	err := r.client.HSet(ctx, key,
		"camera_name", reservation.CameraName,
		"room_name", reservation.RoomName,
		"token", reservation.Token,
		"ingress_id", reservation.IngressID,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	// Set TTL (1 hour)
	r.client.Expire(ctx, key, time.Hour)

	return nil
}

// GetReservation retrieves a stream reservation
// NOTE: This function expects JSON STRING format (legacy)
// For HASH format (stream-counter managed), use GetReservationFromHash
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

// GetReservationFromHash retrieves a stream reservation stored as HASH by stream-counter
// Also fetches go-api metadata from separate key
func (r *StreamRepository) GetReservationFromHash(ctx context.Context, reservationID string) (*domain.StreamReservation, error) {
	key := fmt.Sprintf("stream:reservation:%s", reservationID)

	// Check if key exists
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check reservation existence: %w", err)
	}
	if exists == 0 {
		return nil, fmt.Errorf("reservation not found: %s", reservationID)
	}

	// Get stream-counter HASH fields
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("reservation not found: %s", reservationID)
	}

	// Get go-api metadata from separate key
	metaKey := fmt.Sprintf("stream:metadata:%s", reservationID)
	metadata, err := r.client.HGetAll(ctx, metaKey).Result()
	if err != nil {
		r.logger.Warn().Err(err).Msg("Failed to get metadata")
		metadata = make(map[string]string)
	}

	// Map HASH fields to domain struct (combining both sources)
	reservation := &domain.StreamReservation{
		ID:         reservationID,
		CameraID:   data["camera_id"],
		CameraName: metadata["camera_name"],
		UserID:     data["user_id"],
		Source:     data["source"],
		RoomName:   metadata["room_name"],
		Token:      metadata["token"],
		IngressID:  metadata["ingress_id"],
	}

	return reservation, nil
}

// GetActiveReservations retrieves all active reservations
// Scans all stream:reservation:* keys managed by stream-counter
func (r *StreamRepository) GetActiveReservations(ctx context.Context) ([]*domain.StreamReservation, error) {
	// Scan for all reservation keys (stream-counter managed)
	var cursor uint64
	var reservations []*domain.StreamReservation

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "stream:reservation:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservations: %w", err)
		}

		// Get each reservation using HASH format
		for _, key := range keys {
			// Extract reservation ID from key
			reservationID := key[len("stream:reservation:"):]

			reservation, err := r.GetReservationFromHash(ctx, reservationID)
			if err != nil {
				r.logger.Warn().Err(err).Str("reservation_id", reservationID).Msg("Failed to get reservation")
				continue
			}
			reservations = append(reservations, reservation)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return reservations, nil
}

// GetReservationByCameraID retrieves the first active reservation for a given camera
// Returns nil if no reservation found for this camera
func (r *StreamRepository) GetReservationByCameraID(ctx context.Context, cameraID string) (*domain.StreamReservation, error) {
	// Scan for all reservation keys
	var cursor uint64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "stream:reservation:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservations: %w", err)
		}

		// Check each reservation for matching camera_id
		for _, key := range keys {
			// Extract reservation ID from key
			reservationID := key[len("stream:reservation:"):]

			// Get reservation data from HASH
			data, err := r.client.HGetAll(ctx, key).Result()
			if err != nil {
				r.logger.Warn().Err(err).Str("key", key).Msg("Failed to get reservation HASH")
				continue
			}

			// Check if this reservation is for the requested camera
			if data["camera_id"] == cameraID {
				// Found a reservation for this camera - get full details
				reservation, err := r.GetReservationFromHash(ctx, reservationID)
				if err != nil {
					r.logger.Warn().Err(err).Str("reservation_id", reservationID).Msg("Failed to get full reservation")
					continue
				}
				return reservation, nil
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// No reservation found for this camera
	return nil, nil
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
	// NOTE: Disabled - stream-counter service manages heartbeats via Lua scripts
	// The heartbeat_stream.lua script already handles heartbeat updates properly
	// If we save here, we overwrite the HASH with JSON STRING causing WRONGTYPE errors

	// reservation, err := r.GetReservation(ctx, reservationID)
	// if err != nil {
	// 	return err
	// }

	// reservation.LastHeartbeat = time.Now()

	// // Save updated reservation
	// return r.SaveReservation(ctx, reservation)

	// Just return success - the stream-counter heartbeat is what matters
	return nil
}

// DeleteReservation deletes a stream reservation
// NOTE: Disabled - stream-counter service manages all reservation lifecycle
// The release_stream.lua script handles deletion properly
// Use DeleteReservationMetadata to delete go-api metadata
func (r *StreamRepository) DeleteReservation(ctx context.Context, reservationID string) error {
	// Disabled to prevent conflicts with stream-counter
	return nil
}

// DeleteReservationMetadata deletes go-api specific metadata
func (r *StreamRepository) DeleteReservationMetadata(ctx context.Context, reservationID string) error {
	key := fmt.Sprintf("stream:metadata:%s", reservationID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}
