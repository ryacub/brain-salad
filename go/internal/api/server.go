package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
)

// Server represents the API server
type Server struct {
	repo   *database.Repository
	telos  *models.Telos
	router *chi.Mux
}

// NewServer creates a new API server
func NewServer(repo *database.Repository, telosPath string) (*Server, error) {
	// Load telos configuration
	telosData, err := loadTelos(telosPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load telos: %w", err)
	}

	s := &Server{
		repo:  repo,
		telos: telosData,
	}

	s.setupRouter()

	return s, nil
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

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Get("/health", s.HealthHandler)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
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

// Close closes the database connection
func (s *Server) Close() error {
	return s.repo.Close()
}
