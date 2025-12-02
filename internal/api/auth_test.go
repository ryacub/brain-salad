package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryacub/telos-idea-matrix/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestAuthMiddleware_Disabled tests that auth middleware allows all requests when disabled
func TestAuthMiddleware_Disabled(t *testing.T) {
	cfg := config.DefaultAuthConfig() // disabled by default

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/ideas", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should allow request without auth when disabled
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

// TestAuthMiddleware_Enabled_ValidKey tests successful authentication with valid API key
func TestAuthMiddleware_Enabled_ValidKey(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{
			"test-key-123": "Test Client",
		},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/ideas", nil)
	req.Header.Set("Authorization", "Bearer test-key-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

// TestAuthMiddleware_Enabled_InvalidKey tests failed authentication with invalid API key
func TestAuthMiddleware_Enabled_InvalidKey(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{
			"test-key-123": "Test Client",
		},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/ideas", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid API key")
}

// TestAuthMiddleware_Enabled_MissingHeader tests missing Authorization header
func TestAuthMiddleware_Enabled_MissingHeader(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{
			"test-key-123": "Test Client",
		},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/ideas", nil)
	// No Authorization header set
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Missing authorization header")
}

// TestAuthMiddleware_Enabled_InvalidHeaderFormat tests invalid Authorization header format
func TestAuthMiddleware_Enabled_InvalidHeaderFormat(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{
			"test-key-123": "Test Client",
		},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	testCases := []struct {
		name        string
		authHeader  string
		description string
	}{
		{
			name:        "NoBearer",
			authHeader:  "test-key-123",
			description: "API key without Bearer prefix",
		},
		{
			name:        "WrongScheme",
			authHeader:  "Basic dGVzdDoxMjM=",
			description: "Basic auth instead of Bearer",
		},
		{
			name:        "EmptyBearer",
			authHeader:  "Bearer",
			description: "Bearer with no key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/ideas", nil)
			req.Header.Set("Authorization", tc.authHeader)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code, tc.description)
			assert.Contains(t, rec.Body.String(), "Invalid authorization header format", tc.description)
		})
	}
}

// TestAuthMiddleware_HealthCheckBypass tests that health checks bypass authentication
func TestAuthMiddleware_HealthCheckBypass(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}))

	testCases := []struct {
		name string
		path string
	}{
		{
			name: "HealthEndpoint",
			path: "/health",
		},
		{
			name: "MetricsEndpoint",
			path: "/metrics",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			// No Authorization header
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Should allow without auth
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "healthy", rec.Body.String())
		})
	}
}

// TestAuthMiddleware_MultipleValidKeys tests multiple valid API keys
func TestAuthMiddleware_MultipleValidKeys(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{
			"prod-key-abc": "Production Client",
			"dev-key-xyz":  "Development Client",
			"test-key-123": "Test Client",
		},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	testCases := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "ProductionKey",
			apiKey: "prod-key-abc",
		},
		{
			name:   "DevelopmentKey",
			apiKey: "dev-key-xyz",
		},
		{
			name:   "TestKey",
			apiKey: "test-key-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/ideas", nil)
			req.Header.Set("Authorization", "Bearer "+tc.apiKey)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "success", rec.Body.String())
		})
	}
}

// TestAuthMiddleware_DifferentMethods tests auth with different HTTP methods
func TestAuthMiddleware_DifferentMethods(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{
			"test-key-123": "Test Client",
		},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/ideas", nil)
			req.Header.Set("Authorization", "Bearer test-key-123")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "success", rec.Body.String())
		})
	}
}

// TestAuthMiddleware_CaseSensitiveKeys tests that API keys are case-sensitive
func TestAuthMiddleware_CaseSensitiveKeys(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		Mode:    "api-key",
		APIKeys: map[string]string{
			"test-key-123": "Test Client",
		},
	}

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	// Try uppercase version of key - should fail
	req := httptest.NewRequest("GET", "/api/v1/ideas", nil)
	req.Header.Set("Authorization", "Bearer TEST-KEY-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid API key")
}
