package bulk

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/models"
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
			ideas, err := importCSV(filename)
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

// importCSV reads ideas from a CSV file.
func importCSV(filename string) ([]*models.Idea, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close file")
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("csv file is empty")
	}

	// Skip header row
	if len(records) == 1 {
		// Only header, return empty slice
		return []*models.Idea{}, nil
	}

	ideas := make([]*models.Idea, 0, len(records)-1)

	for i, record := range records[1:] {
		if len(record) < 9 {
			return nil, fmt.Errorf("row %d: invalid format, expected 9 columns, got %d", i+2, len(record))
		}

		// Parse scores (with default 0.0 on error)
		rawScore, _ := strconv.ParseFloat(record[2], 64)
		finalScore, _ := strconv.ParseFloat(record[3], 64)

		// Parse patterns (split by comma)
		var patterns []string
		if record[4] != "" {
			patterns = strings.Split(record[4], ",")
		}

		// Parse timestamp
		createdAt, err := time.Parse(time.RFC3339, record[7])
		if err != nil {
			// Default to current time if parsing fails
			createdAt = time.Now().UTC()
		}

		idea := &models.Idea{
			ID:              record[0],
			Content:         record[1],
			RawScore:        rawScore,
			FinalScore:      finalScore,
			Patterns:        patterns,
			Recommendation:  record[5],
			AnalysisDetails: record[6],
			CreatedAt:       createdAt,
			Status:          record[8],
		}

		ideas = append(ideas, idea)
	}

	return ideas, nil
}
