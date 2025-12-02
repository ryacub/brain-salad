package scoring_test

import (
	"testing"

	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/ryacub/telos-idea-matrix/internal/scoring"
	"github.com/ryacub/telos-idea-matrix/internal/telos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test case from RUST_REFERENCE.md - High Score Example
const highScoreIdea = `Build an AI automation tool using Python and LangChain to help
hotel staff automate guest request routing. Can ship MVP in 30 days.
Target $2K/month recurring revenue. Will build in public on Twitter.`

// Test case from RUST_REFERENCE.md - Low Score Example
const lowScoreIdea = `Learn Rust and build a comprehensive game engine from scratch.
Will need 6 months to learn the basics first, then another 6 months
to build a production-ready system. Personal project for fun.`

// Test case from RUST_REFERENCE.md - Medium Score Example
const mediumScoreIdea = `Create a Python script to automate my daily standup notes.
Will use it personally to save 15 minutes per day. Should take
about 2 weeks to build a working version.`

func loadTestTelos(t *testing.T) *models.Telos {
	t.Helper()
	parser := telos.NewParser()
	telosData, err := parser.ParseFile("testdata/test_telos.md")
	require.NoError(t, err)
	return telosData
}

// ============================================================================
// HIGH SCORE TESTS (Expected: ~8.5-9.0)
// ============================================================================

func TestEngine_CalculateScore_HighScoreIdea_ReturnsHighScore(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(highScoreIdea)

	require.NoError(t, err)
	assert.NotNil(t, analysis)

	// Expected final score: ~8.5-9.0 (per RUST_REFERENCE.md)
	assert.GreaterOrEqual(t, analysis.FinalScore, 8.5, "High score idea should score >= 8.5")
	assert.LessOrEqual(t, analysis.FinalScore, 10.0, "Score should not exceed 10.0")

	// Should get PRIORITIZE recommendation
	assert.Equal(t, "\U0001F525 PRIORITIZE NOW", analysis.GetRecommendation())
}

func TestEngine_CalculateScore_HighScoreIdea_MissionAlignment(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(highScoreIdea)

	require.NoError(t, err)

	// Expected breakdown from RUST_REFERENCE.md:
	// - Domain Expertise: ~1.1 (uses Python, hotel domain)
	// - AI Alignment: ~1.4 (core AI product)
	// - Execution Support: ~0.75 (30-day MVP)
	// - Revenue Potential: ~0.45 ($2K/month)
	// - Total: ~3.7/4.0

	assert.GreaterOrEqual(t, analysis.Mission.DomainExpertise, 0.9, "Should leverage existing skills")
	assert.LessOrEqual(t, analysis.Mission.DomainExpertise, 1.2)

	assert.GreaterOrEqual(t, analysis.Mission.AIAlignment, 1.2, "Core AI product should score high")
	assert.LessOrEqual(t, analysis.Mission.AIAlignment, 1.5)

	assert.GreaterOrEqual(t, analysis.Mission.ExecutionSupport, 0.65, "30-day MVP is fast")
	assert.LessOrEqual(t, analysis.Mission.ExecutionSupport, 0.8)

	assert.GreaterOrEqual(t, analysis.Mission.RevenuePotential, 0.4, "$2K/month is clear monetization")
	assert.LessOrEqual(t, analysis.Mission.RevenuePotential, 0.5)

	assert.GreaterOrEqual(t, analysis.Mission.Total, 3.5, "Mission total should be >= 3.5")
	assert.LessOrEqual(t, analysis.Mission.Total, 4.0)
}

func TestEngine_CalculateScore_HighScoreIdea_AntiChallenge(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(highScoreIdea)

	require.NoError(t, err)

	// Expected breakdown from RUST_REFERENCE.md:
	// - Context Switching: ~1.15 (uses current stack - Python)
	// - Rapid Prototyping: ~0.95 (30-day MVP)
	// - Accountability: ~0.7 (public building on Twitter)
	// - Income Anxiety: ~0.45 (fast revenue)
	// - Total: ~3.25/3.5

	assert.GreaterOrEqual(t, analysis.AntiChallenge.ContextSwitching, 0.95, "Uses current stack")
	assert.LessOrEqual(t, analysis.AntiChallenge.ContextSwitching, 1.2)

	assert.GreaterOrEqual(t, analysis.AntiChallenge.RapidPrototyping, 0.8, "Fast MVP timeline")
	assert.LessOrEqual(t, analysis.AntiChallenge.RapidPrototyping, 1.0)

	assert.GreaterOrEqual(t, analysis.AntiChallenge.Accountability, 0.45, "Public building adds accountability")
	assert.LessOrEqual(t, analysis.AntiChallenge.Accountability, 0.8)

	assert.GreaterOrEqual(t, analysis.AntiChallenge.IncomeAnxiety, 0.25, "Revenue within 30-60 days")
	assert.LessOrEqual(t, analysis.AntiChallenge.IncomeAnxiety, 0.5)

	assert.GreaterOrEqual(t, analysis.AntiChallenge.Total, 3.0, "Anti-challenge total should be >= 3.0")
	assert.LessOrEqual(t, analysis.AntiChallenge.Total, 3.5)
}

func TestEngine_CalculateScore_HighScoreIdea_Strategic(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(highScoreIdea)

	require.NoError(t, err)

	// Expected breakdown from RUST_REFERENCE.md:
	// - Stack Compatibility: ~0.9 (Python flow sessions)
	// - Shipping Habit: ~0.7 (reusable AI components)
	// - Public Accountability: ~0.35 (Twitter validation)
	// - Revenue Testing: ~0.28 (SaaS model)
	// - Total: ~2.23/2.5

	assert.GreaterOrEqual(t, analysis.Strategic.StackCompatibility, 0.8, "Python enables flow")
	assert.LessOrEqual(t, analysis.Strategic.StackCompatibility, 1.0)

	assert.GreaterOrEqual(t, analysis.Strategic.ShippingHabit, 0.45, "Some reusable components")
	assert.LessOrEqual(t, analysis.Strategic.ShippingHabit, 0.8)

	assert.GreaterOrEqual(t, analysis.Strategic.PublicAccountability, 0.22, "Can validate quickly")
	assert.LessOrEqual(t, analysis.Strategic.PublicAccountability, 0.4)

	assert.GreaterOrEqual(t, analysis.Strategic.RevenueTesting, 0.16, "SaaS has leverage")
	assert.LessOrEqual(t, analysis.Strategic.RevenueTesting, 0.3)

	assert.GreaterOrEqual(t, analysis.Strategic.Total, 2.0, "Strategic total should be >= 2.0")
	assert.LessOrEqual(t, analysis.Strategic.Total, 2.5)
}

// ============================================================================
// LOW SCORE TESTS (Expected: ~2.0-3.0)
// ============================================================================

func TestEngine_CalculateScore_LowScoreIdea_ReturnsLowScore(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(lowScoreIdea)

	require.NoError(t, err)
	assert.NotNil(t, analysis)

	// Expected final score: ~1.3/10 (per RUST_REFERENCE.md)
	assert.LessOrEqual(t, analysis.FinalScore, 3.0, "Low score idea should score <= 3.0")
	assert.GreaterOrEqual(t, analysis.FinalScore, 0.0, "Score should not be negative")

	// Should get AVOID recommendation
	assert.Equal(t, "\U0001F6AB AVOID FOR NOW", analysis.GetRecommendation())
}

func TestEngine_CalculateScore_LowScoreIdea_MissionAlignment(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(lowScoreIdea)

	require.NoError(t, err)

	// Expected breakdown from RUST_REFERENCE.md:
	// - Domain Expertise: ~0.1 (no matching skills - Rust)
	// - AI Alignment: ~0.0 (no AI component)
	// - Execution Support: ~0.05 (learning-focused)
	// - Revenue Potential: ~0.0 (no revenue path)
	// - Total: ~0.15/4.0

	assert.LessOrEqual(t, analysis.Mission.DomainExpertise, 0.3, "New stack = low domain score")
	assert.LessOrEqual(t, analysis.Mission.AIAlignment, 0.4, "No AI = low AI score")
	assert.LessOrEqual(t, analysis.Mission.ExecutionSupport, 0.25, "Learning-focused = low execution")
	assert.LessOrEqual(t, analysis.Mission.RevenuePotential, 0.1, "No revenue = low potential")

	assert.LessOrEqual(t, analysis.Mission.Total, 1.0, "Mission total should be <= 1.0 for low score")
}

func TestEngine_CalculateScore_LowScoreIdea_AntiChallenge(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(lowScoreIdea)

	require.NoError(t, err)

	// Expected breakdown from RUST_REFERENCE.md:
	// - Context Switching: ~0.1 (complete stack switch - Rust penalty)
	// - Rapid Prototyping: ~0.1 (perfection-dependent)
	// - Accountability: ~0.05 (personal project)
	// - Income Anxiety: ~0.0 (no revenue)
	// - Total: ~0.25/3.5

	assert.LessOrEqual(t, analysis.AntiChallenge.ContextSwitching, 0.3, "Rust = context switching penalty")
	assert.LessOrEqual(t, analysis.AntiChallenge.RapidPrototyping, 0.25, "6+ months = slow")
	assert.LessOrEqual(t, analysis.AntiChallenge.Accountability, 0.2, "Personal project = low accountability")
	assert.LessOrEqual(t, analysis.AntiChallenge.IncomeAnxiety, 0.1, "No revenue plan")

	assert.LessOrEqual(t, analysis.AntiChallenge.Total, 0.8, "Anti-challenge total should be <= 0.8")
}

// ============================================================================
// MEDIUM SCORE TESTS (Expected: ~5.5-6.5)
// ============================================================================

func TestEngine_CalculateScore_MediumScoreIdea_ReturnsMediumScore(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(mediumScoreIdea)

	require.NoError(t, err)
	assert.NotNil(t, analysis)

	// Expected final score: ~5.7/10 (per RUST_REFERENCE.md)
	assert.GreaterOrEqual(t, analysis.FinalScore, 5.0, "Medium score idea should score >= 5.0")
	assert.LessOrEqual(t, analysis.FinalScore, 7.0, "Medium score idea should score <= 7.0")

	// Should get CONSIDER recommendation
	assert.Equal(t, "\u26A0\uFE0F CONSIDER LATER", analysis.GetRecommendation())
}

// ============================================================================
// EDGE CASES
// ============================================================================

func TestEngine_CalculateScore_EmptyIdea_ReturnsError(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	_, err := engine.CalculateScore("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "idea text cannot be empty")
}

func TestEngine_CalculateScore_NilTelos_ReturnsError(t *testing.T) {
	engine := scoring.NewEngine(nil)

	_, err := engine.CalculateScore("Some idea")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "telos configuration is required")
}

func TestEngine_CalculateScore_SetsAnalyzedAt(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(highScoreIdea)

	require.NoError(t, err)
	assert.False(t, analysis.AnalyzedAt.IsZero(), "AnalyzedAt should be set")
}

func TestEngine_CalculateScore_RawScoreEqualsComponentSum(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(highScoreIdea)

	require.NoError(t, err)

	// Raw score should equal sum of component totals
	expectedRaw := analysis.Mission.Total + analysis.AntiChallenge.Total + analysis.Strategic.Total
	assert.InDelta(t, expectedRaw, analysis.RawScore, 0.01, "Raw score should equal component sum")
}

func TestEngine_CalculateScore_FinalScoreIsScaled(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	analysis, err := engine.CalculateScore(highScoreIdea)

	require.NoError(t, err)

	// Final score should be raw score (already on 0-10 scale per Rust implementation)
	assert.InDelta(t, analysis.RawScore, analysis.FinalScore, 0.01)
}

// ============================================================================
// KEYWORD DETECTION TESTS
// ============================================================================

func TestEngine_DetectsAIKeywords_HighAlignment(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	idea := "Build an AI agent using GPT-4 and LangChain for automation"
	analysis, err := engine.CalculateScore(idea)

	require.NoError(t, err)

	// Should detect AI keywords and score high on AI alignment
	assert.GreaterOrEqual(t, analysis.Mission.AIAlignment, 1.2, "Should detect core AI keywords")
}

func TestEngine_DetectsStackMatch_HighCompatibility(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	idea := "Build a Python CLI tool using LangChain and OpenAI API"
	analysis, err := engine.CalculateScore(idea)

	require.NoError(t, err)

	// Should match current stack (Python, LangChain, OpenAI)
	assert.GreaterOrEqual(t, analysis.AntiChallenge.ContextSwitching, 0.95, "Should match current stack")
}

func TestEngine_DetectsStackMismatch_LowCompatibility(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	idea := "Build a mobile app with React Native and TypeScript"
	analysis, err := engine.CalculateScore(idea)

	require.NoError(t, err)

	// Should penalize for different stack
	assert.LessOrEqual(t, analysis.AntiChallenge.ContextSwitching, 0.65, "Should penalize stack mismatch")
}

// ============================================================================
// TIMELINE DETECTION TESTS
// ============================================================================

func TestEngine_DetectsFastTimeline_HighExecutionScore(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	idea := "Build MVP in 30 days with clear deliverable"
	analysis, err := engine.CalculateScore(idea)

	require.NoError(t, err)

	// Should detect fast timeline
	assert.GreaterOrEqual(t, analysis.Mission.ExecutionSupport, 0.65, "30 days should score high")
}

func TestEngine_DetectsSlowTimeline_LowExecutionScore(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	idea := "Build comprehensive system over 6 months"
	analysis, err := engine.CalculateScore(idea)

	require.NoError(t, err)

	// Should penalize slow timeline
	assert.LessOrEqual(t, analysis.Mission.ExecutionSupport, 0.45, "6 months should score lower")
}

// ============================================================================
// REVENUE DETECTION TESTS
// ============================================================================

func TestEngine_DetectsRevenueModel_HighRevenueScore(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	idea := "Build SaaS with $1000/month subscription recurring revenue"
	analysis, err := engine.CalculateScore(idea)

	require.NoError(t, err)

	// Should detect clear revenue model
	assert.GreaterOrEqual(t, analysis.Mission.RevenuePotential, 0.4, "Clear monetization should score high")
}

func TestEngine_DetectsNoRevenue_LowRevenueScore(t *testing.T) {
	telosData := loadTestTelos(t)
	engine := scoring.NewEngine(telosData)

	idea := "Build free tool for personal use"
	analysis, err := engine.CalculateScore(idea)

	require.NoError(t, err)

	// Should recognize no revenue
	assert.LessOrEqual(t, analysis.Mission.RevenuePotential, 0.1, "No revenue should score low")
}

// ============================================================================
// SINGLETON PATTERN TESTS
// ============================================================================

func TestGetEngine_ReusesSingleton(t *testing.T) {
	scoring.ResetEngine() // Clean slate

	telos := loadTestTelos(t)

	engine1 := scoring.GetEngine(telos)
	engine2 := scoring.GetEngine(telos)

	// Should be same instance (pointer equality)
	assert.Same(t, engine1, engine2, "Expected singleton to return same engine instance")
}

func TestGetEngine_RecreatesOnTelosChange(t *testing.T) {
	scoring.ResetEngine()

	telos1 := loadTestTelos(t)
	engine1 := scoring.GetEngine(telos1)

	// Create a modified telos
	telos2 := loadTestTelos(t)
	// Add an extra mission to trigger change detection
	telos2.Missions = append(telos2.Missions, models.Mission{
		ID:          "extra",
		Description: "Extra mission",
	})
	engine2 := scoring.GetEngine(telos2)

	// Should be different instances
	assert.NotSame(t, engine1, engine2, "Expected new engine when telos changes")
}

func TestGetEngine_ReusesAfterReset(t *testing.T) {
	scoring.ResetEngine()

	telos := loadTestTelos(t)
	engine1 := scoring.GetEngine(telos)

	scoring.ResetEngine()

	engine2 := scoring.GetEngine(telos)

	// Should be different instances after reset
	assert.NotSame(t, engine1, engine2, "Expected new engine after reset")
}

func TestGetEngine_ThreadSafe(t *testing.T) {
	scoring.ResetEngine()

	telos := loadTestTelos(t)

	// Simulate concurrent access
	const goroutines = 10
	engines := make([]*scoring.Engine, goroutines)
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(index int) {
			engines[index] = scoring.GetEngine(telos)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// All should be the same instance
	for i := 1; i < goroutines; i++ {
		assert.Same(t, engines[0], engines[i], "All engines should be the same instance")
	}
}

func TestGetEngine_ProducesSameScores(t *testing.T) {
	scoring.ResetEngine()

	telos := loadTestTelos(t)

	// Get singleton engine
	singletonEngine := scoring.GetEngine(telos)
	analysis1, err := singletonEngine.CalculateScore(highScoreIdea)
	require.NoError(t, err)

	// Create new engine the old way
	directEngine := scoring.NewEngine(telos)
	analysis2, err := directEngine.CalculateScore(highScoreIdea)
	require.NoError(t, err)

	// Should produce identical scores
	assert.InDelta(t, analysis1.FinalScore, analysis2.FinalScore, 0.001, "Singleton and direct engines should produce same scores")
	assert.InDelta(t, analysis1.Mission.Total, analysis2.Mission.Total, 0.001)
	assert.InDelta(t, analysis1.AntiChallenge.Total, analysis2.AntiChallenge.Total, 0.001)
	assert.InDelta(t, analysis1.Strategic.Total, analysis2.Strategic.Total, 0.001)
}
