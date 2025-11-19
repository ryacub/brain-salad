# Track 4C: Background Task Manager

**Phase**: 4 - Production Infrastructure
**Estimated Time**: 8-12 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 4A, 4B)

---

## Mission

You are implementing a background task manager for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation has task supervision in `src/background_tasks.rs`
- We need graceful shutdown, task lifecycle management, and scheduled tasks
- Must handle SIGTERM/SIGINT signals properly
- No goroutine leaks allowed

## Reference Implementation

Review `/home/user/brain-salad/src/background_tasks.rs`

## Your Task

Implement background task manager using strict TDD methodology.

## Directory Structure

Create files in `go/internal/tasks/`:
- `manager.go` - Task supervision and lifecycle
- `task.go` - Task interface and helpers
- `scheduler.go` - Scheduled task execution
- `manager_test.go` - Comprehensive tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/tasks/manager_test.go`:
- `TestTaskManager_NewManager()`
- `TestTaskManager_SpawnTask()`
- `TestTaskManager_GracefulShutdown()`
- `TestTaskManager_TaskCompletion()`
- `TestTaskManager_TaskFailure()`
- `TestTaskManager_ConcurrentTasks()`
- `TestTaskManager_ShutdownTimeout()`
- `TestTaskManager_NoGoroutineLeaks()`

Run: `go test ./internal/tasks -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Implement `go/internal/tasks/task.go`:

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

#### B. Implement `go/internal/tasks/manager.go`:

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

#### C. Implement `go/internal/tasks/scheduler.go`:

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

Run: `go test ./internal/tasks -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add task priority levels
- Add task retry policies
- Extract signal handling utilities
- Optimize goroutine management

## Integration

1. Wire into API server `main.go`:
   - Create TaskManager on startup
   - Add database cleanup scheduled task
   - Add metrics collection scheduled task
   - Call Shutdown on server shutdown
2. Add `tm tasks list` CLI command

## Success Criteria

- ✅ All tests pass with >85% coverage
- ✅ Graceful shutdown in <5 seconds
- ✅ Tasks properly canceled on shutdown
- ✅ No goroutine leaks (verified with -race flag)
- ✅ SIGTERM/SIGINT handling works

## Validation

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

## Deliverables

- `go/internal/tasks/manager.go`
- `go/internal/tasks/task.go`
- `go/internal/tasks/scheduler.go`
- `go/internal/tasks/manager_test.go`
- `go/cmd/web/main.go` (integrate TaskManager)
