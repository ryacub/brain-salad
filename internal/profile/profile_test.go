package profile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProfile_HasValidWeights(t *testing.T) {
	p := DefaultProfile()

	// Weights should sum to 1.0
	sum := 0.0
	for _, weight := range p.Priorities {
		sum += weight
	}
	assert.InDelta(t, 1.0, sum, 0.01, "priorities should sum to 1.0")
}

func TestDefaultProfile_HasAllDimensions(t *testing.T) {
	p := DefaultProfile()

	for _, dim := range AllDimensions() {
		_, ok := p.Priorities[dim]
		assert.True(t, ok, "missing dimension: %s", dim)
	}
}

func TestDefaultProfile_HasVersion(t *testing.T) {
	p := DefaultProfile()
	assert.Equal(t, CurrentVersion, p.Version)
}

func TestValidate_ValidProfile_NoError(t *testing.T) {
	p := DefaultProfile()
	err := Validate(p)
	assert.NoError(t, err)
}

func TestValidate_NilProfile_ReturnsError(t *testing.T) {
	err := Validate(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestValidate_ZeroVersion_ReturnsError(t *testing.T) {
	p := DefaultProfile()
	p.Version = 0

	err := Validate(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestValidate_EmptyPriorities_ReturnsError(t *testing.T) {
	p := DefaultProfile()
	p.Priorities = map[string]float64{}

	err := Validate(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidate_NegativeWeight_ReturnsError(t *testing.T) {
	p := DefaultProfile()
	p.Priorities[DimensionSkillFit] = -0.1

	err := Validate(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "negative")
}

func TestValidate_WeightOver1_ReturnsError(t *testing.T) {
	p := DefaultProfile()
	p.Priorities[DimensionSkillFit] = 1.5

	err := Validate(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceed")
}

func TestValidate_WeightsDontSumTo1_ReturnsError(t *testing.T) {
	p := DefaultProfile()
	// Double one weight without adjusting others
	p.Priorities[DimensionSkillFit] = 0.5

	err := Validate(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sum to 1.0")
}

func TestValidate_MissingDimension_ReturnsError(t *testing.T) {
	p := DefaultProfile()
	// Delete one and redistribute weight to maintain sum of 1.0
	deletedWeight := p.Priorities[DimensionAvoidanceFit]
	delete(p.Priorities, DimensionAvoidanceFit)
	p.Priorities[DimensionSkillFit] += deletedWeight

	err := Validate(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestValidate_InvalidMoneyMatters_ReturnsError(t *testing.T) {
	p := DefaultProfile()
	p.Preferences.MoneyMatters = "invalid"

	err := Validate(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "money_matters")
}

func TestNormalizePriorities_SumsTo1(t *testing.T) {
	p := &Profile{
		Priorities: map[string]float64{
			DimensionCompletionLikelihood: 2.0,
			DimensionSkillFit:             2.0,
			DimensionTimeToDone:           2.0,
			DimensionRewardAlignment:      2.0,
			DimensionSustainability:       1.0,
			DimensionAvoidanceFit:         1.0,
		},
	}

	NormalizePriorities(p)

	sum := 0.0
	for _, weight := range p.Priorities {
		sum += weight
	}
	assert.InDelta(t, 1.0, sum, 0.001)
}

func TestNormalizePriorities_ZeroSum_ResetsToDefaults(t *testing.T) {
	p := &Profile{
		Priorities: map[string]float64{
			DimensionCompletionLikelihood: 0,
			DimensionSkillFit:             0,
			DimensionTimeToDone:           0,
			DimensionRewardAlignment:      0,
			DimensionSustainability:       0,
			DimensionAvoidanceFit:         0,
		},
	}

	NormalizePriorities(p)

	// Should have non-zero weights now
	assert.Greater(t, p.Priorities[DimensionCompletionLikelihood], 0.0)
}

func TestGetPriority_ExistingDimension_ReturnsWeight(t *testing.T) {
	p := DefaultProfile()
	weight := p.GetPriority(DimensionSkillFit)
	assert.Greater(t, weight, 0.0)
}

func TestGetPriority_NonExistentDimension_ReturnsZero(t *testing.T) {
	p := DefaultProfile()
	weight := p.GetPriority("nonexistent")
	assert.Equal(t, 0.0, weight)
}

func TestGetPriority_NilPriorities_ReturnsZero(t *testing.T) {
	p := &Profile{}
	weight := p.GetPriority(DimensionSkillFit)
	assert.Equal(t, 0.0, weight)
}

func TestSetPriority_SetsValue(t *testing.T) {
	p := &Profile{}
	p.SetPriority(DimensionSkillFit, 0.5)
	assert.Equal(t, 0.5, p.Priorities[DimensionSkillFit])
}

func TestAddGoal_AddsNewGoal(t *testing.T) {
	p := &Profile{Goals: []string{}}
	p.AddGoal("sell pottery")
	assert.Contains(t, p.Goals, "sell pottery")
}

func TestAddGoal_DoesNotDuplicate(t *testing.T) {
	p := &Profile{Goals: []string{"sell pottery"}}
	p.AddGoal("sell pottery")
	assert.Len(t, p.Goals, 1)
}

func TestAddAvoid_AddsNewItem(t *testing.T) {
	p := &Profile{Avoid: []string{}}
	p.AddAvoid("wholesale")
	assert.Contains(t, p.Avoid, "wholesale")
}

func TestAddAvoid_DoesNotDuplicate(t *testing.T) {
	p := &Profile{Avoid: []string{"wholesale"}}
	p.AddAvoid("wholesale")
	assert.Len(t, p.Avoid, 1)
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profile.yaml")

	// Create profile
	original := DefaultProfile()
	original.Goals = []string{"sell pottery", "finish projects"}
	original.Avoid = []string{"wholesale", "large inventory"}
	original.Preferences.MoneyMatters = MoneyMattersYes

	// Save
	err := Save(original, path)
	require.NoError(t, err)

	// Load
	loaded, err := Load(path)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.Version, loaded.Version)
	assert.Equal(t, original.Goals, loaded.Goals)
	assert.Equal(t, original.Avoid, loaded.Avoid)
	assert.Equal(t, original.Preferences.MoneyMatters, loaded.Preferences.MoneyMatters)

	for dim, weight := range original.Priorities {
		assert.InDelta(t, weight, loaded.Priorities[dim], 0.001, "weight mismatch for %s", dim)
	}
}

func TestLoad_NonExistentFile_ReturnsError(t *testing.T) {
	_, err := Load("/nonexistent/path/profile.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profile.yaml")

	// Write invalid YAML
	err := os.WriteFile(path, []byte("not: valid: yaml: {{{{"), 0600)
	require.NoError(t, err)

	_, err = Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "profile.yaml")

	p := DefaultProfile()
	err := Save(p, path)
	require.NoError(t, err)

	// Directory should exist
	_, err = os.Stat(filepath.Dir(path))
	assert.NoError(t, err)
}

func TestSave_SetsUpdatedAt(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profile.yaml")

	p := DefaultProfile()
	p.UpdatedAt = time.Time{} // Zero time

	err := Save(p, path)
	require.NoError(t, err)

	assert.False(t, p.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

func TestExists_ExistingFile_ReturnsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profile.yaml")

	p := DefaultProfile()
	err := Save(p, path)
	require.NoError(t, err)

	assert.True(t, Exists(path))
}

func TestExists_NonExistentFile_ReturnsFalse(t *testing.T) {
	assert.False(t, Exists("/nonexistent/profile.yaml"))
}

func TestAllDimensions_Returns6Dimensions(t *testing.T) {
	dims := AllDimensions()
	assert.Len(t, dims, 6)
}

func TestDimensionDescriptions_HasAllDimensions(t *testing.T) {
	for _, dim := range AllDimensions() {
		desc, ok := DimensionDescriptions[dim]
		assert.True(t, ok, "missing description for %s", dim)
		assert.NotEmpty(t, desc)
	}
}

func TestDimensionMaxPoints_SumsTo10(t *testing.T) {
	sum := 0.0
	for _, points := range DimensionMaxPoints {
		sum += points
	}
	assert.Equal(t, 10.0, sum)
}
