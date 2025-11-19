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

## Testing

```bash
go test github.com/rayyacub/telos-idea-matrix/internal/metrics -v -cover
```

Coverage: 91.4% of statements
