package scoring

import "time"

// UniversalScores represents the scoring breakdown using universal dimensions.
// These dimensions are domain-agnostic and work for any type of project.
type UniversalScores struct {
	// CompletionLikelihood answers "Will I actually finish this?"
	// Higher scores for simpler, more achievable projects
	CompletionLikelihood float64 `json:"completion_likelihood"`

	// SkillFit answers "Can I do this with what I know?"
	// Higher scores when project aligns with existing capabilities
	SkillFit float64 `json:"skill_fit"`

	// TimeToDone answers "How long until it's real?"
	// Higher scores for faster timelines
	TimeToDone float64 `json:"time_to_done"`

	// RewardAlignment answers "Does this give me what I want?"
	// Higher scores when project matches stated goals
	RewardAlignment float64 `json:"reward_alignment"`

	// Sustainability answers "Will I stay motivated?"
	// Higher scores for projects with built-in accountability or interest
	Sustainability float64 `json:"sustainability"`

	// AvoidanceFit answers "Does this dodge my pitfalls?"
	// Higher scores when project avoids user's stated concerns
	AvoidanceFit float64 `json:"avoidance_fit"`

	// Total is the weighted sum of all dimensions (0-10 scale)
	Total float64 `json:"total"`
}

// UniversalAnalysis extends the legacy Analysis model with universal scores.
type UniversalAnalysis struct {
	// Universal contains the dimension-based scoring breakdown
	Universal UniversalScores `json:"universal"`

	// FinalScore is the weighted total (0-10 scale)
	FinalScore float64 `json:"final_score"`

	// Recommendation is the human-readable verdict
	Recommendation string `json:"recommendation"`

	// Insights are dimension-specific observations
	Insights map[string]string `json:"insights,omitempty"`

	// AnalyzedAt records when the analysis was performed
	AnalyzedAt time.Time `json:"analyzed_at"`

	// ScoringMode indicates which engine produced this analysis
	ScoringMode string `json:"scoring_mode"` // "universal" or "legacy"
}

// GetRecommendation returns a human-readable recommendation based on score.
func (a *UniversalAnalysis) GetRecommendation() string {
	switch {
	case a.FinalScore >= 8.5:
		return "GREAT FIT - Start this now"
	case a.FinalScore >= 7.0:
		return "GOOD FIT - Worth pursuing"
	case a.FinalScore >= 5.0:
		return "MAYBE - Consider carefully"
	case a.FinalScore >= 3.0:
		return "POOR FIT - Likely to struggle"
	default:
		return "AVOID - Not aligned with your goals"
	}
}

// DimensionScore holds information about a single scored dimension.
type DimensionScore struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	MaxScore    float64 `json:"max_score"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	Insight     string  `json:"insight,omitempty"`
}

// ToSlice converts UniversalScores to a slice of DimensionScore for iteration.
func (s *UniversalScores) ToSlice() []DimensionScore {
	return []DimensionScore{
		{
			Name:        "Completion",
			Score:       s.CompletionLikelihood,
			MaxScore:    2.0,
			Description: "Will I actually finish this?",
		},
		{
			Name:        "Skill Fit",
			Score:       s.SkillFit,
			MaxScore:    2.0,
			Description: "Can I do this with what I know?",
		},
		{
			Name:        "Timeline",
			Score:       s.TimeToDone,
			MaxScore:    2.0,
			Description: "How long until it's real?",
		},
		{
			Name:        "Reward",
			Score:       s.RewardAlignment,
			MaxScore:    2.0,
			Description: "Does this give me what I want?",
		},
		{
			Name:        "Motivation",
			Score:       s.Sustainability,
			MaxScore:    1.0,
			Description: "Will I stay motivated?",
		},
		{
			Name:        "Avoidance",
			Score:       s.AvoidanceFit,
			MaxScore:    1.0,
			Description: "Does this dodge my pitfalls?",
		},
	}
}

// CalculateTotal computes the total score from individual dimensions.
// Note: This is a simple sum. For weighted totals, use the engine.
func (s *UniversalScores) CalculateTotal() float64 {
	return s.CompletionLikelihood +
		s.SkillFit +
		s.TimeToDone +
		s.RewardAlignment +
		s.Sustainability +
		s.AvoidanceFit
}

// MaxPossibleScore returns the maximum achievable score (10.0).
func MaxPossibleScore() float64 {
	return 10.0
}

// ScoreThresholds defines the boundaries for recommendations.
var ScoreThresholds = struct {
	GreatFit  float64
	GoodFit   float64
	Maybe     float64
	PoorFit   float64
	Avoid     float64
}{
	GreatFit: 8.5,
	GoodFit:  7.0,
	Maybe:    5.0,
	PoorFit:  3.0,
	Avoid:    0.0,
}
