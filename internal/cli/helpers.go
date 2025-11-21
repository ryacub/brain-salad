package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rs/zerolog/log"
)

// truncateText truncates text to specified length with ellipsis
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// getScoreColor returns a color based on the score value
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

// getRecommendationColor returns a color based on the recommendation text
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

// displayIdeaAnalysis shows formatted analysis results
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

// confirm prompts the user for yes/no confirmation
func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		log.Warn().Err(err).Msg("failed to read user input")
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
