package bulk

import (
	"fmt"
	"strings"
	"time"

	"github.com/ryacub/telos-idea-matrix/internal/llm"
	"github.com/ryacub/telos-idea-matrix/internal/models"
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

// filterBySearch filters ideas by searching their content, recommendation, or analysis
func filterBySearch(ideas []*models.Idea, searchTerm string) []*models.Idea {
	searchLower := strings.ToLower(searchTerm)
	filtered := make([]*models.Idea, 0, len(ideas)/4)

	for _, idea := range ideas {
		contentLower := strings.ToLower(idea.Content)
		recommendationLower := strings.ToLower(idea.Recommendation)
		analysisLower := strings.ToLower(idea.AnalysisDetails)

		if strings.Contains(contentLower, searchLower) ||
			strings.Contains(recommendationLower, searchLower) ||
			strings.Contains(analysisLower, searchLower) {
			filtered = append(filtered, idea)
		}
	}

	return filtered
}

// filterByAge filters ideas created before the given cutoff date
func filterByAge(ideas []*models.Idea, cutoffDate time.Time) []*models.Idea {
	filtered := make([]*models.Idea, 0, len(ideas)/2)

	for _, idea := range ideas {
		if idea.CreatedAt.Before(cutoffDate) {
			filtered = append(filtered, idea)
		}
	}

	return filtered
}

// updateOptions defines the options for bulk update operations
type updateOptions struct {
	SetStatus      string
	AddPatterns    []string
	RemovePatterns []string
	AddTags        []string
	RemoveTags     []string
}

// applyUpdates applies bulk updates to an idea and returns whether it was modified
func applyUpdates(idea *models.Idea, opts updateOptions) bool {
	modified := false

	// Apply status change
	if opts.SetStatus != "" && idea.Status != opts.SetStatus {
		idea.Status = opts.SetStatus
		modified = true
	}

	// Add patterns
	if len(opts.AddPatterns) > 0 {
		newPatterns := addUniqueStrings(idea.Patterns, opts.AddPatterns)
		if len(newPatterns) > len(idea.Patterns) {
			idea.Patterns = newPatterns
			modified = true
		}
	}

	// Remove patterns
	if len(opts.RemovePatterns) > 0 {
		newPatterns := removeStrings(idea.Patterns, opts.RemovePatterns)
		if len(newPatterns) < len(idea.Patterns) {
			idea.Patterns = newPatterns
			modified = true
		}
	}

	// Add tags
	if len(opts.AddTags) > 0 {
		newTags := addUniqueStrings(idea.Tags, opts.AddTags)
		if len(newTags) > len(idea.Tags) {
			idea.Tags = newTags
			modified = true
		}
	}

	// Remove tags
	if len(opts.RemoveTags) > 0 {
		newTags := removeStrings(idea.Tags, opts.RemoveTags)
		if len(newTags) < len(idea.Tags) {
			idea.Tags = newTags
			modified = true
		}
	}

	return modified
}

// addUniqueStrings adds new items to existing slice, avoiding duplicates
func addUniqueStrings(existing, newItems []string) []string {
	seen := make(map[string]bool, len(existing)+len(newItems))
	result := make([]string, 0, len(existing)+len(newItems))

	// Add existing items
	for _, item := range existing {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// Add new items if not duplicates
	for _, item := range newItems {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// removeStrings removes specified items from a slice
func removeStrings(existing, toRemove []string) []string {
	removeMap := make(map[string]bool, len(toRemove))
	for _, item := range toRemove {
		removeMap[item] = true
	}

	result := make([]string, 0, len(existing))
	for _, item := range existing {
		if !removeMap[item] {
			result = append(result, item)
		}
	}

	return result
}

// splitCommaSeparated splits a comma-separated string and trims whitespace
func splitCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// contains checks if a slice contains a specific item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
