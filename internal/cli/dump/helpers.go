package dump

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// getScoreColor returns a color based on the score value
func getScoreColor(score float64) *color.Color {
	switch {
	case score >= 8.5:
		return color.New(color.FgGreen, color.Bold)
	case score >= 7.0:
		return color.New(color.FgGreen)
	case score >= 5.0:
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgRed)
	}
}

// getRecommendationColor returns a color based on the recommendation text
func getRecommendationColor(recommendation string) *color.Color {
	if strings.Contains(recommendation, "🔥") {
		return color.New(color.FgGreen, color.Bold)
	} else if strings.Contains(recommendation, "✅") {
		return color.New(color.FgGreen)
	} else if strings.Contains(recommendation, "⚠️") {
		return color.New(color.FgYellow)
	}
	return color.New(color.FgRed)
}

// truncateText truncates text to specified length with ellipsis
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

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

// confirm prompts the user for yes/no confirmation
func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Silent error handling - just return false
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
