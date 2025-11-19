package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newScoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "score <idea text>",
		Short: "Score an idea without saving it",
		Long: `Score an idea against your telos without saving it to the database.
Useful for quick idea validation.

Examples:
  tm score "Build a mobile app"
  tm score "Create an AI-powered code reviewer"`,
		Args: cobra.MinimumNArgs(1),
		RunE: runScore,
	}
}

func runScore(cmd *cobra.Command, args []string) error {
	ideaText := strings.Join(args, " ")

	// Show progress
	infoColor.Println("🎯 Scoring idea...")
	fmt.Println()

	// Calculate score
	analysis, err := ctx.Engine.CalculateScore(ideaText)
	if err != nil {
		return fmt.Errorf("failed to score idea: %w", err)
	}

	// Detect patterns
	detectedPatterns := ctx.Detector.DetectPatterns(ideaText)

	// Display results (simplified version)
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("💡 %s\n\n", ideaText)

	// Score with color coding
	scoreColor := getScoreColor(analysis.FinalScore)
	scoreColor.Printf("⭐ Score: %.1f/10.0\n", analysis.FinalScore)

	// Recommendation
	recommendation := analysis.GetRecommendation()
	recommendationColor := getRecommendationColor(recommendation)
	recommendationColor.Printf("%s\n\n", recommendation)

	// Summary scores
	fmt.Printf("📊 Mission Alignment:   %.2f/4.00 (40%%)\n", analysis.Mission.Total)
	fmt.Printf("🎯 Anti-Challenge:      %.2f/3.50 (35%%)\n", analysis.AntiChallenge.Total)
	fmt.Printf("🚀 Strategic Fit:       %.2f/2.50 (25%%)\n\n", analysis.Strategic.Total)

	// Patterns
	if len(detectedPatterns) > 0 {
		warningColor.Println("⚠️  Patterns Detected:")
		for _, p := range detectedPatterns {
			fmt.Printf("  • %s: %s\n", p.Name, p.Description)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 80))
	infoColor.Println("💡 Not saved - use 'tm dump' to save this idea")
	fmt.Println(strings.Repeat("─", 80))

	return nil
}
