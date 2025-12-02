package profile

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// CurrentVersion is the current profile format version
	CurrentVersion = 1

	// DefaultProfileDir is the default directory for brain-salad config
	DefaultProfileDir = ".brain-salad"

	// DefaultProfileFile is the default profile filename
	DefaultProfileFile = "profile.yaml"
)

// DefaultPath returns the default profile path (~/.brain-salad/profile.yaml)
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, DefaultProfileDir, DefaultProfileFile), nil
}

// DefaultDir returns the default profile directory (~/.brain-salad)
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, DefaultProfileDir), nil
}

// Load reads a profile from the specified path.
func Load(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile not found at %s: %w", path, err)
		}
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	if err := Validate(&profile); err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}

	return &profile, nil
}

// Save writes a profile to the specified path.
func Save(profile *Profile, path string) error {
	if err := Validate(profile); err != nil {
		return fmt.Errorf("cannot save invalid profile: %w", err)
	}

	// Ensure directory exists with restricted permissions (user only)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update timestamp
	profile.UpdatedAt = time.Now().UTC()

	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to serialize profile: %w", err)
	}

	// Write with restricted permissions (user read/write only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	return nil
}

// Exists checks if a profile exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExistsAtDefault checks if a profile exists at the default location.
func ExistsAtDefault() bool {
	path, err := DefaultPath()
	if err != nil {
		return false
	}
	return Exists(path)
}

// Validate checks that a profile is valid and internally consistent.
func Validate(p *Profile) error {
	if p == nil {
		return errors.New("profile is nil")
	}

	if p.Version < 1 {
		return errors.New("profile version must be at least 1")
	}

	// Validate priorities exist and sum to 1.0
	if len(p.Priorities) == 0 {
		return errors.New("priorities cannot be empty")
	}

	sum := 0.0
	for dim, weight := range p.Priorities {
		if weight < 0 {
			return fmt.Errorf("priority for %s cannot be negative", dim)
		}
		if weight > 1 {
			return fmt.Errorf("priority for %s cannot exceed 1.0", dim)
		}
		sum += weight
	}

	// Allow small floating point tolerance
	if math.Abs(sum-1.0) > 0.01 {
		return fmt.Errorf("priorities must sum to 1.0, got %.2f", sum)
	}

	// Validate all required dimensions are present
	for _, dim := range AllDimensions() {
		if _, ok := p.Priorities[dim]; !ok {
			return fmt.Errorf("missing required priority: %s", dim)
		}
	}

	// Validate preferences
	validMoneyMatters := map[string]bool{
		MoneyMattersYes:       true,
		MoneyMattersSometimes: true,
		MoneyMattersNotReally: true,
		"":                    true, // Allow empty for backward compatibility
	}
	if !validMoneyMatters[p.Preferences.MoneyMatters] {
		return fmt.Errorf("invalid money_matters value: %s", p.Preferences.MoneyMatters)
	}

	return nil
}

// DefaultProfile creates a profile with sensible default weights.
// This represents a balanced starting point before wizard customization.
func DefaultProfile() *Profile {
	now := time.Now().UTC()
	return &Profile{
		Version: CurrentVersion,
		Priorities: map[string]float64{
			DimensionCompletionLikelihood: 0.20,
			DimensionSkillFit:             0.15,
			DimensionTimeToDone:           0.20,
			DimensionRewardAlignment:      0.20,
			DimensionSustainability:       0.15,
			DimensionAvoidanceFit:         0.10,
		},
		Goals: []string{},
		Avoid: []string{},
		Preferences: Preferences{
			MoneyMatters:    MoneyMattersSometimes,
			PrefersFamiliar: true,
			CompletionFirst: true,
			PushesThrough:   true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NormalizePriorities adjusts priorities to sum to exactly 1.0.
// Useful after manual edits or wizard calculations.
func NormalizePriorities(p *Profile) {
	sum := 0.0
	for _, weight := range p.Priorities {
		sum += weight
	}

	if sum == 0 {
		// Avoid division by zero - reset to defaults
		defaults := DefaultProfile()
		p.Priorities = defaults.Priorities
		return
	}

	for dim := range p.Priorities {
		p.Priorities[dim] = p.Priorities[dim] / sum
	}
}

// GetPriority safely retrieves a priority weight, returning 0 if not found.
func (p *Profile) GetPriority(dimension string) float64 {
	if p.Priorities == nil {
		return 0
	}
	return p.Priorities[dimension]
}

// SetPriority sets a priority weight.
func (p *Profile) SetPriority(dimension string, weight float64) {
	if p.Priorities == nil {
		p.Priorities = make(map[string]float64)
	}
	p.Priorities[dimension] = weight
}

// AddGoal appends a goal if it's not already present.
func (p *Profile) AddGoal(goal string) {
	for _, existing := range p.Goals {
		if existing == goal {
			return
		}
	}
	p.Goals = append(p.Goals, goal)
}

// AddAvoid appends an avoidance item if it's not already present.
func (p *Profile) AddAvoid(item string) {
	for _, existing := range p.Avoid {
		if existing == item {
			return
		}
	}
	p.Avoid = append(p.Avoid, item)
}
