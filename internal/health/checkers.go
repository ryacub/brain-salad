// Package health provides health check functionality for database and system resource monitoring.
package health

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"syscall"
	"time"
)

const (
	// Health check names
	checkNameDatabase  = "database"
	checkNameMemory    = "memory"
	checkNameDiskSpace = "disk_space"
)

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	db *sql.DB
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(db *sql.DB) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{db: db}
}

// Name returns the name of the checker
func (d *DatabaseHealthChecker) Name() string {
	return checkNameDatabase
}

// Check performs the database health check
func (d *DatabaseHealthChecker) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()

	// Try to ping the database with context
	err := d.db.PingContext(ctx)

	result := HealthCheckResult{
		Name:      d.Name(),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Status = Error
		result.Message = fmt.Sprintf("Database ping failed: %v", err)
	} else {
		result.Status = Ok
		result.Message = "Database connection is healthy"
	}

	return result
}

// MemoryHealthChecker checks memory usage
type MemoryHealthChecker struct {
	thresholdMB float64
}

// NewMemoryHealthChecker creates a new memory health checker
// thresholdMB is the warning threshold in megabytes
func NewMemoryHealthChecker(thresholdMB float64) *MemoryHealthChecker {
	return &MemoryHealthChecker{thresholdMB: thresholdMB}
}

// Name returns the name of the checker
func (m *MemoryHealthChecker) Name() string {
	return checkNameMemory
}

// Check performs the memory health check
func (m *MemoryHealthChecker) Check(_ context.Context) HealthCheckResult {
	start := time.Now()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Convert bytes to megabytes
	allocMB := float64(memStats.Alloc) / 1024 / 1024
	sysMB := float64(memStats.Sys) / 1024 / 1024

	result := HealthCheckResult{
		Name:      m.Name(),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	if allocMB > m.thresholdMB {
		result.Status = Warning
		result.Message = fmt.Sprintf("High memory usage: %.2f MB allocated (threshold: %.2f MB), %.2f MB system",
			allocMB, m.thresholdMB, sysMB)
	} else {
		result.Status = Ok
		result.Message = fmt.Sprintf("Memory usage normal: %.2f MB allocated, %.2f MB system",
			allocMB, sysMB)
	}

	return result
}

// DiskSpaceHealthChecker checks available disk space
type DiskSpaceHealthChecker struct {
	path        string
	thresholdMB uint64
}

// NewDiskSpaceHealthChecker creates a new disk space health checker
// path is the directory path to check
// thresholdMB is the minimum free space in megabytes
func NewDiskSpaceHealthChecker(path string, thresholdMB uint64) *DiskSpaceHealthChecker {
	return &DiskSpaceHealthChecker{
		path:        path,
		thresholdMB: thresholdMB,
	}
}

// Name returns the name of the checker
func (d *DiskSpaceHealthChecker) Name() string {
	return checkNameDiskSpace
}

// Check performs the disk space health check
func (d *DiskSpaceHealthChecker) Check(_ context.Context) HealthCheckResult {
	start := time.Now()

	result := HealthCheckResult{
		Name:      d.Name(),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	var stat syscall.Statfs_t
	err := syscall.Statfs(d.path, &stat)

	if err != nil {
		result.Status = Error
		result.Message = fmt.Sprintf("Failed to check disk space for %s: %v", d.path, err)
		return result
	}

	// Calculate available space in MB
	availableMB := (stat.Bavail * uint64(stat.Bsize)) / 1024 / 1024
	totalMB := (stat.Blocks * uint64(stat.Bsize)) / 1024 / 1024

	if availableMB < d.thresholdMB {
		result.Status = Warning
		result.Message = fmt.Sprintf("Low disk space on %s: %d MB available (threshold: %d MB), %d MB total",
			d.path, availableMB, d.thresholdMB, totalMB)
	} else {
		result.Status = Ok
		result.Message = fmt.Sprintf("Disk space OK on %s: %d MB available, %d MB total",
			d.path, availableMB, totalMB)
	}

	result.Duration = time.Since(start)
	return result
}
