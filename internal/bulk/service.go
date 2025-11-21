// Package bulk provides bulk operation services for managing multiple ideas.
package bulk

import (
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Service handles bulk operations on ideas
type Service struct {
	repo *database.Repository
}

// NewService creates a new bulk operations service
func NewService(repo *database.Repository) *Service {
	return &Service{repo: repo}
}

// FilterBySearch filters ideas by searching their content, recommendation, or analysis
func (s *Service) FilterBySearch(ideas []*models.Idea, searchTerm string) []*models.Idea {
	searchLower := strings.ToLower(searchTerm)
	filtered := make([]*models.Idea, 0)

	for _, idea := range ideas {
		if strings.Contains(strings.ToLower(idea.Content), searchLower) ||
			strings.Contains(strings.ToLower(idea.Recommendation), searchLower) ||
			strings.Contains(strings.ToLower(idea.AnalysisDetails), searchLower) {
			filtered = append(filtered, idea)
		}
	}

	return filtered
}

// FilterByAge filters ideas created before the given cutoff date
func (s *Service) FilterByAge(ideas []*models.Idea, cutoffDate time.Time) []*models.Idea {
	filtered := make([]*models.Idea, 0)

	for _, idea := range ideas {
		if idea.CreatedAt.Before(cutoffDate) {
			filtered = append(filtered, idea)
		}
	}

	return filtered
}

// UpdateOptions defines the options for bulk update operations
type UpdateOptions struct {
	SetStatus      string
	AddPatterns    []string
	RemovePatterns []string
	AddTags        []string
	RemoveTags     []string
}

// ApplyUpdates applies bulk updates to an idea and returns whether it was modified
func (s *Service) ApplyUpdates(idea *models.Idea, opts UpdateOptions) bool {
	modified := false

	// Apply status change
	if opts.SetStatus != "" && idea.Status != opts.SetStatus {
		idea.Status = opts.SetStatus
		modified = true
	}

	// Add patterns
	if len(opts.AddPatterns) > 0 {
		newPatterns := AddUniqueStrings(idea.Patterns, opts.AddPatterns)
		if len(newPatterns) > len(idea.Patterns) {
			idea.Patterns = newPatterns
			modified = true
		}
	}

	// Remove patterns
	if len(opts.RemovePatterns) > 0 {
		newPatterns := RemoveStrings(idea.Patterns, opts.RemovePatterns)
		if len(newPatterns) < len(idea.Patterns) {
			idea.Patterns = newPatterns
			modified = true
		}
	}

	// Add tags
	if len(opts.AddTags) > 0 {
		newTags := AddUniqueStrings(idea.Tags, opts.AddTags)
		if len(newTags) > len(idea.Tags) {
			idea.Tags = newTags
			modified = true
		}
	}

	// Remove tags
	if len(opts.RemoveTags) > 0 {
		newTags := RemoveStrings(idea.Tags, opts.RemoveTags)
		if len(newTags) < len(idea.Tags) {
			idea.Tags = newTags
			modified = true
		}
	}

	return modified
}

// String/array manipulation utilities

// AddUniqueStrings adds new items to existing slice, avoiding duplicates
func AddUniqueStrings(existing, newItems []string) []string {
	result := make([]string, len(existing))
	copy(result, existing)

	for _, item := range newItems {
		if !Contains(result, item) {
			result = append(result, item)
		}
	}

	return result
}

// RemoveStrings removes specified items from a slice
func RemoveStrings(existing, toRemove []string) []string {
	result := make([]string, 0, len(existing))

	for _, item := range existing {
		if !Contains(toRemove, item) {
			result = append(result, item)
		}
	}

	return result
}

// Contains checks if a slice contains a specific item
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SplitCommaSeparated splits a comma-separated string and trims whitespace
func SplitCommaSeparated(s string) []string {
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
