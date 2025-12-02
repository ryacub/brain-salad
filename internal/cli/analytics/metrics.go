package analytics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ryacub/telos-idea-matrix/internal/analytics"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/spf13/cobra"
)

type metricsOptions struct {
	format  string
	verbose bool
}

// NewMetricsCommand creates the analytics metrics subcommand
func NewMetricsCommand(getContext func() *CLIContext) *cobra.Command {
	var (
		format  string
		verbose bool
	)

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show comprehensive system metrics",
		Long: `Display system-wide metrics and statistics.

Provides insights into:
- Total counts (ideas, patterns)
- Status distribution
- Score distribution
- Pattern frequency
- Activity timeline
- Database health

Examples:
  # Show basic system metrics
  tm analytics metrics

  # Show verbose metrics with details
  tm analytics metrics --verbose

  # Export as JSON
  tm analytics metrics --format json

  # Export as CSV
  tm analytics metrics --format csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSystemMetrics(getContext, metricsOptions{
				format:  format,
				verbose: verbose,
			})
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format: text|json|csv")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed metrics")

	return cmd
}

func runSystemMetrics(getContext func() *CLIContext, opts metricsOptions) error {
	ctx := getContext()
	if ctx == nil {
		return fmt.Errorf("CLI context not initialized")
	}

	// Fetch all ideas (not just active)
	ideas, err := ctx.Repository.List(database.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch ideas: %w", err)
	}

	if len(ideas) == 0 {
		fmt.Println("No ideas found in the system.")
		return nil
	}

	// Calculate metrics using service
	service := analytics.NewServiceWithDB(ctx.Repository, ctx.DBPath)
	metrics := service.CalculateSystemMetrics(ideas)

	// Output based on format
	switch opts.format {
	case "json":
		return outputMetricsJSON(metrics)
	case "csv":
		return outputMetricsCSV(metrics)
	default:
		return outputMetricsText(metrics, opts)
	}
}

// Output formatters

func outputMetricsText(metrics analytics.SystemMetrics, opts metricsOptions) error {
	fmt.Println("System Metrics")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Overview
	fmt.Println("Overview:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("  Total Ideas:      %d\n", metrics.Overview.TotalIdeas)
	fmt.Printf("  Total Patterns:   %d\n", metrics.Overview.TotalPatterns)
	fmt.Printf("  Average Score:    %.2f\n", metrics.Overview.AverageScore)
	fmt.Printf("  Median Score:     %.2f\n", metrics.Overview.MedianScore)
	fmt.Printf("  Highest Score:    %.2f\n", metrics.Overview.HighestScore)
	fmt.Printf("  Lowest Score:     %.2f\n", metrics.Overview.LowestScore)
	fmt.Println()

	// Status Breakdown
	fmt.Println("Status Breakdown:")
	fmt.Println(strings.Repeat("-", 80))
	total := metrics.Overview.TotalIdeas
	for status, count := range metrics.StatusBreakdown {
		pct := float64(count) / float64(total) * 100
		fmt.Printf("  %-10s: %5d (%.1f%%)\n", status, count, pct)
	}
	fmt.Println()

	// Score Distribution
	fmt.Println("Score Distribution:")
	fmt.Println(strings.Repeat("-", 80))
	bucketOrder := []string{"0-2", "2-4", "4-6", "6-8", "8-10"}
	for _, bucket := range bucketOrder {
		count := metrics.ScoreDistribution.Buckets[bucket]
		pct := float64(count) / float64(total) * 100
		bar := strings.Repeat("█", int(pct/2))
		fmt.Printf("  %5s: %5d (%.1f%%) %s\n", bucket, count, pct, bar)
	}
	fmt.Printf("  StdDev: %.2f\n", metrics.ScoreDistribution.StdDev)
	if opts.verbose {
		fmt.Println("\n  Percentiles:")
		fmt.Printf("    P50: %.2f\n", metrics.ScoreDistribution.Percentiles["P50"])
		fmt.Printf("    P75: %.2f\n", metrics.ScoreDistribution.Percentiles["P75"])
		fmt.Printf("    P90: %.2f\n", metrics.ScoreDistribution.Percentiles["P90"])
		fmt.Printf("    P95: %.2f\n", metrics.ScoreDistribution.Percentiles["P95"])
		fmt.Printf("    P99: %.2f\n", metrics.ScoreDistribution.Percentiles["P99"])
	}
	fmt.Println()

	// Top Patterns
	if len(metrics.PatternStats) > 0 {
		fmt.Println("Top Patterns:")
		fmt.Println(strings.Repeat("-", 80))
		topN := 10
		if opts.verbose {
			topN = 20
		}
		for i, ps := range metrics.PatternStats {
			if i >= topN {
				break
			}
			fmt.Printf("  %2d. %-30s: %4d ideas (%.1f%%)\n",
				i+1, ps.Pattern, ps.Count, ps.Percentage)
		}
		if len(metrics.PatternStats) > topN {
			fmt.Printf("  ... and %d more patterns\n", len(metrics.PatternStats)-topN)
		}
		fmt.Println()
	}

	// Time Metrics
	fmt.Println("Activity Timeline:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("  Oldest Idea:      %s\n", metrics.TimeMetrics.OldestIdea.Format("2006-01-02"))
	fmt.Printf("  Newest Idea:      %s\n", metrics.TimeMetrics.NewestIdea.Format("2006-01-02"))
	fmt.Printf("  Total Days:       %d\n", metrics.TimeMetrics.TotalDays)
	fmt.Printf("  Ideas per Day:    %.2f\n", metrics.TimeMetrics.IdeasPerDay)
	fmt.Printf("  Last 7 Days:      %d ideas\n", metrics.TimeMetrics.IdeasLast7Days)
	fmt.Printf("  Last 30 Days:     %d ideas\n", metrics.TimeMetrics.IdeasLast30Days)
	fmt.Println()

	// Database Stats
	if opts.verbose && metrics.DatabaseStats.SizeFormatted != "Unknown" {
		fmt.Println("Database:")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("  Size:             %s\n", metrics.DatabaseStats.SizeFormatted)
		fmt.Printf("  Tables:           %d\n", metrics.DatabaseStats.TableCount)
		fmt.Printf("  Indexes:          %d\n", metrics.DatabaseStats.IndexCount)
		fmt.Println()
	}

	return nil
}

func outputMetricsJSON(metrics analytics.SystemMetrics) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metrics)
}

func outputMetricsCSV(metrics analytics.SystemMetrics) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Overview section
	if err := writer.Write([]string{"Metric", "Value"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	if err := writer.Write([]string{"Total Ideas", strconv.Itoa(metrics.Overview.TotalIdeas)}); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}
	if err := writer.Write([]string{"Total Patterns", strconv.Itoa(metrics.Overview.TotalPatterns)}); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}
	if err := writer.Write([]string{"Average Score", fmt.Sprintf("%.2f", metrics.Overview.AverageScore)}); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}
	if err := writer.Write([]string{"Median Score", fmt.Sprintf("%.2f", metrics.Overview.MedianScore)}); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}
	if err := writer.Write([]string{}); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}

	// Status breakdown
	if err := writer.Write([]string{"Status", "Count", "Percentage"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	for status, count := range metrics.StatusBreakdown {
		pct := float64(count) / float64(metrics.Overview.TotalIdeas) * 100
		if err := writer.Write([]string{status, strconv.Itoa(count), fmt.Sprintf("%.1f", pct)}); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
