package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/health"
	"github.com/rayyacub/telos-idea-matrix/internal/logging"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
)

// Server represents the API server
type Server struct {
	repo           *database.Repository
	telos          *models.Telos
	router         *chi.Mux
	cache          *Cache
	rateLimiter    *RateLimiter
	csrfProtection *CSRFProtection
	sessionManager *SessionManager
	healthMonitor  *health.HealthMonitor
}

// NewServer creates a new API server from a telos configuration object
func NewServer(repo *database.Repository, telosConfig *models.Telos) *Server {
	// Create health monitor and register checks
	healthMonitor := health.NewHealthMonitor()
	healthMonitor.SetVersion("1.0.0")

	// Add database health checker
	healthMonitor.AddCheck(health.NewDatabaseHealthChecker(repo.DB()))

	// Add memory health checker (warn if using > 500MB)
	healthMonitor.AddCheck(health.NewMemoryHealthChecker(500.0))

	// Add disk space health checker (warn if < 1GB free)
	healthMonitor.AddCheck(health.NewDiskSpaceHealthChecker("/tmp", 1024))

	// Create session manager with secure configuration
	sessionConfig := DefaultSessionConfig()
	// Set SecureCookie based on environment
	// In production (ENV=production), use secure cookies (HTTPS only)
	// In development, allow non-HTTPS cookies for easier testing
	env := strings.ToLower(os.Getenv("ENV"))
	sessionConfig.SecureCookie = (env == "production" || env == "prod")
	sessionManager := NewSessionManager(repo.DB(), sessionConfig)

	s := &Server{
		repo:           repo,
		telos:          telosConfig,
		cache:          NewCache(5 * time.Minute),       // 5-minute cache TTL
		rateLimiter:    NewRateLimiter(100, 10),         // 100 req/min, burst of 10
		csrfProtection: NewCSRFProtection(1 * time.Hour), // 1-hour token TTL
		sessionManager: sessionManager,
		healthMonitor:  healthMonitor,
	}

	s.setupRouter()

	return s
}

// NewServerFromPath creates a new API server from a telos file path
func NewServerFromPath(repo *database.Repository, telosPath string) (*Server, error) {
	// Load telos configuration
	telosData, err := loadTelos(telosPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load telos: %w", err)
	}

	return NewServer(repo, telosData), nil
}

// loadTelos loads and parses the telos configuration file
func loadTelos(path string) (*models.Telos, error) {
	parser := telos.NewParser()
	telosData, err := parser.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse telos: %w", err)
	}

	return telosData, nil
}

// setupRouter configures the Chi router with all routes and middleware
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware (order matters!)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logging.Middleware) // Structured logging middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Session middleware - must come early so session is available for other middleware
	r.Use(SessionMiddleware(s.sessionManager))

	// Security middleware
	r.Use(SecurityHeadersMiddleware)
	r.Use(RateLimitMiddleware(s.rateLimiter))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Cache", "X-RateLimit-Limit"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Caching middleware (only for GET requests)
	r.Use(CacheMiddleware(s.cache))

	// Routes
	r.Get("/health", s.HealthHandler)
	r.Get("/metrics", s.MetricsHandler)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// CSRF token endpoint (for clients that need it)
		r.Get("/csrf-token", s.GetCSRFTokenHandler)

		// Ideas
		r.Post("/ideas", s.CreateIdeaHandler)
		r.Get("/ideas", s.ListIdeasHandler)
		r.Get("/ideas/{id}", s.GetIdeaHandler)
		r.Put("/ideas/{id}", s.UpdateIdeaHandler)
		r.Delete("/ideas/{id}", s.DeleteIdeaHandler)

		// Analysis
		r.Post("/analyze", s.AnalyzeHandler)

		// Analytics
		r.Get("/analytics/stats", s.AnalyticsStatsHandler)
	})

	s.router = r
}

// Router returns the configured Chi router
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// Close closes the database connection and stops background goroutines
func (s *Server) Close() error {
	// Stop background cleanup goroutines
	s.cache.Stop()
	s.rateLimiter.Stop()
	s.sessionManager.Stop()

	// Close database connection
	return s.repo.Close()
}
