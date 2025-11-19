# Custom HTTP Provider

The `CustomProvider` is a flexible HTTP provider that enables integration with any REST API endpoint for LLM analysis. This allows users to connect their own LLM services without modifying code.

## Features

- ✅ Configurable via environment variables
- ✅ Template-based request customization
- ✅ Flexible JSON response parsing with multiple field name variants
- ✅ Text response fallback
- ✅ Custom HTTP headers support
- ✅ Configurable timeout
- ✅ Score validation (ensures scores are within valid ranges)
- ✅ Automatic recommendation generation

## Configuration

All configuration is done via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CUSTOM_LLM_ENDPOINT` | ✅ Yes | - | HTTP endpoint URL for the custom LLM service |
| `CUSTOM_LLM_NAME` | No | "Custom LLM" | Display name for the provider |
| `CUSTOM_LLM_HEADERS` | No | - | Comma-separated key:value HTTP headers |
| `CUSTOM_LLM_PROMPT_TEMPLATE` | No | Default JSON | Go template for request body |
| `CUSTOM_LLM_TIMEOUT` | No | 30 | Request timeout in seconds |

## Basic Setup

### Example 1: Simple Local LLM

For a basic local LLM service that accepts JSON with an `idea` field:

```bash
export CUSTOM_LLM_NAME="Local Llama"
export CUSTOM_LLM_ENDPOINT="http://localhost:8080/v1/analyze"
```

This will send requests in the default format:
```json
{
  "idea": "Your idea content here",
  "missions": [...],
  "challenges": [...],
  "strategies": [...],
  "goals": [...]
}
```

### Example 2: OpenAI-Compatible API

For services that use OpenAI-compatible format:

```bash
export CUSTOM_LLM_NAME="Local OpenAI"
export CUSTOM_LLM_ENDPOINT="http://localhost:8000/v1/chat/completions"
export CUSTOM_LLM_HEADERS="Authorization:Bearer sk-your-api-key,Content-Type:application/json"
export CUSTOM_LLM_PROMPT_TEMPLATE='{"model":"gpt-4","messages":[{"role":"user","content":"Analyze this idea: {{.IdeaContent}}"}]}'
```

### Example 3: Cloud-Hosted Custom Service

For a cloud service with authentication:

```bash
export CUSTOM_LLM_NAME="Company LLM"
export CUSTOM_LLM_ENDPOINT="https://llm.company.com/api/analyze"
export CUSTOM_LLM_HEADERS="Authorization:Bearer your-secret-token,X-API-Version:v2"
export CUSTOM_LLM_TIMEOUT="60"
```

### Example 4: Anthropic Claude-style API

For Claude API or compatible services:

```bash
export CUSTOM_LLM_NAME="Claude Custom"
export CUSTOM_LLM_ENDPOINT="https://api.anthropic.com/v1/messages"
export CUSTOM_LLM_HEADERS="x-api-key:your-api-key,anthropic-version:2023-06-01,Content-Type:application/json"
export CUSTOM_LLM_PROMPT_TEMPLATE='{"model":"claude-3-sonnet-20240229","max_tokens":1024,"messages":[{"role":"user","content":"Analyze idea: {{.IdeaContent}}"}]}'
```

## Request Templates

The `CUSTOM_LLM_PROMPT_TEMPLATE` uses Go's text/template syntax. Available fields:

- `{{.IdeaContent}}` - The idea text to analyze
- `{{.Telos}}` - The full Telos object (use with caution, may be large)
- `{{.TelosJSON}}` - JSON-encoded telos data

### Template Examples

**Simple text prompt:**
```bash
export CUSTOM_LLM_PROMPT_TEMPLATE='{"prompt":"{{.IdeaContent}}","max_tokens":500}'
```

**With system instructions:**
```bash
export CUSTOM_LLM_PROMPT_TEMPLATE='{"system":"You are an idea analyzer.","prompt":"{{.IdeaContent}}"}'
```

**Including telos context:**
```bash
export CUSTOM_LLM_PROMPT_TEMPLATE='{"idea":"{{.IdeaContent}}","context":{{.TelosJSON}}}'
```

## Response Format

The CustomProvider can parse various JSON response formats. It looks for these field names (in order of preference):

### Score Fields

| Category | Field Names (checked in order) |
|----------|-------------------------------|
| Mission Alignment | `mission_alignment`, `missionAlignment`, `mission` |
| Anti-Challenge | `anti_challenge`, `antiChallenge`, `challenges` |
| Strategic Fit | `strategic_fit`, `strategicFit`, `strategic` |
| Final Score | `final_score`, `finalScore`, `score`, `total_score` |
| Recommendation | `recommendation`, `action`, `decision` |

### Expected Score Ranges

- **Mission Alignment**: 0-4.0
- **Anti-Challenge**: 0-3.5
- **Strategic Fit**: 0-2.5
- **Final Score**: 0-10.0 (sum of above)

### Example Response Formats

**Standard format:**
```json
{
  "mission_alignment": 3.5,
  "anti_challenge": 2.8,
  "strategic_fit": 2.0,
  "final_score": 8.3,
  "recommendation": "pursue",
  "reasoning": "This idea aligns well with your mission..."
}
```

**Camel case format:**
```json
{
  "missionAlignment": 3.5,
  "antiChallenge": 2.8,
  "strategicFit": 2.0,
  "score": 8.3,
  "action": "pursue"
}
```

**Minimal format (scores only):**
```json
{
  "mission": 3.5,
  "challenges": 2.8,
  "strategic": 2.0
}
```
*Note: If `final_score` is missing, it will be calculated as the sum of individual scores. If `recommendation` is missing, it will be generated based on the final score.*

### Recommendation Values

Valid recommendation values:
- `strongly_pursue` - Score >= 8.0
- `pursue` - Score >= 6.0
- `review` - Score >= 4.0
- `deprioritize` - Score < 4.0

## Text Response Fallback

If the response is not valid JSON, the provider will fall back to text parsing mode:
- Returns moderate default scores (middle of each range)
- Sets recommendation to "review"
- Stores the entire response text in explanations

This ensures the provider continues to work even with non-JSON responses.

## Headers Configuration

Headers are specified as comma-separated `key:value` pairs:

```bash
export CUSTOM_LLM_HEADERS="Authorization:Bearer token123,Content-Type:application/json,X-Custom:value"
```

**Notes:**
- Spaces around keys and values are automatically trimmed
- If `Content-Type` is not specified, it defaults to `application/json`
- Headers can include authentication tokens, API keys, version info, etc.

## Error Handling

The provider handles various error scenarios:

| Error Type | Behavior |
|------------|----------|
| Endpoint not configured | Returns clear error message |
| Network timeout | Respects `CUSTOM_LLM_TIMEOUT` setting |
| HTTP 4xx/5xx errors | Returns error with status code and response body |
| Invalid JSON response | Falls back to text parsing |
| Invalid score ranges | Returns validation error |
| Template syntax error | Returns error explaining template issue |

## Integration Example

### In Code

```go
package main

import (
    "github.com/rayyacub/telos-idea-matrix/internal/llm"
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

func main() {
    // Create custom provider
    provider := llm.NewCustomProvider()

    // Check if configured
    if !provider.IsAvailable() {
        log.Fatal("Custom provider not configured")
    }

    // Perform analysis
    result, err := provider.Analyze(llm.AnalysisRequest{
        IdeaContent: "Build an AI-powered task manager",
        Telos:       myTelos,
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Score: %.2f\n", result.FinalScore)
    fmt.Printf("Recommendation: %s\n", result.Recommendation)
}
```

### In Fallback Chain

```go
// Add custom provider to fallback chain
providers := []llm.Provider{
    llm.NewOllamaProvider("http://localhost:11434", "llama2"),
    llm.NewCustomProvider(),  // Falls back to custom if Ollama fails
    llm.NewRuleBasedProvider(),
}

chain := llm.NewFallbackProvider(providers...)
```

## Testing Your Configuration

Create a simple test script to verify your configuration:

```bash
#!/bin/bash

# Set your configuration
export CUSTOM_LLM_ENDPOINT="http://localhost:8080/analyze"
export CUSTOM_LLM_NAME="My LLM"

# Run a test
go run cmd/cli/main.go analyze "Test idea: Build a simple web app"
```

## Common Use Cases

### 1. Local Development with Ollama Alternative

Running a different local model server:

```bash
export CUSTOM_LLM_ENDPOINT="http://localhost:5000/v1/analyze"
export CUSTOM_LLM_NAME="LocalAI"
```

### 2. Corporate/Enterprise LLM

Connecting to company-hosted LLM:

```bash
export CUSTOM_LLM_ENDPOINT="https://llm.corp.internal/api/v1/analyze"
export CUSTOM_LLM_HEADERS="Authorization:Bearer ${CORP_TOKEN},X-Department:engineering"
export CUSTOM_LLM_TIMEOUT="120"
```

### 3. Self-Hosted Model

Running your own fine-tuned model:

```bash
export CUSTOM_LLM_ENDPOINT="http://gpu-server:8000/infer"
export CUSTOM_LLM_NAME="Custom Fine-tuned Model"
export CUSTOM_LLM_PROMPT_TEMPLATE='{"text":"{{.IdeaContent}}","temperature":0.7}'
```

### 4. Cloud API with Rate Limiting

For APIs with rate limits, use a higher timeout:

```bash
export CUSTOM_LLM_ENDPOINT="https://api.service.com/v1/completions"
export CUSTOM_LLM_HEADERS="API-Key:${API_KEY}"
export CUSTOM_LLM_TIMEOUT="180"
```

## Troubleshooting

### Provider not available

**Issue:** `IsAvailable()` returns `false`

**Solution:** Ensure `CUSTOM_LLM_ENDPOINT` is set:
```bash
echo $CUSTOM_LLM_ENDPOINT
```

### Connection timeout

**Issue:** Requests timeout before completion

**Solution:** Increase timeout:
```bash
export CUSTOM_LLM_TIMEOUT="120"  # 2 minutes
```

### Invalid score ranges

**Issue:** `validation error: score out of range`

**Solution:** Ensure your API returns scores within valid ranges:
- Mission Alignment: 0-4.0
- Anti-Challenge: 0-3.5
- Strategic Fit: 0-2.5

### Template errors

**Issue:** `failed to build request body: invalid prompt template`

**Solution:** Check your template syntax. Use Go template syntax:
```bash
# Correct
export CUSTOM_LLM_PROMPT_TEMPLATE='{"text":"{{.IdeaContent}}"}'

# Incorrect (JavaScript-style interpolation)
export CUSTOM_LLM_PROMPT_TEMPLATE='{"text":"${idea}"}'
```

### HTTP errors

**Issue:** `API error (HTTP 401): Unauthorized`

**Solution:** Check your headers and authentication:
```bash
export CUSTOM_LLM_HEADERS="Authorization:Bearer correct-token"
```

## Security Considerations

1. **API Keys**: Store sensitive tokens in environment variables, not in code
2. **HTTPS**: Use HTTPS endpoints for production (`https://` not `http://`)
3. **Timeouts**: Set reasonable timeouts to prevent hanging requests
4. **Validation**: The provider validates all scores to prevent injection attacks
5. **Rate Limiting**: Consider implementing rate limiting at the API level

## Performance Tips

1. **Caching**: Use the `CachedProvider` wrapper for frequently analyzed ideas
2. **Timeouts**: Set appropriate timeouts based on your API's response time
3. **Batch Processing**: If your API supports it, consider batching requests
4. **Fallback Chain**: Place faster providers earlier in the chain

## Next Steps

- See `/home/user/brain-salad/go/internal/llm/custom_test.go` for detailed test examples
- Check `/home/user/brain-salad/go/internal/llm/provider.go` for the Provider interface
- Review the integration with fallback chains in your application
