package dump

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/ryacub/telos-idea-matrix/internal/scoring"
	"github.com/ryacub/telos-idea-matrix/internal/utils"
)

// Context interface defines the dependencies needed by dump operations
type Context interface {
	GetRepository() interface{ Create(*models.Idea) error }
}

// runQuickDump performs fast rule-based analysis without LLM
func runQuickDump(content string, toClipboard bool, repo interface{ Create(*models.Idea) error }) error {
	start := time.Now()

	// Step 1: Load telos (optional for rule-based scoring)
	telosContent, err := loadTelosQuiet()
	if err != nil {
		// Continue without telos if not available
		telosContent = ""
	}

	// Step 2: Rule-based scoring
	scorer := scoring.NewRuleBasedScorer()
	score := scorer.Score(content, telosContent)

	// Step 3: Pattern detection
	patterns := detectBasicPatterns(content, telosContent)

	// Step 4: Generate recommendation
	recommendation := generateRecommendation(score, patterns)

	// Step 5: Create idea with quick analysis marker
	reasoning := "Quick analysis (rule-based, no LLM)\nScore based on: keyword matching, length, telos alignment"

	idea := models.NewIdea(content)
	idea.FinalScore = score
	idea.RawScore = score
	idea.Patterns = patterns
	idea.Recommendation = recommendation
	idea.AnalysisDetails = reasoning

	// Step 6: Save to database
	if err := repo.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
	}

	elapsed := time.Since(start)

	// Step 7: Display results
	fmt.Println(strings.Repeat("─", 80))
	successColor := cliutil.GetScoreColor(10.0) // Green color for success
	if _, err := successColor.Printf("✨ Quick Analysis Complete (ID: %s)\n", idea.ID[:8]); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
	fmt.Printf("💡 %s\n\n", idea.Content)

	scoreColor := cliutil.GetScoreColor(idea.FinalScore)
	if _, err := scoreColor.Printf("⭐ Score: %.1f/10.0 (rule-based)\n", score); err != nil {
		log.Warn().Err(err).Msg("failed to print score")
	}

	recommendationColor := cliutil.GetRecommendationColor(recommendation)
	if _, err := recommendationColor.Printf("%s\n\n", recommendation); err != nil {
		log.Warn().Err(err).Msg("failed to print recommendation")
	}

	if len(patterns) > 0 {
		fmt.Println("🏷️  Patterns:")
		for _, pattern := range patterns {
			fmt.Printf("  • %s\n", pattern)
		}
		fmt.Println()
	}

	fmt.Printf("⚡ Completed in %v\n\n", elapsed)

	fmt.Println(strings.Repeat("─", 80))
	if _, err := successColor.Println("✅ Idea saved to database"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	infoColor := cliutil.GetScoreColor(7.0) // Cyan-like color for info
	if _, err := infoColor.Println("💡 Tip: Use 'tm analyze " + idea.ID[:8] + "' to run full LLM analysis later"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))

	// Copy result to clipboard if requested
	if toClipboard {
		summary := fmt.Sprintf("Score: %.1f/10.0 (rule-based)\n%s\n\nIdea: %s",
			idea.FinalScore,
			idea.Recommendation,
			idea.Content)

		if err := utils.CopyToClipboard(summary); err != nil {
			warningColor := cliutil.GetScoreColor(5.0) // Yellow color for warnings
			if _, printErr := warningColor.Printf("⚠️  Warning: failed to copy to clipboard: %v\n", err); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print warning")
			}
		} else {
			if _, err := successColor.Println("✓ Result copied to clipboard"); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
		}
	}

	return nil
}

// loadTelosQuiet loads telos without errors
func loadTelosQuiet() (string, error) {
	telosPath := os.Getenv("TELOS_PATH")
	if telosPath == "" {
		homeDir, _ := os.UserHomeDir()
		telosPath = filepath.Join(homeDir, ".telos", "telos.md")
	}

	data, err := os.ReadFile(telosPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// detectBasicPatterns performs simple pattern detection
func detectBasicPatterns(content, _ string) []string {
	patterns := []string{}
	contentLower := strings.ToLower(content)

	// Pattern keywords with their categories
	patternKeywords := map[string]string{
		"innovation":  "innovation",
		"innovate":    "innovation",
		"novel":       "innovation",
		"new":         "innovation",
		"sustain":     "sustainability",
		"green":       "sustainability",
		"environment": "sustainability",
		"eco":         "sustainability",
		"impact":      "impact",
		"improve":     "impact",
		"benefit":     "impact",
		"help":        "impact",
		"scale":       "scalability",
		"grow":        "scalability",
		"expand":      "scalability",
		"revenue":     "revenue",
		"profit":      "revenue",
		"monetize":    "revenue",
		"income":      "revenue",
		"cost":        "cost-reduction",
		"save":        "cost-reduction",
		"efficient":   "efficiency",
		"optimize":    "efficiency",
		"automate":    "automation",
		"automatic":   "automation",
		"ai":          "ai-ml",
		"machine":     "ai-ml",
		"learning":    "ai-ml",
		"mobile":      "mobile",
		"app":         "mobile",
		"web":         "web",
		"cloud":       "cloud",
		"saas":        "saas",
		"product":     "product",
		"service":     "service",
	}

	// Track found patterns to avoid duplicates
	found := make(map[string]bool)

	for keyword, pattern := range patternKeywords {
		if strings.Contains(contentLower, keyword) && !found[pattern] {
			patterns = append(patterns, pattern)
			found[pattern] = true
		}
	}

	// If no patterns found, add a generic one
	if len(patterns) == 0 {
		patterns = append(patterns, "general")
	}

	// Limit to top 5 patterns
	if len(patterns) > 5 {
		patterns = patterns[:5]
	}

	return patterns
}

// generateRecommendation creates a recommendation based on score and patterns
func generateRecommendation(score float64, patterns []string) string {
	// High score -> pursue
	if score >= 7.0 {
		return "🔥 PURSUE - Strong potential"
	}

	// Low score -> defer
	if score < 4.0 {
		return "❌ DEFER - Low alignment"
	}

	// Medium score -> review (check patterns for tie-breaker)
	strongPatterns := []string{"innovation", "impact", "scalability", "revenue"}
	for _, pattern := range patterns {
		for _, strong := range strongPatterns {
			if pattern == strong {
				return "✅ PURSUE - Good potential with strong patterns"
			}
		}
	}

	return "⚠️ REVIEW - Needs more evaluation"
}
