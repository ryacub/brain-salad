package patterns_test

import (
	"testing"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/patterns"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestTelos(t *testing.T) *models.Telos {
	t.Helper()
	parser := telos.NewParser()
	telosData, err := parser.ParseFile("../scoring/testdata/test_telos.md")
	require.NoError(t, err)
	return telosData
}

// ============================================================================
// CONTEXT-SWITCHING DETECTION
// ============================================================================

func TestDetector_DetectPatterns_ContextSwitching_Negative(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	// Test each stack-switching keyword
	testCases := []struct {
		name string
		idea string
	}{
		{"Rust", "Build a game engine in Rust"},
		{"JavaScript", "Create a JavaScript framework"},
		{"TypeScript", "Build TypeScript library"},
		{"React", "Make a React Native mobile app"},
		{"Flutter", "Develop Flutter mobile app"},
		{"Swift", "iOS app in Swift"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detected := detector.DetectPatterns(tc.idea)

			// Should detect context-switching pattern
			found := false
			for _, p := range detected {
				if p.Name == "Context switching" {
					found = true
					assert.Equal(t, "high", p.Severity, "Context switching should be high severity")
					assert.Contains(t, p.Description, "risk", "Should mention risk")
					break
				}
			}
			assert.True(t, found, "Should detect context-switching pattern for %s", tc.name)
		})
	}
}

func TestDetector_DetectPatterns_ContextSwitching_Positive(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Build automation using Python and LangChain"
	detected := detector.DetectPatterns(idea)

	// Should detect positive stack alignment
	found := false
	for _, p := range detected {
		if p.Name == "Context switching" {
			found = true
			assert.Equal(t, "low", p.Severity, "Stack alignment should be low severity (positive)")
			assert.Contains(t, p.Description, "focused", "Should mention staying focused")
			break
		}
	}
	assert.True(t, found, "Should detect positive stack alignment")
}

func TestDetector_DetectPatterns_NoContextSwitching_NoMatch(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Build a generic tool"
	detected := detector.DetectPatterns(idea)

	// Context switching pattern should not be detected
	for _, p := range detected {
		assert.NotEqual(t, "Context switching", p.Name, "Should not detect context switching for neutral idea")
	}
}

// ============================================================================
// PERFECTIONISM DETECTION
// ============================================================================

func TestDetector_DetectPatterns_Perfectionism_Detected(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	testCases := []struct {
		name string
		idea string
	}{
		{"Comprehensive", "Build a comprehensive system"},
		{"Complete", "Create a complete solution"},
		{"Production-ready", "Production-ready platform"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detected := detector.DetectPatterns(tc.idea)

			found := false
			for _, p := range detected {
				if p.Name == "Perfectionism" {
					found = true
					assert.Equal(t, "high", p.Severity)
					assert.Contains(t, p.Description, "Scope creep", "Should mention scope creep")
					break
				}
			}
			assert.True(t, found, "Should detect perfectionism for %s", tc.name)
		})
	}
}

func TestDetector_DetectPatterns_Perfectionism_NotDetected(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Build a quick MVP in 30 days"
	detected := detector.DetectPatterns(idea)

	for _, p := range detected {
		assert.NotEqual(t, "Perfectionism", p.Name, "Should not detect perfectionism for MVP-focused idea")
	}
}

// ============================================================================
// PROCRASTINATION DETECTION
// ============================================================================

func TestDetector_DetectPatterns_Procrastination_LearnBefore(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	testCases := []struct {
		name string
		idea string
	}{
		{"Learn before building", "Learn Rust before building the app"},
		{"Learn then build", "Learn React then create the website"},
		{"Study first", "Study the framework first, then implement"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detected := detector.DetectPatterns(tc.idea)

			found := false
			for _, p := range detected {
				if p.Name == "Procrastination" {
					found = true
					assert.Equal(t, "critical", p.Severity, "Learning before building should be critical")
					assert.Contains(t, p.Description, "Consumption trap", "Should mention consumption trap")
					break
				}
			}
			assert.True(t, found, "Should detect procrastination for %s", tc.name)
		})
	}
}

func TestDetector_DetectPatterns_Procrastination_NotDetected(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Build first, learn as needed"
	detected := detector.DetectPatterns(idea)

	for _, p := range detected {
		assert.NotEqual(t, "Procrastination", p.Name, "Should not detect procrastination for build-first approach")
	}
}

// ============================================================================
// ACCOUNTABILITY AVOIDANCE DETECTION
// ============================================================================

func TestDetector_DetectPatterns_AccountabilityAvoidance_Negative(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	testCases := []struct {
		name string
		idea string
	}{
		{"Just for me", "Build a tool just for me"},
		{"Personal project", "Personal project to learn"},
		{"Solo work", "Solo project for fun"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detected := detector.DetectPatterns(tc.idea)

			found := false
			for _, p := range detected {
				if p.Name == "Accountability avoidance" {
					found = true
					assert.Equal(t, "medium", p.Severity)
					assert.Contains(t, p.Description, "Solo-only", "Should mention solo-only")
					break
				}
			}
			assert.True(t, found, "Should detect accountability avoidance for %s", tc.name)
		})
	}
}

func TestDetector_DetectPatterns_AccountabilityAvoidance_Positive(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	testCases := []struct {
		name string
		idea string
	}{
		{"Public building", "Build in public on Twitter"},
		{"Share on GitHub", "Share the code on GitHub"},
		{"Public commitment", "Public launch with customers"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detected := detector.DetectPatterns(tc.idea)

			found := false
			for _, p := range detected {
				if p.Name == "Accountability avoidance" {
					found = true
					assert.Equal(t, "low", p.Severity, "Public accountability should be low/positive severity")
					assert.Contains(t, p.Description, "External accountability", "Should mention external accountability")
					break
				}
			}
			assert.True(t, found, "Should detect positive accountability for %s", tc.name)
		})
	}
}

// ============================================================================
// MULTIPLE PATTERNS
// ============================================================================

func TestDetector_DetectPatterns_MultiplePatterns(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Learn Rust before building a comprehensive game engine. Personal project just for me."
	detected := detector.DetectPatterns(idea)

	// Should detect multiple patterns
	assert.GreaterOrEqual(t, len(detected), 3, "Should detect at least 3 patterns")

	patternNames := make(map[string]bool)
	for _, p := range detected {
		patternNames[p.Name] = true
	}

	assert.True(t, patternNames["Context switching"], "Should detect context switching")
	assert.True(t, patternNames["Procrastination"], "Should detect procrastination")
	assert.True(t, patternNames["Perfectionism"], "Should detect perfectionism")
	assert.True(t, patternNames["Accountability avoidance"], "Should detect accountability avoidance")
}

func TestDetector_DetectPatterns_NoPatterns(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Build a Python tool with LangChain, ship MVP in 30 days, share publicly"
	detected := detector.DetectPatterns(idea)

	// May detect positive patterns (stack alignment, public accountability)
	// but no negative patterns
	for _, p := range detected {
		if p.Severity == "high" || p.Severity == "critical" {
			t.Errorf("Should not detect high/critical patterns for ideal idea, found: %s (%s)", p.Name, p.Severity)
		}
	}
}

// ============================================================================
// EDGE CASES
// ============================================================================

func TestDetector_DetectPatterns_EmptyIdea_ReturnsEmpty(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	detected := detector.DetectPatterns("")

	assert.Empty(t, detected, "Empty idea should return no patterns")
}

func TestDetector_DetectPatterns_NilTelos_NoStackPatterns(t *testing.T) {
	detector := patterns.NewDetector(nil)

	idea := "Build something with Python"
	detected := detector.DetectPatterns(idea)

	// Should still detect other patterns (procrastination, perfectionism)
	// but not context-switching (requires telos for stack comparison)
	for _, p := range detected {
		assert.NotEqual(t, "Context switching", p.Name, "Should not detect context switching without telos")
	}
}

func TestDetector_DetectPatterns_ConfidenceValues(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Learn Rust before building"
	detected := detector.DetectPatterns(idea)

	// All patterns should have confidence values
	for _, p := range detected {
		assert.GreaterOrEqual(t, p.Confidence, 0.0, "Confidence should be >= 0")
		assert.LessOrEqual(t, p.Confidence, 1.0, "Confidence should be <= 1")
	}
}

func TestDetector_DetectPatterns_ValidSeverities(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	idea := "Learn Rust before building a comprehensive system just for me"
	detected := detector.DetectPatterns(idea)

	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}

	for _, p := range detected {
		assert.True(t, validSeverities[p.Severity], "Invalid severity: %s", p.Severity)
	}
}

// ============================================================================
// TELOS FAILURE PATTERN MATCHING
// ============================================================================

func TestDetector_DetectPatterns_TelosFailurePatterns_Matched(t *testing.T) {
	telosData := loadTestTelos(t)
	detector := patterns.NewDetector(telosData)

	// Telos has "Context switching" and "Perfectionism" as failure patterns
	idea := "Starting a new project before finishing the current one"
	detected := detector.DetectPatterns(idea)

	// Should detect based on telos failure pattern keywords
	found := false
	for _, p := range detected {
		if p.Name == "Context switching" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect telos failure pattern")
}
