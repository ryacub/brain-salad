package dump

import (
	"fmt"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// runLLMAnalysisWithProvider runs LLM analysis with an optional provider override
func runLLMAnalysisWithProvider(ideaText, provider, model string, manager *llm.Manager, telos *models.Telos) (*models.Analysis, error) {
	// Set provider if specified
	if provider != "" {
		if err := manager.SetPrimaryProvider(provider); err != nil {
			return nil, fmt.Errorf("failed to set provider: %w", err)
		}
	}

	// TODO: Support model selection when LLM providers support it
	_ = model

	// Run LLM analysis
	result, err := manager.AnalyzeWithTelos(ideaText, telos)
	if err != nil {
		return nil, err
	}

	return convertLLMResultToAnalysis(result), nil
}

// convertLLMResultToAnalysis converts an LLM analysis result to the models.Analysis format
func convertLLMResultToAnalysis(result *llm.AnalysisResult) *models.Analysis {
	return &models.Analysis{
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
			ContextSwitching: result.Scores.AntiChallenge * 0.343,
			RapidPrototyping: result.Scores.AntiChallenge * 0.286,
			Accountability:   result.Scores.AntiChallenge * 0.229,
			IncomeAnxiety:    result.Scores.AntiChallenge * 0.143,
		},
		Strategic: models.StrategicScores{
			Total: result.Scores.StrategicFit,
			// Distribute proportionally
			StackCompatibility:   result.Scores.StrategicFit * 0.40,
			ShippingHabit:        result.Scores.StrategicFit * 0.32,
			PublicAccountability: result.Scores.StrategicFit * 0.16,
			RevenueTesting:       result.Scores.StrategicFit * 0.12,
		},
	}
}
