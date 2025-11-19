//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/api"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
	"github.com/stretchr/testify/require"
)

// LoadTestMetrics tracks performance metrics during load testing
type LoadTestMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalLatency    int64 // microseconds
	MinLatency      int64 // microseconds
	MaxLatency      int64 // microseconds
	mu              sync.Mutex
}

func (m *LoadTestMetrics) RecordRequest(latency time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.TotalRequests, 1)
	latencyMicros := latency.Microseconds()
	atomic.AddInt64(&m.TotalLatency, latencyMicros)

	if success {
		atomic.AddInt64(&m.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&m.FailedRequests, 1)
	}

	if m.MinLatency == 0 || latencyMicros < m.MinLatency {
		m.MinLatency = latencyMicros
	}
	if latencyMicros > m.MaxLatency {
		m.MaxLatency = latencyMicros
	}
}

func (m *LoadTestMetrics) Report(t *testing.T) {
	t.Logf("=== Load Test Results ===")
	t.Logf("Total Requests: %d", m.TotalRequests)
	t.Logf("Successful: %d (%.2f%%)", m.SuccessRequests, float64(m.SuccessRequests)/float64(m.TotalRequests)*100)
	t.Logf("Failed: %d (%.2f%%)", m.FailedRequests, float64(m.FailedRequests)/float64(m.TotalRequests)*100)

	if m.TotalRequests > 0 {
		avgLatency := float64(m.TotalLatency) / float64(m.TotalRequests) / 1000.0
		t.Logf("Average Latency: %.2f ms", avgLatency)
		t.Logf("Min Latency: %.2f ms", float64(m.MinLatency)/1000.0)
		t.Logf("Max Latency: %.2f ms", float64(m.MaxLatency)/1000.0)
		t.Logf("Throughput: %.2f req/sec", float64(m.TotalRequests)/10.0) // 10 second test
	}
	t.Logf("========================")
}

// TestLoadCreateIdeas tests sustained load of creating ideas
func TestLoadCreateIdeas(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Setup
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "load_test.db")

	telosPath := filepath.Join(tempDir, "telos.md")
	telosContent := `# Telos: Load Test
## Goals
- G1: Performance
- G2: Scalability
- G3: Reliability
## Strategies
- Efficient algorithms
- Database optimization
- Caching strategies
## Anti-Patterns
- N+1 queries
- Memory leaks
- Blocking operations
`
	require.NoError(t, os.WriteFile(telosPath, []byte(telosContent), 0644))

	telosConfig, err := telos.ParseTelosFile(telosPath)
	require.NoError(t, err)

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	server := api.NewServer(repo, telosConfig)
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Load test parameters
	const (
		numWorkers      = 10
		requestsPerWorker = 50
		testDuration    = 10 * time.Second
	)

	metrics := &LoadTestMetrics{}

	t.Run("SustainedCreateLoad", func(t *testing.T) {
		var wg sync.WaitGroup
		stopChan := make(chan struct{})

		// Start workers
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				requestCount := 0
				for {
					select {
					case <-stopChan:
						return
					default:
						start := time.Now()

						createReq := api.CreateIdeaRequest{
							Content: fmt.Sprintf("Load test idea from worker %d, request %d", workerID, requestCount),
						}
						body, err := json.Marshal(createReq)
						if err != nil {
							metrics.RecordRequest(time.Since(start), false)
							continue
						}

						resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
						success := err == nil && resp != nil && resp.StatusCode == http.StatusCreated

						if resp != nil {
							resp.Body.Close()
						}

						metrics.RecordRequest(time.Since(start), success)
						requestCount++

						if requestCount >= requestsPerWorker {
							return
						}
					}
				}
			}(i)
		}

		// Let it run for test duration or until all workers finish
		time.AfterFunc(testDuration, func() {
			close(stopChan)
		})

		wg.Wait()
		metrics.Report(t)

		// Verify success rate
		successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests)
		require.Greater(t, successRate, 0.95, "Success rate should be above 95%%")
	})
}

// TestLoadMixedOperations tests mixed read/write operations
func TestLoadMixedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Setup
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "mixed_load_test.db")

	telosPath := filepath.Join(tempDir, "telos.md")
	telosContent := `# Telos
## Goals
- G1: Goal 1
- G2: Goal 2
`
	require.NoError(t, os.WriteFile(telosPath, []byte(telosContent), 0644))

	telosConfig, err := telos.ParseTelosFile(telosPath)
	require.NoError(t, err)

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	server := api.NewServer(repo, telosConfig)
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Pre-populate database with some ideas
	ideaIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		createReq := api.CreateIdeaRequest{
			Content: fmt.Sprintf("Pre-populated idea %d", i),
		}
		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)

		var idea api.IdeaResponse
		err = json.NewDecoder(resp.Body).Decode(&idea)
		resp.Body.Close()
		require.NoError(t, err)

		ideaIDs[i] = idea.ID
	}

	const (
		numWorkers   = 10
		testDuration = 10 * time.Second
	)

	metrics := &LoadTestMetrics{}

	t.Run("MixedReadWriteLoad", func(t *testing.T) {
		var wg sync.WaitGroup
		stopChan := make(chan struct{})

		// Start workers with mixed operations
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				requestCount := 0
				for {
					select {
					case <-stopChan:
						return
					default:
						start := time.Now()
						var success bool

						// Randomly choose operation type
						opType := requestCount % 4
						switch opType {
						case 0: // Create
							createReq := api.CreateIdeaRequest{
								Content: fmt.Sprintf("Load test idea from worker %d", workerID),
							}
							body, _ := json.Marshal(createReq)
							resp, err := http.Post(ts.URL+"/api/v1/ideas", "application/json", bytes.NewBuffer(body))
							success = err == nil && resp != nil && resp.StatusCode == http.StatusCreated
							if resp != nil {
								resp.Body.Close()
							}

						case 1: // Read single
							ideaID := ideaIDs[requestCount%len(ideaIDs)]
							resp, err := http.Get(ts.URL + "/api/v1/ideas/" + ideaID)
							success = err == nil && resp != nil && resp.StatusCode == http.StatusOK
							if resp != nil {
								resp.Body.Close()
							}

						case 2: // List
							resp, err := http.Get(ts.URL + "/api/v1/ideas?limit=10")
							success = err == nil && resp != nil && resp.StatusCode == http.StatusOK
							if resp != nil {
								resp.Body.Close()
							}

						case 3: // Update
							ideaID := ideaIDs[requestCount%len(ideaIDs)]
							newStatus := "archived"
							updateReq := api.UpdateIdeaRequest{Status: &newStatus}
							body, _ := json.Marshal(updateReq)
							req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/ideas/"+ideaID, bytes.NewBuffer(body))
							req.Header.Set("Content-Type", "application/json")
							client := &http.Client{}
							resp, err := client.Do(req)
							success = err == nil && resp != nil && resp.StatusCode == http.StatusOK
							if resp != nil {
								resp.Body.Close()
							}
						}

						metrics.RecordRequest(time.Since(start), success)
						requestCount++
					}
				}
			}(i)
		}

		// Run for test duration
		time.AfterFunc(testDuration, func() {
			close(stopChan)
		})

		wg.Wait()
		metrics.Report(t)

		// Verify success rate
		successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests)
		require.Greater(t, successRate, 0.95, "Success rate should be above 95%%")
	})
}

// TestDatabasePerformance tests database query performance
func TestDatabasePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "perf_test.db")

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	// Insert a large number of ideas
	const numIdeas = 1000
	t.Logf("Inserting %d ideas...", numIdeas)

	insertStart := time.Now()
	for i := 0; i < numIdeas; i++ {
		idea := &models.Idea{
			ID:         fmt.Sprintf("idea-%d", i),
			Content:    fmt.Sprintf("Performance test idea %d", i),
			RawScore:   float64(i % 10),
			FinalScore: float64(i%10) + 0.5,
			Patterns:   []string{"pattern1"},
			Status:     "active",
			CreatedAt:  time.Now().UTC(),
		}
		err := repo.Create(idea)
		require.NoError(t, err)
	}
	insertDuration := time.Since(insertStart)
	t.Logf("Inserted %d ideas in %v (%.2f ideas/sec)", numIdeas, insertDuration, float64(numIdeas)/insertDuration.Seconds())

	// Test query performance
	t.Run("ListAllPerformance", func(t *testing.T) {
		start := time.Now()
		ideas, err := repo.List(database.ListOptions{})
		duration := time.Since(start)

		require.NoError(t, err)
		require.Equal(t, numIdeas, len(ideas))
		t.Logf("Listed %d ideas in %v", len(ideas), duration)

		// Should be reasonably fast (< 100ms for 1000 records)
		require.Less(t, duration, 100*time.Millisecond, "List query should complete quickly")
	})

	t.Run("FilteredQueryPerformance", func(t *testing.T) {
		minScore := 5.0
		start := time.Now()
		ideas, err := repo.List(database.ListOptions{
			Status:   "active",
			MinScore: &minScore,
		})
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Filtered query returned %d ideas in %v", len(ideas), duration)

		// Filtered queries should also be fast
		require.Less(t, duration, 50*time.Millisecond, "Filtered query should complete quickly")
	})

	t.Run("GetByIDPerformance", func(t *testing.T) {
		const numLookups = 100
		start := time.Now()

		for i := 0; i < numLookups; i++ {
			ideaID := fmt.Sprintf("idea-%d", i)
			_, err := repo.GetByID(ideaID)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		avgLookup := duration / numLookups
		t.Logf("Average GetByID latency: %v", avgLookup)

		// ID lookups should be very fast
		require.Less(t, avgLookup, 1*time.Millisecond, "ID lookup should be fast")
	})
}

// TestConcurrentDatabaseOperations tests database concurrency
func TestConcurrentDatabaseOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "concurrent_test.db")

	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	const (
		numWriters = 5
		numReaders = 10
		opsPerWorker = 20
	)

	var wg sync.WaitGroup
	errors := make(chan error, (numWriters+numReaders)*opsPerWorker)

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < opsPerWorker; j++ {
				idea := &models.Idea{
					ID:         fmt.Sprintf("writer-%d-idea-%d", workerID, j),
					Content:    fmt.Sprintf("Concurrent write from worker %d, op %d", workerID, j),
					RawScore:   float64(j),
					FinalScore: float64(j) + 0.5,
					Status:     "active",
					CreatedAt:  time.Now().UTC(),
				}

				if err := repo.Create(idea); err != nil {
					errors <- fmt.Errorf("writer %d failed: %w", workerID, err)
				}
			}
		}(i)
	}

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < opsPerWorker; j++ {
				// Try to read random ideas
				ideaID := fmt.Sprintf("writer-%d-idea-%d", workerID%numWriters, j%opsPerWorker)
				_, err := repo.GetByID(ideaID)
				// It's okay if idea doesn't exist yet (race condition)
				if err != nil && !isNotFoundError(err) {
					errors <- fmt.Errorf("reader %d failed: %w", workerID, err)
				}

				// Also list ideas
				_, err = repo.List(database.ListOptions{})
				if err != nil {
					errors <- fmt.Errorf("reader %d list failed: %w", workerID, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
	}

	require.Equal(t, 0, errorCount, "Should have no concurrent operation errors")

	// Verify all ideas were created
	ideas, err := repo.List(database.ListOptions{})
	require.NoError(t, err)
	require.Equal(t, numWriters*opsPerWorker, len(ideas))
}

func isNotFoundError(err error) bool {
	return err != nil && (err.Error() == "idea not found" ||
		bytes.Contains([]byte(err.Error()), []byte("not found")))
}
