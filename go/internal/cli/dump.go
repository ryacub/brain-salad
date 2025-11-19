package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/utils"
	"github.com/spf13/cobra"
)

func newDumpCommand() *cobra.Command {
	var fromClipboard bool
	var toClipboard bool
	var interactive bool
	var quick bool
	var provider string

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
  tm dump --interactive "Start a podcast"
  tm dump --quick "Write a blog post"
  tm dump --from-clipboard
  tm dump "Quick idea" --to-clipboard`,
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
				return runQuickDump(ideaText)
			}

			// Normal dump
			return runNormalDump(ideaText, fromClipboard, toClipboard)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode with step-through analysis")
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, "Quick mode without LLM analysis")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "LLM provider to use (interactive mode only)")
	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read idea from clipboard")
	cmd.Flags().BoolVar(&toClipboard, "to-clipboard", false, "Copy result to clipboard")

	return cmd
}

func runNormalDump(ideaText string, fromClipboard, toClipboard bool) error {
	// Show clipboard info if applicable
	if fromClipboard {
		infoColor.Printf("📋 Read from clipboard: %s\n", truncateText(ideaText, 50))
	}

	// Show progress
	infoColor.Println("📝 Capturing idea...")
	fmt.Println()

	// Calculate score
	analysis, err := ctx.Engine.CalculateScore(ideaText)
	if err != nil {
		return fmt.Errorf("failed to score idea: %w", err)
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
			warningColor.Printf("⚠️  Warning: failed to copy to clipboard: %v\n", err)
		} else {
			successColor.Println("✓ Result copied to clipboard")
		}
	}

	return nil
}

func displayIdeaAnalysis(idea *models.Idea, analysis *models.Analysis) {
	// Header
	fmt.Println(strings.Repeat("─", 80))
	successColor.Printf("✨ Idea Analyzed (ID: %s)\n", idea.ID[:8])
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Content
	fmt.Printf("💡 %s\n\n", idea.Content)

	// Score with color coding
	scoreColor := getScoreColor(idea.FinalScore)
	scoreColor.Printf("⭐ Score: %.1f/10.0\n", idea.FinalScore)

	// Recommendation with emoji
	recommendationColor := getRecommendationColor(idea.Recommendation)
	recommendationColor.Printf("%s\n\n", idea.Recommendation)

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
		warningColor.Println("⚠️  Patterns Detected:")
		for _, pattern := range idea.Patterns {
			fmt.Printf("  • %s\n", pattern)
		}
		fmt.Println()
	}

	// Footer
	fmt.Println(strings.Repeat("─", 80))
	successColor.Println("✅ Idea saved to database")
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

// runQuickDump performs a quick capture without detailed analysis
func runQuickDump(ideaText string) error {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║       Quick Idea Capture                                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create idea without analysis
	idea := models.NewIdea(ideaText)
	idea.RawScore = 0.0
	idea.FinalScore = 0.0
	idea.Recommendation = "Not analyzed (quick mode)"
	idea.Patterns = []string{}

	// Save to database
	if err := ctx.Repository.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
	}

	// Display result
	fmt.Printf("💡 %s\n\n", idea.Content)
	successColor.Printf("✓ Idea saved (ID: %s)\n", idea.ID[:8])
	infoColor.Println("ℹ️  Run 'tm analyze' to analyze this idea later")
	fmt.Println()

	return nil
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
		infoColor.Println("Analysis cancelled.")
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
		infoColor.Println("Analysis cancelled.")
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
		infoColor.Println("Analysis cancelled.")
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

	successColor.Printf("✓ Analysis complete (took %v)\n", duration)
	fmt.Println()

	// Step 5: Display detailed results
	fmt.Println("STEP 5: Analysis Results")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	displayInteractiveAnalysisResults(result)

	if !confirm("Save this idea?") {
		infoColor.Println("Idea not saved.")
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
		warningColor.Printf("⚠️  Warning: failed to serialize analysis: %v\n", err)
	} else {
		idea.AnalysisDetails = string(analysisJSON)
	}

	if err := ctx.Repository.Create(idea); err != nil {
		return fmt.Errorf("failed to save idea: %w", err)
	}

	successColor.Printf("✓ Idea saved successfully\n")
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
		errorColor.Println("No providers available!")
		return nil
	}

	if len(providers) == 1 {
		infoColor.Printf("Using only available provider: %s\n", providers[0].Name())
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
		warningColor.Println("Invalid choice, using default provider")
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
