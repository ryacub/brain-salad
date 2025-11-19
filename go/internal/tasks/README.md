# Background Task Manager

A production-ready background task manager for the Telos Idea Matrix Go application, implemented using Test-Driven Development (TDD).

## Features

- ✅ Task supervision and lifecycle management
- ✅ Graceful shutdown with configurable timeout
- ✅ Context-based cancellation
- ✅ Scheduled/periodic tasks
- ✅ Task result tracking
- ✅ No goroutine leaks
- ✅ Race condition free
- ✅ SIGTERM/SIGINT signal handling
- ✅ Structured logging with zerolog

## Components

### Task Interface

The core `Task` interface that all tasks must implement:

```go
type Task interface {
    Name() string
    Run(ctx context.Context) error
    Timeout() time.Duration
}
```

### TaskManager

Manages the lifecycle of all background tasks:

- **Spawn**: Start a new background task
- **Shutdown**: Gracefully shutdown all tasks with timeout
- **GetResults**: Retrieve results from completed tasks
- **TaskCount**: Get the number of spawned tasks

### ScheduledTask

A task that runs periodically at specified intervals:

```go
task := tasks.NewScheduledTask(
    "cleanup-task",
    1*time.Hour,
    func(ctx context.Context) error {
        // Task logic here
        return nil
    },
)
```

### FuncTask

A simple task wrapper for functions:

```go
task := tasks.NewFuncTask(
    "one-time-task",
    5*time.Second,
    func(ctx context.Context) error {
        // Task logic here
        return nil
    },
)
```

## Usage

### Basic Example

```go
// Create task manager
tm := tasks.NewTaskManager()

// Spawn a one-time task
task := tasks.NewFuncTask("example", 10*time.Second, func(ctx context.Context) error {
    log.Info().Msg("Task running")
    return nil
})
tm.Spawn(task)

// Spawn a periodic task
periodic := tasks.NewScheduledTask("periodic", 1*time.Minute, func(ctx context.Context) error {
    log.Info().Msg("Periodic task running")
    return nil
})
tm.Spawn(periodic)

// Graceful shutdown
if err := tm.Shutdown(5 * time.Second); err != nil {
    log.Error().Err(err).Msg("Shutdown timeout")
}
```

### Integration with Main Application

The task manager is integrated into `cmd/web/main.go` with the following background tasks:

1. **Database Cleanup** - Runs every 1 hour
2. **Metrics Collection** - Runs every 5 minutes
3. **Health Check** - Runs every 30 seconds

## Testing

### Run Tests

```bash
# Run all tests
go test ./internal/tasks -v

# Run with coverage
go test ./internal/tasks -v -cover

# Run with race detection
go test ./internal/tasks -v -race

# Run with both coverage and race detection
go test ./internal/tasks -v -cover -race
```

### Test Coverage

Current coverage: **85.3%**

Tests include:
- Task manager creation
- Task spawning and execution
- Graceful shutdown
- Task completion tracking
- Task failure handling
- Concurrent task execution
- Shutdown timeout
- Goroutine leak detection
- Context cancellation
- Scheduled task execution
- Scheduled task cancellation

## Architecture

The implementation follows the Rust reference implementation in `src/background_tasks.rs` with Go-specific adaptations:

- Uses `context.Context` for cancellation (instead of Tokio channels)
- Uses `sync.WaitGroup` for task tracking
- Uses `sync.Mutex` for thread-safe result storage
- Follows Go concurrency patterns and idioms

## Graceful Shutdown

The task manager ensures graceful shutdown:

1. Cancels all task contexts via `context.CancelFunc`
2. Waits for all tasks to complete with configurable timeout
3. Returns error if timeout is exceeded
4. No goroutine leaks after shutdown

## Signal Handling

The task manager can listen for OS signals:

```go
tm.ListenForShutdown() // Listens for SIGTERM and SIGINT
tm.Wait()              // Blocks until shutdown completes
```

## Error Handling

- Task errors are logged and tracked in `TaskResult`
- Scheduled tasks continue running even if individual iterations fail
- All errors are logged with structured logging

## Performance

- No goroutine leaks (verified with tests)
- Race condition free (verified with `-race` flag)
- Efficient shutdown with configurable timeout
- Minimal overhead for task supervision

## Future Enhancements

Potential improvements for future iterations:

- Task priority levels
- Task retry policies with exponential backoff
- Task dependency management
- Dynamic task scheduling
- Metrics and monitoring integration
- Task result persistence
