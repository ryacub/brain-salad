package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/utils"
	"github.com/spf13/cobra"
)

func newDumpCommand() *cobra.Command {
	var fromClipboard bool
	var toClipboard bool

	cmd := &cobra.Command{
		Use:   "dump <idea text>",
		Short: "Capture and analyze an idea immediately",
		Long: `Capture a new idea, analyze it against your telos, and save it to the database.
The idea will be scored and analyzed for patterns immediately.

Examples:
  tm dump "Build a SaaS product for developers"
  tm dump "Create an AI agent that automates customer support"
  tm dump --from-clipboard
  tm dump "Quick idea" --to-clipboard`,
		Args: func(cmd *cobra.Command, args []string) error {
			fromClipboard, _ := cmd.Flags().GetBool("from-clipboard")
			if !fromClipboard && len(args) < 1 {
				return fmt.Errorf("provide idea text or use --from-clipboard")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDump(cmd, args, fromClipboard, toClipboard)
		},
	}

	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read idea from clipboard")
	cmd.Flags().BoolVar(&toClipboard, "to-clipboard", false, "Copy result to clipboard")

	return cmd
}

func runDump(cmd *cobra.Command, args []string, fromClipboard, toClipboard bool) error {
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
