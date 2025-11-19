package analytics

import (
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTrends_ScoreByWeek tests weekly score trend calculation
func TestTrends_ScoreByWeek(t *testing.T) {
	now := time.Now()

	// Create test ideas across different weeks
	ideas := []*models.Idea{
		{
			ID:         "1",
			Content:    "Idea 1",
			FinalScore: 8.0,
			CreatedAt:  now.AddDate(0, 0, -14), // 2 weeks ago
		},
		{
			ID:         "2",
			Content:    "Idea 2",
			FinalScore: 7.5,
			CreatedAt:  now.AddDate(0, 0, -14), // 2 weeks ago
		},
		{
			ID:         "3",
			Content:    "Idea 3",
			FinalScore: 9.0,
			CreatedAt:  now.AddDate(0, 0, -7), // 1 week ago
		},
		{
			ID:         "4",
			Content:    "Idea 4",
			FinalScore: 6.0,
			CreatedAt:  now, // this week
		},
		{
			ID:         "5",
			Content:    "Idea 5",
			FinalScore: 7.0,
			CreatedAt:  now, // this week
		},
	}

	trends := CalculateScoreTrends(ideas, "week")

	// Should have 3 weeks of data
	assert.Len(t, trends, 3, "should have trends for 3 different weeks")

	// Verify trends are sorted by period
	for i := 0; i < len(trends)-1; i++ {
		assert.LessOrEqual(t, trends[i].Period, trends[i+1].Period,
			"trends should be sorted by period")
	}

	// Verify calculations for each period
	for _, trend := range trends {
		assert.Greater(t, trend.IdeaCount, 0, "each period should have at least one idea")
		assert.Greater(t, trend.AvgScore, 0.0, "average score should be greater than 0")
		assert.LessOrEqual(t, trend.AvgScore, 10.0, "average score should be <= 10")
	}

	// Check specific week calculations
	// Find the week from 2 weeks ago (should have avg of (8.0 + 7.5) / 2 = 7.75)
	found := false
	for _, trend := range trends {
		if trend.IdeaCount == 2 {
			assert.InDelta(t, 7.75, trend.AvgScore, 0.01, "average should be 7.75 for 2-idea week")
			found = true
			break
		}
	}
	assert.True(t, found, "should find the week with 2 ideas")
}

// TestTrends_ScoreByMonth tests monthly score trend calculation
func TestTrends_ScoreByMonth(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	ideas := []*models.Idea{
		{
			ID:         "1",
			Content:    "January idea 1",
			FinalScore: 8.0,
			CreatedAt:  baseTime,
		},
		{
			ID:         "2",
			Content:    "January idea 2",
			FinalScore: 6.0,
			CreatedAt:  baseTime.AddDate(0, 0, 10),
		},
		{
			ID:         "3",
			Content:    "February idea 1",
			FinalScore: 9.0,
			CreatedAt:  baseTime.AddDate(0, 1, 0),
		},
		{
			ID:         "4",
			Content:    "March idea 1",
			FinalScore: 7.0,
			CreatedAt:  baseTime.AddDate(0, 2, 0),
		},
	}

	trends := CalculateScoreTrends(ideas, "month")

	// Should have 3 months of data
	assert.Len(t, trends, 3, "should have trends for 3 different months")

	// Verify sorted order
	assert.Equal(t, "2024-01", trends[0].Period, "first period should be January 2024")
	assert.Equal(t, "2024-02", trends[1].Period, "second period should be February 2024")
	assert.Equal(t, "2024-03", trends[2].Period, "third period should be March 2024")

	// Check January stats: 2 ideas, avg (8.0 + 6.0) / 2 = 7.0
	assert.Equal(t, 2, trends[0].IdeaCount, "January should have 2 ideas")
	assert.InDelta(t, 7.0, trends[0].AvgScore, 0.01, "January average should be 7.0")

	// Check February stats: 1 idea, avg 9.0
	assert.Equal(t, 1, trends[1].IdeaCount, "February should have 1 idea")
	assert.InDelta(t, 9.0, trends[1].AvgScore, 0.01, "February average should be 9.0")

	// Check March stats: 1 idea, avg 7.0
	assert.Equal(t, 1, trends[2].IdeaCount, "March should have 1 idea")
	assert.InDelta(t, 7.0, trends[2].AvgScore, 0.01, "March average should be 7.0")
}

// TestTrends_ScoreByDay tests daily score trend calculation
func TestTrends_ScoreByDay(t *testing.T) {
	baseTime := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	ideas := []*models.Idea{
		{
			ID:         "1",
			Content:    "Day 1 idea 1",
			FinalScore: 8.0,
			CreatedAt:  baseTime,
		},
		{
			ID:         "2",
			Content:    "Day 1 idea 2",
			FinalScore: 6.0,
			CreatedAt:  baseTime.Add(2 * time.Hour),
		},
		{
			ID:         "3",
			Content:    "Day 2 idea 1",
			FinalScore: 9.0,
			CreatedAt:  baseTime.AddDate(0, 0, 1),
		},
	}

	trends := CalculateScoreTrends(ideas, "day")

	assert.Len(t, trends, 2, "should have trends for 2 different days")
	assert.Equal(t, "2024-03-15", trends[0].Period)
	assert.Equal(t, "2024-03-16", trends[1].Period)
	assert.Equal(t, 2, trends[0].IdeaCount)
	assert.Equal(t, 1, trends[1].IdeaCount)
	assert.InDelta(t, 7.0, trends[0].AvgScore, 0.01)
	assert.InDelta(t, 9.0, trends[1].AvgScore, 0.01)
}

// TestTrends_EmptyIdeas tests trend calculation with no ideas
func TestTrends_EmptyIdeas(t *testing.T) {
	ideas := []*models.Idea{}

	trends := CalculateScoreTrends(ideas, "week")

	assert.Empty(t, trends, "should return empty trends for no ideas")
}

// TestTrends_PatternFrequency tests pattern frequency calculation
func TestTrends_PatternFrequency(t *testing.T) {
	ideas := []*models.Idea{
		{
			ID:      "1",
			Content: "Idea 1",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "shiny-object-syndrome", Severity: "high"},
					{Name: "analysis-paralysis", Severity: "medium"},
				},
			},
		},
		{
			ID:      "2",
			Content: "Idea 2",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "shiny-object-syndrome", Severity: "high"},
					{Name: "premature-optimization", Severity: "low"},
				},
			},
		},
		{
			ID:      "3",
			Content: "Idea 3",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "shiny-object-syndrome", Severity: "critical"},
				},
			},
		},
		{
			ID:       "4",
			Content:  "Idea 4",
			Analysis: nil, // No analysis
		},
	}

	freq := CalculatePatternFrequency(ideas)

	// Verify frequency counts
	assert.Equal(t, 3, freq["shiny-object-syndrome"], "shiny-object-syndrome should appear 3 times")
	assert.Equal(t, 1, freq["analysis-paralysis"], "analysis-paralysis should appear 1 time")
	assert.Equal(t, 1, freq["premature-optimization"], "premature-optimization should appear 1 time")
	assert.NotContains(t, freq, "non-existent-pattern", "should not contain non-existent patterns")
}

// TestTrends_PatternFrequency_EmptyAnalysis tests with ideas that have no analysis
func TestTrends_PatternFrequency_EmptyAnalysis(t *testing.T) {
	ideas := []*models.Idea{
		{
			ID:       "1",
			Content:  "Idea without analysis",
			Analysis: nil,
		},
		{
			ID:      "2",
			Content: "Idea with empty patterns",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{},
			},
		},
	}

	freq := CalculatePatternFrequency(ideas)

	assert.Empty(t, freq, "should return empty frequency map when no patterns exist")
}

// TestTrends_IdeaCreationRate tests the creation rate calculation
func TestTrends_IdeaCreationRate(t *testing.T) {
	now := time.Now()

	ideas := []*models.Idea{
		{ID: "1", Content: "Idea 1", CreatedAt: now.AddDate(0, 0, -1)},  // 1 day ago
		{ID: "2", Content: "Idea 2", CreatedAt: now.AddDate(0, 0, -2)},  // 2 days ago
		{ID: "3", Content: "Idea 3", CreatedAt: now.AddDate(0, 0, -5)},  // 5 days ago
		{ID: "4", Content: "Idea 4", CreatedAt: now.AddDate(0, 0, -10)}, // 10 days ago (outside 7-day window)
		{ID: "5", Content: "Idea 5", CreatedAt: now.AddDate(0, 0, -20)}, // 20 days ago (outside 7-day window)
	}

	// Test 7-day creation rate
	rate7 := CalculateCreationRate(ideas, 7)
	assert.InDelta(t, 3.0/7.0, rate7, 0.01, "should have 3 ideas in last 7 days")

	// Test 14-day creation rate
	rate14 := CalculateCreationRate(ideas, 14)
	assert.InDelta(t, 4.0/14.0, rate14, 0.01, "should have 4 ideas in last 14 days")

	// Test 30-day creation rate (all ideas)
	rate30 := CalculateCreationRate(ideas, 30)
	assert.InDelta(t, 5.0/30.0, rate30, 0.01, "should have all 5 ideas in last 30 days")
}

// TestTrends_IdeaCreationRate_EmptyIdeas tests creation rate with no ideas
func TestTrends_IdeaCreationRate_EmptyIdeas(t *testing.T) {
	ideas := []*models.Idea{}

	rate := CalculateCreationRate(ideas, 7)

	assert.Equal(t, 0.0, rate, "should return 0 rate for no ideas")
}

// TestTrends_TopPatterns tests getting top N most frequent patterns
func TestTrends_TopPatterns(t *testing.T) {
	ideas := []*models.Idea{
		{
			ID: "1",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "pattern-a"},
					{Name: "pattern-b"},
				},
			},
		},
		{
			ID: "2",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "pattern-a"},
					{Name: "pattern-c"},
				},
			},
		},
		{
			ID: "3",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "pattern-a"},
				},
			},
		},
		{
			ID: "4",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "pattern-b"},
					{Name: "pattern-d"},
				},
			},
		},
	}

	// Get top 3 patterns
	topPatterns := GetTopPatterns(ideas, 3)

	require.Len(t, topPatterns, 3, "should return top 3 patterns")

	// pattern-a should be first (appears 3 times)
	assert.Equal(t, "pattern-a", topPatterns[0], "pattern-a should be most frequent")

	// pattern-b should be second (appears 2 times)
	assert.Equal(t, "pattern-b", topPatterns[1], "pattern-b should be second most frequent")

	// Third can be pattern-c or pattern-d (both appear 1 time)
	assert.Contains(t, []string{"pattern-c", "pattern-d"}, topPatterns[2],
		"third pattern should be pattern-c or pattern-d")
}
