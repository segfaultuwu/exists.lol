package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/segfaultuwu/exists.lol/internal/registry"
)

type Server struct {
	addr       string
	registry   *registry.Registry
	baseDomain string
}

func New(host string, port int, baseDomain string, reg *registry.Registry) *Server {
	return &Server{
		addr:       fmt.Sprintf("%s:%d", host, port),
		registry:   reg,
		baseDomain: baseDomain,
	}
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

	return http.ListenAndServe(s.addr, r)
}
