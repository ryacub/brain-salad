package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// CacheEntry represents a cached HTTP response
type CacheEntry struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	CachedAt    time.Time
	TTL         time.Duration
}

// IsExpired checks if the cache entry has expired
func (c *CacheEntry) IsExpired() bool {
	return time.Since(c.CachedAt) > c.TTL
}

// Cache is a simple in-memory cache for HTTP responses
type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
	stopCh  chan struct{}
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanupExpired()

	return c
}

// cleanupExpired periodically removes expired cache entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, entry := range c.entries {
				if entry.IsExpired() {
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

// Stop gracefully stops the cache cleanup goroutine
func (c *Cache) Stop() {
	close(c.stopCh)
}

// Get retrieves a cache entry by key
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	return entry, true
}

// Set stores a cache entry
func (c *Cache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = entry
}

// Clear removes all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// CacheSize returns the number of cached entries
func (c *Cache) CacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// cacheKey generates a cache key from request method and URL
func cacheKey(r *http.Request) string {
	h := sha256.New()
	h.Write([]byte(r.Method))
	h.Write([]byte(r.URL.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// responseWriter captures the response for caching
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Write to buffer for caching
	if _, err := rw.body.Write(b); err != nil {
		// If buffer write fails, still try to write to response
		// but this indicates a serious issue
		return rw.ResponseWriter.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// CacheMiddleware is a middleware that caches GET requests
func CacheMiddleware(cache *Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Check cache
			key := cacheKey(r)
			if entry, found := cache.Get(key); found {
				// Serve from cache
				for k, v := range entry.Headers {
					w.Header()[k] = v
				}
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(entry.StatusCode)
				if _, err := w.Write(entry.Body); err != nil {
					log.Warn().Err(err).Msg("failed to write cached response body")
				}
				return
			}

			// Cache miss - capture response
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			// Only cache successful responses (2xx status codes)
			if rw.statusCode >= 200 && rw.statusCode < 300 {
				entry := &CacheEntry{
					StatusCode:  rw.statusCode,
					Headers:     w.Header().Clone(),
					Body:        rw.body.Bytes(),
					CachedAt:    time.Now(),
					TTL:         cache.ttl,
				}
				cache.Set(key, entry)
			}

			w.Header().Set("X-Cache", "MISS")
		})
	}
}

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per minute
	burst    int           // max burst size
	stopCh   chan struct{}
}

type visitor struct {
	lastSeen time.Time
	tokens   float64
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		stopCh:   make(chan struct{}),
	}

	// Cleanup old visitors every 5 minutes
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors removes visitors that haven't been seen in 10 minutes
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

// Stop gracefully stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &visitor{
			lastSeen: now,
			tokens:   float64(rl.burst - 1),
		}
		return true
	}

	// Calculate tokens to add based on time elapsed
	elapsed := now.Sub(v.lastSeen)
	tokensToAdd := elapsed.Seconds() * (float64(rl.rate) / 60.0)
	v.tokens = min(v.tokens+tokensToAdd, float64(rl.burst))
	v.lastSeen = now

	if v.tokens >= 1.0 {
		v.tokens -= 1.0
		return true
	}

	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// getClientIP extracts the real client IP from the request, handling proxy headers
// It checks X-Forwarded-For and X-Real-IP headers with validation
func getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first (most common for proxies)
	// Format: X-Forwarded-For: client, proxy1, proxy2
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain (the original client)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			// Validate it's a proper IP address
			if ip := net.ParseIP(clientIP); ip != nil {
				return clientIP
			}
		}
	}

	// Try X-Real-IP header (used by some proxies like nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		// Validate it's a proper IP address
		if ip := net.ParseIP(xri); ip != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr, but strip the port
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, it might be just an IP without port
		return r.RemoteAddr
	}
	return ip
}

// RateLimitMiddleware is a middleware that rate limits requests
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get real client IP, handling proxy headers with validation
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				w.Header().Set("X-RateLimit-Limit", string(rune(limiter.rate)))
				w.Header().Set("Retry-After", "60")
				respondError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent XSS attacks
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Prevent MIME sniffing
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		next.ServeHTTP(w, r)
	})
}
