package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CSRFToken represents a CSRF token with expiration
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// IsExpired checks if the token has expired
func (t *CSRFToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// CSRFProtection manages CSRF tokens
type CSRFProtection struct {
	tokens map[string]*CSRFToken
	mu     sync.RWMutex
	ttl    time.Duration
}

// NewCSRFProtection creates a new CSRF protection manager
func NewCSRFProtection(ttl time.Duration) *CSRFProtection {
	cp := &CSRFProtection{
		tokens: make(map[string]*CSRFToken),
		ttl:    ttl,
	}

	// Start cleanup goroutine
	go cp.cleanupExpired()

	return cp
}

// cleanupExpired periodically removes expired tokens
func (cp *CSRFProtection) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cp.mu.Lock()
		for key, token := range cp.tokens {
			if token.IsExpired() {
				delete(cp.tokens, key)
			}
		}
		cp.mu.Unlock()
	}
}

// GenerateToken generates a new CSRF token for a session
func (cp *CSRFProtection) GenerateToken(sessionID string) (string, error) {
	// Generate random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)

	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.tokens[sessionID] = &CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(cp.ttl),
	}

	return token, nil
}

// ValidateToken validates a CSRF token for a session
func (cp *CSRFProtection) ValidateToken(sessionID, token string) bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	storedToken, exists := cp.tokens[sessionID]
	if !exists || storedToken.IsExpired() {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(storedToken.Token), []byte(token)) == 1
}

// DeleteToken removes a token for a session
func (cp *CSRFProtection) DeleteToken(sessionID string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	delete(cp.tokens, sessionID)
}

// getSessionID extracts or creates a session ID from the request
// In production, this should use a proper session management system
func getSessionID(r *http.Request) string {
	// Try to get from cookie
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Fallback to IP address (not recommended for production)
	return r.RemoteAddr
}

// CSRFMiddleware protects against CSRF attacks
func CSRFMiddleware(csrf *CSRFProtection) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF check for safe methods
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Get session ID
			sessionID := getSessionID(r)

			// Get token from header
			token := r.Header.Get("X-CSRF-Token")

			// Validate token
			if !csrf.ValidateToken(sessionID, token) {
				respondError(w, http.StatusForbidden, "Invalid or missing CSRF token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetCSRFTokenHandler returns a CSRF token for the current session
func (s *Server) GetCSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionID(r)

	token, err := s.csrfProtection.GenerateToken(sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate CSRF token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"csrf_token": token,
		"expires_in": fmt.Sprintf("%d", int(s.csrfProtection.ttl.Seconds())),
	})
}
