package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/metrics"
	"github.com/spf13/cobra"
)

// newAnalyticsLLMCommand creates the analytics llm subcommand
func newAnalyticsLLMCommand() *cobra.Command {
	var (
		jsonFormat bool
		since      string
	)

	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Show LLM provider telemetry",
		Long: `Display comprehensive telemetry for LLM providers.

Displays:
- Requests per provider (total, success, failure, success rate %)
- Latency statistics (avg, p50, p95, p99) per provider
- Token usage and estimated costs
- Error breakdown by type
- Cache hit rate
- Fallback statistics
- Comparison table showing all providers side-by-side

Examples:
  tm analytics llm                # Show all LLM metrics
  tm analytics llm --json         # Export as JSON
  tm analytics llm --since 24h    # Show metrics from last 24 hours`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLLMAnalytics(llmAnalyticsOptions{
				jsonFormat: jsonFormat,
				since:      since,
			})
		},
	}

	cmd.Flags().BoolVar(&jsonFormat, "json", false, "Output in JSON format")
	cmd.Flags().StringVar(&since, "since", "", "Show metrics since (e.g., 24h, 7d, 30d)")

	return cmd
}

type llmAnalyticsOptions struct {
	jsonFormat bool
	since      string
}

type llmAnalyticsData struct {
	Providers   []providerMetrics `json:"providers"`
	CacheStats  cacheMetrics      `json:"cache_stats"`
	Fallbacks   []fallbackMetric  `json:"fallbacks"`
	TotalCost   costSummary       `json:"total_cost"`
	GeneratedAt time.Time         `json:"generated_at"`
}

type providerMetrics struct {
	Name          string           `json:"name"`
	TotalRequests int64            `json:"total_requests"`
	SuccessCount  int64            `json:"success_count"`
	FailureCount  int64            `json:"failure_count"`
	SuccessRate   float64          `json:"success_rate"`
	AvgLatencyMs  float64          `json:"avg_latency_ms"`
	P50LatencyMs  float64          `json:"p50_latency_ms"`
	P95LatencyMs  float64          `json:"p95_latency_ms"`
	P99LatencyMs  float64          `json:"p99_latency_ms"`
	InputTokens   int64            `json:"input_tokens"`
	OutputTokens  int64            `json:"output_tokens"`
	EstimatedCost float64          `json:"estimated_cost_usd"`
	Errors        map[string]int64 `json:"errors"`
}

type cacheMetrics struct {
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	HitRate float64 `json:"hit_rate"`
}

type fallbackMetric struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Count int64  `json:"count"`
}

type costSummary struct {
	TotalCostUSD float64            `json:"total_cost_usd"`
	ByProvider   map[string]float64 `json:"by_provider"`
}

func runLLMAnalytics(opts llmAnalyticsOptions) error {
	// Get all metrics
	allMetrics := metrics.GetMetrics()

	// Parse since filter if provided
	var sinceTime time.Time
	if opts.since != "" {
		duration, err := parseLLMDuration(opts.since)
		if err != nil {
			return fmt.Errorf("invalid since value: %w", err)
		}
		sinceTime = time.Now().Add(-duration)
	}

	// Extract LLM-specific metrics
	data := extractLLMMetrics(allMetrics, sinceTime)

	// Output based on format
	if opts.jsonFormat {
		return outputLLMAnalyticsJSON(data)
	}
	return outputLLMAnalyticsText(data)
}

func extractLLMMetrics(allMetrics map[string]metrics.Metric, sinceTime time.Time) llmAnalyticsData {
	data := llmAnalyticsData{
		Providers:   make([]providerMetrics, 0),
		GeneratedAt: time.Now(),
	}

	// Provider names
	providerNames := []string{"ollama", "claude", "openai", "custom", "rule_based"}

	// Build metrics per provider
	providerMap := make(map[string]*providerMetrics)
	for _, name := range providerNames {
		providerMap[name] = &providerMetrics{
			Name:   name,
			Errors: make(map[string]int64),
		}
	}

	// Also track OpenAI variants (e.g., openai_gpt-4)
	for metricName := range allMetrics {
		if strings.HasPrefix(metricName, "llm_requests_total_openai_") {
			providerName := extractProviderFromMetricName(metricName, "llm_requests_total_")
			if _, exists := providerMap[providerName]; !exists {
				providerMap[providerName] = &providerMetrics{
					Name:   providerName,
					Errors: make(map[string]int64),
				}
			}
		}
	}

	// Extract metrics for each provider
	for providerName, pm := range providerMap {
		// Total requests
		if m, ok := allMetrics["llm_requests_total_"+providerName]; ok {
			if sinceTime.IsZero() || m.Timestamp.After(sinceTime) {
				pm.TotalRequests = int64(m.Value)
			}
		}

		// Success count
		if m, ok := allMetrics["llm_requests_success_"+providerName]; ok {
			if sinceTime.IsZero() || m.Timestamp.After(sinceTime) {
				pm.SuccessCount = int64(m.Value)
			}
		}

		// Failure count
		if m, ok := allMetrics["llm_requests_failure_"+providerName]; ok {
			if sinceTime.IsZero() || m.Timestamp.After(sinceTime) {
				pm.FailureCount = int64(m.Value)
			}
		}

		// Calculate success rate
		if pm.TotalRequests > 0 {
			pm.SuccessRate = (float64(pm.SuccessCount) / float64(pm.TotalRequests)) * 100
		}

		// Latency statistics
		latencyMetricName := "llm_request_duration_ms_" + providerName
		if m, ok := allMetrics[latencyMetricName]; ok && m.Type == metrics.Histogram {
			collector := metrics.GetGlobalCollector()
			stats := collector.GetHistogramStats(latencyMetricName)
			if stats != nil {
				pm.AvgLatencyMs = stats.Mean
				pm.P50LatencyMs = stats.P50
				pm.P95LatencyMs = stats.P95
				pm.P99LatencyMs = stats.P99
			}
		}

		// Token counts
		if m, ok := allMetrics["llm_input_tokens_"+providerName]; ok {
			pm.InputTokens = int64(m.Value)
		}
		if m, ok := allMetrics["llm_output_tokens_"+providerName]; ok {
			pm.OutputTokens = int64(m.Value)
		}

		// Calculate cost
		if pm.InputTokens > 0 || pm.OutputTokens > 0 {
			costEstimate := metrics.CalculateCost(providerName, int(pm.InputTokens), int(pm.OutputTokens))
			pm.EstimatedCost = costEstimate.TotalCost
		}

		// Extract errors by type
		errorTypes := []string{"timeout", "rate_limit", "auth_error", "network_error", "invalid_response", "provider_error", "unknown"}
		for _, errorType := range errorTypes {
			errorMetricName := "llm_errors_" + providerName + "_" + errorType
			if m, ok := allMetrics[errorMetricName]; ok {
				pm.Errors[errorType] = int64(m.Value)
			}
		}

		// Only add provider if it has any activity
		if pm.TotalRequests > 0 {
			data.Providers = append(data.Providers, *pm)
		}
	}

	// Sort providers by total requests (descending)
	sort.Slice(data.Providers, func(i, j int) bool {
		return data.Providers[i].TotalRequests > data.Providers[j].TotalRequests
	})

	// Extract cache stats
	if hits, ok := allMetrics["llm_cache_hits"]; ok {
		data.CacheStats.Hits = int64(hits.Value)
	}
	if misses, ok := allMetrics["llm_cache_misses"]; ok {
		data.CacheStats.Misses = int64(misses.Value)
	}
	total := data.CacheStats.Hits + data.CacheStats.Misses
	if total > 0 {
		data.CacheStats.HitRate = (float64(data.CacheStats.Hits) / float64(total)) * 100
	}

	// Extract fallback statistics
	data.Fallbacks = make([]fallbackMetric, 0)
	for metricName, metric := range allMetrics {
		if strings.HasPrefix(metricName, "llm_fallback_") {
			// Parse fallback metric name: llm_fallback_<from>_to_<to>
			parts := strings.TrimPrefix(metricName, "llm_fallback_")
			if toIdx := strings.Index(parts, "_to_"); toIdx != -1 {
				from := parts[:toIdx]
				to := parts[toIdx+4:]
				data.Fallbacks = append(data.Fallbacks, fallbackMetric{
					From:  from,
					To:    to,
					Count: int64(metric.Value),
				})
			}
		}
	}

	// Sort fallbacks by count (descending)
	sort.Slice(data.Fallbacks, func(i, j int) bool {
		return data.Fallbacks[i].Count > data.Fallbacks[j].Count
	})

	// Calculate total cost
	data.TotalCost.ByProvider = make(map[string]float64)
	for _, pm := range data.Providers {
		if pm.EstimatedCost > 0 {
			data.TotalCost.TotalCostUSD += pm.EstimatedCost
			data.TotalCost.ByProvider[pm.Name] = pm.EstimatedCost
		}
	}

	return data
}

func outputLLMAnalyticsText(data llmAnalyticsData) error {
	fmt.Println("LLM Provider Telemetry")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println()

	if len(data.Providers) == 0 {
		fmt.Println("No LLM provider metrics available yet.")
		fmt.Println("Metrics will be collected as you use LLM providers.")
		return nil
	}

	// Summary table
	fmt.Println("Provider Summary:")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("%-15s %10s %10s %10s %12s %10s %10s\n",
		"Provider", "Requests", "Success", "Failure", "Success %", "Avg Lat", "Est. Cost")
	fmt.Println(strings.Repeat("-", 100))

	for _, pm := range data.Providers {
		costStr := "-"
		if pm.EstimatedCost > 0 {
			costStr = formatCost(pm.EstimatedCost)
		}

		fmt.Printf("%-15s %10d %10d %10d %11.1f%% %9.0fms %10s\n",
			pm.Name,
			pm.TotalRequests,
			pm.SuccessCount,
			pm.FailureCount,
			pm.SuccessRate,
			pm.AvgLatencyMs,
			costStr,
		)
	}
	fmt.Println()

	// Detailed metrics for each provider
	for _, pm := range data.Providers {
		fmt.Printf("Provider: %s\n", pm.Name)
		fmt.Println(strings.Repeat("-", 100))

		// Request stats
		fmt.Printf("  Requests:      Total: %d | Success: %d | Failure: %d | Success Rate: %.1f%%\n",
			pm.TotalRequests, pm.SuccessCount, pm.FailureCount, pm.SuccessRate)

		// Latency stats
		if pm.AvgLatencyMs > 0 {
			fmt.Printf("  Latency (ms):  Avg: %.0f | P50: %.0f | P95: %.0f | P99: %.0f\n",
				pm.AvgLatencyMs, pm.P50LatencyMs, pm.P95LatencyMs, pm.P99LatencyMs)
		}

		// Token usage
		if pm.InputTokens > 0 || pm.OutputTokens > 0 {
			fmt.Printf("  Tokens:        Input: %d | Output: %d | Total: %d\n",
				pm.InputTokens, pm.OutputTokens, pm.InputTokens+pm.OutputTokens)
			if pm.EstimatedCost > 0 {
				fmt.Printf("  Cost:          %s USD\n", formatCost(pm.EstimatedCost))
			}
		}

		// Errors
		if len(pm.Errors) > 0 {
			hasErrors := false
			for _, count := range pm.Errors {
				if count > 0 {
					hasErrors = true
					break
				}
			}

			if hasErrors {
				fmt.Print("  Errors:        ")
				first := true
				for errorType, count := range pm.Errors {
					if count > 0 {
						if !first {
							fmt.Print(" | ")
						}
						fmt.Printf("%s: %d", errorType, count)
						first = false
					}
				}
				fmt.Println()
			}
		}

		fmt.Println()
	}

	// Cache statistics
	if data.CacheStats.Hits > 0 || data.CacheStats.Misses > 0 {
		fmt.Println("Cache Statistics:")
		fmt.Println(strings.Repeat("-", 100))
		fmt.Printf("  Hits:     %d\n", data.CacheStats.Hits)
		fmt.Printf("  Misses:   %d\n", data.CacheStats.Misses)
		fmt.Printf("  Hit Rate: %.1f%%\n", data.CacheStats.HitRate)
		fmt.Println()
	}

	// Fallback statistics
	if len(data.Fallbacks) > 0 {
		fmt.Println("Fallback Statistics:")
		fmt.Println(strings.Repeat("-", 100))
		for _, fb := range data.Fallbacks {
			fmt.Printf("  %s → %s: %d times\n", fb.From, fb.To, fb.Count)
		}
		fmt.Println()
	}

	// Total cost summary
	if data.TotalCost.TotalCostUSD > 0 {
		fmt.Println("Cost Summary:")
		fmt.Println(strings.Repeat("-", 100))
		fmt.Printf("  Total Estimated Cost: %s USD\n", formatCost(data.TotalCost.TotalCostUSD))
		fmt.Println()
	}

	fmt.Printf("Generated at: %s\n", data.GeneratedAt.Format(time.RFC1123))
	fmt.Println(strings.Repeat("=", 100))

	return nil
}

func outputLLMAnalyticsJSON(data llmAnalyticsData) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Helper functions

func extractProviderFromMetricName(metricName, prefix string) string {
	return strings.TrimPrefix(metricName, prefix)
}

func formatCost(cost float64) string {
	if cost == 0 {
		return "$0.00"
	}
	if cost < 0.01 {
		return fmt.Sprintf("$%.4f", cost)
	}
	if cost < 1.0 {
		return fmt.Sprintf("$%.3f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

func parseLLMDuration(s string) (time.Duration, error) {
	// Handle simple cases like "24h", "7d", "30d"
	if strings.HasSuffix(s, "h") {
		hours := strings.TrimSuffix(s, "h")
		return time.ParseDuration(hours + "h")
	}
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		var d int
		_, err := fmt.Sscanf(days, "%d", &d)
		if err != nil {
			return 0, err
		}
		return time.Duration(d) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func truncateLLMString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
