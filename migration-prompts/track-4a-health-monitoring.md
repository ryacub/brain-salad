# Track 4A: Health Monitoring System

**Phase**: 4 - Production Infrastructure
**Estimated Time**: 8-10 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 4B, 4C)

---

## Mission

You are implementing a production-ready health monitoring system for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation has a comprehensive health monitoring system in `src/health.rs`
- We need to replicate this functionality in Go with proper test coverage (>90%)
- The health system should monitor database, memory, disk space, and application uptime

## Reference Implementation

Review the Rust implementation at `/home/user/brain-salad/src/health.rs` to understand:
- HealthMonitor struct with check registration
- HealthChecker trait/interface pattern
- Health state aggregation (Healthy/Degraded/Unhealthy)
- Built-in checkers (database, memory, disk)
- Uptime tracking

## Your Task

Implement the health monitoring system in Go using strict TDD methodology.

## Directory Structure

Create files in `go/internal/health/`:
- `monitor.go` - Core health monitoring
- `checkers.go` - Built-in health checkers
- `uptime.go` - Uptime tracking
- `monitor_test.go` - Comprehensive tests
- `checkers_test.go` - Checker tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests First)

Create `go/internal/health/monitor_test.go` with tests:
- `TestHealthMonitor_NewMonitor()`
- `TestHealthMonitor_AddCheck()`
- `TestHealthMonitor_RunAllChecks_Healthy()`
- `TestHealthMonitor_RunAllChecks_Degraded()`
- `TestHealthMonitor_RunAllChecks_Unhealthy()`
- `TestHealthCheckState_Aggregation()`
- `TestHealthMonitor_ConcurrentChecks()`

Create `go/internal/health/checkers_test.go` with tests:
- `TestDatabaseHealthChecker()`
- `TestMemoryHealthChecker()`
- `TestDiskSpaceHealthChecker()`
- `TestHealthChecker_Timeout()`
- `TestHealthChecker_Error()`

Run: `go test ./internal/health -v`
Expected: **ALL TESTS FAIL** (code doesn't exist yet)

### STEP 2 - GREEN PHASE (Implement Minimal Code)

#### A. Implement `go/internal/health/monitor.go`:

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

#### B. Implement `go/internal/health/checkers.go`:

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

Run: `go test ./internal/health -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Extract common patterns into helper functions
- Optimize concurrent health check execution
- Add configurable timeouts for each checker
- Improve error messages

## Integration

1. Add health endpoint to API server (`go/internal/api/handlers.go`):
   - `GET /health` → returns HealthStatus JSON
2. Add CLI command (`go/internal/cli/health.go`):
   - `tm health` → displays health status in terminal
3. Wire into graceful shutdown (check health before shutdown)

## Success Criteria

- ✅ All tests pass: `go test ./internal/health -v`
- ✅ Test coverage >90%: `go test ./internal/health -cover`
- ✅ Health checks complete in <100ms
- ✅ Proper state aggregation (any Error = Unhealthy, any Warning = Degraded)
- ✅ GET /health endpoint works
- ✅ `tm health` command displays status

## Validation

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

## Deliverables

- `go/internal/health/monitor.go`
- `go/internal/health/checkers.go`
- `go/internal/health/monitor_test.go`
- `go/internal/health/checkers_test.go`
- `go/internal/api/handlers.go` (add GET /health endpoint)
- `go/internal/cli/health.go` (add tm health command)
