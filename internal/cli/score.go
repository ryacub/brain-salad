package cli

import (
	"fmt"
	"strings"

	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newScoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "score <idea text>",
		Short: "Score an idea without saving it",
		Long: `Score an idea against your profile without saving it to the database.
Useful for quick idea validation.

Examples:
  tm score "Build a mobile app"
  tm score "Sell pottery at the farmer's market"`,
		Args: cobra.MinimumNArgs(1),
		RunE: runScore,
	}
}

func runScore(cmd *cobra.Command, args []string) error {
	ideaText := strings.Join(args, " ")

	// Show progress
	if _, err := cliutil.InfoColor.Println("Scoring idea..."); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println()

	// Route to appropriate scoring mode
	if ctx.ScoringMode == ScoringModeUniversal {
		return runUniversalScore(ideaText)
	}
	return runLegacyScore(ideaText)
}

// runUniversalScore displays scores using the universal scoring engine
func runUniversalScore(ideaText string) error {
	// Calculate score
	analysis, err := ctx.UniversalEngine.Score(ideaText)
	if err != nil {
		return fmt.Errorf("failed to score idea: %w", err)
	}

	// Display results
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("%s\n\n", ideaText)

	// Score with color coding
	scoreColor := cliutil.GetScoreColor(analysis.FinalScore)
	if _, err := scoreColor.Printf("Score: %.1f/10.0 — %s\n\n", analysis.FinalScore, analysis.Recommendation); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}

	// Display dimension breakdown with visual bars
	displayUniversalDimensions(&analysis.Universal)

	// Display insights if any
	if len(analysis.Insights) > 0 {
		fmt.Println()
		_, _ = cliutil.InfoColor.Println("Insights:")
		for _, insight := range analysis.Insights {
			fmt.Printf("  • %s\n", insight)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	if _, err := cliutil.InfoColor.Println("Not saved — use 'tm dump' to save this idea"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}

	return nil
}

// displayUniversalDimensions shows the dimension breakdown with visual bars
func displayUniversalDimensions(scores *scoring.UniversalScores) {
	dimensions := scores.ToSlice()

	for _, dim := range dimensions {
		// Calculate bar width (10 chars = full bar)
		ratio := dim.Score / dim.MaxScore
		filledBars := int(ratio * 10)
		emptyBars := 10 - filledBars

		bar := strings.Repeat("█", filledBars) + strings.Repeat("░", emptyBars)

		// Color based on score ratio
		var dimColor = cliutil.InfoColor
		if ratio >= 0.7 {
			dimColor = cliutil.SuccessColor
		} else if ratio < 0.4 {
			dimColor = cliutil.WarningColor
		}

		// Format: "  Completion    ████████░░  1.6/2.0  Will I finish this?"
		_, _ = dimColor.Printf("  %-12s %s  %.1f/%.1f  %s\n",
			dim.Name, bar, dim.Score, dim.MaxScore, dim.Description)
	}
}

// runLegacyScore displays scores using the legacy telos-based engine
func runLegacyScore(ideaText string) error {
	// Calculate score
	analysis, err := ctx.Engine.CalculateScore(ideaText)
	if err != nil {
		return fmt.Errorf("failed to score idea: %w", err)
	}

	// Detect patterns
	detectedPatterns := ctx.Detector.DetectPatterns(ideaText)

	// Display results (simplified version)
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("%s\n\n", ideaText)

	// Score with color coding
	scoreColor := cliutil.GetScoreColor(analysis.FinalScore)
	if _, err := scoreColor.Printf("Score: %.1f/10.0\n", analysis.FinalScore); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}

	// Recommendation
	recommendation := analysis.GetRecommendation()
	recommendationColor := cliutil.GetRecommendationColor(recommendation)
	if _, err := recommendationColor.Printf("%s\n\n", recommendation); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}

	// Summary scores
	fmt.Printf("Mission Alignment:   %.2f/4.00 (40%%)\n", analysis.Mission.Total)
	fmt.Printf("Anti-Challenge:      %.2f/3.50 (35%%)\n", analysis.AntiChallenge.Total)
	fmt.Printf("Strategic Fit:       %.2f/2.50 (25%%)\n\n", analysis.Strategic.Total)

	// Patterns
	if len(detectedPatterns) > 0 {
		if _, err := cliutil.WarningColor.Println("Patterns Detected:"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		for _, p := range detectedPatterns {
			fmt.Printf("  • %s: %s\n", p.Name, p.Description)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 80))
	if _, err := cliutil.InfoColor.Println("Not saved — use 'tm dump' to save this idea"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))

	return nil
}
