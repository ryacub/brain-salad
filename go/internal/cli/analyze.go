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
	var useAI bool
	var provider string
	var model string

	cmd := &cobra.Command{
		Use:   "analyze [idea text]",
		Short: "Analyze an existing or new idea",
		Long: `Analyze an idea in detail - either by ID, the last saved idea, or new text.
Shows complete scoring breakdown.

Examples:
  tm analyze "Build a new product"               # Analyze new text (rule-based)
  tm analyze "Build a new product" --use-ai      # Analyze with LLM
  tm analyze --id abc123                         # Analyze by ID
  tm analyze --last                              # Analyze last saved idea
  tm analyze "Start a podcast" --use-ai --provider ollama`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(cmd, args, useAI, provider, model)
		},
	}

	cmd.Flags().StringVar(&analyzeID, "id", "", "Idea ID to analyze")
	cmd.Flags().BoolVar(&analyzeLast, "last", false, "Analyze the last saved idea")
	cmd.Flags().BoolVar(&useAI, "use-ai", false, "Use LLM analysis (requires Ollama or API keys)")
	cmd.Flags().StringVar(&provider, "provider", "", "LLM provider to use (ollama|openai|claude|rule_based)")
	cmd.Flags().StringVar(&model, "model", "", "LLM model to use")

	return cmd
}

func runAnalyze(cmd *cobra.Command, args []string, useAI bool, provider, model string) error {
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

		var err error
		if useAI {
			// Use LLM for analysis
			analysis, err = runLLMAnalysisForAnalyze(ideaText, provider, model)
			if err != nil {
				warningColor.Printf("⚠️  LLM analysis failed, falling back to rule-based: %v\n", err)
				// Fall back to rule-based scoring
				analysis, err = ctx.Engine.CalculateScore(ideaText)
				if err != nil {
					return fmt.Errorf("failed to score idea: %w", err)
				}
			}
		} else {
			// Use rule-based scoring (default)
			analysis, err = ctx.Engine.CalculateScore(ideaText)
			if err != nil {
				return fmt.Errorf("failed to score idea: %w", err)
			}
		}

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

// runLLMAnalysisForAnalyze performs LLM-based analysis for the analyze command
func runLLMAnalysisForAnalyze(ideaText, provider, model string) (*models.Analysis, error) {
	// Set provider if specified
	if provider != "" {
		if err := ctx.LLMManager.SetPrimaryProvider(provider); err != nil {
			return nil, fmt.Errorf("failed to set provider: %w", err)
		}
	}

	// TODO: Support model selection when LLM providers support it
	_ = model

	// Run LLM analysis
	result, err := ctx.LLMManager.AnalyzeWithTelos(ideaText, ctx.Telos)
	if err != nil {
		return nil, err
	}

	// Convert LLM result to models.Analysis format
	analysis := &models.Analysis{
		RawScore:   result.FinalScore,
		FinalScore: result.FinalScore,
		Mission: models.MissionScores{
			Total: result.Scores.MissionAlignment,
			// LLM doesn't break down into sub-scores, so distribute proportionally
			DomainExpertise:  result.Scores.MissionAlignment * 0.30,
			AIAlignment:      result.Scores.MissionAlignment * 0.375,
			ExecutionSupport: result.Scores.MissionAlignment * 0.20,
			RevenuePotential: result.Scores.MissionAlignment * 0.125,
		},
		AntiChallenge: models.AntiChallengeScores{
			Total: result.Scores.AntiChallenge,
			// Distribute proportionally
			ContextSwitching:  result.Scores.AntiChallenge * 0.343,
			RapidPrototyping:  result.Scores.AntiChallenge * 0.286,
			Accountability:    result.Scores.AntiChallenge * 0.229,
			IncomeAnxiety:     result.Scores.AntiChallenge * 0.143,
		},
		Strategic: models.StrategicScores{
			Total: result.Scores.StrategicFit,
			// Distribute proportionally
			StackCompatibility:     result.Scores.StrategicFit * 0.40,
			ShippingHabit:          result.Scores.StrategicFit * 0.32,
			PublicAccountability:   result.Scores.StrategicFit * 0.16,
			RevenueTesting:         result.Scores.StrategicFit * 0.12,
		},
	}

	return analysis, nil
}
