// Package profile provides user preference management for the universal scoring system.
package profile

import "time"

// Profile represents user preferences and priorities for idea evaluation.
// This is the core configuration that drives the universal scoring engine.
type Profile struct {
	// Version for future migrations
	Version int `yaml:"version" json:"version"`

	// Priorities are scoring dimension weights (must sum to 1.0)
	// Keys: completion_likelihood, skill_fit, time_to_done, reward_alignment, sustainability, avoidance_fit
	Priorities map[string]float64 `yaml:"priorities" json:"priorities"`

	// Goals are plain English statements of what the user wants to achieve
	Goals []string `yaml:"goals" json:"goals"`

	// Avoid lists things the user wants to steer clear of
	Avoid []string `yaml:"avoid" json:"avoid"`

	// Preferences captured from discovery wizard
	Preferences Preferences `yaml:"preferences" json:"preferences"`

	// CreatedAt tracks when the profile was created
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`

	// UpdatedAt tracks the last modification
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
}

// Preferences holds user tendencies discovered through the wizard.
type Preferences struct {
	// MoneyMatters indicates how important revenue is: "yes", "sometimes", "not_really"
	MoneyMatters string `yaml:"money_matters" json:"money_matters"`

	// PrefersFamiliar indicates if user prefers known skills over learning
	PrefersFamiliar bool `yaml:"prefers_familiar" json:"prefers_familiar"`

	// CompletionFirst indicates if user prioritizes finishing over ambition
	CompletionFirst bool `yaml:"completion_first" json:"completion_first"`

	// PushesThrough indicates if user typically finishes difficult projects
	PushesThrough bool `yaml:"pushes_through" json:"pushes_through"`
}

// Dimension names as constants for consistency
const (
	DimensionCompletionLikelihood = "completion_likelihood"
	DimensionSkillFit             = "skill_fit"
	DimensionTimeToDone           = "time_to_done"
	DimensionRewardAlignment      = "reward_alignment"
	DimensionSustainability       = "sustainability"
	DimensionAvoidanceFit         = "avoidance_fit"
)

// AllDimensions returns all scoring dimension names in order.
func AllDimensions() []string {
	return []string{
		DimensionCompletionLikelihood,
		DimensionSkillFit,
		DimensionTimeToDone,
		DimensionRewardAlignment,
		DimensionSustainability,
		DimensionAvoidanceFit,
	}
}

// DimensionDescriptions maps dimension names to human-readable questions.
var DimensionDescriptions = map[string]string{
	DimensionCompletionLikelihood: "Will I actually finish this?",
	DimensionSkillFit:             "Can I do this with what I know?",
	DimensionTimeToDone:           "How long until it's real?",
	DimensionRewardAlignment:      "Does this give me what I want?",
	DimensionSustainability:       "Will I stay motivated?",
	DimensionAvoidanceFit:         "Does this dodge my pitfalls?",
}

// DimensionMaxPoints defines the maximum score for each dimension.
// Total: 10.0 points (same scale as legacy system)
var DimensionMaxPoints = map[string]float64{
	DimensionCompletionLikelihood: 2.0,
	DimensionSkillFit:             2.0,
	DimensionTimeToDone:           2.0,
	DimensionRewardAlignment:      2.0,
	DimensionSustainability:       1.0,
	DimensionAvoidanceFit:         1.0,
}

// MoneyMatters constants
const (
	MoneyMattersYes       = "yes"
	MoneyMattersSometimes = "sometimes"
	MoneyMattersNotReally = "not_really"
)
