//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/api"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndWorkflow tests the complete workflow from idea creation to analysis
func TestEndToEndWorkflow(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create test telos file
	telosPath := filepath.Join(tempDir, "telos.md")
	telosContent := `# Telos: Technical Excellence

## Core Goals
- Build scalable systems
- Write clean code
- Deliver value

## Strategies
- Test-driven development
- Continuous integration
- Code review process

## Anti-Patterns
- Premature optimization
- Over-engineering
- Technical debt accumulation
`
	require.NoError(t, os.WriteFile(telosPath, []byte(telosContent), 0644))

	// Parse telos
	telosConfig, err := telos.ParseTelosFile(telosPath)
	require.NoError(t, err)

	// Create database repository
	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	// Create API server
	server := api.NewServer(repo, telosConfig)
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Test 1: Create an idea via API
	t.Run("CreateIdea", func(t *testing.T) {
		createReq := api.CreateIdeaRequest{
			Content: "Build a microservices platform with extensive testing",
		}
		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var ideaResp api.IdeaResponse
		err = json.NewDecoder(resp.Body).Decode(&ideaResp)
		require.NoError(t, err)

		assert.NotEmpty(t, ideaResp.ID)
		assert.Equal(t, createReq.Content, ideaResp.Content)
		assert.Greater(t, ideaResp.FinalScore, 0.0)
		assert.Equal(t, "active", ideaResp.Status)
	})

	// Test 2: Analyze an idea without storing it
	t.Run("AnalyzeIdea", func(t *testing.T) {
		analyzeReq := api.AnalyzeRequest{
			Content: "Quick hack without tests to meet deadline",
		}
		body, err := json.Marshal(analyzeReq)
		require.NoError(t, err)

		resp, err := http.Post(ts.URL+"/api/v1/analyze", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var analyzeResp api.AnalyzeResponse
		err = json.NewDecoder(resp.Body).Decode(&analyzeResp)
		require.NoError(t, err)

		assert.NotNil(t, analyzeResp.Analysis)
		assert.Less(t, analyzeResp.Analysis.FinalScore, 5.0) // Should score low due to anti-patterns
	})

	// Test 3: List ideas
	t.Run("ListIdeas", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/ideas")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp api.ListIdeasResponse
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		assert.Greater(t, len(listResp.Ideas), 0)
		assert.Equal(t, len(listResp.Ideas), listResp.Total)
	})

	// Test 4: Get specific idea
	t.Run("GetIdea", func(t *testing.T) {
		// First create an idea
		createReq := api.CreateIdeaRequest{
			Content: "Implement comprehensive testing strategy",
		}
		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		createResp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer createResp.Body.Close()

		var idea api.IdeaResponse
		err = json.NewDecoder(createResp.Body).Decode(&idea)
		require.NoError(t, err)

		// Now get it
		getResp, err := http.Get(ts.URL + "/api/v1/ideas/" + idea.ID)
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var getIdea api.IdeaResponse
		err = json.NewDecoder(getResp.Body).Decode(&getIdea)
		require.NoError(t, err)

		assert.Equal(t, idea.ID, getIdea.ID)
		assert.Equal(t, idea.Content, getIdea.Content)
	})

	// Test 5: Update idea
	t.Run("UpdateIdea", func(t *testing.T) {
		// First create an idea
		createReq := api.CreateIdeaRequest{
			Content: "Initial idea content",
		}
		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		createResp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer createResp.Body.Close()

		var idea api.IdeaResponse
		err = json.NewDecoder(createResp.Body).Decode(&idea)
		require.NoError(t, err)

		// Update it
		newContent := "Updated with better testing approach"
		updateReq := api.UpdateIdeaRequest{
			Content: &newContent,
		}
		updateBody, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/ideas/"+idea.ID, bytes.NewBuffer(updateBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		updateResp, err := client.Do(req)
		require.NoError(t, err)
		defer updateResp.Body.Close()

		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedIdea api.IdeaResponse
		err = json.NewDecoder(updateResp.Body).Decode(&updatedIdea)
		require.NoError(t, err)

		assert.Equal(t, newContent, updatedIdea.Content)
	})

	// Test 6: Delete idea
	t.Run("DeleteIdea", func(t *testing.T) {
		// First create an idea
		createReq := api.CreateIdeaRequest{
			Content: "Idea to be deleted",
		}
		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		createResp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer createResp.Body.Close()

		var idea api.IdeaResponse
		err = json.NewDecoder(createResp.Body).Decode(&idea)
		require.NoError(t, err)

		// Delete it
		req, err := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/ideas/"+idea.ID, nil)
		require.NoError(t, err)

		client := &http.Client{}
		deleteResp, err := client.Do(req)
		require.NoError(t, err)
		defer deleteResp.Body.Close()

		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify it's gone
		getResp, err := http.Get(ts.URL + "/api/v1/ideas/" + idea.ID)
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	// Test 7: Analytics stats
	t.Run("AnalyticsStats", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/analytics/stats")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats api.StatsResponse
		err = json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)

		assert.Greater(t, stats.TotalIdeas, 0)
		assert.GreaterOrEqual(t, stats.TotalIdeas, stats.ActiveIdeas)
	})
}

// TestConcurrentAccess tests concurrent API requests
func TestConcurrentAccess(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	telosPath := filepath.Join(tempDir, "telos.md")
	telosContent := `# Telos
## Core Goals
- Goal 1
## Strategies
- Strategy 1
`
	require.NoError(t, os.WriteFile(telosPath, []byte(telosContent), 0644))

	telosConfig, err := telos.ParseTelosFile(telosPath)
	require.NoError(t, err)

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	server := api.NewServer(repo, telosConfig)
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Test concurrent creates
	t.Run("ConcurrentCreates", func(t *testing.T) {
		const numGoroutines = 10
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()

				createReq := api.CreateIdeaRequest{
					Content: fmt.Sprintf("Concurrent idea %d", n),
				}
				body, err := json.Marshal(createReq)
				if err != nil {
					errors <- err
					return
				}

				resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
					return
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("Error during concurrent creation: %v", err)
		}

		// Verify all ideas were created
		resp, err := http.Get(ts.URL + "/api/v1/ideas")
		require.NoError(t, err)
		defer resp.Body.Close()

		var listResp api.ListIdeasResponse
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		assert.Equal(t, numGoroutines, len(listResp.Ideas))
	})

	// Test concurrent reads and writes
	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
		// Create a few ideas first
		ideaIDs := make([]string, 5)
		for i := 0; i < 5; i++ {
			createReq := api.CreateIdeaRequest{
				Content: fmt.Sprintf("Idea for concurrent test %d", i),
			}
			body, err := json.Marshal(createReq)
			require.NoError(t, err)

			resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
			require.NoError(t, err)

			var idea api.IdeaResponse
			err = json.NewDecoder(resp.Body).Decode(&idea)
			resp.Body.Close()
			require.NoError(t, err)

			ideaIDs[i] = idea.ID
		}

		// Now perform concurrent reads and writes
		const numOperations = 20
		var wg sync.WaitGroup
		errors := make(chan error, numOperations)

		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()

				// Alternate between reads and writes
				if n%2 == 0 {
					// Read operation
					ideaID := ideaIDs[n%len(ideaIDs)]
					resp, err := http.Get(ts.URL + "/api/v1/ideas/" + ideaID)
					if err != nil {
						errors <- err
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						errors <- fmt.Errorf("unexpected status code on read: %d", resp.StatusCode)
						return
					}
				} else {
					// Write operation (update)
					ideaID := ideaIDs[n%len(ideaIDs)]
					newStatus := "archived"
					updateReq := api.UpdateIdeaRequest{
						Status: &newStatus,
					}
					body, err := json.Marshal(updateReq)
					if err != nil {
						errors <- err
						return
					}

					req, err := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/ideas/"+ideaID, bytes.NewBuffer(body))
					if err != nil {
						errors <- err
						return
					}
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						errors <- err
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						errors <- fmt.Errorf("unexpected status code on update: %d", resp.StatusCode)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("Error during concurrent operations: %v", err)
		}
	})
}

// TestDatabaseIntegrity tests database operations and constraints
func TestDatabaseIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	t.Run("CreateAndRetrieve", func(t *testing.T) {
		idea := &models.Idea{
			ID:             "test-id-123",
			Content:        "Test idea content",
			RawScore:       7.5,
			FinalScore:     8.0,
			Patterns:       []string{"pattern1", "pattern2"},
			Recommendation: "Proceed with caution",
			Status:         "active",
			CreatedAt:      time.Now().UTC(),
		}

		err := repo.Create(idea)
		require.NoError(t, err)

		retrieved, err := repo.GetByID("test-id-123")
		require.NoError(t, err)

		assert.Equal(t, idea.ID, retrieved.ID)
		assert.Equal(t, idea.Content, retrieved.Content)
		assert.Equal(t, idea.RawScore, retrieved.RawScore)
		assert.Equal(t, idea.FinalScore, retrieved.FinalScore)
		assert.Equal(t, idea.Status, retrieved.Status)
	})

	t.Run("UpdateIdea", func(t *testing.T) {
		idea, err := repo.GetByID("test-id-123")
		require.NoError(t, err)

		idea.Content = "Updated content"
		idea.FinalScore = 9.0
		idea.Status = "archived"

		err = repo.Update(idea)
		require.NoError(t, err)

		updated, err := repo.GetByID("test-id-123")
		require.NoError(t, err)

		assert.Equal(t, "Updated content", updated.Content)
		assert.Equal(t, 9.0, updated.FinalScore)
		assert.Equal(t, "archived", updated.Status)
	})

	t.Run("DeleteIdea", func(t *testing.T) {
		err := repo.Delete("test-id-123")
		require.NoError(t, err)

		_, err = repo.GetByID("test-id-123")
		assert.Error(t, err)
	})

	t.Run("ListWithFilters", func(t *testing.T) {
		// Create multiple ideas with different scores and statuses
		ideas := []*models.Idea{
			{
				ID:         "idea-1",
				Content:    "Idea 1",
				RawScore:   5.0,
				FinalScore: 5.5,
				Status:     "active",
				CreatedAt:  time.Now().UTC(),
			},
			{
				ID:         "idea-2",
				Content:    "Idea 2",
				RawScore:   8.0,
				FinalScore: 8.5,
				Status:     "active",
				CreatedAt:  time.Now().UTC(),
			},
			{
				ID:         "idea-3",
				Content:    "Idea 3",
				RawScore:   3.0,
				FinalScore: 3.5,
				Status:     "archived",
				CreatedAt:  time.Now().UTC(),
			},
		}

		for _, idea := range ideas {
			err := repo.Create(idea)
			require.NoError(t, err)
		}

		// Test status filter
		activeIdeas, err := repo.List(database.ListOptions{Status: "active"})
		require.NoError(t, err)
		assert.Equal(t, 2, len(activeIdeas))

		// Test score filter
		minScore := 7.0
		highScoreIdeas, err := repo.List(database.ListOptions{MinScore: &minScore})
		require.NoError(t, err)
		assert.Equal(t, 1, len(highScoreIdeas))
		assert.Equal(t, "idea-2", highScoreIdeas[0].ID)

		// Test limit
		limit := 2
		limitedIdeas, err := repo.List(database.ListOptions{Limit: &limit})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(limitedIdeas), 2)
	})
}

// TestAPIErrorHandling tests error scenarios
func TestAPIErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	telosPath := filepath.Join(tempDir, "telos.md")
	telosContent := `# Telos
## Core Goals
- Goal 1
`
	require.NoError(t, os.WriteFile(telosPath, []byte(telosContent), 0644))

	telosConfig, err := telos.ParseTelosFile(telosPath)
	require.NoError(t, err)

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	server := api.NewServer(repo, telosConfig)
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	t.Run("InvalidJSON", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBufferString("invalid json"))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("EmptyContent", func(t *testing.T) {
		createReq := api.CreateIdeaRequest{
			Content: "",
		}
		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NonExistentIdea", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/ideas/non-existent-id")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Invalid UUID format
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/ideas/invalid-uuid-format")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestHealthCheck tests the health endpoint
func TestHealthCheck(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	telosPath := filepath.Join(tempDir, "telos.md")
	require.NoError(t, os.WriteFile(telosPath, []byte("# Telos\n## Core Goals\n- Goal 1"), 0644))

	telosConfig, err := telos.ParseTelosFile(telosPath)
	require.NoError(t, err)

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	server := api.NewServer(repo, telosConfig)
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Contains(t, string(body), "ok")
}
