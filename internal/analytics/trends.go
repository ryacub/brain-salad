package analytics

import (
	"fmt"
	"sort"
	"time"

	"github.com/ryacub/telos-idea-matrix/internal/models"
)

// TrendData represents aggregated metrics for a specific time period
type TrendData struct {
	Period      string   // Time period identifier (e.g., "2024-W12", "2024-01", "2024-01-15")
	AvgScore    float64  // Average score for the period
	IdeaCount   int      // Number of ideas in the period
	TopPatterns []string // Most common patterns in the period
}

// CalculateScoreTrends groups ideas by time period and calculates average scores
// groupBy can be "day", "week", or "month"
func CalculateScoreTrends(ideas []*models.Idea, groupBy string) []TrendData {
	if len(ideas) == 0 {
		return []TrendData{}
	}

	groups := make(map[string][]*models.Idea)

	// Group ideas by time period
	for _, idea := range ideas {
		var key string
		switch groupBy {
		case "week":
			year, week := idea.CreatedAt.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case "month":
			key = idea.CreatedAt.Format("2006-01")
		case "day":
			key = idea.CreatedAt.Format("2006-01-02")
		default:
			// Default to week if invalid groupBy
			year, week := idea.CreatedAt.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		}
		groups[key] = append(groups[key], idea)
	}

	// Calculate trend data for each period
	trends := make([]TrendData, 0, len(groups))
	for period, periodIdeas := range groups {
		totalScore := 0.0
		for _, idea := range periodIdeas {
			totalScore += idea.FinalScore
		}

		avgScore := totalScore / float64(len(periodIdeas))

		trends = append(trends, TrendData{
			Period:    period,
			AvgScore:  avgScore,
			IdeaCount: len(periodIdeas),
		})
	}

	// Sort by period chronologically
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Period < trends[j].Period
	})

	return trends
}

// CalculatePatternFrequency counts how often each pattern appears across all ideas
func CalculatePatternFrequency(ideas []*models.Idea) map[string]int {
	freq := make(map[string]int)

	for _, idea := range ideas {
		if idea.Analysis == nil {
			continue
		}

		for _, pattern := range idea.Analysis.DetectedPatterns {
			freq[pattern.Name]++
		}
	}

	return freq
}

// CalculateCreationRate returns the average number of ideas created per day
// over the specified number of days
func CalculateCreationRate(ideas []*models.Idea, days int) float64 {
	if len(ideas) == 0 || days <= 0 {
		return 0.0
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	count := 0

	for _, idea := range ideas {
		if idea.CreatedAt.After(cutoff) {
			count++
		}
	}

	return float64(count) / float64(days)
}

// GetTopPatterns returns the N most frequently occurring patterns
func GetTopPatterns(ideas []*models.Idea, n int) []string {
	freq := CalculatePatternFrequency(ideas)

	if len(freq) == 0 {
		return []string{}
	}

	// Convert map to slice for sorting
	type patternCount struct {
		name  string
		count int
	}

	patterns := make([]patternCount, 0, len(freq))
	for name, count := range freq {
		patterns = append(patterns, patternCount{name: name, count: count})
	}

	// Sort by count (descending), then by name (ascending) for consistency
	sort.Slice(patterns, func(i, j int) bool {
		if patterns[i].count == patterns[j].count {
			return patterns[i].name < patterns[j].name
		}
		return patterns[i].count > patterns[j].count
	})

	// Return top N
	result := make([]string, 0, n)
	for i := 0; i < len(patterns) && i < n; i++ {
		result = append(result, patterns[i].name)
	}

	return result
}

// CalculateTrendDirection determines if a metric is trending up, down, or neutral
func CalculateTrendDirection(trends []TrendData) string {
	if len(trends) < 2 {
		return "neutral"
	}

	// Compare recent periods vs older periods
	mid := len(trends) / 2
	recentAvg := 0.0
	olderAvg := 0.0

	for i := mid; i < len(trends); i++ {
		recentAvg += trends[i].AvgScore
	}
	recentAvg /= float64(len(trends) - mid)

	for i := 0; i < mid; i++ {
		olderAvg += trends[i].AvgScore
	}
	olderAvg /= float64(mid)

	diff := recentAvg - olderAvg

	if diff > 0.5 {
		return "up"
	} else if diff < -0.5 {
		return "down"
	}
	return "neutral"
}
