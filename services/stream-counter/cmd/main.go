package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	httpdelivery "github.com/rta/cctv/stream-counter/internal/delivery/http"
	"github.com/rta/cctv/stream-counter/internal/domain"
	"github.com/rta/cctv/stream-counter/pkg/valkey"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("No .env file found, using environment variables")
	}

	// Initialize logger
	logLevel := getEnv("LOG_LEVEL", "info")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if getEnv("LOG_FORMAT", "json") == "text" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	logger := log.With().
		Str("service", "stream-counter").
		Str("version", "1.0.0").
		Logger()

	logger.Info().Msg("Starting Stream Counter Service")

	// Initialize Valkey client
	valkeyAddr := getEnv("VALKEY_ADDR", "valkey:6379")
	valkeyPassword := getEnv("VALKEY_PASSWORD", "")
	valkeyDB := getEnvInt("VALKEY_DB", 0)
	poolSize := getEnvInt("VALKEY_POOL_SIZE", 50)

	valkeyClient, err := valkey.NewClient(valkey.Config{
		Addr:     valkeyAddr,
		Password: valkeyPassword,
		DB:       valkeyDB,
		PoolSize: poolSize,
	}, logger)

	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Valkey client")
	}
	defer valkeyClient.Close()

	logger.Info().Str("addr", valkeyAddr).Msg("Connected to Valkey")

	// Initialize stream limits
	limits := domain.LimitConfig{
		DubaiPolice: getEnvInt("LIMIT_DUBAI_POLICE", 50),
		Metro:       getEnvInt("LIMIT_METRO", 30),
		Bus:         getEnvInt("LIMIT_BUS", 20),
		Other:       getEnvInt("LIMIT_OTHER", 400),
		Total:       getEnvInt("LIMIT_TOTAL", 500),
	}

	ctx := context.Background()
	if err := valkeyClient.InitializeLimits(ctx, map[string]int{
		string(domain.SourceDubaiPolice): limits.DubaiPolice,
		string(domain.SourceMetro):       limits.Metro,
		string(domain.SourceBus):         limits.Bus,
		string(domain.SourceOther):       limits.Other,
	}); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize stream limits")
	}

	logger.Info().Interface("limits", limits).Msg("Stream limits initialized")

	// Initialize HTTP handler
	handler := httpdelivery.NewHandler(valkeyClient, logger)

	// Create router
	router := httpdelivery.NewRouter(handler)

	// Start background cleanup job
	go backgroundCleanup(ctx, valkeyClient, logger)

	// Start HTTP server
	port := getEnv("PORT", "8087")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("port", port).Msg("Stream Counter Service listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info().Msg("Shutting down Stream Counter Service")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Stream Counter Service stopped")
}

// backgroundCleanup performs periodic cleanup of stale reservations
func backgroundCleanup(ctx context.Context, client *valkey.Client, logger zerolog.Logger) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Debug().Msg("Starting background cleanup of stale reservations")

			maxAge := 3600 // 1 hour
			cleanedCount, sourcesAffected, err := client.CleanupStale(ctx, maxAge)

			if err != nil {
				logger.Error().Err(err).Msg("Background cleanup failed")
			} else if cleanedCount > 0 {
				logger.Info().
					Int("cleaned_count", cleanedCount).
					Str("sources_affected", sourcesAffected).
					Msg("Cleaned up stale reservations")
			} else {
				logger.Debug().Msg("No stale reservations found")
			}

		case <-ctx.Done():
			logger.Info().Msg("Stopping background cleanup")
			return
		}
	}
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt retrieves integer environment variable with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
