// Package config provides application configuration management with environment variable support.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Telos    TelosConfig
	Auth     AuthConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         int
	Host         string
	AllowOrigins []string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string
}

// TelosConfig holds telos file configuration
type TelosConfig struct {
	FilePath string
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("PORT", 8080),
			Host:         getEnv("HOST", "0.0.0.0"),
			AllowOrigins: getEnvAsSlice("ALLOW_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "data/telos.db"),
		},
		Telos: TelosConfig{
			FilePath: getEnv("TELOS_PATH", "telos.md"),
		},
		Auth: LoadAuthConfig(),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Server.Port)
	}

	if c.Server.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	if c.Telos.FilePath == "" {
		return fmt.Errorf("telos file path cannot be empty")
	}

	return nil
}

// Address returns the server address as "host:port"
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Helper functions for environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// For now, return as single-element slice
	// In production, you might want to split on comma
	return []string{valueStr}
}
