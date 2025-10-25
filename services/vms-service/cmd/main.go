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

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	httpdelivery "github.com/rta/cctv/vms-service/internal/delivery/http"
	"github.com/rta/cctv/vms-service/internal/repository/cache"
	"github.com/rta/cctv/vms-service/internal/repository/postgres"
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
		Str("service", "vms-service").
		Str("version", "1.0.0").
		Logger()

	logger.Info().Msg("Starting VMS Service")

	// Initialize PostgreSQL connection
	ctx := context.Background()
	db, err := initDatabase(ctx, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	logger.Info().Msg("Connected to PostgreSQL database")

	// Initialize PostgreSQL repository
	cameraRepo := postgres.NewPostgresRepository(db, logger)

	// Initialize cache
	cacheTTL := 5 * time.Minute
	cleanupInterval := 10 * time.Minute
	cacheRepo := cache.NewMemoryCache(cacheTTL, cleanupInterval)

	logger.Info().
		Dur("cache_ttl", cacheTTL).
		Msg("Initialized in-memory cache")

	// Initialize HTTP handler
	handler := httpdelivery.NewHandler(cameraRepo, cacheRepo, logger)

	// Create router
	router := httpdelivery.NewRouter(handler)

	// Start HTTP server
	port := getEnv("PORT", "8081")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("port", port).Msg("VMS Service listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info().Msg("Shutting down VMS Service")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("VMS Service stopped")
}

// initDatabase initializes PostgreSQL connection and runs migrations
func initDatabase(ctx context.Context, logger zerolog.Logger) (*sql.DB, error) {
	// Get database connection string from environment
	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbName := getEnv("POSTGRES_DB", "cctv")
	dbUser := getEnv("POSTGRES_USER", "cctv")
	dbPassword := getEnv("POSTGRES_PASSWORD", "changeme_postgres")
	dbSSLMode := getEnv("POSTGRES_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(ctx, db, logger); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations executes database migrations
func runMigrations(ctx context.Context, db *sql.DB, logger zerolog.Logger) error {
	logger.Info().Msg("Running database migrations")

	// Read migration file
	migrationSQL, err := os.ReadFile("migrations/001_create_cameras_table.sql")
	if err != nil {
		logger.Warn().Err(err).Msg("Migration file not found, skipping")
		return nil // Don't fail if migrations directory doesn't exist in container
	}

	// Execute migration
	if _, err := db.ExecContext(ctx, string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	logger.Info().Msg("Database migrations completed successfully")
	return nil
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
