// Package main provides a utility to verify WAL mode configuration for the SQLite database.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ryacub/telos-idea-matrix/internal/database"
)

func main() {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".telos", "ideas.db")

	// Use NewRepository to open the database with WAL mode
	repo, err := database.NewRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Warning: failed to close repository: %v", err)
		}
	}()

	db := repo.DB()

	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		log.Printf("Warning: failed to get journal mode: %v", err)
	}
	fmt.Printf("Journal Mode: %s\n", journalMode)

	var synchronous string
	if err := db.QueryRow("PRAGMA synchronous").Scan(&synchronous); err != nil {
		log.Printf("Warning: failed to get synchronous mode: %v", err)
	}
	fmt.Printf("Synchronous: %s\n", synchronous)

	var busyTimeout int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		log.Printf("Warning: failed to get busy timeout: %v", err)
	}
	fmt.Printf("Busy Timeout: %d ms\n", busyTimeout)

	var cacheSize int
	if err := db.QueryRow("PRAGMA cache_size").Scan(&cacheSize); err != nil {
		log.Printf("Warning: failed to get cache size: %v", err)
	}
	fmt.Printf("Cache Size: %d pages (negative = KB)\n", cacheSize)

	var tempStore int
	if err := db.QueryRow("PRAGMA temp_store").Scan(&tempStore); err != nil {
		log.Printf("Warning: failed to get temp store: %v", err)
	}
	tempStoreStr := "DEFAULT"
	switch tempStore {
	case 1:
		tempStoreStr = "FILE"
	case 2:
		tempStoreStr = "MEMORY"
	}
	fmt.Printf("Temp Store: %s\n", tempStoreStr)

	var foreignKeys int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		log.Printf("Warning: failed to get foreign keys setting: %v", err)
	}
	foreignKeysStr := "OFF"
	if foreignKeys == 1 {
		foreignKeysStr = "ON"
	}
	fmt.Printf("Foreign Keys: %s\n", foreignKeysStr)

	stats := db.Stats()
	fmt.Printf("\nConnection Pool Stats:\n")
	fmt.Printf("  Max Open Connections: %d\n", stats.MaxOpenConnections)
	fmt.Printf("  Open Connections: %d\n", stats.OpenConnections)
	fmt.Printf("  In Use: %d\n", stats.InUse)
	fmt.Printf("  Idle: %d\n", stats.Idle)

	// Check for WAL files
	walPath := dbPath + "-wal"
	shmPath := dbPath + "-shm"

	fmt.Printf("\nWAL Files:\n")
	if _, err := os.Stat(walPath); err == nil {
		info, _ := os.Stat(walPath)
		fmt.Printf("  %s (size: %d bytes)\n", walPath, info.Size())
	} else {
		fmt.Printf("  %s (not found - will be created on first write)\n", walPath)
	}

	if _, err := os.Stat(shmPath); err == nil {
		info, _ := os.Stat(shmPath)
		fmt.Printf("  %s (size: %d bytes)\n", shmPath, info.Size())
	} else {
		fmt.Printf("  %s (not found - will be created on first write)\n", shmPath)
	}
}
