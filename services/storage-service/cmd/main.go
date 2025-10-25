package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	httpDelivery "github.com/rta/cctv/storage-service/internal/delivery/http"
	"github.com/rta/cctv/storage-service/internal/repository/minio"
	"github.com/rta/cctv/storage-service/internal/repository/postgres"
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

	logger.Info().Msg("Starting Storage Service")

	// Database connection
	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "cctv")
	dbPassword := getEnv("POSTGRES_PASSWORD", "changeme")
	dbName := getEnv("POSTGRES_DB", "cctv")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to ping database")
	}
	logger.Info().Msg("Connected to PostgreSQL")

	// MinIO storage
	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "storage-service")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "changeme")
	minioBucket := getEnv("MINIO_BUCKET_RECORDINGS", "cctv-recordings")
	minioUseSSL := getEnv("MINIO_USE_SSL", "false") == "true"

	storageRepo, err := minio.NewMinIOStorage(
		minioEndpoint,
		minioAccessKey,
		minioSecretKey,
		minioBucket,
		minioUseSSL,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MinIO storage")
	}
	logger.Info().Msg("Connected to MinIO")

	// Repositories
	segmentRepo := postgres.NewSegmentRepository(db, logger)
	exportRepo := postgres.NewExportRepository(db, logger)

	// HTTP handler
	handler := httpDelivery.NewHandler(storageRepo, segmentRepo, exportRepo, logger)
	router := httpDelivery.NewRouter(handler)

	// HTTP server
	port := getEnv("PORT", "8082")
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
