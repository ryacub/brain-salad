package cli

import (
	"fmt"
	"strings"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	llmProvider   string
	llmNoFallback bool
	llmVerbose    bool
)

func newAnalyzeLLMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze-llm <idea>",
		Short: "Analyze an idea using LLM provider",
		Long: `Analyze an idea against your telos using an LLM provider.

The system will automatically use the best available provider unless
you specify one with --provider.

Available Providers:
  • ollama      - Local Ollama instance (configure via OllamaBaseURL)
  • claude      - Anthropic Claude API (requires ANTHROPIC_API_KEY env var)
  • openai      - OpenAI GPT models (requires OPENAI_API_KEY env var)
  • custom      - Custom HTTP endpoint (requires CUSTOM_LLM_ENDPOINT env var)
  • rule_based  - Rule-based scoring engine (always available)

Provider Priority: ollama → claude → openai → custom → rule_based

Examples:
  tm analyze-llm "Build a mobile app"
  tm analyze-llm "Start a podcast" --provider claude
  tm analyze-llm "Learn Rust" --provider openai --verbose
  tm analyze-llm "Write a book" --provider rule_based`,
		Args: cobra.ExactArgs(1),
		RunE: runAnalyzeLLM,
	}

	cmd.Flags().StringVarP(&llmProvider, "provider", "p", "", "LLM provider (ollama|claude|openai|custom|rule_based)")
	cmd.Flags().BoolVar(&llmNoFallback, "no-fallback", false, "Disable fallback to other providers")
	cmd.Flags().BoolVarP(&llmVerbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func runAnalyzeLLM(cmd *cobra.Command, args []string) error {
	// Create LLM manager
	manager := llm.NewManager(nil)

	// Set provider if specified
	if llmProvider != "" {
		if err := manager.SetPrimaryProvider(llmProvider); err != nil {
			return fmt.Errorf("failed to set provider: %w", err)
		}
	}

	// Configure fallback
	if llmNoFallback {
		manager.EnableFallback(false)
	}

	// Get current provider
	currentProvider := manager.GetPrimaryProvider()
	if llmVerbose {
		if _, err := infoColor.Printf("🔧 Using provider: %s\n", currentProvider.Name()); err != nil {
			log.Warn().Err(err).Msg("failed to print provider info")
		}
		if !llmNoFallback {
			if _, err := infoColor.Println("🔄 Fallback enabled"); err != nil {
				log.Warn().Err(err).Msg("failed to print fallback info")
			}
		}
		fmt.Println()
	}

	// Analyze
	ideaText := args[0]
	result, err := manager.AnalyzeWithTelos(ideaText, ctx.Telos)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Display result
	displayLLMAnalysisResult(ideaText, result, llmVerbose)

	return nil
}

func displayLLMAnalysisResult(ideaText string, result *llm.AnalysisResult, verbose bool) {
	// Header
	fmt.Println(strings.Repeat("─", 80))
	if _, err := successColor.Printf("✨ LLM Analysis Complete\n"); err != nil {
		log.Warn().Err(err).Msg("failed to print header")
	}
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Content
	fmt.Printf("💡 %s\n\n", ideaText)

	// Provider info
	if verbose {
		if _, err := infoColor.Printf("🤖 Provider: %s\n", result.Provider); err != nil {
			log.Warn().Err(err).Msg("failed to print provider")
		}
		if _, err := infoColor.Printf("⏱️  Duration: %v\n", result.Duration); err != nil {
			log.Warn().Err(err).Msg("failed to print duration")
		}
		if result.FromCache {
			if _, err := infoColor.Println("📦 Result from cache"); err != nil {
				log.Warn().Err(err).Msg("failed to print cache info")
			}
		}
		fmt.Println()
	}

	// Score with color coding
	scoreColor := getScoreColor(result.FinalScore)
	if _, err := scoreColor.Printf("⭐ Score: %.1f/10.0\n", result.FinalScore); err != nil {
		log.Warn().Err(err).Msg("failed to print score")
	}

	// Recommendation with emoji
	recommendationColor := getRecommendationColor(result.Recommendation)
	recommendationEmoji := getRecommendationEmoji(result.Recommendation)
	if _, err := recommendationColor.Printf("%s %s\n\n", recommendationEmoji, result.Recommendation); err != nil {
		log.Warn().Err(err).Msg("failed to print recommendation")
	}

	// Score breakdown
	fmt.Println("📊 Score Breakdown:")
	fmt.Printf("  • Mission Alignment:  %.2f/4.00 (40%%)\n", result.Scores.MissionAlignment)
	fmt.Printf("  • Anti-Challenge:     %.2f/3.50 (35%%)\n", result.Scores.AntiChallenge)
	fmt.Printf("  • Strategic Fit:      %.2f/2.50 (25%%)\n", result.Scores.StrategicFit)
	fmt.Println()

	// Explanations
	if len(result.Explanations) > 0 {
		fmt.Println("💭 Detailed Explanations:")
		for category, explanation := range result.Explanations {
			categoryTitle := formatCategoryTitle(category)
			fmt.Printf("\n%s:\n", categoryTitle)
			fmt.Printf("%s\n", wrapText(explanation, 76, "  "))
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 80))
}

func formatCategoryTitle(category string) string {
	switch category {
	case "mission_alignment":
		return "📊 Mission Alignment"
	case "anti_challenge":
		return "🎯 Anti-Challenge"
	case "strategic_fit":
		return "🚀 Strategic Fit"
	default:
		// Convert underscores to spaces and title case
		title := strings.ReplaceAll(category, "_", " ")
		return strings.Title(title)
	}
}

func getRecommendationEmoji(recommendation string) string {
	rec := strings.ToUpper(recommendation)
	if strings.Contains(rec, "PURSUE") || strings.Contains(rec, "STRONG") {
		return "✅"
	}
	if strings.Contains(rec, "CONSIDER") || strings.Contains(rec, "MODERATE") {
		return "⚠️"
	}
	if strings.Contains(rec, "AVOID") || strings.Contains(rec, "WEAK") {
		return "❌"
	}
	return "ℹ️"
}

func wrapText(text string, width int, indent string) string {
	// Simple word wrapping
	words := strings.Fields(text)
	if len(words) == 0 {
		return indent
	}

	var result strings.Builder
	result.WriteString(indent)
	lineLen := len(indent)

	for i, word := range words {
		wordLen := len(word)
		if lineLen+wordLen+1 > width && lineLen > len(indent) {
			result.WriteString("\n" + indent)
			lineLen = len(indent)
		} else if i > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += wordLen
	}

	return result.String()
}
