package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/metrics"
	"golang.org/x/time/rate"
)

// OpenAIProvider implements the Provider interface for OpenAI GPT models
type OpenAIProvider struct {
	apiKey      string
	model       string
	baseURL     string
	httpClient  *http.Client
	maxRetries  int
	rateLimiter *rate.Limiter
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-5.1" // Default to GPT-5.1 (latest flagship model)
	}

	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1/chat/completions",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries:  3,
		rateLimiter: rate.NewLimiter(rate.Limit(3), 5), // 3 req/sec, burst of 5
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return fmt.Sprintf("openai_%s", p.model)
}

// IsAvailable checks if the provider is available (has API key)
func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// Analyze performs idea analysis using OpenAI GPT models
func (p *OpenAIProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	start := time.Now()

	if !p.IsAvailable() {
		duration := time.Since(start)
		metrics.RecordLLMRequest(p.Name(), false, duration)
		metrics.RecordLLMError(p.Name(), "auth_error")
		return nil, fmt.Errorf("OpenAI provider not available (check OPENAI_API_KEY)")
	}

	// Build the analysis prompt
	prompt, err := BuildAnalysisPrompt(req.IdeaContent, req.Telos)
	if err != nil {
		duration := time.Since(start)
		metrics.RecordLLMRequest(p.Name(), false, duration)
		metrics.RecordLLMError(p.Name(), "invalid_request")
		return nil, fmt.Errorf("build prompt: %w", err)
	}

	// Create OpenAI request
	openAIReq := &openAIRequest{
		Model: p.model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are an expert at analyzing ideas against personal goals and values (telos). Provide structured analysis with scores, patterns, and recommendations.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	// Send request with retries
	var resp *openAIResponse
	var lastErr error

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		resp, lastErr = p.sendRequest(openAIReq)
		if lastErr == nil {
			break
		}

		// Exponential backoff
		if attempt < p.maxRetries-1 {
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
		}
	}

	duration := time.Since(start)

	if lastErr != nil {
		metrics.RecordLLMRequest(p.Name(), false, duration)
		metrics.RecordLLMError(p.Name(), classifyError(lastErr))
		return nil, fmt.Errorf("OpenAI request failed after %d retries: %w", p.maxRetries, lastErr)
	}

	// Parse response
	if len(resp.Choices) == 0 {
		metrics.RecordLLMRequest(p.Name(), false, duration)
		metrics.RecordLLMError(p.Name(), "invalid_response")
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Extract structured result from GPT response
	llmResp, err := ParseLLMResponse(resp.Choices[0].Message.Content)
	if err != nil {
		metrics.RecordLLMRequest(p.Name(), false, duration)
		metrics.RecordLLMError(p.Name(), "invalid_response")
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Record successful request
	metrics.RecordLLMRequest(p.Name(), true, duration)

	// Record token usage
	metrics.RecordLLMTokens(p.Name(), resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

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
		Provider:       p.Name(),
		Duration:       time.Since(start),
		FromCache:      false,
	}

	return result, nil
}

// sendRequest sends an HTTP request to OpenAI API with rate limiting
func (p *OpenAIProvider) sendRequest(req *openAIRequest) (*openAIResponse, error) {
	// Wait for rate limiter
	ctx := context.Background()
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", p.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	// Send request
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var resp openAIResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API error
	if resp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", resp.Error.Message, resp.Error.Type)
	}

	return &resp, nil
}

// SetModel allows changing the model
func (p *OpenAIProvider) SetModel(model string) {
	validModels := []string{
		// Latest GPT-5 models (2025)
		"gpt-5.1",            // GPT-5.1 (latest flagship, Nov 2025)
		"gpt-5.1-instant",    // GPT-5.1 Instant (faster, more conversational)
		"gpt-5.1-thinking",   // GPT-5.1 Thinking (advanced reasoning)
		"gpt-5.1-codex",      // GPT-5.1 Codex (specialized for coding)
		"gpt-5.1-codex-mini", // GPT-5.1 Codex Mini (smaller coding model)
		"gpt-5",              // GPT-5 (Aug 2025)
		"gpt-5-mini",         // GPT-5 Mini (faster variant)
		"gpt-5-nano",         // GPT-5 Nano (smallest variant)
		// GPT-4.5 models
		"gpt-4.5", // GPT-4.5 (research preview)
		// GPT-4 models (2024)
		"gpt-4o",              // GPT-4 Optimized
		"gpt-4o-mini",         // GPT-4 Optimized Mini
		"gpt-4-turbo",         // GPT-4 Turbo (128k context)
		"gpt-4-turbo-preview", // GPT-4 Turbo Preview
		"o1-preview",          // O1 Reasoning Model
		"o1-mini",             // O1 Mini Reasoning Model
		// Legacy models
		"gpt-4",
		"gpt-3.5-turbo",
	}

	for _, valid := range validModels {
		if model == valid {
			p.model = model
			return
		}
	}
}

// GetModel returns the current model
func (p *OpenAIProvider) GetModel() string {
	return p.model
}

// GetAPIKey returns the API key (for config display)
func (p *OpenAIProvider) GetAPIKey() string {
	return p.apiKey
}

// ============================================================================
// REQUEST/RESPONSE STRUCTURES
// ============================================================================

// openAIRequest represents the request to OpenAI API
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// openAIMessage represents a message in the conversation
type openAIMessage struct {
	Role    string `json:"role"` // "system", "user", or "assistant"
	Content string `json:"content"`
}

// openAIResponse represents the response from OpenAI API
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}
