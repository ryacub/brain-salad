package llm

import (
	"testing"
	"time"

	"github.com/ryacub/telos-idea-matrix/internal/models"
)

func TestAnalysisRequest(t *testing.T) {
	req := AnalysisRequest{
		IdeaContent: "Build an AI automation tool",
		Telos:       &models.Telos{},
	}

	if req.IdeaContent == "" {
		t.Error("expected idea content to be set")
	}
	if req.Telos == nil {
		t.Error("expected telos to be set")
	}
}

func TestAnalysisResult(t *testing.T) {
	result := AnalysisResult{
		Scores: ScoreBreakdown{
			MissionAlignment: 3.5,
			AntiChallenge:    2.8,
			StrategicFit:     2.0,
		},
		FinalScore:     8.3,
		Recommendation: "PRIORITIZE NOW",
		Explanations: map[string]string{
			"mission_alignment": "Strong AI focus",
		},
		Provider:  "ollama",
		Duration:  2 * time.Second,
		FromCache: false,
	}

	if result.FinalScore != 8.3 {
		t.Errorf("expected final score 8.3, got %f", result.FinalScore)
	}
	if result.Scores.MissionAlignment != 3.5 {
		t.Errorf("expected mission alignment 3.5, got %f", result.Scores.MissionAlignment)
	}
	if result.Provider != "ollama" {
		t.Errorf("expected provider 'ollama', got %s", result.Provider)
	}
}

func TestScoreBreakdown(t *testing.T) {
	scores := ScoreBreakdown{
		MissionAlignment: 4.0,
		AntiChallenge:    3.5,
		StrategicFit:     2.5,
	}

	total := scores.MissionAlignment + scores.AntiChallenge + scores.StrategicFit
	if total != 10.0 {
		t.Errorf("expected total 10.0, got %f", total)
	}
}

func TestScoreBreakdown_PartialScores(t *testing.T) {
	scores := ScoreBreakdown{
		MissionAlignment: 2.5,
		AntiChallenge:    2.0,
		StrategicFit:     1.5,
	}

	total := scores.MissionAlignment + scores.AntiChallenge + scores.StrategicFit
	expected := 6.0
	if total != expected {
		t.Errorf("expected total %f, got %f", expected, total)
	}
}

func TestDefaultProviderConfig(t *testing.T) {
	config := DefaultProviderConfig()

	if config.OllamaBaseURL != "http://localhost:11434" {
		t.Errorf("expected default ollama base URL, got %s", config.OllamaBaseURL)
	}
	if config.OllamaModel != "llama2" {
		t.Errorf("expected default ollama model 'llama2', got %s", config.OllamaModel)
	}
	if config.OllamaTimeout != 30 {
		t.Errorf("expected default timeout 30, got %d", config.OllamaTimeout)
	}
	if !config.EnableCache {
		t.Error("expected cache to be enabled by default")
	}
	if config.CacheTTL != 3600 {
		t.Errorf("expected default cache TTL 3600, got %d", config.CacheTTL)
	}
}

func TestProviderConfig_CustomValues(t *testing.T) {
	config := ProviderConfig{
		OllamaBaseURL: "http://custom:11434",
		OllamaModel:   "mistral",
		OllamaTimeout: 60,
		ClaudeAPIKey:  "test-key",
		EnableCache:   false,
	}

	if config.OllamaBaseURL != "http://custom:11434" {
		t.Errorf("expected custom ollama base URL, got %s", config.OllamaBaseURL)
	}
	if config.OllamaModel != "mistral" {
		t.Errorf("expected custom model 'mistral', got %s", config.OllamaModel)
	}
	if config.ClaudeAPIKey != "test-key" {
		t.Errorf("expected claude API key 'test-key', got %s", config.ClaudeAPIKey)
	}
	if config.EnableCache {
		t.Error("expected cache to be disabled")
	}
}
