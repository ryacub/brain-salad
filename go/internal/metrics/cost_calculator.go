package metrics

import (
	"fmt"
	"strings"
)

// TokenCost represents the cost of input and output tokens for a provider
type TokenCost struct {
	InputCostPerM  float64 // Cost per million input tokens
	OutputCostPerM float64 // Cost per million output tokens
}

// CostEstimate represents the estimated cost for a provider
type CostEstimate struct {
	Provider     string
	InputTokens  int
	OutputTokens int
	InputCost    float64
	OutputCost   float64
	TotalCost    float64
	Currency     string
}

// Pricing as of 2025 (in USD)
var providerPricing = map[string]TokenCost{
	"claude": {
		InputCostPerM:  3.0,  // $3 per million input tokens (Sonnet 3.5)
		OutputCostPerM: 15.0, // $15 per million output tokens
	},
	"openai": {
		InputCostPerM:  30.0, // $30 per million input tokens (GPT-4)
		OutputCostPerM: 60.0, // $60 per million output tokens
	},
	"ollama": {
		InputCostPerM:  0.0, // Free (local)
		OutputCostPerM: 0.0,
	},
	"custom": {
		InputCostPerM:  0.0, // Unknown
		OutputCostPerM: 0.0,
	},
	"rule_based": {
		InputCostPerM:  0.0, // No LLM used
		OutputCostPerM: 0.0,
	},
}

// CalculateCost estimates the cost based on token usage
func CalculateCost(provider string, inputTokens, outputTokens int) CostEstimate {
	// Normalize provider name (handle variations like "openai_gpt-4")
	normalizedProvider := normalizeProviderName(provider)

	// Get pricing for the provider
	pricing, exists := providerPricing[normalizedProvider]
	if !exists {
		// Default to unknown/custom pricing
		pricing = TokenCost{InputCostPerM: 0.0, OutputCostPerM: 0.0}
	}

	// Calculate costs
	inputCost := (float64(inputTokens) / 1_000_000.0) * pricing.InputCostPerM
	outputCost := (float64(outputTokens) / 1_000_000.0) * pricing.OutputCostPerM
	totalCost := inputCost + outputCost

	return CostEstimate{
		Provider:     provider,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    totalCost,
		Currency:     "USD",
	}
}

// FormatCost formats a cost estimate for display
func (ce CostEstimate) FormatCost() string {
	if ce.TotalCost == 0.0 {
		return "$0.00 (free)"
	}

	// Format with appropriate precision
	if ce.TotalCost < 0.01 {
		return fmt.Sprintf("$%.4f", ce.TotalCost)
	} else if ce.TotalCost < 1.0 {
		return fmt.Sprintf("$%.3f", ce.TotalCost)
	} else {
		return fmt.Sprintf("$%.2f", ce.TotalCost)
	}
}

// FormatDetailed formats detailed cost breakdown
func (ce CostEstimate) FormatDetailed() string {
	if ce.TotalCost == 0.0 {
		return "Free (no cost)"
	}

	return fmt.Sprintf(
		"Input: %d tokens ($%.4f) + Output: %d tokens ($%.4f) = Total: %s",
		ce.InputTokens, ce.InputCost,
		ce.OutputTokens, ce.OutputCost,
		ce.FormatCost(),
	)
}

// normalizeProviderName extracts the base provider name from variants
// e.g., "openai_gpt-4" -> "openai", "claude-3.5-sonnet" -> "claude"
func normalizeProviderName(provider string) string {
	provider = strings.ToLower(provider)

	// Check for common prefixes
	for key := range providerPricing {
		if strings.HasPrefix(provider, key) {
			return key
		}
	}

	return provider
}

// GetProviderPricing returns the pricing information for a provider
func GetProviderPricing(provider string) (TokenCost, bool) {
	normalized := normalizeProviderName(provider)
	pricing, exists := providerPricing[normalized]
	return pricing, exists
}

// EstimateMonthlyCost estimates monthly cost based on daily usage
func EstimateMonthlyCost(dailyInputTokens, dailyOutputTokens int, provider string) float64 {
	dailyCost := CalculateCost(provider, dailyInputTokens, dailyOutputTokens)
	return dailyCost.TotalCost * 30 // Assume 30 days per month
}
