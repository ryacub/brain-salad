package health

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Mock health checker for testing
type mockHealthChecker struct {
	name   string
	status HealthCheckState
	delay  time.Duration
}

func (m *mockHealthChecker) Name() string {
	return m.name
}

func (m *mockHealthChecker) Check(ctx context.Context) HealthCheckResult {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return HealthCheckResult{
		Name:      m.name,
		Status:    m.status,
		Message:   "Mock check result",
		Duration:  m.delay,
		Timestamp: time.Now(),
	}
}

// TestHealthMonitor_NewMonitor tests creating a new health monitor
func TestHealthMonitor_NewMonitor(t *testing.T) {
	monitor := NewHealthMonitor()

	if monitor == nil {
		t.Fatal("NewHealthMonitor returned nil")
		return
	}

	if monitor.startTime.IsZero() {
		t.Error("Start time should be set")
	}

	if monitor.checkers == nil {
		t.Error("Checkers slice should be initialized")
	}

	if len(monitor.checkers) != 0 {
		t.Errorf("Expected 0 checkers, got %d", len(monitor.checkers))
	}
}

// TestHealthMonitor_AddCheck tests adding health checks to the monitor
func TestHealthMonitor_AddCheck(t *testing.T) {
	monitor := NewHealthMonitor()

	checker1 := &mockHealthChecker{name: "check1", status: Ok}
	checker2 := &mockHealthChecker{name: "check2", status: Ok}

	monitor.AddCheck(checker1)
	if len(monitor.checkers) != 1 {
		t.Errorf("Expected 1 checker, got %d", len(monitor.checkers))
	}

	monitor.AddCheck(checker2)
	if len(monitor.checkers) != 2 {
		t.Errorf("Expected 2 checkers, got %d", len(monitor.checkers))
	}
}

// TestHealthMonitor_RunAllChecks_Healthy tests running checks when all are healthy
func TestHealthMonitor_RunAllChecks_Healthy(t *testing.T) {
	monitor := NewHealthMonitor()

	monitor.AddCheck(&mockHealthChecker{name: "check1", status: Ok})
	monitor.AddCheck(&mockHealthChecker{name: "check2", status: Ok})
	monitor.AddCheck(&mockHealthChecker{name: "check3", status: Ok})

	ctx := context.Background()
	status := monitor.RunAllChecks(ctx)

	if status.Status != Healthy {
		t.Errorf("Expected status Healthy, got %s", status.Status)
	}

	if len(status.Checks) != 3 {
		t.Errorf("Expected 3 check results, got %d", len(status.Checks))
	}

	for _, check := range status.Checks {
		if check.Status != Ok {
			t.Errorf("Expected check status Ok, got %s", check.Status)
		}
	}

	if status.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if status.Uptime <= 0 {
		t.Error("Uptime should be greater than 0")
	}
}

// TestHealthMonitor_RunAllChecks_Degraded tests running checks with warnings
func TestHealthMonitor_RunAllChecks_Degraded(t *testing.T) {
	monitor := NewHealthMonitor()

	monitor.AddCheck(&mockHealthChecker{name: "check1", status: Ok})
	monitor.AddCheck(&mockHealthChecker{name: "check2", status: Warning})
	monitor.AddCheck(&mockHealthChecker{name: "check3", status: Ok})

	ctx := context.Background()
	status := monitor.RunAllChecks(ctx)

	if status.Status != Degraded {
		t.Errorf("Expected status Degraded, got %s", status.Status)
	}

	if len(status.Checks) != 3 {
		t.Errorf("Expected 3 check results, got %d", len(status.Checks))
	}
}

// TestHealthMonitor_RunAllChecks_Unhealthy tests running checks with errors
func TestHealthMonitor_RunAllChecks_Unhealthy(t *testing.T) {
	monitor := NewHealthMonitor()

	monitor.AddCheck(&mockHealthChecker{name: "check1", status: Ok})
	monitor.AddCheck(&mockHealthChecker{name: "check2", status: Warning})
	monitor.AddCheck(&mockHealthChecker{name: "check3", status: Error})

	ctx := context.Background()
	status := monitor.RunAllChecks(ctx)

	if status.Status != Unhealthy {
		t.Errorf("Expected status Unhealthy, got %s", status.Status)
	}

	if len(status.Checks) != 3 {
		t.Errorf("Expected 3 check results, got %d", len(status.Checks))
	}
}

// TestHealthCheckState_Aggregation tests state aggregation logic
func TestHealthCheckState_Aggregation(t *testing.T) {
	tests := []struct {
		name           string
		checkStatuses  []HealthCheckState
		expectedStatus HealthState
	}{
		{
			name:           "All OK",
			checkStatuses:  []HealthCheckState{Ok, Ok, Ok},
			expectedStatus: Healthy,
		},
		{
			name:           "One Warning",
			checkStatuses:  []HealthCheckState{Ok, Warning, Ok},
			expectedStatus: Degraded,
		},
		{
			name:           "Multiple Warnings",
			checkStatuses:  []HealthCheckState{Warning, Warning, Ok},
			expectedStatus: Degraded,
		},
		{
			name:           "One Error",
			checkStatuses:  []HealthCheckState{Ok, Ok, Error},
			expectedStatus: Unhealthy,
		},
		{
			name:           "Error and Warning",
			checkStatuses:  []HealthCheckState{Warning, Error, Ok},
			expectedStatus: Unhealthy,
		},
		{
			name:           "Multiple Errors",
			checkStatuses:  []HealthCheckState{Error, Error, Ok},
			expectedStatus: Unhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewHealthMonitor()

			for i, status := range tt.checkStatuses {
				monitor.AddCheck(&mockHealthChecker{
					name:   string(rune('a' + i)),
					status: status,
				})
			}

			ctx := context.Background()
			result := monitor.RunAllChecks(ctx)

			if result.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result.Status)
			}
		})
	}
}

// TestHealthMonitor_ConcurrentChecks tests concurrent health check execution
func TestHealthMonitor_ConcurrentChecks(t *testing.T) {
	monitor := NewHealthMonitor()

	// Add checks with delays to test concurrency
	monitor.AddCheck(&mockHealthChecker{name: "slow1", status: Ok, delay: 50 * time.Millisecond})
	monitor.AddCheck(&mockHealthChecker{name: "slow2", status: Ok, delay: 50 * time.Millisecond})
	monitor.AddCheck(&mockHealthChecker{name: "slow3", status: Ok, delay: 50 * time.Millisecond})

	ctx := context.Background()
	start := time.Now()
	status := monitor.RunAllChecks(ctx)
	elapsed := time.Since(start)

	// If checks run concurrently, total time should be ~50ms, not 150ms
	// Allow some overhead, so check if it's less than 100ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Checks took too long (%v), may not be running concurrently", elapsed)
	}

	if status.Status != Healthy {
		t.Errorf("Expected status Healthy, got %s", status.Status)
	}

	if len(status.Checks) != 3 {
		t.Errorf("Expected 3 check results, got %d", len(status.Checks))
	}
}

// TestHealthMonitor_IsHealthy tests the IsHealthy method
func TestHealthMonitor_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		statuses []HealthCheckState
		expected bool
	}{
		{
			name:     "All healthy",
			statuses: []HealthCheckState{Ok, Ok},
			expected: true,
		},
		{
			name:     "With warning",
			statuses: []HealthCheckState{Ok, Warning},
			expected: false,
		},
		{
			name:     "With error",
			statuses: []HealthCheckState{Ok, Error},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewHealthMonitor()

			for i, status := range tt.statuses {
				monitor.AddCheck(&mockHealthChecker{
					name:   string(rune('a' + i)),
					status: status,
				})
			}

			ctx := context.Background()
			result := monitor.IsHealthy(ctx)

			if result != tt.expected {
				t.Errorf("Expected IsHealthy=%v, got %v", tt.expected, result)
			}
		})
	}
}

// TestHealthMonitor_ContextCancellation tests context cancellation during checks
func TestHealthMonitor_ContextCancellation(t *testing.T) {
	monitor := NewHealthMonitor()

	// Add a slow checker
	monitor.AddCheck(&mockHealthChecker{name: "slow", status: Ok, delay: 1 * time.Second})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	start := time.Now()
	status := monitor.RunAllChecks(ctx)
	elapsed := time.Since(start)

	// Should complete quickly due to context timeout
	if elapsed > 200*time.Millisecond {
		t.Errorf("Check should timeout quickly, took %v", elapsed)
	}

	// Should still return a status (may be degraded/unhealthy due to timeout)
	if len(status.Checks) == 0 {
		t.Error("Should return check results even with timeout")
	}
}

// TestHealthMonitor_ThreadSafety tests concurrent access to the monitor
func TestHealthMonitor_ThreadSafety(t *testing.T) {
	monitor := NewHealthMonitor()

	var wg sync.WaitGroup
	ctx := context.Background()

	// Concurrently add checks and run health checks
	for i := 0; i < 10; i++ {
		wg.Add(2)

		go func(idx int) {
			defer wg.Done()
			monitor.AddCheck(&mockHealthChecker{
				name:   string(rune('a' + idx)),
				status: Ok,
			})
		}(i)

		go func() {
			defer wg.Done()
			_ = monitor.RunAllChecks(ctx)
		}()
	}

	wg.Wait()

	// Should not panic and should have added checkers
	status := monitor.RunAllChecks(ctx)
	if len(status.Checks) == 0 {
		t.Error("Expected some checks to be added")
	}
}

// TestHealthMonitor_Uptime tests uptime tracking
func TestHealthMonitor_Uptime(t *testing.T) {
	monitor := NewHealthMonitor()

	// Wait a bit to ensure uptime is measurable
	time.Sleep(10 * time.Millisecond)

	ctx := context.Background()
	status := monitor.RunAllChecks(ctx)

	if status.Uptime < 10*time.Millisecond {
		t.Errorf("Expected uptime >= 10ms, got %v", status.Uptime)
	}
}

// TestHealthMonitor_CheckDuration tests that check duration is recorded
func TestHealthMonitor_CheckDuration(t *testing.T) {
	monitor := NewHealthMonitor()

	delay := 20 * time.Millisecond
	monitor.AddCheck(&mockHealthChecker{name: "timed", status: Ok, delay: delay})

	ctx := context.Background()
	status := monitor.RunAllChecks(ctx)

	if len(status.Checks) != 1 {
		t.Fatal("Expected 1 check result")
	}

	checkDuration := status.Checks[0].Duration
	if checkDuration < delay {
		t.Errorf("Expected duration >= %v, got %v", delay, checkDuration)
	}
}
