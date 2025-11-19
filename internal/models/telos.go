package models

import (
	"errors"
	"time"
)

// Telos represents the user's goals, strategies, and values.
// This is parsed from the telos.md file.
type Telos struct {
	Problems        []Problem   `json:"problems"`
	Missions        []Mission   `json:"missions"`
	Goals           []Goal      `json:"goals"`
	Challenges      []Challenge `json:"challenges"`
	Strategies      []Strategy  `json:"strategies"`
	Stack           Stack       `json:"stack"`
	FailurePatterns []Pattern   `json:"failure_patterns"`
	LoadedAt        time.Time   `json:"loaded_at"`
}

// Validate validates the telos configuration.
func (t *Telos) Validate() error {
	if len(t.Goals) == 0 {
		return errors.New("at least one goal is required")
	}

	// Validate each problem
	for i, problem := range t.Problems {
		if err := problem.Validate(); err != nil {
			return errors.New("invalid problem at index " + string(rune(i)) + ": " + err.Error())
		}
	}

	// Validate each mission
	for i, mission := range t.Missions {
		if err := mission.Validate(); err != nil {
			return errors.New("invalid mission at index " + string(rune(i)) + ": " + err.Error())
		}
	}

	// Validate each goal
	for i, goal := range t.Goals {
		if err := goal.Validate(); err != nil {
			return errors.New("invalid goal at index " + string(rune(i)) + ": " + err.Error())
		}
	}

	// Validate each challenge
	for i, challenge := range t.Challenges {
		if err := challenge.Validate(); err != nil {
			return errors.New("invalid challenge at index " + string(rune(i)) + ": " + err.Error())
		}
	}

	// Validate each strategy
	for i, strategy := range t.Strategies {
		if err := strategy.Validate(); err != nil {
			return errors.New("invalid strategy at index " + string(rune(i)) + ": " + err.Error())
		}
	}

	// Validate each failure pattern
	for i, pattern := range t.FailurePatterns {
		if err := pattern.Validate(); err != nil {
			return errors.New("invalid failure pattern at index " + string(rune(i)) + ": " + err.Error())
		}
	}

	return nil
}

// Goal represents a user goal with deadline and priority.
type Goal struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Priority    int        `json:"priority"`
}

// Validate validates the goal.
func (g *Goal) Validate() error {
	if g.ID == "" {
		return errors.New("goal ID is required")
	}
	if g.Description == "" {
		return errors.New("goal description is required")
	}
	return nil
}

// Strategy represents a strategic approach or rule.
type Strategy struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Validate validates the strategy.
func (s *Strategy) Validate() error {
	if s.ID == "" {
		return errors.New("strategy ID is required")
	}
	if s.Description == "" {
		return errors.New("strategy description is required")
	}
	return nil
}

// Problem represents a problem or pain point to solve.
type Problem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Validate validates the problem.
func (p *Problem) Validate() error {
	if p.ID == "" {
		return errors.New("problem ID is required")
	}
	if p.Description == "" {
		return errors.New("problem description is required")
	}
	return nil
}

// Mission represents a high-level mission or purpose.
type Mission struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Validate validates the mission.
func (m *Mission) Validate() error {
	if m.ID == "" {
		return errors.New("mission ID is required")
	}
	if m.Description == "" {
		return errors.New("mission description is required")
	}
	return nil
}

// Challenge represents a challenge or constraint.
type Challenge struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Validate validates the challenge.
func (c *Challenge) Validate() error {
	if c.ID == "" {
		return errors.New("challenge ID is required")
	}
	if c.Description == "" {
		return errors.New("challenge description is required")
	}
	return nil
}

// Stack represents technology preferences and current stack.
type Stack struct {
	Primary   []string `json:"primary"`
	Secondary []string `json:"secondary"`
}

// Pattern represents a failure pattern to avoid.
// These are anti-patterns or challenges the user wants to avoid.
type Pattern struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

// Validate validates the pattern.
func (p *Pattern) Validate() error {
	if p.Name == "" {
		return errors.New("pattern name is required")
	}
	if p.Description == "" {
		return errors.New("pattern description is required")
	}
	return nil
}
