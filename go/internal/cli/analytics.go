package cli

import (
	"fmt"
	"os"

	"github.com/rayyacub/telos-idea-matrix/internal/analytics"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
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
