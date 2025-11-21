// Package client provides a Go client library for the Telos Idea Matrix API.
//
// This package provides type-safe access to the tm-web API server, including:
//   - Creating, retrieving, updating, and deleting ideas
//   - Analyzing ideas against telos configuration
//   - Retrieving analytics and statistics
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Client is a Go client for the Telos Idea Matrix API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithTimeout creates a new API client with a custom timeout
func NewClientWithTimeout(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// AnalyzeRequest represents a request to analyze an idea
type AnalyzeRequest struct {
	Content string `json:"content"`
}

// AnalyzeResponse represents the response from analyzing an idea
type AnalyzeResponse struct {
	Analysis *models.Analysis `json:"analysis"`
}

// CreateIdeaRequest represents a request to create an idea
type CreateIdeaRequest struct {
	Content string `json:"content"`
}

// UpdateIdeaRequest represents a request to update an idea
type UpdateIdeaRequest struct {
	Content *string `json:"content,omitempty"`
	Status  *string `json:"status,omitempty"`
}

// IdeaResponse represents an idea in API responses
type IdeaResponse struct {
	ID             string           `json:"id"`
	Content        string           `json:"content"`
	RawScore       float64          `json:"raw_score"`
	NormalizedRank int              `json:"normalized_rank"`
	Status         string           `json:"status"`
	Analysis       *models.Analysis `json:"analysis,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// ListIdeasResponse represents a paginated list of ideas
type ListIdeasResponse struct {
	Ideas  []IdeaResponse `json:"ideas"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// StatsResponse represents analytics statistics
type StatsResponse struct {
	TotalIdeas   int     `json:"total_ideas"`
	ActiveIdeas  int     `json:"active_ideas"`
	AverageScore float64 `json:"average_score"`
	TopIdeas     int     `json:"top_ideas"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// ============================================================================
// API Methods
// ============================================================================

// Health checks the health of the API server
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var response HealthResponse
	if err := c.get(ctx, "/health", &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Analyze analyzes an idea and returns the analysis result
func (c *Client) Analyze(ctx context.Context, content string) (*models.Analysis, error) {
	req := AnalyzeRequest{Content: content}
	var response AnalyzeResponse

	if err := c.post(ctx, "/api/v1/analyze", req, &response); err != nil {
		return nil, err
	}

	return response.Analysis, nil
}

// CreateIdea creates a new idea via the API
func (c *Client) CreateIdea(ctx context.Context, content string) (*IdeaResponse, error) {
	req := CreateIdeaRequest{Content: content}
	var idea IdeaResponse

	if err := c.post(ctx, "/api/v1/ideas", req, &idea); err != nil {
		return nil, err
	}

	return &idea, nil
}

// GetIdea retrieves an idea by ID
func (c *Client) GetIdea(ctx context.Context, id string) (*IdeaResponse, error) {
	var idea IdeaResponse
	path := fmt.Sprintf("/api/v1/ideas/%s", id)

	if err := c.get(ctx, path, &idea); err != nil {
		return nil, err
	}

	return &idea, nil
}

// ListIdeas lists all ideas with optional filtering
func (c *Client) ListIdeas(ctx context.Context, opts *ListOptions) (*ListIdeasResponse, error) {
	query := url.Values{}

	if opts != nil {
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			query.Set("offset", strconv.Itoa(opts.Offset))
		}
		if opts.Status != "" {
			query.Set("status", opts.Status)
		}
		if opts.SortBy != "" {
			query.Set("sort", opts.SortBy)
		}
		if opts.Order != "" {
			query.Set("order", opts.Order)
		}
	}

	path := "/api/v1/ideas"
	if len(query) > 0 {
		path = fmt.Sprintf("%s?%s", path, query.Encode())
	}

	var response ListIdeasResponse
	if err := c.get(ctx, path, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// UpdateIdea updates an existing idea
func (c *Client) UpdateIdea(ctx context.Context, id string, req UpdateIdeaRequest) (*IdeaResponse, error) {
	path := fmt.Sprintf("/api/v1/ideas/%s", id)
	var idea IdeaResponse

	if err := c.put(ctx, path, req, &idea); err != nil {
		return nil, err
	}

	return &idea, nil
}

// DeleteIdea deletes an idea by ID
func (c *Client) DeleteIdea(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/ideas/%s", id)
	return c.delete(ctx, path)
}

// GetAnalyticsStats retrieves analytics statistics
func (c *Client) GetAnalyticsStats(ctx context.Context) (*StatsResponse, error) {
	var stats StatsResponse
	if err := c.get(ctx, "/api/v1/analytics/stats", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// ============================================================================
// Helper Types
// ============================================================================

// ListOptions contains options for listing ideas
type ListOptions struct {
	Limit  int
	Offset int
	Status string // Filter by status: "active", "completed", "archived"
	SortBy string // Sort by field: "score", "created_at", "updated_at"
	Order  string // Sort order: "asc", "desc"
}

// ============================================================================
// Internal HTTP Methods
// ============================================================================

// get performs a GET request
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, "GET", path, nil, result)
}

// post performs a POST request
func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, "POST", path, body, result)
}

// put performs a PUT request
func (c *Client) put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, "PUT", path, body, result)
}

// delete performs a DELETE request
func (c *Client) delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, "DELETE", path, nil, nil)
}

// doRequest performs an HTTP request with the given method, path, and body
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	// Build full URL
	fullURL := c.baseURL + path

	// Marshal request body if present
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to parse error response
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Decode response if result is provided
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
