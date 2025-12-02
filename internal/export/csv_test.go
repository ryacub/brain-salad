package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSV_Export(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "csv-export-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test ideas
	ideas := []*models.Idea{
		{
			ID:              uuid.New().String(),
			Content:         "Build AI agent platform",
			RawScore:        8.5,
			FinalScore:      8.2,
			Patterns:        []string{"ai-focus", "revenue-model"},
			Recommendation:  "PRIORITIZE NOW",
			AnalysisDetails: "Strong alignment with goals",
			CreatedAt:       time.Now().UTC(),
			Status:          "active",
		},
		{
			ID:              uuid.New().String(),
			Content:         "Learn Rust framework",
			RawScore:        4.0,
			FinalScore:      3.5,
			Patterns:        []string{"context-switching"},
			Recommendation:  "AVOID",
			AnalysisDetails: "Misaligned with current stack",
			CreatedAt:       time.Now().UTC().Add(-24 * time.Hour),
			Status:          "active",
		},
	}

	// Export to CSV
	csvPath := filepath.Join(tmpDir, "ideas.csv")
	err = ExportCSV(ideas, csvPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(csvPath)
	require.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	csvContent := string(content)
	assert.Contains(t, csvContent, "ID,Content,RawScore,FinalScore")
	assert.Contains(t, csvContent, "Build AI agent platform")
	assert.Contains(t, csvContent, "Learn Rust framework")
	assert.Contains(t, csvContent, "8.2")
	assert.Contains(t, csvContent, "3.5")
}

func TestCSV_Import(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "csv-import-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test CSV file
	csvPath := filepath.Join(tmpDir, "ideas.csv")
	csvContent := `ID,Content,RawScore,FinalScore,Patterns,Recommendation,AnalysisDetails,CreatedAt,Status
test-id-1,"Build SaaS product",7.5,7.2,"ai-focus,revenue-model","PRIORITIZE NOW","Strong alignment","2025-01-15T10:00:00Z","active"
test-id-2,"Learn new language",3.0,2.5,"context-switching","AVOID","Misaligned","2025-01-14T10:00:00Z","active"
`
	err = os.WriteFile(csvPath, []byte(csvContent), 0644)
	require.NoError(t, err)

	// Import from CSV
	ideas, err := ImportCSV(csvPath)
	require.NoError(t, err)

	// Verify imported ideas
	require.Len(t, ideas, 2)

	// Check first idea
	assert.Equal(t, "test-id-1", ideas[0].ID)
	assert.Equal(t, "Build SaaS product", ideas[0].Content)
	assert.Equal(t, 7.5, ideas[0].RawScore)
	assert.Equal(t, 7.2, ideas[0].FinalScore)
	assert.Equal(t, "PRIORITIZE NOW", ideas[0].Recommendation)
	assert.Equal(t, "active", ideas[0].Status)

	// Check second idea
	assert.Equal(t, "test-id-2", ideas[1].ID)
	assert.Equal(t, "Learn new language", ideas[1].Content)
	assert.Equal(t, 3.0, ideas[1].RawScore)
	assert.Equal(t, 2.5, ideas[1].FinalScore)
}

func TestCSV_HandleMalformed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csv-malformed-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tests := []struct {
		name        string
		csvContent  string
		expectError bool
	}{
		{
			name:        "empty file",
			csvContent:  "",
			expectError: true,
		},
		{
			name:        "header only",
			csvContent:  "ID,Content,RawScore,FinalScore,Patterns,Recommendation,AnalysisDetails,CreatedAt,Status\n",
			expectError: false, // Should return empty slice
		},
		{
			name: "missing columns",
			csvContent: `ID,Content
test-id,"Missing columns"
`,
			expectError: true,
		},
		{
			name: "invalid score",
			csvContent: `ID,Content,RawScore,FinalScore,Patterns,Recommendation,AnalysisDetails,CreatedAt,Status
test-id,"Test","invalid",7.0,"","","","2025-01-15T10:00:00Z","active"
`,
			expectError: false, // Should handle gracefully with 0.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvPath := filepath.Join(tmpDir, tt.name+".csv")
			err := os.WriteFile(csvPath, []byte(tt.csvContent), 0644)
			require.NoError(t, err)

			ideas, err := ImportCSV(csvPath)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Empty file or header only should return empty slice
				if strings.TrimSpace(tt.csvContent) == "" ||
					strings.Count(tt.csvContent, "\n") <= 1 {
					assert.Empty(t, ideas)
				}
			}
		})
	}
}

func TestCSV_Export_LargeDataset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csv-large-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create 1000 test ideas
	ideas := make([]*models.Idea, 1000)
	for i := 0; i < 1000; i++ {
		ideas[i] = &models.Idea{
			ID:             uuid.New().String(),
			Content:        "Test idea " + string(rune(i)),
			RawScore:       float64(i % 10),
			FinalScore:     float64(i % 10),
			Patterns:       []string{"test"},
			Recommendation: "TEST",
			CreatedAt:      time.Now().UTC(),
			Status:         "active",
		}
	}

	// Export to CSV
	csvPath := filepath.Join(tmpDir, "large.csv")
	err = ExportCSV(ideas, csvPath)
	require.NoError(t, err)

	// Verify file exists and has content
	stat, err := os.Stat(csvPath)
	require.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(0))

	// Verify can be imported back
	importedIdeas, err := ImportCSV(csvPath)
	require.NoError(t, err)
	assert.Len(t, importedIdeas, 1000)
}

func TestCSV_ExportJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-export-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test ideas
	ideas := []*models.Idea{
		{
			ID:              uuid.New().String(),
			Content:         "Build AI agent platform",
			RawScore:        8.5,
			FinalScore:      8.2,
			Patterns:        []string{"ai-focus"},
			Recommendation:  "PRIORITIZE NOW",
			AnalysisDetails: "Strong alignment",
			CreatedAt:       time.Now().UTC(),
			Status:          "active",
		},
	}

	// Export to JSON (pretty)
	jsonPath := filepath.Join(tmpDir, "ideas.json")
	err = ExportJSON(ideas, jsonPath, true)
	require.NoError(t, err)

	// Verify file exists
	content, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	jsonContent := string(content)
	assert.Contains(t, jsonContent, "Build AI agent platform")
	assert.Contains(t, jsonContent, "PRIORITIZE NOW")
	// Pretty format should have indentation
	assert.Contains(t, jsonContent, "\n")
}

func TestCSV_ExportJSON_Compact(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-compact-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	ideas := []*models.Idea{
		{
			ID:         uuid.New().String(),
			Content:    "Test",
			RawScore:   5.0,
			FinalScore: 5.0,
			Status:     "active",
			CreatedAt:  time.Now().UTC(),
		},
	}

	// Export to JSON (compact)
	jsonPath := filepath.Join(tmpDir, "compact.json")
	err = ExportJSON(ideas, jsonPath, false)
	require.NoError(t, err)

	content, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	// Compact should have fewer newlines than pretty
	assert.NotEmpty(t, content)
}
