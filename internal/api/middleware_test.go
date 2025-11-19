package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		expectedIP    string
	}{
		{
			name:       "Direct connection without proxy",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For single IP",
			remoteAddr:    "10.0.0.1:8080",
			xForwardedFor: "203.0.113.195",
			expectedIP:    "203.0.113.195",
		},
		{
			name:          "X-Forwarded-For multiple IPs",
			remoteAddr:    "10.0.0.1:8080",
			xForwardedFor: "203.0.113.195, 70.41.3.18, 150.172.238.178",
			expectedIP:    "203.0.113.195",
		},
		{
			name:          "X-Forwarded-For with spaces",
			remoteAddr:    "10.0.0.1:8080",
			xForwardedFor: "  203.0.113.195  ,  70.41.3.18  ",
			expectedIP:    "203.0.113.195",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "10.0.0.1:8080",
			xRealIP:    "203.0.113.195",
			expectedIP: "203.0.113.195",
		},
		{
			name:          "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr:    "10.0.0.1:8080",
			xForwardedFor: "203.0.113.195",
			xRealIP:       "198.51.100.178",
			expectedIP:    "203.0.113.195",
		},
		{
			name:          "Invalid X-Forwarded-For falls back to X-Real-IP",
			remoteAddr:    "10.0.0.1:8080",
			xForwardedFor: "not-an-ip",
			xRealIP:       "203.0.113.195",
			expectedIP:    "203.0.113.195",
		},
		{
			name:       "Invalid headers fall back to RemoteAddr",
			remoteAddr: "192.168.1.1:12345",
			xRealIP:    "not-an-ip",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "IPv6 address",
			remoteAddr: "[2001:db8::1]:8080",
			expectedIP: "2001:db8::1",
		},
		{
			name:          "IPv6 in X-Forwarded-For",
			remoteAddr:    "10.0.0.1:8080",
			xForwardedFor: "2001:db8::1",
			expectedIP:    "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %v, want %v", ip, tt.expectedIP)
			}
		})
	}
}

func TestRateLimitMiddleware_WithProxyHeaders(t *testing.T) {
	limiter := NewRateLimiter(10, 10) // 10 requests per minute, burst of 10

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test that requests from different IPs via X-Forwarded-For are tracked separately
	t.Run("Different IPs via X-Forwarded-For", func(t *testing.T) {
		// Request from IP 1
		req1 := httptest.NewRequest("GET", "/", nil)
		req1.RemoteAddr = "10.0.0.1:8080" // Proxy IP (same for both)
		req1.Header.Set("X-Forwarded-For", "203.0.113.1")
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)

		if rr1.Code != http.StatusOK {
			t.Errorf("First request failed with status %d", rr1.Code)
		}

		// Request from IP 2 (different client, same proxy)
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.RemoteAddr = "10.0.0.1:8080" // Same proxy IP
		req2.Header.Set("X-Forwarded-For", "203.0.113.2")
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		if rr2.Code != http.StatusOK {
			t.Errorf("Second request failed with status %d", rr2.Code)
		}
	})

	// Test that the same IP is rate limited
	t.Run("Same IP is rate limited", func(t *testing.T) {
		limiter := NewRateLimiter(2, 2) // Very low limit for testing
		handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		clientIP := "203.0.113.100"

		// First two requests should succeed
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "10.0.0.1:8080"
			req.Header.Set("X-Forwarded-For", clientIP)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Request %d should succeed, got status %d", i+1, rr.Code)
			}
		}

		// Third request should be rate limited
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		req.Header.Set("X-Forwarded-For", clientIP)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Third request should be rate limited, got status %d", rr.Code)
		}
	})
}
