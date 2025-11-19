package analytics

import (
	"strings"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateReport tests basic report generation
func TestGenerateReport(t *testing.T) {
	baseTime := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	ideas := []*models.Idea{
		{
			ID:         "1",
			Content:    "High scoring idea",
			FinalScore: 9.0,
			CreatedAt:  baseTime,
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "shiny-object-syndrome"},
				},
			},
		},
		{
			ID:         "2",
			Content:    "Medium scoring idea",
			FinalScore: 6.5,
			CreatedAt:  baseTime.AddDate(0, 0, 1),
		},
		{
			ID:         "3",
			Content:    "Low scoring idea",
			FinalScore: 3.0,
			CreatedAt:  baseTime.AddDate(0, 0, 2),
		},
	}

	report := GenerateReport(ideas)

	assert.Equal(t, "Telos Idea Matrix Analytics Report", report.Title)
	assert.Contains(t, report.Summary, "Total Ideas: 3")
	assert.NotEmpty(t, report.Sections, "report should have sections")
	assert.NotZero(t, report.GeneratedAt, "report should have generation timestamp")
}

// TestGenerateReport_EmptyIdeas tests report generation with no ideas
func TestGenerateReport_EmptyIdeas(t *testing.T) {
	ideas := []*models.Idea{}

	report := GenerateReport(ideas)

	assert.Equal(t, "Telos Idea Matrix Analytics Report", report.Title)
	assert.Contains(t, report.Summary, "No ideas found")
	assert.Empty(t, report.Sections, "should have no sections for empty ideas")
}

// TestGenerateScoreDistribution tests the score distribution section
func TestGenerateScoreDistribution(t *testing.T) {
	ideas := []*models.Idea{
		{ID: "1", FinalScore: 9.0},  // high
		{ID: "2", FinalScore: 8.5},  // high
		{ID: "3", FinalScore: 7.0},  // medium
		{ID: "4", FinalScore: 6.0},  // medium
		{ID: "5", FinalScore: 4.0},  // low
		{ID: "6", FinalScore: 3.0},  // low
		{ID: "7", FinalScore: 2.0},  // low
		{ID: "8", FinalScore: 5.5},  // medium
		{ID: "9", FinalScore: 8.0},  // high
		{ID: "10", FinalScore: 7.5}, // medium
	}

	section := generateScoreDistribution(ideas)

	assert.Equal(t, "Score Distribution", section.Title)
	assert.Contains(t, section.Content, "High Score (8.0+):   3 ideas (30%)")
	assert.Contains(t, section.Content, "Medium Score (5-8):  4 ideas (40%)")
	assert.Contains(t, section.Content, "Low Score (<5):      3 ideas (30%)")
	assert.Contains(t, section.Content, "Average Score: 6.0/10.0")
}

// TestGenerateTrendsSection tests the trends section generation
func TestGenerateTrendsSection(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	ideas := []*models.Idea{
		{ID: "1", FinalScore: 7.0, CreatedAt: baseTime},
		{ID: "2", FinalScore: 8.0, CreatedAt: baseTime.AddDate(0, 1, 0)},
		{ID: "3", FinalScore: 9.0, CreatedAt: baseTime.AddDate(0, 2, 0)},
	}

	section := generateTrendsSection(ideas)

	assert.Equal(t, "Trends", section.Title)
	assert.Contains(t, section.Content, "Score Trends by Month")
	assert.Contains(t, section.Content, "2024-01")
	assert.Contains(t, section.Content, "2024-02")
	assert.Contains(t, section.Content, "2024-03")
}

// TestGenerateTrendsSection_EmptyIdeas tests trends with no ideas
func TestGenerateTrendsSection_EmptyIdeas(t *testing.T) {
	ideas := []*models.Idea{}

	section := generateTrendsSection(ideas)

	assert.Equal(t, "Trends", section.Title)
	assert.Contains(t, section.Content, "No trend data available")
}

// TestGeneratePatternSection tests pattern analysis section
func TestGeneratePatternSection(t *testing.T) {
	ideas := []*models.Idea{
		{
			ID: "1",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "shiny-object-syndrome"},
					{Name: "analysis-paralysis"},
				},
			},
		},
		{
			ID: "2",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "shiny-object-syndrome"},
				},
			},
		},
		{
			ID: "3",
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "premature-optimization"},
				},
			},
		},
	}

	section := generatePatternSection(ideas)

	assert.Equal(t, "Pattern Analysis", section.Title)
	assert.Contains(t, section.Content, "Most Common Patterns")
	assert.Contains(t, section.Content, "shiny-object-syndrome")
	assert.Contains(t, section.Content, "2 occurrences")
}

// TestGeneratePatternSection_NoPatterns tests with no patterns
func TestGeneratePatternSection_NoPatterns(t *testing.T) {
	ideas := []*models.Idea{
		{ID: "1", Analysis: nil},
		{ID: "2", Analysis: &models.Analysis{DetectedPatterns: []models.DetectedPattern{}}},
	}

	section := generatePatternSection(ideas)

	assert.Equal(t, "Pattern Analysis", section.Title)
	assert.Contains(t, section.Content, "No patterns detected")
}

// TestGenerateCreationRateSection tests creation rate section
func TestGenerateCreationRateSection(t *testing.T) {
	now := time.Now()

	ideas := []*models.Idea{
		{ID: "1", CreatedAt: now.AddDate(0, 0, -1)},
		{ID: "2", CreatedAt: now.AddDate(0, 0, -3)},
		{ID: "3", CreatedAt: now.AddDate(0, 0, -5)},
		{ID: "4", CreatedAt: now.AddDate(0, 0, -15)},
		{ID: "5", CreatedAt: now.AddDate(0, 0, -25)},
	}

	section := generateCreationRateSection(ideas)

	assert.Equal(t, "Idea Creation Velocity", section.Title)
	assert.Contains(t, section.Content, "Creation Rates")
	assert.Contains(t, section.Content, "Last 7 days")
	assert.Contains(t, section.Content, "Last 30 days")
	assert.Contains(t, section.Content, "Total Ideas Captured: 5")
}

// TestGenerateRecommendationsSection tests recommendations generation
func TestGenerateRecommendationsSection(t *testing.T) {
	t.Run("many high scoring ideas", func(t *testing.T) {
		ideas := []*models.Idea{
			{ID: "1", FinalScore: 9.0},
			{ID: "2", FinalScore: 8.5},
			{ID: "3", FinalScore: 8.2},
			{ID: "4", FinalScore: 7.0},
			{ID: "5", FinalScore: 6.0},
		}

		section := generateRecommendationsSection(ideas)

		assert.Equal(t, "Recommendations", section.Title)
		assert.Contains(t, section.Content, "high-scoring ideas")
	})

	t.Run("many low scoring ideas", func(t *testing.T) {
		ideas := []*models.Idea{
			{ID: "1", FinalScore: 3.0},
			{ID: "2", FinalScore: 2.0},
			{ID: "3", FinalScore: 4.0},
			{ID: "4", FinalScore: 3.5},
			{ID: "5", FinalScore: 6.0},
		}

		section := generateRecommendationsSection(ideas)

		assert.Contains(t, section.Content, "low-scoring ideas")
	})
}

// TestRenderReport tests markdown rendering
func TestRenderReport(t *testing.T) {
	report := Report{
		Title:       "Test Report",
		Summary:     "This is a test summary",
		GeneratedAt: time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC),
		Sections: []ReportSection{
			{
				Title:   "Section 1",
				Content: "Content for section 1",
			},
			{
				Title:   "Section 2",
				Content: "Content for section 2",
			},
		},
	}

	markdown := RenderReport(report)

	assert.Contains(t, markdown, "# Test Report")
	assert.Contains(t, markdown, "**This is a test summary**")
	assert.Contains(t, markdown, "## Section 1")
	assert.Contains(t, markdown, "Content for section 1")
	assert.Contains(t, markdown, "## Section 2")
	assert.Contains(t, markdown, "Content for section 2")
	assert.Contains(t, markdown, "Generated: 2024-03-15")
}

// TestRenderReportPlainText tests plain text rendering
func TestRenderReportPlainText(t *testing.T) {
	report := Report{
		Title:       "Test Report",
		Summary:     "This is a test summary",
		GeneratedAt: time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC),
		Sections: []ReportSection{
			{
				Title:   "Section 1",
				Content: "Content for section 1",
			},
		},
	}

	text := RenderReportPlainText(report)

	assert.Contains(t, text, "TEST REPORT")
	assert.Contains(t, text, "This is a test summary")
	assert.Contains(t, text, "Section 1")
	assert.Contains(t, text, "Content for section 1")
	assert.Contains(t, text, "Generated: 2024-03-15")

	// Check for separator lines
	lines := strings.Split(text, "\n")
	hasSeparators := false
	for _, line := range lines {
		if strings.Contains(line, "---") || strings.Contains(line, "===") {
			hasSeparators = true
			break
		}
	}
	assert.True(t, hasSeparators, "plain text should have separator lines")
}

// TestReportSections_Coverage ensures all section types are generated
func TestReportSections_Coverage(t *testing.T) {
	baseTime := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	ideas := []*models.Idea{
		{
			ID:         "1",
			FinalScore: 9.0,
			CreatedAt:  baseTime,
			Analysis: &models.Analysis{
				DetectedPatterns: []models.DetectedPattern{
					{Name: "test-pattern"},
				},
			},
		},
		{
			ID:         "2",
			FinalScore: 6.0,
			CreatedAt:  baseTime.AddDate(0, 0, 1),
		},
	}

	report := GenerateReport(ideas)

	require.GreaterOrEqual(t, len(report.Sections), 4, "should have at least 4 sections")

	sectionTitles := make([]string, len(report.Sections))
	for i, section := range report.Sections {
		sectionTitles[i] = section.Title
	}

	// Verify all expected sections are present
	assert.Contains(t, sectionTitles, "Score Distribution")
	assert.Contains(t, sectionTitles, "Trends")
	assert.Contains(t, sectionTitles, "Pattern Analysis")
	assert.Contains(t, sectionTitles, "Idea Creation Velocity")
	assert.Contains(t, sectionTitles, "Recommendations")
}
