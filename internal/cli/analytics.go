// Package cli provides command-line interface commands for the Telos Idea Matrix application.
package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/analytics"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newAnalyticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "View statistics about your ideas",
		Long: `Display statistics and insights about your captured ideas.

Examples:
  tm analytics              # Show basic statistics
  tm analytics trends       # Show score trends over time
  tm analytics report       # Generate comprehensive report
  tm analytics patterns     # Show pattern frequency`,
		RunE: runAnalytics,
	}

	// Add subcommands
	cmd.AddCommand(newAnalyticsTrendsCommand())
	cmd.AddCommand(newAnalyticsReportCommand())
	cmd.AddCommand(newAnalyticsPatternsCommand())
	cmd.AddCommand(newAnalyticsPerformanceCommand())
	cmd.AddCommand(newAnalyticsAnomalyCommand())
	cmd.AddCommand(newAnalyticsMetricsCommand())
	cmd.AddCommand(newAnalyticsLLMCommand())

	return cmd
}

func runAnalytics(cmd *cobra.Command, args []string) error {
	// Fetch all active ideas
	ideas, err := ctx.Repository.List(database.ListOptions{
		Status: "active",
	})
	if err != nil {
		return fmt.Errorf("failed to list ideas: %w", err)
	}

	if len(ideas) == 0 {
		if _, err := warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!"); err != nil {
			log.Warn().Err(err).Msg("failed to print warning message")
		}
		return nil
	}

	// Calculate statistics
	var totalScore float64
	var highScore, lowScore float64 = 0, 10
	highCount := 0  // >= 7.0
	mediumCount := 0 // 5.0-7.0
	lowCount := 0    // < 5.0

	for _, idea := range ideas {
		totalScore += idea.FinalScore

		if idea.FinalScore > highScore {
			highScore = idea.FinalScore
		}
		if idea.FinalScore < lowScore {
			lowScore = idea.FinalScore
		}

		switch {
		case idea.FinalScore >= 7.0:
			highCount++
		case idea.FinalScore >= 5.0:
			mediumCount++
		default:
			lowCount++
		}
	}

	avgScore := totalScore / float64(len(ideas))

	// Display statistics
	fmt.Println("📊 Idea Analytics")
	fmt.Println("═════════════════════════════════════════════")
	fmt.Println()

	if _, err := successColor.Printf("Total Ideas: %d\n", len(ideas)); err != nil {
		log.Warn().Err(err).Msg("failed to print total ideas")
	}
	fmt.Printf("Average Score: %.1f/10.0\n", avgScore)
	fmt.Printf("Highest Score: %.1f/10.0\n", highScore)
	fmt.Printf("Lowest Score:  %.1f/10.0\n\n", lowScore)

	fmt.Println("Score Distribution:")

	// Visual distribution bar
	distBar := analytics.RenderDistribution(highCount, mediumCount, lowCount, 50)
	fmt.Printf("%s\n\n", distBar)

	if _, err := successColor.Printf("  🔥 High (>= 7.0):   %d ideas (%.0f%%)\n",
		highCount, float64(highCount)/float64(len(ideas))*100); err != nil {
		log.Warn().Err(err).Msg("failed to print high count")
	}
	if _, err := warningColor.Printf("  ⚠️  Medium (5-7):   %d ideas (%.0f%%)\n",
		mediumCount, float64(mediumCount)/float64(len(ideas))*100); err != nil {
		log.Warn().Err(err).Msg("failed to print medium count")
	}
	if _, err := errorColor.Printf("  🚫 Low (< 5.0):     %d ideas (%.0f%%)\n",
		lowCount, float64(lowCount)/float64(len(ideas))*100); err != nil {
		log.Warn().Err(err).Msg("failed to print low count")
	}
	fmt.Println()

	// Recommendations
	if highCount > 0 {
		if _, err := successColor.Printf("✨ You have %d high-scoring ideas to prioritize!\n", highCount); err != nil {
			log.Warn().Err(err).Msg("failed to print recommendation")
		}
	}
	if lowCount > len(ideas)/2 {
		if _, err := warningColor.Println("💡 Tip: Many ideas are low-scoring. Consider aligning more with your telos."); err != nil {
			log.Warn().Err(err).Msg("failed to print tip")
		}
	}

	fmt.Println("═════════════════════════════════════════════")

	return nil
}

// newAnalyticsTrendsCommand creates the analytics trends subcommand
func newAnalyticsTrendsCommand() *cobra.Command {
	var days int
	var groupBy string

	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Show score trends over time",
		Long: `Display score trends grouped by time period.

The trends command shows how your idea scores have evolved over time,
helping you identify patterns in your ideation process.

Examples:
  tm analytics trends                    # Weekly trends for last 30 days
  tm analytics trends --days 90          # Weekly trends for last 90 days
  tm analytics trends --group-by month   # Monthly trends
  tm analytics trends --group-by day     # Daily trends`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Fetch all active ideas
			ideas, err := ctx.Repository.List(database.ListOptions{
				Status: "active",
			})
			if err != nil {
				return fmt.Errorf("failed to list ideas: %w", err)
			}

			if len(ideas) == 0 {
				if _, err := warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!"); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
				return nil
			}

			// Calculate trends
			trends := analytics.CalculateScoreTrends(ideas, groupBy)

			// Display header
			fmt.Println("📈 Score Trends")
			fmt.Println("═════════════════════════════════════════════")
			fmt.Println()

			if len(trends) == 0 {
				fmt.Println("No trend data available yet.")
				return nil
			}

			// Display trends
			fmt.Printf("Grouping: %s\n\n", groupBy)

			// Generate sparkline
			values := make([]float64, len(trends))
			for i, trend := range trends {
				values[i] = trend.AvgScore
			}
			sparkline := analytics.RenderSparkline(values)

			fmt.Printf("Trend: %s\n\n", sparkline)

			for _, trend := range trends {
				// Color code based on score
				var scoreColor func(string, ...interface{}) string
				if trend.AvgScore >= 7.0 {
					scoreColor = successColor.Sprintf
				} else if trend.AvgScore >= 5.0 {
					scoreColor = warningColor.Sprintf
				} else {
					scoreColor = errorColor.Sprintf
				}

				fmt.Printf("  %s: %s (% d ideas)\n",
					trend.Period,
					scoreColor("%.1f avg", trend.AvgScore),
					trend.IdeaCount,
				)
			}

			// Show trend direction
			direction := analytics.CalculateTrendDirection(trends)
			fmt.Println()
			switch direction {
			case "up":
				if _, err := successColor.Println("📈 Trend: Your idea quality is improving over time!"); err != nil {
					log.Warn().Err(err).Msg("failed to print trend message")
				}
			case "down":
				if _, err := warningColor.Println("📉 Trend: Consider refining your idea capture process."); err != nil {
					log.Warn().Err(err).Msg("failed to print trend message")
				}
			default:
				fmt.Println("➡️  Trend: Your idea quality is stable.")
			}

			fmt.Println("═════════════════════════════════════════════")

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Number of days to analyze")
	cmd.Flags().StringVar(&groupBy, "group-by", "week", "Group by: day, week, or month")

	return cmd
}

// newAnalyticsReportCommand creates the analytics report subcommand
func newAnalyticsReportCommand() *cobra.Command {
	var outputFile string
	var format string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate comprehensive analytics report",
		Long: `Generate a detailed analytics report with insights and recommendations.

The report includes:
- Score distribution analysis
- Trend analysis over time
- Pattern frequency analysis
- Idea creation velocity metrics
- Personalized recommendations

Examples:
  tm analytics report                     # Display report in terminal
  tm analytics report --output report.md  # Save as markdown
  tm analytics report --format plain      # Plain text format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Fetch all active ideas
			ideas, err := ctx.Repository.List(database.ListOptions{
				Status: "active",
			})
			if err != nil {
				return fmt.Errorf("failed to list ideas: %w", err)
			}

			// Generate report
			report := analytics.GenerateReport(ideas)

			// Render report based on format
			var output string
			switch format {
			case "plain":
				output = analytics.RenderReportPlainText(report)
			case "markdown", "md":
				output = analytics.RenderReport(report)
			default:
				// Default to plain text for terminal display
				output = analytics.RenderReportPlainText(report)
			}

			// Output to file or stdout
			if outputFile != "" {
				err := os.WriteFile(outputFile, []byte(output), 0644)
				if err != nil {
					return fmt.Errorf("failed to write report: %w", err)
				}
				if _, err := successColor.Printf("✅ Report saved to: %s\n", outputFile); err != nil {
					log.Warn().Err(err).Msg("failed to print success message")
				}
			} else {
				fmt.Println(output)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path (default: stdout)")
	cmd.Flags().StringVar(&format, "format", "plain", "Output format: plain or markdown")

	return cmd
}

// newAnalyticsPatternsCommand creates the analytics patterns subcommand
func newAnalyticsPatternsCommand() *cobra.Command {
	var topN int

	cmd := &cobra.Command{
		Use:   "patterns",
		Short: "Show pattern frequency analysis",
		Long: `Display the most frequently occurring patterns in your ideas.

This helps identify recurring anti-patterns or common themes
across your idea collection.

Examples:
  tm analytics patterns           # Show top 10 patterns
  tm analytics patterns --top 5   # Show top 5 patterns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Fetch all active ideas
			ideas, err := ctx.Repository.List(database.ListOptions{
				Status: "active",
			})
			if err != nil {
				return fmt.Errorf("failed to list ideas: %w", err)
			}

			if len(ideas) == 0 {
				if _, err := warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!"); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
				return nil
			}

			// Calculate pattern frequency
			freq := analytics.CalculatePatternFrequency(ideas)

			if len(freq) == 0 {
				fmt.Println("No patterns detected in your ideas yet.")
				return nil
			}

			// Get top patterns
			topPatterns := analytics.GetTopPatterns(ideas, topN)

			// Display header
			fmt.Println("🔍 Pattern Frequency Analysis")
			fmt.Println("═════════════════════════════════════════════")
			fmt.Println()

			// Prepare data for bar chart
			labels := make([]string, len(topPatterns))
			values := make([]float64, len(topPatterns))
			for i, pattern := range topPatterns {
				labels[i] = pattern
				values[i] = float64(freq[pattern])
			}

			// Display bar chart
			chart := analytics.RenderBarChart(labels, values, 40)
			fmt.Println(chart)

			// Display patterns with percentages
			for i, pattern := range topPatterns {
				count := freq[pattern]
				percentage := (count * 100) / len(ideas)

				// Highlight high-frequency patterns
				if percentage > 40 {
					if _, err := warningColor.Printf("%d. %s: %d occurrences (%d%% of ideas) ⚠️\n",
						i+1, pattern, count, percentage); err != nil {
						log.Warn().Err(err).Msg("failed to print pattern")
					}
				} else {
					fmt.Printf("%d. %s: %d occurrences (%d%% of ideas)\n",
						i+1, pattern, count, percentage)
				}
			}

			fmt.Println()
			fmt.Printf("Total unique patterns detected: %d\n", len(freq))

			// Show warning if high frequency patterns exist
			hasHighFreq := false
			for _, pattern := range topPatterns {
				percentage := (freq[pattern] * 100) / len(ideas)
				if percentage > 40 {
					hasHighFreq = true
					break
				}
			}

			if hasHighFreq {
				fmt.Println()
				if _, err := warningColor.Println("⚠️  Warning: Some patterns appear very frequently."); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
				fmt.Println("   Consider addressing these recurring anti-patterns in your ideation process.")
			}

			fmt.Println("═════════════════════════════════════════════")

			return nil
		},
	}

	cmd.Flags().IntVar(&topN, "top", 10, "Number of top patterns to display")

	return cmd
}

// --- Metrics Command ---

type metricsOptions struct {
	format  string
	verbose bool
}

type systemMetrics struct {
	Overview          overviewMetrics          `json:"overview"`
	StatusBreakdown   map[string]int           `json:"status_breakdown"`
	ScoreDistribution scoreDistributionMetrics `json:"score_distribution"`
	PatternStats      []patternStat            `json:"pattern_stats"`
	TimeMetrics       timeMetrics              `json:"time_metrics"`
	DatabaseStats     databaseStats            `json:"database_stats"`
}

type overviewMetrics struct {
	TotalIdeas   int     `json:"total_ideas"`
	TotalPatterns int     `json:"total_patterns"`
	AverageScore  float64 `json:"average_score"`
	MedianScore   float64 `json:"median_score"`
	HighestScore  float64 `json:"highest_score"`
	LowestScore   float64 `json:"lowest_score"`
}

type scoreDistributionMetrics struct {
	Buckets     map[string]int     `json:"buckets"`
	StdDev      float64            `json:"std_dev"`
	Percentiles map[string]float64 `json:"percentiles"`
}

type patternStat struct {
	Pattern    string  `json:"pattern"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type timeMetrics struct {
	OldestIdea      time.Time `json:"oldest_idea"`
	NewestIdea      time.Time `json:"newest_idea"`
	TotalDays       int       `json:"total_days"`
	IdeasPerDay     float64   `json:"ideas_per_day"`
	IdeasLast7Days  int       `json:"ideas_last_7_days"`
	IdeasLast30Days int       `json:"ideas_last_30_days"`
}

type databaseStats struct {
	SizeBytes     int64  `json:"size_bytes"`
	SizeFormatted string `json:"size_formatted"`
	TableCount    int    `json:"table_count"`
	IndexCount    int    `json:"index_count"`
}

// newAnalyticsMetricsCommand creates the analytics metrics subcommand
func newAnalyticsMetricsCommand() *cobra.Command {
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
			return runSystemMetrics(metricsOptions{
				format:  format,
				verbose: verbose,
			})
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format: text|json|csv")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed metrics")

	return cmd
}

func runSystemMetrics(opts metricsOptions) error {
	// Fetch all ideas (not just active)
	ideas, err := ctx.Repository.List(database.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch ideas: %w", err)
	}

	if len(ideas) == 0 {
		fmt.Println("No ideas found in the system.")
		return nil
	}

	// Calculate metrics
	metrics := calculateSystemMetrics(ideas)

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

func calculateSystemMetrics(ideas []*models.Idea) systemMetrics {
	metrics := systemMetrics{
		StatusBreakdown: make(map[string]int),
	}

	// Overview metrics
	metrics.Overview = calculateOverviewMetrics(ideas)

	// Status breakdown
	for _, idea := range ideas {
		metrics.StatusBreakdown[idea.Status]++
	}

	// Score distribution
	metrics.ScoreDistribution = calculateScoreDistribution(ideas)

	// Pattern statistics
	metrics.PatternStats = calculatePatternStats(ideas)

	// Time metrics
	metrics.TimeMetrics = calculateTimeMetrics(ideas)

	// Database statistics
	metrics.DatabaseStats = calculateDatabaseStats()

	return metrics
}

func calculateOverviewMetrics(ideas []*models.Idea) overviewMetrics {
	if len(ideas) == 0 {
		return overviewMetrics{}
	}

	scores := make([]float64, len(ideas))
	patternSet := make(map[string]bool)

	sum := 0.0
	highest := ideas[0].FinalScore
	lowest := ideas[0].FinalScore

	for i, idea := range ideas {
		scores[i] = idea.FinalScore
		sum += idea.FinalScore

		if idea.FinalScore > highest {
			highest = idea.FinalScore
		}
		if idea.FinalScore < lowest {
			lowest = idea.FinalScore
		}

		for _, pattern := range idea.Patterns {
			patternSet[pattern] = true
		}
	}

	return overviewMetrics{
		TotalIdeas:    len(ideas),
		TotalPatterns: len(patternSet),
		AverageScore:  sum / float64(len(ideas)),
		MedianScore:   calculateMedian(scores),
		HighestScore:  highest,
		LowestScore:   lowest,
	}
}

func calculateScoreDistribution(ideas []*models.Idea) scoreDistributionMetrics {
	buckets := map[string]int{
		"0-2":  0,
		"2-4":  0,
		"4-6":  0,
		"6-8":  0,
		"8-10": 0,
	}

	scores := make([]float64, len(ideas))

	for i, idea := range ideas {
		scores[i] = idea.FinalScore

		switch {
		case idea.FinalScore < 2:
			buckets["0-2"]++
		case idea.FinalScore < 4:
			buckets["2-4"]++
		case idea.FinalScore < 6:
			buckets["4-6"]++
		case idea.FinalScore < 8:
			buckets["6-8"]++
		default:
			buckets["8-10"]++
		}
	}

	return scoreDistributionMetrics{
		Buckets: buckets,
		StdDev:  calculateStdDev(scores),
		Percentiles: map[string]float64{
			"P50": calculatePercentile(scores, 50),
			"P75": calculatePercentile(scores, 75),
			"P90": calculatePercentile(scores, 90),
			"P95": calculatePercentile(scores, 95),
			"P99": calculatePercentile(scores, 99),
		},
	}
}

func calculatePatternStats(ideas []*models.Idea) []patternStat {
	patternCounts := make(map[string]int)

	for _, idea := range ideas {
		for _, pattern := range idea.Patterns {
			patternCounts[pattern]++
		}
	}

	stats := make([]patternStat, 0, len(patternCounts))
	totalIdeas := float64(len(ideas))

	for pattern, count := range patternCounts {
		stats = append(stats, patternStat{
			Pattern:    pattern,
			Count:      count,
			Percentage: (float64(count) / totalIdeas) * 100,
		})
	}

	// Sort by count (descending)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

func calculateTimeMetrics(ideas []*models.Idea) timeMetrics {
	if len(ideas) == 0 {
		return timeMetrics{}
	}

	oldest := ideas[0].CreatedAt
	newest := ideas[0].CreatedAt

	for _, idea := range ideas {
		if idea.CreatedAt.Before(oldest) {
			oldest = idea.CreatedAt
		}
		if idea.CreatedAt.After(newest) {
			newest = idea.CreatedAt
		}
	}

	totalDays := int(newest.Sub(oldest).Hours() / 24)
	if totalDays == 0 {
		totalDays = 1
	}

	now := time.Now()
	last7Days := now.AddDate(0, 0, -7)
	last30Days := now.AddDate(0, 0, -30)

	count7Days := 0
	count30Days := 0

	for _, idea := range ideas {
		if idea.CreatedAt.After(last7Days) {
			count7Days++
		}
		if idea.CreatedAt.After(last30Days) {
			count30Days++
		}
	}

	return timeMetrics{
		OldestIdea:      oldest,
		NewestIdea:      newest,
		TotalDays:       totalDays,
		IdeasPerDay:     float64(len(ideas)) / float64(totalDays),
		IdeasLast7Days:  count7Days,
		IdeasLast30Days: count30Days,
	}
}

func calculateDatabaseStats() databaseStats {
	// Try to get database file size
	sizeBytes := int64(0)
	sizeFormatted := "Unknown"

	if ctx != nil && ctx.DBPath != "" {
		if info, err := os.Stat(ctx.DBPath); err == nil {
			sizeBytes = info.Size()
			sizeFormatted = formatBytes(sizeBytes)
		}
	}

	return databaseStats{
		SizeBytes:     sizeBytes,
		SizeFormatted: sizeFormatted,
		TableCount:    1, // ideas table
		IndexCount:    0,
	}
}

// Statistical helper functions

func calculateMedian(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}

	sorted := make([]float64, len(scores))
	copy(sorted, scores)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func calculateStdDev(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}

	mean := 0.0
	for _, score := range scores {
		mean += score
	}
	mean /= float64(len(scores))

	variance := 0.0
	for _, score := range scores {
		diff := score - mean
		variance += diff * diff
	}
	variance /= float64(len(scores))

	return math.Sqrt(variance)
}

func calculatePercentile(scores []float64, p int) float64 {
	if len(scores) == 0 {
		return 0
	}

	sorted := make([]float64, len(scores))
	copy(sorted, scores)
	sort.Float64s(sorted)

	if p == 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	// Use nearest-rank method
	rank := float64(p) / 100.0 * float64(len(sorted)-1)
	index := int(rank)

	// Linear interpolation between values
	if index+1 < len(sorted) {
		fraction := rank - float64(index)
		return sorted[index]*(1-fraction) + sorted[index+1]*fraction
	}

	return sorted[index]
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// Output formatters

func outputMetricsText(metrics systemMetrics, opts metricsOptions) error {
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

func outputMetricsJSON(metrics systemMetrics) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metrics)
}

func outputMetricsCSV(metrics systemMetrics) error {
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
func newAnalyticsPerformanceCommand() *cobra.Command {
	var (
		days    int
		groupBy string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "performance",
		Short: "Show scoring performance over time",
		Long: `Analyze how idea scores have trended over time.

Displays:
- Average score trends
- Score distribution changes
- High/low performing periods
- Improvement indicators
- Statistical insights

Examples:
  # Show performance for last 30 days grouped by day
  tm analytics performance

  # Show last 90 days grouped by week
  tm analytics performance --days 90 --group-by week

  # Show last year grouped by month
  tm analytics performance --days 365 --group-by month

  # Export as JSON
  tm analytics performance --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPerformanceAnalytics(performanceOptions{
				days:    days,
				groupBy: groupBy,
				format:  format,
			})
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Number of days to analyze")
	cmd.Flags().StringVar(&groupBy, "group-by", "day", "Group by: day|week|month")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text|json|csv")

	return cmd
}

type performanceOptions struct {
	days    int
	groupBy string
	format  string
}

type performanceReport struct {
	Period       string  `json:"period"`
	TotalIdeas   int     `json:"total_ideas"`
	AverageScore float64 `json:"average_score"`
	HighestScore float64 `json:"highest_score"`
	LowestScore  float64 `json:"lowest_score"`
	MedianScore  float64 `json:"median_score"`
	StdDev       float64 `json:"std_dev"`
}

type timeGroup struct {
	Label string
	Start time.Time
	End   time.Time
	Ideas []*models.Idea
}

func runPerformanceAnalytics(opts performanceOptions) error {
	// Validate groupBy
	validGroups := []string{"day", "week", "month"}
	if !analyticsContains(validGroups, opts.groupBy) {
		return fmt.Errorf("invalid group-by: %s (must be one of: %s)",
			opts.groupBy, strings.Join(validGroups, ", "))
	}

	// Get ideas from time period
	cutoff := time.Now().AddDate(0, 0, -opts.days)

	// Fetch all active ideas
	allIdeas, err := ctx.Repository.List(database.ListOptions{
		Status: "active",
	})
	if err != nil {
		return fmt.Errorf("failed to fetch ideas: %w", err)
	}

	// Filter ideas by created date
	var ideas []*models.Idea
	for _, idea := range allIdeas {
		if idea.CreatedAt.After(cutoff) {
			ideas = append(ideas, idea)
		}
	}

	if len(ideas) == 0 {
		fmt.Printf("No ideas found in the last %d days.\n", opts.days)
		return nil
	}

	// Group by time period
	groups := groupIdeasByTime(ideas, opts.groupBy)

	// Calculate statistics for each group
	reports := make([]performanceReport, 0, len(groups))
	for _, group := range groups {
		report := performanceReport{
			Period:       group.Label,
			TotalIdeas:   len(group.Ideas),
			AverageScore: calculateAverageScore(group.Ideas),
			HighestScore: findHighestScore(group.Ideas),
			LowestScore:  findLowestScore(group.Ideas),
			MedianScore:  calculateMedianScore(group.Ideas),
			StdDev:       calculateStdDevFromIdeas(group.Ideas),
		}
		reports = append(reports, report)
	}

	// Output based on format
	switch opts.format {
	case "json":
		return outputJSON(reports)
	case "csv":
		return outputCSV(reports)
	default:
		return outputText(reports, opts)
	}
}

func groupIdeasByTime(ideas []*models.Idea, groupBy string) []timeGroup {
	groups := make(map[string][]*models.Idea)

	for _, idea := range ideas {
		var key string
		switch groupBy {
		case "day":
			key = idea.CreatedAt.Format("2006-01-02")
		case "week":
			year, week := idea.CreatedAt.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case "month":
			key = idea.CreatedAt.Format("2006-01")
		}

		groups[key] = append(groups[key], idea)
	}

	// Convert to sorted slice
	result := make([]timeGroup, 0, len(groups))
	for key, ideas := range groups {
		result = append(result, timeGroup{
			Label: key,
			Ideas: ideas,
		})
	}

	// Sort by label (time)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Label < result[j].Label
	})

	return result
}

func calculateAverageScore(ideas []*models.Idea) float64 {
	if len(ideas) == 0 {
		return 0
	}

	sum := 0.0
	for _, idea := range ideas {
		sum += idea.FinalScore
	}

	return sum / float64(len(ideas))
}

func findHighestScore(ideas []*models.Idea) float64 {
	if len(ideas) == 0 {
		return 0
	}

	highest := ideas[0].FinalScore
	for _, idea := range ideas[1:] {
		if idea.FinalScore > highest {
			highest = idea.FinalScore
		}
	}

	return highest
}

func findLowestScore(ideas []*models.Idea) float64 {
	if len(ideas) == 0 {
		return 0
	}

	lowest := ideas[0].FinalScore
	for _, idea := range ideas[1:] {
		if idea.FinalScore < lowest {
			lowest = idea.FinalScore
		}
	}

	return lowest
}

func calculateMedianScore(ideas []*models.Idea) float64 {
	if len(ideas) == 0 {
		return 0
	}

	scores := make([]float64, len(ideas))
	for i, idea := range ideas {
		scores[i] = idea.FinalScore
	}

	sort.Float64s(scores)

	mid := len(scores) / 2
	if len(scores)%2 == 0 {
		return (scores[mid-1] + scores[mid]) / 2
	}
	return scores[mid]
}

func calculateStdDevFromIdeas(ideas []*models.Idea) float64 {
	if len(ideas) == 0 {
		return 0
	}

	mean := calculateAverageScore(ideas)
	variance := 0.0

	for _, idea := range ideas {
		diff := idea.FinalScore - mean
		variance += diff * diff
	}

	variance /= float64(len(ideas))
	return math.Sqrt(variance)
}

func outputText(reports []performanceReport, opts performanceOptions) error {
	fmt.Printf("Performance Analysis (Last %d days)\n", opts.days)
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Table header
	fmt.Printf("%-12s %5s %8s %8s %8s %8s %8s\n",
		"Period", "Count", "Average", "Median", "Highest", "Lowest", "StdDev")
	fmt.Println(strings.Repeat("-", 70))

	// Table rows
	for _, report := range reports {
		fmt.Printf("%-12s %5d %8.2f %8.2f %8.2f %8.2f %8.2f\n",
			report.Period,
			report.TotalIdeas,
			report.AverageScore,
			report.MedianScore,
			report.HighestScore,
			report.LowestScore,
			report.StdDev,
		)
	}

	fmt.Println()

	// Show trend indicators
	showTrendIndicators(reports)

	// Show summary statistics
	showSummaryStatistics(reports)

	return nil
}

func outputJSON(reports []performanceReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(reports)
}

func outputCSV(reports []performanceReport) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Header
	if err := writer.Write([]string{
		"Period", "TotalIdeas", "AverageScore", "MedianScore",
		"HighestScore", "LowestScore", "StdDev",
	}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Data rows
	for _, report := range reports {
		if err := writer.Write([]string{
			report.Period,
			strconv.Itoa(report.TotalIdeas),
			fmt.Sprintf("%.2f", report.AverageScore),
			fmt.Sprintf("%.2f", report.MedianScore),
			fmt.Sprintf("%.2f", report.HighestScore),
			fmt.Sprintf("%.2f", report.LowestScore),
			fmt.Sprintf("%.2f", report.StdDev),
		}); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

func showTrendIndicators(reports []performanceReport) {
	if len(reports) < 2 {
		return
	}

	fmt.Println("Trend Analysis:")
	fmt.Println(strings.Repeat("-", 70))

	// Compare first and last periods
	first := reports[0]
	last := reports[len(reports)-1]

	// Score trend
	avgChange := last.AverageScore - first.AverageScore
	avgChangePercent := 0.0
	if first.AverageScore != 0 {
		avgChangePercent = (avgChange / first.AverageScore) * 100
	}

	fmt.Printf("\nAverage Score:\n")
	fmt.Printf("  First period: %.2f\n", first.AverageScore)
	fmt.Printf("  Last period:  %.2f\n", last.AverageScore)
	if avgChange > 0.5 {
		fmt.Printf("  Trend: ↑ Improving (+%.2f, +%.1f%%)\n", avgChange, avgChangePercent)
	} else if avgChange < -0.5 {
		fmt.Printf("  Trend: ↓ Declining (%.2f, %.1f%%)\n", avgChange, avgChangePercent)
	} else {
		fmt.Printf("  Trend: → Stable (%.2f)\n", avgChange)
	}

	// Volume trend
	volumeChange := last.TotalIdeas - first.TotalIdeas
	fmt.Printf("\nIdea Volume:\n")
	fmt.Printf("  First period: %d ideas\n", first.TotalIdeas)
	fmt.Printf("  Last period:  %d ideas\n", last.TotalIdeas)
	if volumeChange > 0 {
		fmt.Printf("  Trend: ↑ Increasing (+%d ideas)\n", volumeChange)
	} else if volumeChange < 0 {
		fmt.Printf("  Trend: ↓ Decreasing (%d ideas)\n", volumeChange)
	} else {
		fmt.Printf("  Trend: → Stable\n")
	}

	// Consistency trend (standard deviation)
	fmt.Printf("\nConsistency (Lower StdDev = More Consistent):\n")
	fmt.Printf("  First period: %.2f\n", first.StdDev)
	fmt.Printf("  Last period:  %.2f\n", last.StdDev)
	if last.StdDev < first.StdDev {
		fmt.Printf("  Trend: ↑ More consistent\n")
	} else if last.StdDev > first.StdDev {
		fmt.Printf("  Trend: ↓ Less consistent\n")
	} else {
		fmt.Printf("  Trend: → Stable\n")
	}

	fmt.Println()
}

func showSummaryStatistics(reports []performanceReport) {
	fmt.Println("Summary Statistics:")
	fmt.Println(strings.Repeat("-", 70))

	totalIdeas := 0
	for _, report := range reports {
		totalIdeas += report.TotalIdeas
	}

	// Calculate overall statistics
	fmt.Printf("  Total ideas analyzed: %d\n", totalIdeas)
	fmt.Printf("  Time periods: %d\n", len(reports))
	fmt.Printf("  Average ideas per period: %.1f\n", float64(totalIdeas)/float64(len(reports)))

	// Find best and worst periods
	bestPeriod := reports[0]
	worstPeriod := reports[0]
	for _, report := range reports[1:] {
		if report.AverageScore > bestPeriod.AverageScore {
			bestPeriod = report
		}
		if report.AverageScore < worstPeriod.AverageScore {
			worstPeriod = report
		}
	}

	fmt.Printf("\n  Best period: %s (avg: %.2f)\n", bestPeriod.Period, bestPeriod.AverageScore)
	fmt.Printf("  Worst period: %s (avg: %.2f)\n", worstPeriod.Period, worstPeriod.AverageScore)
	fmt.Println()
}

func analyticsContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ==================== Anomaly Detection ====================

// Anomaly detection types and data structures
type anomalyOptions struct {
	threshold     float64
	minOccurrence float64
	verbose       bool
	format        string
}

type anomalyReport struct {
	ScoreOutliers        []scoreOutlier        `json:"score_outliers"`
	RarePatterns         []rarePattern         `json:"rare_patterns"`
	TimingAnomalies      []timingAnomaly       `json:"timing_anomalies"`
	RecommendationIssues []recommendationIssue `json:"recommendation_issues"`
}

type scoreOutlier struct {
	IdeaID     string  `json:"idea_id"`
	Score      float64 `json:"score"`
	Deviation  float64 `json:"deviation"`
	Content    string  `json:"content"`
	IsPositive bool    `json:"is_positive"`
}

type rarePattern struct {
	Pattern    string   `json:"pattern"`
	Count      int      `json:"count"`
	Percentage float64  `json:"percentage"`
	IdeaIDs    []string `json:"idea_ids,omitempty"`
}

type timingAnomaly struct {
	Date      string  `json:"date"`
	Count     int     `json:"count"`
	Expected  float64 `json:"expected"`
	Deviation float64 `json:"deviation"`
}

type recommendationIssue struct {
	IdeaID         string  `json:"idea_id"`
	Score          float64 `json:"score"`
	Recommendation string  `json:"recommendation"`
	Issue          string  `json:"issue"`
	Content        string  `json:"content"`
}

// newAnalyticsAnomalyCommand creates the analytics anomaly subcommand
func newAnalyticsAnomalyCommand() *cobra.Command {
	var (
		threshold     float64
		minOccurrence float64
		verbose       bool
		format        string
	)

	cmd := &cobra.Command{
		Use:   "anomaly",
		Short: "Detect unusual patterns and outliers",
		Long: `Identify anomalies in your idea data:

Detection Types:
- Score outliers (unusually high/low scores)
- Rare patterns (uncommon pattern combinations)
- Timing anomalies (unusual creation patterns)
- Recommendation conflicts (score vs recommendation mismatch)

Statistical Methods:
- Standard deviation (σ) for score outliers
- Occurrence frequency for rare patterns
- Time-series analysis for timing

Examples:
  tm analytics anomaly                       # Detect all anomalies with default threshold (2σ)
  tm analytics anomaly --threshold 3.0       # Strict anomaly detection (3σ threshold)
  tm analytics anomaly --min-occurrence 2.0  # Find very rare patterns (<2% occurrence)
  tm analytics anomaly --verbose             # Verbose output with details
  tm analytics anomaly --format json         # Export as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnomalyDetection(anomalyOptions{
				threshold:     threshold,
				minOccurrence: minOccurrence,
				verbose:       verbose,
				format:        format,
			})
		},
	}

	cmd.Flags().Float64Var(&threshold, "threshold", 2.0, "Standard deviation threshold for outliers")
	cmd.Flags().Float64Var(&minOccurrence, "min-occurrence", 5.0, "Minimum occurrence percentage for rare patterns")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed anomaly information")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text|json")

	return cmd
}

func runAnomalyDetection(opts anomalyOptions) error {
	// Get all active ideas
	ideas, err := ctx.Repository.List(database.ListOptions{
		Status: "active",
	})
	if err != nil {
		return fmt.Errorf("failed to fetch ideas: %w", err)
	}

	if len(ideas) == 0 {
		fmt.Println("No ideas found.")
		return nil
	}

	report := anomalyReport{}

	// 1. Detect score outliers
	report.ScoreOutliers = detectScoreOutliers(ideas, opts.threshold)

	// 2. Detect rare patterns
	report.RarePatterns = detectRarePatterns(ideas, opts.minOccurrence, opts.verbose)

	// 3. Detect timing anomalies
	report.TimingAnomalies = detectTimingAnomalies(ideas, opts.threshold)

	// 4. Detect recommendation issues
	report.RecommendationIssues = detectRecommendationIssues(ideas)

	// Output based on format
	switch opts.format {
	case "json":
		return outputAnomalyJSON(report)
	default:
		return outputAnomalyText(report, opts)
	}
}

// detectScoreOutliers finds ideas with scores that deviate significantly from the mean
func detectScoreOutliers(ideas []*models.Idea, threshold float64) []scoreOutlier {
	if len(ideas) < 3 {
		return []scoreOutlier{}
	}

	// Calculate mean and standard deviation
	mean, stdDev := calculateScoreStats(ideas)

	if stdDev == 0 {
		return []scoreOutlier{} // All scores are the same
	}

	outliers := make([]scoreOutlier, 0)

	for _, idea := range ideas {
		deviation := (idea.FinalScore - mean) / stdDev

		if math.Abs(deviation) > threshold {
			outliers = append(outliers, scoreOutlier{
				IdeaID:     idea.ID,
				Score:      idea.FinalScore,
				Deviation:  deviation,
				Content:    truncate(idea.Content, 80),
				IsPositive: deviation > 0,
			})
		}
	}

	// Sort by deviation (highest first)
	sort.Slice(outliers, func(i, j int) bool {
		return math.Abs(outliers[i].Deviation) > math.Abs(outliers[j].Deviation)
	})

	return outliers
}

func calculateScoreStats(ideas []*models.Idea) (mean float64, stdDev float64) {
	if len(ideas) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, idea := range ideas {
		sum += idea.FinalScore
	}
	mean = sum / float64(len(ideas))

	// Calculate standard deviation
	variance := 0.0
	for _, idea := range ideas {
		diff := idea.FinalScore - mean
		variance += diff * diff
	}
	variance /= float64(len(ideas))
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}

// detectRarePatterns finds patterns that occur infrequently
func detectRarePatterns(ideas []*models.Idea, minOccurrence float64, includeIDs bool) []rarePattern {
	// Count pattern occurrences
	patternCounts := make(map[string]int)
	patternIdeas := make(map[string][]string)

	for _, idea := range ideas {
		for _, pattern := range idea.Patterns {
			patternCounts[pattern]++
			if includeIDs {
				patternIdeas[pattern] = append(patternIdeas[pattern], idea.ID)
			}
		}
	}

	totalIdeas := float64(len(ideas))
	rarePatterns := make([]rarePattern, 0)

	for pattern, count := range patternCounts {
		percentage := (float64(count) / totalIdeas) * 100

		if percentage < minOccurrence {
			rp := rarePattern{
				Pattern:    pattern,
				Count:      count,
				Percentage: percentage,
			}
			if includeIDs {
				rp.IdeaIDs = patternIdeas[pattern]
			}
			rarePatterns = append(rarePatterns, rp)
		}
	}

	// Sort by rarity (least common first)
	sort.Slice(rarePatterns, func(i, j int) bool {
		return rarePatterns[i].Percentage < rarePatterns[j].Percentage
	})

	return rarePatterns
}

// detectTimingAnomalies finds unusual patterns in idea creation timing
func detectTimingAnomalies(ideas []*models.Idea, threshold float64) []timingAnomaly {
	if len(ideas) < 7 {
		return []timingAnomaly{} // Need at least a week of data
	}

	// Group ideas by date
	dateCounts := make(map[string]int)
	for _, idea := range ideas {
		date := idea.CreatedAt.Format("2006-01-02")
		dateCounts[date]++
	}

	// Calculate mean and stddev of daily counts
	counts := make([]float64, 0, len(dateCounts))
	for _, count := range dateCounts {
		counts = append(counts, float64(count))
	}

	mean := calculateMean(counts)
	stdDev := calculateStdDevFromValues(counts, mean)

	if stdDev == 0 {
		return []timingAnomaly{} // All days have same count
	}

	anomalies := make([]timingAnomaly, 0)

	for date, count := range dateCounts {
		deviation := (float64(count) - mean) / stdDev

		if math.Abs(deviation) > threshold {
			anomalies = append(anomalies, timingAnomaly{
				Date:      date,
				Count:     count,
				Expected:  mean,
				Deviation: deviation,
			})
		}
	}

	// Sort by deviation
	sort.Slice(anomalies, func(i, j int) bool {
		return math.Abs(anomalies[i].Deviation) > math.Abs(anomalies[j].Deviation)
	})

	return anomalies
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDevFromValues(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

// detectRecommendationIssues finds mismatches between scores and recommendations
func detectRecommendationIssues(ideas []*models.Idea) []recommendationIssue {
	issues := make([]recommendationIssue, 0)

	for _, idea := range ideas {
		var issue string

		// Check for score-recommendation mismatches
		switch idea.Recommendation {
		case "pursue":
			if idea.FinalScore < 7.0 {
				issue = fmt.Sprintf("Low score (%.1f) but recommended to pursue", idea.FinalScore)
			}
		case "reject":
			if idea.FinalScore > 5.0 {
				issue = fmt.Sprintf("High score (%.1f) but recommended to reject", idea.FinalScore)
			}
		case "defer":
			if idea.FinalScore > 8.0 {
				issue = fmt.Sprintf("Very high score (%.1f) but recommended to defer", idea.FinalScore)
			}
		}

		if issue != "" {
			issues = append(issues, recommendationIssue{
				IdeaID:         idea.ID,
				Score:          idea.FinalScore,
				Recommendation: idea.Recommendation,
				Issue:          issue,
				Content:        truncate(idea.Content, 60),
			})
		}
	}

	return issues
}

// outputAnomalyText renders the anomaly report in text format
func outputAnomalyText(report anomalyReport, opts anomalyOptions) error {
	fmt.Println("Anomaly Detection Report")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Score Outliers
	if len(report.ScoreOutliers) > 0 {
		fmt.Printf("Score Outliers (>%.1fσ):\n", opts.threshold)
		fmt.Println(strings.Repeat("-", 80))
		for i, outlier := range report.ScoreOutliers {
			symbol := "↑"
			if !outlier.IsPositive {
				symbol = "↓"
			}
			fmt.Printf("%d. [%s] Score: %.2f (%.1fσ %s)\n",
				i+1, outlier.IdeaID[:8], outlier.Score, math.Abs(outlier.Deviation), symbol)
			if opts.verbose {
				fmt.Printf("   %s\n", outlier.Content)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("No score outliers detected (threshold: %.1fσ)\n\n", opts.threshold)
	}

	// Rare Patterns
	if len(report.RarePatterns) > 0 {
		fmt.Printf("Rare Patterns (<%.1f%% occurrence):\n", opts.minOccurrence)
		fmt.Println(strings.Repeat("-", 80))
		for i, rp := range report.RarePatterns {
			fmt.Printf("%d. %s: %d ideas (%.1f%%)\n",
				i+1, rp.Pattern, rp.Count, rp.Percentage)
			if opts.verbose && len(rp.IdeaIDs) > 0 {
				// Show first few IDs
				idsToShow := rp.IdeaIDs
				if len(idsToShow) > 5 {
					idsToShow = idsToShow[:5]
				}
				shortIDs := make([]string, len(idsToShow))
				for j, id := range idsToShow {
					shortIDs[j] = id[:8]
				}
				fmt.Printf("   IDs: %s", strings.Join(shortIDs, ", "))
				if len(rp.IdeaIDs) > 5 {
					fmt.Printf(" ... and %d more", len(rp.IdeaIDs)-5)
				}
				fmt.Println()
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("No rare patterns detected (threshold: %.1f%%)\n\n", opts.minOccurrence)
	}

	// Timing Anomalies
	if len(report.TimingAnomalies) > 0 {
		fmt.Printf("Timing Anomalies (>%.1fσ from expected):\n", opts.threshold)
		fmt.Println(strings.Repeat("-", 80))
		for i, ta := range report.TimingAnomalies {
			symbol := "↑"
			if ta.Deviation < 0 {
				symbol = "↓"
			}
			fmt.Printf("%d. %s: %d ideas (expected: %.1f, %.1fσ %s)\n",
				i+1, ta.Date, ta.Count, ta.Expected, math.Abs(ta.Deviation), symbol)
		}
		fmt.Println()
	} else {
		fmt.Printf("No timing anomalies detected (threshold: %.1fσ)\n\n", opts.threshold)
	}

	// Recommendation Issues
	if len(report.RecommendationIssues) > 0 {
		fmt.Println("Recommendation Issues:")
		fmt.Println(strings.Repeat("-", 80))
		for i, ri := range report.RecommendationIssues {
			fmt.Printf("%d. [%s] %s\n", i+1, ri.IdeaID[:8], ri.Issue)
			if opts.verbose {
				fmt.Printf("   Content: %s\n", ri.Content)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("No recommendation issues detected")
		fmt.Println()
	}

	// Summary
	totalAnomalies := len(report.ScoreOutliers) + len(report.RarePatterns) +
		len(report.TimingAnomalies) + len(report.RecommendationIssues)

	fmt.Printf("Summary: %d total anomalies detected\n", totalAnomalies)

	return nil
}

// outputAnomalyJSON renders the anomaly report in JSON format
func outputAnomalyJSON(report anomalyReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
