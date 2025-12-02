package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	var minScore float64
	var maxScore float64
	var status string
	var limit int
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved ideas",
		Long: `List and filter your saved ideas.

Examples:
  tm list                      # List recent ideas
  tm list --min-score 7.0      # High-scoring ideas only
  tm list --status archived    # Archived ideas
  tm list --limit 20           # Show more ideas
  tm list --json               # JSON output for scripting
  tm list -q                   # Compact output`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := database.ListOptions{
				Status:  status,
				OrderBy: "final_score DESC",
			}

			if cmd.Flags().Changed("min-score") {
				opts.MinScore = &minScore
			}
			if cmd.Flags().Changed("max-score") {
				opts.MaxScore = &maxScore
			}
			if limit > 0 {
				opts.Limit = &limit
			}

			ideas, err := ctx.Repository.List(opts)
			if err != nil {
				return fmt.Errorf("failed to list: %w", err)
			}

			if len(ideas) == 0 {
				if jsonOutput {
					fmt.Println("[]")
				} else if !quiet {
					_, _ = cliutil.InfoColor.Println("No ideas found.")
				}
				return nil
			}

			// JSON output
			if jsonOutput {
				return outputListJSON(ideas)
			}

			// Quiet output
			if quiet {
				return outputListQuiet(ideas)
			}

			// Full output
			return outputListFull(ideas)
		},
	}

	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum score")
	cmd.Flags().Float64Var(&maxScore, "max-score", 0, "Maximum score")
	cmd.Flags().StringVar(&status, "status", "active", "Status (active|archived|deleted)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Max ideas to show")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Compact output")

	return cmd
}

type listItem struct {
	ID             string   `json:"id"`
	Content        string   `json:"content"`
	Score          float64  `json:"score"`
	Recommendation string   `json:"recommendation"`
	Patterns       []string `json:"patterns,omitempty"`
	CreatedAt      string   `json:"created_at"`
}

func outputListJSON(ideas []*models.Idea) error {
	items := make([]listItem, len(ideas))
	for i, idea := range ideas {
		items[i] = listItem{
			ID:             idea.ID,
			Content:        idea.Content,
			Score:          idea.FinalScore,
			Recommendation: idea.Recommendation,
			Patterns:       idea.Patterns,
			CreatedAt:      idea.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	output, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputListQuiet(ideas []*models.Idea) error {
	for _, idea := range ideas {
		scoreColor := cliutil.GetScoreColor(idea.FinalScore)
		_, _ = scoreColor.Printf("%.1f", idea.FinalScore)
		fmt.Printf(" %s %s\n", idea.ID[:8], cliutil.TruncateText(idea.Content, 50))
	}
	return nil
}

func outputListFull(ideas []*models.Idea) error {
	fmt.Println(strings.Repeat("─", 60))
	_, _ = cliutil.SuccessColor.Printf("%d ideas\n", len(ideas))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	for i, idea := range ideas {
		scoreColor := cliutil.GetScoreColor(idea.FinalScore)

		// Header: "1. 8.5/10 - abc123"
		fmt.Printf("%d. ", i+1)
		_, _ = scoreColor.Printf("%.1f/10", idea.FinalScore)
		fmt.Printf(" - %s\n", idea.ID[:8])

		// Content
		fmt.Printf("   %s\n", cliutil.TruncateText(idea.Content, 55))

		// Recommendation
		if idea.Recommendation != "" {
			recColor := cliutil.GetRecommendationColor(idea.Recommendation)
			if _, err := recColor.Printf("   %s\n", idea.Recommendation); err != nil {
				log.Warn().Err(err).Msg("failed to print")
			}
		}

		// Date
		fmt.Printf("   %s\n\n", idea.CreatedAt.Format("Jan 2, 2006"))
	}

	fmt.Println(strings.Repeat("─", 60))
	_, _ = cliutil.InfoColor.Printf("Use 'tm show <id>' for details\n")

	return nil
}
