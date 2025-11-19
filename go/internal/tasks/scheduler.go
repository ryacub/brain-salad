package tasks

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// ScheduledTask represents a task that runs periodically
type ScheduledTask struct {
	BaseTask
	interval time.Duration
	runFunc  func(ctx context.Context) error
}

// NewScheduledTask creates a new scheduled task that runs at the specified interval
func NewScheduledTask(name string, interval time.Duration, runFunc func(ctx context.Context) error) *ScheduledTask {
	return &ScheduledTask{
		BaseTask: BaseTask{
			name:    name,
			timeout: 5 * time.Minute, // Default timeout for scheduled tasks
		},
		interval: interval,
		runFunc:  runFunc,
	}
}

// Run executes the scheduled task repeatedly until cancelled
func (st *ScheduledTask) Run(ctx context.Context) error {
	ticker := time.NewTicker(st.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("task", st.name).Msg("scheduled task cancelled")
			return ctx.Err()
		case <-ticker.C:
			if err := st.runFunc(ctx); err != nil {
				log.Error().
					Err(err).
					Str("task", st.name).
					Msg("scheduled task iteration failed")
				// Continue running even if one iteration fails
				continue
			}
		}
	}
}

// WithTimeout sets a custom timeout for the scheduled task
func (st *ScheduledTask) WithTimeout(timeout time.Duration) *ScheduledTask {
	st.timeout = timeout
	return st
}
