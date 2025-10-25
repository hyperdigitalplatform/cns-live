package valkey

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

//go:embed scripts/lua/*.lua
var luaScripts embed.FS

// Client wraps redis client with Lua script support
type Client struct {
	rdb     *redis.Client
	scripts map[string]*redis.Script
	logger  zerolog.Logger
}

// Config holds Valkey client configuration
type Config struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// NewClient creates a new Valkey client with Lua scripts loaded
func NewClient(cfg Config, logger zerolog.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	client := &Client{
		rdb:     rdb,
		scripts: make(map[string]*redis.Script),
		logger:  logger,
	}

	// Load Lua scripts
	if err := client.loadScripts(); err != nil {
		return nil, fmt.Errorf("failed to load Lua scripts: %w", err)
	}

	logger.Info().
		Str("addr", cfg.Addr).
		Int("pool_size", cfg.PoolSize).
		Msg("Valkey client initialized")

	return client, nil
}

// loadScripts loads all Lua scripts from embedded filesystem
func (c *Client) loadScripts() error {
	scriptFiles := []string{
		"reserve_stream.lua",
		"release_stream.lua",
		"heartbeat_stream.lua",
		"get_stats.lua",
		"cleanup_stale.lua",
	}

	for _, filename := range scriptFiles {
		scriptContent, err := luaScripts.ReadFile("scripts/lua/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read script %s: %w", filename, err)
		}

		scriptName := filename[:len(filename)-4] // Remove .lua extension
		c.scripts[scriptName] = redis.NewScript(string(scriptContent))

		c.logger.Debug().Str("script", scriptName).Msg("Loaded Lua script")
	}

	return nil
}

// ReserveStream atomically reserves a stream slot
func (c *Client) ReserveStream(ctx context.Context, source, reservationID, cameraID, userID string, ttl int) (success bool, current int, limit int, err error) {
	script := c.scripts["reserve_stream"]
	if script == nil {
		return false, 0, 0, fmt.Errorf("reserve_stream script not loaded")
	}

	result, err := script.Run(ctx, c.rdb, nil, source, reservationID, cameraID, userID, ttl).Result()
	if err != nil {
		return false, 0, 0, fmt.Errorf("reserve script failed: %w", err)
	}

	// Parse result: {success, current, limit}
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 3 {
		return false, 0, 0, fmt.Errorf("invalid reserve script result")
	}

	successInt, _ := resultSlice[0].(int64)
	currentInt, _ := resultSlice[1].(int64)
	limitInt, _ := resultSlice[2].(int64)

	return successInt == 1, int(currentInt), int(limitInt), nil
}

// ReleaseStream atomically releases a stream reservation
func (c *Client) ReleaseStream(ctx context.Context, reservationID string) (success bool, newCount int, source string, err error) {
	script := c.scripts["release_stream"]
	if script == nil {
		return false, 0, "", fmt.Errorf("release_stream script not loaded")
	}

	result, err := script.Run(ctx, c.rdb, nil, reservationID).Result()
	if err != nil {
		return false, 0, "", fmt.Errorf("release script failed: %w", err)
	}

	// Parse result: {success, new_count, source}
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 3 {
		return false, 0, "", fmt.Errorf("invalid release script result")
	}

	successInt, _ := resultSlice[0].(int64)
	newCountInt, _ := resultSlice[1].(int64)
	sourceStr, _ := resultSlice[2].(string)

	return successInt == 1, int(newCountInt), sourceStr, nil
}

// HeartbeatStream updates heartbeat and extends TTL
func (c *Client) HeartbeatStream(ctx context.Context, reservationID string, ttlExtension int) (success bool, remainingTTL int, err error) {
	script := c.scripts["heartbeat_stream"]
	if script == nil {
		return false, 0, fmt.Errorf("heartbeat_stream script not loaded")
	}

	result, err := script.Run(ctx, c.rdb, nil, reservationID, ttlExtension).Result()
	if err != nil {
		return false, 0, fmt.Errorf("heartbeat script failed: %w", err)
	}

	// Parse result: {success, remaining_ttl, message}
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 2 {
		return false, 0, fmt.Errorf("invalid heartbeat script result")
	}

	successInt, _ := resultSlice[0].(int64)
	remainingInt, _ := resultSlice[1].(int64)

	return successInt == 1, int(remainingInt), nil
}

// GetStats retrieves current stream statistics for all sources
func (c *Client) GetStats(ctx context.Context, sources []string) ([][4]interface{}, error) {
	script := c.scripts["get_stats"]
	if script == nil {
		return nil, fmt.Errorf("get_stats script not loaded")
	}

	sourcesStr := ""
	for i, source := range sources {
		if i > 0 {
			sourcesStr += ","
		}
		sourcesStr += source
	}

	result, err := script.Run(ctx, c.rdb, nil, sourcesStr).Result()
	if err != nil {
		return nil, fmt.Errorf("get_stats script failed: %w", err)
	}

	// Parse result: [[source, current, limit, percentage], ...]
	resultSlice, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid stats script result")
	}

	stats := make([][4]interface{}, len(resultSlice))
	for i, item := range resultSlice {
		itemSlice, ok := item.([]interface{})
		if !ok || len(itemSlice) < 4 {
			continue
		}
		copy(stats[i][:], itemSlice)
	}

	return stats, nil
}

// CleanupStale removes stale reservations
func (c *Client) CleanupStale(ctx context.Context, maxAge int) (cleanedCount int, sourcesAffected string, err error) {
	script := c.scripts["cleanup_stale"]
	if script == nil {
		return 0, "", fmt.Errorf("cleanup_stale script not loaded")
	}

	result, err := script.Run(ctx, c.rdb, nil, maxAge).Result()
	if err != nil {
		return 0, "", fmt.Errorf("cleanup script failed: %w", err)
	}

	// Parse result: {cleaned_count, sources_affected}
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 2 {
		return 0, "", fmt.Errorf("invalid cleanup script result")
	}

	cleanedInt, _ := resultSlice[0].(int64)
	sourcesStr, _ := resultSlice[1].(string)

	return int(cleanedInt), sourcesStr, nil
}

// InitializeLimits sets the initial limits for all sources
func (c *Client) InitializeLimits(ctx context.Context, limits map[string]int) error {
	pipe := c.rdb.Pipeline()

	for source, limit := range limits {
		limitKey := fmt.Sprintf("stream:limit:%s", source)
		pipe.Set(ctx, limitKey, limit, 0) // No expiration

		// Initialize count to 0 if not exists
		countKey := fmt.Sprintf("stream:count:%s", source)
		pipe.SetNX(ctx, countKey, 0, 0)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize limits: %w", err)
	}

	c.logger.Info().Interface("limits", limits).Msg("Initialized stream limits")
	return nil
}

// GetCurrentCount retrieves current stream count for a source
func (c *Client) GetCurrentCount(ctx context.Context, source string) (int, error) {
	countKey := fmt.Sprintf("stream:count:%s", source)
	result, err := c.rdb.Get(ctx, countKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get current count: %w", err)
	}
	return result, nil
}

// GetLimit retrieves the limit for a source
func (c *Client) GetLimit(ctx context.Context, source string) (int, error) {
	limitKey := fmt.Sprintf("stream:limit:%s", source)
	result, err := c.rdb.Get(ctx, limitKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get limit: %w", err)
	}
	return result, nil
}

// Ping checks if Valkey is reachable
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Valkey connection
func (c *Client) Close() error {
	return c.rdb.Close()
}
