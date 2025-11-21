package bulk

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rs/zerolog/log"
)

var (
	// Color definitions for bulk commands
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	infoColor    = color.New(color.FgCyan)
	warningColor = color.New(color.FgYellow)
)

// truncate truncates a string to the specified maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// confirm prompts the user for yes/no confirmation
func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		log.Warn().Err(err).Msg("failed to read user input")
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// parseDuration parses duration strings like "7d", "30d", "24h"
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	value := s[:len(s)-1]
	unit := s[len(s)-1:]

	var multiplier time.Duration
	switch unit {
	case "d":
		multiplier = 24 * time.Hour
	case "h":
		multiplier = time.Hour
	case "m":
		multiplier = time.Minute
	case "s":
		multiplier = time.Second
	default:
		// Fallback to standard Go duration parsing
		return time.ParseDuration(s)
	}

	var numValue int
	n, err := fmt.Sscanf(value, "%d", &numValue)
	if err != nil || n != 1 {
		return 0, fmt.Errorf("invalid duration value: %w", err)
	}

	return time.Duration(numValue) * multiplier, nil
}

// createLLMManager creates and configures an LLM manager
func createLLMManager() *llm.Manager {
	return llm.NewManager(nil)
}

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
