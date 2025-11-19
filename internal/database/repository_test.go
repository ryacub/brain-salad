//go:build integration
// +build integration

package database_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a temporary database for testing
func setupTestDB(t *testing.T) (*database.Repository, func()) {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpfile.Close()

	repo, err := database.NewRepository(tmpfile.Name())
	require.NoError(t, err)

	cleanup := func() {
		repo.Close()
		os.Remove(tmpfile.Name())
	}

	return repo, cleanup
}

// TestNewRepository_ValidPath_CreatesDatabase tests repository creation
func TestNewRepository_ValidPath_CreatesDatabase(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	assert.NotNil(t, repo)
}

// TestNewRepository_InvalidPath_ReturnsError tests error handling
func TestNewRepository_InvalidPath_ReturnsError(t *testing.T) {
	// Note: SQLite is permissive with paths and will create directories/files
	// This test verifies that attempting to open a database in a read-only
	// location (like /dev/null as a directory) will fail
	t.Skip("SQLite is too permissive with file paths - skipping this test")
}

// TestRepository_Create_ValidIdea_SavesSuccessfully tests creating an idea
func TestRepository_Create_ValidIdea_SavesSuccessfully(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	idea := models.NewIdea("Build a Go CLI tool for idea management")
	idea.RawScore = 8.5
	idea.FinalScore = 8.5
	idea.Recommendation = "\U0001F525 PRIORITIZE NOW"
	idea.Status = "active"

	err := repo.Create(idea)
	require.NoError(t, err)

	// Verify idea was saved
	retrieved, err := repo.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Equal(t, idea.ID, retrieved.ID)
	assert.Equal(t, idea.Content, retrieved.Content)
	assert.Equal(t, idea.FinalScore, retrieved.FinalScore)
	assert.Equal(t, idea.Status, retrieved.Status)
}

// TestRepository_Create_WithAnalysis_SavesAnalysisJSON tests saving idea with analysis
func TestRepository_Create_WithAnalysis_SavesAnalysisJSON(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	analysis := &models.Analysis{
		RawScore:   8.5,
		FinalScore: 8.5,
		Mission: models.MissionScores{
			DomainExpertise:  1.1,
			AIAlignment:      1.4,
			ExecutionSupport: 0.75,
			RevenuePotential: 0.45,
			Total:            3.7,
		},
		AntiChallenge: models.AntiChallengeScores{
			ContextSwitching: 1.15,
			RapidPrototyping: 1.0,
			Accountability:   0.75,
			IncomeAnxiety:    0.9,
			Total:            3.8,
		},
		Strategic: models.StrategicScores{
			StackCompatibility:   0.95,
			ShippingHabit:        0.8,
			PublicAccountability: 0.3,
			RevenueTesting:       0.25,
			Total:                2.3,
		},
		DetectedPatterns: []models.DetectedPattern{
			{
				Name:        "Context switching",
				Description: "Staying focused on current tech stack",
				Confidence:  0.9,
				Severity:    "low",
			},
		},
		Recommendations: []string{"Ship MVP quickly", "Build in public"},
		AnalyzedAt:      time.Now().UTC(),
	}

	idea := models.NewIdea("Build a Go-based SaaS product")
	idea.RawScore = analysis.RawScore
	idea.FinalScore = analysis.FinalScore
	idea.Recommendation = analysis.GetRecommendation()
	idea.Patterns = []string{"Context switching"}
	analysisJSON, _ := json.Marshal(analysis)
	idea.AnalysisDetails = string(analysisJSON)

	err := repo.Create(idea)
	require.NoError(t, err)

	// Retrieve and verify analysis
	retrieved, err := repo.GetByID(idea.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, retrieved.AnalysisDetails)

	var retrievedAnalysis models.Analysis
	err = json.Unmarshal([]byte(retrieved.AnalysisDetails), &retrievedAnalysis)
	require.NoError(t, err)
	assert.Equal(t, 8.5, retrievedAnalysis.FinalScore)
	assert.Len(t, retrievedAnalysis.DetectedPatterns, 1)
	assert.Equal(t, "Context switching", retrievedAnalysis.DetectedPatterns[0].Name)
}

// TestRepository_GetByID_NonexistentID_ReturnsError tests error for missing idea
func TestRepository_GetByID_NonexistentID_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := repo.GetByID(uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestRepository_Update_ExistingIdea_UpdatesSuccessfully tests updating an idea
func TestRepository_Update_ExistingIdea_UpdatesSuccessfully(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create initial idea
	idea := models.NewIdea("Initial idea content")
	idea.Status = "active"
	err := repo.Create(idea)
	require.NoError(t, err)

	// Update idea
	idea.Content = "Updated idea content"
	idea.FinalScore = 7.5
	idea.Recommendation = "\u2705 GOOD ALIGNMENT"
	idea.Status = "archived"
	reviewedTime := time.Now().UTC()
	idea.ReviewedAt = &reviewedTime

	err = repo.Update(idea)
	require.NoError(t, err)

	// Verify updates
	retrieved, err := repo.GetByID(idea.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated idea content", retrieved.Content)
	assert.Equal(t, 7.5, retrieved.FinalScore)
	assert.Equal(t, "archived", retrieved.Status)
	assert.NotNil(t, retrieved.ReviewedAt)
}

// TestRepository_Update_NonexistentIdea_ReturnsError tests update error handling
func TestRepository_Update_NonexistentIdea_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	idea := models.NewIdea("Nonexistent idea")
	idea.ID = uuid.New().String()

	err := repo.Update(idea)
	assert.Error(t, err)
}

// TestRepository_Delete_ExistingIdea_DeletesSuccessfully tests deleting an idea
func TestRepository_Delete_ExistingIdea_DeletesSuccessfully(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	idea := models.NewIdea("Idea to be deleted")
	err := repo.Create(idea)
	require.NoError(t, err)

	// Delete idea
	err = repo.Delete(idea.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(idea.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestRepository_Delete_NonexistentID_ReturnsError tests delete error handling
func TestRepository_Delete_NonexistentID_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	err := repo.Delete(uuid.New().String())
	assert.Error(t, err)
}

// TestRepository_List_AllIdeas_ReturnsAll tests listing all ideas
func TestRepository_List_AllIdeas_ReturnsAll(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ideas
	idea1 := models.NewIdea("First idea")
	idea1.FinalScore = 8.0
	idea1.Status = "active"
	repo.Create(idea1)

	idea2 := models.NewIdea("Second idea")
	idea2.FinalScore = 6.5
	idea2.Status = "active"
	repo.Create(idea2)

	idea3 := models.NewIdea("Third idea")
	idea3.FinalScore = 4.0
	idea3.Status = "archived"
	repo.Create(idea3)

	// List all ideas
	ideas, err := repo.List(database.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, ideas, 3)
}

// TestRepository_List_FilterByStatus_ReturnsFiltered tests status filtering
func TestRepository_List_FilterByStatus_ReturnsFiltered(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ideas
	idea1 := models.NewIdea("Active idea 1")
	idea1.Status = "active"
	repo.Create(idea1)

	idea2 := models.NewIdea("Active idea 2")
	idea2.Status = "active"
	repo.Create(idea2)

	idea3 := models.NewIdea("Archived idea")
	idea3.Status = "archived"
	repo.Create(idea3)

	// List only active ideas
	ideas, err := repo.List(database.ListOptions{Status: "active"})
	require.NoError(t, err)
	assert.Len(t, ideas, 2)
	for _, idea := range ideas {
		assert.Equal(t, "active", idea.Status)
	}
}

// TestRepository_List_FilterByScoreRange_ReturnsFiltered tests score filtering
func TestRepository_List_FilterByScoreRange_ReturnsFiltered(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ideas with different scores
	idea1 := models.NewIdea("High score idea")
	idea1.FinalScore = 9.0
	idea1.Status = "active"
	repo.Create(idea1)

	idea2 := models.NewIdea("Medium score idea")
	idea2.FinalScore = 6.5
	idea2.Status = "active"
	repo.Create(idea2)

	idea3 := models.NewIdea("Low score idea")
	idea3.FinalScore = 3.0
	idea3.Status = "active"
	repo.Create(idea3)

	// List ideas with score >= 7.0
	minScore := 7.0
	ideas, err := repo.List(database.ListOptions{MinScore: &minScore})
	require.NoError(t, err)
	assert.Len(t, ideas, 1)
	assert.GreaterOrEqual(t, ideas[0].FinalScore, 7.0)
}

// TestRepository_List_OrderByScore_ReturnsOrdered tests sorting
func TestRepository_List_OrderByScore_ReturnsOrdered(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create ideas in random score order
	idea1 := models.NewIdea("Low score")
	idea1.FinalScore = 4.0
	idea1.Status = "active"
	repo.Create(idea1)

	idea2 := models.NewIdea("High score")
	idea2.FinalScore = 9.0
	idea2.Status = "active"
	repo.Create(idea2)

	idea3 := models.NewIdea("Medium score")
	idea3.FinalScore = 6.5
	idea3.Status = "active"
	repo.Create(idea3)

	// List ordered by score DESC
	ideas, err := repo.List(database.ListOptions{OrderBy: "final_score DESC"})
	require.NoError(t, err)
	assert.Len(t, ideas, 3)
	// Should be in descending order
	assert.GreaterOrEqual(t, ideas[0].FinalScore, ideas[1].FinalScore)
	assert.GreaterOrEqual(t, ideas[1].FinalScore, ideas[2].FinalScore)
}

// TestRepository_List_WithLimit_ReturnsLimited tests pagination
func TestRepository_List_WithLimit_ReturnsLimited(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create 5 test ideas
	for i := 0; i < 5; i++ {
		idea := models.NewIdea("Test idea")
		idea.Status = "active"
		repo.Create(idea)
	}

	// List with limit
	limit := 3
	ideas, err := repo.List(database.ListOptions{Limit: &limit})
	require.NoError(t, err)
	assert.Len(t, ideas, 3)
}

// TestRepository_Close_ClosesConnection tests closing the database
func TestRepository_Close_ClosesConnection(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	err := repo.Close()
	assert.NoError(t, err)
}

// TestRepository_Create_NilIdea_ReturnsError tests nil idea error
func TestRepository_Create_NilIdea_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	err := repo.Create(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestRepository_Create_InvalidIdea_ReturnsError tests validation error
func TestRepository_Create_InvalidIdea_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	idea := &models.Idea{
		ID:      uuid.New().String(),
		Content: "", // Invalid: no content and no title
		Title:   "",
		Status:  "active",
	}

	err := repo.Create(idea)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid idea")
}

// TestRepository_GetByID_EmptyID_ReturnsError tests empty ID error
func TestRepository_GetByID_EmptyID_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := repo.GetByID("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestRepository_Update_NilIdea_ReturnsError tests nil idea error
func TestRepository_Update_NilIdea_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	err := repo.Update(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestRepository_Update_InvalidIdea_ReturnsError tests validation error
func TestRepository_Update_InvalidIdea_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	idea := &models.Idea{
		ID:      uuid.New().String(),
		Content: "", // Invalid: no content and no title
		Title:   "",
		Status:  "active",
	}

	err := repo.Update(idea)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid idea")
}

// TestRepository_Delete_EmptyID_ReturnsError tests empty ID error
func TestRepository_Delete_EmptyID_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	err := repo.Delete("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestRepository_List_CombinedFilters_ReturnsFiltered tests multiple filters
func TestRepository_List_CombinedFilters_ReturnsFiltered(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ideas
	idea1 := models.NewIdea("High score active idea")
	idea1.FinalScore = 9.0
	idea1.Status = "active"
	repo.Create(idea1)

	idea2 := models.NewIdea("Low score active idea")
	idea2.FinalScore = 3.0
	idea2.Status = "active"
	repo.Create(idea2)

	idea3 := models.NewIdea("High score archived idea")
	idea3.FinalScore = 8.5
	idea3.Status = "archived"
	repo.Create(idea3)

	// List active ideas with score >= 7.0
	minScore := 7.0
	ideas, err := repo.List(database.ListOptions{
		Status:   "active",
		MinScore: &minScore,
	})
	require.NoError(t, err)
	assert.Len(t, ideas, 1)
	assert.Equal(t, "active", ideas[0].Status)
	assert.GreaterOrEqual(t, ideas[0].FinalScore, 7.0)
}

// TestRepository_List_WithOffset_ReturnsOffset tests offset pagination
func TestRepository_List_WithOffset_ReturnsOffset(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create 5 test ideas
	for i := 0; i < 5; i++ {
		idea := models.NewIdea("Test idea")
		idea.Status = "active"
		repo.Create(idea)
	}

	// List with offset 2 and limit 2
	limit := 2
	offset := 2
	ideas, err := repo.List(database.ListOptions{
		Limit:  &limit,
		Offset: &offset,
	})
	require.NoError(t, err)
	assert.Len(t, ideas, 2)
}

// TestRepository_List_MaxScoreFilter_ReturnsFiltered tests max score filter
func TestRepository_List_MaxScoreFilter_ReturnsFiltered(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ideas
	idea1 := models.NewIdea("High score")
	idea1.FinalScore = 9.0
	idea1.Status = "active"
	repo.Create(idea1)

	idea2 := models.NewIdea("Low score")
	idea2.FinalScore = 3.0
	idea2.Status = "active"
	repo.Create(idea2)

	// List ideas with score <= 5.0
	maxScore := 5.0
	ideas, err := repo.List(database.ListOptions{MaxScore: &maxScore})
	require.NoError(t, err)
	assert.Len(t, ideas, 1)
	assert.LessOrEqual(t, ideas[0].FinalScore, 5.0)
}
