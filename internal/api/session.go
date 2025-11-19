package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// SessionConfig holds configuration for session management
type SessionConfig struct {
	// IdleTimeout is the maximum time a session can be inactive
	IdleTimeout time.Duration
	// AbsoluteTimeout is the maximum lifetime of a session
	AbsoluteTimeout time.Duration
	// CookieName is the name of the session cookie
	CookieName string
	// CookiePath is the path for the session cookie
	CookiePath string
	// CookieDomain is the domain for the session cookie (optional)
	CookieDomain string
	// SecureCookie indicates if the cookie should only be sent over HTTPS
	SecureCookie bool
	// SameSite is the SameSite attribute for the cookie
	SameSite http.SameSite
}

// DefaultSessionConfig returns a secure default configuration
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		IdleTimeout:     1 * time.Hour,  // 1 hour of inactivity
		AbsoluteTimeout: 7 * 24 * time.Hour, // 7 days maximum
		CookieName:      "session_id",
		CookiePath:      "/",
		CookieDomain:    "", // Let browser determine
		SecureCookie:    true, // Should be true in production with HTTPS
		SameSite:        http.SameSiteStrictMode,
	}
}

// Session represents a user session
type Session struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
	LastSeen  time.Time
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionManager manages user sessions backed by SQLite
type SessionManager struct {
	db     *sql.DB
	config SessionConfig
	stopCh chan struct{}
}

// NewSessionManager creates a new session manager
func NewSessionManager(db *sql.DB, config SessionConfig) *SessionManager {
	sm := &SessionManager{
		db:     db,
		config: config,
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go sm.cleanupExpired()

	return sm
}

// generateSessionID generates a cryptographically secure session ID
func (sm *SessionManager) generateSessionID() (string, error) {
	b := make([]byte, 32) // 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateSession creates a new session in the database
func (sm *SessionManager) CreateSession() (*Session, error) {
	sessionID, err := sm.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(sm.config.AbsoluteTimeout)

	session := &Session{
		ID:        sessionID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		LastSeen:  now,
	}

	query := `
		INSERT INTO sessions (id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?)
	`

	_, err = sm.db.Exec(
		query,
		session.ID,
		session.CreatedAt.Format(time.RFC3339),
		session.ExpiresAt.Format(time.RFC3339),
		session.LastSeen.Format(time.RFC3339),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID and updates last_seen
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	query := `
		SELECT id, created_at, expires_at, last_seen
		FROM sessions
		WHERE id = ?
	`

	var session Session
	var createdAt, expiresAt, lastSeen string

	err := sm.db.QueryRow(query, sessionID).Scan(
		&session.ID,
		&createdAt,
		&expiresAt,
		&lastSeen,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Parse timestamps
	session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
	session.LastSeen, _ = time.Parse(time.RFC3339, lastSeen)

	// Check if session has expired
	if session.IsExpired() {
		// Delete expired session
		_ = sm.DeleteSession(sessionID)
		return nil, errors.New("session expired")
	}

	// Check idle timeout
	idleDeadline := session.LastSeen.Add(sm.config.IdleTimeout)
	if time.Now().After(idleDeadline) {
		// Session has been idle too long
		_ = sm.DeleteSession(sessionID)
		return nil, errors.New("session expired due to inactivity")
	}

	// Update last_seen timestamp
	now := time.Now()
	updateQuery := `
		UPDATE sessions
		SET last_seen = ?
		WHERE id = ?
	`
	_, err = sm.db.Exec(updateQuery, now.Format(time.RFC3339), sessionID)
	if err != nil {
		// Log error but don't fail the request
		// The session is still valid even if we couldn't update last_seen
	} else {
		session.LastSeen = now
	}

	return &session, nil
}

// DeleteSession removes a session from the database
func (sm *SessionManager) DeleteSession(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	query := "DELETE FROM sessions WHERE id = ?"
	_, err := sm.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// RefreshSession extends the session expiration time
func (sm *SessionManager) RefreshSession(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	// Get current session to check if it exists and is valid
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Calculate new expiration time
	newExpiresAt := time.Now().Add(sm.config.AbsoluteTimeout)

	// Update expiration
	query := `
		UPDATE sessions
		SET expires_at = ?
		WHERE id = ?
	`
	_, err = sm.db.Exec(query, newExpiresAt.Format(time.RFC3339), session.ID)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	return nil
}

// cleanupExpired periodically removes expired sessions
func (sm *SessionManager) cleanupExpired() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Delete sessions that have exceeded absolute timeout
			deleteQuery := `
				DELETE FROM sessions
				WHERE expires_at < ?
			`
			now := time.Now().Format(time.RFC3339)
			if _, err := sm.db.Exec(deleteQuery, now); err != nil {
				// Log error but continue
				continue
			}

			// Delete sessions that have exceeded idle timeout
			idleDeadline := time.Now().Add(-sm.config.IdleTimeout).Format(time.RFC3339)
			deleteIdleQuery := `
				DELETE FROM sessions
				WHERE last_seen < ?
			`
			_, _ = sm.db.Exec(deleteIdleQuery, idleDeadline)
		case <-sm.stopCh:
			return
		}
	}
}

// Stop gracefully stops the session manager cleanup goroutine
func (sm *SessionManager) Stop() {
	close(sm.stopCh)
}

// GetSessionFromRequest extracts the session ID from a request cookie
func (sm *SessionManager) GetSessionFromRequest(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(sm.config.CookieName)
	if err != nil {
		return nil, errors.New("no session cookie found")
	}

	if cookie.Value == "" {
		return nil, errors.New("session cookie is empty")
	}

	return sm.GetSession(cookie.Value)
}

// GetOrCreateSession gets an existing session or creates a new one
func (sm *SessionManager) GetOrCreateSession(r *http.Request) (*Session, bool, error) {
	// Try to get existing session
	session, err := sm.GetSessionFromRequest(r)
	if err == nil && session != nil {
		return session, false, nil
	}

	// Create new session if none exists or is invalid
	newSession, err := sm.CreateSession()
	if err != nil {
		return nil, false, fmt.Errorf("failed to create session: %w", err)
	}

	return newSession, true, nil
}

// SetSessionCookie sets the session cookie on the response
func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, session *Session) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    session.ID,
		Path:     sm.config.CookiePath,
		Domain:   sm.config.CookieDomain,
		Expires:  session.ExpiresAt,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		HttpOnly: true, // Prevents JavaScript access (XSS protection)
		Secure:   sm.config.SecureCookie, // Only send over HTTPS
		SameSite: sm.config.SameSite, // CSRF protection at cookie level
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie removes the session cookie from the response
func (sm *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    "",
		Path:     sm.config.CookiePath,
		Domain:   sm.config.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   sm.config.SecureCookie,
		SameSite: sm.config.SameSite,
	}

	http.SetCookie(w, cookie)
}

// SessionMiddleware ensures every request has a valid session
func SessionMiddleware(sm *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get or create session
			session, isNew, err := sm.GetOrCreateSession(r)
			if err != nil {
				// If we can't create a session, something is seriously wrong
				respondError(w, http.StatusInternalServerError, "Failed to manage session")
				return
			}

			// Set cookie if it's a new session or needs to be refreshed
			if isNew {
				sm.SetSessionCookie(w, session)
			}

			// Store session ID in request context for use by other handlers
			ctx := r.Context()
			ctx = contextWithSessionID(ctx, session.ID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// Context keys for storing session information
type contextKey string

const sessionIDKey contextKey = "session_id"

// contextWithSessionID adds a session ID to the context
func contextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// getSessionIDFromContext retrieves the session ID from the context
func getSessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

// GetSessionID extracts the session ID from the request
// This replaces the old getSessionID function that used IP addresses
func GetSessionID(r *http.Request) (string, error) {
	// First try to get from context (set by SessionMiddleware)
	if sessionID, ok := getSessionIDFromContext(r.Context()); ok && sessionID != "" {
		return sessionID, nil
	}

	// Fallback: try to get directly from cookie
	// This handles cases where SessionMiddleware wasn't used
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", errors.New("no session found")
}
