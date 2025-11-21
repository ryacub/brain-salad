// Package analytics provides analytics services for idea analysis and statistics.
package analytics

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Service handles analytics operations
type Service struct {
	repo   *database.Repository
	dbPath string
}

// NewService creates a new analytics service
func NewService(repo *database.Repository) *Service {
	return &Service{repo: repo}
}

// NewServiceWithDB creates a new analytics service with database path
func NewServiceWithDB(repo *database.Repository, dbPath string) *Service {
	return &Service{
		repo:   repo,
		dbPath: dbPath,
	}
}

// BasicStats contains basic statistical information about ideas
type BasicStats struct {
	TotalIdeas   int
	AverageScore float64
	HighScore    float64
	LowScore     float64
	HighCount    int // >= 7.0
	MediumCount  int // 5.0-7.0
	LowCount     int // < 5.0
}

// GetBasicStats calculates basic statistics from a set of ideas
func (s *Service) GetBasicStats(ideas []*models.Idea) BasicStats {
	stats := BasicStats{
		TotalIdeas: len(ideas),
		HighScore:  0,
		LowScore:   10,
	}

	if len(ideas) == 0 {
		return stats
	}

	var totalScore float64

	for _, idea := range ideas {
		totalScore += idea.FinalScore

		if idea.FinalScore > stats.HighScore {
			stats.HighScore = idea.FinalScore
		}
		if idea.FinalScore < stats.LowScore {
			stats.LowScore = idea.FinalScore
		}

		switch {
		case idea.FinalScore >= 7.0:
			stats.HighCount++
		case idea.FinalScore >= 5.0:
			stats.MediumCount++
		default:
			stats.LowCount++
		}
	}

	stats.AverageScore = totalScore / float64(len(ideas))

	return stats
}

// ScoreDistribution calculates the distribution percentages
func (s *Service) ScoreDistribution(stats BasicStats) (highPct, mediumPct, lowPct float64) {
	if stats.TotalIdeas == 0 {
		return 0, 0, 0
	}

	total := float64(stats.TotalIdeas)
	highPct = float64(stats.HighCount) / total * 100
	mediumPct = float64(stats.MediumCount) / total * 100
	lowPct = float64(stats.LowCount) / total * 100

	return highPct, mediumPct, lowPct
}

// SystemMetrics contains comprehensive system-wide metrics
type SystemMetrics struct {
	Overview          OverviewMetrics          `json:"overview"`
	StatusBreakdown   map[string]int           `json:"status_breakdown"`
	ScoreDistribution ScoreDistributionMetrics `json:"score_distribution"`
	PatternStats      []PatternStat            `json:"pattern_stats"`
	TimeMetrics       TimeMetrics              `json:"time_metrics"`
	DatabaseStats     DatabaseStats            `json:"database_stats"`
}

// OverviewMetrics contains overview statistics
type OverviewMetrics struct {
	TotalIdeas    int     `json:"total_ideas"`
	TotalPatterns int     `json:"total_patterns"`
	AverageScore  float64 `json:"average_score"`
	MedianScore   float64 `json:"median_score"`
	HighestScore  float64 `json:"highest_score"`
	LowestScore   float64 `json:"lowest_score"`
}

// ScoreDistributionMetrics contains score distribution data
type ScoreDistributionMetrics struct {
	Buckets     map[string]int     `json:"buckets"`
	StdDev      float64            `json:"std_dev"`
	Percentiles map[string]float64 `json:"percentiles"`
}

// PatternStat contains statistics for a single pattern
type PatternStat struct {
	Pattern    string  `json:"pattern"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// TimeMetrics contains time-based statistics
type TimeMetrics struct {
	OldestIdea      time.Time `json:"oldest_idea"`
	NewestIdea      time.Time `json:"newest_idea"`
	TotalDays       int       `json:"total_days"`
	IdeasPerDay     float64   `json:"ideas_per_day"`
	IdeasLast7Days  int       `json:"ideas_last_7_days"`
	IdeasLast30Days int       `json:"ideas_last_30_days"`
}

// DatabaseStats contains database statistics
type DatabaseStats struct {
	SizeBytes     int64  `json:"size_bytes"`
	SizeFormatted string `json:"size_formatted"`
	TableCount    int    `json:"table_count"`
	IndexCount    int    `json:"index_count"`
}

// CalculateSystemMetrics computes all system metrics
func (s *Service) CalculateSystemMetrics(ideas []*models.Idea) SystemMetrics {
	metrics := SystemMetrics{
		StatusBreakdown: make(map[string]int),
	}

	// Overview metrics
	metrics.Overview = s.calculateOverviewMetrics(ideas)

	// Status breakdown
	for _, idea := range ideas {
		metrics.StatusBreakdown[idea.Status]++
	}

	// Score distribution
	metrics.ScoreDistribution = s.calculateScoreDistribution(ideas)

	// Pattern statistics
	metrics.PatternStats = s.calculatePatternStats(ideas)

	// Time metrics
	metrics.TimeMetrics = s.calculateTimeMetrics(ideas)

	// Database statistics
	metrics.DatabaseStats = s.calculateDatabaseStats()

	return metrics
}

func (s *Service) calculateOverviewMetrics(ideas []*models.Idea) OverviewMetrics {
	if len(ideas) == 0 {
		return OverviewMetrics{}
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

	return OverviewMetrics{
		TotalIdeas:    len(ideas),
		TotalPatterns: len(patternSet),
		AverageScore:  sum / float64(len(ideas)),
		MedianScore:   CalculateMedian(scores),
		HighestScore:  highest,
		LowestScore:   lowest,
	}
}

func (s *Service) calculateScoreDistribution(ideas []*models.Idea) ScoreDistributionMetrics {
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

	return ScoreDistributionMetrics{
		Buckets: buckets,
		StdDev:  CalculateStdDev(scores),
		Percentiles: map[string]float64{
			"P50": CalculatePercentile(scores, 50),
			"P75": CalculatePercentile(scores, 75),
			"P90": CalculatePercentile(scores, 90),
			"P95": CalculatePercentile(scores, 95),
			"P99": CalculatePercentile(scores, 99),
		},
	}
}

func (s *Service) calculatePatternStats(ideas []*models.Idea) []PatternStat {
	patternCounts := make(map[string]int)

	for _, idea := range ideas {
		for _, pattern := range idea.Patterns {
			patternCounts[pattern]++
		}
	}

	stats := make([]PatternStat, 0, len(patternCounts))
	totalIdeas := float64(len(ideas))

	for pattern, count := range patternCounts {
		stats = append(stats, PatternStat{
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

func (s *Service) calculateTimeMetrics(ideas []*models.Idea) TimeMetrics {
	if len(ideas) == 0 {
		return TimeMetrics{}
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

	return TimeMetrics{
		OldestIdea:      oldest,
		NewestIdea:      newest,
		TotalDays:       totalDays,
		IdeasPerDay:     float64(len(ideas)) / float64(totalDays),
		IdeasLast7Days:  count7Days,
		IdeasLast30Days: count30Days,
	}
}

func (s *Service) calculateDatabaseStats() DatabaseStats {
	// Try to get database file size
	sizeBytes := int64(0)
	sizeFormatted := "Unknown"

	if s.dbPath != "" {
		if info, err := os.Stat(s.dbPath); err == nil {
			sizeBytes = info.Size()
			sizeFormatted = FormatBytes(sizeBytes)
		}
	}

	return DatabaseStats{
		SizeBytes:     sizeBytes,
		SizeFormatted: sizeFormatted,
		TableCount:    1, // ideas table
		IndexCount:    0,
	}
}

// Statistical helper functions

// CalculateMedian calculates the median of a slice of scores
func CalculateMedian(scores []float64) float64 {
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

// CalculateStdDev calculates the standard deviation of scores
func CalculateStdDev(scores []float64) float64 {
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

// CalculatePercentile calculates the given percentile of scores
func CalculatePercentile(scores []float64, p int) float64 {
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

// FormatBytes formats byte size into human-readable format
func FormatBytes(bytes int64) string {
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
