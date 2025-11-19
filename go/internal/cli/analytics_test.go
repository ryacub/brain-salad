package cli

import (
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

func TestDetectScoreOutliers(t *testing.T) {
	tests := []struct {
		name      string
		ideas     []*models.Idea
		threshold float64
		wantCount int
	}{
		{
			name: "two outliers in dataset",
			ideas: []*models.Idea{
				{ID: "1", FinalScore: 5.0},
				{ID: "2", FinalScore: 5.0},
				{ID: "3", FinalScore: 5.0},
				{ID: "4", FinalScore: 5.0},
				{ID: "5", FinalScore: 5.0},
				{ID: "6", FinalScore: 5.0},
				{ID: "7", FinalScore: 5.0},
				{ID: "8", FinalScore: 5.0},
				{ID: "9", FinalScore: 5.0},
				{ID: "10", FinalScore: 5.0},
				{ID: "11", FinalScore: 10.0}, // Outlier (high)
				{ID: "12", FinalScore: 0.0},  // Outlier (low)
			},
			threshold: 2.0,
			wantCount: 2,
		},
		{
			name: "no outliers",
			ideas: []*models.Idea{
				{ID: "1", FinalScore: 5.0},
				{ID: "2", FinalScore: 5.5},
				{ID: "3", FinalScore: 6.0},
				{ID: "4", FinalScore: 6.5},
			},
			threshold: 2.0,
			wantCount: 0,
		},
		{
			name: "all same scores (no variance)",
			ideas: []*models.Idea{
				{ID: "1", FinalScore: 5.0},
				{ID: "2", FinalScore: 5.0},
				{ID: "3", FinalScore: 5.0},
			},
			threshold: 2.0,
			wantCount: 0,
		},
		{
			name:      "insufficient data",
			ideas:     []*models.Idea{{ID: "1", FinalScore: 5.0}},
			threshold: 2.0,
			wantCount: 0,
		},
		{
			name:      "empty dataset",
			ideas:     []*models.Idea{},
			threshold: 2.0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outliers := detectScoreOutliers(tt.ideas, tt.threshold)
			if len(outliers) != tt.wantCount {
				t.Errorf("detectScoreOutliers() returned %d outliers, want %d", len(outliers), tt.wantCount)
			}

			// Verify outliers are sorted by deviation (highest first)
			for i := 1; i < len(outliers); i++ {
				prevDev := outliers[i-1].Deviation
				currDev := outliers[i].Deviation
				if prevDev < currDev {
					t.Errorf("outliers not sorted by deviation: %.2f > %.2f", prevDev, currDev)
				}
			}
		})
	}
}

func TestCalculateScoreStats(t *testing.T) {
	tests := []struct {
		name         string
		ideas        []*models.Idea
		wantMean     float64
		wantStdDev   float64
		deltaAllowed float64
	}{
		{
			name: "normal distribution",
			ideas: []*models.Idea{
				{FinalScore: 5.0},
				{FinalScore: 6.0},
				{FinalScore: 7.0},
			},
			wantMean:     6.0,
			wantStdDev:   0.816, // Approximately
			deltaAllowed: 0.01,
		},
		{
			name: "all same values",
			ideas: []*models.Idea{
				{FinalScore: 5.0},
				{FinalScore: 5.0},
				{FinalScore: 5.0},
			},
			wantMean:     5.0,
			wantStdDev:   0.0,
			deltaAllowed: 0.01,
		},
		{
			name:         "empty dataset",
			ideas:        []*models.Idea{},
			wantMean:     0.0,
			wantStdDev:   0.0,
			deltaAllowed: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean, stdDev := calculateScoreStats(tt.ideas)

			if mean != tt.wantMean {
				t.Errorf("calculateScoreStats() mean = %.2f, want %.2f", mean, tt.wantMean)
			}

			delta := stdDev - tt.wantStdDev
			if delta < 0 {
				delta = -delta
			}
			if delta > tt.deltaAllowed {
				t.Errorf("calculateScoreStats() stdDev = %.3f, want %.3f (±%.3f)", stdDev, tt.wantStdDev, tt.deltaAllowed)
			}
		})
	}
}

func TestDetectRarePatterns(t *testing.T) {
	tests := []struct {
		name          string
		ideas         []*models.Idea
		minOccurrence float64
		wantCount     int
	}{
		{
			name: "two rare patterns",
			ideas: []*models.Idea{
				{Patterns: []string{"common", "rare1"}},
				{Patterns: []string{"common"}},
				{Patterns: []string{"common"}},
				{Patterns: []string{"common", "rare2"}},
			},
			minOccurrence: 30.0, // rare1 and rare2 appear in 25% of ideas each
			wantCount:     2,
		},
		{
			name: "no rare patterns",
			ideas: []*models.Idea{
				{Patterns: []string{"common"}},
				{Patterns: []string{"common"}},
				{Patterns: []string{"common"}},
			},
			minOccurrence: 10.0,
			wantCount:     0,
		},
		{
			name: "all patterns rare",
			ideas: []*models.Idea{
				{Patterns: []string{"pattern1"}},
				{Patterns: []string{"pattern2"}},
				{Patterns: []string{"pattern3"}},
				{Patterns: []string{"pattern4"}},
			},
			minOccurrence: 30.0, // All patterns appear in 25% of ideas
			wantCount:     4,
		},
		{
			name:          "empty dataset",
			ideas:         []*models.Idea{},
			minOccurrence: 10.0,
			wantCount:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rarePatterns := detectRarePatterns(tt.ideas, tt.minOccurrence, false)
			if len(rarePatterns) != tt.wantCount {
				t.Errorf("detectRarePatterns() returned %d patterns, want %d", len(rarePatterns), tt.wantCount)
			}

			// Verify patterns are sorted by rarity (least common first)
			for i := 1; i < len(rarePatterns); i++ {
				prevPerc := rarePatterns[i-1].Percentage
				currPerc := rarePatterns[i].Percentage
				if prevPerc > currPerc {
					t.Errorf("patterns not sorted by rarity: %.2f <= %.2f", prevPerc, currPerc)
				}
			}
		})
	}
}

func TestDetectRarePatterns_WithIDs(t *testing.T) {
	ideas := []*models.Idea{
		{ID: "id1", Patterns: []string{"rare"}},
		{ID: "id2", Patterns: []string{"common"}},
		{ID: "id3", Patterns: []string{"common"}},
		{ID: "id4", Patterns: []string{"common"}},
		{ID: "id5", Patterns: []string{"common"}},
		{ID: "id6", Patterns: []string{"rare"}},
	}

	rarePatterns := detectRarePatterns(ideas, 40.0, true)

	// "rare" appears in 33% of ideas (2 out of 6), which is less than 40%
	// "common" appears in 67% of ideas (4 out of 6), which is greater than 40%
	if len(rarePatterns) != 1 {
		t.Fatalf("expected 1 rare pattern, got %d", len(rarePatterns))
	}

	if rarePatterns[0].Pattern != "rare" {
		t.Errorf("expected pattern 'rare', got '%s'", rarePatterns[0].Pattern)
	}

	if len(rarePatterns[0].IdeaIDs) != 2 {
		t.Errorf("expected 2 idea IDs, got %d", len(rarePatterns[0].IdeaIDs))
	}
}

func TestDetectTimingAnomalies(t *testing.T) {
	tests := []struct {
		name      string
		ideas     []*models.Idea
		threshold float64
		wantCount int
	}{
		{
			name: "spike detected",
			ideas: []*models.Idea{
				{CreatedAt: parseDate("2025-01-01")},
				{CreatedAt: parseDate("2025-01-01")},
				{CreatedAt: parseDate("2025-01-02")},
				{CreatedAt: parseDate("2025-01-03")},
				{CreatedAt: parseDate("2025-01-04")},
				{CreatedAt: parseDate("2025-01-05")},
				{CreatedAt: parseDate("2025-01-06")},
				{CreatedAt: parseDate("2025-01-07")},
				{CreatedAt: parseDate("2025-01-08")},
				{CreatedAt: parseDate("2025-01-09")}, // 9 more ideas on one day = spike
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
				{CreatedAt: parseDate("2025-01-09")},
			},
			threshold: 2.0,
			wantCount: 1, // One anomaly date (2025-01-09)
		},
		{
			name: "uniform distribution",
			ideas: []*models.Idea{
				{CreatedAt: parseDate("2025-01-01")},
				{CreatedAt: parseDate("2025-01-02")},
				{CreatedAt: parseDate("2025-01-03")},
				{CreatedAt: parseDate("2025-01-04")},
				{CreatedAt: parseDate("2025-01-05")},
				{CreatedAt: parseDate("2025-01-06")},
				{CreatedAt: parseDate("2025-01-07")},
			},
			threshold: 2.0,
			wantCount: 0,
		},
		{
			name: "insufficient data",
			ideas: []*models.Idea{
				{CreatedAt: parseDate("2025-01-01")},
				{CreatedAt: parseDate("2025-01-02")},
			},
			threshold: 2.0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anomalies := detectTimingAnomalies(tt.ideas, tt.threshold)
			if len(anomalies) != tt.wantCount {
				t.Errorf("detectTimingAnomalies() returned %d anomalies, want %d", len(anomalies), tt.wantCount)
			}
		})
	}
}

func TestDetectRecommendationIssues(t *testing.T) {
	tests := []struct {
		name      string
		ideas     []*models.Idea
		wantCount int
	}{
		{
			name: "multiple issues",
			ideas: []*models.Idea{
				{ID: "1", FinalScore: 3.0, Recommendation: "pursue"}, // Issue: low score but pursue
				{ID: "2", FinalScore: 8.0, Recommendation: "pursue"}, // OK
				{ID: "3", FinalScore: 7.0, Recommendation: "reject"}, // Issue: high score but reject
				{ID: "4", FinalScore: 9.0, Recommendation: "defer"},  // Issue: very high score but defer
			},
			wantCount: 3,
		},
		{
			name: "no issues",
			ideas: []*models.Idea{
				{ID: "1", FinalScore: 8.0, Recommendation: "pursue"},
				{ID: "2", FinalScore: 4.0, Recommendation: "reject"},
				{ID: "3", FinalScore: 6.0, Recommendation: "defer"},
			},
			wantCount: 0,
		},
		{
			name: "edge cases",
			ideas: []*models.Idea{
				{ID: "1", FinalScore: 7.0, Recommendation: "pursue"}, // OK (exactly 7.0)
				{ID: "2", FinalScore: 5.0, Recommendation: "reject"}, // OK (exactly 5.0)
				{ID: "3", FinalScore: 8.0, Recommendation: "defer"},  // OK (not very high)
			},
			wantCount: 0,
		},
		{
			name:      "empty dataset",
			ideas:     []*models.Idea{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := detectRecommendationIssues(tt.ideas)
			if len(issues) != tt.wantCount {
				t.Errorf("detectRecommendationIssues() returned %d issues, want %d", len(issues), tt.wantCount)
			}
		})
	}
}

func TestCalculateMean(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{
			name:   "normal values",
			values: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			want:   3.0,
		},
		{
			name:   "single value",
			values: []float64{5.0},
			want:   5.0,
		},
		{
			name:   "empty slice",
			values: []float64{},
			want:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMean(tt.values)
			if got != tt.want {
				t.Errorf("calculateMean() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestCalculateStdDevFromValues(t *testing.T) {
	tests := []struct {
		name         string
		values       []float64
		mean         float64
		want         float64
		deltaAllowed float64
	}{
		{
			name:         "normal distribution",
			values:       []float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0},
			mean:         5.0,
			want:         2.0,
			deltaAllowed: 0.01,
		},
		{
			name:         "no variance",
			values:       []float64{5.0, 5.0, 5.0},
			mean:         5.0,
			want:         0.0,
			deltaAllowed: 0.01,
		},
		{
			name:         "empty slice",
			values:       []float64{},
			mean:         0.0,
			want:         0.0,
			deltaAllowed: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStdDevFromValues(tt.values, tt.mean)

			delta := got - tt.want
			if delta < 0 {
				delta = -delta
			}
			if delta > tt.deltaAllowed {
				t.Errorf("calculateStdDevFromValues() = %.3f, want %.3f (±%.3f)", got, tt.want, tt.deltaAllowed)
			}
		})
	}
}

// Helper function to parse date strings
func parseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}
	return t
}
