# Track 5A: Ollama Client & Provider Abstraction

**Phase**: 5 - LLM Integration
**Estimated Time**: 10-12 hours
**Dependencies**: None (but creates types used by 5B, 5C)
**Can Run in Parallel**: Start first, then 5B/5C can run parallel

---

## Mission

You are implementing Ollama LLM client and provider abstraction for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation uses ollama-rs crate for LLM integration
- We need HTTP client for Ollama API, provider abstraction, and fallback chain
- Provider chain: Ollama → Claude API → rule-based scoring
- Must handle timeouts, connection errors, and model not found errors

## Reference Implementation

Review:
- `/home/user/brain-salad/src/commands/analyze_llm.rs`
- `/home/user/brain-salad/src/llm_fallback.rs`
- `/home/user/brain-salad/src/ai/`

## Your Task

Implement Ollama client and provider abstraction using strict TDD methodology.

**IMPORTANT**: Implement `types.go` FIRST (within 2 hours) and commit/push immediately. This unblocks tracks 5B and 5C which depend on these types.

## Directory Structure

Create files in `go/internal/llm/`:
- `types.go` - **DO THIS FIRST** - Shared types (AnalysisResult, Provider interface)
- `client/ollama.go` - Ollama HTTP client
- `client/ollama_test.go` - Ollama client tests
- `provider.go` - Provider interface and implementations
- `provider_test.go` - Provider tests
- `prompts.go` - Prompt templates

## TDD Workflow (RED → GREEN → REFACTOR)

### PRIORITY: Create types.go First (2 hours)

This must be completed first to unblock 5B and 5C:

#### Implement `go/internal/llm/types.go`:

```go
package llm

import "time"

type AnalysisRequest struct {
    IdeaContent string
    TelosPath   string
}

type AnalysisResult struct {
    Scores        ScoreBreakdown
    FinalScore    float64
    Recommendation string
    Explanations  map[string]string
    Provider      string
    Duration      time.Duration
}

type ScoreBreakdown struct {
    MissionAlignment  float64
    AntiChallenge     float64
    StrategicFit      float64
}

type Provider interface {
    Name() string
    IsAvailable() bool
    Analyze(req AnalysisRequest) (*AnalysisResult, error)
}
```

**COMMIT AND PUSH THIS IMMEDIATELY** after basic tests pass. This unblocks parallel work on 5B and 5C.

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/llm/client/ollama_test.go`:
- `TestOllamaClient_NewClient()`
- `TestOllamaClient_Generate_Success()`
- `TestOllamaClient_Generate_Timeout()`
- `TestOllamaClient_Generate_ConnectionError()`
- `TestOllamaClient_Generate_ModelNotFound()`
- `TestOllamaClient_ListModels()`
- `TestOllamaClient_HealthCheck()`

Create `go/internal/llm/provider_test.go`:
- `TestProvider_OllamaProvider_Analyze()`
- `TestProvider_FallbackChain()`
- `TestProvider_FallbackToRuleBased()`
- `TestProvider_ClaudeProvider_Stub()`

Run: `go test ./internal/llm/... -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Implement `go/internal/llm/client/ollama.go`:

```go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type OllamaClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}

type GenerateRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
}

type GenerateResponse struct {
    Model     string    `json:"model"`
    CreatedAt time.Time `json:"created_at"`
    Response  string    `json:"response"`
    Done      bool      `json:"done"`
}

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

func (c *OllamaClient) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
    req.Stream = false // Disable streaming for simplicity

    payload, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/api/generate", bytes.NewReader(payload))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("ollama error: status %d", resp.StatusCode)
    }

    var result GenerateResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &result, nil
}

func (c *OllamaClient) ListModels(ctx context.Context) ([]string, error) {
    // Implementation
}

func (c *OllamaClient) HealthCheck(ctx context.Context) error {
    // Simple ping to /api/tags
    httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
    if err != nil {
        return err
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
```

#### B. Implement `go/internal/llm/prompts.go`:

```go
package llm

import (
    "fmt"
    "io/ioutil"
)

func BuildAnalysisPrompt(ideaContent string, telosPath string) (string, error) {
    // Read telos file
    telosContent, err := ioutil.ReadFile(telosPath)
    if err != nil {
        return "", fmt.Errorf("read telos file: %w", err)
    }

    prompt := fmt.Sprintf(`You are an expert at evaluating ideas against personal goals and values.

TELOS (Personal Goals & Values):
%s

IDEA TO EVALUATE:
%s

TASK:
Analyze this idea and provide a detailed scoring breakdown:

1. Mission Alignment (0-4.0 points):
   - Domain Expertise (0-1.2)
   - AI Alignment (0-1.5)
   - Execution Support (0-0.8)
   - Revenue Potential (0-0.5)

2. Anti-Challenge Patterns (0-3.5 points):
   - Avoid Context-Switching (0-1.2)
   - Rapid Prototyping (0-1.0)
   - Accountability (0-0.8)
   - Income Anxiety (0-0.5)

3. Strategic Fit (0-2.5 points):
   - Stack Compatibility (0-1.0)
   - Shipping Habit (0-0.8)
   - Public Accountability (0-0.4)
   - Revenue Testing (0-0.3)

Respond with JSON in this exact format:
{
  "scores": {
    "mission_alignment": 2.5,
    "anti_challenge": 2.0,
    "strategic_fit": 1.5
  },
  "final_score": 6.0,
  "recommendation": "CONSIDER LATER",
  "explanations": {
    "mission_alignment": "explanation here",
    "anti_challenge": "explanation here",
    "strategic_fit": "explanation here"
  }
}
`, string(telosContent), ideaContent)

    return prompt, nil
}
```

#### C. Implement `go/internal/llm/provider.go`:

```go
package llm

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/rayyacub/telos-idea-matrix/internal/llm/client"
    "github.com/rayyacub/telos-idea-matrix/internal/scoring"
    "github.com/rayyacub/telos-idea-matrix/internal/telos"
)

type OllamaProvider struct {
    client *client.OllamaClient
    model  string
}

func NewOllamaProvider(baseURL string, model string) *OllamaProvider {
    if model == "" {
        model = "llama2"
    }

    return &OllamaProvider{
        client: client.NewOllamaClient(baseURL, 30*time.Second),
        model:  model,
    }
}

func (op *OllamaProvider) Name() string {
    return "ollama"
}

func (op *OllamaProvider) IsAvailable() bool {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    return op.client.HealthCheck(ctx) == nil
}

func (op *OllamaProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
    start := time.Now()

    // Build prompt
    prompt, err := BuildAnalysisPrompt(req.IdeaContent, req.TelosPath)
    if err != nil {
        return nil, fmt.Errorf("build prompt: %w", err)
    }

    // Generate analysis
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    resp, err := op.client.Generate(ctx, client.GenerateRequest{
        Model:  op.model,
        Prompt: prompt,
    })
    if err != nil {
        return nil, fmt.Errorf("generate: %w", err)
    }

    // Parse response
    var result AnalysisResult
    if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }

    result.Provider = op.Name()
    result.Duration = time.Since(start)

    return &result, nil
}

type FallbackProvider struct {
    providers []Provider
}

func NewFallbackProvider(providers ...Provider) *FallbackProvider {
    return &FallbackProvider{
        providers: providers,
    }
}

func (fp *FallbackProvider) Name() string {
    return "fallback"
}

func (fp *FallbackProvider) IsAvailable() bool {
    for _, p := range fp.providers {
        if p.IsAvailable() {
            return true
        }
    }
    return false
}

func (fp *FallbackProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
    var lastErr error

    for _, provider := range fp.providers {
        if !provider.IsAvailable() {
            continue
        }

        result, err := provider.Analyze(req)
        if err == nil {
            return result, nil
        }
        lastErr = err
    }

    return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// RuleBasedProvider uses the existing scoring engine as fallback
type RuleBasedProvider struct {
    engine *scoring.Engine
}

func NewRuleBasedProvider() *RuleBasedProvider {
    return &RuleBasedProvider{
        engine: scoring.NewEngine(),
    }
}

func (rbp *RuleBasedProvider) Name() string {
    return "rule_based"
}

func (rbp *RuleBasedProvider) IsAvailable() bool {
    return true // Always available
}

func (rbp *RuleBasedProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
    start := time.Now()

    // Parse telos
    telosData, err := telos.ParseFile(req.TelosPath)
    if err != nil {
        return nil, fmt.Errorf("parse telos: %w", err)
    }

    // Score using rule-based engine
    analysis := rbp.engine.Score(req.IdeaContent, telosData)

    result := &AnalysisResult{
        Scores: ScoreBreakdown{
            MissionAlignment: analysis.MissionScores.Total,
            AntiChallenge:    analysis.AntiChallengeScores.Total,
            StrategicFit:     analysis.StrategicScores.Total,
        },
        FinalScore:     analysis.FinalScore,
        Recommendation: analysis.GetRecommendation(),
        Explanations:   make(map[string]string),
        Provider:       rbp.Name(),
        Duration:       time.Since(start),
    }

    return result, nil
}
```

Run: `go test ./internal/llm/... -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add streaming support for real-time feedback
- Optimize prompt templates
- Add provider health monitoring
- Extract HTTP client configuration

## Success Criteria

- ✅ All tests pass with >85% coverage
- ✅ Works with Ollama running locally
- ✅ Proper timeout handling
- ✅ Graceful fallback on connection failure
- ✅ Rule-based provider always works

## Validation

```bash
# Unit tests
go test ./internal/llm/... -v -cover

# Integration test (requires Ollama running)
ollama serve &
sleep 2
go test ./internal/llm/... -v -tags=integration

# Manual test
go run ./cmd/cli/main.go analyze --ai "Build a Python automation tool"
```

## Deliverables

- `go/internal/llm/types.go` **[PRIORITY - Complete within 2 hours]**
- `go/internal/llm/client/ollama.go`
- `go/internal/llm/client/ollama_test.go`
- `go/internal/llm/provider.go`
- `go/internal/llm/provider_test.go`
- `go/internal/llm/prompts.go`
