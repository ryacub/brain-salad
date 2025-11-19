package cli

import (
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

func TestCalculateOverviewMetrics(t *testing.T) {
	t.Run("empty ideas", func(t *testing.T) {
		ideas := []*models.Idea{}
		metrics := calculateOverviewMetrics(ideas)

		if metrics.TotalIdeas != 0 {
			t.Errorf("Expected 0 ideas, got %d", metrics.TotalIdeas)
		}
		if metrics.TotalPatterns != 0 {
			t.Errorf("Expected 0 patterns, got %d", metrics.TotalPatterns)
		}
	})

	t.Run("multiple ideas with patterns", func(t *testing.T) {
		ideas := []*models.Idea{
			{FinalScore: 5.0, Patterns: []string{"p1"}},
			{FinalScore: 7.0, Patterns: []string{"p1", "p2"}},
			{FinalScore: 3.0, Patterns: []string{"p3"}},
		}

		metrics := calculateOverviewMetrics(ideas)

		if metrics.TotalIdeas != 3 {
			t.Errorf("Expected 3 ideas, got %d", metrics.TotalIdeas)
		}
		if metrics.TotalPatterns != 3 {
			t.Errorf("Expected 3 patterns, got %d", metrics.TotalPatterns)
		}
		expectedAvg := 5.0
		if metrics.AverageScore != expectedAvg {
			t.Errorf("Expected avg %.2f, got %.2f", expectedAvg, metrics.AverageScore)
		}
		if metrics.HighestScore != 7.0 {
			t.Errorf("Expected highest 7.0, got %.2f", metrics.HighestScore)
		}
		if metrics.LowestScore != 3.0 {
			t.Errorf("Expected lowest 3.0, got %.2f", metrics.LowestScore)
		}
		if metrics.MedianScore != 5.0 {
			t.Errorf("Expected median 5.0, got %.2f", metrics.MedianScore)
		}
	})

	t.Run("duplicate patterns", func(t *testing.T) {
		ideas := []*models.Idea{
			{FinalScore: 5.0, Patterns: []string{"p1", "p2"}},
			{FinalScore: 7.0, Patterns: []string{"p1", "p2"}},
		}

		metrics := calculateOverviewMetrics(ideas)

		if metrics.TotalPatterns != 2 {
			t.Errorf("Expected 2 unique patterns, got %d", metrics.TotalPatterns)
		}
	})
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name     string
		scores   []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5.0}, 5.0},
		{"odd count", []float64{1, 2, 3, 4, 5}, 3.0},
		{"even count", []float64{1, 2, 3, 4, 5, 6}, 3.5},
		{"unsorted", []float64{5, 1, 3, 2, 4}, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMedian(tt.scores)
			if result != tt.expected {
				t.Errorf("Expected %.1f, got %.1f", tt.expected, result)
			}
		})
	}
}

func TestCalculateStdDev(t *testing.T) {
	t.Run("empty scores", func(t *testing.T) {
		result := calculateStdDev([]float64{})
		if result != 0 {
			t.Errorf("Expected 0, got %.2f", result)
		}
	})

	t.Run("identical values", func(t *testing.T) {
		result := calculateStdDev([]float64{5, 5, 5, 5})
		if result != 0 {
			t.Errorf("Expected 0 for identical values, got %.2f", result)
		}
	})

	t.Run("standard calculation", func(t *testing.T) {
		// Mean = 5, variance = 2, stddev = sqrt(2) ≈ 1.414
		result := calculateStdDev([]float64{3, 4, 5, 6, 7})
		expected := 1.414
		if result < expected-0.01 || result > expected+0.01 {
			t.Errorf("Expected ~%.3f, got %.3f", expected, result)
		}
	})
}

func TestCalculatePercentile(t *testing.T) {
	scores := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		percentile int
		expected   float64
		tolerance  float64
	}{
		{50, 5.5, 0.01},   // Median of 10 values
		{75, 7.75, 0.01},  // 75th percentile
		{90, 9.1, 0.01},   // 90th percentile
		{95, 9.55, 0.01},  // 95th percentile
		{99, 9.91, 0.01},  // 99th percentile
		{0, 1.0, 0.01},    // Minimum
		{100, 10.0, 0.01}, // Maximum
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.percentile)), func(t *testing.T) {
			result := calculatePercentile(scores, tt.percentile)
			if result < tt.expected-tt.tolerance || result > tt.expected+tt.tolerance {
				t.Errorf("P%d: expected %.2f (±%.2f), got %.2f", tt.percentile, tt.expected, tt.tolerance, result)
			}
		})
	}

	t.Run("empty scores", func(t *testing.T) {
		result := calculatePercentile([]float64{}, 50)
		if result != 0 {
			t.Errorf("Expected 0 for empty scores, got %.1f", result)
		}
	})
}

func TestCalculateScoreDistribution(t *testing.T) {
	ideas := []*models.Idea{
		{FinalScore: 1.0},  // 0-2
		{FinalScore: 3.0},  // 2-4
		{FinalScore: 5.0},  // 4-6
		{FinalScore: 7.0},  // 6-8
		{FinalScore: 9.0},  // 8-10
		{FinalScore: 5.5},  // 4-6
	}

	metrics := calculateScoreDistribution(ideas)

	if metrics.Buckets["0-2"] != 1 {
		t.Errorf("Expected 1 in 0-2 bucket, got %d", metrics.Buckets["0-2"])
	}
	if metrics.Buckets["2-4"] != 1 {
		t.Errorf("Expected 1 in 2-4 bucket, got %d", metrics.Buckets["2-4"])
	}
	if metrics.Buckets["4-6"] != 2 {
		t.Errorf("Expected 2 in 4-6 bucket, got %d", metrics.Buckets["4-6"])
	}
	if metrics.Buckets["6-8"] != 1 {
		t.Errorf("Expected 1 in 6-8 bucket, got %d", metrics.Buckets["6-8"])
	}
	if metrics.Buckets["8-10"] != 1 {
		t.Errorf("Expected 1 in 8-10 bucket, got %d", metrics.Buckets["8-10"])
	}

	if metrics.StdDev == 0 {
		t.Error("Expected non-zero standard deviation")
	}

	// P50 should be around 5.5 (median of 6 values: 1, 3, 5, 5.5, 7, 9)
	if metrics.Percentiles["P50"] < 5.0 || metrics.Percentiles["P50"] > 6.0 {
		t.Errorf("Expected P50 between 5.0 and 6.0, got %.2f", metrics.Percentiles["P50"])
	}
}

func TestCalculatePatternStats(t *testing.T) {
	ideas := []*models.Idea{
		{Patterns: []string{"pattern1", "pattern2"}},
		{Patterns: []string{"pattern1"}},
		{Patterns: []string{"pattern3"}},
		{Patterns: []string{"pattern1", "pattern3"}},
	}

	stats := calculatePatternStats(ideas)

	// pattern1 should appear 3 times (75%)
	// pattern2 should appear 1 time (25%)
	// pattern3 should appear 2 times (50%)

	if len(stats) != 3 {
		t.Errorf("Expected 3 patterns, got %d", len(stats))
	}

	// Should be sorted by count (descending)
	if stats[0].Pattern != "pattern1" {
		t.Errorf("Expected pattern1 first, got %s", stats[0].Pattern)
	}
	if stats[0].Count != 3 {
		t.Errorf("Expected pattern1 count 3, got %d", stats[0].Count)
	}
	if stats[0].Percentage != 75.0 {
		t.Errorf("Expected pattern1 percentage 75%%, got %.1f%%", stats[0].Percentage)
	}

	if stats[1].Pattern != "pattern3" {
		t.Errorf("Expected pattern3 second, got %s", stats[1].Pattern)
	}
	if stats[1].Count != 2 {
		t.Errorf("Expected pattern3 count 2, got %d", stats[1].Count)
	}
}

func TestCalculateTimeMetrics(t *testing.T) {
	now := time.Now()
	oldest := now.AddDate(0, 0, -30)
	newest := now.AddDate(0, 0, -1)

	ideas := []*models.Idea{
		{CreatedAt: oldest},
		{CreatedAt: now.AddDate(0, 0, -15)},
		{CreatedAt: newest},
		{CreatedAt: now.AddDate(0, 0, -3)}, // Within last 7 days
		{CreatedAt: now.AddDate(0, 0, -2)}, // Within last 7 days
	}

	metrics := calculateTimeMetrics(ideas)

	if !metrics.OldestIdea.Equal(oldest) {
		t.Errorf("Expected oldest %v, got %v", oldest, metrics.OldestIdea)
	}
	if !metrics.NewestIdea.Equal(newest) {
		t.Errorf("Expected newest %v, got %v", newest, metrics.NewestIdea)
	}

	expectedDays := 29
	if metrics.TotalDays != expectedDays {
		t.Errorf("Expected %d days, got %d", expectedDays, metrics.TotalDays)
	}

	expectedIdeasPerDay := float64(5) / float64(expectedDays)
	if metrics.IdeasPerDay < expectedIdeasPerDay-0.01 || metrics.IdeasPerDay > expectedIdeasPerDay+0.01 {
		t.Errorf("Expected %.2f ideas/day, got %.2f", expectedIdeasPerDay, metrics.IdeasPerDay)
	}

	// Should have 2-3 ideas in last 7 days (depending on exact timing)
	if metrics.IdeasLast7Days < 2 || metrics.IdeasLast7Days > 3 {
		t.Errorf("Expected 2-3 ideas in last 7 days, got %d", metrics.IdeasLast7Days)
	}

	// Should have 4-5 ideas in last 30 days (depending on exact timing)
	if metrics.IdeasLast30Days < 4 || metrics.IdeasLast30Days > 5 {
		t.Errorf("Expected 4-5 ideas in last 30 days, got %d", metrics.IdeasLast30Days)
	}
}

func TestCalculateTimeMetricsEmpty(t *testing.T) {
	metrics := calculateTimeMetrics([]*models.Idea{})

	if !metrics.OldestIdea.IsZero() {
		t.Error("Expected zero time for oldest idea")
	}
	if !metrics.NewestIdea.IsZero() {
		t.Error("Expected zero time for newest idea")
	}
	if metrics.TotalDays != 0 {
		t.Errorf("Expected 0 days, got %d", metrics.TotalDays)
	}
}

func TestCalculateTimeMetricsSingleDay(t *testing.T) {
	now := time.Now()
	ideas := []*models.Idea{
		{CreatedAt: now},
		{CreatedAt: now.Add(time.Hour)},
	}

	metrics := calculateTimeMetrics(ideas)

	// Should default to 1 day to avoid division by zero
	if metrics.TotalDays != 1 {
		t.Errorf("Expected 1 day for same-day ideas, got %d", metrics.TotalDays)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 bytes"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
		{1610612736, "1.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCalculateSystemMetrics(t *testing.T) {
	ideas := []*models.Idea{
		{
			FinalScore: 5.0,
			Status:     "active",
			Patterns:   []string{"p1"},
			CreatedAt:  time.Now(),
		},
		{
			FinalScore: 7.0,
			Status:     "active",
			Patterns:   []string{"p1", "p2"},
			CreatedAt:  time.Now().AddDate(0, 0, -1),
		},
		{
			FinalScore: 3.0,
			Status:     "archived",
			Patterns:   []string{"p3"},
			CreatedAt:  time.Now().AddDate(0, 0, -2),
		},
	}

	metrics := calculateSystemMetrics(ideas)

	// Overview
	if metrics.Overview.TotalIdeas != 3 {
		t.Errorf("Expected 3 total ideas, got %d", metrics.Overview.TotalIdeas)
	}
	if metrics.Overview.TotalPatterns != 3 {
		t.Errorf("Expected 3 unique patterns, got %d", metrics.Overview.TotalPatterns)
	}

	// Status breakdown
	if metrics.StatusBreakdown["active"] != 2 {
		t.Errorf("Expected 2 active ideas, got %d", metrics.StatusBreakdown["active"])
	}
	if metrics.StatusBreakdown["archived"] != 1 {
		t.Errorf("Expected 1 archived idea, got %d", metrics.StatusBreakdown["archived"])
	}

	// Score distribution
	if len(metrics.ScoreDistribution.Buckets) != 5 {
		t.Errorf("Expected 5 buckets, got %d", len(metrics.ScoreDistribution.Buckets))
	}

	// Pattern stats
	if len(metrics.PatternStats) != 3 {
		t.Errorf("Expected 3 pattern stats, got %d", len(metrics.PatternStats))
	}

	// Time metrics should be populated
	if metrics.TimeMetrics.TotalDays == 0 {
		t.Error("Expected non-zero total days")
	}
}
