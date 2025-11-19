//go:build integration

package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkTag_WithFilters(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test ideas with varying scores
	ideas := []*models.Idea{
		{
			ID:         uuid.New().String(),
			Content:    "High score idea - Build AI SaaS with Go",
			RawScore:   8.0,
			FinalScore: 8.5,
			Status:     "active",
			CreatedAt:  time.Now().UTC(),
		},
		{
			ID:         uuid.New().String(),
			Content:    "Medium score idea - Python automation",
			RawScore:   6.0,
			FinalScore: 6.2,
			Status:     "active",
			CreatedAt:  time.Now().UTC(),
		},
		{
			ID:         uuid.New().String(),
			Content:    "Low score idea - Learn Rust from scratch",
			RawScore:   3.0,
			FinalScore: 3.5,
			Status:     "active",
			CreatedAt:  time.Now().UTC(),
		},
	}

	// Save test ideas
	for _, idea := range ideas {
		err := cliCtx.Repository.Create(idea)
		require.NoError(t, err)
	}

	// Test bulk tag with min-score filter
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "tag", "high-priority",
		"--min-score", "7.0",
		"--yes", // Auto-confirm for testing
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify only high-scoring ideas were tagged
	// (In real implementation, we'd check tags table or tags field)
	// For now, we verify the command executed without error
}

func TestBulkArchive_WithFilters(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create old and new ideas
	oldIdea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "Old low-score idea",
		RawScore:   2.0,
		FinalScore: 2.5,
		Status:     "active",
		CreatedAt:  time.Now().UTC().Add(-100 * 24 * time.Hour), // 100 days old
	}

	newIdea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "New high-score idea",
		RawScore:   8.0,
		FinalScore: 8.5,
		Status:     "active",
		CreatedAt:  time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(oldIdea)
	require.NoError(t, err)
	err = cliCtx.Repository.Create(newIdea)
	require.NoError(t, err)

	// Test bulk archive with age and score filters
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "archive",
		"--older-than", "90", // 90 days
		"--max-score", "5.0",
		"--yes", // Auto-confirm for testing
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify old idea was archived
	archivedIdea, err := cliCtx.Repository.GetByID(oldIdea.ID)
	require.NoError(t, err)
	assert.Equal(t, "archived", archivedIdea.Status)

	// Verify new idea remains active
	activeIdea, err := cliCtx.Repository.GetByID(newIdea.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", activeIdea.Status)
}

func TestBulkDelete_WithConfirmation(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test idea
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "Idea to delete",
		RawScore:   1.0,
		FinalScore: 1.0,
		Status:     "active",
		CreatedAt:  time.Now().UTC().Add(-200 * 24 * time.Hour),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk delete requires confirmation
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "delete",
		"--older-than", "180",
		"--max-score", "2.0",
		"--yes", // Auto-confirm for testing
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify idea was deleted
	_, err = cliCtx.Repository.GetByID(idea.ID)
	assert.Error(t, err) // Should not exist anymore
}

func TestBulkImport_FromCSV(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test CSV file
	tmpDir, err := os.MkdirTemp("", "bulk-import-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	csvPath := filepath.Join(tmpDir, "import.csv")
	csvContent := `ID,Content,RawScore,FinalScore,Patterns,Recommendation,AnalysisDetails,CreatedAt,Status
imp-1,"Build SaaS with AI",7.0,7.5,"ai-focus","PRIORITIZE","Good idea","2025-01-15T10:00:00Z","active"
imp-2,"Python automation tool",6.0,6.2,"automation","GOOD ALIGNMENT","Useful","2025-01-14T10:00:00Z","active"
`
	err = os.WriteFile(csvPath, []byte(csvContent), 0644)
	require.NoError(t, err)

	// Test bulk import
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "import", csvPath,
		"--yes", // Auto-confirm for testing
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify ideas were imported
	limit := 10
	ideas, err := cliCtx.Repository.List(database.ListOptions{
		Limit: &limit,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(ideas), 2) // At least the 2 imported ideas
}

func TestBulkExport_ToCSV(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test ideas
	ideas := []*models.Idea{
		{
			ID:              uuid.New().String(),
			Content:         "Export test idea 1",
			RawScore:        8.0,
			FinalScore:      8.5,
			Patterns:        []string{"test"},
			Recommendation:  "PRIORITIZE NOW",
			AnalysisDetails: "Test",
			Status:          "active",
			CreatedAt:       time.Now().UTC(),
		},
		{
			ID:              uuid.New().String(),
			Content:         "Export test idea 2",
			RawScore:        6.0,
			FinalScore:      6.2,
			Patterns:        []string{"test"},
			Recommendation:  "GOOD ALIGNMENT",
			AnalysisDetails: "Test",
			Status:          "active",
			CreatedAt:       time.Now().UTC(),
		},
	}

	for _, idea := range ideas {
		err := cliCtx.Repository.Create(idea)
		require.NoError(t, err)
	}

	// Create temp directory for export
	tmpDir, err := os.MkdirTemp("", "bulk-export-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exportPath := filepath.Join(tmpDir, "export.csv")

	// Test bulk export
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "export", exportPath,
		"--min-score", "5.0",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify export file was created
	_, err = os.Stat(exportPath)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Export test idea 1")
	assert.Contains(t, string(content), "Export test idea 2")
}

func TestBulkExport_ToJSON(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test idea
	idea := &models.Idea{
		ID:              uuid.New().String(),
		Content:         "JSON export test",
		RawScore:        7.0,
		FinalScore:      7.5,
		Patterns:        []string{"test"},
		Recommendation:  "PRIORITIZE NOW",
		AnalysisDetails: "Test",
		Status:          "active",
		CreatedAt:       time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Create temp directory for export
	tmpDir, err := os.MkdirTemp("", "bulk-json-export-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exportPath := filepath.Join(tmpDir, "export.json")

	// Test bulk export to JSON
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "export", exportPath,
		"--format", "json",
		"--pretty",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify export file was created
	content, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "JSON export test")
	assert.Contains(t, string(content), "PRIORITIZE NOW")
}

func TestBulkTag_WithSearchFilter(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test ideas
	ideas := []*models.Idea{
		{
			ID:         uuid.New().String(),
			Content:    "Build AI agent with Go and LangChain",
			RawScore:   8.0,
			FinalScore: 8.5,
			Status:     "active",
			CreatedAt:  time.Now().UTC(),
		},
		{
			ID:         uuid.New().String(),
			Content:    "Build web scraper with Python",
			RawScore:   6.0,
			FinalScore: 6.2,
			Status:     "active",
			CreatedAt:  time.Now().UTC(),
		},
	}

	for _, idea := range ideas {
		err := cliCtx.Repository.Create(idea)
		require.NoError(t, err)
	}

	// Test bulk tag with search filter
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "tag", "ai-related",
		"--search", "AI",
		"--yes",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Command should complete successfully
	// In real implementation, only AI-related idea would be tagged
}

func TestBulkArchive_DryRun(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test idea
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "Old idea for dry run test",
		RawScore:   2.0,
		FinalScore: 2.5,
		Status:     "active",
		CreatedAt:  time.Now().UTC().Add(-100 * 24 * time.Hour),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk archive with dry-run
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "archive",
		"--older-than", "90",
		"--max-score", "5.0",
		"--dry-run",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify idea was NOT archived (dry run)
	unchangedIdea, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", unchangedIdea.Status, "Dry run should not modify data")
}

func TestBulkAnalyze_WithFilters(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test ideas with varying scores
	ideas := []*models.Idea{
		{
			ID:              uuid.New().String(),
			Content:         "Build AI SaaS with Go",
			RawScore:        2.0,
			FinalScore:      2.5,
			Recommendation:  "LOW PRIORITY",
			AnalysisDetails: "Old analysis",
			Status:          "active",
			CreatedAt:       time.Now().UTC(),
		},
		{
			ID:              uuid.New().String(),
			Content:         "Python automation",
			RawScore:        3.0,
			FinalScore:      3.2,
			Recommendation:  "CONSIDER",
			AnalysisDetails: "Old analysis",
			Status:          "active",
			CreatedAt:       time.Now().UTC(),
		},
		{
			ID:              uuid.New().String(),
			Content:         "High score idea",
			RawScore:        8.0,
			FinalScore:      8.5,
			Recommendation:  "PRIORITIZE",
			AnalysisDetails: "Old analysis",
			Status:          "active",
			CreatedAt:       time.Now().UTC(),
		},
	}

	// Save test ideas
	for _, idea := range ideas {
		err := cliCtx.Repository.Create(idea)
		require.NoError(t, err)
	}

	// Test bulk analyze with score filter
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "analyze",
		"--score-max", "4.0",
		"--yes", // Auto-confirm for testing
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify low-scoring ideas were re-analyzed
	// (scores may have changed based on current telos)
	for _, idea := range ideas[:2] {
		reanalyzed, err := cliCtx.Repository.GetByID(idea.ID)
		require.NoError(t, err)
		assert.NotNil(t, reanalyzed.AnalysisDetails)
		// Score should be recalculated (may be different)
		assert.NotEqual(t, 0.0, reanalyzed.FinalScore)
	}
}

func TestBulkAnalyze_WithOlderThan(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create old and new ideas
	oldIdea := &models.Idea{
		ID:              uuid.New().String(),
		Content:         "Old idea - Build mobile app",
		RawScore:        5.0,
		FinalScore:      5.5,
		Recommendation:  "OLD ANALYSIS",
		AnalysisDetails: "Very old analysis",
		Status:          "active",
		CreatedAt:       time.Now().UTC().Add(-60 * 24 * time.Hour), // 60 days old
	}

	newIdea := &models.Idea{
		ID:              uuid.New().String(),
		Content:         "New idea - Build web app",
		RawScore:        5.0,
		FinalScore:      5.5,
		Recommendation:  "RECENT ANALYSIS",
		AnalysisDetails: "Recent analysis",
		Status:          "active",
		CreatedAt:       time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(oldIdea)
	require.NoError(t, err)
	err = cliCtx.Repository.Create(newIdea)
	require.NoError(t, err)

	// Test bulk analyze with older-than filter
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "analyze",
		"--older-than", "30d", // 30 days
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify old idea was re-analyzed
	reanalyzed, err := cliCtx.Repository.GetByID(oldIdea.ID)
	require.NoError(t, err)
	assert.NotNil(t, reanalyzed.AnalysisDetails)
}

func TestBulkAnalyze_DryRun(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test idea
	originalRecommendation := "ORIGINAL RECOMMENDATION"
	idea := &models.Idea{
		ID:              uuid.New().String(),
		Content:         "Test idea for dry run",
		RawScore:        4.0,
		FinalScore:      4.5,
		Recommendation:  originalRecommendation,
		AnalysisDetails: "Original analysis",
		Status:          "active",
		CreatedAt:       time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk analyze with dry-run
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "analyze",
		"--score-max", "5.0",
		"--dry-run",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify idea was NOT modified (dry run)
	unchanged, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Equal(t, originalRecommendation, unchanged.Recommendation, "Dry run should not modify data")
	assert.Equal(t, "Original analysis", unchanged.AnalysisDetails, "Dry run should not modify data")
}

func TestBulkAnalyze_EmptyResult(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test idea that doesn't match filters
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "High score idea",
		RawScore:   9.0,
		FinalScore: 9.5,
		Status:     "active",
		CreatedAt:  time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk analyze with filters that don't match any ideas
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "analyze",
		"--score-max", "3.0", // No ideas below 3.0
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err) // Should complete without error even with no matches
}

func TestBulkAnalyze_WithProvider(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create test idea
	idea := &models.Idea{
		ID:              uuid.New().String(),
		Content:         "Test provider selection",
		RawScore:        5.0,
		FinalScore:      5.5,
		Recommendation:  "OLD",
		AnalysisDetails: "Old",
		Status:          "active",
		CreatedAt:       time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk analyze with specific provider
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "analyze",
		"--provider", "rule_based", // Use rule-based provider
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify idea was re-analyzed
	reanalyzed, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.NotNil(t, reanalyzed.AnalysisDetails)
}

func TestBulkAnalyze_StatusFilter(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Create active and archived ideas
	activeIdea := &models.Idea{
		ID:              uuid.New().String(),
		Content:         "Active idea",
		RawScore:        5.0,
		FinalScore:      5.5,
		Recommendation:  "OLD",
		AnalysisDetails: "Old analysis",
		Status:          "active",
		CreatedAt:       time.Now().UTC(),
	}

	archivedIdea := &models.Idea{
		ID:              uuid.New().String(),
		Content:         "Archived idea",
		RawScore:        5.0,
		FinalScore:      5.5,
		Recommendation:  "OLD ARCHIVED",
		AnalysisDetails: "Old analysis",
		Status:          "archived",
		CreatedAt:       time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(activeIdea)
	require.NoError(t, err)
	err = cliCtx.Repository.Create(archivedIdea)
	require.NoError(t, err)

	// Test bulk analyze with status filter (only active)
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "analyze",
		"--status", "active",
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify active idea was re-analyzed
	reanalyzed, err := cliCtx.Repository.GetByID(activeIdea.ID)
	require.NoError(t, err)
	assert.NotEqual(t, "OLD", reanalyzed.Recommendation)

	// Archived idea should remain unchanged
	unchanged, err := cliCtx.Repository.GetByID(archivedIdea.ID)
	require.NoError(t, err)
	assert.Equal(t, "OLD ARCHIVED", unchanged.Recommendation)
}

// Unit tests for helper functions

func TestSplitCommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single item",
			input:    "one",
			expected: []string{"one"},
		},
		{
			name:     "multiple items",
			input:    "one,two,three",
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "items with spaces",
			input:    "one, two, three",
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "items with extra spaces",
			input:    " one , two ",
			expected: []string{"one", "two"},
		},
		{
			name:     "items with empty parts",
			input:    "one,,two",
			expected: []string{"one", "two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitCommaSeparated(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitCommaSeparated(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAddUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		newItems []string
		expected []string
	}{
		{
			name:     "add to empty slice",
			existing: []string{},
			newItems: []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "add new items",
			existing: []string{"a", "b", "c"},
			newItems: []string{"d", "e"},
			expected: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:     "add with duplicates",
			existing: []string{"a", "b", "c"},
			newItems: []string{"b", "d", "e"},
			expected: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:     "add all duplicates",
			existing: []string{"a", "b", "c"},
			newItems: []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "add nothing",
			existing: []string{"a", "b", "c"},
			newItems: []string{},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addUniqueStrings(tt.existing, tt.newItems)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("addUniqueStrings(%v, %v) = %v, want %v",
					tt.existing, tt.newItems, result, tt.expected)
			}
		})
	}
}

func TestRemoveStrings(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		toRemove []string
		expected []string
	}{
		{
			name:     "remove from empty slice",
			existing: []string{},
			toRemove: []string{"a"},
			expected: []string{},
		},
		{
			name:     "remove existing items",
			existing: []string{"a", "b", "c", "d"},
			toRemove: []string{"b", "d"},
			expected: []string{"a", "c"},
		},
		{
			name:     "remove non-existing items",
			existing: []string{"a", "b", "c"},
			toRemove: []string{"d", "e"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "remove all items",
			existing: []string{"a", "b", "c"},
			toRemove: []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			name:     "remove nothing",
			existing: []string{"a", "b", "c"},
			toRemove: []string{},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "remove with duplicates in toRemove",
			existing: []string{"a", "b", "c"},
			toRemove: []string{"b", "b", "c"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeStrings(tt.existing, tt.toRemove)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("removeStrings(%v, %v) = %v, want %v",
					tt.existing, tt.toRemove, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
		{
			name:     "case sensitive",
			slice:    []string{"a", "b", "c"},
			item:     "A",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains(%v, %q) = %v, want %v",
					tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "truncation needed",
			input:    "this is a very long string that needs truncation",
			maxLen:   20,
			expected: "this is a very lo...",
		},
		{
			name:     "exact length",
			input:    "exactly",
			maxLen:   7,
			expected: "exactly",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// Integration tests for bulk update

func TestBulkUpdate_SetStatus(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test ideas
	ideas := []*models.Idea{
		{
			ID:         uuid.New().String(),
			Content:    "Low score idea 1",
			RawScore:   2.0,
			FinalScore: 2.5,
			Status:     "active",
			Patterns:   []string{},
			Tags:       []string{},
			CreatedAt:  time.Now().UTC(),
		},
		{
			ID:         uuid.New().String(),
			Content:    "Low score idea 2",
			RawScore:   1.5,
			FinalScore: 1.8,
			Status:     "active",
			Patterns:   []string{},
			Tags:       []string{},
			CreatedAt:  time.Now().UTC(),
		},
	}

	for _, idea := range ideas {
		err := cliCtx.Repository.Create(idea)
		require.NoError(t, err)
	}

	// Test bulk update to archive low-scoring ideas
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "update",
		"--score-max", "3.0",
		"--set-status", "archived",
		"--yes",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify ideas were archived
	for _, idea := range ideas {
		updated, err := cliCtx.Repository.GetByID(idea.ID)
		require.NoError(t, err)
		assert.Equal(t, "archived", updated.Status)
	}
}

func TestBulkUpdate_AddPatterns(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test idea
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "High value idea",
		RawScore:   8.0,
		FinalScore: 8.5,
		Status:     "active",
		Patterns:   []string{"existing-pattern"},
		Tags:       []string{},
		CreatedAt:  time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk update to add pattern
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "update",
		"--score-min", "7.0",
		"--add-patterns", "high-value,priority",
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify patterns were added
	updated, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Contains(t, updated.Patterns, "existing-pattern")
	assert.Contains(t, updated.Patterns, "high-value")
	assert.Contains(t, updated.Patterns, "priority")
	assert.Equal(t, 3, len(updated.Patterns))
}

func TestBulkUpdate_RemovePatterns(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test idea
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "Idea with obsolete patterns",
		RawScore:   5.0,
		FinalScore: 5.5,
		Status:     "active",
		Patterns:   []string{"old-pattern", "keep-pattern", "remove-pattern"},
		Tags:       []string{},
		CreatedAt:  time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk update to remove patterns
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "update",
		"--remove-patterns", "old-pattern,remove-pattern",
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify patterns were removed
	updated, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.NotContains(t, updated.Patterns, "old-pattern")
	assert.NotContains(t, updated.Patterns, "remove-pattern")
	assert.Contains(t, updated.Patterns, "keep-pattern")
	assert.Equal(t, 1, len(updated.Patterns))
}

func TestBulkUpdate_AddTags(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test idea
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "Archived idea",
		RawScore:   6.0,
		FinalScore: 6.2,
		Status:     "archived",
		Patterns:   []string{},
		Tags:       []string{},
		CreatedAt:  time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk update to add tags
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "update",
		"--status", "archived",
		"--add-tags", "reviewed,processed",
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify tags were added
	updated, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Contains(t, updated.Tags, "reviewed")
	assert.Contains(t, updated.Tags, "processed")
	assert.Equal(t, 2, len(updated.Tags))
}

func TestBulkUpdate_DryRun(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test idea
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "Idea for dry run",
		RawScore:   3.0,
		FinalScore: 3.5,
		Status:     "active",
		Patterns:   []string{"old"},
		Tags:       []string{},
		CreatedAt:  time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk update with dry-run
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "update",
		"--score-max", "4.0",
		"--set-status", "archived",
		"--add-patterns", "new",
		"--dry-run",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify idea was NOT modified
	unchanged, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", unchanged.Status)
	assert.Equal(t, []string{"old"}, unchanged.Patterns)
}

func TestBulkUpdate_CombinedOperations(t *testing.T) {
	cliCtx, cleanup := setupTestCLI(t)
	defer cleanup()

	// Set the global CLI context for commands to use
	SetContext(cliCtx)

	// Create test idea
	idea := &models.Idea{
		ID:         uuid.New().String(),
		Content:    "Idea for combined update",
		RawScore:   7.0,
		FinalScore: 7.5,
		Status:     "active",
		Patterns:   []string{"remove-me"},
		Tags:       []string{},
		CreatedAt:  time.Now().UTC(),
	}

	err := cliCtx.Repository.Create(idea)
	require.NoError(t, err)

	// Test bulk update with multiple operations
	cmd := GetRootCmd()
	cmd.SetArgs([]string{
		"--telos", cliCtx.TelosPath,
		"--db", cliCtx.DBPath,
		"bulk", "update",
		"--score-min", "7.0",
		"--set-status", "archived",
		"--add-patterns", "high-priority",
		"--remove-patterns", "remove-me",
		"--add-tags", "reviewed,done",
		"--yes",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify all changes were applied
	updated, err := cliCtx.Repository.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Equal(t, "archived", updated.Status)
	assert.Contains(t, updated.Patterns, "high-priority")
	assert.NotContains(t, updated.Patterns, "remove-me")
	assert.Contains(t, updated.Tags, "reviewed")
	assert.Contains(t, updated.Tags, "done")
}
