package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
	"github.com/rayyacub/telos-idea-matrix/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newDumpCommand() *cobra.Command {
	var fromClipboard bool
	var toClipboard bool
	var interactive bool
	var quick bool
	var useAI bool
	var provider string
	var model string

	cmd := &cobra.Command{
		Use:   "dump <idea text>",
		Short: "Capture and analyze an idea immediately",
		Long: `Capture a new idea, analyze it against your telos, and save it to the database.
The idea will be scored and analyzed for patterns immediately.

Modes:
  Normal      - Standard analysis with scoring engine (default)
  Interactive - Step-by-step analysis with LLM and user confirmations
  Quick       - Fast capture without detailed analysis

Examples:
  tm dump "Build a SaaS product for developers"
  tm dump "Start a podcast" --use-ai
  tm dump "Learn Rust" --use-ai --provider ollama
  tm dump --interactive "Start a podcast"
  tm dump --quick "Write a blog post"
  tm dump --from-clipboard
  tm dump "Quick idea" --to-clipboard
  tm dump --quick "Fast idea capture"`,
		Args: func(cmd *cobra.Command, args []string) error {
			fromClipboard, _ := cmd.Flags().GetBool("from-clipboard")
			if !fromClipboard && len(args) < 1 {
				return fmt.Errorf("provide idea text or use --from-clipboard")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get idea text from clipboard or arguments
			var ideaText string
			if fromClipboard {
				text, err := utils.PasteFromClipboard()
				if err != nil {
					return fmt.Errorf("read clipboard: %w", err)
				}
				ideaText = strings.TrimSpace(text)
				if ideaText == "" {
					return fmt.Errorf("clipboard is empty")
				}
			} else {
				ideaText = strings.Join(args, " ")
			}

			// Route to appropriate mode
			if interactive {
				return runInteractiveDump(ideaText, provider)
			}

			if quick {
				return runQuickDump(ideaText, toClipboard)
			}

			// Normal dump
			return runNormalDump(ideaText, fromClipboard, toClipboard, useAI, provider, model)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode with step-through analysis")
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, "Quick mode with rule-based analysis")
	cmd.Flags().BoolVar(&useAI, "use-ai", false, "Use LLM analysis (requires Ollama or API keys)")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "LLM provider to use (ollama|openai|claude|rule_based)")
	cmd.Flags().StringVar(&model, "model", "", "LLM model to use")
	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read idea from clipboard")
	cmd.Flags().BoolVar(&toClipboard, "to-clipboard", false, "Copy result to clipboard")

	return cmd
}

func runNormalDump(ideaText string, fromClipboard, toClipboard, useAI bool, provider, model string) error {
	// Show clipboard info if applicable
	if fromClipboard {
		if _, err := infoColor.Printf("📋 Read from clipboard: %s\n", truncateText(ideaText, 50)); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
	}

	// Show progress
	if _, err := infoColor.Println("📝 Capturing idea..."); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println()

	var analysis *models.Analysis
	var err error

	if useAI {
		// Use LLM for analysis
		analysis, err = runLLMAnalysis(ideaText, provider, model)
		if err != nil {
			if _, printErr := warningColor.Printf("⚠️  LLM analysis failed, falling back to rule-based: %v\n", err); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print warning")
			}
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
		return fmt.Errorf("failed to serialize analysis: %w", err)
	}
	idea.AnalysisDetails = string(analysisJSON)

	// Save to database
	if err := ctx.Repository.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
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
			if _, err := successColor.Println("✓ Result copied to clipboard"); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
		}
	}

	return nil
}

// runLLMAnalysis performs LLM-based analysis and converts result to models.Analysis
func runLLMAnalysis(ideaText, provider, model string) (*models.Analysis, error) {
	return runLLMAnalysisWithProvider(ideaText, provider, model, ctx.LLMManager, ctx.Telos)
}

func displayIdeaAnalysis(idea *models.Idea, analysis *models.Analysis) {
	// Header
	fmt.Println(strings.Repeat("─", 80))
	if _, err := successColor.Printf("✨ Idea Analyzed (ID: %s)\n", idea.ID[:8]); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Content
	fmt.Printf("💡 %s\n\n", idea.Content)

	// Score with color coding
	scoreColor := getScoreColor(idea.FinalScore)
	if _, err := scoreColor.Printf("⭐ Score: %.1f/10.0\n", idea.FinalScore); err != nil {
		log.Warn().Err(err).Msg("failed to print score")
	}

	// Recommendation with emoji
	recommendationColor := getRecommendationColor(idea.Recommendation)
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
	if _, err := successColor.Println("✅ Idea saved to database"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))
}

func getScoreColor(score float64) *color.Color {
	switch {
	case score >= 8.5:
		return color.New(color.FgGreen, color.Bold)
	case score >= 7.0:
		return color.New(color.FgGreen)
	case score >= 5.0:
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgRed)
	}
}

func getRecommendationColor(recommendation string) *color.Color {
	if strings.Contains(recommendation, "🔥") {
		return color.New(color.FgGreen, color.Bold)
	} else if strings.Contains(recommendation, "✅") {
		return color.New(color.FgGreen)
	} else if strings.Contains(recommendation, "⚠️") {
		return color.New(color.FgYellow)
	}
	return color.New(color.FgRed)
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// runQuickDump performs fast rule-based analysis without LLM
func runQuickDump(content string, toClipboard bool) error {
	start := time.Now()

	// Step 1: Load telos (optional for rule-based scoring)
	telosContent, err := loadTelosQuiet()
	if err != nil {
		// Continue without telos if not available
		telosContent = ""
	}

	// Step 2: Rule-based scoring
	scorer := scoring.NewRuleBasedScorer()
	score := scorer.Score(content, telosContent)

	// Step 3: Pattern detection
	patterns := detectBasicPatterns(content, telosContent)

	// Step 4: Generate recommendation
	recommendation := generateRecommendation(score, patterns)

	// Step 5: Create idea with quick analysis marker
	reasoning := "Quick analysis (rule-based, no LLM)\nScore based on: keyword matching, length, telos alignment"

	idea := models.NewIdea(content)
	idea.FinalScore = score
	idea.RawScore = score
	idea.Patterns = patterns
	idea.Recommendation = recommendation
	idea.AnalysisDetails = reasoning

	// Step 6: Save to database
	if err := ctx.Repository.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
	}

	elapsed := time.Since(start)

	// Step 7: Display results
	fmt.Println(strings.Repeat("─", 80))
	if _, err := successColor.Printf("✨ Quick Analysis Complete (ID: %s)\n", idea.ID[:8]); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
	fmt.Printf("💡 %s\n\n", idea.Content)

	scoreColor := getScoreColor(idea.FinalScore)
	if _, err := scoreColor.Printf("⭐ Score: %.1f/10.0 (rule-based)\n", score); err != nil {
		log.Warn().Err(err).Msg("failed to print score")
	}

	recommendationColor := getRecommendationColor(recommendation)
	if _, err := recommendationColor.Printf("%s\n\n", recommendation); err != nil {
		log.Warn().Err(err).Msg("failed to print recommendation")
	}

	if len(patterns) > 0 {
		fmt.Println("🏷️  Patterns:")
		for _, pattern := range patterns {
			fmt.Printf("  • %s\n", pattern)
		}
		fmt.Println()
	}

	fmt.Printf("⚡ Completed in %v\n\n", elapsed)

	fmt.Println(strings.Repeat("─", 80))
	if _, err := successColor.Println("✅ Idea saved to database"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	if _, err := infoColor.Println("💡 Tip: Use 'tm analyze " + idea.ID[:8] + "' to run full LLM analysis later"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println(strings.Repeat("─", 80))

	// Copy result to clipboard if requested
	if toClipboard {
		summary := fmt.Sprintf("Score: %.1f/10.0 (rule-based)\n%s\n\nIdea: %s",
			idea.FinalScore,
			idea.Recommendation,
			idea.Content)

		if err := utils.CopyToClipboard(summary); err != nil {
			if _, printErr := warningColor.Printf("⚠️  Warning: failed to copy to clipboard: %v\n", err); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print warning")
			}
		} else {
			if _, err := successColor.Println("✓ Result copied to clipboard"); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
		}
	}

	return nil
}

// loadTelosQuiet loads telos without errors
func loadTelosQuiet() (string, error) {
	telosPath := os.Getenv("TELOS_PATH")
	if telosPath == "" {
		homeDir, _ := os.UserHomeDir()
		telosPath = filepath.Join(homeDir, ".telos", "telos.md")
	}

	data, err := os.ReadFile(telosPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// detectBasicPatterns performs simple pattern detection
func detectBasicPatterns(content, _ string) []string {
	patterns := []string{}
	contentLower := strings.ToLower(content)

	// Pattern keywords with their categories
	patternKeywords := map[string]string{
		"innovation":  "innovation",
		"innovate":    "innovation",
		"novel":       "innovation",
		"new":         "innovation",
		"sustain":     "sustainability",
		"green":       "sustainability",
		"environment": "sustainability",
		"eco":         "sustainability",
		"impact":      "impact",
		"improve":     "impact",
		"benefit":     "impact",
		"help":        "impact",
		"scale":       "scalability",
		"grow":        "scalability",
		"expand":      "scalability",
		"revenue":     "revenue",
		"profit":      "revenue",
		"monetize":    "revenue",
		"income":      "revenue",
		"cost":        "cost-reduction",
		"save":        "cost-reduction",
		"efficient":   "efficiency",
		"optimize":    "efficiency",
		"automate":    "automation",
		"automatic":   "automation",
		"ai":          "ai-ml",
		"machine":     "ai-ml",
		"learning":    "ai-ml",
		"mobile":      "mobile",
		"app":         "mobile",
		"web":         "web",
		"cloud":       "cloud",
		"saas":        "saas",
		"product":     "product",
		"service":     "service",
	}

	// Track found patterns to avoid duplicates
	found := make(map[string]bool)

	for keyword, pattern := range patternKeywords {
		if strings.Contains(contentLower, keyword) && !found[pattern] {
			patterns = append(patterns, pattern)
			found[pattern] = true
		}
	}

	// If no patterns found, add a generic one
	if len(patterns) == 0 {
		patterns = append(patterns, "general")
	}

	// Limit to top 5 patterns
	if len(patterns) > 5 {
		patterns = patterns[:5]
	}

	return patterns
}

// generateRecommendation creates a recommendation based on score and patterns
func generateRecommendation(score float64, patterns []string) string {
	// High score -> pursue
	if score >= 7.0 {
		return "🔥 PURSUE - Strong potential"
	}

	// Low score -> defer
	if score < 4.0 {
		return "❌ DEFER - Low alignment"
	}

	// Medium score -> review (check patterns for tie-breaker)
	strongPatterns := []string{"innovation", "impact", "scalability", "revenue"}
	for _, pattern := range patterns {
		for _, strong := range strongPatterns {
			if pattern == strong {
				return "✅ PURSUE - Good potential with strong patterns"
			}
		}
	}

	return "⚠️ REVIEW - Needs more evaluation"
}

// runInteractiveDump performs step-by-step interactive analysis using LLM
func runInteractiveDump(ideaText string, providerName string) error {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║       Interactive Idea Analysis                           ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Step 1: Display and confirm idea
	fmt.Println("STEP 1: Idea Content")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(ideaText)
	fmt.Println()

	if !confirm("Continue to telos loading?") {
		if _, err := infoColor.Println("Analysis cancelled."); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		return nil
	}

	// Step 2: Load and display telos
	fmt.Println("\nSTEP 2: Loading Telos")
	fmt.Println(strings.Repeat("─", 60))

	if ctx.Telos == nil {
		return fmt.Errorf("telos not loaded")
	}

	// Display telos summary
	fmt.Printf("Mission Elements: %d\n", len(ctx.Telos.Missions))
	fmt.Printf("Failure Patterns: %d\n", len(ctx.Telos.FailurePatterns))
	stackCount := len(ctx.Telos.Stack.Primary) + len(ctx.Telos.Stack.Secondary)
	fmt.Printf("Stack Items: %d\n", stackCount)
	fmt.Println()

	if !confirm("Continue with this telos?") {
		if _, err := infoColor.Println("Analysis cancelled."); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		return nil
	}

	// Step 3: Select and display provider
	fmt.Println("\nSTEP 3: LLM Provider Selection")
	fmt.Println(strings.Repeat("─", 60))

	// Create LLM manager
	manager := llm.NewManager(nil)

	var provider llm.Provider
	if providerName == "" {
		provider = selectProviderInteractive(manager)
		if provider == nil {
			return fmt.Errorf("no provider selected")
		}
	} else {
		if err := manager.SetPrimaryProvider(providerName); err != nil {
			return fmt.Errorf("provider not found: %s", providerName)
		}
		provider = manager.GetPrimaryProvider()
	}

	fmt.Printf("Selected Provider: %s\n", provider.Name())
	fmt.Printf("Status: %s\n", getProviderStatus(provider))
	fmt.Println()

	if !confirm("Continue with this provider?") {
		if _, err := infoColor.Println("Analysis cancelled."); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		return nil
	}

	// Step 4: Run analysis with progress indicator
	fmt.Println("\nSTEP 4: Running Analysis")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("Sending request to LLM...")
	fmt.Println("(This may take 10-30 seconds depending on the provider)")
	fmt.Println()

	startTime := time.Now()

	result, err := manager.AnalyzeWithTelos(ideaText, ctx.Telos)

	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if _, err := successColor.Printf("✓ Analysis complete (took %v)\n", duration); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println()

	// Step 5: Display detailed results
	fmt.Println("STEP 5: Analysis Results")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	displayInteractiveAnalysisResults(result)

	if !confirm("Save this idea?") {
		if _, err := infoColor.Println("Idea not saved."); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		return nil
	}

	// Step 6: Save idea
	fmt.Println("\nSTEP 6: Saving Idea")
	fmt.Println(strings.Repeat("─", 60))

	idea := models.NewIdea(ideaText)
	idea.RawScore = result.FinalScore
	idea.FinalScore = result.FinalScore
	idea.Recommendation = result.Recommendation
	idea.Patterns = []string{} // LLM result doesn't have patterns in the same format

	// Serialize analysis details
	analysisJSON, err := json.Marshal(result)
	if err != nil {
		if _, printErr := warningColor.Printf("⚠️  Warning: failed to serialize analysis: %v\n", err); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print warning")
		}
	} else {
		idea.AnalysisDetails = string(analysisJSON)
	}

	if err := ctx.Repository.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
	}

	if _, err := successColor.Printf("✓ Idea saved successfully\n"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Printf("  ID: %s\n", idea.ID[:8])
	fmt.Printf("  Score: %.1f/10\n", idea.FinalScore)
	fmt.Printf("  Recommendation: %s\n", idea.Recommendation)
	fmt.Printf("  Created: %s\n", idea.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║       Interactive Analysis Complete                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	return nil
}

// selectProviderInteractive allows user to choose from available providers
func selectProviderInteractive(manager *llm.Manager) llm.Provider {
	providers := manager.GetAvailableProviders()

	if len(providers) == 0 {
		if _, err := errorColor.Println("No providers available!"); err != nil {
			log.Warn().Err(err).Msg("failed to print error message")
		}
		return nil
	}

	if len(providers) == 1 {
		if _, err := infoColor.Printf("Using only available provider: %s\n", providers[0].Name()); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		return providers[0]
	}

	fmt.Println("Available providers:")
	for i, p := range providers {
		status := "✓"
		if !p.IsAvailable() {
			status = "✗"
		}
		fmt.Printf("  %d. %s %s\n", i+1, status, p.Name())
	}
	fmt.Println()
	fmt.Print("Select provider (number): ")

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(providers) {
		if _, printErr := warningColor.Println("Invalid choice, using default provider"); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print warning")
		}
		return providers[0]
	}

	return providers[choice-1]
}

// getProviderStatus returns a human-readable status string
func getProviderStatus(provider llm.Provider) string {
	if provider.IsAvailable() {
		return "✓ Available and ready"
	}
	return "✗ Not available"
}

// displayInteractiveAnalysisResults shows formatted analysis results
func displayInteractiveAnalysisResults(result *llm.AnalysisResult) {
	// Score with visual indicator
	fmt.Printf("Score: %.1f/10 ", result.FinalScore)
	fmt.Println(getScoreIndicator(result.FinalScore))
	fmt.Println()

	// Recommendation with indicator
	fmt.Printf("Recommendation: %s %s\n",
		result.Recommendation,
		getRecommendationIndicator(result.Recommendation))
	fmt.Println()

	// Score breakdown
	fmt.Println("Score Breakdown:")
	fmt.Printf("  • Mission Alignment:  %.2f/4.00 (40%%)\n", result.Scores.MissionAlignment)
	fmt.Printf("  • Anti-Challenge:     %.2f/3.50 (35%%)\n", result.Scores.AntiChallenge)
	fmt.Printf("  • Strategic Fit:      %.2f/2.50 (25%%)\n", result.Scores.StrategicFit)
	fmt.Println()

	// Explanations
	if len(result.Explanations) > 0 {
		fmt.Println("Detailed Explanations:")
		for category, explanation := range result.Explanations {
			categoryTitle := formatCategoryTitle(category)
			fmt.Printf("\n%s:\n", categoryTitle)
			fmt.Println(wrapTextSimple(explanation, 58))
		}
		fmt.Println()
	}
}

// getScoreIndicator returns a visual bar for the score
func getScoreIndicator(score float64) string {
	bars := int(score)
	if bars > 10 {
		bars = 10
	}
	if bars < 0 {
		bars = 0
	}
	filled := strings.Repeat("█", bars)
	empty := strings.Repeat("░", 10-bars)
	return fmt.Sprintf("[%s%s]", filled, empty)
}

// getRecommendationIndicator returns an indicator for the recommendation
func getRecommendationIndicator(rec string) string {
	recUpper := strings.ToUpper(rec)
	if strings.Contains(recUpper, "PURSUE") || strings.Contains(recUpper, "STRONG") {
		return "✓ (Go for it!)"
	}
	if strings.Contains(recUpper, "CONSIDER") || strings.Contains(recUpper, "MODERATE") {
		return "⏸ (Consider carefully)"
	}
	if strings.Contains(recUpper, "AVOID") || strings.Contains(recUpper, "WEAK") || strings.Contains(recUpper, "DEFER") {
		return "✗ (Skip this)"
	}
	return "?"
}

// wrapTextSimple wraps text to specified width
func wrapTextSimple(text string, width int) string {
	if len(text) <= width {
		return "  " + text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	result.WriteString("  ") // Initial indent
	lineLen = 2

	for i, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > width && lineLen > 2 {
			result.WriteString("\n  ")
			lineLen = 2
		} else if i > 0 {
			result.WriteString(" ")
			lineLen++
		}

		result.WriteString(word)
		lineLen += wordLen
	}

	return result.String()
}
