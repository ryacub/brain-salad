// Package analytics provides comprehensive analytics reporting for idea analysis and trends.
package analytics

import (
	"fmt"
	"strings"
	"time"

	"github.com/ryacub/telos-idea-matrix/internal/models"
)

// Report represents a comprehensive analytics report
type Report struct {
	Title       string
	Summary     string
	Sections    []ReportSection
	GeneratedAt time.Time
}

// ReportSection represents a section in the report
type ReportSection struct {
	Title   string
	Content string
}

// GenerateReport creates a comprehensive analytics report from a set of ideas
func GenerateReport(ideas []*models.Idea) Report {
	report := Report{
		Title:       "Telos Idea Matrix Analytics Report",
		GeneratedAt: time.Now(),
		Sections:    make([]ReportSection, 0),
	}

	if len(ideas) == 0 {
		report.Summary = "No ideas found. Start capturing ideas to see analytics!"
		return report
	}

	// Build summary
	report.Summary = fmt.Sprintf(
		"Total Ideas: %d\nReport Generated: %s",
		len(ideas),
		time.Now().Format("2006-01-02 15:04"),
	)

	// Add score distribution section
	report.Sections = append(report.Sections, generateScoreDistribution(ideas))

	// Add trends section
	report.Sections = append(report.Sections, generateTrendsSection(ideas))

	// Add pattern analysis section
	report.Sections = append(report.Sections, generatePatternSection(ideas))

	// Add creation rate section
	report.Sections = append(report.Sections, generateCreationRateSection(ideas))

	// Add recommendations section
	report.Sections = append(report.Sections, generateRecommendationsSection(ideas))

	return report
}

// generateScoreDistribution creates the score distribution section
func generateScoreDistribution(ideas []*models.Idea) ReportSection {
	high := 0   // >= 8.0
	medium := 0 // 5.0 - 8.0
	low := 0    // < 5.0

	totalScore := 0.0
	for _, idea := range ideas {
		totalScore += idea.FinalScore

		if idea.FinalScore >= 8.0 {
			high++
		} else if idea.FinalScore >= 5.0 {
			medium++
		} else {
			low++
		}
	}

	avgScore := totalScore / float64(len(ideas))

	content := fmt.Sprintf(`
Average Score: %.1f/10.0

Distribution:
  🔥 High Score (8.0+):   %d ideas (%d%%)
  ✅ Medium Score (5-8):  %d ideas (%d%%)
  🚫 Low Score (<5):      %d ideas (%d%%)
`,
		avgScore,
		high, (high*100)/len(ideas),
		medium, (medium*100)/len(ideas),
		low, (low*100)/len(ideas),
	)

	return ReportSection{
		Title:   "Score Distribution",
		Content: strings.TrimSpace(content),
	}
}

// generateTrendsSection creates the trends analysis section
func generateTrendsSection(ideas []*models.Idea) ReportSection {
	trends := CalculateScoreTrends(ideas, "month")

	if len(trends) == 0 {
		return ReportSection{
			Title:   "Trends",
			Content: "No trend data available yet.",
		}
	}

	var content strings.Builder
	content.WriteString("Score Trends by Month:\n\n")

	for _, trend := range trends {
		content.WriteString(fmt.Sprintf(
			"  %s: %.1f avg (%d ideas)\n",
			trend.Period,
			trend.AvgScore,
			trend.IdeaCount,
		))
	}

	// Add trend direction
	direction := CalculateTrendDirection(trends)
	var directionMsg string
	switch direction {
	case "up":
		directionMsg = "📈 Your idea quality is improving over time!"
	case "down":
		directionMsg = "📉 Consider refining your idea capture process."
	default:
		directionMsg = "➡️  Your idea quality is stable."
	}
	content.WriteString(fmt.Sprintf("\nTrend Direction: %s\n", directionMsg))

	return ReportSection{
		Title:   "Trends",
		Content: content.String(),
	}
}

// generatePatternSection creates the pattern analysis section
func generatePatternSection(ideas []*models.Idea) ReportSection {
	freq := CalculatePatternFrequency(ideas)

	if len(freq) == 0 {
		return ReportSection{
			Title:   "Pattern Analysis",
			Content: "No patterns detected yet.",
		}
	}

	topPatterns := GetTopPatterns(ideas, 5)

	var content strings.Builder
	content.WriteString("Most Common Patterns:\n\n")

	for i, pattern := range topPatterns {
		count := freq[pattern]
		percentage := (count * 100) / len(ideas)
		content.WriteString(fmt.Sprintf(
			"  %d. %s: %d occurrences (%d%% of ideas)\n",
			i+1,
			pattern,
			count,
			percentage,
		))
	}

	return ReportSection{
		Title:   "Pattern Analysis",
		Content: content.String(),
	}
}

// generateCreationRateSection creates the idea creation rate section
func generateCreationRateSection(ideas []*models.Idea) ReportSection {
	rate7 := CalculateCreationRate(ideas, 7)
	rate30 := CalculateCreationRate(ideas, 30)

	content := fmt.Sprintf(`
Creation Rates:
  Last 7 days:  %.1f ideas/day
  Last 30 days: %.1f ideas/day

Total Ideas Captured: %d
`,
		rate7,
		rate30,
		len(ideas),
	)

	return ReportSection{
		Title:   "Idea Creation Velocity",
		Content: strings.TrimSpace(content),
	}
}

// generateRecommendationsSection creates recommendations based on analysis
func generateRecommendationsSection(ideas []*models.Idea) ReportSection {
	var recommendations []string

	// Analyze score distribution
	high := 0
	low := 0
	for _, idea := range ideas {
		if idea.FinalScore >= 8.0 {
			high++
		} else if idea.FinalScore < 5.0 {
			low++
		}
	}

	highPercent := (high * 100) / len(ideas)
	lowPercent := (low * 100) / len(ideas)

	if highPercent > 30 {
		recommendations = append(recommendations,
			"🌟 Excellent! You have many high-scoring ideas. Prioritize and execute on them!")
	}

	if lowPercent > 50 {
		recommendations = append(recommendations,
			"⚠️  Many low-scoring ideas detected. Review alignment with your telos and mission.")
	}

	// Analyze creation rate
	rate7 := CalculateCreationRate(ideas, 7)
	if rate7 < 0.5 {
		recommendations = append(recommendations,
			"💡 Consider increasing your idea capture frequency for better insights.")
	} else if rate7 > 5.0 {
		recommendations = append(recommendations,
			"🔥 High idea velocity! Make sure to review and prioritize regularly.")
	}

	// Analyze patterns
	freq := CalculatePatternFrequency(ideas)
	for pattern, count := range freq {
		percentage := (count * 100) / len(ideas)
		if percentage > 40 {
			recommendations = append(recommendations,
				fmt.Sprintf("⚠️  Pattern '%s' appears in %d%% of ideas - watch for recurring anti-patterns!", pattern, percentage))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"✅ Your idea generation process looks healthy. Keep it up!")
	}

	var content strings.Builder
	for i, rec := range recommendations {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}

	return ReportSection{
		Title:   "Recommendations",
		Content: content.String(),
	}
}

// RenderReport converts a report to markdown format
func RenderReport(report Report) string {
	var sb strings.Builder

	// Title and header
	sb.WriteString(fmt.Sprintf("# %s\n\n", report.Title))

	// Summary
	sb.WriteString(fmt.Sprintf("**%s**\n\n", report.Summary))
	sb.WriteString("---\n\n")

	// Sections
	for _, section := range report.Sections {
		sb.WriteString(fmt.Sprintf("## %s\n\n", section.Title))
		sb.WriteString(section.Content)
		sb.WriteString("\n\n---\n\n")
	}

	// Footer
	sb.WriteString(fmt.Sprintf("*Generated: %s*\n",
		report.GeneratedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}

// RenderReportPlainText converts a report to plain text format
func RenderReportPlainText(report Report) string {
	var sb strings.Builder

	// Title
	sb.WriteString(strings.ToUpper(report.Title))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("=", len(report.Title)))
	sb.WriteString("\n\n")

	// Summary
	sb.WriteString(report.Summary)
	sb.WriteString("\n\n")
	sb.WriteString(strings.Repeat("-", 60))
	sb.WriteString("\n\n")

	// Sections
	for _, section := range report.Sections {
		sb.WriteString(section.Title)
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("-", len(section.Title)))
		sb.WriteString("\n\n")
		sb.WriteString(section.Content)
		sb.WriteString("\n\n")
		sb.WriteString(strings.Repeat("-", 60))
		sb.WriteString("\n\n")
	}

	// Footer
	sb.WriteString(fmt.Sprintf("Generated: %s\n",
		report.GeneratedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}
