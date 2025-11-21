package cli

import (
	"fmt"
	"strings"

	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rs/zerolog/log"
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
	if _, err := cliutil.InfoColor.Println("🎯 Scoring idea..."); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
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
	scoreColor := cliutil.GetScoreColor(analysis.FinalScore)
	if _, err := scoreColor.Printf("⭐ Score: %.1f/10.0\n", analysis.FinalScore); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}

	// Recommendation
	recommendation := analysis.GetRecommendation()
	recommendationColor := cliutil.GetRecommendationColor(recommendation)
	if _, err := recommendationColor.Printf("%s\n\n", recommendation); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}

	// Summary scores
	fmt.Printf("📊 Mission Alignment:   %.2f/4.00 (40%%)\n", analysis.Mission.Total)
	fmt.Printf("🎯 Anti-Challenge:      %.2f/3.50 (35%%)\n", analysis.AntiChallenge.Total)
	fmt.Printf("🚀 Strategic Fit:       %.2f/2.50 (25%%)\n\n", analysis.Strategic.Total)

	// Patterns
	if len(detectedPatterns) > 0 {
		if _, err := cliutil.WarningColor.Println("⚠️  Patterns Detected:"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		for _, p := range detectedPatterns {
			fmt.Printf("  • %s: %s\n", p.Name, p.Description)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 80))
	if _, err := cliutil.InfoColor.Println("💡 Not saved - use 'tm dump' to save this idea"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))

	return nil
}
