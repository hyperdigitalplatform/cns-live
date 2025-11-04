package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	Milestone MilestoneConfig
}

// ServerConfig holds the HTTP server configuration
type ServerConfig struct {
	Port int
}

// MilestoneConfig holds the Milestone XProtect configuration
type MilestoneConfig struct {
	BaseURL  string
	Username string
	Password string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvInt("PORT", 8080),
		},
		Milestone: MilestoneConfig{
			BaseURL:  getEnv("MILESTONE_BASE_URL", "https://192.168.1.9"),
			Username: getEnv("MILESTONE_USERNAME", ""),
			Password: getEnv("MILESTONE_PASSWORD", ""),
		},
	}

	// Validate required fields
	if cfg.Milestone.Username == "" {
		return nil, fmt.Errorf("MILESTONE_USERNAME is required")
	}
	if cfg.Milestone.Password == "" {
		return nil, fmt.Errorf("MILESTONE_PASSWORD is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
