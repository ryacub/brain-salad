package llm

import (
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// AnalysisRequest contains the information needed to analyze an idea using an LLM.
type AnalysisRequest struct {
	IdeaContent string        // The idea text to analyze
	Telos       *models.Telos // The parsed telos configuration
}

// AnalysisResult represents the result of an LLM analysis.
// This is the simplified output from any LLM provider.
type AnalysisResult struct {
	Scores         ScoreBreakdown    // Breakdown of scores by category
	FinalScore     float64           // Total score (0-10 scale)
	Recommendation string            // Textual recommendation
	Explanations   map[string]string // Explanations for each score category
	Provider       string            // Which provider generated this result
	Duration       time.Duration     // How long the analysis took
	FromCache      bool              // Whether result came from cache
}

// ScoreBreakdown contains the three main scoring categories.
type ScoreBreakdown struct {
	MissionAlignment float64 // 0-4.0 points max (40%)
	AntiChallenge    float64 // 0-3.5 points max (35%)
	StrategicFit     float64 // 0-2.5 points max (25%)
}

// Provider is the interface that all LLM providers must implement.
// This enables the fallback chain pattern: Ollama → Claude API → Rule-based.
type Provider interface {
	// Name returns the provider name (e.g., "ollama", "claude", "rule_based")
	Name() string

	// IsAvailable checks if the provider is currently available.
	// For Ollama, this would check if the server is running.
	// For Claude API, this would check if an API key is configured.
	// For rule-based, this always returns true.
	IsAvailable() bool

	// Analyze performs the idea analysis and returns the result.
	// Returns an error if the analysis fails.
	Analyze(req AnalysisRequest) (*AnalysisResult, error)
}

// ProviderConfig contains configuration for LLM providers.
type ProviderConfig struct {
	// Ollama configuration
	OllamaBaseURL string // Default: http://localhost:11434
	OllamaModel   string // Default: llama2
	OllamaTimeout int    // Timeout in seconds, default: 30

	// Claude API configuration
	ClaudeAPIKey  string // Claude API key (or use ANTHROPIC_API_KEY env var)
	ClaudeModel   string // Default: claude-3-5-sonnet-20241022
	ClaudeTimeout int    // Timeout in seconds, default: 30

	// OpenAI API configuration
	OpenAIAPIKey  string // OpenAI API key (or use OPENAI_API_KEY env var)
	OpenAIModel   string // Default: gpt-5.1
	OpenAITimeout int    // Timeout in seconds, default: 30

	// Custom provider configuration
	CustomEndpoint       string // Custom LLM endpoint URL (or use CUSTOM_LLM_ENDPOINT env var)
	CustomHeaders        string // Custom headers as comma-separated key:value pairs
	CustomPromptTemplate string // Go template for request body
	CustomTimeout        int    // Timeout in seconds, default: 30

	// General configuration
	EnableCache bool // Whether to cache results
	CacheTTL    int  // Cache TTL in seconds
}

// DefaultProviderConfig returns the default provider configuration.
func DefaultProviderConfig() ProviderConfig {
	return ProviderConfig{
		OllamaBaseURL: "http://localhost:11434",
		OllamaModel:   "llama2",
		OllamaTimeout: 30,
		ClaudeModel:   "claude-3-5-sonnet-20241022",
		ClaudeTimeout: 30,
		OpenAIModel:   "gpt-5.1",
		OpenAITimeout: 30,
		CustomTimeout: 30,
		EnableCache:   true,
		CacheTTL:      3600, // 1 hour
	}
}
