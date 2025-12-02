package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/ryacub/telos-idea-matrix/internal/scoring"
	"github.com/ryacub/telos-idea-matrix/internal/utils"
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	var dryRun bool
	var useAI bool
	var provider string
	var quiet bool
	var jsonOutput bool
	var fromClipboard bool
	var toClipboard bool

	cmd := &cobra.Command{
		Use:   "add <idea>",
		Short: "Add and score an idea",
		Long: `Add an idea, score it against your goals, and save it.

Examples:
  tm add "Build a mobile app"              # Add and save
  tm add "Start a podcast" --ai            # Add with AI analysis
  tm add "Learn Rust" -n                   # Dry-run: score without saving
  tm add "Quick idea" -q                   # Quiet: minimal output
  tm add --from-clipboard                  # Read from clipboard
  tm add "My idea" --json                  # Output as JSON

Flags:
  -n, --dry-run       Score without saving (preview mode)
  -q, --quiet         Minimal output
      --ai            Use AI for deeper analysis
      --json          Output as JSON (for scripting)`,
		Args: func(cmd *cobra.Command, args []string) error {
			fromClip, _ := cmd.Flags().GetBool("from-clipboard")
			if !fromClip && len(args) < 1 {
				return fmt.Errorf("provide an idea or use --from-clipboard")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get idea text
			var ideaText string
			if fromClipboard {
				text, err := utils.PasteFromClipboard()
				if err != nil {
					return fmt.Errorf("read clipboard: %w", err)
				}
				ideaText = strings.TrimSpace(text)
				if ideaText == "" {
					return fmt.Errorf("clipboard is empty")
				}
			} else {
				ideaText = strings.Join(args, " ")
			}

			return runAdd(ideaText, addOptions{
				dryRun:      dryRun,
				useAI:       useAI,
				provider:    provider,
				quiet:       quiet,
				jsonOutput:  jsonOutput,
				toClipboard: toClipboard,
			})
		},
	}

	// Standard flags per clig.dev
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Score without saving (preview mode)")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Minimal output")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	// Feature flags
	cmd.Flags().BoolVar(&useAI, "ai", false, "Use AI for deeper analysis")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "AI provider (ollama|openai|claude)")

	// Clipboard flags
	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read idea from clipboard")
	cmd.Flags().BoolVar(&toClipboard, "to-clipboard", false, "Copy result to clipboard")

	return cmd
}

type addOptions struct {
	dryRun      bool
	useAI       bool
	provider    string
	quiet       bool
	jsonOutput  bool
	toClipboard bool
}

type addResult struct {
	ID             string   `json:"id,omitempty"`
	Content        string   `json:"content"`
	Score          float64  `json:"score"`
	Recommendation string   `json:"recommendation"`
	Saved          bool     `json:"saved"`
	Insights       []string `json:"insights,omitempty"`
}

func runAdd(ideaText string, opts addOptions) error {
	var result addResult
	result.Content = ideaText
	result.Saved = !opts.dryRun

	// Score the idea based on mode
	if ctx.ScoringMode == ScoringModeUniversal {
		return runAddUniversal(ideaText, opts)
	}
	return runAddLegacy(ideaText, opts)
}

func runAddUniversal(ideaText string, opts addOptions) error {
	// Calculate score
	analysis, err := ctx.UniversalEngine.Score(ideaText)
	if err != nil {
		return fmt.Errorf("failed to score: %w", err)
	}

	// Create idea
	idea := models.NewIdea(ideaText)
	idea.FinalScore = analysis.FinalScore
	idea.Recommendation = analysis.Recommendation

	// Serialize analysis
	analysisJSON, _ := json.Marshal(analysis)
	idea.AnalysisDetails = string(analysisJSON)

	// Save unless dry-run
	if !opts.dryRun {
		if err := ctx.Repository.Create(idea); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
	}

	// Convert insights map to slice
	var insights []string
	for _, v := range analysis.Insights {
		insights = append(insights, v)
	}

	// Output
	if opts.jsonOutput {
		return outputAddJSON(idea, insights, opts.dryRun)
	}

	if opts.quiet {
		return outputAddQuiet(idea, opts.dryRun)
	}

	return outputAddFull(idea, &analysis.Universal, insights, opts)
}

func runAddLegacy(ideaText string, opts addOptions) error {
	// Use AI if requested
	var analysis *models.Analysis
	var err error

	if opts.useAI {
		analysis, err = ctx.LLMManager.AnalyzeWithProviderOverride(ideaText, opts.provider, "", ctx.Telos)
		if err != nil {
			if !opts.quiet {
				_, _ = cliutil.WarningColor.Printf("AI unavailable, using rule-based: %v\n", err)
			}
			analysis, err = ctx.Engine.CalculateScore(ideaText)
		}
	} else {
		analysis, err = ctx.Engine.CalculateScore(ideaText)
	}

	if err != nil {
		return fmt.Errorf("failed to score: %w", err)
	}

	// Create idea
	idea := models.NewIdea(ideaText)
	idea.FinalScore = analysis.FinalScore
	idea.Recommendation = analysis.GetRecommendation()

	// Detect patterns
	detectedPatterns := ctx.Detector.DetectPatterns(ideaText)
	patternStrings := make([]string, len(detectedPatterns))
	for i, p := range detectedPatterns {
		patternStrings[i] = fmt.Sprintf("%s: %s", p.Name, p.Description)
	}
	idea.Patterns = patternStrings

	// Serialize analysis
	analysisJSON, _ := json.Marshal(analysis)
	idea.AnalysisDetails = string(analysisJSON)

	// Save unless dry-run
	if !opts.dryRun {
		if err := ctx.Repository.Create(idea); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
	}

	// Output
	if opts.jsonOutput {
		return outputAddJSON(idea, nil, opts.dryRun)
	}

	if opts.quiet {
		return outputAddQuiet(idea, opts.dryRun)
	}

	return outputAddFullLegacy(idea, analysis, opts)
}

func outputAddJSON(idea *models.Idea, insights []string, dryRun bool) error {
	result := addResult{
		Content:        idea.Content,
		Score:          idea.FinalScore,
		Recommendation: idea.Recommendation,
		Saved:          !dryRun,
		Insights:       insights,
	}
	if !dryRun {
		result.ID = idea.ID
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputAddQuiet(idea *models.Idea, dryRun bool) error {
	scoreColor := cliutil.GetScoreColor(idea.FinalScore)
	if dryRun {
		_, _ = scoreColor.Printf("%.1f", idea.FinalScore)
		fmt.Printf(" %s\n", idea.Recommendation)
	} else {
		_, _ = scoreColor.Printf("%.1f", idea.FinalScore)
		fmt.Printf(" %s [%s]\n", idea.Recommendation, idea.ID[:8])
	}
	return nil
}

func outputAddFull(idea *models.Idea, scores *scoring.UniversalScores, insights []string, opts addOptions) error {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("%s\n\n", idea.Content)

	// Score with color
	scoreColor := cliutil.GetScoreColor(idea.FinalScore)
	_, _ = scoreColor.Printf("Score: %.1f/10.0 — %s\n\n", idea.FinalScore, idea.Recommendation)

	// Dimension breakdown
	displayUniversalDimensions(scores)

	// Insights
	if len(insights) > 0 {
		fmt.Println()
		_, _ = cliutil.InfoColor.Println("Insights:")
		for _, insight := range insights {
			fmt.Printf("  • %s\n", insight)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))

	// Status message
	if opts.dryRun {
		_, _ = cliutil.InfoColor.Println("Preview only — use 'tm add' without -n to save")
	} else {
		_, _ = cliutil.SuccessColor.Printf("Saved [%s]\n", idea.ID[:8])
	}

	// Clipboard
	if opts.toClipboard {
		summary := fmt.Sprintf("%.1f/10 - %s: %s", idea.FinalScore, idea.Recommendation, idea.Content)
		if err := utils.CopyToClipboard(summary); err != nil {
			log.Warn().Err(err).Msg("failed to copy to clipboard")
		} else {
			_, _ = cliutil.InfoColor.Println("Copied to clipboard")
		}
	}

	return nil
}

func outputAddFullLegacy(idea *models.Idea, analysis *models.Analysis, opts addOptions) error {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("%s\n\n", idea.Content)

	// Score
	scoreColor := cliutil.GetScoreColor(idea.FinalScore)
	_, _ = scoreColor.Printf("Score: %.1f/10.0\n", idea.FinalScore)

	// Recommendation
	recColor := cliutil.GetRecommendationColor(idea.Recommendation)
	_, _ = recColor.Printf("%s\n\n", idea.Recommendation)

	// Score breakdown
	fmt.Printf("Mission:       %.2f/4.00\n", analysis.Mission.Total)
	fmt.Printf("Anti-Challenge: %.2f/3.50\n", analysis.AntiChallenge.Total)
	fmt.Printf("Strategic:     %.2f/2.50\n", analysis.Strategic.Total)

	// Patterns
	if len(idea.Patterns) > 0 {
		fmt.Println()
		_, _ = cliutil.WarningColor.Println("Patterns:")
		for _, p := range idea.Patterns {
			fmt.Printf("  • %s\n", p)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))

	// Status
	if opts.dryRun {
		_, _ = cliutil.InfoColor.Println("Preview only — use 'tm add' without -n to save")
	} else {
		_, _ = cliutil.SuccessColor.Printf("Saved [%s]\n", idea.ID[:8])
	}

	return nil
}

// displayUniversalDimensions shows a visual breakdown of universal scoring dimensions
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
