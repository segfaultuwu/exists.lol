package api

// DomainResponse represents a domain in the API response
type DomainResponse struct {
	Subdomain string              `json:"subdomain"`
	FQDN      string              `json:"fqdn"`
	Records   map[string][]string `json:"records"`
	Owner     Owner               `json:"owner"`
	Status    string              `json:"status"`
}

// Owner represents the owner of a domain in API responses
type Owner struct {
	Username string `json:"username,omitempty"`
	Discord  string `json:"discord,omitempty"`
	GitHub   string `json:"github,omitempty"`
}

// StatsResponse represents registry statistics
type StatsResponse struct {
	DomainsTotal int `json:"domains_total"`
	Active       int `json:"active"`
	Suspended    int `json:"suspended"`
}

// StatusResponse represents the API status
type StatusResponse struct {
	OK       bool   `json:"ok"`
	Service  string `json:"service"`
	Registry string `json:"registry"`
}

// ValidateRequest represents a validation request
type ValidateRequest struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

// ValidateResponse represents a validation response
type ValidateResponse struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

// CreateDomainRequest represents a request to create a new domain
type CreateDomainRequest struct {
	DiscordID      string              `json:"discord_id"`
	Username       string              `json:"username"`
	GitHubUsername string              `json:"github_username"`
	Subdomain      string              `json:"subdomain"`
	Records        map[string][]string `json:"records"`
	PRTitle        string              `json:"pr_title,omitempty"`       // Optional custom PR title
	PRDescription  string              `json:"pr_description,omitempty"` // Optional custom PR description
}

// CreateDomainResponse represents the response from creating a domain
type CreateDomainResponse struct {
	OK          bool    `json:"ok"`
	Subdomain   string  `json:"subdomain"`
	FQDN        string  `json:"fqdn"`
	Message     string  `json:"message"`
	PullRequest *PRInfo `json:"pull_request,omitempty"` // Information about the created PR
	Error       string  `json:"error,omitempty"`        // Error message if creation failed
}

// PRInfo contains information about a created pull request
type PRInfo struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Branch string `json:"branch"`
}

// ReloadRegistryResponse represents the response from reloading the registry
type ReloadRegistryResponse struct {
	OK      bool `json:"ok"`
	Domains int  `json:"domains"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
	OK     bool   `json:"ok"`
}
