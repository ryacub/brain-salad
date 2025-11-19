# Metrics Package

Application metrics collection and reporting for the Telos Idea Matrix application.

## Features

- **Thread-safe** metrics collection using mutexes
- **Multiple metric types**: Counter, Gauge, Histogram
- **Histogram statistics**: min, max, mean, percentiles (p50, p95, p99)
- **Global collector** singleton pattern
- **Application-specific helpers** for common operations
- **HTTP endpoint** for metrics retrieval

## Metric Types

### Counter
Incremental values that only go up (e.g., total requests, total ideas created).

```go
metrics.RecordIdeaCreated()
metrics.RecordHTTPRequest("GET", "/api/v1/ideas", 200)
```

### Gauge
Current value that can go up or down (e.g., active connections, memory usage).

```go
metrics.RecordActiveConnections(10)
metrics.RecordMemoryUsage(1024 * 1024 * 100) // 100 MB
```

### Histogram
Distribution of values over time (e.g., response times, query durations).

```go
start := time.Now()
// ... do work ...
metrics.RecordScoringDuration(time.Since(start))
metrics.RecordDatabaseQueryDuration(time.Since(start))
```

## Usage

### Application Metrics

Pre-built functions for common application metrics:

```go
import "github.com/rayyacub/telos-idea-matrix/internal/metrics"

// Ideas
metrics.RecordIdeaCreated()
metrics.RecordIdeaUpdated()
metrics.RecordIdeaDeleted()

// Performance
metrics.RecordScoringDuration(duration)
metrics.RecordDatabaseQueryDuration(duration)
metrics.RecordAPIRequestDuration(duration)

// HTTP
metrics.RecordHTTPRequest("POST", "/api/v1/ideas", 201)

// System
metrics.RecordActiveConnections(5)
metrics.RecordMemoryUsage(bytes)
metrics.RecordGoroutineCount(count)
```

### Custom Metrics

```go
collector := metrics.GetGlobalCollector()

// Record counter
collector.RecordCounter("custom_counter", 1.0)

// Record gauge
collector.RecordGauge("custom_gauge", 42.0)

// Record histogram value
collector.RecordHistogram("custom_duration", 123.4)

// Get snapshot
snapshot := collector.GetSnapshot()
```

### Retrieve Metrics

```go
// Get all metrics
allMetrics := metrics.GetMetrics()

// Get histogram statistics
stats := collector.GetHistogramStats("scoring_duration_ms")
if stats != nil {
    fmt.Printf("Min: %f, Max: %f, Mean: %f\n", stats.Min, stats.Max, stats.Mean)
    fmt.Printf("P50: %f, P95: %f, P99: %f\n", stats.P50, stats.P95, stats.P99)
}

// Reset all metrics (useful for testing)
metrics.ResetMetrics()
```

## HTTP Endpoint

Access metrics via HTTP:

```bash
curl http://localhost:8080/metrics | jq
```

Response format:
```json
{
  "ideas_created_total": {
    "type": "counter",
    "value": 42,
    "count": 42,
    "timestamp": "2025-01-19T10:30:00Z"
  },
  "scoring_duration_ms": {
    "type": "histogram",
    "count": 100,
    "timestamp": "2025-01-19T10:30:00Z",
    "stats": {
      "count": 100,
      "min": 10.5,
      "max": 250.3,
      "avg": 85.7
    }
  },
  "active_connections": {
    "type": "gauge",
    "value": 5,
    "timestamp": "2025-01-19T10:30:00Z"
  }
}
```

## Performance

- Thread-safe using RWMutex
- O(1) counter and gauge operations
- O(n) histogram operations (sorted for percentiles)
- Minimal memory overhead
- Snapshot operation creates a copy (non-blocking)

## LLM Telemetry

Comprehensive telemetry for tracking LLM provider usage, performance, and costs.

### What Metrics Are Collected

**Per Provider (Ollama, Claude, OpenAI, Custom, Rule-based):**
- Total requests (success/failure counts)
- Request latency (average, P50, P95, P99)
- Token usage (input/output tokens)
- Estimated costs (for paid providers)
- Error breakdown by type (timeout, rate_limit, auth_error, etc.)

**Global Metrics:**
- Cache hit/miss rates
- Provider fallback events

### Privacy Guarantees

The metrics system only captures aggregated statistics. **No sensitive data is captured:**

❌ Never captured:
- Actual idea content
- API keys or credentials
- User-identifiable information
- Error messages containing PII

✅ Only captured:
- Aggregated counts
- Timing data
- Token counts (numbers only)
- Error types (categorized)

### Usage

#### Recording Metrics

```go
import "github.com/rayyacub/telos-idea-matrix/internal/metrics"

// Record a successful LLM request
start := time.Now()
// ... call LLM provider ...
duration := time.Since(start)
metrics.RecordLLMRequest("claude", true, duration)

// Record token usage (for cost estimation)
metrics.RecordLLMTokens("claude", 1500, 800) // input: 1500, output: 800

// Record errors
metrics.RecordLLMError("openai", "timeout")
metrics.RecordLLMError("claude", "rate_limit")

// Record cache events
metrics.RecordLLMCacheHit(true)  // cache hit
metrics.RecordLLMCacheHit(false) // cache miss

// Record fallback events
metrics.RecordLLMFallback("ollama", "claude") // fell back from ollama to claude
```

#### Error Types

Standard error classifications for consistent tracking:

- `timeout` - Request exceeded deadline
- `rate_limit` - Provider rate limit hit
- `auth_error` - Authentication/API key issues
- `network_error` - Connection failures
- `invalid_response` - Provider returned malformed data
- `provider_error` - Provider-side errors (5xx)
- `unknown` - Unclassified errors

#### Cost Estimation

```go
// Calculate cost for a single request
cost := metrics.CalculateCost("claude", 1000, 500) // 1000 input, 500 output tokens
fmt.Printf("Cost: %s\n", cost.FormatCost()) // e.g., "$0.0105"
fmt.Printf("Details: %s\n", cost.FormatDetailed())

// Estimate monthly cost based on daily usage
monthlyCost := metrics.EstimateMonthlyCost(1_000_000, 500_000, "claude")
fmt.Printf("Estimated monthly: $%.2f\n", monthlyCost)
```

#### Pricing (as of 2025)

Provider | Input Cost (per 1M tokens) | Output Cost (per 1M tokens)
---------|---------------------------|---------------------------
Claude (Sonnet 3.5) | $3.00 | $15.00
OpenAI (GPT-4) | $30.00 | $60.00
Ollama | $0.00 (local) | $0.00 (local)
Custom | Unknown | Unknown
Rule-based | $0.00 (no LLM) | $0.00 (no LLM)

### Viewing Metrics

#### CLI Command

```bash
# View all LLM metrics
tm analytics llm

# Export as JSON
tm analytics llm --json

# View metrics from last 24 hours
tm analytics llm --since 24h
```

#### Sample Output

```
LLM Provider Telemetry
====================================================================================================

Provider Summary:
----------------------------------------------------------------------------------------------------
Provider        Requests    Success    Failure   Success %   Avg Lat   Est. Cost
----------------------------------------------------------------------------------------------------
claude                42         40          2       95.2%      450ms      $0.105
ollama                15         15          0      100.0%      200ms          -
rule_based             8          8          0      100.0%       10ms          -

Provider: claude
----------------------------------------------------------------------------------------------------
  Requests:      Total: 42 | Success: 40 | Failure: 2 | Success Rate: 95.2%
  Latency (ms):  Avg: 450 | P50: 420 | P95: 580 | P99: 650
  Tokens:        Input: 35000 | Output: 17500 | Total: 52500
  Cost:          $0.105 USD
  Errors:        timeout: 1 | network_error: 1

Provider: ollama
----------------------------------------------------------------------------------------------------
  Requests:      Total: 15 | Success: 15 | Failure: 0 | Success Rate: 100.0%
  Latency (ms):  Avg: 200 | P50: 195 | P95: 220 | P99: 230

Cache Statistics:
----------------------------------------------------------------------------------------------------
  Hits:     12
  Misses:   53
  Hit Rate: 18.5%

Fallback Statistics:
----------------------------------------------------------------------------------------------------
  ollama → claude: 3 times

Cost Summary:
----------------------------------------------------------------------------------------------------
  Total Estimated Cost: $0.105 USD
```

### Implementation Details

All LLM providers are automatically instrumented:
- Ollama: Tracks requests and latency (no token counts available)
- Claude: Tracks requests, latency, tokens, and costs
- OpenAI: Tracks requests, latency, tokens, and costs
- Custom: Tracks requests, latency, and errors
- Rule-based: Tracks requests and latency (no costs)

The LLM manager automatically records fallback events when a primary provider fails and a backup is used.

### Testing

```bash
# Run all metrics tests including LLM tests
go test github.com/rayyacub/telos-idea-matrix/internal/metrics -v -cover

# Run only LLM metrics tests
go test github.com/rayyacub/telos-idea-matrix/internal/metrics -run LLM -v
```

## Testing

```bash
go test github.com/rayyacub/telos-idea-matrix/internal/metrics -v -cover
```

Coverage: 91.4% of statements
