package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rta/cctv/playback-service/internal/cache"
	"github.com/rta/cctv/playback-service/internal/client"
	deliveryHttp "github.com/rta/cctv/playback-service/internal/delivery/http"
	"github.com/rta/cctv/playback-service/internal/transmux"
	"github.com/rta/cctv/playback-service/internal/usecase"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Load configuration
	config := loadConfig()

	ctx := context.Background()

	// Initialize MinIO client
	minioClient, err := client.NewMinIOClient(
		config.MinIOEndpoint,
		config.MinIOAccessKey,
		config.MinIOSecretKey,
		config.MinioBucket,
		config.MinIOUseSSL,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MinIO client")
	}

	// Check if bucket exists
	exists, err := minioClient.CheckBucketExists(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to check bucket existence")
	}
	if !exists {
		logger.Fatal().Str("bucket", config.MinioBucket).Msg("MinIO bucket does not exist")
	}
	logger.Info().Str("bucket", config.MinioBucket).Msg("Connected to MinIO")

	// Initialize Milestone client
	milestoneClient := client.NewMilestoneClient(
		config.MilestoneURL,
		config.MilestoneUsername,
		config.MilestonePassword,
		logger,
	)

	// Initialize source detector
	sourceDetector := usecase.NewSourceDetector(
		minioClient,
		milestoneClient,
		logger,
	)

	// Initialize FFmpeg transmuxer
	ffmpegTransmuxer := transmux.NewFFmpegTransmuxer(config.WorkDir, logger)

	// Initialize segment cache (10 GB max)
	segmentCache, err := cache.NewSegmentCache(
		config.CacheDir,
		10*1024*1024*1024, // 10 GB
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create segment cache")
	}

	// Start cache cleanup worker (every 1 hour, remove entries older than 2 hours)
	segmentCache.StartCleanupWorker(1*time.Hour, 2*time.Hour)

	logger.Info().Msg("Segment cache initialized")

	// Initialize playback use case
	playbackUseCase := usecase.NewPlaybackUseCase(
		sourceDetector,
		ffmpegTransmuxer,
		segmentCache,
		minioClient,
		config.WorkDir,
		config.HLSBaseURL,
		logger,
	)

	// Initialize memory cache for milestone playback queries
	queryCache := cache.NewMemoryCache()

	// Initialize milestone playback use case
	milestonePlaybackUseCase := usecase.NewMilestonePlaybackUsecase(
		milestoneClient,
		queryCache,
		logger,
	)

	// Initialize milestone playback handler
	milestonePlaybackHandler := deliveryHttp.NewMilestonePlaybackHandler(milestonePlaybackUseCase, logger)

	// Initialize HTTP handler
	playbackHandler := deliveryHttp.NewPlaybackHandler(playbackUseCase, logger)

	// Setup router
	router := deliveryHttp.NewRouter(playbackHandler, milestonePlaybackHandler)

	// Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // Longer for video streaming
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("port", config.Port).Msg("Starting Playback Service")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

type Config struct {
	Port              string
	WorkDir           string
	CacheDir          string
	HLSBaseURL        string
	MinIOEndpoint     string
	MinIOAccessKey    string
	MinIOSecretKey    string
	MinioBucket       string
	MinIOUseSSL       bool
	MilestoneURL      string
	MilestoneUsername string
	MilestonePassword string
}

func loadConfig() Config {
	return Config{
		Port:              getEnv("PORT", "8090"),
		WorkDir:           getEnv("WORK_DIR", "/tmp/playback"),
		CacheDir:          getEnv("CACHE_DIR", "/tmp/playback/cache"),
		HLSBaseURL:        getEnv("HLS_BASE_URL", "http://localhost:8090/hls"),
		MinIOEndpoint:     getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey:    getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:    getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:       getEnv("MINIO_BUCKET", "cctv-recordings"),
		MinIOUseSSL:       getEnvBool("MINIO_USE_SSL", false),
		MilestoneURL:      getEnv("MILESTONE_URL", ""),
		MilestoneUsername: getEnv("MILESTONE_USERNAME", ""),
		MilestonePassword: getEnv("MILESTONE_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}
