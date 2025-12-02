package analytics

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

	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

// NewPerformanceCommand creates the analytics performance subcommand
func NewPerformanceCommand(getContext func() *CLIContext) *cobra.Command {
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
			return runPerformanceAnalytics(getContext, performanceOptions{
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

func runPerformanceAnalytics(getContext func() *CLIContext, opts performanceOptions) error {
	ctx := getContext()
	if ctx == nil {
		return fmt.Errorf("CLI context not initialized")
	}

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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
