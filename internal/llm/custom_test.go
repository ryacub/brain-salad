package llm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// TestCustomProvider_Name tests the Name() method
func TestCustomProvider_Name(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "custom name from env",
			envValue: "My Custom LLM",
			expected: "My Custom LLM",
		},
		{
			name:     "default name",
			envValue: "",
			expected: "Custom LLM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				_ = os.Setenv("CUSTOM_LLM_NAME", tt.envValue)
				defer func() { _ = os.Unsetenv("CUSTOM_LLM_NAME") }()
			}

			provider := NewCustomProvider()
			if got := provider.Name(); got != tt.expected {
				t.Errorf("Name() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCustomProvider_IsAvailable tests the IsAvailable() method
func TestCustomProvider_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected bool
	}{
		{
			name:     "endpoint configured",
			endpoint: "http://localhost:8080/api",
			expected: true,
		},
		{
			name:     "endpoint not configured",
			endpoint: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.endpoint != "" {
				_ = os.Setenv("CUSTOM_LLM_ENDPOINT", tt.endpoint)
				defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()
			}

			provider := NewCustomProvider()
			if got := provider.IsAvailable(); got != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCustomProvider_Analyze_Success tests successful analysis with JSON response
func TestCustomProvider_Analyze_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify Content-Type header
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", ct)
		}

		// Return mock response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"mission_alignment": 3.5,
			"anti_challenge": 2.8,
			"strategic_fit": 2.0,
			"final_score": 8.3,
			"recommendation": "pursue",
			"reasoning": "This is a great idea that aligns well with the mission."
		}`))
	}))
	defer server.Close()

	// Configure provider
	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()

	provider := NewCustomProvider()

	// Create test request
	req := AnalysisRequest{
		IdeaContent: "Build an AI-powered task manager",
		Telos: &models.Telos{
			Missions: []models.Mission{
				{ID: "m1", Description: "Build AI tools"},
			},
			Challenges: []models.Challenge{
				{ID: "c1", Description: "Context switching"},
			},
			Strategies: []models.Strategy{
				{ID: "s1", Description: "Leverage AI expertise"},
			},
		},
	}

	// Perform analysis
	result, err := provider.Analyze(req)

	// Verify no error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify scores
	if result.Scores.MissionAlignment != 3.5 {
		t.Errorf("Expected mission_alignment 3.5, got %f", result.Scores.MissionAlignment)
	}
	if result.Scores.AntiChallenge != 2.8 {
		t.Errorf("Expected anti_challenge 2.8, got %f", result.Scores.AntiChallenge)
	}
	if result.Scores.StrategicFit != 2.0 {
		t.Errorf("Expected strategic_fit 2.0, got %f", result.Scores.StrategicFit)
	}

	// Verify final score
	if result.FinalScore != 8.3 {
		t.Errorf("Expected final_score 8.3, got %f", result.FinalScore)
	}

	// Verify recommendation
	if result.Recommendation != "pursue" {
		t.Errorf("Expected recommendation 'pursue', got '%s'", result.Recommendation)
	}

	// Verify explanation
	if result.Explanations["overall"] == "" {
		t.Error("Expected reasoning in explanations")
	}

	// Verify provider name
	if result.Provider != "Custom LLM" {
		t.Errorf("Expected provider 'Custom LLM', got '%s'", result.Provider)
	}

	// Verify duration is set
	if result.Duration == 0 {
		t.Error("Expected duration to be set")
	}

	// Verify not from cache
	if result.FromCache {
		t.Error("Expected FromCache to be false")
	}
}

// TestCustomProvider_Analyze_AlternativeFieldNames tests JSON parsing with alternative field names
func TestCustomProvider_Analyze_AlternativeFieldNames(t *testing.T) {
	// Create mock server with alternative field names
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"missionAlignment": 3.0,
			"antiChallenge": 2.5,
			"strategicFit": 1.5,
			"score": 7.0,
			"action": "review"
		}`))
	}))
	defer server.Close()

	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()

	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Test idea"}

	result, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Scores.MissionAlignment != 3.0 {
		t.Errorf("Expected mission_alignment 3.0, got %f", result.Scores.MissionAlignment)
	}
	if result.FinalScore != 7.0 {
		t.Errorf("Expected final_score 7.0, got %f", result.FinalScore)
	}
	if result.Recommendation != "review" {
		t.Errorf("Expected recommendation 'review', got '%s'", result.Recommendation)
	}
}

// TestCustomProvider_Analyze_TextResponse tests fallback to text parsing
func TestCustomProvider_Analyze_TextResponse(t *testing.T) {
	// Create mock server returning plain text
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("This idea shows promise but needs more validation."))
	}))
	defer server.Close()

	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()

	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Test idea"}

	result, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify default scores are set
	if result.Scores.MissionAlignment != 2.0 {
		t.Errorf("Expected default mission_alignment 2.0, got %f", result.Scores.MissionAlignment)
	}
	if result.FinalScore != 5.0 {
		t.Errorf("Expected default final_score 5.0, got %f", result.FinalScore)
	}
	if result.Recommendation != "review" {
		t.Errorf("Expected default recommendation 'review', got '%s'", result.Recommendation)
	}

	// Verify text is in explanations
	if result.Explanations["overall"] == "" {
		t.Error("Expected text in explanations")
	}
}

// TestCustomProvider_Analyze_HTTPError tests handling of HTTP errors
func TestCustomProvider_Analyze_HTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			body:       "Invalid request format",
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       "Invalid API key",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			body:       "Internal server error",
		},
		{
			name:       "503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
			body:       "Service temporarily unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
			defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()

			provider := NewCustomProvider()
			req := AnalysisRequest{IdeaContent: "Test idea"}

			_, err := provider.Analyze(req)
			if err == nil {
				t.Fatal("Expected error for HTTP error status, got nil")
			}
		})
	}
}

// TestCustomProvider_Analyze_NotConfigured tests error when provider is not configured
func TestCustomProvider_Analyze_NotConfigured(t *testing.T) {
	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Test idea"}

	_, err := provider.Analyze(req)
	if err == nil {
		t.Fatal("Expected error for unconfigured provider, got nil")
	}
}

// TestCustomProvider_Analyze_WithHeaders tests custom headers
func TestCustomProvider_Analyze_WithHeaders(t *testing.T) {
	receivedHeaders := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture headers
		receivedHeaders["Authorization"] = r.Header.Get("Authorization")
		receivedHeaders["X-Custom-Header"] = r.Header.Get("X-Custom-Header")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"score": 7.0, "recommendation": "review"}`))
	}))
	defer server.Close()

	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	_ = os.Setenv("CUSTOM_LLM_HEADERS", "Authorization:Bearer secret123,X-Custom-Header:custom-value")
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_HEADERS") }()

	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Test idea"}

	_, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if receivedHeaders["Authorization"] != "Bearer secret123" {
		t.Errorf("Expected Authorization header 'Bearer secret123', got '%s'", receivedHeaders["Authorization"])
	}
	if receivedHeaders["X-Custom-Header"] != "custom-value" {
		t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", receivedHeaders["X-Custom-Header"])
	}
}

// TestCustomProvider_Analyze_WithTemplate tests custom request template
func TestCustomProvider_Analyze_WithTemplate(t *testing.T) {
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"score": 7.0, "recommendation": "review"}`))
	}))
	defer server.Close()

	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	_ = os.Setenv("CUSTOM_LLM_PROMPT_TEMPLATE", `{"text": "{{.IdeaContent}}", "max_tokens": 500}`)
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_PROMPT_TEMPLATE") }()

	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Build a task manager"}

	_, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if receivedBody["text"] != "Build a task manager" {
		t.Errorf("Expected text 'Build a task manager', got '%v'", receivedBody["text"])
	}
	if receivedBody["max_tokens"] != float64(500) {
		t.Errorf("Expected max_tokens 500, got '%v'", receivedBody["max_tokens"])
	}
}

// TestCustomProvider_Analyze_Timeout tests request timeout
func TestCustomProvider_Analyze_Timeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"score": 7.0}`))
	}))
	defer server.Close()

	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	_ = os.Setenv("CUSTOM_LLM_TIMEOUT", "1") // 1 second timeout
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_TIMEOUT") }()

	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Test idea"}

	_, err := provider.Analyze(req)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

// TestCustomProvider_ScoreValidation tests score range validation
func TestCustomProvider_ScoreValidation(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		shouldError bool
	}{
		{
			name:        "valid scores",
			response:    `{"mission_alignment": 3.5, "anti_challenge": 2.8, "strategic_fit": 2.0}`,
			shouldError: false,
		},
		{
			name:        "mission_alignment too high",
			response:    `{"mission_alignment": 5.0, "anti_challenge": 2.0, "strategic_fit": 1.0}`,
			shouldError: true,
		},
		{
			name:        "anti_challenge too high",
			response:    `{"mission_alignment": 3.0, "anti_challenge": 4.0, "strategic_fit": 1.0}`,
			shouldError: true,
		},
		{
			name:        "strategic_fit too high",
			response:    `{"mission_alignment": 3.0, "anti_challenge": 2.0, "strategic_fit": 3.0}`,
			shouldError: true,
		},
		{
			name:        "negative score",
			response:    `{"mission_alignment": -1.0, "anti_challenge": 2.0, "strategic_fit": 1.0}`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
			defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()

			provider := NewCustomProvider()
			req := AnalysisRequest{IdeaContent: "Test idea"}

			_, err := provider.Analyze(req)
			if tt.shouldError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestParseHeaders tests header parsing
func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "single header",
			input: "Authorization:Bearer token123",
			expected: map[string]string{
				"Authorization": "Bearer token123",
			},
		},
		{
			name:  "multiple headers",
			input: "Authorization:Bearer token123,Content-Type:application/json,X-Custom:value",
			expected: map[string]string{
				"Authorization": "Bearer token123",
				"Content-Type":  "application/json",
				"X-Custom":      "value",
			},
		},
		{
			name:  "headers with spaces",
			input: "Authorization: Bearer token123 , Content-Type: application/json",
			expected: map[string]string{
				"Authorization": "Bearer token123",
				"Content-Type":  "application/json",
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "malformed header",
			input:    "InvalidHeader",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHeaders(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d headers, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if gotValue, ok := result[key]; !ok {
					t.Errorf("Expected header '%s' not found", key)
				} else if gotValue != expectedValue {
					t.Errorf("Header '%s': expected '%s', got '%s'", key, expectedValue, gotValue)
				}
			}
		})
	}
}

// TestGenerateRecommendation tests recommendation generation based on score
func TestGenerateRecommendation(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{score: 9.0, expected: "strongly_pursue"},
		{score: 8.0, expected: "strongly_pursue"},
		{score: 7.5, expected: "pursue"},
		{score: 6.0, expected: "pursue"},
		{score: 5.5, expected: "review"},
		{score: 4.0, expected: "review"},
		{score: 3.0, expected: "deprioritize"},
		{score: 0.0, expected: "deprioritize"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%.1f", tt.score), func(t *testing.T) {
			result := generateRecommendation(tt.score)
			if result != tt.expected {
				t.Errorf("For score %.1f, expected '%s', got '%s'", tt.score, tt.expected, result)
			}
		})
	}
}

// TestCustomProvider_CalculatedFinalScore tests final score calculation when not provided
func TestCustomProvider_CalculatedFinalScore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No final_score in response
		_, _ = w.Write([]byte(`{
			"mission_alignment": 3.0,
			"anti_challenge": 2.5,
			"strategic_fit": 1.5
		}`))
	}))
	defer server.Close()

	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()

	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Test idea"}

	result, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedScore := 3.0 + 2.5 + 1.5
	if result.FinalScore != expectedScore {
		t.Errorf("Expected calculated final_score %.1f, got %.1f", expectedScore, result.FinalScore)
	}
}

// TestCustomProvider_DefaultRecommendation tests default recommendation when not provided
func TestCustomProvider_DefaultRecommendation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No recommendation in response
		_, _ = w.Write([]byte(`{
			"mission_alignment": 3.5,
			"anti_challenge": 3.0,
			"strategic_fit": 2.0
		}`))
	}))
	defer server.Close()

	_ = os.Setenv("CUSTOM_LLM_ENDPOINT", server.URL)
	defer func() { _ = os.Unsetenv("CUSTOM_LLM_ENDPOINT") }()

	provider := NewCustomProvider()
	req := AnalysisRequest{IdeaContent: "Test idea"}

	result, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Score is 8.5, should generate "strongly_pursue"
	if result.Recommendation != "strongly_pursue" {
		t.Errorf("Expected generated recommendation 'strongly_pursue', got '%s'", result.Recommendation)
	}
}
