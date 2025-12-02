package cli

import (
	"testing"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
)

func TestAnalyzeLLMCommand(t *testing.T) {
	cmd := newAnalyzeLLMCommand()

	// Test command creation
	if cmd == nil {
		t.Fatal("Command creation failed")
		return
	}

	// Test command name
	if cmd.Use != "analyze-llm <idea>" {
		t.Errorf("Expected command name 'analyze-llm <idea>', got '%s'", cmd.Use)
	}

	// Test flags exist
	flags := cmd.Flags()
	if flags.Lookup("provider") == nil {
		t.Error("Missing --provider flag")
	}
	if flags.Lookup("no-fallback") == nil {
		t.Error("Missing --no-fallback flag")
	}
	if flags.Lookup("verbose") == nil {
		t.Error("Missing --verbose flag")
	}
}

func TestLLMListCommand(t *testing.T) {
	cmd := newLLMListCommand()

	// Test command creation
	if cmd == nil {
		t.Fatal("Command creation failed")
		return
	}

	// Test command name
	if cmd.Use != "llm-list" {
		t.Errorf("Expected command name 'llm-list', got '%s'", cmd.Use)
	}

	// Test flags exist
	flags := cmd.Flags()
	if flags.Lookup("health") == nil {
		t.Error("Missing --health flag")
	}
}

func TestLLMConfigCommand(t *testing.T) {
	cmd := newLLMConfigCommand()

	// Test command creation
	if cmd == nil {
		t.Fatal("Command creation failed")
		return
	}

	// Test command name
	if cmd.Use != "llm-config" {
		t.Errorf("Expected command name 'llm-config', got '%s'", cmd.Use)
	}

	// Test subcommands exist
	if !cmd.HasSubCommands() {
		t.Error("Expected subcommands but none found")
	}

	// Check for specific subcommands
	foundSetDefault := false
	foundShow := false
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "set-default <provider>" {
			foundSetDefault = true
		}
		if subcmd.Use == "show" {
			foundShow = true
		}
	}

	if !foundSetDefault {
		t.Error("Missing 'set-default' subcommand")
	}
	if !foundShow {
		t.Error("Missing 'show' subcommand")
	}
}

func TestLLMHealthCommand(t *testing.T) {
	cmd := newLLMHealthCommand()

	// Test command creation
	if cmd == nil {
		t.Fatal("Command creation failed")
		return
	}

	// Test command name
	if cmd.Use != "llm-health" {
		t.Errorf("Expected command name 'llm-health', got '%s'", cmd.Use)
	}

	// Test flags exist
	flags := cmd.Flags()
	if flags.Lookup("watch") == nil {
		t.Error("Missing --watch flag")
	}
	if flags.Lookup("interval") == nil {
		t.Error("Missing --interval flag")
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short key",
			input:    "abc",
			expected: "***",
		},
		{
			name:     "normal key",
			input:    "sk-1234567890abcdef",
			expected: "sk-1...cdef",
		},
		{
			name:     "long key",
			input:    "sk-proj-1234567890abcdefghijklmnop",
			expected: "sk-p...mnop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskAPIKey(tt.input)
			if result != tt.expected {
				t.Errorf("maskAPIKey(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatCategoryTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"mission_alignment", "📊 Mission Alignment"},
		{"anti_challenge", "🎯 Anti-Challenge"},
		{"strategic_fit", "🚀 Strategic Fit"},
		{"unknown_category", "Unknown Category"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatCategoryTitle(tt.input)
			if result != tt.expected {
				t.Errorf("formatCategoryTitle(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		indent   string
		expected string
	}{
		{
			name:     "short text",
			text:     "Hello world",
			width:    20,
			indent:   "  ",
			expected: "  Hello world",
		},
		{
			name:     "long text wraps",
			text:     "This is a very long line that should wrap at the specified width",
			width:    30,
			indent:   "  ",
			expected: "  This is a very long line\n  that should wrap at the\n  specified width",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width, tt.indent)
			if result != tt.expected {
				t.Errorf("wrapText() failed\nGot:\n%s\nExpected:\n%s", result, tt.expected)
			}
		})
	}
}

func TestLLMManagerIntegration(t *testing.T) {
	// Test that manager can be created
	manager := llm.NewManager(nil)
	if manager == nil {
		t.Fatal("Failed to create LLM manager")
	}

	// Test that rule_based provider is always available
	providers := manager.GetAllProviders()
	if _, exists := providers["rule_based"]; !exists {
		t.Error("rule_based provider should always be registered")
	}

	// Test health check
	health := manager.HealthCheck()
	if len(health) == 0 {
		t.Error("Health check should return at least one provider")
	}

	// rule_based should always be healthy
	if !health["rule_based"] {
		t.Error("rule_based provider should always be healthy")
	}
}
