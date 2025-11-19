// Package patterns provides anti-pattern and positive pattern detection for idea analysis.
package patterns

import (
	"regexp"
	"strings"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Detector detects anti-patterns and positive patterns in ideas.
type Detector struct {
	telos *models.Telos

	// Compiled regex patterns
	contextSwitchingRegex  *regexp.Regexp
	perfectionismRegex     *regexp.Regexp
	procrastinationRegex   *regexp.Regexp
	accountabilityNegRegex *regexp.Regexp
	accountabilityPosRegex *regexp.Regexp
}

// NewDetector creates a new pattern detector with the given telos configuration.
func NewDetector(telos *models.Telos) *Detector {
	return &Detector{
		telos: telos,
		// Context switching penalty keywords
		contextSwitchingRegex: regexp.MustCompile(`(?i)(rust|javascript|typescript|react|flutter|swift|mobile\s+app|game\s+development)`),
		// Perfectionism keywords
		perfectionismRegex: regexp.MustCompile(`(?i)(comprehensive|complete|production-ready|fully-featured|everything)`),
		// Procrastination pattern (learn before/then)
		procrastinationRegex: regexp.MustCompile(`(?i)(learn.*(before|then|first)|study.*(before|then|first))`),
		// Accountability avoidance (negative)
		accountabilityNegRegex: regexp.MustCompile(`(?i)(just for me|personal\s+project|solo\s+project|only\s+for\s+me|private\s+project)`),
		// Accountability (positive)
		accountabilityPosRegex: regexp.MustCompile(`(?i)(public|share|github|twitter|build\s+in\s+public|customer|client)`),
	}
}

// DetectPatterns analyzes an idea and returns all detected patterns.
func (d *Detector) DetectPatterns(ideaText string) []models.DetectedPattern {
	if ideaText == "" {
		return []models.DetectedPattern{}
	}

	var patterns []models.DetectedPattern
	ideaLower := strings.ToLower(ideaText)

	// Detect context switching
	if pattern := d.detectContextSwitching(ideaLower); pattern != nil {
		patterns = append(patterns, *pattern)
	}

	// Detect perfectionism
	if pattern := d.detectPerfectionism(ideaLower); pattern != nil {
		patterns = append(patterns, *pattern)
	}

	// Detect procrastination
	if pattern := d.detectProcrastination(ideaLower); pattern != nil {
		patterns = append(patterns, *pattern)
	}

	// Detect accountability avoidance
	if pattern := d.detectAccountabilityAvoidance(ideaLower); pattern != nil {
		patterns = append(patterns, *pattern)
	}

	// Detect telos failure patterns
	patterns = append(patterns, d.detectTelosFailurePatterns(ideaLower)...)

	return patterns
}

// detectContextSwitching detects stack switching anti-patterns.
// From RUST_REFERENCE.md:
// - Negative (High): Stack-switching keywords (rust, javascript, react, etc.)
// - Positive: Matches current stack keywords
func (d *Detector) detectContextSwitching(ideaLower string) *models.DetectedPattern {
	// Check for stack-switching keywords (negative pattern)
	if d.contextSwitchingRegex.MatchString(ideaLower) {
		return &models.DetectedPattern{
			Name:        "Context switching",
			Description: "Context-switching risk detected - using different tech stack",
			Confidence:  0.9,
			Severity:    "high",
		}
	}

	// Check if matches current stack (positive pattern)
	if d.telos != nil && len(d.telos.Stack.Primary) > 0 {
		matchCount := 0
		for _, tech := range d.telos.Stack.Primary {
			if strings.Contains(ideaLower, strings.ToLower(tech)) {
				matchCount++
			}
		}

		// If using 2+ primary stack items, it's staying focused
		if matchCount >= 2 {
			return &models.DetectedPattern{
				Name:        "Context switching",
				Description: "Staying focused on current tech stack",
				Confidence:  0.9,
				Severity:    "low", // Low severity = positive pattern
			}
		}
	}

	return nil
}

// detectPerfectionism detects scope creep and perfectionism anti-patterns.
// From RUST_REFERENCE.md:
// - Negative (High): "comprehensive", "complete", "production-ready"
// - Message: "Scope creep risk - over-engineering detected"
func (d *Detector) detectPerfectionism(ideaLower string) *models.DetectedPattern {
	if d.perfectionismRegex.MatchString(ideaLower) {
		return &models.DetectedPattern{
			Name:        "Perfectionism",
			Description: "Scope creep risk - over-engineering detected",
			Confidence:  0.85,
			Severity:    "high",
		}
	}
	return nil
}

// detectProcrastination detects learning-before-building anti-patterns.
// From RUST_REFERENCE.md:
// - Negative (Critical): Pattern "learn" + ("before" OR "then")
// - Message: "Consumption trap - learning before building"
func (d *Detector) detectProcrastination(ideaLower string) *models.DetectedPattern {
	if d.procrastinationRegex.MatchString(ideaLower) {
		return &models.DetectedPattern{
			Name:        "Procrastination",
			Description: "Consumption trap - learning before building instead of building to learn",
			Confidence:  0.95,
			Severity:    "critical",
		}
	}
	return nil
}

// detectAccountabilityAvoidance detects solo/personal project anti-patterns.
// From RUST_REFERENCE.md:
// - Negative (Medium): "just for me", "personal project"
// - Positive: "public", "share", "github"
func (d *Detector) detectAccountabilityAvoidance(ideaLower string) *models.DetectedPattern {
	// Check for positive accountability first
	if d.accountabilityPosRegex.MatchString(ideaLower) {
		return &models.DetectedPattern{
			Name:        "Accountability avoidance",
			Description: "External accountability component detected - building in public or with customers",
			Confidence:  0.8,
			Severity:    "low", // Low severity = positive pattern
		}
	}

	// Check for negative accountability avoidance
	if d.accountabilityNegRegex.MatchString(ideaLower) {
		return &models.DetectedPattern{
			Name:        "Accountability avoidance",
			Description: "Solo-only project - no external accountability detected",
			Confidence:  0.75,
			Severity:    "medium",
		}
	}

	return nil
}

// detectTelosFailurePatterns checks idea against telos failure patterns.
// Matches keywords from telos failure patterns against the idea.
func (d *Detector) detectTelosFailurePatterns(ideaLower string) []models.DetectedPattern {
	if d.telos == nil {
		return nil
	}

	var patterns []models.DetectedPattern

	for _, failurePattern := range d.telos.FailurePatterns {
		// Check if any keywords from the failure pattern match the idea
		matchedKeywords := []string{}
		for _, keyword := range failurePattern.Keywords {
			if strings.Contains(ideaLower, strings.ToLower(keyword)) {
				matchedKeywords = append(matchedKeywords, keyword)
			}
		}

		// If at least 2 keywords match (or 1 keyword for short patterns), flag it
		threshold := 2
		if len(failurePattern.Keywords) <= 3 {
			threshold = 1
		}

		if len(matchedKeywords) >= threshold {
			// Determine severity based on number of matches
			severity := "medium"
			confidence := float64(len(matchedKeywords)) / float64(len(failurePattern.Keywords))
			if confidence > 0.7 {
				severity = "high"
			}

			patterns = append(patterns, models.DetectedPattern{
				Name:        failurePattern.Name,
				Description: failurePattern.Description,
				Confidence:  confidence,
				Severity:    severity,
			})
		}
	}

	return patterns
}
