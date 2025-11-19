package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

func TestNewClaudeProvider(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		model         string
		expectedModel string
	}{
		{
			name:          "default model",
			apiKey:        "test-key",
			model:         "",
			expectedModel: "claude-3-5-sonnet-20241022",
		},
		{
			name:          "custom model",
			apiKey:        "test-key",
			model:         "claude-3-opus-20240229",
			expectedModel: "claude-3-opus-20240229",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewClaudeProvider(tt.apiKey, tt.model)
			if provider.model != tt.expectedModel {
				t.Errorf("expected model %s, got %s", tt.expectedModel, provider.model)
			}
			if provider.apiKey != tt.apiKey {
				t.Errorf("expected apiKey %s, got %s", tt.apiKey, provider.apiKey)
			}
		})
	}
}

func TestClaudeProvider_Name(t *testing.T) {
	provider := NewClaudeProvider("test-key", "")
	if provider.Name() != "claude" {
		t.Errorf("expected name 'claude', got '%s'", provider.Name())
	}
}

func TestClaudeProvider_IsAvailable_NoAPIKey(t *testing.T) {
	provider := &ClaudeProvider{apiKey: ""}
	if provider.IsAvailable() {
		t.Error("should not be available without API key")
	}
}

func TestClaudeProvider_IsAvailable_WithMockServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]string{
				{"type": "text", "text": "test"},
			},
			"model":       "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
		})
	}))
	defer server.Close()

	provider := &ClaudeProvider{
		apiKey:     "test-key",
		model:      "claude-3-5-sonnet-20241022",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		maxRetries: 1,
	}

	if !provider.IsAvailable() {
		t.Error("should be available with valid API key and server")
	}
}

func TestClaudeProvider_Headers(t *testing.T) {
	// Create mock server to verify headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify required headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing or incorrect x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("missing or incorrect anthropic-version header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing or incorrect Content-Type header")
		}

		// Return valid response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": `{
						"scores": {
							"mission_alignment": 3.5,
							"anti_challenge": 2.8,
							"strategic_fit": 2.0
						},
						"final_score": 8.3,
						"recommendation": "GOOD ALIGNMENT",
						"explanations": {
							"mission_alignment": "Strong AI alignment",
							"anti_challenge": "Good execution support",
							"strategic_fit": "Fits tech stack well"
						}
					}`,
				},
			},
			"model":       "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
		})
	}))
	defer server.Close()

	// Create provider with mock server
	provider := NewClaudeProvider("test-key", "")
	provider.baseURL = server.URL

	// Create test telos
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "g1", Description: "Build AI products"},
		},
	}

	// Make request
	_, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "Build an AI-powered tool",
		Telos:       telos,
	})

	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
}

func TestClaudeProvider_Analyze_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "msg_test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": `{
						"scores": {
							"mission_alignment": 3.5,
							"anti_challenge": 2.8,
							"strategic_fit": 2.0
						},
						"final_score": 8.3,
						"recommendation": "GOOD ALIGNMENT",
						"explanations": {
							"mission_alignment": "Strong AI alignment",
							"anti_challenge": "Good execution support",
							"strategic_fit": "Fits tech stack well"
						}
					}`,
				},
			},
			"model":       "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
			"usage": map[string]int{
				"input_tokens":  100,
				"output_tokens": 200,
			},
		})
	}))
	defer server.Close()

	// Create provider
	provider := NewClaudeProvider("test-key", "")
	provider.baseURL = server.URL

	// Create test telos
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "g1", Description: "Build AI products"},
		},
	}

	// Analyze
	result, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "Build an AI-powered tool",
		Telos:       telos,
	})

	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Verify result
	if result.Scores.MissionAlignment != 3.5 {
		t.Errorf("expected mission_alignment 3.5, got %f", result.Scores.MissionAlignment)
	}
	if result.Scores.AntiChallenge != 2.8 {
		t.Errorf("expected anti_challenge 2.8, got %f", result.Scores.AntiChallenge)
	}
	if result.Scores.StrategicFit != 2.0 {
		t.Errorf("expected strategic_fit 2.0, got %f", result.Scores.StrategicFit)
	}
	if result.FinalScore != 8.3 {
		t.Errorf("expected final_score 8.3, got %f", result.FinalScore)
	}
	if result.Recommendation != "GOOD ALIGNMENT" {
		t.Errorf("expected recommendation 'GOOD ALIGNMENT', got '%s'", result.Recommendation)
	}
	if result.Provider != "claude" {
		t.Errorf("expected provider 'claude', got '%s'", result.Provider)
	}
}

func TestClaudeProvider_Analyze_APIError(t *testing.T) {
	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"type": "error",
			"error": map[string]string{
				"type":    "authentication_error",
				"message": "invalid API key",
			},
		})
	}))
	defer server.Close()

	// Create provider
	provider := NewClaudeProvider("invalid-key", "")
	provider.baseURL = server.URL

	// Create test telos
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "g1", Description: "Build AI products"},
		},
	}

	// Analyze should fail
	_, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "Build an AI-powered tool",
		Telos:       telos,
	})

	if err == nil {
		t.Error("expected error for invalid API key")
	}
}

func TestClaudeProvider_SetModel(t *testing.T) {
	provider := NewClaudeProvider("test-key", "")

	tests := []struct {
		name          string
		model         string
		expectedModel string
	}{
		{
			name:          "valid sonnet model",
			model:         "claude-3-5-sonnet-20241022",
			expectedModel: "claude-3-5-sonnet-20241022",
		},
		{
			name:          "valid opus model",
			model:         "claude-3-opus-20240229",
			expectedModel: "claude-3-opus-20240229",
		},
		{
			name:          "valid haiku model",
			model:         "claude-3-haiku-20240307",
			expectedModel: "claude-3-haiku-20240307",
		},
		{
			name:          "invalid model (should not change)",
			model:         "invalid-model",
			expectedModel: "claude-3-5-sonnet-20241022", // Should remain default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default
			provider.model = "claude-3-5-sonnet-20241022"

			// Set model
			provider.SetModel(tt.model)

			// Verify
			if provider.GetModel() != tt.expectedModel {
				t.Errorf("expected model %s, got %s", tt.expectedModel, provider.GetModel())
			}
		})
	}
}

func TestClaudeProvider_GetAPIKey(t *testing.T) {
	provider := NewClaudeProvider("test-api-key-123", "")
	if provider.GetAPIKey() != "test-api-key-123" {
		t.Errorf("expected API key 'test-api-key-123', got '%s'", provider.GetAPIKey())
	}
}

func TestClaudeProvider_extractPrompts(t *testing.T) {
	provider := NewClaudeProvider("test-key", "")

	tests := []struct {
		name           string
		fullPrompt     string
		expectedSystem string
		expectedUser   string
	}{
		{
			name: "prompt with TASK section",
			fullPrompt: `You are an expert at evaluating ideas.

TELOS:
Goals and values here.

TASK:
Analyze this idea and provide scores.`,
			expectedSystem: `You are an expert at evaluating ideas.

TELOS:
Goals and values here.`,
			expectedUser: `TASK:
Analyze this idea and provide scores.`,
		},
		{
			name:           "prompt without TASK section",
			fullPrompt:     "Simple prompt without task section",
			expectedSystem: "",
			expectedUser:   "Simple prompt without task section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system, user := provider.extractPrompts(tt.fullPrompt)
			if system != tt.expectedSystem {
				t.Errorf("expected system:\n%s\n\ngot:\n%s", tt.expectedSystem, system)
			}
			if user != tt.expectedUser {
				t.Errorf("expected user:\n%s\n\ngot:\n%s", tt.expectedUser, user)
			}
		})
	}
}

func TestClaudeProvider_RetryLogic(t *testing.T) {
	attempts := 0
	// Account for IsAvailable check + actual analysis attempts
	// IsAvailable makes 1 call, then Analyze makes up to maxRetries calls
	isAvailableAttempts := 0

	// Create mock server that fails initially then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		// First request is from IsAvailable - always succeed
		if isAvailableAttempts == 0 {
			isAvailableAttempts++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "test",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{"type": "text", "text": "ok"},
				},
				"model":       "claude-3-5-sonnet-20241022",
				"stop_reason": "end_turn",
			})
			return
		}

		// For actual analysis requests, fail first 2 times, succeed on 3rd
		analyzeAttempt := attempts - isAvailableAttempts
		if analyzeAttempt < 2 {
			// Fail the first two analysis attempts
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Succeed on the third analysis attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "msg_test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": `{
						"scores": {
							"mission_alignment": 3.0,
							"anti_challenge": 2.5,
							"strategic_fit": 1.8
						},
						"final_score": 7.3,
						"recommendation": "GOOD ALIGNMENT",
						"explanations": {}
					}`,
				},
			},
			"model":       "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
		})
	}))
	defer server.Close()

	// Create provider
	provider := NewClaudeProvider("test-key", "")
	provider.baseURL = server.URL
	provider.maxRetries = 3

	// Create test telos
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "g1", Description: "Test goal"},
		},
	}

	// This should succeed after retries
	result, err := provider.Analyze(AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	})

	if err != nil {
		t.Fatalf("Analyze should succeed after retries, got error: %v", err)
	}

	if result.FinalScore != 7.3 {
		t.Errorf("expected final_score 7.3, got %f", result.FinalScore)
	}

	// Should have 1 IsAvailable attempt + 2 analyze attempts (1 failure + 1 success)
	expectedAttempts := 3
	if attempts != expectedAttempts {
		t.Errorf("expected %d total attempts, got %d", expectedAttempts, attempts)
	}
}
