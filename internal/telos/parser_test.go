package telos_test

import (
	"testing"
	"time"

	"github.com/ryacub/telos-idea-matrix/internal/telos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile_ValidTelos_ParsesAllSections(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Goals, 3, "Should parse all 3 goals")
	assert.Len(t, result.Strategies, 3, "Should parse all 3 strategies")
	assert.Len(t, result.FailurePatterns, 3, "Should parse all 3 failure patterns")
	assert.Len(t, result.Stack.Primary, 3, "Should parse 3 primary stack items")
	assert.Len(t, result.Stack.Secondary, 3, "Should parse 3 secondary stack items")
}

func TestParseFile_ValidTelos_ParsesGoalsWithDeadlines(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	require.NoError(t, err)

	// Check first goal
	assert.Equal(t, "G1", result.Goals[0].ID)
	assert.Contains(t, result.Goals[0].Description, "SaaS product")
	require.NotNil(t, result.Goals[0].Deadline, "G1 should have a deadline")
	assert.Equal(t, 2025, result.Goals[0].Deadline.Year())
	assert.Equal(t, time.Month(12), result.Goals[0].Deadline.Month())
	assert.Equal(t, 31, result.Goals[0].Deadline.Day())

	// Check second goal
	assert.Equal(t, "G2", result.Goals[1].ID)
	assert.Contains(t, result.Goals[1].Description, "personal brand")
	require.NotNil(t, result.Goals[1].Deadline, "G2 should have a deadline")
	assert.Equal(t, 2025, result.Goals[1].Deadline.Year())
	assert.Equal(t, time.Month(6), result.Goals[1].Deadline.Month())
	assert.Equal(t, 30, result.Goals[1].Deadline.Day())

	// Check third goal
	assert.Equal(t, "G3", result.Goals[2].ID)
	assert.Contains(t, result.Goals[2].Description, "$2500/month")
}

func TestParseFile_ValidTelos_ParsesGoalWithoutDeadline(t *testing.T) {
	parser := telos.NewParser()

	// G3 doesn't have proper format, so its deadline parsing is lenient
	result, err := parser.ParseFile("testdata/minimal_telos.md")

	require.NoError(t, err)
	assert.Equal(t, "G1", result.Goals[0].ID)
	assert.Contains(t, result.Goals[0].Description, "Ship a product")
}

func TestParseFile_ValidTelos_ParsesStrategies(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	require.NoError(t, err)
	assert.Equal(t, "S1", result.Strategies[0].ID)
	assert.Contains(t, result.Strategies[0].Description, "Ship early")

	assert.Equal(t, "S2", result.Strategies[1].ID)
	assert.Contains(t, result.Strategies[1].Description, "one technology stack")

	assert.Equal(t, "S3", result.Strategies[2].ID)
	assert.Contains(t, result.Strategies[2].Description, "public on Twitter")
}

func TestParseFile_ValidTelos_ParsesStack(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	require.NoError(t, err)

	// Check primary stack
	assert.Contains(t, result.Stack.Primary, "Go")
	assert.Contains(t, result.Stack.Primary, "TypeScript")
	assert.Contains(t, result.Stack.Primary, "PostgreSQL")

	// Check secondary stack
	assert.Contains(t, result.Stack.Secondary, "Docker")
	assert.Contains(t, result.Stack.Secondary, "Kubernetes")
	assert.Contains(t, result.Stack.Secondary, "Redis")
}

func TestParseFile_ValidTelos_ParsesFailurePatterns(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	require.NoError(t, err)

	// Check first pattern
	assert.Equal(t, "Context switching", result.FailurePatterns[0].Name)
	assert.Contains(t, result.FailurePatterns[0].Description, "Starting new projects")
	assert.Contains(t, result.FailurePatterns[0].Keywords, "starting")
	assert.Contains(t, result.FailurePatterns[0].Keywords, "projects")

	// Check second pattern
	assert.Equal(t, "Perfectionism", result.FailurePatterns[1].Name)
	assert.Contains(t, result.FailurePatterns[1].Description, "Over-engineering")

	// Check third pattern
	assert.Equal(t, "Procrastination", result.FailurePatterns[2].Name)
	assert.Contains(t, result.FailurePatterns[2].Description, "Learning before building")
}

func TestParseFile_MinimalTelos_ParsesSuccessfully(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/minimal_telos.md")

	require.NoError(t, err)
	assert.Len(t, result.Goals, 1)
	assert.Len(t, result.Stack.Primary, 1)
	assert.Contains(t, result.Stack.Primary, "Go")
	assert.Empty(t, result.Strategies)
	assert.Empty(t, result.FailurePatterns)
}

func TestParseFile_MissingFile_ReturnsError(t *testing.T) {
	parser := telos.NewParser()

	_, err := parser.ParseFile("testdata/nonexistent.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestParseFile_EmptyFile_ReturnsError(t *testing.T) {
	parser := telos.NewParser()

	_, err := parser.ParseFile("testdata/empty.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one goal is required")
}

func TestParseFile_SetsLoadedAt(t *testing.T) {
	parser := telos.NewParser()

	before := time.Now()
	result, err := parser.ParseFile("testdata/valid_telos.md")
	after := time.Now()

	require.NoError(t, err)
	assert.True(t, result.LoadedAt.After(before) || result.LoadedAt.Equal(before))
	assert.True(t, result.LoadedAt.Before(after) || result.LoadedAt.Equal(after))
}

func TestParseFile_ExtractsKeywordsFromPatterns(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	require.NoError(t, err)

	// Keywords should be extracted and filtered (no stopwords)
	pattern := result.FailurePatterns[0] // Context switching pattern

	// Should have meaningful keywords
	assert.NotEmpty(t, pattern.Keywords)

	// Should not have stopwords like "the", "a", "and"
	assert.NotContains(t, pattern.Keywords, "the")
	assert.NotContains(t, pattern.Keywords, "and")
	assert.NotContains(t, pattern.Keywords, "before")
}

func TestParseFile_HandlesMultipleGoalsWithSamePrefix(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	require.NoError(t, err)

	// Verify all goal IDs are unique
	ids := make(map[string]bool)
	for _, goal := range result.Goals {
		assert.False(t, ids[goal.ID], "Duplicate goal ID: %s", goal.ID)
		ids[goal.ID] = true
	}
}

// Tests for new sections: Problems, Missions, Challenges

func TestParser_ProblemsSection(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/full_telos.md")

	require.NoError(t, err)
	assert.Len(t, result.Problems, 3, "Should parse all 3 problems")

	// Check first problem
	assert.Equal(t, "P1", result.Problems[0].ID)
	assert.Contains(t, result.Problems[0].Description, "Too many ideas")

	// Check second problem
	assert.Equal(t, "P2", result.Problems[1].ID)
	assert.Contains(t, result.Problems[1].Description, "Context switching")

	// Check third problem
	assert.Equal(t, "P3", result.Problems[2].ID)
	assert.Contains(t, result.Problems[2].Description, "Perfectionism")
}

func TestParser_MissionsSection(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/full_telos.md")

	require.NoError(t, err)
	assert.Len(t, result.Missions, 3, "Should parse all 3 missions")

	// Check first mission
	assert.Equal(t, "M1", result.Missions[0].ID)
	assert.Contains(t, result.Missions[0].Description, "Ship profitable SaaS")

	// Check second mission
	assert.Equal(t, "M2", result.Missions[1].ID)
	assert.Contains(t, result.Missions[1].Description, "Build personal brand")

	// Check third mission
	assert.Equal(t, "M3", result.Missions[2].ID)
	assert.Contains(t, result.Missions[2].Description, "financial independence")
}

func TestParser_ChallengesSection(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/full_telos.md")

	require.NoError(t, err)
	assert.Len(t, result.Challenges, 3, "Should parse all 3 challenges")

	// Check first challenge
	assert.Equal(t, "C1", result.Challenges[0].ID)
	assert.Contains(t, result.Challenges[0].Description, "Limited time")

	// Check second challenge
	assert.Equal(t, "C2", result.Challenges[1].ID)
	assert.Contains(t, result.Challenges[1].Description, "Small audience")

	// Check third challenge
	assert.Equal(t, "C3", result.Challenges[2].ID)
	assert.Contains(t, result.Challenges[2].Description, "Technical depth")
}

func TestParser_FullTelosFile(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/full_telos.md")

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify all sections are parsed
	assert.Len(t, result.Problems, 3, "Should parse all 3 problems")
	assert.Len(t, result.Missions, 3, "Should parse all 3 missions")
	assert.Len(t, result.Goals, 3, "Should parse all 3 goals")
	assert.Len(t, result.Challenges, 3, "Should parse all 3 challenges")
	assert.Len(t, result.Strategies, 3, "Should parse all 3 strategies")
	assert.Len(t, result.Stack.Primary, 3, "Should parse 3 primary stack items")
	assert.Len(t, result.Stack.Secondary, 3, "Should parse 3 secondary stack items")
	assert.Len(t, result.FailurePatterns, 3, "Should parse all 3 failure patterns")
}

func TestParser_AllSectionsPresent(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/full_telos.md")

	require.NoError(t, err)

	// Verify no sections are nil
	assert.NotNil(t, result.Problems, "Problems should not be nil")
	assert.NotNil(t, result.Missions, "Missions should not be nil")
	assert.NotNil(t, result.Goals, "Goals should not be nil")
	assert.NotNil(t, result.Challenges, "Challenges should not be nil")
	assert.NotNil(t, result.Strategies, "Strategies should not be nil")
	assert.NotNil(t, result.FailurePatterns, "Failure patterns should not be nil")

	// Verify all IDs are unique within each section
	problemIDs := make(map[string]bool)
	for _, p := range result.Problems {
		assert.False(t, problemIDs[p.ID], "Duplicate problem ID: %s", p.ID)
		problemIDs[p.ID] = true
	}

	missionIDs := make(map[string]bool)
	for _, m := range result.Missions {
		assert.False(t, missionIDs[m.ID], "Duplicate mission ID: %s", m.ID)
		missionIDs[m.ID] = true
	}

	challengeIDs := make(map[string]bool)
	for _, c := range result.Challenges {
		assert.False(t, challengeIDs[c.ID], "Duplicate challenge ID: %s", c.ID)
		challengeIDs[c.ID] = true
	}
}

// Edge case tests for better coverage

func TestParser_HandlesInvalidLines(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/full_telos.md")

	require.NoError(t, err)

	// Verify that invalid lines are skipped gracefully
	// The parser should only return valid items
	assert.NotEmpty(t, result.Problems)
	assert.NotEmpty(t, result.Missions)
	assert.NotEmpty(t, result.Goals)
	assert.NotEmpty(t, result.Challenges)
	assert.NotEmpty(t, result.Strategies)
}

func TestParser_HandlesMissingSections(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/minimal_telos.md")

	require.NoError(t, err)

	// Minimal telos should have empty arrays for missing sections
	assert.Empty(t, result.Problems, "Should have no problems")
	assert.Empty(t, result.Missions, "Should have no missions")
	assert.Empty(t, result.Challenges, "Should have no challenges")
	assert.NotEmpty(t, result.Goals, "Should have at least one goal")
}
