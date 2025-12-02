// Package wizard provides an interactive discovery wizard for creating user profiles.
package wizard

// Question represents a single discovery question.
type Question struct {
	ID          string
	Text        string
	Subtext     string // Optional clarifying text
	Options     []Option
	AllowCustom bool // Whether user can type a custom response
}

// Option represents a possible answer to a question.
type Option struct {
	Key         string // A, B, C, etc.
	Text        string
	Description string // Optional longer description
}

// Questions returns the ordered list of discovery questions.
func Questions() []Question {
	return []Question{
		{
			ID:   "project_style",
			Text: "You have two project ideas. Which do you start?",
			Options: []Option{
				{Key: "A", Text: "Simple and done by next week"},
				{Key: "B", Text: "Ambitious and done in a few months"},
			},
		},
		{
			ID:   "failure_mode",
			Text: "Which feels worse?",
			Options: []Option{
				{Key: "A", Text: "Finishing something nobody cared about"},
				{Key: "B", Text: "Never finishing at all"},
			},
		},
		{
			ID:   "persistence",
			Text: "When a project gets hard, you usually:",
			Options: []Option{
				{Key: "A", Text: "Push through to the end"},
				{Key: "B", Text: "Pause and come back later"},
				{Key: "C", Text: "Move on to something new"},
			},
		},
		{
			ID:   "money_matters",
			Text: "Does making money from your projects matter?",
			Options: []Option{
				{Key: "A", Text: "Yes — that's the point"},
				{Key: "B", Text: "Sometimes — depends on the project"},
				{Key: "C", Text: "Not really — I do this for other reasons"},
			},
		},
		{
			ID:   "skill_preference",
			Text: "Do you prefer working with what you already know?",
			Options: []Option{
				{Key: "A", Text: "Yes — I like using my strengths"},
				{Key: "B", Text: "No — I like learning new things"},
				{Key: "C", Text: "Mix of both"},
			},
		},
	}
}

// GoalPrompt returns the prompt for collecting user goals.
type GoalPrompt struct {
	Text       string
	Subtext    string
	MaxGoals   int
	MinGoals   int
	Placeholder string
}

// GetGoalPrompt returns the goal collection prompt.
func GetGoalPrompt() GoalPrompt {
	return GoalPrompt{
		Text:       "What are you trying to achieve right now?",
		Subtext:    "(write 1-3 goals — anything goes)",
		MaxGoals:   3,
		MinGoals:   1,
		Placeholder: "sell pottery at farmer's market",
	}
}

// GoalExamples returns diverse examples to show the system works for any domain.
func GoalExamples() []string {
	return []string{
		"sell pottery at farmer's market",
		"launch my app this month",
		"finish writing my novel",
		"start a consulting side hustle",
		"learn woodworking",
	}
}

// AvoidPrompt returns the prompt for collecting things to avoid.
type AvoidPrompt struct {
	Text       string
	Subtext    string
	Optional   bool
	Placeholder string
}

// GetAvoidPrompt returns the avoidance collection prompt.
func GetAvoidPrompt() AvoidPrompt {
	return AvoidPrompt{
		Text:       "Anything you want to avoid?",
		Subtext:    "(optional — separate with commas, or press Enter to skip)",
		Optional:   true,
		Placeholder: "large inventory, wholesale accounts",
	}
}

// AvoidExamples returns diverse examples for things to avoid.
func AvoidExamples() []string {
	return []string{
		"large upfront costs",
		"managing employees",
		"technical complexity",
		"long-term commitments",
	}
}

// Answer represents a user's answer to a question.
type Answer struct {
	QuestionID string
	OptionKey  string // A, B, C, etc.
	CustomText string // If AllowCustom was true and user typed something
}

// WizardAnswers collects all answers from the wizard flow.
type WizardAnswers struct {
	Answers []Answer
	Goals   []string
	Avoid   []string
}

// GetAnswer retrieves an answer by question ID.
func (wa *WizardAnswers) GetAnswer(questionID string) *Answer {
	for _, a := range wa.Answers {
		if a.QuestionID == questionID {
			return &a
		}
	}
	return nil
}

// GetAnswerKey returns just the option key for a question, or empty string if not found.
func (wa *WizardAnswers) GetAnswerKey(questionID string) string {
	a := wa.GetAnswer(questionID)
	if a == nil {
		return ""
	}
	return a.OptionKey
}
