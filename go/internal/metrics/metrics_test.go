package metrics

import (
	"testing"
	"time"
)

func TestRecordIdeaCreated(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	RecordIdeaCreated()
	RecordIdeaCreated()

	snapshot := GetMetrics()
	metric, exists := snapshot["ideas_created_total"]
	if !exists {
		t.Fatal("Expected metric 'ideas_created_total' to exist")
	}

	if metric.Value != 2 {
		t.Errorf("Expected value 2, got %v", metric.Value)
	}
}

func TestRecordIdeaUpdated(t *testing.T) {
	ResetMetrics()

	RecordIdeaUpdated()

	snapshot := GetMetrics()
	metric, exists := snapshot["ideas_updated_total"]
	if !exists {
		t.Fatal("Expected metric 'ideas_updated_total' to exist")
	}

	if metric.Value != 1 {
		t.Errorf("Expected value 1, got %v", metric.Value)
	}
}

func TestRecordIdeaDeleted(t *testing.T) {
	ResetMetrics()

	RecordIdeaDeleted()

	snapshot := GetMetrics()
	metric, exists := snapshot["ideas_deleted_total"]
	if !exists {
		t.Fatal("Expected metric 'ideas_deleted_total' to exist")
	}

	if metric.Value != 1 {
		t.Errorf("Expected value 1, got %v", metric.Value)
	}
}

func TestRecordScoringDuration(t *testing.T) {
	ResetMetrics()

	RecordScoringDuration(100 * time.Millisecond)

	snapshot := GetMetrics()
	metric, exists := snapshot["scoring_duration_ms"]
	if !exists {
		t.Fatal("Expected metric 'scoring_duration_ms' to exist")
	}

	if metric.Type != Histogram {
		t.Errorf("Expected type Histogram, got %v", metric.Type)
	}

	if len(metric.Values) != 1 {
		t.Errorf("Expected 1 value, got %d", len(metric.Values))
	}
}

func TestRecordDatabaseQueryDuration(t *testing.T) {
	ResetMetrics()

	RecordDatabaseQueryDuration(50 * time.Millisecond)

	snapshot := GetMetrics()
	metric, exists := snapshot["database_query_duration_ms"]
	if !exists {
		t.Fatal("Expected metric 'database_query_duration_ms' to exist")
	}

	if len(metric.Values) != 1 {
		t.Errorf("Expected 1 value, got %d", len(metric.Values))
	}
}

func TestRecordHTTPRequest(t *testing.T) {
	ResetMetrics()

	RecordHTTPRequest("GET", "/api/v1/ideas", 200)
	RecordHTTPRequest("POST", "/api/v1/ideas", 201)
	RecordHTTPRequest("GET", "/api/v1/ideas", 404)
	RecordHTTPRequest("GET", "/api/v1/ideas", 500)

	snapshot := GetMetrics()

	// Check total requests
	total := snapshot["http_requests_total"]
	if total.Value != 4 {
		t.Errorf("Expected 4 total requests, got %v", total.Value)
	}

	// Check 2xx requests
	success := snapshot["http_requests_2xx"]
	if success.Value != 2 {
		t.Errorf("Expected 2 successful requests, got %v", success.Value)
	}

	// Check 4xx requests
	clientErr := snapshot["http_requests_4xx"]
	if clientErr.Value != 1 {
		t.Errorf("Expected 1 client error, got %v", clientErr.Value)
	}

	// Check 5xx requests
	serverErr := snapshot["http_requests_5xx"]
	if serverErr.Value != 1 {
		t.Errorf("Expected 1 server error, got %v", serverErr.Value)
	}
}

func TestRecordActiveConnections(t *testing.T) {
	ResetMetrics()

	RecordActiveConnections(5)

	snapshot := GetMetrics()
	metric := snapshot["active_connections"]

	if metric.Type != Gauge {
		t.Errorf("Expected type Gauge, got %v", metric.Type)
	}

	if metric.Value != 5 {
		t.Errorf("Expected value 5, got %v", metric.Value)
	}

	// Update gauge
	RecordActiveConnections(10)

	snapshot = GetMetrics()
	metric = snapshot["active_connections"]

	if metric.Value != 10 {
		t.Errorf("Expected updated value 10, got %v", metric.Value)
	}
}

func TestGetGlobalCollector(t *testing.T) {
	collector1 := GetGlobalCollector()
	collector2 := GetGlobalCollector()

	if collector1 != collector2 {
		t.Error("Expected singleton pattern - both collectors should be the same instance")
	}
}
