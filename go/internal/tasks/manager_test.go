package tasks

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// TestTaskManager_NewManager verifies that a new TaskManager can be created
func TestTaskManager_NewManager(t *testing.T) {
	tm := NewTaskManager()
	if tm == nil {
		t.Fatal("NewTaskManager() returned nil")
	}

	// Verify initial state
	results := tm.GetResults()
	if len(results) != 0 {
		t.Errorf("new TaskManager should have 0 results, got %d", len(results))
	}
}

// TestTaskManager_SpawnTask verifies that tasks can be spawned successfully
func TestTaskManager_SpawnTask(t *testing.T) {
	tm := NewTaskManager()

	executed := atomic.Bool{}
	task := &mockTask{
		name: "test-task",
		runFunc: func(ctx context.Context) error {
			executed.Store(true)
			return nil
		},
		timeout: 5 * time.Second,
	}

	tm.Spawn(task)

	// Wait a bit for task to execute
	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("task was not executed")
	}

	// Shutdown gracefully
	if err := tm.Shutdown(2 * time.Second); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

// TestTaskManager_GracefulShutdown verifies graceful shutdown works correctly
func TestTaskManager_GracefulShutdown(t *testing.T) {
	tm := NewTaskManager()

	started := atomic.Bool{}
	cancelled := atomic.Bool{}

	task := &mockTask{
		name: "long-running-task",
		runFunc: func(ctx context.Context) error {
			started.Store(true)
			<-ctx.Done()
			cancelled.Store(true)
			return ctx.Err()
		},
		timeout: 10 * time.Second,
	}

	tm.Spawn(task)

	// Wait for task to start
	time.Sleep(50 * time.Millisecond)

	if !started.Load() {
		t.Fatal("task did not start")
	}

	// Shutdown should cancel the task
	start := time.Now()
	err := tm.Shutdown(2 * time.Second)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	if duration > 3*time.Second {
		t.Errorf("Shutdown took too long: %v", duration)
	}

	if !cancelled.Load() {
		t.Error("task was not cancelled")
	}
}

// TestTaskManager_TaskCompletion verifies task completion is tracked
func TestTaskManager_TaskCompletion(t *testing.T) {
	tm := NewTaskManager()

	task := &mockTask{
		name: "quick-task",
		runFunc: func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
		timeout: 2 * time.Second,
	}

	tm.Spawn(task)

	// Wait for task to complete
	time.Sleep(200 * time.Millisecond)

	results := tm.GetResults()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Name != "quick-task" {
		t.Errorf("result.Name = %s, want quick-task", result.Name)
	}
	if result.Error != nil {
		t.Errorf("result.Error = %v, want nil", result.Error)
	}
	if result.Duration < 50*time.Millisecond {
		t.Errorf("result.Duration = %v, want >= 50ms", result.Duration)
	}

	tm.Shutdown(2 * time.Second)
}

// TestTaskManager_TaskFailure verifies task failures are tracked
func TestTaskManager_TaskFailure(t *testing.T) {
	tm := NewTaskManager()

	expectedErr := errors.New("task failed")
	task := &mockTask{
		name: "failing-task",
		runFunc: func(ctx context.Context) error {
			return expectedErr
		},
		timeout: 2 * time.Second,
	}

	tm.Spawn(task)

	// Wait for task to complete
	time.Sleep(100 * time.Millisecond)

	results := tm.GetResults()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Name != "failing-task" {
		t.Errorf("result.Name = %s, want failing-task", result.Name)
	}
	if result.Error == nil {
		t.Error("result.Error = nil, want error")
	}

	tm.Shutdown(2 * time.Second)
}

// TestTaskManager_ConcurrentTasks verifies multiple tasks can run concurrently
func TestTaskManager_ConcurrentTasks(t *testing.T) {
	tm := NewTaskManager()

	numTasks := 10
	counter := atomic.Int32{}

	for i := 0; i < numTasks; i++ {
		task := &mockTask{
			name: "concurrent-task",
			runFunc: func(ctx context.Context) error {
				counter.Add(1)
				time.Sleep(50 * time.Millisecond)
				return nil
			},
			timeout: 2 * time.Second,
		}
		tm.Spawn(task)
	}

	// Wait for all tasks to complete
	time.Sleep(200 * time.Millisecond)

	if count := counter.Load(); count != int32(numTasks) {
		t.Errorf("counter = %d, want %d", count, numTasks)
	}

	results := tm.GetResults()
	if len(results) != numTasks {
		t.Errorf("len(results) = %d, want %d", len(results), numTasks)
	}

	tm.Shutdown(2 * time.Second)
}

// TestTaskManager_ShutdownTimeout verifies shutdown timeout works
func TestTaskManager_ShutdownTimeout(t *testing.T) {
	tm := NewTaskManager()

	task := &mockTask{
		name: "stuck-task",
		runFunc: func(ctx context.Context) error {
			// Ignore cancellation and sleep
			time.Sleep(5 * time.Second)
			return nil
		},
		timeout: 10 * time.Second,
	}

	tm.Spawn(task)
	time.Sleep(50 * time.Millisecond)

	// Shutdown with short timeout
	start := time.Now()
	err := tm.Shutdown(500 * time.Millisecond)
	duration := time.Since(start)

	// Should timeout
	if err == nil {
		t.Error("Shutdown() should have timed out")
	}

	// Should take about 500ms
	if duration < 400*time.Millisecond || duration > 700*time.Millisecond {
		t.Errorf("Shutdown duration = %v, want ~500ms", duration)
	}
}

// TestTaskManager_NoGoroutineLeaks verifies no goroutine leaks occur
func TestTaskManager_NoGoroutineLeaks(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	tm := NewTaskManager()

	// Spawn multiple tasks
	for i := 0; i < 20; i++ {
		task := &mockTask{
			name: "leak-test-task",
			runFunc: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			timeout: 2 * time.Second,
		}
		tm.Spawn(task)
	}

	// Wait for tasks to complete
	time.Sleep(200 * time.Millisecond)

	// Shutdown
	tm.Shutdown(2 * time.Second)

	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// Allow for some variance, but should be close to initial
	if finalGoroutines > initialGoroutines+2 {
		t.Errorf("goroutine leak detected: initial=%d, final=%d", initialGoroutines, finalGoroutines)
	}
}

// TestTaskManager_ContextCancellation verifies tasks respect context cancellation
func TestTaskManager_ContextCancellation(t *testing.T) {
	tm := NewTaskManager()

	cancelled := atomic.Bool{}
	task := &mockTask{
		name: "cancellable-task",
		runFunc: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				cancelled.Store(true)
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return errors.New("task did not respect cancellation")
			}
		},
		timeout: 10 * time.Second,
	}

	tm.Spawn(task)
	time.Sleep(50 * time.Millisecond)

	// Cancel via shutdown
	tm.Shutdown(1 * time.Second)

	if !cancelled.Load() {
		t.Error("task did not receive cancellation")
	}
}

// TestScheduledTask verifies scheduled tasks run periodically
func TestScheduledTask(t *testing.T) {
	tm := NewTaskManager()

	counter := atomic.Int32{}
	task := NewScheduledTask("periodic-task", 100*time.Millisecond, func(ctx context.Context) error {
		counter.Add(1)
		return nil
	})

	tm.Spawn(task)

	// Let it run for ~350ms (should execute 3-4 times)
	time.Sleep(350 * time.Millisecond)

	tm.Shutdown(1 * time.Second)

	count := counter.Load()
	if count < 2 || count > 5 {
		t.Errorf("counter = %d, want 2-5 (periodic task should run multiple times)", count)
	}
}

// TestScheduledTask_Cancellation verifies scheduled tasks can be cancelled
func TestScheduledTask_Cancellation(t *testing.T) {
	tm := NewTaskManager()

	counter := atomic.Int32{}
	task := NewScheduledTask("periodic-task", 50*time.Millisecond, func(ctx context.Context) error {
		counter.Add(1)
		return nil
	})

	tm.Spawn(task)

	// Let it run for a bit
	time.Sleep(175 * time.Millisecond)

	// Cancel via shutdown
	tm.Shutdown(1 * time.Second)

	countAfterShutdown := counter.Load()

	// Wait more time
	time.Sleep(200 * time.Millisecond)

	// Counter should not increase after shutdown
	if counter.Load() != countAfterShutdown {
		t.Error("scheduled task continued running after shutdown")
	}
}

// TestFuncTask verifies FuncTask works correctly
func TestFuncTask(t *testing.T) {
	executed := atomic.Bool{}
	task := NewFuncTask("func-task", 2*time.Second, func(ctx context.Context) error {
		executed.Store(true)
		return nil
	})

	if task.Name() != "func-task" {
		t.Errorf("task.Name() = %s, want func-task", task.Name())
	}

	if task.Timeout() != 2*time.Second {
		t.Errorf("task.Timeout() = %v, want 2s", task.Timeout())
	}

	ctx := context.Background()
	if err := task.Run(ctx); err != nil {
		t.Errorf("task.Run() error = %v, want nil", err)
	}

	if !executed.Load() {
		t.Error("task was not executed")
	}
}

// TestBaseTask verifies BaseTask accessors work
func TestBaseTask(t *testing.T) {
	task := NewBaseTask("base-task", 3*time.Second)

	if task.Name() != "base-task" {
		t.Errorf("task.Name() = %s, want base-task", task.Name())
	}

	if task.Timeout() != 3*time.Second {
		t.Errorf("task.Timeout() = %v, want 3s", task.Timeout())
	}
}

// TestScheduledTask_WithTimeout verifies WithTimeout works
func TestScheduledTask_WithTimeout(t *testing.T) {
	task := NewScheduledTask("test", 100*time.Millisecond, func(ctx context.Context) error {
		return nil
	}).WithTimeout(10 * time.Second)

	if task.Timeout() != 10*time.Second {
		t.Errorf("task.Timeout() = %v, want 10s", task.Timeout())
	}
}

// TestTaskManager_TaskCount verifies TaskCount works
func TestTaskManager_TaskCount(t *testing.T) {
	tm := NewTaskManager()

	if count := tm.TaskCount(); count != 0 {
		t.Errorf("initial TaskCount() = %d, want 0", count)
	}

	for i := 0; i < 5; i++ {
		task := &mockTask{
			name: "test-task",
			runFunc: func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			},
			timeout: 2 * time.Second,
		}
		tm.Spawn(task)
	}

	if count := tm.TaskCount(); count != 5 {
		t.Errorf("TaskCount() = %d, want 5", count)
	}

	time.Sleep(200 * time.Millisecond)
	tm.Shutdown(2 * time.Second)
}

// TestScheduledTask_ErrorHandling verifies errors in scheduled tasks are handled
func TestScheduledTask_ErrorHandling(t *testing.T) {
	tm := NewTaskManager()

	counter := atomic.Int32{}
	task := NewScheduledTask("error-task", 50*time.Millisecond, func(ctx context.Context) error {
		count := counter.Add(1)
		if count%2 == 0 {
			return errors.New("simulated error")
		}
		return nil
	})

	tm.Spawn(task)

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)

	tm.Shutdown(1 * time.Second)

	// Should have run multiple times despite errors
	if count := counter.Load(); count < 2 {
		t.Errorf("counter = %d, want >= 2", count)
	}
}

// mockTask is a test implementation of the Task interface
type mockTask struct {
	name    string
	runFunc func(ctx context.Context) error
	timeout time.Duration
}

func (m *mockTask) Name() string {
	return m.name
}

func (m *mockTask) Run(ctx context.Context) error {
	return m.runFunc(ctx)
}

func (m *mockTask) Timeout() time.Duration {
	return m.timeout
}
