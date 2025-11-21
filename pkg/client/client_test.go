package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	client := NewClientWithTimeout("http://localhost:8080", timeout)
	assert.NotNil(t, client)
	assert.Equal(t, timeout, client.httpClient.Timeout)
}

func TestClientHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := HealthResponse{
			Status: "healthy",
			Checks: map[string]string{
				"database": "ok",
				"memory":   "ok",
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	health, err := client.Health(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "ok", health.Checks["database"])
}

func TestClientAnalyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/analyze", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var req AnalyzeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "test idea", req.Content)

		// Return mock analysis
		response := AnalyzeResponse{
			Analysis: &models.Analysis{
				Mission: models.MissionScores{
					Total: 3.5,
				},
				AntiChallenge: models.AntiChallengeScores{
					Total: 2.8,
				},
				Strategic: models.StrategicScores{
					Total: 2.0,
				},
				FinalScore:      8.3,
				Recommendations: []string{"PRIORITIZE NOW"},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	analysis, err := client.Analyze(context.Background(), "test idea")

	require.NoError(t, err)
	assert.Equal(t, 3.5, analysis.Mission.Total)
	assert.Equal(t, 8.3, analysis.FinalScore)
	assert.Contains(t, analysis.Recommendations, "PRIORITIZE NOW")
}

func TestClientCreateIdea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/ideas", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req CreateIdeaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "new idea", req.Content)

		idea := IdeaResponse{
			ID:             "test-id-123",
			Content:        "new idea",
			RawScore:       8.3,
			NormalizedRank: 95,
			Status:         "active",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(idea)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	idea, err := client.CreateIdea(context.Background(), "new idea")

	require.NoError(t, err)
	assert.Equal(t, "test-id-123", idea.ID)
	assert.Equal(t, "new idea", idea.Content)
	assert.Equal(t, 8.3, idea.RawScore)
	assert.Equal(t, 95, idea.NormalizedRank)
}

func TestClientGetIdea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/ideas/test-id", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		idea := IdeaResponse{
			ID:             "test-id",
			Content:        "existing idea",
			RawScore:       7.5,
			NormalizedRank: 85,
			Status:         "active",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(idea)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	idea, err := client.GetIdea(context.Background(), "test-id")

	require.NoError(t, err)
	assert.Equal(t, "test-id", idea.ID)
	assert.Equal(t, "existing idea", idea.Content)
}

func TestClientListIdeas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/ideas", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters
		query := r.URL.Query()
		assert.Equal(t, "10", query.Get("limit"))
		// Offset 0 is not included in query params (it's the default)
		assert.Equal(t, "active", query.Get("status"))
		assert.Equal(t, "score", query.Get("sort"))
		assert.Equal(t, "desc", query.Get("order"))

		response := ListIdeasResponse{
			Ideas: []IdeaResponse{
				{ID: "1", Content: "idea 1", RawScore: 9.0, Status: "active"},
				{ID: "2", Content: "idea 2", RawScore: 8.5, Status: "active"},
			},
			Total:  2,
			Limit:  10,
			Offset: 0,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	opts := &ListOptions{
		Limit:  10,
		Offset: 0,
		Status: "active",
		SortBy: "score",
		Order:  "desc",
	}
	response, err := client.ListIdeas(context.Background(), opts)

	require.NoError(t, err)
	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Ideas, 2)
	assert.Equal(t, "idea 1", response.Ideas[0].Content)
}

func TestClientListIdeasNoOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/ideas", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)

		response := ListIdeasResponse{
			Ideas:  []IdeaResponse{},
			Total:  0,
			Limit:  100,
			Offset: 0,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	response, err := client.ListIdeas(context.Background(), nil)

	require.NoError(t, err)
	assert.Equal(t, 0, response.Total)
}

func TestClientUpdateIdea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/ideas/test-id", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req UpdateIdeaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.NotNil(t, req.Content)
		assert.Equal(t, "updated content", *req.Content)

		idea := IdeaResponse{
			ID:      "test-id",
			Content: "updated content",
			Status:  "active",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(idea)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	content := "updated content"
	req := UpdateIdeaRequest{Content: &content}
	idea, err := client.UpdateIdea(context.Background(), "test-id", req)

	require.NoError(t, err)
	assert.Equal(t, "test-id", idea.ID)
	assert.Equal(t, "updated content", idea.Content)
}

func TestClientDeleteIdea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/ideas/test-id", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteIdea(context.Background(), "test-id")

	require.NoError(t, err)
}

func TestClientGetAnalyticsStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/analytics/stats", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		stats := StatsResponse{
			TotalIdeas:   100,
			ActiveIdeas:  75,
			AverageScore: 7.5,
			TopIdeas:     25,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(stats)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	stats, err := client.GetAnalyticsStats(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 100, stats.TotalIdeas)
	assert.Equal(t, 75, stats.ActiveIdeas)
	assert.Equal(t, 7.5, stats.AverageScore)
	assert.Equal(t, 25, stats.TopIdeas)
}

func TestClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.CreateIdea(context.Background(), "test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request")
	assert.Contains(t, err.Error(), "400")
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := NewClientWithTimeout(server.URL, 10*time.Millisecond)

	ctx := context.Background()
	_, err := client.Health(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Create context that cancels immediately
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Health(ctx)

	require.Error(t, err)
}
