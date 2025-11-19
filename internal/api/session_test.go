package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) (*database.Repository, func()) {
	// Create temp database file
	tmpfile, err := os.CreateTemp("", "session_test_*.db")
	require.NoError(t, err)
	_ = tmpfile.Close()

	dbPath := tmpfile.Name()

	// Create repository (which runs migrations)
	repo, err := database.NewRepository(dbPath)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		_ = repo.Close()
		_ = os.Remove(dbPath)
		// Clean up WAL files
		_ = os.Remove(dbPath + "-wal")
		_ = os.Remove(dbPath + "-shm")
	}

	return repo, cleanup
}

func TestSessionManager_CreateSession(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false // For testing
	sm := NewSessionManager(repo.DB(), config)

	session, err := sm.CreateSession()
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.False(t, session.CreatedAt.IsZero())
	assert.False(t, session.ExpiresAt.IsZero())
	assert.False(t, session.LastSeen.IsZero())
	assert.True(t, session.ExpiresAt.After(session.CreatedAt))
}

func TestSessionManager_GetSession(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create a session
	created, err := sm.CreateSession()
	require.NoError(t, err)

	// Retrieve it
	retrieved, err := sm.GetSession(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.WithinDuration(t, created.CreatedAt, retrieved.CreatedAt, time.Second)
	assert.WithinDuration(t, created.ExpiresAt, retrieved.ExpiresAt, time.Second)
}

func TestSessionManager_GetSession_NotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	_, err := sm.GetSession("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSessionManager_GetSession_Expired(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	config.AbsoluteTimeout = 1 * time.Millisecond // Very short timeout
	sm := NewSessionManager(repo.DB(), config)

	// Create session
	session, err := sm.CreateSession()
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Try to retrieve expired session
	_, err = sm.GetSession(session.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestSessionManager_GetSession_IdleTimeout(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	config.IdleTimeout = 1 * time.Millisecond // Very short idle timeout
	config.AbsoluteTimeout = 1 * time.Hour    // Long absolute timeout
	sm := NewSessionManager(repo.DB(), config)

	// Create session
	session, err := sm.CreateSession()
	require.NoError(t, err)

	// Wait for idle timeout
	time.Sleep(10 * time.Millisecond)

	// Try to retrieve session after idle timeout
	_, err = sm.GetSession(session.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inactivity")
}

func TestSessionManager_DeleteSession(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create session
	session, err := sm.CreateSession()
	require.NoError(t, err)

	// Delete it
	err = sm.DeleteSession(session.ID)
	require.NoError(t, err)

	// Try to retrieve deleted session
	_, err = sm.GetSession(session.ID)
	assert.Error(t, err)
}

func TestSessionManager_RefreshSession(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	config.AbsoluteTimeout = 1 * time.Hour // Long timeout
	config.IdleTimeout = 1 * time.Hour     // Long idle timeout
	sm := NewSessionManager(repo.DB(), config)

	// Create session
	session, err := sm.CreateSession()
	require.NoError(t, err)

	// Truncate to second precision (RFC3339 loses microseconds)
	originalExpiry := session.ExpiresAt.Truncate(time.Second)

	// Wait enough time to ensure a full second difference
	time.Sleep(1100 * time.Millisecond)

	// Refresh session
	err = sm.RefreshSession(session.ID)
	require.NoError(t, err)

	// Query database directly to check expiry (avoid GetSession which may update timestamps)
	query := `SELECT expires_at FROM sessions WHERE id = ?`
	var expiresAtStr string
	err = repo.DB().QueryRow(query, session.ID).Scan(&expiresAtStr)
	require.NoError(t, err)

	newExpiry, err := time.Parse(time.RFC3339, expiresAtStr)
	require.NoError(t, err)

	// Expiry should be updated to be at least 1 second later
	assert.True(t, newExpiry.After(originalExpiry),
		"New expiry %v should be after original %v", newExpiry, originalExpiry)

	// Should be roughly 1 hour in the future from now
	assert.WithinDuration(t, time.Now().Add(config.AbsoluteTimeout), newExpiry, 2*time.Second)
}

func TestSessionManager_GetOrCreateSession_NoExisting(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create request without session cookie
	req := httptest.NewRequest("GET", "/", nil)

	session, isNew, err := sm.GetOrCreateSession(req)
	require.NoError(t, err)
	assert.True(t, isNew)
	assert.NotEmpty(t, session.ID)
}

func TestSessionManager_GetOrCreateSession_Existing(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create a session first
	existingSession, err := sm.CreateSession()
	require.NoError(t, err)

	// Create request with session cookie
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: existingSession.ID,
	})

	session, isNew, err := sm.GetOrCreateSession(req)
	require.NoError(t, err)
	assert.False(t, isNew)
	assert.Equal(t, existingSession.ID, session.ID)
}

func TestSessionManager_SetSessionCookie(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create session
	session, err := sm.CreateSession()
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set cookie
	sm.SetSessionCookie(w, session)

	// Check response
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, config.CookieName, cookie.Name)
	assert.Equal(t, session.ID, cookie.Value)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, config.SecureCookie, cookie.Secure)
	assert.Equal(t, config.SameSite, cookie.SameSite)
}

func TestSessionManager_ClearSessionCookie(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create response recorder
	w := httptest.NewRecorder()

	// Clear cookie
	sm.ClearSessionCookie(w)

	// Check response
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, config.CookieName, cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
}

func TestSessionMiddleware(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create a handler that checks for session
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Try to get session ID from request
		sessionID, err := GetSessionID(r)
		assert.NoError(t, err)
		assert.NotEmpty(t, sessionID)
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with session middleware
	middleware := SessionMiddleware(sm)
	wrappedHandler := middleware(handler)

	// Create request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Check results
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that cookie was set
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, config.CookieName, cookies[0].Name)
}

func TestGetSessionID(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	sm := NewSessionManager(repo.DB(), config)

	// Create session
	session, err := sm.CreateSession()
	require.NoError(t, err)

	t.Run("from cookie", func(_ *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  config.CookieName,
			Value: session.ID,
		})

		sessionID, err := GetSessionID(req)
		require.NoError(t, err)
		assert.Equal(t, session.ID, sessionID)
	})

	t.Run("no session", func(_ *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		_, err := GetSessionID(req)
		assert.Error(t, err)
	})
}

func TestSessionManager_CleanupExpired(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	config := DefaultSessionConfig()
	config.SecureCookie = false
	config.AbsoluteTimeout = 1 * time.Millisecond
	sm := NewSessionManager(repo.DB(), config)

	// Create a session
	session, err := sm.CreateSession()
	require.NoError(t, err)

	// Wait for it to expire
	time.Sleep(10 * time.Millisecond)

	// Manually trigger cleanup
	deleteQuery := `DELETE FROM sessions WHERE expires_at < ?`
	now := time.Now().Format(time.RFC3339)
	_, err = repo.DB().Exec(deleteQuery, now)
	require.NoError(t, err)

	// Try to retrieve - should fail
	_, err = sm.GetSession(session.ID)
	assert.Error(t, err)
}

func TestSessionSecurityAttributes(t *testing.T) {
	config := DefaultSessionConfig()

	// Verify secure defaults
	assert.Equal(t, 1*time.Hour, config.IdleTimeout)
	assert.Equal(t, 7*24*time.Hour, config.AbsoluteTimeout)
	assert.Equal(t, "session_id", config.CookieName)
	assert.Equal(t, "/", config.CookiePath)
	assert.True(t, config.SecureCookie)
	assert.Equal(t, http.SameSiteStrictMode, config.SameSite)
}
