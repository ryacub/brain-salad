package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestMetricsCollector_RecordCounter(t *testing.T) {
	collector := NewCollector()

	collector.RecordCounter("test_counter", 1.0)
	collector.RecordCounter("test_counter", 2.0)
	collector.RecordCounter("test_counter", 3.0)

	snapshot := collector.GetSnapshot()
	metric, exists := snapshot["test_counter"]
	if !exists {
		t.Fatal("Expected metric 'test_counter' to exist")
	}

	if metric.Type != Counter {
		t.Errorf("Expected metric type Counter, got %v", metric.Type)
	}

	if metric.Value != 6.0 {
		t.Errorf("Expected value 6.0, got %v", metric.Value)
	}

	if metric.Count != 3 {
		t.Errorf("Expected count 3, got %v", metric.Count)
	}
}

func TestMetricsCollector_RecordGauge(t *testing.T) {
	collector := NewCollector()

	collector.RecordGauge("test_gauge", 42.0)
	collector.RecordGauge("test_gauge", 100.0)

	snapshot := collector.GetSnapshot()
	metric, exists := snapshot["test_gauge"]
	if !exists {
		t.Fatal("Expected metric 'test_gauge' to exist")
	}

	if metric.Type != Gauge {
		t.Errorf("Expected metric type Gauge, got %v", metric.Type)
	}

	// Gauge should be set to latest value
	if metric.Value != 100.0 {
		t.Errorf("Expected value 100.0, got %v", metric.Value)
	}
}

func TestMetricsCollector_RecordHistogram(t *testing.T) {
	collector := NewCollector()

	collector.RecordHistogram("test_histogram", 10.0)
	collector.RecordHistogram("test_histogram", 20.0)
	collector.RecordHistogram("test_histogram", 30.0)

	snapshot := collector.GetSnapshot()
	metric, exists := snapshot["test_histogram"]
	if !exists {
		t.Fatal("Expected metric 'test_histogram' to exist")
	}

	if metric.Type != Histogram {
		t.Errorf("Expected metric type Histogram, got %v", metric.Type)
	}

	if metric.Count != 3 {
		t.Errorf("Expected count 3, got %v", metric.Count)
	}
}

func TestMetricsCollector_GetSnapshot(t *testing.T) {
	collector := NewCollector()

	collector.RecordCounter("counter1", 5.0)
	collector.RecordGauge("gauge1", 10.0)
	collector.RecordHistogram("histogram1", 15.0)

	snapshot := collector.GetSnapshot()

	if len(snapshot) != 3 {
		t.Errorf("Expected 3 metrics in snapshot, got %d", len(snapshot))
	}

	// Verify all metrics are present
	if _, exists := snapshot["counter1"]; !exists {
		t.Error("Expected counter1 in snapshot")
	}
	if _, exists := snapshot["gauge1"]; !exists {
		t.Error("Expected gauge1 in snapshot")
	}
	if _, exists := snapshot["histogram1"]; !exists {
		t.Error("Expected histogram1 in snapshot")
	}
}

func TestMetricsCollector_Reset(t *testing.T) {
	collector := NewCollector()

	collector.RecordCounter("test_counter", 10.0)
	collector.RecordGauge("test_gauge", 20.0)

	snapshot1 := collector.GetSnapshot()
	if len(snapshot1) == 0 {
		t.Error("Expected metrics before reset")
	}

	collector.Reset()

	snapshot2 := collector.GetSnapshot()
	if len(snapshot2) != 0 {
		t.Errorf("Expected 0 metrics after reset, got %d", len(snapshot2))
	}
}

func TestMetricsCollector_Concurrency(t *testing.T) {
	collector := NewCollector()
	var wg sync.WaitGroup

	// Test concurrent counter updates
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collector.RecordCounter("concurrent_counter", 1.0)
		}()
	}

	wg.Wait()

	snapshot := collector.GetSnapshot()
	metric, exists := snapshot["concurrent_counter"]
	if !exists {
		t.Fatal("Expected metric 'concurrent_counter' to exist")
	}

	if metric.Value != 100.0 {
		t.Errorf("Expected value 100.0, got %v", metric.Value)
	}

	if metric.Count != 100 {
		t.Errorf("Expected count 100, got %v", metric.Count)
	}
}

func TestMetricsCollector_HistogramStats(t *testing.T) {
	collector := NewCollector()

	values := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	for _, v := range values {
		collector.RecordHistogram("latency", v)
	}

	snapshot := collector.GetSnapshot()
	metric, exists := snapshot["latency"]
	if !exists {
		t.Fatal("Expected metric 'latency' to exist")
	}

	if metric.Count != 10 {
		t.Errorf("Expected count 10, got %v", metric.Count)
	}

	// Verify we can calculate statistics
	stats := collector.GetHistogramStats("latency")
	if stats == nil {
		t.Fatal("Expected histogram stats to exist")
	}

	if stats.Count != 10 {
		t.Errorf("Expected stats count 10, got %v", stats.Count)
	}
}

func TestMetricsCollector_MetricTypes(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*Collector)
		metricName   string
		expectedType MetricType
	}{
		{
			name: "counter type",
			setup: func(c *Collector) {
				c.RecordCounter("test", 1.0)
			},
			metricName:   "test",
			expectedType: Counter,
		},
		{
			name: "gauge type",
			setup: func(c *Collector) {
				c.RecordGauge("test", 1.0)
			},
			metricName:   "test",
			expectedType: Gauge,
		},
		{
			name: "histogram type",
			setup: func(c *Collector) {
				c.RecordHistogram("test", 1.0)
			},
			metricName:   "test",
			expectedType: Histogram,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewCollector()
			tt.setup(collector)

			snapshot := collector.GetSnapshot()
			metric := snapshot[tt.metricName]

			if metric.Type != tt.expectedType {
				t.Errorf("Expected type %v, got %v", tt.expectedType, metric.Type)
			}
		})
	}
}

func TestMetricsCollector_Timestamp(t *testing.T) {
	collector := NewCollector()

	beforeRecord := time.Now()
	collector.RecordCounter("test", 1.0)
	afterRecord := time.Now()

	snapshot := collector.GetSnapshot()
	metric := snapshot["test"]

	if metric.Timestamp.Before(beforeRecord) || metric.Timestamp.After(afterRecord) {
		t.Errorf("Timestamp not within expected range: got %v, expected between %v and %v",
			metric.Timestamp, beforeRecord, afterRecord)
	}
}
