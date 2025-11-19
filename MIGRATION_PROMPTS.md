# Migration Prompts for Go Implementation

This document contains detailed, self-contained prompts for each track in the Go migration game plan. Each prompt can be used to launch a subagent or assign to a developer.

---

## PHASE 4: PRODUCTION INFRASTRUCTURE

### Track 4A: Health Monitoring System

**Subagent Prompt:**

```
You are implementing a production-ready health monitoring system for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

CONTEXT:
- The Rust implementation has a comprehensive health monitoring system in src/health.rs
- We need to replicate this functionality in Go with proper test coverage (>90%)
- The health system should monitor database, memory, disk space, and application uptime

REFERENCE IMPLEMENTATION:
Review the Rust implementation at /home/user/brain-salad/src/health.rs to understand:
- HealthMonitor struct with check registration
- HealthChecker trait/interface pattern
- Health state aggregation (Healthy/Degraded/Unhealthy)
- Built-in checkers (database, memory, disk)
- Uptime tracking

YOUR TASK:
Implement the health monitoring system in Go using strict TDD methodology.

DIRECTORY STRUCTURE:
Create files in go/internal/health/:
- monitor.go - Core health monitoring
- checkers.go - Built-in health checkers
- uptime.go - Uptime tracking
- monitor_test.go - Comprehensive tests
- checkers_test.go - Checker tests

TDD WORKFLOW (RED → GREEN → REFACTOR):

STEP 1 - RED PHASE (Write Failing Tests First):
Create go/internal/health/monitor_test.go with tests:
- TestHealthMonitor_NewMonitor()
- TestHealthMonitor_AddCheck()
- TestHealthMonitor_RunAllChecks_Healthy()
- TestHealthMonitor_RunAllChecks_Degraded()
- TestHealthMonitor_RunAllChecks_Unhealthy()
- TestHealthCheckState_Aggregation()
- TestHealthMonitor_ConcurrentChecks()

Create go/internal/health/checkers_test.go with tests:
- TestDatabaseHealthChecker()
- TestMemoryHealthChecker()
- TestDiskSpaceHealthChecker()
- TestHealthChecker_Timeout()
- TestHealthChecker_Error()

Run: go test ./internal/health -v
Expected: ALL TESTS FAIL (code doesn't exist yet)

STEP 2 - GREEN PHASE (Implement Minimal Code):

A. Implement go/internal/health/monitor.go:
```go
package health

import (
    "context"
    "sync"
    "time"
)

type HealthState string

const (
    Healthy   HealthState = "healthy"
    Degraded  HealthState = "degraded"
    Unhealthy HealthState = "unhealthy"
)

type HealthCheckState string

const (
    Ok      HealthCheckState = "ok"
    Warning HealthCheckState = "warning"
    Error   HealthCheckState = "error"
)

type HealthCheckResult struct {
    Name      string           `json:"name"`
    Status    HealthCheckState `json:"status"`
    Message   string           `json:"message,omitempty"`
    Duration  time.Duration    `json:"duration"`
    Timestamp time.Time        `json:"timestamp"`
}

type HealthStatus struct {
    Status    HealthState         `json:"status"`
    Timestamp time.Time           `json:"timestamp"`
    Checks    []HealthCheckResult `json:"checks"`
    Uptime    time.Duration       `json:"uptime"`
    Version   string              `json:"version"`
}

type HealthChecker interface {
    Name() string
    Check(ctx context.Context) HealthCheckResult
}

type HealthMonitor struct {
    startTime time.Time
    checkers  []HealthChecker
    mu        sync.RWMutex
}

func NewHealthMonitor() *HealthMonitor {
    return &HealthMonitor{
        startTime: time.Now(),
        checkers:  make([]HealthChecker, 0),
    }
}

func (hm *HealthMonitor) AddCheck(checker HealthChecker) {
    hm.mu.Lock()
    defer hm.mu.Unlock()
    hm.checkers = append(hm.checkers, checker)
}

func (hm *HealthMonitor) RunAllChecks(ctx context.Context) HealthStatus {
    // Implementation here
}

func (hm *HealthMonitor) IsHealthy(ctx context.Context) bool {
    status := hm.RunAllChecks(ctx)
    return status.Status == Healthy
}
```

B. Implement go/internal/health/checkers.go:
```go
package health

import (
    "context"
    "database/sql"
    "fmt"
    "runtime"
    "syscall"
    "time"
)

type DatabaseHealthChecker struct {
    db *sql.DB
}

func NewDatabaseHealthChecker(db *sql.DB) *DatabaseHealthChecker {
    return &DatabaseHealthChecker{db: db}
}

func (d *DatabaseHealthChecker) Name() string {
    return "database"
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) HealthCheckResult {
    // Implementation
}

type MemoryHealthChecker struct {
    thresholdMB float64
}

func NewMemoryHealthChecker(thresholdMB float64) *MemoryHealthChecker {
    return &MemoryHealthChecker{thresholdMB: thresholdMB}
}

func (m *MemoryHealthChecker) Name() string {
    return "memory"
}

func (m *MemoryHealthChecker) Check(ctx context.Context) HealthCheckResult {
    // Implementation using runtime.MemStats
}

type DiskSpaceHealthChecker struct {
    path          string
    thresholdMB   uint64
}

func NewDiskSpaceHealthChecker(path string, thresholdMB uint64) *DiskSpaceHealthChecker {
    return &DiskSpaceHealthChecker{path: path, thresholdMB: thresholdMB}
}

func (d *DiskSpaceHealthChecker) Name() string {
    return "disk_space"
}

func (d *DiskSpaceHealthChecker) Check(ctx context.Context) HealthCheckResult {
    // Implementation using syscall.Statfs
}
```

Run: go test ./internal/health -v
Expected: ALL TESTS PASS

STEP 3 - REFACTOR PHASE:
- Extract common patterns into helper functions
- Optimize concurrent health check execution
- Add configurable timeouts for each checker
- Improve error messages

INTEGRATION:
1. Add health endpoint to API server (go/internal/api/handlers.go):
   - GET /health → returns HealthStatus JSON
2. Add CLI command (go/internal/cli/health.go):
   - `tm health` → displays health status in terminal
3. Wire into graceful shutdown (check health before shutdown)

SUCCESS CRITERIA:
- ✅ All tests pass: go test ./internal/health -v
- ✅ Test coverage >90%: go test ./internal/health -cover
- ✅ Health checks complete in <100ms
- ✅ Proper state aggregation (any Error = Unhealthy, any Warning = Degraded)
- ✅ GET /health endpoint works
- ✅ tm health command displays status

VALIDATION:
```bash
# Unit tests
go test ./internal/health -v -cover

# Integration test
go run ./cmd/web/main.go &
sleep 2
curl http://localhost:8080/health | jq
# Should return: {"status":"healthy",...}

# CLI test
go run ./cmd/cli/main.go health
# Should display health status with colored output
```

ESTIMATED TIME: 8-10 hours

DELIVERABLES:
- go/internal/health/monitor.go
- go/internal/health/checkers.go
- go/internal/health/monitor_test.go
- go/internal/health/checkers_test.go
- go/internal/api/handlers.go (add GET /health endpoint)
- go/internal/cli/health.go (add tm health command)
```

---

### Track 4B: Structured Logging & Metrics

**Subagent Prompt:**

```
You are implementing structured logging and metrics collection for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

CONTEXT:
- The Rust implementation uses tracing/tracing-subscriber for structured logging (src/logging.rs)
- The Rust implementation tracks metrics in src/metrics.rs
- We need JSON logging, log levels, file rotation, and basic metrics collection
- Use zerolog (github.com/rs/zerolog) for structured logging

REFERENCE IMPLEMENTATION:
Review:
- /home/user/brain-salad/src/logging.rs - Structured logging setup
- /home/user/brain-salad/src/metrics.rs - Metrics collection

YOUR TASK:
Implement structured logging and metrics using strict TDD methodology.

DIRECTORY STRUCTURE:
Create files in go/internal/logging/ and go/internal/metrics/:
- logging/logger.go - Logger setup and configuration
- logging/middleware.go - HTTP request logging middleware
- logging/logger_test.go - Logger tests
- metrics/collector.go - Metrics collection
- metrics/metrics.go - Application-specific metrics
- metrics/collector_test.go - Metrics tests

TDD WORKFLOW (RED → GREEN → REFACTOR):

STEP 1 - RED PHASE (Write Failing Tests):

Create go/internal/logging/logger_test.go:
- TestLogger_NewLogger()
- TestLogger_Levels() - debug, info, warn, error
- TestLogger_StructuredFields()
- TestLogger_FileOutput()
- TestLogger_JsonFormat()
- TestLogger_ContextIntegration()

Create go/internal/metrics/collector_test.go:
- TestMetricsCollector_RecordCounter()
- TestMetricsCollector_RecordGauge()
- TestMetricsCollector_RecordHistogram()
- TestMetricsCollector_GetSnapshot()
- TestMetricsCollector_Reset()

Run: go test ./internal/logging ./internal/metrics -v
Expected: ALL TESTS FAIL

STEP 2 - GREEN PHASE (Implement):

A. Install dependencies:
```bash
cd go
go get github.com/rs/zerolog
go get github.com/rs/zerolog/log
```

B. Implement go/internal/logging/logger.go:
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

C. Implement go/internal/logging/middleware.go:
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

D. Implement go/internal/metrics/collector.go:
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

E. Implement go/internal/metrics/metrics.go (application metrics):
```go
package metrics

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

Run: go test ./internal/logging ./internal/metrics -v
Expected: ALL TESTS PASS

STEP 3 - REFACTOR PHASE:
- Add log sampling for high-volume logs
- Optimize metrics storage (use atomic operations)
- Add configurable metrics retention
- Extract common patterns

INTEGRATION:
1. Replace all fmt.Println with log.Info() throughout codebase
2. Add logging middleware to API server
3. Add /metrics endpoint to API server
4. Add `tm logs tail` CLI command (tail log file)

SUCCESS CRITERIA:
- ✅ All tests pass with >85% coverage
- ✅ JSON log output works
- ✅ Log rotation works (max 10MB per file, 7 days retention)
- ✅ Metrics endpoint returns data
- ✅ No performance impact (<1ms per log statement)

VALIDATION:
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

ESTIMATED TIME: 8-10 hours

DELIVERABLES:
- go/internal/logging/logger.go
- go/internal/logging/middleware.go
- go/internal/logging/logger_test.go
- go/internal/metrics/collector.go
- go/internal/metrics/metrics.go
- go/internal/metrics/collector_test.go
- go/internal/api/handlers.go (add GET /metrics endpoint)
```

---

### Track 4C: Background Task Manager

**Subagent Prompt:**

```
You are implementing a background task manager for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

CONTEXT:
- The Rust implementation has task supervision in src/background_tasks.rs
- We need graceful shutdown, task lifecycle management, and scheduled tasks
- Must handle SIGTERM/SIGINT signals properly
- No goroutine leaks allowed

REFERENCE IMPLEMENTATION:
Review /home/user/brain-salad/src/background_tasks.rs

YOUR TASK:
Implement background task manager using strict TDD methodology.

DIRECTORY STRUCTURE:
Create files in go/internal/tasks/:
- manager.go - Task supervision and lifecycle
- task.go - Task interface and helpers
- scheduler.go - Scheduled task execution
- manager_test.go - Comprehensive tests

TDD WORKFLOW (RED → GREEN → REFACTOR):

STEP 1 - RED PHASE (Write Failing Tests):

Create go/internal/tasks/manager_test.go:
- TestTaskManager_NewManager()
- TestTaskManager_SpawnTask()
- TestTaskManager_GracefulShutdown()
- TestTaskManager_TaskCompletion()
- TestTaskManager_TaskFailure()
- TestTaskManager_ConcurrentTasks()
- TestTaskManager_ShutdownTimeout()
- TestTaskManager_NoGoroutineLeaks()

Run: go test ./internal/tasks -v
Expected: ALL TESTS FAIL

STEP 2 - GREEN PHASE (Implement):

A. Implement go/internal/tasks/task.go:
```go
package tasks

import (
    "context"
    "time"
)

type Task interface {
    Name() string
    Run(ctx context.Context) error
    Timeout() time.Duration
}

type TaskResult struct {
    Name      string
    Error     error
    Duration  time.Duration
    StartedAt time.Time
    EndedAt   time.Time
}

type BaseTask struct {
    name    string
    timeout time.Duration
}

func NewBaseTask(name string, timeout time.Duration) *BaseTask {
    return &BaseTask{
        name:    name,
        timeout: timeout,
    }
}

func (bt *BaseTask) Name() string {
    return bt.name
}

func (bt *BaseTask) Timeout() time.Duration {
    return bt.timeout
}
```

B. Implement go/internal/tasks/manager.go:
```go
package tasks

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/rs/zerolog/log"
)

type TaskManager struct {
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
    tasks      []Task
    results    []TaskResult
    mu         sync.Mutex
    shutdownCh chan struct{}
}

func NewTaskManager() *TaskManager {
    ctx, cancel := context.WithCancel(context.Background())

    return &TaskManager{
        ctx:        ctx,
        cancel:     cancel,
        tasks:      make([]Task, 0),
        results:    make([]TaskResult, 0),
        shutdownCh: make(chan struct{}),
    }
}

func (tm *TaskManager) Spawn(task Task) {
    tm.mu.Lock()
    tm.tasks = append(tm.tasks, task)
    tm.mu.Unlock()

    tm.wg.Add(1)
    go tm.runTask(task)
}

func (tm *TaskManager) runTask(task Task) {
    defer tm.wg.Done()

    startTime := time.Now()

    log.Info().Str("task", task.Name()).Msg("task started")

    // Create task context with timeout
    taskCtx, cancel := context.WithTimeout(tm.ctx, task.Timeout())
    defer cancel()

    // Run task
    err := task.Run(taskCtx)

    duration := time.Since(startTime)

    // Record result
    result := TaskResult{
        Name:      task.Name(),
        Error:     err,
        Duration:  duration,
        StartedAt: startTime,
        EndedAt:   time.Now(),
    }

    tm.mu.Lock()
    tm.results = append(tm.results, result)
    tm.mu.Unlock()

    if err != nil {
        log.Error().
            Err(err).
            Str("task", task.Name()).
            Dur("duration", duration).
            Msg("task failed")
    } else {
        log.Info().
            Str("task", task.Name()).
            Dur("duration", duration).
            Msg("task completed")
    }
}

func (tm *TaskManager) Shutdown(timeout time.Duration) error {
    log.Info().Msg("initiating graceful shutdown")

    // Cancel all task contexts
    tm.cancel()

    // Wait for tasks to complete with timeout
    done := make(chan struct{})
    go func() {
        tm.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        log.Info().Msg("all tasks completed gracefully")
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("shutdown timeout exceeded")
    }
}

func (tm *TaskManager) ListenForShutdown() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

    go func() {
        sig := <-sigCh
        log.Info().Str("signal", sig.String()).Msg("received shutdown signal")

        if err := tm.Shutdown(5 * time.Second); err != nil {
            log.Error().Err(err).Msg("shutdown error")
            os.Exit(1)
        }

        close(tm.shutdownCh)
    }()
}

func (tm *TaskManager) Wait() {
    <-tm.shutdownCh
}

func (tm *TaskManager) GetResults() []TaskResult {
    tm.mu.Lock()
    defer tm.mu.Unlock()

    results := make([]TaskResult, len(tm.results))
    copy(results, tm.results)
    return results
}
```

C. Implement go/internal/tasks/scheduler.go:
```go
package tasks

import (
    "context"
    "time"
)

type ScheduledTask struct {
    BaseTask
    interval time.Duration
    runFunc  func(ctx context.Context) error
}

func NewScheduledTask(name string, interval time.Duration, runFunc func(ctx context.Context) error) *ScheduledTask {
    return &ScheduledTask{
        BaseTask: BaseTask{
            name:    name,
            timeout: 5 * time.Minute,
        },
        interval: interval,
        runFunc:  runFunc,
    }
}

func (st *ScheduledTask) Run(ctx context.Context) error {
    ticker := time.NewTicker(st.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := st.runFunc(ctx); err != nil {
                // Log error but continue running
                continue
            }
        }
    }
}
```

Run: go test ./internal/tasks -v
Expected: ALL TESTS PASS

STEP 3 - REFACTOR PHASE:
- Add task priority levels
- Add task retry policies
- Extract signal handling utilities
- Optimize goroutine management

INTEGRATION:
1. Wire into API server main.go:
   - Create TaskManager on startup
   - Add database cleanup scheduled task
   - Add metrics collection scheduled task
   - Call Shutdown on server shutdown
2. Add `tm tasks list` CLI command

SUCCESS CRITERIA:
- ✅ All tests pass with >85% coverage
- ✅ Graceful shutdown in <5 seconds
- ✅ Tasks properly canceled on shutdown
- ✅ No goroutine leaks (verified with -race flag)
- ✅ SIGTERM/SIGINT handling works

VALIDATION:
```bash
# Unit tests
go test ./internal/tasks -v -cover -race

# Integration test
go run ./cmd/web/main.go &
PID=$!
sleep 5
kill -TERM $PID
# Should shut down gracefully within 5 seconds
```

ESTIMATED TIME: 8-12 hours

DELIVERABLES:
- go/internal/tasks/manager.go
- go/internal/tasks/task.go
- go/internal/tasks/scheduler.go
- go/internal/tasks/manager_test.go
- go/cmd/web/main.go (integrate TaskManager)
```

---

## PHASE 5: LLM INTEGRATION

### Track 5A: Ollama Client & Provider Abstraction

**Subagent Prompt:**

```
You are implementing Ollama LLM client and provider abstraction for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

CONTEXT:
- The Rust implementation uses ollama-rs crate for LLM integration
- We need HTTP client for Ollama API, provider abstraction, and fallback chain
- Provider chain: Ollama → Claude API → rule-based scoring
- Must handle timeouts, connection errors, and model not found errors

REFERENCE IMPLEMENTATION:
Review:
- /home/user/brain-salad/src/commands/analyze_llm.rs
- /home/user/brain-salad/src/llm_fallback.rs
- /home/user/brain-salad/src/ai/

YOUR TASK:
Implement Ollama client and provider abstraction using strict TDD methodology.

DIRECTORY STRUCTURE:
Create files in go/internal/llm/:
- client/ollama.go - Ollama HTTP client
- client/ollama_test.go - Ollama client tests
- provider.go - Provider interface and implementations
- provider_test.go - Provider tests
- prompts.go - Prompt templates
- types.go - Shared types

TDD WORKFLOW (RED → GREEN → REFACTOR):

STEP 1 - RED PHASE (Write Failing Tests):

Create go/internal/llm/client/ollama_test.go:
- TestOllamaClient_NewClient()
- TestOllamaClient_Generate_Success()
- TestOllamaClient_Generate_Timeout()
- TestOllamaClient_Generate_ConnectionError()
- TestOllamaClient_Generate_ModelNotFound()
- TestOllamaClient_ListModels()
- TestOllamaClient_HealthCheck()

Create go/internal/llm/provider_test.go:
- TestProvider_OllamaProvider_Analyze()
- TestProvider_FallbackChain()
- TestProvider_FallbackToRuleBased()
- TestProvider_ClaudeProvider_Stub()

Run: go test ./internal/llm/... -v
Expected: ALL TESTS FAIL

STEP 2 - GREEN PHASE (Implement):

A. Implement go/internal/llm/types.go:
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

B. Implement go/internal/llm/client/ollama.go:
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

C. Implement go/internal/llm/prompts.go:
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

D. Implement go/internal/llm/provider.go:
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

Run: go test ./internal/llm/... -v
Expected: ALL TESTS PASS

STEP 3 - REFACTOR PHASE:
- Add streaming support for real-time feedback
- Optimize prompt templates
- Add provider health monitoring
- Extract HTTP client configuration

SUCCESS CRITERIA:
- ✅ All tests pass with >85% coverage
- ✅ Works with Ollama running locally
- ✅ Proper timeout handling
- ✅ Graceful fallback on connection failure
- ✅ Rule-based provider always works

VALIDATION:
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

ESTIMATED TIME: 10-12 hours

DELIVERABLES:
- go/internal/llm/client/ollama.go
- go/internal/llm/client/ollama_test.go
- go/internal/llm/provider.go
- go/internal/llm/provider_test.go
- go/internal/llm/prompts.go
- go/internal/llm/types.go
```

---

### Track 5B: Semantic Cache System

**Subagent Prompt:**

```
You are implementing a semantic similarity-based cache system for LLM responses in the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

CONTEXT:
- The Rust implementation has semantic caching in src/llm_cache.rs
- Cache should match similar ideas using Jaccard similarity (>0.85 threshold)
- 24-hour TTL, max 1000 entries per type, LRU eviction
- Cache hits should return in <5ms

REFERENCE IMPLEMENTATION:
Review /home/user/brain-salad/src/llm_cache.rs for:
- Similarity threshold (0.85)
- TTL (24 hours)
- Max cache size (1000)
- Hit count tracking
- Normalization logic

YOUR TASK:
Implement semantic cache system using strict TDD methodology.

DIRECTORY STRUCTURE:
Create files in go/internal/llm/cache/:
- cache.go - Core cache implementation
- similarity.go - Text similarity algorithms
- stats.go - Cache statistics
- cache_test.go - Comprehensive tests

TDD WORKFLOW (RED → GREEN → REFACTOR):

STEP 1 - RED PHASE (Write Failing Tests):

Create go/internal/llm/cache/cache_test.go:
- TestCache_NewCache()
- TestCache_StoreAndRetrieve()
- TestCache_SimilarityMatching_AboveThreshold()
- TestCache_SimilarityMatching_BelowThreshold()
- TestCache_TTL_Expiration()
- TestCache_MaxSize_LRU_Eviction()
- TestCache_HitCount_Tracking()
- TestCache_ConcurrentAccess()
- TestCache_Stats()

Create go/internal/llm/cache/similarity_test.go:
- TestJaccardSimilarity()
- TestNormalizeText()
- TestTokenize()

Run: go test ./internal/llm/cache -v
Expected: ALL TESTS FAIL

STEP 2 - GREEN PHASE (Implement):

A. Implement go/internal/llm/cache/similarity.go:
```go
package cache

import (
    "regexp"
    "strings"
    "unicode"
)

var (
    // Common English stopwords to filter out
    stopwords = map[string]bool{
        "a": true, "an": true, "and": true, "are": true, "as": true,
        "at": true, "be": true, "by": true, "for": true, "from": true,
        "has": true, "he": true, "in": true, "is": true, "it": true,
        "its": true, "of": true, "on": true, "that": true, "the": true,
        "to": true, "was": true, "will": true, "with": true,
    }

    nonAlphanumeric = regexp.MustCompile(`[^a-z0-9\s]+`)
    multipleSpaces  = regexp.MustCompile(`\s+`)
)

// NormalizeText canonicalizes text for similarity comparison
func NormalizeText(text string) string {
    // Lowercase
    text = strings.ToLower(text)

    // Remove non-alphanumeric characters
    text = nonAlphanumeric.ReplaceAllString(text, " ")

    // Collapse multiple spaces
    text = multipleSpaces.ReplaceAllString(text, " ")

    // Trim
    text = strings.TrimSpace(text)

    return text
}

// Tokenize splits text into words and removes stopwords
func Tokenize(text string) []string {
    normalized := NormalizeText(text)
    words := strings.Fields(normalized)

    // Filter stopwords
    filtered := make([]string, 0, len(words))
    for _, word := range words {
        if !stopwords[word] && len(word) > 1 {
            filtered = append(filtered, word)
        }
    }

    return filtered
}

// JaccardSimilarity computes Jaccard similarity between two texts
// Returns value between 0.0 (no similarity) and 1.0 (identical)
func JaccardSimilarity(text1, text2 string) float64 {
    tokens1 := Tokenize(text1)
    tokens2 := Tokenize(text2)

    if len(tokens1) == 0 && len(tokens2) == 0 {
        return 1.0
    }
    if len(tokens1) == 0 || len(tokens2) == 0 {
        return 0.0
    }

    // Create sets
    set1 := make(map[string]bool)
    for _, token := range tokens1 {
        set1[token] = true
    }

    set2 := make(map[string]bool)
    for _, token := range tokens2 {
        set2[token] = true
    }

    // Compute intersection
    intersection := 0
    for token := range set1 {
        if set2[token] {
            intersection++
        }
    }

    // Compute union
    union := len(set1) + len(set2) - intersection

    if union == 0 {
        return 0.0
    }

    return float64(intersection) / float64(union)
}
```

B. Implement go/internal/llm/cache/cache.go:
```go
package cache

import (
    "container/list"
    "sync"
    "time"

    "github.com/rayyacub/telos-idea-matrix/internal/llm"
)

const (
    DefaultSimilarityThreshold = 0.85
    DefaultTTL                 = 24 * time.Hour
    DefaultMaxSize             = 1000
)

type CacheEntry struct {
    Key            string
    NormalizedText string
    Result         *llm.AnalysisResult
    CachedAt       time.Time
    HitCount       int64
    LastSimilarity float64

    // For LRU
    element *list.Element
}

type Cache struct {
    entries            map[string]*CacheEntry
    lru                *list.List
    mu                 sync.RWMutex
    similarityThreshold float64
    ttl                time.Duration
    maxSize            int

    // Stats
    hits   int64
    misses int64
}

func NewCache() *Cache {
    return &Cache{
        entries:            make(map[string]*CacheEntry),
        lru:                list.New(),
        similarityThreshold: DefaultSimilarityThreshold,
        ttl:                DefaultTTL,
        maxSize:            DefaultMaxSize,
    }
}

func (c *Cache) Store(ideaContent string, result *llm.AnalysisResult) {
    c.mu.Lock()
    defer c.mu.Unlock()

    normalized := NormalizeText(ideaContent)
    key := normalized // Use normalized text as key

    entry := &CacheEntry{
        Key:            key,
        NormalizedText: normalized,
        Result:         result,
        CachedAt:       time.Now(),
        HitCount:       0,
    }

    // Add to LRU front
    entry.element = c.lru.PushFront(entry)
    c.entries[key] = entry

    // Evict if over max size
    if c.lru.Len() > c.maxSize {
        c.evictOldest()
    }
}

func (c *Cache) Get(ideaContent string) (*llm.AnalysisResult, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    normalized := NormalizeText(ideaContent)

    // Try exact match first
    if entry, exists := c.entries[normalized]; exists {
        if !c.isExpired(entry) {
            entry.HitCount++
            entry.LastSimilarity = 1.0
            c.lru.MoveToFront(entry.element)
            c.hits++
            return entry.Result, true
        }
        // Expired, remove it
        c.removeEntry(entry)
    }

    // Try similarity match
    bestMatch := c.findSimilarEntry(normalized)
    if bestMatch != nil {
        bestMatch.HitCount++
        c.lru.MoveToFront(bestMatch.element)
        c.hits++
        return bestMatch.Result, true
    }

    c.misses++
    return nil, false
}

func (c *Cache) findSimilarEntry(normalized string) *CacheEntry {
    var bestMatch *CacheEntry
    var bestSimilarity float64

    for _, entry := range c.entries {
        if c.isExpired(entry) {
            continue
        }

        similarity := JaccardSimilarity(normalized, entry.NormalizedText)

        if similarity >= c.similarityThreshold && similarity > bestSimilarity {
            bestSimilarity = similarity
            bestMatch = entry
        }
    }

    if bestMatch != nil {
        bestMatch.LastSimilarity = bestSimilarity
    }

    return bestMatch
}

func (c *Cache) isExpired(entry *CacheEntry) bool {
    return time.Since(entry.CachedAt) > c.ttl
}

func (c *Cache) evictOldest() {
    element := c.lru.Back()
    if element != nil {
        entry := element.Value.(*CacheEntry)
        c.removeEntry(entry)
    }
}

func (c *Cache) removeEntry(entry *CacheEntry) {
    c.lru.Remove(entry.element)
    delete(c.entries, entry.Key)
}

func (c *Cache) Size() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.entries)
}

func (c *Cache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.entries = make(map[string]*CacheEntry)
    c.lru = list.New()
    c.hits = 0
    c.misses = 0
}
```

C. Implement go/internal/llm/cache/stats.go:
```go
package cache

type CacheStats struct {
    Size        int
    Hits        int64
    Misses      int64
    HitRate     float64
    AvgHitCount float64
}

func (c *Cache) GetStats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()

    total := c.hits + c.misses
    hitRate := 0.0
    if total > 0 {
        hitRate = float64(c.hits) / float64(total)
    }

    totalHitCount := int64(0)
    for _, entry := range c.entries {
        totalHitCount += entry.HitCount
    }

    avgHitCount := 0.0
    if len(c.entries) > 0 {
        avgHitCount = float64(totalHitCount) / float64(len(c.entries))
    }

    return CacheStats{
        Size:        len(c.entries),
        Hits:        c.hits,
        Misses:      c.misses,
        HitRate:     hitRate,
        AvgHitCount: avgHitCount,
    }
}
```

Run: go test ./internal/llm/cache -v
Expected: ALL TESTS PASS

STEP 3 - REFACTOR PHASE:
- Add cache persistence (save/load from disk)
- Optimize similarity calculation (skip low-similarity candidates early)
- Add cache warming (preload common queries)
- Extract configuration (threshold, TTL, max size)

INTEGRATION:
1. Wire into Ollama provider:
   - Check cache before calling Ollama
   - Store result in cache after analysis
2. Add cache stats to metrics endpoint
3. Add `tm cache stats` CLI command
4. Add `tm cache clear` CLI command

SUCCESS CRITERIA:
- ✅ All tests pass with >90% coverage
- ✅ Cache hits return in <5ms
- ✅ Similarity matching accuracy >90%
- ✅ Proper LRU eviction
- ✅ Thread-safe (verified with -race flag)
- ✅ Hit rate >60% in production use

VALIDATION:
```bash
# Unit tests
go test ./internal/llm/cache -v -cover -race

# Performance test
go test ./internal/llm/cache -v -bench=. -benchmem

# Integration test
go run ./cmd/cli/main.go analyze --ai "Build automation tool"
go run ./cmd/cli/main.go analyze --ai "Create automation tool"
# Second call should be instant (cache hit)
```

ESTIMATED TIME: 10-12 hours

DELIVERABLES:
- go/internal/llm/cache/cache.go
- go/internal/llm/cache/similarity.go
- go/internal/llm/cache/stats.go
- go/internal/llm/cache/cache_test.go
- go/internal/llm/cache/similarity_test.go
- Integration into Ollama provider
```

---

*[Continuing with remaining tracks 5C, 5D, 6A, 6B, 7, 8A, 8B, 8C in next section due to length...]*

Would you like me to:
1. Continue with the remaining track prompts (5C-8C)?
2. Create the parallel execution game plan now?
3. Both?
