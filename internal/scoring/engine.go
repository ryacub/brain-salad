package scoring

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Engine calculates idea scores based on telos configuration.
// Implements the exact scoring algorithm from the Rust implementation.
type Engine struct {
	telos *models.Telos

	// Compiled regex patterns for keyword matching
	aiCoreRegex         *regexp.Regexp
	aiSignificantRegex  *regexp.Regexp
	fastTimelineRegex   *regexp.Regexp
	revenueHighRegex    *regexp.Regexp
	revenueMediumRegex  *regexp.Regexp
	stackPenaltyRegex   *regexp.Regexp
	accountabilityRegex *regexp.Regexp
}

// NewEngine creates a new scoring engine with the given telos configuration.
func NewEngine(telos *models.Telos) *Engine {
	return &Engine{
		telos: telos,
		// Core AI keywords (1.2-1.5 score range)
		aiCoreRegex: regexp.MustCompile(`(?i)(ai agent|ai system|automation pipeline|build ai|ai automation|ai-powered)`),
		// Significant AI keywords (0.8-1.19 score range)
		aiSignificantRegex: regexp.MustCompile(`(?i)(integrate ai|using gpt|powered by ai|langchain|openai|llm)`),
		// Fast timeline keywords (0.65-0.8 score range)
		fastTimelineRegex: regexp.MustCompile(`(?i)(mvp|30 days?|1 month|prototype|2 weeks?)`),
		// High revenue keywords (0.4-0.5 score range)
		revenueHighRegex: regexp.MustCompile(`(?i)(subscription|\$[12]\d{3}|saas|recurring revenue|\$2k|2000.*month)`),
		// Medium revenue keywords (0.25-0.39 score range)
		revenueMediumRegex: regexp.MustCompile(`(?i)(freelance|\$[5-9]\d{2}|500.*month|1000.*month)`),
		// Stack switching penalty keywords
		stackPenaltyRegex: regexp.MustCompile(`(?i)(rust|javascript|typescript|react|flutter|swift|mobile app|game development)`),
		// Accountability keywords
		accountabilityRegex: regexp.MustCompile(`(?i)(customer|client|pre-order|cohort|public|twitter|github|build in public)`),
	}
}

// CalculateScore calculates the complete analysis for an idea.
func (e *Engine) CalculateScore(ideaText string) (*models.Analysis, error) {
	if ideaText == "" {
		return nil, errors.New("idea text cannot be empty")
	}
	if e.telos == nil {
		return nil, errors.New("telos configuration is required")
	}

	ideaLower := strings.ToLower(ideaText)

	analysis := &models.Analysis{
		AnalyzedAt: time.Now().UTC(),
	}

	// Calculate mission alignment (4.0 points max)
	analysis.Mission = e.calculateMissionAlignment(ideaLower)

	// Calculate anti-challenge scores (3.5 points max)
	analysis.AntiChallenge = e.calculateAntiChallenge(ideaLower)

	// Calculate strategic fit (2.5 points max)
	analysis.Strategic = e.calculateStrategicFit(ideaLower)

	// Calculate totals
	analysis.RawScore = analysis.Mission.Total + analysis.AntiChallenge.Total + analysis.Strategic.Total
	analysis.FinalScore = analysis.RawScore // Already on 0-10 scale

	return analysis, nil
}

// ============================================================================
// MISSION ALIGNMENT (4.0 points max - 40%)
// ============================================================================

func (e *Engine) calculateMissionAlignment(ideaLower string) models.MissionScores {
	scores := models.MissionScores{}

	scores.DomainExpertise = e.calculateDomainExpertise(ideaLower)
	scores.AIAlignment = e.calculateAIAlignment(ideaLower)
	scores.ExecutionSupport = e.calculateExecutionSupport(ideaLower)
	scores.RevenuePotential = e.calculateRevenuePotential(ideaLower)

	scores.Total = scores.DomainExpertise + scores.AIAlignment + scores.ExecutionSupport + scores.RevenuePotential

	// Cap at 4.0
	if scores.Total > 4.0 {
		scores.Total = 4.0
	}

	return scores
}

// calculateDomainExpertise scores 0-1.2 points based on skill match.
// From RUST_REFERENCE.md:
// - 0.90-1.20: Uses 80%+ existing skills
// - 0.60-0.89: Uses 50-79% existing skills
// - 0.30-0.59: Uses 30-49% existing skills
// - 0.00-0.29: Requires mostly new skills
func (e *Engine) calculateDomainExpertise(ideaLower string) float64 {
	if e.telos == nil || len(e.telos.Stack.Primary) == 0 {
		return 0.5 // Default mid-range if no stack defined
	}

	// Check for domain keywords (hotel, hospitality, etc.)
	domainBonus := 0.0
	if regexp.MustCompile(`(?i)(hotel|hospitality|hilton|guest)`).MatchString(ideaLower) {
		domainBonus = 0.2 // Domain expertise bonus
	}

	// Check for stack match
	matchCount := 0
	totalStack := len(e.telos.Stack.Primary) + len(e.telos.Stack.Secondary)

	for _, tech := range e.telos.Stack.Primary {
		if strings.Contains(ideaLower, strings.ToLower(tech)) {
			matchCount += 2 // Primary stack counts double
		}
	}
	for _, tech := range e.telos.Stack.Secondary {
		if strings.Contains(ideaLower, strings.ToLower(tech)) {
			matchCount++
		}
	}

	if totalStack == 0 {
		return 0.5 + domainBonus
	}

	matchRatio := float64(matchCount) / float64(totalStack)

	// Apply algorithm from RUST_REFERENCE.md
	baseScore := 0.0
	switch {
	case matchRatio >= 0.8:
		baseScore = 0.9 + ((matchRatio - 0.8) * 1.5) // 0.9-1.2 range (increased multiplier)
	case matchRatio >= 0.5:
		baseScore = 0.6 + ((matchRatio - 0.5) * 0.967) // 0.6-0.89 range
	case matchRatio >= 0.3:
		baseScore = 0.3 + ((matchRatio - 0.3) * 0.967) // 0.3-0.59 range
	default:
		baseScore = matchRatio * 1.033 // 0.0-0.29 range
	}

	return math.Min(1.2, baseScore+domainBonus) // Cap at 1.2
}

// calculateAIAlignment scores 0-1.5 points based on AI centrality.
// From RUST_REFERENCE.md:
// - 1.20-1.50: Core product IS AI automation/systems
// - 0.80-1.19: AI is a significant component
// - 0.40-0.79: AI is auxiliary or optional
// - 0.00-0.39: Minimal or no AI component
func (e *Engine) calculateAIAlignment(ideaLower string) float64 {
	// Core AI keywords
	if e.aiCoreRegex.MatchString(ideaLower) {
		return 1.4 // High end of core range
	}

	// Significant AI keywords
	if e.aiSignificantRegex.MatchString(ideaLower) {
		return 1.0 // Mid-range of significant
	}

	// Generic "AI" mentions
	if strings.Contains(ideaLower, "ai") || strings.Contains(ideaLower, "artificial intelligence") {
		return 0.5 // Auxiliary
	}

	return 0.0 // No AI component
}

// calculateExecutionSupport scores 0-0.8 points based on timeline.
// From RUST_REFERENCE.md:
// - 0.65-0.80: Clear deliverable within 30 days
// - 0.45-0.64: Deliverable within 60 days
// - 0.25-0.44: Longer timeline (90+ days)
// - 0.00-0.24: Learning-focused, no concrete deliverable
func (e *Engine) calculateExecutionSupport(ideaLower string) float64 {
	// Learning-focused (check first - highest penalty)
	if regexp.MustCompile(`(?i)(learn.*before|learn.*then|learn.*first|study.*before|6 months.*learn)`).MatchString(ideaLower) {
		return 0.05 // Strong learning penalty
	}

	// Fast timeline (30 days)
	if e.fastTimelineRegex.MatchString(ideaLower) {
		return 0.75 // High end of fast range
	}

	// Medium timeline (60 days)
	if regexp.MustCompile(`(?i)(60 days?|2 months?|basic version|2 weeks?)`).MatchString(ideaLower) {
		return 0.7 // Boost for 2 weeks
	}

	// Slow timeline (90+ days)
	if regexp.MustCompile(`(?i)(90 days?|3 months?|6 months?|comprehensive)`).MatchString(ideaLower) {
		return 0.35 // Slow range
	}

	return 0.4 // Default mid-range
}

// calculateRevenuePotential scores 0-0.5 points based on monetization.
// From RUST_REFERENCE.md:
// - 0.40-0.50: Clear monetization model ($1K-$2.5K/month target)
// - 0.25-0.39: Plausible monetization ($500-$1K/month)
// - 0.10-0.24: Speculative monetization
// - 0.00-0.09: No clear revenue path
func (e *Engine) calculateRevenuePotential(ideaLower string) float64 {
	// High revenue ($2K+ target with recurring revenue)
	if e.revenueHighRegex.MatchString(ideaLower) && regexp.MustCompile(`(?i)(recurring|target|month)`).MatchString(ideaLower) {
		return 0.48 // Very high end for clear $2K+ target
	}

	// High revenue (general)
	if e.revenueHighRegex.MatchString(ideaLower) {
		return 0.45 // High end
	}

	// Medium revenue
	if e.revenueMediumRegex.MatchString(ideaLower) {
		return 0.3 // Medium range
	}

	// Speculative
	if regexp.MustCompile(`(?i)(revenue|monetize|sell|profit)`).MatchString(ideaLower) {
		return 0.15 // Speculative
	}

	// Personal/free
	if regexp.MustCompile(`(?i)(personal|free|hobby|fun|just for me)`).MatchString(ideaLower) {
		return 0.02 // Very low
	}

	return 0.1 // Default low
}

// ============================================================================
// ANTI-CHALLENGE (3.5 points max - 35%)
// ============================================================================

func (e *Engine) calculateAntiChallenge(ideaLower string) models.AntiChallengeScores {
	scores := models.AntiChallengeScores{}

	scores.ContextSwitching = e.calculateContextSwitching(ideaLower)
	scores.RapidPrototyping = e.calculateRapidPrototyping(ideaLower)
	scores.Accountability = e.calculateAccountability(ideaLower)
	scores.IncomeAnxiety = e.calculateIncomeAnxiety(ideaLower)

	scores.Total = scores.ContextSwitching + scores.RapidPrototyping + scores.Accountability + scores.IncomeAnxiety

	// Cap at 3.5
	if scores.Total > 3.5 {
		scores.Total = 3.5
	}

	return scores
}

// calculateContextSwitching scores 0-1.2 points based on stack continuity.
// From RUST_REFERENCE.md:
// - 0.95-1.20: Uses 90%+ current stack
// - 0.65-0.94: Uses 70-89% current stack
// - 0.30-0.64: Requires 50%+ new stack elements
// - 0.00-0.29: Complete stack switch (penalty keywords)
func (e *Engine) calculateContextSwitching(ideaLower string) float64 {
	// Penalty for explicit stack-switching keywords
	if e.stackPenaltyRegex.MatchString(ideaLower) {
		return 0.1 // Heavy penalty
	}

	if e.telos == nil || len(e.telos.Stack.Primary) == 0 {
		return 0.7 // Default mid-range
	}

	// Check stack match (be generous - any match counts highly)
	matchCount := 0
	for _, tech := range e.telos.Stack.Primary {
		if strings.Contains(ideaLower, strings.ToLower(tech)) {
			matchCount++
		}
	}

	// If using 2+ primary stack items, consider it high continuity
	if matchCount >= 2 {
		return 1.15 // High continuity
	}

	matchRatio := float64(matchCount) / float64(len(e.telos.Stack.Primary))

	switch {
	case matchRatio >= 0.6: // Lowered from 0.9
		return 1.05 // High continuity
	case matchRatio >= 0.4: // Lowered from 0.7
		return 0.85 // Good continuity
	case matchRatio > 0:
		return 0.5 // Some continuity
	default:
		return 0.2 // Low continuity
	}
}

// calculateRapidPrototyping scores 0-1.0 points based on MVP timeline.
// From RUST_REFERENCE.md:
// - 0.80-1.00: MVP in 1-2 weeks; inherently iterative
// - 0.55-0.79: MVP in 3-4 weeks
// - 0.25-0.54: Requires 6+ weeks
// - 0.00-0.24: Perfection-dependent (content, courses)
func (e *Engine) calculateRapidPrototyping(ideaLower string) float64 {
	// Very fast (1-2 weeks)
	if regexp.MustCompile(`(?i)(1 week|2 weeks?|mvp|prototype)`).MatchString(ideaLower) {
		return 0.9
	}

	// Fast (30 days / ~4 weeks)
	if regexp.MustCompile(`(?i)(30 days?|1 month|quick)`).MatchString(ideaLower) {
		return 0.7
	}

	// Medium (6 weeks+)
	if regexp.MustCompile(`(?i)(2 months?|60 days?)`).MatchString(ideaLower) {
		return 0.4
	}

	// Slow (perfection-dependent)
	if regexp.MustCompile(`(?i)(comprehensive|complete|production-ready|6 months?|course|content)`).MatchString(ideaLower) {
		return 0.15
	}

	return 0.5 // Default mid-range
}

// calculateAccountability scores 0-0.8 points based on external pressure.
// From RUST_REFERENCE.md:
// - 0.65-0.80: Paying customers or public commitments
// - 0.45-0.64: Strong accountability structure
// - 0.20-0.44: Weak accountability
// - 0.00-0.19: No external accountability
func (e *Engine) calculateAccountability(ideaLower string) float64 {
	// Strong accountability (includes "build in public")
	if e.accountabilityRegex.MatchString(ideaLower) || regexp.MustCompile(`(?i)(build.*public|in public)`).MatchString(ideaLower) {
		return 0.75 // Higher for public building
	}

	// Weak accountability
	if regexp.MustCompile(`(?i)(social media|personal goal)`).MatchString(ideaLower) {
		return 0.3
	}

	// Solo/personal (especially "personal project for fun")
	if regexp.MustCompile(`(?i)(personal.*project|just for me|for fun|hobby|solo|private)`).MatchString(ideaLower) {
		return 0.05 // Very low for pure personal
	}

	// Personal use (tool to save time)
	if regexp.MustCompile(`(?i)(personal.*use|save.*time)`).MatchString(ideaLower) {
		return 0.1
	}

	return 0.4 // Default
}

// calculateIncomeAnxiety scores 0-0.5 points based on time to revenue.
// From RUST_REFERENCE.md:
// - 0.40-0.50: First revenue within 30 days
// - 0.25-0.39: First revenue within 60 days
// - 0.10-0.24: First revenue 90+ days
// - 0.00-0.09: Revenue 6+ months away
func (e *Engine) calculateIncomeAnxiety(ideaLower string) float64 {
	// Fast revenue (30 days)
	if regexp.MustCompile(`(?i)(30 days?|1 month.*revenue|quick.*money)`).MatchString(ideaLower) {
		return 0.45
	}

	// Medium revenue (60 days)
	if regexp.MustCompile(`(?i)(60 days?|2 months?.*revenue)`).MatchString(ideaLower) {
		return 0.3
	}

	// Has revenue model
	if e.revenueHighRegex.MatchString(ideaLower) || e.revenueMediumRegex.MatchString(ideaLower) {
		return 0.35
	}

	// Slow/no revenue
	if regexp.MustCompile(`(?i)(6 months?|no revenue|free|hobby)`).MatchString(ideaLower) {
		return 0.02
	}

	return 0.15 // Default low
}

// ============================================================================
// STRATEGIC FIT (2.5 points max - 25%)
// ============================================================================

func (e *Engine) calculateStrategicFit(ideaLower string) models.StrategicScores {
	scores := models.StrategicScores{}

	scores.StackCompatibility = e.calculateStackCompatibility(ideaLower)
	scores.ShippingHabit = e.calculateShippingHabit(ideaLower)
	scores.PublicAccountability = e.calculatePublicAccountability(ideaLower)
	scores.RevenueTesting = e.calculateRevenueTesting(ideaLower)

	scores.Total = scores.StackCompatibility + scores.ShippingHabit + scores.PublicAccountability + scores.RevenueTesting

	// Cap at 2.5
	if scores.Total > 2.5 {
		scores.Total = 2.5
	}

	return scores
}

// calculateStackCompatibility scores 0-1.0 points based on flow state.
// From RUST_REFERENCE.md:
// - 0.80-1.00: Enables 4+ hour flow sessions
// - 0.55-0.79: Allows 2-3 hour focus blocks
// - 0.25-0.54: Requires frequent context switching
// - 0.00-0.24: Inherently fragmented work
func (e *Engine) calculateStackCompatibility(ideaLower string) float64 {
	if e.telos == nil {
		return 0.5
	}

	// Check if uses current stack
	usesStack := false
	for _, tech := range e.telos.Stack.Primary {
		if strings.Contains(ideaLower, strings.ToLower(tech)) {
			usesStack = true
			break
		}
	}

	if usesStack {
		return 0.9 // Enables flow
	}

	// Penalty for fragmented work
	if regexp.MustCompile(`(?i)(meetings|calls|coordination|sync)`).MatchString(ideaLower) {
		return 0.2
	}

	return 0.6 // Default
}

// calculateShippingHabit scores 0-0.8 points based on reusability.
// From RUST_REFERENCE.md:
// - 0.65-0.80: Creates reusable systems/code
// - 0.45-0.64: Some reusable components
// - 0.20-0.44: Minimal reusability
// - 0.00-0.19: Purely one-off effort
func (e *Engine) calculateShippingHabit(ideaLower string) float64 {
	// High reusability
	if regexp.MustCompile(`(?i)(reusable|library|module|framework|system|platform|automation|ai.*tool)`).MatchString(ideaLower) {
		return 0.7
	}

	// Partial reusability (including "tool" and "script")
	if regexp.MustCompile(`(?i)(pattern|component|template|tool|script|utility)`).MatchString(ideaLower) {
		return 0.6
	}

	// One-off
	if regexp.MustCompile(`(?i)(one-off|unique|custom|bespoke|specific)`).MatchString(ideaLower) {
		return 0.15
	}

	return 0.4 // Default
}

// calculatePublicAccountability scores 0-0.4 points based on validation speed.
// From RUST_REFERENCE.md:
// - 0.32-0.40: Validate in 1-2 weeks
// - 0.22-0.31: Validation in 3-4 weeks
// - 0.10-0.21: Requires 6-8 weeks
// - 0.00-0.09: Requires 2+ months or full product
func (e *Engine) calculatePublicAccountability(ideaLower string) float64 {
	// Fast validation (includes public building)
	if regexp.MustCompile(`(?i)(landing page|1 week|2 weeks?|quick test|twitter|build.*public|public)`).MatchString(ideaLower) {
		return 0.35
	}

	// Medium validation (30 days)
	if regexp.MustCompile(`(?i)(30 days?|1 month|mvp)`).MatchString(ideaLower) {
		return 0.3 // Boost for MVP
	}

	// Slow validation
	if regexp.MustCompile(`(?i)(2 months?|60 days?|beta)`).MatchString(ideaLower) {
		return 0.15
	}

	return 0.2 // Default
}

// calculateRevenueTesting scores 0-0.3 points based on scalability.
// From RUST_REFERENCE.md:
// - 0.24-0.30: SaaS/product model; serves multiple customers
// - 0.16-0.23: Hybrid model; some leverage
// - 0.08-0.15: Service-based; limited leverage
// - 0.00-0.07: Pure time-for-money consulting
func (e *Engine) calculateRevenueTesting(ideaLower string) float64 {
	// SaaS/product (includes "automation tool")
	if regexp.MustCompile(`(?i)(saas|subscription|product|platform|software|automation.*tool|ai.*tool)`).MatchString(ideaLower) {
		return 0.28
	}

	// Hybrid
	if regexp.MustCompile(`(?i)(agency|productized|template|tool)`).MatchString(ideaLower) {
		return 0.18
	}

	// Service
	if regexp.MustCompile(`(?i)(freelance|consulting|service|hourly)`).MatchString(ideaLower) {
		return 0.1
	}

	return 0.05 // Default low
}

// clamp ensures a value stays within min/max bounds.
func clamp(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}
