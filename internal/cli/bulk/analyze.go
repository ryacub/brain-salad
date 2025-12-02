package bulk

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/patterns"
	"github.com/spf13/cobra"
)

// NewAnalyzeCommand creates the bulk analyze command
func NewAnalyzeCommand(getContext func() *CLIContext) *cobra.Command {
	var (
		scoreMin  float64
		scoreMax  float64
		status    string
		olderThan string
		dryRun    bool
		provider  string
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Re-analyze multiple ideas with updated criteria",
		Long: `Re-analyze multiple ideas using current telos and LLM provider.

This is useful when:
- Your telos has changed
- You want to use a different LLM provider
- You've improved the analysis algorithm
- You want to refresh old analyses

Examples:
  # Re-analyze all low-scoring ideas
  telos bulk analyze --score-max 5.0

  # Re-analyze ideas from last month
  telos bulk analyze --older-than 30d

  # Re-analyze with specific provider
  telos bulk analyze --provider ollama

  # Dry-run to see what would be analyzed
  telos bulk analyze --score-max 5.0 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBulkAnalyze(getContext, bulkAnalyzeOptions{
				scoreMin:  scoreMin,
				scoreMax:  scoreMax,
				status:    status,
				olderThan: olderThan,
				dryRun:    dryRun,
				provider:  provider,
				yes:       yes,
			})
		},
	}

	cmd.Flags().Float64Var(&scoreMin, "score-min", 0, "Minimum score (inclusive)")
	cmd.Flags().Float64Var(&scoreMax, "score-max", 10, "Maximum score (inclusive)")
	cmd.Flags().StringVar(&status, "status", "active", "Filter by status (active|archived|deleted)")
	cmd.Flags().StringVar(&olderThan, "older-than", "", "Re-analyze ideas older than duration (e.g., 30d, 6h)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be analyzed without making changes")
	cmd.Flags().StringVar(&provider, "provider", "", "LLM provider to use (ollama|claude|openai|rule_based)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}

// bulkAnalyzeOptions contains options for bulk analysis
type bulkAnalyzeOptions struct {
	scoreMin  float64
	scoreMax  float64
	status    string
	olderThan string
	dryRun    bool
	provider  string
	yes       bool
}

// runBulkAnalyze performs bulk re-analysis of ideas
func runBulkAnalyze(getContext func() *CLIContext, opts bulkAnalyzeOptions) error {
	ctx := getContext()
	if ctx == nil {
		return fmt.Errorf("CLI context not initialized")
	}

	// Create service once

	// Parse olderThan duration if specified
	var cutoffTime time.Time
	if opts.olderThan != "" {
		duration, err := parseDuration(opts.olderThan)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		cutoffTime = time.Now().UTC().Add(-duration)
	}

	// Build filter criteria
	minScorePtr := &opts.scoreMin
	maxScorePtr := &opts.scoreMax
	limit := 1000 // Safety limit

	ideas, err := ctx.Repository.List(database.ListOptions{
		Status:   opts.status,
		MinScore: minScorePtr,
		MaxScore: maxScorePtr,
		Limit:    &limit,
		OrderBy:  "created_at ASC",
	})
	if err != nil {
		return fmt.Errorf("failed to find ideas: %w", err)
	}

	// Filter by age if specified
	if !cutoffTime.IsZero() {
		ideas = filterByAge(ideas, cutoffTime)
	}

	if len(ideas) == 0 {
		fmt.Println("📭 No ideas match the criteria.")
		return nil
	}

	// Show summary
	fmt.Printf("🔍 Found %s ideas matching criteria:\n",
		color.CyanString("%d", len(ideas)))
	fmt.Printf("  Score range: %.1f - %.1f\n", opts.scoreMin, opts.scoreMax)
	if opts.status != "" {
		fmt.Printf("  Status: %s\n", opts.status)
	}
	if opts.olderThan != "" {
		fmt.Printf("  Older than: %s\n", opts.olderThan)
	}
	fmt.Println()

	if opts.dryRun {
		if _, err := cliutil.InfoColor.Println("🔍 DRY RUN - No changes will be made"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		fmt.Println()
		for i, idea := range ideas {
			if i < 10 { // Show first 10
				age := time.Since(idea.CreatedAt).Hours() / 24
				fmt.Printf("%d. [%s] %s (score: %.1f, age: %.0fd)\n",
					i+1, idea.ID[:8], cliutil.TruncateText(idea.Content, 60), idea.FinalScore, age)
			}
		}
		if len(ideas) > 10 {
			fmt.Printf("... and %d more\n", len(ideas)-10)
		}
		return nil
	}

	// Confirm with user
	if !opts.yes && !cliutil.Confirm(fmt.Sprintf("Re-analyze %d ideas?", len(ideas))) {
		fmt.Println("❌ Cancelled")
		return nil
	}

	// Create LLM manager
	llmManager := ctx.LLMManager
	if llmManager == nil {
		llmManager = createLLMManager()
	}

	// Set provider if specified
	if opts.provider != "" {
		if err := llmManager.SetPrimaryProvider(opts.provider); err != nil {
			return fmt.Errorf("failed to set provider: %w", err)
		}
		if _, err := cliutil.InfoColor.Printf("🤖 Using provider: %s\n", opts.provider); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
	} else {
		primaryProvider := llmManager.GetPrimaryProvider()
		if primaryProvider != nil {
			if _, err := cliutil.InfoColor.Printf("🤖 Using provider: %s\n", primaryProvider.Name()); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
		}
	}
	fmt.Println()

	// Create detector from telos
	detector := patterns.NewDetector(ctx.Telos)

	// Analyze ideas with progress tracking
	successful := 0
	failed := 0
	errors := make([]string, 0)

	for i, idea := range ideas {
		// Show progress
		progress := float64(i+1) / float64(len(ideas)) * 100
		fmt.Printf("\r[%d/%d] 🔄 Analyzing ideas... %.1f%%",
			i+1, len(ideas), progress)

		// Re-analyze using LLM
		result, err := llmManager.AnalyzeWithTelos(idea.Content, ctx.Telos)
		if err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", idea.ID[:8], err))
			continue
		}

		// Detect patterns
		detectedPatterns := detector.DetectPatterns(idea.Content)
		patternStrings := make([]string, len(detectedPatterns))
		for j, p := range detectedPatterns {
			patternStrings[j] = fmt.Sprintf("%s: %s", p.Name, p.Description)
		}

		// Format explanations as JSON for storage
		analysisDetails := ""
		if len(result.Explanations) > 0 {
			detailsMap := map[string]interface{}{
				"explanations": result.Explanations,
				"provider":     result.Provider,
				"scores": map[string]float64{
					"mission_alignment": result.Scores.MissionAlignment,
					"anti_challenge":    result.Scores.AntiChallenge,
					"strategic_fit":     result.Scores.StrategicFit,
				},
			}
			detailsBytes, _ := json.Marshal(detailsMap)
			analysisDetails = string(detailsBytes)
		} else {
			analysisDetails = result.Recommendation
		}

		// Update idea
		idea.FinalScore = result.FinalScore
		idea.Patterns = patternStrings
		idea.Recommendation = result.Recommendation
		idea.AnalysisDetails = analysisDetails

		if err := ctx.Repository.Update(idea); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: failed to save: %v", idea.ID[:8], err))
			continue
		}

		successful++
	}

	fmt.Println() // New line after progress
	fmt.Println()

	// Show summary
	if _, err := cliutil.SuccessColor.Printf("✅ Re-analysis complete:\n"); err != nil {
		log.Warn().Err(err).Msg("failed to print success message")
	}
	fmt.Printf("  ✓ Successful: %d\n", successful)
	if failed > 0 {
		if _, err := cliutil.WarningColor.Printf("  ✗ Failed: %d\n", failed); err != nil {
			log.Warn().Err(err).Msg("failed to print failed count")
		}
		if len(errors) > 0 && len(errors) <= 10 {
			fmt.Println("\nErrors:")
			for _, errMsg := range errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		} else if len(errors) > 10 {
			fmt.Printf("\n  (Showing first 10 of %d errors)\n", len(errors))
			for i := 0; i < 10; i++ {
				fmt.Printf("  - %s\n", errors[i])
			}
		}
	}

	return nil
}
