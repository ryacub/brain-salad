package cli

import (
	"testing"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
)

func TestNewLLMCommand(t *testing.T) {
	cmd := NewLLMCommand()

	if cmd == nil {
		t.Fatal("Command should not be nil")
		return
	}

	if cmd.Use != "llm" {
		t.Errorf("Expected Use='llm', got '%s'", cmd.Use)
	}

	// Check that all subcommands are registered
	expectedSubcommands := []string{"list", "test", "set-default", "config", "health"}
	for _, expected := range expectedSubcommands {
		found := false
		for _, subcmd := range cmd.Commands() {
			if subcmd.Use == expected || subcmd.Use == expected+" <provider-name>" || subcmd.Use == expected+" [provider-name]" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

func TestLLMListSubcommand(t *testing.T) {
	cmd := newLLMListSubcommand()

	if cmd == nil {
		t.Fatal("Command should not be nil")
		return
	}

	if cmd.Use != "list" {
		t.Errorf("Expected Use='list', got '%s'", cmd.Use)
	}

	// Check that --all flag is registered
	flag := cmd.Flags().Lookup("all")
	if flag == nil {
		t.Error("Expected --all flag to be registered")
	}
}

func TestLLMTestSubcommand(t *testing.T) {
	cmd := newLLMTestSubcommand()

	if cmd == nil {
		t.Fatal("Command should not be nil")
		return
	}

	if cmd.Use != "test <provider-name>" {
		t.Errorf("Expected Use='test <provider-name>', got '%s'", cmd.Use)
	}
}

func TestLLMSetDefaultSubcommand(t *testing.T) {
	cmd := newLLMSetDefaultSubcommand()

	if cmd == nil {
		t.Fatal("Command should not be nil")
		return
	}

	if cmd.Use != "set-default <provider-name>" {
		t.Errorf("Expected Use='set-default <provider-name>', got '%s'", cmd.Use)
	}
}

func TestLLMConfigSubcommand(t *testing.T) {
	cmd := newLLMConfigSubcommand()

	if cmd == nil {
		t.Fatal("Command should not be nil")
		return
	}

	if cmd.Use != "config [provider-name]" {
		t.Errorf("Expected Use='config [provider-name]', got '%s'", cmd.Use)
	}
}

func TestMaskAPIKeyLLM(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "(not set)"},
		{"short", "***"},
		{"sk-1234567890abcdef", "sk-1...cdef"},
		{"abcdefghijklmnop", "abcd...mnop"},
	}

	for _, tt := range tests {
		result := maskAPIKeyLLM(tt.input)
		if result != tt.expected {
			t.Errorf("maskAPIKeyLLM(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestGetProviderConfigHelp(t *testing.T) {
	tests := []string{"openai", "claude", "ollama", "custom", "rule_based"}

	for _, provider := range tests {
		help := getProviderConfigHelp(provider)
		if help == "" {
			t.Errorf("Help should not be empty for provider: %s", provider)
		}
		if help == "  No configuration help available for this provider." {
			t.Errorf("Expected specific help for provider: %s, got default message", provider)
		}
	}
}

func TestGetProviderConfigHelpUnknown(t *testing.T) {
	help := getProviderConfigHelp("unknown-provider")
	expected := "  No configuration help available for this provider."
	if help != expected {
		t.Errorf("Expected default help message, got: %s", help)
	}
}

func TestGetProviderByName(t *testing.T) {
	config := llm.DefaultManagerConfig()
	manager := llm.NewManager(config)

	// Test finding a provider that should exist
	provider := getProviderByName(manager, "rule_based")
	if provider == nil {
		t.Error("Expected to find rule_based provider")
	}

	// Test finding a provider that doesn't exist
	provider = getProviderByName(manager, "nonexistent")
	if provider != nil {
		t.Error("Expected nil for nonexistent provider")
	}
}

func TestGetProviderList(t *testing.T) {
	config := llm.DefaultManagerConfig()
	manager := llm.NewManager(config)

	list := getProviderList(manager)
	if list == "" {
		t.Error("Provider list should not be empty")
	}

	// Should contain at least rule_based provider
	if !stringContains(list, "rule_based") {
		t.Error("Provider list should contain rule_based")
	}
}

func TestGetFirstExplanation(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: "",
		},
		{
			name: "single explanation",
			input: map[string]string{
				"test": "explanation",
			},
			expected: "explanation",
		},
		{
			name: "multiple explanations",
			input: map[string]string{
				"test1": "explanation1",
				"test2": "explanation2",
			},
			expected: "", // Will be one of them, but we can't predict which
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFirstExplanation(tt.input)
			if tt.name == "empty map" && result != tt.expected {
				t.Errorf("Expected empty string, got: %s", result)
			}
			if tt.name == "single explanation" && result != tt.expected {
				t.Errorf("Expected %s, got: %s", tt.expected, result)
			}
			if tt.name == "multiple explanations" && result == "" && len(tt.input) > 0 {
				t.Error("Expected non-empty result for non-empty map")
			}
		})
	}
}

// Helper function for tests
func stringContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
