package health

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestDatabaseHealthChecker tests the database health checker
func TestDatabaseHealthChecker(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() { _ = db.Close() }()

	checker := NewDatabaseHealthChecker(db)

	if checker == nil {
		t.Fatal("NewDatabaseHealthChecker returned nil")
	}

	if checker.Name() != "database" {
		t.Errorf("Expected name 'database', got '%s'", checker.Name())
	}

	ctx := context.Background()
	result := checker.Check(ctx)

	if result.Name != "database" {
		t.Errorf("Expected result name 'database', got '%s'", result.Name)
	}

	if result.Status != Ok {
		t.Errorf("Expected status Ok for healthy database, got %s", result.Status)
	}

	if result.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if result.Duration < 0 {
		t.Error("Duration should be non-negative")
	}
}

// TestDatabaseHealthChecker_ClosedConnection tests database checker with closed connection
func TestDatabaseHealthChecker_ClosedConnection(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Close the database before checking
	_ = db.Close()

	checker := NewDatabaseHealthChecker(db)
	ctx := context.Background()
	result := checker.Check(ctx)

	if result.Status != Error {
		t.Errorf("Expected status Error for closed database, got %s", result.Status)
	}

	if result.Message == "" {
		t.Error("Expected error message for failed check")
	}
}

// TestDatabaseHealthChecker_Timeout tests database checker with context timeout
func TestDatabaseHealthChecker_Timeout(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() { _ = db.Close() }()

	checker := NewDatabaseHealthChecker(db)

	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout has passed

	result := checker.Check(ctx)

	// Should handle timeout gracefully
	if result.Name != "database" {
		t.Errorf("Expected result name 'database', got '%s'", result.Name)
	}
}

// TestMemoryHealthChecker tests the memory health checker
func TestMemoryHealthChecker(t *testing.T) {
	// Test with a very high threshold (should always pass)
	checker := NewMemoryHealthChecker(10000.0) // 10GB threshold

	if checker == nil {
		t.Fatal("NewMemoryHealthChecker returned nil")
	}

	if checker.Name() != "memory" {
		t.Errorf("Expected name 'memory', got '%s'", checker.Name())
	}

	ctx := context.Background()
	result := checker.Check(ctx)

	if result.Name != "memory" {
		t.Errorf("Expected result name 'memory', got '%s'", result.Name)
	}

	if result.Status != Ok {
		t.Errorf("Expected status Ok with high threshold, got %s", result.Status)
	}

	if result.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

// TestMemoryHealthChecker_Warning tests memory checker warning threshold
func TestMemoryHealthChecker_Warning(t *testing.T) {
	// Test with a very low threshold (should trigger warning)
	checker := NewMemoryHealthChecker(0.001) // 1KB threshold - should always exceed

	ctx := context.Background()
	result := checker.Check(ctx)

	if result.Status != Warning {
		t.Errorf("Expected status Warning with low threshold, got %s", result.Status)
	}

	if result.Message == "" {
		t.Error("Expected message explaining high memory usage")
	}
}

// TestMemoryHealthChecker_Values tests memory checker reports actual values
func TestMemoryHealthChecker_Values(t *testing.T) {
	checker := NewMemoryHealthChecker(10000.0)

	ctx := context.Background()
	result := checker.Check(ctx)

	// The message should contain actual memory usage information
	if result.Message == "" {
		t.Error("Expected message with memory usage information")
	}

	// Duration should be very fast (memory check is quick)
	if result.Duration > 100*time.Millisecond {
		t.Errorf("Memory check took too long: %v", result.Duration)
	}
}

// TestDiskSpaceHealthChecker tests the disk space health checker
func TestDiskSpaceHealthChecker(t *testing.T) {
	// Test with current directory and low threshold
	checker := NewDiskSpaceHealthChecker("/tmp", 1) // 1MB threshold

	if checker == nil {
		t.Fatal("NewDiskSpaceHealthChecker returned nil")
	}

	if checker.Name() != "disk_space" {
		t.Errorf("Expected name 'disk_space', got '%s'", checker.Name())
	}

	ctx := context.Background()
	result := checker.Check(ctx)

	if result.Name != "disk_space" {
		t.Errorf("Expected result name 'disk_space', got '%s'", result.Name)
	}

	// With 1MB threshold on /tmp, should have plenty of space
	if result.Status == Error {
		t.Errorf("Unexpected error status: %s", result.Message)
	}

	if result.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

// TestDiskSpaceHealthChecker_Warning tests disk checker warning threshold
func TestDiskSpaceHealthChecker_Warning(t *testing.T) {
	// Test with an impossibly high threshold
	checker := NewDiskSpaceHealthChecker("/tmp", 1000000000) // 1TB threshold

	ctx := context.Background()
	result := checker.Check(ctx)

	// Should trigger warning (not enough free space)
	if result.Status != Warning {
		t.Errorf("Expected status Warning with high threshold, got %s", result.Status)
	}

	if result.Message == "" {
		t.Error("Expected message explaining low disk space")
	}
}

// TestDiskSpaceHealthChecker_InvalidPath tests disk checker with invalid path
func TestDiskSpaceHealthChecker_InvalidPath(t *testing.T) {
	checker := NewDiskSpaceHealthChecker("/nonexistent/path/that/does/not/exist", 100)

	ctx := context.Background()
	result := checker.Check(ctx)

	if result.Status != Error {
		t.Errorf("Expected status Error for invalid path, got %s", result.Status)
	}

	if result.Message == "" {
		t.Error("Expected error message for invalid path")
	}
}

// TestDiskSpaceHealthChecker_Values tests disk checker reports actual values
func TestDiskSpaceHealthChecker_Values(t *testing.T) {
	checker := NewDiskSpaceHealthChecker("/tmp", 1)

	ctx := context.Background()
	result := checker.Check(ctx)

	// The message should contain disk space information
	if result.Message == "" {
		t.Error("Expected message with disk space information")
	}

	// Duration should be very fast (disk check is quick)
	if result.Duration > 100*time.Millisecond {
		t.Errorf("Disk space check took too long: %v", result.Duration)
	}
}

// TestHealthChecker_Timeout tests that checkers respect context timeout
func TestHealthChecker_Timeout(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() { _ = db.Close() }()

	checker := NewDatabaseHealthChecker(db)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	start := time.Now()
	result := checker.Check(ctx)
	elapsed := time.Since(start)

	// Should complete quickly even with timeout
	if elapsed > 100*time.Millisecond {
		t.Errorf("Check with timeout took too long: %v", elapsed)
	}

	// Should return a result (not panic or hang)
	if result.Name == "" {
		t.Error("Expected check result even with timeout")
	}
}

// TestHealthChecker_Error tests checker error handling
func TestHealthChecker_Error(t *testing.T) {
	// Create a database and immediately close it
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	_ = db.Close()

	checker := NewDatabaseHealthChecker(db)

	ctx := context.Background()
	result := checker.Check(ctx)

	if result.Status != Error {
		t.Errorf("Expected status Error, got %s", result.Status)
	}

	if result.Message == "" {
		t.Error("Expected error message")
	}

	if result.Name != "database" {
		t.Errorf("Expected name 'database', got '%s'", result.Name)
	}
}

// TestHealthChecker_ConcurrentExecution tests concurrent checker execution
func TestHealthChecker_ConcurrentExecution(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() { _ = db.Close() }()

	checker := NewDatabaseHealthChecker(db)
	ctx := context.Background()

	// Run multiple checks concurrently
	done := make(chan HealthCheckResult, 10)

	for i := 0; i < 10; i++ {
		go func() {
			result := checker.Check(ctx)
			done <- result
		}()
	}

	// Collect results
	for i := 0; i < 10; i++ {
		result := <-done
		if result.Status != Ok {
			t.Errorf("Expected Ok status, got %s", result.Status)
		}
	}
}

// TestMemoryHealthChecker_NoThreshold tests memory checker with zero threshold
func TestMemoryHealthChecker_NoThreshold(t *testing.T) {
	checker := NewMemoryHealthChecker(0)

	ctx := context.Background()
	result := checker.Check(ctx)

	// With zero threshold, any memory usage should trigger warning
	if result.Status != Warning {
		t.Errorf("Expected Warning with zero threshold, got %s", result.Status)
	}
}

// TestDiskSpaceHealthChecker_MultipleChecks tests disk checker stability
func TestDiskSpaceHealthChecker_MultipleChecks(t *testing.T) {
	checker := NewDiskSpaceHealthChecker("/tmp", 100)

	ctx := context.Background()

	// Run check multiple times
	for i := 0; i < 5; i++ {
		result := checker.Check(ctx)

		if result.Name != "disk_space" {
			t.Errorf("Expected name 'disk_space', got '%s'", result.Name)
		}

		// Results should be consistent
		if result.Status == Error {
			t.Errorf("Unexpected error on check %d: %s", i, result.Message)
		}
	}
}
