package bulk

import (
	"fmt"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/bulk"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the bulk delete command
func NewDeleteCommand(getContext func() *CLIContext) *cobra.Command {
	var olderThan int
	var maxScore float64
	var search string
	var limit int
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete multiple ideas",
		Long: `Permanently delete multiple ideas based on filters.
⚠️  WARNING: This operation cannot be undone!
Always requires confirmation for safety.`,
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
			limitPtr := &limit

			ideas, err := ctx.Repository.List(database.ListOptions{
				Status:   "active",
				MaxScore: maxScorePtr,
				Limit:    limitPtr,
				OrderBy:  "created_at ASC",
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
				fmt.Println("📭 No ideas match your criteria for deletion.")
				return nil
			}

			// Show preview
			if _, err := errorColor.Printf("⚠️  WARNING: About to PERMANENTLY DELETE %d ideas:\n", len(ideas)); err != nil {
				log.Warn().Err(err).Msg("failed to print warning message")
			}
			for i, idea := range ideas {
				if i < 5 {
					fmt.Printf("  - %s (score: %.1f)\n",
						truncate(idea.Content, 50),
						idea.FinalScore)
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			// Always require confirmation for delete
			if !yes {
				fmt.Println()
				if !confirm("⚠️  PERMANENTLY DELETE these ideas? This CANNOT be undone!") {
					fmt.Println("❌ Cancelled")
					return nil
				}
			}

			// Delete ideas
			successCount := 0
			errorCount := 0
			for i, idea := range ideas {
				if err := ctx.Repository.Delete(idea.ID); err != nil {
					if _, printErr := warningColor.Printf("⚠  Failed to delete idea %s: %v\n", idea.ID, err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print error message")
					}
					errorCount++
					continue
				}
				successCount++

				// Show progress for large batches
				if len(ideas) > 10 && (i+1)%10 == 0 {
					fmt.Printf("  Progress: %d/%d deleted\n", i+1, len(ideas))
				}
			}

			if errorCount > 0 {
				if _, err := warningColor.Printf("⚠  %d ideas failed to delete\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := errorColor.Printf("🗑️  Permanently deleted %d ideas\n", successCount); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&olderThan, "older-than", 0, "Delete ideas older than N days")
	cmd.Flags().Float64Var(&maxScore, "max-score", 0, "Maximum score threshold")
	cmd.Flags().StringVar(&search, "search", "", "Search term to filter ideas")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum ideas to process")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}
