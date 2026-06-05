package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/segfaultuwu/exists.lol/internal/registry"
)

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, StatusResponse{
		OK:       true,
		Service:  "exists.lol-api",
		Registry: "loaded",
	})
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://docs.exists.lol", http.StatusFound)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	domains := s.registry.All()

	writeJSON(w, http.StatusOK, StatsResponse{
		DomainsTotal: len(domains),
		Active:       len(domains),
		Suspended:    0,
	})
}

func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	var req CreateDomainRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Subdomain = strings.ToLower(strings.TrimSpace(req.Subdomain))

	if req.Subdomain == "" {
		writeError(w, http.StatusBadRequest, "subdomain is required")
		return
	}

	if len(req.Records) == 0 {
		writeError(w, http.StatusBadRequest, "records are required")
		return
	}

	if _, exists := s.registry.Get(req.Subdomain); exists {
		writeError(w, http.StatusConflict, "subdomain already exists")
		return
	}

	domain := registry.DomainFile{
		Owner: registry.Owner{
			Username:       req.Username,
			GitHubUsername: req.GitHubUsername,
			DiscordID:      req.DiscordID,
		},
		Records: req.Records,
	}

	for recordType, values := range req.Records {
		recordType = strings.ToUpper(strings.TrimSpace(recordType))

		switch recordType {
		case "A", "AAAA", "CNAME", "TXT":
		default:
			writeError(w, http.StatusBadRequest, "unsupported record type: "+recordType)
			return
		}

		if len(values) == 0 {
			writeError(w, http.StatusBadRequest, "record has no values: "+recordType)
			return
		}
	}

	if err := s.registry.Add(s.registryDir, req.Subdomain, domain); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CreateDomainResponse{
		OK:        true,
		Subdomain: req.Subdomain,
		FQDN:      req.Subdomain + "." + s.baseDomain,
		Message:   "domain created",
	})
}

func (s *Server) handleListDomains(w http.ResponseWriter, r *http.Request) {
	domains := s.registry.All()

	out := make([]DomainResponse, 0, len(domains))

	for subdomain, domain := range domains {
		out = append(out, s.domainResponse(subdomain, domain))
	}

	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetDomain(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.TrimSpace(chi.URLParam(r, "subdomain"))

	if subdomain == "" {
		writeError(w, http.StatusBadRequest, "missing subdomain")
		return
	}

	domain, ok := s.registry.Get(subdomain)
	if !ok {
		writeError(w, http.StatusNotFound, "domain not found")
		return
	}

	writeJSON(w, http.StatusOK, s.domainResponse(subdomain, domain))
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	var errors []string

	req.Subdomain = strings.ToLower(strings.TrimSpace(req.Subdomain))
	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	req.Value = strings.TrimSpace(req.Value)

	if req.Subdomain == "" {
		errors = append(errors, "subdomain is required")
	}

	if req.Type == "" {
		errors = append(errors, "type is required")
	}

	if req.Value == "" {
		errors = append(errors, "value is required")
	}

	if _, exists := s.registry.Get(req.Subdomain); exists {
		errors = append(errors, "subdomain already exists")
	}

	switch req.Type {
	case "A", "AAAA", "CNAME", "TXT":
	default:
		errors = append(errors, "unsupported record type")
	}

	writeJSON(w, http.StatusOK, ValidateResponse{
		OK:     len(errors) == 0,
		Errors: errors,
	})
}

func (s *Server) domainResponse(subdomain string, domain registry.DomainFile) DomainResponse {
	return DomainResponse{
		Subdomain: subdomain,
		FQDN:      subdomain + "." + s.baseDomain,
		Records:   domain.Records,
		Status:    "active",
		Owner: Owner{
			Username: domain.Owner.Username,
			Discord:  domain.Owner.DiscordID,
			GitHub:   domain.Owner.GitHubUsername,
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
