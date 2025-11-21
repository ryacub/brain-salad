package bulk

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/bulk"
	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewArchiveCommand creates the bulk archive command
func NewArchiveCommand(getContext func() *CLIContext) *cobra.Command {
	var olderThan int
	var maxScore float64
	var minScore float64
	var search string
	var limit int
	var yes bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive multiple old/low-scoring ideas",
		Long: `Archive multiple ideas based on age and score filters.
Use --older-than to archive ideas older than N days.
Use --max-score to archive ideas below a score threshold.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext()
			if ctx == nil {
				return fmt.Errorf("CLI context not initialized")
			}

			// Create service once
			service := bulk.NewService(ctx.Repository)

			// Build filter options
			maxScorePtr := &maxScore
			if maxScore == 0 {
				maxScorePtr = nil
			}
			minScorePtr := &minScore
			if minScore == 0 {
				minScorePtr = nil
			}
			limitPtr := &limit

			ideas, err := ctx.Repository.List(database.ListOptions{
				Status:   "active",
				MinScore: minScorePtr,
				MaxScore: maxScorePtr,
				Limit:    limitPtr,
				OrderBy:  "created_at ASC", // Oldest first
			})
			if err != nil {
				return fmt.Errorf("failed to list ideas: %w", err)
			}

			// Filter by age if specified
			if olderThan > 0 {
				cutoffDate := time.Now().UTC().Add(-time.Duration(olderThan) * 24 * time.Hour)
				ideas = service.FilterByAge(ideas, cutoffDate)
			}

			// Filter by search if provided
			if search != "" {
				ideas = service.FilterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria for archiving.")
				return nil
			}

			// Show preview
			fmt.Printf("📦 Found %s ideas to archive:\n", color.CyanString("%d", len(ideas)))
			for i, idea := range ideas {
				if i < 5 {
					age := time.Since(idea.CreatedAt).Hours() / 24
					fmt.Printf("  - %s (score: %.1f, age: %.0f days)\n",
						cliutil.TruncateText(idea.Content, 50),
						idea.FinalScore,
						age)
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			if dryRun {
				if _, err := cliutil.InfoColor.Println("\n🔍 DRY RUN - No changes will be made"); err != nil {
					log.Warn().Err(err).Msg("failed to print message")
				}
				return nil
			}

			// Confirm
			if !yes && !cliutil.Confirm("Proceed with archiving?") {
				fmt.Println("❌ Cancelled")
				return nil
			}

			// Archive ideas
			successCount := 0
			errorCount := 0
			for i, idea := range ideas {
				idea.Status = "archived"
				if err := ctx.Repository.Update(idea); err != nil {
					if _, printErr := cliutil.WarningColor.Printf("⚠  Failed to archive idea %s: %v\n", idea.ID, err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print error message")
					}
					errorCount++
					continue
				}
				successCount++

				// Show progress for large batches
				if len(ideas) > 10 && (i+1)%10 == 0 {
					fmt.Printf("  Progress: %d/%d archived\n", i+1, len(ideas))
				}
			}

			if errorCount > 0 {
				if _, err := cliutil.WarningColor.Printf("⚠  %d ideas failed to archive\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := cliutil.SuccessColor.Printf("✅ Archived %d ideas\n", successCount); err != nil {
				log.Warn().Err(err).Msg("failed to print success message")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&olderThan, "older-than", 0, "Archive ideas older than N days")
	cmd.Flags().Float64Var(&maxScore, "max-score", 0, "Maximum score threshold")
	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum score threshold")
	cmd.Flags().StringVar(&search, "search", "", "Search term to filter ideas")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum ideas to process")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be archived without making changes")

	return cmd
}
