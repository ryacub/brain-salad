package llm

import (
	"fmt"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// ConvertResultToAnalysis converts an LLM analysis result to the models.Analysis format
func ConvertResultToAnalysis(result *AnalysisResult) *models.Analysis {
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

// AnalyzeWithProviderOverride runs LLM analysis with an optional provider override
// Note: model parameter is reserved for future use when providers support model selection
func (m *Manager) AnalyzeWithProviderOverride(ideaText, provider, model string, telos *models.Telos) (*models.Analysis, error) {
	// Set provider if specified
	if provider != "" {
		if err := m.SetPrimaryProvider(provider); err != nil {
			return nil, fmt.Errorf("failed to set provider: %w", err)
		}
	}

	// Model selection is not yet supported by providers
	_ = model

	// Run LLM analysis
	result, err := m.AnalyzeWithTelos(ideaText, telos)
	if err != nil {
		return nil, err
	}

	return ConvertResultToAnalysis(result), nil
}
