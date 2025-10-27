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
	"github.com/redis/go-redis/v9"
	"github.com/rta/cctv/go-api/internal/client"
	deliveryHttp "github.com/rta/cctv/go-api/internal/delivery/http"
	deliveryWS "github.com/rta/cctv/go-api/internal/delivery/websocket"
	"github.com/rta/cctv/go-api/internal/repository/postgres"
	"github.com/rta/cctv/go-api/internal/repository/valkey"
	"github.com/rta/cctv/go-api/internal/usecase"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Load configuration
	config := loadConfig()

	// Connect to Valkey
	valkeyClient := redis.NewClient(&redis.Options{
		Addr:     config.ValkeyAddr,
		Password: config.ValkeyPassword,
		DB:       config.ValkeyDB,
	})

	ctx := context.Background()
	if err := valkeyClient.Ping(ctx).Err(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Valkey")
	}
	defer valkeyClient.Close()

	logger.Info().Msg("Connected to Valkey")

	// Connect to PostgreSQL
	db, err := initPostgreSQL(ctx, config, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()

	logger.Info().Msg("Connected to PostgreSQL")

	// Initialize repositories
	streamRepo := valkey.NewStreamRepository(valkeyClient, logger)
	layoutRepo := postgres.NewLayoutRepository(db, logger)

	// Initialize clients
	streamCounterClient := client.NewStreamCounterClient(config.StreamCounterURL, logger)
	vmsClient := client.NewVMSClient(config.VMSServiceURL, logger)
	livekitClient := client.NewLiveKitClient(
		config.LiveKitURL,
		config.LiveKitAPIKey,
		config.LiveKitAPISecret,
		logger,
	)
	mediaMTXClient := client.NewMediaMTXClient(config.MediaMTXURL, logger)
	livekitIngressClient := client.NewLiveKitIngressClient(
		config.LiveKitURL,
		config.LiveKitAPIKey,
		config.LiveKitAPISecret,
		logger,
	)

	// Initialize Docker client for managing WHIP pusher containers
	dockerClient, err := client.NewDockerClient(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Docker client")
	}
	defer dockerClient.Close()

	// Initialize use cases
	streamUseCase := usecase.NewStreamUseCase(
		streamCounterClient,
		vmsClient,
		livekitClient,
		mediaMTXClient,
		livekitIngressClient,
		dockerClient,
		streamRepo,
		config.LiveKitWSURL,
		logger,
	)
	layoutUseCase := usecase.NewLayoutUseCase(layoutRepo, logger)

	// Initialize WebSocket hub
	wsHub := deliveryWS.NewHub(streamUseCase, logger)
	go wsHub.Run(ctx)

	// Initialize HTTP handlers
	streamHandler := deliveryHttp.NewStreamHandler(streamUseCase, logger)
	cameraHandler := deliveryHttp.NewCameraHandler(vmsClient, logger)
	wsHandler := deliveryWS.NewHandler(wsHub, logger)
	layoutHandler := deliveryHttp.NewLayoutHandler(layoutUseCase, logger)

	// Setup router
	router := deliveryHttp.NewRouter(streamHandler, cameraHandler, wsHandler, layoutHandler)

	// Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second, // Increased for slow PTZ camera responses (TrueView ~16s)
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("port", config.Port).Msg("Starting Go API Service")
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
	Port               string
	ValkeyAddr         string
	ValkeyPassword     string
	ValkeyDB           int
	PostgresHost       string
	PostgresPort       string
	PostgresDB         string
	PostgresUser       string
	PostgresPassword   string
	StreamCounterURL   string
	VMSServiceURL      string
	MediaMTXURL        string // MediaMTX API URL
	LiveKitURL         string // Internal LiveKit URL for API calls
	LiveKitWSURL       string // External LiveKit WebSocket URL for clients
	LiveKitAPIKey      string
	LiveKitAPISecret   string
}

func loadConfig() Config {
	return Config{
		Port:               getEnv("PORT", "8086"),
		ValkeyAddr:         getEnv("VALKEY_ADDR", "localhost:6379"),
		ValkeyPassword:     getEnv("VALKEY_PASSWORD", ""),
		ValkeyDB:           getEnvInt("VALKEY_DB", 0),
		PostgresHost:       getEnv("POSTGRES_HOST", "postgres"),
		PostgresPort:       getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:         getEnv("POSTGRES_DB", "cctv"),
		PostgresUser:       getEnv("POSTGRES_USER", "cctv"),
		PostgresPassword:   getEnv("POSTGRES_PASSWORD", "changeme_postgres"),
		StreamCounterURL:   getEnv("STREAM_COUNTER_URL", "http://localhost:8087"),
		VMSServiceURL:      getEnv("VMS_SERVICE_URL", "http://localhost:8081"),
		MediaMTXURL:        getEnv("MEDIAMTX_URL", "http://localhost:9997"),
		LiveKitURL:         getEnv("LIVEKIT_URL", "http://localhost:7880"),
		LiveKitWSURL:       getEnv("LIVEKIT_WS_URL", "ws://localhost:7880"),
		LiveKitAPIKey:      getEnv("LIVEKIT_API_KEY", "devkey"),
		LiveKitAPISecret:   getEnv("LIVEKIT_API_SECRET", "devsecret"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		fmt.Sscanf(value, "%d", &intVal)
		return intVal
	}
	return defaultValue
}

func initPostgreSQL(ctx context.Context, config Config, logger zerolog.Logger) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDB,
	)

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

	return db, nil
}
