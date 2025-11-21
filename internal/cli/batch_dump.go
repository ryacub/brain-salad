package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/cli/dump"
	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newBatchDumpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch <file>",
		Short: "Quickly dump multiple ideas from a file",
		Long: `Quickly dump multiple ideas from a file (one per line).

Uses quick mode for fast batch processing.

Example:
  tm dump batch ideas.txt

File format:
  - One idea per line
  - Empty lines are skipped
  - Lines starting with # are treated as comments`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchDump(args[0])
		},
	}

	return cmd
}

func runBatchDump(filename string) error {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Split into lines
	lines := strings.Split(string(data), "\n")

	start := time.Now()
	successCount := 0
	errorCount := 0

	fmt.Printf("Processing %d lines from %s...\n\n", len(lines), filename)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(lines), dump.TruncateText(line, 60))

		err := dump.RunQuickDump(line, false, ctx.Repository)
		if err != nil {
			if _, printErr := cliutil.ErrorColor.Printf("  ✗ Error: %v\n", err); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print error message")
			}
			errorCount++
		} else {
			if _, printErr := cliutil.SuccessColor.Printf("  ✓ Saved\n"); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print success message")
			}
			successCount++
		}
		fmt.Println()
	}

	elapsed := time.Since(start)

	fmt.Println(strings.Repeat("═", 80))
	fmt.Println("Batch processing complete:")
	if _, err := cliutil.SuccessColor.Printf("  ✓ Success: %d ideas\n", successCount); err != nil {
		log.Warn().Err(err).Msg("failed to print success message")
	}
	if errorCount > 0 {
		if _, err := cliutil.ErrorColor.Printf("  ✗ Errors: %d\n", errorCount); err != nil {
			log.Warn().Err(err).Msg("failed to print error count")
		}
	}
	fmt.Printf("  ⚡ Time: %v (%.1f ideas/sec)\n", elapsed, float64(successCount)/elapsed.Seconds())
	fmt.Println(strings.Repeat("═", 80))

	return nil
}
