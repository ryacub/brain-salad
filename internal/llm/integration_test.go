//go:build integration
// +build integration

package llm

import (
	"context"
	"testing"
	"time"

	llmclient "github.com/rayyacub/telos-idea-matrix/internal/llm/client"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// TestOllamaIntegration_EndToEnd tests the full integration with Ollama
// Run with: go test -tags=integration ./internal/llm -v
func TestOllamaIntegration_EndToEnd(t *testing.T) {
	// Skip if Ollama is not available
	ollamaClient := llmclient.NewOllamaClient("http://localhost:11434", 5*time.Second)
	ctx := context.Background()

	if err := ollamaClient.HealthCheck(ctx); err != nil {
		t.Skip("Ollama is not running - skipping integration test")
	}

	// Create Ollama provider
	provider := NewOllamaProvider("http://localhost:11434", "llama2")

	if !provider.IsAvailable() {
		t.Skip("Ollama provider not available")
	}

	// Create test request
	req := AnalysisRequest{
		IdeaContent: "Build an AI automation tool using Python and GPT-4 with $2000/month subscription model",
		Telos:       createIntegrationTelos(),
	}

	// Perform analysis
	result, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	// Validate result
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Provider != "ollama" {
		t.Errorf("expected provider 'ollama', got %s", result.Provider)
	}
	if result.FinalScore < 0 || result.FinalScore > 10 {
		t.Errorf("final score out of range: %f", result.FinalScore)
	}
	if result.Scores.MissionAlignment < 0 || result.Scores.MissionAlignment > 4.0 {
		t.Errorf("mission alignment out of range: %f", result.Scores.MissionAlignment)
	}
	if result.Scores.AntiChallenge < 0 || result.Scores.AntiChallenge > 3.5 {
		t.Errorf("anti challenge out of range: %f", result.Scores.AntiChallenge)
	}
	if result.Scores.StrategicFit < 0 || result.Scores.StrategicFit > 2.5 {
		t.Errorf("strategic fit out of range: %f", result.Scores.StrategicFit)
	}
	if result.Duration == 0 {
		t.Error("expected duration to be set")
	}

	t.Logf("Analysis completed successfully:")
	t.Logf("  Final Score: %.2f/10", result.FinalScore)
	t.Logf("  Recommendation: %s", result.Recommendation)
	t.Logf("  Duration: %v", result.Duration)
	t.Logf("  Mission Alignment: %.2f", result.Scores.MissionAlignment)
	t.Logf("  Anti-Challenge: %.2f", result.Scores.AntiChallenge)
	t.Logf("  Strategic Fit: %.2f", result.Scores.StrategicFit)
}

// TestFallbackChainIntegration tests the fallback chain with real Ollama
func TestFallbackChainIntegration(t *testing.T) {
	// Create fallback chain
	config := DefaultProviderConfig()
	telos := createIntegrationTelos()
	provider := CreateDefaultFallbackChain(config, telos)

	if !provider.IsAvailable() {
		t.Skip("No providers available")
	}

	// Test analysis
	req := AnalysisRequest{
		IdeaContent: "Create a simple Python automation script for personal use",
		Telos:       telos,
	}

	result, err := provider.Analyze(req)
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	// Validate result
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}

	// Should use Ollama if available, otherwise rule-based
	if result.Provider != "ollama" && result.Provider != "rule_based" {
		t.Errorf("unexpected provider: %s", result.Provider)
	}

	t.Logf("Fallback chain used provider: %s", result.Provider)
	t.Logf("Final Score: %.2f/10", result.FinalScore)
}

// TestOllamaClient_ListModelsIntegration tests listing available models
func TestOllamaClient_ListModelsIntegration(t *testing.T) {
	ollamaClient := llmclient.NewOllamaClient("http://localhost:11434", 5*time.Second)
	ctx := context.Background()

	// Check if Ollama is available
	if err := ollamaClient.HealthCheck(ctx); err != nil {
		t.Skip("Ollama is not running")
	}

	// List models
	models, err := ollamaClient.ListModels(ctx)
	if err != nil {
		t.Fatalf("failed to list models: %v", err)
	}

	if len(models) == 0 {
		t.Skip("No models available in Ollama")
	}

	t.Logf("Available models: %v", models)

	// Verify we can use at least one model
	for _, model := range models {
		resp, err := ollamaClient.Generate(ctx, llmclient.GenerateRequest{
			Model:  model,
			Prompt: "Say 'Hello, World!' and nothing else.",
		})
		if err != nil {
			t.Logf("Model %s failed: %v", model, err)
			continue
		}
		if resp.Response == "" {
			t.Errorf("Model %s returned empty response", model)
		}
		t.Logf("Model %s works: %s", model, resp.Response)
		break
	}
}

// TestPromptBuilding tests that prompts are built correctly
func TestPromptBuildingIntegration(t *testing.T) {
	telos := createIntegrationTelos()
	idea := "Build an AI tool"

	prompt, err := BuildAnalysisPrompt(idea, telos)
	if err != nil {
		t.Fatalf("failed to build prompt: %v", err)
	}

	// Validate prompt contains key elements
	if prompt == "" {
		t.Fatal("prompt is empty")
	}
	if len(prompt) < 100 {
		t.Error("prompt seems too short")
	}

	// Should contain the idea
	if !contains(prompt, idea) {
		t.Error("prompt doesn't contain the idea")
	}

	// Should contain scoring framework
	if !contains(prompt, "Mission Alignment") {
		t.Error("prompt doesn't contain 'Mission Alignment'")
	}
	if !contains(prompt, "Anti-Challenge") {
		t.Error("prompt doesn't contain 'Anti-Challenge'")
	}
	if !contains(prompt, "Strategic Fit") {
		t.Error("prompt doesn't contain 'Strategic Fit'")
	}

	t.Logf("Prompt length: %d characters", len(prompt))
}

// TestRuleBasedFallbackIntegration tests that rule-based always works
func TestRuleBasedFallbackIntegration(t *testing.T) {
	// Create a provider that will fail
	failingProvider := &MockProvider{
		name:      "failing",
		available: false,
	}

	// Create fallback chain with failing provider + rule-based
	ruleProvider := NewRuleBasedProvider()
	fallback := NewFallbackProvider(failingProvider, ruleProvider)

	req := AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       createIntegrationTelos(),
	}

	result, err := fallback.Analyze(req)
	if err != nil {
		t.Fatalf("fallback to rule-based failed: %v", err)
	}

	if result.Provider != "rule_based" {
		t.Errorf("expected rule_based provider, got %s", result.Provider)
	}

	t.Logf("Rule-based fallback worked successfully: %.2f/10", result.FinalScore)
}

// Helper to create a test telos
func createIntegrationTelos() *models.Telos {
	deadline := time.Now().Add(90 * 24 * time.Hour)
	return &models.Telos{
		Goals: []models.Goal{
			{
				ID:          "G1",
				Description: "Ship AI-powered automation tools that generate recurring revenue",
				Priority:    1,
				Deadline:    &deadline,
			},
			{
				ID:          "G2",
				Description: "Build expertise in AI/ML and modern Python development",
				Priority:    2,
			},
		},
		Strategies: []models.Strategy{
			{
				ID:          "S1",
				Description: "Focus on rapid prototyping and MVP delivery",
			},
			{
				ID:          "S2",
				Description: "Leverage existing Python and AI skills",
			},
			{
				ID:          "S3",
				Description: "Build in public for accountability",
			},
		},
		Stack: models.Stack{
			Primary:   []string{"Python", "Go", "PostgreSQL", "Docker"},
			Secondary: []string{"TypeScript", "Redis", "AWS"},
		},
		FailurePatterns: []models.Pattern{
			{
				Name:        "Context Switching",
				Description: "Avoid switching to completely new tech stacks mid-project",
				Keywords:    []string{"context", "switching", "distraction"},
			},
			{
				Name:        "Perfection Paralysis",
				Description: "Ship MVPs instead of waiting for perfect solutions",
				Keywords:    []string{"perfect", "complete", "polished"},
			},
		},
		LoadedAt: time.Now(),
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
