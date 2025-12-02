package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ryacub/telos-idea-matrix/internal/llm/client"
	"github.com/ryacub/telos-idea-matrix/internal/llm/processing"
	"github.com/ryacub/telos-idea-matrix/internal/llm/quality"
	"github.com/ryacub/telos-idea-matrix/internal/metrics"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/ryacub/telos-idea-matrix/internal/scoring"
)

// Error types for better error classification
var (
	ErrTimeout         = errors.New("request timeout")
	ErrRateLimit       = errors.New("rate limit exceeded")
	ErrAuth            = errors.New("authentication failed")
	ErrNetwork         = errors.New("network error")
	ErrInvalidResponse = errors.New("invalid response")
	ErrProvider        = errors.New("provider error")
)

// Global quality tracker for all LLM analyses
var (
	globalQualityTracker *quality.SimpleTracker
	trackerOnce          sync.Once
)

// GetQualityTracker returns the global quality tracker with thread-safe initialization
func GetQualityTracker() *quality.SimpleTracker {
	trackerOnce.Do(func() {
		globalQualityTracker = quality.NewSimpleTracker()
	})
	return globalQualityTracker
}

// ============================================================================
// OLLAMA PROVIDER
// ============================================================================

// OllamaProvider implements the Provider interface using Ollama.
type OllamaProvider struct {
	client    *client.OllamaClient
	model     string
	processor *processing.SimpleProcessor
}

// NewOllamaProvider creates a new Ollama provider with the given configuration.
// If model is empty, defaults to "llama2".
func NewOllamaProvider(baseURL string, model string) *OllamaProvider {
	if model == "" {
		model = "llama2"
	}

	// Create processor with rule-based fallback function
	fallbackFunc := func(ideaContent string, telos interface{}) (*processing.ProcessedResult, error) {
		// Use rule-based provider as fallback
		ruleProvider := NewRuleBasedProvider()

		// Convert telos to proper type
		var telosModel *models.Telos
		if t, ok := telos.(*models.Telos); ok {
			telosModel = t
		}

		result, err := ruleProvider.Analyze(AnalysisRequest{
			IdeaContent: ideaContent,
			Telos:       telosModel,
		})
		if err != nil {
			return nil, err
		}

		// Convert to ProcessedResult
		return &processing.ProcessedResult{
			Scores: processing.ScoreBreakdown{
				MissionAlignment: result.Scores.MissionAlignment,
				AntiChallenge:    result.Scores.AntiChallenge,
				StrategicFit:     result.Scores.StrategicFit,
			},
			FinalScore:     result.FinalScore,
			Recommendation: result.Recommendation,
			Explanations:   result.Explanations,
			Provider:       result.Provider,
			UsedFallback:   true,
		}, nil
	}

	processor := processing.NewSimpleProcessor(fallbackFunc)

	return &OllamaProvider{
		client:    client.NewOllamaClient(baseURL, 30*time.Second),
		model:     model,
		processor: processor,
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

	duration := time.Since(start)

	// Record metrics
	if err != nil {
		// Import metrics package at the top if not already imported
		// Record failure
		metrics.RecordLLMRequest(op.Name(), false, duration)
		metrics.RecordLLMError(op.Name(), classifyError(err))
		return nil, fmt.Errorf("generate: %w", err)
	}

	// Process LLM response with fallback support
	processed, err := op.processor.Process(resp.Response, req.IdeaContent, req.Telos)
	if err != nil {
		// Record failure
		metrics.RecordLLMRequest(op.Name(), false, duration)
		metrics.RecordLLMError(op.Name(), "invalid_response")
		return nil, fmt.Errorf("process response: %w", err)
	}

	// Record successful request
	metrics.RecordLLMRequest(op.Name(), true, duration)

	// Note: Ollama doesn't provide token counts in response, so we can't track tokens

	// Convert to AnalysisResult
	result := &AnalysisResult{
		Scores: ScoreBreakdown{
			MissionAlignment: processed.Scores.MissionAlignment,
			AntiChallenge:    processed.Scores.AntiChallenge,
			StrategicFit:     processed.Scores.StrategicFit,
		},
		FinalScore:     processed.FinalScore,
		Recommendation: processed.Recommendation,
		Explanations:   processed.Explanations,
		Provider:       op.Name(),
		Duration:       time.Since(start),
		FromCache:      false,
	}

	// Track quality metrics
	simpleResult := &quality.SimpleResult{
		MissionAlignment: result.Scores.MissionAlignment,
		AntiChallenge:    result.Scores.AntiChallenge,
		StrategicFit:     result.Scores.StrategicFit,
		FinalScore:       result.FinalScore,
		Explanations:     result.Explanations,
		Provider:         result.Provider,
	}
	qualityMetrics := GetQualityTracker().Record(simpleResult)

	// Log quality metrics (optional - could be removed in production)
	_ = qualityMetrics // Suppress unused variable warning

	return result, nil
}

// ============================================================================
// RULE-BASED PROVIDER
// ============================================================================

// RuleBasedProvider implements the Provider interface using rule-based scoring.
// This provider always works and serves as the ultimate fallback.
type RuleBasedProvider struct {
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
		duration := time.Since(start)
		metrics.RecordLLMRequest(rbp.Name(), false, duration)
		metrics.RecordLLMError(rbp.Name(), "invalid_request")
		return nil, fmt.Errorf("telos is required for rule-based analysis")
	}

	// Create scoring engine with telos
	engine := scoring.NewEngine(req.Telos)

	// Calculate scores
	analysis, err := engine.CalculateScore(req.IdeaContent)
	duration := time.Since(start)

	if err != nil {
		metrics.RecordLLMRequest(rbp.Name(), false, duration)
		metrics.RecordLLMError(rbp.Name(), "calculation_error")
		return nil, fmt.Errorf("calculate score: %w", err)
	}

	// Record successful request
	metrics.RecordLLMRequest(rbp.Name(), true, duration)

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
// Ollama → Claude API → Rule-based
func CreateDefaultFallbackChain(config ProviderConfig, _ *models.Telos) *FallbackProvider {
	providers := []Provider{
		NewOllamaProvider(config.OllamaBaseURL, config.OllamaModel),
		NewClaudeProvider(config.ClaudeAPIKey, config.ClaudeModel),
		NewRuleBasedProvider(),
	}

	return NewFallbackProvider(providers...)
}

// classifyError categorizes errors into standard types for metrics tracking
// Uses error type checking first, then falls back to string matching for compatibility
func classifyError(err error) string {
	if err == nil {
		return "unknown"
	}

	// Check for specific error types using errors.Is()
	switch {
	case errors.Is(err, ErrTimeout):
		return "timeout"
	case errors.Is(err, ErrRateLimit):
		return "rate_limit"
	case errors.Is(err, ErrAuth):
		return "auth_error"
	case errors.Is(err, ErrNetwork):
		return "network_error"
	case errors.Is(err, ErrInvalidResponse):
		return "invalid_response"
	case errors.Is(err, ErrProvider):
		return "provider_error"
	}

	// Fall back to string matching for errors that don't use typed errors
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		return "timeout"
	case strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "too many requests") || strings.Contains(errStr, "429"):
		return "rate_limit"
	case strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "api key") || strings.Contains(errStr, "authentication") || strings.Contains(errStr, "401"):
		return "auth_error"
	case strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") || strings.Contains(errStr, "dial"):
		return "network_error"
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "malformed") || strings.Contains(errStr, "parse"):
		return "invalid_response"
	case strings.Contains(errStr, "500") || strings.Contains(errStr, "502") || strings.Contains(errStr, "503") || strings.Contains(errStr, "504"):
		return "provider_error"
	default:
		return "unknown"
	}
}
