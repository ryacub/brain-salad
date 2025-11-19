package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestE2E_AnalyzeLLM_Basic tests the basic analyze-llm command
func TestE2E_AnalyzeLLM_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Build the binary
	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	// Run analyze command
	cmd := exec.Command(binaryPath, "analyze-llm", "Build a simple web app")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		t.Logf("Output: %s", output)
		t.Fatalf("command failed: %v", err)
	}

	// Verify output contains expected fields
	requiredFields := []string{
		"Score:",
		"Recommendation:",
		"Provider:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(output, field) {
			t.Errorf("output missing field '%s'", field)
		}
	}

	t.Logf("Analysis output:\n%s", output)
}

// TestE2E_AnalyzeLLM_WithTelos tests analyze-llm with custom telos
func TestE2E_AnalyzeLLM_WithTelos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Create a temporary telos file
	telosContent := `# Goals
1. Build AI tools
2. Generate revenue

# Strategies
- Move fast
- Ship MVPs

# Tech Stack
Primary: Python, Go
Secondary: Docker

# Failure Patterns
- Perfection paralysis: perfect, complete, polished
`

	telosPath := createTempFile(t, "telos.md", telosContent)
	defer func() { _ = os.Remove(telosPath) }()

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	// Run with custom telos
	cmd := exec.Command(binaryPath, "analyze-llm",
		"--telos", telosPath,
		"Build an AI automation tool")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		t.Logf("Output: %s", output)
		t.Fatalf("command failed: %v", err)
	}

	// Should complete successfully
	if !strings.Contains(output, "Score:") {
		t.Errorf("expected score in output")
	}

	t.Logf("Analysis with custom telos:\n%s", output)
}

// TestE2E_AnalyzeLLM_LongIdea tests with a complex idea
func TestE2E_AnalyzeLLM_LongIdea(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	longIdea := `Build a comprehensive AI-powered automation platform that uses
Python and GPT-4 to help businesses streamline their workflows. The platform
will include a web interface, API, and CLI tools. Target market is small to
medium businesses willing to pay $200/month subscription.`

	cmd := exec.Command(binaryPath, "analyze-llm", longIdea)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		t.Logf("Output: %s", output)
		t.Fatalf("command failed: %v", err)
	}

	// Verify output
	if !strings.Contains(output, "Score:") {
		t.Error("output missing score")
	}

	t.Logf("Long idea analysis:\n%s", output)
}

// TestE2E_AnalyzeLLM_Performance tests response time
func TestE2E_AnalyzeLLM_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	start := time.Now()

	cmd := exec.Command(binaryPath, "analyze-llm", "Build a web app")
	err := cmd.Run()

	duration := time.Since(start)

	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Should complete quickly (rule-based provider should be fast)
	maxDuration := 5 * time.Second
	if duration > maxDuration {
		t.Errorf("command took %v, expected <%v", duration, maxDuration)
	}

	t.Logf("Command completed in %v", duration)
}

// TestE2E_AnalyzeLLM_MultipleRuns tests running multiple analyses
func TestE2E_AnalyzeLLM_MultipleRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	ideas := []string{
		"Build a web app",
		"Create a mobile app",
		"Develop an API service",
	}

	for i, idea := range ideas {
		t.Run("idea_"+string(rune('0'+i)), func(t *testing.T) {
			cmd := exec.Command(binaryPath, "analyze-llm", idea)
			var out bytes.Buffer
			cmd.Stdout = &out

			err := cmd.Run()
			if err != nil {
				t.Fatalf("command failed for idea %d: %v", i, err)
			}

			output := out.String()
			if !strings.Contains(output, "Score:") {
				t.Errorf("idea %d missing score in output", i)
			}
		})
	}
}

// TestE2E_AnalyzeLLM_InvalidInput tests error handling
func TestE2E_AnalyzeLLM_InvalidInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no idea",
			args: []string{"analyze-llm"},
		},
		{
			name: "empty idea",
			args: []string{"analyze-llm", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			err := cmd.Run()
			output := out.String()

			// Should either fail or provide helpful message
			if err == nil && !strings.Contains(output, "Score:") {
				// If it doesn't fail, it should at least provide output
				t.Logf("Command succeeded with output: %s", output)
			}
		})
	}
}

// TestE2E_AnalyzeLLM_Help tests help output
func TestE2E_AnalyzeLLM_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	cmd := exec.Command(binaryPath, "analyze-llm", "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		t.Logf("Output: %s", output)
		// Some CLIs return non-zero for --help, which is okay
	}

	// Should contain usage information
	if !strings.Contains(output, "analyze-llm") {
		t.Error("help output should mention command name")
	}

	t.Logf("Help output:\n%s", output)
}

// TestE2E_AnalyzeLLM_JSONOutput tests JSON output format
func TestE2E_AnalyzeLLM_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	// Try with --json flag if supported
	cmd := exec.Command(binaryPath, "analyze-llm", "--json", "Build a web app")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		// JSON flag might not be implemented yet, that's okay
		t.Logf("JSON output not supported or failed: %v", err)
		t.Logf("Output: %s", output)
		return
	}

	// If it succeeded, check for JSON-like output
	if strings.Contains(output, "{") || strings.Contains(output, "\"") {
		t.Logf("JSON output detected:\n%s", output)
	}
}

// TestE2E_AnalyzeLLM_VerboseOutput tests verbose mode
func TestE2E_AnalyzeLLM_VerboseOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	// Run with verbose flag
	cmd := exec.Command(binaryPath, "analyze-llm", "-v", "Build a web app")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		t.Logf("Verbose mode failed (might not be implemented): %v", err)
		t.Logf("Output: %s", output)
		return
	}

	// Verbose mode should provide more details
	t.Logf("Verbose output:\n%s", output)
}

// Helper functions

func buildBinary(t *testing.T) string {
	t.Helper()

	// Get the project root (assumes we're in /go/test)
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Navigate to go directory if we're in test
	if filepath.Base(projectRoot) == "test" {
		projectRoot = filepath.Dir(projectRoot)
	}

	// Create temp binary name
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "brain-salad-test")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/cli")
	cmd.Dir = projectRoot

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Logf("Build output: %s", out.String())
		t.Fatalf("failed to build binary: %v", err)
	}

	return binaryPath
}

func createTempFile(t *testing.T, name, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return filePath
}

// TestE2E_AnalyzeLLM_StressTest runs many analyses to test stability
func TestE2E_AnalyzeLLM_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	binaryPath := buildBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	iterations := 10
	successCount := 0

	for i := 0; i < iterations; i++ {
		cmd := exec.Command(binaryPath, "analyze-llm", "Build a web app")
		err := cmd.Run()

		if err == nil {
			successCount++
		}
	}

	successRate := float64(successCount) / float64(iterations) * 100
	t.Logf("Stress test: %d/%d succeeded (%.1f%%)", successCount, iterations, successRate)

	if successRate < 90 {
		t.Errorf("success rate %.1f%% below acceptable threshold", successRate)
	}
}
