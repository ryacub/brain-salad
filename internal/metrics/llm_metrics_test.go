package metrics

import (
	"testing"
	"time"
)

func TestRecordLLMRequest(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	provider := "claude"
	duration := 500 * time.Millisecond

	// Test successful request
	RecordLLMRequest(provider, true, duration)

	metrics := GetMetrics()

	// Check total requests counter
	if m, ok := metrics["llm_requests_total_"+provider]; !ok {
		t.Errorf("Expected llm_requests_total_%s to exist", provider)
	} else if m.Value != 1 {
		t.Errorf("Expected llm_requests_total_%s = 1, got %f", provider, m.Value)
	}

	// Check success counter
	if m, ok := metrics["llm_requests_success_"+provider]; !ok {
		t.Errorf("Expected llm_requests_success_%s to exist", provider)
	} else if m.Value != 1 {
		t.Errorf("Expected llm_requests_success_%s = 1, got %f", provider, m.Value)
	}

	// Check duration histogram
	if m, ok := metrics["llm_request_duration_ms_"+provider]; !ok {
		t.Errorf("Expected llm_request_duration_ms_%s to exist", provider)
	} else if m.Type != Histogram {
		t.Errorf("Expected histogram type for latency metric")
	}

	// Test failed request
	RecordLLMRequest(provider, false, duration)

	metrics = GetMetrics()

	// Check failure counter
	if m, ok := metrics["llm_requests_failure_"+provider]; !ok {
		t.Errorf("Expected llm_requests_failure_%s to exist", provider)
	} else if m.Value != 1 {
		t.Errorf("Expected llm_requests_failure_%s = 1, got %f", provider, m.Value)
	}

	// Check total is now 2
	if m, ok := metrics["llm_requests_total_"+provider]; !ok {
		t.Errorf("Expected llm_requests_total_%s to exist", provider)
	} else if m.Value != 2 {
		t.Errorf("Expected llm_requests_total_%s = 2, got %f", provider, m.Value)
	}
}

func TestRecordLLMTokens(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	provider := "openai"
	inputTokens := 1000
	outputTokens := 500

	RecordLLMTokens(provider, inputTokens, outputTokens)

	metrics := GetMetrics()

	// Check input tokens
	if m, ok := metrics["llm_input_tokens_"+provider]; !ok {
		t.Errorf("Expected llm_input_tokens_%s to exist", provider)
	} else if m.Value != float64(inputTokens) {
		t.Errorf("Expected llm_input_tokens_%s = %d, got %f", provider, inputTokens, m.Value)
	}

	// Check output tokens
	if m, ok := metrics["llm_output_tokens_"+provider]; !ok {
		t.Errorf("Expected llm_output_tokens_%s to exist", provider)
	} else if m.Value != float64(outputTokens) {
		t.Errorf("Expected llm_output_tokens_%s = %d, got %f", provider, outputTokens, m.Value)
	}

	// Check total tokens
	if m, ok := metrics["llm_total_tokens_"+provider]; !ok {
		t.Errorf("Expected llm_total_tokens_%s to exist", provider)
	} else if m.Value != float64(inputTokens+outputTokens) {
		t.Errorf("Expected llm_total_tokens_%s = %d, got %f", provider, inputTokens+outputTokens, m.Value)
	}

	// Test accumulation
	RecordLLMTokens(provider, inputTokens, outputTokens)

	metrics = GetMetrics()

	// Check that tokens are accumulated
	if m, ok := metrics["llm_total_tokens_"+provider]; !ok {
		t.Errorf("Expected llm_total_tokens_%s to exist", provider)
	} else if m.Value != float64((inputTokens+outputTokens)*2) {
		t.Errorf("Expected llm_total_tokens_%s = %d, got %f", provider, (inputTokens+outputTokens)*2, m.Value)
	}
}

func TestRecordLLMError(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	provider := "claude"
	errorType := "timeout"

	RecordLLMError(provider, errorType)

	metrics := GetMetrics()

	metricName := "llm_errors_" + provider + "_" + errorType
	if m, ok := metrics[metricName]; !ok {
		t.Errorf("Expected %s to exist", metricName)
	} else if m.Value != 1 {
		t.Errorf("Expected %s = 1, got %f", metricName, m.Value)
	}

	// Test multiple error types
	RecordLLMError(provider, "rate_limit")
	RecordLLMError(provider, "timeout") // Another timeout

	metrics = GetMetrics()

	// Check timeout is now 2
	if m, ok := metrics[metricName]; !ok {
		t.Errorf("Expected %s to exist", metricName)
	} else if m.Value != 2 {
		t.Errorf("Expected %s = 2, got %f", metricName, m.Value)
	}

	// Check rate_limit is 1
	rateLimitMetric := "llm_errors_" + provider + "_rate_limit"
	if m, ok := metrics[rateLimitMetric]; !ok {
		t.Errorf("Expected %s to exist", rateLimitMetric)
	} else if m.Value != 1 {
		t.Errorf("Expected %s = 1, got %f", rateLimitMetric, m.Value)
	}
}

func TestRecordLLMCacheHit(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	// Record hits and misses
	RecordLLMCacheHit(true)
	RecordLLMCacheHit(true)
	RecordLLMCacheHit(false)

	metrics := GetMetrics()

	// Check hits
	if m, ok := metrics["llm_cache_hits"]; !ok {
		t.Error("Expected llm_cache_hits to exist")
	} else if m.Value != 2 {
		t.Errorf("Expected llm_cache_hits = 2, got %f", m.Value)
	}

	// Check misses
	if m, ok := metrics["llm_cache_misses"]; !ok {
		t.Error("Expected llm_cache_misses to exist")
	} else if m.Value != 1 {
		t.Errorf("Expected llm_cache_misses = 1, got %f", m.Value)
	}
}

func TestRecordLLMFallback(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	fromProvider := "ollama"
	toProvider := "claude"

	RecordLLMFallback(fromProvider, toProvider)

	metrics := GetMetrics()

	metricName := "llm_fallback_" + fromProvider + "_to_" + toProvider
	if m, ok := metrics[metricName]; !ok {
		t.Errorf("Expected %s to exist", metricName)
	} else if m.Value != 1 {
		t.Errorf("Expected %s = 1, got %f", metricName, m.Value)
	}

	// Test accumulation
	RecordLLMFallback(fromProvider, toProvider)

	metrics = GetMetrics()

	if m, ok := metrics[metricName]; !ok {
		t.Errorf("Expected %s to exist", metricName)
	} else if m.Value != 2 {
		t.Errorf("Expected %s = 2, got %f", metricName, m.Value)
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		inputTokens  int
		outputTokens int
		wantCost     float64
	}{
		{
			name:         "Claude cost calculation",
			provider:     "claude",
			inputTokens:  1_000_000,
			outputTokens: 1_000_000,
			wantCost:     18.0, // $3 + $15
		},
		{
			name:         "OpenAI cost calculation",
			provider:     "openai",
			inputTokens:  1_000_000,
			outputTokens: 1_000_000,
			wantCost:     90.0, // $30 + $60
		},
		{
			name:         "Ollama should be free",
			provider:     "ollama",
			inputTokens:  1_000_000,
			outputTokens: 1_000_000,
			wantCost:     0.0,
		},
		{
			name:         "Rule-based should be free",
			provider:     "rule_based",
			inputTokens:  1_000_000,
			outputTokens: 1_000_000,
			wantCost:     0.0,
		},
		{
			name:         "Small token count",
			provider:     "claude",
			inputTokens:  1000,
			outputTokens: 500,
			wantCost:     0.0105, // ($3/1M * 1000) + ($15/1M * 500) = 0.003 + 0.0075
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateCost(tt.provider, tt.inputTokens, tt.outputTokens)
			// Use tolerance for floating point comparison
			tolerance := 0.0001
			if diff := cost.TotalCost - tt.wantCost; diff < -tolerance || diff > tolerance {
				t.Errorf("CalculateCost() = %.6f, want %.6f", cost.TotalCost, tt.wantCost)
			}
			if cost.Provider != tt.provider {
				t.Errorf("CalculateCost() provider = %s, want %s", cost.Provider, tt.provider)
			}
		})
	}
}

func TestNormalizeProviderName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"claude", "claude"},
		{"claude-3.5-sonnet", "claude"},
		{"openai", "openai"},
		{"openai_gpt-4", "openai"},
		{"ollama", "ollama"},
		{"custom", "custom"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeProviderName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeProviderName(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestCostEstimateFormatting(t *testing.T) {
	tests := []struct {
		name string
		cost CostEstimate
		want string
	}{
		{
			name: "Zero cost",
			cost: CostEstimate{TotalCost: 0.0},
			want: "$0.00 (free)",
		},
		{
			name: "Very small cost",
			cost: CostEstimate{TotalCost: 0.0012},
			want: "$0.0012",
		},
		{
			name: "Small cost",
			cost: CostEstimate{TotalCost: 0.123},
			want: "$0.123",
		},
		{
			name: "Medium cost",
			cost: CostEstimate{TotalCost: 1.23},
			want: "$1.23",
		},
		{
			name: "Large cost",
			cost: CostEstimate{TotalCost: 123.45},
			want: "$123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cost.FormatCost()
			if got != tt.want {
				t.Errorf("FormatCost() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestConcurrentMetricRecording(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	provider := "claude"
	numGoroutines := 100
	requestsPerGoroutine := 10

	// Record metrics concurrently
	done := make(chan bool)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < requestsPerGoroutine; j++ {
				RecordLLMRequest(provider, true, 100*time.Millisecond)
				RecordLLMTokens(provider, 100, 50)
				RecordLLMError(provider, "timeout")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	metrics := GetMetrics()

	expectedRequests := numGoroutines * requestsPerGoroutine
	if m, ok := metrics["llm_requests_total_"+provider]; !ok {
		t.Error("Expected llm_requests_total to exist")
	} else if int(m.Value) != expectedRequests {
		t.Errorf("Expected llm_requests_total = %d, got %f", expectedRequests, m.Value)
	}

	expectedTokens := numGoroutines * requestsPerGoroutine * 150 // 100 + 50
	if m, ok := metrics["llm_total_tokens_"+provider]; !ok {
		t.Error("Expected llm_total_tokens to exist")
	} else if int(m.Value) != expectedTokens {
		t.Errorf("Expected llm_total_tokens = %d, got %f", expectedTokens, m.Value)
	}
}

func TestHistogramStatistics(t *testing.T) {
	// Reset metrics before test
	ResetMetrics()

	provider := "test_provider"

	// Record a series of latencies
	latencies := []int{100, 150, 200, 250, 300, 350, 400, 450, 500, 1000}
	for _, latency := range latencies {
		RecordLLMRequest(provider, true, time.Duration(latency)*time.Millisecond)
	}

	collector := GetGlobalCollector()
	stats := collector.GetHistogramStats("llm_request_duration_ms_" + provider)

	if stats == nil {
		t.Fatal("Expected histogram stats to exist")
	}

	// Check count
	if stats.Count != int64(len(latencies)) {
		t.Errorf("Expected count = %d, got %d", len(latencies), stats.Count)
	}

	// Check min and max
	if stats.Min != 100 {
		t.Errorf("Expected min = 100, got %f", stats.Min)
	}
	if stats.Max != 1000 {
		t.Errorf("Expected max = 1000, got %f", stats.Max)
	}

	// Check mean (sum = 3700, count = 10, mean = 370)
	expectedMean := 370.0
	if stats.Mean != expectedMean {
		t.Errorf("Expected mean = %f, got %f", expectedMean, stats.Mean)
	}

	// Check percentiles (approximate)
	if stats.P50 < 250 || stats.P50 > 350 {
		t.Errorf("Expected P50 to be around 300, got %f", stats.P50)
	}
	if stats.P95 < 900 || stats.P95 > 1000 {
		t.Errorf("Expected P95 to be around 1000, got %f", stats.P95)
	}
}

func TestEstimateMonthlyCost(t *testing.T) {
	dailyInput := 1_000_000  // 1M tokens per day
	dailyOutput := 500_000   // 500K tokens per day

	monthlyCost := EstimateMonthlyCost(dailyInput, dailyOutput, "claude")

	// Expected: (1M * $3/M + 0.5M * $15/M) * 30 = (3 + 7.5) * 30 = 315
	expected := 315.0

	if monthlyCost != expected {
		t.Errorf("EstimateMonthlyCost() = %f, want %f", monthlyCost, expected)
	}
}
