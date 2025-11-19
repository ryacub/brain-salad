package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

var (
	analyzeID   string
	analyzeLast bool
)

func newAnalyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [idea text]",
		Short: "Analyze an existing or new idea",
		Long: `Analyze an idea in detail - either by ID, the last saved idea, or new text.
Shows complete scoring breakdown.

Examples:
  tm analyze "Build a new product"    # Analyze new text
  tm analyze --id abc123              # Analyze by ID
  tm analyze --last                   # Analyze last saved idea`,
		RunE: runAnalyze,
	}

	cmd.Flags().StringVar(&analyzeID, "id", "", "Idea ID to analyze")
	cmd.Flags().BoolVar(&analyzeLast, "last", false, "Analyze the last saved idea")

	return cmd
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	var idea *models.Idea
	var analysis *models.Analysis

	// Determine what to analyze
	if analyzeID != "" {
		// Analyze existing idea by ID
		fetchedIdea, err := ctx.Repository.GetByID(analyzeID)
		if err != nil {
			return fmt.Errorf("failed to fetch idea: %w", err)
		}
		idea = fetchedIdea

		// Deserialize analysis
		if idea.AnalysisDetails != "" {
			var a models.Analysis
			if err := json.Unmarshal([]byte(idea.AnalysisDetails), &a); err != nil {
				return fmt.Errorf("failed to parse analysis: %w", err)
			}
			analysis = &a
		}

	} else if analyzeLast {
		// Analyze last saved idea
		limit := 1
		ideas, err := ctx.Repository.List(database.ListOptions{
			Status:  "active",
			OrderBy: "created_at DESC",
			Limit:   &limit,
		})
		if err != nil {
			return fmt.Errorf("failed to list ideas: %w", err)
		}
		if len(ideas) == 0 {
			return fmt.Errorf("no ideas found")
		}
		idea = ideas[0]

		// Deserialize analysis
		if idea.AnalysisDetails != "" {
			var a models.Analysis
			if err := json.Unmarshal([]byte(idea.AnalysisDetails), &a); err != nil {
				return fmt.Errorf("failed to parse analysis: %w", err)
			}
			analysis = &a
		}

	} else if len(args) > 0 {
		// Analyze new text (don't save)
		ideaText := strings.Join(args, " ")

		a, err := ctx.Engine.CalculateScore(ideaText)
		if err != nil {
			return fmt.Errorf("failed to score idea: %w", err)
		}
		analysis = a

		// Create temporary idea for display
		idea = models.NewIdea(ideaText)
		idea.FinalScore = analysis.FinalScore
		idea.Recommendation = analysis.GetRecommendation()

		detectedPatterns := ctx.Detector.DetectPatterns(ideaText)
		patternStrings := make([]string, len(detectedPatterns))
		for i, p := range detectedPatterns {
			patternStrings[i] = fmt.Sprintf("%s: %s", p.Name, p.Description)
		}
		idea.Patterns = patternStrings

	} else {
		return fmt.Errorf("provide idea text, --id, or --last flag")
	}

	// Display full analysis
	displayIdeaAnalysis(idea, analysis)

	return nil
}
