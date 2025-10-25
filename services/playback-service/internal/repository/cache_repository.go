package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// CacheRepository handles caching for playback sessions and HLS segments
type CacheRepository struct {
	client *redis.Client
	logger zerolog.Logger
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(client *redis.Client, logger zerolog.Logger) *CacheRepository {
	return &CacheRepository{
		client: client,
		logger: logger,
	}
}

// SetSession stores a playback session
func (r *CacheRepository) SetSession(ctx context.Context, sessionID string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := fmt.Sprintf("playback:session:%s", sessionID)
	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache session: %w", err)
	}

	r.logger.Info().Str("session_id", sessionID).Msg("Session cached")
	return nil
}

// GetSession retrieves a playback session
func (r *CacheRepository) GetSession(ctx context.Context, sessionID string, result interface{}) error {
	key := fmt.Sprintf("playback:session:%s", sessionID)
	jsonData, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return nil
}

// SetHLSManifest stores an HLS manifest
func (r *CacheRepository) SetHLSManifest(ctx context.Context, sessionID string, manifest string, ttl time.Duration) error {
	key := fmt.Sprintf("playback:hls:manifest:%s", sessionID)
	err := r.client.Set(ctx, key, manifest, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache HLS manifest: %w", err)
	}

	r.logger.Debug().Str("session_id", sessionID).Msg("HLS manifest cached")
	return nil
}

// GetHLSManifest retrieves an HLS manifest
func (r *CacheRepository) GetHLSManifest(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("playback:hls:manifest:%s", sessionID)
	manifest, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("manifest not found: %s", sessionID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get manifest: %w", err)
	}

	return manifest, nil
}

// SetHLSSegment stores an HLS segment URL
func (r *CacheRepository) SetHLSSegment(ctx context.Context, sessionID string, segmentIndex int, url string, ttl time.Duration) error {
	key := fmt.Sprintf("playback:hls:segment:%s:%d", sessionID, segmentIndex)
	err := r.client.Set(ctx, key, url, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache HLS segment: %w", err)
	}

	return nil
}

// GetHLSSegment retrieves an HLS segment URL
func (r *CacheRepository) GetHLSSegment(ctx context.Context, sessionID string, segmentIndex int) (string, error) {
	key := fmt.Sprintf("playback:hls:segment:%s:%d", sessionID, segmentIndex)
	url, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("segment not found: %s:%d", sessionID, segmentIndex)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get segment: %w", err)
	}

	return url, nil
}

// DeleteSession removes a playback session and related data
func (r *CacheRepository) DeleteSession(ctx context.Context, sessionID string) error {
	// Delete session key
	sessionKey := fmt.Sprintf("playback:session:%s", sessionID)
	manifestKey := fmt.Sprintf("playback:hls:manifest:%s", sessionID)

	// Delete using pipeline
	pipe := r.client.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.Del(ctx, manifestKey)

	// Delete all segment keys using pattern
	segmentPattern := fmt.Sprintf("playback:hls:segment:%s:*", sessionID)
	iter := r.client.Scan(ctx, 0, segmentPattern, 0).Iterator()
	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		r.logger.Error().Err(err).Msg("Failed to scan segment keys")
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	r.logger.Info().Str("session_id", sessionID).Msg("Session deleted")
	return nil
}

// ExtendSessionTTL extends the TTL of a session
func (r *CacheRepository) ExtendSessionTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf("playback:session:%s", sessionID)
	err := r.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to extend session TTL: %w", err)
	}

	return nil
}
