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

// TaskManager manages the lifecycle of background tasks
type TaskManager struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	tasks      []Task
	results    []TaskResult
	mu         sync.Mutex
	shutdownCh chan struct{}
}

// NewTaskManager creates a new TaskManager
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

// Spawn starts a new background task
func (tm *TaskManager) Spawn(task Task) {
	tm.mu.Lock()
	tm.tasks = append(tm.tasks, task)
	tm.mu.Unlock()

	tm.wg.Add(1)
	go tm.runTask(task)
}

// runTask executes a task with proper supervision
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

// Shutdown gracefully shuts down all tasks
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
		close(tm.shutdownCh)
		return nil
	case <-time.After(timeout):
		close(tm.shutdownCh)
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// ListenForShutdown sets up signal handlers for graceful shutdown
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
	}()
}

// Wait blocks until shutdown is complete
func (tm *TaskManager) Wait() {
	<-tm.shutdownCh
}

// GetResults returns a copy of all task results
func (tm *TaskManager) GetResults() []TaskResult {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	results := make([]TaskResult, len(tm.results))
	copy(results, tm.results)
	return results
}

// TaskCount returns the number of tasks spawned
func (tm *TaskManager) TaskCount() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return len(tm.tasks)
}
