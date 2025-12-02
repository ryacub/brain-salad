package cli

import (
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// runLLMAnalysisWithProvider runs LLM analysis with an optional provider override
func runLLMAnalysisWithProvider(ideaText, provider, model string, manager *llm.Manager, telos *models.Telos) (*models.Analysis, error) {
	return manager.AnalyzeWithProviderOverride(ideaText, provider, model, telos)
}
