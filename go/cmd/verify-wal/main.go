package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
)

func main() {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".telos", "ideas.db")

	// Use NewRepository to open the database with WAL mode
	repo, err := database.NewRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	db := repo.DB()

	var journalMode string
	db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	fmt.Printf("Journal Mode: %s\n", journalMode)

	var synchronous string
	db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	fmt.Printf("Synchronous: %s\n", synchronous)

	var busyTimeout int
	db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	fmt.Printf("Busy Timeout: %d ms\n", busyTimeout)

	var cacheSize int
	db.QueryRow("PRAGMA cache_size").Scan(&cacheSize)
	fmt.Printf("Cache Size: %d pages (negative = KB)\n", cacheSize)

	var tempStore int
	db.QueryRow("PRAGMA temp_store").Scan(&tempStore)
	tempStoreStr := "DEFAULT"
	if tempStore == 1 {
		tempStoreStr = "FILE"
	} else if tempStore == 2 {
		tempStoreStr = "MEMORY"
	}
	fmt.Printf("Temp Store: %s\n", tempStoreStr)

	var foreignKeys int
	db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
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
