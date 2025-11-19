package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/rayyacub/telos-idea-matrix/internal/api"
	"github.com/rayyacub/telos-idea-matrix/internal/config"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/logging"
	"github.com/rs/zerolog/log"
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
	os.MkdirAll(logDir, 0755)

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
	defer repo.Close()

	log.Info().Str("database_path", cfg.Database.Path).Msg("Database initialized")

	// Check if telos file exists
	if !config.FileExists(cfg.Telos.FilePath) {
		log.Warn().Str("telos_path", cfg.Telos.FilePath).Msg("Telos file not found")
		log.Warn().Msg("The server will start, but idea analysis may not work correctly")
		log.Warn().Msg("Create a telos.md file with your goals, strategies, and failure patterns")
	}

	// Create API server
	server, err := api.NewServerFromPath(repo, cfg.Telos.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer server.Close()

	// Setup graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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

	return nil
}
