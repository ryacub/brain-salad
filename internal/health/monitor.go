package health

import (
	"context"
	"sync"
	"time"
)

// HealthState represents the overall health state of the system
type HealthState string

const (
	Healthy   HealthState = "healthy"
	Degraded  HealthState = "degraded"
	Unhealthy HealthState = "unhealthy"
)

// HealthCheckState represents the state of an individual health check
type HealthCheckState string

const (
	Ok      HealthCheckState = "ok"
	Warning HealthCheckState = "warning"
	Error   HealthCheckState = "error"
)

// HealthCheckResult represents the result of a single health check
type HealthCheckResult struct {
	Name      string           `json:"name"`
	Status    HealthCheckState `json:"status"`
	Message   string           `json:"message,omitempty"`
	Duration  time.Duration    `json:"duration"`
	Timestamp time.Time        `json:"timestamp"`
}

// HealthStatus represents the overall health status of the system
type HealthStatus struct {
	Status    HealthState         `json:"status"`
	Timestamp time.Time           `json:"timestamp"`
	Checks    []HealthCheckResult `json:"checks"`
	Uptime    time.Duration       `json:"uptime"`
	Version   string              `json:"version"`
}

// HealthChecker is the interface that health checkers must implement
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) HealthCheckResult
}

// HealthMonitor manages and executes health checks
type HealthMonitor struct {
	startTime time.Time
	checkers  []HealthChecker
	mu        sync.RWMutex
	version   string
}

// NewHealthMonitor creates a new health monitor instance
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		startTime: time.Now(),
		checkers:  make([]HealthChecker, 0),
		version:   "1.0.0", // Default version
	}
}

// SetVersion sets the version string for health status
func (hm *HealthMonitor) SetVersion(version string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.version = version
}

// AddCheck adds a health checker to the monitor
func (hm *HealthMonitor) AddCheck(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers = append(hm.checkers, checker)
}

// RunAllChecks executes all registered health checks concurrently
func (hm *HealthMonitor) RunAllChecks(ctx context.Context) HealthStatus {
	hm.mu.RLock()
	checkers := make([]HealthChecker, len(hm.checkers))
	copy(checkers, hm.checkers)
	version := hm.version
	hm.mu.RUnlock()

	// Run all checks concurrently
	results := make(chan HealthCheckResult, len(checkers))
	var wg sync.WaitGroup

	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()

			start := time.Now()

			// Create a channel to receive the result
			resultChan := make(chan HealthCheckResult, 1)

			// Run the check in a goroutine
			go func() {
				result := c.Check(ctx)
				result.Duration = time.Since(start)
				result.Timestamp = time.Now()
				resultChan <- result
			}()

			// Wait for either the result or context cancellation
			select {
			case result := <-resultChan:
				results <- result
			case <-ctx.Done():
				// Context cancelled, return error result
				results <- HealthCheckResult{
					Name:      c.Name(),
					Status:    Error,
					Message:   "Health check timed out: " + ctx.Err().Error(),
					Duration:  time.Since(start),
					Timestamp: time.Now(),
				}
			}
		}(checker)
	}

	// Wait for all checks to complete
	wg.Wait()
	close(results)

	// Collect results and determine overall state
	checkResults := make([]HealthCheckResult, 0, len(checkers))
	overallState := Healthy

	for result := range results {
		checkResults = append(checkResults, result)

		// Aggregate state based on check results
		switch result.Status {
		case Error:
			overallState = Unhealthy
		case Warning:
			if overallState == Healthy {
				overallState = Degraded
			}
		case Ok:
			// Keep current state
		}
	}

	return HealthStatus{
		Status:    overallState,
		Timestamp: time.Now(),
		Checks:    checkResults,
		Uptime:    time.Since(hm.startTime),
		Version:   version,
	}
}

// IsHealthy returns true if the system is healthy (all checks pass)
func (hm *HealthMonitor) IsHealthy(ctx context.Context) bool {
	status := hm.RunAllChecks(ctx)
	return status.Status == Healthy
}
