package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/api"
	"github.com/rayyacub/telos-idea-matrix/internal/config"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/logging"
	"github.com/rayyacub/telos-idea-matrix/internal/tasks"
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

	// Create task manager for background tasks
	taskManager := tasks.NewTaskManager()
	log.Info().Msg("Task manager initialized")

	// Spawn background tasks
	setupBackgroundTasks(taskManager, repo)

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

	// Shutdown background tasks
	if err := taskManager.Shutdown(5 * time.Second); err != nil {
		log.Warn().Err(err).Msg("Task manager shutdown warning")
	}

	return nil
}

// setupBackgroundTasks initializes and spawns background tasks
func setupBackgroundTasks(tm *tasks.TaskManager, repo *database.Repository) {
	// Database cleanup task - runs every hour
	cleanupTask := tasks.NewScheduledTask(
		"database-cleanup",
		1*time.Hour,
		func(ctx context.Context) error {
			log.Info().Msg("Running database cleanup task")
			// Placeholder for actual cleanup logic
			// Could vacuum, remove old records, optimize indexes, etc.
			return nil
		},
	).WithTimeout(10 * time.Minute)

	tm.Spawn(cleanupTask)
	log.Info().Msg("Spawned database cleanup task (runs every 1 hour)")

	// Metrics collection task - runs every 5 minutes
	metricsTask := tasks.NewScheduledTask(
		"metrics-collection",
		5*time.Minute,
		func(ctx context.Context) error {
			log.Debug().Msg("Collecting metrics")
			// Placeholder for metrics collection
			// Could collect database stats, memory usage, etc.
			return nil
		},
	).WithTimeout(1 * time.Minute)

	tm.Spawn(metricsTask)
	log.Info().Msg("Spawned metrics collection task (runs every 5 minutes)")

	// Health check task - runs every 30 seconds
	healthCheckTask := tasks.NewScheduledTask(
		"health-check",
		30*time.Second,
		func(ctx context.Context) error {
			// Verify database connection
			if err := repo.Ping(); err != nil {
				log.Error().Err(err).Msg("Database health check failed")
				return err
			}
			return nil
		},
	).WithTimeout(10 * time.Second)

	tm.Spawn(healthCheckTask)
	log.Info().Msg("Spawned health check task (runs every 30 seconds)")
}
