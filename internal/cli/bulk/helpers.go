package bulk

import (
	"fmt"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
)

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
