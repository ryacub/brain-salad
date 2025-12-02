package analytics

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

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

// NewAnomalyCommand creates the analytics anomaly subcommand
func NewAnomalyCommand(getContext func() *CLIContext) *cobra.Command {
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
			return runAnomalyDetection(getContext, anomalyOptions{
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

func runAnomalyDetection(getContext func() *CLIContext, opts anomalyOptions) error {
	ctx := getContext()
	if ctx == nil {
		return fmt.Errorf("CLI context not initialized")
	}

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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
