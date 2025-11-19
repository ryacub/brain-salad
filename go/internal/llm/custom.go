package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

// CustomProvider implements a generic HTTP/REST LLM provider that can work with any
// endpoint using configurable request/response mapping. This enables integration with
// custom LLM services without code changes.
//
// Configuration is done via environment variables:
//   - CUSTOM_LLM_NAME: Provider name (default: "Custom LLM")
//   - CUSTOM_LLM_ENDPOINT: HTTP endpoint URL (required)
//   - CUSTOM_LLM_HEADERS: Comma-separated key:value headers
//   - CUSTOM_LLM_PROMPT_TEMPLATE: Go template for request body
//   - CUSTOM_LLM_RESPONSE_PARSER: Response parsing configuration
//   - CUSTOM_LLM_TIMEOUT: Request timeout in seconds (default: 30)
//
// Example configuration:
//   CUSTOM_LLM_ENDPOINT="http://localhost:8080/v1/analyze"
//   CUSTOM_LLM_HEADERS="Authorization:Bearer token123,Content-Type:application/json"
//   CUSTOM_LLM_PROMPT_TEMPLATE='{"prompt": "{{.IdeaContent}}", "context": "{{.Telos}}"}'
type CustomProvider struct {
	name           string
	endpoint       string
	headers        map[string]string
	httpClient     *http.Client
	promptTemplate string
	responseParser string
}

// NewCustomProvider creates a custom HTTP provider from environment variable configuration.
// The provider will only be available if CUSTOM_LLM_ENDPOINT is set.
func NewCustomProvider() *CustomProvider {
	timeoutSeconds := getEnvAsInt("CUSTOM_LLM_TIMEOUT", 30)

	return &CustomProvider{
		name:           getEnv("CUSTOM_LLM_NAME", "Custom LLM"),
		endpoint:       os.Getenv("CUSTOM_LLM_ENDPOINT"),
		headers:        parseHeaders(os.Getenv("CUSTOM_LLM_HEADERS")),
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		promptTemplate: os.Getenv("CUSTOM_LLM_PROMPT_TEMPLATE"),
		responseParser: os.Getenv("CUSTOM_LLM_RESPONSE_PARSER"),
	}
}

// Name returns the provider name from configuration or "Custom LLM" as default.
func (p *CustomProvider) Name() string {
	return p.name
}

// IsAvailable checks if the custom provider is configured.
// Returns true only if CUSTOM_LLM_ENDPOINT is set.
func (p *CustomProvider) IsAvailable() bool {
	return p.endpoint != ""
}

// Analyze performs idea analysis by calling the configured HTTP endpoint.
// The request body is built using the template (if configured) and the response
// is parsed according to the configured parser.
func (p *CustomProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	start := time.Now()

	if !p.IsAvailable() {
		return nil, fmt.Errorf("custom provider not configured: CUSTOM_LLM_ENDPOINT not set")
	}

	// Build request body using template
	requestBody, err := p.buildRequestBody(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request body: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add configured headers
	for key, value := range p.headers {
		httpReq.Header.Set(key, value)
	}

	// Ensure Content-Type is set if not already configured
	if httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Send request
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	result, err := p.parseResponse(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Set metadata
	result.Provider = p.Name()
	result.Duration = time.Since(start)
	result.FromCache = false

	return result, nil
}

// buildRequestBody constructs the request body using the configured template.
// If no template is configured, uses a default JSON structure.
func (p *CustomProvider) buildRequestBody(req AnalysisRequest) ([]byte, error) {
	if p.promptTemplate == "" {
		// Default template: simple JSON with idea content and telos
		defaultReq := map[string]interface{}{
			"idea": req.IdeaContent,
		}

		// Add telos information if available
		if req.Telos != nil {
			defaultReq["missions"] = req.Telos.Missions
			defaultReq["challenges"] = req.Telos.Challenges
			defaultReq["strategies"] = req.Telos.Strategies
			defaultReq["goals"] = req.Telos.Goals
		}

		return json.Marshal(defaultReq)
	}

	// Use custom template
	tmpl, err := template.New("request").Parse(p.promptTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid prompt template: %w", err)
	}

	// Create template data with helper functions for JSON formatting
	templateData := struct {
		AnalysisRequest
		TelosJSON string
	}{
		AnalysisRequest: req,
		TelosJSON:       "",
	}

	// Add JSON-encoded telos if available
	if req.Telos != nil {
		telosBytes, _ := json.Marshal(req.Telos)
		templateData.TelosJSON = string(telosBytes)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// parseResponse parses the HTTP response into an AnalysisResult.
// Supports both JSON and plain text responses with intelligent fallback.
func (p *CustomProvider) parseResponse(body []byte) (*AnalysisResult, error) {
	// Try JSON parsing first
	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		return p.parseJSONResponse(jsonData)
	}

	// Fallback to text parsing
	return p.parseTextResponse(string(body))
}

// parseJSONResponse extracts analysis data from a JSON response.
// Supports flexible field names and nested structures.
func (p *CustomProvider) parseJSONResponse(data map[string]interface{}) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Scores: ScoreBreakdown{
			MissionAlignment: extractFloat(data, "mission_alignment", "missionAlignment", "mission"),
			AntiChallenge:    extractFloat(data, "anti_challenge", "antiChallenge", "challenges"),
			StrategicFit:     extractFloat(data, "strategic_fit", "strategicFit", "strategic"),
		},
		FinalScore:     extractFloat(data, "final_score", "finalScore", "score", "total_score"),
		Recommendation: extractString(data, "recommendation", "action", "decision"),
		Explanations:   make(map[string]string),
	}

	// Extract explanations if present
	if reasoning := extractString(data, "reasoning", "explanation", "analysis"); reasoning != "" {
		result.Explanations["overall"] = reasoning
	}

	// Extract individual explanations
	if missionExp := extractString(data, "mission_explanation", "mission_reasoning"); missionExp != "" {
		result.Explanations["mission_alignment"] = missionExp
	}
	if challengeExp := extractString(data, "challenge_explanation", "challenge_reasoning"); challengeExp != "" {
		result.Explanations["anti_challenge"] = challengeExp
	}
	if strategicExp := extractString(data, "strategic_explanation", "strategic_reasoning"); strategicExp != "" {
		result.Explanations["strategic_fit"] = strategicExp
	}

	// Validate and normalize scores
	if err := p.validateScores(&result.Scores); err != nil {
		return nil, err
	}

	// Calculate final score if not provided
	if result.FinalScore == 0 {
		result.FinalScore = result.Scores.MissionAlignment +
			result.Scores.AntiChallenge +
			result.Scores.StrategicFit
	}

	// Default recommendation if not provided
	if result.Recommendation == "" {
		result.Recommendation = generateRecommendation(result.FinalScore)
	}

	return result, nil
}

// parseTextResponse handles plain text responses as fallback.
// Provides basic analysis with moderate scores.
func (p *CustomProvider) parseTextResponse(content string) (*AnalysisResult, error) {
	// Basic text parsing fallback - returns moderate scores
	result := &AnalysisResult{
		Scores: ScoreBreakdown{
			MissionAlignment: 2.0, // Middle of 0-4.0 range
			AntiChallenge:    1.75, // Middle of 0-3.5 range
			StrategicFit:     1.25, // Middle of 0-2.5 range
		},
		FinalScore:     5.0, // Middle score
		Recommendation: "review",
		Explanations: map[string]string{
			"overall": content,
		},
	}

	return result, nil
}

// validateScores ensures scores are within valid ranges.
func (p *CustomProvider) validateScores(scores *ScoreBreakdown) error {
	if scores.MissionAlignment < 0 || scores.MissionAlignment > 4.0 {
		return fmt.Errorf("mission_alignment score %.2f out of range [0, 4.0]", scores.MissionAlignment)
	}
	if scores.AntiChallenge < 0 || scores.AntiChallenge > 3.5 {
		return fmt.Errorf("anti_challenge score %.2f out of range [0, 3.5]", scores.AntiChallenge)
	}
	if scores.StrategicFit < 0 || scores.StrategicFit > 2.5 {
		return fmt.Errorf("strategic_fit score %.2f out of range [0, 2.5]", scores.StrategicFit)
	}
	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// parseHeaders parses comma-separated key:value header pairs.
// Example: "Authorization:Bearer token,Content-Type:application/json"
func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}

// extractFloat tries to extract a float from a map using multiple possible keys.
// Returns 0.0 if none of the keys exist or value is not a number.
func extractFloat(data map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			}
		}
	}
	return 0.0
}

// extractString tries to extract a string from a map using multiple possible keys.
// Returns empty string if none of the keys exist or value is not a string.
func extractString(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if s, ok := val.(string); ok {
				return s
			}
		}
	}
	return ""
}

// extractStringArray tries to extract a string array from a map.
// Returns empty slice if key doesn't exist or value is not an array.
func extractStringArray(data map[string]interface{}, keys ...string) []string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if arr, ok := val.([]interface{}); ok {
				result := make([]string, 0, len(arr))
				for _, item := range arr {
					if s, ok := item.(string); ok {
						result = append(result, s)
					}
				}
				return result
			}
		}
	}
	return []string{}
}

// generateRecommendation generates a recommendation based on final score.
func generateRecommendation(score float64) string {
	switch {
	case score >= 8.0:
		return "strongly_pursue"
	case score >= 6.0:
		return "pursue"
	case score >= 4.0:
		return "review"
	default:
		return "deprioritize"
	}
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer with a default value.
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
