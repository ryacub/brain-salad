package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// MockProvider is a test implementation of the Provider interface
type MockProvider struct {
	name      string
	available bool
	result    *AnalysisResult
	err       error
	callCount int
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) IsAvailable() bool {
	return m.available
}

func (m *MockProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestProvider_OllamaProvider_Analyze(t *testing.T) {
	// This test will fail until we implement OllamaProvider
	provider := NewOllamaProvider("http://localhost:11434", "llama2")

	if provider.Name() != "ollama" {
		t.Errorf("expected provider name 'ollama', got %s", provider.Name())
	}

	// Skip actual analysis in unit tests unless Ollama is running
	// We'll test this in integration tests
	t.Skip("requires Ollama to be running - covered by integration tests")
}

func TestProvider_FallbackChain(t *testing.T) {
	// Create mock providers
	successProvider := &MockProvider{
		name:      "success",
		available: true,
		result: &AnalysisResult{
			FinalScore:     8.0,
			Recommendation: "PRIORITIZE NOW",
			Provider:       "success",
		},
	}

	failingProvider := &MockProvider{
		name:      "failing",
		available: true,
		err:       errors.New("provider failed"),
	}

	unavailableProvider := &MockProvider{
		name:      "unavailable",
		available: false,
	}

	tests := []struct {
		name          string
		providers     []Provider
		wantProvider  string
		wantError     bool
		wantCallCount map[string]int
	}{
		{
			name:         "first provider succeeds",
			providers:    []Provider{successProvider, failingProvider},
			wantProvider: "success",
			wantError:    false,
			wantCallCount: map[string]int{
				"success": 1,
				"failing": 0,
			},
		},
		{
			name:         "first provider fails, second succeeds",
			providers:    []Provider{failingProvider, successProvider},
			wantProvider: "success",
			wantError:    false,
			wantCallCount: map[string]int{
				"failing": 1,
				"success": 1,
			},
		},
		{
			name:         "skip unavailable provider",
			providers:    []Provider{unavailableProvider, successProvider},
			wantProvider: "success",
			wantError:    false,
			wantCallCount: map[string]int{
				"unavailable": 0,
				"success":     1,
			},
		},
		{
			name:      "all providers fail",
			providers: []Provider{failingProvider, unavailableProvider},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset call counts
			for _, p := range tt.providers {
				if mp, ok := p.(*MockProvider); ok {
					mp.callCount = 0
				}
			}

			fallback := NewFallbackProvider(tt.providers...)
			req := AnalysisRequest{
				IdeaContent: "test idea",
				Telos:       createMockTelos(),
			}

			result, err := fallback.Analyze(req)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if !tt.wantError {
				if result == nil {
					t.Fatal("expected result to be non-nil")
					return
				}
				if result.Provider != tt.wantProvider {
					t.Errorf("expected provider %s, got %s", tt.wantProvider, result.Provider)
				}
			}

			// Check call counts
			for providerName, expectedCount := range tt.wantCallCount {
				for _, p := range tt.providers {
					if mp, ok := p.(*MockProvider); ok && mp.name == providerName {
						if mp.callCount != expectedCount {
							t.Errorf("expected %s to be called %d times, got %d",
								providerName, expectedCount, mp.callCount)
						}
					}
				}
			}
		})
	}
}

func TestProvider_FallbackToRuleBased(t *testing.T) {
	// Create a rule-based provider (always available)
	ruleProvider := NewRuleBasedProvider()

	if ruleProvider.Name() != "rule_based" {
		t.Errorf("expected provider name 'rule_based', got %s", ruleProvider.Name())
	}

	if !ruleProvider.IsAvailable() {
		t.Error("expected rule-based provider to always be available")
	}

	// Test analysis
	req := AnalysisRequest{
		IdeaContent: "Build an AI automation tool using Python and GPT-4",
		Telos:       createMockTelos(),
	}

	result, err := ruleProvider.Analyze(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
		return
	}
	if result.Provider != "rule_based" {
		t.Errorf("expected provider 'rule_based', got %s", result.Provider)
	}
	if result.FinalScore < 0 || result.FinalScore > 10 {
		t.Errorf("expected final score between 0-10, got %f", result.FinalScore)
	}
	if result.Scores.MissionAlignment < 0 || result.Scores.MissionAlignment > 4.0 {
		t.Errorf("expected mission alignment 0-4.0, got %f", result.Scores.MissionAlignment)
	}
	if result.Scores.AntiChallenge < 0 || result.Scores.AntiChallenge > 3.5 {
		t.Errorf("expected anti-challenge 0-3.5, got %f", result.Scores.AntiChallenge)
	}
	if result.Scores.StrategicFit < 0 || result.Scores.StrategicFit > 2.5 {
		t.Errorf("expected strategic fit 0-2.5, got %f", result.Scores.StrategicFit)
	}
}

func TestProvider_ClaudeProvider_Stub(t *testing.T) {
	// This test is a stub for Track 5B
	// It will be implemented when Claude provider is added
	t.Skip("Claude provider implementation is part of Track 5B")
}

func TestProvider_Timeouts(t *testing.T) {
	slowProvider := &MockProvider{
		name:      "slow",
		available: true,
		err:       context.DeadlineExceeded,
	}

	fastProvider := &MockProvider{
		name:      "fast",
		available: true,
		result: &AnalysisResult{
			FinalScore: 7.0,
			Provider:   "fast",
		},
	}

	fallback := NewFallbackProvider(slowProvider, fastProvider)
	req := AnalysisRequest{
		IdeaContent: "test idea",
		Telos:       createMockTelos(),
	}

	result, err := fallback.Analyze(req)
	if err != nil {
		t.Fatalf("expected fallback to succeed, got error: %v", err)
	}
	if result.Provider != "fast" {
		t.Errorf("expected fast provider to be used, got %s", result.Provider)
	}
}

func TestRuleBasedProvider_WithDifferentIdeas(t *testing.T) {
	provider := NewRuleBasedProvider()

	tests := []struct {
		name     string
		idea     string
		minScore float64
		maxScore float64
	}{
		{
			name:     "high quality idea with AI and revenue",
			idea:     "Build an AI automation tool using Python with $2000/month subscription",
			minScore: 6.0,
			maxScore: 10.0,
		},
		{
			name:     "medium quality idea",
			idea:     "Create a simple automation script",
			minScore: 2.0,
			maxScore: 5.0,
		},
		{
			name:     "low quality idea",
			idea:     "Learn Rust for 6 months before starting project",
			minScore: 0.0,
			maxScore: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := AnalysisRequest{
				IdeaContent: tt.idea,
				Telos:       createMockTelos(),
			}

			result, err := provider.Analyze(req)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if result.FinalScore < tt.minScore || result.FinalScore > tt.maxScore {
				t.Errorf("expected score between %f-%f, got %f",
					tt.minScore, tt.maxScore, result.FinalScore)
			}
		})
	}
}

func TestFallbackProvider_Name(t *testing.T) {
	provider1 := &MockProvider{name: "provider1", available: true}
	provider2 := &MockProvider{name: "provider2", available: true}

	fallback := NewFallbackProvider(provider1, provider2)
	if fallback.Name() != "fallback" {
		t.Errorf("expected name 'fallback', got %s", fallback.Name())
	}
}

func TestFallbackProvider_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		providers []Provider
		want      bool
	}{
		{
			name: "at least one available",
			providers: []Provider{
				&MockProvider{name: "p1", available: false},
				&MockProvider{name: "p2", available: true},
			},
			want: true,
		},
		{
			name: "none available",
			providers: []Provider{
				&MockProvider{name: "p1", available: false},
				&MockProvider{name: "p2", available: false},
			},
			want: false,
		},
		{
			name: "all available",
			providers: []Provider{
				&MockProvider{name: "p1", available: true},
				&MockProvider{name: "p2", available: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallback := NewFallbackProvider(tt.providers...)
			if got := fallback.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create a mock Telos for testing
func createMockTelos() *models.Telos {
	return &models.Telos{
		Goals: []models.Goal{
			{
				ID:          "G1",
				Description: "Ship AI-powered automation tools",
				Priority:    1,
			},
		},
		Strategies: []models.Strategy{
			{
				ID:          "S1",
				Description: "Focus on rapid prototyping",
			},
		},
		Stack: models.Stack{
			Primary:   []string{"Python", "Go", "PostgreSQL"},
			Secondary: []string{"Docker", "Redis"},
		},
		FailurePatterns: []models.Pattern{
			{
				Name:        "Context Switching",
				Description: "Avoid switching between too many technologies",
				Keywords:    []string{"context", "switching"},
			},
		},
		LoadedAt: time.Now(),
	}
}
