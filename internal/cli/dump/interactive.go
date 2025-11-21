package dump

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rs/zerolog/log"
)

var (
	successColor = color.New(color.FgGreen, color.Bold)
	infoColor    = color.New(color.FgCyan)
)

// runInteractiveDump performs step-by-step interactive analysis using LLM
func runInteractiveDump(ideaText string, providerName string, telos *models.Telos, repo interface{ Create(*models.Idea) error }) error {
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

	if telos == nil {
		return fmt.Errorf("telos not loaded")
	}

	// Display telos summary
	fmt.Printf("Mission Elements: %d\n", len(telos.Missions))
	fmt.Printf("Failure Patterns: %d\n", len(telos.FailurePatterns))
	stackCount := len(telos.Stack.Primary) + len(telos.Stack.Secondary)
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

	result, err := manager.AnalyzeWithTelos(ideaText, telos)

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
		warningColor := getScoreColor(5.0)
		if _, printErr := warningColor.Printf("⚠️  Warning: failed to serialize analysis: %v\n", err); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print warning")
		}
	} else {
		idea.AnalysisDetails = string(analysisJSON)
	}

	if err := repo.Create(idea); err != nil {
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
		errorColor := color.New(color.FgRed, color.Bold)
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
		warningColor := color.New(color.FgYellow)
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
