package metrics

import (
	"sync"
	"time"
)

var (
	globalCollector     *Collector
	globalCollectorOnce sync.Once
)

// GetGlobalCollector returns the global metrics collector instance
func GetGlobalCollector() *Collector {
	globalCollectorOnce.Do(func() {
		globalCollector = NewCollector()
	})
	return globalCollector
}

// Application-specific metric recording functions

// RecordIdeaCreated increments the counter for created ideas
func RecordIdeaCreated() {
	GetGlobalCollector().RecordCounter("ideas_created_total", 1)
}

// RecordIdeaUpdated increments the counter for updated ideas
func RecordIdeaUpdated() {
	GetGlobalCollector().RecordCounter("ideas_updated_total", 1)
}

// RecordIdeaDeleted increments the counter for deleted ideas
func RecordIdeaDeleted() {
	GetGlobalCollector().RecordCounter("ideas_deleted_total", 1)
}

// RecordScoringDuration records the duration of a scoring operation
func RecordScoringDuration(duration time.Duration) {
	GetGlobalCollector().RecordHistogram("scoring_duration_ms", float64(duration.Milliseconds()))
}

// RecordDatabaseQueryDuration records the duration of a database query
func RecordDatabaseQueryDuration(duration time.Duration) {
	GetGlobalCollector().RecordHistogram("database_query_duration_ms", float64(duration.Milliseconds()))
}

// RecordAPIRequestDuration records the duration of an API request
func RecordAPIRequestDuration(duration time.Duration) {
	GetGlobalCollector().RecordHistogram("api_request_duration_ms", float64(duration.Milliseconds()))
}

// RecordHTTPRequest increments the counter for HTTP requests
func RecordHTTPRequest(_, path string, statusCode int) {
	GetGlobalCollector().RecordCounter("http_requests_total", 1)

	// Also track by status code category
	if statusCode >= 200 && statusCode < 300 {
		GetGlobalCollector().RecordCounter("http_requests_2xx", 1)
	} else if statusCode >= 300 && statusCode < 400 {
		GetGlobalCollector().RecordCounter("http_requests_3xx", 1)
	} else if statusCode >= 400 && statusCode < 500 {
		GetGlobalCollector().RecordCounter("http_requests_4xx", 1)
	} else if statusCode >= 500 {
		GetGlobalCollector().RecordCounter("http_requests_5xx", 1)
	}
}

// RecordActiveConnections sets the gauge for active connections
func RecordActiveConnections(count int) {
	GetGlobalCollector().RecordGauge("active_connections", float64(count))
}

// RecordMemoryUsage sets the gauge for memory usage in bytes
func RecordMemoryUsage(bytes uint64) {
	GetGlobalCollector().RecordGauge("memory_usage_bytes", float64(bytes))
}

// RecordGoroutineCount sets the gauge for goroutine count
func RecordGoroutineCount(count int) {
	GetGlobalCollector().RecordGauge("goroutine_count", float64(count))
}

// GetMetrics returns a snapshot of all current metrics
func GetMetrics() map[string]Metric {
	return GetGlobalCollector().GetSnapshot()
}

// ResetMetrics clears all metrics (useful for testing)
func ResetMetrics() {
	GetGlobalCollector().Reset()
}

// ============================================================================
// LLM-SPECIFIC METRICS
// ============================================================================

// RecordLLMRequest tracks an LLM provider request
func RecordLLMRequest(provider string, success bool, duration time.Duration) {
	collector := GetGlobalCollector()

	// Track total requests per provider
	collector.RecordCounter("llm_requests_total_"+provider, 1)

	// Track success/failure
	if success {
		collector.RecordCounter("llm_requests_success_"+provider, 1)
	} else {
		collector.RecordCounter("llm_requests_failure_"+provider, 1)
	}

	// Track request duration (latency)
	collector.RecordHistogram("llm_request_duration_ms_"+provider, float64(duration.Milliseconds()))
}

// RecordLLMTokens tracks token usage for cost estimation
func RecordLLMTokens(provider string, inputTokens, outputTokens int) {
	collector := GetGlobalCollector()

	// Track input tokens
	collector.RecordCounter("llm_input_tokens_"+provider, float64(inputTokens))

	// Track output tokens
	collector.RecordCounter("llm_output_tokens_"+provider, float64(outputTokens))

	// Track total tokens
	collector.RecordCounter("llm_total_tokens_"+provider, float64(inputTokens+outputTokens))
}

// RecordLLMError tracks specific error types
func RecordLLMError(provider string, errorType string) {
	collector := GetGlobalCollector()

	// Track errors by type: timeout, rate_limit, auth_error, network_error, invalid_response, provider_error, unknown
	metricName := "llm_errors_" + provider + "_" + errorType
	collector.RecordCounter(metricName, 1)
}

// RecordLLMCacheHit tracks cache hits/misses
func RecordLLMCacheHit(hit bool) {
	collector := GetGlobalCollector()

	if hit {
		collector.RecordCounter("llm_cache_hits", 1)
	} else {
		collector.RecordCounter("llm_cache_misses", 1)
	}
}

// RecordLLMFallback tracks when a provider fallback occurs
func RecordLLMFallback(fromProvider, toProvider string) {
	collector := GetGlobalCollector()

	// Track fallback events
	metricName := "llm_fallback_" + fromProvider + "_to_" + toProvider
	collector.RecordCounter(metricName, 1)
}
