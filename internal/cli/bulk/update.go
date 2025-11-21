package bulk

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/bulk"
	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the bulk update command
func NewUpdateCommand(getContext func() *CLIContext) *cobra.Command {
	var (
		setStatus      string
		addPatterns    string
		removePatterns string
		addTags        string
		removeTags     string
		scoreMin       float64
		scoreMax       float64
		statusFilter   string
		dryRun         bool
		yes            bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update multiple ideas in batch",
		Long: `Update fields for multiple ideas based on filtering criteria.

Supports updating:
- Status (active/archived/deleted)
- Patterns (add/remove)
- Tags (add/remove)

Examples:
  # Archive all low-scoring ideas
  telos bulk update --score-max 3.0 --set-status archived

  # Add pattern to all ideas in score range
  telos bulk update --score-min 7.0 --add-patterns "high-value"

  # Remove obsolete pattern from all ideas
  telos bulk update --remove-patterns "old-pattern"

  # Dry-run to preview changes
  telos bulk update --score-max 3.0 --set-status archived --dry-run

  # Add tag to archived ideas
  telos bulk update --status archived --add-tags "reviewed"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBulkUpdate(getContext, bulkUpdateOptions{
				setStatus:      setStatus,
				addPatterns:    bulk.SplitCommaSeparated(addPatterns),
				removePatterns: bulk.SplitCommaSeparated(removePatterns),
				addTags:        bulk.SplitCommaSeparated(addTags),
				removeTags:     bulk.SplitCommaSeparated(removeTags),
				scoreMin:       scoreMin,
				scoreMax:       scoreMax,
				statusFilter:   statusFilter,
				dryRun:         dryRun,
				yes:            yes,
			})
		},
	}

	// Update operations
	cmd.Flags().StringVar(&setStatus, "set-status", "", "Set status (active|archived|deleted)")
	cmd.Flags().StringVar(&addPatterns, "add-patterns", "", "Add patterns (comma-separated)")
	cmd.Flags().StringVar(&removePatterns, "remove-patterns", "", "Remove patterns (comma-separated)")
	cmd.Flags().StringVar(&addTags, "add-tags", "", "Add tags (comma-separated)")
	cmd.Flags().StringVar(&removeTags, "remove-tags", "", "Remove tags (comma-separated)")

	// Filters
	cmd.Flags().Float64Var(&scoreMin, "score-min", 0, "Minimum score filter")
	cmd.Flags().Float64Var(&scoreMax, "score-max", 10, "Maximum score filter")
	cmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status")

	// Options
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}

type bulkUpdateOptions struct {
	setStatus      string
	addPatterns    []string
	removePatterns []string
	addTags        []string
	removeTags     []string
	scoreMin       float64
	scoreMax       float64
	statusFilter   string
	dryRun         bool
	yes            bool
}

func runBulkUpdate(getContext func() *CLIContext, opts bulkUpdateOptions) error {
	// Validate that at least one update operation is specified
	if opts.setStatus == "" &&
		len(opts.addPatterns) == 0 &&
		len(opts.removePatterns) == 0 &&
		len(opts.addTags) == 0 &&
		len(opts.removeTags) == 0 {
		return fmt.Errorf("no updates specified (use --set-status, --add-patterns, --remove-patterns, --add-tags, or --remove-tags)")
	}

	// Validate status value if provided
	if opts.setStatus != "" {
		validStatuses := []string{"active", "archived", "deleted"}
		if !bulk.Contains(validStatuses, opts.setStatus) {
			return fmt.Errorf("invalid status: %s (must be one of: %s)",
				opts.setStatus, strings.Join(validStatuses, ", "))
		}
	}

	ctx := getContext()
	if ctx == nil {
		return fmt.Errorf("CLI context not initialized")
	}

	// Build filter
	minScorePtr := &opts.scoreMin
	maxScorePtr := &opts.scoreMax
	limitPtr := new(int)
	*limitPtr = 10000 // High limit for bulk operations

	ideas, err := ctx.Repository.List(database.ListOptions{
		Status:   opts.statusFilter,
		MinScore: minScorePtr,
		MaxScore: maxScorePtr,
		Limit:    limitPtr,
		OrderBy:  "final_score DESC",
	})
	if err != nil {
		return fmt.Errorf("failed to find ideas: %w", err)
	}

	if len(ideas) == 0 {
		fmt.Println("📭 No ideas match the criteria.")
		return nil
	}

	// Show preview
	fmt.Printf("🎯 Found %s ideas to update:\n\n",
		color.CyanString("%d", len(ideas)))
	fmt.Println("Filters applied:")
	fmt.Printf("  Score range: %.1f - %.1f\n", opts.scoreMin, opts.scoreMax)
	if opts.statusFilter != "" {
		fmt.Printf("  Status: %s\n", opts.statusFilter)
	}
	fmt.Println()

	fmt.Println("Updates to apply:")
	if opts.setStatus != "" {
		fmt.Printf("  - Set status: %s\n", color.GreenString(opts.setStatus))
	}
	if len(opts.addPatterns) > 0 {
		fmt.Printf("  - Add patterns: %s\n", color.GreenString(strings.Join(opts.addPatterns, ", ")))
	}
	if len(opts.removePatterns) > 0 {
		fmt.Printf("  - Remove patterns: %s\n", color.YellowString(strings.Join(opts.removePatterns, ", ")))
	}
	if len(opts.addTags) > 0 {
		fmt.Printf("  - Add tags: %s\n", color.GreenString(strings.Join(opts.addTags, ", ")))
	}
	if len(opts.removeTags) > 0 {
		fmt.Printf("  - Remove tags: %s\n", color.YellowString(strings.Join(opts.removeTags, ", ")))
	}
	fmt.Println()

	if opts.dryRun {
		if _, err := cliutil.InfoColor.Println("🔍 DRY RUN - Showing affected ideas and changes:"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		for i, idea := range ideas {
			if i >= 10 {
				fmt.Printf("\n... and %d more ideas\n", len(ideas)-10)
				break
			}
			fmt.Printf("\n%d. [%s] %s\n", i+1, idea.ID[:8], cliutil.TruncateText(idea.Content, 60))
			fmt.Printf("   Current - Score: %.1f, Status: %s\n", idea.FinalScore, idea.Status)

			if opts.setStatus != "" && idea.Status != opts.setStatus {
				fmt.Printf("   %s Status change: %s → %s\n",
					color.CyanString("→"), idea.Status, opts.setStatus)
			}
			if len(opts.addPatterns) > 0 {
				newPatterns := bulk.AddUniqueStrings(idea.Patterns, opts.addPatterns)
				if len(newPatterns) > len(idea.Patterns) {
					fmt.Printf("   %s Patterns: %v → %v\n",
						color.CyanString("→"), idea.Patterns, newPatterns)
				}
			}
			if len(opts.removePatterns) > 0 {
				newPatterns := bulk.RemoveStrings(idea.Patterns, opts.removePatterns)
				if len(newPatterns) < len(idea.Patterns) {
					fmt.Printf("   %s Patterns: %v → %v\n",
						color.CyanString("→"), idea.Patterns, newPatterns)
				}
			}
			if len(opts.addTags) > 0 {
				newTags := bulk.AddUniqueStrings(idea.Tags, opts.addTags)
				if len(newTags) > len(idea.Tags) {
					fmt.Printf("   %s Tags: %v → %v\n",
						color.CyanString("→"), idea.Tags, newTags)
				}
			}
			if len(opts.removeTags) > 0 {
				newTags := bulk.RemoveStrings(idea.Tags, opts.removeTags)
				if len(newTags) < len(idea.Tags) {
					fmt.Printf("   %s Tags: %v → %v\n",
						color.CyanString("→"), idea.Tags, newTags)
				}
			}
		}
		return nil
	}

	// Confirm
	if !opts.yes && !cliutil.Confirm(fmt.Sprintf("Update %d ideas?", len(ideas))) {
		fmt.Println("❌ Cancelled")
		return nil
	}

	// Apply updates
	updated := 0
	unchanged := 0
	failed := 0
	errors := make([]string, 0)

	service := bulk.NewService(ctx.Repository)
	updateOpts := bulk.UpdateOptions{
		SetStatus:      opts.setStatus,
		AddPatterns:    opts.addPatterns,
		RemovePatterns: opts.removePatterns,
		AddTags:        opts.addTags,
		RemoveTags:     opts.removeTags,
	}

	for i, idea := range ideas {
		// Apply updates using service
		modified := service.ApplyUpdates(idea, updateOpts)

		// Only save if something actually changed
		if modified {
			if err := ctx.Repository.Update(idea); err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("%s: %v", idea.ID[:8], err))
				continue
			}
			updated++
		} else {
			unchanged++
		}

		// Show progress for large batches
		if len(ideas) > 10 && (i+1)%10 == 0 {
			fmt.Printf("  Progress: %d/%d processed\n", i+1, len(ideas))
		}
	}

	fmt.Printf("\n%s Update complete:\n", cliutil.SuccessColor.Sprint("✅"))
	fmt.Printf("  ✓ Updated: %s\n", color.GreenString("%d", updated))
	if unchanged > 0 {
		fmt.Printf("  - Unchanged: %s (no modifications needed)\n", color.CyanString("%d", unchanged))
	}
	if failed > 0 {
		fmt.Printf("  ✗ Failed: %s\n", cliutil.ErrorColor.Sprint(failed))
		if len(errors) > 0 && len(errors) <= 10 {
			fmt.Println("\nErrors:")
			for _, errMsg := range errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
	}

	return nil
}
