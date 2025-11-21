// Package config provides authentication configuration management.
package config

import (
	"os"
	"strings"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Enabled controls whether authentication is required
	// Default: false (local CLI tool, no auth needed)
	// Production: true (public API deployment)
	Enabled bool

	// Mode specifies the authentication mechanism
	// Supported: "api-key", "jwt" (future: "oauth2")
	Mode string

	// APIKeys holds valid API keys (if Mode == "api-key")
	// Format: map[key]description
	// Example: {"sk_prod_abc123": "Production client", "sk_dev_xyz789": "Development"}
	APIKeys map[string]string

	// JWTSecret is the secret key for JWT signing (if Mode == "jwt")
	JWTSecret string
}

// DefaultAuthConfig returns auth config for local development (disabled)
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Enabled: false,
		Mode:    "api-key",
		APIKeys: make(map[string]string),
	}
}

// LoadAuthConfig loads authentication configuration from environment variables
func LoadAuthConfig() AuthConfig {
	cfg := DefaultAuthConfig()

	// Check if authentication is enabled
	if os.Getenv("AUTH_ENABLED") == "true" {
		cfg.Enabled = true
		cfg.Mode = getEnvOrDefault("AUTH_MODE", "api-key")

		// Load API keys from comma-separated env var
		// Format: AUTH_API_KEYS="key1:desc1,key2:desc2"
		if keys := os.Getenv("AUTH_API_KEYS"); keys != "" {
			cfg.APIKeys = parseAPIKeys(keys)
		}

		// Load JWT secret if in JWT mode
		if cfg.Mode == "jwt" {
			cfg.JWTSecret = os.Getenv("JWT_SECRET")
		}
	}

	return cfg
}

// parseAPIKeys parses API keys from the format "key1:desc1,key2:desc2"
func parseAPIKeys(input string) map[string]string {
	keys := make(map[string]string)

	// Split by comma to get individual key:description pairs
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		// Split by colon to separate key and description
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			desc := strings.TrimSpace(parts[1])
			if key != "" {
				keys[key] = desc
			}
		} else if len(parts) == 1 {
			// If no description provided, use empty string
			key := strings.TrimSpace(parts[0])
			if key != "" {
				keys[key] = ""
			}
		}
	}

	return keys
}

// getEnvOrDefault retrieves an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
