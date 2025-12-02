package scoring

import (
	"regexp"
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/profile"
)

// UniversalEngine scores ideas based on user-defined profiles.
// Unlike the legacy Engine, this uses universal dimensions that work for any domain.
type UniversalEngine struct {
	profile *profile.Profile

	// Extracted keywords from profile for matching
	goalKeywords  []string
	avoidKeywords []string

	// Pre-compiled regex patterns for efficiency
	timelinePatterns   map[*regexp.Regexp]float64
	completionPatterns map[*regexp.Regexp]float64
	revenuePatterns    map[*regexp.Regexp]float64
	motivationPatterns map[*regexp.Regexp]float64
}

// NewUniversalEngine creates a new universal scoring engine with the given profile.
func NewUniversalEngine(p *profile.Profile) *UniversalEngine {
	goalKw, avoidKw := profile.ExtractKeywordsFromProfile(p)

	engine := &UniversalEngine{
		profile:       p,
		goalKeywords:  goalKw,
		avoidKeywords: avoidKw,
	}

	engine.compilePatterns()
	return engine
}

// compilePatterns pre-compiles regex patterns for performance.
func (e *UniversalEngine) compilePatterns() {
	// Timeline patterns: shorter = higher score
	e.timelinePatterns = map[*regexp.Regexp]float64{
		regexp.MustCompile(`(?i)(today|tonight|this evening)`):                   1.0,
		regexp.MustCompile(`(?i)(tomorrow|next day)`):                            0.95,
		regexp.MustCompile(`(?i)(this week|few days|couple days)`):               0.90,
		regexp.MustCompile(`(?i)(next week|1 week|one week|weekend)`):            0.85,
		regexp.MustCompile(`(?i)(2 weeks|two weeks|couple weeks|fortnight)`):     0.75,
		regexp.MustCompile(`(?i)(this month|1 month|one month|30 days)`):         0.60,
		regexp.MustCompile(`(?i)(few months|2-3 months|couple months|60 days)`):  0.40,
		regexp.MustCompile(`(?i)(this year|6 months|half year|quarter)`):         0.25,
		regexp.MustCompile(`(?i)(next year|long.?term|years?|eventually)`):       0.10,
	}

	// Completion likelihood patterns: simpler = higher score
	e.completionPatterns = map[*regexp.Regexp]float64{
		regexp.MustCompile(`(?i)(simple|quick|easy|basic|minimal|tiny)`):         0.95,
		regexp.MustCompile(`(?i)(mvp|prototype|proof of concept|v1|first version)`): 0.90,
		regexp.MustCompile(`(?i)(small|focused|straightforward|lean)`):           0.85,
		regexp.MustCompile(`(?i)(doable|achievable|manageable|realistic)`):       0.80,
		regexp.MustCompile(`(?i)(moderate|medium|reasonable|standard)`):          0.50,
		regexp.MustCompile(`(?i)(comprehensive|complete|full|extensive|thorough)`): 0.30,
		regexp.MustCompile(`(?i)(complex|ambitious|large.?scale|massive)`):       0.20,
		regexp.MustCompile(`(?i)(enterprise|production.?ready|perfect|polished)`): 0.15,
	}

	// Revenue/reward patterns (used when money matters)
	e.revenuePatterns = map[*regexp.Regexp]float64{
		regexp.MustCompile(`(?i)(sell|selling|sales|revenue|income|profit|money)`): 0.85,
		regexp.MustCompile(`(?i)(customers?|clients?|buyers?|paying)`):             0.80,
		regexp.MustCompile(`(?i)(market|business|commercial|monetize)`):            0.70,
		regexp.MustCompile(`(?i)(subscription|recurring|saas|mrr)`):                0.90,
		regexp.MustCompile(`(?i)(freelance|contract|gig|commission)`):              0.60,
		regexp.MustCompile(`(?i)(free|hobby|personal|fun|learning|practice)`):      0.20,
	}

	// Motivation/accountability patterns
	e.motivationPatterns = map[*regexp.Regexp]float64{
		regexp.MustCompile(`(?i)(deadline|due date|committed|promised)`):          0.90,
		regexp.MustCompile(`(?i)(public|share|publish|launch|announce)`):          0.85,
		regexp.MustCompile(`(?i)(team|partner|collaborat|together)`):              0.80,
		regexp.MustCompile(`(?i)(customer|client|user|audience)`):                 0.85,
		regexp.MustCompile(`(?i)(excited|passionate|love|enjoy|fun)`):             0.75,
		regexp.MustCompile(`(?i)(curious|interesting|explore|experiment)`):        0.70,
		regexp.MustCompile(`(?i)(solo|alone|private|secret|just me)`):             0.30,
		regexp.MustCompile(`(?i)(obligation|have to|should|must|forced)`):         0.25,
	}
}

// Score evaluates an idea text against the user's profile.
func (e *UniversalEngine) Score(ideaText string) (*UniversalAnalysis, error) {
	ideaLower := strings.ToLower(ideaText)

	scores := &UniversalScores{}

	// Calculate each dimension
	scores.CompletionLikelihood = e.scoreCompletion(ideaLower)
	scores.SkillFit = e.scoreSkillFit(ideaLower)
	scores.TimeToDone = e.scoreTimeline(ideaLower)
	scores.RewardAlignment = e.scoreRewardAlignment(ideaLower)
	scores.Sustainability = e.scoreSustainability(ideaLower)
	scores.AvoidanceFit = e.scoreAvoidance(ideaLower)

	// Apply user's priority weights to get final score
	scores.Total = e.applyWeights(scores)

	analysis := &UniversalAnalysis{
		Universal:   *scores,
		FinalScore:  scores.Total,
		AnalyzedAt:  time.Now().UTC(),
		ScoringMode: "universal",
		Insights:    e.generateInsights(scores, ideaLower),
	}

	analysis.Recommendation = analysis.GetRecommendation()

	return analysis, nil
}

// scoreCompletion evaluates "Will I actually finish this?"
func (e *UniversalEngine) scoreCompletion(ideaLower string) float64 {
	maxScore := 2.0
	baseScore := 0.5 // Default mid-range

	// Check completion patterns
	for pattern, score := range e.completionPatterns {
		if pattern.MatchString(ideaLower) {
			if score > baseScore {
				baseScore = score
			}
		}
	}

	// Adjust based on user preference
	if e.profile.Preferences.CompletionFirst {
		// User values completion - be more strict about complexity
		if baseScore < 0.5 {
			baseScore *= 0.8 // Penalize complex ideas more
		}
	}

	return baseScore * maxScore
}

// scoreSkillFit evaluates "Can I do this with what I know?"
func (e *UniversalEngine) scoreSkillFit(ideaLower string) float64 {
	maxScore := 2.0

	// Match against goal keywords to infer domain familiarity
	goalMatch := profile.MatchScore(ideaLower, e.goalKeywords)

	// Adjust based on familiar preference
	var baseScore float64
	if e.profile.Preferences.PrefersFamiliar {
		// User prefers familiar - boost matching scores
		baseScore = 0.3 + (goalMatch * 0.7)
	} else {
		// User likes learning - don't penalize unfamiliar
		baseScore = 0.5 + (goalMatch * 0.4)
	}

	// Clamp to valid range
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore * maxScore
}

// scoreTimeline evaluates "How long until it's real?"
func (e *UniversalEngine) scoreTimeline(ideaLower string) float64 {
	maxScore := 2.0
	baseScore := 0.5 // Default mid-range (no timeline mentioned)

	// Check timeline patterns
	for pattern, score := range e.timelinePatterns {
		if pattern.MatchString(ideaLower) {
			if score > baseScore || (score < baseScore && baseScore == 0.5) {
				baseScore = score
			}
		}
	}

	// Adjust based on completion preference
	if e.profile.Preferences.CompletionFirst {
		// User wants to finish things - prefer shorter timelines
		if baseScore < 0.5 {
			baseScore *= 0.9 // Slight penalty for long timelines
		}
	}

	return baseScore * maxScore
}

// scoreRewardAlignment evaluates "Does this give me what I want?"
func (e *UniversalEngine) scoreRewardAlignment(ideaLower string) float64 {
	maxScore := 2.0

	// Primary: match against user's stated goals
	goalMatch := profile.MatchScore(ideaLower, e.goalKeywords)
	baseScore := goalMatch

	// Secondary: check revenue patterns based on money preference
	switch e.profile.Preferences.MoneyMatters {
	case profile.MoneyMattersYes:
		// User wants money - boost revenue-related ideas
		revenueScore := 0.0
		for pattern, score := range e.revenuePatterns {
			if pattern.MatchString(ideaLower) {
				if score > revenueScore {
					revenueScore = score
				}
			}
		}
		// Blend goal match with revenue signals
		baseScore = (baseScore * 0.6) + (revenueScore * 0.4)
	case profile.MoneyMattersNotReally:
		// User doesn't care about money - don't penalize non-revenue ideas
		// Just use goal matching
		baseScore = goalMatch
	default:
		// Sometimes - slight revenue consideration
		revenueScore := 0.3 // Neutral default
		for pattern, score := range e.revenuePatterns {
			if pattern.MatchString(ideaLower) {
				if score > revenueScore {
					revenueScore = score
				}
			}
		}
		baseScore = (baseScore * 0.8) + (revenueScore * 0.2)
	}

	// Clamp to valid range
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	if baseScore < 0 {
		baseScore = 0
	}

	return baseScore * maxScore
}

// scoreSustainability evaluates "Will I stay motivated?"
func (e *UniversalEngine) scoreSustainability(ideaLower string) float64 {
	maxScore := 1.0
	baseScore := 0.5 // Default mid-range

	// Check motivation patterns
	for pattern, score := range e.motivationPatterns {
		if pattern.MatchString(ideaLower) {
			if score > baseScore {
				baseScore = score
			}
		}
	}

	// Adjust based on push-through preference
	if e.profile.Preferences.PushesThrough {
		// User tends to finish - less dependent on motivation signals
		baseScore = 0.4 + (baseScore * 0.6) // Compress range upward
	}

	return baseScore * maxScore
}

// scoreAvoidance evaluates "Does this dodge my pitfalls?"
func (e *UniversalEngine) scoreAvoidance(ideaLower string) float64 {
	maxScore := 1.0

	// Check against user's avoid list
	avoidScore := profile.AvoidanceScore(ideaLower, e.avoidKeywords)

	// Also check against raw avoid strings (for phrases)
	for _, avoid := range e.profile.Avoid {
		avoidLower := strings.ToLower(avoid)
		if strings.Contains(ideaLower, avoidLower) {
			avoidScore *= 0.5 // Significant penalty for direct match
		}
	}

	if avoidScore < 0 {
		avoidScore = 0
	}

	return avoidScore * maxScore
}

// Scale factors convert raw dimension scores to a 0-10 total scale.
// Major dimensions (max 2.0) use ScaleMajor: 2.0 * 0.20 (priority) * 5 = 2.0 contribution
// Minor dimensions (max 1.0) use ScaleMinor: 1.0 * 0.10 (priority) * 10 = 1.0 contribution
const (
	ScaleMajor = 5.0  // For dimensions with max score 2.0 (completion, skill, time, reward)
	ScaleMinor = 10.0 // For dimensions with max score 1.0 (sustainability, avoidance)
)

// applyWeights calculates the weighted total score.
func (e *UniversalEngine) applyWeights(scores *UniversalScores) float64 {
	total := 0.0

	// Apply user's priority weights with appropriate scale factors
	// Major dimensions (max raw score 2.0)
	total += scores.CompletionLikelihood * e.profile.GetPriority(profile.DimensionCompletionLikelihood) * ScaleMajor
	total += scores.SkillFit * e.profile.GetPriority(profile.DimensionSkillFit) * ScaleMajor
	total += scores.TimeToDone * e.profile.GetPriority(profile.DimensionTimeToDone) * ScaleMajor
	total += scores.RewardAlignment * e.profile.GetPriority(profile.DimensionRewardAlignment) * ScaleMajor

	// Minor dimensions (max raw score 1.0)
	total += scores.Sustainability * e.profile.GetPriority(profile.DimensionSustainability) * ScaleMinor
	total += scores.AvoidanceFit * e.profile.GetPriority(profile.DimensionAvoidanceFit) * ScaleMinor

	// Clamp to 0-10 range
	if total > 10.0 {
		total = 10.0
	}
	if total < 0 {
		total = 0
	}

	return total
}

// generateInsights creates human-readable observations about the scores.
func (e *UniversalEngine) generateInsights(scores *UniversalScores, ideaLower string) map[string]string {
	insights := make(map[string]string)

	// Completion insight
	if scores.CompletionLikelihood >= 1.6 {
		insights["completion"] = "This looks achievable and well-scoped"
	} else if scores.CompletionLikelihood <= 0.8 {
		insights["completion"] = "This might be too ambitious to finish"
	}

	// Timeline insight
	if scores.TimeToDone >= 1.6 {
		insights["timeline"] = "Fast timeline - you'll see results quickly"
	} else if scores.TimeToDone <= 0.6 {
		insights["timeline"] = "Long timeline - patience required"
	}

	// Reward insight
	if scores.RewardAlignment >= 1.6 {
		insights["reward"] = "Strong alignment with your stated goals"
	} else if scores.RewardAlignment <= 0.6 {
		insights["reward"] = "Doesn't clearly match your goals"
	}

	// Avoidance insight
	if scores.AvoidanceFit <= 0.5 {
		insights["avoidance"] = "Warning: touches things you wanted to avoid"
	}

	return insights
}

// GetProfile returns the engine's profile for inspection.
func (e *UniversalEngine) GetProfile() *profile.Profile {
	return e.profile
}
