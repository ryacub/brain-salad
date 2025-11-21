package dump

import (
	"encoding/json"
	"fmt"
	"strings"

	clierrors "github.com/rayyacub/telos-idea-matrix/internal/cli/errors"
	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/patterns"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
	"github.com/rayyacub/telos-idea-matrix/internal/utils"
	"github.com/rs/zerolog/log"
)

var (
	warningColor = cliutil.GetScoreColor(5.0) // Yellow color for warnings
)

// DumpContext holds dependencies for dump operations
type DumpContext struct {
	Repository interface{ Create(*models.Idea) error }
	Engine     *scoring.Engine
	Detector   *patterns.Detector
}

// runNormalDump performs standard dump with scoring engine
func runNormalDump(ideaText string, fromClipboard, toClipboard, useAI bool, provider, model string,
	ctx DumpContext, llmAnalyzer func(string, string, string) (*models.Analysis, error)) error {

	// Show clipboard info if applicable
	if fromClipboard {
		if _, err := cliutil.InfoColor.Printf("📋 Read from clipboard: %s\n", cliutil.TruncateText(ideaText, 50)); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
	}

	// Show progress
	if _, err := cliutil.InfoColor.Println("📝 Capturing idea..."); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println()

	var analysis *models.Analysis
	var err error

	if useAI {
		// Use LLM for analysis
		analysis, err = llmAnalyzer(ideaText, provider, model)
		if err != nil {
			if _, printErr := warningColor.Printf("⚠️  LLM analysis failed, falling back to rule-based: %v\n", err); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print warning")
			}
			// Fall back to rule-based scoring
			analysis, err = ctx.Engine.CalculateScore(ideaText)
			if err != nil {
				return clierrors.WrapError(err, "Failed to score idea")
			}
		}
	} else {
		// Use rule-based scoring (default)
		analysis, err = ctx.Engine.CalculateScore(ideaText)
		if err != nil {
			return clierrors.WrapError(err, "Failed to score idea")
		}
	}

	// Detect patterns
	detectedPatterns := ctx.Detector.DetectPatterns(ideaText)

	// Create idea
	idea := models.NewIdea(ideaText)
	idea.RawScore = analysis.RawScore
	idea.FinalScore = analysis.FinalScore
	idea.Recommendation = analysis.GetRecommendation()

	// Convert detected patterns to strings
	patternStrings := make([]string, len(detectedPatterns))
	for i, p := range detectedPatterns {
		patternStrings[i] = fmt.Sprintf("%s: %s", p.Name, p.Description)
	}
	idea.Patterns = patternStrings

	// Serialize analysis details
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return clierrors.WrapError(err, "Failed to serialize analysis")
	}
	idea.AnalysisDetails = string(analysisJSON)

	// Save to database
	if err := ctx.Repository.Create(idea); err != nil {
		return clierrors.WrapError(err, "Failed to save idea")
	}

	// Display results
	displayIdeaAnalysis(idea, analysis)

	// Copy result to clipboard if requested
	if toClipboard {
		summary := fmt.Sprintf("Score: %.1f/10.0\n%s\n\nIdea: %s",
			idea.FinalScore,
			idea.Recommendation,
			idea.Content)

		if err := utils.CopyToClipboard(summary); err != nil {
			if _, printErr := warningColor.Printf("⚠️  Warning: failed to copy to clipboard: %v\n", err); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print warning")
			}
		} else {
			if _, err := cliutil.SuccessColor.Println("✓ Result copied to clipboard"); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
		}
	}

	return nil
}

// displayIdeaAnalysis shows formatted analysis results
func displayIdeaAnalysis(idea *models.Idea, analysis *models.Analysis) {
	// Header
	fmt.Println(strings.Repeat("─", 80))
	if _, err := cliutil.SuccessColor.Printf("✨ Idea Analyzed (ID: %s)\n", idea.ID[:8]); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Content
	fmt.Printf("💡 %s\n\n", idea.Content)

	// Score with color coding
	scoreColor := cliutil.GetScoreColor(idea.FinalScore)
	if _, err := scoreColor.Printf("⭐ Score: %.1f/10.0\n", idea.FinalScore); err != nil {
		log.Warn().Err(err).Msg("failed to print score")
	}

	// Recommendation with emoji
	recommendationColor := cliutil.GetRecommendationColor(idea.Recommendation)
	if _, err := recommendationColor.Printf("%s\n\n", idea.Recommendation); err != nil {
		log.Warn().Err(err).Msg("failed to print recommendation")
	}

	// Mission Alignment breakdown
	fmt.Println("📊 Mission Alignment (40%):")
	fmt.Printf("  • Domain Expertise:   %.2f/1.20\n", analysis.Mission.DomainExpertise)
	fmt.Printf("  • AI Alignment:       %.2f/1.50\n", analysis.Mission.AIAlignment)
	fmt.Printf("  • Execution Support:  %.2f/0.80\n", analysis.Mission.ExecutionSupport)
	fmt.Printf("  • Revenue Potential:  %.2f/0.50\n", analysis.Mission.RevenuePotential)
	fmt.Printf("  Total: %.2f/4.00\n\n", analysis.Mission.Total)

	// Anti-Challenge Scores breakdown
	fmt.Println("🎯 Anti-Challenge Scores (35%):")
	fmt.Printf("  • Context Switching:  %.2f/1.20\n", analysis.AntiChallenge.ContextSwitching)
	fmt.Printf("  • Rapid Prototyping:  %.2f/1.00\n", analysis.AntiChallenge.RapidPrototyping)
	fmt.Printf("  • Accountability:     %.2f/0.80\n", analysis.AntiChallenge.Accountability)
	fmt.Printf("  • Income Anxiety:     %.2f/0.50\n", analysis.AntiChallenge.IncomeAnxiety)
	fmt.Printf("  Total: %.2f/3.50\n\n", analysis.AntiChallenge.Total)

	// Strategic Fit breakdown
	fmt.Println("🚀 Strategic Fit (25%):")
	fmt.Printf("  • Stack Compatibility: %.2f/1.00\n", analysis.Strategic.StackCompatibility)
	fmt.Printf("  • Shipping Habit:      %.2f/0.80\n", analysis.Strategic.ShippingHabit)
	fmt.Printf("  • Public Accountability: %.2f/0.40\n", analysis.Strategic.PublicAccountability)
	fmt.Printf("  • Revenue Testing:     %.2f/0.30\n", analysis.Strategic.RevenueTesting)
	fmt.Printf("  Total: %.2f/2.50\n\n", analysis.Strategic.Total)

	// Patterns detected
	if len(idea.Patterns) > 0 {
		if _, err := warningColor.Println("⚠️  Patterns Detected:"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		for _, pattern := range idea.Patterns {
			fmt.Printf("  • %s\n", pattern)
		}
		fmt.Println()
	}

	// Footer
	fmt.Println(strings.Repeat("─", 80))
	if _, err := cliutil.SuccessColor.Println("✅ Idea saved to database"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))
}
