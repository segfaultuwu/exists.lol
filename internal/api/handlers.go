package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/segfaultuwu/exists.lol/internal/registry"
	"github.com/segfaultuwu/exists.lol/internal/service"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status: "healthy",
		OK:     true,
	})
}

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
	stats := s.service.GetStats()

	writeJSON(w, http.StatusOK, StatsResponse{
		DomainsTotal: stats.DomainsTotal,
		Active:       stats.Active,
		Suspended:    stats.Suspended,
	})
}

// handleCreateDomain creates a new domain via GitHub pull request
func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	var req CreateDomainRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Use service to create domain via PR
	svcReq := service.CreateDomainRequest{
		Subdomain:      req.Subdomain,
		Username:       req.Username,
		GitHubUsername: req.GitHubUsername,
		DiscordID:      req.DiscordID,
		Records:        req.Records,
		PRTitle:        req.PRTitle,
		PRDescription:  req.PRDescription,
	}

	response, err := s.service.CreateDomain(r.Context(), svcReq)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Convert service response to API response
	apiResponse := CreateDomainResponse{
		OK:        response.Success,
		Subdomain: response.Subdomain,
		FQDN:      response.FQDN,
		Message:   response.Message,
	}

	if response.PullRequest != nil {
		apiResponse.PullRequest = &PRInfo{
			URL:    response.PullRequest.URL,
			Number: response.PullRequest.Number,
			Title:  response.PullRequest.Title,
			Branch: response.PullRequest.Branch,
		}
	}

	if response.Error != "" {
		apiResponse.Error = response.Error
	}

	statusCode := http.StatusCreated
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	writeJSON(w, statusCode, apiResponse)
}

func (s *Server) handleReloadRegistry(w http.ResponseWriter, r *http.Request) {
	if err := s.service.ReloadRegistry(s.registryDir); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ReloadRegistryResponse{
		OK:      true,
		Domains: s.service.GetStats().DomainsTotal,
	})
}

func (s *Server) handleListDomains(w http.ResponseWriter, r *http.Request) {
	domains := s.service.ListDomains()

	out := make([]DomainResponse, 0, len(domains))

	for _, domain := range domains {
		out = append(out, DomainResponse{
			Subdomain: domain.Subdomain,
			FQDN:      domain.FQDN,
			Records:   domain.Records,
			Status:    domain.Status,
			Owner: Owner{
				Username: domain.Owner.Username,
				Discord:  domain.Owner.DiscordID,
				GitHub:   domain.Owner.GitHubUsername,
			},
		})
	}

	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetDomain(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.TrimSpace(chi.URLParam(r, "subdomain"))

	if subdomain == "" {
		writeError(w, http.StatusBadRequest, "missing subdomain")
		return
	}

	domain, ok := s.service.GetDomain(subdomain)
	if !ok {
		writeError(w, http.StatusNotFound, "domain not found")
		return
	}

	writeJSON(w, http.StatusOK, DomainResponse{
		Subdomain: domain.Subdomain,
		FQDN:      domain.FQDN,
		Records:   domain.Records,
		Status:    domain.Status,
		Owner: Owner{
			Username: domain.Owner.Username,
			Discord:  domain.Owner.DiscordID,
			GitHub:   domain.Owner.GitHubUsername,
		},
	})
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	var errors []string

	req.Subdomain = normalizeSubdomain(req.Subdomain)
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

	if s.registry.Contains(req.Subdomain) {
		errors = append(errors, "subdomain already exists")
	}

	switch req.Type {
	case "A", "AAAA", "CNAME", "TXT", "MX", "REDIRECT":
	default:
		errors = append(errors, "unsupported record type")
	}

	writeJSON(w, http.StatusOK, ValidateResponse{
		OK:     len(errors) == 0,
		Errors: errors,
	})
}

// Legacy handlers for backwards compatibility

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

func normalizeSubdomain(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	value = strings.TrimSuffix(value, ".exists.lol")
	value = strings.ToLower(value)
	return value
}

func normalizeRecords(records map[string][]string) map[string][]string {
	out := make(map[string][]string, len(records))

	for recordType, values := range records {
		recordType = strings.ToUpper(strings.TrimSpace(recordType))
		if recordType == "" {
			continue
		}

		cleanValues := make([]string, 0, len(values))

		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}

			cleanValues = append(cleanValues, value)
		}

		if len(cleanValues) > 0 {
			out[recordType] = cleanValues
		}
	}

	return out
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
