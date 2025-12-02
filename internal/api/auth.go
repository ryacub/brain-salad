package api

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/config"
)

// AuthMiddleware checks API key authentication if enabled
func AuthMiddleware(cfg config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth for health checks and metrics (monitoring endpoints)
			// These endpoints should always be accessible for monitoring systems
			if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract API key from Authorization header
			// Format: "Authorization: Bearer sk_prod_abc123"
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Str("remote_addr", r.RemoteAddr).
					Msg("Authentication failed: missing authorization header")
				respondError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			// Parse Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Str("remote_addr", r.RemoteAddr).
					Msg("Authentication failed: invalid authorization header format")
				respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			apiKey := parts[1]

			// Validate API key
			if _, valid := cfg.APIKeys[apiKey]; !valid {
				log.Warn().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Str("remote_addr", r.RemoteAddr).
					Msg("Authentication failed: invalid API key")
				respondError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			// API key is valid, log successful authentication
			log.Debug().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Msg("Authentication successful")

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
