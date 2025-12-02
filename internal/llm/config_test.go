package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	if path == "" {
		t.Error("Config path should not be empty")
	}

	if !filepath.IsAbs(path) {
		t.Error("Config path should be absolute")
	}

	expectedSuffix := filepath.Join(".telos", "llm-config.json")
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("Config path should end with %s, got %s", expectedSuffix, path)
	}
}

func TestLoadConfig_NewFile(t *testing.T) {
	// Setup: ensure no config exists
	configPath, _ := GetConfigPath()
	_ = os.Remove(configPath)
	defer func() { _ = os.Remove(configPath) }()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
		return
	}

	if config.DefaultProvider != "" {
		t.Errorf("New config should have empty DefaultProvider, got %s", config.DefaultProvider)
	}

	if config.Version != configVersion {
		t.Errorf("Expected version %s, got %s", configVersion, config.Version)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Create config
	config := &Config{
		DefaultProvider:  "openai",
		ProviderSettings: map[string]string{"test": "value"},
		Version:          configVersion,
	}

	// Save
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.DefaultProvider != "openai" {
		t.Errorf("Expected DefaultProvider 'openai', got '%s'", loaded.DefaultProvider)
	}

	if loaded.ProviderSettings["test"] != "value" {
		t.Error("Provider settings not preserved")
	}
}

func TestSetAndGetDefaultProvider(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Set
	err := SetDefaultProvider("claude")
	if err != nil {
		t.Fatalf("SetDefaultProvider failed: %v", err)
	}

	// Get
	provider, err := GetDefaultProvider()
	if err != nil {
		t.Fatalf("GetDefaultProvider failed: %v", err)
	}

	if provider != "claude" {
		t.Errorf("Expected 'claude', got '%s'", provider)
	}
}

func TestClearDefaultProvider(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Set a provider
	err := SetDefaultProvider("claude")
	if err != nil {
		t.Fatalf("SetDefaultProvider failed: %v", err)
	}

	// Clear it
	err = ClearDefaultProvider()
	if err != nil {
		t.Fatalf("ClearDefaultProvider failed: %v", err)
	}

	// Verify it's cleared
	provider, err := GetDefaultProvider()
	if err != nil {
		t.Fatalf("GetDefaultProvider failed: %v", err)
	}

	if provider != "" {
		t.Errorf("Expected empty provider after clear, got '%s'", provider)
	}
}

func TestProviderSettings(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Set settings
	err := SetProviderSetting("openai", "model", "gpt-4")
	if err != nil {
		t.Fatalf("SetProviderSetting failed: %v", err)
	}

	err = SetProviderSetting("openai", "temperature", "0.7")
	if err != nil {
		t.Fatalf("SetProviderSetting failed: %v", err)
	}

	// Get specific setting
	model, err := GetProviderSetting("openai", "model")
	if err != nil {
		t.Fatalf("GetProviderSetting failed: %v", err)
	}

	if model != "gpt-4" {
		t.Errorf("Expected 'gpt-4', got '%s'", model)
	}

	// Get all settings
	settings, err := GetAllProviderSettings("openai")
	if err != nil {
		t.Fatalf("GetAllProviderSettings failed: %v", err)
	}

	if len(settings) != 2 {
		t.Errorf("Expected 2 settings, got %d", len(settings))
	}

	if settings["model"] != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", settings["model"])
	}

	if settings["temperature"] != "0.7" {
		t.Errorf("Expected temperature '0.7', got '%s'", settings["temperature"])
	}
}

func TestGetProviderSetting_NotFound(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Try to get non-existent setting
	_, err := GetProviderSetting("openai", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent setting, got nil")
	}

	expectedErrMsg := "setting not found"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestGetAllProviderSettings_Empty(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Get settings for provider with no settings
	settings, err := GetAllProviderSettings("openai")
	if err != nil {
		t.Fatalf("GetAllProviderSettings failed: %v", err)
	}

	if len(settings) != 0 {
		t.Errorf("Expected 0 settings, got %d", len(settings))
	}
}

func TestGetAllProviderSettings_MultipleProviders(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Set settings for multiple providers
	_ = SetProviderSetting("openai", "model", "gpt-4")
	_ = SetProviderSetting("claude", "model", "claude-3")
	_ = SetProviderSetting("openai", "temperature", "0.7")

	// Get settings for openai only
	settings, err := GetAllProviderSettings("openai")
	if err != nil {
		t.Fatalf("GetAllProviderSettings failed: %v", err)
	}

	if len(settings) != 2 {
		t.Errorf("Expected 2 settings for openai, got %d", len(settings))
	}

	// Get settings for claude only
	settings, err = GetAllProviderSettings("claude")
	if err != nil {
		t.Fatalf("GetAllProviderSettings failed: %v", err)
	}

	if len(settings) != 1 {
		t.Errorf("Expected 1 setting for claude, got %d", len(settings))
	}

	if settings["model"] != "claude-3" {
		t.Errorf("Expected model 'claude-3', got '%s'", settings["model"])
	}
}

func TestResetConfig(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()

	// Create config
	err := SetDefaultProvider("openai")
	if err != nil {
		t.Fatalf("SetDefaultProvider failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file should exist")
	}

	// Reset
	err = ResetConfig()
	if err != nil {
		t.Fatalf("ResetConfig failed: %v", err)
	}

	// Verify file deleted
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Config file should be deleted")
	}
}

func TestResetConfig_AlreadyReset(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	_ = os.Remove(configPath)

	// Reset when file doesn't exist
	err := ResetConfig()
	if err != nil {
		t.Fatalf("ResetConfig failed when file doesn't exist: %v", err)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with openai",
			config: &Config{
				DefaultProvider: "openai",
				Version:         configVersion,
			},
			wantErr: false,
		},
		{
			name: "valid config with claude",
			config: &Config{
				DefaultProvider: "claude",
				Version:         configVersion,
			},
			wantErr: false,
		},
		{
			name: "valid config with ollama",
			config: &Config{
				DefaultProvider: "ollama",
				Version:         configVersion,
			},
			wantErr: false,
		},
		{
			name: "valid config with custom",
			config: &Config{
				DefaultProvider: "custom",
				Version:         configVersion,
			},
			wantErr: false,
		},
		{
			name: "valid config with empty provider",
			config: &Config{
				DefaultProvider: "",
				Version:         configVersion,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "invalid provider",
			config: &Config{
				DefaultProvider: "invalid",
				Version:         configVersion,
			},
			wantErr: true,
		},
		{
			name: "invalid version",
			config: &Config{
				DefaultProvider: "openai",
				Version:         "0.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetConfigSummary(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Set some config
	_ = SetDefaultProvider("claude")
	_ = SetProviderSetting("openai", "model", "gpt-4")

	// Get summary
	summary, err := GetConfigSummary()
	if err != nil {
		t.Fatalf("GetConfigSummary failed: %v", err)
	}

	// Verify summary contains key information
	if !strings.Contains(summary, "claude") {
		t.Error("Summary should contain default provider 'claude'")
	}

	if !strings.Contains(summary, "openai.model") {
		t.Error("Summary should contain provider setting 'openai.model'")
	}

	if !strings.Contains(summary, "gpt-4") {
		t.Error("Summary should contain setting value 'gpt-4'")
	}

	if !strings.Contains(summary, configPath) {
		t.Error("Summary should contain config file path")
	}
}

func TestGetConfigSummary_Empty(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Get summary with no config
	summary, err := GetConfigSummary()
	if err != nil {
		t.Fatalf("GetConfigSummary failed: %v", err)
	}

	// Verify summary indicates no default provider
	if !strings.Contains(summary, "(not set)") {
		t.Error("Summary should indicate no default provider is set")
	}
}

func TestConfigPersistence(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Set default provider
	err := SetDefaultProvider("claude")
	if err != nil {
		t.Fatalf("SetDefaultProvider failed: %v", err)
	}

	// Load config again (simulating new session)
	provider, err := GetDefaultProvider()
	if err != nil {
		t.Fatalf("GetDefaultProvider failed: %v", err)
	}

	if provider != "claude" {
		t.Errorf("Provider should persist across loads, expected 'claude', got '%s'", provider)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Save config
	config := &Config{
		DefaultProvider: "openai",
		Version:         configVersion,
	}
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Verify permissions are 0600 (owner read/write only)
	mode := info.Mode().Perm()
	expected := os.FileMode(0600)
	if mode != expected {
		t.Errorf("Expected file permissions %o, got %o", expected, mode)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("invalid json {{{"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Try to load
	_, err = LoadConfig()
	if err == nil {
		t.Error("Expected error when loading invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse config file") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestLoadConfig_InitializesNilMap(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Write config with null provider_settings
	jsonData := `{
		"default_provider": "openai",
		"provider_settings": null,
		"version": "1.0"
	}`
	err := os.WriteFile(configPath, []byte(jsonData), 0600)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify map is initialized
	if config.ProviderSettings == nil {
		t.Error("ProviderSettings should be initialized, not nil")
	}
}

func TestLoadConfig_MigratesVersion(t *testing.T) {
	// Setup
	configPath, _ := GetConfigPath()
	defer func() { _ = os.Remove(configPath) }()

	// Write config without version
	jsonData := `{
		"default_provider": "openai"
	}`
	err := os.WriteFile(configPath, []byte(jsonData), 0600)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify version is set
	if config.Version != configVersion {
		t.Errorf("Expected version to be migrated to %s, got %s", configVersion, config.Version)
	}
}
