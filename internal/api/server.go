package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/segfaultuwu/exists.lol/internal/registry"
	"github.com/segfaultuwu/exists.lol/internal/service"
)

// Server represents the API HTTP server
type Server struct {
	addr        string
	service     *service.DomainService
	registry    *registry.Registry
	baseDomain  string
	registryDir string
	apiToken    string
}

// Config contains configuration for the API server
type Config struct {
	Host        string
	Port        int
	BaseDomain  string
	RootDomain  string
	RegistryDir string
	APIToken    string
	GitHubToken string
	GitHubOwner string
	GitHubRepo  string
}

// New creates a new API server
func New(cfg Config, reg *registry.Registry) *Server {
	// Create domain service
	svcConfig := service.DomainServiceConfig{
		GitHubToken: cfg.GitHubToken,
		GitHubOwner: cfg.GitHubOwner,
		GitHubRepo:  cfg.GitHubRepo,
		BaseDomain:  cfg.BaseDomain,
		RootDomain:  cfg.RootDomain,
	}

	svc := service.NewDomainService(svcConfig, reg)

	return &Server{
		addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		service:     svc,
		registry:    reg,
		baseDomain:  cfg.BaseDomain,
		registryDir: cfg.RegistryDir,
		apiToken:    cfg.APIToken,
	}
}

// NewLegacy creates a server with the old configuration format (for backwards compatibility)
// Note: This doesn't support GitHub PR creation, only local file storage
func NewLegacy(host string, port int, baseDomain string, registryDir string, reg *registry.Registry) *Server {
	return &Server{
		addr:        fmt.Sprintf("%s:%d", host, port),
		registry:    reg,
		baseDomain:  baseDomain,
		registryDir: registryDir,
		apiToken:    reg.Token(),
		// service is nil - legacy mode doesn't support PR creation
	}
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check
		if strings.HasPrefix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+s.apiToken {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %s %s (%v)\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

func (s *Server) timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// Start launches the HTTP server
func (s *Server) Start() error {
	r := chi.NewRouter()

	// Apply middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	r.Use(s.loggingMiddleware)
	r.Use(s.timeoutMiddleware(30 * time.Second))

	// Health check (public)
	r.Get("/health", s.handleHealth)

	// Root redirect
	r.Get("/", s.handleRoot)

	// Public endpoints
	r.Get("/api/status", s.handleStatus)
	r.Get("/api/stats", s.handleStats)
	r.Get("/api/domains", s.handleListDomains)
	r.Get("/api/domains/{subdomain}", s.handleGetDomain)

	// Protected endpoints
	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		// Domain creation now creates a PR instead of local file
		r.Post("/api/domains", s.handleCreateDomain)
		r.Post("/api/validate", s.handleValidate)
		r.Post("/api/registry/reload", s.handleReloadRegistry)
	})

	fmt.Printf("API server listening on %s\n", s.addr)
	return http.ListenAndServe(s.addr, r)
}
