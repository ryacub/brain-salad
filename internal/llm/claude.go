package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/llm/processing"
	"github.com/rayyacub/telos-idea-matrix/internal/metrics"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// ============================================================================
// CLAUDE PROVIDER
// ============================================================================

// ClaudeProvider implements the Provider interface using Anthropic Claude API.
type ClaudeProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	maxRetries int
	processor  *processing.SimpleProcessor
}

// NewClaudeProvider creates a new Claude provider with the given configuration.
// If apiKey is empty, it will try to read from ANTHROPIC_API_KEY environment variable.
// If model is empty, defaults to "claude-3-5-sonnet-20241022".
func NewClaudeProvider(apiKey string, model string) *ClaudeProvider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
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

	return &ClaudeProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,             // Max idle connections across all hosts
				MaxIdleConnsPerHost: 10,              // Max idle connections per host
				MaxConnsPerHost:     10,              // Max total connections per host
				IdleConnTimeout:     90 * time.Second, // Keep idle connections for 90s
				DisableKeepAlives:   false,           // Enable connection reuse
			},
		},
		maxRetries: 3,
		processor:  processor,
	}
}

// Name returns the provider name.
func (cp *ClaudeProvider) Name() string {
	return "claude"
}

// IsAvailable checks if Claude API is accessible.
// It verifies that an API key is configured and can make a test request.
func (cp *ClaudeProvider) IsAvailable() bool {
	if cp.apiKey == "" {
		return false
	}

	// Test with minimal request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &claudeRequest{
		Model: cp.model,
		Messages: []claudeMessage{
			{Role: "user", Content: "test"},
		},
		MaxTokens: 10,
	}

	_, err := cp.sendRequest(ctx, req)
	return err == nil
}

// Analyze performs idea analysis using Claude API.
func (cp *ClaudeProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	start := time.Now()

	if !cp.IsAvailable() {
		duration := time.Since(start)
		metrics.RecordLLMRequest(cp.Name(), false, duration)
		metrics.RecordLLMError(cp.Name(), "auth_error")
		return nil, fmt.Errorf("claude provider not available (check ANTHROPIC_API_KEY)")
	}

	// Build prompt
	prompt, err := BuildAnalysisPrompt(req.IdeaContent, req.Telos)
	if err != nil {
		duration := time.Since(start)
		metrics.RecordLLMRequest(cp.Name(), false, duration)
		metrics.RecordLLMError(cp.Name(), "invalid_request")
		return nil, fmt.Errorf("build prompt: %w", err)
	}

	// Extract system prompt and user prompt
	systemPrompt, userPrompt := cp.extractPrompts(prompt)

	// Create Claude request
	claudeReq := &claudeRequest{
		Model:  cp.model,
		System: systemPrompt,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	// Send request with retries and exponential backoff
	ctx := context.Background()
	var resp *claudeResponse
	var lastErr error

	for attempt := 0; attempt < cp.maxRetries; attempt++ {
		// Add timeout for each attempt
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		resp, lastErr = cp.sendRequest(attemptCtx, claudeReq)
		cancel()

		if lastErr == nil {
			break
		}

		// Exponential backoff before retry
		if attempt < cp.maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	duration := time.Since(start)

	if lastErr != nil {
		metrics.RecordLLMRequest(cp.Name(), false, duration)
		metrics.RecordLLMError(cp.Name(), classifyError(lastErr))
		return nil, fmt.Errorf("claude request failed after %d retries: %w", cp.maxRetries, lastErr)
	}

	// Extract text from response
	if len(resp.Content) == 0 {
		metrics.RecordLLMRequest(cp.Name(), false, duration)
		metrics.RecordLLMError(cp.Name(), "invalid_response")
		return nil, fmt.Errorf("no response content from Claude")
	}

	responseText := resp.Content[0].Text

	// Process LLM response with fallback support
	processed, err := cp.processor.Process(responseText, req.IdeaContent, req.Telos)
	if err != nil {
		metrics.RecordLLMRequest(cp.Name(), false, duration)
		metrics.RecordLLMError(cp.Name(), "invalid_response")
		return nil, fmt.Errorf("process response: %w", err)
	}

	// Record successful request
	metrics.RecordLLMRequest(cp.Name(), true, duration)

	// Record token usage
	metrics.RecordLLMTokens(cp.Name(), resp.Usage.InputTokens, resp.Usage.OutputTokens)

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
		Provider:       cp.Name(),
		Duration:       time.Since(start),
		FromCache:      false,
	}

	return result, nil
}

// ============================================================================
// CLAUDE API STRUCTURES
// ============================================================================

// claudeRequest represents a request to Claude API.
type claudeRequest struct {
	Model       string          `json:"model"`
	Messages    []claudeMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	System      string          `json:"system,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// claudeMessage represents a message in the Claude API request.
type claudeMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// claudeResponse represents a response from Claude API.
type claudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// extractPrompts extracts system and user prompts from the full prompt.
// Claude API requires system prompt to be separate from messages.
func (cp *ClaudeProvider) extractPrompts(fullPrompt string) (string, string) {
	// Look for the task section to split system from user content
	taskIndex := strings.Index(fullPrompt, "TASK:")
	if taskIndex == -1 {
		// If no TASK section, treat entire prompt as user message
		return "", fullPrompt
	}

	// Everything before TASK is system prompt, everything from TASK onwards is user
	systemPrompt := strings.TrimSpace(fullPrompt[:taskIndex])
	userPrompt := strings.TrimSpace(fullPrompt[taskIndex:])

	return systemPrompt, userPrompt
}

// sendRequest sends a request to Claude API.
func (cp *ClaudeProvider) sendRequest(ctx context.Context, req *claudeRequest) (*claudeResponse, error) {
	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", cp.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set Claude-specific headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", cp.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	httpResp, err := cp.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var resp claudeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Check for API error
	if resp.Error != nil {
		return nil, fmt.Errorf("claude API error: %s (type: %s)", resp.Error.Message, resp.Error.Type)
	}

	return &resp, nil
}

// SetModel updates the Claude model being used.
func (cp *ClaudeProvider) SetModel(model string) {
	validModels := []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}

	for _, valid := range validModels {
		if model == valid {
			cp.model = model
			return
		}
	}
}

// GetModel returns the current Claude model.
func (cp *ClaudeProvider) GetModel() string {
	return cp.model
}

// GetAPIKey returns the configured API key.
func (cp *ClaudeProvider) GetAPIKey() string {
	return cp.apiKey
}
