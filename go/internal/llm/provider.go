package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/llm/client"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
)

// ============================================================================
// OLLAMA PROVIDER
// ============================================================================

// OllamaProvider implements the Provider interface using Ollama.
type OllamaProvider struct {
	client *client.OllamaClient
	model  string
}

// NewOllamaProvider creates a new Ollama provider with the given configuration.
// If model is empty, defaults to "llama2".
func NewOllamaProvider(baseURL string, model string) *OllamaProvider {
	if model == "" {
		model = "llama2"
	}

	return &OllamaProvider{
		client: client.NewOllamaClient(baseURL, 30*time.Second),
		model:  model,
	}
}

// Name returns the provider name.
func (op *OllamaProvider) Name() string {
	return "ollama"
}

// IsAvailable checks if Ollama is running and accessible.
func (op *OllamaProvider) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return op.client.HealthCheck(ctx) == nil
}

// Analyze performs idea analysis using Ollama.
func (op *OllamaProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	start := time.Now()

	// Build prompt
	prompt, err := BuildAnalysisPrompt(req.IdeaContent, req.Telos)
	if err != nil {
		return nil, fmt.Errorf("build prompt: %w", err)
	}

	// Generate analysis using Ollama
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := op.client.Generate(ctx, client.GenerateRequest{
		Model:  op.model,
		Prompt: prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	// Parse LLM response
	llmResp, err := ParseLLMResponse(resp.Response)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Convert to AnalysisResult
	result := &AnalysisResult{
		Scores: ScoreBreakdown{
			MissionAlignment: llmResp.Scores.MissionAlignment,
			AntiChallenge:    llmResp.Scores.AntiChallenge,
			StrategicFit:     llmResp.Scores.StrategicFit,
		},
		FinalScore:     llmResp.FinalScore,
		Recommendation: llmResp.Recommendation,
		Explanations:   llmResp.Explanations,
		Provider:       op.Name(),
		Duration:       time.Since(start),
		FromCache:      false,
	}

	return result, nil
}

// ============================================================================
// RULE-BASED PROVIDER
// ============================================================================

// RuleBasedProvider implements the Provider interface using rule-based scoring.
// This provider always works and serves as the ultimate fallback.
type RuleBasedProvider struct {
	engine *scoring.Engine
}

// NewRuleBasedProvider creates a new rule-based provider.
// Note: The engine will be initialized with a telos when Analyze is called.
func NewRuleBasedProvider() *RuleBasedProvider {
	return &RuleBasedProvider{}
}

// Name returns the provider name.
func (rbp *RuleBasedProvider) Name() string {
	return "rule_based"
}

// IsAvailable always returns true as this provider is always available.
func (rbp *RuleBasedProvider) IsAvailable() bool {
	return true
}

// Analyze performs rule-based analysis using the scoring engine.
func (rbp *RuleBasedProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	start := time.Now()

	if req.Telos == nil {
		return nil, fmt.Errorf("telos is required for rule-based analysis")
	}

	// Create scoring engine with telos
	engine := scoring.NewEngine(req.Telos)

	// Calculate scores
	analysis, err := engine.CalculateScore(req.IdeaContent)
	if err != nil {
		return nil, fmt.Errorf("calculate score: %w", err)
	}

	// Convert to AnalysisResult
	result := &AnalysisResult{
		Scores: ScoreBreakdown{
			MissionAlignment: analysis.Mission.Total,
			AntiChallenge:    analysis.AntiChallenge.Total,
			StrategicFit:     analysis.Strategic.Total,
		},
		FinalScore:     analysis.FinalScore,
		Recommendation: analysis.GetRecommendation(),
		Explanations:   make(map[string]string),
		Provider:       rbp.Name(),
		Duration:       time.Since(start),
		FromCache:      false,
	}

	// Add basic explanations based on scores
	result.Explanations["mission_alignment"] = fmt.Sprintf(
		"Score: %.2f/4.0 (Domain: %.2f, AI: %.2f, Execution: %.2f, Revenue: %.2f)",
		analysis.Mission.Total,
		analysis.Mission.DomainExpertise,
		analysis.Mission.AIAlignment,
		analysis.Mission.ExecutionSupport,
		analysis.Mission.RevenuePotential,
	)
	result.Explanations["anti_challenge"] = fmt.Sprintf(
		"Score: %.2f/3.5 (Context: %.2f, Prototyping: %.2f, Accountability: %.2f, Income: %.2f)",
		analysis.AntiChallenge.Total,
		analysis.AntiChallenge.ContextSwitching,
		analysis.AntiChallenge.RapidPrototyping,
		analysis.AntiChallenge.Accountability,
		analysis.AntiChallenge.IncomeAnxiety,
	)
	result.Explanations["strategic_fit"] = fmt.Sprintf(
		"Score: %.2f/2.5 (Stack: %.2f, Shipping: %.2f, Public: %.2f, Revenue: %.2f)",
		analysis.Strategic.Total,
		analysis.Strategic.StackCompatibility,
		analysis.Strategic.ShippingHabit,
		analysis.Strategic.PublicAccountability,
		analysis.Strategic.RevenueTesting,
	)

	return result, nil
}

// ============================================================================
// FALLBACK PROVIDER
// ============================================================================

// FallbackProvider implements the Provider interface by chaining multiple providers.
// It tries each provider in order until one succeeds.
type FallbackProvider struct {
	providers []Provider
}

// NewFallbackProvider creates a new fallback provider with the given providers.
// Providers are tried in the order they are provided.
func NewFallbackProvider(providers ...Provider) *FallbackProvider {
	return &FallbackProvider{
		providers: providers,
	}
}

// Name returns the provider name.
func (fp *FallbackProvider) Name() string {
	return "fallback"
}

// IsAvailable returns true if at least one provider is available.
func (fp *FallbackProvider) IsAvailable() bool {
	for _, p := range fp.providers {
		if p.IsAvailable() {
			return true
		}
	}
	return false
}

// Analyze tries each provider in order until one succeeds.
// If all providers fail, returns the last error encountered.
func (fp *FallbackProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	var lastErr error

	for _, provider := range fp.providers {
		// Skip unavailable providers
		if !provider.IsAvailable() {
			continue
		}

		// Try to analyze with this provider
		result, err := provider.Analyze(req)
		if err == nil {
			// Success! Return the result
			return result, nil
		}

		// Save the error and try the next provider
		lastErr = err
	}

	// All providers failed
	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("no providers available")
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// CreateDefaultFallbackChain creates the default fallback chain:
// Ollama → Claude API (stub) → Rule-based
func CreateDefaultFallbackChain(config ProviderConfig, telos *models.Telos) *FallbackProvider {
	providers := []Provider{
		NewOllamaProvider(config.OllamaBaseURL, config.OllamaModel),
		// TODO: Add Claude provider when Track 5B is implemented
		NewRuleBasedProvider(),
	}

	return NewFallbackProvider(providers...)
}
