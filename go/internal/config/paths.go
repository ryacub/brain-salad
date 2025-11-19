package config

import (
	"os"
	"path/filepath"
)

// EnsureDataDir ensures the data directory exists
func EnsureDataDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}

// DefaultTelosPath returns the default path to telos.md
func DefaultTelosPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "telos.md"
	}
	return filepath.Join(home, "telos.md")
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
