# LLM Package

This package provides LLM integration for the Telos Idea Matrix, implementing the provider abstraction pattern with fallback chain support.

## Architecture

### Provider Interface

All LLM providers implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    IsAvailable() bool
    Analyze(req AnalysisRequest) (*AnalysisResult, error)
}
```

### Provider Implementations

1. **OllamaProvider** - Uses local Ollama for LLM analysis
2. **RuleBasedProvider** - Uses rule-based scoring engine (always available)
3. **FallbackProvider** - Chains multiple providers with automatic fallback

### Fallback Chain

The default fallback chain is:

```
Ollama → Claude API (Track 5B) → Rule-based
```

If Ollama is unavailable or fails, it automatically falls back to the next provider.

## Usage

### Basic Usage

```go
import (
    "github.com/rayyacub/telos-idea-matrix/internal/llm"
    "github.com/rayyacub/telos-idea-matrix/internal/telos"
)

// Load telos configuration
parser := telos.NewParser()
telosData, err := parser.ParseFile("telos.md")
if err != nil {
    log.Fatal(err)
}

// Create fallback provider chain
config := llm.DefaultProviderConfig()
provider := llm.CreateDefaultFallbackChain(config, telosData)

// Analyze an idea
req := llm.AnalysisRequest{
    IdeaContent: "Build an AI automation tool using Python",
    Telos:       telosData,
}

result, err := provider.Analyze(req)
if err != nil {
    log.Fatal(err)
}

// Use the result
fmt.Printf("Score: %.2f/10\n", result.FinalScore)
fmt.Printf("Recommendation: %s\n", result.Recommendation)
fmt.Printf("Provider: %s\n", result.Provider)
```

### Using Specific Providers

```go
// Use only Ollama
ollamaProvider := llm.NewOllamaProvider("http://localhost:11434", "llama2")
if ollamaProvider.IsAvailable() {
    result, err := ollamaProvider.Analyze(req)
    // ...
}

// Use only rule-based
ruleProvider := llm.NewRuleBasedProvider()
result, err := ruleProvider.Analyze(req)
// ...

// Custom fallback chain
customChain := llm.NewFallbackProvider(
    llm.NewOllamaProvider("http://localhost:11434", "mistral"),
    llm.NewRuleBasedProvider(),
)
result, err := customChain.Analyze(req)
// ...
```

## Configuration

### Provider Configuration

```go
config := llm.ProviderConfig{
    // Ollama settings
    OllamaBaseURL: "http://localhost:11434",
    OllamaModel:   "llama2",
    OllamaTimeout: 30, // seconds

    // Claude API settings (Track 5B)
    ClaudeAPIKey:  os.Getenv("CLAUDE_API_KEY"),
    ClaudeModel:   "claude-3-5-sonnet-20241022",
    ClaudeTimeout: 30,

    // Cache settings
    EnableCache: true,
    CacheTTL:    3600, // 1 hour
}
```

## Testing

### Unit Tests

```bash
go test ./internal/llm/... -v
```

### Integration Tests

Integration tests require Ollama to be running:

```bash
# Start Ollama
ollama serve &

# Run integration tests
go test ./internal/llm/... -v -tags=integration
```

### Coverage

```bash
go test ./internal/llm/... -cover
```

## Ollama Client

The `client` package provides a low-level HTTP client for Ollama:

```go
import "github.com/rayyacub/telos-idea-matrix/internal/llm/client"

// Create client
ollamaClient := client.NewOllamaClient("http://localhost:11434", 30*time.Second)

// Check health
ctx := context.Background()
if err := ollamaClient.HealthCheck(ctx); err != nil {
    log.Fatal("Ollama not available:", err)
}

// Generate text
resp, err := ollamaClient.Generate(ctx, client.GenerateRequest{
    Model:  "llama2",
    Prompt: "Hello, world!",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Response)

// List models
models, err := ollamaClient.ListModels(ctx)
fmt.Println("Available models:", models)
```

## Scoring Framework

The LLM analyzes ideas using a three-category framework:

### 1. Mission Alignment (0-4.0 points, 40%)
- **Domain Expertise** (0-1.2): Leverages existing skills?
- **AI Alignment** (0-1.5): How central is AI?
- **Execution Support** (0-0.8): Can deliver quickly?
- **Revenue Potential** (0-0.5): Clear path to revenue?

### 2. Anti-Challenge Patterns (0-3.5 points, 35%)
- **Avoid Context-Switching** (0-1.2): Uses current stack?
- **Rapid Prototyping** (0-1.0): Build MVP quickly?
- **Accountability** (0-0.8): External accountability?
- **Income Anxiety** (0-0.5): Quick revenue?

### 3. Strategic Fit (0-2.5 points, 25%)
- **Stack Compatibility** (0-1.0): Enables flow state?
- **Shipping Habit** (0-0.8): Reusable systems?
- **Public Accountability** (0-0.4): Validate quickly?
- **Revenue Testing** (0-0.3): Scalable model?

## Error Handling

The fallback chain automatically handles:
- Connection errors (Ollama not running)
- Timeout errors (slow responses)
- Model not found errors
- Invalid JSON responses

## Future Enhancements (Track 5C)

- [ ] Caching support
- [ ] Streaming responses
- [ ] Provider health monitoring
- [ ] Retry logic with exponential backoff
- [ ] Response quality metrics
