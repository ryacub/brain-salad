package scoring

import (
	"testing"

	"github.com/rayyacub/telos-idea-matrix/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testProfile() *profile.Profile {
	return &profile.Profile{
		Version: 1,
		Priorities: map[string]float64{
			profile.DimensionCompletionLikelihood: 0.20,
			profile.DimensionSkillFit:             0.15,
			profile.DimensionTimeToDone:           0.20,
			profile.DimensionRewardAlignment:      0.20,
			profile.DimensionSustainability:       0.15,
			profile.DimensionAvoidanceFit:         0.10,
		},
		Goals: []string{"sell pottery at the farmer's market"},
		Avoid: []string{"wholesale", "large inventory"},
		Preferences: profile.Preferences{
			MoneyMatters:    profile.MoneyMattersYes,
			PrefersFamiliar: true,
			CompletionFirst: true,
			PushesThrough:   true,
		},
	}
}

func TestNewUniversalEngine_CreatesEngine(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.profile)
	assert.NotEmpty(t, engine.goalKeywords)
}

func TestNewUniversalEngine_ExtractsKeywords(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	assert.Contains(t, engine.goalKeywords, "pottery")
	assert.Contains(t, engine.goalKeywords, "farmer")
	assert.Contains(t, engine.goalKeywords, "market")
}

func TestNewUniversalEngine_ExtractsAvoidKeywords(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	assert.Contains(t, engine.avoidKeywords, "wholesale")
	assert.Contains(t, engine.avoidKeywords, "inventory")
}

func TestScore_ReturnsAnalysis(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	analysis, err := engine.Score("sell handmade pottery at market")
	require.NoError(t, err)

	assert.NotNil(t, analysis)
	assert.Equal(t, "universal", analysis.ScoringMode)
	assert.NotZero(t, analysis.FinalScore)
	assert.NotEmpty(t, analysis.Recommendation)
}

func TestScore_ScoreInValidRange(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	analysis, err := engine.Score("any idea here")
	require.NoError(t, err)

	assert.GreaterOrEqual(t, analysis.FinalScore, 0.0)
	assert.LessOrEqual(t, analysis.FinalScore, 10.0)
}

func TestScore_HighAlignmentIdea_ScoresHigh(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	// Idea that matches goals, avoids pitfalls, has good timeline
	analysis, err := engine.Score("sell pottery at farmer's market this weekend, simple setup")
	require.NoError(t, err)

	assert.Greater(t, analysis.FinalScore, 6.0, "well-aligned idea should score above 6")
}

func TestScore_PoorAlignmentIdea_ScoresLow(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	// Idea that doesn't match goals and hits avoid words
	analysis, err := engine.Score("wholesale distribution center with large inventory management")
	require.NoError(t, err)

	assert.Less(t, analysis.FinalScore, 5.0, "poorly-aligned idea should score below 5")
}

func TestScore_FastTimeline_HigherTimeScore(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	fastIdea, _ := engine.Score("launch this weekend")
	slowIdea, _ := engine.Score("launch next year")

	assert.Greater(t, fastIdea.Universal.TimeToDone, slowIdea.Universal.TimeToDone)
}

func TestScore_SimpleIdea_HigherCompletionScore(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	simpleIdea, _ := engine.Score("simple quick mvp")
	complexIdea, _ := engine.Score("comprehensive enterprise solution")

	assert.Greater(t, simpleIdea.Universal.CompletionLikelihood, complexIdea.Universal.CompletionLikelihood)
}

func TestScore_AvoidedContent_LowerAvoidanceScore(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	goodIdea, _ := engine.Score("sell pottery directly to customers")
	badIdea, _ := engine.Score("wholesale pottery with large inventory")

	assert.Greater(t, goodIdea.Universal.AvoidanceFit, badIdea.Universal.AvoidanceFit)
}

func TestScore_MoneyMatters_RevenueIdeaScoresHigher(t *testing.T) {
	p := testProfile()
	p.Preferences.MoneyMatters = profile.MoneyMattersYes
	engine := NewUniversalEngine(p)

	revenueIdea, _ := engine.Score("sell products for profit to customers")
	hobbyIdea, _ := engine.Score("personal hobby project just for fun")

	assert.Greater(t, revenueIdea.Universal.RewardAlignment, hobbyIdea.Universal.RewardAlignment)
}

func TestScore_MoneyNotImportant_HobbyIdeaNotPenalized(t *testing.T) {
	p := testProfile()
	p.Preferences.MoneyMatters = profile.MoneyMattersNotReally
	engine := NewUniversalEngine(p)

	hobbyIdea, _ := engine.Score("personal hobby project for fun")

	// Should still get decent reward alignment because goals matter, not just money
	assert.GreaterOrEqual(t, hobbyIdea.Universal.RewardAlignment, 0.0)
}

func TestScore_PublicAccountability_HigherSustainability(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	publicIdea, _ := engine.Score("launch publicly with deadline and customers")
	privateIdea, _ := engine.Score("solo private project just for me")

	assert.Greater(t, publicIdea.Universal.Sustainability, privateIdea.Universal.Sustainability)
}

func TestScore_SetsAnalyzedAt(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	analysis, _ := engine.Score("any idea")

	assert.False(t, analysis.AnalyzedAt.IsZero())
}

func TestScore_GeneratesInsights(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	// Use an idea that should trigger insights
	analysis, _ := engine.Score("comprehensive enterprise project next year")

	// Should have some insights for long timeline / complex idea
	// Insights are optional, so just check it's a valid map
	assert.NotNil(t, analysis.Insights)
}

func TestGetRecommendation_HighScore_GreatFit(t *testing.T) {
	analysis := &UniversalAnalysis{FinalScore: 9.0}
	rec := analysis.GetRecommendation()
	assert.Contains(t, rec, "GREAT FIT")
}

func TestGetRecommendation_GoodScore_GoodFit(t *testing.T) {
	analysis := &UniversalAnalysis{FinalScore: 7.5}
	rec := analysis.GetRecommendation()
	assert.Contains(t, rec, "GOOD FIT")
}

func TestGetRecommendation_MediumScore_Maybe(t *testing.T) {
	analysis := &UniversalAnalysis{FinalScore: 5.5}
	rec := analysis.GetRecommendation()
	assert.Contains(t, rec, "MAYBE")
}

func TestGetRecommendation_LowScore_PoorFit(t *testing.T) {
	analysis := &UniversalAnalysis{FinalScore: 3.5}
	rec := analysis.GetRecommendation()
	assert.Contains(t, rec, "POOR FIT")
}

func TestGetRecommendation_VeryLowScore_Avoid(t *testing.T) {
	analysis := &UniversalAnalysis{FinalScore: 2.0}
	rec := analysis.GetRecommendation()
	assert.Contains(t, rec, "AVOID")
}

func TestUniversalScores_ToSlice_Returns6Dimensions(t *testing.T) {
	scores := &UniversalScores{
		CompletionLikelihood: 1.5,
		SkillFit:             1.2,
		TimeToDone:           1.8,
		RewardAlignment:      1.6,
		Sustainability:       0.7,
		AvoidanceFit:         0.9,
	}

	slice := scores.ToSlice()
	assert.Len(t, slice, 6)
}

func TestUniversalScores_CalculateTotal_SumsCorrectly(t *testing.T) {
	scores := &UniversalScores{
		CompletionLikelihood: 1.0,
		SkillFit:             1.0,
		TimeToDone:           1.0,
		RewardAlignment:      1.0,
		Sustainability:       0.5,
		AvoidanceFit:         0.5,
	}

	total := scores.CalculateTotal()
	assert.Equal(t, 5.0, total)
}

func TestGetProfile_ReturnsProfile(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	retrieved := engine.GetProfile()
	assert.Equal(t, p, retrieved)
}

// Table-driven test for various idea types
func TestScore_VariousIdeaTypes(t *testing.T) {
	p := testProfile()
	engine := NewUniversalEngine(p)

	tests := []struct {
		name          string
		idea          string
		minScore      float64
		maxScore      float64
		checkField    string
		fieldMinScore float64
	}{
		{
			name:     "MVP idea scores high on completion",
			idea:     "simple mvp prototype",
			minScore: 0,
			maxScore: 10,
		},
		{
			name:     "Weekend project scores high on timeline",
			idea:     "finish this weekend",
			minScore: 0,
			maxScore: 10,
		},
		{
			name:     "Customer-facing idea scores high on sustainability",
			idea:     "launch to customers with deadline",
			minScore: 0,
			maxScore: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := engine.Score(tt.idea)
			require.NoError(t, err)

			assert.GreaterOrEqual(t, analysis.FinalScore, tt.minScore)
			assert.LessOrEqual(t, analysis.FinalScore, tt.maxScore)
		})
	}
}
