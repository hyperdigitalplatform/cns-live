package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	httpDelivery "github.com/rta/cctv/recording-service/internal/delivery/http"
	"github.com/rta/cctv/recording-service/internal/manager"
)

func main() {
	// Setup logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if os.Getenv("LOG_FORMAT") == "text" {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err == nil {
		zerolog.SetGlobalLevel(level)
	}

	logger.Info().Msg("Starting Recording Service")

	// Configuration
	outputDir := getEnv("OUTPUT_DIR", "/tmp/recordings")
	segmentSeconds := 3600 // 1 hour segments
	storageURL := getEnv("STORAGE_SERVICE_URL", "http://storage-service:8082")
	vmsURL := getEnv("VMS_SERVICE_URL", "http://vms-service:8081")
	streamCounterURL := getEnv("STREAM_COUNTER_URL", "http://stream-counter:8087")

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Fatal().Err(err).Msg("Failed to create output directory")
	}

	// Recording manager
	recordingManager := manager.NewRecordingManager(
		outputDir,
		segmentSeconds,
		storageURL,
		vmsURL,
		streamCounterURL,
		logger,
	)

	// HTTP handler
	handler := httpDelivery.NewHandler(recordingManager, logger)
	router := httpDelivery.NewRouter(handler)

	// HTTP server
	port := getEnv("PORT", "8083")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("port", port).Msg("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Stop all recordings
	for cameraID := range recordingManager.GetAllRecordings() {
		logger.Info().Str("camera_id", cameraID).Msg("Stopping recording")
		if err := recordingManager.StopRecording(cameraID); err != nil {
			logger.Warn().Err(err).Str("camera_id", cameraID).Msg("Error stopping recording")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
