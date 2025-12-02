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

// Suppress unused import warning
var _ = log.Logger

func newShowCommand() *cobra.Command {
	var last bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show idea details",
		Long: `Show detailed analysis for a saved idea.

Examples:
  tm show abc123              # Show idea by ID
  tm show --last              # Show most recent idea
  tm show abc123 --json       # JSON output`,
		Aliases: []string{"view", "get"},
		Args: func(cmd *cobra.Command, args []string) error {
			lastFlag, _ := cmd.Flags().GetBool("last")
			if !lastFlag && len(args) < 1 {
				return fmt.Errorf("provide an idea ID or use --last")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var idea *models.Idea
			var err error

			if last {
				// Get most recent idea
				limit := 1
				ideas, err := ctx.Repository.List(database.ListOptions{
					Status:  "active",
					OrderBy: "created_at DESC",
					Limit:   &limit,
				})
				if err != nil {
					return fmt.Errorf("failed to fetch: %w", err)
				}
				if len(ideas) == 0 {
					return fmt.Errorf("no ideas found")
				}
				idea = ideas[0]
			} else {
				// Get by ID (support partial IDs)
				ideaID := args[0]
				idea, err = ctx.Repository.GetByID(ideaID)
				if err != nil {
					// Try partial match
					idea, err = ctx.Repository.GetByPartialID(ideaID)
					if err != nil {
						return fmt.Errorf("idea not found: %s", ideaID)
					}
				}
			}

			if jsonOutput {
				return outputShowJSON(idea)
			}
			return outputShowFull(idea)
		},
	}

	cmd.Flags().BoolVar(&last, "last", false, "Show most recent idea")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

type showResult struct {
	ID              string                 `json:"id"`
	Content         string                 `json:"content"`
	Score           float64                `json:"score"`
	Recommendation  string                 `json:"recommendation"`
	Patterns        []string               `json:"patterns,omitempty"`
	AnalysisDetails map[string]interface{} `json:"analysis,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

func outputShowJSON(idea *models.Idea) error {
	updatedAt := idea.CreatedAt
	if idea.ReviewedAt != nil {
		updatedAt = *idea.ReviewedAt
	}

	result := showResult{
		ID:             idea.ID,
		Content:        idea.Content,
		Score:          idea.FinalScore,
		Recommendation: idea.Recommendation,
		Patterns:       idea.Patterns,
		CreatedAt:      idea.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      updatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Parse analysis details if available
	if idea.AnalysisDetails != "" {
		var analysis map[string]interface{}
		if err := json.Unmarshal([]byte(idea.AnalysisDetails), &analysis); err == nil {
			result.AnalysisDetails = analysis
		}
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputShowFull(idea *models.Idea) error {
	fmt.Println(strings.Repeat("═", 60))

	// Header
	_, _ = cliutil.InfoColor.Printf("Idea: %s\n", idea.ID[:8])
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Content
	fmt.Printf("%s\n\n", idea.Content)

	// Score
	scoreColor := cliutil.GetScoreColor(idea.FinalScore)
	_, _ = scoreColor.Printf("Score: %.1f/10.0\n", idea.FinalScore)

	// Recommendation
	if idea.Recommendation != "" {
		recColor := cliutil.GetRecommendationColor(idea.Recommendation)
		_, _ = recColor.Printf("%s\n", idea.Recommendation)
	}
	fmt.Println()

	// Analysis details
	if idea.AnalysisDetails != "" {
		displayStoredAnalysis(idea.AnalysisDetails)
	}

	// Patterns
	if len(idea.Patterns) > 0 {
		_, _ = cliutil.WarningColor.Println("Patterns Detected:")
		for _, p := range idea.Patterns {
			fmt.Printf("  • %s\n", p)
		}
		fmt.Println()
	}

	// Metadata
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Created: %s\n", idea.CreatedAt.Format("Jan 2, 2006 3:04 PM"))
	if idea.ReviewedAt != nil {
		fmt.Printf("Updated: %s\n", idea.ReviewedAt.Format("Jan 2, 2006 3:04 PM"))
	}
	fmt.Printf("ID: %s\n", idea.ID)
	fmt.Println(strings.Repeat("═", 60))

	return nil
}

func displayStoredAnalysis(analysisJSON string) {
	// Try to parse as universal analysis first
	var universalAnalysis struct {
		Universal struct {
			CompletionLikelihood float64 `json:"completion_likelihood"`
			SkillFit             float64 `json:"skill_fit"`
			TimeToDone           float64 `json:"time_to_done"`
			RewardAlignment      float64 `json:"reward_alignment"`
			Sustainability       float64 `json:"sustainability"`
			AvoidanceFit         float64 `json:"avoidance_fit"`
		} `json:"universal"`
	}

	if err := json.Unmarshal([]byte(analysisJSON), &universalAnalysis); err == nil {
		u := universalAnalysis.Universal
		if u.CompletionLikelihood > 0 || u.SkillFit > 0 {
			_, _ = cliutil.InfoColor.Println("Score Breakdown:")
			displayUniversalScoresFromStored(u.CompletionLikelihood, u.SkillFit, u.TimeToDone, u.RewardAlignment, u.Sustainability, u.AvoidanceFit)
			fmt.Println()
			return
		}
	}

	// Try legacy mode
	var analysis models.Analysis
	if err := json.Unmarshal([]byte(analysisJSON), &analysis); err != nil {
		return
	}

	_, _ = cliutil.InfoColor.Println("Score Breakdown:")
	fmt.Printf("  Mission Alignment:  %.2f/4.00\n", analysis.Mission.Total)
	fmt.Printf("  Anti-Challenge:     %.2f/3.50\n", analysis.AntiChallenge.Total)
	fmt.Printf("  Strategic Fit:      %.2f/2.50\n", analysis.Strategic.Total)
	fmt.Println()
}

func displayUniversalScoresFromStored(completion, skillFit, timeline, reward, sustainability, avoidance float64) {
	dimensions := []struct {
		name     string
		score    float64
		maxScore float64
		desc     string
	}{
		{"Completion", completion, 2.0, "Will I finish this?"},
		{"Skill Fit", skillFit, 2.0, "Can I do this?"},
		{"Timeline", timeline, 2.0, "How long?"},
		{"Reward", reward, 2.0, "What I want?"},
		{"Sustainability", sustainability, 1.0, "Stay motivated?"},
		{"Avoidance", avoidance, 1.0, "Dodges pitfalls?"},
	}

	for _, dim := range dimensions {
		ratio := dim.score / dim.maxScore
		filledBars := int(ratio * 10)
		emptyBars := 10 - filledBars
		bar := strings.Repeat("█", filledBars) + strings.Repeat("░", emptyBars)

		var dimColor = cliutil.InfoColor
		if ratio >= 0.7 {
			dimColor = cliutil.SuccessColor
		} else if ratio < 0.4 {
			dimColor = cliutil.WarningColor
		}

		_, err := dimColor.Printf("  %-12s %s  %.1f/%.1f  %s\n",
			dim.name, bar, dim.score, dim.maxScore, dim.desc)
		if err != nil {
			log.Warn().Err(err).Msg("failed to print")
		}
	}
}
