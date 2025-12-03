//go:build integration

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/patterns"
	"github.com/ryacub/telos-idea-matrix/internal/scoring"
	"github.com/ryacub/telos-idea-matrix/internal/telos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCLI(t *testing.T) (*CLIContext, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tm-cli-test-*")
	require.NoError(t, err)

	// Create test telos file
	telosPath := filepath.Join(tmpDir, "telos.md")
	telosContent := `## Goals
- G1: Generate $2500/month revenue (Deadline: 2025-03-31)

## Strategies
- S1: The "One Stack, One Month" Rule

## Stack
- Primary: Go, Python, LangChain, OpenAI
- Secondary: Docker, PostgreSQL

## Failure Patterns
- Context switching: Starting new projects before finishing current ones
- Perfectionism: Over-engineering before validating market fit
`
	err = os.WriteFile(telosPath, []byte(telosContent), 0644)
	require.NoError(t, err)

	// Parse telos
	parser := telos.NewParser()
	telosData, err := parser.ParseFile(telosPath)
	require.NoError(t, err)

	// Create test database
	dbPath := filepath.Join(tmpDir, "test.db")
	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)

	// Create engine and detector
	engine := scoring.NewEngine(telosData)
	detector := patterns.NewDetector(telosData)

	cliCtx := &CLIContext{
		Repository: repo,
		Engine:     engine,
		Detector:   detector,
		Telos:      telosData,
		DBPath:     dbPath,
		TelosPath:  telosPath,
	}

	cleanup := func() {
		ClearContext()
		os.RemoveAll(tmpDir)
	}

	return cliCtx, cleanup
}

func TestAddCommand_WithIdeaText_SavesAndDisplays(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Execute add command with test telos and db paths
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"add", "Build a SaaS product using Go and AI agents",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify idea was saved
	limit := 10
	ideas, err := cliCtx.Repository.List(database.ListOptions{
		Limit: &limit,
	})
	require.NoError(t, err)
	assert.Len(t, ideas, 1)

	// Verify idea content
	idea := ideas[0]
	assert.Contains(t, idea.Content, "SaaS")
	assert.Contains(t, idea.Content, "Go")
	assert.Greater(t, idea.FinalScore, 0.0)
	assert.NotEmpty(t, idea.Recommendation)
}

func TestAddCommand_HighScoreIdea_GetsCorrectRecommendation(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// High-scoring idea (uses primary stack, AI-focused, revenue potential)
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"add", "Build an AI agent using Go and LangChain with $2000/month SaaS subscription model, MVP in 30 days",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify high score
	limit := 1
	ideas, err := cliCtx.Repository.List(database.ListOptions{
		Limit: &limit,
	})
	require.NoError(t, err)
	require.Len(t, ideas, 1)

	idea := ideas[0]
	assert.Greater(t, idea.FinalScore, 7.0, "Expected high score for well-aligned idea")
	// Should be either PRIORITIZE NOW or GOOD ALIGNMENT (both are good outcomes)
	assert.True(t,
		strings.Contains(idea.Recommendation, "PRIORITIZE") || strings.Contains(idea.Recommendation, "GOOD ALIGNMENT"),
		"Expected PRIORITIZE NOW or GOOD ALIGNMENT recommendation, got: %s", idea.Recommendation)
}

func TestAddCommand_LowScoreIdea_GetsCorrectRecommendation(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Low-scoring idea (wrong stack, vague, no revenue model)
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"add", "Learn Rust and build a comprehensive framework from scratch",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify low score
	limit := 1
	ideas, err := cliCtx.Repository.List(database.ListOptions{
		Limit: &limit,
	})
	require.NoError(t, err)
	require.Len(t, ideas, 1)

	idea := ideas[0]
	assert.Less(t, idea.FinalScore, 6.0, "Expected low score for misaligned idea")
	assert.True(t,
		strings.Contains(idea.Recommendation, "AVOID") || strings.Contains(idea.Recommendation, "CONSIDER"),
		"Expected AVOID or CONSIDER LATER recommendation")
}

func TestAddCommand_DetectsPatterns(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Idea with context switching pattern
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"add", "Learn Rust and Flutter before building anything",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify patterns detected
	limit := 1
	ideas, err := cliCtx.Repository.List(database.ListOptions{
		Limit: &limit,
	})
	require.NoError(t, err)
	require.Len(t, ideas, 1)

	idea := ideas[0]
	assert.NotEmpty(t, idea.Patterns, "Expected patterns to be detected")
}

func TestAddCommand_NoArgs_ReturnsError(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"add",
	})

	err := cmd.Execute()
	assert.Error(t, err, "Expected error when no idea text provided")
}
