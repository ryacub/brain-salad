package wizard

import (
	"sort"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/profile"
)

// MapAnswersToProfile converts wizard answers into a Profile.
func MapAnswersToProfile(answers *WizardAnswers) *profile.Profile {
	p := profile.DefaultProfile()
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt

	// Start with balanced weights
	weights := map[string]float64{
		profile.DimensionCompletionLikelihood: 0.20,
		profile.DimensionSkillFit:             0.15,
		profile.DimensionTimeToDone:           0.20,
		profile.DimensionRewardAlignment:      0.20,
		profile.DimensionSustainability:       0.15,
		profile.DimensionAvoidanceFit:         0.10,
	}

	// Apply question-specific adjustments
	applyProjectStyleAnswer(answers, weights, p)
	applyFailureModeAnswer(answers, weights, p)
	applyPersistenceAnswer(answers, weights, p)
	applyMoneyMattersAnswer(answers, weights, p)
	applySkillPreferenceAnswer(answers, weights, p)

	// Normalize weights to sum to 1.0
	p.Priorities = weights
	profile.NormalizePriorities(p)

	// Set goals and avoid lists
	p.Goals = answers.Goals
	p.Avoid = answers.Avoid

	return p
}

// applyProjectStyleAnswer adjusts weights based on simple vs ambitious preference.
func applyProjectStyleAnswer(answers *WizardAnswers, weights map[string]float64, p *profile.Profile) {
	key := answers.GetAnswerKey("project_style")

	switch key {
	case "A": // Simple and done by next week
		weights[profile.DimensionCompletionLikelihood] += 0.08
		weights[profile.DimensionTimeToDone] += 0.05
		p.Preferences.CompletionFirst = true

	case "B": // Ambitious and done in a few months
		weights[profile.DimensionCompletionLikelihood] -= 0.05
		weights[profile.DimensionTimeToDone] -= 0.05
		weights[profile.DimensionRewardAlignment] += 0.05
		p.Preferences.CompletionFirst = false
	}
}

// applyFailureModeAnswer adjusts weights based on which failure feels worse.
func applyFailureModeAnswer(answers *WizardAnswers, weights map[string]float64, p *profile.Profile) {
	key := answers.GetAnswerKey("failure_mode")

	switch key {
	case "A": // Finishing something nobody cared about
		// User fears wasted effort on wrong thing
		weights[profile.DimensionRewardAlignment] += 0.08
		weights[profile.DimensionCompletionLikelihood] -= 0.03

	case "B": // Never finishing at all
		// User fears abandoning projects
		weights[profile.DimensionCompletionLikelihood] += 0.08
		weights[profile.DimensionSustainability] += 0.05
		p.Preferences.CompletionFirst = true
	}
}

// applyPersistenceAnswer adjusts weights based on how user handles difficulty.
func applyPersistenceAnswer(answers *WizardAnswers, weights map[string]float64, p *profile.Profile) {
	key := answers.GetAnswerKey("persistence")

	switch key {
	case "A": // Push through to the end
		weights[profile.DimensionSustainability] -= 0.05 // Less dependent on motivation
		p.Preferences.PushesThrough = true

	case "B": // Pause and come back later
		weights[profile.DimensionSustainability] += 0.03
		p.Preferences.PushesThrough = false

	case "C": // Move on to something new
		weights[profile.DimensionCompletionLikelihood] += 0.08 // Need simpler projects
		weights[profile.DimensionSustainability] += 0.05       // Need more motivation
		p.Preferences.PushesThrough = false
	}
}

// applyMoneyMattersAnswer adjusts weights based on revenue importance.
func applyMoneyMattersAnswer(answers *WizardAnswers, weights map[string]float64, p *profile.Profile) {
	key := answers.GetAnswerKey("money_matters")

	switch key {
	case "A": // Yes — that's the point
		weights[profile.DimensionRewardAlignment] += 0.10
		weights[profile.DimensionTimeToDone] += 0.05 // Faster = faster revenue
		p.Preferences.MoneyMatters = profile.MoneyMattersYes

	case "B": // Sometimes — depends on the project
		p.Preferences.MoneyMatters = profile.MoneyMattersSometimes

	case "C": // Not really — I do this for other reasons
		weights[profile.DimensionRewardAlignment] -= 0.05
		weights[profile.DimensionSustainability] += 0.05 // Motivation matters more
		p.Preferences.MoneyMatters = profile.MoneyMattersNotReally
	}
}

// applySkillPreferenceAnswer adjusts weights based on familiar vs learning preference.
func applySkillPreferenceAnswer(answers *WizardAnswers, weights map[string]float64, p *profile.Profile) {
	key := answers.GetAnswerKey("skill_preference")

	switch key {
	case "A": // Yes — I like using my strengths
		weights[profile.DimensionSkillFit] += 0.08
		weights[profile.DimensionCompletionLikelihood] += 0.03 // Familiar = faster completion
		p.Preferences.PrefersFamiliar = true

	case "B": // No — I like learning new things
		weights[profile.DimensionSkillFit] -= 0.05
		weights[profile.DimensionRewardAlignment] += 0.03 // Learning itself is rewarding
		p.Preferences.PrefersFamiliar = false

	case "C": // Mix of both
		// Keep balanced, slight preference for familiar
		p.Preferences.PrefersFamiliar = true
	}
}

// GenerateSummary creates a human-readable summary of what was learned.
func GenerateSummary(p *profile.Profile) []string {
	summary := []string{}

	// Completion preference
	if p.Preferences.CompletionFirst {
		summary = append(summary, "You value finishing over ambition")
	} else {
		summary = append(summary, "You're willing to tackle ambitious projects")
	}

	// Money preference
	switch p.Preferences.MoneyMatters {
	case profile.MoneyMattersYes:
		summary = append(summary, "Making money matters to you")
	case profile.MoneyMattersNotReally:
		summary = append(summary, "You're motivated by things other than money")
	}

	// Skill preference
	if p.Preferences.PrefersFamiliar {
		summary = append(summary, "You prefer using skills you already have")
	} else {
		summary = append(summary, "You enjoy learning new things")
	}

	// Persistence
	if p.Preferences.PushesThrough {
		summary = append(summary, "You tend to push through when things get hard")
	}

	// Top priorities
	topDimensions := getTopPriorities(p, 2)
	if len(topDimensions) > 0 {
		summary = append(summary, "Your top priorities: "+formatDimensions(topDimensions))
	}

	return summary
}

// getTopPriorities returns the n highest-weighted dimensions.
func getTopPriorities(p *profile.Profile, n int) []string {
	type dimWeight struct {
		name   string
		weight float64
	}

	dims := []dimWeight{}
	for name, weight := range p.Priorities {
		dims = append(dims, dimWeight{name, weight})
	}

	// Sort by weight descending
	sort.Slice(dims, func(i, j int) bool {
		return dims[i].weight > dims[j].weight
	})

	result := []string{}
	for i := 0; i < n && i < len(dims); i++ {
		result = append(result, dims[i].name)
	}

	return result
}

// formatDimensions converts dimension keys to human-readable names.
func formatDimensions(dims []string) string {
	names := map[string]string{
		profile.DimensionCompletionLikelihood: "finishing projects",
		profile.DimensionSkillFit:             "skill match",
		profile.DimensionTimeToDone:           "fast results",
		profile.DimensionRewardAlignment:      "goal alignment",
		profile.DimensionSustainability:       "staying motivated",
		profile.DimensionAvoidanceFit:         "avoiding pitfalls",
	}

	result := ""
	for i, dim := range dims {
		if i > 0 {
			result += ", "
		}
		if name, ok := names[dim]; ok {
			result += name
		} else {
			result += dim
		}
	}

	return result
}
