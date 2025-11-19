package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test server with an in-memory database
func setupTestServer(t *testing.T) (*Server, *database.Repository, func()) {
	t.Helper()

	// Create temp directory for test database
	tempDir, err := os.MkdirTemp("", "api-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)

	// Create test telos file
	telosPath := filepath.Join(tempDir, "test_telos.md")
	telosContent := `# Telos

## Goals
- G1: Build AI-powered developer tools (Deadline: 2025-12-31)
- G2: Launch SaaS product (Deadline: 2025-06-30)

## Strategies
- S1: Ship fast, iterate based on feedback
- S2: Build in public

## Stack
- Primary: Go, Svelte, TypeScript
- Secondary: Python, React

## Failure Patterns
- Context Switching: Jumping between too many technologies
- Perfectionism: Over-engineering before validation
`
	err = os.WriteFile(telosPath, []byte(telosContent), 0644)
	require.NoError(t, err)

	// Create server
	server, err := NewServerFromPath(repo, telosPath)
	require.NoError(t, err)

	cleanup := func() {
		if err := repo.Close(); err != nil {
			t.Logf("failed to close repository: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp directory: %v", err)
		}
	}

	return server, repo, cleanup
}

// Test Health Check
func TestHealthHandler(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response["status"])
	assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, response["status"])
}

// Test Analyze Endpoint
func TestAnalyzeHandler(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "valid idea analysis",
			body: `{"content":"Build a Go-based AI code review tool that ships in 2 weeks"}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response AnalyzeResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Analysis)
				assert.Greater(t, response.Analysis.FinalScore, 0.0)
			},
		},
		{
			name:           "empty content",
			body:           `{"content":""}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error, "content")
			},
		},
		{
			name:           "invalid json",
			body:           `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// Test Create Idea Endpoint
func TestCreateIdeaHandler(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "valid idea creation",
			body: `{"content":"Build AI-powered Go code reviewer"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response IdeaResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "Build AI-powered Go code reviewer", response.Content)
				assert.Equal(t, "active", response.Status)
				assert.NotNil(t, response.Analysis)
			},
		},
		{
			name:           "empty content",
			body:           `{"content":""}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
		{
			name:           "invalid json",
			body:           `{not valid json}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/ideas", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// Test Get Idea Endpoint
func TestGetIdeaHandler(t *testing.T) {
	server, repo, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test idea
	idea := &models.Idea{
		ID:      uuid.New().String(),
		Content: "Test idea",
		Status:  "active",
	}
	err := repo.Create(idea)
	require.NoError(t, err)

	tests := []struct {
		name           string
		ideaID         string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "existing idea",
			ideaID:         idea.ID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response IdeaResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, idea.ID, response.ID)
				assert.Equal(t, "Test idea", response.Content)
			},
		},
		{
			name:           "non-existent idea",
			ideaID:         uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
		{
			name:           "invalid uuid",
			ideaID:         "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/ideas/"+tt.ideaID, nil)
			w := httptest.NewRecorder()

			server.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// Test List Ideas Endpoint
func TestListIdeasHandler(t *testing.T) {
	server, repo, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test ideas
	ideas := []*models.Idea{
		{ID: uuid.New().String(), Content: "Idea 1", Status: "active", FinalScore: 8.5},
		{ID: uuid.New().String(), Content: "Idea 2", Status: "active", FinalScore: 6.0},
		{ID: uuid.New().String(), Content: "Idea 3", Status: "archived", FinalScore: 7.0},
	}
	for _, idea := range ideas {
		err := repo.Create(idea)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "list all ideas",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response ListIdeasResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, 3, response.Total)
				assert.Len(t, response.Ideas, 3)
			},
		},
		{
			name:           "filter by status",
			queryParams:    "?status=active",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response ListIdeasResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, 2, response.Total)
			},
		},
		{
			name:           "pagination",
			queryParams:    "?limit=2&offset=1",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response ListIdeasResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Len(t, response.Ideas, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/ideas"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			server.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// Test Update Idea Endpoint
func TestUpdateIdeaHandler(t *testing.T) {
	server, repo, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test idea
	idea := &models.Idea{
		ID:      uuid.New().String(),
		Content: "Original content",
		Status:  "active",
	}
	err := repo.Create(idea)
	require.NoError(t, err)

	tests := []struct {
		name           string
		ideaID         string
		body           string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "update content",
			ideaID:         idea.ID,
			body:           `{"content":"Updated content"}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response IdeaResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "Updated content", response.Content)
			},
		},
		{
			name:           "update status",
			ideaID:         idea.ID,
			body:           `{"status":"archived"}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response IdeaResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "archived", response.Status)
			},
		},
		{
			name:           "non-existent idea",
			ideaID:         uuid.New().String(),
			body:           `{"content":"New content"}`,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
		{
			name:           "invalid UUID",
			ideaID:         "not-a-uuid",
			body:           `{"content":"New content"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
		{
			name:           "invalid json",
			ideaID:         idea.ID,
			body:           `{not valid}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/api/v1/ideas/"+tt.ideaID, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// Test Delete Idea Endpoint
func TestDeleteIdeaHandler(t *testing.T) {
	server, repo, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test idea
	idea := &models.Idea{
		ID:      uuid.New().String(),
		Content: "To be deleted",
		Status:  "active",
	}
	err := repo.Create(idea)
	require.NoError(t, err)

	tests := []struct {
		name           string
		ideaID         string
		expectedStatus int
	}{
		{
			name:           "delete existing idea",
			ideaID:         idea.ID,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existent idea",
			ideaID:         uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "delete with invalid UUID",
			ideaID:         "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/v1/ideas/"+tt.ideaID, nil)
			w := httptest.NewRecorder()

			server.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// Test Analytics Stats Endpoint
func TestAnalyticsStatsHandler(t *testing.T) {
	server, repo, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test ideas with various scores
	ideas := []*models.Idea{
		{ID: uuid.New().String(), Content: "Idea 1", Status: "active", FinalScore: 8.5},
		{ID: uuid.New().String(), Content: "Idea 2", Status: "active", FinalScore: 6.0},
		{ID: uuid.New().String(), Content: "Idea 3", Status: "active", FinalScore: 7.5},
		{ID: uuid.New().String(), Content: "Idea 4", Status: "archived", FinalScore: 5.0},
	}
	for _, idea := range ideas {
		err := repo.Create(idea)
		require.NoError(t, err)
	}

	req := httptest.NewRequest("GET", "/api/v1/analytics/stats", nil)
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response StatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 4, response.TotalIdeas)
	assert.Equal(t, 3, response.ActiveIdeas)
	assert.Greater(t, response.AverageScore, 0.0)
}
