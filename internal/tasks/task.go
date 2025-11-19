package tasks

import (
	"context"
	"errors"
	"time"
)

// Task represents a background task that can be executed
type Task interface {
	Name() string
	Run(ctx context.Context) error
	Timeout() time.Duration
}

// TaskResult contains the result of a completed task
type TaskResult struct {
	Name      string
	Error     error
	Duration  time.Duration
	StartedAt time.Time
	EndedAt   time.Time
}

// BaseTask provides a basic implementation of the Task interface
type BaseTask struct {
	name    string
	timeout time.Duration
}

// NewBaseTask creates a new BaseTask
func NewBaseTask(name string, timeout time.Duration) *BaseTask {
	return &BaseTask{
		name:    name,
		timeout: timeout,
	}
}

// Name returns the task name
func (bt *BaseTask) Name() string {
	return bt.name
}

// Timeout returns the task timeout
func (bt *BaseTask) Timeout() time.Duration {
	return bt.timeout
}

// Run must be implemented by tasks that embed BaseTask
func (bt *BaseTask) Run(_ context.Context) error {
	return errors.New("Run() must be implemented by tasks that embed BaseTask")
}

// FuncTask wraps a function as a Task
type FuncTask struct {
	BaseTask
	runFunc func(ctx context.Context) error
}

// NewFuncTask creates a new FuncTask
func NewFuncTask(name string, timeout time.Duration, runFunc func(ctx context.Context) error) *FuncTask {
	return &FuncTask{
		BaseTask: BaseTask{
			name:    name,
			timeout: timeout,
		},
		runFunc: runFunc,
	}
}

// Run executes the function
func (ft *FuncTask) Run(ctx context.Context) error {
	return ft.runFunc(ctx)
}
