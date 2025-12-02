package wizard

import (
	"testing"

	"github.com/rayyacub/telos-idea-matrix/internal/profile"
	"github.com/stretchr/testify/assert"
)

func TestMapAnswersToProfile_ReturnsValidProfile(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "project_style", OptionKey: "A"},
			{QuestionID: "failure_mode", OptionKey: "B"},
			{QuestionID: "persistence", OptionKey: "A"},
			{QuestionID: "money_matters", OptionKey: "A"},
			{QuestionID: "skill_preference", OptionKey: "A"},
		},
		Goals: []string{"sell pottery"},
		Avoid: []string{"wholesale"},
	}

	p := MapAnswersToProfile(answers)

	assert.NotNil(t, p)
	assert.Equal(t, 1, p.Version)
	err := profile.Validate(p)
	assert.NoError(t, err)
}

func TestMapAnswersToProfile_SetsGoals(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{},
		Goals:   []string{"sell pottery", "finish projects"},
		Avoid:   []string{},
	}

	p := MapAnswersToProfile(answers)

	assert.Equal(t, []string{"sell pottery", "finish projects"}, p.Goals)
}

func TestMapAnswersToProfile_SetsAvoid(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{},
		Goals:   []string{},
		Avoid:   []string{"wholesale", "inventory"},
	}

	p := MapAnswersToProfile(answers)

	assert.Equal(t, []string{"wholesale", "inventory"}, p.Avoid)
}

func TestMapAnswersToProfile_SimpleProject_BoostsCompletion(t *testing.T) {
	simpleAnswers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "project_style", OptionKey: "A"}, // Simple
		},
	}

	ambitiousAnswers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "project_style", OptionKey: "B"}, // Ambitious
		},
	}

	simpleProfile := MapAnswersToProfile(simpleAnswers)
	ambitiousProfile := MapAnswersToProfile(ambitiousAnswers)

	assert.Greater(t,
		simpleProfile.Priorities[profile.DimensionCompletionLikelihood],
		ambitiousProfile.Priorities[profile.DimensionCompletionLikelihood],
	)
}

func TestMapAnswersToProfile_FearsNotFinishing_BoostsCompletion(t *testing.T) {
	fearsNotFinishing := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "failure_mode", OptionKey: "B"}, // Never finishing
		},
	}

	fearsWastedEffort := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "failure_mode", OptionKey: "A"}, // Nobody cared
		},
	}

	notFinishingProfile := MapAnswersToProfile(fearsNotFinishing)
	wastedEffortProfile := MapAnswersToProfile(fearsWastedEffort)

	assert.Greater(t,
		notFinishingProfile.Priorities[profile.DimensionCompletionLikelihood],
		wastedEffortProfile.Priorities[profile.DimensionCompletionLikelihood],
	)
}

func TestMapAnswersToProfile_MoneyMattersYes_BoostsReward(t *testing.T) {
	moneyMatters := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "money_matters", OptionKey: "A"}, // Yes
		},
	}

	moneyNotImportant := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "money_matters", OptionKey: "C"}, // Not really
		},
	}

	moneyProfile := MapAnswersToProfile(moneyMatters)
	noMoneyProfile := MapAnswersToProfile(moneyNotImportant)

	assert.Greater(t,
		moneyProfile.Priorities[profile.DimensionRewardAlignment],
		noMoneyProfile.Priorities[profile.DimensionRewardAlignment],
	)
}

func TestMapAnswersToProfile_MoneyMattersYes_SetsPreference(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "money_matters", OptionKey: "A"},
		},
	}

	p := MapAnswersToProfile(answers)

	assert.Equal(t, profile.MoneyMattersYes, p.Preferences.MoneyMatters)
}

func TestMapAnswersToProfile_MoneyMattersSometimes_SetsPreference(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "money_matters", OptionKey: "B"},
		},
	}

	p := MapAnswersToProfile(answers)

	assert.Equal(t, profile.MoneyMattersSometimes, p.Preferences.MoneyMatters)
}

func TestMapAnswersToProfile_MoneyMattersNotReally_SetsPreference(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "money_matters", OptionKey: "C"},
		},
	}

	p := MapAnswersToProfile(answers)

	assert.Equal(t, profile.MoneyMattersNotReally, p.Preferences.MoneyMatters)
}

func TestMapAnswersToProfile_PrefersFamiliar_SetsPreference(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "skill_preference", OptionKey: "A"}, // Yes, familiar
		},
	}

	p := MapAnswersToProfile(answers)

	assert.True(t, p.Preferences.PrefersFamiliar)
}

func TestMapAnswersToProfile_PrefersLearning_SetsPreference(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "skill_preference", OptionKey: "B"}, // No, learning
		},
	}

	p := MapAnswersToProfile(answers)

	assert.False(t, p.Preferences.PrefersFamiliar)
}

func TestMapAnswersToProfile_PushesThrough_SetsPreference(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "persistence", OptionKey: "A"}, // Push through
		},
	}

	p := MapAnswersToProfile(answers)

	assert.True(t, p.Preferences.PushesThrough)
}

func TestMapAnswersToProfile_MovesOn_SetsPreference(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "persistence", OptionKey: "C"}, // Move on
		},
	}

	p := MapAnswersToProfile(answers)

	assert.False(t, p.Preferences.PushesThrough)
}

func TestMapAnswersToProfile_PrioritiesSumTo1(t *testing.T) {
	// Test various answer combinations
	combinations := [][]Answer{
		{{QuestionID: "project_style", OptionKey: "A"}},
		{{QuestionID: "project_style", OptionKey: "B"}},
		{
			{QuestionID: "project_style", OptionKey: "A"},
			{QuestionID: "failure_mode", OptionKey: "B"},
			{QuestionID: "persistence", OptionKey: "C"},
			{QuestionID: "money_matters", OptionKey: "A"},
			{QuestionID: "skill_preference", OptionKey: "B"},
		},
	}

	for _, combo := range combinations {
		answers := &WizardAnswers{Answers: combo}
		p := MapAnswersToProfile(answers)

		sum := 0.0
		for _, weight := range p.Priorities {
			sum += weight
		}

		assert.InDelta(t, 1.0, sum, 0.01, "priorities should sum to 1.0")
	}
}

func TestGenerateSummary_ReturnsStrings(t *testing.T) {
	p := profile.DefaultProfile()
	p.Preferences.CompletionFirst = true
	p.Preferences.MoneyMatters = profile.MoneyMattersYes

	summary := GenerateSummary(p)

	assert.NotEmpty(t, summary)
	assert.IsType(t, []string{}, summary)
}

func TestGenerateSummary_IncludesCompletionPreference(t *testing.T) {
	p := profile.DefaultProfile()
	p.Preferences.CompletionFirst = true

	summary := GenerateSummary(p)

	found := false
	for _, line := range summary {
		if line == "You value finishing over ambition" {
			found = true
			break
		}
	}
	assert.True(t, found, "should include completion preference")
}

func TestGenerateSummary_IncludesMoneyPreference(t *testing.T) {
	p := profile.DefaultProfile()
	p.Preferences.MoneyMatters = profile.MoneyMattersYes

	summary := GenerateSummary(p)

	found := false
	for _, line := range summary {
		if line == "Making money matters to you" {
			found = true
			break
		}
	}
	assert.True(t, found, "should include money preference")
}

func TestWizardAnswers_GetAnswer_Found(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "test_q", OptionKey: "A"},
		},
	}

	a := answers.GetAnswer("test_q")
	assert.NotNil(t, a)
	assert.Equal(t, "A", a.OptionKey)
}

func TestWizardAnswers_GetAnswer_NotFound(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{},
	}

	a := answers.GetAnswer("nonexistent")
	assert.Nil(t, a)
}

func TestWizardAnswers_GetAnswerKey_Found(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{
			{QuestionID: "test_q", OptionKey: "B"},
		},
	}

	key := answers.GetAnswerKey("test_q")
	assert.Equal(t, "B", key)
}

func TestWizardAnswers_GetAnswerKey_NotFound(t *testing.T) {
	answers := &WizardAnswers{
		Answers: []Answer{},
	}

	key := answers.GetAnswerKey("nonexistent")
	assert.Equal(t, "", key)
}
