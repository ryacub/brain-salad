package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
	"github.com/rayyacub/telos-idea-matrix/internal/utils"
	"github.com/spf13/cobra"
)

func newDumpCommand() *cobra.Command {
	var fromClipboard bool
	var toClipboard bool
	var quick bool

	cmd := &cobra.Command{
		Use:   "dump <idea text>",
		Short: "Capture and analyze an idea immediately",
		Long: `Capture a new idea, analyze it against your telos, and save it to the database.
The idea will be scored and analyzed for patterns immediately.

Examples:
  tm dump "Build a SaaS product for developers"
  tm dump "Create an AI agent that automates customer support"
  tm dump --from-clipboard
  tm dump "Quick idea" --to-clipboard
  tm dump --quick "Fast idea capture"`,
		Args: func(cmd *cobra.Command, args []string) error {
			fromClipboard, _ := cmd.Flags().GetBool("from-clipboard")
			if !fromClipboard && len(args) < 1 {
				return fmt.Errorf("provide idea text or use --from-clipboard")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDump(cmd, args, fromClipboard, toClipboard, quick)
		},
	}

	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read idea from clipboard")
	cmd.Flags().BoolVar(&toClipboard, "to-clipboard", false, "Copy result to clipboard")
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, "Quick mode without LLM analysis")

	return cmd
}

func runDump(cmd *cobra.Command, args []string, fromClipboard, toClipboard, quick bool) error {
	var ideaText string

	// Get idea content from clipboard or arguments
	if fromClipboard {
		text, err := utils.PasteFromClipboard()
		if err != nil {
			return fmt.Errorf("read clipboard: %w", err)
		}
		ideaText = strings.TrimSpace(text)
		if ideaText == "" {
			return fmt.Errorf("clipboard is empty")
		}
		infoColor.Printf("📋 Read from clipboard: %s\n", truncateText(ideaText, 50))
	} else {
		ideaText = strings.Join(args, " ")
	}

	// Use quick mode if requested
	if quick {
		return runQuickDump(ideaText, toClipboard)
	}

	// Show progress
	infoColor.Println("📝 Capturing idea...")
	fmt.Println()

	// Calculate score
	analysis, err := ctx.Engine.CalculateScore(ideaText)
	if err != nil {
		return fmt.Errorf("failed to score idea: %w", err)
	}

	// Detect patterns
	detectedPatterns := ctx.Detector.DetectPatterns(ideaText)

	// Create idea
	idea := models.NewIdea(ideaText)
	idea.RawScore = analysis.RawScore
	idea.FinalScore = analysis.FinalScore
	idea.Recommendation = analysis.GetRecommendation()

	// Convert detected patterns to strings
	patternStrings := make([]string, len(detectedPatterns))
	for i, p := range detectedPatterns {
		patternStrings[i] = fmt.Sprintf("%s: %s", p.Name, p.Description)
	}
	idea.Patterns = patternStrings

	// Serialize analysis details
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("failed to serialize analysis: %w", err)
	}
	idea.AnalysisDetails = string(analysisJSON)

	// Save to database
	if err := ctx.Repository.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
	}

	// Display results
	displayIdeaAnalysis(idea, analysis)

	// Copy result to clipboard if requested
	if toClipboard {
		summary := fmt.Sprintf("Score: %.1f/10.0\n%s\n\nIdea: %s",
			idea.FinalScore,
			idea.Recommendation,
			idea.Content)

		if err := utils.CopyToClipboard(summary); err != nil {
			warningColor.Printf("⚠️  Warning: failed to copy to clipboard: %v\n", err)
		} else {
			successColor.Println("✓ Result copied to clipboard")
		}
	}

	return nil
}

func displayIdeaAnalysis(idea *models.Idea, analysis *models.Analysis) {
	// Header
	fmt.Println(strings.Repeat("─", 80))
	successColor.Printf("✨ Idea Analyzed (ID: %s)\n", idea.ID[:8])
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Content
	fmt.Printf("💡 %s\n\n", idea.Content)

	// Score with color coding
	scoreColor := getScoreColor(idea.FinalScore)
	scoreColor.Printf("⭐ Score: %.1f/10.0\n", idea.FinalScore)

	// Recommendation with emoji
	recommendationColor := getRecommendationColor(idea.Recommendation)
	recommendationColor.Printf("%s\n\n", idea.Recommendation)

	// Mission Alignment breakdown
	fmt.Println("📊 Mission Alignment (40%):")
	fmt.Printf("  • Domain Expertise:   %.2f/1.20\n", analysis.Mission.DomainExpertise)
	fmt.Printf("  • AI Alignment:       %.2f/1.50\n", analysis.Mission.AIAlignment)
	fmt.Printf("  • Execution Support:  %.2f/0.80\n", analysis.Mission.ExecutionSupport)
	fmt.Printf("  • Revenue Potential:  %.2f/0.50\n", analysis.Mission.RevenuePotential)
	fmt.Printf("  Total: %.2f/4.00\n\n", analysis.Mission.Total)

	// Anti-Challenge Scores breakdown
	fmt.Println("🎯 Anti-Challenge Scores (35%):")
	fmt.Printf("  • Context Switching:  %.2f/1.20\n", analysis.AntiChallenge.ContextSwitching)
	fmt.Printf("  • Rapid Prototyping:  %.2f/1.00\n", analysis.AntiChallenge.RapidPrototyping)
	fmt.Printf("  • Accountability:     %.2f/0.80\n", analysis.AntiChallenge.Accountability)
	fmt.Printf("  • Income Anxiety:     %.2f/0.50\n", analysis.AntiChallenge.IncomeAnxiety)
	fmt.Printf("  Total: %.2f/3.50\n\n", analysis.AntiChallenge.Total)

	// Strategic Fit breakdown
	fmt.Println("🚀 Strategic Fit (25%):")
	fmt.Printf("  • Stack Compatibility: %.2f/1.00\n", analysis.Strategic.StackCompatibility)
	fmt.Printf("  • Shipping Habit:      %.2f/0.80\n", analysis.Strategic.ShippingHabit)
	fmt.Printf("  • Public Accountability: %.2f/0.40\n", analysis.Strategic.PublicAccountability)
	fmt.Printf("  • Revenue Testing:     %.2f/0.30\n", analysis.Strategic.RevenueTesting)
	fmt.Printf("  Total: %.2f/2.50\n\n", analysis.Strategic.Total)

	// Patterns detected
	if len(idea.Patterns) > 0 {
		warningColor.Println("⚠️  Patterns Detected:")
		for _, pattern := range idea.Patterns {
			fmt.Printf("  • %s\n", pattern)
		}
		fmt.Println()
	}

	// Footer
	fmt.Println(strings.Repeat("─", 80))
	successColor.Println("✅ Idea saved to database")
	fmt.Println(strings.Repeat("─", 80))
}

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

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// runQuickDump performs fast rule-based analysis without LLM
func runQuickDump(content string, toClipboard bool) error {
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
	if err := ctx.Repository.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
	}

	elapsed := time.Since(start)

	// Step 7: Display results
	fmt.Println(strings.Repeat("─", 80))
	successColor.Printf("✨ Quick Analysis Complete (ID: %s)\n", idea.ID[:8])
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
	fmt.Printf("💡 %s\n\n", idea.Content)

	scoreColor := getScoreColor(idea.FinalScore)
	scoreColor.Printf("⭐ Score: %.1f/10.0 (rule-based)\n", score)

	recommendationColor := getRecommendationColor(recommendation)
	recommendationColor.Printf("%s\n\n", recommendation)

	if len(patterns) > 0 {
		fmt.Println("🏷️  Patterns:")
		for _, pattern := range patterns {
			fmt.Printf("  • %s\n", pattern)
		}
		fmt.Println()
	}

	fmt.Printf("⚡ Completed in %v\n\n", elapsed)

	fmt.Println(strings.Repeat("─", 80))
	successColor.Println("✅ Idea saved to database")
	infoColor.Println("💡 Tip: Use 'tm analyze " + idea.ID[:8] + "' to run full LLM analysis later")
	fmt.Println(strings.Repeat("─", 80))

	// Copy result to clipboard if requested
	if toClipboard {
		summary := fmt.Sprintf("Score: %.1f/10.0 (rule-based)\n%s\n\nIdea: %s",
			idea.FinalScore,
			idea.Recommendation,
			idea.Content)

		if err := utils.CopyToClipboard(summary); err != nil {
			warningColor.Printf("⚠️  Warning: failed to copy to clipboard: %v\n", err)
		} else {
			successColor.Println("✓ Result copied to clipboard")
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
func detectBasicPatterns(content, telos string) []string {
	patterns := []string{}
	contentLower := strings.ToLower(content)

	// Pattern keywords with their categories
	patternKeywords := map[string]string{
		"innovation":   "innovation",
		"innovate":     "innovation",
		"novel":        "innovation",
		"new":          "innovation",
		"sustain":      "sustainability",
		"green":        "sustainability",
		"environment":  "sustainability",
		"eco":          "sustainability",
		"impact":       "impact",
		"improve":      "impact",
		"benefit":      "impact",
		"help":         "impact",
		"scale":        "scalability",
		"grow":         "scalability",
		"expand":       "scalability",
		"revenue":      "revenue",
		"profit":       "revenue",
		"monetize":     "revenue",
		"income":       "revenue",
		"cost":         "cost-reduction",
		"save":         "cost-reduction",
		"efficient":    "efficiency",
		"optimize":     "efficiency",
		"automate":     "automation",
		"automatic":    "automation",
		"ai":           "ai-ml",
		"machine":      "ai-ml",
		"learning":     "ai-ml",
		"mobile":       "mobile",
		"app":          "mobile",
		"web":          "web",
		"cloud":        "cloud",
		"saas":         "saas",
		"product":      "product",
		"service":      "service",
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
