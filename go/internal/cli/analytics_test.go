package cli

import (
	"math"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

func TestGroupIdeasByTime_Day(t *testing.T) {
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), FinalScore: 7.0},
		{CreatedAt: time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC), FinalScore: 8.0},
		{CreatedAt: time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC), FinalScore: 6.0},
	}

	groups := groupIdeasByTime(ideas, "day")

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Verify first group
	if groups[0].Label != "2025-01-01" {
		t.Errorf("Expected first group label to be 2025-01-01, got %s", groups[0].Label)
	}
	if len(groups[0].Ideas) != 2 {
		t.Errorf("Expected first group to have 2 ideas, got %d", len(groups[0].Ideas))
	}

	// Verify second group
	if groups[1].Label != "2025-01-02" {
		t.Errorf("Expected second group label to be 2025-01-02, got %s", groups[1].Label)
	}
	if len(groups[1].Ideas) != 1 {
		t.Errorf("Expected second group to have 1 idea, got %d", len(groups[1].Ideas))
	}
}

func TestGroupIdeasByTime_Week(t *testing.T) {
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC), FinalScore: 7.0},  // Week 02
		{CreatedAt: time.Date(2025, 1, 7, 10, 0, 0, 0, time.UTC), FinalScore: 8.0},  // Week 02
		{CreatedAt: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC), FinalScore: 6.0}, // Week 03
	}

	groups := groupIdeasByTime(ideas, "week")

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Verify week grouping
	if len(groups[0].Ideas) != 2 {
		t.Errorf("Expected first group to have 2 ideas, got %d", len(groups[0].Ideas))
	}
	if len(groups[1].Ideas) != 1 {
		t.Errorf("Expected second group to have 1 idea, got %d", len(groups[1].Ideas))
	}
}

func TestGroupIdeasByTime_Month(t *testing.T) {
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), FinalScore: 7.0},
		{CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), FinalScore: 8.0},
		{CreatedAt: time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC), FinalScore: 6.0},
	}

	groups := groupIdeasByTime(ideas, "month")

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Verify first group (January)
	if groups[0].Label != "2025-01" {
		t.Errorf("Expected first group label to be 2025-01, got %s", groups[0].Label)
	}
	if len(groups[0].Ideas) != 2 {
		t.Errorf("Expected first group to have 2 ideas, got %d", len(groups[0].Ideas))
	}

	// Verify second group (February)
	if groups[1].Label != "2025-02" {
		t.Errorf("Expected second group label to be 2025-02, got %s", groups[1].Label)
	}
	if len(groups[1].Ideas) != 1 {
		t.Errorf("Expected second group to have 1 idea, got %d", len(groups[1].Ideas))
	}
}

func TestGroupIdeasByTime_Sorted(t *testing.T) {
	// Create ideas in random order
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), FinalScore: 7.0},
		{CreatedAt: time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC), FinalScore: 8.0},
		{CreatedAt: time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC), FinalScore: 6.0},
	}

	groups := groupIdeasByTime(ideas, "day")

	// Verify groups are sorted by date
	if groups[0].Label >= groups[1].Label {
		t.Errorf("Groups should be sorted by date, but %s >= %s", groups[0].Label, groups[1].Label)
	}
	if groups[1].Label >= groups[2].Label {
		t.Errorf("Groups should be sorted by date, but %s >= %s", groups[1].Label, groups[2].Label)
	}
}

func TestCalculateAverageScore(t *testing.T) {
	tests := []struct {
		name     string
		scores   []float64
		expected float64
	}{
		{"Single score", []float64{5.0}, 5.0},
		{"Multiple scores", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 3.0},
		{"All same", []float64{7.0, 7.0, 7.0}, 7.0},
		{"Empty", []float64{}, 0.0},
		{"Decimals", []float64{7.5, 8.0, 6.5}, 7.333333333333333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ideas := make([]*models.Idea, len(tt.scores))
			for i, score := range tt.scores {
				ideas[i] = &models.Idea{FinalScore: score}
			}

			result := calculateAverageScore(ideas)
			if math.Abs(result-tt.expected) > 0.000001 {
				t.Errorf("calculateAverageScore(%v) = %f, want %f", tt.scores, result, tt.expected)
			}
		})
	}
}

func TestFindHighestScore(t *testing.T) {
	tests := []struct {
		name     string
		scores   []float64
		expected float64
	}{
		{"Single score", []float64{5.0}, 5.0},
		{"Multiple scores", []float64{1.0, 9.0, 3.0, 4.0, 5.0}, 9.0},
		{"All same", []float64{7.0, 7.0, 7.0}, 7.0},
		{"Empty", []float64{}, 0.0},
		{"First is highest", []float64{9.0, 5.0, 3.0}, 9.0},
		{"Last is highest", []float64{3.0, 5.0, 9.0}, 9.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ideas := make([]*models.Idea, len(tt.scores))
			for i, score := range tt.scores {
				ideas[i] = &models.Idea{FinalScore: score}
			}

			result := findHighestScore(ideas)
			if result != tt.expected {
				t.Errorf("findHighestScore(%v) = %f, want %f", tt.scores, result, tt.expected)
			}
		})
	}
}

func TestFindLowestScore(t *testing.T) {
	tests := []struct {
		name     string
		scores   []float64
		expected float64
	}{
		{"Single score", []float64{5.0}, 5.0},
		{"Multiple scores", []float64{9.0, 1.0, 3.0, 4.0, 5.0}, 1.0},
		{"All same", []float64{7.0, 7.0, 7.0}, 7.0},
		{"Empty", []float64{}, 0.0},
		{"First is lowest", []float64{1.0, 5.0, 9.0}, 1.0},
		{"Last is lowest", []float64{9.0, 5.0, 1.0}, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ideas := make([]*models.Idea, len(tt.scores))
			for i, score := range tt.scores {
				ideas[i] = &models.Idea{FinalScore: score}
			}

			result := findLowestScore(ideas)
			if result != tt.expected {
				t.Errorf("findLowestScore(%v) = %f, want %f", tt.scores, result, tt.expected)
			}
		})
	}
}

func TestCalculateMedianScore(t *testing.T) {
	tests := []struct {
		name     string
		scores   []float64
		expected float64
	}{
		{"Odd number of scores", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 3.0},
		{"Even number of scores", []float64{1.0, 2.0, 3.0, 4.0}, 2.5},
		{"Single score", []float64{5.0}, 5.0},
		{"Two scores", []float64{5.0, 7.0}, 6.0},
		{"Empty", []float64{}, 0.0},
		{"Unsorted input", []float64{5.0, 1.0, 3.0, 2.0, 4.0}, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ideas := make([]*models.Idea, len(tt.scores))
			for i, score := range tt.scores {
				ideas[i] = &models.Idea{FinalScore: score}
			}

			result := calculateMedianScore(ideas)
			if result != tt.expected {
				t.Errorf("calculateMedianScore(%v) = %f, want %f", tt.scores, result, tt.expected)
			}
		})
	}
}

func TestCalculateStdDev(t *testing.T) {
	tests := []struct {
		name      string
		scores    []float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Classic example",
			scores:    []float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0},
			expected:  2.0,
			tolerance: 0.1,
		},
		{
			name:      "All same values",
			scores:    []float64{5.0, 5.0, 5.0, 5.0},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "Two values",
			scores:    []float64{3.0, 7.0},
			expected:  2.0,
			tolerance: 0.001,
		},
		{
			name:      "Single value",
			scores:    []float64{5.0},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "Empty",
			scores:    []float64{},
			expected:  0.0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ideas := make([]*models.Idea, len(tt.scores))
			for i, score := range tt.scores {
				ideas[i] = &models.Idea{FinalScore: score}
			}

			result := calculateStdDev(ideas)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("calculateStdDev(%v) = %f, want ~%f (tolerance: %f)",
					tt.scores, result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"Item exists", []string{"day", "week", "month"}, "week", true},
		{"Item does not exist", []string{"day", "week", "month"}, "year", false},
		{"Empty slice", []string{}, "day", false},
		{"First item", []string{"day", "week", "month"}, "day", true},
		{"Last item", []string{"day", "week", "month"}, "month", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains(%v, %s) = %v, want %v", tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}

func TestPerformanceReport_Statistics(t *testing.T) {
	// Create a set of ideas for comprehensive testing
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), FinalScore: 5.0},
		{CreatedAt: time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC), FinalScore: 7.0},
		{CreatedAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), FinalScore: 6.0},
		{CreatedAt: time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC), FinalScore: 8.0},
		{CreatedAt: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC), FinalScore: 4.0},
	}

	// Test all statistical functions together
	avg := calculateAverageScore(ideas)
	if avg != 6.0 {
		t.Errorf("Average should be 6.0, got %f", avg)
	}

	highest := findHighestScore(ideas)
	if highest != 8.0 {
		t.Errorf("Highest should be 8.0, got %f", highest)
	}

	lowest := findLowestScore(ideas)
	if lowest != 4.0 {
		t.Errorf("Lowest should be 4.0, got %f", lowest)
	}

	median := calculateMedianScore(ideas)
	if median != 6.0 {
		t.Errorf("Median should be 6.0, got %f", median)
	}

	stdDev := calculateStdDev(ideas)
	// Standard deviation of [4, 5, 6, 7, 8] is approximately 1.414
	if math.Abs(stdDev-1.414) > 0.1 {
		t.Errorf("StdDev should be ~1.414, got %f", stdDev)
	}
}

func TestGroupIdeasByTime_EmptyInput(t *testing.T) {
	ideas := []*models.Idea{}

	for _, groupBy := range []string{"day", "week", "month"} {
		t.Run(groupBy, func(t *testing.T) {
			groups := groupIdeasByTime(ideas, groupBy)
			if len(groups) != 0 {
				t.Errorf("Expected 0 groups for empty input, got %d", len(groups))
			}
		})
	}
}

func TestGroupIdeasByTime_SingleIdea(t *testing.T) {
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), FinalScore: 7.0},
	}

	for _, groupBy := range []string{"day", "week", "month"} {
		t.Run(groupBy, func(t *testing.T) {
			groups := groupIdeasByTime(ideas, groupBy)
			if len(groups) != 1 {
				t.Errorf("Expected 1 group for single idea, got %d", len(groups))
			}
			if len(groups[0].Ideas) != 1 {
				t.Errorf("Expected group to contain 1 idea, got %d", len(groups[0].Ideas))
			}
		})
	}
}

func TestGroupIdeasByTime_LeapYear(t *testing.T) {
	// Test with leap year (2024 is a leap year)
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2024, 2, 28, 10, 0, 0, 0, time.UTC), FinalScore: 7.0},
		{CreatedAt: time.Date(2024, 2, 29, 10, 0, 0, 0, time.UTC), FinalScore: 8.0},
		{CreatedAt: time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC), FinalScore: 6.0},
	}

	groups := groupIdeasByTime(ideas, "day")

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups (Feb 28, Feb 29, Mar 1), got %d", len(groups))
	}

	// Verify Feb 29 is correctly grouped
	hasLeapDay := false
	for _, group := range groups {
		if group.Label == "2024-02-29" {
			hasLeapDay = true
			if len(group.Ideas) != 1 {
				t.Errorf("Expected leap day group to have 1 idea, got %d", len(group.Ideas))
			}
		}
	}
	if !hasLeapDay {
		t.Error("Leap day (Feb 29) was not found in groups")
	}
}

func TestGroupIdeasByTime_YearBoundary(t *testing.T) {
	ideas := []*models.Idea{
		{CreatedAt: time.Date(2024, 12, 31, 23, 0, 0, 0, time.UTC), FinalScore: 7.0},
		{CreatedAt: time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC), FinalScore: 8.0},
	}

	groups := groupIdeasByTime(ideas, "day")

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups across year boundary, got %d", len(groups))
	}

	if groups[0].Label != "2024-12-31" {
		t.Errorf("Expected first group to be 2024-12-31, got %s", groups[0].Label)
	}
	if groups[1].Label != "2025-01-01" {
		t.Errorf("Expected second group to be 2025-01-01, got %s", groups[1].Label)
	}
}
