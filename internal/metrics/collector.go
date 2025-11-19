package metrics

import (
	"sort"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	Counter   MetricType = "counter"
	Gauge     MetricType = "gauge"
	Histogram MetricType = "histogram"
)

// Metric represents a single metric with its metadata
type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Count     int64
	Timestamp time.Time
	Values    []float64 // For histograms
}

// HistogramStats contains statistical data for histogram metrics
type HistogramStats struct {
	Count int64
	Min   float64
	Max   float64
	Mean  float64
	P50   float64
	P95   float64
	P99   float64
}

// Collector manages metrics collection in a thread-safe manner
type Collector struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		metrics: make(map[string]*Metric),
	}
}

// RecordCounter increments a counter metric
func (c *Collector) RecordCounter(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, exists := c.metrics[name]; exists {
		m.Value += value
		m.Count++
		m.Timestamp = time.Now()
	} else {
		c.metrics[name] = &Metric{
			Name:      name,
			Type:      Counter,
			Value:     value,
			Count:     1,
			Timestamp: time.Now(),
		}
	}
}

// RecordGauge sets a gauge metric to a specific value
func (c *Collector) RecordGauge(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, exists := c.metrics[name]; exists {
		m.Value = value
		m.Count++
		m.Timestamp = time.Now()
	} else {
		c.metrics[name] = &Metric{
			Name:      name,
			Type:      Gauge,
			Value:     value,
			Count:     1,
			Timestamp: time.Now(),
		}
	}
}

// RecordHistogram records a value in a histogram metric
func (c *Collector) RecordHistogram(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, exists := c.metrics[name]; exists {
		m.Values = append(m.Values, value)
		m.Count++
		m.Timestamp = time.Now()
	} else {
		c.metrics[name] = &Metric{
			Name:      name,
			Type:      Histogram,
			Values:    []float64{value},
			Count:     1,
			Timestamp: time.Now(),
		}
	}
}

// GetSnapshot returns a snapshot of all metrics
func (c *Collector) GetSnapshot() map[string]Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshot := make(map[string]Metric, len(c.metrics))
	for name, metric := range c.metrics {
		// Deep copy the metric
		m := Metric{
			Name:      metric.Name,
			Type:      metric.Type,
			Value:     metric.Value,
			Count:     metric.Count,
			Timestamp: metric.Timestamp,
		}
		if metric.Values != nil {
			m.Values = make([]float64, len(metric.Values))
			copy(m.Values, metric.Values)
		}
		snapshot[name] = m
	}
	return snapshot
}

// Reset clears all metrics
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = make(map[string]*Metric)
}

// GetHistogramStats calculates statistics for a histogram metric
func (c *Collector) GetHistogramStats(name string) *HistogramStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metric, exists := c.metrics[name]
	if !exists || metric.Type != Histogram || len(metric.Values) == 0 {
		return nil
	}

	// Sort values for percentile calculations
	sorted := make([]float64, len(metric.Values))
	copy(sorted, metric.Values)
	sort.Float64s(sorted)

	stats := &HistogramStats{
		Count: metric.Count,
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
	}

	// Calculate mean
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	stats.Mean = sum / float64(len(sorted))

	// Calculate percentiles
	stats.P50 = percentile(sorted, 0.50)
	stats.P95 = percentile(sorted, 0.95)
	stats.P99 = percentile(sorted, 0.99)

	return stats
}

// percentile calculates the value at a given percentile
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := int(float64(len(sorted)) * p)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}
