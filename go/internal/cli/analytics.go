package cli

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

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
	cmd.AddCommand(newAnalyticsAnomalyCommand())

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
