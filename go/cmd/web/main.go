package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rayyacub/telos-idea-matrix/internal/api"
	"github.com/rayyacub/telos-idea-matrix/internal/config"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
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

	log.Printf("Database initialized: %s", cfg.Database.Path)

	// Check if telos file exists
	if !config.FileExists(cfg.Telos.FilePath) {
		log.Printf("Warning: Telos file not found at %s", cfg.Telos.FilePath)
		log.Printf("The server will start, but idea analysis may not work correctly.")
		log.Printf("Create a telos.md file with your goals, strategies, and failure patterns.")
	}

	// Create API server
	server, err := api.NewServer(repo, cfg.Telos.FilePath)
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
		log.Printf("Starting Telos Idea Matrix API server on %s", addr)
		log.Printf("Health check: http://%s/health", addr)
		log.Printf("API endpoints: http://%s/api/v1/*", addr)

		if err := server.Start(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("Server started. Press Ctrl+C to shutdown.")

	// Wait for interrupt signal
	<-done
	log.Println("Shutting down gracefully...")

	return nil
}
