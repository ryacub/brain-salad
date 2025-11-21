package cliutil

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

// Shared color definitions for CLI commands
var (
	SuccessColor = color.New(color.FgGreen, color.Bold)
	ErrorColor   = color.New(color.FgRed, color.Bold)
	InfoColor    = color.New(color.FgCyan)
	WarningColor = color.New(color.FgYellow)
)

// GetScoreColor returns a color based on the score value
func GetScoreColor(score float64) *color.Color {
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

// GetRecommendationColor returns a color based on the recommendation text
func GetRecommendationColor(recommendation string) *color.Color {
	if strings.Contains(recommendation, "🔥") {
		return color.New(color.FgGreen, color.Bold)
	} else if strings.Contains(recommendation, "✅") {
		return color.New(color.FgGreen)
	} else if strings.Contains(recommendation, "⚠️") {
		return color.New(color.FgYellow)
	}
	return color.New(color.FgRed)
}

// TruncateText truncates text to specified length with ellipsis
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// Confirm prompts the user for yes/no confirmation
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		log.Warn().Err(err).Msg("failed to read user input")
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
