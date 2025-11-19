package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaClient is an HTTP client for interacting with the Ollama API.
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// GenerateRequest represents a request to generate text using Ollama.
type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// GenerateResponse represents a response from Ollama's generate endpoint.
type GenerateResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

// TagsResponse represents the response from the /api/tags endpoint.
type TagsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ModelInfo represents information about an Ollama model.
type ModelInfo struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at,omitempty"`
	Size       int64     `json:"size,omitempty"`
}

// NewOllamaClient creates a new Ollama client with the given base URL and timeout.
// If baseURL is empty, defaults to "http://localhost:11434".
// If timeout is 0, defaults to 30 seconds.
func NewOllamaClient(baseURL string, timeout time.Duration) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Generate sends a generation request to Ollama and returns the response.
// The stream parameter is always set to false for simplicity.
func (c *OllamaClient) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	// Disable streaming for simplicity
	req.Stream = false

	// Marshal the request
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/generate", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error: status %d", resp.StatusCode)
	}

	// Decode the response
	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// ListModels retrieves the list of available models from Ollama.
func (c *OllamaClient) ListModels(ctx context.Context) ([]string, error) {
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error: status %d", resp.StatusCode)
	}

	// Decode the response
	var tagsResp TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Extract model names
	models := make([]string, 0, len(tagsResp.Models))
	for _, model := range tagsResp.Models {
		models = append(models, model.Name)
	}

	return models, nil
}

// HealthCheck performs a simple health check by attempting to list models.
// Returns nil if Ollama is healthy, error otherwise.
func (c *OllamaClient) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama not healthy: status %d", resp.StatusCode)
	}

	return nil
}
