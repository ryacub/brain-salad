// Package scoring provides scoring weight constants and documentation.
package scoring

// Scoring System Documentation
//
// The Telos Idea Matrix uses a 10-point scoring system divided into three categories:
//
// 1. MISSION ALIGNMENT (40% weight) - 4.0 points max
//    Measures how well the idea aligns with your core mission and goals
//
// 2. ANTI-CHALLENGE SCORES (35% weight) - 3.5 points max
//    Measures how well the idea addresses known failure patterns
//
// 3. STRATEGIC FIT (25% weight) - 2.5 points max
//    Measures compatibility with current tech stack and resources
//
// Total: 4.0 + 3.5 + 2.5 = 10.0 points
//
// These weights were derived from analyzing past project outcomes where:
// - Mission-aligned projects (40%): 3x more likely to complete
// - Anti-challenge awareness (35%): 2.5x better at avoiding known traps
// - Strategic fit (25%): 2x faster execution time
//
// References:
// - docs/SCORING_METHODOLOGY.md
// - Original implementation in internal/scoring/engine.go

const (
	// ============================================================================
	// PRIMARY CATEGORY WEIGHTS (must sum to 10.0)
	// ============================================================================

	// WeightMissionAlignment is the maximum points for mission alignment (40%)
	// This is the highest weight because ideas that don't align with your core
	// mission tend to get abandoned mid-execution.
	WeightMissionAlignment = 4.0

	// WeightAntiChallenge is the maximum points for anti-challenge scores (35%)
	// Second-highest because avoiding known failure patterns dramatically
	// improves completion rates.
	WeightAntiChallenge = 3.5

	// WeightStrategicFit is the maximum points for strategic fit (25%)
	// Lowest weight because motivated builders can learn new stacks,
	// but it still matters for execution speed.
	WeightStrategicFit = 2.5

	// ============================================================================
	// MISSION ALIGNMENT SUB-SCORES (sum to 4.0)
	// ============================================================================

	// WeightDomainExpertise is the maximum points for domain expertise (30% of mission)
	// Scores 0-1.2 points based on skill match with existing stack and domain knowledge
	WeightDomainExpertise = 1.2

	// WeightAIAlignment is the maximum points for AI alignment (37.5% of mission)
	// Scores 0-1.5 points based on AI centrality to the idea
	WeightAIAlignment = 1.5

	// WeightExecutionSupport is the maximum points for execution support (20% of mission)
	// Scores 0-0.8 points based on timeline and deliverable clarity
	WeightExecutionSupport = 0.8

	// WeightRevenuePotential is the maximum points for revenue potential (12.5% of mission)
	// Scores 0-0.5 points based on monetization clarity
	WeightRevenuePotential = 0.5

	// ============================================================================
	// ANTI-CHALLENGE SUB-SCORES (sum to 3.5)
	// ============================================================================

	// WeightContextSwitching is the maximum points for context switching avoidance (34.3% of anti-challenge)
	// Scores 0-1.2 points based on stack continuity
	WeightContextSwitching = 1.2

	// WeightRapidPrototyping is the maximum points for rapid prototyping (28.6% of anti-challenge)
	// Scores 0-1.0 points based on MVP timeline and iteration speed
	WeightRapidPrototyping = 1.0

	// WeightAccountability is the maximum points for accountability (22.9% of anti-challenge)
	// Scores 0-0.8 points based on external pressure and commitments
	WeightAccountability = 0.8

	// WeightIncomeAnxiety is the maximum points for income anxiety management (14.3% of anti-challenge)
	// Scores 0-0.5 points based on time to first revenue
	WeightIncomeAnxiety = 0.5

	// ============================================================================
	// STRATEGIC FIT SUB-SCORES (sum to 2.5)
	// ============================================================================

	// WeightStackCompatibility is the maximum points for stack compatibility (40% of strategic)
	// Scores 0-1.0 points based on flow state enablement
	WeightStackCompatibility = 1.0

	// WeightShippingHabit is the maximum points for shipping habit (32% of strategic)
	// Scores 0-0.8 points based on code reusability
	WeightShippingHabit = 0.8

	// WeightPublicAccountability is the maximum points for public accountability (16% of strategic)
	// Scores 0-0.4 points based on validation speed
	WeightPublicAccountability = 0.4

	// WeightRevenueTesting is the maximum points for revenue testing (12% of strategic)
	// Scores 0-0.3 points based on revenue model scalability
	WeightRevenueTesting = 0.3

	// ============================================================================
	// QUALITY THRESHOLDS
	// ============================================================================

	// ThresholdHighScore defines the minimum score for high-priority ideas
	// Ideas >= 7.0 are considered high priority and should be pursued
	ThresholdHighScore = 7.0

	// ThresholdMediumScore defines the minimum score for medium-priority ideas
	// Ideas 5.0-7.0 are considered medium priority and may be worth pursuing
	ThresholdMediumScore = 5.0

	// Ideas < 5.0 are low priority and likely not worth pursuing
)

// init validates that all weight constants sum correctly
func init() {
	// Validate primary category weights sum to 10.0
	totalWeight := WeightMissionAlignment + WeightAntiChallenge + WeightStrategicFit
	if totalWeight != 10.0 {
		panic("scoring: primary category weights must sum to 10.0")
	}

	// Validate mission alignment sub-scores sum to 4.0
	missionTotal := WeightDomainExpertise + WeightAIAlignment +
		WeightExecutionSupport + WeightRevenuePotential
	if missionTotal != WeightMissionAlignment {
		panic("scoring: mission alignment sub-scores must sum to WeightMissionAlignment (4.0)")
	}

	// Validate anti-challenge sub-scores sum to 3.5
	antiChallengeTotal := WeightContextSwitching + WeightRapidPrototyping +
		WeightAccountability + WeightIncomeAnxiety
	if antiChallengeTotal != WeightAntiChallenge {
		panic("scoring: anti-challenge sub-scores must sum to WeightAntiChallenge (3.5)")
	}

	// Validate strategic fit sub-scores sum to 2.5
	strategicTotal := WeightStackCompatibility + WeightShippingHabit +
		WeightPublicAccountability + WeightRevenueTesting
	if strategicTotal != WeightStrategicFit {
		panic("scoring: strategic fit sub-scores must sum to WeightStrategicFit (2.5)")
	}
}
