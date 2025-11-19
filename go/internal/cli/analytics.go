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
	cmd.AddCommand(newAnalyticsMetricsCommand())

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
		warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!")
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

	successColor.Printf("Total Ideas: %d\n", len(ideas))
	fmt.Printf("Average Score: %.1f/10.0\n", avgScore)
	fmt.Printf("Highest Score: %.1f/10.0\n", highScore)
	fmt.Printf("Lowest Score:  %.1f/10.0\n\n", lowScore)

	fmt.Println("Score Distribution:")

	// Visual distribution bar
	distBar := analytics.RenderDistribution(highCount, mediumCount, lowCount, 50)
	fmt.Printf("%s\n\n", distBar)

	successColor.Printf("  🔥 High (>= 7.0):   %d ideas (%.0f%%)\n",
		highCount, float64(highCount)/float64(len(ideas))*100)
	warningColor.Printf("  ⚠️  Medium (5-7):   %d ideas (%.0f%%)\n",
		mediumCount, float64(mediumCount)/float64(len(ideas))*100)
	errorColor.Printf("  🚫 Low (< 5.0):     %d ideas (%.0f%%)\n",
		lowCount, float64(lowCount)/float64(len(ideas))*100)
	fmt.Println()

	// Recommendations
	if highCount > 0 {
		successColor.Printf("✨ You have %d high-scoring ideas to prioritize!\n", highCount)
	}
	if lowCount > len(ideas)/2 {
		warningColor.Println("💡 Tip: Many ideas are low-scoring. Consider aligning more with your telos.")
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
				warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!")
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
				successColor.Println("📈 Trend: Your idea quality is improving over time!")
			case "down":
				warningColor.Println("📉 Trend: Consider refining your idea capture process.")
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
				successColor.Printf("✅ Report saved to: %s\n", outputFile)
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
				warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!")
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
					warningColor.Printf("%d. %s: %d occurrences (%d%% of ideas) ⚠️\n",
						i+1, pattern, count, percentage)
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
				warningColor.Println("⚠️  Warning: Some patterns appear very frequently.")
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
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Ideas", strconv.Itoa(metrics.Overview.TotalIdeas)})
	writer.Write([]string{"Total Patterns", strconv.Itoa(metrics.Overview.TotalPatterns)})
	writer.Write([]string{"Average Score", fmt.Sprintf("%.2f", metrics.Overview.AverageScore)})
	writer.Write([]string{"Median Score", fmt.Sprintf("%.2f", metrics.Overview.MedianScore)})
	writer.Write([]string{})

	// Status breakdown
	writer.Write([]string{"Status", "Count", "Percentage"})
	for status, count := range metrics.StatusBreakdown {
		pct := float64(count) / float64(metrics.Overview.TotalIdeas) * 100
		writer.Write([]string{status, strconv.Itoa(count), fmt.Sprintf("%.1f", pct)})
	}

	return nil
}
