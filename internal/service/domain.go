package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/segfaultuwu/exists.lol/internal/github"
	"github.com/segfaultuwu/exists.lol/internal/registry"
)

// DomainService handles domain creation and management business logic
type DomainService struct {
	ghClient   *github.Client
	registry   *registry.Registry
	baseDomain string
	rootDomain string
}

// DomainServiceConfig contains configuration for the domain service
type DomainServiceConfig struct {
	GitHubToken string
	GitHubOwner string
	GitHubRepo  string
	BaseDomain  string
	RootDomain  string
}

// NewDomainService creates a new domain service
func NewDomainService(cfg DomainServiceConfig, reg *registry.Registry) *DomainService {
	ghClient := github.New(cfg.GitHubToken, cfg.GitHubOwner, cfg.GitHubRepo)

	return &DomainService{
		ghClient:   ghClient,
		registry:   reg,
		baseDomain: cfg.BaseDomain,
		rootDomain: cfg.RootDomain,
	}
}

// CreateDomainRequest represents a request to create a new domain
type CreateDomainRequest struct {
	Subdomain      string              `json:"subdomain"`
	Username       string              `json:"username"`
	GitHubUsername string              `json:"github_username"`
	DiscordID      string              `json:"discord_id"`
	Records        map[string][]string `json:"records"`
	PRTitle        string              `json:"pr_title,omitempty"`
	PRDescription  string              `json:"pr_description,omitempty"`
}

// CreateDomainResponse represents the response from creating a domain
type CreateDomainResponse struct {
	Success     bool    `json:"success"`
	Subdomain   string  `json:"subdomain"`
	FQDN        string  `json:"fqdn"`
	Message     string  `json:"message"`
	PullRequest *PRInfo `json:"pull_request,omitempty"`
	Error       string  `json:"error,omitempty"`
}

// PRInfo contains information about a created pull request
type PRInfo struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Branch string `json:"branch"`
}

// CreateDomain creates a new domain via GitHub pull request
func (s *DomainService) CreateDomain(ctx context.Context, req CreateDomainRequest) (*CreateDomainResponse, error) {
	// Normalize subdomain
	subdomain := normalizeSubdomain(req.Subdomain)
	if subdomain == "" {
		return nil, fmt.Errorf("subdomain is required")
	}

	// Validate required fields
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.GitHubUsername == "" {
		return nil, fmt.Errorf("github_username is required")
	}
	if req.DiscordID == "" {
		return nil, fmt.Errorf("discord_id is required")
	}
	if len(req.Records) == 0 {
		return nil, fmt.Errorf("records are required")
	}

	// Check if domain already exists
	if s.registry.Contains(subdomain) {
		return nil, fmt.Errorf("subdomain %q already exists", subdomain)
	}

	// Normalize records
	normalizedRecords := registry.NormalizeRecords(req.Records)
	if len(normalizedRecords) == 0 {
		return nil, fmt.Errorf("no valid records after normalization")
	}

	// Create domain file
	owner := registry.Owner{
		Username:       strings.TrimSpace(req.Username),
		GitHubUsername: strings.TrimSpace(req.GitHubUsername),
		DiscordID:      strings.TrimSpace(req.DiscordID),
	}

	// Validate domain configuration
	if err := registry.ValidateDomainFile(subdomain, registry.DomainFile{
		Owner:   owner,
		Records: normalizedRecords,
	}); err != nil {
		return nil, fmt.Errorf("invalid domain configuration: %w", err)
	}

	// Create GitHub PR
	ghReq := github.DomainRequest{
		Subdomain:     subdomain,
		Owner:         owner,
		Records:       normalizedRecords,
		PRTitle:       req.PRTitle,
		PRDescription: req.PRDescription,
	}

	prResult, err := s.ghClient.CreateDomainPR(ctx, ghReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	// Build response
	fqdn := subdomain + "." + s.getRootDomain()

	return &CreateDomainResponse{
		Success:   true,
		Subdomain: subdomain,
		FQDN:      fqdn,
		Message:   "Domain creation requested via pull request",
		PullRequest: &PRInfo{
			URL:    prResult.URL,
			Number: prResult.Number,
			Title:  prResult.Title,
			Branch: prResult.Branch,
		},
	}, nil
}

// GetDomain retrieves a domain by subdomain
func (s *DomainService) GetDomain(subdomain string) (*DomainInfo, bool) {
	domain, ok := s.registry.Get(normalizeSubdomain(subdomain))
	if !ok {
		return nil, false
	}

	return &DomainInfo{
		Subdomain: subdomain,
		FQDN:      subdomain + "." + s.getRootDomain(),
		Owner: OwnerInfo{
			Username:       domain.Owner.Username,
			GitHubUsername: domain.Owner.GitHubUsername,
			DiscordID:      domain.Owner.DiscordID,
		},
		Records: domain.Records,
		Status:  "active",
	}, true
}

// ListDomains returns all domains
func (s *DomainService) ListDomains() []DomainInfo {
	all := s.registry.All()
	result := make([]DomainInfo, 0, len(all))

	for subdomain, domain := range all {
		result = append(result, DomainInfo{
			Subdomain: subdomain,
			FQDN:      subdomain + "." + s.getRootDomain(),
			Owner: OwnerInfo{
				Username:       domain.Owner.Username,
				GitHubUsername: domain.Owner.GitHubUsername,
				DiscordID:      domain.Owner.DiscordID,
			},
			Records: domain.Records,
			Status:  "active",
		})
	}

	return result
}

// GetDomainsByDiscordID returns all domains owned by a Discord user
func (s *DomainService) GetDomainsByDiscordID(discordID string) []DomainInfo {
	all := s.registry.ByDiscordID(discordID)
	result := make([]DomainInfo, 0, len(all))

	for subdomain, domain := range all {
		result = append(result, DomainInfo{
			Subdomain: subdomain,
			FQDN:      subdomain + "." + s.getRootDomain(),
			Owner: OwnerInfo{
				Username:       domain.Owner.Username,
				GitHubUsername: domain.Owner.GitHubUsername,
				DiscordID:      domain.Owner.DiscordID,
			},
			Records: domain.Records,
			Status:  "active",
		})
	}

	return result
}

// ValidateDomain validates a domain configuration without creating it
func (s *DomainService) ValidateDomain(subdomain string, owner registry.Owner, records map[string][]string) error {
	subdomain = normalizeSubdomain(subdomain)
	if subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}

	if s.registry.Contains(subdomain) {
		return fmt.Errorf("subdomain %q already exists", subdomain)
	}

	normalizedRecords := registry.NormalizeRecords(records)
	if len(normalizedRecords) == 0 {
		return fmt.Errorf("no valid records")
	}

	return registry.ValidateDomainFile(subdomain, registry.DomainFile{
		Owner:   owner,
		Records: normalizedRecords,
	})
}

// GetStats returns statistics about the registry
func (s *DomainService) GetStats() Stats {
	return Stats{
		DomainsTotal: s.registry.Count(),
		Active:       s.registry.Count(),
		Suspended:    0,
	}
}

// ReloadRegistry reloads the registry from disk
func (s *DomainService) ReloadRegistry(dir string) error {
	return s.registry.Reload(dir)
}

// getRootDomain returns the root domain, ensuring it doesn't end with a dot
func (s *DomainService) getRootDomain() string {
	if s.rootDomain != "" {
		return strings.TrimSuffix(s.rootDomain, ".")
	}
	return strings.TrimSuffix(s.baseDomain, ".")
}

// DomainInfo represents a domain with its information
type DomainInfo struct {
	Subdomain string              `json:"subdomain"`
	FQDN      string              `json:"fqdn"`
	Records   map[string][]string `json:"records"`
	Owner     OwnerInfo           `json:"owner"`
	Status    string              `json:"status"`
}

// OwnerInfo represents the owner of a domain
type OwnerInfo struct {
	Username       string `json:"username,omitempty"`
	GitHubUsername string `json:"github,omitempty"`
	DiscordID      string `json:"discord,omitempty"`
}

// Stats represents registry statistics
type Stats struct {
	DomainsTotal int `json:"domains_total"`
	Active       int `json:"active"`
	Suspended    int `json:"suspended"`
}

// normalizeSubdomain normalizes a subdomain string
func normalizeSubdomain(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	value = strings.TrimSuffix(value, ".exists.lol")
	value = strings.ToLower(value)
	return value
}
