package llm

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestOpenAIProvider_Name(t *testing.T) {
	provider := NewOpenAIProvider()
	assert.NotEmpty(t, provider.Name())
	assert.Contains(t, provider.Name(), "openai")
}

func TestOpenAIProvider_IsAvailable_NoAPIKey(t *testing.T) {
	provider := &OpenAIProvider{apiKey: ""}
	assert.False(t, provider.IsAvailable(), "Provider should not be available without API key")
}

func TestOpenAIProvider_Analyze_MockServer(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "{\"scores\": {\"mission_alignment\": 3.5, \"anti_challenge\": 2.8, \"strategic_fit\": 2.0}, \"final_score\": 8.3, \"recommendation\": \"GOOD ALIGNMENT\", \"explanations\": {\"mission_alignment\": \"Strong AI alignment\", \"anti_challenge\": \"Good stack fit\", \"strategic_fit\": \"Revenue potential\"}}"
				},
				"finish_reason": "stop"
			}]
		}`))
	}))
	defer server.Close()

	provider := &OpenAIProvider{
		apiKey:      "test-key",
		model:       "gpt-4",
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		maxRetries:  1,
		rateLimiter: rate.NewLimiter(rate.Inf, 1), // No rate limiting for tests
	}

	// Create a minimal telos for testing
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "test-goal", Description: "Test goal"},
		},
	}

	result, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	})

	require.NoError(t, err)
	assert.Equal(t, 3.5, result.Scores.MissionAlignment)
	assert.Equal(t, 2.8, result.Scores.AntiChallenge)
	assert.Equal(t, 2.0, result.Scores.StrategicFit)
	assert.Equal(t, 8.3, result.FinalScore)
	assert.Equal(t, "GOOD ALIGNMENT", result.Recommendation)
	assert.Contains(t, result.Provider, "openai")
	assert.False(t, result.FromCache)
}

func TestOpenAIProvider_RetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "{\"scores\": {\"mission_alignment\": 2.0, \"anti_challenge\": 2.0, \"strategic_fit\": 1.5}, \"final_score\": 5.5, \"recommendation\": \"CONSIDER LATER\", \"explanations\": {\"mission_alignment\": \"test\", \"anti_challenge\": \"test\", \"strategic_fit\": \"test\"}}"
				},
				"finish_reason": "stop"
			}]
		}`))
	}))
	defer server.Close()

	provider := &OpenAIProvider{
		apiKey:      "test-key",
		model:       "gpt-4",
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		maxRetries:  3,
		rateLimiter: rate.NewLimiter(rate.Inf, 1), // No rate limiting for tests
	}

	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "test-goal", Description: "Test goal"},
		},
	}

	result, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "test",
		Telos:       telos,
	})

	require.NoError(t, err, "Should succeed after retries")
	assert.Equal(t, 3, attempts, "Expected 3 attempts")
	assert.Equal(t, 5.5, result.FinalScore)
}

func TestOpenAIProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{
			"error": {
				"message": "Invalid API key",
				"type": "invalid_request_error",
				"code": "invalid_api_key"
			}
		}`))
	}))
	defer server.Close()

	provider := &OpenAIProvider{
		apiKey:      "invalid-key",
		model:       "gpt-4",
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		maxRetries:  1,
		rateLimiter: rate.NewLimiter(rate.Inf, 1),
	}

	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "test-goal", Description: "Test goal"},
		},
	}

	_, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "test",
		Telos:       telos,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestOpenAIProvider_SetModel(t *testing.T) {
	provider := NewOpenAIProvider()

	// Test GPT-5.1 models
	provider.SetModel("gpt-5.1")
	assert.Equal(t, "gpt-5.1", provider.GetModel())

	provider.SetModel("gpt-5.1-instant")
	assert.Equal(t, "gpt-5.1-instant", provider.GetModel())

	provider.SetModel("gpt-5.1-thinking")
	assert.Equal(t, "gpt-5.1-thinking", provider.GetModel())

	// Test GPT-5 models
	provider.SetModel("gpt-5")
	assert.Equal(t, "gpt-5", provider.GetModel())

	provider.SetModel("gpt-5-mini")
	assert.Equal(t, "gpt-5-mini", provider.GetModel())

	// Test GPT-4 models
	provider.SetModel("gpt-4o")
	assert.Equal(t, "gpt-4o", provider.GetModel())

	// Test invalid model (should not change)
	provider.SetModel("invalid-model")
	assert.Equal(t, "gpt-4o", provider.GetModel(), "Invalid model should not change current model")
}

func TestOpenAIProvider_RateLimit(t *testing.T) {
	callTimes := []time.Time{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callTimes = append(callTimes, time.Now())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "test"
				},
				"finish_reason": "stop"
			}]
		}`))
	}))
	defer server.Close()

	// Create provider with strict rate limit (1 req/sec)
	provider := &OpenAIProvider{
		apiKey:      "test-key",
		model:       "gpt-4",
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		maxRetries:  1,
		rateLimiter: rate.NewLimiter(rate.Limit(1), 1), // 1 req/sec
	}

	// Make 3 requests
	for i := 0; i < 3; i++ {
		_, _ = provider.sendRequest(&openAIRequest{
			Model:    "gpt-4",
			Messages: []openAIMessage{{Role: "user", Content: "test"}},
		})
	}

	// Verify rate limiting occurred
	assert.Equal(t, 3, len(callTimes), "Should have made 3 calls")

	// Check that there's a delay between calls
	if len(callTimes) >= 2 {
		delay := callTimes[1].Sub(callTimes[0])
		// Allow some margin for test execution
		assert.Greater(t, delay, 500*time.Millisecond, "Should have delay due to rate limiting")
	}
}

func TestOpenAIProvider_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "This is not valid JSON"
				},
				"finish_reason": "stop"
			}]
		}`))
	}))
	defer server.Close()

	provider := &OpenAIProvider{
		apiKey:      "test-key",
		model:       "gpt-4",
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		maxRetries:  1,
		rateLimiter: rate.NewLimiter(rate.Inf, 1),
	}

	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "test-goal", Description: "Test goal"},
		},
	}

	_, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "test",
		Telos:       telos,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestOpenAIProvider_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "test",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "gpt-4",
			"choices": []
		}`))
	}))
	defer server.Close()

	provider := &OpenAIProvider{
		apiKey:      "test-key",
		model:       "gpt-4",
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		maxRetries:  1,
		rateLimiter: rate.NewLimiter(rate.Inf, 1),
	}

	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "test-goal", Description: "Test goal"},
		},
	}

	_, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "test",
		Telos:       telos,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no response")
}
