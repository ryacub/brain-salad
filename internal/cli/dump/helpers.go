package dump

import (
	"fmt"
	"strings"
)

// getScoreIndicator returns a visual bar for the score
func getScoreIndicator(score float64) string {
	bars := int(score)
	if bars > 10 {
		bars = 10
	}
	if bars < 0 {
		bars = 0
	}
	filled := strings.Repeat("█", bars)
	empty := strings.Repeat("░", 10-bars)
	return fmt.Sprintf("[%s%s]", filled, empty)
}

// getRecommendationIndicator returns an indicator for the recommendation
func getRecommendationIndicator(rec string) string {
	recUpper := strings.ToUpper(rec)
	if strings.Contains(recUpper, "PURSUE") || strings.Contains(recUpper, "STRONG") {
		return "✓ (Go for it!)"
	}
	if strings.Contains(recUpper, "CONSIDER") || strings.Contains(recUpper, "MODERATE") {
		return "⏸ (Consider carefully)"
	}
	if strings.Contains(recUpper, "AVOID") || strings.Contains(recUpper, "WEAK") || strings.Contains(recUpper, "DEFER") {
		return "✗ (Skip this)"
	}
	return "?"
}

// wrapTextSimple wraps text to specified width
func wrapTextSimple(text string, width int) string {
	if len(text) <= width {
		return "  " + text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	result.WriteString("  ") // Initial indent
	lineLen = 2

	for i, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > width && lineLen > 2 {
			result.WriteString("\n  ")
			lineLen = 2
		} else if i > 0 {
			result.WriteString(" ")
			lineLen++
		}

		result.WriteString(word)
		lineLen += wordLen
	}

	return result.String()
}

// formatCategoryTitle formats a category name for display
func formatCategoryTitle(category string) string {
	// Convert snake_case or camelCase to Title Case
	words := strings.FieldsFunc(category, func(r rune) bool {
		return r == '_' || r == '-'
	})

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}
