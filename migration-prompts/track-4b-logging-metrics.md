# Track 4B: Structured Logging & Metrics

**Phase**: 4 - Production Infrastructure
**Estimated Time**: 8-10 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 4A, 4C)

---

## Mission

You are implementing structured logging and metrics collection for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation uses tracing/tracing-subscriber for structured logging (`src/logging.rs`)
- The Rust implementation tracks metrics in `src/metrics.rs`
- We need JSON logging, log levels, file rotation, and basic metrics collection
- Use zerolog (`github.com/rs/zerolog`) for structured logging

## Reference Implementation

Review:
- `/home/user/brain-salad/src/logging.rs` - Structured logging setup
- `/home/user/brain-salad/src/metrics.rs` - Metrics collection

## Your Task

Implement structured logging and metrics using strict TDD methodology.

## Directory Structure

Create files in `go/internal/logging/` and `go/internal/metrics/`:
- `logging/logger.go` - Logger setup and configuration
- `logging/middleware.go` - HTTP request logging middleware
- `logging/logger_test.go` - Logger tests
- `metrics/collector.go` - Metrics collection
- `metrics/metrics.go` - Application-specific metrics
- `metrics/collector_test.go` - Metrics tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/logging/logger_test.go`:
- `TestLogger_NewLogger()`
- `TestLogger_Levels()` - debug, info, warn, error
- `TestLogger_StructuredFields()`
- `TestLogger_FileOutput()`
- `TestLogger_JsonFormat()`
- `TestLogger_ContextIntegration()`

Create `go/internal/metrics/collector_test.go`:
- `TestMetricsCollector_RecordCounter()`
- `TestMetricsCollector_RecordGauge()`
- `TestMetricsCollector_RecordHistogram()`
- `TestMetricsCollector_GetSnapshot()`
- `TestMetricsCollector_Reset()`

Run: `go test ./internal/logging ./internal/metrics -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Install dependencies:

```bash
cd go
go get github.com/rs/zerolog
go get github.com/rs/zerolog/log
```

#### B. Implement `go/internal/logging/logger.go`:

```go
package logging

import (
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
    Level      string
    Format     string // "json" or "console"
    OutputPath string // file path or "stdout"
    MaxSizeMB  int
    MaxBackups int
    MaxAgeDays int
}

func NewLogger(cfg Config) zerolog.Logger {
    // Set log level
    level := parseLogLevel(cfg.Level)
    zerolog.SetGlobalLevel(level)

    // Configure output
    var output io.Writer
    if cfg.OutputPath == "stdout" || cfg.OutputPath == "" {
        output = os.Stdout
    } else {
        // File output with rotation
        output = &lumberjack.Logger{
            Filename:   cfg.OutputPath,
            MaxSize:    cfg.MaxSizeMB,
            MaxBackups: cfg.MaxBackups,
            MaxAge:     cfg.MaxAgeDays,
            Compress:   true,
        }
    }

    // Format
    if cfg.Format == "console" {
        output = zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339}
    }

    logger := zerolog.New(output).With().Timestamp().Caller().Logger()

    // Set global logger
    log.Logger = logger

    return logger
}

func parseLogLevel(level string) zerolog.Level {
    switch level {
    case "debug":
        return zerolog.DebugLevel
    case "info":
        return zerolog.InfoLevel
    case "warn":
        return zerolog.WarnLevel
    case "error":
        return zerolog.ErrorLevel
    default:
        return zerolog.InfoLevel
    }
}
```

#### C. Implement `go/internal/logging/middleware.go`:

```go
package logging

import (
    "net/http"
    "time"

    "github.com/rs/zerolog/log"
)

func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Log request
        log.Info().
            Str("method", r.Method).
            Str("path", r.URL.Path).
            Str("remote_addr", r.RemoteAddr).
            Msg("request started")

        // Handle request
        next.ServeHTTP(wrapped, r)

        // Log response
        duration := time.Since(start)
        log.Info().
            Str("method", r.Method).
            Str("path", r.URL.Path).
            Int("status", wrapped.statusCode).
            Dur("duration_ms", duration).
            Msg("request completed")
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

#### D. Implement `go/internal/metrics/collector.go`:

```go
package metrics

import (
    "sync"
    "time"
)

type MetricType string

const (
    Counter   MetricType = "counter"
    Gauge     MetricType = "gauge"
    Histogram MetricType = "histogram"
)

type Metric struct {
    Name      string
    Type      MetricType
    Value     float64
    Count     int64
    Timestamp time.Time
}

type Collector struct {
    metrics map[string]*Metric
    mu      sync.RWMutex
}

func NewCollector() *Collector {
    return &Collector{
        metrics: make(map[string]*Metric),
    }
}

func (c *Collector) RecordCounter(name string, value float64) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if m, exists := c.metrics[name]; exists {
        m.Value += value
        m.Count++
        m.Timestamp = time.Now()
    } else {
        c.metrics[name] = &Metric{
            Name:      name,
            Type:      Counter,
            Value:     value,
            Count:     1,
            Timestamp: time.Now(),
        }
    }
}

func (c *Collector) RecordGauge(name string, value float64) {
    // Implementation
}

func (c *Collector) RecordHistogram(name string, value float64) {
    // Implementation
}

func (c *Collector) GetSnapshot() map[string]Metric {
    // Implementation
}
```

#### E. Implement `go/internal/metrics/metrics.go` (application metrics):

```go
package metrics

import "time"

var (
    globalCollector *Collector
)

func init() {
    globalCollector = NewCollector()
}

// Application-specific metrics
func RecordIdeaCreated() {
    globalCollector.RecordCounter("ideas_created_total", 1)
}

func RecordIdeaUpdated() {
    globalCollector.RecordCounter("ideas_updated_total", 1)
}

func RecordScoringDuration(duration time.Duration) {
    globalCollector.RecordHistogram("scoring_duration_ms", float64(duration.Milliseconds()))
}

func RecordDatabaseQueryDuration(duration time.Duration) {
    globalCollector.RecordHistogram("database_query_duration_ms", float64(duration.Milliseconds()))
}

func GetMetrics() map[string]Metric {
    return globalCollector.GetSnapshot()
}
```

Run: `go test ./internal/logging ./internal/metrics -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add log sampling for high-volume logs
- Optimize metrics storage (use atomic operations)
- Add configurable metrics retention
- Extract common patterns

## Integration

1. Replace all `fmt.Println` with `log.Info()` throughout codebase
2. Add logging middleware to API server
3. Add `/metrics` endpoint to API server
4. Add `tm logs tail` CLI command (tail log file)

## Success Criteria

- ✅ All tests pass with >85% coverage
- ✅ JSON log output works
- ✅ Log rotation works (max 10MB per file, 7 days retention)
- ✅ Metrics endpoint returns data
- ✅ No performance impact (<1ms per log statement)

## Validation

```bash
# Test logging
go test ./internal/logging -v -cover

# Test metrics
go test ./internal/metrics -v -cover

# Integration test
go run ./cmd/web/main.go &
sleep 2
curl http://localhost:8080/metrics | jq

# Check log file
cat logs/telos-matrix.log | jq
```

## Deliverables

- `go/internal/logging/logger.go`
- `go/internal/logging/middleware.go`
- `go/internal/logging/logger_test.go`
- `go/internal/metrics/collector.go`
- `go/internal/metrics/metrics.go`
- `go/internal/metrics/collector_test.go`
- `go/internal/api/handlers.go` (add GET /metrics endpoint)
