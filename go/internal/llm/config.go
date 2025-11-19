package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config stores LLM configuration preferences
type Config struct {
	DefaultProvider  string            `json:"default_provider"`
	ProviderSettings map[string]string `json:"provider_settings,omitempty"`
	Version          string            `json:"version"`
}

const configVersion = "1.0"

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".telos")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "llm-config.json"), nil
}

// LoadConfig loads the LLM configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Return default config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Version:          configVersion,
			ProviderSettings: make(map[string]string),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize map if nil
	if config.ProviderSettings == nil {
		config.ProviderSettings = make(map[string]string)
	}

	// Handle version migration if needed
	if config.Version == "" {
		config.Version = configVersion
	}

	return &config, nil
}

// SaveConfig saves the LLM configuration to disk
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Set version
	config.Version = configVersion

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file with secure permissions (user read/write only)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetDefaultProvider saves the default provider preference
func SetDefaultProvider(providerName string) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	config.DefaultProvider = providerName

	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetDefaultProvider retrieves the default provider preference
func GetDefaultProvider() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	return config.DefaultProvider, nil
}

// ClearDefaultProvider removes the default provider preference
func ClearDefaultProvider() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.DefaultProvider = ""

	return SaveConfig(config)
}

// SetProviderSetting stores a provider-specific setting
func SetProviderSetting(provider, key, value string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	settingKey := fmt.Sprintf("%s.%s", provider, key)
	config.ProviderSettings[settingKey] = value

	return SaveConfig(config)
}

// GetProviderSetting retrieves a provider-specific setting
func GetProviderSetting(provider, key string) (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	settingKey := fmt.Sprintf("%s.%s", provider, key)
	value, ok := config.ProviderSettings[settingKey]
	if !ok {
		return "", fmt.Errorf("setting not found: %s", settingKey)
	}

	return value, nil
}

// GetAllProviderSettings returns all settings for a provider
func GetAllProviderSettings(provider string) (map[string]string, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	prefix := provider + "."
	settings := make(map[string]string)

	for k, v := range config.ProviderSettings {
		if strings.HasPrefix(k, prefix) {
			key := strings.TrimPrefix(k, prefix)
			settings[key] = v
		}
	}

	return settings, nil
}

// ValidateConfig checks if the configuration is valid
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// Validate version
	if config.Version != configVersion {
		return fmt.Errorf("unsupported config version: %s (expected: %s)",
			config.Version, configVersion)
	}

	// Validate default provider if set
	if config.DefaultProvider != "" {
		validProviders := []string{"ollama", "claude", "openai", "custom", "rule_based"}
		valid := false
		for _, p := range validProviders {
			if config.DefaultProvider == p {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid default provider: %s", config.DefaultProvider)
		}
	}

	return nil
}

// ResetConfig deletes the configuration file and returns to defaults
func ResetConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Already reset
	}

	// Delete config file
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// GetConfigSummary returns a human-readable summary of the configuration
func GetConfigSummary() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	var summary strings.Builder

	summary.WriteString("LLM Configuration\n")
	summary.WriteString(strings.Repeat("-", 40) + "\n")

	if config.DefaultProvider != "" {
		summary.WriteString(fmt.Sprintf("Default Provider: %s\n", config.DefaultProvider))
	} else {
		summary.WriteString("Default Provider: (not set)\n")
	}

	if len(config.ProviderSettings) > 0 {
		summary.WriteString("\nProvider Settings:\n")
		for k, v := range config.ProviderSettings {
			summary.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	configPath, _ := GetConfigPath()
	summary.WriteString(fmt.Sprintf("\nConfig File: %s\n", configPath))

	return summary.String(), nil
}
