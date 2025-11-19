//go:build integration
// +build integration

package integration

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rayyacub/telos-idea-matrix/internal/api"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
	"github.com/stretchr/testify/require"
)

// testServerConfig holds configuration for test server setup
type testServerConfig struct {
	telosContent string
}

// setupTestServer creates a test server with standard configuration
// It disables rate limiting and returns a test server with cleanup function
func setupTestServer(t *testing.T, config *testServerConfig) (*httptest.Server, *database.Repository) {
	t.Helper()

	// Disable rate limiting for all tests
	t.Setenv("DISABLE_RATE_LIMIT", "true")

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	telosPath := filepath.Join(tempDir, "telos.md")

	// Use provided telos content or default
	telosContent := config.telosContent
	if telosContent == "" {
		telosContent = `# Telos
## Goals
- G1: Goal 1
`
	}

	require.NoError(t, os.WriteFile(telosPath, []byte(telosContent), 0644))

	telosConfig, err := telos.ParseTelosFile(telosPath)
	require.NoError(t, err)

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })

	server := api.NewServer(repo, telosConfig)
	ts := httptest.NewServer(server.Router())
	t.Cleanup(func() { ts.Close() })

	return ts, repo
}
