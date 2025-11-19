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
	cmd.AddCommand(newAnalyticsPerformanceCommand())

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

// newAnalyticsPerformanceCommand creates the analytics performance subcommand
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
	if !contains(validGroups, opts.groupBy) {
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
			StdDev:       calculateStdDev(group.Ideas),
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

func calculateStdDev(ideas []*models.Idea) float64 {
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
	writer.Write([]string{
		"Period", "TotalIdeas", "AverageScore", "MedianScore",
		"HighestScore", "LowestScore", "StdDev",
	})

	// Data rows
	for _, report := range reports {
		writer.Write([]string{
			report.Period,
			strconv.Itoa(report.TotalIdeas),
			fmt.Sprintf("%.2f", report.AverageScore),
			fmt.Sprintf("%.2f", report.MedianScore),
			fmt.Sprintf("%.2f", report.HighestScore),
			fmt.Sprintf("%.2f", report.LowestScore),
			fmt.Sprintf("%.2f", report.StdDev),
		})
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
