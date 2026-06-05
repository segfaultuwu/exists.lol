package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/segfaultuwu/exists.lol/internal/registry"
)

type Server struct {
	addr        string
	registry    *registry.Registry
	baseDomain  string
	registryDir string
}

func New(host string, port int, baseDomain string, registryDir string, reg *registry.Registry) *Server {
	return &Server{
		addr:        fmt.Sprintf("%s:%d", host, port),
		registry:    reg,
		baseDomain:  baseDomain,
		registryDir: registryDir,
	}
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+s.registry.Token() {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Get("/", s.handleRoot)
	r.Get("/api/status", s.handleStatus)
	r.Get("/api/stats", s.handleStats)
	r.Get("/api/domains", s.handleListDomains)
	r.Get("/api/domains/{subdomain}", s.handleGetDomain)
	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		r.Post("/api/domains", s.handleCreateDomain)
	})

	return http.ListenAndServe(s.addr, r)
}
