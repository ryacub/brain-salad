package bulk

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/export"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the bulk import command
func NewImportCommand(getContext func() *CLIContext) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import ideas from CSV",
		Long: `Import ideas from a CSV file.
The CSV file should have the following columns:
ID,Content,RawScore,FinalScore,Patterns,Recommendation,AnalysisDetails,CreatedAt,Status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext()
			if ctx == nil {
				return fmt.Errorf("CLI context not initialized")
			}

			filename := args[0]

			// Import from CSV
			ideas, err := export.ImportCSV(filename)
			if err != nil {
				return fmt.Errorf("failed to import CSV: %w", err)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas found in CSV file.")
				return nil
			}

			// Show preview
			fmt.Printf("📥 Found %s ideas to import from '%s':\n",
				color.CyanString("%d", len(ideas)),
				filename)
			for i, idea := range ideas {
				if i < 5 {
					fmt.Printf("  - %s\n", cliutil.TruncateText(idea.Content, 60))
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			// Confirm
			if !yes && !cliutil.Confirm("Proceed with import?") {
				fmt.Println("❌ Cancelled")
				return nil
			}

			// Import ideas
			successCount := 0
			errorCount := 0
			for i, idea := range ideas {
				// Validate idea before import
				if err := idea.Validate(); err != nil {
					if _, printErr := cliutil.WarningColor.Printf("⚠  Skipping invalid idea: %v\n", err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print warning")
					}
					errorCount++
					continue
				}

				if err := ctx.Repository.Create(idea); err != nil {
					if _, printErr := cliutil.WarningColor.Printf("⚠  Failed to import idea: %v\n", err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print error message")
					}
					errorCount++
					continue
				}
				successCount++

				// Show progress for large batches
				if len(ideas) > 10 && (i+1)%10 == 0 {
					fmt.Printf("  Progress: %d/%d imported\n", i+1, len(ideas))
				}
			}

			if errorCount > 0 {
				if _, err := cliutil.WarningColor.Printf("⚠  %d ideas failed to import\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := cliutil.SuccessColor.Printf("✅ Imported %d ideas from '%s'\n", successCount, filename); err != nil {
				log.Warn().Err(err).Msg("failed to print success message")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}
