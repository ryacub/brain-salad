// Package main provides the web server entry point for the Telos Idea Matrix application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/api"
	"github.com/ryacub/telos-idea-matrix/internal/config"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/logging"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Application error")
	}
}

func run() error {
	// Initialize logging
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	logDir := filepath.Join(homeDir, ".telos-idea-matrix", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Warn().Err(err).Str("log_dir", logDir).Msg("failed to create log directory")
	}

	logCfg := logging.Config{
		Level:      "info",
		Format:     "json",
		OutputPath: filepath.Join(logDir, "telos-matrix.log"),
		MaxSizeMB:  10,
		MaxBackups: 7,
		MaxAgeDays: 7,
	}
	logging.NewLogger(logCfg)

	log.Info().Msg("Telos Idea Matrix starting...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Ensure data directory exists
	if err := config.EnsureDataDir(cfg.Database.Path); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize database
	repo, err := database.NewRepository(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close repository")
		}
	}()

	log.Info().Str("database_path", cfg.Database.Path).Msg("Database initialized")

	// Check if telos file exists
	if !config.FileExists(cfg.Telos.FilePath) {
		log.Warn().Str("telos_path", cfg.Telos.FilePath).Msg("Telos file not found")
		log.Warn().Msg("The server will start, but idea analysis may not work correctly")
		log.Warn().Msg("Create a telos.md file with your goals, strategies, and failure patterns")
	}

	// Log authentication status
	if cfg.Auth.Enabled {
		log.Info().Str("mode", cfg.Auth.Mode).Msg("Authentication enabled")
		log.Info().Int("api_keys", len(cfg.Auth.APIKeys)).Msg("API keys configured")
	} else {
		log.Info().Msg("Authentication disabled (local development mode)")
	}

	// Create API server
	server, err := api.NewServerFromPath(repo, cfg.Telos.FilePath, cfg.Auth)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close server")
		}
	}()

	// Setup graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Spawn background tasks with simple tickers
	stopTasks := make(chan struct{})
	setupBackgroundTasks(repo, stopTasks)

	// Start server in goroutine
	go func() {
		addr := cfg.Address()
		log.Info().Str("address", addr).Msg("Starting Telos Idea Matrix API server")
		log.Info().Str("health_check", "http://"+addr+"/health").Msg("Health check endpoint")
		log.Info().Str("api_endpoints", "http://"+addr+"/api/v1/*").Msg("API endpoints")
		log.Info().Str("metrics", "http://"+addr+"/metrics").Msg("Metrics endpoint")

		if err := server.Start(addr); err != nil {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	log.Info().Msg("Server started. Press Ctrl+C to shutdown")

	// Wait for interrupt signal
	<-done
	log.Info().Msg("Shutting down gracefully...")

	// Stop background tasks
	close(stopTasks)

	return nil
}

// setupBackgroundTasks initializes and spawns background tasks using simple tickers
func setupBackgroundTasks(repo *database.Repository, stop <-chan struct{}) {
	// Database cleanup task - runs once per day
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		log.Info().Msg("Started database cleanup task (runs every 24 hours)")

		for {
			select {
			case <-ticker.C:
				log.Info().Msg("Running database vacuum")
				if _, err := repo.DB().Exec("VACUUM"); err != nil {
					log.Warn().Err(err).Msg("Database vacuum failed")
				} else {
					log.Info().Msg("Database vacuum completed")
				}
			case <-stop:
				log.Info().Msg("Stopping database cleanup task")
				return
			}
		}
	}()

	// Metrics collection task - runs every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		log.Info().Msg("Started metrics collection task (runs every 5 minutes)")

		for {
			select {
			case <-ticker.C:
				stats := repo.DB().Stats()
				log.Debug().
					Int("open_connections", stats.OpenConnections).
					Int("in_use", stats.InUse).
					Msg("Database connection stats")
			case <-stop:
				log.Info().Msg("Stopping metrics collection task")
				return
			}
		}
	}()

	// Health check task - runs every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		log.Info().Msg("Started health check task (runs every 30 seconds)")

		for {
			select {
			case <-ticker.C:
				if err := repo.Ping(); err != nil {
					log.Error().Err(err).Msg("Database health check failed")
				}
			case <-stop:
				log.Info().Msg("Stopping health check task")
				return
			}
		}
	}()
}
